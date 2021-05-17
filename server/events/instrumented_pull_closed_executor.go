package events

import (
	"strconv"

	stats "github.com/lyft/gostats"
	"github.com/runatlantis/atlantis/server/events/metrics"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/logging"
)

type InstrumentedPullClosedExecutor struct {
	scope   stats.Scope
	log     logging.SimpleLogging
	cleaner PullCleaner
}

func NewInstrumentedPullClosedExecutor(
	scope stats.Scope, log logging.SimpleLogging, cleaner PullCleaner,
) PullCleaner {

	return &InstrumentedPullClosedExecutor{
		scope:   scope.Scope("pullclosed.cleanup"),
		log:     log,
		cleaner: cleaner,
	}
}

func (e *InstrumentedPullClosedExecutor) CleanUpPull(repo models.Repo, pull models.PullRequest) error {
	log := e.log.With(
		"repository", repo.FullName,
		"pull-num", strconv.Itoa(pull.Num),
	)

	executionSuccess := e.scope.NewCounter(metrics.ExecutionSuccessMetric)
	executionError := e.scope.NewCounter(metrics.ExecutionErrorMetric)
	executionTime := e.scope.NewTimer(metrics.ExecutionTimeMetric).AllocateSpan()
	defer executionTime.Complete()

	log.Info("Initiating cleanup of pull data.")

	err := e.cleaner.CleanUpPull(repo, pull)

	if err != nil {
		executionError.Inc()
		log.Err("error during cleanup of pull data", err)
		return err
	}

	executionSuccess.Inc()

	return nil
}
