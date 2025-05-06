package deletefile

import (
	"context"
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

	database data.Database
}

func NewDeleteFilesWf(
	uuid string,
	config *config.Config,
	jobCtx context.Context,
	jobIn chan<- interface{},
	job_out <-chan interface{},
	database data.Database,
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

	return object
}

func (w *DeleteFilesWf) Run(wg *sync.WaitGroup, request *jobmessages.Request) {
	defer wg.Done()
	err := w.database.DeleteFiles(request.Ids)

	select {
	case <-w.jobCtx.Done():
		w.log.Debugf("workflow cancelled: %v", w.jobCtx.Err().Error())

		switch w.jobCtx.Err() {
		case context.DeadlineExceeded:
			w.jobIn <- &cjmessages.Error{Reason: "Timeout exceeded"}
		case context.Canceled:
			w.jobIn <- &cjmessages.Error{Reason: "Workflow cancelled"}
		}
	default:
	}

	if err != nil {
		w.jobIn <- &cjmessages.Error{Reason: err.Error()}
		return
	}

	w.jobIn <- &cjmessages.Done{}
}
