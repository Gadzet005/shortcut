package configutils

import (
	"strings"

	errorsutils "github.com/Gadzet005/shortcut/shortcut/pkg/utils/errors"
)

type Env uint8

const (
	envUnknown = "unknown"
	envDev     = "dev"
	envProd    = "prod"

	EnvUnknown Env = iota
	EnvDev
	EnvProd
)

func (e Env) String() string {
	switch e {
	case EnvDev:
		return envDev
	case EnvProd:
		return envProd
	default:
		return envUnknown
	}
}

func (e Env) IsDev() bool {
	return e == EnvDev
}

func (e Env) IsProd() bool {
	return e == EnvProd
}

func (e *Env) UnmarshalYAML(unmarshal func(any) error) error {
	var str string
	if err := unmarshal(&str); err != nil {
		return err
	}
	str = strings.ToLower(str)

	switch str {
	case envDev:
		*e = EnvDev
	case envProd:
		*e = EnvProd
	default:
		return errorsutils.Errorf("unknown env value: %s", str)
	}

	return nil
}
