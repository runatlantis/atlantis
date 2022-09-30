package job

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	"github.com/runatlantis/atlantis/server/core/config/valid"
	"github.com/runatlantis/atlantis/server/events/terraform/filter"
	"github.com/runatlantis/atlantis/server/logging"
)

type OutputLine struct {
	JobID string
	Line  string
}

func NewStreamHandler(
	jobStore Store,
	receiverRegistry receiverRegistry,
	logFilters valid.TerraformLogFilters,
	logger logging.Logger,
) *StreamHandler {

	logFilter := filter.LogFilter{
		Regexes: logFilters.Regexes,
	}

	jobOutputChan := make(chan *OutputLine)
	return &StreamHandler{
		JobOutput:        jobOutputChan,
		Store:            jobStore,
		ReceiverRegistry: receiverRegistry,
		LogFilter:        logFilter,
		Logger:           logger,
	}
}

type StreamHandler struct {
	JobOutput        chan *OutputLine
	Store            Store
	ReceiverRegistry receiverRegistry
	LogFilter        filter.LogFilter
	Logger           logging.Logger
}

func (s *StreamHandler) Stream(jobID string, msg string) {
	s.JobOutput <- &OutputLine{
		JobID: jobID,
		Line:  msg,
	}
}

func (s *StreamHandler) Handle() {
	for msg := range s.JobOutput {
		// Filter out log lines from job output
		if s.LogFilter.ShouldFilterLine(msg.Line) {
			continue
		}

		s.ReceiverRegistry.Broadcast(*msg)

		// Append new log to the output buffer for the job
		err := s.Store.Write(msg.JobID, msg.Line)
		if err != nil {
			s.Logger.Warn(fmt.Sprintf("appending log: %s for job: %s: %v", msg.Line, msg.JobID, err))
		}
	}
}

// Activity context since it's called from within an activity
func (s *StreamHandler) Close(ctx context.Context, jobID string) error {
	s.ReceiverRegistry.Close(ctx, jobID)

	// Update job status and persist to storage if configured
	if err := s.Store.Close(ctx, jobID, Complete); err != nil {
		return errors.Wrapf(err, "closing job store for ID: %s", jobID)
	}
	return nil
}
