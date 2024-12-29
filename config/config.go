package config

import (
	"os"
	"server/common/loggers"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Port int16
}

func (config *Config) parse(path string) {
	loggers.ApplicationLogger.Infof("Reading config file %s", path)

	file, err := os.ReadFile(path)

	if err != nil {
		loggers.ApplicationLogger.Fatal(err)
	}

	yaml.Unmarshal(file, config)
}

func NewConfig(path string) *Config {
	config := &Config{}
	config.parse(path)
	return config
}
