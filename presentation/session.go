package presentation

import (
	"errors"
	"net"
	"server/common/loggers"
	"server/config"
	"server/presentation/messages"
	"time"

	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
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
	factory *HandlerFactory
	send    chan *messages.Message
}

func NewSession(
	config *config.Config,
	conn *websocket.Conn,
	peer string,
	factory *HandlerFactory) *Session {
	object := &Session{}

	object.log = loggers.PresentationLogger
	object.config = config
	object.conn = conn
	object.peer = peer
	object.factory = factory
	object.send = make(chan *messages.Message, messageLimit)

	return object
}

func (s *Session) readPump() {
	defer func() {
		s.conn.Close()
	}()

	if err := s.conn.SetReadDeadline(time.Now().Add(pongWait)); err != nil {
		s.log.Error("SetReadDeadline error: ", err)
		return
	}
	s.conn.SetPongHandler(func(string) error {
		if err := s.conn.SetReadDeadline(time.Now().Add(pongWait)); err != nil {
			s.log.Error("SetReadDeadline error: ", err)
		}
		return nil
	})

	for {
		_, message, err := s.conn.ReadMessage()
		if err != nil {
			switch err.(type) {
			case *websocket.CloseError:
				s.log.Infof("Connection from %s is closed", s.peer)
				return
			default:
				if errors.Is(err, net.ErrClosed) {
					s.log.Infof("Connection from %s is closed", s.peer)
					return
				}

				s.log.Fatal("read error: ", err)
				return
			}
		}

		go func() {
			msg, err := messages.ParseMessage(message)

			if err != nil {
				s.log.Error(err)
				s.conn.Close()
				return
			}

			handler, err := s.factory.CreateHandler(msg)

			if err != nil {
				s.log.Error(err)
				s.conn.Close()
				return
			}

			err = handler.Handle(msg, s.send)

			if err != nil {
				s.log.Error(err)
				s.conn.Close()
				return
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
		case message, ok := <-s.send:
			if !ok {
				if err := s.conn.WriteMessage(websocket.CloseMessage, []byte{}); err != nil {
					s.log.Fatal("WriteMessage error: ", err)
				}
				return
			}

			w, err := s.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				s.log.Fatal(err)
			}

			data := message.Serialize()

			if _, err := w.Write(data); err != nil {
				s.log.Error("Write error: ", err)
				return
			}

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			if err := s.conn.SetWriteDeadline(time.Now().Add(writeWait)); err != nil {
				s.log.Fatal("SetWriteDeadline error: ", err)
				return
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
