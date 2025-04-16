package presentation

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"

	"uv_server/internal/uv_server/common/loggers"
	"uv_server/internal/uv_server/config"
	"uv_server/internal/uv_server/data"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(c *http.Request) bool { return true },
}

type Server struct {
	log        *logrus.Entry
	config     *config.Config
	jobBuilder *JobBuilder
	sessions   []*Session

	srv *http.Server
}

func NewServer(config *config.Config, resources *data.Resources) *Server {
	object := &Server{}

	object.log = loggers.PresentationLogger
	object.config = config
	object.jobBuilder = NewJobBuilder(config, resources)
	object.sessions = make([]*Session, 0)

	return object
}

func (s *Server) Run() error {
	http.HandleFunc("/ws", s.handleConnection)

	addr := fmt.Sprintf(":%v", s.config.Port)
	s.srv = &http.Server{Addr: addr}

	s.log.Infof("websocket server started on %s", addr)

	return s.srv.ListenAndServe()
}

func (s *Server) handleConnection(w http.ResponseWriter, r *http.Request) {
	s.log.Infof("handling connection from %s", r.RemoteAddr)

	ws, err := upgrader.Upgrade(w, r, nil)

	if err != nil {
		log.Fatal(err)
	}

	// Could add session removal later, but as I'm expecting to have
	// only one client at the moment so I do not really care

	session := NewSession(s.config, ws, r.RemoteAddr, s.jobBuilder, s.srv)

	s.sessions = append(s.sessions, session)
	session.Run()
}
