package temporal

import (
	"github.com/uber-go/tally/v4"
	sdkclient "go.temporal.io/sdk/client"
	sdktally "go.temporal.io/sdk/contrib/tally"
)

type Option func(*sdkclient.Options)

func WithMetricsScope(scope tally.Scope) Option {
	return func(opts *sdkclient.Options) {
		opts.MetricsHandler = sdktally.NewMetricsHandler(scope)
	}
}
