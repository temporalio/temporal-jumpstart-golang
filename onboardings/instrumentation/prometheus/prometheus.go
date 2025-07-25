package prometheus

import (
	"context"
	"fmt"
	prom "github.com/prometheus/client_golang/prometheus"
	"github.com/uber-go/tally/v4"
	"github.com/uber-go/tally/v4/prometheus"
	sdktally "go.temporal.io/sdk/contrib/tally"
	"log/slog"
	"time"
)

func NewScope(_ context.Context, root tally.Scope, opts ...ScopeOption) (tally.Scope, error) {
	scope := sdktally.NewPrometheusNamingScope(root)
	for _, opt := range opts {
		scope = opt(scope)
	}
	return scope, nil
}
func NewReporter(ctx context.Context, opts ...ReporterOption) (prometheus.Reporter, error) {
	c := &prometheus.Configuration{}
	for _, opt := range opts {
		opt(c)
	}
	c.TimerType = "histogram"

	if c.ListenAddress == "" {
		c.ListenAddress = "0.0.0.0:9090"
		slog.Info("prometheus metrics listener address not set, using default", "address", c.ListenAddress)
	}

	reporter, err := c.NewReporter(
		prometheus.ConfigurationOptions{
			Registry: prom.NewRegistry(),
			OnError: func(err error) {
				slog.Error("error in prometheus reporter", "error", err)
			},
		},
	)
	if err != nil {
		return nil, fmt.Errorf("error creating prometheus reporter %w", err)
	}
	return reporter, nil
}
func NewRootScope(ctx context.Context, opts ...RootOption) (tally.Scope, error) {

	c := &tally.ScopeOptions{}
	for _, opt := range opts {
		opt(c)
	}
	if c.CachedReporter == nil {
		var rerr error
		c.CachedReporter, rerr = NewReporter(ctx)
		if rerr != nil {
			return nil, fmt.Errorf("error creating default prometheus reporter %w", rerr)
		}
	}
	c.Separator = prometheus.DefaultSeparator
	c.SanitizeOptions = &sdktally.PrometheusSanitizeOptions
	scope, _ := tally.NewRootScope(*c, time.Second)

	return scope, nil
}
