package main

import (
	"path/filepath"
	"server/common/loggers"
	"server/config"
	"server/presentation"
)

func main() {
	loggers.Init("logs", "log.txt")
	defer loggers.CloseLogFile()

	loggers.ApplicationLogger.Info("Starting...")

	config := config.NewConfig(filepath.Join("config", "config.yaml"))

	server := presentation.NewServer(config)

	err := server.Run()

	if err != nil {
		loggers.PresentationLogger.Error(err)
	}
}
