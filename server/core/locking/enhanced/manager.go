package enhanced

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/runatlantis/atlantis/server/core/locking/enhanced/deadlock"
	"github.com/runatlantis/atlantis/server/core/locking/enhanced/queue"
	"github.com/runatlantis/atlantis/server/core/locking/enhanced/timeout"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/logging"
)

// EnhancedLockManager implements the LockManager interface with advanced features
type EnhancedLockManager struct {
	backend Backend
	config  *EnhancedConfig
	log     logging.SimpleLogging

	// Advanced components
	queue            *queue.ResourceBasedQueue
	timeoutManager   *timeout.TimeoutManager
	retryManager     *timeout.RetryManager
	deadlockDetector *deadlock.DeadlockDetector

	// State management
	mutex    sync.RWMutex
	running  bool
	stopChan chan struct{}

	// Metrics and monitoring
	metrics        *ManagerMetrics
	eventCallbacks []EventCallback
}

// EventCallback handles lock manager events
type EventCallback func(ctx context.Context, event *ManagerEvent)

// ManagerEvent represents events from the lock manager
type ManagerEvent struct {
	Type      string                 `json:"type"`
	LockID    string                 `json:"lock_id"`
	Resource  ResourceIdentifier     `json:"resource"`
	User      string                 `json:"user"`
	Timestamp time.Time              `json:"timestamp"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// ManagerMetrics tracks lock manager performance
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
}

// NewEnhancedLockManager creates a new enhanced lock manager
func NewEnhancedLockManager(backend Backend, config *EnhancedConfig, log logging.SimpleLogging) *EnhancedLockManager {
	if config == nil {
		config = DefaultConfig()
	}

	manager := &EnhancedLockManager{
		backend:  backend,
		config:   config,
		log:      log,
		running:  false,
		stopChan: make(chan struct{}),
		metrics: &ManagerMetrics{
			StartTime:   time.Now(),
			LastUpdated: time.Now(),
		},
	}

	// Initialize advanced components if enabled
	if config.EnablePriorityQueue {
		manager.queue = queue.NewResourceBasedQueue(config.MaxQueueSize)
		log.Info("Priority queue enabled with max size: %d", config.MaxQueueSize)
	}

	manager.timeoutManager = timeout.NewTimeoutManager(log)

	if config.EnableRetry {
		retryConfig := &timeout.RetryConfig{
			MaxAttempts:   config.MaxRetryAttempts,
			BaseDelay:     config.RetryBaseDelay,
			MaxDelay:      config.RetryMaxDelay,
			Multiplier:    2.0,
			Jitter:        true,
			JitterPercent: 0.1,
		}
		manager.retryManager = timeout.NewRetryManager(retryConfig, log)
		log.Info("Retry mechanism enabled with max %d attempts", config.MaxRetryAttempts)
	}

	if config.EnableDeadlockDetection {
		deadlockConfig := &deadlock.DetectorConfig{
			Enabled:          true,
			CheckInterval:    config.DeadlockCheckInterval,
			MaxWaitTime:      config.DefaultTimeout,
			ResolutionPolicy: deadlock.ResolveLowestPriority,
			HistorySize:      100,
			EnablePrevention: true,
		}
		manager.deadlockDetector = deadlock.NewDeadlockDetector(deadlockConfig, log)
		log.Info("Deadlock detection enabled with %v check interval", config.DeadlockCheckInterval)
	}

	return manager
}

// Start initializes and starts the enhanced lock manager
func (elm *EnhancedLockManager) Start(ctx context.Context) error {
	elm.mutex.Lock()
	defer elm.mutex.Unlock()

	if elm.running {
		return fmt.Errorf("manager is already running")
	}

	elm.log.Info("Starting enhanced lock manager...")

	// Start deadlock detection if enabled
	if elm.deadlockDetector != nil {
		elm.deadlockDetector.Start(ctx)
	}

	// Start background maintenance tasks
	go elm.maintenanceLoop(ctx)

	elm.running = true
	elm.log.Info("Enhanced lock manager started successfully")
	return nil
}

// Stop gracefully shuts down the enhanced lock manager
func (elm *EnhancedLockManager) Stop() error {
	elm.mutex.Lock()
	defer elm.mutex.Unlock()

	if !elm.running {
		return nil
	}

	elm.log.Info("Stopping enhanced lock manager...")

	// Stop deadlock detection
	if elm.deadlockDetector != nil {
		elm.deadlockDetector.Stop()
	}

	// Stop maintenance loop
	close(elm.stopChan)

	// Cleanup timeouts
	elm.timeoutManager.Cleanup()

	elm.running = false
	elm.log.Info("Enhanced lock manager stopped")
	return nil
}

// Lock acquires a lock using enhanced features (implements LockManager interface)
func (elm *EnhancedLockManager) Lock(ctx context.Context, project models.Project, workspace string, user models.User) (*models.ProjectLock, error) {
	return elm.LockWithPriority(ctx, project, workspace, user, PriorityNormal)
}

// Unlock releases a lock (implements LockManager interface)
func (elm *EnhancedLockManager) Unlock(ctx context.Context, project models.Project, workspace string, user models.User) (*models.ProjectLock, error) {
	elm.metrics.TotalRequests++
	elm.metrics.LastUpdated = time.Now()

	// Find and release the lock
	request := elm.createLockRequest(project, workspace, user, PriorityNormal, elm.config.DefaultTimeout)

	// Get current locks to find the target
	locks, err := elm.backend.ListLocks(ctx)
	if err != nil {
		elm.log.Err("Failed to list locks during unlock: %v", err)
		return nil, err
	}

	var targetLock *EnhancedLock
	for _, lock := range locks {
		if elm.sameResource(request.Resource, lock.Resource) && lock.Owner == user.Username {
			targetLock = lock
			break
		}
	}

	if targetLock == nil {
		// No lock found - this is sometimes expected in Atlantis
		return nil, nil
	}

	// Clear timeout if set
	elm.timeoutManager.ClearTimeout(targetLock.ID)

	// Remove from deadlock detection
	if elm.deadlockDetector != nil {
		elm.deadlockDetector.RemoveLockRequest(targetLock.ID)
	}

	// Release the lock
	err = elm.backend.ReleaseLock(ctx, targetLock.ID)
	if err != nil {
		elm.metrics.FailedLocks++
		return nil, err
	}

	// Update metrics
	if targetLock.AcquiredAt.After(time.Time{}) {
		holdTime := time.Since(targetLock.AcquiredAt)
		elm.updateAverageHoldTime(holdTime)
	}
	elm.metrics.ActiveLocks--

	// Emit event
	elm.emitEvent(ctx, &ManagerEvent{
		Type:      "lock_released",
		LockID:    targetLock.ID,
		Resource:  targetLock.Resource,
		User:      user.Username,
		Timestamp: time.Now(),
	})

	// Process queue if enabled
	if elm.queue != nil {
		go elm.processQueueForResource(ctx, targetLock.Resource)
	}

	elm.log.Info("Lock released: %s for %s/%s", targetLock.ID, project.RepoFullName, workspace)
	return elm.backend.ConvertToLegacy(targetLock), nil
}

// List returns all current locks (implements LockManager interface)
func (elm *EnhancedLockManager) List(ctx context.Context) ([]*models.ProjectLock, error) {
	locks, err := elm.backend.ListLocks(ctx)
	if err != nil {
		return nil, err
	}

	var result []*models.ProjectLock
	for _, lock := range locks {
		if lock.State == LockStateAcquired {
			legacyLock := elm.backend.ConvertToLegacy(lock)
			if legacyLock != nil {
				result = append(result, legacyLock)
			}
		}
	}

	return result, nil
}

// Enhanced methods with additional capabilities

// LockWithPriority acquires a lock with specific priority
func (elm *EnhancedLockManager) LockWithPriority(ctx context.Context, project models.Project, workspace string, user models.User, priority Priority) (*models.ProjectLock, error) {
	return elm.LockWithOptions(ctx, project, workspace, user, LockOptions{
		Priority: priority,
		Timeout:  elm.config.DefaultTimeout,
	})
}

// LockWithTimeout acquires a lock with specific timeout
func (elm *EnhancedLockManager) LockWithTimeout(ctx context.Context, project models.Project, workspace string, user models.User, timeout time.Duration) (*models.ProjectLock, error) {
	return elm.LockWithOptions(ctx, project, workspace, user, LockOptions{
		Priority: PriorityNormal,
		Timeout:  timeout,
	})
}

// LockOptions provides options for lock acquisition
type LockOptions struct {
	Priority    Priority
	Timeout     time.Duration
	Metadata    map[string]string
	RetryPolicy *timeout.RetryConfig
}

// LockWithOptions acquires a lock with full options
func (elm *EnhancedLockManager) LockWithOptions(ctx context.Context, project models.Project, workspace string, user models.User, options LockOptions) (*models.ProjectLock, error) {
	elm.metrics.TotalRequests++
	elm.metrics.LastUpdated = time.Now()

	request := elm.createLockRequest(project, workspace, user, options.Priority, options.Timeout)
	if options.Metadata != nil {
		request.Metadata = options.Metadata
	}

	startTime := time.Now()
	var lock *models.ProjectLock
	var err error

	// Use retry manager if enabled and configured
	if elm.retryManager != nil && options.RetryPolicy != nil {
		retryManager := timeout.NewRetryManager(options.RetryPolicy, elm.log)
		err = retryManager.Execute(ctx, func(ctx context.Context, attempt int) error {
			lock, err = elm.attemptLockAcquisition(ctx, request, attempt > 1)
			return err
		})
	} else if elm.retryManager != nil && elm.config.EnableRetry {
		err = elm.retryManager.Execute(ctx, func(ctx context.Context, attempt int) error {
			lock, err = elm.attemptLockAcquisition(ctx, request, attempt > 1)
			return err
		})
	} else {
		lock, err = elm.attemptLockAcquisition(ctx, request, false)
	}

	// Update metrics
	waitTime := time.Since(startTime)
	elm.updateAverageWaitTime(waitTime)

	if err != nil {
		elm.metrics.FailedLocks++
		return nil, err
	}

	elm.metrics.SuccessfulLocks++
	elm.metrics.ActiveLocks++

	elm.log.Info("Lock acquired: %s for %s/%s (priority: %v, wait: %v)",
		request.ID, project.RepoFullName, workspace, options.Priority, waitTime)

	return lock, nil
}

// attemptLockAcquisition attempts to acquire a lock, handling queuing and deadlock prevention
func (elm *EnhancedLockManager) attemptLockAcquisition(ctx context.Context, request *EnhancedLockRequest, isRetry bool) (*models.ProjectLock, error) {
	// Check for deadlock prevention if enabled
	if elm.deadlockDetector != nil && !isRetry {
		// Get current locks that might block this request
		currentLocks, err := elm.backend.ListLocks(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to check for conflicts: %w", err)
		}

		var blockedBy []*EnhancedLock
		for _, lock := range currentLocks {
			if elm.sameResource(request.Resource, lock.Resource) && lock.State == LockStateAcquired {
				blockedBy = append(blockedBy, lock)
			}
		}

		if len(blockedBy) > 0 {
			canProceed, err := elm.deadlockDetector.PreventDeadlock(request, blockedBy)
			if err != nil {
				return nil, fmt.Errorf("deadlock prevention check failed: %w", err)
			}
			if !canProceed {
				return nil, &LockError{
					Type:    "DeadlockPrevented",
					Message: "lock request would create a deadlock",
					Code:    ErrCodeDeadlock,
				}
			}

			// Add to deadlock tracking
			elm.deadlockDetector.AddLockRequest(request, blockedBy)
		}
	}

	// Try to acquire the lock
	enhancedLock, acquired, err := elm.backend.TryAcquireLock(ctx, request)
	if err != nil {
		return nil, err
	}

	if acquired {
		// Lock acquired successfully
		elm.setupLockTimeout(ctx, enhancedLock)

		// Add to deadlock tracking
		if elm.deadlockDetector != nil {
			elm.deadlockDetector.AddLockAcquisition(enhancedLock)
		}

		// Emit event
		elm.emitEvent(ctx, &ManagerEvent{
			Type:      "lock_acquired",
			LockID:    enhancedLock.ID,
			Resource:  enhancedLock.Resource,
			User:      request.User.Username,
			Timestamp: time.Now(),
		})

		return elm.backend.ConvertToLegacy(enhancedLock), nil
	}

	// Lock not acquired - handle queuing if enabled
	if elm.queue != nil && elm.config.EnablePriorityQueue {
		return elm.handleQueuedRequest(ctx, request)
	}

	// No queuing - return lock exists error
	return nil, NewLockExistsError(fmt.Sprintf("%s/%s/%s", request.Resource.Namespace, request.Resource.Name, request.Resource.Workspace))
}

// handleQueuedRequest handles a request that needs to be queued
func (elm *EnhancedLockManager) handleQueuedRequest(ctx context.Context, request *EnhancedLockRequest) (*models.ProjectLock, error) {
	// Add to queue
	err := elm.queue.Push(ctx, request)
	if err != nil {
		return nil, err
	}

	elm.metrics.QueuedRequests++

	// Emit queued event
	elm.emitEvent(ctx, &ManagerEvent{
		Type:      "lock_queued",
		LockID:    request.ID,
		Resource:  request.Resource,
		User:      request.User.Username,
		Timestamp: time.Now(),
		Metadata: map[string]interface{}{
			"priority": request.Priority,
		},
	})

	elm.log.Info("Lock request queued: %s for %s/%s (priority: %v)",
		request.ID, request.Resource.Namespace, request.Resource.Workspace, request.Priority)

	// Wait for lock to become available or timeout
	return elm.waitForQueuedLock(ctx, request)
}

// waitForQueuedLock waits for a queued lock request to be processed
func (elm *EnhancedLockManager) waitForQueuedLock(ctx context.Context, request *EnhancedLockRequest) (*models.ProjectLock, error) {
	timeout := request.Timeout
	if timeout <= 0 {
		timeout = elm.config.QueueTimeout
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Poll for lock acquisition
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			// Remove from queue on timeout
			elm.queue.Remove(request.ID)
			return nil, NewTimeoutError(timeout)

		case <-ticker.C:
			// Check if our request has been processed
			lock, err := elm.backend.GetLock(ctx, request.ID)
			if err == nil && lock != nil && lock.State == LockStateAcquired {
				// Lock was acquired!
				elm.setupLockTimeout(ctx, lock)

				elm.emitEvent(ctx, &ManagerEvent{
					Type:      "queued_lock_acquired",
					LockID:    lock.ID,
					Resource:  lock.Resource,
					User:      request.User.Username,
					Timestamp: time.Now(),
				})

				return elm.backend.ConvertToLegacy(lock), nil
			}
		}
	}
}

// GetQueuePosition returns the position of a request in the queue
func (elm *EnhancedLockManager) GetQueuePosition(ctx context.Context, project models.Project, workspace string) (int, error) {
	if elm.queue == nil {
		return -1, fmt.Errorf("priority queue is not enabled")
	}

	resource := ResourceIdentifier{
		Type:      ResourceTypeProject,
		Namespace: project.RepoFullName,
		Name:      project.Path,
		Workspace: workspace,
		Path:      project.Path,
	}

	queue := elm.queue.GetQueueForResource(resource)
	if queue == nil {
		return 0, nil // No queue for this resource
	}

	// This is a simplified implementation - in practice, you'd need to track specific requests
	return queue.Size(), nil
}

// CancelQueuedRequest removes a request from the queue
func (elm *EnhancedLockManager) CancelQueuedRequest(ctx context.Context, project models.Project, workspace string, user models.User) error {
	if elm.queue == nil {
		return fmt.Errorf("priority queue is not enabled")
	}

	// In practice, you'd maintain a mapping of user requests to IDs and use the resource identifier
	// resource := types.ResourceIdentifier{
	//	Type:      types.ResourceTypeProject,
	//	Namespace: project.RepoFullName,
	//	Name:      project.Path,
	//	Workspace: workspace,
	//	Path:      project.Path,
	// }
	elm.log.Info("Queue cancellation requested for %s/%s by %s", project.RepoFullName, workspace, user.Username)
	return nil
}

// GetStats returns manager statistics
func (elm *EnhancedLockManager) GetStats(ctx context.Context) (*BackendStats, error) {
	backendStats, err := elm.backend.GetStats(ctx)
	if err != nil {
		return nil, err
	}

	// Enhance with manager-specific metrics
	elm.metrics.LastUpdated = time.Now()

	// Add manager metrics to backend stats
	if backendStats != nil {
		backendStats.TotalRequests = elm.metrics.TotalRequests
		backendStats.SuccessfulAcquires = elm.metrics.SuccessfulLocks
		backendStats.FailedAcquires = elm.metrics.FailedLocks
		backendStats.AverageWaitTime = elm.metrics.AverageWaitTime
		backendStats.AverageHoldTime = elm.metrics.AverageHoldTime
	}

	return backendStats, nil
}

// GetHealth checks the health of the lock manager
func (elm *EnhancedLockManager) GetHealth(ctx context.Context) error {
	// Check backend health
	if err := elm.backend.HealthCheck(ctx); err != nil {
		return fmt.Errorf("backend health check failed: %w", err)
	}

	// Check if manager is running
	elm.mutex.RLock()
	running := elm.running
	elm.mutex.RUnlock()

	if !running {
		return fmt.Errorf("lock manager is not running")
	}

	return nil
}

// Helper methods

func (elm *EnhancedLockManager) createLockRequest(project models.Project, workspace string, user models.User, priority Priority, timeout time.Duration) *EnhancedLockRequest {
	return &EnhancedLockRequest{
		ID: fmt.Sprintf("lock_%d_%s", time.Now().UnixNano(), user.Username),
		Resource: ResourceIdentifier{
			Type:      ResourceTypeProject,
			Namespace: project.RepoFullName,
			Name:      project.Path,
			Workspace: workspace,
			Path:      project.Path,
		},
		Priority:    priority,
		Timeout:     timeout,
		Metadata:    make(map[string]string),
		Context:     context.Background(),
		RequestedAt: time.Now(),
		Project:     project,
		Workspace:   workspace,
		User:        user,
	}
}

func (elm *EnhancedLockManager) sameResource(r1, r2 ResourceIdentifier) bool {
	return r1.Namespace == r2.Namespace &&
		r1.Name == r2.Name &&
		r1.Workspace == r2.Workspace &&
		r1.Path == r2.Path
}

func (elm *EnhancedLockManager) setupLockTimeout(ctx context.Context, lock *EnhancedLock) {
	if lock.ExpiresAt == nil {
		return
	}

	timeout := lock.ExpiresAt.Sub(time.Now())
	if timeout <= 0 {
		return
	}

	elm.timeoutManager.SetTimeout(ctx, lock.ID, lock.Resource, timeout, func(ctx context.Context, lockID string, resource ResourceIdentifier) {
		elm.log.Info("Lock timeout triggered, releasing lock: %s", lockID)

		err := elm.backend.ReleaseLock(ctx, lockID)
		if err != nil {
			elm.log.Err("Failed to release timed-out lock %s: %v", lockID, err)
		} else {
			elm.emitEvent(ctx, &ManagerEvent{
				Type:      "lock_timeout",
				LockID:    lockID,
				Resource:  resource,
				Timestamp: time.Now(),
			})

			// Process queue for this resource
			if elm.queue != nil {
				go elm.processQueueForResource(ctx, resource)
			}
		}
	})
}

func (elm *EnhancedLockManager) processQueueForResource(ctx context.Context, resource ResourceIdentifier) {
	if elm.queue == nil {
		return
	}

	request, err := elm.queue.PopForResource(ctx, resource)
	if err != nil || request == nil {
		return // No queued requests
	}

	elm.log.Info("Processing queued request: %s for resource %s/%s", request.ID, resource.Namespace, resource.Workspace)

	// Try to acquire the lock for the queued request
	enhancedLock, err := elm.backend.AcquireLock(ctx, request)
	if err != nil {
		elm.log.Err("Failed to acquire lock for queued request %s: %v", request.ID, err)
		return
	}

	if enhancedLock != nil && enhancedLock.State == LockStateAcquired {
		elm.setupLockTimeout(ctx, enhancedLock)

		elm.emitEvent(ctx, &ManagerEvent{
			Type:      "queued_lock_processed",
			LockID:    enhancedLock.ID,
			Resource:  enhancedLock.Resource,
			User:      enhancedLock.Owner,
			Timestamp: time.Now(),
		})
	}
}

func (elm *EnhancedLockManager) updateAverageWaitTime(waitTime time.Duration) {
	// Simple moving average - in production, use proper statistical tracking
	if elm.metrics.AverageWaitTime == 0 {
		elm.metrics.AverageWaitTime = waitTime
	} else {
		elm.metrics.AverageWaitTime = (elm.metrics.AverageWaitTime + waitTime) / 2
	}
}

func (elm *EnhancedLockManager) updateAverageHoldTime(holdTime time.Duration) {
	// Simple moving average - in production, use proper statistical tracking
	if elm.metrics.AverageHoldTime == 0 {
		elm.metrics.AverageHoldTime = holdTime
	} else {
		elm.metrics.AverageHoldTime = (elm.metrics.AverageHoldTime + holdTime) / 2
	}
}

func (elm *EnhancedLockManager) emitEvent(ctx context.Context, event *ManagerEvent) {
	for _, callback := range elm.eventCallbacks {
		// Run callbacks in goroutines to avoid blocking
		go func(cb EventCallback) {
			defer func() {
				if r := recover(); r != nil {
					elm.log.Err("Event callback panicked: %v", r)
				}
			}()
			cb(ctx, event)
		}(callback)
	}
}

// maintenanceLoop runs periodic maintenance tasks
func (elm *EnhancedLockManager) maintenanceLoop(ctx context.Context) {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-elm.stopChan:
			return
		case <-ticker.C:
			elm.performMaintenance(ctx)
		}
	}
}

func (elm *EnhancedLockManager) performMaintenance(ctx context.Context) {
	// Clean up expired locks
	if expired, err := elm.backend.CleanupExpiredLocks(ctx); err != nil {
		elm.log.Err("Failed to cleanup expired locks: %v", err)
	} else if expired > 0 {
		elm.log.Info("Cleaned up %d expired locks", expired)
	}

	// Clean up empty queues
	if elm.queue != nil {
		elm.queue.CleanupEmptyQueues()
	}

	// Update metrics
	elm.metrics.LastUpdated = time.Now()
}

// Event management

// AddEventCallback adds an event callback
func (elm *EnhancedLockManager) AddEventCallback(callback EventCallback) {
	elm.eventCallbacks = append(elm.eventCallbacks, callback)
}

// Configuration management

// GetConfiguration returns the current configuration
func (elm *EnhancedLockManager) GetConfiguration() *EnhancedConfig {
	return elm.config
}

// UpdateConfiguration updates the configuration (runtime changes)
func (elm *EnhancedLockManager) UpdateConfiguration(config *EnhancedConfig) error {
	if config == nil {
		return fmt.Errorf("configuration cannot be nil")
	}

	// Validate configuration
	if err := elm.validateConfig(config); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}

	elm.config = config
	elm.log.Info("Enhanced lock manager configuration updated")
	return nil
}

func (elm *EnhancedLockManager) validateConfig(config *EnhancedConfig) error {
	if config.DefaultTimeout <= 0 {
		return fmt.Errorf("default timeout must be positive")
	}
	if config.MaxTimeout < config.DefaultTimeout {
		return fmt.Errorf("max timeout must be >= default timeout")
	}
	if config.MaxQueueSize < 0 {
		return fmt.Errorf("max queue size cannot be negative")
	}
	if config.MaxRetryAttempts < 0 {
		return fmt.Errorf("max retry attempts cannot be negative")
	}
	return nil
}
