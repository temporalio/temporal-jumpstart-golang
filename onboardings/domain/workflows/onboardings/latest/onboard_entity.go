package latest

import (
	"fmt"
	workflows2 "github.com/temporalio/temporal-jumpstart-golang/onboardings/domain/workflows"
	"github.com/temporalio/temporal-jumpstart-golang/onboardings/domain/workflows/onboardings"
	commandsv1 "github.com/temporalio/temporal-jumpstart-golang/onboardings/generated/onboardings/domain/commands/v1"
	commandsv2 "github.com/temporalio/temporal-jumpstart-golang/onboardings/generated/onboardings/domain/commands/v2"
	queriesv2 "github.com/temporalio/temporal-jumpstart-golang/onboardings/generated/onboardings/domain/queries/v2"
	valuesv1 "github.com/temporalio/temporal-jumpstart-golang/onboardings/generated/onboardings/domain/values/v1"
	workflowsv1 "github.com/temporalio/temporal-jumpstart-golang/onboardings/generated/onboardings/domain/workflows/v1"
	workflowsv2 "github.com/temporalio/temporal-jumpstart-golang/onboardings/generated/onboardings/domain/workflows/v2"
	"go.temporal.io/sdk/log"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
	"strings"
	"time"
)

// assertValidArgsv2 is a poor-man's implementation not providing very specific details on which args are "invalid"
// but instead just failing hard in the event of missing or bad args
func assertValidArgsv2(args *workflowsv2.OnboardEntityRequest) error {
	if strings.TrimSpace(args.Id) != "" &&
		strings.TrimSpace(args.Value) != "" &&
		!args.Timestamp.AsTime().IsZero() {
		return nil
	}
	return temporal.NewApplicationError(workflowsv1.Errors_ERRORS_ONBOARD_ENTITY_INVALID_ARGS.String(),
		workflowsv1.Errors_ERRORS_ONBOARD_ENTITY_INVALID_ARGS.String())
}

// OnboardEntity is always the `latest` of this Application.
// In our scenario this is `V2`.
// Previous versions, even when using "Patched" strategy, should be copied over in the file system to support
// Replay testing. This allows no loading of history via JSON files and history can be piped in directly.
func OnboardEntity(ctx workflow.Context, args *workflowsv2.OnboardEntityRequest) error {

	// 1. initialize state ASAP
	state := &queriesv2.EntityOnboardingStateResponse{
		Id:          args.Id,
		SentRequest: args,
		Approval: &valuesv1.Approval{
			Status:  valuesv1.ApprovalStatus_APPROVAL_STATUS_PENDING,
			Comment: "",
		},
		ApprovalTimeRemainingSeconds: args.GetCompletionTimeoutSeconds(),
	}
	calculator := onboardings.OnboardEntityDurationCalculator{
		CompletionTimeoutSeconds: args.CompletionTimeoutSeconds,
		SkipApproval:             args.SkipApproval,
		Timestamp:                args.Timestamp.AsTime(),
		HasDeputyOwner:           len(strings.TrimSpace(args.DeputyOwnerEmail)) > 0,
	}

	waitSeconds := calculator.CalculateWaitSeconds(workflow.Now(ctx))
	notifyDeputy := !args.SkipApproval && len(strings.TrimSpace(args.DeputyOwnerEmail)) > 0
	logger := log.With(workflow.GetLogger(ctx), "waitSeconds", waitSeconds, "notifyDeputy", notifyDeputy)
	logger.Info("Starting OnboardEntity")

	// 2. configure reads
	err3 := setQueryHandlers(ctx, args, state, calculator)
	if err3 != nil {
		return err3
	}

	// 3. validate
	if err := assertValidArgsv2(args); err != nil {
		return err
	}
	if calculator.CalculateCompletionThreshold(workflow.Now(ctx)).Sub(workflow.Now(ctx)).Seconds() <= 0 {
		return temporal.NewApplicationError(workflowsv1.Errors_ERRORS_ONBOARD_ENTITY_TIMED_OUT.String(),
			workflowsv1.Errors_ERRORS_ONBOARD_ENTITY_TIMED_OUT.String())
	}

	// create Context the main WF can interact with
	approvalCtx, cancelApprovalWindow := workflow.WithCancel(ctx)

	// 4 - Updates: configure write handlers
	err2 := setUpdateHandlers(approvalCtx, state)
	if err2 != nil {
		return err2
	}

	// 5. perform workflow behavior

	conditionMet, err := workflow.AwaitWithTimeout(ctx, time.Duration(waitSeconds)*time.Second, func() bool {
		return state.Approval != nil && state.Approval.Status != valuesv1.ApprovalStatus_APPROVAL_STATUS_PENDING
	})
	// no longer eligible for approval while we figure out what to do next
	cancelApprovalWindow()
	if err = onboardings.TryHandleCancellation(err); err != nil {
		return err
	}

	if !conditionMet {
		if !notifyDeputy {
			return temporal.NewApplicationError("Approval was not met in time", workflowsv1.Errors_ERRORS_ONBOARD_ENTITY_TIMED_OUT.String())
		}
		notificationCtx := workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
			StartToCloseTimeout: time.Second * 30,
			// prevent spamming the deputy owner with our request
			RetryPolicy: &temporal.RetryPolicy{
				MaximumAttempts: 2,
			},
		})

		if err := workflow.ExecuteActivity(notificationCtx, onboardings.TypeOnboardingsActivities.SendDeputyOwnerApprovalRequest, &commandsv1.RequestDeputyOwnerRequest{
			Id:               args.Id,
			DeputyOwnerEmail: args.DeputyOwnerEmail,
		}).Get(notificationCtx, nil); err != nil {
			logger.Error("failed to notify deputy owner", "err", err)
			return err
		}

		err = workflow.Await(ctx, func() bool {
			// block until any buffered (late) messages might have come in
			// But note this is somewhat redundant since we explicitly `cancelApprovalWindow` when
			// we have not made progress in time.
			return workflow.AllHandlersFinished(ctx)
		})
		if err = onboardings.TryHandleCancellation(err); err != nil {
			return err
		}

		// 1. send email to the deputy owner to request approval.
		// 2. continue this workflow as new without the deputy owner email and reduce the amount of time we are willing to wait.
		return workflow.NewContinueAsNewError(ctx, OnboardEntity, &workflowsv1.OnboardEntityRequest{
			Id:                       args.Id,
			Value:                    args.Value,
			CompletionTimeoutSeconds: args.CompletionTimeoutSeconds - waitSeconds, // offset how long we will wait
			DeputyOwnerEmail:         "",                                          // blank out deputy approval for next run
			SkipApproval:             args.SkipApproval,
		})
	}
	if state.Approval.Status == valuesv1.ApprovalStatus_APPROVAL_STATUS_REJECTED {
		logger.Info("Entity was rejected.")
		return temporal.NewApplicationError(workflowsv1.Errors_ERRORS_ONBOARD_ENTITY_REJECTED.String(), workflowsv1.Errors_ERRORS_ONBOARD_ENTITY_REJECTED.String())
	}
	if state.Approval.Status != valuesv1.ApprovalStatus_APPROVAL_STATUS_APPROVED {
		logger.Info("Entity was either rejected or was not approved in time.")
		return temporal.NewApplicationError(workflowsv1.Errors_ERRORS_ONBOARD_ENTITY_TIMED_OUT.String(), workflowsv1.Errors_ERRORS_ONBOARD_ENTITY_TIMED_OUT.String())
	}
	// the rest of this is for approved onboardings
	actCtx, _ := workflow.NewDisconnectedContext(ctx)
	ao := workflow.WithActivityOptions(actCtx, workflow.ActivityOptions{StartToCloseTimeout: time.Second * 30})
	if err = workflow.ExecuteActivity(ao, onboardings.OnboardEntityActivities.RegisterCrmEntity, &commandsv1.RegisterCrmEntityRequest{
		Id:    args.Id,
		Value: args.Value,
	}).Get(actCtx, nil); err != nil {
		logger.Error("RegisterCrmEntity failed", "err", err)
		return err
	}
	notificationCtx := workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		StartToCloseTimeout: time.Second * 30,
		RetryPolicy:         &temporal.RetryPolicy{MaximumAttempts: 2}})

	notifyVersion := workflow.GetVersion(ctx, "notifyOnboardEntityCompleted", workflow.DefaultVersion, 0)
	if notifyVersion == workflow.DefaultVersion {
		return nil
	}
	if err = workflow.ExecuteActivity(notificationCtx, "NotifyOnboardEntityCompleted", &commandsv2.NotifyOnboardEntityCompletedRequest{
		Id:       args.Id,
		Email:    args.Email,
		Value:    args.Value,
		Approval: state.Approval,
	}).Get(notificationCtx, nil); err != nil {
		logger.Error("NotifyOnboardEntity failed", "err", err)
		// notification failure will not fail the entire workflow
	}

	return nil
}

func setQueryHandlers(ctx workflow.Context, args *workflowsv2.OnboardEntityRequest, state *queriesv2.EntityOnboardingStateResponse, calculator onboardings.OnboardEntityDurationCalculator) error {
	if err := workflow.SetQueryHandlerWithOptions(ctx, onboardings.QueryEntityOnboardingState, func() (*queriesv2.EntityOnboardingStateResponse, error) {
		threshold := calculator.CalculateCompletionThreshold(workflow.Now(ctx))

		timeRemaining := threshold.Sub(workflow.Now(ctx))
		return &queriesv2.EntityOnboardingStateResponse{
			Id:                           state.Id,
			SentRequest:                  args,
			Approval:                     state.Approval,
			ApprovalTimeRemainingSeconds: uint64(timeRemaining.Seconds()),
		}, nil
	}, workflow.QueryHandlerOptions{
		Description: "Get Entity Onboarding State",
	}); err != nil {
		return err
	}
	return nil
}

func setUpdateHandlers(approvalCtx workflow.Context, state *queriesv2.EntityOnboardingStateResponse) error {
	if err := workflow.SetUpdateHandlerWithOptions(
		approvalCtx,
		workflows2.UpdateName(&commandsv1.ApproveEntityRequest{}),
		func(ctx workflow.Context, cmd *commandsv1.ApproveEntityRequest) error {
			state.Approval = &valuesv1.Approval{
				Status:  valuesv1.ApprovalStatus_APPROVAL_STATUS_APPROVED,
				Comment: cmd.GetComment(),
			}
			return nil
		},
		workflow.UpdateHandlerOptions{
			Description: "Approve Entity Onboarding",
			Validator: func(ctx workflow.Context, cmd *commandsv1.ApproveEntityRequest) error {
				if state.Approval == nil || (state.Approval.Status == valuesv1.ApprovalStatus_APPROVAL_STATUS_PENDING ||
					state.Approval.Status == valuesv1.ApprovalStatus_APPROVAL_STATUS_UNSPECIFIED) {
					return nil
				}
				return fmt.Errorf("Onboarding Approval or Rejection has already been completed: %v", state.Approval.Status)
			},
		}); err != nil {
		return err
	}
	if err := workflow.SetUpdateHandlerWithOptions(
		approvalCtx,
		workflows2.UpdateName(&commandsv1.RejectEntityRequest{}),
		func(ctx workflow.Context, cmd *commandsv1.RejectEntityRequest) error {
			state.Approval = &valuesv1.Approval{
				Status:  valuesv1.ApprovalStatus_APPROVAL_STATUS_REJECTED,
				Comment: cmd.GetComment(),
			}
			return nil
		},
		workflow.UpdateHandlerOptions{
			Description: "Reject Entity Onboarding",
			Validator: func(ctx workflow.Context, cmd *commandsv1.RejectEntityRequest) error {
				if state.Approval == nil || (state.Approval.Status == valuesv1.ApprovalStatus_APPROVAL_STATUS_PENDING ||
					state.Approval.Status == valuesv1.ApprovalStatus_APPROVAL_STATUS_UNSPECIFIED) {
					return nil
				}
				return fmt.Errorf("Onboarding Approval or Rejection has already been completed: %v", state.Approval.Status)
			},
		}); err != nil {
		return err
	}
	return nil
}
