package job

import (
	"context"
	"encoding/json"
	"sync"
	"time"
	commonJobMessages "uv_server/internal/uv_server/business/common_job_messages"
	"uv_server/internal/uv_server/common/loggers"
	"uv_server/internal/uv_server/config"
	"uv_server/internal/uv_server/presentation/messages"

	"github.com/sirupsen/logrus"
)

type Message struct {
	Msg  *messages.Message
	Done bool
}

type Job struct {
	uuid string

	session_in  chan<- *Message
	session_out chan *messages.Message

	log    *logrus.Entry
	config *config.Config

	wf_in  chan interface{}
	wf_out chan interface{}

	wf_adatapter WorkflowAdapter
}

func NewJob(
	uuid string,
	config *config.Config,
	session_in chan<- *Message,
	wf_adatapter WorkflowAdapter,
) *Job {
	object := &Job{}

	object.log = loggers.PresentationLogger.WithFields(
		logrus.Fields{
			"component": "Job",
			"uuid":      uuid})
	object.config = config
	object.uuid = uuid
	object.session_in = session_in

	object.session_out = make(chan *messages.Message, 1)

	object.wf_in = make(chan interface{}, 1)
	object.wf_out = make(chan interface{}, 1)

	object.wf_adatapter = wf_adatapter

	return object
}

func (j *Job) Notify(m *messages.Message) {
	j.log.Tracef("Notify: handling message %v", m)
	j.session_out <- m
}

func (j *Job) Run(m *messages.Message) {
	j.log.Tracef("Run: handling message %v", m)

	if m.Header.Type != messages.DownloadingRequest {
		j.log.Fatalf("Run: unextected message type, got %v instead of Download", m.Header.Type)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	j.wf_adatapter.CreateWf(
		j.uuid,
		j.config,
		ctx,
		cancel,
		j.wf_in,
		j.wf_out,
	)

	var wg sync.WaitGroup
	wg.Add(1)
	go j.wf_adatapter.RunWf(&wg)

	j.active(ctx, cancel, &wg)
}

func (j *Job) active(
	ctx context.Context, cancel context.CancelFunc, wg *sync.WaitGroup) {
	j.log.Trace("entering active state")

	for {
		select {
		case <-ctx.Done():
			j.canceled(ctx, wg)
			return
		case msg := <-j.session_out:
			if msg.Header.Type == messages.CancelRequest {
				cancel()
				j.canceled(ctx, wg)
				return
			}
			_ = msg
		case msg := <-j.wf_out:
			_ = msg
		}
	}
}

func (j *Job) canceled(ctx context.Context, wg *sync.WaitGroup) {
	j.log.Trace("entering canceled state")

	wg.Wait()

	select {
	case msg := <-j.wf_out:
		_ = msg
	default:
		j.log.Warnf("workflow exited with no user notification")

		business_msg := commonJobMessages.Canceled{}

		switch ctx.Err() {
		case context.DeadlineExceeded:
			j.log.Debugf("job canceled due to the timeout, uuid is %v", j.uuid)
			business_msg.Reason = "tiemout"
		case context.Canceled:
			j.log.Debugf("job canceled, uuid is %v", j.uuid)
			business_msg.Reason = "cancelled"
		}

		payload, err := json.Marshal(business_msg)
		if err != nil {
			j.log.Fatalf("Failed to serialize message: %v", err)
		}

		msg := Message{
			Msg: &messages.Message{
				Header: &messages.Header{
					Uuid: &j.uuid,
					Type: messages.Canceled,
				},
				Payload: payload,
			},
			Done: true,
		}

		j.session_in <- &msg
	}

	j.releaseResources()
}

// func (j *Job) done(wg *sync.WaitGroup) {
// 	j.log.Trace("entering done state")

// 	wg.Wait()
// }

func (j *Job) releaseResources() {
	close(j.session_out)
	close(j.wf_in)
	close(j.wf_out)
}
