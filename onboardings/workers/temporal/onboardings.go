package temporal

import (
	"context"
	"github.com/temporalio/temporal-jumpstart-golang/onboardings/clients"
	"github.com/temporalio/temporal-jumpstart-golang/onboardings/config"
	"github.com/temporalio/temporal-jumpstart-golang/onboardings/domain/workflows"
	"github.com/temporalio/temporal-jumpstart-golang/onboardings/domain/workflows/onboardings"
	"github.com/temporalio/temporal-jumpstart-golang/onboardings/domain/workflows/onboardings/latest"
	"go.temporal.io/sdk/worker"
	"go.temporal.io/sdk/workflow"
)

type OnboardingsBuilder struct {
	clients *clients.Clients
}

func NewOnboardingsBuilder(clients *clients.Clients) *OnboardingsBuilder {
	return &OnboardingsBuilder{
		clients: clients,
	}
}
func (o *OnboardingsBuilder) Build(ctx context.Context, cfg *config.Config, target *WorkerClient, options worker.Options) (worker.Worker, error) {
	w := worker.New(target.TemporalClient(), target.TaskQueue, options)
	acts, err := onboardings.NewOnboardingsActivities(o.clients.Snailforce())
	if err != nil {
		return nil, err
	}
	w.RegisterWorkflowWithOptions(latest.OnboardEntity,
		workflow.RegisterOptions{Name: onboardings.TypeWorkflowOnboardEntity})
	w.RegisterWorkflow(workflows.Ping)
	w.RegisterActivity(acts)

	return w, nil
}
