package scheduled

import (
	"testing"

	"github.com/uber-go/tally"
)

func TestRuntimeStatCollector_Run(t *testing.T) {
	scope := tally.NewTestScope("test", nil)
	r := NewRuntimeStats(scope)
	r.Run()

	expGaugeCount := 25
	if len(scope.Snapshot().Gauges()) != expGaugeCount {
		t.Errorf("Expected %d gauges but got %d", expGaugeCount, len(scope.Snapshot().Gauges()))
	}
}
