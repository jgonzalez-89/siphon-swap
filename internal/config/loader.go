package config

import (
	"errors"
	"os"

	ymlConf "github.com/Mth-Ryan/go-yaml-cfg"
)

const PATH = "CONFIG_PATH"

func LoadConfig() (Config, error) {
	path := os.Getenv(PATH)
	if path == "" {
		return Config{}, errors.New("CONFIG_PATH is not set")
	}

	if err := ymlConf.InitializeConfigSingleton[Config](path); err != nil {
		return Config{}, err
	}

	return ymlConf.GetConfigFromSingleton[Config]()
}
