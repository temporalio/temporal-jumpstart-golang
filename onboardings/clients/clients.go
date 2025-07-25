package clients

import (
	"context"
	"github.com/temporalio/temporal-jumpstart-golang/app/clients/temporal"
	"github.com/temporalio/temporal-jumpstart-golang/app/config"
	"go.temporal.io/sdk/client"
	"log"
	"sync"
)

// a way of singleton
var once sync.Once
var oneClients *Clients

type Clients struct {
	temporals           []client.Client
	temporalOptions     [][]temporal.Option
	temporalClientCount int
}

func (c *Clients) Temporals() []client.Client {
	return c.temporals
}

func (c *Clients) Close() {
	for _, t := range c.temporals {
		t.Close()
	}
}

func NewClients(ctx context.Context,
	cfg *config.Config, opts ...Option) (*Clients, error) {
	out := &Clients{}
	for _, opt := range opts {
		opt(out)
	}
	out.temporals = make([]client.Client, out.temporalClientCount)
	for i := 0; i < out.temporalClientCount; i++ {
		t, err := temporal.NewClient(ctx, cfg, i, out.temporalOptions[i]...)

		if err != nil {
			return nil, err
		}
		out.temporals[i] = t
	}

	return out, nil
}

func MustGetClients(
	ctx context.Context,
	cfg *config.Config,
	opts ...Option) *Clients {
	once.Do(func() {

		var err error
		if oneClients, err = NewClients(ctx, cfg, opts...); err != nil {
			log.Fatalf("Failed to create clients: %v", err)
		}

	})

	return oneClients

}
