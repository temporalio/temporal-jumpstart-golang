package onboardings

import (
	workflows2 "github.com/temporalio/temporal-jumpstart-golang/onboardings/domain/workflows"
	v1 "github.com/temporalio/temporal-jumpstart-golang/onboardings/generated/domain/v1"
	"go.temporal.io/sdk/log"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
	"strings"
	"time"
)

// OnboardEntityV1 is the initial launch (v1) of this workflow
func (workflows *Workflows) OnboardEntityV1(ctx workflow.Context, args *v1.OnboardEntityRequest) error {

	// 1. initialize state ASAP
	state := &v1.EntityOnboardingStateResponse{
		Id:          args.Id,
		SentRequest: args,
		Approval: &v1.Approval{
			Status:  v1.ApprovalStatus_APPROVAL_STATUS_PENDING,
			Comment: "",
		},
		ApprovalTimeRemainingSeconds: args.GetCompletionTimeoutSeconds(),
	}

	waitSeconds := calculateWaitSeconds(workflow.Now(ctx), args)
	notifyDeputy := !args.SkipApproval && len(strings.TrimSpace(args.DeputyOwnerEmail)) > 0
	logger := log.With(workflow.GetLogger(ctx), "waitSeconds", waitSeconds, "notifyDeputy", notifyDeputy)
	logger.Info("Starting OnboardEntity")

	// 2. configure reads
	if err := workflow.SetQueryHandlerWithOptions(ctx, QueryEntityOnboardingState, func() (*v1.EntityOnboardingStateResponse, error) {
		now := workflow.Now(ctx)
		logger.Debug("Onboarding State", "now", now.String())
		threshold := calculateCompletionThreshold(args)

		timeRemaining := threshold.Sub(workflow.Now(ctx))
		return &v1.EntityOnboardingStateResponse{
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
	if err := assertValidArgs(args); err != nil {
		return err
	}
	if calculateCompletionThreshold(args).Sub(workflow.Now(ctx)).Seconds() <= 0 {
		return temporal.NewApplicationError(v1.Errors_ERRORS_ONBOARD_ENTITY_TIMED_OUT.String(), v1.Errors_ERRORS_ONBOARD_ENTITY_TIMED_OUT.String())
	}

	// 4. configure write handlers
	approvalCtx, cancelApprovalWindow := workflow.WithCancel(ctx)
	approvalChan := workflow.GetSignalChannel(approvalCtx, workflows2.SignalName(&v1.ApproveEntityRequest{}))
	rejectChan := workflow.GetSignalChannel(approvalCtx, workflows2.SignalName(&v1.RejectEntityRequest{}))

	approvalsSelector := workflow.NewNamedSelector(approvalCtx, "approvals")
	workflow.GoNamed(ctx, "approvals", func(ctx workflow.Context) {
		approvalsSelector.AddReceive(approvalChan, func(c workflow.ReceiveChannel, more bool) {
			var approval *v1.ApproveEntityRequest
			approvalChan.Receive(ctx, &approval)
			state.Approval = &v1.Approval{
				Status:  v1.ApprovalStatus_APPROVAL_STATUS_APPROVED,
				Comment: approval.GetComment(),
			}
		})
		approvalsSelector.AddReceive(rejectChan, func(c workflow.ReceiveChannel, more bool) {
			var rejection *v1.RejectEntityRequest
			rejectChan.Receive(ctx, &rejection)
			state.Approval = &v1.Approval{
				Status:  v1.ApprovalStatus_APPROVAL_STATUS_REJECTED,
				Comment: rejection.GetComment(),
			}
		})
		approvalsSelector.Select(ctx)
	})

	// 5. perform workflow behavior

	conditionMet, _ := workflow.AwaitWithTimeout(ctx, time.Duration(waitSeconds)*time.Second, func() bool {
		return state.Approval != nil && state.Approval.Status != v1.ApprovalStatus_APPROVAL_STATUS_PENDING
	})
	// no longer eligible for approval while we figure out what to do next
	cancelApprovalWindow()

	if !conditionMet {
		if !notifyDeputy {
			return temporal.NewApplicationError("Approval was not met in time", v1.Errors_ERRORS_ONBOARD_ENTITY_TIMED_OUT.String())
		}
		notificationCtx := workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
			StartToCloseTimeout: time.Second * 30,
			// prevent spamming the deputy owner with our request
			RetryPolicy: &temporal.RetryPolicy{
				MaximumAttempts: 2,
			},
		})
		if err := workflow.ExecuteActivity(notificationCtx, TypeOnboardActivities.SendEmail, &v1.RequestDeputyOwnerRequest{
			Id:               args.Id,
			DeputyOwnerEmail: args.DeputyOwnerEmail,
		}).Get(notificationCtx, nil); err != nil {
			logger.Error("failed to notify deputy owner", "err", err)
			return err
		}
		// 1. send email to the deputy owner to request approval.
		// 2. continue this workflow as new without the deputy owner email and reduce the amount of time we are willing to wait.
		return workflow.NewContinueAsNewError(ctx, TypeWorkflows.OnboardEntity, &v1.OnboardEntityRequest{
			Id:                       args.Id,
			Value:                    args.Value,
			CompletionTimeoutSeconds: args.CompletionTimeoutSeconds - waitSeconds, // offset how long we will wait
			DeputyOwnerEmail:         "",                                          // blank out deputy approval for next run
			SkipApproval:             args.SkipApproval,
		})
	}
	if state.Approval.Status == v1.ApprovalStatus_APPROVAL_STATUS_REJECTED {
		logger.Info("Entity was rejected.")
		return temporal.NewApplicationError(v1.Errors_ERRORS_ONBOARD_ENTITY_REJECTED.String(), v1.Errors_ERRORS_ONBOARD_ENTITY_REJECTED.String())
	}
	if state.Approval.Status != v1.ApprovalStatus_APPROVAL_STATUS_APPROVED {
		logger.Info("Entity was either rejected or was not approved in time.")
		return temporal.NewApplicationError(v1.Errors_ERRORS_ONBOARD_ENTITY_TIMED_OUT.String(), v1.Errors_ERRORS_ONBOARD_ENTITY_TIMED_OUT.String())
	}
	// the rest of this is only despite approval status
	ao := workflow.WithActivityOptions(ctx, workflow.ActivityOptions{StartToCloseTimeout: time.Second * 30})
	if err := workflow.ExecuteActivity(ao, OnboardEntityActivities.RegisterCrmEntity, &v1.RegisterCrmEntityRequest{
		Id:    args.Id,
		Value: args.Value,
	}).Get(ctx, nil); err != nil {
		logger.Error("RegisterCrmEntity failed", "err", err)
		return err
	}
	return nil
}
