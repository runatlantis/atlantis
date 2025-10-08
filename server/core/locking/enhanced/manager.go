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

// EnhancedLockManager provides centralized orchestration for the enhanced locking system
// This is the main integration point for PR #4 - Enhanced Manager and Events
type EnhancedLockManager struct {
	// Core components
	backend Backend
	config  *EnhancedConfig
	logger  logging.SimpleLogging

	// PR #4 Enhanced components - Event system and metrics collection
	eventManager     *EventManager
	metricsCollector *MetricsCollector

	// Legacy components (from previous PRs)
	legacyManager locking.LockManager

	// Manager state
	mu       sync.RWMutex
	started  bool
	stopCh   chan struct{}
	workers  int
	workerWg sync.WaitGroup
}

// NewEnhancedLockManager creates a new enhanced lock manager
func NewEnhancedLockManager(
	backend Backend,
	config *EnhancedConfig,
	legacyManager locking.LockManager,
	logger logging.SimpleLogging,
) *EnhancedLockManager {
	if config == nil {
		config = DefaultConfig()
	}

	manager := &EnhancedLockManager{
		backend:       backend,
		config:        config,
		legacyManager: legacyManager,
		logger:        logger,
		stopCh:        make(chan struct{}),
		workers:       4, // Default worker pool size
	}

	// Initialize PR #4 enhanced components
	if config.EnableEvents {
		manager.eventManager = NewEventManager(config.EventBufferSize, logger)
		logger.Info("Enhanced event manager enabled for PR #4")
	}

	if config.EnableMetrics {
		manager.metricsCollector = NewMetricsCollector(logger)
		logger.Info("Enhanced metrics collector enabled for PR #4")
	}

	return manager
}

// Start starts the enhanced lock manager and all its components
func (m *EnhancedLockManager) Start(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.started {
		return fmt.Errorf("manager already started")
	}

	m.logger.Info("Starting Enhanced Lock Manager (PR #4)")

	// Start event manager if enabled
	if m.eventManager != nil {
		if err := m.eventManager.Start(ctx); err != nil {
			return fmt.Errorf("failed to start event manager: %w", err)
		}
		m.logger.Info("Event manager started")
	}

	// Start metrics collector if enabled
	if m.metricsCollector != nil {
		if err := m.metricsCollector.Start(ctx); err != nil {
			return fmt.Errorf("failed to start metrics collector: %w", err)
		}
		m.logger.Info("Metrics collector started")
	}

	// Start worker pool
	m.startWorkers(ctx)

	// Emit startup event
	if m.eventManager != nil {
		event := &LockEvent{
			Type:      "manager_started",
			Timestamp: time.Now(),
			Metadata: map[string]string{
				"pr":      "4",
				"version": "enhanced",
				"workers": fmt.Sprintf("%d", m.workers),
			},
		}
		m.eventManager.Emit(event)
	}

	m.started = true
	m.logger.Info("Enhanced Lock Manager started successfully")
	return nil
}

// Stop gracefully stops the enhanced lock manager
func (m *EnhancedLockManager) Stop(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.started {
		return nil
	}

	m.logger.Info("Stopping Enhanced Lock Manager (PR #4)")

	// Signal workers to stop
	close(m.stopCh)

	// Wait for workers to finish
	done := make(chan struct{})
	go func() {
		m.workerWg.Wait()
		close(done)
	}()

	select {
	case <-done:
		m.logger.Info("All workers stopped")
	case <-ctx.Done():
		m.logger.Warn("Context cancelled while waiting for workers")
	}

	// Stop components
	if m.metricsCollector != nil {
		if err := m.metricsCollector.Stop(ctx); err != nil {
			m.logger.Err("Failed to stop metrics collector: %v", err)
		}
	}

	if m.eventManager != nil {
		// Emit shutdown event before stopping
		event := &LockEvent{
			Type:      "manager_stopped",
			Timestamp: time.Now(),
			Metadata: map[string]string{
				"pr":      "4",
				"version": "enhanced",
			},
		}
		m.eventManager.Emit(event)

		if err := m.eventManager.Stop(ctx); err != nil {
			m.logger.Err("Failed to stop event manager: %v", err)
		}
	}

	m.started = false
	m.logger.Info("Enhanced Lock Manager stopped")
	return nil
}

// Lock implements the standard LockManager interface (backward compatibility)
func (m *EnhancedLockManager) Lock(ctx context.Context, project models.Project, workspace string, user models.User) (*models.ProjectLock, error) {
	startTime := time.Now()

	// Record metrics
	if m.metricsCollector != nil {
		m.metricsCollector.RecordLockRequest("standard")
	}

	// Emit lock requested event
	if m.eventManager != nil {
		event := &LockEvent{
			Type:      "lock_requested",
			Timestamp: time.Now(),
			Owner:     user.Username,
			Resource: ResourceIdentifier{
				Type:      ResourceTypeProject,
				Namespace: project.RepoFullName,
				Name:      project.Name,
				Workspace: workspace,
			},
			Metadata: map[string]string{
				"method": "standard",
				"pr":     "4",
			},
		}
		m.eventManager.Emit(event)
	}

	// Use legacy manager for actual locking
	lock, err := m.legacyManager.Lock(ctx, project, workspace, user)

	// Record metrics and events
	duration := time.Since(startTime)
	if err != nil {
		if m.metricsCollector != nil {
			m.metricsCollector.RecordLockFailure("standard", duration)
		}
		if m.eventManager != nil {
			event := &LockEvent{
				Type:      "lock_failed",
				Timestamp: time.Now(),
				Owner:     user.Username,
				Resource: ResourceIdentifier{
					Type:      ResourceTypeProject,
					Namespace: project.RepoFullName,
					Name:      project.Name,
					Workspace: workspace,
				},
				Metadata: map[string]string{
					"error": err.Error(),
					"pr":    "4",
				},
			}
			m.eventManager.Emit(event)
		}
		return nil, err
	}

	// Record successful acquisition
	if m.metricsCollector != nil {
		m.metricsCollector.RecordLockAcquisition("standard", duration)
	}

	if m.eventManager != nil {
		event := &LockEvent{
			Type:      "lock_acquired",
			LockID:    fmt.Sprintf("%s:%s", project.RepoFullName, workspace),
			Timestamp: time.Now(),
			Owner:     user.Username,
			Resource: ResourceIdentifier{
				Type:      ResourceTypeProject,
				Namespace: project.RepoFullName,
				Name:      project.Name,
				Workspace: workspace,
			},
			Metadata: map[string]string{
				"duration_ms": fmt.Sprintf("%d", duration.Milliseconds()),
				"pr":          "4",
			},
		}
		m.eventManager.Emit(event)
	}

	return lock, nil
}

// Unlock implements the standard LockManager interface (backward compatibility)
func (m *EnhancedLockManager) Unlock(ctx context.Context, project models.Project, workspace string, user models.User) (*models.ProjectLock, error) {
	startTime := time.Now()

	// Emit unlock event
	if m.eventManager != nil {
		event := &LockEvent{
			Type:      "lock_release_requested",
			Timestamp: time.Now(),
			Owner:     user.Username,
			Resource: ResourceIdentifier{
				Type:      ResourceTypeProject,
				Namespace: project.RepoFullName,
				Name:      project.Name,
				Workspace: workspace,
			},
			Metadata: map[string]string{
				"pr": "4",
			},
		}
		m.eventManager.Emit(event)
	}

	// Use legacy manager for actual unlocking
	lock, err := m.legacyManager.Unlock(ctx, project, workspace, user)

	// Record metrics and events
	duration := time.Since(startTime)
	if err != nil {
		m.logger.Err("Failed to unlock %s:%s for user %s: %v", project.RepoFullName, workspace, user.Username, err)
		return nil, err
	}

	// Record successful release
	if m.metricsCollector != nil {
		m.metricsCollector.RecordLockRelease(duration)
	}

	if m.eventManager != nil {
		event := &LockEvent{
			Type:      "lock_released",
			LockID:    fmt.Sprintf("%s:%s", project.RepoFullName, workspace),
			Timestamp: time.Now(),
			Owner:     user.Username,
			Resource: ResourceIdentifier{
				Type:      ResourceTypeProject,
				Namespace: project.RepoFullName,
				Name:      project.Name,
				Workspace: workspace,
			},
			Metadata: map[string]string{
				"pr": "4",
			},
		}
		m.eventManager.Emit(event)
	}

	return lock, nil
}

// List implements the standard LockManager interface (backward compatibility)
func (m *EnhancedLockManager) List(ctx context.Context) ([]*models.ProjectLock, error) {
	if m.metricsCollector != nil {
		m.metricsCollector.RecordLockRequest("list")
	}

	return m.legacyManager.List(ctx)
}

// LockWithPriority provides enhanced locking with priority support
func (m *EnhancedLockManager) LockWithPriority(ctx context.Context, project models.Project, workspace string, user models.User, priority Priority) (*models.ProjectLock, error) {
	startTime := time.Now()

	// Record enhanced metrics
	if m.metricsCollector != nil {
		m.metricsCollector.RecordLockRequest(fmt.Sprintf("priority_%s", priority.String()))
	}

	// Emit enhanced lock requested event
	if m.eventManager != nil {
		event := &LockEvent{
			Type:      "enhanced_lock_requested",
			Timestamp: time.Now(),
			Owner:     user.Username,
			Resource: ResourceIdentifier{
				Type:      ResourceTypeProject,
				Namespace: project.RepoFullName,
				Name:      project.Name,
				Workspace: workspace,
			},
			Metadata: map[string]string{
				"priority": priority.String(),
				"method":   "enhanced",
				"pr":       "4",
			},
		}
		m.eventManager.Emit(event)
	}

	// For now, delegate to standard lock (enhanced priority logic would go here in future PRs)
	lock, err := m.Lock(ctx, project, workspace, user)

	duration := time.Since(startTime)
	if err != nil {
		if m.metricsCollector != nil {
			m.metricsCollector.RecordLockFailure(fmt.Sprintf("priority_%s", priority.String()), duration)
		}
		return nil, err
	}

	if m.metricsCollector != nil {
		m.metricsCollector.RecordLockAcquisition(fmt.Sprintf("priority_%s", priority.String()), duration)
	}

	return lock, nil
}

// GetStats returns performance and operational statistics
func (m *EnhancedLockManager) GetStats(ctx context.Context) (*ManagerStats, error) {
	stats := &ManagerStats{
		ManagerVersion: "enhanced-pr4",
		Started:        m.started,
		Workers:        m.workers,
		Timestamp:      time.Now(),
	}

	// Get metrics if available
	if m.metricsCollector != nil {
		metricsStats := m.metricsCollector.GetStats()
		stats.Metrics = metricsStats
		stats.HealthScore = metricsStats.HealthScore
	}

	// Get event stats if available
	if m.eventManager != nil {
		eventStats := m.eventManager.GetStats()
		stats.Events = eventStats
	}

	return stats, nil
}

// GetHealth performs health check on all components
func (m *EnhancedLockManager) GetHealth(ctx context.Context) error {
	if !m.started {
		return fmt.Errorf("manager not started")
	}

	// Check backend health
	if m.backend != nil {
		if err := m.backend.HealthCheck(ctx); err != nil {
			return fmt.Errorf("backend health check failed: %w", err)
		}
	}

	// Check legacy manager health if available
	if healthChecker, ok := m.legacyManager.(interface{ GetHealth(context.Context) error }); ok {
		if err := healthChecker.GetHealth(ctx); err != nil {
			return fmt.Errorf("legacy manager health check failed: %w", err)
		}
	}

	return nil
}

// GetEventManager returns the event manager for advanced usage
func (m *EnhancedLockManager) GetEventManager() *EventManager {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.eventManager
}

// GetMetricsCollector returns the metrics collector for advanced usage
func (m *EnhancedLockManager) GetMetricsCollector() *MetricsCollector {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.metricsCollector
}

// startWorkers starts the worker pool for handling concurrent requests
func (m *EnhancedLockManager) startWorkers(ctx context.Context) {
	for i := 0; i < m.workers; i++ {
		m.workerWg.Add(1)
		go m.worker(ctx, i)
	}
	m.logger.Info("Started %d workers for enhanced lock manager", m.workers)
}

// worker processes lock requests from the queue
func (m *EnhancedLockManager) worker(ctx context.Context, workerID int) {
	defer m.workerWg.Done()

	m.logger.Debug("Worker %d started", workerID)
	defer m.logger.Debug("Worker %d stopped", workerID)

	for {
		select {
		case <-m.stopCh:
			return
		case <-ctx.Done():
			return
		default:
			// Worker logic for processing queued requests would go here
			// For now, just sleep to prevent busy waiting
			time.Sleep(100 * time.Millisecond)
		}
	}
}

// ManagerStats provides comprehensive statistics about the enhanced lock manager
type ManagerStats struct {
	ManagerVersion string                 `json:"manager_version"`
	Started        bool                   `json:"started"`
	Workers        int                    `json:"workers"`
	HealthScore    int                    `json:"health_score"`
	Timestamp      time.Time              `json:"timestamp"`
	Metrics        *MetricsStats          `json:"metrics,omitempty"`
	Events         *EventStats            `json:"events,omitempty"`
	Components     map[string]interface{} `json:"components,omitempty"`
}