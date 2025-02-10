package activities

import (
	"github.com/stretchr/testify/suite"
	"github.com/temporalio/temporal-jumpstart-golang/app/domain/scaffold/messages/commands"
	"github.com/temporalio/temporal-jumpstart-golang/app/domain/scaffold/messages/queries"
	"github.com/temporalio/temporal-jumpstart-golang/app/testhelper"
	"go.temporal.io/sdk/testsuite"
	"reflect"
	"testing"
)

type MyWorkflowActivitiesTestSuite struct {
	suite.Suite
	testsuite.WorkflowTestSuite
	env *testsuite.TestActivityEnvironment
}

// SetupSuite https://pkg.go.dev/github.com/stretchr/testify/suite#SetupAllSuite
func (s *MyWorkflowActivitiesTestSuite) SetupSuite() {

}

// SetupTest https://pkg.go.dev/github.com/stretchr/testify/suite#SetupTestSuite
// CAREFUL not to put this `env` inside the SetupSuite or else you will
// get interleaved test times between parallel tests (testify runs suite tests in parallel)
func (s *MyWorkflowActivitiesTestSuite) SetupTest() {
}

// BeforeTest https://pkg.go.dev/github.com/stretchr/testify/suite#BeforeTest
func (s *MyWorkflowActivitiesTestSuite) BeforeTest(suiteName, testName string) {
	s.env = s.NewTestActivityEnvironment()
}

// AfterTest https://pkg.go.dev/github.com/stretchr/testify/suite#AfterTest
func (s *MyWorkflowActivitiesTestSuite) AfterTest(suiteName, testName string) {

}

func TestMyWorkflowActivities(t *testing.T) {
	suite.Run(t, &MyWorkflowActivitiesTestSuite{})
}

/* ============= TESTS =================== */
func (s *MyWorkflowActivitiesTestSuite) Test_QueryActivity() {
	s.env.RegisterActivity(TypeMyWorkflowActivities.QueryActivity)

	args := &queries.QueryActivityRequest{ID: testhelper.RandomString()}
	val, err := s.env.ExecuteActivity(TypeMyWorkflowActivities.QueryActivity, args)
	s.NoError(err)

	expect := &queries.QueryActivityResponse{
		ID: args.ID,
	}

	var actual *queries.QueryActivityResponse

	s.NoError(val.Get(&actual))
	s.True(reflect.DeepEqual(expect, actual))
}
func (s *MyWorkflowActivitiesTestSuite) Test_MutateActivity() {
	s.env.RegisterActivity(TypeMyWorkflowActivities.MutateActivity)

	args := &commands.MutateActivityRequest{ID: testhelper.RandomString()}
	_, err := s.env.ExecuteActivity(TypeMyWorkflowActivities.MutateActivity, args)
	s.NoError(err)
}
