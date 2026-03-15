package config

import (
	"os"
	"path"
	"path/filepath"

	"dario.cat/mergo"
	"github.com/Gadzet005/shortcut/pkg/app/env"
	"github.com/Gadzet005/shortcut/pkg/errors"
	"gopkg.in/yaml.v3"
)

const (
	defaultConfigDir = "./configs"
	configDirEnvVar  = "CONFIGS_DIR"
	baseConfigName   = "base.yaml"
)

func LoadServiceConfig[T any](serviceName string, env env.Env) (T, error) {
	configsDir := os.Getenv(configDirEnvVar)
	if configsDir == "" {
		configsDir = path.Join(defaultConfigDir, serviceName)
	}
	configPath := filepath.Join(configsDir, env.String()+".yaml")
	baseConfigPath := filepath.Join(configsDir, baseConfigName)

	baseConfig, err := Load[T](baseConfigPath)
	if err != nil {
		return baseConfig, errors.Wrap(err, "load base config")
	}
	overrideConfig, err := Load[T](configPath)
	if err != nil {
		return overrideConfig, errors.Wrap(err, "load override config")
	}
	if err := mergo.Merge(&baseConfig, overrideConfig, mergo.WithOverride); err != nil {
		return baseConfig, errors.Wrap(err, "merge configs")
	}
	return baseConfig, nil
}

func Load[T any](path string) (T, error) {
	var config T
	yamlFile, err := os.ReadFile(path)
	if err != nil {
		return config, errors.Wrap(err, "read config file")
	}
	if err := yaml.Unmarshal(yamlFile, &config); err != nil {
		return config, errors.Wrap(err, "unmarshal config file")
	}
	return config, nil
}
