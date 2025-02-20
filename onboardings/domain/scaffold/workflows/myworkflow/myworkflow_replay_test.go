package myworkflow

import (
	"context"
	"github.com/stretchr/testify/suite"
	"github.com/temporalio/temporal-jumpstart-golang/onboardings/domain/scaffold/messages/workflows"
	"github.com/temporalio/temporal-jumpstart-golang/onboardings/testhelper"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/testsuite"
	"go.temporal.io/sdk/worker"
	"go.temporal.io/sdk/workflow"
	"testing"
)

// MyWorkflowReplayTestSuite
// https://docs.temporal.io/docs/go/testing/
type MyWorkflowReplayTestSuite struct {
	suite.Suite
	taskQueue string
	server    *testsuite.DevServer
	client    client.Client
	worker    worker.Worker
}

// SetupSuite https://pkg.go.dev/github.com/stretchr/testify/suite#SetupAllSuite
func (s *MyWorkflowReplayTestSuite) SetupSuite() {
	server, err := testsuite.StartDevServer(context.Background(), testsuite.DevServerOptions{
		LogLevel: "error",
	})
	s.Require().NoError(err)
	s.taskQueue = testhelper.RandomString()
	s.server = server
	s.client = server.Client()
	s.worker = worker.New(s.client, s.taskQueue, worker.Options{})
	s.Require().NoError(s.worker.Start())
}

// SetupTest https://pkg.go.dev/github.com/stretchr/testify/suite#SetupTestSuite
func (s *MyWorkflowReplayTestSuite) SetupTest() {
}

// BeforeTest https://pkg.go.dev/github.com/stretchr/testify/suite#BeforeTest
func (s *MyWorkflowReplayTestSuite) BeforeTest(suiteName, testName string) {

}

// AfterTest https://pkg.go.dev/github.com/stretchr/testify/suite#AfterTest
func (s *MyWorkflowReplayTestSuite) AfterTest(suiteName, testName string) {
}
func (s *MyWorkflowReplayTestSuite) TearDownSuite() {
	s.worker.Stop()

	err := s.server.Stop()
	if err != nil {
		s.Fail("Failed to stop server: %w", err)
	}
}
func (s *MyWorkflowReplayTestSuite) Test_ReplayExposesNDE() {
	var historySource = TypeWorkflows.MyWorkflow
	var historyTarget = TypeWorkflows.MyWorkflowV1

	workflowTypeName, _ := testhelper.GetFunctionName(historySource)

	args := &workflows.StartMyWorkflowRequest{
		ID:    testhelper.RandomString(),
		Value: testhelper.RandomString(),
	}

	ctx := context.Background()

	s.worker.RegisterWorkflow(historySource)
	run, err := s.client.ExecuteWorkflow(ctx,
		client.StartWorkflowOptions{
			ID:        args.ID,
			TaskQueue: s.taskQueue,
		},
		historySource, args)
	s.NoError(err)
	// either wait for the WF to complete or give some time for the Task to reply back with events
	s.NoError(run.Get(ctx, nil))

	// grab the WF history
	history, err := testhelper.GetWorkflowHistory(ctx, s.client, args.ID)

	// attempt replay
	replayer := worker.NewWorkflowReplayer()

	// here we use the RegisterOptions to slip in our new implementation registered by the old name
	replayer.RegisterWorkflowWithOptions(historyTarget,
		workflow.RegisterOptions{Name: workflowTypeName})

	s.ErrorContains(replayer.ReplayWorkflowHistoryWithOptions(
		nil,
		history,
		worker.ReplayWorkflowHistoryOptions{}), "nondeterministic workflow definition code")
}

func TestMyWorkflowReplay(t *testing.T) {
	suite.Run(t, &MyWorkflowReplayTestSuite{})
}
