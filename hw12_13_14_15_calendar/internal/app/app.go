package app

import (
	"context"

	"github.com/YuriyNazarov/otus_home_work/hw12_13_14_15_calendar/internal/storage"
)

type App struct { // TODO
}

type Logger interface { // TODO
}

type Storage interface {
	storage.EventRepository
}

func New(logger Logger, storage Storage) *App {
	return &App{}
}

func (a *App) CreateEvent(ctx context.Context, id, title string) error {
	// TODO
	return nil
	// return a.storage.CreateEvent(storage.Event{ID: id, Title: title})
}

// TODO
