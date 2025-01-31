package main

import (
	"context"
	"fmt"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
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
	router := mux.NewRouter()
	creds := handlers.AllowCredentials()
	headers := handlers.AllowedHeaders([]string{"X-Requested-With", "Content-Type", "Authorization", "Content-Type"})
	methods := handlers.AllowedMethods([]string{"GET", "POST", "PUT", "PATCH", "HEAD", "OPTIONS"})
	origins := handlers.AllowedOrigins([]string{"*"})
	ttl := handlers.MaxAge(3600)

	v1Router := router.PathPrefix("/api/v1").Subrouter()
	v1Router = api.CreateV1Router(ctx, &api.V1Dependencies{
		Clients: clients,
		Config:  cfg,
	}, v1Router)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", cfg.API.Port),
		handlers.CORS(creds, ttl, headers, methods, origins)(router)))
}
