package onboardings

import (
	"context"
	commandsv1 "github.com/temporalio/temporal-jumpstart-golang/onboardings/generated/onboardings/domain/commands/v1"
	"time"
)

var TypeWorkflows *Workflows

const QueryEntityOnboardingState = "entityOnboardingState"

// DefaultCompletionTimeoutSeconds specs say to wait 7 days before giving up
var DefaultCompletionTimeoutSeconds = uint64((7 * 24 * time.Hour).Seconds())

type OnboardEntityActivities interface {
	RegisterCrmEntity(ctx context.Context, q *commandsv1.RegisterCrmEntityRequest) error
	SendEmail(ctx context.Context, cmd *commandsv1.RequestDeputyOwnerRequest) error
}

type Workflows struct{}
