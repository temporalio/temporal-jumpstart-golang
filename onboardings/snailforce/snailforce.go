package snailforce

import (
	"connectrpc.com/connect"
	"context"
	v1 "github.com/temporalio/temporal-jumpstart-golang/onboardings/generated/snailforce/v1"
	"github.com/temporalio/temporal-jumpstart-golang/onboardings/generated/snailforce/v1/snailforcev1connect"
	"log"
)

type SnailforceService struct {
	snailforcev1connect.UnimplementedSnailforceServiceHandler
}

func (s *SnailforceService) Register(ctx context.Context, c *connect.Request[v1.RegisterRequest]) (*connect.Response[v1.RegisterResponse], error) {
	log.Println("Registered", c.Msg.GetId(), "with value", c.Msg.GetValue())
	return connect.NewResponse(&v1.RegisterResponse{Id: c.Msg.GetId()}), nil
}
