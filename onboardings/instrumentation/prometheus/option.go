package prometheus

import (
	"github.com/uber-go/tally/v4"
	"github.com/uber-go/tally/v4/prometheus"
)

type ReporterOption func(*prometheus.Configuration)

func WithListenAddress(addr string) ReporterOption {
	return func(c *prometheus.Configuration) {
		c.ListenAddress = addr
	}
}

type RootOption func(*tally.ScopeOptions)

func WithRootTags(tags map[string]string) RootOption {
	return func(s *tally.ScopeOptions) {
		s.Tags = tags
	}
}
func WithReporter(reporter prometheus.Reporter) RootOption {
	return func(s *tally.ScopeOptions) {
		s.CachedReporter = reporter
	}
}

type ScopeOption func(tally.Scope) tally.Scope

func WithTags(tags map[string]string) ScopeOption {
	return func(s tally.Scope) tally.Scope {
		return s.Tagged(tags)
	}
}
