package enhanced

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/runatlantis/atlantis/server/logging"
)

// MetricsCollector aggregates and tracks performance metrics for the enhanced locking system
type MetricsCollector struct {
	manager        *ManagerMetrics
	backend        *BackendMetrics
	deadlock       *DeadlockMetrics
	queue          *QueueMetrics
	events         *EventMetrics
	system         *SystemMetrics
	mutex          sync.RWMutex
	log            logging.SimpleLogging
	collectionInterval time.Duration
	retentionPeriod    time.Duration
	running        bool
	stopChan       chan struct{}
	wg             sync.WaitGroup
}

// ManagerMetrics tracks lock manager performance (redefined for completeness)
type ManagerMetrics struct {
	TotalRequests     int64         `json:"total_requests"`
	SuccessfulLocks   int64         `json:"successful_locks"`
	FailedLocks       int64         `json:"failed_locks"`
	QueuedRequests    int64         `json:"queued_requests"`
	AverageWaitTime   time.Duration `json:"average_wait_time"`
	AverageHoldTime   time.Duration `json:"average_hold_time"`
	ActiveLocks       int64         `json:"active_locks"`
	DeadlocksDetected int64         `json:"deadlocks_detected"`
	DeadlocksResolved int64         `json:"deadlocks_resolved"`
	StartTime         time.Time     `json:"start_time"`
	LastUpdated       time.Time     `json:"last_updated"`
	mutex             sync.RWMutex
}

// BackendMetrics tracks backend performance
type BackendMetrics struct {
	ConnectionPool    *ConnectionPoolMetrics `json:"connection_pool,omitempty"`
	OperationLatency  *LatencyMetrics        `json:"operation_latency"`
	ErrorRates        *ErrorRateMetrics      `json:"error_rates"`
	ThroughputMetrics *ThroughputMetrics     `json:"throughput_metrics"`
	ResourceUsage     *ResourceUsageMetrics  `json:"resource_usage"`
	LastUpdated       time.Time              `json:"last_updated"`
	mutex             sync.RWMutex
}

// ConnectionPoolMetrics tracks connection pool performance
type ConnectionPoolMetrics struct {
	ActiveConnections int           `json:"active_connections"`
	IdleConnections   int           `json:"idle_connections"`
	MaxConnections    int           `json:"max_connections"`
	ConnectionErrors  int64         `json:"connection_errors"`
	AverageLatency    time.Duration `json:"average_latency"`
	PoolUtilization   float64       `json:"pool_utilization"`
}

// LatencyMetrics tracks operation latency
type LatencyMetrics struct {
	AcquireLock  *LatencyStats `json:"acquire_lock"`
	ReleaseLock  *LatencyStats `json:"release_lock"`
	ListLocks    *LatencyStats `json:"list_locks"`
	HealthCheck  *LatencyStats `json:"health_check"`
	QueuePush    *LatencyStats `json:"queue_push"`
	QueuePop     *LatencyStats `json:"queue_pop"`
}

// LatencyStats contains statistical data for latencies
type LatencyStats struct {
	Count     int64         `json:"count"`
	Average   time.Duration `json:"average"`
	Min       time.Duration `json:"min"`
	Max       time.Duration `json:"max"`
	P50       time.Duration `json:"p50"`
	P90       time.Duration `json:"p90"`
	P95       time.Duration `json:"p95"`
	P99       time.Duration `json:"p99"`
	Recent    []time.Duration `json:"-"` // Keep recent samples for percentile calculation
	MaxSamples int            `json:"-"`
}

// ErrorRateMetrics tracks error rates by operation
type ErrorRateMetrics struct {
	AcquireErrors   *ErrorStats `json:"acquire_errors"`
	ReleaseErrors   *ErrorStats `json:"release_errors"`
	TimeoutErrors   *ErrorStats `json:"timeout_errors"`
	BackendErrors   *ErrorStats `json:"backend_errors"`
	ValidationErrors *ErrorStats `json:"validation_errors"`
	TotalErrorRate  float64     `json:"total_error_rate"`
	LastErrorTime   *time.Time  `json:"last_error_time,omitempty"`
}

// ErrorStats contains error statistics
type ErrorStats struct {
	Count      int64              `json:"count"`
	Rate       float64            `json:"rate"` // errors per second
	LastError  string             `json:"last_error,omitempty"`
	ErrorTypes map[string]int64   `json:"error_types"`
	Recent     []time.Time        `json:"-"` // Recent error timestamps
	MaxSamples int                `json:"-"`
}

// ThroughputMetrics tracks system throughput
type ThroughputMetrics struct {
	RequestsPerSecond  float64 `json:"requests_per_second"`
	LocksPerSecond     float64 `json:"locks_per_second"`
	ReleasesPerSecond  float64 `json:"releases_per_second"`
	QueueOpsPerSecond  float64 `json:"queue_ops_per_second"`
	PeakThroughput     float64 `json:"peak_throughput"`
	CurrentLoad        float64 `json:"current_load"` // 0.0 to 1.0
}

// ResourceUsageMetrics tracks resource consumption
type ResourceUsageMetrics struct {
	MemoryUsage    *MemoryMetrics `json:"memory_usage"`
	CPUUsage       *CPUMetrics    `json:"cpu_usage"`
	NetworkUsage   *NetworkMetrics `json:"network_usage"`
	StorageUsage   *StorageMetrics `json:"storage_usage"`
}

// MemoryMetrics tracks memory usage
type MemoryMetrics struct {
	AllocatedBytes int64   `json:"allocated_bytes"`
	InUseBytes     int64   `json:"in_use_bytes"`
	SystemBytes    int64   `json:"system_bytes"`
	GCPauses       int64   `json:"gc_pauses"`
	HeapSize       int64   `json:"heap_size"`
	Utilization    float64 `json:"utilization"`
}

// CPUMetrics tracks CPU usage
type CPUMetrics struct {
	Usage         float64 `json:"usage"` // 0.0 to 1.0
	SystemTime    float64 `json:"system_time"`
	UserTime      float64 `json:"user_time"`
	IdleTime      float64 `json:"idle_time"`
	LoadAverage   float64 `json:"load_average"`
}

// NetworkMetrics tracks network usage
type NetworkMetrics struct {
	BytesSent     int64 `json:"bytes_sent"`
	BytesReceived int64 `json:"bytes_received"`
	PacketsSent   int64 `json:"packets_sent"`
	PacketsReceived int64 `json:"packets_received"`
	ConnectionCount int   `json:"connection_count"`
	Bandwidth     float64 `json:"bandwidth_utilization"`
}

// StorageMetrics tracks storage usage
type StorageMetrics struct {
	DiskUsage      int64   `json:"disk_usage"`
	DiskAvailable  int64   `json:"disk_available"`
	DiskTotal      int64   `json:"disk_total"`
	IOOperations   int64   `json:"io_operations"`
	IOBandwidth    float64 `json:"io_bandwidth"`
	Utilization    float64 `json:"utilization"`
}

// DeadlockMetrics tracks deadlock detection and resolution
type DeadlockMetrics struct {
	TotalDetected      int64              `json:"total_detected"`
	TotalResolved      int64              `json:"total_resolved"`
	ResolutionMethods  map[string]int64   `json:"resolution_methods"`
	AverageResolutionTime time.Duration   `json:"average_resolution_time"`
	CurrentDeadlocks   int                `json:"current_deadlocks"`
	PreventedDeadlocks int64              `json:"prevented_deadlocks"`
	FalsePositives     int64              `json:"false_positives"`
	DetectionLatency   time.Duration      `json:"detection_latency"`
	LastDetected       *time.Time         `json:"last_detected,omitempty"`
	mutex              sync.RWMutex
}

// QueueMetrics tracks queue performance
type QueueMetrics struct {
	TotalQueued       int64                  `json:"total_queued"`
	TotalProcessed    int64                  `json:"total_processed"`
	CurrentDepth      int                    `json:"current_depth"`
	MaxDepth          int                    `json:"max_depth"`
	AverageWaitTime   time.Duration          `json:"average_wait_time"`
	DepthByPriority   map[Priority]int       `json:"depth_by_priority"`
	ThroughputPerSec  float64                `json:"throughput_per_second"`
	OldestQueueTime   *time.Time             `json:"oldest_queue_time,omitempty"`
	QueueUtilization  float64                `json:"queue_utilization"`
	DroppedRequests   int64                  `json:"dropped_requests"`
	mutex             sync.RWMutex
}

// SystemMetrics tracks overall system health
type SystemMetrics struct {
	Uptime            time.Duration      `json:"uptime"`
	HealthScore       int                `json:"health_score"` // 0-100
	ComponentHealth   map[string]int     `json:"component_health"`
	AlertLevel        AlertLevel         `json:"alert_level"`
	ActiveAlerts      []string           `json:"active_alerts"`
	LastHealthCheck   time.Time          `json:"last_health_check"`
	PerformanceIndex  float64            `json:"performance_index"` // 0.0 to 1.0
	ReliabilityScore  float64            `json:"reliability_score"` // 0.0 to 1.0
	mutex             sync.RWMutex
}

// AlertLevel defines system alert levels
type AlertLevel string

const (
	AlertLevelNone     AlertLevel = "none"
	AlertLevelInfo     AlertLevel = "info"
	AlertLevelWarning  AlertLevel = "warning"
	AlertLevelError    AlertLevel = "error"
	AlertLevelCritical AlertLevel = "critical"
)

// MetricsSnapshot represents a point-in-time view of all metrics
type MetricsSnapshot struct {
	Timestamp    time.Time        `json:"timestamp"`
	Manager      *ManagerMetrics  `json:"manager"`
	Backend      *BackendMetrics  `json:"backend"`
	Deadlock     *DeadlockMetrics `json:"deadlock"`
	Queue        *QueueMetrics    `json:"queue"`
	Events       *EventMetrics    `json:"events"`
	System       *SystemMetrics   `json:"system"`
	Summary      *MetricsSummary  `json:"summary"`
}

// MetricsSummary provides high-level system metrics
type MetricsSummary struct {
	OverallHealth       int           `json:"overall_health"` // 0-100
	PerformanceScore    float64       `json:"performance_score"` // 0.0 to 1.0
	ReliabilityScore    float64       `json:"reliability_score"` // 0.0 to 1.0
	TotalOperations     int64         `json:"total_operations"`
	ErrorRate           float64       `json:"error_rate"`
	AverageLatency      time.Duration `json:"average_latency"`
	ThroughputPerSecond float64       `json:"throughput_per_second"`
	ResourceUtilization float64       `json:"resource_utilization"`
	AlertLevel          AlertLevel    `json:"alert_level"`
	Recommendations     []string      `json:"recommendations,omitempty"`
}

// NewMetricsCollector creates a new metrics collector
func NewMetricsCollector(log logging.SimpleLogging) *MetricsCollector {
	return &MetricsCollector{
		manager:  &ManagerMetrics{StartTime: time.Now(), LastUpdated: time.Now()},
		backend:  &BackendMetrics{LastUpdated: time.Now()},
		deadlock: &DeadlockMetrics{ResolutionMethods: make(map[string]int64)},
		queue:    &QueueMetrics{DepthByPriority: make(map[Priority]int)},
		events:   &EventMetrics{EventsByType: make(map[EventType]int64), LastUpdated: time.Now()},
		system:   &SystemMetrics{ComponentHealth: make(map[string]int), LastUpdated: time.Now()},
		log:      log,
		collectionInterval: 30 * time.Second,
		retentionPeriod:    24 * time.Hour,
		stopChan: make(chan struct{}),
	}
}

// Start begins metrics collection
func (mc *MetricsCollector) Start(ctx context.Context) error {
	mc.mutex.Lock()
	defer mc.mutex.Unlock()

	if mc.running {
		return fmt.Errorf("metrics collector is already running")
	}

	mc.running = true
	mc.wg.Add(1)
	go mc.collectMetrics(ctx)

	mc.log.Info("Metrics collector started with %v collection interval", mc.collectionInterval)
	return nil
}

// Stop gracefully shuts down the metrics collector
func (mc *MetricsCollector) Stop() error {
	mc.mutex.Lock()
	defer mc.mutex.Unlock()

	if !mc.running {
		return nil
	}

	mc.log.Info("Stopping metrics collector...")
	close(mc.stopChan)
	mc.running = false

	mc.wg.Wait()

	mc.log.Info("Metrics collector stopped")
	return nil
}

// GetSnapshot returns a current snapshot of all metrics
func (mc *MetricsCollector) GetSnapshot() *MetricsSnapshot {
	mc.mutex.RLock()
	defer mc.mutex.RUnlock()

	snapshot := &MetricsSnapshot{
		Timestamp: time.Now(),
		Manager:   mc.copyManagerMetrics(),
		Backend:   mc.copyBackendMetrics(),
		Deadlock:  mc.copyDeadlockMetrics(),
		Queue:     mc.copyQueueMetrics(),
		Events:    mc.copyEventMetrics(),
		System:    mc.copySystemMetrics(),
	}

	snapshot.Summary = mc.calculateSummary(snapshot)
	return snapshot
}

// GetManagerMetrics returns current manager metrics
func (mc *MetricsCollector) GetManagerMetrics() *ManagerMetrics {
	mc.mutex.RLock()
	defer mc.mutex.RUnlock()
	return mc.copyManagerMetrics()
}

// UpdateManagerMetrics updates manager metrics
func (mc *MetricsCollector) UpdateManagerMetrics(update func(*ManagerMetrics)) {
	mc.manager.mutex.Lock()
	defer mc.manager.mutex.Unlock()
	update(mc.manager)
	mc.manager.LastUpdated = time.Now()
}

// RecordLatency records operation latency
func (mc *MetricsCollector) RecordLatency(operation string, latency time.Duration) {
	mc.backend.mutex.Lock()
	defer mc.backend.mutex.Unlock()

	if mc.backend.OperationLatency == nil {
		mc.backend.OperationLatency = &LatencyMetrics{}
	}

	var stats *LatencyStats
	switch operation {
	case "acquire_lock":
		if mc.backend.OperationLatency.AcquireLock == nil {
			mc.backend.OperationLatency.AcquireLock = newLatencyStats()
		}
		stats = mc.backend.OperationLatency.AcquireLock
	case "release_lock":
		if mc.backend.OperationLatency.ReleaseLock == nil {
			mc.backend.OperationLatency.ReleaseLock = newLatencyStats()
		}
		stats = mc.backend.OperationLatency.ReleaseLock
	case "list_locks":
		if mc.backend.OperationLatency.ListLocks == nil {
			mc.backend.OperationLatency.ListLocks = newLatencyStats()
		}
		stats = mc.backend.OperationLatency.ListLocks
	case "health_check":
		if mc.backend.OperationLatency.HealthCheck == nil {
			mc.backend.OperationLatency.HealthCheck = newLatencyStats()
		}
		stats = mc.backend.OperationLatency.HealthCheck
	default:
		return
	}

	mc.updateLatencyStats(stats, latency)
	mc.backend.LastUpdated = time.Now()
}

// RecordError records an error occurrence
func (mc *MetricsCollector) RecordError(errorType string, errorMsg string) {
	mc.backend.mutex.Lock()
	defer mc.backend.mutex.Unlock()

	if mc.backend.ErrorRates == nil {
		mc.backend.ErrorRates = &ErrorRateMetrics{}
	}

	var errorStats *ErrorStats
	switch errorType {
	case "acquire":
		if mc.backend.ErrorRates.AcquireErrors == nil {
			mc.backend.ErrorRates.AcquireErrors = newErrorStats()
		}
		errorStats = mc.backend.ErrorRates.AcquireErrors
	case "release":
		if mc.backend.ErrorRates.ReleaseErrors == nil {
			mc.backend.ErrorRates.ReleaseErrors = newErrorStats()
		}
		errorStats = mc.backend.ErrorRates.ReleaseErrors
	case "timeout":
		if mc.backend.ErrorRates.TimeoutErrors == nil {
			mc.backend.ErrorRates.TimeoutErrors = newErrorStats()
		}
		errorStats = mc.backend.ErrorRates.TimeoutErrors
	case "backend":
		if mc.backend.ErrorRates.BackendErrors == nil {
			mc.backend.ErrorRates.BackendErrors = newErrorStats()
		}
		errorStats = mc.backend.ErrorRates.BackendErrors
	case "validation":
		if mc.backend.ErrorRates.ValidationErrors == nil {
			mc.backend.ErrorRates.ValidationErrors = newErrorStats()
		}
		errorStats = mc.backend.ErrorRates.ValidationErrors
	default:
		return
	}

	mc.updateErrorStats(errorStats, errorMsg)
	now := time.Now()
	mc.backend.ErrorRates.LastErrorTime = &now
	mc.backend.LastUpdated = time.Now()
}

// GetHealthScore calculates overall system health score (0-100)
func (mc *MetricsCollector) GetHealthScore() int {
	mc.mutex.RLock()
	defer mc.mutex.RUnlock()

	return mc.calculateHealthScore()
}

// GetRecommendations returns system optimization recommendations
func (mc *MetricsCollector) GetRecommendations() []string {
	snapshot := mc.GetSnapshot()
	var recommendations []string

	// Check error rates
	if snapshot.Backend.ErrorRates != nil && snapshot.Backend.ErrorRates.TotalErrorRate > 0.05 {
		recommendations = append(recommendations, "High error rate detected. Consider investigating backend connectivity or configuration.")
	}

	// Check queue depth
	if snapshot.Queue.CurrentDepth > snapshot.Queue.MaxDepth*8/10 {
		recommendations = append(recommendations, "Queue utilization is high. Consider increasing queue capacity or optimizing lock processing.")
	}

	// Check deadlocks
	if snapshot.Deadlock.CurrentDeadlocks > 0 {
		recommendations = append(recommendations, "Active deadlocks detected. Review lock acquisition patterns and consider implementing timeout strategies.")
	}

	// Check latency
	if snapshot.Backend.OperationLatency != nil && snapshot.Backend.OperationLatency.AcquireLock != nil &&
		snapshot.Backend.OperationLatency.AcquireLock.Average > 5*time.Second {
		recommendations = append(recommendations, "High lock acquisition latency. Consider optimizing backend configuration or scaling resources.")
	}

	// Check resource utilization
	if snapshot.Summary.ResourceUtilization > 0.9 {
		recommendations = append(recommendations, "High resource utilization. Consider scaling the system or optimizing resource usage.")
	}

	return recommendations
}

// Helper methods

func (mc *MetricsCollector) collectMetrics(ctx context.Context) {
	defer mc.wg.Done()

	ticker := time.NewTicker(mc.collectionInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-mc.stopChan:
			return
		case <-ticker.C:
			mc.performCollection()
		}
	}
}

func (mc *MetricsCollector) performCollection() {
	// Update system metrics
	mc.system.mutex.Lock()
	mc.system.Uptime = time.Since(mc.manager.StartTime)
	mc.system.HealthScore = mc.calculateHealthScore()
	mc.system.LastHealthCheck = time.Now()
	mc.system.mutex.Unlock()

	// Update throughput metrics
	mc.updateThroughputMetrics()

	// Clean up old samples
	mc.cleanupOldSamples()
}

func (mc *MetricsCollector) calculateHealthScore() int {
	score := 100

	// Deduct for errors
	if mc.backend.ErrorRates != nil && mc.backend.ErrorRates.TotalErrorRate > 0 {
		score -= int(mc.backend.ErrorRates.TotalErrorRate * 50) // Max 50 point deduction
	}

	// Deduct for deadlocks
	if mc.deadlock.CurrentDeadlocks > 0 {
		score -= mc.deadlock.CurrentDeadlocks * 10 // 10 points per deadlock
	}

	// Deduct for high queue utilization
	if mc.queue.QueueUtilization > 0.8 {
		score -= int((mc.queue.QueueUtilization - 0.8) * 100) // Up to 20 point deduction
	}

	if score < 0 {
		score = 0
	}

	return score
}

func (mc *MetricsCollector) updateThroughputMetrics() {
	if mc.backend.ThroughputMetrics == nil {
		mc.backend.ThroughputMetrics = &ThroughputMetrics{}
	}

	// Calculate requests per second based on recent activity
	// This is a simplified calculation - in production, use a sliding window
	timeDiff := time.Since(mc.manager.LastUpdated).Seconds()
	if timeDiff > 0 {
		mc.backend.ThroughputMetrics.RequestsPerSecond = float64(mc.manager.TotalRequests) / timeDiff
		mc.backend.ThroughputMetrics.LocksPerSecond = float64(mc.manager.SuccessfulLocks) / timeDiff
	}
}

func (mc *MetricsCollector) cleanupOldSamples() {
	cutoff := time.Now().Add(-mc.retentionPeriod)

	// Clean up latency samples
	if mc.backend.OperationLatency != nil {
		if mc.backend.OperationLatency.AcquireLock != nil {
			mc.cleanupLatencySamples(mc.backend.OperationLatency.AcquireLock, cutoff)
		}
		if mc.backend.OperationLatency.ReleaseLock != nil {
			mc.cleanupLatencySamples(mc.backend.OperationLatency.ReleaseLock, cutoff)
		}
	}

	// Clean up error samples
	if mc.backend.ErrorRates != nil {
		if mc.backend.ErrorRates.AcquireErrors != nil {
			mc.cleanupErrorSamples(mc.backend.ErrorRates.AcquireErrors, cutoff)
		}
		if mc.backend.ErrorRates.ReleaseErrors != nil {
			mc.cleanupErrorSamples(mc.backend.ErrorRates.ReleaseErrors, cutoff)
		}
	}
}

func (mc *MetricsCollector) cleanupLatencySamples(stats *LatencyStats, cutoff time.Time) {
	// This is a simplified cleanup - in production, you'd track timestamps with samples
	if len(stats.Recent) > stats.MaxSamples {
		stats.Recent = stats.Recent[len(stats.Recent)-stats.MaxSamples:]
	}
}

func (mc *MetricsCollector) cleanupErrorSamples(stats *ErrorStats, cutoff time.Time) {
	var keep []time.Time
	for _, timestamp := range stats.Recent {
		if timestamp.After(cutoff) {
			keep = append(keep, timestamp)
		}
	}
	stats.Recent = keep
}

// Copy methods for thread-safe access

func (mc *MetricsCollector) copyManagerMetrics() *ManagerMetrics {
	mc.manager.mutex.RLock()
	defer mc.manager.mutex.RUnlock()

	return &ManagerMetrics{
		TotalRequests:     mc.manager.TotalRequests,
		SuccessfulLocks:   mc.manager.SuccessfulLocks,
		FailedLocks:       mc.manager.FailedLocks,
		QueuedRequests:    mc.manager.QueuedRequests,
		AverageWaitTime:   mc.manager.AverageWaitTime,
		AverageHoldTime:   mc.manager.AverageHoldTime,
		ActiveLocks:       mc.manager.ActiveLocks,
		DeadlocksDetected: mc.manager.DeadlocksDetected,
		DeadlocksResolved: mc.manager.DeadlocksResolved,
		StartTime:         mc.manager.StartTime,
		LastUpdated:       mc.manager.LastUpdated,
	}
}

func (mc *MetricsCollector) copyBackendMetrics() *BackendMetrics {
	mc.backend.mutex.RLock()
	defer mc.backend.mutex.RUnlock()

	// Deep copy of backend metrics
	copied := &BackendMetrics{
		LastUpdated: mc.backend.LastUpdated,
	}

	if mc.backend.OperationLatency != nil {
		copied.OperationLatency = &LatencyMetrics{}
		// Copy each latency stat if it exists
		// ... (implement deep copy as needed)
	}

	return copied
}

func (mc *MetricsCollector) copyDeadlockMetrics() *DeadlockMetrics {
	mc.deadlock.mutex.RLock()
	defer mc.deadlock.mutex.RUnlock()

	methods := make(map[string]int64)
	for k, v := range mc.deadlock.ResolutionMethods {
		methods[k] = v
	}

	return &DeadlockMetrics{
		TotalDetected:         mc.deadlock.TotalDetected,
		TotalResolved:         mc.deadlock.TotalResolved,
		ResolutionMethods:     methods,
		AverageResolutionTime: mc.deadlock.AverageResolutionTime,
		CurrentDeadlocks:      mc.deadlock.CurrentDeadlocks,
		PreventedDeadlocks:    mc.deadlock.PreventedDeadlocks,
		FalsePositives:        mc.deadlock.FalsePositives,
		DetectionLatency:      mc.deadlock.DetectionLatency,
		LastDetected:          mc.deadlock.LastDetected,
	}
}

func (mc *MetricsCollector) copyQueueMetrics() *QueueMetrics {
	mc.queue.mutex.RLock()
	defer mc.queue.mutex.RUnlock()

	depth := make(map[Priority]int)
	for k, v := range mc.queue.DepthByPriority {
		depth[k] = v
	}

	return &QueueMetrics{
		TotalQueued:       mc.queue.TotalQueued,
		TotalProcessed:    mc.queue.TotalProcessed,
		CurrentDepth:      mc.queue.CurrentDepth,
		MaxDepth:          mc.queue.MaxDepth,
		AverageWaitTime:   mc.queue.AverageWaitTime,
		DepthByPriority:   depth,
		ThroughputPerSec:  mc.queue.ThroughputPerSec,
		OldestQueueTime:   mc.queue.OldestQueueTime,
		QueueUtilization:  mc.queue.QueueUtilization,
		DroppedRequests:   mc.queue.DroppedRequests,
	}
}

func (mc *MetricsCollector) copyEventMetrics() *EventMetrics {
	mc.events.mutex.RLock()
	defer mc.events.mutex.RUnlock()

	events := make(map[EventType]int64)
	for k, v := range mc.events.EventsByType {
		events[k] = v
	}

	return &EventMetrics{
		TotalEvents:         mc.events.TotalEvents,
		EventsByType:        events,
		SubscriptionCount:   mc.events.SubscriptionCount,
		ActiveSubscriptions: mc.events.ActiveSubscriptions,
		FailedDeliveries:    mc.events.FailedDeliveries,
		AverageDeliveryTime: mc.events.AverageDeliveryTime,
		BufferUtilization:   mc.events.BufferUtilization,
		LastUpdated:         mc.events.LastUpdated,
	}
}

func (mc *MetricsCollector) copySystemMetrics() *SystemMetrics {
	mc.system.mutex.RLock()
	defer mc.system.mutex.RUnlock()

	health := make(map[string]int)
	for k, v := range mc.system.ComponentHealth {
		health[k] = v
	}

	alerts := make([]string, len(mc.system.ActiveAlerts))
	copy(alerts, mc.system.ActiveAlerts)

	return &SystemMetrics{
		Uptime:           mc.system.Uptime,
		HealthScore:      mc.system.HealthScore,
		ComponentHealth:  health,
		AlertLevel:       mc.system.AlertLevel,
		ActiveAlerts:     alerts,
		LastHealthCheck:  mc.system.LastHealthCheck,
		PerformanceIndex: mc.system.PerformanceIndex,
		ReliabilityScore: mc.system.ReliabilityScore,
	}
}

func (mc *MetricsCollector) calculateSummary(snapshot *MetricsSnapshot) *MetricsSummary {
	summary := &MetricsSummary{
		OverallHealth:    snapshot.System.HealthScore,
		PerformanceScore: snapshot.System.PerformanceIndex,
		ReliabilityScore: snapshot.System.ReliabilityScore,
		TotalOperations:  snapshot.Manager.TotalRequests,
		AlertLevel:       snapshot.System.AlertLevel,
	}

	// Calculate error rate
	if snapshot.Manager.TotalRequests > 0 {
		summary.ErrorRate = float64(snapshot.Manager.FailedLocks) / float64(snapshot.Manager.TotalRequests)
	}

	// Set average latency
	if snapshot.Backend.OperationLatency != nil && snapshot.Backend.OperationLatency.AcquireLock != nil {
		summary.AverageLatency = snapshot.Backend.OperationLatency.AcquireLock.Average
	}

	// Set throughput
	if snapshot.Backend.ThroughputMetrics != nil {
		summary.ThroughputPerSecond = snapshot.Backend.ThroughputMetrics.RequestsPerSecond
	}

	// Get recommendations
	summary.Recommendations = mc.GetRecommendations()

	return summary
}

// Helper functions for creating new metric structures

func newLatencyStats() *LatencyStats {
	return &LatencyStats{
		MaxSamples: 1000,
		Recent:     make([]time.Duration, 0, 1000),
		Min:        time.Hour, // Initialize to high value
	}
}

func newErrorStats() *ErrorStats {
	return &ErrorStats{
		ErrorTypes: make(map[string]int64),
		Recent:     make([]time.Time, 0, 1000),
		MaxSamples: 1000,
	}
}

func (mc *MetricsCollector) updateLatencyStats(stats *LatencyStats, latency time.Duration) {
	stats.Count++

	// Update min/max
	if latency < stats.Min {
		stats.Min = latency
	}
	if latency > stats.Max {
		stats.Max = latency
	}

	// Update average
	if stats.Average == 0 {
		stats.Average = latency
	} else {
		stats.Average = (stats.Average + latency) / 2
	}

	// Add to recent samples for percentile calculation
	stats.Recent = append(stats.Recent, latency)
	if len(stats.Recent) > stats.MaxSamples {
		stats.Recent = stats.Recent[1:]
	}

	// Calculate percentiles (simplified)
	mc.calculatePercentiles(stats)
}

func (mc *MetricsCollector) updateErrorStats(stats *ErrorStats, errorMsg string) {
	stats.Count++
	stats.LastError = errorMsg

	// Update error types
	if stats.ErrorTypes == nil {
		stats.ErrorTypes = make(map[string]int64)
	}
	stats.ErrorTypes[errorMsg]++

	// Add to recent timestamps
	stats.Recent = append(stats.Recent, time.Now())
	if len(stats.Recent) > stats.MaxSamples {
		stats.Recent = stats.Recent[1:]
	}

	// Calculate rate (errors per second in last minute)
	cutoff := time.Now().Add(-time.Minute)
	var recentCount int64
	for _, timestamp := range stats.Recent {
		if timestamp.After(cutoff) {
			recentCount++
		}
	}
	stats.Rate = float64(recentCount) / 60.0
}

func (mc *MetricsCollector) calculatePercentiles(stats *LatencyStats) {
	if len(stats.Recent) == 0 {
		return
	}

	// Sort samples for percentile calculation
	samples := make([]time.Duration, len(stats.Recent))
	copy(samples, stats.Recent)

	// Simple sort (in production, use more efficient method)
	for i := 0; i < len(samples); i++ {
		for j := i + 1; j < len(samples); j++ {
			if samples[i] > samples[j] {
				samples[i], samples[j] = samples[j], samples[i]
			}
		}
	}

	n := len(samples)
	stats.P50 = samples[n/2]
	stats.P90 = samples[int(float64(n)*0.9)]
	stats.P95 = samples[int(float64(n)*0.95)]
	stats.P99 = samples[int(float64(n)*0.99)]
}

// ExportMetrics exports metrics in various formats
func (mc *MetricsCollector) ExportMetrics(format string) ([]byte, error) {
	snapshot := mc.GetSnapshot()

	switch format {
	case "json":
		return json.MarshalIndent(snapshot, "", "  ")
	case "prometheus":
		return mc.exportPrometheus(snapshot), nil
	default:
		return nil, fmt.Errorf("unsupported export format: %s", format)
	}
}

func (mc *MetricsCollector) exportPrometheus(snapshot *MetricsSnapshot) []byte {
	// Simplified Prometheus export - in production, use proper Prometheus client
	var output []string

	output = append(output, fmt.Sprintf("# HELP atlantis_locks_total Total number of lock requests"))
	output = append(output, fmt.Sprintf("# TYPE atlantis_locks_total counter"))
	output = append(output, fmt.Sprintf("atlantis_locks_total %d", snapshot.Manager.TotalRequests))

	output = append(output, fmt.Sprintf("# HELP atlantis_locks_active Number of currently active locks"))
	output = append(output, fmt.Sprintf("# TYPE atlantis_locks_active gauge"))
	output = append(output, fmt.Sprintf("atlantis_locks_active %d", snapshot.Manager.ActiveLocks))

	output = append(output, fmt.Sprintf("# HELP atlantis_system_health System health score"))
	output = append(output, fmt.Sprintf("# TYPE atlantis_system_health gauge"))
	output = append(output, fmt.Sprintf("atlantis_system_health %d", snapshot.System.HealthScore))

	return []byte(fmt.Sprintf("%s\n", output))
}