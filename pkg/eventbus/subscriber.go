package eventbus

import "context"

type Subscriber interface {
	ID() string
	Handle(context.Context, *Event) error
}
