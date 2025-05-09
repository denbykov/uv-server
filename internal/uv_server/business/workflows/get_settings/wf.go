package getsettings

import (
	"context"
	"sync"
	cjmessages "uv_server/internal/uv_server/business/common_job_messages"
	"uv_server/internal/uv_server/business/data"
	"uv_server/internal/uv_server/common/loggers"
	"uv_server/internal/uv_server/config"

	"github.com/sirupsen/logrus"
)

type GetSettingsWf struct {
	uuid string

	log    *logrus.Entry
	config *config.Config

	jobCtx context.Context
	jobIn  chan<- interface{}

	database data.Database
}

func NewGetSettingsWf(
	uuid string,
	config *config.Config,
	jobCtx context.Context,
	jobIn chan<- interface{},
	job_out <-chan interface{},
	database data.Database,
) *GetSettingsWf {
	object := &GetSettingsWf{}
	object.uuid = uuid
	object.log = loggers.BusinessLogger.WithFields(
		logrus.Fields{
			"component": "GetSettingsWf",
			"uuid":      uuid,
		},
	)
	object.config = config
	object.jobCtx = jobCtx
	object.jobIn = jobIn
	_ = job_out
	object.database = database
	return object
}

func (w *GetSettingsWf) Run(wg *sync.WaitGroup) {
	defer wg.Done()
	result, err := w.database.GetSettings()

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
