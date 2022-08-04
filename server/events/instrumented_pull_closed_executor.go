package events

import (
	"strconv"

	"github.com/runatlantis/atlantis/server/events/metrics"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/logging"
	"github.com/uber-go/tally/v4"
)

type InstrumentedPullClosedExecutor struct {
	scope   tally.Scope
	log     logging.Logger
	cleaner PullCleaner
}

func NewInstrumentedPullClosedExecutor(
	scope tally.Scope, log logging.Logger, cleaner PullCleaner,
) PullCleaner {

	return &InstrumentedPullClosedExecutor{
		scope:   scope.SubScope("pullclosed.cleanup"),
		log:     log,
		cleaner: cleaner,
	}
}

func (e *InstrumentedPullClosedExecutor) CleanUpPull(repo models.Repo, pull models.PullRequest) error {
	executionSuccess := e.scope.Counter(metrics.ExecutionSuccessMetric)
	executionError := e.scope.Counter(metrics.ExecutionErrorMetric)
	executionTime := e.scope.Timer(metrics.ExecutionTimeMetric).Start()
	defer executionTime.Stop()

	e.log.Info("Initiating cleanup of pull data.", map[string]interface{}{
		"repository": repo.FullName,
		"pull-num":   strconv.Itoa(pull.Num),
	})

	err := e.cleaner.CleanUpPull(repo, pull)

	if err != nil {
		executionError.Inc(1)
		e.log.Error("error during cleanup of pull data", map[string]interface{}{
			"repository": repo.FullName,
			"pull-num":   strconv.Itoa(pull.Num),
			"err":        err,
		})
		return err
	}

	executionSuccess.Inc(1)

	return nil
}
