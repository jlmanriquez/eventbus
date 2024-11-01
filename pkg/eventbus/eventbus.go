package eventbus

import (
	"context"
	"errors"
	"maps"
	"sync"
)

var ErrSubscriberAlreadyExist = errors.New("already exist")

type EventBus struct {
	m           *sync.RWMutex
	subscribers map[string]Subscriber
}

func New() *EventBus {
	return &EventBus{
		m:           &sync.RWMutex{},
		subscribers: map[string]Subscriber{},
	}
}

func NewWithSubscribers(subscribers ...Subscriber) *EventBus {
	eventBusSubscribers := make(map[string]Subscriber, 0)

	for _, s := range subscribers {
		eventBusSubscribers[s.ID()] = s
	}

	return &EventBus{
		m:           &sync.RWMutex{},
		subscribers: eventBusSubscribers,
	}
}

func (eb *EventBus) Subscribe(s Subscriber) error {
	eb.m.Lock()
	defer eb.m.Unlock()

	if s == nil {
		return errors.New("subscriber can not nil")
	}

	if _, has := eb.subscribers[s.ID()]; has {
		return ErrSubscriberAlreadyExist
	}

	eb.subscribers[s.ID()] = s

	return nil
}

func (eb *EventBus) Unsubscribe(id string) {
	eb.m.Lock()
	defer eb.m.Unlock()

	delete(eb.subscribers, id)
}

func (eb *EventBus) Publish(ctx context.Context, topic string, ev Event) {
	for _, s := range eb.subscribers {
		if s.ShouldHandle(ev) {
			go s.Handle(ctx, ev)
		}
	}
}

func (eb *EventBus) SubscribersIDs() []string {
	eb.m.RLock()
	defer eb.m.RUnlock()

	ids := []string{}

	for k := range maps.Keys(eb.subscribers) {
		ids = append(ids, k)
	}

	return ids
}
