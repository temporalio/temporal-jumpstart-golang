package onboardings

import (
	"connectrpc.com/connect"
	"context"
	commandsv1 "github.com/temporalio/temporal-jumpstart-golang/onboardings/generated/onboardings/domain/commands/v1"
	snailforcev1 "github.com/temporalio/temporal-jumpstart-golang/onboardings/generated/snailforce/v1"
	"github.com/temporalio/temporal-jumpstart-golang/onboardings/generated/snailforce/v1/snailforcev1connect"
	"log"
)

var TypeOnboardActivities *Activities

type Activities struct {
	snailforce snailforcev1connect.SnailforceServiceClient
}

func (a *Activities) RegisterCrmEntity(ctx context.Context, q *commandsv1.RegisterCrmEntityRequest) error {
	if _, err := a.snailforce.Register(ctx, connect.NewRequest(&snailforcev1.RegisterRequest{
		Id:    q.GetId(),
		Value: q.GetValue(),
	})); err != nil {
		log.Println("ERROR", err)
	}
	return nil
}

func (a *Activities) SendEmail(ctx context.Context, cmd *commandsv1.RequestDeputyOwnerRequest) error {
	//TODO implement me
	panic("implement me")
}
