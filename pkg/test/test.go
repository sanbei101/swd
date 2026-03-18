package test

import (
	"os"
	"path/filepath"

	"swd-new/internal/repository"
	"swd-new/pkg/log"

	"github.com/spf13/viper"
)

type TestEnvironment struct {
	TestDB     *repository.Repository
	TestLogger *log.Logger
	TestConfig *viper.Viper
}

func SetupTestEnvironment() (*TestEnvironment, error) {
	conf, err := loadLocalConfig()
	if err != nil {
		return nil, err
	}

	logger := log.NewLog(conf)
	repo := repository.NewRepository(logger, conf)

	return &TestEnvironment{
		TestDB:     repo,
		TestLogger: logger,
		TestConfig: conf,
	}, nil
}

func loadLocalConfig() (*viper.Viper, error) {
	conf := viper.New()
	conf.SetConfigFile(filepath.Join(projectRoot(), "config", "local.yml"))
	if err := conf.ReadInConfig(); err != nil {
		return nil, err
	}
	return conf, nil
}

func projectRoot() string {
	dir, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			panic("project root not found")
		}
		dir = parent
	}
}
