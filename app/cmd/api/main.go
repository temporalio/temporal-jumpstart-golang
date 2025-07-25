package main

import (
	"context"
	"fmt"
	"github.com/temporalio/temporal-jumpstart-golang/app/api"
	appclients "github.com/temporalio/temporal-jumpstart-golang/app/clients"
	"github.com/temporalio/temporal-jumpstart-golang/app/config"
	"log"
	"log/slog"
	"net/http"
	"net/url"
)

func main() {

	ctx := context.Background()
	// TODO pass in env params with defaults of `default`
	cfg := config.MustNewConfig("config", "default")
	clients, err := appclients.NewClients(ctx, cfg)
	if err != nil {
		log.Fatal(err)
	}
	slog.Info("starting api on", "url", cfg.API.URL)

	parsedUrl, err := url.Parse(cfg.API.URL)
	if err != nil {
		log.Fatal(err)
	}
	cfg.API.URL = parsedUrl.Host
	apiRouter, err := api.CreateAPIRouter(ctx, cfg, clients)
	if err != nil {
		log.Fatal(err)
	}
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", parsedUrl.Port()), apiRouter))
}
