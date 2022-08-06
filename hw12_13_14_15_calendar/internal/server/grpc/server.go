//go:generate protoc --go_out=. --go-grpc_out=. ../../../api/EventService.proto --proto_path=../../../api

package internalgrpc

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strconv"
	"time"

	"github.com/YuriyNazarov/otus_home_work/hw12_13_14_15_calendar/internal/app"
	"github.com/YuriyNazarov/otus_home_work/hw12_13_14_15_calendar/internal/storage"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	errInvalidDateTimeFormat = errors.New("failed to parse date time: expecting \"YYYY-MM-DD HH:MM\" format")
	errInvalidDateFormat     = errors.New("failed to parse date: expecting \"YYYY-MM-DD\" format")
	errInvalidDuration       = errors.New("failed to parse duration: expecting \"_h_m_s\" format")
	errDataMissing           = errors.New("some required fields are not filled")
)

type Server struct {
	UnimplementedEventServiceServer
	host   string
	port   int
	server *grpc.Server
	app    *app.App
	logg   app.Logger
}

func NewLoggingInterceptor(logger app.Logger) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (resp interface{}, err error) {
		start := time.Now()
		result, err := handler(ctx, req)
		logger.Info(fmt.Sprintf("grpc request: %v done in %s", req, time.Since(start).String()))
		return result, err
	}
}

func NewServer(logger app.Logger, app *app.App, host string, port int) *Server {
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(
			NewLoggingInterceptor(logger),
		),
	)
	s := &Server{
		host:   host,
		port:   port,
		server: grpcServer,
		app:    app,
		logg:   logger,
	}
	RegisterEventServiceServer(s.server, s)
	return s
}

func (s *Server) Start() error {
	lsn, err := net.Listen("tcp", net.JoinHostPort(s.host, strconv.Itoa(s.port)))
	if err != nil {
		return err
	}
	return s.server.Serve(lsn)
}

func (s *Server) Stop() {
	s.server.GracefulStop()
}

func (s *Server) Create(ctx context.Context, inp *Event) (*EventResponse, error) {
	title := inp.GetTitle()
	start := inp.GetStart()
	end := inp.GetEnd()
	ownerID := inp.GetOwnerId()
	if title == "" || start == "" || end == "" || ownerID == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "%s", errDataMissing)
	}
	tStart, err := time.Parse("2006-01-02 15:04", start)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "field 'start': %s", errInvalidDateTimeFormat)
	}
	tEnd, err := time.Parse("2006-01-02 15:04", end)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "field 'end': %s", errInvalidDateTimeFormat)
	}
	remind, err := time.ParseDuration(inp.GetRemindBefore())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "%s", errInvalidDuration)
	}

	id, err := s.app.CreateEvent(
		ctx,
		title,
		inp.GetDescription(),
		tStart,
		tEnd,
		int(ownerID),
		remind,
	)
	if err != nil {
		return &EventResponse{Result: &EventResponse_Error{err.Error()}},
			status.Errorf(codes.Internal, "oops: %s", err)
	}
	return &EventResponse{Result: &EventResponse_Id{id}}, nil
}

func (s *Server) Update(ctx context.Context, inp *Event) (*EventResponse, error) {
	title := inp.GetTitle()
	start := inp.GetStart()
	end := inp.GetEnd()
	id := inp.GetId()
	if title == "" || start == "" || end == "" || id == "" {
		return nil, status.Errorf(codes.InvalidArgument, "%s", errDataMissing)
	}
	tStart, err := time.Parse("2006-01-02 15:04", start)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "field 'start': %s", errInvalidDateTimeFormat)
	}
	tEnd, err := time.Parse("2006-01-02 15:04", end)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "field 'end': %s", errInvalidDateTimeFormat)
	}
	remind, err := time.ParseDuration(inp.GetRemindBefore())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "%s", errInvalidDuration)
	}
	err = s.app.UpdateEvent(
		ctx,
		id,
		title,
		inp.GetDescription(),
		tStart,
		tEnd,
		remind,
	)
	if err != nil {
		return &EventResponse{Result: &EventResponse_Error{err.Error()}},
			status.Errorf(codes.Internal, "oops: %s", err)
	}
	return &EventResponse{Result: &EventResponse_Id{id}}, nil
}

func (s *Server) Delete(ctx context.Context, inp *EventIdRequest) (*EventResponse, error) {
	id := inp.GetId()
	if id == "" {
		return nil, status.Errorf(codes.InvalidArgument, "%s: ID required", errDataMissing)
	}
	if err := s.app.DeleteEvent(id); err != nil {
		return &EventResponse{Result: &EventResponse_Error{err.Error()}},
			status.Errorf(codes.Internal, "oops: %s", err)
	}
	return &EventResponse{Result: &EventResponse_Id{id}}, nil
}

func (s *Server) GetByID(ctx context.Context, inp *EventIdRequest) (*Event, error) {
	id := inp.GetId()
	if id == "" {
		return nil, status.Errorf(codes.InvalidArgument, "%s: ID required", errDataMissing)
	}
	event, err := s.app.GetByID(id)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "oops: %s", err)
	}
	return &Event{
		Id:           event.ID,
		Title:        event.Title,
		Description:  event.Description,
		Start:        event.Start.Format("2006-01-02 15:04"),
		End:          event.End.Format("2006-01-02 15:04"),
		OwnerId:      int32(event.OwnerID),
		RemindBefore: event.RemindBefore.String(),
	}, nil
}

func (s *Server) DayEvents(ctx context.Context, inp *EventsRequest) (*EventsResponse, error) {
	date, err := time.Parse("2006-01-02", inp.GetDate())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "%s", errInvalidDateFormat)
	}
	events, err := s.app.GetList(date, "day")
	if err != nil {
		return nil, status.Errorf(codes.Internal, "oops: %s", err)
	}
	return &EventsResponse{Events: convertEvents(events)}, nil
}

func (s *Server) WeekEvents(ctx context.Context, inp *EventsRequest) (*EventsResponse, error) {
	date, err := time.Parse("2006-01-02", inp.GetDate())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "%s", errInvalidDateFormat)
	}
	events, err := s.app.GetList(date, "week")
	if err != nil {
		return nil, status.Errorf(codes.Internal, "oops: %s", err)
	}
	return &EventsResponse{Events: convertEvents(events)}, nil
}

func (s *Server) MonthEvents(ctx context.Context, inp *EventsRequest) (*EventsResponse, error) {
	date, err := time.Parse("2006-01-02", inp.GetDate())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "%s", errInvalidDateFormat)
	}
	events, err := s.app.GetList(date, "month")
	if err != nil {
		return nil, status.Errorf(codes.Internal, "oops: %s", err)
	}
	return &EventsResponse{Events: convertEvents(events)}, nil
}

func convertEvents(events []storage.Event) []*Event {
	var (
		result []*Event
		gEvent Event
	)

	for i := 0; i < len(events); i++ {
		gEvent = Event{
			Id:           events[i].ID,
			Title:        events[i].Title,
			Description:  events[i].Description,
			Start:        events[i].Start.Format("2006-01-02 15:04"),
			End:          events[i].End.Format("2006-01-02 15:04"),
			OwnerId:      int32(events[i].OwnerID),
			RemindBefore: events[i].RemindBefore.String(),
		}
		result = append(result, &gEvent)
	}
	return result
}
