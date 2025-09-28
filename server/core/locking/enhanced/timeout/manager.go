package timeout

import (
	"context"
	"math"
	"sync"
	"sync/atomic"
	"time"

	"github.com/runatlantis/atlantis/server/core/locking/types"
	"github.com/runatlantis/atlantis/server/logging"
)

// TimeoutManager manages timeouts for lock operations
type TimeoutManager struct {
	timers  sync.Map // map[string]*time.Timer
	log     logging.SimpleLogging
	metrics *TimeoutMetrics
}

// TimeoutCallback is called when a timeout occurs
type TimeoutCallback func(ctx context.Context, lockID string, resource types.ResourceIdentifier)

// NewTimeoutManager creates a new timeout manager
func NewTimeoutManager(log logging.SimpleLogging) *TimeoutManager {
	return &TimeoutManager{
		log:     log,
		metrics: NewTimeoutMetrics(),
	}
}

// SetTimeout sets a timeout for a lock operation
func (tm *TimeoutManager) SetTimeout(ctx context.Context, lockID string, resource types.ResourceIdentifier, timeout time.Duration, callback TimeoutCallback) {
	if timeout <= 0 {
		return
	}

	timer := time.AfterFunc(timeout, func() {
		tm.log.Info("Lock timeout triggered for lock: %s", lockID)
		tm.metrics.IncrementTimeouts()

		// Remove the timer
		tm.timers.Delete(lockID)

		// Execute callback
		if callback != nil {
			callback(ctx, lockID, resource)
		}
	})

	// Store the timer
	if existing, loaded := tm.timers.LoadOrStore(lockID, timer); loaded {
		// If a timer already exists, stop the old one
		if existingTimer, ok := existing.(*time.Timer); ok {
			existingTimer.Stop()
		}
		tm.timers.Store(lockID, timer)
	}

	tm.metrics.IncrementCreated()
}

// ClearTimeout removes a timeout for a lock
func (tm *TimeoutManager) ClearTimeout(lockID string) bool {
	if timer, ok := tm.timers.LoadAndDelete(lockID); ok {
		if t, ok := timer.(*time.Timer); ok {
			stopped := t.Stop()
			if stopped {
				tm.metrics.IncrementCleared()
			}
			return stopped
		}
	}
	return false
}

// ExtendTimeout extends an existing timeout
func (tm *TimeoutManager) ExtendTimeout(ctx context.Context, lockID string, resource types.ResourceIdentifier, extension time.Duration, callback TimeoutCallback) bool {
	if val, ok := tm.timers.Load(lockID); ok {
		if timer, ok := val.(*time.Timer); ok {
			timer.Stop()
			newTimer := time.AfterFunc(extension, func() {
				tm.log.Info("Extended lock timeout triggered for lock: %s", lockID)
				tm.metrics.IncrementTimeouts()
				tm.timers.Delete(lockID)

				if callback != nil {
					callback(ctx, lockID, resource)
				}
			})

			tm.timers.Store(lockID, newTimer)
			tm.metrics.IncrementExtended()
			return true
		}
	}
	return false
}

// GetActiveTimeouts returns the number of active timeouts
func (tm *TimeoutManager) GetActiveTimeouts() int {
	count := 0
	tm.timers.Range(func(_, _ interface{}) bool {
		count++
		return true
	})
	return count
}

// GetMetrics returns timeout metrics
func (tm *TimeoutManager) GetMetrics() *TimeoutMetrics {
	return tm.metrics
}

// Cleanup removes all active timeouts
func (tm *TimeoutManager) Cleanup() {
	tm.timers.Range(func(key, value interface{}) bool {
		if timer, ok := value.(*time.Timer); ok {
			timer.Stop()
		}
		tm.timers.Delete(key)
		return true
	})
}

// TimeoutMetrics tracks timeout-related metrics
type TimeoutMetrics struct {
	Created  int64 `json:"created"`
	Cleared  int64 `json:"cleared"`
	Extended int64 `json:"extended"`
	Timeouts int64 `json:"timeouts"`
}

// NewTimeoutMetrics creates new timeout metrics
func NewTimeoutMetrics() *TimeoutMetrics {
	return &TimeoutMetrics{}
}

func (tm *TimeoutMetrics) IncrementCreated() {
	atomic.AddInt64(&tm.Created, 1)
}

func (tm *TimeoutMetrics) IncrementCleared() {
	atomic.AddInt64(&tm.Cleared, 1)
}

func (tm *TimeoutMetrics) IncrementExtended() {
	atomic.AddInt64(&tm.Extended, 1)
}

func (tm *TimeoutMetrics) IncrementTimeouts() {
	atomic.AddInt64(&tm.Timeouts, 1)
}

// RetryManager handles retry logic with exponential backoff
type RetryManager struct {
	config  *RetryConfig
	metrics *RetryMetrics
	log     logging.SimpleLogging
}

// RetryConfig configures retry behavior
type RetryConfig struct {
	MaxAttempts   int           `json:"max_attempts"`
	BaseDelay     time.Duration `json:"base_delay"`
	MaxDelay      time.Duration `json:"max_delay"`
	Multiplier    float64       `json:"multiplier"`
	Jitter        bool          `json:"jitter"`
	JitterPercent float64       `json:"jitter_percent"`
}

// DefaultRetryConfig returns default retry configuration
func DefaultRetryConfig() *RetryConfig {
	return &RetryConfig{
		MaxAttempts:   3,
		BaseDelay:     time.Second,
		MaxDelay:      30 * time.Second,
		Multiplier:    2.0,
		Jitter:        true,
		JitterPercent: 0.1, // 10% jitter
	}
}

// NewRetryManager creates a new retry manager
func NewRetryManager(config *RetryConfig, log logging.SimpleLogging) *RetryManager {
	if config == nil {
		config = DefaultRetryConfig()
	}

	return &RetryManager{
		config:  config,
		metrics: NewRetryMetrics(),
		log:     log,
	}
}

// RetryOperation represents an operation that can be retried
type RetryOperation func(ctx context.Context, attempt int) error

// Execute executes an operation with retry logic
func (rm *RetryManager) Execute(ctx context.Context, operation RetryOperation) error {
	var lastErr error

	for attempt := 1; attempt <= rm.config.MaxAttempts; attempt++ {
		rm.metrics.IncrementAttempts()

		err := operation(ctx, attempt)
		if err == nil {
			if attempt > 1 {
				rm.metrics.IncrementRetrySuccess()
				rm.log.Info("Operation succeeded after %d attempts", attempt)
			}
			return nil
		}

		lastErr = err
		rm.metrics.IncrementFailures()

		// Don't retry on context cancellation
		if ctx.Err() != nil {
			return ctx.Err()
		}

		// Check if error is retryable
		if !rm.isRetryableError(err) {
			rm.log.Info("Non-retryable error, stopping retry: %v", err)
			return err
		}

		// Don't sleep after the last attempt
		if attempt < rm.config.MaxAttempts {
			delay := rm.calculateDelay(attempt)
			rm.log.Info("Operation failed (attempt %d/%d), retrying in %v: %v",
				attempt, rm.config.MaxAttempts, delay, err)

			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(delay):
				// Continue to next attempt
			}
		}
	}

	rm.metrics.IncrementRetryFailures()
	rm.log.Warn("Operation failed after %d attempts: %v", rm.config.MaxAttempts, lastErr)
	return lastErr
}

// calculateDelay calculates the delay before the next retry attempt
func (rm *RetryManager) calculateDelay(attempt int) time.Duration {
	delay := float64(rm.config.BaseDelay) * math.Pow(rm.config.Multiplier, float64(attempt-1))

	// Apply maximum delay
	if delay > float64(rm.config.MaxDelay) {
		delay = float64(rm.config.MaxDelay)
	}

	// Apply jitter if enabled
	if rm.config.Jitter {
		jitterAmount := delay * rm.config.JitterPercent
		jitter := (2 * time.Now().UnixNano() % int64(jitterAmount)) - int64(jitterAmount)
		delay += float64(jitter)
	}

	// Ensure delay is not negative
	if delay < 0 {
		delay = float64(rm.config.BaseDelay)
	}

	return time.Duration(delay)
}

// isRetryableError determines if an error should trigger a retry
func (rm *RetryManager) isRetryableError(err error) bool {
	if lockErr, ok := err.(*types.LockError); ok {
		switch lockErr.Code {
		case types.ErrCodeTimeout, types.ErrCodeBackendError:
			return true
		case types.ErrCodeLockExists, types.ErrCodeLockNotFound,
			types.ErrCodeInvalidRequest, types.ErrCodePermissionDenied:
			return false
		default:
			return true
		}
	}

	// For non-lock errors, be conservative and retry
	return true
}

// GetMetrics returns retry metrics
func (rm *RetryManager) GetMetrics() *RetryMetrics {
	return rm.metrics
}

// RetryMetrics tracks retry-related metrics
type RetryMetrics struct {
	Attempts      int64 `json:"attempts"`
	Failures      int64 `json:"failures"`
	RetrySuccess  int64 `json:"retry_success"`
	RetryFailures int64 `json:"retry_failures"`
}

// NewRetryMetrics creates new retry metrics
func NewRetryMetrics() *RetryMetrics {
	return &RetryMetrics{}
}

func (rm *RetryMetrics) IncrementAttempts() {
	atomic.AddInt64(&rm.Attempts, 1)
}

func (rm *RetryMetrics) IncrementFailures() {
	atomic.AddInt64(&rm.Failures, 1)
}

func (rm *RetryMetrics) IncrementRetrySuccess() {
	atomic.AddInt64(&rm.RetrySuccess, 1)
}

func (rm *RetryMetrics) IncrementRetryFailures() {
	atomic.AddInt64(&rm.RetryFailures, 1)
}

// CircuitBreaker implements the circuit breaker pattern for fault tolerance
type CircuitBreaker struct {
	config    *CircuitBreakerConfig
	state     CircuitState
	mutex     sync.RWMutex
	failures  int64
	successes int64
	lastState time.Time
	nextCheck time.Time
}

type CircuitState int

const (
	CircuitClosed CircuitState = iota
	CircuitOpen
	CircuitHalfOpen
)

// CircuitBreakerConfig configures circuit breaker behavior
type CircuitBreakerConfig struct {
	FailureThreshold int           `json:"failure_threshold"`
	SuccessThreshold int           `json:"success_threshold"`
	Timeout          time.Duration `json:"timeout"`
	ResetTimeout     time.Duration `json:"reset_timeout"`
}

// DefaultCircuitBreakerConfig returns default circuit breaker configuration
func DefaultCircuitBreakerConfig() *CircuitBreakerConfig {
	return &CircuitBreakerConfig{
		FailureThreshold: 5,
		SuccessThreshold: 3,
		Timeout:          30 * time.Second,
		ResetTimeout:     60 * time.Second,
	}
}

// NewCircuitBreaker creates a new circuit breaker
func NewCircuitBreaker(config *CircuitBreakerConfig) *CircuitBreaker {
	if config == nil {
		config = DefaultCircuitBreakerConfig()
	}

	return &CircuitBreaker{
		config:    config,
		state:     CircuitClosed,
		lastState: time.Now(),
	}
}

// Execute executes an operation through the circuit breaker
func (cb *CircuitBreaker) Execute(ctx context.Context, operation func() error) error {
	if !cb.allowRequest() {
		return &types.LockError{
			Type:    "CircuitOpen",
			Message: "circuit breaker is open",
			Code:    "CIRCUIT_OPEN",
		}
	}

	err := operation()
	cb.recordResult(err == nil)
	return err
}

// allowRequest determines if a request should be allowed
func (cb *CircuitBreaker) allowRequest() bool {
	cb.mutex.RLock()
	defer cb.mutex.RUnlock()

	switch cb.state {
	case CircuitClosed:
		return true
	case CircuitOpen:
		return time.Now().After(cb.nextCheck)
	case CircuitHalfOpen:
		return true
	default:
		return false
	}
}

// recordResult records the result of an operation
func (cb *CircuitBreaker) recordResult(success bool) {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	now := time.Now()

	if success {
		cb.successes++
		cb.failures = 0 // Reset consecutive failures

		if cb.state == CircuitHalfOpen && cb.successes >= int64(cb.config.SuccessThreshold) {
			cb.state = CircuitClosed
			cb.lastState = now
			cb.successes = 0
		}
	} else {
		cb.failures++
		cb.successes = 0 // Reset consecutive successes

		if cb.state == CircuitClosed && cb.failures >= int64(cb.config.FailureThreshold) {
			cb.state = CircuitOpen
			cb.lastState = now
			cb.nextCheck = now.Add(cb.config.ResetTimeout)
		} else if cb.state == CircuitHalfOpen {
			cb.state = CircuitOpen
			cb.lastState = now
			cb.nextCheck = now.Add(cb.config.ResetTimeout)
		}
	}

	// Transition from Open to HalfOpen
	if cb.state == CircuitOpen && now.After(cb.nextCheck) {
		cb.state = CircuitHalfOpen
		cb.lastState = now
		cb.failures = 0
		cb.successes = 0
	}
}

// GetState returns the current circuit breaker state
func (cb *CircuitBreaker) GetState() CircuitState {
	cb.mutex.RLock()
	defer cb.mutex.RUnlock()
	return cb.state
}

// GetMetrics returns circuit breaker metrics
func (cb *CircuitBreaker) GetMetrics() map[string]interface{} {
	cb.mutex.RLock()
	defer cb.mutex.RUnlock()

	return map[string]interface{}{
		"state":             cb.state,
		"failures":          cb.failures,
		"successes":         cb.successes,
		"last_state_change": cb.lastState,
		"next_check":        cb.nextCheck,
	}
}

// RateLimiter implements token bucket rate limiting
type RateLimiter struct {
	tokens     float64
	maxTokens  float64
	refillRate float64
	lastRefill time.Time
	mutex      sync.Mutex
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(maxTokens, refillRate float64) *RateLimiter {
	return &RateLimiter{
		tokens:     maxTokens,
		maxTokens:  maxTokens,
		refillRate: refillRate,
		lastRefill: time.Now(),
	}
}

// Allow checks if a request should be allowed
func (rl *RateLimiter) Allow() bool {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	now := time.Now()
	elapsed := now.Sub(rl.lastRefill).Seconds()
	rl.lastRefill = now

	// Refill tokens
	rl.tokens += elapsed * rl.refillRate
	if rl.tokens > rl.maxTokens {
		rl.tokens = rl.maxTokens
	}

	// Check if we have tokens available
	if rl.tokens >= 1.0 {
		rl.tokens -= 1.0
		return true
	}

	return false
}

// GetTokens returns the current number of tokens
func (rl *RateLimiter) GetTokens() float64 {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()
	return rl.tokens
}

// AdaptiveTimeoutManager adjusts timeouts based on system load and performance
type AdaptiveTimeoutManager struct {
	baseTimeout     time.Duration
	minTimeout      time.Duration
	maxTimeout      time.Duration
	loadFactor      float64
	successRate     float64
	averageLatency  time.Duration
	mutex           sync.RWMutex
	measurements    []time.Duration
	maxMeasurements int
}

// NewAdaptiveTimeoutManager creates a new adaptive timeout manager
func NewAdaptiveTimeoutManager(baseTimeout, minTimeout, maxTimeout time.Duration) *AdaptiveTimeoutManager {
	return &AdaptiveTimeoutManager{
		baseTimeout:     baseTimeout,
		minTimeout:      minTimeout,
		maxTimeout:      maxTimeout,
		loadFactor:      1.0,
		successRate:     1.0,
		measurements:    make([]time.Duration, 0, 100),
		maxMeasurements: 100,
	}
}

// GetTimeout returns the current adaptive timeout
func (atm *AdaptiveTimeoutManager) GetTimeout() time.Duration {
	atm.mutex.RLock()
	defer atm.mutex.RUnlock()

	// Calculate adaptive timeout based on load and success rate
	adaptiveFactor := atm.loadFactor * (2.0 - atm.successRate)
	timeout := time.Duration(float64(atm.baseTimeout) * adaptiveFactor)

	// Apply bounds
	if timeout < atm.minTimeout {
		timeout = atm.minTimeout
	}
	if timeout > atm.maxTimeout {
		timeout = atm.maxTimeout
	}

	return timeout
}

// RecordLatency records a latency measurement
func (atm *AdaptiveTimeoutManager) RecordLatency(latency time.Duration) {
	atm.mutex.Lock()
	defer atm.mutex.Unlock()

	atm.measurements = append(atm.measurements, latency)

	// Keep only recent measurements
	if len(atm.measurements) > atm.maxMeasurements {
		atm.measurements = atm.measurements[1:]
	}

	// Update average latency
	if len(atm.measurements) > 0 {
		var total time.Duration
		for _, m := range atm.measurements {
			total += m
		}
		atm.averageLatency = total / time.Duration(len(atm.measurements))
	}
}

// UpdateLoadFactor updates the system load factor
func (atm *AdaptiveTimeoutManager) UpdateLoadFactor(factor float64) {
	atm.mutex.Lock()
	defer atm.mutex.Unlock()
	atm.loadFactor = factor
}

// UpdateSuccessRate updates the success rate
func (atm *AdaptiveTimeoutManager) UpdateSuccessRate(rate float64) {
	atm.mutex.Lock()
	defer atm.mutex.Unlock()
	atm.successRate = rate
}

// GetMetrics returns adaptive timeout metrics
func (atm *AdaptiveTimeoutManager) GetMetrics() map[string]interface{} {
	atm.mutex.RLock()
	defer atm.mutex.RUnlock()

	return map[string]interface{}{
		"base_timeout":    atm.baseTimeout,
		"current_timeout": atm.GetTimeout(),
		"load_factor":     atm.loadFactor,
		"success_rate":    atm.successRate,
		"average_latency": atm.averageLatency,
		"measurements":    len(atm.measurements),
	}
}
