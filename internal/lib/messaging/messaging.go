package messaging

import (
	"encoding/json"
	"fmt"
	"time"
)

// DefaultConfig returns a default configuration for RabbitMQ
func DefaultConfig() Config {
	return Config{
		URL:            "amqp://guest:guest@localhost:5672/",
		PrefetchCount:  1,
		ReconnectDelay: 5 * time.Second,
		MaxReconnects:  10,
	}
}

// NewMessage creates a new message with JSON payload
func NewMessage(data interface{}) (Message, error) {
	body, err := json.Marshal(data)
	if err != nil {
		return Message{}, fmt.Errorf("failed to marshal JSON: %w", err)
	}

	return Message{
		Body:      body,
		Timestamp: time.Now(),
	}, nil
}

// Unmarshal parses a JSON message into the provided struct
func Unmarshal(msg Message, dest interface{}) error {
	if err := json.Unmarshal(msg.Body, dest); err != nil {
		return fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	return nil
}

// MessageBuilder helps build messages fluently
type MessageBuilder struct {
	msg Message
}

// NewMessageBuilder creates a new message builder
func NewMessageBuilder() *MessageBuilder {
	return &MessageBuilder{
		msg: Message{
			Timestamp: time.Now(),
		},
	}
}

// WithBody sets the message body
func (mb *MessageBuilder) WithBody(body []byte) *MessageBuilder {
	mb.msg.Body = body
	return mb
}

// WithJSONBody sets the message body as JSON
func (mb *MessageBuilder) WithJSONBody(data any) *MessageBuilder {
	body, _ := json.Marshal(data)
	mb.msg.Body = body
	return mb
}

// WithMessageID sets the message ID
func (mb *MessageBuilder) WithRequestId(id string) *MessageBuilder {
	mb.msg.RequestId = id
	return mb
}

// WithRoutingKey sets the routing key
func (mb *MessageBuilder) WithRoutingKey(routingKey string) *MessageBuilder {
	mb.msg.RoutingKey = routingKey
	return mb
}

// Build returns the constructed message
func (mb *MessageBuilder) Build() Message {
	return mb.msg
}
