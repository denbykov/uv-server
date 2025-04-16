package data

import (
	"os"
	"path/filepath"

	"uv_server/internal/uv_server/common/loggers"

	"github.com/sirupsen/logrus"
)

var (
	log *logrus.Entry
)

func InitializeAndCleanDirectories() {
	log = loggers.DataLogger
	storagePath := "storage"
	tmpPath := "tmp"
	log.Info("initializing directories")
	if err := createDirIfNotExists(storagePath); err != nil {
		log.Fatalf("failed to create storage directory: %v", err)
	}
	if err := createDirIfNotExists(tmpPath); err != nil {
		log.Fatalf("failed to create tmp directory: %v", err)
	}
	log.Info("cleaning tmp directory")
	if err := CleanDirectory(tmpPath); err != nil {
		log.Fatalf("failed to clean tmp directory: %v", err)
	}
}

func createDirIfNotExists(path string) error {
	if _, err := os.Stat(path); err != nil {
		err := os.Mkdir(path, 0755)
		if err != nil {
			return err
		}
	}
	return nil
}

func RemoveItem(path string) error {
	if err := os.RemoveAll(path); err != nil {
		return err
	}
	return nil
}

func CleanDirectory(path string) error {
	dir, err := os.Open(path)
	if err != nil {
		return err
	}
	defer dir.Close()
	names, err := dir.Readdirnames(-1)
	if err != nil {
		return err
	}
	for _, name := range names {
		itemPath := filepath.Join(path, name)
		if err := RemoveItem(itemPath); err != nil {
			return err
		}
	}
	return nil
}
