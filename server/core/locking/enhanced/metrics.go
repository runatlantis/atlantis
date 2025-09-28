package enhanced

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/runatlantis/atlantis/server/logging"
)

// MetricsCollector provides comprehensive performance monitoring and analysis
// This is a key component of PR #4 - Enhanced Manager and Events
type MetricsCollector struct {
	logger logging.SimpleLogging

	// Core metrics
	mu               sync.RWMutex
	lockRequests     int64
	lockAcquisitions int64
	lockFailures     int64
	lockReleases     int64

	// Performance metrics
	totalWaitTime time.Duration
	totalHoldTime time.Duration
	minWaitTime   time.Duration
	maxWaitTime   time.Duration
	minHoldTime   time.Duration
	maxHoldTime   time.Duration

	// Priority-based metrics
	priorityMetrics map[Priority]*PriorityMetrics

	// Request type metrics
	requestTypeMetrics map[string]*RequestTypeMetrics

	// Time-based metrics for rate calculations
	startTime      time.Time
	lastUpdateTime time.Time
	metricsHistory []*MetricsSnapshot
	maxHistorySize int

	// Manager state
	started  bool
	stopCh   chan struct{}
	workerWg sync.WaitGroup
}

// PriorityMetrics tracks metrics for each priority level
type PriorityMetrics struct {
	Requests      int64         `json:"requests"`
	Acquisitions  int64         `json:"acquisitions"`
	Failures      int64         `json:"failures"`
	TotalWaitTime time.Duration `json:"total_wait_time"`
	AvgWaitTime   time.Duration `json:"avg_wait_time"`
}

// RequestTypeMetrics tracks metrics for different request types
type RequestTypeMetrics struct {
	Requests      int64         `json:"requests"`
	Acquisitions  int64         `json:"acquisitions"`
	Failures      int64         `json:"failures"`
	TotalWaitTime time.Duration `json:"total_wait_time"`
	AvgWaitTime   time.Duration `json:"avg_wait_time"`
}

// MetricsSnapshot represents metrics at a point in time
type MetricsSnapshot struct {
	Timestamp     time.Time     `json:"timestamp"`
	TotalRequests int64         `json:"total_requests"`
	SuccessRate   float64       `json:"success_rate"`
	AvgWaitTime   time.Duration `json:"avg_wait_time"`
	RequestRate   float64       `json:"request_rate"` // requests per second
	HealthScore   int           `json:"health_score"`
}

// MetricsStats provides comprehensive metrics statistics
type MetricsStats struct {
	// Core counters
	LockRequests     int64 `json:"lock_requests"`
	LockAcquisitions int64 `json:"lock_acquisitions"`
	LockFailures     int64 `json:"lock_failures"`
	LockReleases     int64 `json:"lock_releases"`

	// Success rates
	SuccessRate float64 `json:"success_rate"`
	FailureRate float64 `json:"failure_rate"`

	// Performance metrics
	AvgWaitTime time.Duration `json:"avg_wait_time"`
	MinWaitTime time.Duration `json:"min_wait_time"`
	MaxWaitTime time.Duration `json:"max_wait_time"`
	AvgHoldTime time.Duration `json:"avg_hold_time"`
	MinHoldTime time.Duration `json:"min_hold_time"`
	MaxHoldTime time.Duration `json:"max_hold_time"`

	// Rates
	RequestRate    float64 `json:"request_rate"`    // requests per second
	ThroughputRate float64 `json:"throughput_rate"` // successful acquisitions per second

	// Health and quality
	HealthScore int `json:"health_score"` // 0-100

	// Priority breakdown
	PriorityBreakdown map[Priority]*PriorityMetrics `json:"priority_breakdown"`

	// Request type breakdown
	RequestTypeBreakdown map[string]*RequestTypeMetrics `json:"request_type_breakdown"`

	// Time information
	CollectionStartTime time.Time     `json:"collection_start_time"`
	LastUpdateTime      time.Time     `json:"last_update_time"`
	UptimeDuration      time.Duration `json:"uptime_duration"`
}

// NewMetricsCollector creates a new metrics collector
func NewMetricsCollector(logger logging.SimpleLogging) *MetricsCollector {
	return &MetricsCollector{
		logger:             logger,
		priorityMetrics:    make(map[Priority]*PriorityMetrics),
		requestTypeMetrics: make(map[string]*RequestTypeMetrics),
		metricsHistory:     make([]*MetricsSnapshot, 0),
		maxHistorySize:     1000, // Keep last 1000 snapshots
		stopCh:             make(chan struct{}),
		minWaitTime:        time.Duration(^uint64(0) >> 1), // Max duration
		minHoldTime:        time.Duration(^uint64(0) >> 1), // Max duration
	}
}

// Start starts the metrics collector
func (mc *MetricsCollector) Start(ctx context.Context) error {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	if mc.started {
		return fmt.Errorf("metrics collector already started")
	}

	mc.logger.Info("Starting Enhanced Metrics Collector (PR #4)")

	mc.startTime = time.Now()
	mc.lastUpdateTime = mc.startTime

	// Initialize priority metrics
	for priority := PriorityLow; priority <= PriorityCritical; priority++ {
		mc.priorityMetrics[priority] = &PriorityMetrics{}
	}

	// Start background collector
	mc.workerWg.Add(1)
	go mc.backgroundCollector(ctx)

	mc.started = true
	mc.logger.Info("Enhanced Metrics Collector started")
	return nil
}

// Stop stops the metrics collector
func (mc *MetricsCollector) Stop(ctx context.Context) error {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	if !mc.started {
		return nil
	}

	mc.logger.Info("Stopping Enhanced Metrics Collector (PR #4)")

	// Signal stop
	close(mc.stopCh)

	// Wait for background worker
	done := make(chan struct{})
	go func() {
		mc.workerWg.Wait()
		close(done)
	}()

	select {
	case <-done:
		mc.logger.Info("Metrics collector stopped")
	case <-ctx.Done():
		mc.logger.Warn("Context cancelled while stopping metrics collector")
	}

	mc.started = false
	return nil
}

// RecordLockRequest records a lock request
func (mc *MetricsCollector) RecordLockRequest(requestType string) {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	mc.lockRequests++
	mc.lastUpdateTime = time.Now()

	// Update request type metrics
	if _, exists := mc.requestTypeMetrics[requestType]; !exists {
		mc.requestTypeMetrics[requestType] = &RequestTypeMetrics{}
	}
	mc.requestTypeMetrics[requestType].Requests++
}

// RecordLockAcquisition records a successful lock acquisition
func (mc *MetricsCollector) RecordLockAcquisition(requestType string, waitTime time.Duration) {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	mc.lockAcquisitions++
	mc.totalWaitTime += waitTime
	mc.lastUpdateTime = time.Now()

	// Update min/max wait times
	if waitTime < mc.minWaitTime {
		mc.minWaitTime = waitTime
	}
	if waitTime > mc.maxWaitTime {
		mc.maxWaitTime = waitTime
	}

	// Update request type metrics
	if rtm, exists := mc.requestTypeMetrics[requestType]; exists {
		rtm.Acquisitions++
		rtm.TotalWaitTime += waitTime
		if rtm.Acquisitions > 0 {
			rtm.AvgWaitTime = time.Duration(rtm.TotalWaitTime.Nanoseconds() / rtm.Acquisitions)
		}
	}
}

// RecordLockFailure records a failed lock acquisition
func (mc *MetricsCollector) RecordLockFailure(requestType string, waitTime time.Duration) {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	mc.lockFailures++
	mc.totalWaitTime += waitTime
	mc.lastUpdateTime = time.Now()

	// Update request type metrics
	if rtm, exists := mc.requestTypeMetrics[requestType]; exists {
		rtm.Failures++
		rtm.TotalWaitTime += waitTime
	}
}

// RecordLockRelease records a lock release
func (mc *MetricsCollector) RecordLockRelease(holdTime time.Duration) {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	mc.lockReleases++
	mc.totalHoldTime += holdTime
	mc.lastUpdateTime = time.Now()

	// Update min/max hold times
	if holdTime < mc.minHoldTime {
		mc.minHoldTime = holdTime
	}
	if holdTime > mc.maxHoldTime {
		mc.maxHoldTime = holdTime
	}
}

// RecordPriorityMetrics records metrics for a specific priority
func (mc *MetricsCollector) RecordPriorityMetrics(priority Priority, acquisition bool, waitTime time.Duration) {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	if pm, exists := mc.priorityMetrics[priority]; exists {
		pm.Requests++
		pm.TotalWaitTime += waitTime

		if acquisition {
			pm.Acquisitions++
		} else {
			pm.Failures++
		}

		// Update average wait time
		if pm.Requests > 0 {
			pm.AvgWaitTime = time.Duration(pm.TotalWaitTime.Nanoseconds() / pm.Requests)
		}
	}
}

// RecordEvent records metrics from lock events
func (mc *MetricsCollector) RecordEvent(event *LockEvent) {
	if event == nil {
		return
	}

	// Extract metrics from event metadata
	requestType := "unknown"
	if rt, exists := event.Metadata["method"]; exists {
		requestType = rt
	}

	switch event.Type {
	case "lock_requested", "enhanced_lock_requested":
		mc.RecordLockRequest(requestType)

	case "lock_acquired":
		// Try to extract duration from metadata
		waitTime := time.Duration(0)
		if durationStr, exists := event.Metadata["duration_ms"]; exists {
			if ms, err := time.ParseDuration(durationStr + "ms"); err == nil {
				waitTime = ms
			}
		}
		mc.RecordLockAcquisition(requestType, waitTime)

	case "lock_failed":
		mc.RecordLockFailure(requestType, time.Duration(0))

	case "lock_released":
		mc.RecordLockRelease(time.Duration(0))
	}
}

// GetStats returns comprehensive metrics statistics
func (mc *MetricsCollector) GetStats() *MetricsStats {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	stats := &MetricsStats{
		LockRequests:         mc.lockRequests,
		LockAcquisitions:     mc.lockAcquisitions,
		LockFailures:         mc.lockFailures,
		LockReleases:         mc.lockReleases,
		CollectionStartTime:  mc.startTime,
		LastUpdateTime:       mc.lastUpdateTime,
		UptimeDuration:       time.Since(mc.startTime),
		PriorityBreakdown:    make(map[Priority]*PriorityMetrics),
		RequestTypeBreakdown: make(map[string]*RequestTypeMetrics),
	}

	// Calculate success and failure rates
	if mc.lockRequests > 0 {
		stats.SuccessRate = float64(mc.lockAcquisitions) / float64(mc.lockRequests) * 100
		stats.FailureRate = float64(mc.lockFailures) / float64(mc.lockRequests) * 100
	}

	// Calculate average wait and hold times
	if mc.lockAcquisitions > 0 {
		stats.AvgWaitTime = time.Duration(mc.totalWaitTime.Nanoseconds() / mc.lockAcquisitions)
	}
	if mc.lockReleases > 0 {
		stats.AvgHoldTime = time.Duration(mc.totalHoldTime.Nanoseconds() / mc.lockReleases)
	}

	// Set min/max times (handle initial max duration values)
	if mc.minWaitTime != time.Duration(^uint64(0)>>1) {
		stats.MinWaitTime = mc.minWaitTime
	}
	stats.MaxWaitTime = mc.maxWaitTime

	if mc.minHoldTime != time.Duration(^uint64(0)>>1) {
		stats.MinHoldTime = mc.minHoldTime
	}
	stats.MaxHoldTime = mc.maxHoldTime

	// Calculate rates
	uptime := time.Since(mc.startTime).Seconds()
	if uptime > 0 {
		stats.RequestRate = float64(mc.lockRequests) / uptime
		stats.ThroughputRate = float64(mc.lockAcquisitions) / uptime
	}

	// Calculate health score (0-100)
	stats.HealthScore = mc.calculateHealthScore()

	// Copy priority metrics
	for priority, pm := range mc.priorityMetrics {
		pmCopy := *pm
		stats.PriorityBreakdown[priority] = &pmCopy
	}

	// Copy request type metrics
	for requestType, rtm := range mc.requestTypeMetrics {
		rtmCopy := *rtm
		stats.RequestTypeBreakdown[requestType] = &rtmCopy
	}

	return stats
}

// calculateHealthScore calculates a health score from 0-100 based on various metrics
func (mc *MetricsCollector) calculateHealthScore() int {
	score := 100

	// Deduct points based on failure rate
	if mc.lockRequests > 0 {
		failureRate := float64(mc.lockFailures) / float64(mc.lockRequests) * 100
		score -= int(failureRate * 0.5) // Max 50 points deduction for 100% failure rate
	}

	// Deduct points for very high wait times
	if mc.lockAcquisitions > 0 {
		avgWaitTime := time.Duration(mc.totalWaitTime.Nanoseconds() / mc.lockAcquisitions)
		if avgWaitTime > 30*time.Second {
			score -= 20 // Deduct 20 points for very slow responses
		} else if avgWaitTime > 10*time.Second {
			score -= 10 // Deduct 10 points for slow responses
		}
	}

	// Deduct points for very long max wait times
	if mc.maxWaitTime > 60*time.Second {
		score -= 15 // Deduct 15 points for very long max wait times
	}

	// Ensure score is within bounds
	if score < 0 {
		score = 0
	}
	if score > 100 {
		score = 100
	}

	return score
}

// backgroundCollector runs periodic tasks for metrics collection
func (mc *MetricsCollector) backgroundCollector(ctx context.Context) {
	defer mc.workerWg.Done()

	ticker := time.NewTicker(30 * time.Second) // Take snapshot every 30 seconds
	defer ticker.Stop()

	mc.logger.Debug("Metrics background collector started")
	defer mc.logger.Debug("Metrics background collector stopped")

	for {
		select {
		case <-mc.stopCh:
			return
		case <-ctx.Done():
			return
		case <-ticker.C:
			mc.takeSnapshot()
		}
	}
}

// takeSnapshot creates a snapshot of current metrics
func (mc *MetricsCollector) takeSnapshot() {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	snapshot := &MetricsSnapshot{
		Timestamp:     time.Now(),
		TotalRequests: mc.lockRequests,
		HealthScore:   mc.calculateHealthScore(),
	}

	// Calculate success rate
	if mc.lockRequests > 0 {
		snapshot.SuccessRate = float64(mc.lockAcquisitions) / float64(mc.lockRequests) * 100
	}

	// Calculate average wait time
	if mc.lockAcquisitions > 0 {
		snapshot.AvgWaitTime = time.Duration(mc.totalWaitTime.Nanoseconds() / mc.lockAcquisitions)
	}

	// Calculate request rate (based on last snapshot)
	if len(mc.metricsHistory) > 0 {
		lastSnapshot := mc.metricsHistory[len(mc.metricsHistory)-1]
		timeDiff := snapshot.Timestamp.Sub(lastSnapshot.Timestamp).Seconds()
		if timeDiff > 0 {
			requestDiff := snapshot.TotalRequests - lastSnapshot.TotalRequests
			snapshot.RequestRate = float64(requestDiff) / timeDiff
		}
	}

	// Add to history
	mc.metricsHistory = append(mc.metricsHistory, snapshot)

	// Trim history if too large
	if len(mc.metricsHistory) > mc.maxHistorySize {
		mc.metricsHistory = mc.metricsHistory[1:]
	}
}

// GetHistory returns metrics history
func (mc *MetricsCollector) GetHistory(limit int) []*MetricsSnapshot {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	if limit <= 0 || limit > len(mc.metricsHistory) {
		limit = len(mc.metricsHistory)
	}

	// Return the most recent snapshots
	start := len(mc.metricsHistory) - limit
	if start < 0 {
		start = 0
	}

	history := make([]*MetricsSnapshot, limit)
	copy(history, mc.metricsHistory[start:])
	return history
}

// Reset resets all metrics (useful for testing)
func (mc *MetricsCollector) Reset() {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	mc.lockRequests = 0
	mc.lockAcquisitions = 0
	mc.lockFailures = 0
	mc.lockReleases = 0
	mc.totalWaitTime = 0
	mc.totalHoldTime = 0
	mc.minWaitTime = time.Duration(^uint64(0) >> 1)
	mc.maxWaitTime = 0
	mc.minHoldTime = time.Duration(^uint64(0) >> 1)
	mc.maxHoldTime = 0

	// Reset priority metrics
	for priority := range mc.priorityMetrics {
		mc.priorityMetrics[priority] = &PriorityMetrics{}
	}

	// Reset request type metrics
	mc.requestTypeMetrics = make(map[string]*RequestTypeMetrics)

	// Clear history
	mc.metricsHistory = make([]*MetricsSnapshot, 0)

	mc.startTime = time.Now()
	mc.lastUpdateTime = mc.startTime

	mc.logger.Info("Metrics collector reset")
}
