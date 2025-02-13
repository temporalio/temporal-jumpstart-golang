package connect

import (
	"context"
	"fmt"
	"github.com/temporalio/temporal-shop/services/go/pkg/instrumentation/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"net"
)

func NewDefaultGRPCServer(ctx context.Context, opts ...GRPCServerOption) (*DefaultGRPCServer, error) {

	defaultOpts := []GRPCServerOption{
		WithServer(grpc.NewServer()),
	}
	opts = append(defaultOpts, opts...)
	out := &DefaultGRPCServer{}
	for _, o := range opts {
		o(out)
	}
	if out.listener == nil {
		lis, err := net.Listen("tcp", fmt.Sprintf(":%d", out.cfg.Port))
		if err != nil {
			return nil, fmt.Errorf("failed to listen: %v", err)
		}
		out.listener = lis
	}
	return out, nil
}

type DefaultGRPCServer struct {
	server   *grpc.Server
	listener net.Listener
	services []registerable
	cfg      *Config
}

func (s *DefaultGRPCServer) Start(ctx context.Context) error {
	for _, r := range s.services {
		r.Register(s.server)
	}
	logger := log.WithFields(
		log.GetLogger(ctx),
		log.Fields{"addr": s.listener.Addr()},
	)
	reflection.Register(s.server)
	logger.Info("starting grpc server")
	return s.server.Serve(s.listener)
}
func (s *DefaultGRPCServer) Shutdown(ctx context.Context) {
	logger := log.WithFields(
		log.GetLogger(ctx),
		log.Fields{"addr": s.listener.Addr()},
	)
	logger.Info("shutting down")
	s.server.GracefulStop()
}
