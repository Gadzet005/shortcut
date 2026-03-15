package env

import "os"

const (
	envVarName = "ENV"
)

type Env string

const (
	EnvUnknown Env = "unknown"
	EnvDev     Env = "dev"
	EnvTesting Env = "testing"
	EnvProd    Env = "prod"
)

func (e Env) String() string {
	return string(e)
}

func (e Env) IsProd() bool {
	return e == EnvProd
}

func (e Env) IsDev() bool {
	return e == EnvDev
}

func (e Env) IsTesting() bool {
	return e == EnvTesting
}

func (e Env) IsUnknown() bool {
	return e == EnvUnknown
}

func ParseEnv(env string) Env {
	switch env {
	case EnvDev.String():
		return EnvDev
	case EnvTesting.String():
		return EnvTesting
	case EnvProd.String():
		return EnvProd
	}
	return EnvUnknown
}

func LoadFromEnvVar() Env {
	envRaw := os.Getenv(envVarName)
	if envRaw == "" {
		return EnvDev
	}
	return ParseEnv(envRaw)
}
