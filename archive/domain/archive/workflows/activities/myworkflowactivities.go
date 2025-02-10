package activities

import (
	"context"
	"github.com/temporalio/temporal-jumpstart-golang/app/domain/scaffold/messages/commands"
	"github.com/temporalio/temporal-jumpstart-golang/app/domain/scaffold/messages/queries"
)

var TypeArchiveActivities *ArchiveActivities

type ArchiveActivities struct {
}

func (m *ArchiveActivities) QueryActivity(ctx context.Context, q *queries.QueryActivityRequest) (*queries.QueryActivityResponse, error) {
	return &queries.QueryActivityResponse{ID: q.ID}, nil
}

func (m *ArchiveActivities) MutateActivity(ctx context.Context, cmd *commands.MutateActivityRequest) error {
	//TODO implement me
	return nil
}
