package eventbus

import "fmt"

type ErrSubscriberAlreadyExist struct {
	ID string
}

func (e ErrSubscriberAlreadyExist) Error() string {
	return fmt.Sprintf("subscriber with id %s already exists", e.ID)
}
