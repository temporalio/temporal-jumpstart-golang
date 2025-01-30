package clients

import (
	"context"
	"github.com/temporalio/temporal-jumpstart-golang/app/clients/temporal"
	"github.com/temporalio/temporal-jumpstart-golang/app/config"
	"go.temporal.io/sdk/client"
)

type Clients struct {
	Temporal client.Client
}

func NewClients(ctx context.Context, cfg *config.Config) (*Clients, error) {
	t, err := temporal.NewClient(ctx, cfg)
	if err != nil {
		return nil, err
	}
	return &Clients{
		Temporal: t,
	}, nil
}
