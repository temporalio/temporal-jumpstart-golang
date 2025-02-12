package onboardings

import (
	v1 "github.com/temporalio/temporal-jumpstart-golang/onboardings/generated/domain/v1"
	"go.temporal.io/sdk/temporal"
	"math"
	"strings"
	"time"
)

// assertValidArgs is a poor-man's implementation not providing very specific details on which args are "invalid"
// but instead just failing hard in the event of missing or bad args
func assertValidArgs(args *v1.OnboardEntityRequest) error {
	if strings.TrimSpace(args.Id) != "" &&
		strings.TrimSpace(args.Value) != "" &&
		!args.Timestamp.AsTime().IsZero() {
		return nil
	}
	return temporal.NewApplicationError(v1.Errors_ERRORS_ONBOARD_ENTITY_INVALID_ARGS.String(), v1.Errors_ERRORS_ONBOARD_ENTITY_INVALID_ARGS.String())
}

// resolveCompletionTimeoutSeconds calculates a valid completion timeout duration
func resolveCompletionTimeoutSeconds(args *v1.OnboardEntityRequest) uint64 {
	if args.SkipApproval {
		return 0
	}
	completionTimeoutSeconds := args.CompletionTimeoutSeconds
	if completionTimeoutSeconds == 0 {
		completionTimeoutSeconds = DefaultCompletionTimeoutSeconds
	}
	return completionTimeoutSeconds
}

// calculateWaitSeconds calcs the time remaining before the Workflow shall make progress
// taking into consideration the DeputyOwner argument and the Product requirement that we shall
// give 40% of the time before reaching out to the deputy to ask for approval.
func calculateWaitSeconds(now time.Time, args *v1.OnboardEntityRequest) uint64 {
	completionTimeoutSeconds := resolveCompletionTimeoutSeconds(args)
	if completionTimeoutSeconds == 0 {
		return completionTimeoutSeconds
	}
	waitSeconds := completionTimeoutSeconds
	if len(strings.TrimSpace(args.DeputyOwnerEmail)) != 0 {
		// if a deputy owner is provided, wait for 40% of the time (per the spec) to approve before next steps
		waitSeconds = uint64(.4 * float64(waitSeconds))
	}
	threshold := args.Timestamp.AsTime().Add(time.Second * time.Duration(waitSeconds))

	return uint64(math.Max(threshold.Sub(now).Seconds(), 0))
}

func calculateCompletionThreshold(args *v1.OnboardEntityRequest) time.Time {
	return args.Timestamp.AsTime().Add(time.Second * time.Duration(resolveCompletionTimeoutSeconds(args)))
}
func calculateElapsedTimeSinceStarted(now time.Time, args *v1.OnboardEntityRequest) uint64 {
	return uint64(now.Sub(args.Timestamp.AsTime()).Seconds())
}

// tryHandleCancellation
func tryHandleCancellation(err error) error {
	if temporal.IsCanceledError(err) {
		return err
	}
	return nil
}
