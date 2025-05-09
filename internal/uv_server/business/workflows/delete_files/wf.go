package deletefile

import (
	"context"
	"errors"
	"os"
	"path"
	"sync"
	cjmessages "uv_server/internal/uv_server/business/common_job_messages"
	"uv_server/internal/uv_server/business/data"
	jobmessages "uv_server/internal/uv_server/business/workflows/delete_files/job_messages"
	"uv_server/internal/uv_server/common/loggers"
	"uv_server/internal/uv_server/config"

	"github.com/sirupsen/logrus"
)

type DeleteFilesWf struct {
	uuid string

	log    *logrus.Entry
	config *config.Config

	jobCtx context.Context
	jobIn  chan<- interface{}

	database   data.Database
	filesystem data.Filesystem
}

func NewDeleteFilesWf(
	uuid string,
	config *config.Config,
	jobCtx context.Context,
	jobIn chan<- interface{},
	job_out <-chan interface{},
	database data.Database,
	filesystem data.Filesystem,
) *DeleteFilesWf {
	object := &DeleteFilesWf{}

	object.uuid = uuid
	object.log = loggers.BusinessLogger.WithFields(
		logrus.Fields{
			"component": "DeleteFilesWf",
			"uuid":      uuid},
	)
	object.config = config

	object.jobCtx = jobCtx

	object.jobIn = jobIn
	_ = job_out

	object.database = database
	object.filesystem = filesystem

	return object
}

func (w *DeleteFilesWf) Run(wg *sync.WaitGroup, request *jobmessages.Request) {
	defer wg.Done()

	failedFiles := []int64{}

	for _, id := range request.Ids {
		err := w.deleteFile(id)
		if err != nil {
			failedFiles = append(failedFiles, id)
		}
	}

	select {
	case <-w.jobCtx.Done():
		w.log.Debugf("workflow cancelled: %v", w.jobCtx.Err().Error())

		switch w.jobCtx.Err() {
		case context.DeadlineExceeded:
			w.jobIn <- &cjmessages.Error{Reason: "Timeout exceeded"}
		case context.Canceled:
			w.jobIn <- &cjmessages.Canceled{}
		}
	default:
	}

	if len(failedFiles) != 0 {
		w.jobIn <- &jobmessages.Error{FailedIds: failedFiles}
		return
	}

	w.jobIn <- &cjmessages.Done{}
}

func (w *DeleteFilesWf) deleteFile(id int64) error {
	log := w.log.WithField("id", id)

	file, err := w.database.GetFile(id)
	if err != nil {
		log.Errorf("failed to get file, error is: %v", err)
		return errors.New("failed to get file from database")
	}

	if file.Status != data.FsFinished {
		err := errors.New("file is not downloaded yet")
		log.Error(err)
		return err
	}

	if !file.Path.Valid {
		err := errors.New("file does not have a path")
		log.Error(err)
		return err
	}

	//  ToDo: integrate with settings
	wd, err := os.Getwd()
	if err != nil {
		w.log.Fatal(err)
	}

	storageDir := path.Join(wd, "storage")
	path := path.Join(storageDir, file.Path.String)
	//

	err = w.filesystem.DeleteFile(path)
	if err != nil {
		log.Errorf("failed to delete file from filesystem, error is: %v", err)
		return errors.New("failed to delete file from filesystem")
	}

	err = w.database.DeleteFile(&data.File{Id: id})
	if err != nil {
		log.Errorf("failed to delete file from database, error is: %v", err)
		return errors.New("failed to delete file from database")
	}

	return nil
}
