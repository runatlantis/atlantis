# Enhanced Locking System - Integration Examples

## Overview

This document provides practical code examples for integrating with the Enhanced Locking System, including API usage, custom implementations, monitoring integration, and advanced use cases.

## Table of Contents

1. [Basic API Integration](#basic-api-integration)
2. [Custom Backend Implementation](#custom-backend-implementation)
3. [Event System Integration](#event-system-integration)
4. [Monitoring and Metrics](#monitoring-and-metrics)
5. [Priority Queue Usage](#priority-queue-usage)
6. [Deadlock Detection Hooks](#deadlock-detection-hooks)
7. [Migration Helper Scripts](#migration-helper-scripts)
8. [Testing Utilities](#testing-utilities)
9. [Advanced Use Cases](#advanced-use-cases)

## Basic API Integration

### Lock Acquisition and Release

```go
package main

import (
    "context"
    "fmt"
    "time"

    "github.com/runatlantis/atlantis/server/enhanced"
    "github.com/runatlantis/atlantis/server/enhanced/locking"
)

func basicLockingExample() error {
    // Initialize enhanced locking system
    config := enhanced.DefaultConfig()
    config.Backend = "redis"
    config.EnablePriorityQueue = true

    lockManager, err := locking.NewEnhancedLockManager(config)
    if err != nil {
        return fmt.Errorf("failed to create lock manager: %w", err)
    }
    defer lockManager.Close()

    // Create lock request
    lockReq := &locking.LockRequest{
        Project:   "my-project",
        Workspace: "production",
        User:      "john.doe",
        Priority:  locking.PriorityHigh,
        Timeout:   30 * time.Second,
    }

    ctx := context.Background()

    // Acquire lock
    lock, err := lockManager.AcquireLock(ctx, lockReq)
    if err != nil {
        return fmt.Errorf("failed to acquire lock: %w", err)
    }

    fmt.Printf("Lock acquired: %s\n", lock.ID)

    // Perform work while holding lock
    defer func() {
        if err := lockManager.ReleaseLock(ctx, lock.ID); err != nil {
            fmt.Printf("Failed to release lock: %v\n", err)
        }
    }()

    // Your work here
    time.Sleep(5 * time.Second)

    return nil
}
```

### Async Lock Operations

```go
func asyncLockingExample() error {
    lockManager, err := locking.NewEnhancedLockManager(enhanced.DefaultConfig())
    if err != nil {
        return err
    }
    defer lockManager.Close()

    ctx := context.Background()

    // Submit lock request to queue
    lockReq := &locking.LockRequest{
        Project:   "my-project",
        Workspace: "staging",
        User:      "ci-system",
        Priority:  locking.PriorityMedium,
        Async:     true,
    }

    // Get future for async operation
    future, err := lockManager.AcquireLockAsync(ctx, lockReq)
    if err != nil {
        return err
    }

    // Wait for lock with timeout
    select {
    case lock := <-future.Result():
        fmt.Printf("Async lock acquired: %s\n", lock.ID)
        defer lockManager.ReleaseLock(ctx, lock.ID)

        // Your work here

    case err := <-future.Error():
        return fmt.Errorf("async lock failed: %w", err)

    case <-time.After(60 * time.Second):
        future.Cancel()
        return fmt.Errorf("async lock timed out")
    }

    return nil
}
```

### Bulk Lock Operations

```go
func bulkLockingExample() error {
    lockManager, err := locking.NewEnhancedLockManager(enhanced.DefaultConfig())
    if err != nil {
        return err
    }
    defer lockManager.Close()

    ctx := context.Background()

    // Prepare multiple lock requests
    requests := []*locking.LockRequest{
        {Project: "project-a", Workspace: "production", User: "deploy-bot", Priority: locking.PriorityHigh},
        {Project: "project-b", Workspace: "production", User: "deploy-bot", Priority: locking.PriorityHigh},
        {Project: "project-c", Workspace: "production", User: "deploy-bot", Priority: locking.PriorityHigh},
    }

    // Acquire all locks atomically
    locks, err := lockManager.AcquireMultipleLocks(ctx, requests)
    if err != nil {
        return fmt.Errorf("failed to acquire multiple locks: %w", err)
    }

    fmt.Printf("Acquired %d locks\n", len(locks))

    // Release all locks
    defer func() {
        lockIDs := make([]string, len(locks))
        for i, lock := range locks {
            lockIDs[i] = lock.ID
        }

        if err := lockManager.ReleaseMultipleLocks(ctx, lockIDs); err != nil {
            fmt.Printf("Failed to release locks: %v\n", err)
        }
    }()

    // Perform coordinated work
    return nil
}
```

## Custom Backend Implementation

### Implementing a Custom Backend

```go
package custombackend

import (
    "context"
    "fmt"
    "sync"

    "github.com/runatlantis/atlantis/server/enhanced/locking"
)

// CustomBackend implements the Backend interface
type CustomBackend struct {
    locks map[string]*locking.Lock
    mutex sync.RWMutex

    // Your custom storage (database, etc.)
    storage CustomStorage
}

func NewCustomBackend(config *CustomConfig) *CustomBackend {
    return &CustomBackend{
        locks:   make(map[string]*locking.Lock),
        storage: NewCustomStorage(config),
    }
}

// AcquireLock implements Backend.AcquireLock
func (cb *CustomBackend) AcquireLock(ctx context.Context, req *locking.LockRequest) (*locking.Lock, error) {
    cb.mutex.Lock()
    defer cb.mutex.Unlock()

    // Check if lock already exists
    lockKey := fmt.Sprintf("%s:%s", req.Project, req.Workspace)
    if existingLock, exists := cb.locks[lockKey]; exists {
        return nil, &locking.LockConflictError{
            ExistingLock: existingLock,
            Request:      req,
        }
    }

    // Create new lock
    lock := &locking.Lock{
        ID:          generateLockID(),
        Project:     req.Project,
        Workspace:   req.Workspace,
        User:        req.User,
        AcquiredAt:  time.Now(),
        Priority:    req.Priority,
    }

    // Persist to custom storage
    if err := cb.storage.SaveLock(ctx, lock); err != nil {
        return nil, fmt.Errorf("failed to persist lock: %w", err)
    }

    // Store in memory
    cb.locks[lockKey] = lock

    return lock, nil
}

// ReleaseLock implements Backend.ReleaseLock
func (cb *CustomBackend) ReleaseLock(ctx context.Context, lockID string) error {
    cb.mutex.Lock()
    defer cb.mutex.Unlock()

    // Find and remove lock
    var lockKey string
    var lock *locking.Lock

    for key, l := range cb.locks {
        if l.ID == lockID {
            lockKey = key
            lock = l
            break
        }
    }

    if lock == nil {
        return &locking.LockNotFoundError{LockID: lockID}
    }

    // Remove from storage
    if err := cb.storage.DeleteLock(ctx, lockID); err != nil {
        return fmt.Errorf("failed to delete lock from storage: %w", err)
    }

    // Remove from memory
    delete(cb.locks, lockKey)

    return nil
}

// ListLocks implements Backend.ListLocks
func (cb *CustomBackend) ListLocks(ctx context.Context) ([]*locking.Lock, error) {
    cb.mutex.RLock()
    defer cb.mutex.RUnlock()

    locks := make([]*locking.Lock, 0, len(cb.locks))
    for _, lock := range cb.locks {
        locks = append(locks, lock)
    }

    return locks, nil
}

// Health implements Backend.Health
func (cb *CustomBackend) Health(ctx context.Context) error {
    return cb.storage.Ping(ctx)
}
```

### Backend Registration

```go
func init() {
    // Register custom backend
    locking.RegisterBackend("custom", func(config map[string]interface{}) (locking.Backend, error) {
        customConfig, err := parseCustomConfig(config)
        if err != nil {
            return nil, err
        }

        return NewCustomBackend(customConfig), nil
    })
}
```

## Event System Integration

### Event Listener Implementation

```go
package events

import (
    "context"
    "fmt"
    "log"

    "github.com/runatlantis/atlantis/server/enhanced/events"
)

// CustomEventListener handles enhanced locking events
type CustomEventListener struct {
    notifier NotificationService
    metrics  MetricsCollector
}

func NewCustomEventListener(notifier NotificationService, metrics MetricsCollector) *CustomEventListener {
    return &CustomEventListener{
        notifier: notifier,
        metrics:  metrics,
    }
}

// HandleEvent processes incoming events
func (cel *CustomEventListener) HandleEvent(ctx context.Context, event *events.Event) error {
    switch event.Type {
    case events.LockAcquired:
        return cel.handleLockAcquired(ctx, event)
    case events.LockReleased:
        return cel.handleLockReleased(ctx, event)
    case events.LockTimeout:
        return cel.handleLockTimeout(ctx, event)
    case events.DeadlockDetected:
        return cel.handleDeadlockDetected(ctx, event)
    case events.QueueOverflow:
        return cel.handleQueueOverflow(ctx, event)
    default:
        log.Printf("Unknown event type: %s", event.Type)
    }

    return nil
}

func (cel *CustomEventListener) handleLockAcquired(ctx context.Context, event *events.Event) error {
    lockEvent := event.Data.(*events.LockAcquiredEvent)

    // Send notification
    message := fmt.Sprintf("Lock acquired: %s by %s", lockEvent.Lock.ID, lockEvent.Lock.User)
    if err := cel.notifier.SendSlackMessage(message); err != nil {
        log.Printf("Failed to send notification: %v", err)
    }

    // Record metrics
    cel.metrics.IncrementCounter("locks_acquired_total", map[string]string{
        "project":   lockEvent.Lock.Project,
        "workspace": lockEvent.Lock.Workspace,
        "user":      lockEvent.Lock.User,
    })

    return nil
}

func (cel *CustomEventListener) handleDeadlockDetected(ctx context.Context, event *events.Event) error {
    deadlockEvent := event.Data.(*events.DeadlockDetectedEvent)

    // Send urgent notification
    message := fmt.Sprintf("🚨 DEADLOCK DETECTED: %s (participants: %d)",
        deadlockEvent.DeadlockID, len(deadlockEvent.Participants))
    if err := cel.notifier.SendUrgentAlert(message); err != nil {
        log.Printf("Failed to send urgent alert: %v", err)
    }

    // Record critical metric
    cel.metrics.IncrementCounter("deadlocks_detected_total", map[string]string{
        "resolution_policy": deadlockEvent.ResolutionPolicy,
    })

    return nil
}
```

### Event Stream Consumer

```go
func startEventConsumer() error {
    // Create event manager
    eventManager := events.NewEventManager(events.DefaultConfig())

    // Create custom listener
    listener := NewCustomEventListener(notificationService, metricsCollector)

    // Subscribe to all lock events
    subscription := &events.Subscription{
        Name: "custom-integration",
        Filter: events.FilterBuilder().
            WithEventTypes(events.LockAcquired, events.LockReleased, events.DeadlockDetected).
            WithProjects("critical-project").
            Build(),
        Handler: listener.HandleEvent,
    }

    if err := eventManager.Subscribe(subscription); err != nil {
        return fmt.Errorf("failed to subscribe to events: %w", err)
    }

    // Start consuming events
    ctx := context.Background()
    return eventManager.Start(ctx)
}
```

## Monitoring and Metrics

### Prometheus Integration

```go
package monitoring

import (
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
)

var (
    lockDuration = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "atlantis_enhanced_lock_duration_seconds",
            Help:    "Time spent holding locks",
            Buckets: prometheus.DefBuckets,
        },
        []string{"project", "workspace", "user"},
    )

    queueDepth = promauto.NewGaugeVec(
        prometheus.GaugeOpts{
            Name: "atlantis_enhanced_queue_depth",
            Help: "Current queue depth",
        },
        []string{"priority"},
    )

    deadlockResolutions = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "atlantis_enhanced_deadlocks_resolved_total",
            Help: "Total number of deadlocks resolved",
        },
        []string{"policy", "success"},
    )
)

// MetricsCollector collects enhanced locking metrics
type MetricsCollector struct {
    manager *locking.EnhancedLockManager
}

func NewMetricsCollector(manager *locking.EnhancedLockManager) *MetricsCollector {
    return &MetricsCollector{manager: manager}
}

func (mc *MetricsCollector) CollectMetrics() {
    // Collect queue metrics
    queueStats := mc.manager.GetQueueStats()
    for priority, depth := range queueStats.DepthByPriority {
        queueDepth.WithLabelValues(priority.String()).Set(float64(depth))
    }

    // Collect lock metrics
    lockStats := mc.manager.GetLockStats()
    for _, stat := range lockStats.ActiveLocks {
        duration := time.Since(stat.AcquiredAt).Seconds()
        lockDuration.WithLabelValues(stat.Project, stat.Workspace, stat.User).Observe(duration)
    }
}

func (mc *MetricsCollector) Start(ctx context.Context) {
    ticker := time.NewTicker(30 * time.Second)
    defer ticker.Stop()

    for {
        select {
        case <-ticker.C:
            mc.CollectMetrics()
        case <-ctx.Done():
            return
        }
    }
}
```

### Custom Health Checks

```go
package health

import (
    "context"
    "fmt"
    "net/http"
    "time"

    "github.com/runatlantis/atlantis/server/enhanced/locking"
)

type HealthChecker struct {
    lockManager *locking.EnhancedLockManager
    thresholds  *HealthThresholds
}

type HealthThresholds struct {
    MaxQueueDepth     int
    MaxErrorRate      float64
    MaxAverageLatency time.Duration
}

func (hc *HealthChecker) CheckHealth(ctx context.Context) (*HealthReport, error) {
    report := &HealthReport{
        Timestamp: time.Now(),
        Status:    "healthy",
        Checks:    make(map[string]*HealthCheck),
    }

    // Check backend connectivity
    if err := hc.checkBackend(ctx, report); err != nil {
        report.Status = "unhealthy"
    }

    // Check queue health
    if err := hc.checkQueue(ctx, report); err != nil {
        report.Status = "degraded"
    }

    // Check performance metrics
    if err := hc.checkPerformance(ctx, report); err != nil {
        report.Status = "degraded"
    }

    return report, nil
}

func (hc *HealthChecker) checkBackend(ctx context.Context, report *HealthReport) error {
    start := time.Now()
    err := hc.lockManager.HealthCheck(ctx)
    duration := time.Since(start)

    check := &HealthCheck{
        Name:     "backend",
        Status:   "pass",
        Duration: duration,
    }

    if err != nil {
        check.Status = "fail"
        check.Output = err.Error()
        return err
    }

    report.Checks["backend"] = check
    return nil
}

func (hc *HealthChecker) checkQueue(ctx context.Context, report *HealthReport) error {
    stats := hc.lockManager.GetQueueStats()

    check := &HealthCheck{
        Name:   "queue",
        Status: "pass",
        Output: fmt.Sprintf("Depth: %d", stats.TotalDepth),
    }

    if stats.TotalDepth > hc.thresholds.MaxQueueDepth {
        check.Status = "warn"
        check.Output += fmt.Sprintf(" (exceeds threshold: %d)", hc.thresholds.MaxQueueDepth)
        return fmt.Errorf("queue depth exceeds threshold")
    }

    report.Checks["queue"] = check
    return nil
}

// HTTP handler for health endpoint
func (hc *HealthChecker) HealthHandler(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    report, err := hc.CheckHealth(ctx)

    statusCode := http.StatusOK
    if report.Status == "unhealthy" {
        statusCode = http.StatusServiceUnavailable
    } else if report.Status == "degraded" {
        statusCode = http.StatusPartialContent
    }

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(statusCode)

    if err := json.NewEncoder(w).Encode(report); err != nil {
        http.Error(w, "Failed to encode health report", http.StatusInternalServerError)
    }
}
```

## Priority Queue Usage

### Custom Priority Implementation

```go
package priority

import (
    "time"

    "github.com/runatlantis/atlantis/server/enhanced/locking"
)

// CustomPriorityCalculator implements dynamic priority calculation
type CustomPriorityCalculator struct {
    weights map[string]float64
}

func NewCustomPriorityCalculator() *CustomPriorityCalculator {
    return &CustomPriorityCalculator{
        weights: map[string]float64{
            "production": 1.0,
            "staging":    0.7,
            "development": 0.3,
        },
    }
}

// CalculatePriority determines priority based on multiple factors
func (cpc *CustomPriorityCalculator) CalculatePriority(req *locking.LockRequest) locking.Priority {
    score := 0.0

    // Base priority from request
    score += float64(req.Priority) * 10

    // Workspace weight
    if weight, exists := cpc.weights[req.Workspace]; exists {
        score += weight * 5
    }

    // Time-based urgency (longer waiting = higher priority)
    waitTime := time.Since(req.CreatedAt)
    score += waitTime.Minutes() * 0.1

    // User-based priority (emergency users get boost)
    if cpc.isEmergencyUser(req.User) {
        score += 20
    }

    // Convert score to priority level
    switch {
    case score >= 50:
        return locking.PriorityCritical
    case score >= 30:
        return locking.PriorityHigh
    case score >= 15:
        return locking.PriorityMedium
    case score >= 5:
        return locking.PriorityLow
    default:
        return locking.PriorityVeryLow
    }
}

func (cpc *CustomPriorityCalculator) isEmergencyUser(user string) bool {
    emergencyUsers := []string{"incident-commander", "security-team", "cto"}
    for _, emergencyUser := range emergencyUsers {
        if user == emergencyUser {
            return true
        }
    }
    return false
}
```

### Queue Management

```go
func advancedQueueManagement() error {
    config := enhanced.DefaultConfig()
    config.EnablePriorityQueue = true

    // Configure custom priority calculator
    priorityCalculator := NewCustomPriorityCalculator()
    config.PriorityCalculator = priorityCalculator

    lockManager, err := locking.NewEnhancedLockManager(config)
    if err != nil {
        return err
    }
    defer lockManager.Close()

    // Get queue manager
    queueManager := lockManager.GetQueueManager()

    // Monitor queue health
    go func() {
        ticker := time.NewTicker(10 * time.Second)
        defer ticker.Stop()

        for range ticker.C {
            stats := queueManager.GetStats()

            fmt.Printf("Queue Stats:\n")
            fmt.Printf("  Total Depth: %d\n", stats.TotalDepth)
            fmt.Printf("  Processing Rate: %.2f/sec\n", stats.ProcessingRate)
            fmt.Printf("  Average Wait Time: %v\n", stats.AverageWaitTime)

            // Check for queue health issues
            if stats.TotalDepth > 100 {
                fmt.Printf("⚠️  Queue depth high: %d\n", stats.TotalDepth)

                // Take corrective action
                queueManager.OptimizeProcessing()
            }

            if stats.AverageWaitTime > 60*time.Second {
                fmt.Printf("⚠️  High wait times: %v\n", stats.AverageWaitTime)

                // Increase processing capacity
                queueManager.ScaleWorkers(8)
            }
        }
    }()

    return nil
}
```

## Deadlock Detection Hooks

### Custom Resolution Hook

```go
package deadlock

import (
    "context"
    "fmt"
    "time"

    "github.com/runatlantis/atlantis/server/enhanced/deadlock"
)

// CustomResolutionHook provides custom deadlock handling
type CustomResolutionHook struct {
    notifier   NotificationService
    approver   ApprovalService
    analytics  AnalyticsService
}

func NewCustomResolutionHook(notifier NotificationService, approver ApprovalService, analytics AnalyticsService) *CustomResolutionHook {
    return &CustomResolutionHook{
        notifier:  notifier,
        approver:  approver,
        analytics: analytics,
    }
}

// BeforeDetection is called before deadlock detection runs
func (crh *CustomResolutionHook) BeforeDetection(ctx context.Context) error {
    // Log detection cycle start
    crh.analytics.RecordEvent("deadlock_detection_started", map[string]interface{}{
        "timestamp": time.Now(),
    })

    return nil
}

// OnDeadlockDetected is called when a deadlock is found
func (crh *CustomResolutionHook) OnDeadlockDetected(ctx context.Context, deadlock *deadlock.Deadlock) error {
    // Send immediate notification
    message := fmt.Sprintf("🔒 Deadlock detected: %s\nParticipants: %d\nCycle: %s",
        deadlock.ID, len(deadlock.Participants), deadlock.CycleDescription())

    if err := crh.notifier.SendUrgentAlert(message); err != nil {
        return fmt.Errorf("failed to send deadlock notification: %w", err)
    }

    // Record analytics
    crh.analytics.RecordEvent("deadlock_detected", map[string]interface{}{
        "deadlock_id":     deadlock.ID,
        "participants":    len(deadlock.Participants),
        "cycle_length":    deadlock.CycleLength(),
        "detection_time":  deadlock.DetectedAt,
    })

    return nil
}

// BeforeResolution is called before attempting resolution
func (crh *CustomResolutionHook) BeforeResolution(ctx context.Context, deadlock *deadlock.Deadlock, victim string) error {
    // For high-impact deadlocks, require manual approval
    if crh.isHighImpact(deadlock) {
        message := fmt.Sprintf("High-impact deadlock resolution requires approval.\nDeadlock: %s\nProposed victim: %s",
            deadlock.ID, victim)

        if err := crh.notifier.RequestApproval(message); err != nil {
            return fmt.Errorf("failed to request approval: %w", err)
        }

        // Wait for approval with timeout
        approved, err := crh.approver.WaitForApproval(ctx, deadlock.ID, 5*time.Minute)
        if err != nil {
            return fmt.Errorf("approval failed: %w", err)
        }

        if !approved {
            return fmt.Errorf("resolution not approved")
        }
    }

    return nil
}

// AfterResolution is called after resolution attempt
func (crh *CustomResolutionHook) AfterResolution(ctx context.Context, deadlock *deadlock.Deadlock, success bool, victim string) {
    status := "success"
    if !success {
        status = "failed"
    }

    // Send resolution notification
    message := fmt.Sprintf("Deadlock resolution %s\nDeadlock: %s\nVictim: %s",
        status, deadlock.ID, victim)

    if success {
        crh.notifier.SendInfoMessage(message)
    } else {
        crh.notifier.SendErrorAlert(message)
    }

    // Record resolution outcome
    crh.analytics.RecordEvent("deadlock_resolution", map[string]interface{}{
        "deadlock_id":     deadlock.ID,
        "success":         success,
        "victim":          victim,
        "resolution_time": time.Now(),
        "policy_used":     deadlock.ResolutionPolicy,
    })
}

func (crh *CustomResolutionHook) isHighImpact(deadlock *deadlock.Deadlock) bool {
    // Consider high impact if involves production workspaces
    for _, participant := range deadlock.Participants {
        if participant.Workspace == "production" {
            return true
        }
    }

    // Or if cycle involves many participants
    return len(deadlock.Participants) > 5
}
```

### Deadlock Prevention

```go
package prevention

import (
    "context"
    "time"

    "github.com/runatlantis/atlantis/server/enhanced/deadlock"
    "github.com/runatlantis/atlantis/server/enhanced/locking"
)

// DeadlockPrevention implements proactive deadlock avoidance
type DeadlockPrevention struct {
    detector    *deadlock.Detector
    waitGraph   *deadlock.WaitForGraph
    predictions map[string]*DeadlockPrediction
}

type DeadlockPrediction struct {
    Probability float64
    TimeToDeadlock time.Duration
    SuggestedAction string
}

func NewDeadlockPrevention(detector *deadlock.Detector) *DeadlockPrevention {
    return &DeadlockPrevention{
        detector:    detector,
        waitGraph:   detector.GetWaitGraph(),
        predictions: make(map[string]*DeadlockPrediction),
    }
}

// PredictDeadlock analyzes current wait graph for potential deadlocks
func (dp *DeadlockPrevention) PredictDeadlock(ctx context.Context, newRequest *locking.LockRequest) (*DeadlockPrediction, error) {
    // Create hypothetical wait graph with new request
    hypotheticalGraph := dp.waitGraph.Clone()
    hypotheticalGraph.AddEdge(newRequest.User, newRequest.Project+":"+newRequest.Workspace)

    // Analyze for potential cycles
    cycles := hypotheticalGraph.FindPotentialCycles()

    if len(cycles) == 0 {
        return &DeadlockPrediction{
            Probability: 0.0,
            SuggestedAction: "proceed",
        }, nil
    }

    // Calculate deadlock probability based on historical data
    probability := dp.calculateDeadlockProbability(cycles, newRequest)

    prediction := &DeadlockPrediction{
        Probability: probability,
        TimeToDeadlock: dp.estimateTimeToDeadlock(cycles),
    }

    // Suggest action based on probability
    switch {
    case probability > 0.8:
        prediction.SuggestedAction = "deny_request"
    case probability > 0.5:
        prediction.SuggestedAction = "delay_request"
    case probability > 0.2:
        prediction.SuggestedAction = "monitor_closely"
    default:
        prediction.SuggestedAction = "proceed"
    }

    return prediction, nil
}

func (dp *DeadlockPrevention) calculateDeadlockProbability(cycles [][]string, request *locking.LockRequest) float64 {
    // Base probability on cycle characteristics
    baseProbability := 0.1

    // Increase probability for each cycle found
    baseProbability += float64(len(cycles)) * 0.2

    // Historical factors
    if dp.hasRecentDeadlocks(request.Project) {
        baseProbability += 0.3
    }

    // Time-based factors (more likely during busy periods)
    if dp.isBusyPeriod() {
        baseProbability += 0.2
    }

    // Cap at 1.0
    if baseProbability > 1.0 {
        baseProbability = 1.0
    }

    return baseProbability
}

func (dp *DeadlockPrevention) estimateTimeToDeadlock(cycles [][]string) time.Duration {
    // Estimate based on average lock hold times and cycle complexity
    avgHoldTime := 30 * time.Second // from historical data
    longestCycle := 0

    for _, cycle := range cycles {
        if len(cycle) > longestCycle {
            longestCycle = len(cycle)
        }
    }

    // Longer cycles take more time to form
    return time.Duration(longestCycle) * avgHoldTime / 2
}
```

## Migration Helper Scripts

### Data Migration Script

```go
package migration

import (
    "context"
    "fmt"
    "log"
    "time"

    "github.com/runatlantis/atlantis/server/enhanced/locking"
)

// MigrationManager handles data migration between backends
type MigrationManager struct {
    sourceBackend locking.Backend
    targetBackend locking.Backend
    config        *MigrationConfig
}

type MigrationConfig struct {
    BatchSize       int
    VerifyData      bool
    DryRun          bool
    ContinueOnError bool
    ProgressCallback func(progress *MigrationProgress)
}

type MigrationProgress struct {
    TotalLocks     int
    MigratedLocks  int
    FailedLocks    int
    CurrentLock    string
    StartTime      time.Time
    ElapsedTime    time.Duration
}

func NewMigrationManager(source, target locking.Backend, config *MigrationConfig) *MigrationManager {
    return &MigrationManager{
        sourceBackend: source,
        targetBackend: target,
        config:        config,
    }
}

func (mm *MigrationManager) MigrateData(ctx context.Context) error {
    // Get all locks from source
    sourceLocks, err := mm.sourceBackend.ListLocks(ctx)
    if err != nil {
        return fmt.Errorf("failed to list source locks: %w", err)
    }

    progress := &MigrationProgress{
        TotalLocks: len(sourceLocks),
        StartTime:  time.Now(),
    }

    log.Printf("Starting migration of %d locks", progress.TotalLocks)

    // Migrate in batches
    for i := 0; i < len(sourceLocks); i += mm.config.BatchSize {
        end := i + mm.config.BatchSize
        if end > len(sourceLocks) {
            end = len(sourceLocks)
        }

        batch := sourceLocks[i:end]
        if err := mm.migrateBatch(ctx, batch, progress); err != nil && !mm.config.ContinueOnError {
            return fmt.Errorf("migration failed at batch %d-%d: %w", i, end-1, err)
        }

        // Update progress
        progress.ElapsedTime = time.Since(progress.StartTime)
        if mm.config.ProgressCallback != nil {
            mm.config.ProgressCallback(progress)
        }
    }

    log.Printf("Migration completed: %d/%d locks migrated, %d failed",
        progress.MigratedLocks, progress.TotalLocks, progress.FailedLocks)

    return nil
}

func (mm *MigrationManager) migrateBatch(ctx context.Context, locks []*locking.Lock, progress *MigrationProgress) error {
    for _, lock := range locks {
        progress.CurrentLock = lock.ID

        if mm.config.DryRun {
            log.Printf("DRY RUN: Would migrate lock %s", lock.ID)
            progress.MigratedLocks++
            continue
        }

        // Migrate lock to target backend
        if err := mm.migrateLock(ctx, lock); err != nil {
            log.Printf("Failed to migrate lock %s: %v", lock.ID, err)
            progress.FailedLocks++
            continue
        }

        progress.MigratedLocks++
    }

    return nil
}

func (mm *MigrationManager) migrateLock(ctx context.Context, lock *locking.Lock) error {
    // Create lock request for target backend
    req := &locking.LockRequest{
        Project:   lock.Project,
        Workspace: lock.Workspace,
        User:      lock.User,
        Priority:  lock.Priority,
    }

    // Acquire lock in target backend
    targetLock, err := mm.targetBackend.AcquireLock(ctx, req)
    if err != nil {
        return fmt.Errorf("failed to acquire lock in target: %w", err)
    }

    // Update target lock with source metadata
    targetLock.AcquiredAt = lock.AcquiredAt
    targetLock.Metadata = lock.Metadata

    // Verify migration if enabled
    if mm.config.VerifyData {
        if err := mm.verifyLock(ctx, lock, targetLock); err != nil {
            // Cleanup target lock on verification failure
            mm.targetBackend.ReleaseLock(ctx, targetLock.ID)
            return fmt.Errorf("verification failed: %w", err)
        }
    }

    return nil
}

func (mm *MigrationManager) verifyLock(ctx context.Context, source, target *locking.Lock) error {
    if source.Project != target.Project {
        return fmt.Errorf("project mismatch: %s != %s", source.Project, target.Project)
    }

    if source.Workspace != target.Workspace {
        return fmt.Errorf("workspace mismatch: %s != %s", source.Workspace, target.Workspace)
    }

    if source.User != target.User {
        return fmt.Errorf("user mismatch: %s != %s", source.User, target.User)
    }

    return nil
}
```

### Migration Status Monitor

```go
func runMigrationWithMonitoring() error {
    // Setup backends
    sourceBackend := boltdb.NewBoltDBBackend("/var/lib/atlantis/atlantis.db")
    targetBackend := redis.NewRedisBackend(redisConfig)

    // Configure migration
    config := &MigrationConfig{
        BatchSize:       10,
        VerifyData:      true,
        DryRun:          false,
        ContinueOnError: true,
        ProgressCallback: func(progress *MigrationProgress) {
            percentage := float64(progress.MigratedLocks) / float64(progress.TotalLocks) * 100

            fmt.Printf("Migration Progress: %.1f%% (%d/%d locks, %d failed)\n",
                percentage, progress.MigratedLocks, progress.TotalLocks, progress.FailedLocks)

            if progress.ElapsedTime > 0 {
                rate := float64(progress.MigratedLocks) / progress.ElapsedTime.Seconds()
                remaining := time.Duration(float64(progress.TotalLocks-progress.MigratedLocks) / rate * float64(time.Second))
                fmt.Printf("Rate: %.2f locks/sec, ETA: %v\n", rate, remaining)
            }
        },
    }

    // Run migration
    migrationManager := NewMigrationManager(sourceBackend, targetBackend, config)

    ctx := context.Background()
    return migrationManager.MigrateData(ctx)
}
```

## Testing Utilities

### Load Testing Framework

```go
package testing

import (
    "context"
    "fmt"
    "sync"
    "time"

    "github.com/runatlantis/atlantis/server/enhanced/locking"
)

// LoadTestRunner executes load tests against the enhanced locking system
type LoadTestRunner struct {
    lockManager *locking.EnhancedLockManager
    config      *LoadTestConfig
    results     *LoadTestResults
}

type LoadTestConfig struct {
    ConcurrentUsers int
    TestDuration    time.Duration
    OperationsPerUser int
    LockHoldTime    time.Duration
    Projects        []string
    Workspaces      []string
}

type LoadTestResults struct {
    TotalOperations   int
    SuccessfulLocks   int
    FailedLocks       int
    Timeouts          int
    Deadlocks         int
    AverageLatency    time.Duration
    ThroughputPerSec  float64
    ErrorRate         float64
}

func NewLoadTestRunner(lockManager *locking.EnhancedLockManager, config *LoadTestConfig) *LoadTestRunner {
    return &LoadTestRunner{
        lockManager: lockManager,
        config:      config,
        results:     &LoadTestResults{},
    }
}

func (ltr *LoadTestRunner) RunLoadTest(ctx context.Context) (*LoadTestResults, error) {
    fmt.Printf("Starting load test: %d users, %d ops/user, %v duration\n",
        ltr.config.ConcurrentUsers, ltr.config.OperationsPerUser, ltr.config.TestDuration)

    var wg sync.WaitGroup
    resultsChan := make(chan *UserTestResult, ltr.config.ConcurrentUsers)

    startTime := time.Now()

    // Start concurrent users
    for i := 0; i < ltr.config.ConcurrentUsers; i++ {
        wg.Add(1)
        go func(userID int) {
            defer wg.Done()
            result := ltr.runUserTest(ctx, userID)
            resultsChan <- result
        }(i)
    }

    // Wait for all users to complete
    go func() {
        wg.Wait()
        close(resultsChan)
    }()

    // Collect results
    ltr.collectResults(resultsChan)

    // Calculate final metrics
    totalDuration := time.Since(startTime)
    ltr.results.ThroughputPerSec = float64(ltr.results.TotalOperations) / totalDuration.Seconds()
    ltr.results.ErrorRate = float64(ltr.results.FailedLocks) / float64(ltr.results.TotalOperations)

    return ltr.results, nil
}

func (ltr *LoadTestRunner) runUserTest(ctx context.Context, userID int) *UserTestResult {
    result := &UserTestResult{UserID: userID}

    for op := 0; op < ltr.config.OperationsPerUser; op++ {
        // Random project and workspace
        project := ltr.config.Projects[op%len(ltr.config.Projects)]
        workspace := ltr.config.Workspaces[op%len(ltr.config.Workspaces)]

        opResult := ltr.performLockOperation(ctx, userID, project, workspace)
        result.Operations = append(result.Operations, opResult)

        // Small delay between operations
        time.Sleep(100 * time.Millisecond)
    }

    return result
}

func (ltr *LoadTestRunner) performLockOperation(ctx context.Context, userID int, project, workspace string) *OperationResult {
    start := time.Now()

    req := &locking.LockRequest{
        Project:   project,
        Workspace: workspace,
        User:      fmt.Sprintf("loadtest-user-%d", userID),
        Priority:  locking.PriorityMedium,
        Timeout:   30 * time.Second,
    }

    // Acquire lock
    lock, err := ltr.lockManager.AcquireLock(ctx, req)
    if err != nil {
        return &OperationResult{
            Success:   false,
            Error:     err.Error(),
            Duration:  time.Since(start),
        }
    }

    // Hold lock for configured time
    time.Sleep(ltr.config.LockHoldTime)

    // Release lock
    releaseErr := ltr.lockManager.ReleaseLock(ctx, lock.ID)

    return &OperationResult{
        Success:  releaseErr == nil,
        LockID:   lock.ID,
        Duration: time.Since(start),
        Error:    func() string {
            if releaseErr != nil {
                return releaseErr.Error()
            }
            return ""
        }(),
    }
}
```

### Integration Test Suite

```go
package integration

import (
    "context"
    "testing"
    "time"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"

    "github.com/runatlantis/atlantis/server/enhanced"
    "github.com/runatlantis/atlantis/server/enhanced/locking"
)

func TestEnhancedLockingIntegration(t *testing.T) {
    // Setup test environment
    config := enhanced.DefaultConfig()
    config.Backend = "memory"  // Use in-memory backend for tests

    lockManager, err := locking.NewEnhancedLockManager(config)
    require.NoError(t, err)
    defer lockManager.Close()

    ctx := context.Background()

    t.Run("BasicLockAcquisition", func(t *testing.T) {
        testBasicLockAcquisition(t, ctx, lockManager)
    })

    t.Run("ConcurrentLocking", func(t *testing.T) {
        testConcurrentLocking(t, ctx, lockManager)
    })

    t.Run("PriorityQueuing", func(t *testing.T) {
        testPriorityQueuing(t, ctx, lockManager)
    })

    t.Run("DeadlockDetection", func(t *testing.T) {
        testDeadlockDetection(t, ctx, lockManager)
    })
}

func testBasicLockAcquisition(t *testing.T, ctx context.Context, lockManager *locking.EnhancedLockManager) {
    req := &locking.LockRequest{
        Project:   "test-project",
        Workspace: "test-workspace",
        User:      "test-user",
        Priority:  locking.PriorityMedium,
    }

    // Acquire lock
    lock, err := lockManager.AcquireLock(ctx, req)
    require.NoError(t, err)
    assert.NotEmpty(t, lock.ID)
    assert.Equal(t, req.Project, lock.Project)
    assert.Equal(t, req.Workspace, lock.Workspace)
    assert.Equal(t, req.User, lock.User)

    // Verify lock is active
    locks, err := lockManager.ListLocks(ctx)
    require.NoError(t, err)
    assert.Len(t, locks, 1)
    assert.Equal(t, lock.ID, locks[0].ID)

    // Release lock
    err = lockManager.ReleaseLock(ctx, lock.ID)
    require.NoError(t, err)

    // Verify lock is released
    locks, err = lockManager.ListLocks(ctx)
    require.NoError(t, err)
    assert.Len(t, locks, 0)
}

func testConcurrentLocking(t *testing.T, ctx context.Context, lockManager *locking.EnhancedLockManager) {
    const numGoroutines = 10
    results := make(chan error, numGoroutines)

    // Start concurrent lock attempts
    for i := 0; i < numGoroutines; i++ {
        go func(id int) {
            req := &locking.LockRequest{
                Project:   "concurrent-project",
                Workspace: "concurrent-workspace",
                User:      fmt.Sprintf("user-%d", id),
                Priority:  locking.PriorityMedium,
                Timeout:   5 * time.Second,
            }

            lock, err := lockManager.AcquireLock(ctx, req)
            if err != nil {
                results <- err
                return
            }

            // Hold lock briefly
            time.Sleep(100 * time.Millisecond)

            err = lockManager.ReleaseLock(ctx, lock.ID)
            results <- err
        }(i)
    }

    // Collect results
    successCount := 0
    for i := 0; i < numGoroutines; i++ {
        err := <-results
        if err == nil {
            successCount++
        }
    }

    // At least some should succeed (depends on timing and queue)
    assert.Greater(t, successCount, 0)
}
```

## Advanced Use Cases

### Multi-Tenant Lock Management

```go
package multitenant

import (
    "context"
    "fmt"

    "github.com/runatlantis/atlantis/server/enhanced/locking"
)

// TenantAwareLockManager provides tenant isolation for locking
type TenantAwareLockManager struct {
    baseLockManager *locking.EnhancedLockManager
    tenantIsolation bool
}

func NewTenantAwareLockManager(baseLockManager *locking.EnhancedLockManager, tenantIsolation bool) *TenantAwareLockManager {
    return &TenantAwareLockManager{
        baseLockManager: baseLockManager,
        tenantIsolation: tenantIsolation,
    }
}

func (tlm *TenantAwareLockManager) AcquireLock(ctx context.Context, tenantID string, req *locking.LockRequest) (*locking.Lock, error) {
    // Add tenant prefix to isolate resources
    if tlm.tenantIsolation {
        req.Project = fmt.Sprintf("%s:%s", tenantID, req.Project)
    }

    // Add tenant metadata
    if req.Metadata == nil {
        req.Metadata = make(map[string]string)
    }
    req.Metadata["tenant_id"] = tenantID

    // Check tenant quotas
    if err := tlm.checkTenantQuota(ctx, tenantID); err != nil {
        return nil, fmt.Errorf("tenant quota exceeded: %w", err)
    }

    return tlm.baseLockManager.AcquireLock(ctx, req)
}

func (tlm *TenantAwareLockManager) ListTenantLocks(ctx context.Context, tenantID string) ([]*locking.Lock, error) {
    allLocks, err := tlm.baseLockManager.ListLocks(ctx)
    if err != nil {
        return nil, err
    }

    var tenantLocks []*locking.Lock
    for _, lock := range allLocks {
        if lock.Metadata["tenant_id"] == tenantID {
            tenantLocks = append(tenantLocks, lock)
        }
    }

    return tenantLocks, nil
}

func (tlm *TenantAwareLockManager) checkTenantQuota(ctx context.Context, tenantID string) error {
    currentLocks, err := tlm.ListTenantLocks(ctx, tenantID)
    if err != nil {
        return err
    }

    maxLocks := tlm.getTenantQuota(tenantID)
    if len(currentLocks) >= maxLocks {
        return fmt.Errorf("tenant %s has reached maximum locks quota: %d", tenantID, maxLocks)
    }

    return nil
}

func (tlm *TenantAwareLockManager) getTenantQuota(tenantID string) int {
    // Implementation would check tenant configuration
    // For example, from database or configuration service
    defaultQuota := 10
    return defaultQuota
}
```

### Workflow Integration

```go
package workflow

import (
    "context"
    "fmt"
    "time"

    "github.com/runatlantis/atlantis/server/enhanced/locking"
)

// WorkflowLockManager integrates with CI/CD workflows
type WorkflowLockManager struct {
    lockManager    *locking.EnhancedLockManager
    workflowTracker WorkflowTracker
}

type WorkflowTracker interface {
    StartWorkflow(ctx context.Context, workflowID string, metadata map[string]string) error
    CompleteWorkflow(ctx context.Context, workflowID string, success bool) error
    GetWorkflowStatus(ctx context.Context, workflowID string) (*WorkflowStatus, error)
}

type WorkflowStatus struct {
    ID        string
    Status    string
    StartTime time.Time
    EndTime   *time.Time
    Metadata  map[string]string
}

func NewWorkflowLockManager(lockManager *locking.EnhancedLockManager, tracker WorkflowTracker) *WorkflowLockManager {
    return &WorkflowLockManager{
        lockManager:     lockManager,
        workflowTracker: tracker,
    }
}

// AcquireLockForWorkflow acquires a lock and starts workflow tracking
func (wlm *WorkflowLockManager) AcquireLockForWorkflow(ctx context.Context, workflowID string, req *locking.LockRequest) (*locking.Lock, error) {
    // Start workflow tracking
    metadata := map[string]string{
        "project":   req.Project,
        "workspace": req.Workspace,
        "user":      req.User,
        "priority":  req.Priority.String(),
    }

    if err := wlm.workflowTracker.StartWorkflow(ctx, workflowID, metadata); err != nil {
        return nil, fmt.Errorf("failed to start workflow tracking: %w", err)
    }

    // Add workflow ID to lock metadata
    if req.Metadata == nil {
        req.Metadata = make(map[string]string)
    }
    req.Metadata["workflow_id"] = workflowID

    // Acquire lock
    lock, err := wlm.lockManager.AcquireLock(ctx, req)
    if err != nil {
        // Mark workflow as failed if lock acquisition fails
        wlm.workflowTracker.CompleteWorkflow(ctx, workflowID, false)
        return nil, err
    }

    return lock, nil
}

// ReleaseLockAndCompleteWorkflow releases lock and completes workflow
func (wlm *WorkflowLockManager) ReleaseLockAndCompleteWorkflow(ctx context.Context, lockID string, success bool) error {
    // Get lock to extract workflow ID
    locks, err := wlm.lockManager.ListLocks(ctx)
    if err != nil {
        return fmt.Errorf("failed to list locks: %w", err)
    }

    var workflowID string
    for _, lock := range locks {
        if lock.ID == lockID {
            workflowID = lock.Metadata["workflow_id"]
            break
        }
    }

    // Release lock
    if err := wlm.lockManager.ReleaseLock(ctx, lockID); err != nil {
        return fmt.Errorf("failed to release lock: %w", err)
    }

    // Complete workflow tracking
    if workflowID != "" {
        if err := wlm.workflowTracker.CompleteWorkflow(ctx, workflowID, success); err != nil {
            return fmt.Errorf("failed to complete workflow tracking: %w", err)
        }
    }

    return nil
}

// GetWorkflowLocks returns all locks associated with a workflow
func (wlm *WorkflowLockManager) GetWorkflowLocks(ctx context.Context, workflowID string) ([]*locking.Lock, error) {
    allLocks, err := wlm.lockManager.ListLocks(ctx)
    if err != nil {
        return nil, err
    }

    var workflowLocks []*locking.Lock
    for _, lock := range allLocks {
        if lock.Metadata["workflow_id"] == workflowID {
            workflowLocks = append(workflowLocks, lock)
        }
    }

    return workflowLocks, nil
}
```

---

These integration examples provide comprehensive coverage of how to integrate with the Enhanced Locking System across various use cases. For more specific integration scenarios or additional examples, refer to the [Enhanced Locking Documentation](../README.md).