package workflows

import (
	"context"
	"github.com/stretchr/testify/suite"
	"github.com/temporalio/temporal-jumpstart-golang/app/domain/testhelper"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/testsuite"
	"go.temporal.io/sdk/worker"
	"testing"
)

// ScaffoldReplayTestSuite
// https://docs.temporal.io/docs/go/testing/
// This utilizes the DevServer https://pkg.go.dev/go.temporal.io/sdk@v1.32.1/testsuite#DevServer and the
// https://pkg.go.dev/go.temporal.io/sdk/internal#WorkflowReplayer to validate Workflow Code against
// Non-Determinism Exceptions.
// The general flow of this kind of test is:
// 1. Execute the source Workflow Type
// 2. Extract History
// 3. Replay the extracted history against the target Workflow Type
type ScaffoldReplayTestSuite struct {
	suite.Suite
	taskQueue string
	server    *testsuite.DevServer
	client    client.Client
	worker    worker.Worker
}

// SetupSuite https://pkg.go.dev/github.com/stretchr/testify/suite#SetupAllSuite
func (s *ScaffoldReplayTestSuite) SetupSuite() {
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
func (s *ScaffoldReplayTestSuite) SetupTest() {
}

// BeforeTest https://pkg.go.dev/github.com/stretchr/testify/suite#BeforeTest
func (s *ScaffoldReplayTestSuite) BeforeTest(suiteName, testName string) {

}

// AfterTest https://pkg.go.dev/github.com/stretchr/testify/suite#AfterTest
func (s *ScaffoldReplayTestSuite) AfterTest(suiteName, testName string) {
}
func (s *ScaffoldReplayTestSuite) TearDownSuite() {
	s.worker.Stop()

	err := s.server.Stop()
	if err != nil {
		s.Fail("Failed to stop server: %w", err)
	}
}
func (s *ScaffoldReplayTestSuite) Test_ReplayExposesNDE() {

	// execute
	// extract
	// replay
}

func TestScaffoldReplayTestSuite(t *testing.T) {
	suite.Run(t, &ScaffoldReplayTestSuite{})
}
