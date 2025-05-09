package job

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"sync"
	"time"
	"uv_server/internal/uv_protocol"
	cjmessages "uv_server/internal/uv_server/business/common_job_messages"
	"uv_server/internal/uv_server/common/loggers"
	"uv_server/internal/uv_server/config"

	"github.com/sirupsen/logrus"
)

type State int

const (
	Active State = iota
	Canceled
	Done
	None
)

func (t State) String() string {
	switch t {
	case Active:
		return "Active"
	case Canceled:
		return "Canceled"
	case Done:
		return "Done"
	case None:
		return "None"
	default:
		return fmt.Sprintf("Unknown: %d", t)
	}
}

type Message struct {
	Msg  *uv_protocol.Message
	Done bool
}

type Job struct {
	uuid string

	session_in  chan<- *Message
	session_out chan *uv_protocol.Message

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

	object.session_out = make(chan *uv_protocol.Message, 1)

	object.wf_in = make(chan interface{}, 1)
	object.wf_out = make(chan interface{}, 1)

	object.wf_adatapter = wf_adatapter

	return object
}

func (j *Job) Notify(m *uv_protocol.Message) {
	j.log.Tracef("Notify: handling message %v", m)
	j.session_out <- m
}

func (j *Job) Run(m *uv_protocol.Message) {
	j.log.Tracef("Run: handling message %v", m)

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	j.wf_adatapter.CreateWf(
		j.uuid,
		j.config,
		ctx,
		j.wf_in,
		j.wf_out,
	)

	var wg sync.WaitGroup
	err := j.wf_adatapter.RunWf(&wg, m)

	if err != nil {
		err_msg := j.buildErrorMessage(err.Error())
		j.session_in <- err_msg
		return
	}

	nextState := Active
	for {
		switch nextState {
		case Active:
			nextState = j.active(ctx, cancel)
		case Canceled:
			nextState = j.canceled(ctx, &wg)
		case Done:
			nextState = j.done(&wg)
		case None:
			j.releaseResources()
			return
		}
	}
}

func (j *Job) buildErrorMessage(reason string) *Message {
	payload, err := json.Marshal(cjmessages.Error{Reason: reason})
	if err != nil {
		j.log.Fatalf("failed to serialize message: %v", err)
	}

	msg := &Message{
		Msg: &uv_protocol.Message{
			Header: &uv_protocol.Header{
				Uuid: &j.uuid,
				Type: uv_protocol.Error,
			},
			Payload: payload,
		},
		Done: true,
	}

	return msg
}

func (j *Job) buildCanceledMessage() *Message {
	msg := &Message{
		Msg: &uv_protocol.Message{
			Header: &uv_protocol.Header{
				Uuid: &j.uuid,
				Type: uv_protocol.Canceled,
			},
			Payload: nil,
		},
		Done: true,
	}

	return msg
}

func (j *Job) buildDoneMessage() *Message {
	msg := &Message{
		Msg: &uv_protocol.Message{
			Header: &uv_protocol.Header{
				Uuid: &j.uuid,
				Type: uv_protocol.Done,
			},
			Payload: nil,
		},
		Done: true,
	}

	return msg
}

func (j *Job) active(
	ctx context.Context, cancel context.CancelFunc) State {
	j.log.Trace("entering active state")

	for {
		select {
		case <-ctx.Done():
			return Canceled
		case msg := <-j.session_out:
			if msg.Header.Type == uv_protocol.CancelRequest {
				cancel()
				return Canceled
			}

			err := j.wf_adatapter.HandleSessionMessage(msg)

			if err != nil {
				j.log.Errorf("failed to handle message: %v", err)
				err_msg := j.buildErrorMessage(err.Error())
				j.session_in <- err_msg
			}
		case msg := <-j.wf_out:
			if tMsg, ok := msg.(*cjmessages.Error); ok {
				err_msg := j.buildErrorMessage(tMsg.Reason)
				j.session_in <- err_msg
				return None
			} else if _, ok := msg.(*cjmessages.Done); ok {
				j.session_in <- j.buildDoneMessage()
				return None
			} else {
				state, err := j.wf_adatapter.HandleWfMessage(msg)

				if err != nil {
					err_msg := j.buildErrorMessage(err.Error())
					j.session_in <- err_msg
					return None
				}

				if state != Active {
					return state
				}
			}
		}
	}
}

func (j *Job) canceled(ctx context.Context, wg *sync.WaitGroup) State {
	j.log.Trace("entering canceled state")

	wg.Wait()

	select {
	case msg := <-j.wf_out:
		if tMsg, ok := msg.(*cjmessages.Error); ok {
			err_msg := j.buildErrorMessage(tMsg.Reason)
			j.session_in <- err_msg
		} else if _, ok := msg.(*cjmessages.Canceled); ok {
			err_msg := j.buildCanceledMessage()
			j.session_in <- err_msg
		} else {
			j.log.Fatalf("Unexpected workflow message: %v %v", reflect.TypeOf(msg), msg)
		}
	default:
		j.log.Warnf("workflow exited with no user notification")

		var reason string

		switch ctx.Err() {
		case context.DeadlineExceeded:
			j.log.Debugf("job canceled due to the timeout, uuid is %v", j.uuid)
			reason = "tiemout"
		case context.Canceled:
			j.log.Debugf("job canceled, uuid is %v", j.uuid)
			reason = "cancelled"
		}

		err_msg := j.buildErrorMessage(reason)
		j.session_in <- err_msg
	}

	return None
}

func (j *Job) done(wg *sync.WaitGroup) State {
	j.log.Trace("entering done state")
	wg.Wait()
	return None
}

func (j *Job) releaseResources() {
	close(j.session_out)
	close(j.wf_in)
	close(j.wf_out)
}
