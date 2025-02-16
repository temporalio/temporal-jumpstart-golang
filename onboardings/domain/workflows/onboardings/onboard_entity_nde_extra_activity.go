package onboardings

import (
	commandsv2 "github.com/temporalio/temporal-jumpstart-golang/onboardings/generated/onboardings/domain/commands/v2"
	workflowsv1 "github.com/temporalio/temporal-jumpstart-golang/onboardings/generated/onboardings/domain/workflows/v1"
	workflowsv2 "github.com/temporalio/temporal-jumpstart-golang/onboardings/generated/onboardings/domain/workflows/v2"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
	"time"
)

// OnboardEntityNDEExtraActivity is always the `latest` of this Application.
// In our scenario this is `V2` logic without a patch.
// Previous versions, even when using "Patched" strategy, should be copied over in the file system to support
// Replay testing. This allows no loading of history via JSON files and history can be piped in directly.
func OnboardEntityNDEExtraActivity(ctx workflow.Context, args *workflowsv2.OnboardEntityRequest) error {

	// first call the V1 as at first
	if err := OnboardEntityV1(ctx, &workflowsv1.OnboardEntityRequest{
		Timestamp:                args.Timestamp,
		Id:                       args.Id,
		Value:                    args.Value,
		CompletionTimeoutSeconds: args.CompletionTimeoutSeconds,
		DeputyOwnerEmail:         args.DeputyOwnerEmail,
		SkipApproval:             args.SkipApproval,
	}); err != nil {
		return err
	}
	notificationCtx := workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		StartToCloseTimeout: time.Second * 30,
		RetryPolicy:         &temporal.RetryPolicy{MaximumAttempts: 2}})

	// this should raise an NDE since we introduce this without a GetVersion call (patch)
	if err := workflow.ExecuteActivity(notificationCtx, "NotifyOnboardEntityCompleted", &commandsv2.NotifyOnboardEntityCompletedRequest{
		Id:       args.Id,
		Email:    args.Email,
		Value:    args.Value,
		Approval: nil, // we shouldnt get this far
	}).Get(notificationCtx, nil); err != nil {
		workflow.GetLogger(ctx).Error("NotifyOnboardEntity failed", "err", err)
		// notification failure will not fail the entire workflow
	}

	return nil
}
