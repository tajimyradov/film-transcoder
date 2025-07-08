package broker

import (
	"fmt"

	"github.com/tajimyradov/transcoder/models"

	"github.com/streadway/amqp"
)

type RabbitConsumer struct {
	Connection *amqp.Connection
	Channel    *amqp.Channel
	Messages   <-chan amqp.Delivery
}

func NewRabbitMQ(cfg models.RabbitMQ) (*RabbitConsumer, error) {
	amqpURL := fmt.Sprintf("amqp://%s:%s@%s/", cfg.Username, cfg.Password, cfg.Url)

	conn, err := amqp.Dial(amqpURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to open a channel: %w", err)
	}

	queue, err := ch.QueueDeclare(
		cfg.QueueName,
		true,  // durable
		false, // autoDelete
		false, // exclusive
		false, // noWait
		nil,
	)
	if err != nil {
		ch.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to declare queue: %w", err)
	}

	// Limit to 1 unacked message at a time
	if err := ch.Qos(1, 0, false); err != nil {
		ch.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to set QoS: %w", err)
	}

	msgs, err := ch.Consume(
		queue.Name,
		"",    // consumer tag
		false, // auto-ack
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,   // args
	)
	if err != nil {
		ch.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to register consumer: %w", err)
	}

	return &RabbitConsumer{
		Connection: conn,
		Channel:    ch,
		Messages:   msgs,
	}, nil
}
