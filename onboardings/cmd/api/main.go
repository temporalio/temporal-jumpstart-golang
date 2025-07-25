package main

import (
	"context"
	"github.com/temporalio/temporal-jumpstart-golang/onboardings/api"
	appclients "github.com/temporalio/temporal-jumpstart-golang/onboardings/clients"
	"github.com/temporalio/temporal-jumpstart-golang/onboardings/config"
	"log"
	"net/http"
	"net/url"
)

func main() {

	ctx := context.Background()
	cfg := config.MustNewConfig("config", "default")
	clients, err := appclients.NewClients(ctx, cfg)
	if err != nil {
		log.Fatal(err)
	}
	parsedUrl, err := url.Parse(cfg.API.URL)
	if err != nil {
		log.Fatal(err)
	}

	apiRouter, err := api.CreateAPIRouter(ctx, cfg, clients)
	if err != nil {
		log.Fatal(err)
	}
	log.Fatal(http.ListenAndServe(parsedUrl.Host, apiRouter))
}
