package bootstrap

import (
	"os"

	"uv_server/internal/uv_server/common/loggers"

	"github.com/sirupsen/logrus"
)

var (
	bootstrapLogger *logrus.Entry
)

func Run() {
	bootstrapLogger = loggers.BootstrapLogger
	bootstrapLogger.Info("initializing project directories")
	storagePath := "storage"
	tmpPath := "tmp"
	createDirIfNotExists(storagePath)
	createDirIfNotExists(tmpPath)
	bootstrapLogger.Info("cleaning tmp directory")
	сleanDirectory(tmpPath)
}

func InitProjectDirectories() error {
	storagePath := "storage"
	tmpPath := "tmp"
	createDirIfNotExists(storagePath)
	createDirIfNotExists(tmpPath)
	return nil
}

func createDirIfNotExists(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		err := os.Mkdir(path, 0755)
		if err != nil {
			bootstrapLogger.Panic(err, "failed to create "+path+"directory")
		}
	}
	return nil
}

func сleanDirectory(path string) error {
	dir, err := os.Open(path)
	if err != nil {
		bootstrapLogger.Panic(err, "failed to open "+path+" directory")
	}
	defer dir.Close()
	names, err := dir.Readdirnames(-1)
	if err != nil {
		bootstrapLogger.WithError(err).Errorf("failed to read %s directory", path)
	}
	for _, name := range names {
		err = os.RemoveAll(path + "/" + name)
		if err != nil {
			bootstrapLogger.Panic(err, "failed to remove "+name+" from tmp directory")
		}
	}
	return nil
}
