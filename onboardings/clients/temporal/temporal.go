package temporal

import (
	"context"
	"crypto/tls"
	"fmt"
	"github.com/temporalio/temporal-jumpstart-golang/app/config"
	sdkclient "go.temporal.io/sdk/client"
	"log/slog"
	"os"
)

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
	cfg *config.Config
}

func (A Headers) GetHeaders(ctx context.Context) (map[string]string, error) {
	return map[string]string{
		"temporal-namespace": A.cfg.Temporal.Connection.Namespace,
	}, nil
}

// NewClient creates the temporal client
func NewClient(ctx context.Context, cfg *config.Config, index int, options ...Option) (sdkclient.Client, error) {

	opts := &sdkclient.Options{}
	for _, option := range options {
		option(opts)
	}
	// map
	if opts.HostPort == "" {
		opts.HostPort = cfg.Temporal.Connection.Target
	}
	if opts.Namespace == "" {
		opts.Namespace = cfg.Temporal.Connection.Namespace
	}
	if opts.Identity == "" {
		// same behavior as SDK
		opts.Identity = GetIdentity("", index)
	}
	if cfg.Temporal.Connection.MTLS == nil && cfg.Temporal.Connection.APIKey != "" {
		// api key connection setup
		opts.ConnectionOptions.TLS = &tls.Config{}
		if cfg.Temporal.Connection.APIKey != "" {
			opts.Credentials = sdkclient.NewAPIKeyStaticCredentials(cfg.Temporal.Connection.APIKey)
			opts.HeadersProvider = &Headers{cfg: cfg}
		}
	}

	slog.Info("creating temporal client",
		"hostport", opts.HostPort,
		"namespace", opts.Namespace,
		"with api-key", opts.Credentials != nil,
		"identity", opts.Identity)

	result, err := sdkclient.Dial(*opts)
	if err != nil {
		return nil, fmt.Errorf("failed to new temporal client %w", err)
	}

	slog.Info("Connected to Temporal server", "address", opts.HostPort)
	return result, nil
}
