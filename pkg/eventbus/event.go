package eventbus

type Event struct {
	Topic   string
	Payload any
}
