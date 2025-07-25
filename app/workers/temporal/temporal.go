package temporal

import (
	"context"
	"fmt"
	"github.com/temporalio/temporal-jumpstart-golang/app/clients/temporal"
	"github.com/temporalio/temporal-jumpstart-golang/app/config"
	"github.com/temporalio/temporal-jumpstart-golang/app/instrumentation/prometheus"
	"github.com/uber-go/tally/v4"
	sdkclient "go.temporal.io/sdk/client"
)

type WorkerTarget struct {
	Namespace string
	TaskQueue string
}
type WorkerClient struct {
	WorkerTarget
	Client *temporal.Client
}

func (w *WorkerClient) TemporalClient() sdkclient.Client {
	return w.Client.Client
}

func NewClients(ctx context.Context,
	cfg *config.Config,
	rootScope tally.Scope,
	targets []WorkerTarget) ([]*WorkerClient, error) {

	result := make([]*WorkerClient, 0)
	for _, ntq := range targets {
		tcfg, exists := cfg.Temporal.Namespaces[ntq.Namespace]
		if !exists {
			return nil, fmt.Errorf("namespace '%s' not found in config", ntq.Namespace)
		}
		tqcfg, exists := tcfg.TaskQueues[ntq.TaskQueue]
		if !exists {
			return nil, fmt.Errorf("task queue '%s' not found in config", ntq.TaskQueue)
		}
		for i := 0; i < tqcfg.WorkerCount; i++ {
			scope, serr := prometheus.NewScope(ctx, rootScope, prometheus.WithTags(
				map[string]string{
					// used for disambiguating metrics
					"worker_client_id": fmt.Sprintf("client-%d", i),
				}))
			if serr != nil {
				return nil, fmt.Errorf("failed to create prometheus scope: %w", serr)
			}
			opts := []temporal.Option{
				temporal.WithMetricsScope(scope),
			}
			c, cerr := temporal.NewClient(ctx, ntq.Namespace, cfg.Temporal, i, opts...)
			if cerr != nil {
				return nil, fmt.Errorf("failed to create client: %w", cerr)
			}
			result = append(result, &WorkerClient{WorkerTarget: ntq, Client: c})
		}
	}
	return result, nil
}
