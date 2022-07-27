package storage

import (
	"errors"
	"time"
)

type Event struct {
	ID              string
	Title           string
	Description     string
	Start           time.Time
	End             time.Time
	OwnerID         int
	RemindBefore    time.Duration
	IntRemindBefore int64
	RemindSent      bool
	RemindReceived  bool
}

var (
	ErrEventNotFound    = errors.New("event not found")
	ErrDateOverlap      = errors.New("event overlaps with another event")
	ErrEventDataMissing = errors.New("some required field of event are not filled")
	ErrConnFailed       = errors.New("database connection is dead")
	ErrMigrationFailed  = errors.New("db startup migration failed")
	ErrDBOperationFail  = errors.New("something went wrong on DB operation")
)

type EventRepository interface {
	AddEvent(event *Event) error
	UpdateEvent(event Event) error
	DeleteEvent(event Event) error
	ListEvents(from time.Time, to time.Time, ownerID int) ([]Event, error)
	GetByID(id string) (Event, error)
	Close() error
}

func (event Event) IsRequiredFilled() bool {
	return event.Title != "" && event.Start != time.Time{} && event.End != time.Time{} && event.OwnerID != 0
}

func (event Event) IsOverlapsDateRange(from time.Time, to time.Time) bool {
	return (event.Start.After(from) && event.Start.Before(to)) || // Начало события в диапазоне
		(event.End.After(from) && event.End.Before(to)) || // Конец события в диапазоне
		(event.Start.Equal(from) && event.End.Equal(to)) // Совпадает с диапазоном
}
