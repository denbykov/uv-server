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

	log.Info("starting...")

	config := config.NewConfig(filepath.Join("config", "config.yaml"))

	db, err := sql.Open("sqlite3", "app.db")
	if err != nil {
		log.Fatal(err)
	}

	DbMigrator := data.NewDbMigrator(
		config,
		db,
	)
	DbMigrator.MigrateIfNeeded()

	to_clean := make(chan string, 5)

	cleaner := data.NewFileCleaner(to_clean)
	cleaner.InitializeAndCleanDirectories()
	go cleaner.CleanUpLoop()

	resources := data.Resources{
		Db:       db,
		To_clean: to_clean,
	}

	server := presentation.NewServer(config, &resources)

	err = server.Run()

	if err != nil {
		log.Fatal(err)
	}
}
