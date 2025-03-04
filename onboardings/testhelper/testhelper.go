package testhelper

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/temporalio/temporal-jumpstart-golang/onboardings/domain/workflows"
	"go.temporal.io/sdk/client"
	"reflect"
)

// NonDeterminismError https://github.com/temporalio/rules/blob/299fd78a45a5b342d9460cc2155ff4e50d1b2e96/rules/TMPRL1100.md
const NonDeterminismError = "TMPRL1100"

func RandomString() string {
	return uuid.New().String()
}

// GetFunctionName shamelessly lifted from sdk-go
var GetFunctionName = workflows.GetFunctionName

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

// TestWorkflowRun implements WorkflowRun for testing
type TestWorkflowRun struct {
	WorkflowID string
	RunID      string
}

func (t *TestWorkflowRun) GetID() string {
	//TODO implement me
	return t.WorkflowID
}

func (t *TestWorkflowRun) GetRunID() string {
	//TODO implement me
	return t.RunID
}

func (t *TestWorkflowRun) Get(ctx context.Context, valuePtr interface{}) error {
	//TODO implement me
	panic("implement me")
}

func (t *TestWorkflowRun) GetWithOptions(ctx context.Context, valuePtr interface{}, options client.WorkflowRunGetOptions) error {
	//TODO implement me
	panic("implement me")
}

// TestWorkflowUpdateHandler implements WorkflowUpdateHandle for testing
type TestWorkflowUpdateHandle struct {
	WorkflowIDToUse string
}

func (t *TestWorkflowUpdateHandle) WorkflowID() string {
	return t.WorkflowIDToUse
}

func (t *TestWorkflowUpdateHandle) RunID() string {
	//TODO implement me
	panic("implement me")
}

func (t *TestWorkflowUpdateHandle) UpdateID() string {
	//TODO implement me
	panic("implement me")
}

func (t *TestWorkflowUpdateHandle) Get(ctx context.Context, valuePtr interface{}) error {
	//TODO implement me
	panic("implement me")
}
