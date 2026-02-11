// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package metrics

import (
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/runatlantis/atlantis/server/logging"
	tally "github.com/uber-go/tally/v4"
	tallyprom "github.com/uber-go/tally/v4/prometheus"
)

// PRScopeManager manages separate root scopes for each PR, allowing them to be
// individually closed when PRs are done, preventing exessive resource consumption over time.
type PRScopeManager struct {
	logger          logging.SimpleLogging
	baseReporter    tally.BaseStatsReporter
	promReporter    tallyprom.Reporter // Optional: for cleaning up Prometheus metrics
	scopeOptions    tally.ScopeOptions
	reportInterval  time.Duration
	retentionPeriod time.Duration

	mu       sync.RWMutex
	prScopes map[string]*prScopeEntry // key: "repo/pullnum"
}

type prScopeEntry struct {
	scope      *trackingScope
	closer     io.Closer
	lastAccess time.Time
	tags       map[string]string // Store tags for Prometheus cleanup
}

// trackingScope wraps a tally.Scope to track which subscopes are created.
// This allows us to clean up Prometheus metrics without hardcoding subscope names.
type trackingScope struct {
	tally.Scope
	mu        sync.RWMutex
	subscopes map[string]bool // Set of subscope names that have been created
}

func newTrackingScope(scope tally.Scope) *trackingScope {
	return &trackingScope{
		Scope:     scope,
		subscopes: make(map[string]bool),
	}
}

// SubScope wraps the underlying SubScope call and tracks the subscope name.
// When Prometheus cleanup runs, only metrics from subscopes that were actually
// created will be targeted for deletion, avoiding unnecessary cleanup attempts.
func (ts *trackingScope) SubScope(name string) tally.Scope {
	ts.mu.Lock()
	ts.subscopes[name] = true
	ts.mu.Unlock()
	return ts.Scope.SubScope(name)
}

// getSubscopes returns a list of all subscope names that were created.
func (ts *trackingScope) getSubscopes() []string {
	ts.mu.RLock()
	defer ts.mu.RUnlock()

	names := make([]string, 0, len(ts.subscopes))
	for name := range ts.subscopes {
		names = append(names, name)
	}
	return names
}

// NewPRScopeManager creates a manager for PR-specific root scopes.
func NewPRScopeManager(
	logger logging.SimpleLogging,
	baseReporter tally.BaseStatsReporter,
	scopeOptions tally.ScopeOptions,
	reportInterval time.Duration,
	retentionPeriod time.Duration,
) *PRScopeManager {
	manager := &PRScopeManager{
		logger:          logger,
		baseReporter:    baseReporter,
		scopeOptions:    scopeOptions,
		reportInterval:  reportInterval,
		retentionPeriod: retentionPeriod,
		prScopes:        make(map[string]*prScopeEntry),
	}

	// If the reporter is a Prometheus reporter, store it for metric cleanup
	if promReporter, ok := baseReporter.(tallyprom.Reporter); ok {
		manager.promReporter = promReporter
	}

	return manager
}

// GetOrCreatePRScope returns a root scope for the given PR with the specified tags.
// If a scope already exists for this PR, it returns the existing one.
// Each PR gets its own root scope that can be independently closed.
func (m *PRScopeManager) GetOrCreatePRScope(repo string, prNum int, tags map[string]string) tally.Scope {
	key := m.prKey(repo, prNum)

	m.mu.Lock()
	if entry, exists := m.prScopes[key]; exists {
		// Update last access time
		entry.lastAccess = time.Now()
		m.mu.Unlock()
		return entry.scope
	}
	m.mu.Unlock()

	// Create a new root scope for this PR
	m.mu.Lock()
	defer m.mu.Unlock()

	// Double-check after acquiring write lock
	if entry, exists := m.prScopes[key]; exists {
		return entry.scope
	}

	// Create new root scope with the PR-specific tags
	// IMPORTANT: Each root scope shares the same reporter, so we must
	// use tags to differentiate metrics (for CachedStatsReporter like Prometheus)
	// or rely on metric name prefixes (for StatsReporter like StatsD).
	opts := m.scopeOptions

	// Set the appropriate reporter field based on type
	if cachedReporter, ok := m.baseReporter.(tally.CachedStatsReporter); ok {
		opts.CachedReporter = cachedReporter
	} else if statsReporter, ok := m.baseReporter.(tally.StatsReporter); ok {
		opts.Reporter = statsReporter
	}

	opts.Tags = mergeTags(opts.Tags, tags)

	scope, closer := tally.NewRootScope(opts, m.reportInterval)

	// Wrap the scope to track subscopes for Prometheus cleanup
	ts := newTrackingScope(scope)

	m.prScopes[key] = &prScopeEntry{
		scope:      ts,
		closer:     closer,
		lastAccess: time.Now(),
		tags:       tags, // Store tags for Prometheus cleanup
	}

	m.logger.Debug("created new root scope for pr %s", key)
	return ts
}

// MarkPRClosed immediately closes a PR's scope and removes it.
// For explicitly closed PRs, we don't need to wait for a retention period.
func (m *PRScopeManager) MarkPRClosed(repo string, prNum int) {
	key := m.prKey(repo, prNum)

	m.mu.Lock()
	defer m.mu.Unlock()

	entry, exists := m.prScopes[key]
	if !exists {
		// PR never had any commands run, no scope to clean up
		return
	}

	// Clean up Prometheus metrics before closing the scope
	m.deletePrometheusMetrics(entry)

	// Close the scope immediately
	if err := entry.closer.Close(); err != nil {
		m.logger.Err("error closing scope for pr %s: %s", key, err)
	} else {
		m.logger.Debug("closed scope for pr %s (explicitly closed)", key)
	}

	// Remove from active scopes
	delete(m.prScopes, key)
}

// Run implements the scheduled.Job interface for periodic cleanup.
func (m *PRScopeManager) Run() {
	m.CleanupStaleScopes()
}

// CleanupStaleMetrics is an alias for CleanupStaleScopes to satisfy the ScopeCleaner interface.
func (m *PRScopeManager) CleanupStaleMetrics() int {
	return m.CleanupStaleScopes()
}

// CleanupStaleScopes closes and removes scopes for inactive PRs
// (PRs with no activity for longer than retention period - abandoned PRs, deleted repos, etc.)
// Note: Explicitly closed PRs are handled immediately in MarkPRClosed().
func (m *PRScopeManager) CleanupStaleScopes() int {
	// If retention period is 0, cleanup is disabled
	if m.retentionPeriod == 0 {
		return 0
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()
	cleaned := 0

	// Clean up active PR scopes with no recent activity (abandoned/stale PRs)
	for key, entry := range m.prScopes {
		if now.Sub(entry.lastAccess) > m.retentionPeriod {
			m.logger.Debug("removed stale scope for pr %s %v", key, entry.scope)

			// Clean up Prometheus metrics before closing the scope
			m.deletePrometheusMetrics(entry)

			if err := entry.closer.Close(); err != nil {
				m.logger.Err("error closing scope for inactive pr %s: %s", key, err)
			} else {
				m.logger.Debug("closed scope for inactive pr %s (no activity for %s)", key, m.retentionPeriod)
				cleaned++
			}
			delete(m.prScopes, key)
		}
	}

	if cleaned > 0 {
		m.logger.Info("closed and cleaned up %d inactive pr root scopes", cleaned)
	}

	return cleaned
}

// GetStats returns the number of active PR scopes.
func (m *PRScopeManager) GetStats() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.prScopes)
}

func (m *PRScopeManager) prKey(repo string, prNum int) string {
	return fmt.Sprintf("%s/%d", repo, prNum)
}

// deletePrometheusMetrics removes Prometheus metric label values for a PR to prevent metric bloat.
// This is called before closing a PR scope to clean up stale metrics from the /metrics endpoint.
func (m *PRScopeManager) deletePrometheusMetrics(prScopeEntry *prScopeEntry) {
	if m.promReporter == nil {
		return
	}

	// Get the list of subscopes that were actually created during this PR's lifetime
	subscopes := prScopeEntry.scope.getSubscopes()
	if len(subscopes) == 0 {
		// No subscopes created, nothing to clean up
		return
	}

	// Build the full metric name with prefix and subscope
	prefix := m.scopeOptions.Prefix
	separator := m.scopeOptions.Separator
	if separator == "" {
		separator = "."
	}

	deletedCount := 0

	// Use Delete(Labels) instead of DeleteLabelValues() to avoid label key ordering issues.
	// Tally's keysFromMap() doesn't sort keys, so we can't reliably match the order used
	// when the collector was created. Delete(Labels) uses a map and is order-independent.
	//
	// IMPORTANT: Prometheus sanitizes label values (e.g., "owner/repo" becomes "owner_repo").
	// We must apply the same sanitization to our label values before calling Delete(),
	// otherwise the labels won't match and deletion will fail.
	labels := make(map[string]string, len(prScopeEntry.tags))
	if m.scopeOptions.SanitizeOptions != nil {
		sanitizer := tally.NewSanitizer(*m.scopeOptions.SanitizeOptions)
		for k, v := range prScopeEntry.tags {
			// Sanitize the label value to match what Prometheus actually stored
			labels[k] = sanitizer.Value(v)
		}
	} else {
		// No sanitizer configured - use original values
		for k, v := range prScopeEntry.tags {
			labels[k] = v
		}
	}

	// Extract all tag keys for RegisterCounter/RegisterTimer (order doesn't matter for registration)
	tagKeys := make([]string, 0, len(prScopeEntry.tags))
	for k := range prScopeEntry.tags {
		tagKeys = append(tagKeys, k)
	}

	for _, subscope := range subscopes {
		// Delete counters: execution_success and execution_error
		for _, metricName := range []string{ExecutionSuccessMetric, ExecutionErrorMetric} {
			fullName := prefix
			if fullName != "" {
				fullName += separator
			}
			fullName += subscope + separator + metricName

			// Try to get the counter and delete using label map
			if counterVec, err := m.promReporter.RegisterCounter(fullName, tagKeys, ""); err == nil {
				if counterVec.Delete(labels) {
					deletedCount++
				}
			}
		}

		// Delete timer (histogram or summary): execution_time
		fullName := prefix
		if fullName != "" {
			fullName += separator
		}
		fullName += subscope + separator + ExecutionTimeMetric

		// Try to get the timer and delete using label map
		if timerUnion, err := m.promReporter.RegisterTimer(fullName, tagKeys, "", nil); err == nil {
			if timerUnion.Histogram != nil && timerUnion.Histogram.Delete(labels) {
				deletedCount++
			} else if timerUnion.Summary != nil && timerUnion.Summary.Delete(labels) {
				deletedCount++
			}
		}
	}

	if deletedCount > 0 {
		m.logger.Debug("deleted %d Prometheus metric label values for PR", deletedCount)
	}
}

func mergeTags(base, additional map[string]string) map[string]string {
	result := make(map[string]string, len(base)+len(additional))
	for k, v := range base {
		result[k] = v
	}
	for k, v := range additional {
		result[k] = v
	}
	return result
}
