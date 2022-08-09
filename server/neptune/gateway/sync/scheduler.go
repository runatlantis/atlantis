package sync

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/runatlantis/atlantis/server/logging"
	contextUtils "github.com/runatlantis/atlantis/server/neptune/gateway/context"
	"github.com/runatlantis/atlantis/server/recovery"
)

type Executor func(ctx context.Context) error

// AsyncScheduler handles scheduling background work with the correct
// context values while ensuring work can gracefully exit when
// necessary.
type AsyncScheduler struct {
	delegate  *SynchronousScheduler
	poolCtx   context.Context
	cancelCtx context.CancelFunc
	wg        sync.WaitGroup
}

func NewAsyncScheduler(logger logging.Logger) *AsyncScheduler {
	ctx, cancel := context.WithCancel(context.Background())

	return &AsyncScheduler{
		delegate: &SynchronousScheduler{
			logger: logger,
		},
		poolCtx:   ctx,
		cancelCtx: cancel,
	}
}

func (s *AsyncScheduler) Schedule(ctx context.Context, f Executor) {

	// copy relevant context fields to a new ctx based off a single parent
	// for easy cancellation when shutting down.
	ctx = contextUtils.CopyFields(s.poolCtx, ctx)

	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		s.delegate.Schedule(ctx, f)
	}()
}

func (s *AsyncScheduler) Shutdown(t time.Duration) {
	done := make(chan struct{})
	go func() {
		s.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		return
	case <-time.After(t):
		s.cancelCtx()
	}
}

// SynchronousScheduler schedules work and handles panics/logging in a consistent manner.
type SynchronousScheduler struct {
	logger logging.Logger
}

func (s *SynchronousScheduler) Schedule(ctx context.Context, f Executor) {
	var err error
	defer func() {
		if r := recover(); r != nil {
			stack := recovery.Stack(3)
			s.logger.ErrorContext(ctx, fmt.Sprintf("PANIC: %s\n%s", r, stack))
		}
		if err != nil {
			s.logger.ErrorContext(context.WithValue(ctx, contextUtils.Err, err), "error running handle")
		}
	}()
	err = f(ctx)
}
