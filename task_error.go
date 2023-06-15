package task

import (
	"errors"
	"fmt"
)

var canceledError = errors.New("Task canceled")

func Canceled() error {
	return canceledError
}

func toError(msg interface{}) error {
	switch v := msg.(type) {
	case error:
		return v
	default:
		return errors.New(toString(v))
	}
}

type ForcePanic struct {
	msg interface{}
}

func PanicForce(a interface{}) {
	panic(&ForcePanic{msg: a})
}

func internalPanicForce(msg interface{}) {
	panic(&ForcePanic{msg: fmt.Sprintf("Task: %v", msg)})
}
