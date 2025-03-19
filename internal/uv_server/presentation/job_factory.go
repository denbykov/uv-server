package presentation

import (
	"github.com/sirupsen/logrus"

	"uv_server/internal/uv_protocol/presentation/messages"
	"uv_server/internal/uv_server/common/loggers"
	"uv_server/internal/uv_server/config"
	"uv_server/internal/uv_server/presentation/jobs"
)

type JobFactory struct {
	log    *logrus.Entry
	config *config.Config
}

func NewJobFactory(config *config.Config) *JobFactory {
	object := &JobFactory{}
	object.log = loggers.PresentationLogger
	object.config = config

	return object
}

func (f *JobFactory) CreateJob(
	message *messages.Message, session_in chan *jobs.JobMessage) (jobs.Job, error) {
	typ := message.Header.Type

	f.log.Debugf("Creating Job for message type %v", typ)

	var Job jobs.Job = nil

	switch typ {
	case messages.DownloadingRequest:
		return jobs.NewDownloadingJob(*message.Header.Uuid, f.config, session_in), nil
	default:
		f.log.Fatalf("Unercognized message type %v", typ)
	}

	return Job, nil
}
