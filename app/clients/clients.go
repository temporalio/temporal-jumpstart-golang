package clients

import (
	"context"
	"github.com/temporalio/temporal-jumpstart-golang/app/clients/temporal"
	"github.com/temporalio/temporal-jumpstart-golang/app/config"
	"log"
	"sync"
)

// a way of singleton
var once sync.Once
var oneClients *Clients

type Clients struct {
	temporals           map[string]*temporal.Client
	temporalOptions     map[string][]temporal.Option
	temporalClientCount int
}

func (c *Clients) Temporals() map[string]*temporal.Client {
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
	out.temporals = make(map[string]*temporal.Client)
	for ns, nscfg := range cfg.Temporal.Namespaces {
		// create N Workflow Clients
		if nscfg.Workflows != nil {
			for i := 0; i < nscfg.Workflows.ClientCount; i++ {
				c, err := temporal.NewClient(ctx, ns, cfg.Temporal, i)
				if err != nil {
				}
				out.temporals[ns] = c
			}
		}
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
