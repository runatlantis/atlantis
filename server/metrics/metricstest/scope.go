package metricstest

import (
	"testing"

	"github.com/runatlantis/atlantis/server/logging"
	"github.com/runatlantis/atlantis/server/metrics"
	tally "github.com/uber-go/tally/v4"
)

func NewLoggingScope(t *testing.T, logger logging.SimpleLogging, statsNamespace string) tally.Scope {
	t.Helper()
	scope, closer, err := metrics.NewLoggingScope(logger, "atlantis")
	if err != nil {
		t.Fatalf("failed to create metrics logging scope: %v", err)
	}
	t.Cleanup(func() {
		closer.Close()
	})
	return scope
}
