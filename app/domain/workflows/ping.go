package workflows

import (
	"fmt"
	"go.temporal.io/sdk/log"
	"go.temporal.io/sdk/workflow"
)

const QueryPing = "ping"

func Ping(ctx workflow.Context, args string) (string, error) {

	logger := log.With(workflow.GetLogger(ctx), "arg1", args)
	logger.Info("Ping")
	state := fmt.Sprintf("pong: %s", args)
	err := workflow.SetQueryHandler(ctx, QueryPing, func() (string, error) {
		return state, nil
	})
	if err != nil {
		return "", err
	}

	return state, nil
}
