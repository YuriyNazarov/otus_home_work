package memorystorage

import (
	"sync"
	"time"

	"github.com/YuriyNazarov/otus_home_work/hw12_13_14_15_calendar/internal/logger"
	"github.com/YuriyNazarov/otus_home_work/hw12_13_14_15_calendar/internal/storage"
	uuid "github.com/satori/go.uuid"
)

type Storage struct {
	events map[string]storage.Event
	mu     sync.RWMutex
	log    logger.Logger
}

func (s *Storage) AddEvent(event *storage.Event) error {
	if !event.IsRequiredFilled() {
		return storage.ErrEventDataMissing
	}
	exEvents, err := s.ListEvents(event.Start, event.End, event.OwnerID)
	if err != nil {
		return err
	}
	if len(exEvents) > 0 {
		return storage.ErrDateOverlap
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	event.ID = uuid.NewV4().String()
	s.events[event.ID] = *event
	s.log.Debug("stored event id=" + event.ID)
	return nil
}

func (s *Storage) UpdateEvent(event storage.Event) error {
	if !event.IsRequiredFilled() {
		return storage.ErrEventDataMissing
	}
	events, _ := s.listEventsExclude(event.Start, event.End, event.OwnerID, event.ID)
	if len(events) > 0 {
		return storage.ErrDateOverlap
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.events[event.ID] = event
	s.log.Debug("updated event id=" + event.ID)
	return nil
}

func (s *Storage) DeleteEvent(event storage.Event) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.events, event.ID)
	s.log.Debug("deleted event id=" + event.ID)
	return nil
}

func (s *Storage) ListEvents(from time.Time, to time.Time, ownerID int) ([]storage.Event, error) {
	return s.listEventsExclude(from, to, ownerID, "")
}

func (s *Storage) listEventsExclude(
	from time.Time,
	to time.Time,
	ownerID int,
	excludeID string,
) ([]storage.Event, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	var events []storage.Event
	for _, event := range s.events {
		if event.IsOverlapsDateRange(from, to) && event.OwnerID == ownerID && event.ID != excludeID {
			events = append(events, event)
		}
	}
	return events, nil
}

func (s *Storage) GetByID(id string) (storage.Event, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	event, ok := s.events[id]
	if !ok {
		return storage.Event{}, storage.ErrEventNotFound
	}
	return event, nil
}

func New(log logger.Logger) *Storage {
	return &Storage{
		events: make(map[string]storage.Event),
		mu:     sync.RWMutex{},
		log:    log,
	}
}

func (s *Storage) Close() error {
	s.mu.Lock()
	s.events = make(map[string]storage.Event)
	s.mu.Unlock()
	return nil
}
