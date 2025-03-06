package jobs

import (
	"context"
	"time"
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
	session_out chan *JobMessage

	log    *logrus.Entry
	config *config.Config

	ctx    context.Context
	cancel context.CancelFunc
}

func NewDownloadingJob(
	config *config.Config,
	session_out chan *JobMessage,
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

func (j *DownloadingJob) Run(m *messages.Message) {
	j.log.Tracef("Run: handling message %v", m)

	j.ctx, j.cancel = context.WithTimeout(context.Background(), 20*time.Second)
	defer j.cancel()

	if m.Header.Type != messages.Download {
		j.log.Fatalf("Run: unextected message type, got %v instead of Download", m.Header.Type)
	}
}

func (j *DownloadingJob) Notify(m *messages.Message) {
	j.log.Tracef("Notify: handling message %v", m)

	if m.Header.Type == messages.Download {
		j.log.Warnf("Notify: unextected start job message %v", m.Header.Type)
		return
	}
}
