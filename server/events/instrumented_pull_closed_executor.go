package events

import (
	"strconv"

	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/logging"
	"github.com/runatlantis/atlantis/server/metrics"
	"github.com/uber-go/tally"
)

type InstrumentedPullClosedExecutor struct {
	scope   tally.Scope
	log     logging.SimpleLogging
	cleaner PullCleaner
}

func NewInstrumentedPullClosedExecutor(
	scope tally.Scope, log logging.SimpleLogging, cleaner PullCleaner,
) PullCleaner {

	return &InstrumentedPullClosedExecutor{
		scope:   scope.SubScope("pullclosed_cleanup"),
		log:     log,
		cleaner: cleaner,
	}
}

func (e *InstrumentedPullClosedExecutor) CleanUpPull(repo models.Repo, pull models.PullRequest) error {
	log := e.log.With(
		"repository", repo.FullName,
		"pull-num", strconv.Itoa(pull.Num),
	)

	executionSuccess := e.scope.Counter(metrics.ExecutionSuccessMetric)
	executionError := e.scope.Counter(metrics.ExecutionErrorMetric)
	executionTime := e.scope.Timer(metrics.ExecutionTimeMetric).Start()
	defer executionTime.Stop()

	log.Info("Initiating cleanup of pull data.")

	err := e.cleaner.CleanUpPull(repo, pull)

	if err != nil {
		executionError.Inc(1)
		log.Err("error during cleanup of pull data", err)
		return err
	}

	executionSuccess.Inc(1)

	return nil
}
