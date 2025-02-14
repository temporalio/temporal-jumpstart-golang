package onboardings

import (
	workflowsv1 "github.com/temporalio/temporal-jumpstart-golang/onboardings/generated/onboardings/domain/workflows/v1"
	workflowsv2 "github.com/temporalio/temporal-jumpstart-golang/onboardings/generated/onboardings/domain/workflows/v2"
	"go.temporal.io/sdk/temporal"
	"math"
	"strings"
	"time"
)

// assertValidArgsv1 is a poor-man's implementation not providing very specific details on which args are "invalid"
// but instead just failing hard in the event of missing or bad args
func assertValidArgsv1(args *workflowsv1.OnboardEntityRequest) error {
	if strings.TrimSpace(args.Id) != "" &&
		strings.TrimSpace(args.Value) != "" &&
		!args.Timestamp.AsTime().IsZero() {
		return nil
	}
	return temporal.NewApplicationError(workflowsv1.Errors_ERRORS_ONBOARD_ENTITY_INVALID_ARGS.String(), workflowsv1.Errors_ERRORS_ONBOARD_ENTITY_INVALID_ARGS.String())
}

// assertValidArgsv2 is a poor-man's implementation not providing very specific details on which args are "invalid"
// but instead just failing hard in the event of missing or bad args
func assertValidArgsv2(args *workflowsv2.OnboardEntityRequest) error {
	if strings.TrimSpace(args.Id) != "" &&
		strings.TrimSpace(args.Value) != "" &&
		!args.Timestamp.AsTime().IsZero() {
		return nil
	}
	return temporal.NewApplicationError(workflowsv1.Errors_ERRORS_ONBOARD_ENTITY_INVALID_ARGS.String(), workflowsv1.Errors_ERRORS_ONBOARD_ENTITY_INVALID_ARGS.String())
}

// onboardEntityDurationCalculator is message version agnostic and calculations time spans and enforces product requirements
// for things like timeouts
type onboardEntityDurationCalculator struct {
	completionTimeoutSeconds uint64
	skipApproval             bool
	timestamp                time.Time
	hasDeputyOwner           bool
}

// resolveCompletionTimeoutSeconds calculates a valid completion timeout duration
func (c *onboardEntityDurationCalculator) resolveCompletionTimeoutSeconds(now time.Time) uint64 {
	if c.skipApproval {
		return 0
	}
	completionTimeoutSeconds := c.completionTimeoutSeconds
	if completionTimeoutSeconds == 0 {
		completionTimeoutSeconds = DefaultCompletionTimeoutSeconds
	}
	return completionTimeoutSeconds
}

// calculateWaitSeconds calcs the time remaining before the Workflow shall make progress
// taking into consideration the DeputyOwner argument and the Product requirement that we shall
// give 40% of the time before reaching out to the deputy to ask for approval.
func (c *onboardEntityDurationCalculator) calculateWaitSeconds(now time.Time) uint64 {
	completionTimeoutSeconds := c.resolveCompletionTimeoutSeconds(now)
	if completionTimeoutSeconds == 0 {
		return completionTimeoutSeconds
	}
	waitSeconds := completionTimeoutSeconds
	if c.hasDeputyOwner {
		// if a deputy owner is provided, wait for 40% of the time (per the spec) to approve before next steps
		waitSeconds = uint64(.4 * float64(waitSeconds))
	}
	threshold := c.timestamp.Add(time.Second * time.Duration(waitSeconds))

	return uint64(math.Max(threshold.Sub(now).Seconds(), 0))
}

func (c *onboardEntityDurationCalculator) calculateCompletionThreshold(now time.Time) time.Time {
	return c.timestamp.Add(time.Second * time.Duration(c.resolveCompletionTimeoutSeconds(now)))
}

// tryHandleCancellation
func tryHandleCancellation(err error) error {
	if temporal.IsCanceledError(err) {
		return err
	}
	return nil
}
