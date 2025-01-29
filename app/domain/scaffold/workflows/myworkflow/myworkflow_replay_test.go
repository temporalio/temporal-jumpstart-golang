package myworkflow

import (
	"github.com/stretchr/testify/suite"
	"github.com/temporalio/temporal-jumpstart-golang/app/domain/scaffold/messages/workflows"
	"github.com/temporalio/temporal-jumpstart-golang/app/domain/testhelper"
	"go.temporal.io/sdk/testsuite"
	"testing"
)

// MyWorkflowReplayTestSuite
// https://docs.temporal.io/docs/go/testing/
type MyWorkflowReplayTestSuite struct {
	suite.Suite
	testsuite.WorkflowTestSuite
	env *testsuite.TestWorkflowEnvironment
}

// SetupSuite https://pkg.go.dev/github.com/stretchr/testify/suite#SetupAllSuite
func (s *MyWorkflowReplayTestSuite) SetupSuite() {

}

// SetupTest https://pkg.go.dev/github.com/stretchr/testify/suite#SetupTestSuite
// CAREFUL not to put this `env` inside the SetupSuite or else you will
// get interleaved test times between parallel tests (testify runs suite tests in parallel)
func (s *MyWorkflowReplayTestSuite) SetupTest() {
	s.env = s.NewTestWorkflowEnvironment()
}

// BeforeTest https://pkg.go.dev/github.com/stretchr/testify/suite#BeforeTest
func (s *MyWorkflowReplayTestSuite) BeforeTest(suiteName, testName string) {

}

// AfterTest https://pkg.go.dev/github.com/stretchr/testify/suite#AfterTest
func (s *MyWorkflowReplayTestSuite) AfterTest(suiteName, testName string) {
	s.env.AssertExpectations(s.T())
}

func (s *MyWorkflowReplayTestSuite) Test_Replay() {
	s.env.RegisterWorkflow(TypeWorkflows.MyWorkflow)
	args := &workflows.StartMyWorkflowRequest{
		ID:    testhelper.RandomString(),
		Value: testhelper.RandomString(),
	}

	s.env.ExecuteWorkflow(TypeWorkflows.MyWorkflow, args)
}
func TestMyWorkflowReplay(t *testing.T) {
	suite.Run(t, &MyWorkflowReplayTestSuite{})
}
