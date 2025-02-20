package main

import (
	"context"
	"github.com/temporalio/temporal-jumpstart-golang/onboardings/api"
	appclients "github.com/temporalio/temporal-jumpstart-golang/onboardings/clients"
	"github.com/temporalio/temporal-jumpstart-golang/onboardings/config"
	"log"
	"net/http"
)

func main() {

	ctx := context.Background()
	cfg := config.MustNewConfig()
	clients, err := appclients.NewClients(ctx, cfg)
	if err != nil {
		log.Fatal(err)
	}

	apiRouter, err := api.CreateAPIRouter(ctx, cfg, clients)
	if err != nil {
		log.Fatal(err)
	}
	log.Fatal(http.ListenAndServe(cfg.API.URL.Host, apiRouter))
}
