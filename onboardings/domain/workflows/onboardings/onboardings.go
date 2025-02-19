package onboardings

import (
	"go.temporal.io/sdk/temporal"
)

// TryHandleCancellation
func TryHandleCancellation(err error) error {
	if temporal.IsCanceledError(err) {
		return err
	}
	return nil
}
