//go:build wireinject
// +build wireinject

// The build tag makes sure the stub is not built in the final build.

package main

import (
	"log/slog"

	"github.com/go-kratos/kratos/v3"
	"github.com/google/wire"
	"jieyuc.cn/muxu/studio/internal/biz"
	"jieyuc.cn/muxu/studio/internal/conf"
	"jieyuc.cn/muxu/studio/internal/data"
	"jieyuc.cn/muxu/studio/internal/server"
	"jieyuc.cn/muxu/studio/internal/service"
)

// wireApp init kratos application.
func wireApp(*conf.Server, *conf.Data, *slog.Logger) (*kratos.App, func(), error) {
	panic(wire.Build(server.ProviderSet, data.ProviderSet, biz.ProviderSet, service.ProviderSet, newApp))
}
