package job

import (
	"context"
	"fmt"
	"sync"

	"github.com/runatlantis/atlantis/server/core/config/valid"
	"github.com/runatlantis/atlantis/server/events/terraform/filter"
	"github.com/runatlantis/atlantis/server/logging"
)

type StreamCloserFn func()

type OutputLine struct {
	JobID string
	Line  string
}

func NewStreamHandler(
	jobStore Store,
	receiverRegistry ReceiverRegistry,
	logFilters valid.TerraformLogFilters,
	logger logging.Logger,
) *StreamHandler {

	logFilter := filter.LogFilter{
		Regexes: logFilters.Regexes,
	}

	return &StreamHandler{
		Store:            jobStore,
		ReceiverRegistry: receiverRegistry,
		LogFilter:        logFilter,
		Logger:           logger,
	}
}

type StreamHandler struct {
	Store            Store
	wg               sync.WaitGroup
	ReceiverRegistry ReceiverRegistry
	LogFilter        filter.LogFilter
	Logger           logging.Logger
}

// RegisterJob returns a channel and creates a receiver go routine
// that reads from the associated channel.
//
// New resources are created each time this is called so it is up to the consumer to ensure
// this is not being called more than once per job id.
//
// Note: Cleanup will be blocked until these routines are complete, therefore
// it is essential these channels are closed or we risk deadlock on shutdown
func (s *StreamHandler) RegisterJob(id string) chan string {
	jobOutput := make(chan string)
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		for line := range jobOutput {
			s.handle(&OutputLine{
				JobID: id,
				Line:  line,
			})
		}
	}()

	return jobOutput
}

// CloseJob closes any receivers within the registry, this exists outside of RegisterJob
// because atm our definition of job is conflated.  In certain cases jobs are mapped to individual operations (ie. TerraformPlan/TerraformApply)
// however, in other cases they represent a series of steps in a group (plan group -> [init, scripts, plan])
// in order to support streaming all logs for a group to the same job id, we call this after all the activities of a group have occurred.
// This actually is kind of broken because we don't checkpoint these across activities, so on shutdown we would lose logs for previous steps.
// TODO: we need to rethink this a bit and create clear definitions for our data models.
func (s *StreamHandler) CloseJob(ctx context.Context, jobID string) error {
	s.ReceiverRegistry.Close(ctx, jobID)
	return s.Store.Close(ctx, jobID, Complete)
}

func (s *StreamHandler) CleanUp(ctx context.Context) error {
	s.Logger.InfoContext(ctx, "waiting for jobs to finish")
	// Should we timeout here? We do risk a deadlock without this however adding a timeout also risks
	// contention with the worker stop timeout and therefore dropped logs
	s.wg.Wait()

	s.Logger.InfoContext(ctx, "cleaning up receiver registry and persisting to store")
	s.ReceiverRegistry.CleanUp()
	return s.Store.Cleanup(ctx)
}

func (s *StreamHandler) handle(msg *OutputLine) {
	// Filter out log lines from job output
	if s.LogFilter.ShouldFilterLine(msg.Line) {
		return
	}

	s.ReceiverRegistry.Broadcast(*msg)

	// Append new log to the output buffer for the job
	err := s.Store.Write(context.Background(), msg.JobID, msg.Line)
	if err != nil {
		s.Logger.Warn(fmt.Sprintf("appending log: %s for job: %s: %v", msg.Line, msg.JobID, err))
	}
}
