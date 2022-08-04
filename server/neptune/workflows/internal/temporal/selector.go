package temporal

import (
	"go.temporal.io/sdk/workflow"
	"time"
)

// SelectorWithTimeout makes it a bit easier to add Timeout futures
// by ensuring that we are creating a new cancellable context and returning
// the cancel func to callers.
type SelectorWithTimeout struct {
	workflow.Selector
}

func (s *SelectorWithTimeout) AddTimeout(ctx workflow.Context, timeout time.Duration, onTimeout func(f workflow.Future)) (workflow.CancelFunc, *SelectorWithTimeout) {
	ctx, cancel := workflow.WithCancel(ctx)
	delegate := s.AddFuture(workflow.NewTimer(ctx, timeout), onTimeout)

	return cancel, &SelectorWithTimeout{
		delegate,
	}
}
