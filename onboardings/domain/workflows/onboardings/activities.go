package onboardings

import (
	"context"
	v1 "github.com/temporalio/temporal-jumpstart-golang/onboardings/generated/domain/v1"
)

var TypeOnboardActivities *Activities

type Activities struct {
}

func (a *Activities) RegisterCrmEntity(ctx context.Context, q *v1.RegisterCrmEntityRequest) error {
	//TODO implement me
	panic("implement me")
}

func (a *Activities) SendEmail(ctx context.Context, cmd *v1.RequestDeputyOwnerRequest) error {
	//TODO implement me
	panic("implement me")
}
