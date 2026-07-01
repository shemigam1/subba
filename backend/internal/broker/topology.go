package broker

import (
	amqp "github.com/rabbitmq/amqp091-go"
)

// DeclareTopology declares the full exchange/queue/binding topology.
// Safe to call on every startup — idempotent.
func DeclareTopology(ch *amqp.Channel) error {
	// ── Exchange ────────────────────────────────────────────────────────────
	if err := ch.ExchangeDeclare(
		ExchangeName, // "subba.events"
		"topic",
		true,  // durable
		false, // auto-delete
		false, // internal
		false, // no-wait
		nil,
	); err != nil {
		return err
	}

	// ── Dead-letter exchange (all DLQs route through one fanout) ────────────
	if err := ch.ExchangeDeclare(
		DlxExchangeName,
		"direct",
		true,
		false,
		false,
		false,
		nil,
	); err != nil {
		return err
	}

	// ── Declare each queue group ─────────────────────────────────────────────
	queues := []struct {
		main    string
		retry   string
		dead    string
		binding string
	}{
		{QueueInvoicing, QueueInvoicingRetry, QueueInvoicingDead, RoutingKeyPaymentSucceeded},
		{QueuePayouts, QueuePayoutsRetry, QueuePayoutsDead, RoutingKeyPaymentSucceeded},
		{QueueSubscriptionState, QueueSubscriptionStateRetry, QueueSubscriptionStateDead, RoutingKeyPaymentSucceeded},
	}

	for _, q := range queues {
		// Dead queue — terminal, no further routing
		if _, err := ch.QueueDeclare(
			q.dead,
			true,  // durable
			false, // auto-delete
			false, // exclusive
			false, // no-wait
			nil,
		); err != nil {
			return err
		}
		if err := ch.QueueBind(q.dead, q.dead, "subba.dlx", false, nil); err != nil {
			return err
		}

		// Retry queue — messages sit here for TTL then go back to main queue
		retryArgs := amqp.Table{
			"x-dead-letter-exchange":    ExchangeName,
			"x-dead-letter-routing-key": q.binding,
			"x-message-ttl":             int32(30_000), // 30s first retry; workers increment on nack
		}
		if _, err := ch.QueueDeclare(
			q.retry,
			true,
			false,
			false,
			false,
			retryArgs,
		); err != nil {
			return err
		}

		// Main queue — workers consume from here
		mainArgs := amqp.Table{
			"x-dead-letter-exchange":    "subba.dlx",
			"x-dead-letter-routing-key": q.dead,
		}
		if _, err := ch.QueueDeclare(
			q.main,
			true,
			false,
			false,
			false,
			mainArgs,
		); err != nil {
			return err
		}

		// Bind main queue to the events exchange
		if err := ch.QueueBind(q.main, q.binding, ExchangeName, false, nil); err != nil {
			return err
		}
	}

	// subscription_state also binds to subscription.renew
	if err := ch.QueueBind(
		QueueSubscriptionState,
		RoutingKeySubscriptionRenew,
		ExchangeName,
		false,
		nil,
	); err != nil {
		return err
	}

	return nil
}
