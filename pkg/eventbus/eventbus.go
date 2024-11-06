package eventbus

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"

	"github.com/alitto/pond/v2"
)

const (
	defaultPoolSize int = 10

	defaultMaxRetries = 1
)

type EventBus struct {
	// m is a mutex to manage the concurrent access
	m *sync.RWMutex

	// publishCh is the channel to receive the event
	publishCh chan *Event

	//
	stopCh chan struct{}

	subscribers map[string]Subscriber

	// pool is the worker pool
	pool pond.Pool

	// poolSize is the size of the worker pool
	poolSize int

	// maxRetries is the max number of retries when a notification fails
	maxRetries int
}

func New(opts ...Option) *EventBus {
	eventBus := &EventBus{
		m:         &sync.RWMutex{},
		publishCh: make(chan *Event),
		stopCh:    make(chan struct{}),
	}

	for _, opt := range opts {
		opt(eventBus)
	}

	if eventBus.poolSize == 0 {
		eventBus.poolSize = defaultPoolSize
	}
	eventBus.pool = pond.NewPool(eventBus.poolSize)

	if eventBus.maxRetries == 0 {
		eventBus.maxRetries = defaultMaxRetries
	}

	return eventBus
}

func (eb *EventBus) Subscribe(s Subscriber) error {
	eb.m.Lock()
	defer eb.m.Unlock()

	if s == nil {
		return errors.New("subscriber can not nil")
	}

	if _, has := eb.subscribers[s.ID()]; has {
		return ErrSubscriberAlreadyExist{ID: s.ID()}
	}

	eb.subscribers[s.ID()] = s

	return nil
}

func (eb *EventBus) Publish() chan<- *Event {
	return eb.publishCh
}

func (eb *EventBus) Unsubscribe(id string) {
	eb.m.Lock()
	defer eb.m.Unlock()

	delete(eb.subscribers, id)
}

func (eb *EventBus) notify(ev *Event) {
	eb.m.RLock()
	defer eb.m.RUnlock()

	for _, s := range eb.subscribers {
		retries := 0

		err := eb.pool.Go(func() {
			for retries < eb.maxRetries {
				if err := s.Handle(ev.Ctx, ev); err == nil {
					return
				}

				retries++
				log.Printf("retry %d for subscriber %s due to error: %v\n", retries, s.ID(), ev)

				if retries == eb.maxRetries {
					log.Printf("failed to handle event after %d retries for subscriber %s\n", eb.maxRetries, s.ID())
				}
			}
		})

		if err != nil {
			log.Printf("failed to notify eventbus: %v", err)
		}
	}
}

func (eb *EventBus) Start(ctx context.Context) {
	counter := 0

	for {
		select {
		case ev := <-eb.publishCh:
			counter += 1
			eb.notify(ev)

		case <-ctx.Done():
			close(eb.publishCh)

			fmt.Printf("total message published: %d\n", counter)

			eb.pool.StopAndWait()

			eb.stopCh <- struct{}{}
		}
	}
}

func (eb *EventBus) StopAndWait() <-chan struct{} {
	return eb.stopCh
}

func (eb *EventBus) SubscribersIDs() []string {
	eb.m.RLock()
	defer eb.m.RUnlock()

	var ids []string

	for k := range eb.subscribers {
		ids = append(ids, k)
	}

	return ids
}

type Option func(*EventBus)

func WithPoolSize(size int) Option {
	return func(eb *EventBus) {
		eb.poolSize = size
	}
}

func WithSubscribers(subscribers ...Subscriber) Option {
	return func(eb *EventBus) {
		eventBusSubscribers := make(map[string]Subscriber)

		for _, s := range subscribers {
			eventBusSubscribers[s.ID()] = s
		}

		eb.subscribers = eventBusSubscribers
	}
}

func WithRetries(retries int) Option {
	return func(eb *EventBus) {
		eb.maxRetries = retries
	}
}
