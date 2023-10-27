package server

import (
	"github.com/clusterlink-net/clusterlink/pkg/util"
	"github.com/clusterlink-net/clusterlink/pkg/utils/netutils"
	"github.com/go-chi/chi"
	"github.com/sirupsen/logrus"
)

type Server struct {
	router         *chi.Mux
	parsedCertData *util.ParsedCertData
	logger         *logrus.Entry
}

// StartFrelayServer starts the Dataplane server
func (s *Server) StartFrelayServer(frelayServerAddress string) error {
	s.logger.Infof("Dataplane server starting at %s.", frelayServerAddress)
	server := netutils.CreateResilientHTTPServer(frelayServerAddress, s.router, s.parsedCertData.ServerConfig(), nil, nil, nil)

	return server.ListenAndServeTLS("", "")
}

func (s *Server) addApiHandlers() {
	s.router.Post("/user", s.addUser)
}

func (s *Server) addAuthHandlers() {
	s.router.Post("/", s.ingressAuthorize)
}

// NewDataplane returns a new dataplane HTTP server.
func NewFrelay(parsedCertData *util.ParsedCertData) *Server {
	d := &Server{
		router:         chi.NewRouter(),
		parsedCertData: parsedCertData,
		logger:         logrus.WithField("component", "server.frelay"),
	}

	d.addAuthHandlers()
	d.addApiHandlers()

	return d
}
