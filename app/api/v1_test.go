package api

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/temporalio/temporal-jumpstart-golang/app/api/messages"
	"github.com/temporalio/temporal-jumpstart-golang/app/clients"
	"github.com/temporalio/temporal-jumpstart-golang/app/config"
	"github.com/temporalio/temporal-jumpstart-golang/app/testhelper"
	"go.temporal.io/api/enums/v1"
	"go.temporal.io/sdk/client"
	"net/http"
	"net/http/httptest"
	"net/url"
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
func TestV1PutPing(t *testing.T) {
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
		}), []interface{}{body.Ping}).Return(&TestWorkflowRun{id: workflowId}, nil)
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

	resp, err := testserver.Client().Do(req)
	A.NoError(err)

	A.Equal(202, resp.StatusCode)

}
