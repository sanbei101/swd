package repository

import (
	"swd-new/pkg/log"

	"github.com/spf13/viper"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Repository struct {
	db     *gorm.DB
	logger *log.Logger
}

func NewRepository(logger *log.Logger, conf *viper.Viper) *Repository {
	db, err := gorm.Open(postgres.Open(conf.GetString("data.postgres.dsn")), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	return &Repository{
		db:     db,
		logger: logger,
	}
}
