package config

import (
	"errors"
	"os"
)

type Secret struct {
	Env string `yaml:"env"`
}

func (s Secret) Load() (string, error) {
	if s.Env != "" {
		return s.loadEnv(s.Env)
	}
	return "", errors.New("secret config is not set")
}

func (s Secret) loadEnv(env string) (string, error) {
	envVar := os.Getenv(env)
	if envVar == "" {
		return "", errors.New("secret is not set")
	}
	return envVar, nil
}
