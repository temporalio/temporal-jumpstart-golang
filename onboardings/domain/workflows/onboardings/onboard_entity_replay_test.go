package onboardings

import (
	"context"
	"github.com/stretchr/testify/suite"
	"github.com/temporalio/temporal-jumpstart-golang/onboardings/domain/workflows"
	commandsv1 "github.com/temporalio/temporal-jumpstart-golang/onboardings/generated/onboardings/domain/commands/v1"
	commandsv2 "github.com/temporalio/temporal-jumpstart-golang/onboardings/generated/onboardings/domain/commands/v2"
	workflowsv1 "github.com/temporalio/temporal-jumpstart-golang/onboardings/generated/onboardings/domain/workflows/v1"
	"github.com/temporalio/temporal-jumpstart-golang/onboardings/testhelper"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/testsuite"
	"go.temporal.io/sdk/worker"
	"go.temporal.io/sdk/workflow"
	"google.golang.org/protobuf/types/known/timestamppb"
	"testing"
	"time"
)

type testWorker struct {
	taskQueue string
	worker.Worker
}

func (w *testWorker) Run(t *testing.T, runner func()) func() {
	w.RegisterWorkflowWithOptions(OnboardEntityV1, workflow.RegisterOptions{Name: workflows.MustGetFunctionName(OnboardEntity)})
	if err := w.Worker.Start(); err != nil {
		t.Fatal(err)
	}
	runner()
	return func() {
		w.Worker.Stop()
	}
}

type testWorkerFactory func(taskQueue string) *testWorker

// OnboardEntityReplayTestSuite
// https://docs.temporal.io/docs/go/testing/
type OnboardEntityReplayTestSuite struct {
	suite.Suite
	server              *testsuite.DevServer
	client              client.Client
	sutWorkflowTypeName string
	workerFactory       testWorkerFactory
}

// SetupSuite https://pkg.go.dev/github.com/stretchr/testify/suite#SetupAllSuite
func (s *OnboardEntityReplayTestSuite) SetupSuite() {
	server, err := testsuite.StartDevServer(context.Background(), testsuite.DevServerOptions{
		LogLevel: "error",
	})
	s.Require().NoError(err)

	s.server = server
	s.client = server.Client()
	s.sutWorkflowTypeName, _ = testhelper.GetFunctionName(OnboardEntity)
	s.workerFactory = func(taskQueue string) *testWorker {
		w := worker.New(s.client, taskQueue, worker.Options{})
		return &testWorker{
			Worker:    w,
			taskQueue: taskQueue,
		}
	}
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
	err := s.server.Stop()
	if err != nil {
		s.Fail("Failed to stop server: %w", err)
	}
}

type activitiesDouble struct {
	registerCrmEntityError error
}

func (a *activitiesDouble) NotifyOnboardEntityCompleted(ctx context.Context, cmd *commandsv2.NotifyOnboardEntityCompletedRequest) error {
	return nil
}

func (a *activitiesDouble) RegisterCrmEntity(ctx context.Context, q *commandsv1.RegisterCrmEntityRequest) error {
	if a.registerCrmEntityError != nil {
		return a.registerCrmEntityError
	}
	return nil
}

func (a *activitiesDouble) SendDeputyOwnerApprovalRequest(ctx context.Context, cmd *commandsv1.RequestDeputyOwnerRequest) error {
	//TODO implement me
	panic("implement me")
}

func (s *OnboardEntityReplayTestSuite) Test_OnboardEntityV1_ToV2_ReplayWithApproval_NoPatch_ExposesNDE() {
	var historyTarget = OnboardEntityNDEExtraActivity

	var taskQueue = testhelper.RandomString()
	w := s.workerFactory(taskQueue)

	args := &workflowsv1.OnboardEntityRequest{
		Id:                       testhelper.RandomString(),
		Value:                    testhelper.RandomString(),
		CompletionTimeoutSeconds: 5,
		Timestamp:                timestamppb.Now(),
		DeputyOwnerEmail:         "",
	}

	ctx := context.Background()

	acts := &activitiesDouble{}

	w.RegisterActivity(acts)

	w.Run(s.T(), func() {
		run, err := s.client.ExecuteWorkflow(ctx,
			client.StartWorkflowOptions{
				ID:        args.Id,
				TaskQueue: taskQueue,
			},
			s.sutWorkflowTypeName, args)
		s.NoError(err)
		time.Sleep(time.Second * 2)
		var approval = &commandsv1.ApproveEntityRequest{Comment: testhelper.RandomString()}
		s.NoError(s.client.SignalWorkflow(ctx, run.GetID(), "", workflows.SignalName(approval), approval))
		// either wait for the WF to complete or give some time for the Task to reply back with events
		s.NoError(run.Get(ctx, nil))

		s.T().Log("Workflow completed successfully")

		// grab the WF history
		history, err := testhelper.GetWorkflowHistory(ctx, s.client, args.Id)

		// attempt replay
		replayer := worker.NewWorkflowReplayer()

		// here we use the RegisterOptions to slip in our new implementation registered by the old name
		replayer.RegisterWorkflowWithOptions(historyTarget,
			workflow.RegisterOptions{Name: s.sutWorkflowTypeName})

		s.ErrorContains(replayer.ReplayWorkflowHistoryWithOptions(
			nil,
			history,
			worker.ReplayWorkflowHistoryOptions{}), testhelper.NonDeterminismError)
	})

}

func (s *OnboardEntityReplayTestSuite) Test_OnboardEntityV1_ToV2_GivenPatched_DoesNotNDE() {
	var historyTarget = OnboardEntity

	args := &workflowsv1.OnboardEntityRequest{
		Id:                       testhelper.RandomString(),
		Value:                    testhelper.RandomString(),
		CompletionTimeoutSeconds: 5,
		Timestamp:                timestamppb.Now(),
		DeputyOwnerEmail:         "",
	}

	ctx := context.Background()

	var taskQueue = testhelper.RandomString()
	w := s.workerFactory(taskQueue)

	acts := &activitiesDouble{}

	w.RegisterActivity(acts)
	w.Run(s.T(), func() {
		run, err := s.client.ExecuteWorkflow(ctx,
			client.StartWorkflowOptions{
				ID:        args.Id,
				TaskQueue: taskQueue,
			},
			s.sutWorkflowTypeName, args)
		s.NoError(err)
		time.Sleep(time.Second * 2)
		var approval = &commandsv1.ApproveEntityRequest{Comment: testhelper.RandomString()}
		s.NoError(s.client.SignalWorkflow(ctx, run.GetID(), "", workflows.SignalName(approval), approval))
		// either wait for the WF to complete or give some time for the Task to reply back with events
		s.NoError(run.Get(ctx, nil))

		s.T().Log("Workflow completed successfully")

		// grab the WF history
		history, err := testhelper.GetWorkflowHistory(ctx, s.client, args.Id)

		// attempt replay
		replayer := worker.NewWorkflowReplayer()

		// here we use the RegisterOptions to slip in our new implementation registered by the old name
		replayer.RegisterWorkflowWithOptions(historyTarget,
			workflow.RegisterOptions{Name: s.sutWorkflowTypeName})

		s.NoError(replayer.ReplayWorkflowHistoryWithOptions(
			nil,
			history,
			worker.ReplayWorkflowHistoryOptions{}), testhelper.NonDeterminismError)
	})

}

func TestMyWorkflowReplay(t *testing.T) {
	suite.Run(t, &OnboardEntityReplayTestSuite{})
}
