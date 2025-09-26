// PR #3: Adapter Layer Activation - Runtime Switching and Compatibility Layer
// This file implements the adapter that bridges enhanced and legacy locking systems

package enhanced

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/runatlantis/atlantis/server/core/locking"
	"github.com/runatlantis/atlantis/server/models"
	"go.uber.org/zap"
)

// LockingAdapter provides backward compatibility between enhanced and legacy systems
type LockingAdapter struct {
	mu sync.RWMutex
	
	// Configuration
	config AdapterConfig
	
	// Backend references
	enhanced EnhancedBackend
	legacy locking.Backend
	
	// Runtime state
	currentBackend string // "enhanced" or "legacy"
	health HealthTracker
	logger *zap.Logger
	
	// Metrics and monitoring
	metrics *AdapterMetrics
	eventBus *EventBus
}

// AdapterConfig controls adapter behavior
type AdapterConfig struct {
	// Backend selection
	Enhanced EnhancedBackend
	Legacy locking.Backend
	
	// Fallback behavior
	FallbackEnabled bool
	PreserveFormat bool
	AutoFallback bool
	
	// Health monitoring
	HealthCheckInterval time.Duration
	FailureThreshold int
	RecoveryThreshold int
	
	// Performance settings
	MaxRetries int
	RetryDelay time.Duration
	OperationTimeout time.Duration
}

// EnhancedBackend represents the enhanced locking interface
type EnhancedBackend interface {
	locking.Backend
	
	// Enhanced operations
	TryLockWithPriority(lock models.ProjectLock, priority int) (bool, locking.LockingError)
	TryLockWithTimeout(lock models.ProjectLock, timeout time.Duration) (bool, locking.LockingError)
	ListWithDetails() ([]LockDetails, error)
	GetQueueStatus(project models.Project, workspace string) (*QueueStatus, error)
	
	// Health and monitoring
	Health() HealthStatus
	Metrics() BackendMetrics
}

// LockDetails provides enhanced lock information
type LockDetails struct {
	models.ProjectLock
	AcquiredAt time.Time
	ExpiresAt *time.Time
	QueuePosition *int
	Priority int
}

// QueueStatus shows queue information for a project/workspace
type QueueStatus struct {
	QueueDepth int
	EstimatedWait time.Duration
	CurrentPosition int
	AverageProcessingTime time.Duration
}

// HealthTracker monitors backend health
type HealthTracker struct {
	enhancedFailures int
	legacyFailures int
	lastCheck time.Time
	lastError error
}

// AdapterMetrics tracks adapter performance
type AdapterMetrics struct {
	TotalRequests int64
	EnhancedRequests int64
	LegacyRequests int64
	FallbackEvents int64
	ErrorCount int64
	AverageResponseTime time.Duration
}

// EventBus handles adapter events
type EventBus struct {
	subscribers []EventSubscriber
	mu sync.RWMutex
}

type EventSubscriber interface {
	OnBackendSwitch(from, to string)
	OnFallback(reason string)
	OnError(err error)
}

// NewLockingAdapter creates a new adapter instance
func NewLockingAdapter(config AdapterConfig) *LockingAdapter {
	return &LockingAdapter{
		config: config,
		enhanced: config.Enhanced,
		legacy: config.Legacy,
		currentBackend: "enhanced", // Start with enhanced if available
		health: HealthTracker{},
		metrics: &AdapterMetrics{},
		eventBus: &EventBus{},
		logger: zap.NewNop(), // Will be configured properly
	}
}

// TryLock implements locking.Backend interface with intelligent routing
func (a *LockingAdapter) TryLock(lock models.ProjectLock) (bool, locking.LockingError) {
	a.mu.RLock()
	defer a.mu.RUnlock()
	
	a.metrics.TotalRequests++
	start := time.Now()
	defer func() {
		a.metrics.AverageResponseTime = time.Since(start)
	}()
	
	// Try enhanced backend first
	if a.shouldUseEnhanced() {
		return a.tryLockEnhanced(lock)
	}
	
	// Fall back to legacy backend
	return a.tryLockLegacy(lock)
}

// tryLockEnhanced attempts lock with enhanced backend
func (a *LockingAdapter) tryLockEnhanced(lock models.ProjectLock) (bool, locking.LockingError) {
	a.metrics.EnhancedRequests++
	
	// Try enhanced backend with retry logic
	for attempt := 0; attempt < a.config.MaxRetries; attempt++ {
		acquired, err := a.enhanced.TryLock(lock)
		
		if err == nil {
			// Success - reset failure count
			a.health.enhancedFailures = 0
			return acquired, nil
		}
		
		// Handle error
		a.health.enhancedFailures++
		a.health.lastError = err
		a.metrics.ErrorCount++
		
		// Check if we should fallback
		if a.shouldFallbackToLegacy(err) {
			a.logger.Warn("Enhanced backend failed, falling back to legacy",
				zap.Error(err),
				zap.Int("attempt", attempt+1))
			
			return a.fallbackToLegacy(lock)
		}
		
		// Retry with delay
		if attempt < a.config.MaxRetries-1 {
			time.Sleep(a.config.RetryDelay)
		}
	}
	
	// All retries failed - fallback if enabled
	if a.config.FallbackEnabled {
		return a.fallbackToLegacy(lock)
	}
	
	return false, locking.NewLockingError("enhanced backend failed after all retries", a.health.lastError)
}

// tryLockLegacy attempts lock with legacy backend
func (a *LockingAdapter) tryLockLegacy(lock models.ProjectLock) (bool, locking.LockingError) {
	a.metrics.LegacyRequests++
	
	acquired, err := a.legacy.TryLock(lock)
	if err != nil {
		a.health.legacyFailures++
		a.health.lastError = err
		a.metrics.ErrorCount++
	}
	
	return acquired, err
}

// fallbackToLegacy switches to legacy backend temporarily
func (a *LockingAdapter) fallbackToLegacy(lock models.ProjectLock) (bool, locking.LockingError) {
	a.metrics.FallbackEvents++
	
	// Notify subscribers of fallback
	a.eventBus.NotifyFallback("enhanced backend failure")
	
	// Temporarily switch to legacy
	previousBackend := a.currentBackend
	a.currentBackend = "legacy"
	
	// Restore backend after operation (if auto-recovery enabled)
	defer func() {
		if a.health.enhancedFailures < a.config.RecoveryThreshold {
			a.currentBackend = previousBackend
		}
	}()
	
	return a.tryLockLegacy(lock)
}

// Unlock implements locking.Backend interface
func (a *LockingAdapter) Unlock(lock models.ProjectLock) locking.LockingError {
	a.mu.RLock()
	defer a.mu.RUnlock()
	
	// Try both backends to ensure cleanup
	// This handles cases where locks might exist in either system
	var errors []error
	
	// Try enhanced backend first
	if a.enhanced != nil {
		if err := a.enhanced.Unlock(lock); err != nil {
			errors = append(errors, err)
		}
	}
	
	// Try legacy backend
	if a.legacy != nil {
		if err := a.legacy.Unlock(lock); err != nil {
			errors = append(errors, err)
		}
	}
	
	// Return error only if both backends failed
	if len(errors) == 2 {
		return locking.NewLockingError("unlock failed on both backends", 
			fmt.Errorf("enhanced: %v, legacy: %v", errors[0], errors[1]))
	}
	
	return nil
}

// List implements locking.Backend interface
func (a *LockingAdapter) List() (map[string]models.ProjectLock, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()
	
	// Aggregate locks from both backends
	allLocks := make(map[string]models.ProjectLock)
	
	// Get locks from enhanced backend
	if a.enhanced != nil && a.shouldUseEnhanced() {
		enhancedLocks, err := a.enhanced.List()
		if err == nil {
			for id, lock := range enhancedLocks {
				allLocks[id] = lock
			}
		} else {
			a.logger.Warn("Failed to list enhanced locks", zap.Error(err))
		}
	}
	
	// Get locks from legacy backend (for fallback or dual mode)
	if a.legacy != nil && (a.config.FallbackEnabled || !a.shouldUseEnhanced()) {
		legacyLocks, err := a.legacy.List()
		if err == nil {
			// Merge with enhanced locks (enhanced takes precedence)
			for id, lock := range legacyLocks {
				if _, exists := allLocks[id]; !exists {
					allLocks[id] = lock
				}
			}
		} else {
			a.logger.Warn("Failed to list legacy locks", zap.Error(err))
		}
	}
	
	return allLocks, nil
}

// UnlockByPull implements locking.Backend interface
func (a *LockingAdapter) UnlockByPull(repoFullName string, pullNum int) ([]models.ProjectLock, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()
	
	var allUnlocked []models.ProjectLock
	
	// Unlock from enhanced backend
	if a.enhanced != nil {
		unlocked, err := a.enhanced.UnlockByPull(repoFullName, pullNum)
		if err == nil {
			allUnlocked = append(allUnlocked, unlocked...)
		} else {
			a.logger.Warn("Failed to unlock from enhanced backend", zap.Error(err))
		}
	}
	
	// Unlock from legacy backend
	if a.legacy != nil {
		unlocked, err := a.legacy.UnlockByPull(repoFullName, pullNum)
		if err == nil {
			allUnlocked = append(allUnlocked, unlocked...)
		} else {
			a.logger.Warn("Failed to unlock from legacy backend", zap.Error(err))
		}
	}
	
	return allUnlocked, nil
}

// Enhanced interface methods

// TryLockWithPriority attempts lock with priority (enhanced feature)
func (a *LockingAdapter) TryLockWithPriority(lock models.ProjectLock, priority int) (bool, locking.LockingError) {
	if a.enhanced != nil && a.shouldUseEnhanced() {
		return a.enhanced.TryLockWithPriority(lock, priority)
	}
	
	// Fallback to regular lock
	return a.TryLock(lock)
}

// TryLockWithTimeout attempts lock with timeout (enhanced feature)
func (a *LockingAdapter) TryLockWithTimeout(lock models.ProjectLock, timeout time.Duration) (bool, locking.LockingError) {
	if a.enhanced != nil && a.shouldUseEnhanced() {
		return a.enhanced.TryLockWithTimeout(lock, timeout)
	}
	
	// Fallback to regular lock with context timeout
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	
	done := make(chan struct {bool, locking.LockingError})
	go func() {
		acquired, err := a.TryLock(lock)
		done <- struct {bool, locking.LockingError}{acquired, err}
	}()
	
	select {
	case result := <-done:
		return result.bool, result.locking.LockingError
	case <-ctx.Done():
		return false, locking.NewLockingError("lock acquisition timeout", ctx.Err())
	}
}

// GetQueueStatus returns queue information (enhanced feature)
func (a *LockingAdapter) GetQueueStatus(project models.Project, workspace string) (*QueueStatus, error) {
	if a.enhanced != nil && a.shouldUseEnhanced() {
		return a.enhanced.GetQueueStatus(project, workspace)
	}
	
	// Legacy backend doesn't support queues
	return &QueueStatus{
		QueueDepth: 0,
		EstimatedWait: 0,
		CurrentPosition: 0,
		AverageProcessingTime: 0,
	}, nil
}

// Management methods

// SwitchBackend manually switches between backends
func (a *LockingAdapter) SwitchBackend(backend string) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	
	if backend != "enhanced" && backend != "legacy" {
		return fmt.Errorf("invalid backend: %s", backend)
	}
	
	previous := a.currentBackend
	a.currentBackend = backend
	
	// Notify subscribers
	a.eventBus.NotifyBackendSwitch(previous, backend)
	
	return nil
}

// GetCurrentBackend returns the currently active backend
func (a *LockingAdapter) GetCurrentBackend() string {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.currentBackend
}

// GetMetrics returns adapter performance metrics
func (a *LockingAdapter) GetMetrics() AdapterMetrics {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return *a.metrics
}

// GetHealth returns adapter health status
func (a *LockingAdapter) GetHealth() HealthTracker {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.health
}

// Helper methods

func (a *LockingAdapter) shouldUseEnhanced() bool {
	// Check if enhanced backend is available and healthy
	if a.enhanced == nil {
		return false
	}
	
	// Check current backend selection
	if a.currentBackend == "legacy" {
		return false
	}
	
	// Check failure threshold
	if a.health.enhancedFailures >= a.config.FailureThreshold {
		return false
	}
	
	return true
}

func (a *LockingAdapter) shouldFallbackToLegacy(err error) bool {
	if !a.config.FallbackEnabled {
		return false
	}
	
	// Define conditions for fallback
	// For example: connection errors, timeout errors, etc.
	return true
}

// Event bus methods

func (e *EventBus) Subscribe(subscriber EventSubscriber) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.subscribers = append(e.subscribers, subscriber)
}

func (e *EventBus) NotifyBackendSwitch(from, to string) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	
	for _, subscriber := range e.subscribers {
		subscriber.OnBackendSwitch(from, to)
	}
}

func (e *EventBus) NotifyFallback(reason string) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	
	for _, subscriber := range e.subscribers {
		subscriber.OnFallback(reason)
	}
}

func (e *EventBus) NotifyError(err error) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	
	for _, subscriber := range e.subscribers {
		subscriber.OnError(err)
	}
}
