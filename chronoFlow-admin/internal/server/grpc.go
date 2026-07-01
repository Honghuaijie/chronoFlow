package server

import (
	v1 "chronoFlow-admin/api/all-pb-go/v1"
	"chronoFlow-admin/internal/conf"
	"chronoFlow-admin/internal/service"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/transport/grpc"
)

func NewGRPCServer(
	c *conf.Server,
	authSvc *service.AuthService,
	userSvc *service.UserService,
	executorSvc *service.ExecutorService,
	jobSvc *service.JobService,
	glueSvc *service.GlueService,
	jobLogSvc *service.JobLogService,
	callbackSvc *service.CallbackService,
	systemSettingsSvc *service.SystemSettingsService,
	logger log.Logger,
) *grpc.Server {
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
	v1.RegisterAuthServer(srv, authSvc)
	v1.RegisterUserServer(srv, userSvc)
	v1.RegisterExecutorServer(srv, executorSvc)
	v1.RegisterJobServer(srv, jobSvc)
	v1.RegisterGlueServer(srv, glueSvc)
	v1.RegisterJobLogServer(srv, jobLogSvc)
	v1.RegisterJobRunCallbackServer(srv, callbackSvc)
	v1.RegisterSystemSettingsServer(srv, systemSettingsSvc)
	return srv
}
