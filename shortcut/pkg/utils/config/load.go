package configutils

import (
	"os"

	"dario.cat/mergo"
	errorsutils "github.com/Gadzet005/shortcut/shortcut/pkg/utils/errors"
	"gopkg.in/yaml.v3"
)

// LoadConfig загружает несколько конфигурационных файлов последовательно.
// Каждый последующий файл переопределяет значения из предыдущих.
// Первый файл в списке является базовым конфигом.
func LoadConfig[Config any](configPaths []string) (Config, error) {
	var config Config

	if len(configPaths) == 0 {
		return config, errorsutils.Error("no config paths provided")
	}

	config, err := ReadConfig[Config](configPaths[0])
	if err != nil {
		return config, err
	}

	for i := 1; i < len(configPaths); i++ {
		overlay, err := ReadConfig[Config](configPaths[i])
		if err != nil {
			return config, err
		}

		if err := mergo.Merge(&config, overlay, mergo.WithOverride); err != nil {
			return config, errorsutils.WrapFail(err, "merge config %s", configPaths[i])
		}
	}

	return config, nil
}

func ReadConfig[Config any](configPath string) (Config, error) {
	var config Config

	data, err := os.ReadFile(configPath)
	if err != nil {
		return config, errorsutils.WrapFail(err, "read config %s", configPath)
	}

	if err := yaml.Unmarshal(data, &config); err != nil {
		return config, errorsutils.WrapFail(err, "unmarshal config %s", configPath)
	}

	return config, nil
}
