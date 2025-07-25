package main

import (
	"context"
	"fmt"
	"github.com/temporalio/temporal-jumpstart-golang/app/api"
	appclients "github.com/temporalio/temporal-jumpstart-golang/app/clients"
	"github.com/temporalio/temporal-jumpstart-golang/app/config"
	"log"
	"net/http"
	"net/url"
	"os"
)

func main() {

	ctx := context.Background()
	cfg := config.MustNewConfig(os.Getenv("CONFIG_DIR"), os.Getenv("ENV"))
	clients, err := appclients.NewClients(ctx, cfg)
	if err != nil {
		log.Fatal(err)
	}
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
