package presentation

import (
	"errors"
	"net"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"

	"uv_server/internal/uv_protocol/presentation/messages"
	"uv_server/internal/uv_server/common/loggers"
	"uv_server/internal/uv_server/config"
	"uv_server/internal/uv_server/presentation/jobs"
)

const (
	messageLimit = 5
	writeWait    = 10 * time.Second
	pongWait     = 60 * time.Second
	pingPeriod   = (pongWait * 9) / 10
)

type Session struct {
	log     *logrus.Entry
	config  *config.Config
	conn    *websocket.Conn
	peer    string
	factory *JobFactory

	job_out chan *jobs.JobMessage

	jobs_mx sync.Mutex
	jobs    map[string]jobs.Job
}

func NewSession(
	config *config.Config,
	conn *websocket.Conn,
	peer string,
	factory *JobFactory) *Session {
	object := &Session{}

	object.log = loggers.PresentationLogger
	object.config = config
	object.conn = conn
	object.peer = peer
	object.factory = factory
	object.job_out = make(chan *jobs.JobMessage, messageLimit)

	object.jobs_mx = sync.Mutex{}
	object.jobs = make(map[string]jobs.Job)

	return object
}

func (s *Session) readPump() {
	defer func() {
		s.conn.Close()
	}()

	err := s.conn.SetReadDeadline(time.Now().Add(pongWait))
	if err != nil {
		s.log.Fatalf("failed to set read dead line: %v", err)
	}

	s.conn.SetPongHandler(func(string) error {
		return s.conn.SetReadDeadline(time.Now().Add(pongWait))
	})

	for {
		_, message, err := s.conn.ReadMessage()
		if err != nil {
			switch err.(type) {
			case *websocket.CloseError:
				s.log.Infof("connection from %s is closed", s.peer)
				return
			default:
				if errors.Is(err, net.ErrClosed) {
					s.log.Infof("connection from %s is closed", s.peer)
					return
				}

				s.log.Fatal("read error: ", err)
				return
			}
		}

		go func() {
			msg, err := messages.ParseMessage(message)
			_ = msg

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
				s.log.Tracef("creating new job for: %v", *msg.Header.Uuid)
				job, err := s.factory.CreateJob(msg, s.job_out)

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
		}()
	}
}

func (s *Session) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		s.conn.Close()
	}()

	for {
		select {
		case j_message, ok := <-s.job_out:
			if j_message.Done {
				s.jobs_mx.Lock()
				s.log.Tracef("removing job: %v", *j_message.Msg.Header.Uuid)
				delete(s.jobs, *j_message.Msg.Header.Uuid)
				s.jobs_mx.Unlock()
			}

			if !ok {
				err := s.conn.WriteMessage(websocket.CloseMessage, []byte{})
				if err != nil {
					s.log.Errorf("failed to write message: %v", err)
				}
				return
			}

			w, err := s.conn.NextWriter(websocket.BinaryMessage)
			if err != nil {
				s.log.Fatal(err)
			}

			message := j_message.Msg
			data := message.Serialize()

			_, err = w.Write(data)
			if err != nil {
				s.log.Errorf("failed to write message: %v", err)
			}

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			err := s.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err != nil {
				s.log.Fatalf("failed to set write dead line: %v", err)
			}

			if err := s.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (s *Session) Run() {
	go s.readPump()
	go s.writePump()
}
