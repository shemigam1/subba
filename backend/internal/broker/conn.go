package broker

import (
	amqp "github.com/rabbitmq/amqp091-go"
)

// Conn wraps a RabbitMQ connection and a single channel.
// One Conn per process is enough for topology + publishing.
type Conn struct {
	conn *amqp.Connection
	Ch   *amqp.Channel
}

// Connect dials RabbitMQ and opens a channel.
func Connect(url string) (*Conn, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, err
	}

	return &Conn{conn: conn, Ch: ch}, nil
}

// NewChannel opens an additional AMQP channel on the same connection. Use this
// when consumers need isolated QoS settings (one channel per consumer queue).
func (c *Conn) NewChannel() (*amqp.Channel, error) {
	return c.conn.Channel()
}

// Close shuts down the channel and connection cleanly.
func (c *Conn) Close() error {
	if err := c.Ch.Close(); err != nil {
		return err
	}
	return c.conn.Close()
}
