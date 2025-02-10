package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/temporalio/temporal-jumpstart-golang/app/api/messages"
	"github.com/temporalio/temporal-jumpstart-golang/app/clients"
	"github.com/temporalio/temporal-jumpstart-golang/app/config"
	"github.com/temporalio/temporal-jumpstart-golang/app/domain/workflows"
	"github.com/temporalio/temporal-jumpstart-golang/app/testhelper"
	"go.temporal.io/api/enums/v1"
	"go.temporal.io/api/serviceerror"
	"go.temporal.io/sdk/client"
	"io"
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
func TestV1PutPingAcceptsTheRequestAndReturnsResourceLocation(t *testing.T) {
	A := assert.New(t)
	ctx := context.Background()
	workflowId := testhelper.RandomString()
	body := messages.PutPing{Ping: testhelper.RandomString()}
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
	parsedUrl, err := url.Parse(testserver.URL + "/pings/" + workflowId)
	A.NoError(err)
	req, err := http.NewRequest("PUT", parsedUrl.String(), bytes.NewReader(jsonBody))
	A.NoError(err)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Host", testserver.URL)

	resp, err := testserver.Client().Do(req)
	A.NoError(err)

	A.Equal(202, resp.StatusCode)
	A.Equal(fmt.Sprintf("%s/pings/%s", strings.Replace(testserver.URL, "http", "https", -1), workflowId), resp.Header.Get("Location"))
}

// resource behavior test
func TestV1PutPingStartsAPingWorkflowWithCorrectParams(t *testing.T) {
	A := assert.New(t)
	ctx := context.Background()
	workflowId := testhelper.RandomString()
	body := messages.PutPing{Ping: testhelper.RandomString()}
	cfg := &config.Config{
		Temporal: &config.TemporalConfig{
			Worker: &config.TemporalWorker{TaskQueue: testhelper.RandomString()},
		},
	}
	temporalClient := &testhelper.MockTemporalClient{}
	temporalClient.On("ExecuteWorkflow", mock.Anything,
		mock.MatchedBy(func(opts client.StartWorkflowOptions) bool {
			return opts.ID == workflowId &&
				opts.TaskQueue == cfg.Temporal.Worker.TaskQueue &&
				opts.WorkflowIDReusePolicy == enums.WORKFLOW_ID_REUSE_POLICY_ALLOW_DUPLICATE_FAILED_ONLY &&
				opts.WorkflowIDConflictPolicy == enums.WORKFLOW_ID_CONFLICT_POLICY_FAIL
		}), mock.MatchedBy(func(fn interface{}) bool {
			funcName, _ := testhelper.GetFunctionName(fn)
			return funcName == "Ping"
		}), []interface{}{body.Ping}).Once().Return(&TestWorkflowRun{id: workflowId}, nil)
	c := &clients.Clients{Temporal: temporalClient}
	sut := createV1Router(ctx, &V1Dependencies{
		Clients: c,
		Config:  cfg,
	}, mux.NewRouter())
	testserver := httptest.NewServer(sut)
	defer testserver.Close()
	jsonBody, err := json.Marshal(&body)
	A.NoError(err)
	parsedUrl, err := url.Parse(testserver.URL + "/pings/" + workflowId)
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
func TestV1GetExistingPingRepliesWithPong(t *testing.T) {
	A := assert.New(t)
	ctx := context.Background()
	workflowId := testhelper.RandomString()
	cfg := &config.Config{
		Temporal: &config.TemporalConfig{
			Worker: &config.TemporalWorker{TaskQueue: testhelper.RandomString()},
		},
	}
	var expect string = "pong: bar"
	queryResult := &testhelper.TestEncodedValue{Value: &expect}

	temporalClient := &testhelper.MockTemporalClient{}
	temporalClient.On("QueryWorkflow",
		mock.Anything,
		mock.Anything,
		mock.Anything,
		workflows.QueryPing,
		mock.Anything).Return(queryResult, nil)
	c := &clients.Clients{Temporal: temporalClient}
	sut := createV1Router(ctx, &V1Dependencies{
		Clients: c,
		Config:  cfg,
	}, mux.NewRouter())
	testserver := httptest.NewServer(sut)
	defer testserver.Close()
	parsedUrl, err := url.Parse(testserver.URL + "/pings/" + workflowId)
	A.NoError(err)
	req, err := http.NewRequest("GET", parsedUrl.String(), nil)
	A.NoError(err)

	resp, err := testserver.Client().Do(req)
	A.NoError(err)

	A.Equal(200, resp.StatusCode)
	actual, err := io.ReadAll(resp.Body)
	A.NoError(err)
	A.Equal(expect, string(actual))
}

// behavior test
func TestV1GetGivenExistingPingFetchesCurrentStateOfPing(t *testing.T) {
	A := assert.New(t)
	ctx := context.Background()
	workflowId := testhelper.RandomString()
	cfg := &config.Config{
		Temporal: &config.TemporalConfig{
			Worker: &config.TemporalWorker{TaskQueue: testhelper.RandomString()},
		},
	}
	var expect string = "pong: bar"
	queryResult := &testhelper.TestEncodedValue{Value: &expect}

	temporalClient := &testhelper.MockTemporalClient{}
	temporalClient.On("QueryWorkflow",
		mock.Anything,
		workflowId,
		mock.Anything,
		workflows.QueryPing,
		mock.Anything).Once().Return(queryResult, nil)
	c := &clients.Clients{Temporal: temporalClient}
	sut := createV1Router(ctx, &V1Dependencies{
		Clients: c,
		Config:  cfg,
	}, mux.NewRouter())
	testserver := httptest.NewServer(sut)
	defer testserver.Close()
	parsedUrl, err := url.Parse(testserver.URL + "/pings/" + workflowId)
	A.NoError(err)
	req, err := http.NewRequest("GET", parsedUrl.String(), nil)
	A.NoError(err)

	_, err = testserver.Client().Do(req)
	A.NoError(err)

	temporalClient.AssertExpectations(t)
}

// state
func TestV1GetGivenNonExistingPingReturns404(t *testing.T) {
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
		workflows.QueryPing,
		mock.Anything).Once().Return(nil, serviceerror.NewNotFound("workflow not found"))
	c := &clients.Clients{Temporal: temporalClient}
	sut := createV1Router(ctx, &V1Dependencies{
		Clients: c,
		Config:  cfg,
	}, mux.NewRouter())
	testserver := httptest.NewServer(sut)
	defer testserver.Close()
	parsedUrl, err := url.Parse(testserver.URL + "/pings/" + workflowId)
	A.NoError(err)
	req, err := http.NewRequest("GET", parsedUrl.String(), nil)
	A.NoError(err)

	resp, err := testserver.Client().Do(req)
	A.NoError(err)

	A.Equal(http.StatusNotFound, resp.StatusCode)
}
