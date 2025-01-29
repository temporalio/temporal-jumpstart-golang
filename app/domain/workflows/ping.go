package workflows

import (
	"fmt"
	"go.temporal.io/sdk/workflow"
)

const QueryPing = "ping"

func Ping(ctx workflow.Context, args string) (string, error) {
	state := fmt.Sprintf("pong: %s", args)
	err := workflow.SetQueryHandler(ctx, QueryPing, func() (string, error) {
		return state, nil
	})
	if err != nil {
		return "", err
	}

	return state, nil
}
