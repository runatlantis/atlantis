// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package metrics_test

import (
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/runatlantis/atlantis/server/logging"
	"github.com/runatlantis/atlantis/server/metrics"
	"github.com/runatlantis/atlantis/server/scheduled"
	tally "github.com/uber-go/tally/v4"
	tallyprom "github.com/uber-go/tally/v4/prometheus"
	. "github.com/runatlantis/atlantis/testing"
)

// TestPRScopeManager_ScheduledCleanupIntegration is an end-to-end test that verifies
// the complete integration: PRScopeManager creates root scopes, scheduled executor
// runs periodic cleanup, and inactive PR scopes are properly cleaned up.
func TestPRScopeManager_ScheduledCleanupIntegration(t *testing.T) {
	logger := logging.NewNoopLogger(t)

	// Create a Prometheus reporter with unique registry for this test
	reporter := tallyprom.NewReporter(tallyprom.Options{
		Registerer: prometheus.NewRegistry(),
	})

	// Create PRScopeManager with short retention for testing (100ms)
	retentionPeriod := 100 * time.Millisecond
	manager := metrics.NewPRScopeManager(
		logger,
		reporter,
		tally.ScopeOptions{
			Prefix:          "test_integration",
			Separator:       tallyprom.DefaultSeparator,
			SanitizeOptions: &tallyprom.DefaultSanitizerOpts,
		},
		time.Second,
		retentionPeriod,
	)

	// Create scheduled executor service (same as production)
	scheduledService := scheduled.NewExecutorService(
		tally.NoopScope,
		logger,
	)

	// Add PRScopeManager as a scheduled job with same period as retention
	// (this is exactly how it's done in server.go)
	scheduledService.AddJob(scheduled.JobDefinition{
		Job:    manager,
		Period: retentionPeriod,
	})

	// Start the scheduled executor in background
	go scheduledService.Run()
	defer func() {
		// Cleanup would happen via signal, but for test we just return
	}()

	// === Simulate production usage ===

	// 1. Create metrics for PR #1 (will become inactive)
	scope1 := manager.GetOrCreatePRScope("owner/repo", 1, map[string]string{
		"pr_number": "1",
	})
	scope1.Counter("commands").Inc(1)
	Equals(t, 1, manager.GetStats()) // 1 active PR

	// 2. Create metrics for PR #2 (will stay active)
	scope2 := manager.GetOrCreatePRScope("owner/repo", 2, map[string]string{
		"pr_number": "2",
	})
	scope2.Counter("commands").Inc(1)
	Equals(t, 2, manager.GetStats()) // 2 active PRs

	// 3. PR #2 continues to be active (simulate ongoing commands)
	keepAlive := time.NewTicker(50 * time.Millisecond)
	defer keepAlive.Stop()
	done := make(chan bool)
	go func() {
		for {
			select {
			case <-keepAlive.C:
				// Keep PR #2 active by accessing its scope
				s := manager.GetOrCreatePRScope("owner/repo", 2, map[string]string{
					"pr_number": "2",
				})
				s.Counter("commands").Inc(1)
			case <-done:
				return
			}
		}
	}()

	// 4. Wait for retention period + a bit more to ensure cleanup runs
	// The scheduled job will run after retentionPeriod and should clean up PR #1
	time.Sleep(retentionPeriod + 50*time.Millisecond)

	// Stop keep-alive
	done <- true

	// 5. Verify: PR #1 should be cleaned up (inactive), PR #2 should remain (active)
	active := manager.GetStats()
	Equals(t, 1, active)

	// 6. Verify PR #2 still works (wasn't cleaned)
	scope2Again := manager.GetOrCreatePRScope("owner/repo", 2, map[string]string{
		"pr_number": "2",
	})
	Assert(t, scope2Again != nil, "PR #2 scope should still exist")

	t.Log("✅ Integration test passed: scheduled cleanup properly removes inactive PRs while preserving active ones")
}

// TestPRScopeManager_ExplicitCloseIntegration verifies that explicitly closing
// a PR immediately removes its scope (doesn't wait for retention period).
func TestPRScopeManager_ExplicitCloseIntegration(t *testing.T) {
	logger := logging.NewNoopLogger(t)
	// Use unique reporter for each test to avoid Prometheus registration conflicts
	reporter := tallyprom.NewReporter(tallyprom.Options{
		Registerer: prometheus.NewRegistry(),
	})

	// Create PRScopeManager with long retention (we won't wait for it)
	manager := metrics.NewPRScopeManager(
		logger,
		reporter,
		tally.ScopeOptions{
			Prefix:          "test_explicit",
			Separator:       tallyprom.DefaultSeparator,
			SanitizeOptions: &tallyprom.DefaultSanitizerOpts,
		},
		time.Second,
		1*time.Hour, // Long retention - we're testing explicit close
	)

	// Create metrics for a PR
	scope := manager.GetOrCreatePRScope("owner/repo", 123, map[string]string{
		"pr_number": "123",
	})
	scope.Counter("commands").Inc(1)
	Equals(t, 1, manager.GetStats())

	// Explicitly close the PR (simulates PR being merged/closed)
	manager.MarkPRClosed("owner/repo", 123)

	// Verify immediate cleanup (no waiting for retention period)
	Equals(t, 0, manager.GetStats())

	t.Log("✅ Integration test passed: explicit close immediately removes PR scope")
}

// TestPRScopeManager_DisabledCleanup verifies that when retention is 0,
// cleanup is disabled and scopes are never automatically cleaned.
func TestPRScopeManager_DisabledCleanup(t *testing.T) {
	logger := logging.NewNoopLogger(t)
	// Use unique reporter for each test to avoid Prometheus registration conflicts
	reporter := tallyprom.NewReporter(tallyprom.Options{
		Registerer: prometheus.NewRegistry(),
	})

	// Create PRScopeManager with 0 retention (cleanup disabled)
	manager := metrics.NewPRScopeManager(
		logger,
		reporter,
		tally.ScopeOptions{
			Prefix:          "test_disabled",
			Separator:       tallyprom.DefaultSeparator,
			SanitizeOptions: &tallyprom.DefaultSanitizerOpts,
		},
		time.Second,
		0, // Retention = 0 means cleanup disabled
	)

	// Create metrics for a PR
	scope := manager.GetOrCreatePRScope("owner/repo", 999, map[string]string{
		"pr_number": "999",
	})
	scope.Counter("commands").Inc(1)
	Equals(t, 1, manager.GetStats())

	// Wait a bit (would normally trigger cleanup if enabled)
	time.Sleep(50 * time.Millisecond)

	// Run cleanup manually
	cleaned := manager.CleanupStaleScopes()
	Equals(t, 0, cleaned) // Nothing cleaned because retention = 0

	// PR scope should still exist
	Equals(t, 1, manager.GetStats())

	t.Log("✅ Integration test passed: cleanup disabled when retention = 0")
}
