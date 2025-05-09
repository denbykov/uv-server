package getfiles

import (
	"context"
	"sync"
	cjmessages "uv_server/internal/uv_server/business/common_job_messages"
	"uv_server/internal/uv_server/business/data"
	jobmessages "uv_server/internal/uv_server/business/workflows/get_files/job_messages"
	"uv_server/internal/uv_server/common/loggers"
	"uv_server/internal/uv_server/config"

	"github.com/sirupsen/logrus"
)

type GetFilesWf struct {
	uuid string

	log    *logrus.Entry
	config *config.Config

	jobCtx context.Context
	jobIn  chan<- interface{}

	database data.Database
}

func NewGetFilesWf(
	uuid string,
	config *config.Config,
	jobCtx context.Context,
	jobIn chan<- interface{},
	job_out <-chan interface{},
	database data.Database,
) *GetFilesWf {
	object := &GetFilesWf{}

	object.uuid = uuid
	object.log = loggers.BusinessLogger.WithFields(
		logrus.Fields{
			"component": "GetFilesWf",
			"uuid":      uuid},
	)
	object.config = config

	object.jobCtx = jobCtx

	object.jobIn = jobIn
	_ = job_out

	object.database = database

	return object
}

func (w *GetFilesWf) Run(wg *sync.WaitGroup, request *jobmessages.Request) {
	defer wg.Done()
	result, err := w.database.GetFilesForGFW(request)

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

	if err != nil {
		w.jobIn <- &cjmessages.Error{Reason: err.Error()}
		return
	}

	w.jobIn <- result
}
