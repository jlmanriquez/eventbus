package main

import (
	"context"
	"fmt"
	"github.com/jlmanriquez/eventbus/pkg/eventbus"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type MySubscriber struct {
	IDValue string
}

func (s *MySubscriber) ID() string {
	return s.IDValue
}

func (s *MySubscriber) Handle(ctx context.Context, ev *eventbus.Event) error {
	if ev.Topic != "topic1" {
		return nil
	}

	select {
	case <-ctx.Done():
		log.Printf("subscriber %s: context canceled", s.ID())
		return ctx.Err()
	case <-time.After(2 * time.Second):
		log.Printf("subscriber %s: Event handled, Topic: %s, Payload: %v", s.ID(), ev.Topic, ev.Payload)
		return nil
	}
}

func main() {
	subscriber1 := &MySubscriber{IDValue: "subscriber1"}
	subscriber2 := &MySubscriber{IDValue: "subscriber2"}

	bus := eventbus.New(
		eventbus.WithPoolSize(5),
		eventbus.WithRetries(3),
		eventbus.WithSubscribers(subscriber1, subscriber2),
	)

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		bus.Start(ctx)
	}()

	stopCh := make(chan os.Signal, 1)
	signal.Notify(stopCh, syscall.SIGINT, syscall.SIGTERM)

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

	sigReceived := <-stopCh
	log.Printf("Received signal: %v", sigReceived)

	log.Println("Initiating graceful shutdown...")

	cancel()

	<-bus.StopAndWait()

	log.Println("EventBus stopped successfully")
}
