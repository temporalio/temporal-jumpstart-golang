package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/temporalio/temporal-jumpstart-golang/onboardings/api/encoding"
	"github.com/temporalio/temporal-jumpstart-golang/onboardings/clients"
	"github.com/temporalio/temporal-jumpstart-golang/onboardings/config"
	"github.com/temporalio/temporal-jumpstart-golang/onboardings/domain/workflows/onboardings"
	apiv1 "github.com/temporalio/temporal-jumpstart-golang/onboardings/generated/onboardings/api/v1"
	queriesv1 "github.com/temporalio/temporal-jumpstart-golang/onboardings/generated/onboardings/domain/queries/v1"
	valuesv1 "github.com/temporalio/temporal-jumpstart-golang/onboardings/generated/onboardings/domain/values/v1"
	workflowsv1 "github.com/temporalio/temporal-jumpstart-golang/onboardings/generated/onboardings/domain/workflows/v1"
	"github.com/temporalio/temporal-jumpstart-golang/onboardings/testhelper"
	"go.temporal.io/api/enums/v1"
	"go.temporal.io/api/serviceerror"
	"go.temporal.io/sdk/client"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

type TestWorkflowRun struct {
	id    string
	runId string
}

func (t *TestWorkflowRun) GetID() string {
	//TODO implement me
	return t.id
}

func (t *TestWorkflowRun) GetRunID() string {
	//TODO implement me
	return t.runId
}

func (t *TestWorkflowRun) Get(ctx context.Context, valuePtr interface{}) error {
	//TODO implement me
	panic("implement me")
}

func (t *TestWorkflowRun) GetWithOptions(ctx context.Context, valuePtr interface{}, options client.WorkflowRunGetOptions) error {
	//TODO implement me
	panic("implement me")
}

// resource state test
func TestV1StartOnboardingAcceptsTheRequestAndReturnsOnboardingLocation(t *testing.T) {
	A := assert.New(t)
	ctx := context.Background()
	workflowId := testhelper.RandomString()
	body := &apiv1.OnboardingsPut{Value: testhelper.RandomString()}
	cfg := &config.Config{
		Temporal: &config.TemporalConfig{
			Worker: &config.TemporalWorker{TaskQueue: testhelper.RandomString()},
		},
	}
	temporalClient := &testhelper.MockTemporalClient{}
	temporalClient.On("ExecuteWorkflow", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(&TestWorkflowRun{id: workflowId}, nil)
	c := &clients.Clients{Temporal: temporalClient}
	sut := createV1Router(ctx, &V1Dependencies{
		Clients: c,
		Config:  cfg,
	}, mux.NewRouter())
	testserver := httptest.NewServer(sut)
	defer testserver.Close()
	jsonBody, err := json.Marshal(&body)
	A.NoError(err)
	parsedUrl, err := url.Parse(testserver.URL + "/onboardings/" + workflowId)
	A.NoError(err)
	req, err := http.NewRequest("PUT", parsedUrl.String(), bytes.NewReader(jsonBody))
	A.NoError(err)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Host", testserver.URL)

	resp, err := testserver.Client().Do(req)
	A.NoError(err)

	A.Equal(202, resp.StatusCode)
	A.Equal(fmt.Sprintf("%s/onboardings/%s", strings.Replace(testserver.URL, "http", "https", -1), workflowId), resp.Header.Get("Location"))
}

// resource behavior test
func testV1PutApprovalParamsSignalsRelatedOnboarding(t *testing.T) {
	A := assert.New(t)
	ctx := context.Background()
	workflowId := testhelper.RandomString()
	body := &apiv1.ApprovalsPut{
		Id: workflowId,
		Approval: &valuesv1.Approval{
			Status:  valuesv1.ApprovalStatus_APPROVAL_STATUS_APPROVED,
			Comment: testhelper.RandomString(),
		},
	}
	cfg := &config.Config{
		Temporal: &config.TemporalConfig{
			Worker: &config.TemporalWorker{TaskQueue: testhelper.RandomString()},
		},
	}
	temporalClient := &testhelper.MockTemporalClient{}
	temporalClient.On("ExecuteWorkflow", mock.Anything,
		mock.MatchedBy(func(opts client.StartWorkflowOptions) bool {
			// check the workflow options we are configuring
			return opts.ID == workflowId &&
				opts.TaskQueue == cfg.Temporal.Worker.TaskQueue &&
				opts.WorkflowIDReusePolicy == enums.WORKFLOW_ID_REUSE_POLICY_ALLOW_DUPLICATE_FAILED_ONLY &&
				opts.WorkflowIDConflictPolicy == enums.WORKFLOW_ID_CONFLICT_POLICY_FAIL &&
				opts.WorkflowExecutionErrorWhenAlreadyStarted == true
		}), mock.MatchedBy(func(fn interface{}) bool {
			funcName, _ := testhelper.GetFunctionName(fn)
			return funcName == "OnboardEntityV1"
		}), mock.MatchedBy(func(params []interface{}) bool {
			panic("not implemented")

			//// check input argument to our Workflow
			//if len(params) != 1 {
			//	return false
			//}
			//arg, ok := params[0].(*workflowsv1.OnboardEntityRequest)
			//if !ok {
			//	return false
			//}
			//panic("not implemented")
			//return arg.Value == body.GetValue() && arg.Id == workflowId
		})).Once().Return(&TestWorkflowRun{id: workflowId}, nil)
	c := &clients.Clients{Temporal: temporalClient}
	sut := createV1Router(ctx, &V1Dependencies{
		Clients: c,
		Config:  cfg,
	}, mux.NewRouter())
	testserver := httptest.NewServer(sut)
	defer testserver.Close()
	jsonBody, err := json.Marshal(&body)
	A.NoError(err)
	parsedUrl, err := url.Parse(testserver.URL + "/onboardings/" + workflowId)
	A.NoError(err)
	req, err := http.NewRequest("PUT", parsedUrl.String(), bytes.NewReader(jsonBody))
	A.NoError(err)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Host", testserver.URL)
	_, err = testserver.Client().Do(req)
	A.NoError(err)

	temporalClient.AssertExpectations(t)
}

// resource behavior test
func TestV1PutOnboardingStartsOnboardEntityWithCorrectParams(t *testing.T) {
	A := assert.New(t)
	ctx := context.Background()
	workflowId := testhelper.RandomString()
	body := &apiv1.OnboardingsPut{Value: testhelper.RandomString()}
	cfg := &config.Config{
		Temporal: &config.TemporalConfig{
			Worker: &config.TemporalWorker{TaskQueue: testhelper.RandomString()},
		},
	}
	temporalClient := &testhelper.MockTemporalClient{}
	temporalClient.On("ExecuteWorkflow", mock.Anything,
		mock.MatchedBy(func(opts client.StartWorkflowOptions) bool {
			// check the workflow options we are configuring
			return opts.ID == workflowId &&
				opts.TaskQueue == cfg.Temporal.Worker.TaskQueue &&
				opts.WorkflowIDReusePolicy == enums.WORKFLOW_ID_REUSE_POLICY_ALLOW_DUPLICATE_FAILED_ONLY &&
				opts.WorkflowIDConflictPolicy == enums.WORKFLOW_ID_CONFLICT_POLICY_FAIL &&
				opts.WorkflowExecutionErrorWhenAlreadyStarted == true
		}), mock.MatchedBy(func(fn interface{}) bool {
			funcName, _ := testhelper.GetFunctionName(fn)
			return funcName == "OnboardEntity"
		}), mock.MatchedBy(func(params []interface{}) bool {
			// check input argument to our Workflow
			if len(params) != 1 {
				return false
			}
			arg, ok := params[0].(*workflowsv1.OnboardEntityRequest)
			if !ok {
				return false
			}
			return arg.Value == body.GetValue() && arg.Id == workflowId
		})).Once().Return(&TestWorkflowRun{id: workflowId}, nil)
	c := &clients.Clients{Temporal: temporalClient}
	sut := createV1Router(ctx, &V1Dependencies{
		Clients: c,
		Config:  cfg,
	}, mux.NewRouter())
	testserver := httptest.NewServer(sut)
	defer testserver.Close()
	jsonBody, err := json.Marshal(&body)
	A.NoError(err)
	parsedUrl, err := url.Parse(testserver.URL + "/onboardings/" + workflowId)
	A.NoError(err)
	req, err := http.NewRequest("PUT", parsedUrl.String(), bytes.NewReader(jsonBody))
	A.NoError(err)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Host", testserver.URL)
	_, err = testserver.Client().Do(req)
	A.NoError(err)

	temporalClient.AssertExpectations(t)
}

// state test
func TestV1GetOnboardingState(t *testing.T) {
	A := assert.New(t)
	ctx := context.Background()
	workflowId := testhelper.RandomString()
	cfg := &config.Config{
		Temporal: &config.TemporalConfig{
			Worker: &config.TemporalWorker{TaskQueue: testhelper.RandomString()},
		},
	}
	state := &queriesv1.EntityOnboardingStateResponse{
		Id: workflowId,
		SentRequest: &workflowsv1.OnboardEntityRequest{
			Id:                       workflowId,
			Value:                    testhelper.RandomString(),
			CompletionTimeoutSeconds: 0,
			DeputyOwnerEmail:         "",
			SkipApproval:             false,
		},
		Approval: &valuesv1.Approval{
			Status:  valuesv1.ApprovalStatus_APPROVAL_STATUS_PENDING,
			Comment: "",
		},
	}

	queryResult := &testhelper.TestEncodedValue{Value: &state}

	temporalClient := &testhelper.MockTemporalClient{}
	temporalClient.On("QueryWorkflow",
		mock.Anything,
		mock.Anything,
		mock.Anything,
		onboardings.QueryEntityOnboardingState,
		mock.Anything).Return(queryResult, nil)
	c := &clients.Clients{Temporal: temporalClient}
	sut := createV1Router(ctx, &V1Dependencies{
		Clients: c,
		Config:  cfg,
	}, mux.NewRouter())
	testserver := httptest.NewServer(sut)
	defer testserver.Close()
	parsedUrl, err := url.Parse(testserver.URL + "/onboardings/" + workflowId)
	A.NoError(err)
	req, err := http.NewRequest("GET", parsedUrl.String(), nil)
	A.NoError(err)

	resp, err := testserver.Client().Do(req)
	A.NoError(err)

	A.Equal(200, resp.StatusCode)
	actual := &apiv1.OnboardingsGet{}
	A.NoError(encoding.DecodeJSONResponse(resp, &actual))
	A.NoError(err)
	var expect *apiv1.OnboardingsGet = &apiv1.OnboardingsGet{
		Id:           workflowId,
		CurrentValue: state.SentRequest.GetValue(),
		Approval:     state.Approval,
	}
	A.Equal(expect.CurrentValue, actual.CurrentValue)
	A.Equal(expect.Approval.Status, actual.Approval.Status)
	A.Equal(expect.Approval.Comment, actual.Approval.Comment)
}

// behavior test
func TestV1GetGivenExistingOnboardingFetchesCurrentState(t *testing.T) {
	A := assert.New(t)
	ctx := context.Background()
	workflowId := testhelper.RandomString()
	cfg := &config.Config{
		Temporal: &config.TemporalConfig{
			Worker: &config.TemporalWorker{TaskQueue: testhelper.RandomString()},
		},
	}
	state := &queriesv1.EntityOnboardingStateResponse{
		Id: workflowId,
		SentRequest: &workflowsv1.OnboardEntityRequest{
			Id:                       workflowId,
			Value:                    testhelper.RandomString(),
			CompletionTimeoutSeconds: 0,
			DeputyOwnerEmail:         "",
			SkipApproval:             false,
		},
		Approval: &valuesv1.Approval{
			Status:  valuesv1.ApprovalStatus_APPROVAL_STATUS_PENDING,
			Comment: "",
		},
	}

	queryResult := &testhelper.TestEncodedValue{Value: &state}

	temporalClient := &testhelper.MockTemporalClient{}
	temporalClient.On("QueryWorkflow",
		mock.Anything,
		mock.Anything,
		mock.Anything,
		onboardings.QueryEntityOnboardingState,
		mock.Anything).Return(queryResult, nil)

	c := &clients.Clients{Temporal: temporalClient}
	sut := createV1Router(ctx, &V1Dependencies{
		Clients: c,
		Config:  cfg,
	}, mux.NewRouter())
	testserver := httptest.NewServer(sut)
	defer testserver.Close()
	parsedUrl, err := url.Parse(testserver.URL + "/onboardings/" + workflowId)
	A.NoError(err)
	req, err := http.NewRequest("GET", parsedUrl.String(), nil)
	A.NoError(err)

	_, err = testserver.Client().Do(req)
	A.NoError(err)

	temporalClient.AssertExpectations(t)
}

// state
func TestV1GetGivenNonExistingOnboardingReturns404(t *testing.T) {
	A := assert.New(t)
	ctx := context.Background()
	workflowId := testhelper.RandomString()
	cfg := &config.Config{
		Temporal: &config.TemporalConfig{
			Worker: &config.TemporalWorker{TaskQueue: testhelper.RandomString()},
		},
	}

	temporalClient := &testhelper.MockTemporalClient{}
	temporalClient.On("QueryWorkflow",
		mock.Anything,
		workflowId,
		mock.Anything,
		onboardings.QueryEntityOnboardingState,
		mock.Anything).Once().Return(nil, serviceerror.NewNotFound("workflow not found"))
	c := &clients.Clients{Temporal: temporalClient}
	sut := createV1Router(ctx, &V1Dependencies{
		Clients: c,
		Config:  cfg,
	}, mux.NewRouter())
	testserver := httptest.NewServer(sut)
	defer testserver.Close()
	parsedUrl, err := url.Parse(testserver.URL + "/onboardings/" + workflowId)
	A.NoError(err)
	req, err := http.NewRequest("GET", parsedUrl.String(), nil)
	A.NoError(err)

	resp, err := testserver.Client().Do(req)
	A.NoError(err)

	A.Equal(http.StatusNotFound, resp.StatusCode)
}
