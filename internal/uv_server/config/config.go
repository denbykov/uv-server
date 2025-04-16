package config

import (
	"os"

	"github.com/sirupsen/logrus"

	"uv_server/internal/uv_server/common"
	"uv_server/internal/uv_server/common/loggers"
)

type Config struct {
	log *logrus.Entry

	Port int16 `yaml:"port"`

	FfmpegLocation     string `yaml:"ffmpegLocation"`
	ToolsLocation      string
	ChangesetsLocation string `yaml:"changesetsLocation"`

	AllowClientReconnect bool `yaml:"allowClientReconnect"`
}

func (config *Config) parse(path string) {

	file, err := os.ReadFile(path)

	if err != nil {
		config.log.Fatal(err)
	}

	err = common.UnmarshalYamlStrict(file, config)

	if err != nil {
		config.log.Fatal(err)
	}
}

func (config *Config) validateSriptsLocation() {
	if config.ToolsLocation == "" {
		config.log.Fatal(
			"scripts location is not specified")
	}

	stat, err := os.Stat(config.ToolsLocation)

	if os.IsNotExist(err) {
		config.log.Fatalf(
			"scripts location does not exist: '%v'",
			config.ToolsLocation)
	}

	if !stat.IsDir() {
		config.log.Fatalf(
			"scripts location is not a directory: '%v'",
			config.ToolsLocation)
	}
}

func (config *Config) validateFfmpegLocation() {
	if config.FfmpegLocation == "" {
		config.log.Fatal(
			"ffmpeg location is not specified")
	}

	stat, err := os.Stat(config.FfmpegLocation)

	if os.IsNotExist(err) {
		config.log.Fatalf(
			"ffmpeg location does not exist: '%v'",
			config.FfmpegLocation)
	}

	if !stat.IsDir() {
		config.log.Fatalf(
			"ffmpeg location is not a directory: '%v'",
			config.FfmpegLocation)
	}
}

func NewConfig(path string) *Config {
	config := &Config{}
	config.log = loggers.ApplicationLogger

	config.parse(path)

	if config.Port == 0 {
		config.log.Fatal("port is not specified")
	}

	config.validateFfmpegLocation()

	config.ToolsLocation = "tools"
	config.validateSriptsLocation()

	return config
}
