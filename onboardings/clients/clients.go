package clients

import (
	"context"
	"github.com/temporalio/temporal-jumpstart-golang/onboardings/clients/snailforce"
	"github.com/temporalio/temporal-jumpstart-golang/onboardings/clients/temporal"
	"github.com/temporalio/temporal-jumpstart-golang/onboardings/config"
	"github.com/temporalio/temporal-jumpstart-golang/onboardings/generated/snailforce/v1/snailforcev1connect"
	"log"
	"log/slog"
	"sync"
)

// a way of singleton
var once sync.Once
var oneClients *Clients

type Clients struct {
	temporals       map[string][]*temporal.Client
	temporalOptions map[string][]temporal.Option
	snailforce      snailforcev1connect.SnailforceServiceClient
}

func (c *Clients) Temporals() map[string][]*temporal.Client {
	return c.temporals
}
func (c *Clients) Snailforce() snailforcev1connect.SnailforceServiceClient {
	return c.snailforce
}
func (c *Clients) Close() {
	for _, cl := range c.temporals {
		for _, t := range cl {
			t.Close()
		}
	}
}

func NewClients(ctx context.Context,
	cfg *config.Config, opts ...Option) (*Clients, error) {
	out := &Clients{
		temporalOptions: make(map[string][]temporal.Option),
		temporals:       make(map[string][]*temporal.Client),
	}
	for _, opt := range opts {
		opt(out)
	}
	out.temporals = make(map[string][]*temporal.Client)
	for ns, nscfg := range cfg.Temporal.Namespaces {
		// create N Workflow Clients
		if nscfg.Workflows != nil {
			for i := 0; i < nscfg.Workflows.ClientCount; i++ {
				clientOpts := out.temporalOptions[ns]
				c, err := temporal.NewClient(ctx, ns, cfg.Temporal, i, clientOpts...)
				if err != nil {
					return nil, err
				}
				out.temporals[ns] = append(out.temporals[ns], c)
				slog.Info("connected workflow client", "namespace", ns, "client", i)
			}
		}
	}
	var err error
	out.snailforce, err = snailforce.NewClient(ctx, cfg, nil)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func MustNewClients(
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
