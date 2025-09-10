package messaging

import (
	"context"
	"fmt"
	"time"

	"github.com/streadway/amqp"
)

// Consume starts consuming messages from the configured queue
func (r *RabbitMQConnection) Consume(ctx context.Context, handler Handler) error {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if r.closed || r.channel == nil {
		return fmt.Errorf("connection is closed")
	}

	for _, queue := range r.config.Queues {
		if queue.Name == "" {
			return fmt.Errorf("queue name is required for consuming")
		}

		if err := r.consumeFromQueue(ctx, queue.Name, handler); err != nil {
			return err
		}
	}
	return nil
}

// handleMessage processes a single message
func (r *RabbitMQConnection) handleMessage(ctx context.Context, delivery amqp.Delivery, handler Handler) {
	// Create our message struct
	msg := Message{
		Body:       delivery.Body,
		Timestamp:  delivery.Timestamp,
		RoutingKey: delivery.RoutingKey,
	}

	// Create a timeout context for message processing
	msgCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// Process the message
	if handerErr := handler(msgCtx, msg); handerErr != nil {
		r.logger.Errorf(ctx, "Error processing message: %v", handerErr)

		// Reject the message and requeue it
		if nackErr := delivery.Nack(false, true); nackErr != nil {
			r.logger.Errorf(ctx, "Error nacking message: %v", nackErr)
		}
		return
	}

	// Acknowledge the message
	if ackErr := delivery.Ack(false); ackErr != nil {
		r.logger.Errorf(ctx, "Error acking message: %v", ackErr)
	}
}

// consumeFromQueue starts consuming from a single queue (internal helper)
func (r *RabbitMQConnection) consumeFromQueue(ctx context.Context, queueName string, handler Handler) error {
	msgs, err := r.channel.Consume(
		queueName, // queue
		"",        // consumer tag (auto-generated)
		false,     // auto-ack (we'll ack manually)
		false,     // exclusive (changed to false for multi-queue)
		false,     // no-local
		true,      // no-wait
		nil,       // args
	)
	if err != nil {
		return fmt.Errorf("failed to start consuming from queue %s: %w", queueName, err)
	}

	r.logger.Infof(ctx, "Started consuming from queue: %s", queueName)

	// Process messages in a separate goroutine for each queue
	go func(queueName string, msgs <-chan amqp.Delivery) {
		for {
			select {
			case <-ctx.Done():
				r.logger.Infof(ctx, "Consumer context cancelled, stopping consumer for queue: %s", queueName)
				return
			case msg, ok := <-msgs:
				if !ok {
					r.logger.Infof(ctx, "Message channel closed for queue: %s", queueName)
					return
				}
				r.handleMessage(ctx, msg, handler)
			}
		}
	}(queueName, msgs)
	return nil
}
