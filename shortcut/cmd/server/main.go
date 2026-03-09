package main

import (
	"context"
	"log"
	"os"

	"github.com/Gadzet005/shortcut/shortcut/internal/app"
	"github.com/Gadzet005/shortcut/shortcut/pkg/lifecycle"
	"github.com/Gadzet005/shortcut/shortcut/pkg/metrics"
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

	metricsService, err := metrics.NewService(2112, logger)
	if err != nil {
		logger.Fatal("failed to run metrics service", zap.Error(err))
	}

	if err := metricsService.Start(context.Background()); err != nil {
		logger.Fatal("failed to run metrics service", zap.Error(err))
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
	f, err := os.OpenFile("/var/log/shortcut/app.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
    if err != nil {
        log.Fatal(err)
    }

    defer f.Close()

	writers := []zapcore.WriteSyncer{zapcore.AddSync(os.Stdout)}
    
    if config.LogPath != "" {
        fileWriter := &lumberjack.Logger{
            Filename:   config.LogPath,
            MaxSize:    100,
            MaxBackups: 3,
            MaxAge:     28,
            Compress:   true,
        }
        writers = append(writers, zapcore.AddSync(fileWriter))
    }

	if config.Env.IsDev() {
		return zap.NewDevelopment()
	}

	return zap.NewProduction()
}
