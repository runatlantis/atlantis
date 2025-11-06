package enhanced

import (
	"context"
	"sync"
	"time"

	"github.com/runatlantis/atlantis/server/logging"
)

// MetricsCollector collects enhanced locking metrics for PR #4
type MetricsCollector struct {
	logger  logging.SimpleLogging
	mu      sync.RWMutex
	running bool

	// Core metrics
	lockRequests     int64
	lockAcquisitions int64
	lockFailures     int64
	lockReleases     int64

	// Performance metrics
	totalWaitTime time.Duration
	totalHoldTime time.Duration
	maxWaitTime   time.Duration
	maxHoldTime   time.Duration

	// Priority metrics
	priorityStats map[Priority]*PriorityMetrics

	// Health metrics
	healthScore      float64
	lastHealthUpdate time.Time
}

// PriorityMetrics tracks per-priority performance
type PriorityMetrics struct {
	Requests      int64         `json:"requests"`
	Acquisitions  int64         `json:"acquisitions"`
	Failures      int64         `json:"failures"`
	AverageWait   time.Duration `json:"average_wait"`
	TotalWaitTime time.Duration `json:"total_wait_time"`
}

// MetricsSummary provides system performance overview
type MetricsSummary struct {
	TotalRequests   int64   `json:"total_requests"`
	SuccessRate     float64 `json:"success_rate"`
	AverageWaitTime int64   `json:"average_wait_time_ms"`
	AverageHoldTime int64   `json:"average_hold_time_ms"`
	HealthScore     float64 `json:"health_score"`
	LastUpdated     time.Time `json:"last_updated"`
}

// NewMetricsCollector creates enhanced metrics collector
func NewMetricsCollector(logger logging.SimpleLogging) *MetricsCollector {
	return &MetricsCollector{
		logger:        logger,
		priorityStats: make(map[Priority]*PriorityMetrics),
		healthScore:   100.0,
		lastHealthUpdate: time.Now(),
	}
}

// Start begins metrics collection
func (mc *MetricsCollector) Start(ctx context.Context) error {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	if mc.running {
		return nil
	}

	mc.running = true
	mc.logger.Info("Enhanced locking metrics collector started (PR #4)")
	return nil
}

// Stop gracefully stops metrics collection
func (mc *MetricsCollector) Stop(ctx context.Context) error {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	mc.running = false
	mc.logger.Info("Enhanced locking metrics collector stopped")
	return nil
}

// RecordLockRequest records a lock request
func (mc *MetricsCollector) RecordLockRequest(priority Priority) {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	mc.lockRequests++
	mc.updatePriorityStats(priority, func(p *PriorityMetrics) {
		p.Requests++
	})
}

// RecordLockAcquisition records successful lock acquisition
func (mc *MetricsCollector) RecordLockAcquisition(priority Priority, waitTime time.Duration) {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	mc.lockAcquisitions++
	mc.totalWaitTime += waitTime

	if waitTime > mc.maxWaitTime {
		mc.maxWaitTime = waitTime
	}

	mc.updatePriorityStats(priority, func(p *PriorityMetrics) {
		p.Acquisitions++
		p.TotalWaitTime += waitTime
		if p.Acquisitions > 0 {
			p.AverageWait = p.TotalWaitTime / time.Duration(p.Acquisitions)
		}
	})
}

// RecordLockFailure records failed lock attempt
func (mc *MetricsCollector) RecordLockFailure(priority Priority, waitTime time.Duration) {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	mc.lockFailures++
	mc.updatePriorityStats(priority, func(p *PriorityMetrics) {
		p.Failures++
	})
}

// RecordLockRelease records lock release
func (mc *MetricsCollector) RecordLockRelease(holdTime time.Duration) {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	mc.lockReleases++
	mc.totalHoldTime += holdTime

	if holdTime > mc.maxHoldTime {
		mc.maxHoldTime = holdTime
	}
}

// RecordCleanup records cleanup operations
func (mc *MetricsCollector) RecordCleanup(cleanedLocks int) {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	// Record cleanup for health scoring
}

// GetMetrics returns current metrics summary
func (mc *MetricsCollector) GetMetrics() *MetricsSummary {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	summary := &MetricsSummary{
		TotalRequests:   mc.lockRequests,
		AverageWaitTime: mc.calculateAverageWaitTime().Milliseconds(),
		AverageHoldTime: mc.calculateAverageHoldTime().Milliseconds(),
		HealthScore:     mc.healthScore,
		LastUpdated:     time.Now(),
	}

	// Calculate success rate
	if mc.lockRequests > 0 {
		summary.SuccessRate = float64(mc.lockAcquisitions) / float64(mc.lockRequests) * 100
	}

	return summary
}

// IsHealthy checks if metrics collector is healthy
func (mc *MetricsCollector) IsHealthy() bool {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	return mc.running && mc.healthScore > 70.0
}

// Private helper methods

func (mc *MetricsCollector) updatePriorityStats(priority Priority, updateFunc func(*PriorityMetrics)) {
	if _, exists := mc.priorityStats[priority]; !exists {
		mc.priorityStats[priority] = &PriorityMetrics{}
	}
	updateFunc(mc.priorityStats[priority])
}

func (mc *MetricsCollector) calculateAverageWaitTime() time.Duration {
	if mc.lockAcquisitions == 0 {
		return 0
	}
	return mc.totalWaitTime / time.Duration(mc.lockAcquisitions)
}

func (mc *MetricsCollector) calculateAverageHoldTime() time.Duration {
	if mc.lockReleases == 0 {
		return 0
	}
	return mc.totalHoldTime / time.Duration(mc.lockReleases)
}