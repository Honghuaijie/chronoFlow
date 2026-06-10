//go:build wireinject
// +build wireinject

package main

import (
	"chronoFlow-admin/internal/biz"
	"chronoFlow-admin/internal/conf"
	"chronoFlow-admin/internal/data"
	"chronoFlow-admin/internal/server"
	"chronoFlow-admin/internal/service"

	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
)

func wireApp(*conf.Server, *conf.Data, log.Logger) (*kratos.App, func(), error) {
	panic(wire.Build(
		server.ProviderSet,
		data.ProviderSet,
		biz.ProviderSet,
		service.ProviderSet,
		newApp,
	))
}
