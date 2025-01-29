package workflows

import (
	"github.com/stretchr/testify/suite"
	"go.temporal.io/sdk/testsuite"
	"testing"
)

// ScaffoldWorkflowTestSuite
// https://docs.temporal.io/docs/go/testing/
type ScaffoldWorkflowTestSuite struct {
	suite.Suite
	testsuite.WorkflowTestSuite
	env *testsuite.TestWorkflowEnvironment
}

// SetupSuite https://pkg.go.dev/github.com/stretchr/testify/suite#SetupAllSuite
func (s *ScaffoldWorkflowTestSuite) SetupSuite() {

}

// SetupTest https://pkg.go.dev/github.com/stretchr/testify/suite#SetupTestSuite
// CAREFUL not to put this `env` inside the SetupSuite or else you will
// get interleaved test times between parallel tests (testify runs suite tests in parallel)
func (s *ScaffoldWorkflowTestSuite) SetupTest() {
	s.env = s.NewTestWorkflowEnvironment()
}

// BeforeTest https://pkg.go.dev/github.com/stretchr/testify/suite#BeforeTest
func (s *ScaffoldWorkflowTestSuite) BeforeTest(suiteName, testName string) {

}

// AfterTest https://pkg.go.dev/github.com/stretchr/testify/suite#AfterTest
func (s *ScaffoldWorkflowTestSuite) AfterTest(suiteName, testName string) {
	s.env.AssertExpectations(s.T())
}

func TestScaffoldWorkflow(t *testing.T) {
	suite.Run(t, &ScaffoldWorkflowTestSuite{})
}
