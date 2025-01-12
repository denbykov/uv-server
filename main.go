package main

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"

	"path/filepath"
	"server/common/loggers"
	"server/config"
	"server/data"
	"server/presentation"
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

	migrator := data.NewMigrator(config,
		data.NewMigrationRepositry(db))
	migrator.MigrateIfNeeded()

	server := presentation.NewServer(config)

	err = server.Run()

	if err != nil {
		log.Fatal(err)
	}
}
