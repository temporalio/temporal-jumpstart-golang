package onboardings

import (
	workflows2 "github.com/temporalio/temporal-jumpstart-golang/onboardings/domain/workflows"
	commandsv1 "github.com/temporalio/temporal-jumpstart-golang/onboardings/generated/onboardings/domain/commands/v1"
	queriesv1 "github.com/temporalio/temporal-jumpstart-golang/onboardings/generated/onboardings/domain/queries/v1"
	valuesv1 "github.com/temporalio/temporal-jumpstart-golang/onboardings/generated/onboardings/domain/values/v1"
	workflowsv1 "github.com/temporalio/temporal-jumpstart-golang/onboardings/generated/onboardings/domain/workflows/v1"
	"go.temporal.io/sdk/log"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
	"strings"
	"time"
)

// OnboardEntityV1 is the initial launch (v1) of this workflow
func OnboardEntityV1(ctx workflow.Context, args *workflowsv1.OnboardEntityRequest) error {

	// 1. initialize state ASAP
	state := &queriesv1.EntityOnboardingStateResponse{
		Id:          args.Id,
		SentRequest: args,
		Approval: &valuesv1.Approval{
			Status:  valuesv1.ApprovalStatus_APPROVAL_STATUS_PENDING,
			Comment: "",
		},
		ApprovalTimeRemainingSeconds: args.GetCompletionTimeoutSeconds(),
	}
	calculator := onboardEntityDurationCalculator{
		completionTimeoutSeconds: args.CompletionTimeoutSeconds,
		skipApproval:             args.SkipApproval,
		timestamp:                args.Timestamp.AsTime(),
		hasDeputyOwner:           len(strings.TrimSpace(args.DeputyOwnerEmail)) > 0,
	}

	waitSeconds := calculator.calculateWaitSeconds(workflow.Now(ctx))
	notifyDeputy := !args.SkipApproval && len(strings.TrimSpace(args.DeputyOwnerEmail)) > 0
	logger := log.With(workflow.GetLogger(ctx), "waitSeconds", waitSeconds, "notifyDeputy", notifyDeputy)
	logger.Info("Starting OnboardEntity")

	// 2. configure reads
	if err := workflow.SetQueryHandlerWithOptions(ctx, QueryEntityOnboardingState, func() (*queriesv1.EntityOnboardingStateResponse, error) {
		now := workflow.Now(ctx)
		logger.Debug("Onboarding State", "now", now.String())
		threshold := calculator.calculateCompletionThreshold(workflow.Now(ctx))

		timeRemaining := threshold.Sub(workflow.Now(ctx))
		return &queriesv1.EntityOnboardingStateResponse{
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

	// 3. validate
	if err := assertValidArgsv1(args); err != nil {
		return err
	}
	if calculator.calculateCompletionThreshold(workflow.Now(ctx)).Sub(workflow.Now(ctx)).Seconds() <= 0 {
		return temporal.NewApplicationError(workflowsv1.Errors_ERRORS_ONBOARD_ENTITY_TIMED_OUT.String(), workflowsv1.Errors_ERRORS_ONBOARD_ENTITY_TIMED_OUT.String())
	}

	// 4. configure write handlers
	approvalCtx, cancelApprovalWindow := workflow.WithCancel(ctx)
	approvalChan := workflow.GetSignalChannel(approvalCtx, workflows2.SignalName(&commandsv1.ApproveEntityRequest{}))
	rejectChan := workflow.GetSignalChannel(approvalCtx, workflows2.SignalName(&commandsv1.RejectEntityRequest{}))

	approvalsSelector := workflow.NewNamedSelector(approvalCtx, "approvals")
	workflow.GoNamed(ctx, "approvals", func(ctx workflow.Context) {
		approvalsSelector.AddReceive(approvalChan, func(c workflow.ReceiveChannel, more bool) {
			var approval *commandsv1.ApproveEntityRequest
			approvalChan.Receive(ctx, &approval)
			state.Approval = &valuesv1.Approval{
				Status:  valuesv1.ApprovalStatus_APPROVAL_STATUS_APPROVED,
				Comment: approval.GetComment(),
			}
		})
		approvalsSelector.AddReceive(rejectChan, func(c workflow.ReceiveChannel, more bool) {
			var rejection *commandsv1.RejectEntityRequest
			rejectChan.Receive(ctx, &rejection)
			state.Approval = &valuesv1.Approval{
				Status:  valuesv1.ApprovalStatus_APPROVAL_STATUS_REJECTED,
				Comment: rejection.GetComment(),
			}
		})
		approvalsSelector.Select(ctx)
	})

	// 5. perform workflow behavior

	conditionMet, err := workflow.AwaitWithTimeout(ctx, time.Duration(waitSeconds)*time.Second, func() bool {
		return state.Approval != nil && state.Approval.Status != valuesv1.ApprovalStatus_APPROVAL_STATUS_PENDING
	})
	// no longer eligible for approval while we figure out what to do next
	cancelApprovalWindow()
	if err = tryHandleCancellation(err); err != nil {
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

		if err := workflow.ExecuteActivity(notificationCtx, TypeOnboardingsActivities.SendDeputyOwnerApprovalRequest, &commandsv1.RequestDeputyOwnerRequest{
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
		if err = tryHandleCancellation(err); err != nil {
			return err
		}

		// 1. send email to the deputy owner to request approval.
		// 2. continue this workflow as new without the deputy owner email and reduce the amount of time we are willing to wait.
		return workflow.NewContinueAsNewError(ctx, workflow.GetInfo(ctx).WorkflowType.Name, &workflowsv1.OnboardEntityRequest{
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
	if err = workflow.ExecuteActivity(ao, OnboardEntityActivities.RegisterCrmEntity, &commandsv1.RegisterCrmEntityRequest{
		Id:    args.Id,
		Value: args.Value,
	}).Get(actCtx, nil); err != nil {
		logger.Error("RegisterCrmEntity failed", "err", err)
		return err
	}
	return nil
}
