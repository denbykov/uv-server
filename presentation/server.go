package presentation

import (
	"fmt"
	"log"
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
	log      *logrus.Entry
	config   *config.Config
	factory  *HandlerFactory
	sessions []*Session
}

func NewServer(config *config.Config) *Server {
	object := &Server{}

	object.log = loggers.PresentationLogger
	object.config = config
	object.factory = NewHandlerFactory(config)
	object.sessions = make([]*Session, 0)

	return object
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
		log.Fatal(err)
	}

	// Could add session removal later, but as I'm expecting to have
	// only one client at the moment I do not really care
	session := NewSession(s.config, ws, r.Host, s.factory)
	s.sessions = append(s.sessions, session)
	session.Run()
}
