package presentation

import (
	"fmt"
	"net/http"
	"server/common/loggers"
	"server/config"

	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(c *http.Request) bool { return true },
}

type Server struct {
	log    *logrus.Entry
	config *config.Config
}

func NewServer(config *config.Config) *Server {
	server := &Server{}
	server.log = loggers.PresentationLogger
	server.config = config

	return server
}

func (s *Server) Run() error {
	http.HandleFunc("/ws", s.handleConnection)

	port := fmt.Sprintf(":%v", s.config.Port)

	s.log.Infof("Websocket server started on %s", port)
	return http.ListenAndServe(port, nil)
}

func (s *Server) handleConnection(w http.ResponseWriter, r *http.Request) {
	s.log.Infof("Handling connection from %s", r.Host)

	ws, err := upgrader.Upgrade(w, r, nil)

	if err != nil {
		fmt.Println(err)
		return
	}
	defer ws.Close()

	for {
		_, msg, err := ws.ReadMessage()
		if err != nil {
			switch err.(type) {
			case *websocket.CloseError:
				s.log.Infof("Connection from %s is closed", r.Host)
				return
			default:
			}

			s.log.Error("Read error: ", err)
			return
		}

		s.log.Tracef("Received: %s", msg)

		if err := ws.WriteMessage(websocket.TextMessage, msg); err != nil {
			s.log.Error("Write error:", err)
			return
		}
	}
}
