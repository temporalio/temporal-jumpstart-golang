package onboardings

import (
	"errors"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"github.com/temporalio/temporal-jumpstart-golang/app/testhelper"
	"github.com/temporalio/temporal-jumpstart-golang/onboardings/domain/workflows"
	v1 "github.com/temporalio/temporal-jumpstart-golang/onboardings/generated/domain/v1"
	"go.temporal.io/sdk/converter"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/testsuite"
	"go.temporal.io/sdk/workflow"
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
	//s.env.AssertExpectations(s.T())
}

func TestMyWorkflow(t *testing.T) {
	suite.Run(t, &OnboardEntityTestSuite{})
}

/* ============= TESTS =================== */
// [state]
// This demonstrates using golang table tests to perform validation
// but notice that we  disregard the underlying TestSuite environment to get complete isolation.
// This is not completely necessary but here for demonstration.
func (s *OnboardEntityTestSuite) Test_GivenInvalidArgs_ShouldFailTheWorkflow() {

	type test struct {
		args *v1.OnboardEntityRequest
		name string
	}
	cases := []test{
		{
			name: "missing id",
			args: &v1.OnboardEntityRequest{
				Id:                       "",
				Value:                    testhelper.RandomString(),
				CompletionTimeoutSeconds: 0,
				DeputyOwnerEmail:         "",
				SkipApproval:             false,
			},
		},
		{
			name: "missing value",
			args: &v1.OnboardEntityRequest{
				Id:                       testhelper.RandomString(),
				Value:                    "",
				CompletionTimeoutSeconds: 0,
				DeputyOwnerEmail:         "",
				SkipApproval:             false,
			},
		},
	}
	for _, tc := range cases {
		s.Run(tc.name, func() {
			// disregard the suite test environment for these
			caseEnv := s.NewTestWorkflowEnvironment()
			caseEnv.RegisterWorkflow(TypeWorkflows.OnboardEntity)
			caseEnv.OnActivity(OnboardEntityActivities.RegisterCrmEntity, mock.Anything, mock.Anything).Never()
			caseEnv.OnActivity(OnboardEntityActivities.SendEmail, mock.Anything, mock.Anything).Never()
			caseEnv.ExecuteWorkflow(TypeWorkflows.OnboardEntity, tc.args)
			s.True(caseEnv.IsWorkflowCompleted())
			werr := caseEnv.GetWorkflowError()
			s.NotNil(werr)
			var appErr *temporal.ApplicationError
			errors.As(werr, &appErr)
			s.Equal(v1.Errors_ERR_ONBOARD_ENTITY_INVALID_ARGS.String(), appErr.Type())
		})
	}

}

// [behavior]
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

// [behavior]
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

// [state]: ContinueAsNew Test
func (s *OnboardEntityTestSuite) Test_GivenDeputyWithNoApprovalReceived_ShouldContinueAsNewWithNewArgs() {
	s.env.RegisterWorkflow(TypeWorkflows.OnboardEntity)
	args := &v1.OnboardEntityRequest{
		Id:                       testhelper.RandomString(),
		Value:                    testhelper.RandomString(),
		CompletionTimeoutSeconds: 3000,
		DeputyOwnerEmail:         "deputydawg@temporal.io",
		SkipApproval:             false,
	}
	s.env.OnActivity(TypeOnboardActivities.RegisterCrmEntity, mock.Anything, mock.Anything).Never()
	s.env.OnActivity(TypeOnboardActivities.SendEmail, mock.Anything, mock.Anything).Once().Return(nil)

	s.env.ExecuteWorkflow(TypeWorkflows.OnboardEntity, args)
	s.True(s.env.IsWorkflowCompleted())

	expectCompletionTimeoutSeconds := args.CompletionTimeoutSeconds - calculateWaitSeconds(args)
	werr := s.env.GetWorkflowError()
	// this shows how to test for a ContinueAsNew
	can := &workflow.ContinueAsNewError{}
	s.True(errors.As(werr, &can))
	canWFType, _ := testhelper.GetFunctionName(TypeWorkflows.OnboardEntity)
	s.Equal(canWFType, can.WorkflowType.Name)
	canParams := &v1.OnboardEntityRequest{}
	dc := converter.GetDefaultDataConverter()
	s.NoError(dc.FromPayloads(can.Input, canParams))
	s.Equal(expectCompletionTimeoutSeconds, canParams.CompletionTimeoutSeconds)
	s.Equal("", canParams.DeputyOwnerEmail)
	s.Equal(false, canParams.SkipApproval)
}

// [behavior]: ContinueAsNew Test
func (s *OnboardEntityTestSuite) Test_GivenDeputyWithNoApprovalReceived_ShouldRequestDeputyApproval() {
	s.env.RegisterWorkflow(TypeWorkflows.OnboardEntity)
	args := &v1.OnboardEntityRequest{
		Id:                       testhelper.RandomString(),
		Value:                    testhelper.RandomString(),
		CompletionTimeoutSeconds: 3000,
		DeputyOwnerEmail:         "deputydawg@temporal.io",
		SkipApproval:             false,
	}
	s.env.OnActivity(TypeOnboardActivities.RegisterCrmEntity, mock.Anything, mock.Anything).Never()
	s.env.OnActivity(TypeOnboardActivities.SendEmail, mock.Anything, &v1.RequestDeputyOwnerRequest{
		Id:               args.Id,
		DeputyOwnerEmail: args.DeputyOwnerEmail,
	}).Once().Return(nil)

	s.env.ExecuteWorkflow(TypeWorkflows.OnboardEntity, args)
	s.True(s.env.IsWorkflowCompleted())

	s.env.AssertExpectations(s.T())
}
func (s *OnboardEntityTestSuite) Test_GivenRejection_ShouldNotPerformOnboardingTasks() {

}
