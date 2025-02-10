package onboardings

import (
	"context"
	workflows2 "github.com/temporalio/temporal-jumpstart-golang/onboardings/domain/workflows"
	v1 "github.com/temporalio/temporal-jumpstart-golang/onboardings/generated/domain/v1"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
	"math"
	"strings"
	"time"
)

const QueryEntityOnboardingState = "entityOnboardingState"

type OnboardEntityActivities interface {
	RegisterCrmEntity(ctx context.Context, q *v1.RegisterCrmEntityRequest) error
	SendEmail(ctx context.Context, cmd *v1.RequestDeputyOwnerRequest) error
}

func (workflows *Workflows) OnboardEntity(ctx workflow.Context, args *v1.OnboardEntityRequest) error {

	state := &v1.EntityOnboardingStateResponse{
		Id:          args.Id,
		Status:      "",
		SentRequest: args,
		Approval: &v1.Approval{
			Status:  v1.ApprovalStatus_PENDING,
			Comment: "",
		},
	}
	logger := workflow.GetLogger(ctx)
	waitSeconds := int64(math.Abs(math.Round(float64(args.CompletionTimeoutSeconds))))
	notifyDeputy := !args.SkipApproval && len(strings.Trim(args.DeputyOwnerEmail, " ")) > 0

	approvalCtx, cancelApprovalWindow := workflow.WithCancel(ctx)
	approvalChan := workflow.GetSignalChannel(approvalCtx, workflows2.SignalName(&v1.ApproveEntityRequest{}))
	rejectChan := workflow.GetSignalChannel(approvalCtx, workflows2.SignalName(&v1.RejectEntityRequest{}))

	approvalsSelector := workflow.NewNamedSelector(approvalCtx, "approvals")
	workflow.GoNamed(ctx, "approvals", func(ctx workflow.Context) {
		approvalsSelector.AddReceive(approvalChan, func(c workflow.ReceiveChannel, more bool) {
			var approval *v1.ApproveEntityRequest
			approvalChan.Receive(ctx, &approval)
			state.Approval = &v1.Approval{
				Status:  v1.ApprovalStatus_APPROVED,
				Comment: approval.GetComment(),
			}
		})
		approvalsSelector.AddReceive(approvalChan, func(c workflow.ReceiveChannel, more bool) {
			var rejection *v1.RejectEntityRequest
			rejectChan.Receive(ctx, &rejection)
			state.Approval = &v1.Approval{
				Status:  v1.ApprovalStatus_REJECTED,
				Comment: rejection.GetComment(),
			}
		})
		approvalsSelector.Select(ctx)
	})
	conditionMet, _ := workflow.AwaitWithTimeout(ctx, time.Duration(waitSeconds)*time.Second, func() bool {
		return state.Approval != nil && state.Approval.Status != v1.ApprovalStatus_PENDING
	})
	// no longer eligible for approval while we figure out what to do next
	cancelApprovalWindow()

	if !conditionMet {
		if !notifyDeputy {
			return temporal.NewApplicationError("Approval was not met in time", v1.Errors_ERR_ONBOARD_ENTITY_TIMED_OUT.String())
		}
		return temporal.NewApplicationError("not implemented", "not implemented")
	}
	if state.Approval.Status != v1.ApprovalStatus_APPROVED {
		logger.Info("Entity was rejected")
		return nil
	}
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
