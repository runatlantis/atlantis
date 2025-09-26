package enhanced

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/runatlantis/atlantis/server/core/locking/enhanced/backends"
	"github.com/runatlantis/atlantis/server/core/locking/enhanced/deadlock"
	"github.com/runatlantis/atlantis/server/core/locking/enhanced/queue"
	"github.com/runatlantis/atlantis/server/core/locking/enhanced/timeout"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/logging"
)

// LockOrchestrator coordinates all enhanced locking components and provides centralized management
type LockOrchestrator struct {
	// Core components
	backend          Backend
	manager          *EnhancedLockManager
	config          *EnhancedConfig
	log             logging.SimpleLogging

	// Advanced features
	eventBus        *EventBus
	metricsCollector *MetricsCollector
	queue           *queue.ResourceBasedQueue
	timeoutManager  *timeout.TimeoutManager
	retryManager    *timeout.RetryManager
	deadlockDetector *deadlock.DeadlockDetector

	// State management
	mutex           sync.RWMutex
	running         bool
	startTime       time.Time
	stopChan        chan struct{}
	components      map[string]ComponentStatus

	// Coordination
	requestChan     chan *LockOrchestrationRequest
	responseChan    chan *LockOrchestrationResponse
	workerPool      *WorkerPool
}

// ComponentStatus tracks the status of orchestrator components
type ComponentStatus struct {
	Name        string    `json:"name"`
	Status      string    `json:"status"` // "starting", "running", "stopped", "error"
	Health      int       `json:"health"` // 0-100
	LastCheck   time.Time `json:"last_check"`
	Error       string    `json:"error,omitempty"`
	StartTime   time.Time `json:"start_time"`
	Restarts    int       `json:"restarts"`
}

// LockOrchestrationRequest represents a request to the orchestrator
type LockOrchestrationRequest struct {
	ID           string                     `json:"id"`
	Type         OrchestrationRequestType   `json:"type"`
	Project      models.Project             `json:"project"`
	Workspace    string                     `json:"workspace"`
	User         models.User                `json:"user"`
	Priority     Priority                   `json:"priority"`
	Timeout      time.Duration              `json:"timeout"`
	Metadata     map[string]string          `json:"metadata,omitempty"`
	Context      context.Context            `json:"-"`
	ResponseChan chan *LockOrchestrationResponse `json:"-"`
	RequestedAt  time.Time                  `json:"requested_at"`
}

// LockOrchestrationResponse represents a response from the orchestrator
type LockOrchestrationResponse struct {
	RequestID   string                 `json:"request_id"`
	Success     bool                   `json:"success"`
	Lock        *models.ProjectLock    `json:"lock,omitempty"`
	Error       string                 `json:"error,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	ProcessedAt time.Time              `json:"processed_at"`
	Duration    time.Duration          `json:"duration"`
}

// OrchestrationRequestType defines types of orchestration requests
type OrchestrationRequestType string

const (
	RequestTypeLock       OrchestrationRequestType = "lock"
	RequestTypeUnlock     OrchestrationRequestType = "unlock"
	RequestTypeList       OrchestrationRequestType = "list"
	RequestTypeStatus     OrchestrationRequestType = "status"
	RequestTypeHealth     OrchestrationRequestType = "health"
	RequestTypeMetrics    OrchestrationRequestType = "metrics"
	RequestTypeCancel     OrchestrationRequestType = "cancel"
	RequestTypeRefresh    OrchestrationRequestType = "refresh"
	RequestTypeTransfer   OrchestrationRequestType = "transfer"
)

// WorkerPool manages a pool of workers for processing orchestration requests
type WorkerPool struct {
	workers    int
	requestChan chan *LockOrchestrationRequest
	workerWg   sync.WaitGroup
	stopChan   chan struct{}
	log        logging.SimpleLogging
	orchestrator *LockOrchestrator
}

// NewLockOrchestrator creates a new lock orchestrator with all components
func NewLockOrchestrator(backend Backend, config *EnhancedConfig, log logging.SimpleLogging) *LockOrchestrator {
	if config == nil {
		config = DefaultConfig()
	}

	orchestrator := &LockOrchestrator{
		backend:      backend,
		config:       config,
		log:          log,
		stopChan:     make(chan struct{}),
		components:   make(map[string]ComponentStatus),
		requestChan:  make(chan *LockOrchestrationRequest, 1000),
		responseChan: make(chan *LockOrchestrationResponse, 1000),
	}

	// Initialize components
	orchestrator.initializeComponents()

	return orchestrator
}

// initializeComponents sets up all orchestrator components
func (lo *LockOrchestrator) initializeComponents() {
	lo.log.Info("Initializing orchestrator components...")

	// Initialize event bus
	lo.eventBus = NewEventBus(lo.config.EventBufferSize, lo.log)
	lo.components["eventBus"] = ComponentStatus{
		Name:      "EventBus",
		Status:    "initialized",
		Health:    100,
		LastCheck: time.Now(),
		StartTime: time.Now(),
	}

	// Initialize metrics collector
	lo.metricsCollector = NewMetricsCollector(lo.log)
	lo.components["metricsCollector"] = ComponentStatus{
		Name:      "MetricsCollector",
		Status:    "initialized",
		Health:    100,
		LastCheck: time.Now(),
		StartTime: time.Now(),
	}

	// Initialize enhanced lock manager
	lo.manager = NewEnhancedLockManager(lo.backend, lo.config, lo.log)
	lo.components["lockManager"] = ComponentStatus{
		Name:      "EnhancedLockManager",
		Status:    "initialized",
		Health:    100,
		LastCheck: time.Now(),
		StartTime: time.Now(),
	}

	// Initialize advanced components if enabled
	if lo.config.EnablePriorityQueue {
		lo.queue = queue.NewResourceBasedQueue(lo.config.MaxQueueSize)
		lo.components["queue"] = ComponentStatus{
			Name:      "PriorityQueue",
			Status:    "initialized",
			Health:    100,
			LastCheck: time.Now(),
			StartTime: time.Now(),
		}
	}

	lo.timeoutManager = timeout.NewTimeoutManager(lo.log)
	lo.components["timeoutManager"] = ComponentStatus{
		Name:      "TimeoutManager",
		Status:    "initialized",
		Health:    100,
		LastCheck: time.Now(),
		StartTime: time.Now(),
	}

	if lo.config.EnableRetry {
		retryConfig := &timeout.RetryConfig{
			MaxAttempts:   lo.config.MaxRetryAttempts,
			BaseDelay:     lo.config.RetryBaseDelay,
			MaxDelay:      lo.config.RetryMaxDelay,
			Multiplier:    2.0,
			Jitter:        true,
			JitterPercent: 0.1,
		}
		lo.retryManager = timeout.NewRetryManager(retryConfig, lo.log)
		lo.components["retryManager"] = ComponentStatus{
			Name:      "RetryManager",
			Status:    "initialized",
			Health:    100,
			LastCheck: time.Now(),
			StartTime: time.Now(),
		}
	}

	if lo.config.EnableDeadlockDetection {
		deadlockConfig := &deadlock.DetectorConfig{
			Enabled:           true,
			CheckInterval:     lo.config.DeadlockCheckInterval,
			MaxWaitTime:       lo.config.DefaultTimeout,
			ResolutionPolicy:  deadlock.ResolveLowestPriority,
			HistorySize:       100,
			EnablePrevention:  true,
		}
		lo.deadlockDetector = deadlock.NewDeadlockDetector(deadlockConfig, lo.log)
		lo.components["deadlockDetector"] = ComponentStatus{
			Name:      "DeadlockDetector",
			Status:    "initialized",
			Health:    100,
			LastCheck: time.Now(),
			StartTime: time.Now(),
		}
	}

	// Initialize worker pool
	lo.workerPool = NewWorkerPool(8, lo.requestChan, lo, lo.log)
	lo.components["workerPool"] = ComponentStatus{
		Name:      "WorkerPool",
		Status:    "initialized",
		Health:    100,
		LastCheck: time.Now(),
		StartTime: time.Now(),
	}

	lo.log.Info("Orchestrator components initialized successfully")
}

// Start begins orchestration of all components
func (lo *LockOrchestrator) Start(ctx context.Context) error {
	lo.mutex.Lock()
	defer lo.mutex.Unlock()

	if lo.running {
		return fmt.Errorf("orchestrator is already running")
	}

	lo.log.Info("Starting lock orchestrator...")
	lo.startTime = time.Now()
	lo.running = true

	// Start components in dependency order
	if err := lo.startComponents(ctx); err != nil {
		lo.running = false
		return fmt.Errorf("failed to start components: %w", err)
	}

	// Setup event handlers
	lo.setupEventHandlers()

	// Start health monitoring
	go lo.healthMonitorLoop(ctx)

	// Start worker pool
	lo.workerPool.Start()

	lo.log.Info("Lock orchestrator started successfully")
	return nil
}

// Stop gracefully shuts down the orchestrator and all components
func (lo *LockOrchestrator) Stop() error {
	lo.mutex.Lock()
	defer lo.mutex.Unlock()

	if !lo.running {
		return nil
	}

	lo.log.Info("Stopping lock orchestrator...")

	// Stop worker pool first
	lo.workerPool.Stop()

	// Stop components in reverse dependency order
	lo.stopComponents()

	// Signal stop
	close(lo.stopChan)
	lo.running = false

	lo.log.Info("Lock orchestrator stopped")
	return nil
}

// Lock acquires a lock through the orchestrator
func (lo *LockOrchestrator) Lock(ctx context.Context, project models.Project, workspace string, user models.User) (*models.ProjectLock, error) {
	return lo.LockWithOptions(ctx, project, workspace, user, LockRequestOptions{
		Priority: PriorityNormal,
		Timeout:  lo.config.DefaultTimeout,
	})
}

// LockRequestOptions provides options for lock requests
type LockRequestOptions struct {
	Priority Priority
	Timeout  time.Duration
	Metadata map[string]string
}

// LockWithOptions acquires a lock with specific options through the orchestrator
func (lo *LockOrchestrator) LockWithOptions(ctx context.Context, project models.Project, workspace string, user models.User, options LockRequestOptions) (*models.ProjectLock, error) {
	request := &LockOrchestrationRequest{
		ID:           lo.generateRequestID(),
		Type:         RequestTypeLock,
		Project:      project,
		Workspace:    workspace,
		User:         user,
		Priority:     options.Priority,
		Timeout:      options.Timeout,
		Metadata:     options.Metadata,
		Context:      ctx,
		ResponseChan: make(chan *LockOrchestrationResponse, 1),
		RequestedAt:  time.Now(),
	}

	return lo.processRequest(ctx, request)
}

// Unlock releases a lock through the orchestrator
func (lo *LockOrchestrator) Unlock(ctx context.Context, project models.Project, workspace string, user models.User) (*models.ProjectLock, error) {
	request := &LockOrchestrationRequest{
		ID:           lo.generateRequestID(),
		Type:         RequestTypeUnlock,
		Project:      project,
		Workspace:    workspace,
		User:         user,
		Context:      ctx,
		ResponseChan: make(chan *LockOrchestrationResponse, 1),
		RequestedAt:  time.Now(),
	}

	return lo.processRequest(ctx, request)
}

// List returns all locks through the orchestrator
func (lo *LockOrchestrator) List(ctx context.Context) ([]*models.ProjectLock, error) {
	request := &LockOrchestrationRequest{
		ID:           lo.generateRequestID(),
		Type:         RequestTypeList,
		Context:      ctx,
		ResponseChan: make(chan *LockOrchestrationResponse, 1),
		RequestedAt:  time.Now(),
	}

	lock, err := lo.processRequest(ctx, request)
	if err != nil {
		return nil, err
	}

	// For list operations, we need to use the manager directly
	return lo.manager.List(ctx)
}

// GetStatus returns orchestrator status and component health
func (lo *LockOrchestrator) GetStatus() *OrchestratorStatus {
	lo.mutex.RLock()
	defer lo.mutex.RUnlock()

	status := &OrchestratorStatus{
		Running:     lo.running,
		StartTime:   lo.startTime,
		Uptime:      time.Since(lo.startTime),
		Components:  make(map[string]ComponentStatus),
		HealthScore: lo.calculateHealthScore(),
		LastUpdated: time.Now(),
	}

	for name, component := range lo.components {
		status.Components[name] = component
	}

	return status
}

// OrchestratorStatus represents the current status of the orchestrator
type OrchestratorStatus struct {
	Running     bool                         `json:"running"`
	StartTime   time.Time                    `json:"start_time"`
	Uptime      time.Duration                `json:"uptime"`
	Components  map[string]ComponentStatus   `json:"components"`
	HealthScore int                          `json:"health_score"`
	LastUpdated time.Time                    `json:"last_updated"`
	Metrics     *MetricsSnapshot             `json:"metrics,omitempty"`
}

// GetMetrics returns current metrics snapshot
func (lo *LockOrchestrator) GetMetrics() *MetricsSnapshot {
	if lo.metricsCollector == nil {
		return nil
	}
	return lo.metricsCollector.GetSnapshot()
}

// GetComponentHealth returns health information for all components
func (lo *LockOrchestrator) GetComponentHealth() map[string]ComponentStatus {
	lo.mutex.RLock()
	defer lo.mutex.RUnlock()

	health := make(map[string]ComponentStatus)
	for name, status := range lo.components {
		health[name] = status
	}

	return health
}

// Private methods

func (lo *LockOrchestrator) processRequest(ctx context.Context, request *LockOrchestrationRequest) (*models.ProjectLock, error) {
	startTime := time.Now()

	// Validate request
	if err := lo.validateRequest(request); err != nil {
		return nil, err
	}

	// Add timeout to context if specified
	if request.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, request.Timeout)
		defer cancel()
		request.Context = ctx
	}

	// Emit request event
	lo.emitEvent(&LockEventData{
		Type:      EventLockRequested,
		RequestID: request.ID,
		Resource:  lo.createResourceIdentifier(request.Project, request.Workspace),
		Owner:     request.User.Username,
		Priority:  request.Priority,
		Timestamp: time.Now(),
		Metadata: map[string]interface{}{
			"request_type": string(request.Type),
		},
	})

	// Send request to worker pool
	select {
	case lo.requestChan <- request:
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	// Wait for response
	select {
	case response := <-request.ResponseChan:
		duration := time.Since(startTime)

		// Update metrics
		lo.metricsCollector.UpdateManagerMetrics(func(metrics *ManagerMetrics) {
			metrics.TotalRequests++
			if response.Success {
				if request.Type == RequestTypeLock {
					metrics.SuccessfulLocks++
				}
			} else {
				metrics.FailedLocks++
			}
			// Update average wait time
			if metrics.AverageWaitTime == 0 {
				metrics.AverageWaitTime = duration
			} else {
				metrics.AverageWaitTime = (metrics.AverageWaitTime + duration) / 2
			}
		})

		if !response.Success {
			return nil, fmt.Errorf("orchestration request failed: %s", response.Error)
		}

		return response.Lock, nil

	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func (lo *LockOrchestrator) validateRequest(request *LockOrchestrationRequest) error {
	if request == nil {
		return fmt.Errorf("request cannot be nil")
	}

	if request.ID == "" {
		request.ID = lo.generateRequestID()
	}

	if request.Type == "" {
		return fmt.Errorf("request type is required")
	}

	if request.Type == RequestTypeLock || request.Type == RequestTypeUnlock {
		if request.Project.RepoFullName == "" {
			return fmt.Errorf("project repository name is required")
		}
		if request.Workspace == "" {
			return fmt.Errorf("workspace is required")
		}
		if request.User.Username == "" {
			return fmt.Errorf("user is required")
		}
	}

	return nil
}

func (lo *LockOrchestrator) startComponents(ctx context.Context) error {
	// Start event bus
	if err := lo.eventBus.Start(ctx); err != nil {
		return fmt.Errorf("failed to start event bus: %w", err)
	}
	lo.updateComponentStatus("eventBus", "running", 100, "")

	// Start metrics collector
	if err := lo.metricsCollector.Start(ctx); err != nil {
		return fmt.Errorf("failed to start metrics collector: %w", err)
	}
	lo.updateComponentStatus("metricsCollector", "running", 100, "")

	// Start lock manager
	if err := lo.manager.Start(ctx); err != nil {
		return fmt.Errorf("failed to start lock manager: %w", err)
	}
	lo.updateComponentStatus("lockManager", "running", 100, "")

	// Start deadlock detector if enabled
	if lo.deadlockDetector != nil {
		lo.deadlockDetector.Start(ctx)
		lo.updateComponentStatus("deadlockDetector", "running", 100, "")
	}

	return nil
}

func (lo *LockOrchestrator) stopComponents() {
	// Stop in reverse order

	// Stop deadlock detector
	if lo.deadlockDetector != nil {
		lo.deadlockDetector.Stop()
		lo.updateComponentStatus("deadlockDetector", "stopped", 0, "")
	}

	// Stop lock manager
	if lo.manager != nil {
		lo.manager.Stop()
		lo.updateComponentStatus("lockManager", "stopped", 0, "")
	}

	// Stop metrics collector
	if lo.metricsCollector != nil {
		lo.metricsCollector.Stop()
		lo.updateComponentStatus("metricsCollector", "stopped", 0, "")
	}

	// Stop event bus
	if lo.eventBus != nil {
		lo.eventBus.Stop()
		lo.updateComponentStatus("eventBus", "stopped", 0, "")
	}
}

func (lo *LockOrchestrator) setupEventHandlers() {
	// Setup logging event handler
	loggingHandler := NewLoggingEventHandler(lo.log, nil)
	lo.eventBus.Subscribe(loggingHandler, nil)

	// Setup metrics event handler
	if lo.metricsCollector != nil {
		metricsHandler := NewMetricsEventHandler(lo.metricsCollector.GetManagerMetrics())
		lo.eventBus.Subscribe(metricsHandler, nil)
	}
}

func (lo *LockOrchestrator) healthMonitorLoop(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-lo.stopChan:
			return
		case <-ticker.C:
			lo.performHealthChecks()
		}
	}
}

func (lo *LockOrchestrator) performHealthChecks() {
	// Check backend health
	if err := lo.backend.HealthCheck(context.Background()); err != nil {
		lo.log.Warn("Backend health check failed: %v", err)
		lo.updateComponentStatus("backend", "error", 0, err.Error())
	} else {
		lo.updateComponentStatus("backend", "running", 100, "")
	}

	// Check other components
	lo.checkComponentHealth()
}

func (lo *LockOrchestrator) checkComponentHealth() {
	// This is a simplified health check - in production, implement proper health checks
	for name, component := range lo.components {
		if component.Status == "running" && time.Since(component.LastCheck) > 2*time.Minute {
			lo.updateComponentStatus(name, "warning", 50, "No recent activity")
		}
	}
}

func (lo *LockOrchestrator) updateComponentStatus(name, status string, health int, errorMsg string) {
	lo.mutex.Lock()
	defer lo.mutex.Unlock()

	component := lo.components[name]
	component.Status = status
	component.Health = health
	component.LastCheck = time.Now()
	if errorMsg != "" {
		component.Error = errorMsg
	} else {
		component.Error = ""
	}

	lo.components[name] = component
}

func (lo *LockOrchestrator) calculateHealthScore() int {
	if len(lo.components) == 0 {
		return 0
	}

	total := 0
	for _, component := range lo.components {
		total += component.Health
	}

	return total / len(lo.components)
}

func (lo *LockOrchestrator) createResourceIdentifier(project models.Project, workspace string) ResourceIdentifier {
	return ResourceIdentifier{
		Type:      ResourceTypeProject,
		Namespace: project.RepoFullName,
		Name:      project.Path,
		Workspace: workspace,
		Path:      project.Path,
	}
}

func (lo *LockOrchestrator) emitEvent(event *LockEventData) {
	if lo.eventBus != nil {
		lo.eventBus.Publish(context.Background(), event)
	}
}

func (lo *LockOrchestrator) generateRequestID() string {
	return fmt.Sprintf("req_%d_%d", time.Now().UnixNano(), len(lo.requestChan))
}

// Worker Pool Implementation

// NewWorkerPool creates a new worker pool
func NewWorkerPool(workers int, requestChan chan *LockOrchestrationRequest, orchestrator *LockOrchestrator, log logging.SimpleLogging) *WorkerPool {
	return &WorkerPool{
		workers:     workers,
		requestChan: requestChan,
		stopChan:    make(chan struct{}),
		log:         log,
		orchestrator: orchestrator,
	}
}

// Start begins processing requests with worker pool
func (wp *WorkerPool) Start() {
	wp.log.Info("Starting worker pool with %d workers", wp.workers)

	for i := 0; i < wp.workers; i++ {
		wp.workerWg.Add(1)
		go wp.worker(i)
	}
}

// Stop gracefully shuts down the worker pool
func (wp *WorkerPool) Stop() {
	wp.log.Info("Stopping worker pool...")
	close(wp.stopChan)
	wp.workerWg.Wait()
	wp.log.Info("Worker pool stopped")
}

// worker processes orchestration requests
func (wp *WorkerPool) worker(id int) {
	defer wp.workerWg.Done()

	wp.log.Debug("Worker %d started", id)

	for {
		select {
		case <-wp.stopChan:
			wp.log.Debug("Worker %d stopping", id)
			return
		case request := <-wp.requestChan:
			wp.processRequest(id, request)
		}
	}
}

// processRequest processes a single orchestration request
func (wp *WorkerPool) processRequest(workerID int, request *LockOrchestrationRequest) {
	startTime := time.Now()

	wp.log.Debug("Worker %d processing request %s of type %s", workerID, request.ID, request.Type)

	response := &LockOrchestrationResponse{
		RequestID:   request.ID,
		ProcessedAt: time.Now(),
	}

	defer func() {
		response.Duration = time.Since(startTime)

		// Send response
		select {
		case request.ResponseChan <- response:
		case <-time.After(5 * time.Second):
			wp.log.Error("Response timeout for request %s", request.ID)
		}
	}()

	// Process request based on type
	switch request.Type {
	case RequestTypeLock:
		lock, err := wp.orchestrator.manager.LockWithOptions(request.Context, request.Project, request.Workspace, request.User, LockOptions{
			Priority: request.Priority,
			Timeout:  request.Timeout,
			Metadata: request.Metadata,
		})
		if err != nil {
			response.Success = false
			response.Error = err.Error()
		} else {
			response.Success = true
			response.Lock = lock
		}

	case RequestTypeUnlock:
		lock, err := wp.orchestrator.manager.Unlock(request.Context, request.Project, request.Workspace, request.User)
		if err != nil {
			response.Success = false
			response.Error = err.Error()
		} else {
			response.Success = true
			response.Lock = lock
		}

	case RequestTypeList:
		locks, err := wp.orchestrator.manager.List(request.Context)
		if err != nil {
			response.Success = false
			response.Error = err.Error()
		} else {
			response.Success = true
			response.Metadata = map[string]interface{}{
				"locks": locks,
				"count": len(locks),
			}
		}

	case RequestTypeHealth:
		err := wp.orchestrator.manager.GetHealth(request.Context)
		if err != nil {
			response.Success = false
			response.Error = err.Error()
		} else {
			response.Success = true
			response.Metadata = map[string]interface{}{
				"health": "ok",
				"status": wp.orchestrator.GetStatus(),
			}
		}

	case RequestTypeMetrics:
		metrics := wp.orchestrator.GetMetrics()
		response.Success = true
		response.Metadata = map[string]interface{}{
			"metrics": metrics,
		}

	default:
		response.Success = false
		response.Error = fmt.Sprintf("unsupported request type: %s", request.Type)
	}

	wp.log.Debug("Worker %d completed request %s in %v", workerID, request.ID, response.Duration)
}