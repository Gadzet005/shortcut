package main

import (
	"context"
	"path/filepath"
	"fmt"
	"log"
	"os"

	"github.com/Gadzet005/shortcut/shortcut/internal/app"
	graphconfig "github.com/Gadzet005/shortcut/shortcut/internal/domain/graph/config"
	"github.com/Gadzet005/shortcut/shortcut/pkg/lifecycle"
	"github.com/Gadzet005/shortcut/shortcut/pkg/metrics"
	configutils "github.com/Gadzet005/shortcut/shortcut/pkg/utils/config"
	"github.com/spf13/pflag"
	"gopkg.in/natefinch/lumberjack.v2"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func main() {
	configPaths := pflag.StringSliceP(
		"configs",
		"c",
		[]string{"./tests/e2e/configs/base.yaml", "./tests/e2e/configs/prod.yaml"},
		"paths to config files (comma-separated)",
	)

	graphConfigPaths := pflag.StringSliceP(
		"graphconfigs",
		"g",
		[]string{"./tests/e2e/configs/graph.yaml"},
		"paths to graph config files (comma-separated)",
	)
	pflag.Parse()

	config, err := configutils.LoadConfig[app.Config](*configPaths)
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	graphConfig, err := configutils.LoadConfig[graphconfig.Config](*graphConfigPaths)
	if err != nil {
		log.Fatalf("failed to load graph config: %v", err)
	}

	logger, err := newLogger(config)
	if err != nil {
		log.Fatalf("failed to create logger: %v", err)
	}

	metricsService, err := metrics.NewService(config.MetricsPort, logger)
	if err != nil {
		logger.Fatal("failed to run metrics service", zap.Error(err))
	}

	if err := metricsService.Start(context.Background()); err != nil {
		logger.Fatal("failed to run metrics service", zap.Error(err))
	}

	service, err := app.NewService(config, graphConfig, logger)
	if err != nil {
		logger.Fatal("failed to create service", zap.Error(err))
	}

	if err := lifecycle.Run(service); err != nil {
		logger.Fatal("failed to run service", zap.Error(err))
	}
}

func newLogger(config app.Config) (*zap.Logger, error) {
    var writers []zapcore.WriteSyncer
    writers = append(writers, zapcore.AddSync(os.Stdout))
    
    if config.LogPath != "" {
        logDir := filepath.Dir(config.LogPath)
        if err := os.MkdirAll(logDir, 0755); err != nil {
            return nil, fmt.Errorf("failed to create log directory: %w", err)
        }
        
        fileWriter := &lumberjack.Logger{
            Filename:   config.LogPath,
            MaxSize:    100,
            MaxBackups: 3,
            MaxAge:     28,
            Compress:   true,
            LocalTime:  true,
        }
        writers = append(writers, zapcore.AddSync(fileWriter))
    }

    core := zapcore.NewCore(
        zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()),
        zapcore.NewMultiWriteSyncer(writers...),
        zap.InfoLevel,
    )
    
    if config.Env.IsDev() {
        developmentConfig := zap.NewDevelopmentConfig()
        developmentConfig.OutputPaths = []string{"stdout"}
        if config.LogPath != "" {
            developmentConfig.OutputPaths = append(developmentConfig.OutputPaths, config.LogPath)
        }
        return developmentConfig.Build()
    }
    
    logger := zap.New(core, zap.AddCaller(), zap.AddStacktrace(zap.ErrorLevel))
    
    return logger, nil
}
