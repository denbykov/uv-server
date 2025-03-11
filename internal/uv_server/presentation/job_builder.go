package presentation

import (
	"database/sql"
	"fmt"

	"github.com/sirupsen/logrus"

	"uv_server/internal/uv_server/common/loggers"
	"uv_server/internal/uv_server/config"
	"uv_server/internal/uv_server/presentation/job"
	"uv_server/internal/uv_server/presentation/messages"
)

type JobBuilder struct {
	log    *logrus.Entry
	config *config.Config

	db *sql.DB
}

func NewJobBuilder(config *config.Config, db *sql.DB) *JobBuilder {
	object := &JobBuilder{}
	object.log = loggers.PresentationLogger
	object.config = config

	object.db = db

	return object
}

func (b *JobBuilder) CreateJob(
	message *messages.Message,
	session_in chan<- *job.Message,
) (*job.Job, error) {
	typ := message.Header.Type

	b.log.Debugf("Creating Job for message type %v", typ)

	var j *job.Job = nil

	uuid := *message.Header.Uuid

	var wa job.WorkflowAdapter

	switch typ {
	case messages.DownloadingRequest:
		wa = job.NewDownloadingWfAdapter(
			uuid,
			b.config,
			session_in,
			b.db,
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
