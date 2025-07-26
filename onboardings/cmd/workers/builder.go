package main

import (
	"context"
	"github.com/temporalio/temporal-jumpstart-golang/onboardings/clients"
	"github.com/temporalio/temporal-jumpstart-golang/onboardings/config"
	worker "github.com/temporalio/temporal-jumpstart-golang/onboardings/workers/temporal"
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
	clients *clients.Clients
}

func (d *DefaultBuilder) Build(ctx context.Context,
	cfg *config.Config,
	target *worker.WorkerClient,
	options sdkworker.Options) (sdkworker.Worker, error) {

	var onboardings = worker.NewOnboardingsBuilder(d.clients)
	return onboardings.Build(ctx, cfg, target, options)
}
