package server

import (
	"context"
	"net/http"
	"time"
)

const (
	defaultReadTimeout  = 5 * time.Second
	defaultWriteTimeout = 5 * time.Second
	defaultPort         = "8080"
)

type Server struct {
	server *http.Server
}

func New(handler http.Handler) *Server {
	return &Server{
		server: &http.Server{
			Handler:      handler,
			ReadTimeout:  defaultReadTimeout,
			WriteTimeout: defaultWriteTimeout,
			Addr:         ":" + defaultPort,
		},
	}
}

func (s *Server) Run() error {
	return s.server.ListenAndServe()
}

func (s *Server) Stop(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}
