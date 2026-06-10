package server

import (
	v1 "chronoFlow-exec/api/all-pb-go/v1"
	"chronoFlow-exec/internal/conf"
	"chronoFlow-exec/internal/service"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/transport/grpc"
)

func NewGRPCServer(c *conf.Server, userSvc *service.UserService, logger log.Logger) *grpc.Server {
	opts := []grpc.ServerOption{
		grpc.Middleware(
			requestLogMiddleware(logger),
			recovery.Recovery(),
		),
	}
	if c != nil && c.Grpc != nil {
		if c.Grpc.Network != "" {
			opts = append(opts, grpc.Network(c.Grpc.Network))
		}
		if c.Grpc.Addr != "" {
			opts = append(opts, grpc.Address(c.Grpc.Addr))
		}
		if c.Grpc.Timeout != nil {
			opts = append(opts, grpc.Timeout(c.Grpc.Timeout.AsDuration()))
		}
	}

	srv := grpc.NewServer(opts...)
	v1.RegisterUserServer(srv, userSvc)
	return srv
}
