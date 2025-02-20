package temporal

import (
	"context"
	"crypto/tls"
	"fmt"
	"github.com/temporalio/temporal-jumpstart-golang/onboardings/config"
	"github.com/uber-go/tally/v4"
	sdkclient "go.temporal.io/sdk/client"
	sdktally "go.temporal.io/sdk/contrib/tally"
	"os"
)

func GetIdentity(taskQueue string) string {
	return fmt.Sprintf("%d@%s@%s", os.Getpid(), getHostName(), taskQueue)

}

func NewMetricsHandler(scope tally.Scope) sdkclient.MetricsHandler {
	return sdktally.NewMetricsHandler(scope)
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

// NewClient creates the temporal client
func NewClient(ctx context.Context, cfg *config.Config) (sdkclient.Client, error) {

	opts := sdkclient.Options{}
	var cert tls.Certificate

	// map
	if opts.HostPort == "" {
		opts.HostPort = cfg.Temporal.Connection.Target
	}
	if opts.Namespace == "" {
		opts.Namespace = cfg.Temporal.Connection.Namespace
	}
	if opts.Identity == "" {
		// same behavior as SDK
		opts.Identity = GetIdentity("")
	}

	if cfg.Temporal.Connection.MTLS != nil &&
		cfg.Temporal.Connection.MTLS.CertChainFile != "" &&
		cfg.Temporal.Connection.MTLS.KeyFile != "" {
		var err error

		cert, err = tls.LoadX509KeyPair(cfg.Temporal.Connection.MTLS.CertChainFile, cfg.Temporal.Connection.MTLS.KeyFile)
		if err != nil {
			return nil, fmt.Errorf("failed to load TLS from files: %w", err)
		}
	} else if cfg.Temporal.Connection.MTLS != nil &&
		len(cfg.Temporal.Connection.MTLS.CertChain) > 0 &&
		len(cfg.Temporal.Connection.MTLS.Key) > 0 {
		var err error
		cert, err = tls.X509KeyPair(cfg.Temporal.Connection.MTLS.CertChain, cfg.Temporal.Connection.MTLS.Key)
		if err != nil {
			return nil, fmt.Errorf("failed to load Cloud TLS from data: %w", err)
		}
	}

	if len(cert.Certificate) > 0 {
		opts.ConnectionOptions.TLS = &tls.Config{Certificates: []tls.Certificate{cert}}
	}

	result, err := sdkclient.Dial(opts)
	if err != nil {
		return nil, fmt.Errorf("failed to new temporal client %w", err)
	}
	return result, nil
}
