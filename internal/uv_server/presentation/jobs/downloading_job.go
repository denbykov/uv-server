package jobs

import (
	"context"
	"log"
	"uv_server/internal/uv_server/common/loggers"
	"uv_server/internal/uv_server/config"
	"uv_server/internal/uv_server/presentation/messages"

	"github.com/sirupsen/logrus"
)

type DownloadingJob struct {
	uuid string
	// in
	// out

	session_in  chan *messages.Message
	session_out chan *messages.Message

	log    *logrus.Entry
	config *config.Config

	ctx    context.Context
	cancel context.CancelFunc
}

func NewDownloadingJob(
	config *config.Config,
	session_out chan *messages.Message,
	uuid string,
) *DownloadingJob {
	object := &DownloadingJob{}

	object.log = loggers.PresentationLogger.WithFields(
		logrus.Fields{
			"component": "DownloadingJob",
			"uuid":      uuid})
	object.config = config
	object.uuid = uuid
	object.session_out = session_out

	object.session_in = make(chan *messages.Message, 1)

	return object
}

func (j *DownloadingJob) Run(ctx context.Context, cancel context.CancelFunc, m *messages.Message) error {
	defer cancel()

	j.ctx = ctx
	j.cancel = cancel

	if m.Header.Type != messages.Download {
		log.Fatalf("unextected message type, got %v instead of Download", m.Header.Type)
	}

	return nil
}
