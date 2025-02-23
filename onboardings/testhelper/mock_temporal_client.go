package testhelper

import (
	"context"
	"github.com/stretchr/testify/mock"
	"go.temporal.io/api/enums/v1"
	"go.temporal.io/api/operatorservice/v1"
	"go.temporal.io/api/workflowservice/v1"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/converter"
)

// MockTemporalClient captures calls for the `client.Client`. Implement these operations as needed.
type MockTemporalClient struct {
	mock.Mock
}

func (m *MockTemporalClient) ExecuteWorkflow(ctx context.Context, options client.StartWorkflowOptions, workflow interface{}, args ...interface{}) (client.WorkflowRun, error) {
	a := m.Called(ctx, options, workflow, args)
	if a.Get(0) == nil {
		return nil, a.Error(1)
	}
	return a.Get(0).(client.WorkflowRun), a.Error(1)
}

func (m *MockTemporalClient) GetWorkflow(ctx context.Context, workflowID string, runID string) client.WorkflowRun {
	//TODO implement me
	panic("implement me")
}

func (m *MockTemporalClient) SignalWorkflow(ctx context.Context, workflowID string, runID string, signalName string, arg interface{}) error {
	//TODO implement me
	panic("implement me")
}

func (m *MockTemporalClient) SignalWithStartWorkflow(ctx context.Context, workflowID string, signalName string, signalArg interface{}, options client.StartWorkflowOptions, workflow interface{}, workflowArgs ...interface{}) (client.WorkflowRun, error) {
	//TODO implement me
	panic("implement me")
}

func (m *MockTemporalClient) NewWithStartWorkflowOperation(options client.StartWorkflowOptions, workflow interface{}, args ...interface{}) client.WithStartWorkflowOperation {
	//TODO implement me
	panic("implement me")
}

func (m *MockTemporalClient) CancelWorkflow(ctx context.Context, workflowID string, runID string) error {
	//TODO implement me
	panic("implement me")
}

func (m *MockTemporalClient) TerminateWorkflow(ctx context.Context, workflowID string, runID string, reason string, details ...interface{}) error {
	//TODO implement me
	panic("implement me")
}

func (m *MockTemporalClient) GetWorkflowHistory(ctx context.Context, workflowID string, runID string, isLongPoll bool, filterType enums.HistoryEventFilterType) client.HistoryEventIterator {
	//TODO implement me
	panic("implement me")
}

func (m *MockTemporalClient) CompleteActivity(ctx context.Context, taskToken []byte, result interface{}, err error) error {
	//TODO implement me
	panic("implement me")
}

func (m *MockTemporalClient) CompleteActivityByID(ctx context.Context, namespace, workflowID, runID, activityID string, result interface{}, err error) error {
	//TODO implement me
	panic("implement me")
}

func (m *MockTemporalClient) RecordActivityHeartbeat(ctx context.Context, taskToken []byte, details ...interface{}) error {
	//TODO implement me
	panic("implement me")
}

func (m *MockTemporalClient) RecordActivityHeartbeatByID(ctx context.Context, namespace, workflowID, runID, activityID string, details ...interface{}) error {
	//TODO implement me
	panic("implement me")
}

func (m *MockTemporalClient) ListClosedWorkflow(ctx context.Context, request *workflowservice.ListClosedWorkflowExecutionsRequest) (*workflowservice.ListClosedWorkflowExecutionsResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (m *MockTemporalClient) ListOpenWorkflow(ctx context.Context, request *workflowservice.ListOpenWorkflowExecutionsRequest) (*workflowservice.ListOpenWorkflowExecutionsResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (m *MockTemporalClient) ListWorkflow(ctx context.Context, request *workflowservice.ListWorkflowExecutionsRequest) (*workflowservice.ListWorkflowExecutionsResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (m *MockTemporalClient) ListArchivedWorkflow(ctx context.Context, request *workflowservice.ListArchivedWorkflowExecutionsRequest) (*workflowservice.ListArchivedWorkflowExecutionsResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (m *MockTemporalClient) ScanWorkflow(ctx context.Context, request *workflowservice.ScanWorkflowExecutionsRequest) (*workflowservice.ScanWorkflowExecutionsResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (m *MockTemporalClient) CountWorkflow(ctx context.Context, request *workflowservice.CountWorkflowExecutionsRequest) (*workflowservice.CountWorkflowExecutionsResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (m *MockTemporalClient) GetSearchAttributes(ctx context.Context) (*workflowservice.GetSearchAttributesResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (m *MockTemporalClient) QueryWorkflow(ctx context.Context, workflowID string, runID string, queryType string, args ...interface{}) (converter.EncodedValue, error) {
	a := m.Called(ctx, workflowID, runID, queryType, args)
	if a.Get(0) == nil {
		return nil, a.Error(1)
	}
	return a.Get(0).(converter.EncodedValue), a.Error(1)
}

func (m *MockTemporalClient) QueryWorkflowWithOptions(ctx context.Context, request *client.QueryWorkflowWithOptionsRequest) (*client.QueryWorkflowWithOptionsResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (m *MockTemporalClient) DescribeWorkflowExecution(ctx context.Context, workflowID, runID string) (*workflowservice.DescribeWorkflowExecutionResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (m *MockTemporalClient) DescribeTaskQueue(ctx context.Context, taskqueue string, taskqueueType enums.TaskQueueType) (*workflowservice.DescribeTaskQueueResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (m *MockTemporalClient) DescribeTaskQueueEnhanced(ctx context.Context, options client.DescribeTaskQueueEnhancedOptions) (client.TaskQueueDescription, error) {
	//TODO implement me
	panic("implement me")
}

func (m *MockTemporalClient) ResetWorkflowExecution(ctx context.Context, request *workflowservice.ResetWorkflowExecutionRequest) (*workflowservice.ResetWorkflowExecutionResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (m *MockTemporalClient) UpdateWorkerBuildIdCompatibility(ctx context.Context, options *client.UpdateWorkerBuildIdCompatibilityOptions) error {
	//TODO implement me
	panic("implement me")
}

func (m *MockTemporalClient) GetWorkerBuildIdCompatibility(ctx context.Context, options *client.GetWorkerBuildIdCompatibilityOptions) (*client.WorkerBuildIDVersionSets, error) {
	//TODO implement me
	panic("implement me")
}

func (m *MockTemporalClient) GetWorkerTaskReachability(ctx context.Context, options *client.GetWorkerTaskReachabilityOptions) (*client.WorkerTaskReachability, error) {
	//TODO implement me
	panic("implement me")
}

func (m *MockTemporalClient) UpdateWorkerVersioningRules(ctx context.Context, options client.UpdateWorkerVersioningRulesOptions) (*client.WorkerVersioningRules, error) {
	//TODO implement me
	panic("implement me")
}

func (m *MockTemporalClient) GetWorkerVersioningRules(ctx context.Context, options client.GetWorkerVersioningOptions) (*client.WorkerVersioningRules, error) {
	//TODO implement me
	panic("implement me")
}

func (m *MockTemporalClient) CheckHealth(ctx context.Context, request *client.CheckHealthRequest) (*client.CheckHealthResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (m *MockTemporalClient) UpdateWorkflow(ctx context.Context, options client.UpdateWorkflowOptions) (client.WorkflowUpdateHandle, error) {
	a := m.Called(ctx, options)
	if a.Get(0) == nil {
		return nil, a.Error(1)
	}
	return a.Get(0).(client.WorkflowUpdateHandle), a.Error(1)
}

func (m *MockTemporalClient) UpdateWorkflowExecutionOptions(ctx context.Context, options client.UpdateWorkflowExecutionOptionsRequest) (client.WorkflowExecutionOptions, error) {
	//TODO implement me
	panic("implement me")
}

func (m *MockTemporalClient) UpdateWithStartWorkflow(ctx context.Context, options client.UpdateWithStartWorkflowOptions) (client.WorkflowUpdateHandle, error) {
	//TODO implement me
	panic("implement me")
}

func (m *MockTemporalClient) GetWorkflowUpdateHandle(ref client.GetWorkflowUpdateHandleOptions) client.WorkflowUpdateHandle {
	//TODO implement me
	panic("implement me")
}

func (m *MockTemporalClient) WorkflowService() workflowservice.WorkflowServiceClient {
	//TODO implement me
	panic("implement me")
}

func (m *MockTemporalClient) OperatorService() operatorservice.OperatorServiceClient {
	//TODO implement me
	panic("implement me")
}

func (m *MockTemporalClient) ScheduleClient() client.ScheduleClient {
	//TODO implement me
	panic("implement me")
}

func (m *MockTemporalClient) DeploymentClient() client.DeploymentClient {
	//TODO implement me
	panic("implement me")
}

func (m *MockTemporalClient) Close() {
	//TODO implement me
	panic("implement me")
}
