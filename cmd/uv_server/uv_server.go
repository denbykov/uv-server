package main

import (
	"database/sql"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"

	"uv_server/internal/uv_server/common/loggers"
	"uv_server/internal/uv_server/config"
	"uv_server/internal/uv_server/data"
	"uv_server/internal/uv_server/presentation"
)

func main() {
	loggers.Init("logs", "log.txt")
	defer loggers.CloseLogFile()

	log := loggers.ApplicationLogger

	log.Info("Starting...")

	config := config.NewConfig(filepath.Join("config", "config.yaml"))

	db, err := sql.Open("sqlite3", "app.db")
	if err != nil {
		log.Fatal(err)
	}

	migrator := data.NewMigrator(
		config,
		data.NewMigrationRepositry(db),
	)
	migrator.MigrateIfNeeded()

	server := presentation.NewServer(config)

	err = server.Run()

	if err != nil {
		log.Fatal(err)
	}
}
