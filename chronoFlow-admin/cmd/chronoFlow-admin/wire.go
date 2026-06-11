//go:build wireinject
// +build wireinject

package main

import (
	"chronoFlow-admin/internal/biz"
	"chronoFlow-admin/internal/conf"
	"chronoFlow-admin/internal/data"
	"chronoFlow-admin/internal/executorclient"
	"chronoFlow-admin/internal/logstore"
	"chronoFlow-admin/internal/scheduler"
	"chronoFlow-admin/internal/security"
	"chronoFlow-admin/internal/server"
	"chronoFlow-admin/internal/service"
	"chronoFlow-admin/internal/worker"

	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
)

func wireApp(*conf.Server, *conf.Data, *conf.Security, *conf.Logs, *conf.Executor, *conf.Recovery, *conf.Scheduler, log.Logger) (*kratos.App, func(), error) {
	panic(wire.Build(
		server.ProviderSet,
		data.ProviderSet,
		security.ProviderSet,
		logstore.ProviderSet,
		executorclient.ProviderSet,
		scheduler.ProviderSet,
		biz.ProviderSet,
		service.ProviderSet,
		worker.ProviderSet,
		newApp,
	))
}
