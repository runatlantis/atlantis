package events

import (
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/logging"
	"github.com/runatlantis/atlantis/server/metrics"
	tally "github.com/uber-go/tally/v4"
)

type InstrumentedPullClosedExecutor struct {
	scope   tally.Scope
	log     logging.SimpleLogging
	cleaner PullCleaner
}

func NewInstrumentedPullClosedExecutor(
	scope tally.Scope, log logging.SimpleLogging, cleaner PullCleaner,
) PullCleaner {
	scope = scope.SubScope("pullclosed_cleanup")

	for _, m := range []string{metrics.ExecutionSuccessMetric, metrics.ExecutionErrorMetric} {
		metrics.InitCounter(scope, m)
	}

	return &InstrumentedPullClosedExecutor{
		scope:   scope,
		log:     log,
		cleaner: cleaner,
	}
}

func (e *InstrumentedPullClosedExecutor) CleanUpPull(logger logging.SimpleLogging, repo models.Repo, pull models.PullRequest) error {

	executionSuccess := e.scope.Counter(metrics.ExecutionSuccessMetric)
	executionError := e.scope.Counter(metrics.ExecutionErrorMetric)
	executionTime := e.scope.Timer(metrics.ExecutionTimeMetric).Start()
	defer executionTime.Stop()

	logger.Info("Initiating cleanup of pull data.")

	err := e.cleaner.CleanUpPull(logger, repo, pull)

	if err != nil {
		executionError.Inc(1)
		logger.Err("error during cleanup of pull data", err)
		return err
	}

	executionSuccess.Inc(1)

	return nil
}
