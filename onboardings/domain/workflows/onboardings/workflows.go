package onboardings

import (
	"context"
	commandsv1 "github.com/temporalio/temporal-jumpstart-golang/onboardings/generated/onboardings/domain/commands/v1"
	commandsv2 "github.com/temporalio/temporal-jumpstart-golang/onboardings/generated/onboardings/domain/commands/v2"
	"time"
)

var TypeWorkflows *Workflows

const QueryEntityOnboardingState = "entityOnboardingState"

// DefaultCompletionTimeoutSeconds specs say to wait 7 days before giving up
var DefaultCompletionTimeoutSeconds = uint64((7 * 24 * time.Hour).Seconds())

type OnboardEntityActivities interface {
	RegisterCrmEntity(ctx context.Context, q *commandsv1.RegisterCrmEntityRequest) error
	SendDeputyOwnerApprovalRequest(ctx context.Context, cmd *commandsv1.RequestDeputyOwnerRequest) error
}
type OnboardEntityNotificationActivities interface {
	NotifyOnboardEntityCompleted(ctx context.Context, cmd *commandsv2.NotifyOnboardEntityCompletedRequest) error
}

type Workflows struct{}
