package app

import "fmt"

type Sender struct {
	source EventsSource
	logger Logger
}

type EventsSource interface {
	GetReminders() (<-chan Reminder, error)
}

func NewSender(source EventsSource, logger Logger) *Sender {
	return &Sender{
		source: source,
		logger: logger,
	}
}

func (s *Sender) Run() {
	s.logger.Info("sender is running")
	channel, err := s.source.GetReminders()
	if err != nil {
		s.logger.Error(fmt.Sprintf("failed to get events: %s", err))
		return
	}

	for notification := range channel {
		s.logger.Info(notification.String())
	}
}
