package internalhttp

import (
	"context"
	"net/http"
	"time"

	"github.com/YuriyNazarov/otus_home_work/hw12_13_14_15_calendar/internal/storage"
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

type Application interface {
	CreateEvent(
		ctx context.Context,
		title,
		description string,
		start,
		end time.Time,
		ownerID int,
		remindBefore time.Duration,
	) (string, error)
	GetByID(ID string) (storage.Event, error)
	UpdateEvent(
		ctx context.Context,
		ID,
		title,
		description string,
		start,
		end time.Time,
		remindBefore time.Duration,
	) error
	DeleteEvent(ID string) error
	GetList(day time.Time, interval string) ([]storage.Event, error)
}

func NewServer(logger Logger, app Application, addr string) *Server {
	server := &Server{
		logger: logger,
		app:    app,
		addr:   addr,
	}
	mux := NewMux(app)
	server.server = &http.Server{
		Addr:    addr,
		Handler: loggingMiddleware(mux.mux, logger),
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
