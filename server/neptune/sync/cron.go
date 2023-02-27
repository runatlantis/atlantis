package sync

import (
	"context"
	"sync"
	"time"

	"github.com/runatlantis/atlantis/server/logging"
)

type Cron struct {
	Executor
	Frequency time.Duration
}

// AsyncScheduler handles scheduling background work with the correct
// context values while ensuring work can gracefully exit when
// necessary.
type CronScheduler struct {
	delegate  *SynchronousScheduler
	logger    logging.Logger
	ctx       context.Context
	cancelCtx context.CancelFunc
	wg        sync.WaitGroup
}

func NewCronScheduler(logger logging.Logger) *CronScheduler {
	ctx, cancel := context.WithCancel(context.Background())
	return &CronScheduler{
		delegate:  &SynchronousScheduler{Logger: logger},
		ctx:       ctx,
		cancelCtx: cancel,
		logger:    logger,
	}
}

func (s *CronScheduler) Schedule(cron *Cron) {
	ticker := time.NewTicker(cron.Frequency)

	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		defer ticker.Stop()
		for {
			select {
			case <-s.ctx.Done():
				s.logger.Warn("Received interrupt, cancelling job")
				return
			case <-ticker.C:
				_ = s.delegate.Schedule(s.ctx, cron.Executor)
			}
		}
	}()
}

func (s *CronScheduler) Shutdown(t time.Duration) {
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
