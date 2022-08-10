package rabbit

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/YuriyNazarov/otus_home_work/hw12_13_14_15_calendar/internal/app"
	amqp "github.com/rabbitmq/amqp091-go"
)

type Rabbit struct {
	exchange string
	queue    string
	consumer string
	channel  *amqp.Channel
	logger   app.Logger
}

func NewRabbit(
	ctx context.Context,
	dsn string,
	exchange string,
	queueName string,
	logger app.Logger,
) *Rabbit {
	conn, err := amqp.Dial(dsn)
	if err != nil {
		logger.Error(fmt.Sprintf("failed on connect to rabblitmq: %s", err))
		return nil
	}

	chanel, err := conn.Channel()
	if err != nil {
		logger.Error(fmt.Sprintf("failed on opening chanel: %s", err))
		return nil
	}
	err = chanel.ExchangeDeclare(
		exchange,
		amqp.ExchangeFanout,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		logger.Error(fmt.Sprintf("failed on creating exchange: %s", err))
		return nil
	}

	queue, err := chanel.QueueDeclare(
		queueName,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		logger.Error(fmt.Sprintf("failed on creating queue: %s", err))
		return nil
	}

	err = chanel.QueueBind(
		queue.Name,
		queue.Name,
		exchange,
		false,
		nil,
	)
	if err != nil {
		logger.Error(fmt.Sprintf("failed on binding queue: %s", err))
		return nil
	}

	go func() {
		<-ctx.Done()
		chanel.Close()
		conn.Close()
	}()

	return &Rabbit{
		exchange: exchange,
		queue:    queueName,
		consumer: "calendar-consumer",
		channel:  chanel,
		logger:   logger,
	}
}

func (q *Rabbit) Add(ctx context.Context, msg app.Reminder) error {
	payload, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	err = q.channel.PublishWithContext(
		ctx,
		q.exchange,
		q.queue,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        payload,
		})
	if err != nil {
		return fmt.Errorf("failed send notification: %w", err)
	}

	return nil
}

func (q *Rabbit) GetReminders() (<-chan app.Reminder, error) {
	ch := make(chan app.Reminder)
	deliveries, err := q.channel.Consume(
		q.queue,
		q.consumer,
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		close(ch)
		return ch, err
	}

	go func() {
		var reminder app.Reminder
		for d := range deliveries {
			err := json.Unmarshal(d.Body, &reminder)
			if err != nil {
				q.logger.Error(fmt.Sprintf("failed to unmarshal event: %s", err))
				continue
			}

			ch <- reminder

			err = d.Ack(false)
			if err != nil {
				q.logger.Error(fmt.Sprintf("failed on ack: %s", err))
			}
		}

		close(ch)
	}()
	return ch, err
}
