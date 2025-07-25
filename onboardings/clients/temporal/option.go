package temporal

import (
	"github.com/uber-go/tally/v4"
	sdkclient "go.temporal.io/sdk/client"
	sdktally "go.temporal.io/sdk/contrib/tally"
)

type Option func(*Client, *sdkclient.Options)

func WithRootScope(scope tally.Scope) Option {
	return func(c *Client, _ *sdkclient.Options) {
		c.rootScope = scope
	}
}
func WithMetricsScope(scope tally.Scope) Option {
	return func(c *Client, opts *sdkclient.Options) {
		opts.MetricsHandler = sdktally.NewMetricsHandler(scope)
	}
}
