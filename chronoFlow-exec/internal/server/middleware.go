package server

import (
	"context"
	"crypto/subtle"
	"strings"
	"time"

	"chronoFlow-exec/internal/conf"
	httpErrors "chronoFlow-exec/internal/errors"
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

func executorTokenMiddleware(c *conf.Executor) middleware.Middleware {
	return func(next middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			httpReq, ok := httpCtx.RequestFromServerContext(ctx)
			if !ok || httpReq == nil {
				return next(ctx, req)
			}
			path := httpReq.URL.Path
			if path != "/health" && path != "/run" && path != "/kill" {
				return next(ctx, req)
			}
			expected := ""
			if c != nil {
				expected = c.Token
			}
			got := strings.TrimSpace(httpReq.Header.Get("X-Executor-Token"))
			if expected == "" || got == "" {
				return nil, httpErrors.E(httpErrors.ErrMissingToken)
			}
			if subtle.ConstantTimeCompare([]byte(got), []byte(expected)) != 1 {
				return nil, httpErrors.E(httpErrors.ErrInvalidToken)
			}
			return next(ctx, req)
		}
	}
}
