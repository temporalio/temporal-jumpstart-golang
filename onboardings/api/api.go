package api

import (
	"context"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/temporalio/temporal-jumpstart-golang/onboardings/clients"
	"github.com/temporalio/temporal-jumpstart-golang/onboardings/config"
	"net/http"
)

func CreateAPIRouter(ctx context.Context, cfg *config.Config, clients *clients.Clients) (http.Handler, error) {
	router := mux.NewRouter()
	creds := handlers.AllowCredentials()
	headers := handlers.AllowedHeaders([]string{"X-Requested-With", "Content-Type", "Authorization", "Content-Type"})
	methods := handlers.AllowedMethods([]string{"GET", "POST", "PUT", "PATCH", "HEAD", "OPTIONS"})
	origins := handlers.AllowedOrigins([]string{"*"})
	ttl := handlers.MaxAge(3600)

	v1Router := router.PathPrefix("/api/v1").Subrouter()
	v1Router = createV1Router(ctx, &V1Dependencies{
		Clients: clients,
		Config:  cfg,
	}, v1Router)
	return handlers.CORS(creds, headers, methods, origins, ttl)(router), nil
}
