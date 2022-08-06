package app

import (
	"context"
	"fmt"
	"time"

	"github.com/YuriyNazarov/otus_home_work/hw12_13_14_15_calendar/internal/storage"
)

type App struct {
	logg    Logger
	storage Storage
}

type Logger interface {
	Debug(msg string)
	Info(msg string)
	Warn(msg string)
	Error(msg string)
}

type Storage interface {
	AddEvent(event *storage.Event) error
	UpdateEvent(event storage.Event) error
	DeleteEvent(event storage.Event) error
	ListEvents(from time.Time, to time.Time, ownerID int) ([]storage.Event, error)
	GetByID(id string) (storage.Event, error)
	Close() error
}

func New(logger Logger, storage storage.EventRepository) *App {
	return &App{
		logg:    logger,
		storage: storage,
	}
}

func (a *App) CreateEvent(
	ctx context.Context,
	title,
	description string,
	start,
	end time.Time,
	ownerID int,
	remindBefore time.Duration,
) (string, error) {
	event := storage.Event{
		Title:           title,
		Description:     description,
		Start:           start,
		End:             end,
		OwnerID:         ownerID,
		RemindBefore:    remindBefore,
		IntRemindBefore: remindBefore.Nanoseconds(),
	}
	err := a.storage.AddEvent(&event)
	if err != nil {
		return "", fmt.Errorf("failed to create event: %w", err)
	}
	return event.ID, nil
}

func (a *App) GetByID(id string) (storage.Event, error) {
	event, err := a.storage.GetByID(id)
	if err != nil {
		return storage.Event{}, fmt.Errorf("event not found: %w", err)
	}
	return event, nil
}

func (a *App) UpdateEvent(
	ctx context.Context,
	id,
	title,
	description string,
	start,
	end time.Time,
	remindBefore time.Duration,
) error {
	event, err := a.storage.GetByID(id)
	if err != nil {
		return fmt.Errorf("requested to update event not found: %w", err)
	}
	event.Title = title
	event.Description = description
	event.Start = start
	event.End = end
	event.RemindBefore = remindBefore
	event.IntRemindBefore = remindBefore.Nanoseconds()
	err = a.storage.UpdateEvent(event)
	if err != nil {
		return fmt.Errorf("failed to update event: %w", err)
	}
	return nil
}

func (a *App) DeleteEvent(id string) error {
	event, err := a.storage.GetByID(id)
	if err != nil {
		return fmt.Errorf("requested to delete event not found: %w", err)
	}
	err = a.storage.DeleteEvent(event)
	if err != nil {
		return fmt.Errorf("failed to delete event: %w", err)
	}
	return nil
}

func (a *App) GetList(day time.Time, interval string) ([]storage.Event, error) {
	day = day.Truncate(24 * time.Hour)
	var endDay time.Time
	switch interval {
	case "day":
		endDay = day.Add(24 * time.Hour)
	case "week":
		endDay = day.AddDate(0, 0, 7)
	case "month":
		endDay = day.AddDate(0, 1, 0)
	}
	events, err := a.storage.ListEvents(day, endDay, 0)
	if err != nil {
		return []storage.Event{}, fmt.Errorf("failed to get list of events: %w", err)
	}
	return events, nil
}
