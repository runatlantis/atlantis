package test

import (
	"time"

	"go.temporal.io/sdk/workflow"
)

// An emptyCtx is never canceled, has no values, and has no deadline.  It is not
// struct{}, since vars of this type must have distinct addresses.
//
// Note this is copied from the temporal internal lib, in order to allow us to test
// small parts of our workflow without testing the whole thing.
type emptyCtx int

func (*emptyCtx) Deadline() (deadline time.Time, ok bool) {
	return
}

func (*emptyCtx) Done() workflow.Channel {
	return nil
}

func (*emptyCtx) Err() error {
	return nil
}

func (*emptyCtx) Value(_ interface{}) interface{} {
	return nil
}

func (e *emptyCtx) String() string {
	switch e {
	case background:
		return "context.Background"
	}
	return "unknown empty Context"
}

var (
	background = new(emptyCtx)
)

// Background returns a non-nil, empty Context. It is never canceled, has no
// values, and has no deadline
func Background() workflow.Context {
	return background
}
