// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package scheduled

import (
	"testing"

	tally "github.com/uber-go/tally/v4"
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
