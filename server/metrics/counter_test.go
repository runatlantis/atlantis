package metrics

import (
	"testing"

	tally "github.com/uber-go/tally/v4"
)

func TestInitCounter(t *testing.T) {
	scope := tally.NewTestScope("test", nil)

	InitCounter(scope, "counter")

	counter, ok := scope.Snapshot().Counters()["test.counter+"]
	if !ok {
		t.Errorf("Counter not found")
	}
	if counter.Value() != 0 {
		t.Errorf("Counter is not initialized")
	}
}
