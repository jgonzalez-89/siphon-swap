package messaging

import (
	"context"
	"cryptoswap/internal/lib/logger"
	"fmt"
	"sync"
	"time"

	"github.com/streadway/amqp"
)

// RabbitMQConnection manages RabbitMQ connection and channels
type RabbitMQConnection struct {
	conn       *amqp.Connection
	channel    *amqp.Channel
	mu         sync.RWMutex
	closed     bool
	reconnects int
	logger     logger.Logger
	config     Config
}

// NewConnection creates a new RabbitMQ connection
func NewConnection(logger logger.Logger, config Config) (*RabbitMQConnection, error) {
	if config.ReconnectDelay == 0 {
		config.ReconnectDelay = 5 * time.Second
	}
	if config.MaxReconnects == 0 {
		config.MaxReconnects = 10
	}
	if config.PrefetchCount == 0 {
		config.PrefetchCount = 1
	}
	rmq := &RabbitMQConnection{
		logger: logger,
		config: config,
	}

	if err := rmq.connect(); err != nil {
		return nil, fmt.Errorf("failed to establish initial connection: %w", err)
	}

	return rmq, nil
}

// connect establishes connection to RabbitMQ
func (r *RabbitMQConnection) connect() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	ctx := context.Background()

	var err error
	r.conn, err = amqp.Dial(r.config.URL)
	if err != nil {
		return fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	r.channel, err = r.conn.Channel()
	if err != nil {
		r.conn.Close()
		return fmt.Errorf("failed to open channel: %w", err)
	}

	// Set QoS
	if err := r.channel.Qos(r.config.PrefetchCount, 0, false); err != nil {
		r.channel.Close()
		r.conn.Close()
		return fmt.Errorf("failed to set QoS: %w", err)
	}

	r.closed = false
	r.reconnects = 0

	// Setup connection close handler
	go r.handleReconnect()

	// Setup exchange and queue if configured
	if err := r.setupTopology(); err != nil {
		r.channel.Close()
		r.conn.Close()
		return fmt.Errorf("failed to setup topology: %w", err)
	}

	r.logger.Infof(ctx, "Connected to RabbitMQ: %s", r.config.URL)
	return nil
}

// setupTopology declares exchange and queue if they are configured
func (r *RabbitMQConnection) setupTopology() error {
	// Skip topology setup - assume exchange and queue are already configured
	// via RabbitMQ definitions file or external setup
	return nil
}

// handleReconnect handles automatic reconnection
func (r *RabbitMQConnection) handleReconnect() {
	ctx := context.Background()

	for {
		reason, ok := <-r.conn.NotifyClose(make(chan *amqp.Error))
		if !ok {
			break
		}

		r.mu.Lock()
		if r.closed {
			r.mu.Unlock()
			break
		}
		r.mu.Unlock()

		r.logger.Warningf(ctx, "RabbitMQ connection closed: %v. Attempting to reconnect...", reason)

		for r.reconnects < r.config.MaxReconnects {
			r.reconnects++
			r.logger.Infof(ctx, "Reconnection attempt %d/%d", r.reconnects, r.config.MaxReconnects)

			if err := r.connect(); err != nil {
				r.logger.Errorf(ctx, "Reconnection failed: %v", err)
				time.Sleep(r.config.ReconnectDelay)
				continue
			}

			r.logger.Info(ctx, "Successfully reconnected to RabbitMQ")
			break
		}

		if r.reconnects >= r.config.MaxReconnects {
			r.logger.Error(ctx, "Max reconnection attempts reached. Giving up.")
			break
		}
	}
}

// IsConnected returns true if connection is active
func (r *RabbitMQConnection) IsConnected() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.conn != nil && !r.conn.IsClosed() && !r.closed
}

// Reconnect manually triggers a reconnection
func (r *RabbitMQConnection) Reconnect() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.conn != nil && !r.conn.IsClosed() {
		r.conn.Close()
	}

	return r.connect()
}

// Close closes the connection
func (r *RabbitMQConnection) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	ctx := context.Background()
	r.closed = true

	if r.channel != nil {
		if err := r.channel.Close(); err != nil {
			r.logger.Errorf(ctx, "Error closing channel: %v", err)
		}
	}

	if r.conn != nil {
		if err := r.conn.Close(); err != nil {
			r.logger.Errorf(ctx, "Error closing connection: %v", err)
			return err
		}
	}

	r.logger.Info(ctx, "RabbitMQ connection closed")
	return nil
}
