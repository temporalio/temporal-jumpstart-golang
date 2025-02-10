package main

import (
	"context"
	appclients "github.com/temporalio/temporal-jumpstart-golang/app/clients"
	"github.com/temporalio/temporal-jumpstart-golang/app/config"
	temporalworker "github.com/temporalio/temporal-jumpstart-golang/app/workers/temporal"
	"go.temporal.io/sdk/worker"
	"log"
)

func main() {
	ctx := context.Background()
	cfg := config.MustNewConfig()
	clients, err := appclients.NewClients(ctx, cfg)
	if err != nil {
		log.Fatalln(err)
	}
	w, err := temporalworker.NewAppsWorker(ctx, cfg, clients)
	if err != nil {
		log.Fatalln(err)
	}

	if err := w.Run(worker.InterruptCh()); err != nil {
		log.Fatal(err)
	}
}
