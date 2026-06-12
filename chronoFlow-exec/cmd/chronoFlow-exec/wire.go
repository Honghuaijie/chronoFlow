//go:build wireinject
// +build wireinject

package main

import (
	"chronoFlow-exec/internal/callback"
	"chronoFlow-exec/internal/conf"
	"chronoFlow-exec/internal/process"
	"chronoFlow-exec/internal/server"
	"chronoFlow-exec/internal/service"
	"chronoFlow-exec/internal/store"

	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
)

func wireApp(*conf.Server, *conf.Executor, *conf.Callback, log.Logger) (*kratos.App, func(), error) {
	panic(wire.Build(
		server.ProviderSet,
		process.ProviderSet,
		store.ProviderSet,
		callback.ProviderSet,
		callback.WorkerProviderSet,
		service.ProviderSet,
		newApp,
	))
}
