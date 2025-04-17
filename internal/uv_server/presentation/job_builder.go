package presentation

import (
	"fmt"

	"github.com/sirupsen/logrus"

	"uv_server/internal/uv_protocol"
	"uv_server/internal/uv_server/common/loggers"
	"uv_server/internal/uv_server/config"
	"uv_server/internal/uv_server/data"
	"uv_server/internal/uv_server/presentation/job"
)

type JobBuilder struct {
	log    *logrus.Entry
	config *config.Config

	resources *data.Resources
}

func NewJobBuilder(config *config.Config, resources *data.Resources) *JobBuilder {
	object := &JobBuilder{}
	object.log = loggers.PresentationLogger
	object.config = config

	object.resources = resources

	return object
}

func (b *JobBuilder) CreateJob(
	message *uv_protocol.Message,
	session_in chan<- *job.Message,
) (*job.Job, error) {
	typ := message.Header.Type

	b.log.Debugf("Creating Job for message type %v", typ)

	var j *job.Job = nil

	uuid := *message.Header.Uuid

	var wa job.WorkflowAdapter

	switch typ {
	case uv_protocol.DownloadingRequest:
		wa = job.NewDownloadingWfAdapter(
			uuid,
			b.config,
			session_in,
			b.resources,
		)
	default:
		return j, fmt.Errorf("unable to create job for message type %v", typ)
	}

	job := job.NewJob(
		uuid,
		b.config,
		session_in,
		wa,
	)

	return job, nil
}
