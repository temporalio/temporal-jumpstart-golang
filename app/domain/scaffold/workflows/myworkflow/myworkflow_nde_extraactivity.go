package myworkflow

import (
	"github.com/temporalio/temporal-jumpstart-golang/app/domain/scaffold/messages/commands"
	"github.com/temporalio/temporal-jumpstart-golang/app/domain/scaffold/messages/workflows"
	"go.temporal.io/sdk/workflow"
	"time"
)

func (workflows *Workflows) MyWorkflowV1(ctx workflow.Context, args *workflows.StartMyWorkflowRequest) error {
	// activity being added without GetVersion
	workflow.ExecuteActivity(ctx, MyWorkflowActivities.MutateActivity, &commands.MutateActivityRequest{})
	workflow.Sleep(ctx, time.Second)
	return nil
}
