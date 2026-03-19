package main

import (
	"context"
	"swd-new/cmd/server/wire"
	"swd-new/pkg/config"
	"swd-new/pkg/log"

	"go.uber.org/zap"
)

func main() {
	conf := config.NewConfig()
	logger := log.NewLog(conf)

	migration, cleanup, err := wire.NewMigrationWire(conf, logger)
	if err != nil {
		panic(err)
	}
	defer cleanup()

	if err := migration.SyncSensitiveWordsFromAssets(context.Background()); err != nil {
		logger.Error("sync sensitive words from assets failed", zap.Error(err))
		panic(err)
	}

	logger.Info("sync sensitive words from assets completed")
}
