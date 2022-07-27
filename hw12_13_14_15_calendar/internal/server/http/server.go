package internalhttp

import (
	"context"
	"net/http"
)

type Server struct {
	logger Logger
	app    Application
	addr   string
	server *http.Server
}

type Logger interface {
	Info(ip, method, path, protocol string, respCode int, latency, userAgent string)
}

type Application interface { // TODO
}

func NewServer(logger Logger, app Application, addr string) *Server {
	server := &Server{
		logger: logger,
		app:    app,
		addr:   addr,
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/", server.hello)

	server.server = &http.Server{
		Addr:    addr,
		Handler: loggingMiddleware(mux, logger),
	}
	return server
}

func (s *Server) Start(ctx context.Context) error {
	err := s.server.ListenAndServe()
	<-ctx.Done()
	if err == http.ErrServerClosed {
		return nil
	}
	return err
}

func (s *Server) Stop(ctx context.Context) error {
	s.server.Close()
	return nil
}

func (s *Server) hello(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Hello world"))
}
