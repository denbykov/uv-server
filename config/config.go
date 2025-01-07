package config

import (
	"os"
	"server/common"
	"server/common/loggers"

	"github.com/sirupsen/logrus"
)

type Config struct {
	log *logrus.Entry

	Port           int16  `yaml:"port"`
	FfmpegLocation string `yaml:"ffmpegLocation"`

	ScriptsLocation string
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
	if config.ScriptsLocation == "" {
		config.log.Fatal(
			"scripts location is not specified")
	}

	stat, err := os.Stat(config.ScriptsLocation)

	if os.IsNotExist(err) {
		config.log.Fatalf(
			"scripts location does not exist: '%v'",
			config.ScriptsLocation)
	}

	if !stat.IsDir() {
		config.log.Fatalf(
			"scripts location is not a directory: '%v'",
			config.ScriptsLocation)
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

	config.ScriptsLocation = "scripts"
	config.validateSriptsLocation()

	return config
}
