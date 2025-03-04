package testhelper

import (
	"context"
	"go.temporal.io/api/enums/v1"
	"go.temporal.io/api/history/v1"
	"go.temporal.io/sdk/client"
	"google.golang.org/protobuf/encoding/protojson"
	"os"
)

// GetWorkflowHistory is a utility for fetching all history events out of a workflow execution
func GetWorkflowHistory(ctx context.Context,
	client client.Client,
	workflowID string) (*history.History, error) {
	result := &history.History{
		Events: []*history.HistoryEvent{},
	}
	iter := client.GetWorkflowHistory(ctx, workflowID, "", true, enums.HISTORY_EVENT_FILTER_TYPE_ALL_EVENT)
	for iter.HasNext() {
		event, err := iter.Next()
		if err != nil {
			return nil, err
		}
		result.Events = append(result.Events, event)
	}
	return result, nil
}

// DumpWorkflowHistory is a utility for dumping all history events out of a workflow execution
// Note that destinationPath is relative to any test's directory that is being run.
func DumpWorkflowHistory(ctx context.Context,
	client client.Client,
	workflowID string,
	destinationPath string) (*history.History, error) {
	hist, err := GetWorkflowHistory(ctx, client, workflowID)
	if err != nil {
		return nil, err
	}
	data, err := protojson.Marshal(hist)
	if err != nil {
		return nil, err
	}
	err = os.WriteFile(destinationPath, data, os.ModePerm)
	if err != nil {
		return nil, err
	}
	return hist, nil
}
