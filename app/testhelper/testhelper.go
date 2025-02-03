package testhelper

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"go.temporal.io/api/enums/v1"
	"go.temporal.io/api/history/v1"
	"go.temporal.io/sdk/client"
	"reflect"
	"runtime"
	"strings"
)

func RandomString() string {
	return uuid.New().String()
}

// GetFunctionName shamelessly lifted from sdk-go
func GetFunctionName(i interface{}) (name string, isMethod bool) {
	if fullName, ok := i.(string); ok {
		return fullName, false
	}
	fullName := runtime.FuncForPC(reflect.ValueOf(i).Pointer()).Name()
	// Full function name that has a struct pointer receiver has the following format
	// <prefix>.(*<type>).<function>
	isMethod = strings.ContainsAny(fullName, "*")
	elements := strings.Split(fullName, ".")
	shortName := elements[len(elements)-1]
	// This allows to call activities by method pointer
	// Compiler adds -fm suffix to a function name which has a receiver
	// Note that this works even if struct pointer used to get the function is nil
	// It is possible because nil receivers are allowed.
	// For example:
	// var a *Activities
	// ExecuteActivity(ctx, a.Foo)
	// will call this function which is going to return "Foo"
	return strings.TrimSuffix(shortName, "-fm"), isMethod
}

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
