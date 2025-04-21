package presentation

import (
	"context"
	"time"

	"errors"
	"net"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"

	"uv_server/internal/uv_protocol"
	"uv_server/internal/uv_server/common/loggers"
	"uv_server/internal/uv_server/config"
	"uv_server/internal/uv_server/presentation/job"
)

const (
	messageLimit = 5
	// writeWait    = 10 * time.Second
	// pongWait     = 60 * time.Second
	// pingPeriod   = (pongWait * 9) / 10
)

type Session struct {
	log     *logrus.Entry
	config  *config.Config
	conn    *websocket.Conn
	peer    string
	builder *JobBuilder
	srv     *http.Server

	job_out chan *job.Message

	jobs_mx sync.Mutex
	jobs    map[string]*job.Job
}

func NewSession(
	config *config.Config,
	conn *websocket.Conn,
	peer string,
	builder *JobBuilder,
	srv *http.Server,
) *Session {
	object := &Session{}

	object.log = loggers.PresentationLogger
	object.config = config
	object.conn = conn
	object.peer = peer
	object.builder = builder
	object.job_out = make(chan *job.Message, messageLimit)

	object.jobs_mx = sync.Mutex{}
	object.jobs = make(map[string]*job.Job)

	object.srv = srv

	return object
}

func (s *Session) readPump() {
	defer func() {
		s.conn.Close()
	}()

	for {
		_, message, err := s.conn.ReadMessage()
		if err != nil {
			switch err.(type) {
			case *websocket.CloseError:
				s.log.Infof("connection from %s is closed %T: %v", s.peer, err, err)

				if !s.config.AllowClientReconnect {
					err = s.srv.Shutdown(context.TODO())
					if err != nil {
						s.log.Fatalf("failed to shut server down gracefully: %v", err)
					}
				}

				return
			default:
				if errors.Is(err, net.ErrClosed) {
					s.log.Infof("connection from %s is closed %T: %v", s.peer, err, err)

					if !s.config.AllowClientReconnect {
						err = s.srv.Shutdown(context.TODO())
						if err != nil {
							s.log.Fatalf("failed to shut server down gracefully: %v", err)
						}
					}

					return
				}

				s.log.Fatal("read error: ", err)
				return
			}
		}
		go s.handleIncomingMessage(message)
	}
}

func (s *Session) handleIncomingMessage(raw_msg []byte) {
	msg, err := uv_protocol.ParseMessage(raw_msg)

	if err != nil {
		s.log.Error(err)
		s.conn.Close()
		return
	}

	s.jobs_mx.Lock()
	job, ok := s.jobs[*msg.Header.Uuid]
	s.jobs_mx.Unlock()

	if ok {
		job.Notify(msg)
	} else {
		if msg.Header.Type == uv_protocol.CancelRequest {
			s.log.Debugf("received cancel request for non existing job: %v", *msg.Header.Uuid)
			return
		}

		s.log.Tracef("creating new job for: %v", *msg.Header.Uuid)
		job, err := s.builder.CreateJob(msg, s.job_out)

		if err != nil {
			s.log.Error(err)
			s.conn.Close()
			return
		}

		go job.Run(msg)

		s.jobs_mx.Lock()
		s.jobs[*msg.Header.Uuid] = job
		s.jobs_mx.Unlock()
	}
}

func (s *Session) writePump() {
	defer func() {
		s.conn.Close()
	}()

	for j_message := range s.job_out {
		if j_message.Done {
			s.jobs_mx.Lock()
			s.log.Tracef("removing job: %v", *j_message.Msg.Header.Uuid)

			_, ok := s.jobs[*j_message.Msg.Header.Uuid]
			if !ok {
				s.log.Fatalf(
					"trying to remove non-existing job: %v",
					*j_message.Msg.Header.Uuid)
			}
			delete(s.jobs, *j_message.Msg.Header.Uuid)

			s.jobs_mx.Unlock()
		}

		w, err := s.conn.NextWriter(websocket.BinaryMessage)
		if err != nil {
			s.log.Fatal(err)
		}

		message := j_message.Msg
		data := message.Serialize()

		s.conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
		_, err = w.Write(data)
		if err != nil {
			s.log.Errorf("failed to write message: %v", err)
		}

		if err := w.Close(); err != nil {
			return
		}
	}
}

func (s *Session) Run() {
	go s.readPump()
	go s.writePump()
}
