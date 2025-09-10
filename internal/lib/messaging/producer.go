package messaging

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/streadway/amqp"
)

// TODO: we should retry to publish failed messages for a given amount of times

// PublishWithRoutingKey publishes a message with a specific routing key
func (r *RabbitMQConnection) Publish(ctx context.Context, msg Message) error {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if r.closed || r.channel == nil {
		return errors.New("connection is closed")
	}

	if msg.RoutingKey == "" {
		return errors.New("routing key is required")
	}

	if msg.Timestamp.IsZero() {
		msg.Timestamp = time.Now()
	}

	// Create AMQP message
	amqpMsg := amqp.Publishing{
		Body:         msg.Body,
		Timestamp:    msg.Timestamp,
		MessageId:    msg.RequestId,
		DeliveryMode: amqp.Persistent, // Make messages persistent
	}

	return r.channel.Publish(
		r.config.Exchange, // exchange
		msg.RoutingKey,    // routing key
		false,             // mandatory
		false,             // immediate
		amqpMsg,
	)
}

// PublishWithConfirm publishes a message with publisher confirmation
func (r *RabbitMQConnection) PublishWithConfirm(ctx context.Context, msg Message) error {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if r.closed || r.channel == nil {
		return fmt.Errorf("connection is closed")
	}

	// Enable publisher confirms
	if err := r.channel.Confirm(false); err != nil {
		return fmt.Errorf("failed to enable publisher confirms: %w", err)
	}

	confirms := r.channel.NotifyPublish(make(chan amqp.Confirmation, 1))

	// Publish the message
	if err := r.Publish(ctx, msg); err != nil {
		return err
	}

	// Wait for confirmation
	select {
	case confirm := <-confirms:
		if !confirm.Ack {
			return fmt.Errorf("message was not acknowledged by broker")
		}
		return nil
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(30 * time.Second):
		return fmt.Errorf("timeout waiting for publisher confirmation")
	}
}
