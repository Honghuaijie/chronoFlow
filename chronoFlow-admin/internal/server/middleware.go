package server

import (
	"context"
	"strings"
	"time"

	"chronoFlow-admin/internal/auth"
	"chronoFlow-admin/internal/conf"
	httpErrors "chronoFlow-admin/internal/errors"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/transport"
	httpCtx "github.com/go-kratos/kratos/v2/transport/http"
)

func requestLogMiddleware(logger log.Logger) middleware.Middleware {
	helper := log.NewHelper(logger)
	return func(next middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (reply interface{}, err error) {
			start := time.Now()
			reply, err = next(ctx, req)
			latency := time.Since(start)

			transportKind := ""
			operation := ""
			if tr, ok := transport.FromServerContext(ctx); ok {
				transportKind = tr.Kind().String()
				operation = tr.Operation()
			}

			fields := []interface{}{
				"component", "server_middleware",
				"transport", transportKind,
				"operation", operation,
				"latency_ms", latency.Milliseconds(),
			}

			if err != nil {
				se := httpErrors.FromError(err)
				fields = append(fields,
					"code", se.Code,
					"http_code", se.HttpCode,
					"err", err.Error(),
				)
				helper.Errorw(fields...)
				return reply, err
			}

			fields = append(fields, "code", 0)
			helper.Infow(fields...)
			return reply, nil
		}
	}
}

func adminAuthMiddleware(security *conf.Security) middleware.Middleware {
	return func(next middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			httpReq, ok := httpCtx.RequestFromServerContext(ctx)
			if !ok || httpReq == nil || !strings.HasPrefix(httpReq.URL.Path, "/v1/admin/") {
				return next(ctx, req)
			}
			header := strings.TrimSpace(httpReq.Header.Get("Authorization"))
			if !strings.HasPrefix(header, "Bearer ") {
				return nil, httpErrors.E(httpErrors.ErrMissingToken)
			}
			token := strings.TrimSpace(strings.TrimPrefix(header, "Bearer "))
			if token == "" {
				return nil, httpErrors.E(httpErrors.ErrMissingToken)
			}
			if security == nil {
				return nil, httpErrors.E(httpErrors.ErrConfigInvalid)
			}
			if _, err := auth.ParseJWT(security.JwtSecret, token); err != nil {
				if err == auth.ErrExpiredToken {
					return nil, httpErrors.E(httpErrors.ErrExpiredToken)
				}
				return nil, httpErrors.E(httpErrors.ErrInvalidToken)
			}
			return next(ctx, req)
		}
	}
}
