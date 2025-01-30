package temporal

import (
	"github.com/temporalio/temporal-jumpstart-golang/app/clients"
	"github.com/temporalio/temporal-jumpstart-golang/app/config"
	"go.temporal.io/sdk/contrib/resourcetuner"
	"go.temporal.io/sdk/worker"
)

func NewWorker(cfg *config.Config, clients *clients.Clients) (worker.Worker, error) {
	if cfg.Temporal.Worker.Capacity.MaxCachedWorkflows > 0 {
		worker.SetStickyWorkflowCacheSize(cfg.Temporal.Worker.Capacity.MaxCachedWorkflows)
	}

	opts := worker.Options{}

	// capacity Tuner versus discrete Executor config is mutually exclusive
	if cfg.Temporal.Worker.Tuner != nil && (cfg.Temporal.Worker.Tuner.TargetCPU > 0 || cfg.Temporal.Worker.Tuner.TargetMem > 0) {
		// Using the ResourceBasedTuner in worker options
		tuner, err := resourcetuner.NewResourceBasedTuner(resourcetuner.ResourceBasedTunerOptions{
			TargetMem: cfg.Temporal.Worker.Tuner.TargetMem,
			TargetCpu: cfg.Temporal.Worker.Tuner.TargetCPU,
		})
		if err != nil {
			return nil, err
		}
		opts.Tuner = tuner
	} else {
		opts.MaxConcurrentActivityExecutionSize = cfg.Temporal.Worker.Capacity.MaxConcurrentActivityExecutors
		opts.MaxConcurrentWorkflowTaskExecutionSize = cfg.Temporal.Worker.Capacity.MaxConcurrentWorkflowTaskExecutions
		opts.MaxConcurrentLocalActivityExecutionSize = cfg.Temporal.Worker.Capacity.MaxConcurrentLocalActivityExecutors
		opts.MaxConcurrentActivityTaskPollers = cfg.Temporal.Worker.Capacity.MaxConcurrentActivityTaskPollers
		opts.MaxConcurrentWorkflowTaskPollers = cfg.Temporal.Worker.Capacity.MaxConcurrentWorkflowTaskPollers
	}
	if cfg.Temporal.Worker.RateLimits != nil {
		opts.WorkerActivitiesPerSecond = float64(cfg.Temporal.Worker.RateLimits.MaxWorkerActivitiesPerSecond)
		opts.TaskQueueActivitiesPerSecond = float64(cfg.Temporal.Worker.RateLimits.MaxTaskQueueActivitiesPerSecond)
	}

	w := worker.New(clients.Temporal, cfg.Temporal.Worker.TaskQueue, opts)

	return w, nil
}
