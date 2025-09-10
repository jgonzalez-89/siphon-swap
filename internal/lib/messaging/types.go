package messaging

import (
	"context"
	"cryptoswap/internal/config"
	"time"
)

func NewConfig(config config.RabbitMQ, queues []Queue) Config {
	url := "amqp://" + config.User + ":" + config.Password + "@" + config.Host + ":" + config.Port + "/"
	return Config{
		URL:            url,
		Exchange:       config.Exchange,
		Queues:         queues,
		PrefetchCount:  config.GetPrefetchCount(),
		ReconnectDelay: time.Duration(config.GetReconnectDelay()) * time.Second,
		MaxReconnects:  config.GetMaxReconnects(),
	}
}

// Config holds RabbitMQ connection configuration
type Config struct {
	URL            string
	Exchange       string
	Queues         []Queue
	PrefetchCount  int
	ReconnectDelay time.Duration
	MaxReconnects  int
}

type Queue struct {
	Name      string
	AutoAck   bool
	Exclusive bool
	NoLocal   bool
	NoWait    bool
}

func NewQueue(name string) Queue {
	return Queue{
		Name:      name,
		AutoAck:   false,
		Exclusive: false,
		NoLocal:   false,
		NoWait:    true,
	}
}

// Message represents a message to be published or consumed
type Message struct {
	Timestamp  time.Time
	RequestId  string
	RoutingKey string
	Body       []byte
}

type ConsumerBuilder interface {
	Build() Handler
}

// Handler defines the function signature for message handlers
type Handler func(ctx context.Context, msg Message) error

// Publisher interface for message publishing
type Publisher interface {
	Publish(ctx context.Context, msg Message) error
	Close() error
}

type Consumer interface {
	Consume(ctx context.Context, handler Handler) error
	Close() error
}

// Connection interface for RabbitMQ connection management
type Connection interface {
	IsConnected() bool
	Publisher
	Consumer
}
