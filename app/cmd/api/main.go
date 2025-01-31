package main

import (
	"context"
	"fmt"
	"github.com/temporalio/temporal-jumpstart-golang/app/api"
	"github.com/temporalio/temporal-jumpstart-golang/app/clients"
	"github.com/temporalio/temporal-jumpstart-golang/app/config"
	"log"
	"net/http"
)

func main() {

	ctx := context.Background()
	cfg := config.MustNewConfig()
	clients, err := clients.NewClients(ctx, cfg)
	if err != nil {
		log.Fatal(err)
	}

	apiRouter, err := api.CreateAPIRouter(ctx, cfg, clients)
	if err != nil {
		log.Fatal(err)
	}
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", cfg.API.Port), apiRouter))
}
