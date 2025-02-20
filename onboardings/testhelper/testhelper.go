package testhelper

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/temporalio/temporal-jumpstart-golang/onboardings/domain/workflows"
	"go.temporal.io/api/enums/v1"
	"go.temporal.io/api/history/v1"
	"go.temporal.io/sdk/client"
	"google.golang.org/protobuf/encoding/protojson"
	"os"
	"reflect"
)

// NonDeterminismError https://github.com/temporalio/rules/blob/299fd78a45a5b342d9460cc2155ff4e50d1b2e96/rules/TMPRL1100.md
const NonDeterminismError = "TMPRL1100"

func RandomString() string {
	return uuid.New().String()
}

// GetFunctionName shamelessly lifted from sdk-go
var GetFunctionName = workflows.GetFunctionName

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

// TestEncodedValue simplifies testing with this result from the Temporal Client
type TestEncodedValue struct {
	Value interface{}
}

func (val *TestEncodedValue) HasValue() bool {
	return val.Value != nil
}

func (val *TestEncodedValue) Get(valuePtr interface{}) error {
	if !val.HasValue() {
		return fmt.Errorf("no value present")
	}
	if reflect.TypeOf(valuePtr) != reflect.TypeOf(val.Value) {
		return fmt.Errorf("wrong type of value. received %T but got %T", valuePtr, val.Value)
	}
	result := reflect.ValueOf(val.Value).Elem()
	reflect.ValueOf(valuePtr).Elem().Set(result)
	return nil
}
