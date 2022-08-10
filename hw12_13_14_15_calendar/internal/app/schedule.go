package app

import (
	"context"
	"fmt"
	"time"

	"github.com/YuriyNazarov/otus_home_work/hw12_13_14_15_calendar/internal/storage"
)

type Scheduler struct {
	logg     Logger
	storage  storage.NotificationsStorage
	notifier Notifier
}

type Notifier interface {
	Add(ctx context.Context, msg Reminder) error
}

type Reminder struct {
	EventID   string
	OwnerID   int
	Title     string
	StartTime time.Time
}

func (r Reminder) String() string {
	return fmt.Sprintf("Scheduled event \"%s\" starts at %s", r.Title, r.StartTime.Format("2006-01-02 15:04"))
}

func NewScheduler(logger Logger, storage storage.NotificationsStorage, notifier Notifier) *Scheduler {
	return &Scheduler{
		logg:     logger,
		storage:  storage,
		notifier: notifier,
	}
}

func (s *Scheduler) Notify(ctx context.Context) {
	events, err := s.storage.GetEventsToNotify()
	if err != nil {
		s.logg.Error(fmt.Sprintf("Failed to get events for notification: %s", err))
	}
	if len(events) > 0 {
		var (
			reminder Reminder
			err      error
		)
		for i := 0; i < len(events); i++ {
			reminder = Reminder{
				EventID:   events[i].ID,
				OwnerID:   events[i].OwnerID,
				Title:     events[i].Title,
				StartTime: events[i].Start,
			}
			err = s.notifier.Add(ctx, reminder)
			if err != nil {
				s.logg.Error(fmt.Sprintf("failed notification: %s", err))
				continue
			}
			err = s.storage.SetNotifiedFlag(events[i])
			if err != nil {
				s.logg.Error(fmt.Sprintf("err on setting flag to event: %s", err))
			}
		}
	}
	s.logg.Info("notifications has been sent")
}

func (s *Scheduler) RemoveOldEvents() {
	err := s.storage.DropOldEvents()
	if err != nil {
		s.logg.Error(fmt.Sprintf("Error on deleting old events: %s", err))
	} else {
		s.logg.Info("deleted old events")
	}
}
