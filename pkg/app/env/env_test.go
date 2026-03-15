package env

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseEnv(t *testing.T) {
	tests := []struct {
		name string
		env  string
		want Env
	}{
		{"dev", "dev", EnvDev},
		{"testing", "testing", EnvTesting},
		{"prod", "prod", EnvProd},
		{"unknown empty", "", EnvUnknown},
		{"unknown value", "staging", EnvUnknown},
		{"case sensitive", "DEV", EnvUnknown},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseEnv(tt.env)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestEnv_String(t *testing.T) {
	tests := []struct {
		env  Env
		want string
	}{
		{EnvDev, "dev"},
		{EnvProd, "prod"},
		{EnvUnknown, "unknown"},
	}
	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.env.String())
		})
	}
}

func TestLoadFromEnvVar(t *testing.T) {
	const key = "ENV"
	old := os.Getenv(key)
	defer func() { _ = os.Setenv(key, old) }()

	tests := []struct {
		name   string
		setEnv string
		want   Env
	}{
		{"unset -> dev", "", EnvDev},
		{"prod", "prod", EnvProd},
		{"dev", "dev", EnvDev},
		{"testing", "testing", EnvTesting},
		{"invalid", "invalid", EnvUnknown},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setEnv == "" {
				_ = os.Unsetenv(key)
			} else {
				_ = os.Setenv(key, tt.setEnv)
			}
			got := LoadFromEnvVar()
			assert.Equal(t, tt.want, got)
		})
	}
}
