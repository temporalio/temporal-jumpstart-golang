package activities

import (
	"context"
	"github.com/temporalio/temporal-jumpstart-golang/app/domain/scaffold/messages/commands"
	"github.com/temporalio/temporal-jumpstart-golang/app/domain/scaffold/messages/queries"
)

var TypeMyWorkflowActivities *MyWorkflowActivities

type MyWorkflowActivities struct {
}

func (m *MyWorkflowActivities) QueryActivity(ctx context.Context, q *queries.QueryActivityRequest) (*queries.QueryActivityResponse, error) {
	return &queries.QueryActivityResponse{ID: q.ID}, nil
}

func (m *MyWorkflowActivities) MutateActivity(ctx context.Context, cmd *commands.MutateActivityRequest) error {
	//TODO implement me
	return nil
}
