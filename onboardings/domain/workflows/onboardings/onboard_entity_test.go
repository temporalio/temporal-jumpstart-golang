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
	"google.golang.org/protobuf/types/known/timestamppb"
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
// This demonstrates using table tests to perform validation;
// but notice that we  disregard the underlying TestSuite environment to get complete isolation.
// This is not completely necessary but here for demonstration.
func (s *OnboardEntityTestSuite) Test_GivenInvalidArgs_ShouldFailTheWorkflow() {

	type test struct {
		args *v1.OnboardEntityRequest
		name string
	}
	// all of these cases should raise failures
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
		{
			name: "missing timestamp",
			args: &v1.OnboardEntityRequest{
				Id:                       testhelper.RandomString(),
				Value:                    testhelper.RandomString(),
				CompletionTimeoutSeconds: 3000,
				DeputyOwnerEmail:         "",
				SkipApproval:             false,
				Timestamp:                timestamppb.New(time.Time{}),
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
			s.Equal(v1.Errors_ERRORS_ONBOARD_ENTITY_INVALID_ARGS.String(), appErr.Type())
		})
	}

}

// [state]
func (s *OnboardEntityTestSuite) Test_OnboardingThresholdHasAlreadyPassed() {
	s.env.RegisterWorkflow(TypeWorkflows.OnboardEntity)
	past := time.Now().AddDate(0, 0, -1)

	args := &v1.OnboardEntityRequest{
		Id:    testhelper.RandomString(),
		Value: testhelper.RandomString(),
		// always be late
		CompletionTimeoutSeconds: uint64(time.Now().Sub(past).Seconds() / 2),
		DeputyOwnerEmail:         "deputydawg@temporal.io",
		SkipApproval:             false,
		Timestamp:                timestamppb.New(past),
	}

	s.env.ExecuteWorkflow(TypeWorkflows.OnboardEntity, args)
	s.True(s.env.IsWorkflowCompleted())
	werr := s.env.GetWorkflowError()
	s.NotNil(werr)
	var appErr *temporal.ApplicationError
	errors.As(werr, &appErr)
	// this will be raised when there is no approval too, so we need to check
	// that this is being raised due to late workflow run OR we could raise a more specific error type.
	//
	s.Equal(v1.Errors_ERRORS_ONBOARD_ENTITY_TIMED_OUT.String(), appErr.Type())
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
	s.env.OnActivity(OnboardEntityActivities.SendEmail, mock.Anything).Never()

	s.env.ExecuteWorkflow(TypeWorkflows.OnboardEntity, args)
	s.True(s.env.IsWorkflowCompleted())
	s.NoError(s.env.GetWorkflowError())

	s.env.AssertExpectations(s.T())
}

// [state]
func (s *OnboardEntityTestSuite) Test_GivenPendingApproval_OnboardShouldExposeTimeRemainingForApproval() {
	s.env.RegisterWorkflow(TypeWorkflows.OnboardEntity)

	// we want to complete this within 7 days
	completionTimeoutSeconds := uint64((7 * 24 * time.Hour).Seconds())
	// message was minted with Jan 1, 2025 @ 6am as start time
	// that means we have until Jan 8, 2025 @ 6am to get approved
	actualStartTime := time.Date(2025, time.January, 1, 6, 0, 0, 0, time.UTC)

	// worker did not run this workflow until Jan 2, 2025 @ 6am
	workerAliveTime := time.Date(2025, time.January, 2, 6, 0, 0, 0, time.UTC)
	s.env.SetStartTime(workerAliveTime)

	// We will wait two days before querying for the time remaining;
	// this means two days after the Worker has picked up the Task to start the Workflow.
	// Since the Worker was down for a day from the time the message was sent, that means
	// the Query is requested THREE days from the time we actually started the onboard process (Jan 4, 2025)
	// which means we have FOUR days remaining before giving up
	waitBeforeQuery := 2 * 24 * time.Hour
	expectApprovalTimeRemaining := 4 * 24 * time.Hour // completionTimeoutSeconds - (uint64(waitBeforeQuery + ((1 * 24 * time.Hour) / time.Second)))
	var actualState *v1.EntityOnboardingStateResponse
	s.env.RegisterDelayedCallback(func() {
		val, err := s.env.QueryWorkflow(QueryEntityOnboardingState)
		s.NoError(err)
		s.NoError(val.Get(&actualState))
	}, waitBeforeQuery)

	args := &v1.OnboardEntityRequest{
		Id:                       testhelper.RandomString(),
		Value:                    testhelper.RandomString(),
		CompletionTimeoutSeconds: completionTimeoutSeconds,
		DeputyOwnerEmail:         "",
		SkipApproval:             false,
		Timestamp:                timestamppb.New(actualStartTime),
	}
	s.env.OnActivity(TypeOnboardActivities.RegisterCrmEntity, mock.Anything, &v1.RegisterCrmEntityRequest{
		Id:    args.Id,
		Value: args.Value,
	}).Once().Return(nil)
	s.env.OnActivity(TypeOnboardActivities.SendEmail, mock.Anything, &v1.RequestDeputyOwnerRequest{
		Id:               args.Id,
		DeputyOwnerEmail: args.DeputyOwnerEmail,
	}).Never()

	s.env.ExecuteWorkflow(TypeWorkflows.OnboardEntity, args)
	s.True(s.env.IsWorkflowCompleted())
	// check the expected approval time remaining overall but accommodate task timeout as reasonable drift
	s.InDelta(int(expectApprovalTimeRemaining.Seconds()), int(actualState.GetApprovalTimeRemainingSeconds()), 10)
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
		Timestamp:                timestamppb.New(time.Now()),
	}
	s.env.OnActivity(TypeOnboardActivities.RegisterCrmEntity, mock.Anything, mock.Anything).Never()
	s.env.OnActivity(TypeOnboardActivities.SendEmail, mock.Anything, mock.Anything).Once().Return(nil)

	s.env.ExecuteWorkflow(TypeWorkflows.OnboardEntity, args)
	s.True(s.env.IsWorkflowCompleted())

	expectCompletionTimeoutSeconds := args.CompletionTimeoutSeconds - calculateWaitSeconds(args.Timestamp.AsTime(), args)
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

// [state]
func (s *OnboardEntityTestSuite) Test_GivenRejection_ShouldRevealRejectionComment_NotPerformOnboardingTasks() {
	s.env.RegisterWorkflow(TypeWorkflows.OnboardEntity)
	args := &v1.OnboardEntityRequest{
		Id:                       testhelper.RandomString(),
		Value:                    testhelper.RandomString(),
		CompletionTimeoutSeconds: 3000,
		DeputyOwnerEmail:         "",
		SkipApproval:             false,
	}
	rejection := &v1.RejectEntityRequest{Comment: testhelper.RandomString()}
	state := &v1.EntityOnboardingStateResponse{}
	s.env.RegisterDelayedCallback(func() {
		s.env.SignalWorkflow(workflows.SignalName(rejection), rejection)
	}, time.Second*2)
	s.env.RegisterDelayedCallback(func() {
		val, err := s.env.QueryWorkflow(QueryEntityOnboardingState)
		s.NoError(err)
		s.NoError(val.Get(&state))
	}, time.Second*3)

	s.env.ExecuteWorkflow(TypeWorkflows.OnboardEntity, args)
	s.True(s.env.IsWorkflowCompleted())
	s.Error(s.env.GetWorkflowError())

	s.Equal(rejection.Comment, state.Approval.GetComment())
}

// [behavior]
func (s *OnboardEntityTestSuite) Test_GivenRejection_ShouldNotPerformOnboardingTasks() {
	s.env.RegisterWorkflow(TypeWorkflows.OnboardEntity)
	args := &v1.OnboardEntityRequest{
		Id:                       testhelper.RandomString(),
		Value:                    testhelper.RandomString(),
		CompletionTimeoutSeconds: 3000,
		DeputyOwnerEmail:         "",
		SkipApproval:             false,
	}
	approval := &v1.RejectEntityRequest{Comment: testhelper.RandomString()}
	s.env.OnActivity(TypeOnboardActivities.RegisterCrmEntity, mock.Anything, mock.Anything).Never()
	s.env.OnActivity(TypeOnboardActivities.SendEmail, mock.Anything, mock.Anything).Never()
	s.env.RegisterDelayedCallback(func() {
		s.env.SignalWorkflow(workflows.SignalName(approval), approval)
	}, time.Second*2)

	s.env.ExecuteWorkflow(TypeWorkflows.OnboardEntity, args)
	s.True(s.env.IsWorkflowCompleted())
	s.NoError(s.env.GetWorkflowError())

	s.env.AssertExpectations(s.T())
}
