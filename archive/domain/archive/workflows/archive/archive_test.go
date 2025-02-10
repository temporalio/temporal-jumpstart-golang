package archive

import (
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"github.com/temporalio/temporal-jumpstart-golang/app/testhelper"
	"github.com/temporalio/temporal-jumpstart-golang/archive/domain/archive/workflows/activities"
	v1 "github.com/temporalio/temporal-jumpstart-golang/archive/generated/domain/v1"
	"go.temporal.io/sdk/testsuite"
	"testing"
)

// ArchiveTestSuite
// https://docs.temporal.io/docs/go/testing/
type ArchiveTestSuite struct {
	suite.Suite
	testsuite.WorkflowTestSuite
	env *testsuite.TestWorkflowEnvironment
}

// SetupSuite https://pkg.go.dev/github.com/stretchr/testify/suite#SetupAllSuite
func (s *ArchiveTestSuite) SetupSuite() {

}

// SetupTest https://pkg.go.dev/github.com/stretchr/testify/suite#SetupTestSuite
// CAREFUL not to put this `env` inside the SetupSuite or else you will
// get interleaved test times between parallel tests (testify runs suite tests in parallel)
func (s *ArchiveTestSuite) SetupTest() {
	s.env = s.NewTestWorkflowEnvironment()
}

// BeforeTest https://pkg.go.dev/github.com/stretchr/testify/suite#BeforeTest
func (s *ArchiveTestSuite) BeforeTest(suiteName, testName string) {

}

// AfterTest https://pkg.go.dev/github.com/stretchr/testify/suite#AfterTest
func (s *ArchiveTestSuite) AfterTest(suiteName, testName string) {
	s.env.AssertExpectations(s.T())
}

func TestMyWorkflow(t *testing.T) {
	suite.Run(t, &ArchiveTestSuite{})
}

/* ============= TESTS =================== */
func (s *ArchiveTestSuite) Test_ArchiveWorkflow() {
	s.env.RegisterWorkflow(TypeWorkflows.Archive)
	args := &v1.StartArchiveRequest{
		Id: testhelper.RandomString(),
	}
	s.env.OnActivity(activities.TypeArchiveActivities.MutateActivity,
		mock.Anything,
	)
	s.env.ExecuteWorkflow(TypeWorkflows.Archive, args)
	s.True(s.env.IsWorkflowCompleted())
	s.NoError(s.env.GetWorkflowError())

}
