# EventBus in Go

This is an EventBus implementation in Go that allows for publishing and subscribing to events, with support for concurrency via a worker pool, subscriber management, and **graceful shutdown**. The goal is to provide an efficient and scalable infrastructure for handling events in concurrent applications.

## Features

- **Event publishing and subscription**: Subscribers can register to listen for and handle published events.
- **Concurrency**: Uses a worker pool to process events concurrently.
- **Graceful Shutdown**: The system can shut down cleanly, waiting for workers to finish their tasks.
- **Flexible configuration**: You can configure the size of the worker pool and the initial set of subscribers via customizable options.

## Installation

To install this EventBus, you can clone the repository and add it as a dependency to your Go project.

1. Clone the repository:
    ```bash
       git clone https://github.com/yourusername/eventbus.git
    ```

2. If you're using Go Modules (recommended for modern Go projects), you can add it to your project as a dependency:
    ```bash
        go mod tidy
    ```

## Usage

Here’s a basic example of how to use the EventBus in your application:

```go
    package main

import (
	"context"
	"fmt"
	"github.com/jlmanriquez/eventbus/pkg/eventbus"
	"log"
	"time"
)

type MySubscriber struct {
	IDValue string
}

func (s *MySubscriber) ID() string {
	return s.IDValue
}

func (s *MySubscriber) Handle(ctx context.Context, ev *eventbus.Event) error {
	select {
	case <-ctx.Done():
		log.Printf("Subscriber %s: context canceled", s.ID())
		return ctx.Err()
	case <-time.After(2 * time.Second): // Simulate work taking time
		log.Printf("Subscriber %s: Event handled, Topic: %s, Payload: %v", s.ID(), ev.Topic, ev.Payload)
		return nil
	}
}

func (s *MySubscriber) ShouldHandle(ev *eventbus.Event) bool {
	// Simulate this subscriber only handling certain topics
	return ev.Topic == "topic1"
}

func main() {
	// Create an EventBus with a pool size of 5 workers
	bus := eventbus.New(
		eventbus.WithPoolSize(5), // Configure the worker pool size
	)

	// Create and subscribe the subscribers
	subscriber1 := &MySubscriber{IDValue: "subscriber1"}
	subscriber2 := &MySubscriber{IDValue: "subscriber2"}
	bus.Subscribe(subscriber1)
	bus.Subscribe(subscriber2)

	// Create a context and start the EventBus
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		bus.Start(ctx)
	}()

	// Publish events
	go func() {
		publishCh := bus.Publish()
		for i := 1; i <= 5; i++ {
			event := &eventbus.Event{
				Ctx:     ctx,
				Topic:   "topic1",
				Payload: fmt.Sprintf("Message %d", i),
			}
			publishCh <- event
			log.Printf("Published event: %v", event)
			time.Sleep(1 * time.Second)
		}
	}()

	// Wait for interruption signals for graceful shutdown
	// Code to listen for signals omitted for simplicity

	// Cancel the context when needed
	cancel()
	<-bus.StopAndWait()
	log.Println("EventBus stopped successfully")
}
```

## Configuration Options:

* WithPoolSize(size int): Configures the size of the worker pool.
* WithSubscribers(subscribers ...Subscriber): Registers initial subscribers to the EventBus.

[//]: # (## Contributing)

[//]: # ()
[//]: # (Contributions are welcome. If you have suggestions, bug fixes, or improvements, feel free to open an issue or submit a pull request.)

---

Project maintained by jlmanriquez. License [MIT](LICENSE)


