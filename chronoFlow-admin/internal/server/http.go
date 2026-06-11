package server

import (
	v1 "chronoFlow-admin/api/all-pb-go/v1"
	"chronoFlow-admin/internal/conf"
	"chronoFlow-admin/internal/handler"
	"chronoFlow-admin/internal/service"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/middleware/validate"
	"github.com/go-kratos/kratos/v2/transport/http"
	"github.com/gorilla/handlers"
)

func NewHTTPServer(
	c *conf.Server,
	security *conf.Security,
	authSvc *service.AuthService,
	userSvc *service.UserService,
	executorSvc *service.ExecutorService,
	jobSvc *service.JobService,
	glueSvc *service.GlueService,
	jobLogSvc *service.JobLogService,
	callbackSvc *service.CallbackService,
	logger log.Logger,
) *http.Server {
	opts := []http.ServerOption{
		// 给kratos的http服务注册一个“统一错误处理函数”--只要业务层返回err，都会调用这个方法
		http.ErrorEncoder(errorEncoder),
		http.Middleware(
			requestLogMiddleware(logger),
			adminAuthMiddleware(security),
			recovery.Recovery(),
			validate.Validator(),
		),
		http.Filter(
			handlers.CORS(
				handlers.AllowedHeaders([]string{"X-Requested-With", "Content-Type", "Authorization"}),
				handlers.AllowedMethods([]string{"GET", "POST", "HEAD", "OPTIONS"}),
				handlers.AllowedOrigins([]string{"*"}),
			),
		),
	}
	if c != nil && c.Http != nil {
		if c.Http.Network != "" {
			opts = append(opts, http.Network(c.Http.Network))
		}
		if c.Http.Addr != "" {
			opts = append(opts, http.Address(c.Http.Addr))
		}
		if c.Http.Timeout != nil {
			opts = append(opts, http.Timeout(c.Http.Timeout.AsDuration()))
		}
	}

	srv := http.NewServer(opts...)
	v1.RegisterAuthHTTPServer(srv, authSvc)
	v1.RegisterUserHTTPServer(srv, userSvc)
	v1.RegisterExecutorHTTPServer(srv, executorSvc)
	v1.RegisterJobHTTPServer(srv, jobSvc)
	v1.RegisterGlueHTTPServer(srv, glueSvc)
	v1.RegisterJobLogHTTPServer(srv, jobLogSvc)
	v1.RegisterJobRunCallbackHTTPServer(srv, callbackSvc)
	srv.Route("").GET("/health", handler.HealthCheckHandler)
	srv.Route("").POST("/v1/users/avatarUpload", userSvc.AvatarUpload)
	log.NewHelper(logger).Info("http routes registered")
	return srv
}
