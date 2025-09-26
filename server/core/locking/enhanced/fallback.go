package enhanced

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/runatlantis/atlantis/server/core/locking"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/logging"
)

// FallbackSystem manages automatic fallback to legacy locking when enhanced system fails
type FallbackSystem struct {
	enhanced    Backend
	legacy      locking.Backend
	config      *EnhancedConfig
	log         logging.SimpleLogging

	// Circuit breaker state
	circuitBreaker *CircuitBreaker
	mu             sync.RWMutex

	// Fallback statistics
	stats *FallbackStats
}

// CircuitBreaker implements circuit breaker pattern for enhanced backend
type CircuitBreaker struct {
	failures     int64
	lastFailure  time.Time
	state        CircuitState
	threshold    int64
	timeout      time.Duration
	mu           sync.RWMutex
}

// CircuitState represents the state of the circuit breaker
type CircuitState string

const (
	CircuitStateClosed   CircuitState = "closed"   // Normal operation
	CircuitStateOpen     CircuitState = "open"     // Failing, reject requests
	CircuitStateHalfOpen CircuitState = "half_open" // Testing if service recovered
)

// FallbackStats tracks fallback system performance
type FallbackStats struct {
	EnhancedAttempts  int64     `json:"enhanced_attempts"`
	EnhancedFailures  int64     `json:"enhanced_failures"`
	FallbackAttempts  int64     `json:"fallback_attempts"`
	FallbackFailures  int64     `json:"fallback_failures"`
	LastFallback      time.Time `json:"last_fallback"`
	CircuitState      CircuitState `json:"circuit_state"`
	mu                sync.RWMutex
}

// NewFallbackSystem creates a new fallback system
func NewFallbackSystem(enhanced Backend, legacy locking.Backend, config *EnhancedConfig, log logging.SimpleLogging) *FallbackSystem {
	return &FallbackSystem{
		enhanced: enhanced,
		legacy:   legacy,
		config:   config,
		log:      log,
		circuitBreaker: &CircuitBreaker{
			threshold: 5,
			timeout:   60 * time.Second,
			state:     CircuitStateClosed,
		},
		stats: &FallbackStats{},
	}
}

// TryLock attempts lock with automatic fallback
func (fs *FallbackSystem) TryLock(lock models.ProjectLock) (bool, models.ProjectLock, error) {
	// Check circuit breaker state
	if fs.circuitBreaker.ShouldReject() {
		fs.log.Debug("Circuit breaker open, using fallback for TryLock")
		return fs.tryLockFallback(lock)
	}

	// Try enhanced system first
	acquired, currLock, err := fs.tryLockEnhanced(lock)
	if err != nil {
		fs.circuitBreaker.RecordFailure()
		fs.recordEnhancedFailure()

		fs.log.Warn("Enhanced TryLock failed, falling back: %v", err)
		return fs.tryLockFallback(lock)
	}

	fs.circuitBreaker.RecordSuccess()
	fs.recordEnhancedSuccess()
	return acquired, currLock, nil
}

// tryLockEnhanced attempts to acquire lock using enhanced backend
func (fs *FallbackSystem) tryLockEnhanced(lock models.ProjectLock) (bool, models.ProjectLock, error) {
	fs.recordEnhancedAttempt()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Convert to enhanced request
	request := &EnhancedLockRequest{
		ID: generateRequestID(),
		Resource: ResourceIdentifier{
			Type:      ResourceTypeProject,
			Namespace: lock.Project.RepoFullName,
			Name:      lock.Project.Path,
			Workspace: lock.Workspace,
			Path:      lock.Project.Path,
		},
		Priority:    PriorityNormal,
		Timeout:     fs.config.DefaultTimeout,
		Metadata:    make(map[string]string),
		Context:     ctx,
		RequestedAt: lock.Time,
		Project:     lock.Project,
		Workspace:   lock.Workspace,
		User:        lock.User,
	}

	enhancedLock, acquired, err := fs.enhanced.TryAcquireLock(ctx, request)
	if err != nil {
		return false, models.ProjectLock{}, err
	}

	if acquired {
		return true, lock, nil
	}

	// Lock not acquired - find existing lock
	if enhancedLock != nil {
		legacyLock := fs.enhanced.ConvertToLegacy(enhancedLock)
		if legacyLock != nil {
			return false, *legacyLock, nil
		}
	}

	// Check for other locks on the same resource
	locks, err := fs.enhanced.ListLocks(ctx)
	if err != nil {
		return false, models.ProjectLock{}, err
	}

	for _, existingLock := range locks {
		if existingLock.Resource.Namespace == request.Resource.Namespace &&
		   existingLock.Resource.Name == request.Resource.Name &&
		   existingLock.Resource.Workspace == request.Resource.Workspace &&
		   existingLock.State == LockStateAcquired {
			legacyLock := fs.enhanced.ConvertToLegacy(existingLock)
			if legacyLock != nil {
				return false, *legacyLock, nil
			}
		}
	}

	return false, models.ProjectLock{}, nil
}

// tryLockFallback attempts to acquire lock using legacy backend
func (fs *FallbackSystem) tryLockFallback(lock models.ProjectLock) (bool, models.ProjectLock, error) {
	fs.recordFallbackAttempt()

	acquired, currLock, err := fs.legacy.TryLock(lock)
	if err != nil {
		fs.recordFallbackFailure()
		return false, models.ProjectLock{}, err
	}

	return acquired, currLock, nil
}

// Unlock releases a lock with automatic fallback
func (fs *FallbackSystem) Unlock(project models.Project, workspace string) (*models.ProjectLock, error) {
	// Try enhanced system first (unless circuit breaker is open)
	if !fs.circuitBreaker.ShouldReject() {
		lock, err := fs.unlockEnhanced(project, workspace)
		if err == nil {
			fs.circuitBreaker.RecordSuccess()
			fs.recordEnhancedSuccess()
			return lock, nil
		}

		fs.circuitBreaker.RecordFailure()
		fs.recordEnhancedFailure()
		fs.log.Warn("Enhanced Unlock failed, falling back: %v", err)
	}

	// Fall back to legacy
	return fs.unlockFallback(project, workspace)
}

// unlockEnhanced releases lock using enhanced backend
func (fs *FallbackSystem) unlockEnhanced(project models.Project, workspace string) (*models.ProjectLock, error) {
	fs.recordEnhancedAttempt()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Find the lock to unlock
	locks, err := fs.enhanced.ListLocks(ctx)
	if err != nil {
		return nil, err
	}

	var targetLock *EnhancedLock
	for _, lock := range locks {
		if lock.Resource.Namespace == project.RepoFullName &&
		   lock.Resource.Path == project.Path &&
		   lock.Resource.Workspace == workspace {
			targetLock = lock
			break
		}
	}

	if targetLock == nil {
		return nil, nil // No lock found
	}

	err = fs.enhanced.ReleaseLock(ctx, targetLock.ID)
	if err != nil {
		return nil, err
	}

	return fs.enhanced.ConvertToLegacy(targetLock), nil
}

// unlockFallback releases lock using legacy backend
func (fs *FallbackSystem) unlockFallback(project models.Project, workspace string) (*models.ProjectLock, error) {
	fs.recordFallbackAttempt()

	lock, err := fs.legacy.Unlock(project, workspace)
	if err != nil {
		fs.recordFallbackFailure()
		return nil, err
	}

	return lock, nil
}

// List returns all locks with automatic fallback
func (fs *FallbackSystem) List() ([]models.ProjectLock, error) {
	// Try enhanced system first
	if !fs.circuitBreaker.ShouldReject() {
		locks, err := fs.listEnhanced()
		if err == nil {
			fs.circuitBreaker.RecordSuccess()
			fs.recordEnhancedSuccess()
			return locks, nil
		}

		fs.circuitBreaker.RecordFailure()
		fs.recordEnhancedFailure()
		fs.log.Warn("Enhanced List failed, falling back: %v", err)
	}

	// Fall back to legacy
	return fs.listFallback()
}

// listEnhanced lists locks using enhanced backend
func (fs *FallbackSystem) listEnhanced() ([]models.ProjectLock, error) {
	fs.recordEnhancedAttempt()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	locks, err := fs.enhanced.ListLocks(ctx)
	if err != nil {
		return nil, err
	}

	var legacyLocks []models.ProjectLock
	for _, lock := range locks {
		if lock.State == LockStateAcquired {
			legacyLock := fs.enhanced.ConvertToLegacy(lock)
			if legacyLock != nil {
				legacyLocks = append(legacyLocks, *legacyLock)
			}
		}
	}

	return legacyLocks, nil
}

// listFallback lists locks using legacy backend
func (fs *FallbackSystem) listFallback() ([]models.ProjectLock, error) {
	fs.recordFallbackAttempt()

	locks, err := fs.legacy.List()
	if err != nil {
		fs.recordFallbackFailure()
		return nil, err
	}

	return locks, nil
}

// GetLock retrieves a specific lock with automatic fallback
func (fs *FallbackSystem) GetLock(project models.Project, workspace string) (*models.ProjectLock, error) {
	// Try enhanced system first
	if !fs.circuitBreaker.ShouldReject() {
		lock, err := fs.getLockEnhanced(project, workspace)
		if err == nil {
			fs.circuitBreaker.RecordSuccess()
			fs.recordEnhancedSuccess()
			return lock, nil
		}

		fs.circuitBreaker.RecordFailure()
		fs.recordEnhancedFailure()
		fs.log.Debug("Enhanced GetLock failed, falling back: %v", err)
	}

	// Fall back to legacy
	return fs.getLockFallback(project, workspace)
}

// getLockEnhanced retrieves lock using enhanced backend
func (fs *FallbackSystem) getLockEnhanced(project models.Project, workspace string) (*models.ProjectLock, error) {
	fs.recordEnhancedAttempt()
	return fs.enhanced.GetLegacyLock(project, workspace)
}

// getLockFallback retrieves lock using legacy backend
func (fs *FallbackSystem) getLockFallback(project models.Project, workspace string) (*models.ProjectLock, error) {
	fs.recordFallbackAttempt()

	lock, err := fs.legacy.GetLock(project, workspace)
	if err != nil {
		fs.recordFallbackFailure()
		return nil, err
	}

	return lock, nil
}

// UnlockByPull releases all locks for a pull request with automatic fallback
func (fs *FallbackSystem) UnlockByPull(repoFullName string, pullNum int) ([]models.ProjectLock, error) {
	// Try enhanced system first
	if !fs.circuitBreaker.ShouldReject() {
		locks, err := fs.unlockByPullEnhanced(repoFullName, pullNum)
		if err == nil {
			fs.circuitBreaker.RecordSuccess()
			fs.recordEnhancedSuccess()
			return locks, nil
		}

		fs.circuitBreaker.RecordFailure()
		fs.recordEnhancedFailure()
		fs.log.Warn("Enhanced UnlockByPull failed, falling back: %v", err)
	}

	// Fall back to legacy
	return fs.unlockByPullFallback(repoFullName, pullNum)
}

// unlockByPullEnhanced releases locks using enhanced backend
func (fs *FallbackSystem) unlockByPullEnhanced(repoFullName string, pullNum int) ([]models.ProjectLock, error) {
	fs.recordEnhancedAttempt()

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Get all locks for this repository
	locks, err := fs.enhanced.ListLocks(ctx)
	if err != nil {
		return nil, err
	}

	var unlockedLocks []models.ProjectLock

	// Find and release locks for this repository
	for _, lock := range locks {
		if lock.Resource.Namespace == repoFullName && lock.State == LockStateAcquired {
			err := fs.enhanced.ReleaseLock(ctx, lock.ID)
			if err != nil {
				fs.log.Warn("Failed to release lock %s during UnlockByPull: %v", lock.ID, err)
				continue
			}

			legacyLock := fs.enhanced.ConvertToLegacy(lock)
			if legacyLock != nil {
				unlockedLocks = append(unlockedLocks, *legacyLock)
			}
		}
	}

	return unlockedLocks, nil
}

// unlockByPullFallback releases locks using legacy backend
func (fs *FallbackSystem) unlockByPullFallback(repoFullName string, pullNum int) ([]models.ProjectLock, error) {
	fs.recordFallbackAttempt()

	locks, err := fs.legacy.UnlockByPull(repoFullName, pullNum)
	if err != nil {
		fs.recordFallbackFailure()
		return nil, err
	}

	return locks, nil
}

// Circuit Breaker Implementation

// ShouldReject determines if requests should be rejected due to circuit breaker
func (cb *CircuitBreaker) ShouldReject() bool {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	switch cb.state {
	case CircuitStateClosed:
		return false
	case CircuitStateOpen:
		return time.Since(cb.lastFailure) < cb.timeout
	case CircuitStateHalfOpen:
		return false // Allow limited requests to test recovery
	default:
		return false
	}
}

// RecordFailure records a failure and potentially opens the circuit
func (cb *CircuitBreaker) RecordFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.failures++
	cb.lastFailure = time.Now()

	if cb.state == CircuitStateClosed && cb.failures >= cb.threshold {
		cb.state = CircuitStateOpen
	} else if cb.state == CircuitStateHalfOpen {
		cb.state = CircuitStateOpen
	}
}

// RecordSuccess records a success and potentially closes the circuit
func (cb *CircuitBreaker) RecordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	if cb.state == CircuitStateHalfOpen {
		cb.state = CircuitStateClosed
		cb.failures = 0
	} else if cb.state == CircuitStateOpen && time.Since(cb.lastFailure) >= cb.timeout {
		cb.state = CircuitStateHalfOpen
	}
}

// GetState returns the current circuit breaker state
func (cb *CircuitBreaker) GetState() CircuitState {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state
}

// GetFailureCount returns the current failure count
func (cb *CircuitBreaker) GetFailureCount() int64 {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.failures
}

// Statistics tracking methods

func (fs *FallbackSystem) recordEnhancedAttempt() {
	fs.stats.mu.Lock()
	defer fs.stats.mu.Unlock()
	fs.stats.EnhancedAttempts++
}

func (fs *FallbackSystem) recordEnhancedSuccess() {
	// Success is recorded when no failure is recorded
}

func (fs *FallbackSystem) recordEnhancedFailure() {
	fs.stats.mu.Lock()
	defer fs.stats.mu.Unlock()
	fs.stats.EnhancedFailures++
}

func (fs *FallbackSystem) recordFallbackAttempt() {
	fs.stats.mu.Lock()
	defer fs.stats.mu.Unlock()
	fs.stats.FallbackAttempts++
	fs.stats.LastFallback = time.Now()
}

func (fs *FallbackSystem) recordFallbackFailure() {
	fs.stats.mu.Lock()
	defer fs.stats.mu.Unlock()
	fs.stats.FallbackFailures++
}

// GetStats returns fallback system statistics
func (fs *FallbackSystem) GetStats() *FallbackStats {
	fs.stats.mu.RLock()
	defer fs.stats.mu.RUnlock()

	// Return a copy
	stats := *fs.stats
	stats.CircuitState = fs.circuitBreaker.GetState()
	return &stats
}

// HealthCheck performs a health check on both backends
func (fs *FallbackSystem) HealthCheck(ctx context.Context) *HealthCheckResult {
	result := &HealthCheckResult{
		Timestamp: time.Now(),
		Enhanced:  fs.checkEnhancedHealth(ctx),
		Legacy:    fs.checkLegacyHealth(ctx),
		Circuit:   fs.circuitBreaker.GetState(),
	}

	result.Overall = result.Enhanced.Healthy || result.Legacy.Healthy
	return result
}

// checkEnhancedHealth checks the health of the enhanced backend
func (fs *FallbackSystem) checkEnhancedHealth(ctx context.Context) *BackendHealth {
	start := time.Now()
	err := fs.enhanced.HealthCheck(ctx)
	duration := time.Since(start)

	return &BackendHealth{
		Healthy:      err == nil,
		Error:        err,
		ResponseTime: duration,
	}
}

// checkLegacyHealth checks the health of the legacy backend
func (fs *FallbackSystem) checkLegacyHealth(ctx context.Context) *BackendHealth {
	start := time.Now()

	// Legacy backends may not have health checks, so we test with a simple operation
	_, err := fs.legacy.List()
	duration := time.Since(start)

	return &BackendHealth{
		Healthy:      err == nil,
		Error:        err,
		ResponseTime: duration,
	}
}

// ResetCircuitBreaker manually resets the circuit breaker
func (fs *FallbackSystem) ResetCircuitBreaker() {
	fs.circuitBreaker.mu.Lock()
	defer fs.circuitBreaker.mu.Unlock()

	fs.circuitBreaker.state = CircuitStateClosed
	fs.circuitBreaker.failures = 0
	fs.log.Info("Circuit breaker manually reset")
}

// Configure updates the fallback system configuration
func (fs *FallbackSystem) Configure(threshold int64, timeout time.Duration) {
	fs.circuitBreaker.mu.Lock()
	defer fs.circuitBreaker.mu.Unlock()

	fs.circuitBreaker.threshold = threshold
	fs.circuitBreaker.timeout = timeout
	fs.log.Info("Fallback system reconfigured: threshold=%d, timeout=%v", threshold, timeout)
}

// Health check result structures

// HealthCheckResult contains health check results for both backends
type HealthCheckResult struct {
	Overall   bool           `json:"overall"`
	Enhanced  *BackendHealth `json:"enhanced"`
	Legacy    *BackendHealth `json:"legacy"`
	Circuit   CircuitState   `json:"circuit_state"`
	Timestamp time.Time      `json:"timestamp"`
}

// BackendHealth contains health information for a single backend
type BackendHealth struct {
	Healthy      bool          `json:"healthy"`
	Error        error         `json:"error,omitempty"`
	ResponseTime time.Duration `json:"response_time"`
}

// String returns a human-readable representation of the health check result
func (hcr *HealthCheckResult) String() string {
	enhancedStatus := "healthy"
	if !hcr.Enhanced.Healthy {
		enhancedStatus = fmt.Sprintf("unhealthy (%v)", hcr.Enhanced.Error)
	}

	legacyStatus := "healthy"
	if !hcr.Legacy.Healthy {
		legacyStatus = fmt.Sprintf("unhealthy (%v)", hcr.Legacy.Error)
	}

	return fmt.Sprintf("Fallback System Health: Overall=%v, Enhanced=%s, Legacy=%s, Circuit=%s",
		hcr.Overall, enhancedStatus, legacyStatus, hcr.Circuit)
}

// IsReadyForMigration checks if the system is ready to migrate from legacy to enhanced
func (fs *FallbackSystem) IsReadyForMigration() (bool, string) {
	health := fs.HealthCheck(context.Background())

	if !health.Enhanced.Healthy {
		return false, fmt.Sprintf("Enhanced backend is not healthy: %v", health.Enhanced.Error)
	}

	if fs.circuitBreaker.GetState() == CircuitStateOpen {
		return false, "Circuit breaker is open indicating recent failures"
	}

	stats := fs.GetStats()
	if stats.EnhancedFailures > 0 && stats.EnhancedAttempts > 0 {
		failureRate := float64(stats.EnhancedFailures) / float64(stats.EnhancedAttempts)
		if failureRate > 0.1 { // More than 10% failure rate
			return false, fmt.Sprintf("Enhanced backend failure rate too high: %.1f%%", failureRate*100)
		}
	}

	return true, "System ready for migration"
}