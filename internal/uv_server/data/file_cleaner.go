package data

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"time"

	"uv_server/internal/uv_server/common/loggers"

	"github.com/sirupsen/logrus"
)

type FileCleaner struct {
	log      *logrus.Entry
	to_clean <-chan string
}

func NewFileCleaner(to_clean chan string) *FileCleaner {
	object := &FileCleaner{}
	object.log = loggers.DataLogger
	object.to_clean = to_clean
	return object
}

func (f *FileCleaner) InitializeAndCleanDirectories() {
	storagePath := "storage"
	tmpPath := "tmp"
	f.log.Info("initializing directories")
	if err := f.createDirIfNotExists(storagePath); err != nil {
		f.log.Fatalf("failed to create storage directory: %v", err)
	}
	if err := f.createDirIfNotExists(tmpPath); err != nil {
		f.log.Fatalf("failed to create tmp directory: %v", err)
	}
	f.log.Info("cleaning tmp directory")
	if err := CleanDirectory(tmpPath); err != nil {
		f.log.Fatalf("failed to clean tmp directory: %v", err)
	}
}

func (f *FileCleaner) createDirIfNotExists(path string) error {
	info, err := os.Stat(path)
	if err != nil && errors.Is(err, fs.ErrNotExist) {
		return os.Mkdir(path, 0755)
	} else if err != nil {
		f.log.Fatalf("failed to stat file %v", err)
	}
	if !info.IsDir() {
		f.log.Fatalf("path exists and it's not a directory")
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
		if err := os.RemoveAll(itemPath); err != nil {
			return err
		}
	}
	return nil
}

func (f *FileCleaner) CleanUpLoop() {
	for path := range f.to_clean {
		var pathIsDeleted bool
		for range 5 {
			if err := os.RemoveAll(path); err != nil {
				time.Sleep(1 * time.Second)
				continue
			}
			_, err := os.Stat(path)
			if err == nil {
				time.Sleep(1 * time.Second)
				continue
			}
			if errors.Is(err, fs.ErrNotExist) {
				pathIsDeleted = true
				break
			}
		}
		if !pathIsDeleted {
			f.log.Errorf("failed to delete: %v", path)
		}
	}
}
