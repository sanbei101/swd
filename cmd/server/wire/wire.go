//go:build wireinject
// +build wireinject

package wire

import (
	"swd-new/internal/handler"
	"swd-new/internal/repository"
	"swd-new/internal/server"
	"swd-new/internal/service"
	"swd-new/pkg/log"

	"github.com/gin-gonic/gin"
	"github.com/google/wire"
	"github.com/spf13/viper"
)

var ServerSet = wire.NewSet(server.NewServerHTTP)

var RepositorySet = wire.NewSet(
	repository.NewRepository,
	repository.NewSensitiveWordRepository,
)

var ServiceSet = wire.NewSet(
	service.NewService,
	service.NewSensitiveWordService,
)

var HandlerSet = wire.NewSet(
	handler.NewHandler,
	handler.NewSensitiveCheckHandler,
)

func NewWire(*viper.Viper, *log.Logger) (*gin.Engine, func(), error) {
	panic(wire.Build(
		ServerSet,
		RepositorySet,
		ServiceSet,
		HandlerSet,
	))
}
