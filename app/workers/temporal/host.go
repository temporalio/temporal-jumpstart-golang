package temporal

import (
	"context"
	"fmt"
	"github.com/temporalio/temporal-jumpstart-golang/app/clients"
	"github.com/temporalio/temporal-jumpstart-golang/app/config"
	"go.temporal.io/sdk/contrib/resourcetuner"
	"golang.org/x/sync/errgroup"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.temporal.io/sdk/client"
	sdkworker "go.temporal.io/sdk/worker"
)

// WorkerFactory creates a worker for a specific client
type WorkerFactory func(ctx context.Context,
	cfg *config.Config,
	clients *clients.Clients,
	client client.Client) (sdkworker.Worker, error)

type Builder interface {
	Build(ctx context.Context, cfg *config.Config, target *WorkerClient, options sdkworker.Options) (sdkworker.Worker, error)
}

// closeClients closes all client connections for the given targets
func closeClients(targets []*WorkerClient) {
	for _, target := range targets {
		if target.Client != nil {
			target.Client.Close()
			slog.Debug("Closed client connection",
				"namespace", target.Namespace,
				"taskqueue", target.TaskQueue)
		}
	}
}

// HostConfig contains configuration for the worker host
type HostConfig struct {
	ShutdownTimeout time.Duration
	WorkerType      string // for logging
}

// WorkerHost manages the lifecycle of multiple temporal workers
type WorkerHost struct {
	config  *HostConfig
	builder Builder
}

// NewWorkerHost creates a new worker host
func NewWorkerHost(ctx context.Context,
	config *HostConfig,
	builder Builder) *WorkerHost {
	if config.ShutdownTimeout == 0 {
		config.ShutdownTimeout = 10 * time.Second
	}
	if builder == nil {
		panic("builder is required")
	}

	return &WorkerHost{
		config:  config,
		builder: builder,
	}
}
func (wh *WorkerHost) buildOptions(ctx context.Context,
	cfg *config.Config,
	client *WorkerClient) (sdkworker.Options, error) {
	wcfg := cfg.Temporal.Namespaces[client.Namespace].TaskQueues[client.TaskQueue]

	opts := sdkworker.Options{}
	// capacity Tuner versus discrete Executor config is mutually exclusive
	if wcfg.Tuner != nil && (wcfg.Tuner.TargetCPU > 0 || wcfg.Tuner.TargetMem > 0) {
		// Using the ResourceBasedTuner in worker options
		tuner, err := resourcetuner.NewResourceBasedTuner(resourcetuner.ResourceBasedTunerOptions{
			TargetMem: wcfg.Tuner.TargetMem,
			TargetCpu: wcfg.Tuner.TargetCPU,
		})
		if err != nil {
			return opts, err
		}
		opts.Tuner = tuner
	} else {
		opts.MaxConcurrentActivityExecutionSize = wcfg.Capacity.MaxConcurrentActivityTaskExecutors
		opts.MaxConcurrentWorkflowTaskExecutionSize = wcfg.Capacity.MaxConcurrentWorkflowTaskExecutors
		opts.MaxConcurrentLocalActivityExecutionSize = wcfg.Capacity.MaxConcurrentLocalActivityExecutors
		opts.MaxConcurrentActivityTaskPollers = wcfg.Capacity.MaxConcurrentActivityTaskPollers
		opts.MaxConcurrentWorkflowTaskPollers = wcfg.Capacity.MaxConcurrentWorkflowTaskPollers
	}
	if wcfg.RateLimits != nil {
		opts.WorkerActivitiesPerSecond = float64(wcfg.RateLimits.MaxWorkerActivitiesPerSecond)
		opts.TaskQueueActivitiesPerSecond = float64(wcfg.RateLimits.MaxTaskQueueActivitiesPerSecond)
	}
	return opts, nil
}

// Run starts all workers and handles graceful shutdown
func (wh *WorkerHost) Run(ctx context.Context,
	cfg *config.Config,
	clients ...*WorkerClient) error {
	g, ctx := errgroup.WithContext(ctx)

	// Set up signal handling
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(quit)

	// Create all workers
	workers := make([]sdkworker.Worker, 0)
	for i, c := range clients {
		opts, err := wh.buildOptions(ctx, cfg, c)
		if err != nil {
			return fmt.Errorf(
				"failed to build %s worker %d options: %w",
				wh.config.WorkerType, i, err)
		}
		w, err := wh.builder.Build(ctx, cfg, c, opts)
		if err != nil {
			return fmt.Errorf(
				"failed to create %s worker %d: %w",
				wh.config.WorkerType, i, err)
		}
		workers = append(workers, w)
	}

	// Create interrupt channel for temporal workers
	interrupt := make(chan interface{}, 1)

	// Start all workers
	for i, worker := range workers {
		i := i           // capture loop variable
		worker := worker // capture loop variable
		g.Go(func() error {
			slog.Info("Starting worker",
				"worker_type", wh.config.WorkerType,
				"worker_id", i)

			runErr := worker.Run(interrupt)
			if runErr != nil {
				return fmt.Errorf("%s worker %d failed: %w",
					wh.config.WorkerType, i, runErr)
			}

			slog.Info("Worker shut down gracefully",
				"worker_type", wh.config.WorkerType,
				"worker_id", i)
			return nil
		})
	}

	// Wait for signal or context cancellation
	select {
	case sig := <-quit:
		slog.Info("Received signal, initiating graceful shutdown",
			"signal", sig,
			"worker_type", wh.config.WorkerType)
		close(interrupt)
	case <-ctx.Done():
		slog.Info("Context cancelled, initiating graceful shutdown",
			"worker_type", wh.config.WorkerType)
		close(interrupt)
	}

	// Wait for workers to complete with timeout
	shutdownTimeout := wh.config.ShutdownTimeout
	timeoutCtx, timeoutCancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer timeoutCancel()

	workersDone := make(chan error, 1)
	go func() {
		workersDone <- g.Wait()
	}()

	select {
	case <-timeoutCtx.Done():
		slog.Error("Shutdown timeout reached, forcing exit",
			"timeout", shutdownTimeout,
			"worker_type", wh.config.WorkerType)
		// Close all clients even on timeout
		closeClients(clients)
		return fmt.Errorf("shutdown timeout for %s workers", wh.config.WorkerType)
	case shutdownErr := <-workersDone:
		// Close all clients after workers shutdown
		closeClients(clients)
		if shutdownErr != nil {
			slog.Error("Some workers failed during shutdown",
				"error", shutdownErr,
				"worker_type", wh.config.WorkerType)
			return shutdownErr
		}
		slog.Info("All workers shut down gracefully",
			"worker_type", wh.config.WorkerType)
	}

	return nil
}
