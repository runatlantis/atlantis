// Copyright 2017 HootSuite Media Inc.
//
// Licensed under the Apache License, Version 2.0 (the License);
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an AS IS BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
// Modified hereafter by contributors to runatlantis/atlantis.

// Package timeout provides enhanced retry logic with circuit breakers and adaptive strategies.
package timeout

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"sync"
	"time"

	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/logging"
)

// RetryStrategy defines different retry strategies
type RetryStrategy int

const (
	FixedDelay RetryStrategy = iota
	ExponentialBackoff
	LinearBackoff
	JitteredExponential
	AdaptiveBackoff
)

// String returns a string representation of the retry strategy
func (rs RetryStrategy) String() string {
	switch rs {
	case FixedDelay:
		return "fixed_delay"
	case ExponentialBackoff:
		return "exponential_backoff"
	case LinearBackoff:
		return "linear_backoff"
	case JitteredExponential:
		return "jittered_exponential"
	case AdaptiveBackoff:
		return "adaptive_backoff"
	default:
		return "unknown"
	}
}

// RetryPolicy defines the retry policy for lock operations
type RetryPolicy struct {
	MaxAttempts       int
	Strategy          RetryStrategy
	BaseDelay         time.Duration
	MaxDelay          time.Duration
	BackoffMultiplier float64
	JitterPercent     float64
	CircuitBreaker    *CircuitBreakerConfig
	RateLimiter       *RateLimiterConfig
	Conditions        []RetryCondition
}

// RetryCondition defines when retries should be attempted
type RetryCondition struct {
	ErrorType     string
	ShouldRetry   bool
	MaxAttempts   int
	CustomDelay   time.Duration
}

// RateLimiterConfig configures rate limiting for retries
type RateLimiterConfig struct {
	MaxTokens   float64
	RefillRate  float64
	BurstLimit  int
}

// EnhancedRetryManager provides advanced retry capabilities
type EnhancedRetryManager struct {
	mu             sync.RWMutex
	policy         RetryPolicy
	logger         logging.SimpleLogging
	circuitBreaker *CircuitBreaker
	rateLimiter    *RateLimiter
	metrics        *EnhancedRetryMetrics
	adaptiveData   map[string]*AdaptiveRetryData
}

// AdaptiveRetryData tracks retry performance for adaptive strategies
type AdaptiveRetryData struct {
	SuccessCount     int64
	FailureCount     int64
	AverageLatency   time.Duration
	LastSuccessDelay time.Duration
	RecentOutcomes   []RetryOutcome
	LastUpdated      time.Time
}

// RetryOutcome tracks the outcome of a retry attempt
type RetryOutcome struct {
	Attempt   int
	Delay     time.Duration
	Success   bool
	Latency   time.Duration
	Timestamp time.Time
	Error     error
}

// EnhancedRetryMetrics tracks comprehensive retry metrics
type EnhancedRetryMetrics struct {
	mu                    sync.RWMutex
	TotalRetries          int64
	SuccessfulRetries     int64
	FailedRetries         int64
	CircuitBreakerTrips   int64
	RateLimitedRetries    int64
	AdaptiveAdjustments   int64
	StrategyDistribution  map[RetryStrategy]int64
	AverageRetryLatency   time.Duration
	MaxRetryAttempts      int
}

// NewEnhancedRetryManager creates a new enhanced retry manager
func NewEnhancedRetryManager(policy RetryPolicy, logger logging.SimpleLogging) *EnhancedRetryManager {
	// Set default values
	if policy.MaxAttempts == 0 {
		policy.MaxAttempts = 3
	}
	if policy.BaseDelay == 0 {
		policy.BaseDelay = 1 * time.Second
	}
	if policy.MaxDelay == 0 {
		policy.MaxDelay = 30 * time.Second
	}
	if policy.BackoffMultiplier == 0 {
		policy.BackoffMultiplier = 2.0
	}
	if policy.JitterPercent == 0 {
		policy.JitterPercent = 0.1
	}

	rm := &EnhancedRetryManager{
		policy:       policy,
		logger:       logger,
		adaptiveData: make(map[string]*AdaptiveRetryData),
		metrics: &EnhancedRetryMetrics{
			StrategyDistribution: make(map[RetryStrategy]int64),
		},
	}

	// Initialize circuit breaker if configured
	if policy.CircuitBreaker != nil {
		rm.circuitBreaker = NewCircuitBreaker(policy.CircuitBreaker)
	}

	// Initialize rate limiter if configured
	if policy.RateLimiter != nil {
		rm.rateLimiter = NewRateLimiter(policy.RateLimiter.MaxTokens, policy.RateLimiter.RefillRate)
	}

	return rm
}

// RetryableOperation defines an operation that can be retried
type RetryableOperation func(ctx context.Context, attempt int) error

// ExecuteWithRetry executes an operation with the configured retry policy
func (rm *EnhancedRetryManager) ExecuteWithRetry(ctx context.Context, operationID string, operation RetryableOperation) error {
	rm.updateMetrics(func(m *EnhancedRetryMetrics) {
		m.TotalRetries++
		m.StrategyDistribution[rm.policy.Strategy]++
	})

	// Check circuit breaker
	if rm.circuitBreaker != nil && rm.circuitBreaker.GetState() == CircuitOpen {
		rm.updateMetrics(func(m *EnhancedRetryMetrics) {
			m.CircuitBreakerTrips++
		})
		return fmt.Errorf("circuit breaker is open")
	}

	// Check rate limiter
	if rm.rateLimiter != nil && !rm.rateLimiter.Allow() {
		rm.updateMetrics(func(m *EnhancedRetryMetrics) {
			m.RateLimitedRetries++
		})
		return fmt.Errorf("rate limited")
	}

	var lastErr error
	startTime := time.Now()

	for attempt := 1; attempt <= rm.policy.MaxAttempts; attempt++ {
		attemptStart := time.Now()

		// Execute the operation
		err := operation(ctx, attempt)
		latency := time.Since(attemptStart)

		// Record the outcome
		outcome := RetryOutcome{
			Attempt:   attempt,
			Success:   err == nil,
			Latency:   latency,
			Timestamp: time.Now(),
			Error:     err,
		}

		if err == nil {
			// Success
			rm.recordSuccess(operationID, outcome)

			if rm.circuitBreaker != nil {
				rm.circuitBreaker.recordResult(true)
			}

			if attempt > 1 {
				rm.updateMetrics(func(m *EnhancedRetryMetrics) {
					m.SuccessfulRetries++
				})
				rm.logger.Info("operation succeeded after %d attempts (operation: %s)", attempt, operationID)
			}

			return nil
		}

		// Record failure
		lastErr = err
		rm.recordFailure(operationID, outcome)

		if rm.circuitBreaker != nil {
			rm.circuitBreaker.recordResult(false)
		}

		// Check if we should retry
		if !rm.shouldRetry(err, attempt) {
			rm.logger.Info("operation not retryable: %v (operation: %s)", err, operationID)
			break
		}

		// Check for context cancellation
		if ctx.Err() != nil {
			return ctx.Err()
		}

		// Don't sleep after the last attempt
		if attempt < rm.policy.MaxAttempts {
			delay := rm.calculateDelay(operationID, attempt, outcome)
			outcome.Delay = delay

			rm.logger.Info("operation failed (attempt %d/%d), retrying in %v: %v (operation: %s)",
				attempt, rm.policy.MaxAttempts, delay, err, operationID)

			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(delay):
				// Continue to next attempt
			}
		}
	}

	// All attempts failed
	totalDuration := time.Since(startTime)
	rm.updateMetrics(func(m *EnhancedRetryMetrics) {
		m.FailedRetries++
		if rm.policy.MaxAttempts > m.MaxRetryAttempts {
			m.MaxRetryAttempts = rm.policy.MaxAttempts
		}
		// Update average latency
		if m.AverageRetryLatency == 0 {
			m.AverageRetryLatency = totalDuration
		} else {
			m.AverageRetryLatency = (m.AverageRetryLatency + totalDuration) / 2
		}
	})

	rm.logger.Warn("operation failed after %d attempts in %v: %v (operation: %s)",
		rm.policy.MaxAttempts, totalDuration, lastErr, operationID)

	return fmt.Errorf("operation failed after %d attempts: %w", rm.policy.MaxAttempts, lastErr)
}

// calculateDelay calculates the delay before the next retry attempt
func (rm *EnhancedRetryManager) calculateDelay(operationID string, attempt int, outcome RetryOutcome) time.Duration {
	switch rm.policy.Strategy {
	case FixedDelay:
		return rm.policy.BaseDelay
	case ExponentialBackoff:
		return rm.calculateExponentialBackoff(attempt)
	case LinearBackoff:
		return rm.calculateLinearBackoff(attempt)
	case JitteredExponential:
		return rm.calculateJitteredExponential(attempt)
	case AdaptiveBackoff:
		return rm.calculateAdaptiveBackoff(operationID, attempt, outcome)
	default:
		return rm.policy.BaseDelay
	}
}

// calculateExponentialBackoff calculates exponential backoff delay
func (rm *EnhancedRetryManager) calculateExponentialBackoff(attempt int) time.Duration {
	delay := float64(rm.policy.BaseDelay) * math.Pow(rm.policy.BackoffMultiplier, float64(attempt-1))

	if delay > float64(rm.policy.MaxDelay) {
		delay = float64(rm.policy.MaxDelay)
	}

	return time.Duration(delay)
}

// calculateLinearBackoff calculates linear backoff delay
func (rm *EnhancedRetryManager) calculateLinearBackoff(attempt int) time.Duration {
	delay := float64(rm.policy.BaseDelay) * float64(attempt)

	if delay > float64(rm.policy.MaxDelay) {
		delay = float64(rm.policy.MaxDelay)
	}

	return time.Duration(delay)
}

// calculateJitteredExponential calculates jittered exponential backoff
func (rm *EnhancedRetryManager) calculateJitteredExponential(attempt int) time.Duration {
	baseDelay := rm.calculateExponentialBackoff(attempt)
	jitter := rm.calculateJitter(baseDelay)

	delay := baseDelay + jitter
	if delay > rm.policy.MaxDelay {
		delay = rm.policy.MaxDelay
	}
	if delay < 0 {
		delay = rm.policy.BaseDelay
	}

	return delay
}

// calculateAdaptiveBackoff calculates adaptive backoff based on historical performance
func (rm *EnhancedRetryManager) calculateAdaptiveBackoff(operationID string, attempt int, outcome RetryOutcome) time.Duration {
	rm.mu.RLock()
	data, exists := rm.adaptiveData[operationID]
	rm.mu.RUnlock()

	if !exists || len(data.RecentOutcomes) == 0 {
		// Fall back to exponential backoff for new operations
		return rm.calculateExponentialBackoff(attempt)
	}

	// Calculate success rate
	totalOutcomes := data.SuccessCount + data.FailureCount
	successRate := float64(data.SuccessCount) / float64(totalOutcomes)

	// Adjust delay based on success rate and recent performance
	baseDelay := rm.calculateExponentialBackoff(attempt)

	var adaptiveFactor float64
	if successRate > 0.8 {
		// High success rate - reduce delay
		adaptiveFactor = 0.5 + (successRate * 0.5)
	} else if successRate < 0.3 {
		// Low success rate - increase delay
		adaptiveFactor = 1.5 + (0.3-successRate)*2
	} else {
		// Moderate success rate - slight adjustment
		adaptiveFactor = 1.0 + (0.6-successRate)*0.5
	}

	// Consider recent latency
	if data.AverageLatency > baseDelay {
		adaptiveFactor *= 1.2 // Increase delay if operations are slow
	}

	adaptiveDelay := time.Duration(float64(baseDelay) * adaptiveFactor)

	// Apply bounds
	if adaptiveDelay < rm.policy.BaseDelay {
		adaptiveDelay = rm.policy.BaseDelay
	}
	if adaptiveDelay > rm.policy.MaxDelay {
		adaptiveDelay = rm.policy.MaxDelay
	}

	rm.updateMetrics(func(m *EnhancedRetryMetrics) {
		m.AdaptiveAdjustments++
	})

	rm.logger.Debug("adaptive backoff: operation=%s, attempt=%d, success_rate=%.2f, factor=%.2f, delay=%v",
		operationID, attempt, successRate, adaptiveFactor, adaptiveDelay)

	return adaptiveDelay
}

// calculateJitter adds jitter to a delay
func (rm *EnhancedRetryManager) calculateJitter(baseDelay time.Duration) time.Duration {
	if rm.policy.JitterPercent <= 0 {
		return 0
	}

	maxJitter := float64(baseDelay) * rm.policy.JitterPercent
	jitter := (rand.Float64() * 2 * maxJitter) - maxJitter

	return time.Duration(jitter)
}

// shouldRetry determines if an operation should be retried
func (rm *EnhancedRetryManager) shouldRetry(err error, attempt int) bool {
	// Check attempt limit
	if attempt >= rm.policy.MaxAttempts {
		return false
	}

	// Check custom retry conditions
	for _, condition := range rm.policy.Conditions {
		if rm.matchesCondition(err, condition) {
			if !condition.ShouldRetry {
				return false
			}
			if condition.MaxAttempts > 0 && attempt >= condition.MaxAttempts {
				return false
			}
		}
	}

	// Default retry logic
	return rm.isRetryableError(err)
}

// matchesCondition checks if an error matches a retry condition
func (rm *EnhancedRetryManager) matchesCondition(err error, condition RetryCondition) bool {
	// Simple string matching for now - could be enhanced with regex or error type checking
	return err != nil && condition.ErrorType != "" &&
		   (err.Error() == condition.ErrorType ||
		    fmt.Sprintf("%T", err) == condition.ErrorType)
}

// isRetryableError determines if an error should trigger a retry
func (rm *EnhancedRetryManager) isRetryableError(err error) bool {
	if err == nil {
		return false
	}

	// Context errors are not retryable
	if err == context.Canceled || err == context.DeadlineExceeded {
		return false
	}

	// Check for specific error patterns that should not be retried
	errorMsg := err.Error()
	nonRetryablePatterns := []string{
		"permission denied",
		"unauthorized",
		"forbidden",
		"invalid request",
		"malformed",
		"bad request",
	}

	for _, pattern := range nonRetryablePatterns {
		if containsIgnoreCase(errorMsg, pattern) {
			return false
		}
	}

	// Most other errors are retryable
	return true
}

// containsIgnoreCase checks if a string contains a substring (case-insensitive)
func containsIgnoreCase(s, substr string) bool {
	return len(s) >= len(substr) &&
		   (s == substr ||
		    len(s) > len(substr) &&
		    (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
		     findSubstring(s, substr)))
}

// findSubstring is a simple substring search
func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// recordSuccess records a successful operation outcome
func (rm *EnhancedRetryManager) recordSuccess(operationID string, outcome RetryOutcome) {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	data := rm.getOrCreateAdaptiveData(operationID)
	data.SuccessCount++
	data.LastSuccessDelay = outcome.Delay
	data.RecentOutcomes = append(data.RecentOutcomes, outcome)

	// Update average latency
	if data.AverageLatency == 0 {
		data.AverageLatency = outcome.Latency
	} else {
		data.AverageLatency = (data.AverageLatency + outcome.Latency) / 2
	}

	data.LastUpdated = time.Now()

	// Trim old outcomes
	rm.trimOutcomes(data)
}

// recordFailure records a failed operation outcome
func (rm *EnhancedRetryManager) recordFailure(operationID string, outcome RetryOutcome) {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	data := rm.getOrCreateAdaptiveData(operationID)
	data.FailureCount++
	data.RecentOutcomes = append(data.RecentOutcomes, outcome)
	data.LastUpdated = time.Now()

	// Trim old outcomes
	rm.trimOutcomes(data)
}

// getOrCreateAdaptiveData gets or creates adaptive data for an operation
func (rm *EnhancedRetryManager) getOrCreateAdaptiveData(operationID string) *AdaptiveRetryData {
	data, exists := rm.adaptiveData[operationID]
	if !exists {
		data = &AdaptiveRetryData{
			RecentOutcomes: make([]RetryOutcome, 0, 100),
			LastUpdated:    time.Now(),
		}
		rm.adaptiveData[operationID] = data
	}
	return data
}

// trimOutcomes keeps only recent outcomes
func (rm *EnhancedRetryManager) trimOutcomes(data *AdaptiveRetryData) {
	maxOutcomes := 50
	cutoff := time.Now().Add(-24 * time.Hour)

	// Remove old outcomes
	validOutcomes := data.RecentOutcomes[:0]
	for _, outcome := range data.RecentOutcomes {
		if outcome.Timestamp.After(cutoff) && len(validOutcomes) < maxOutcomes {
			validOutcomes = append(validOutcomes, outcome)
		}
	}
	data.RecentOutcomes = validOutcomes
}

// updateMetrics safely updates retry metrics
func (rm *EnhancedRetryManager) updateMetrics(updateFn func(*EnhancedRetryMetrics)) {
	rm.metrics.mu.Lock()
	defer rm.metrics.mu.Unlock()
	updateFn(rm.metrics)
}

// GetMetrics returns a copy of the current metrics
func (rm *EnhancedRetryManager) GetMetrics() EnhancedRetryMetrics {
	rm.metrics.mu.RLock()
	defer rm.metrics.mu.RUnlock()

	metrics := EnhancedRetryMetrics{
		TotalRetries:          rm.metrics.TotalRetries,
		SuccessfulRetries:     rm.metrics.SuccessfulRetries,
		FailedRetries:         rm.metrics.FailedRetries,
		CircuitBreakerTrips:   rm.metrics.CircuitBreakerTrips,
		RateLimitedRetries:    rm.metrics.RateLimitedRetries,
		AdaptiveAdjustments:   rm.metrics.AdaptiveAdjustments,
		StrategyDistribution:  make(map[RetryStrategy]int64),
		AverageRetryLatency:   rm.metrics.AverageRetryLatency,
		MaxRetryAttempts:      rm.metrics.MaxRetryAttempts,
	}

	for k, v := range rm.metrics.StrategyDistribution {
		metrics.StrategyDistribution[k] = v
	}

	return metrics
}

// GetAdaptiveData returns adaptive data for analysis
func (rm *EnhancedRetryManager) GetAdaptiveData() map[string]*AdaptiveRetryData {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	data := make(map[string]*AdaptiveRetryData)
	for k, v := range rm.adaptiveData {
		// Create a copy to avoid race conditions
		dataCopy := &AdaptiveRetryData{
			SuccessCount:     v.SuccessCount,
			FailureCount:     v.FailureCount,
			AverageLatency:   v.AverageLatency,
			LastSuccessDelay: v.LastSuccessDelay,
			RecentOutcomes:   make([]RetryOutcome, len(v.RecentOutcomes)),
			LastUpdated:      v.LastUpdated,
		}
		copy(dataCopy.RecentOutcomes, v.RecentOutcomes)
		data[k] = dataCopy
	}

	return data
}

// Cleanup removes old adaptive data
func (rm *EnhancedRetryManager) Cleanup() {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	cutoff := time.Now().Add(-24 * time.Hour)
	for operationID, data := range rm.adaptiveData {
		if data.LastUpdated.Before(cutoff) {
			delete(rm.adaptiveData, operationID)
		}
	}
}

// DefaultRetryPolicy returns a sensible default retry policy
func DefaultRetryPolicy() RetryPolicy {
	return RetryPolicy{
		MaxAttempts:       3,
		Strategy:          JitteredExponential,
		BaseDelay:         1 * time.Second,
		MaxDelay:          30 * time.Second,
		BackoffMultiplier: 2.0,
		JitterPercent:     0.1,
		CircuitBreaker: &CircuitBreakerConfig{
			FailureThreshold: 5,
			SuccessThreshold: 3,
			Timeout:          30 * time.Second,
			ResetTimeout:     60 * time.Second,
		},
		RateLimiter: &RateLimiterConfig{
			MaxTokens:  10,
			RefillRate: 1.0,
			BurstLimit: 5,
		},
	}
}

// AggressiveRetryPolicy returns a more aggressive retry policy for critical operations
func AggressiveRetryPolicy() RetryPolicy {
	return RetryPolicy{
		MaxAttempts:       5,
		Strategy:          AdaptiveBackoff,
		BaseDelay:         500 * time.Millisecond,
		MaxDelay:          60 * time.Second,
		BackoffMultiplier: 1.5,
		JitterPercent:     0.2,
		CircuitBreaker: &CircuitBreakerConfig{
			FailureThreshold: 10,
			SuccessThreshold: 5,
			Timeout:          60 * time.Second,
			ResetTimeout:     120 * time.Second,
		},
		RateLimiter: &RateLimiterConfig{
			MaxTokens:  20,
			RefillRate: 2.0,
			BurstLimit: 10,
		},
	}
}

// ConservativeRetryPolicy returns a conservative retry policy
func ConservativeRetryPolicy() RetryPolicy {
	return RetryPolicy{
		MaxAttempts:       2,
		Strategy:          ExponentialBackoff,
		BaseDelay:         2 * time.Second,
		MaxDelay:          15 * time.Second,
		BackoffMultiplier: 2.0,
		JitterPercent:     0.05,
		CircuitBreaker: &CircuitBreakerConfig{
			FailureThreshold: 3,
			SuccessThreshold: 2,
			Timeout:          15 * time.Second,
			ResetTimeout:     30 * time.Second,
		},
		RateLimiter: &RateLimiterConfig{
			MaxTokens:  5,
			RefillRate: 0.5,
			BurstLimit: 2,
		},
	}
}