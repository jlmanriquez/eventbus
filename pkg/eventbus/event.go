package eventbus

import "context"

type Event struct {
	Ctx     context.Context
	Topic   string
	Payload any
}
