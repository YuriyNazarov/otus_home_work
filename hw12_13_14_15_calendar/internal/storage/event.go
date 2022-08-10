package storage

import (
	"errors"
	"time"
)

type Event struct {
	ID              string        `json:"id"`
	Title           string        `json:"title"`
	Description     string        `json:"description"`
	Start           time.Time     `json:"start"`
	End             time.Time     `json:"end"`
	OwnerID         int           `json:"ownerId"`
	RemindBefore    time.Duration `json:"remindBefore"`
	IntRemindBefore int64
	RemindSent      bool `json:"remindSent"`
	RemindReceived  bool `json:"remindReceived"`
}

var (
	ErrEventNotFound    = errors.New("event not found")
	ErrDateOverlap      = errors.New("event overlaps with another event")
	ErrEventDataMissing = errors.New("some required field of event are not filled")
	ErrConnFailed       = errors.New("database connection is dead")
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

type NotificationsStorage interface {
	GetEventsToNotify() ([]Event, error)
	DropOldEvents() error
	SetNotifiedFlag(event Event) error
}

func (event Event) IsRequiredFilled() bool {
	return event.Title != "" && event.Start != time.Time{} && event.End != time.Time{} && event.OwnerID != 0
}

func (event Event) IsOverlapsDateRange(from time.Time, to time.Time) bool {
	return (event.Start.After(from) && event.Start.Before(to)) || // Начало события в диапазоне
		(event.End.After(from) && event.End.Before(to)) || // Конец события в диапазоне
		(event.Start.Equal(from) && event.End.Equal(to)) // Совпадает с диапазоном
}
