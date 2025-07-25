package main

import (
	"context"
	"fmt"
	"github.com/temporalio/temporal-jumpstart-golang/app/clients"
	"github.com/temporalio/temporal-jumpstart-golang/app/clients/temporal"
	"github.com/temporalio/temporal-jumpstart-golang/app/config"
	"github.com/temporalio/temporal-jumpstart-golang/app/instrumentation/prometheus"
	worker "github.com/temporalio/temporal-jumpstart-golang/app/workers/temporal"
	"github.com/temporalio/temporal-jumpstart-golang/app/workers/temporal/apps"
	"log"
	"log/slog"
	"os"
	"time"
)

func main() {
	// Basic setup (same for all workers)
	setupLogging()
	configDir := os.Getenv("CONFIG_DIR")
	if configDir == "" {
		panic("CONFIG_DIR is not set")
	}

	env := os.Getenv("ENV")
	if env == "" {
		panic("ENV is not set")
	}
	cfg := config.MustNewConfig(configDir, env)
	ctx := context.Background()

	// Setup instrumentation and clients - different metrics endpoint
	clients, err := setupClientsWithMetrics(ctx, cfg)
	if err != nil {
		log.Fatalf("Failed to setup clients: %v", err)
	}
	defer clients.Close()

	builder, err := apps.NewBuilder(ctx, clients)
	if err != nil {
		log.Fatalf("Failed to create builder: %v", err)
	}
	// Create and run worker host - different factory and config
	host := worker.NewWorkerHost(
		ctx,
		&worker.HostConfig{
			WorkerCount:     len(clients.Temporals()), // Different config path
			ShutdownTimeout: 15 * time.Second,         // Maybe longer timeout for file workers
			WorkerType:      "file_monitoring",
		},
		builder, // Different factory function
	)

	if err := host.Run(ctx, cfg, clients); err != nil {
		slog.Error("Builder host failed", "error", err)
		os.Exit(1)
	}

	slog.Info("Goodbye")
}

// Same helper functions but parameterized
func setupLogging() {
	log.SetOutput(os.Stdout)
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	handler := slog.NewJSONHandler(os.Stdout, nil)
	slog.SetDefault(slog.New(handler))
}

func setupClientsWithMetrics(ctx context.Context, cfg *config.Config) (*clients.Clients, error) {
	// Same logic but with parameterized metrics address

	reporter, err := prometheus.NewReporter(ctx, cfg,
		prometheus.WithListenAddress(cfg.Temporal.Prometheus.Address))
	if err != nil {
		return nil, fmt.Errorf("failed to create prometheus reporter: %w", err)
	}

	root, err := prometheus.NewRootScope(ctx, cfg, prometheus.WithReporter(reporter))
	if err != nil {
		return nil, fmt.Errorf("failed to create prometheus root scope: %w", err)
	}

	// Use file monitoring worker count instead
	workerCount := cfg.Temporal.Workers.Apps.WorkerCount
	topts := make([][]temporal.Option, workerCount)
	for i := 0; i < workerCount; i++ {
		scope, serr := prometheus.NewScope(ctx, cfg, root, prometheus.WithTags(
			map[string]string{"client_id": fmt.Sprintf("client-%d", i)}))
		if serr != nil {
			return nil, fmt.Errorf("failed to create prometheus scope: %w", serr)
		}
		topts[i] = []temporal.Option{
			temporal.WithMetricsScope(scope),
		}
	}

	return clients2.NewClients(ctx, cfg,
		clients2.WithTemporalClientCount(workerCount),
		clients2.WithTemporalOptions(topts...))
}
