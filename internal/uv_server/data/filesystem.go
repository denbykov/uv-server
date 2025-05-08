package data

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"uv_server/internal/uv_server/common/loggers"

	"github.com/sirupsen/logrus"
)

type Filesystem struct {
	log *logrus.Entry
}

func NewFilesystem() *Filesystem {
	object := &Filesystem{}

	object.log = loggers.DataLogger.
		WithField("component", "filesystem")

	return object
}

func (f *Filesystem) DeleteFile(path string) error {
	return f.DeleteLocalFile(path)
}

func (f *Filesystem) DeleteLocalFile(path string) error {
	info, err := os.Stat(path)
	if err != nil && errors.Is(err, fs.ErrNotExist) {
		err := fmt.Errorf("trying to delete non-existing file: %v", path)
		f.log.Error(err)
		return err
	} else if err != nil {
		f.log.Fatalf("failed to stat file %v", err)
	}

	if info.IsDir() {
		f.log.Fatalf("path exists but it's a directory")
	}

	return os.Remove(path)
}
