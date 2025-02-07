package onboardings

import (
	"github.com/stretchr/testify/suite"
	"github.com/temporalio/temporal-jumpstart-golang/app/domain/scaffold/messages/workflows"
	"github.com/temporalio/temporal-jumpstart-golang/app/testhelper"
	"go.temporal.io/sdk/testsuite"
	"testing"
)

// MyWorkflowTestSuite
// https://docs.temporal.io/docs/go/testing/
type MyWorkflowTestSuite struct {
	suite.Suite
	testsuite.WorkflowTestSuite
	env *testsuite.TestWorkflowEnvironment
}

// SetupSuite https://pkg.go.dev/github.com/stretchr/testify/suite#SetupAllSuite
func (s *MyWorkflowTestSuite) SetupSuite() {

}

// SetupTest https://pkg.go.dev/github.com/stretchr/testify/suite#SetupTestSuite
// CAREFUL not to put this `env` inside the SetupSuite or else you will
// get interleaved test times between parallel tests (testify runs suite tests in parallel)
func (s *MyWorkflowTestSuite) SetupTest() {
	s.env = s.NewTestWorkflowEnvironment()
}

// BeforeTest https://pkg.go.dev/github.com/stretchr/testify/suite#BeforeTest
func (s *MyWorkflowTestSuite) BeforeTest(suiteName, testName string) {

}

// AfterTest https://pkg.go.dev/github.com/stretchr/testify/suite#AfterTest
func (s *MyWorkflowTestSuite) AfterTest(suiteName, testName string) {
	s.env.AssertExpectations(s.T())
}

func TestMyWorkflow(t *testing.T) {
	suite.Run(t, &MyWorkflowTestSuite{})
}

/* ============= TESTS =================== */
func (s *MyWorkflowTestSuite) Test_WorkflowExecutes() {
	s.env.RegisterWorkflow(TypeWorkflows.OnboardEntity)
	args := &workflows.StartMyWorkflowRequest{
		ID:    testhelper.RandomString(),
		Value: testhelper.RandomString(),
	}
	s.env.ExecuteWorkflow(TypeWorkflows.OnboardEntity, args)
	s.True(s.env.IsWorkflowCompleted())
	s.NoError(s.env.GetWorkflowError())
}
