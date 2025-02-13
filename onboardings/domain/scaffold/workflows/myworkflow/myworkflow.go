package myworkflow

import (
	"context"
	"github.com/temporalio/temporal-jumpstart-golang/onboardings/domain/scaffold/messages/commands"
	"github.com/temporalio/temporal-jumpstart-golang/onboardings/domain/scaffold/messages/queries"
	"github.com/temporalio/temporal-jumpstart-golang/onboardings/domain/scaffold/messages/workflows"
	"go.temporal.io/sdk/workflow"
	"time"
)

type MyWorkflowActivities interface {
	QueryActivity(ctx context.Context, q *queries.QueryActivityRequest) (*queries.QueryActivityResponse, error)
	MutateActivity(ctx context.Context, cmd *commands.MutateActivityRequest) error
}

func (workflows *Workflows) MyWorkflow(ctx workflow.Context, args *workflows.StartMyWorkflowRequest) error {
	workflow.Sleep(ctx, time.Second*1)
	return nil
}
