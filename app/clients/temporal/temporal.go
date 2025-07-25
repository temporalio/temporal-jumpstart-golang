package temporal

import (
	"context"
	"crypto/tls"
	"fmt"
	"github.com/temporalio/temporal-jumpstart-golang/app/config"
	"github.com/temporalio/temporal-jumpstart-golang/app/instrumentation/prometheus"
	"github.com/uber-go/tally/v4"
	sdkclient "go.temporal.io/sdk/client"
	sdktally "go.temporal.io/sdk/contrib/tally"
	"log/slog"
	"os"
)

type Client struct {
	sdkclient.Client
	Options   sdkclient.Options
	rootScope tally.Scope
}

func GetIdentity(taskQueue string, index int) string {
	return fmt.Sprintf("%d@%s@%s-%d", os.Getpid(), getHostName(), taskQueue, index)
}

type Clients struct {
	Client        sdkclient.Client
	Config        *config.Config
	ClientOptions sdkclient.Options
}

func (c *Clients) Close() error {
	if c.Client != nil {
		c.Client.Close()
	}

	return nil
}

func getHostName() string {
	hostName, err := os.Hostname()
	if err != nil {
		hostName = "Unknown"
	}
	return hostName
}

type Headers struct {
	namespace string
	nsCfg     *config.NamespaceConfig
}

func (A Headers) GetHeaders(ctx context.Context) (map[string]string, error) {
	return map[string]string{
		"temporal-namespace": A.namespace,
	}, nil
}

// NewClient creates the temporal client
func NewClient(ctx context.Context,
	namespace string,
	cfg *config.TemporalConfig,
	index int,
	options ...Option) (*Client, error) {

	result := &Client{}
	opts := &sdkclient.Options{}
	for _, option := range options {
		option(result, opts)
	}
	nsCfg, exists := cfg.Namespaces[namespace]
	if !exists {
		return nil, fmt.Errorf("namespace %s config not found", namespace)
	}
	// map
	if opts.HostPort == "" {
		opts.HostPort = nsCfg.Connection.Target
	}
	if opts.Namespace == "" {
		opts.Namespace = namespace
	}
	if opts.Identity == "" {
		// same behavior as SDK
		opts.Identity = GetIdentity("", index)
	}
	if nsCfg.Connection.MTLS == nil && nsCfg.Connection.APIKey != "" {
		// api key connection setup
		opts.ConnectionOptions.TLS = &tls.Config{}
		if nsCfg.Connection.APIKey != "" {
			opts.Credentials = sdkclient.NewAPIKeyStaticCredentials(nsCfg.Connection.APIKey)
			opts.HeadersProvider = &Headers{namespace: namespace, nsCfg: nsCfg}
		}
	}
	if result.rootScope != nil {
		// if a rootScope was passed in as an option, tag this client so N clients can report metrics
		scope, err := prometheus.NewScope(ctx, result.rootScope, prometheus.WithTags(
			map[string]string{
				// used for disambiguating metrics
				"client_id": fmt.Sprintf("wf-client-%d", index),
			}))
		if err != nil {
			return nil, fmt.Errorf("failed to create prometheus scope: %w", err)
		}
		opts.MetricsHandler = sdktally.NewMetricsHandler(scope)
	}

	slog.Info("creating temporal client",
		"hostport", opts.HostPort,
		"namespace", opts.Namespace,
		"with api-key", opts.Credentials != nil,
		"identity", opts.Identity)

	inner, err := sdkclient.Dial(*opts)
	if err != nil {
		return nil, fmt.Errorf("failed to new temporal client %w", err)
	}
	result.Client = inner
	result.Options = *opts
	slog.Info("Connected to Temporal server", "address", opts.HostPort)
	return result, nil
}
