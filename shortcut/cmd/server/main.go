package main

import (
	"log"

	"github.com/Gadzet005/shortcut/shortcut/internal/app"
	"github.com/Gadzet005/shortcut/shortcut/pkg/lifecycle"
	configutils "github.com/Gadzet005/shortcut/shortcut/pkg/utils/config"
	"github.com/spf13/pflag"
	"go.uber.org/zap"
)

func main() {
	configPaths := pflag.StringSliceP(
		"configs",
		"c",
		[]string{"./shortcut/configs/base.yaml", "./shortcut/configs/dev.yaml"},
		"paths to config files (comma-separated)",
	)
	pflag.Parse()

	config, err := configutils.LoadConfig[app.Config](*configPaths)
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	logger, err := newLogger(config)
	if err != nil {
		log.Fatalf("failed to create logger: %v", err)
	}

	service, err := app.NewService(config, logger)
	if err != nil {
		logger.Fatal("failed to create service", zap.Error(err))
	}

	if err := lifecycle.Run(service); err != nil {
		logger.Fatal("failed to run service", zap.Error(err))
	}
}

func newLogger(config app.Config) (*zap.Logger, error) {
	if config.Env.IsDev() {
		return zap.NewDevelopment()
	}
	return zap.NewProduction()
}
