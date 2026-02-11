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

// TestPRScopeManager_InactivePRCleanup verifies that scopes are cleaned up
// even if the PR is never explicitly closed (abandoned PRs, deleted repos, etc.)
func TestPRScopeManager_InactivePRCleanup(t *testing.T) {
	logger := logging.NewNoopLogger(t)
	reporter := promreporter.NewReporter(promreporter.Options{})

	manager := metrics.NewPRScopeManager(
		logger,
		reporter,
		tally.ScopeOptions{
			Prefix:                 "atlantis_inactive",
			Separator:              "_",
			OmitCardinalityMetrics: true,
		},
		100*time.Millisecond,
		50*time.Millisecond, // Short retention for testing
	)

	// Create scope for PR that will never be explicitly closed
	scope := manager.GetOrCreatePRScope("owner/repo", 999, map[string]string{
		"pr_number": "999",
	})
	Assert(t, scope != nil, "scope created")

	// Use it once
	scope.Counter("operations").Inc(1)

	active := manager.GetStats()
	Equals(t, 1, active)

	// Don't call MarkPRClosed - simulate abandoned PR

	// Cleanup immediately shouldn't remove (recently used)
	cleaned := manager.CleanupStaleScopes()
	Equals(t, 0, cleaned)

	active = manager.GetStats()
	Equals(t, 1, active)

	// Wait for retention period without any activity
	time.Sleep(60 * time.Millisecond)

	// Now cleanup should remove the inactive scope
	cleaned = manager.CleanupStaleScopes()
	Equals(t, 1, cleaned)

	active = manager.GetStats()
	Equals(t, 0, active)

	t.Log("Successfully cleaned up inactive PR scope without explicit close")
}

// TestPRScopeManager_ActivePRNotCleaned verifies that PRs with recent activity
// are NOT cleaned up, even if they've been open a long time
func TestPRScopeManager_ActivePRNotCleaned(t *testing.T) {
	logger := logging.NewNoopLogger(t)
	reporter := promreporter.NewReporter(promreporter.Options{})

	manager := metrics.NewPRScopeManager(
		logger,
		reporter,
		tally.ScopeOptions{
			Prefix:                 "atlantis_active",
			Separator:              "_",
			OmitCardinalityMetrics: true,
		},
		100*time.Millisecond,
		100*time.Millisecond, // retention period
	)

	// Create scope
	scope := manager.GetOrCreatePRScope("owner/repo", 123, map[string]string{
		"pr_number": "123",
	})

	// Simulate ongoing activity - use unique counter names to avoid Prometheus conflicts
	for i := 0; i < 5; i++ {
		time.Sleep(30 * time.Millisecond)

		// Access the scope (updates lastAccess)
		scope2 := manager.GetOrCreatePRScope("owner/repo", 123, map[string]string{
			"pr_number": "123",
		})
		Assert(t, scope == scope2, "should return same scope")
		scope2.Counter(fmt.Sprintf("operations_%d", i)).Inc(1)

		// Try cleanup
		cleaned := manager.CleanupStaleScopes()
		Equals(t, 0, cleaned) // Should not clean up active PR

		active := manager.GetStats()
		Equals(t, 1, active)
	}

	t.Log("Active PR with recent activity was not cleaned up")
}

// TestPRScopeManager_MixedClosedAndInactive verifies cleanup handles both
// explicitly closed PRs and inactive PRs correctly
func TestPRScopeManager_MixedClosedAndInactive(t *testing.T) {
	logger := logging.NewNoopLogger(t)
	reporter := promreporter.NewReporter(promreporter.Options{})

	manager := metrics.NewPRScopeManager(
		logger,
		reporter,
		tally.ScopeOptions{
			Prefix:                 "atlantis_mixed",
			Separator:              "_",
			OmitCardinalityMetrics: true,
		},
		100*time.Millisecond,
		50*time.Millisecond,
	)

	// PR 1: Will be explicitly closed
	scope1 := manager.GetOrCreatePRScope("owner/repo", 1, map[string]string{"pr_number": "1"})
	scope1.Counter("ops").Inc(1)

	// PR 2: Will be abandoned (inactive)
	scope2 := manager.GetOrCreatePRScope("owner/repo", 2, map[string]string{"pr_number": "2"})
	scope2.Counter("ops").Inc(1)

	// PR 3: Will remain active
	scope3 := manager.GetOrCreatePRScope("owner/repo", 3, map[string]string{"pr_number": "3"})
	scope3.Counter("ops").Inc(1)

	active := manager.GetStats()
	Equals(t, 3, active)

	// Close PR 1 - closes immediately
	manager.MarkPRClosed("owner/repo", 1)

	active = manager.GetStats()
	Equals(t, 2, active) // PR 1 closed immediately, PRs 2 and 3 remain

	// Keep PR 3 active
	time.Sleep(30 * time.Millisecond)
	scope3 = manager.GetOrCreatePRScope("owner/repo", 3, map[string]string{"pr_number": "3"})
	scope3.Counter("ops2").Inc(1)

	// Wait for retention to expire for PR 2 (inactive)
	time.Sleep(30 * time.Millisecond)

	// Cleanup should remove:
	// - PR 2 (inactive + retention expired)
	// But NOT PR 3 (active with recent access)
	cleaned := manager.CleanupStaleScopes()
	Equals(t, 1, cleaned)

	active = manager.GetStats()
	Equals(t, 1, active) // Only PR 3 remains

	t.Log("Successfully cleaned up closed and inactive PRs while preserving active PR")
}
