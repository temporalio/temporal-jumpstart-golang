package connect

import (
	"google.golang.org/grpc"
	"net"
)

type registerable interface {
	Register(srv *grpc.Server)
}
type GRPCServerOption func(server *DefaultGRPCServer)

func WithConfig(cfg *Config) GRPCServerOption {
	return func(s *DefaultGRPCServer) {
		s.cfg = cfg
	}
}
func WithServer(srv *grpc.Server) GRPCServerOption {
	return func(s *DefaultGRPCServer) {
		s.server = srv
	}
}
func WithListener(lis net.Listener) GRPCServerOption {
	return func(s *DefaultGRPCServer) {
		s.listener = lis
	}
}
func WithServices(r ...registerable) GRPCServerOption {
	return func(s *DefaultGRPCServer) {
		s.services = r
	}
}
