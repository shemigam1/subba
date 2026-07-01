package broker

import (
	"context"
	"encoding/json"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

// Publish publishes an event to the exchange and waits for broker confirmation.
// Returns only after RabbitMQ acks the message — guarantees at-least-once delivery.
func Publish(ctx context.Context, ch *amqp.Channel, routingKey string, event *Event) error {
	body, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal event: %w", err)
	}

	if err := ch.Confirm(false); err != nil {
		return fmt.Errorf("confirm mode: %w", err)
	}

	confirms := ch.NotifyPublish(make(chan amqp.Confirmation, 1))

	if err := ch.PublishWithContext(
		ctx,
		ExchangeName, // "subba.events"
		routingKey,
		true,  // mandatory — fail if no queue is bound
		false, // immediate — not supported in modern RabbitMQ
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent, // survive broker restart
			Body:         body,
		},
	); err != nil {
		return fmt.Errorf("publish: %w", err)
	}

	select {
	case confirm := <-confirms:
		if !confirm.Ack {
			return fmt.Errorf("broker nacked the message")
		}
		return nil
	case <-ctx.Done():
		return fmt.Errorf("publish confirm timeout: %w", ctx.Err())
	}
}
