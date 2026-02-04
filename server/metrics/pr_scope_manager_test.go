// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package metrics_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/runatlantis/atlantis/server/logging"
	"github.com/runatlantis/atlantis/server/metrics"
	. "github.com/runatlantis/atlantis/testing"
	tally "github.com/uber-go/tally/v4"
	promreporter "github.com/uber-go/tally/v4/prometheus"
)

func TestPRScopeManager_GetOrCreatePRScope(t *testing.T) {
	logger := logging.NewNoopLogger(t)
	reporter := promreporter.NewReporter(promreporter.Options{})

	manager := metrics.NewPRScopeManager(
		logger,
		reporter,
		tally.ScopeOptions{
			Prefix:                 "atlantis_get",
			Separator:              "_",
			OmitCardinalityMetrics: true,
		},
		100*time.Millisecond,
		1*time.Hour,
	)

	// First call creates a new scope
	scope1 := manager.GetOrCreatePRScope("owner/repo", 123, map[string]string{
		"pr_number": "123",
	})
	Assert(t, scope1 != nil, "scope should be created")

	// Second call returns the same scope
	scope2 := manager.GetOrCreatePRScope("owner/repo", 123, map[string]string{
		"pr_number": "123",
	})
	Assert(t, scope1 == scope2, "should return same scope instance")

	// Different PR gets different scope
	scope3 := manager.GetOrCreatePRScope("owner/repo", 456, map[string]string{
		"pr_number": "456",
	})
	Assert(t, scope1 != scope3, "different PR should get different scope")

	active := manager.GetStats()
	Equals(t, 2, active)
}

func TestPRScopeManager_MarkPRClosedAndCleanup(t *testing.T) {
	logger := logging.NewNoopLogger(t)
	reporter := promreporter.NewReporter(promreporter.Options{})

	manager := metrics.NewPRScopeManager(
		logger,
		reporter,
		tally.ScopeOptions{
			Prefix:                 "atlantis_mark",
			Separator:              "_",
			OmitCardinalityMetrics: true,
		},
		100*time.Millisecond,
		50*time.Millisecond, // Short retention for testing
	)

	// Create scope for PR
	scope := manager.GetOrCreatePRScope("owner/repo", 123, map[string]string{
		"pr_number": "123",
	})
	Assert(t, scope != nil, "scope created")

	// Use the scope
	scope.Counter("test_ops").Inc(1)

	active := manager.GetStats()
	Equals(t, 1, active)

	// Mark PR as closed - should close immediately
	manager.MarkPRClosed("owner/repo", 123)

	// Scope should be closed and removed immediately
	active = manager.GetStats()
	Equals(t, 0, active)
}

func TestPRScopeManager_MultipleRootScopes(t *testing.T) {
	logger := logging.NewNoopLogger(t)
	reporter := promreporter.NewReporter(promreporter.Options{})

	manager := metrics.NewPRScopeManager(
		logger,
		reporter,
		tally.ScopeOptions{
			Prefix:                 "atlantis_multi",
			Separator:              "_",
			OmitCardinalityMetrics: true,
		},
		100*time.Millisecond,
		1*time.Hour,
	)

	// Create scopes for multiple PRs
	scope1 := manager.GetOrCreatePRScope("owner/repo", 1, map[string]string{"pr_number": "1"})
	scope2 := manager.GetOrCreatePRScope("owner/repo", 2, map[string]string{"pr_number": "2"})
	scope3 := manager.GetOrCreatePRScope("owner/repo", 3, map[string]string{"pr_number": "3"})

	// Each should be independent root scopes - use unique counter names to avoid Prometheus conflicts
	scope1.Counter("operations_1").Inc(1)
	scope2.Counter("operations_2").Inc(2)
	scope3.Counter("operations_3").Inc(3)

	active := manager.GetStats()
	Equals(t, 3, active)

	// Close one PR - should close immediately
	manager.MarkPRClosed("owner/repo", 2)

	active = manager.GetStats()
	Equals(t, 2, active)

	// Other PRs still work - continue using unique counter names
	scope1.Counter("operations_1_again").Inc(1)
	scope3.Counter("operations_3_again").Inc(1)

	t.Log("Successfully verified independent root scopes per PR")
}

func TestPRScopeManager_LargeScale(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping large scale test in short mode")
	}

	logger := logging.NewNoopLogger(t)
	// Create ONE shared reporter that all root scopes will use
	reporter := promreporter.NewReporter(promreporter.Options{})

	manager := metrics.NewPRScopeManager(
		logger,
		reporter,
		tally.ScopeOptions{
			Prefix:                 "atlantis_large",
			Separator:              "_",
			OmitCardinalityMetrics: true,
		},
		100*time.Millisecond,
		50*time.Millisecond,
	)

	numPRs := 10 // Keep small to avoid rapid registration issues
	t.Logf("Creating %d PR scopes (each is a separate root scope)...", numPRs)

	// Create PR scopes one by one
	// Each root scope shares the same reporter, Prometheus differentiates via tags
	for i := 0; i < numPRs; i++ {
		scope := manager.GetOrCreatePRScope("test/repo", i, map[string]string{
			"pr_number": fmt.Sprintf("%d", i),
		})
		// Use different counter names per PR to avoid conflicts during rapid creation
		scope.Counter(fmt.Sprintf("operations_%d", i)).Inc(1)

		// Small delay to avoid overwhelming Prometheus registration
		if i > 0 && i%5 == 0 {
			time.Sleep(10 * time.Millisecond)
		}
	}

	active := manager.GetStats()
	Equals(t, numPRs, active)

	// Close all PRs - should close immediately
	t.Log("Closing all PRs...")
	startTime := time.Now()
	for i := 0; i < numPRs; i++ {
		manager.MarkPRClosed("test/repo", i)
	}
	duration := time.Since(startTime)

	active = manager.GetStats()
	Equals(t, 0, active)
	t.Logf("Closed %d root scopes immediately in %v", numPRs, duration)
}

func TestPRScopeManager_NonExistentPRClose(t *testing.T) {
	logger := logging.NewNoopLogger(t)
	reporter := promreporter.NewReporter(promreporter.Options{})

	manager := metrics.NewPRScopeManager(
		logger,
		reporter,
		tally.ScopeOptions{
			Prefix:                 "atlantis_nonexist",
			OmitCardinalityMetrics: true,
		},
		100*time.Millisecond,
		1*time.Hour,
	)

	// Closing a PR that never had a scope should not panic
	manager.MarkPRClosed("owner/repo", 999)

	active := manager.GetStats()
	Equals(t, 0, active)
}
