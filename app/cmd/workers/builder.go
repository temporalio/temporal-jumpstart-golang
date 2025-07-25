package main

import (
	"context"
	"github.com/temporalio/temporal-jumpstart-golang/app/clients"
	"github.com/temporalio/temporal-jumpstart-golang/app/config"
	worker "github.com/temporalio/temporal-jumpstart-golang/app/workers/temporal"
	"github.com/temporalio/temporal-jumpstart-golang/onboardings/domain/workflows"
	sdkworker "go.temporal.io/sdk/worker"
)

type NullBuilder struct {
	// demonstrates injection of dependency for activities, etc
	Clients *clients.Clients
}

func (n *NullBuilder) Build(ctx context.Context, cfg *config.Config, target *worker.WorkerClient, options sdkworker.Options) (sdkworker.Worker, error) {
	return sdkworker.New(target.TemporalClient(), target.TaskQueue, options), nil
}

type DefaultBuilder struct {
}

func (d *DefaultBuilder) Build(ctx context.Context, cfg *config.Config, target *worker.WorkerClient, options sdkworker.Options) (sdkworker.Worker, error) {
	w := sdkworker.New(target.TemporalClient(), target.TaskQueue, options)
	w.RegisterWorkflow(workflows.Ping)
	return w, nil
}
