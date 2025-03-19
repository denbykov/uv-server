package jobs

import (
	"context"
	"sync"
	"time"
	"uv_server/internal/uv_protocol/presentation/messages"
	"uv_server/internal/uv_server/business/workflows/downloading"
	"uv_server/internal/uv_server/common/loggers"
	"uv_server/internal/uv_server/config"

	"github.com/sirupsen/logrus"
)

type DownloadingJob struct {
	uuid string

	session_in  chan<- *JobMessage
	session_out chan *messages.Message

	log    *logrus.Entry
	config *config.Config

	wf     *downloading.DownloadingWf
	wf_in  chan interface{}
	wf_out chan interface{}
}

func NewDownloadingJob(
	uuid string,
	config *config.Config,
	session_in chan<- *JobMessage,
) *DownloadingJob {
	object := &DownloadingJob{}

	object.log = loggers.PresentationLogger.WithFields(
		logrus.Fields{
			"component": "DownloadingJob",
			"uuid":      uuid})
	object.config = config
	object.uuid = uuid
	object.session_in = session_in

	object.session_out = make(chan *messages.Message, 1)

	return object
}

func (j *DownloadingJob) Notify(m *messages.Message) {
	j.log.Tracef("Notify: handling message %v", m)

	if m.Header.Type == messages.DownloadingRequest {
		j.log.Warnf("Notify: unextected start job message %v", m.Header.Type)
		return
	}

	j.session_out <- m
}

func (j *DownloadingJob) Run(m *messages.Message) {
	j.log.Tracef("Run: handling message %v", m)

	if m.Header.Type != messages.DownloadingRequest {
		j.log.Fatalf("Run: unextected message type, got %v instead of Download", m.Header.Type)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	j.wf = downloading.NewDownloadingWf(
		j.uuid,
		j.config,
		ctx,
		cancel,
	)
	j.wf_in = make(chan interface{}, 1)
	j.wf_out = make(chan interface{}, 1)

	var wg sync.WaitGroup
	wg.Add(1)
	go j.wf.Run(&wg)

	j.active(ctx, cancel, &wg)
}

func (j *DownloadingJob) active(
	ctx context.Context, cancel context.CancelFunc, wg *sync.WaitGroup) {
	j.log.Trace("entering active state")

	for {
		select {
		case <-ctx.Done():
			j.cancelled(wg)
			return
		case msg := <-j.session_out:
			_ = msg
		case msg := <-j.wf_out:
			_ = msg
		}
	}
}

func (j *DownloadingJob) cancelled(wg *sync.WaitGroup) {
	j.log.Trace("entering cancelled state")

	wg.Wait()

	select {
	case msg := <-j.session_out:
		_ = msg
	default:
		j.log.Warnf("workflow exited with no user notification")

		msg := JobMessage{
			Msg: &messages.Message{
				Header: &messages.Header{
					Uuid: &j.uuid,
					Type: messages.DownloadingCancelled,
				},
			},
			Done: true,
		}

		j.session_in <- &msg
	}

	j.releaseResources()
}

// func (j *DownloadingJob) done(wg *sync.WaitGroup) {
// 	j.log.Trace("entering done state")

// 	wg.Wait()
// }

func (j *DownloadingJob) releaseResources() {
	close(j.session_out)
	close(j.wf_in)
	close(j.wf_out)
}
