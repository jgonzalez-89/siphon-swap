# RabbitMQ Messaging Library

A comprehensive RabbitMQ client library for Go with producer and consumer functionality, automatic reconnection, retry logic, and health checking.

**Note**: This library assumes that exchanges, queues, and bindings are pre-configured in RabbitMQ. It focuses on message publishing and consumption without managing infrastructure.

## Features

- **Producer & Consumer**: Full-featured message publishing and consumption
- **Automatic Reconnection**: Handles connection drops with configurable retry logic
- **Message Retry**: Built-in retry mechanism for failed message processing
- **Dead Letter Support**: Send failed messages to dead letter exchanges
- **Health Checking**: Monitor connection health
- **Batch Publishing**: Publish multiple messages efficiently
- **Publisher Confirms**: Ensure message delivery with confirmations
- **Fluent Message Builder**: Easy message construction
- **JSON Support**: Built-in JSON marshaling/unmarshaling

## Quick Start

### Basic Producer

```go
package main

import (
    "context"
    "your-project/internal/lib/messaging"
    "your-project/internal/lib/logger"
)

func main() {
    // Create logger
    loggerFactory := logger.NewLoggerFactory("messaging", "info")
    log := loggerFactory.NewLogger("producer")

    // Configure connection
    config := messaging.DefaultConfig(log)
    config.Exchange = "events"
    config.RoutingKey = "user.created"

    // Create producer
    producer, err := messaging.NewProducer(config)
    if err != nil {
        log.Fatalf(context.Background(), "Failed to create producer: %v", err)
    }
    defer producer.Close()

    // Create and publish message
    msg, err := messaging.NewJSONMessage(map[string]interface{}{
        "user_id": "123",
        "action":  "created",
    })
    if err != nil {
        log.Fatalf(context.Background(), "Failed to create message: %v", err)
    }

    ctx := context.Background()
    if err := producer.Publish(ctx, msg); err != nil {
        log.Fatalf(ctx, "Failed to publish message: %v", err)
    }

    log.Info(ctx, "Message published!")
}
```

### Basic Consumer

```go
package main

import (
    "context"
    "your-project/internal/lib/messaging"
    "your-project/internal/lib/logger"
)

func main() {
    // Create logger
    loggerFactory := logger.NewLoggerFactory("messaging", "info")
    log := loggerFactory.NewLogger("consumer")

    // Configure connection
    config := messaging.DefaultConfig(log)
    config.Exchange = "events"
    config.QueueName = "user_events"
    config.RoutingKey = "user.*"

    // Create consumer
    consumer, err := messaging.NewConsumer(config)
    if err != nil {
        log.Fatalf(context.Background(), "Failed to create consumer: %v", err)
    }
    defer consumer.Close()

    // Define message handler
    handler := func(ctx context.Context, msg messaging.Message) error {
        var data map[string]interface{}
        if err := messaging.ParseJSONMessage(msg, &data); err != nil {
            return err
        }

        log.Infof(ctx, "Received: %+v", data)
        return nil
    }

    // Start consuming
    ctx := context.Background()
    if err := consumer.Consume(ctx, handler); err != nil {
        log.Fatalf(ctx, "Failed to start consuming: %v", err)
    }

    // Keep running
    select {}
}
```

## Configuration

```go
// Create logger first
loggerFactory := logger.NewLoggerFactory("messaging", "info")
log := loggerFactory.NewLogger("rabbitmq")

config := messaging.Config{
    URL:             "amqp://user:pass@localhost:5672/",
    Exchange:        "my_exchange",
    RoutingKey:      "my.routing.key",
    QueueName:       "my_queue",
    PrefetchCount:   10,             // Number of unacked messages
    ReconnectDelay:  5 * time.Second,
    MaxReconnects:   10,
    Logger:          log,            // Required logger instance
}
```

## Advanced Usage

### Message Builder

```go
msg := messaging.NewMessageBuilder().
    WithJSONBody(map[string]interface{}{
        "event": "user_login",
        "user_id": "123",
    }).
    WithHeader("priority", "high").
    WithPriority(5).
    WithMessageID("msg-123").
    WithRoutingKey("user.login").
    Build()

producer.Publish(ctx, msg)
```

### Consumer with Retry

```go
handler := func(ctx context.Context, msg messaging.Message) error {
    // Your processing logic that might fail
    return processMessage(msg)
}

maxRetries := 3
retryDelay := 2 * time.Second

conn.ConsumeWithRetry(ctx, handler, maxRetries, retryDelay)
```

### Publisher Confirms

```go
// Ensure message delivery
if err := conn.PublishWithConfirm(ctx, msg); err != nil {
    log.Printf("Message was not delivered: %v", err)
}
```

### Batch Publishing

```go
var messages []messaging.Message
for i := 0; i < 100; i++ {
    msg, _ := messaging.NewJSONMessage(map[string]interface{}{
        "id": i,
        "data": fmt.Sprintf("Message %d", i),
    })
    messages = append(messages, msg)
}

conn.PublishBatch(ctx, messages)
```

### Health Checking

```go
healthChecker := messaging.NewHealthChecker(conn)

// Simple check
if healthChecker.IsHealthy() {
    log.Println("Connection is healthy")
}

// Comprehensive check
if err := healthChecker.HealthCheck(ctx); err != nil {
    log.Printf("Health check failed: %v", err)
}
```

### Dead Letter Handling

```go
deadLetterExchange := "failed_messages"

handler := func(ctx context.Context, msg messaging.Message) error {
    // Processing logic that might fail
    return processMessage(msg)
}

// Failed messages will be sent to dead letter exchange
conn.ConsumeWithDeadLetter(ctx, handler, deadLetterExchange)
```

## Error Handling

The library provides comprehensive error handling:

- **Connection errors**: Automatic reconnection with exponential backoff
- **Publishing errors**: Detailed error messages with context
- **Consumption errors**: Message rejection and requeuing
- **JSON errors**: Clear parsing error messages

## Best Practices

1. **Always close connections**: Use `defer conn.Close()`
2. **Handle context cancellation**: Respect context timeouts
3. **Use publisher confirms**: For critical messages
4. **Set appropriate prefetch**: Balance throughput and memory
5. **Monitor health**: Implement health checks in your services
6. **Use dead letters**: For failed message handling
7. **Configure retries**: Set reasonable retry limits

## Dependencies

```bash
go get github.com/streadway/amqp
```

## Examples

See `examples.go` for complete working examples of all features.
