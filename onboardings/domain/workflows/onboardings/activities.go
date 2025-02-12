package onboardings

import (
	"context"
	commandsv1 "github.com/temporalio/temporal-jumpstart-golang/onboardings/generated/onboardings/domain/commands/v1"
)

var TypeOnboardActivities *Activities

type Activities struct {
}

func (a *Activities) RegisterCrmEntity(ctx context.Context, q *commandscommandsv1.RegisterCrmEntityRequest) error {
	//TODO implement me
	panic("implement me")
}

func (a *Activities) SendEmail(ctx context.Context, cmd *commandscommandsv1.RequestDeputyOwnerRequest) error {
	//TODO implement me
	panic("implement me")
}
