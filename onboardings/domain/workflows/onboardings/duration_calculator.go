package onboardings

import (
	"math"
	"time"
)

// OnboardEntityDurationCalculator is message version agnostic and calculations time spans and enforces product requirements
// for things like timeouts
type OnboardEntityDurationCalculator struct {
	CompletionTimeoutSeconds uint64
	SkipApproval             bool
	Timestamp                time.Time
	HasDeputyOwner           bool
}

// ResolveCompletionTimeoutSeconds calculates a valid completion timeout duration
func (c *OnboardEntityDurationCalculator) ResolveCompletionTimeoutSeconds(now time.Time) uint64 {
	if c.SkipApproval {
		return 0
	}
	completionTimeoutSeconds := c.CompletionTimeoutSeconds
	if completionTimeoutSeconds == 0 {
		completionTimeoutSeconds = DefaultCompletionTimeoutSeconds
	}
	return completionTimeoutSeconds
}

// CalculateWaitSeconds calcs the time remaining before the Workflow shall make progress
// taking into consideration the DeputyOwner argument and the Product requirement that we shall
// give 40% of the time before reaching out to the deputy to ask for approval.
func (c *OnboardEntityDurationCalculator) CalculateWaitSeconds(now time.Time) uint64 {
	completionTimeoutSeconds := c.ResolveCompletionTimeoutSeconds(now)
	if completionTimeoutSeconds == 0 {
		return completionTimeoutSeconds
	}
	waitSeconds := completionTimeoutSeconds
	if c.HasDeputyOwner {
		// if a deputy owner is provided, wait for 40% of the time (per the spec) to approve before next steps
		waitSeconds = uint64(.4 * float64(waitSeconds))
	}
	threshold := c.Timestamp.Add(time.Second * time.Duration(waitSeconds))

	return uint64(math.Max(threshold.Sub(now).Seconds(), 0))
}

func (c *OnboardEntityDurationCalculator) CalculateCompletionThreshold(now time.Time) time.Time {
	return c.Timestamp.Add(time.Second * time.Duration(c.ResolveCompletionTimeoutSeconds(now)))
}
