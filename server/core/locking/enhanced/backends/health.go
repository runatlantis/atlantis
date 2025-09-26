package backends

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/runatlantis/atlantis/server/logging"
)

// HealthMonitor provides comprehensive health monitoring for Redis backend
type HealthMonitor struct {
	client       redis.UniversalClient
	log          logging.SimpleLogging
	scriptMgr    *ScriptManager

	// Circuit breaker
	circuitBreaker *CircuitBreaker

	// Health metrics
	metrics      *HealthMetrics
	mutex        sync.RWMutex

	// Monitoring configuration
	config       *HealthConfig

	// Background monitoring
	stopChan     chan struct{}
	running      bool
}

// HealthConfig configures health monitoring behavior
type HealthConfig struct {
	CheckInterval        time.Duration `mapstructure:"check_interval"`
	TimeoutThreshold     time.Duration `mapstructure:"timeout_threshold"`
	ErrorThreshold       int           `mapstructure:"error_threshold"`
	RecoveryThreshold    int           `mapstructure:"recovery_threshold"`
	CircuitBreakerEnabled bool         `mapstructure:"circuit_breaker_enabled"`
	PerformanceMonitoring bool         `mapstructure:"performance_monitoring"`
	AlertingEnabled      bool          `mapstructure:"alerting_enabled"`
	MaxSlowQueries       int           `mapstructure:"max_slow_queries"`
	MemoryThresholdMB    int64         `mapstructure:"memory_threshold_mb"`
}

// DefaultHealthConfig returns default health monitoring configuration
func DefaultHealthConfig() *HealthConfig {
	return &HealthConfig{
		CheckInterval:         30 * time.Second,
		TimeoutThreshold:      5 * time.Second,
		ErrorThreshold:        5,
		RecoveryThreshold:     3,
		CircuitBreakerEnabled: true,
		PerformanceMonitoring: true,
		AlertingEnabled:       false,
		MaxSlowQueries:        10,
		MemoryThresholdMB:     1024,
	}
}

// HealthMetrics tracks health monitoring statistics
type HealthMetrics struct {
	Status                string        `json:"status"`
	LastCheck             time.Time     `json:"last_check"`
	ConsecutiveFailures   int           `json:"consecutive_failures"`
	ConsecutiveSuccesses  int           `json:"consecutive_successes"`
	TotalChecks           int64         `json:"total_checks"`
	SuccessfulChecks      int64         `json:"successful_checks"`
	FailedChecks          int64         `json:"failed_checks"`
	AverageResponseTime   time.Duration `json:"average_response_time"`
	MaxResponseTime       time.Duration `json:"max_response_time"`
	MinResponseTime       time.Duration `json:"min_response_time"`
	CircuitBreakerState   string        `json:"circuit_breaker_state"`
	MemoryUsage           int64         `json:"memory_usage_bytes"`
	ConnectionCount       int           `json:"connection_count"`
	SlowQueryCount        int           `json:"slow_query_count"`
	ReplicationLag        time.Duration `json:"replication_lag"`
	ClusterHealth         string        `json:"cluster_health"`
	LastError             string        `json:"last_error,omitempty"`
	Uptime                time.Duration `json:"uptime"`
	StartTime             time.Time     `json:"start_time"`
}

// HealthStatus represents the overall health status
type HealthStatus string

const (
	HealthStatusHealthy   HealthStatus = "healthy"
	HealthStatusDegraded  HealthStatus = "degraded"
	HealthStatusUnhealthy HealthStatus = "unhealthy"
	HealthStatusDown      HealthStatus = "down"
)

// CircuitBreakerState represents circuit breaker states
type CircuitBreakerState string

const (
	CircuitBreakerClosed    CircuitBreakerState = "closed"
	CircuitBreakerOpen      CircuitBreakerState = "open"
	CircuitBreakerHalfOpen  CircuitBreakerState = "half_open"
)

// CircuitBreaker implements circuit breaker pattern for Redis operations
type CircuitBreaker struct {
	state                CircuitBreakerState
	failureCount         int
	successCount         int
	lastFailureTime      time.Time
	timeout              time.Duration
	maxFailures          int
	recoveryThreshold    int
	mutex                sync.RWMutex
}

// NewCircuitBreaker creates a new circuit breaker
func NewCircuitBreaker(maxFailures int, timeout time.Duration, recoveryThreshold int) *CircuitBreaker {
	return &CircuitBreaker{
		state:             CircuitBreakerClosed,
		timeout:           timeout,
		maxFailures:       maxFailures,
		recoveryThreshold: recoveryThreshold,
	}
}

// Execute wraps an operation with circuit breaker logic
func (cb *CircuitBreaker) Execute(operation func() error) error {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	// Check if circuit breaker should allow the operation
	if cb.state == CircuitBreakerOpen {
		if time.Since(cb.lastFailureTime) < cb.timeout {
			return fmt.Errorf("circuit breaker is open")
		}
		// Transition to half-open
		cb.state = CircuitBreakerHalfOpen
		cb.successCount = 0
	}

	// Execute the operation
	err := operation()

	if err != nil {
		cb.recordFailure()
		return err
	}

	cb.recordSuccess()
	return nil
}

// recordFailure records a failure and potentially opens the circuit
func (cb *CircuitBreaker) recordFailure() {
	cb.failureCount++
	cb.lastFailureTime = time.Now()
	cb.successCount = 0

	if cb.failureCount >= cb.maxFailures {
		cb.state = CircuitBreakerOpen
	}
}

// recordSuccess records a success and potentially closes the circuit
func (cb *CircuitBreaker) recordSuccess() {
	cb.failureCount = 0
	cb.successCount++

	if cb.state == CircuitBreakerHalfOpen && cb.successCount >= cb.recoveryThreshold {
		cb.state = CircuitBreakerClosed
	}
}

// GetState returns the current circuit breaker state
func (cb *CircuitBreaker) GetState() CircuitBreakerState {
	cb.mutex.RLock()
	defer cb.mutex.RUnlock()
	return cb.state
}

// NewHealthMonitor creates a new health monitor
func NewHealthMonitor(client redis.UniversalClient, config *HealthConfig, log logging.SimpleLogging) *HealthMonitor {
	if config == nil {
		config = DefaultHealthConfig()
	}

	hm := &HealthMonitor{
		client:    client,
		log:       log,
		config:    config,
		stopChan:  make(chan struct{}),
		scriptMgr: NewScriptManager(client),
		metrics: &HealthMetrics{
			Status:      string(HealthStatusHealthy),
			StartTime:   time.Now(),
			MinResponseTime: time.Hour, // Initialize to high value
		},
	}

	if config.CircuitBreakerEnabled {
		hm.circuitBreaker = NewCircuitBreaker(
			config.ErrorThreshold,
			config.TimeoutThreshold,
			config.RecoveryThreshold,
		)
	}

	return hm
}

// Start begins health monitoring
func (hm *HealthMonitor) Start(ctx context.Context) error {
	hm.mutex.Lock()
	defer hm.mutex.Unlock()

	if hm.running {
		return fmt.Errorf("health monitor is already running")
	}

	// Load health check scripts
	if err := hm.scriptMgr.LoadScripts(ctx); err != nil {
		hm.log.Warn("Failed to preload health check scripts: %v", err)
	}

	hm.running = true

	// Start background monitoring
	go hm.monitoringLoop(ctx)

	hm.log.Info("Health monitor started with %v check interval", hm.config.CheckInterval)
	return nil
}

// Stop stops health monitoring
func (hm *HealthMonitor) Stop() {
	hm.mutex.Lock()
	defer hm.mutex.Unlock()

	if !hm.running {
		return
	}

	close(hm.stopChan)
	hm.running = false
	hm.log.Info("Health monitor stopped")
}

// CheckHealth performs a comprehensive health check
func (hm *HealthMonitor) CheckHealth(ctx context.Context) error {
	if hm.circuitBreaker != nil {
		return hm.circuitBreaker.Execute(func() error {
			return hm.performHealthCheck(ctx)
		})
	}

	return hm.performHealthCheck(ctx)
}

// performHealthCheck executes the actual health check
func (hm *HealthMonitor) performHealthCheck(ctx context.Context) error {
	startTime := time.Now()

	hm.mutex.Lock()
	hm.metrics.TotalChecks++
	hm.mutex.Unlock()

	// Create timeout context
	checkCtx, cancel := context.WithTimeout(ctx, hm.config.TimeoutThreshold)
	defer cancel()

	// Basic connectivity test
	if err := hm.basicHealthCheck(checkCtx); err != nil {
		hm.recordCheckResult(false, time.Since(startTime), err)
		return fmt.Errorf("basic health check failed: %w", err)
	}

	// Extended health check if enabled
	if hm.config.PerformanceMonitoring {
		if err := hm.extendedHealthCheck(checkCtx); err != nil {
			hm.log.Warn("Extended health check failed: %v", err)
			// Don't fail the entire check for extended failures
		}
	}

	hm.recordCheckResult(true, time.Since(startTime), nil)
	return nil
}

// basicHealthCheck performs basic connectivity tests
func (hm *HealthMonitor) basicHealthCheck(ctx context.Context) error {
	// Ping test
	if err := hm.client.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("ping failed: %w", err)
	}

	// Simple set/get test
	testKey := "atlantis:health:check"
	testValue := fmt.Sprintf("health-check-%d", time.Now().Unix())

	if err := hm.client.Set(ctx, testKey, testValue, time.Minute).Err(); err != nil {
		return fmt.Errorf("set operation failed: %w", err)
	}

	retrievedValue, err := hm.client.Get(ctx, testKey).Result()
	if err != nil {
		return fmt.Errorf("get operation failed: %w", err)
	}

	if retrievedValue != testValue {
		return fmt.Errorf("value mismatch: expected %s, got %s", testValue, retrievedValue)
	}

	// Cleanup test key
	hm.client.Del(ctx, testKey)

	return nil
}

// extendedHealthCheck performs comprehensive health analysis
func (hm *HealthMonitor) extendedHealthCheck(ctx context.Context) error {
	// Use Lua script for comprehensive health check
	result, err := hm.scriptMgr.Execute(ctx, "health_check", []string{}, "performance")
	if err != nil {
		return fmt.Errorf("health check script failed: %w", err)
	}

	var healthData map[string]interface{}
	if err := json.Unmarshal([]byte(result.(string)), &healthData); err != nil {
		return fmt.Errorf("failed to parse health check results: %w", err)
	}

	hm.mutex.Lock()
	defer hm.mutex.Unlock()

	// Update metrics from health check results
	if memUsage, ok := healthData["memory_usage"].(float64); ok {
		hm.metrics.MemoryUsage = int64(memUsage)
	}

	if perfTime, ok := healthData["performance_test_microseconds"].(float64); ok {
		perfDuration := time.Duration(perfTime) * time.Microsecond
		if perfDuration > hm.config.TimeoutThreshold {
			hm.log.Warn("Performance test took %v, threshold is %v", perfDuration, hm.config.TimeoutThreshold)
		}
	}

	if slowQueries, ok := healthData["recent_slow_queries"].(float64); ok {
		hm.metrics.SlowQueryCount = int(slowQueries)
		if int(slowQueries) > hm.config.MaxSlowQueries {
			return fmt.Errorf("too many slow queries: %d > %d", int(slowQueries), hm.config.MaxSlowQueries)
		}
	}

	// Check memory threshold
	if hm.metrics.MemoryUsage > hm.config.MemoryThresholdMB*1024*1024 {
		return fmt.Errorf("memory usage too high: %d MB > %d MB",
			hm.metrics.MemoryUsage/(1024*1024), hm.config.MemoryThresholdMB)
	}

	return nil
}

// recordCheckResult updates health metrics
func (hm *HealthMonitor) recordCheckResult(success bool, duration time.Duration, err error) {
	hm.mutex.Lock()
	defer hm.mutex.Unlock()

	hm.metrics.LastCheck = time.Now()
	hm.metrics.Uptime = time.Since(hm.metrics.StartTime)

	// Update response time metrics
	if duration > hm.metrics.MaxResponseTime {
		hm.metrics.MaxResponseTime = duration
	}
	if duration < hm.metrics.MinResponseTime {
		hm.metrics.MinResponseTime = duration
	}

	// Update average response time (simple moving average)
	if hm.metrics.AverageResponseTime == 0 {
		hm.metrics.AverageResponseTime = duration
	} else {
		hm.metrics.AverageResponseTime = (hm.metrics.AverageResponseTime + duration) / 2
	}

	if success {
		hm.metrics.SuccessfulChecks++
		hm.metrics.ConsecutiveSuccesses++
		hm.metrics.ConsecutiveFailures = 0
		hm.metrics.LastError = ""

		// Update status based on consecutive successes
		if hm.metrics.ConsecutiveSuccesses >= hm.config.RecoveryThreshold {
			hm.metrics.Status = string(HealthStatusHealthy)
		}
	} else {
		hm.metrics.FailedChecks++
		hm.metrics.ConsecutiveFailures++
		hm.metrics.ConsecutiveSuccesses = 0

		if err != nil {
			hm.metrics.LastError = err.Error()
		}

		// Update status based on consecutive failures
		if hm.metrics.ConsecutiveFailures >= hm.config.ErrorThreshold {
			hm.metrics.Status = string(HealthStatusUnhealthy)
		} else if hm.metrics.ConsecutiveFailures > 1 {
			hm.metrics.Status = string(HealthStatusDegraded)
		}
	}

	// Update circuit breaker state in metrics
	if hm.circuitBreaker != nil {
		hm.metrics.CircuitBreakerState = string(hm.circuitBreaker.GetState())
	}
}

// GetMetrics returns current health metrics
func (hm *HealthMonitor) GetMetrics() *HealthMetrics {
	hm.mutex.RLock()
	defer hm.mutex.RUnlock()

	// Create a copy to avoid race conditions
	metrics := *hm.metrics
	return &metrics
}

// GetStatus returns the current health status
func (hm *HealthMonitor) GetStatus() HealthStatus {
	hm.mutex.RLock()
	defer hm.mutex.RUnlock()

	return HealthStatus(hm.metrics.Status)
}

// IsHealthy returns true if the system is healthy
func (hm *HealthMonitor) IsHealthy() bool {
	status := hm.GetStatus()
	return status == HealthStatusHealthy || status == HealthStatusDegraded
}

// monitoringLoop runs continuous health monitoring
func (hm *HealthMonitor) monitoringLoop(ctx context.Context) {
	ticker := time.NewTicker(hm.config.CheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-hm.stopChan:
			return
		case <-ticker.C:
			if err := hm.CheckHealth(ctx); err != nil {
				hm.log.Warn("Health check failed: %v", err)

				if hm.config.AlertingEnabled {
					hm.triggerAlert(err)
				}
			}
		}
	}
}

// triggerAlert sends alerts when health issues are detected
func (hm *HealthMonitor) triggerAlert(err error) {
	hm.log.Error("HEALTH ALERT: Redis backend unhealthy - %v", err)

	// TODO: Implement alerting mechanisms (email, slack, webhook, etc.)
	// This would integrate with external alerting systems
}

// GetDetailedReport returns a comprehensive health report
func (hm *HealthMonitor) GetDetailedReport(ctx context.Context) (map[string]interface{}, error) {
	metrics := hm.GetMetrics()

	report := map[string]interface{}{
		"overall_status":         metrics.Status,
		"last_check":            metrics.LastCheck,
		"uptime":                metrics.Uptime.String(),
		"total_checks":          metrics.TotalChecks,
		"success_rate":          float64(metrics.SuccessfulChecks) / float64(metrics.TotalChecks) * 100,
		"consecutive_failures":  metrics.ConsecutiveFailures,
		"consecutive_successes": metrics.ConsecutiveSuccesses,
		"response_times": map[string]interface{}{
			"average": metrics.AverageResponseTime.String(),
			"min":     metrics.MinResponseTime.String(),
			"max":     metrics.MaxResponseTime.String(),
		},
		"circuit_breaker_state": metrics.CircuitBreakerState,
		"memory_usage_mb":       metrics.MemoryUsage / (1024 * 1024),
		"slow_query_count":      metrics.SlowQueryCount,
		"last_error":           metrics.LastError,
	}

	// Add Redis-specific information
	if hm.config.PerformanceMonitoring {
		info, err := hm.client.Info(ctx).Result()
		if err == nil {
			report["redis_info"] = hm.parseRedisInfo(info)
		}

		// Get configuration
		config, err := hm.client.ConfigGet(ctx, "*").Result()
		if err == nil {
			report["redis_config"] = config
		}
	}

	return report, nil
}

// parseRedisInfo parses Redis INFO output into structured data
func (hm *HealthMonitor) parseRedisInfo(info string) map[string]interface{} {
	result := make(map[string]interface{})
	lines := strings.Split(info, "\r\n")

	var currentSection string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "#") {
			currentSection = strings.TrimPrefix(line, "# ")
			continue
		}

		if strings.Contains(line, ":") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				key := currentSection + "_" + parts[0]
				result[key] = parts[1]
			}
		}
	}

	return result
}

// SetCircuitBreakerEnabled enables or disables the circuit breaker
func (hm *HealthMonitor) SetCircuitBreakerEnabled(enabled bool) {
	hm.mutex.Lock()
	defer hm.mutex.Unlock()

	if enabled && hm.circuitBreaker == nil {
		hm.circuitBreaker = NewCircuitBreaker(
			hm.config.ErrorThreshold,
			hm.config.TimeoutThreshold,
			hm.config.RecoveryThreshold,
		)
	} else if !enabled {
		hm.circuitBreaker = nil
	}
}

// ExecuteWithHealthCheck wraps Redis operations with health monitoring
func (hm *HealthMonitor) ExecuteWithHealthCheck(ctx context.Context, operation func(context.Context) error) error {
	startTime := time.Now()

	// Check circuit breaker if enabled
	if hm.circuitBreaker != nil && hm.circuitBreaker.GetState() == CircuitBreakerOpen {
		return fmt.Errorf("operation blocked by circuit breaker")
	}

	err := operation(ctx)
	duration := time.Since(startTime)

	// Update metrics based on operation result
	hm.recordOperationResult(err == nil, duration, err)

	return err
}

// recordOperationResult updates metrics for individual operations
func (hm *HealthMonitor) recordOperationResult(success bool, duration time.Duration, err error) {
	hm.mutex.Lock()
	defer hm.mutex.Unlock()

	// Update response time tracking
	if duration > hm.metrics.MaxResponseTime {
		hm.metrics.MaxResponseTime = duration
	}

	// Record circuit breaker events
	if hm.circuitBreaker != nil {
		if success {
			hm.circuitBreaker.recordSuccess()
		} else {
			hm.circuitBreaker.recordFailure()
		}
		hm.metrics.CircuitBreakerState = string(hm.circuitBreaker.GetState())
	}
}