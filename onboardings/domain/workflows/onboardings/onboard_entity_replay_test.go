package onboardings

import (
	"context"
	"github.com/stretchr/testify/suite"
	"github.com/temporalio/temporal-jumpstart-golang/app/testhelper"
	v1 "github.com/temporalio/temporal-jumpstart-golang/onboardings/generated/domain/v1"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/testsuite"
	"go.temporal.io/sdk/worker"
	"go.temporal.io/sdk/workflow"
	"testing"
)

// OnboardEntityReplayTestSuite
// https://docs.temporal.io/docs/go/testing/
type OnboardEntityReplayTestSuite struct {
	suite.Suite
	taskQueue string
	server    *testsuite.DevServer
	client    client.Client
	worker    worker.Worker
}

// SetupSuite https://pkg.go.dev/github.com/stretchr/testify/suite#SetupAllSuite
func (s *OnboardEntityReplayTestSuite) SetupSuite() {
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
func (s *OnboardEntityReplayTestSuite) SetupTest() {
}

// BeforeTest https://pkg.go.dev/github.com/stretchr/testify/suite#BeforeTest
func (s *OnboardEntityReplayTestSuite) BeforeTest(suiteName, testName string) {

}

// AfterTest https://pkg.go.dev/github.com/stretchr/testify/suite#AfterTest
func (s *OnboardEntityReplayTestSuite) AfterTest(suiteName, testName string) {
}
func (s *OnboardEntityReplayTestSuite) TearDownSuite() {
	s.worker.Stop()

	err := s.server.Stop()
	if err != nil {
		s.Fail("Failed to stop server: %w", err)
	}
}
func (s *OnboardEntityReplayTestSuite) Test_ReplayExposesNDE() {
	var historySource = TypeWorkflows.OnboardEntityV1
	var historyTarget = TypeWorkflows.OnboardEntity

	workflowTypeName, _ := testhelper.GetFunctionName(historySource)

	args := &v1.OnboardEntityRequest{
		Id:    testhelper.RandomString(),
		Value: testhelper.RandomString(),
	}

	ctx := context.Background()

	s.worker.RegisterWorkflow(historySource)
	run, err := s.client.ExecuteWorkflow(ctx,
		client.StartWorkflowOptions{
			ID:        args.Id,
			TaskQueue: s.taskQueue,
		},
		historySource, args)
	s.NoError(err)
	// either wait for the WF to complete or give some time for the Task to reply back with events
	s.NoError(run.Get(ctx, nil))

	// grab the WF history
	history, err := testhelper.GetWorkflowHistory(ctx, s.client, args.Id)

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
	suite.Run(t, &OnboardEntityReplayTestSuite{})
}
