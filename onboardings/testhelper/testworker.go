package testhelper

import (
	"github.com/temporalio/temporal-jumpstart-golang/onboardings/domain/workflows"
	"go.temporal.io/sdk/worker"
	"go.temporal.io/sdk/workflow"
	"testing"
)

// TestWorker simplifies running a Workflow in a scoped Worker on a custom TaskQueue
type TestWorker struct {
	worker.Worker
	TaskQueue        string
	WorkflowType     interface{}
	WorkflowTypeName string
}

func (w *TestWorker) Run(t *testing.T, runner func()) func() {
	wfTypeName := w.WorkflowTypeName
	if wfTypeName == "" {
		wfTypeName = workflows.MustGetFunctionName(w.WorkflowType)
	}
	w.RegisterWorkflowWithOptions(w.WorkflowType, workflow.RegisterOptions{Name: wfTypeName})
	if err := w.Worker.Start(); err != nil {
		t.Fatal(err)
	}
	runner()
	return func() {
		w.Worker.Stop()
	}
}

// TestWorkerFactory creates a TestWorker given a TaskQueue
type TestWorkerFactory func(taskQueue string) *TestWorker
