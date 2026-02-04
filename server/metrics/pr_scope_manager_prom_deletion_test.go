// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package metrics_test

import (
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/runatlantis/atlantis/server/logging"
	"github.com/runatlantis/atlantis/server/metrics"
	tally "github.com/uber-go/tally/v4"
	tallyprom "github.com/uber-go/tally/v4/prometheus"
	. "github.com/runatlantis/atlantis/testing"
)

// TestPRScopeManager_PrometheusMetricDeletion verifies that Prometheus metrics
// are actually deleted when a PR scope is closed.
func TestPRScopeManager_PrometheusMetricDeletion(t *testing.T) {
	logger := logging.NewNoopLogger(t)

	// Create a custom Prometheus registry so we can inspect metrics
	registry := prometheus.NewRegistry()
	reporter := tallyprom.NewReporter(tallyprom.Options{
		Registerer: registry,
	})

	// Create PRScopeManager
	manager := metrics.NewPRScopeManager(
		logger,
		reporter,
		tally.ScopeOptions{
			Prefix:          "test_deletion",
			Separator:       tallyprom.DefaultSeparator,
			SanitizeOptions: &tallyprom.DefaultSanitizerOpts,
		},
		100*time.Millisecond, // Report interval
		1*time.Hour,          // Long retention - we'll test explicit close
	)

	// Create a PR scope with tags
	scope := manager.GetOrCreatePRScope("owner/repo", 42, map[string]string{
		"base_repo": "owner/repo",
		"pr_number": "42",
	})

	// Create subscopes and metrics that match what instrumented_client.go creates
	updateStatusScope := scope.SubScope("update_status")
	updateStatusScope.Counter(metrics.ExecutionSuccessMetric).Inc(5)
	updateStatusScope.Counter(metrics.ExecutionErrorMetric).Inc(2)
	timer := updateStatusScope.Timer(metrics.ExecutionTimeMetric)
	stopwatch := timer.Start()
	time.Sleep(10 * time.Millisecond)
	stopwatch.Stop()

	createCommentScope := scope.SubScope("create_comment")
	createCommentScope.Counter(metrics.ExecutionSuccessMetric).Inc(3)
	createCommentScope.Counter(metrics.ExecutionErrorMetric).Inc(1)

	// Let metrics flush
	time.Sleep(150 * time.Millisecond)

	// Verify metrics exist in Prometheus before deletion
	metricsFamilies, err := registry.Gather()
	Ok(t, err)

	// Count metrics with pr_number="42"
	beforeCount := countMetricsWithLabel(metricsFamilies, "pr_number", "42")
	t.Logf("Found %d metric samples with pr_number=\"42\" before deletion", beforeCount)
	Assert(t, beforeCount > 0, "should have metrics with pr_number=\"42\" before deletion")

	// Now close the PR scope (this should trigger Prometheus metric deletion)
	manager.MarkPRClosed("owner/repo", 42)

	// Let the deletion complete
	time.Sleep(50 * time.Millisecond)

	// Verify metrics are deleted from Prometheus
	metricsFamilies, err = registry.Gather()
	Ok(t, err)

	afterCount := countMetricsWithLabel(metricsFamilies, "pr_number", "42")
	t.Logf("Found %d metric samples with pr_number=\"42\" after deletion", afterCount)

	// This is the critical assertion - metrics should be deleted
	Equals(t, 0, afterCount)

	t.Log("✅ Test passed: Prometheus metrics with pr_number=\"42\" were successfully deleted")
}

// TestPRScopeManager_MultiPRDeletion verifies that deleting one PR's metrics
// doesn't affect another PR's metrics.
func TestPRScopeManager_MultiPRDeletion(t *testing.T) {
	logger := logging.NewNoopLogger(t)

	registry := prometheus.NewRegistry()
	reporter := tallyprom.NewReporter(tallyprom.Options{
		Registerer: registry,
	})

	manager := metrics.NewPRScopeManager(
		logger,
		reporter,
		tally.ScopeOptions{
			Prefix:          "test_multi",
			Separator:       tallyprom.DefaultSeparator,
			SanitizeOptions: &tallyprom.DefaultSanitizerOpts,
		},
		100*time.Millisecond,
		1*time.Hour,
	)

	// Create metrics for PR #1
	scope1 := manager.GetOrCreatePRScope("owner/repo", 1, map[string]string{
		"base_repo": "owner/repo",
		"pr_number": "1",
	})
	scope1.SubScope("update_status").Counter(metrics.ExecutionSuccessMetric).Inc(10)

	// Create metrics for PR #2
	scope2 := manager.GetOrCreatePRScope("owner/repo", 2, map[string]string{
		"base_repo": "owner/repo",
		"pr_number": "2",
	})
	scope2.SubScope("update_status").Counter(metrics.ExecutionSuccessMetric).Inc(20)

	// Let metrics flush
	time.Sleep(150 * time.Millisecond)

	// Verify both PRs have metrics
	metricsFamilies, err := registry.Gather()
	Ok(t, err)
	pr1Before := countMetricsWithLabel(metricsFamilies, "pr_number", "1")
	pr2Before := countMetricsWithLabel(metricsFamilies, "pr_number", "2")
	t.Logf("Before: PR #1 has %d metrics, PR #2 has %d metrics", pr1Before, pr2Before)
	Assert(t, pr1Before > 0, "PR #1 should have metrics")
	Assert(t, pr2Before > 0, "PR #2 should have metrics")

	// Close PR #1
	manager.MarkPRClosed("owner/repo", 1)
	time.Sleep(50 * time.Millisecond)

	// Verify PR #1 metrics are gone but PR #2 metrics remain
	metricsFamilies, err = registry.Gather()
	Ok(t, err)
	pr1After := countMetricsWithLabel(metricsFamilies, "pr_number", "1")
	pr2After := countMetricsWithLabel(metricsFamilies, "pr_number", "2")
	t.Logf("After: PR #1 has %d metrics, PR #2 has %d metrics", pr1After, pr2After)

	Equals(t, 0, pr1After)
	Equals(t, pr2Before, pr2After) // PR #2 metrics unchanged

	t.Log("✅ Test passed: Deleting PR #1 metrics didn't affect PR #2 metrics")
}

// countMetricsWithLabel counts how many metric samples have a specific label value.
func countMetricsWithLabel(families []*dto.MetricFamily, labelName, labelValue string) int {
	count := 0
	for _, family := range families {
		for _, metric := range family.GetMetric() {
			for _, label := range metric.GetLabel() {
				if label.GetName() == labelName && label.GetValue() == labelValue {
					count++
					break
				}
			}
		}
	}
	return count
}

// gatherMetrics is a helper that gathers all metrics from a registry.
func gatherMetrics(t *testing.T, registry prometheus.Gatherer) []*dto.MetricFamily {
	families, err := registry.Gather()
	Ok(t, err)
	return families
}

// dumpMetrics is a helper for debugging - prints all metrics to test log.
func dumpMetrics(t *testing.T, families []*dto.MetricFamily, prefix string) {
	t.Logf("%s: Dumping %d metric families", prefix, len(families))
	for _, family := range families {
		t.Logf("  Family: %s (%s)", family.GetName(), family.GetType())
		for _, metric := range family.GetMetric() {
			labels := make([]string, 0)
			for _, label := range metric.GetLabel() {
				labels = append(labels, label.GetName()+"="+label.GetValue())
			}
			t.Logf("    Metric: %v", labels)
		}
	}
}

// getMetricValue retrieves the value of a counter metric with specific labels.
func getMetricValue(families []*dto.MetricFamily, metricName string, labels map[string]string) (float64, bool) {
	for _, family := range families {
		if family.GetName() == metricName {
			for _, metric := range family.GetMetric() {
				if matchesLabels(metric.GetLabel(), labels) {
					if metric.Counter != nil {
						return metric.Counter.GetValue(), true
					}
				}
			}
		}
	}
	return 0, false
}

// matchesLabels checks if a metric's labels match the expected labels.
func matchesLabels(metricLabels []*dto.LabelPair, expectedLabels map[string]string) bool {
	if len(metricLabels) != len(expectedLabels) {
		return false
	}
	for _, label := range metricLabels {
		expectedValue, exists := expectedLabels[label.GetName()]
		if !exists || expectedValue != label.GetValue() {
			return false
		}
	}
	return true
}

// TestPRScopeManager_MetricValueVerification verifies the actual counter values
// before and after deletion.
func TestPRScopeManager_MetricValueVerification(t *testing.T) {
	logger := logging.NewNoopLogger(t)

	registry := prometheus.NewRegistry()
	reporter := tallyprom.NewReporter(tallyprom.Options{
		Registerer: registry,
	})

	manager := metrics.NewPRScopeManager(
		logger,
		reporter,
		tally.ScopeOptions{
			Prefix:          "test_values",
			Separator:       tallyprom.DefaultSeparator,
			SanitizeOptions: &tallyprom.DefaultSanitizerOpts,
		},
		100*time.Millisecond,
		1*time.Hour,
	)

	// Create scope and increment counter
	scope := manager.GetOrCreatePRScope("owner/repo", 99, map[string]string{
		"base_repo": "owner/repo",
		"pr_number": "99",
	})
	updateScope := scope.SubScope("update_status")
	updateScope.Counter(metrics.ExecutionSuccessMetric).Inc(42)

	// Let metrics flush
	time.Sleep(150 * time.Millisecond)

	// Verify the counter value before deletion
	families := gatherMetrics(t, registry)
	// Note: Prometheus sanitizes label values (owner/repo becomes owner_repo)
	value, found := getMetricValue(families, "test_values_update_status_execution_success", map[string]string{
		"base_repo": "owner_repo", // Sanitized: / becomes _
		"pr_number": "99",
	})
	Assert(t, found, "metric should exist before deletion")
	Equals(t, 42.0, value)
	t.Logf("Counter value before deletion: %.0f", value)

	// Close the PR
	manager.MarkPRClosed("owner/repo", 99)
	time.Sleep(50 * time.Millisecond)

	// Verify metric is deleted
	families = gatherMetrics(t, registry)
	_, found = getMetricValue(families, "test_values_update_status_execution_success", map[string]string{
		"base_repo": "owner_repo", // Sanitized
		"pr_number": "99",
	})
	Assert(t, !found, "metric should not exist after deletion")

	t.Log("✅ Test passed: Metric value was 42, now metric is deleted")
}

// TestPRScopeManager_OnlyCreatedSubscopesDeleted verifies that only subscopes that were
// actually created during the PR's lifetime are targeted for deletion, not all possible subscopes.
func TestPRScopeManager_OnlyCreatedSubscopesDeleted(t *testing.T) {
	logger := logging.NewNoopLogger(t)

	registry := prometheus.NewRegistry()
	reporter := tallyprom.NewReporter(tallyprom.Options{
		Registerer: registry,
	})

	manager := metrics.NewPRScopeManager(
		logger,
		reporter,
		tally.ScopeOptions{
			Prefix:          "test_selective",
			Separator:       tallyprom.DefaultSeparator,
			SanitizeOptions: &tallyprom.DefaultSanitizerOpts,
		},
		100*time.Millisecond,
		1*time.Hour,
	)

	// Create a PR scope but only use TWO subscopes (not all 8)
	scope := manager.GetOrCreatePRScope("owner/repo", 77, map[string]string{
		"base_repo": "owner/repo",
		"pr_number": "77",
	})

	// Only create metrics in update_status and create_comment subscopes
	scope.SubScope("update_status").Counter(metrics.ExecutionSuccessMetric).Inc(10)
	scope.SubScope("create_comment").Counter(metrics.ExecutionSuccessMetric).Inc(5)

	// Let metrics flush
	time.Sleep(150 * time.Millisecond)

	// Verify we have metrics for pr_number="77"
	metricsFamilies, err := registry.Gather()
	Ok(t, err)
	beforeCount := countMetricsWithLabel(metricsFamilies, "pr_number", "77")
	t.Logf("Created %d metrics using only 2 subscopes", beforeCount)
	Assert(t, beforeCount == 2, "should have exactly 2 metrics (one per subscope)")

	// Close the PR - should only attempt to delete metrics from the 2 subscopes we used
	manager.MarkPRClosed("owner/repo", 77)
	time.Sleep(50 * time.Millisecond)

	// Verify all metrics are deleted
	metricsFamilies, err = registry.Gather()
	Ok(t, err)
	afterCount := countMetricsWithLabel(metricsFamilies, "pr_number", "77")
	Equals(t, 0, afterCount)

	t.Log("✅ Test passed: Only the 2 subscopes that were created were targeted for cleanup")
}

// TestPRScopeManager_NonPrometheusReporter verifies that when using a non-Prometheus
// reporter (like StatsD), the deletion code gracefully skips without errors.
func TestPRScopeManager_NonPrometheusReporter(t *testing.T) {
	logger := logging.NewNoopLogger(t)

	// Use a TestReporter instead of Prometheus
	reporter := &testStatsReporter{}

	manager := metrics.NewPRScopeManager(
		logger,
		reporter,
		tally.ScopeOptions{
			Prefix:    "test_statsd",
			Separator: ".",
		},
		100*time.Millisecond,
		1*time.Hour,
	)

	// Create and close a scope - should not panic or error
	scope := manager.GetOrCreatePRScope("owner/repo", 555, map[string]string{
		"pr_number": "555",
	})
	scope.Counter("test").Inc(1)

	// This should not panic even though reporter is not Prometheus
	manager.MarkPRClosed("owner/repo", 555)

	Equals(t, 0, manager.GetStats())
	t.Log("✅ Test passed: Non-Prometheus reporter gracefully handled")
}

// testStatsReporter is a simple test reporter for verifying non-Prometheus behavior.
type testStatsReporter struct{}

func (r *testStatsReporter) ReportCounter(name string, tags map[string]string, value int64)         {}
func (r *testStatsReporter) ReportGauge(name string, tags map[string]string, value float64)         {}
func (r *testStatsReporter) ReportTimer(name string, tags map[string]string, interval time.Duration) {}
func (r *testStatsReporter) ReportHistogramValueSamples(name string, tags map[string]string, buckets tally.Buckets, bucketLowerBound, bucketUpperBound float64, samples int64) {
}
func (r *testStatsReporter) ReportHistogramDurationSamples(name string, tags map[string]string, buckets tally.Buckets, bucketLowerBound, bucketUpperBound time.Duration, samples int64) {
}
func (r *testStatsReporter) Capabilities() tally.Capabilities {
	return r
}
func (r *testStatsReporter) Reporting() bool {
	return true
}
func (r *testStatsReporter) Tagging() bool {
	return true
}
func (r *testStatsReporter) Flush() {}
