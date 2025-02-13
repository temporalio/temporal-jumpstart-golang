package connect

import (
	"connectrpc.com/connect"
	"context"
	"fmt"
	"github.com/temporalio/temporal-shop/services/go/pkg/instrumentation/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func NewClientConnection(ctx context.Context, cfg *Config) (*connect.Client[].ClientConn, error) {
	logger := log.GetLogger(ctx)
	defaultDialOpts := []grpc.DialOption{
		//grpc.WithBlock(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}
	addr := fmt.Sprintf(":%d", cfg.Port)
	dialer, err := grpc.DialContext(ctx, addr, defaultDialOpts...)
	dialer.Connect()
	logger.Info("grpc_client_conn created", map[string]interface{}{"grpc_client_conn_state": dialer.GetState().String()})
	return dialer, err
}
