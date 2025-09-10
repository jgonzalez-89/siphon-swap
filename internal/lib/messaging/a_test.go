package messaging

import (
	"context"
	"cryptoswap/internal/lib/logger"
	"fmt"
	"testing"
	"time"
)

func getConn() (Connection, error) {
	logger := logger.NewLoggerFactory("test", "info").
		NewLogger("test")
	return NewConnection(logger, Config{
		URL:      "amqp://myuser:mypassword@localhost:5672/",
		Exchange: "app.events",
		Queues:   []Queue{NewQueue("app.events.q")},
	})
}

func TestProduceConsume(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	conn, err := getConn()
	if err != nil {
		t.Fatalf("Failed to get connection: %v", err)
	}
	defer conn.Close()

	// Start consumer first
	if err = conn.Consume(ctx, handler); err != nil {
		t.Fatalf("Failed to start consumer: %v", err)
	}

	// Give consumer time to start
	time.Sleep(100 * time.Millisecond)

	// Publish a few test messages with correct routing key
	for i := 0; i < 3; i++ {
		fmt.Printf("Publishing message %d\n", i+1)
		msg := NewMessageBuilder().
			WithJSONBody(map[string]any{
				"test":    "iker",
				"message": i + 1,
			}).
			WithRoutingKey("cryptoswap.test").
			WithRequestId("RQ_001").
			Build()
		if err != nil {
			t.Fatalf("Failed to create message: %v", err)
		}

		// Use routing key that matches the binding pattern "cryptoswap.*"
		if err := conn.Publish(ctx, msg); err != nil {
			t.Fatalf("Failed to publish message: %v", err)
		}
		time.Sleep(500 * time.Millisecond)
	}

	// Wait for messages to be processed
	time.Sleep(2 * time.Second)
}

func handler(ctx context.Context, msg Message) error {
	fmt.Printf("Received message: %s\n", string(msg.Body))
	fmt.Printf("Routing key: %s\n", msg.RoutingKey)
	return nil
}
