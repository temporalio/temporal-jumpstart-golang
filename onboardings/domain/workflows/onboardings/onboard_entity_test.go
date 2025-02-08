package onboardings

import (
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"github.com/temporalio/temporal-jumpstart-golang/app/testhelper"
	"github.com/temporalio/temporal-jumpstart-golang/onboardings/domain/workflows"
	v1 "github.com/temporalio/temporal-jumpstart-golang/onboardings/generated/domain/v1"
	"go.temporal.io/sdk/testsuite"
	"testing"
	"time"
)

// OnboardEntityTestSuite
// https://docs.temporal.io/docs/go/testing/
type OnboardEntityTestSuite struct {
	suite.Suite
	testsuite.WorkflowTestSuite
	env *testsuite.TestWorkflowEnvironment
}

// SetupSuite https://pkg.go.dev/github.com/stretchr/testify/suite#SetupAllSuite
func (s *OnboardEntityTestSuite) SetupSuite() {

}

// SetupTest https://pkg.go.dev/github.com/stretchr/testify/suite#SetupTestSuite
// CAREFUL not to put this `env` inside the SetupSuite or else you will
// get interleaved test times between parallel tests (testify runs suite tests in parallel)
func (s *OnboardEntityTestSuite) SetupTest() {
	s.env = s.NewTestWorkflowEnvironment()
}

// BeforeTest https://pkg.go.dev/github.com/stretchr/testify/suite#BeforeTest
func (s *OnboardEntityTestSuite) BeforeTest(suiteName, testName string) {

}

// AfterTest https://pkg.go.dev/github.com/stretchr/testify/suite#AfterTest
func (s *OnboardEntityTestSuite) AfterTest(suiteName, testName string) {
	s.env.AssertExpectations(s.T())
}

func TestMyWorkflow(t *testing.T) {
	suite.Run(t, &OnboardEntityTestSuite{})
}

/* ============= TESTS =================== */
// behavior
func (s *OnboardEntityTestSuite) Test_GivenNeverApproved_DoesNotPerformOnboardingTasks() {
	s.env.RegisterWorkflow(TypeWorkflows.OnboardEntity)
	args := &v1.OnboardEntityRequest{
		Id:                       testhelper.RandomString(),
		Value:                    testhelper.RandomString(),
		CompletionTimeoutSeconds: 3000,
		DeputyOwnerEmail:         "",
		SkipApproval:             false,
	}
	s.env.OnActivity(OnboardEntityActivities.RegisterCrmEntity, mock.Anything).Never()
	s.env.OnActivity(OnboardEntityActivities.RegisterCrmEntity, mock.Anything).Never()

	s.env.ExecuteWorkflow(TypeWorkflows.OnboardEntity, args)
	s.True(s.env.IsWorkflowCompleted())
	s.NoError(s.env.GetWorkflowError())

	s.env.AssertExpectations(s.T())
}

// behavior
func (s *OnboardEntityTestSuite) Test_GivenApprovedNoDeputy_ShouldPerformOnboardingTasks() {
	s.env.RegisterWorkflow(TypeWorkflows.OnboardEntity)
	args := &v1.OnboardEntityRequest{
		Id:                       testhelper.RandomString(),
		Value:                    testhelper.RandomString(),
		CompletionTimeoutSeconds: 3000,
		DeputyOwnerEmail:         "",
		SkipApproval:             false,
	}
	approval := &v1.ApproveEntityRequest{Comment: testhelper.RandomString()}
	s.env.OnActivity(TypeOnboardActivities.RegisterCrmEntity, mock.Anything, &v1.RegisterCrmEntityRequest{
		Id:    args.Id,
		Value: args.Value,
	}).Once().Return(nil)
	s.env.OnActivity(TypeOnboardActivities.SendEmail, mock.Anything, &v1.RequestDeputyOwnerRequest{
		Id:               args.Id,
		DeputyOwnerEmail: args.DeputyOwnerEmail,
	}).Never()
	s.env.RegisterDelayedCallback(func() {
		s.env.SignalWorkflow(workflows.SignalName(approval), approval)
	}, time.Second*2)

	s.env.ExecuteWorkflow(TypeWorkflows.OnboardEntity, args)
	s.True(s.env.IsWorkflowCompleted())
	s.NoError(s.env.GetWorkflowError())

	s.env.AssertExpectations(s.T())
}
