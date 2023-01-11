package scheduled

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/runatlantis/atlantis/server/logging"
	"github.com/uber-go/tally"
)

type ExecutorService struct {
	log logging.SimpleLogging

	// jobs
	runtimeStatsPublisher JobDefinition
}

func NewExecutorService(
	statsScope tally.Scope,
	log logging.SimpleLogging,
) *ExecutorService {

	scheduledScope := statsScope.SubScope("scheduled")
	runtimeStatsPublisher := NewRuntimeStats(scheduledScope)

	runtimeStatsPublisherJob := JobDefinition{
		Job:    runtimeStatsPublisher,
		Period: 10 * time.Second,
	}

	return &ExecutorService{
		log:                   log,
		runtimeStatsPublisher: runtimeStatsPublisherJob,
	}
}

type JobDefinition struct {
	Job    Job
	Period time.Duration
}

func (s *ExecutorService) Run() {
	s.log.Info("Scheduled Executor Service started")

	ctx, cancel := context.WithCancel(context.Background())

	var wg sync.WaitGroup

	s.runScheduledJob(ctx, &wg, s.runtimeStatsPublisher)

	interrupt := make(chan os.Signal, 1)

	// Stop on SIGINTs and SIGTERMs.
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	<-interrupt

	s.log.Warn("Received interrupt. Attempting to Shut down scheduled executor service")

	cancel()
	wg.Wait()

	s.log.Warn("All jobs completed, exiting.")
}

func (s *ExecutorService) runScheduledJob(ctx context.Context, wg *sync.WaitGroup, jd JobDefinition) {
	ticker := time.NewTicker(jd.Period)
	wg.Add(1)

	go func() {
		defer wg.Done()
		defer ticker.Stop()

		// Ensure we recover from any panics to keep the jobs isolated.
		// Keep the recovery outside the select to ensure that we don't infinitely panic.
		defer func() {
			if r := recover(); r != nil {
				s.log.Err("Recovered from panic: %v", r)
			}
		}()

		for {
			select {
			case <-ctx.Done():
				s.log.Warn("Received interrupt, cancelling job")
				return
			case <-ticker.C:
				jd.Job.Run()
			}
		}
	}()

}

type Job interface {
	Run()
}
