# Enhanced Locking Manager and Events System

## Overview

The Enhanced Locking Manager and Events System (PR #4) provides centralized orchestration and comprehensive event tracking for the Atlantis enhanced locking system. This component serves as the main coordination hub, managing all locking operations, events, and metrics collection.

## Architecture

### Core Components

#### 1. Lock Orchestrator
The central coordination component that manages all enhanced locking operations:

```go
type LockOrchestrator struct {
    // Core components
    backend          Backend
    manager          *EnhancedLockManager
    config          *EnhancedConfig

    // Advanced features
    eventBus        *EventBus
    metricsCollector *MetricsCollector
    queue           *queue.ResourceBasedQueue
    timeoutManager  *timeout.TimeoutManager
    deadlockDetector *deadlock.DeadlockDetector

    // Coordination
    workerPool      *WorkerPool
}
```

**Key Features:**
- Centralized request processing with worker pool
- Component health monitoring and management
- Graceful startup and shutdown coordination
- Request validation and routing
- Event orchestration and metrics collection

#### 2. Event Bus System
Comprehensive event management for lock lifecycle tracking:

```go
type EventBus struct {
    subscriptions map[string]*EventSubscription
    eventBuffer   chan *LockEventData
    metrics       *EventMetrics
}
```

**Supported Events:**
- Lock lifecycle: `lock_requested`, `lock_acquired`, `lock_released`, `lock_timeout`
- Queue operations: `queued_request`, `queue_processed`, `queue_timeout`
- System events: `deadlock_detected`, `system_health_check`, `metrics_updated`
- Error tracking: `backend_error`, `timeout_error`, `validation_error`

#### 3. Metrics Collection
Real-time performance monitoring and analytics:

```go
type MetricsCollector struct {
    manager  *ManagerMetrics
    backend  *BackendMetrics
    deadlock *DeadlockMetrics
    queue    *QueueMetrics
    events   *EventMetrics
    system   *SystemMetrics
}
```

**Metrics Categories:**
- **Manager Metrics**: Request counts, success rates, timing
- **Backend Metrics**: Connection pool, latency, error rates
- **Deadlock Metrics**: Detection and resolution statistics
- **Queue Metrics**: Depth, throughput, wait times
- **System Metrics**: Health scores, uptime, alerts

## Implementation Details

### Request Processing Flow

1. **Request Validation**
   ```go
   func (lo *LockOrchestrator) validateRequest(request *LockOrchestrationRequest) error {
       // Validate required fields
       // Check permissions
       // Verify resource availability
   }
   ```

2. **Worker Pool Distribution**
   ```go
   func (wp *WorkerPool) processRequest(workerID int, request *LockOrchestrationRequest) {
       // Route to appropriate handler
       // Apply timeout and retry logic
       // Emit events for lifecycle tracking
   }
   ```

3. **Response Coordination**
   ```go
   type LockOrchestrationResponse struct {
       RequestID   string
       Success     bool
       Lock        *models.ProjectLock
       Error       string
       Duration    time.Duration
   }
   ```

### Event System Architecture

#### Event Types and Handlers
```go
const (
    EventLockRequested    EventType = "lock_requested"
    EventLockAcquired     EventType = "lock_acquired"
    EventLockReleased     EventType = "lock_released"
    EventQueuedRequest    EventType = "queued_request"
    EventDeadlockDetected EventType = "deadlock_detected"
)
```

#### Built-in Event Handlers
1. **LoggingEventHandler**: Structured logging for all events
2. **MetricsEventHandler**: Automatic metrics updates
3. **Custom handlers**: Webhook notifications, audit trails

### Metrics and Monitoring

#### Health Scoring Algorithm
```go
func (mc *MetricsCollector) calculateHealthScore() int {
    score := 100

    // Deduct for errors (max 50 points)
    if errorRate > 0 {
        score -= int(errorRate * 50)
    }

    // Deduct for deadlocks (10 points each)
    score -= currentDeadlocks * 10

    // Deduct for high queue utilization (max 20 points)
    if queueUtilization > 0.8 {
        score -= int((queueUtilization - 0.8) * 100)
    }

    return max(score, 0)
}
```

#### Performance Tracking
- **Latency Percentiles**: P50, P90, P95, P99 for all operations
- **Throughput Metrics**: Requests/second, locks/second
- **Resource Utilization**: Memory, CPU, network, storage
- **Error Analysis**: Rate tracking, categorization, patterns

## Configuration

### Enhanced Config Structure
```yaml
enhanced_locking:
  enabled: true
  backend: "redis"

  # Event system
  enable_events: true
  event_buffer_size: 1000

  # Metrics collection
  collection_interval: "30s"
  retention_period: "24h"

  # Worker pool
  worker_count: 8
  request_buffer_size: 1000

  # Health monitoring
  health_check_interval: "30s"
  component_timeout: "5s"
```

### Component Configuration
```go
type EnhancedConfig struct {
    // Event configuration
    EnableEvents      bool
    EventBufferSize   int

    // Metrics configuration
    MetricsInterval   time.Duration
    MetricsRetention  time.Duration

    // Worker pool configuration
    WorkerCount       int
    RequestBufferSize int

    // Health monitoring
    HealthCheckInterval time.Duration
    ComponentTimeout    time.Duration
}
```

## API Interface

### Orchestrator Methods
```go
// Primary lock operations
func (lo *LockOrchestrator) Lock(ctx context.Context, project models.Project, workspace string, user models.User) (*models.ProjectLock, error)
func (lo *LockOrchestrator) LockWithOptions(ctx context.Context, project models.Project, workspace string, user models.User, options LockRequestOptions) (*models.ProjectLock, error)
func (lo *LockOrchestrator) Unlock(ctx context.Context, project models.Project, workspace string, user models.User) (*models.ProjectLock, error)

// Status and monitoring
func (lo *LockOrchestrator) GetStatus() *OrchestratorStatus
func (lo *LockOrchestrator) GetMetrics() *MetricsSnapshot
func (lo *LockOrchestrator) GetComponentHealth() map[string]ComponentStatus
```

### Event Subscription
```go
// Subscribe to events
subscription := eventBus.Subscribe(handler, filter)

// Built-in handlers
loggingHandler := NewLoggingEventHandler(log, eventTypes)
metricsHandler := NewMetricsEventHandler(metrics)
```

### Metrics Export
```go
// Export in various formats
jsonData, err := metricsCollector.ExportMetrics("json")
prometheusData, err := metricsCollector.ExportMetrics("prometheus")
```

## Performance Characteristics

### Throughput
- **Target**: 1000+ requests/second per node
- **Scaling**: Linear with worker pool size
- **Bottlenecks**: Backend latency, event processing

### Latency
- **P50**: < 10ms for cache hits
- **P95**: < 100ms for typical operations
- **P99**: < 500ms including queue processing

### Resource Usage
- **Memory**: ~50MB baseline + 1KB per active lock
- **CPU**: ~5% for normal operations
- **Network**: Depends on backend (Redis: ~1KB per operation)

## Error Handling and Recovery

### Component Failure Handling
```go
func (lo *LockOrchestrator) handleComponentFailure(componentName string, err error) {
    // Log error and update status
    lo.log.Error("Component %s failed: %v", componentName, err)
    lo.updateComponentStatus(componentName, "error", 0, err.Error())

    // Attempt restart if configured
    if lo.config.AutoRestart {
        go lo.restartComponent(componentName)
    }

    // Emit system event
    lo.emitEvent(CreateErrorEvent(EventSystemError, ResourceIdentifier{}, err.Error()))
}
```

### Graceful Degradation
- **Event system failure**: Continue operations, log locally
- **Metrics failure**: Maintain basic counters
- **Queue failure**: Fall back to direct processing
- **Backend issues**: Use circuit breaker pattern

## Integration Points

### Backward Compatibility
The orchestrator maintains full compatibility with existing Atlantis lock interfaces:

```go
// Standard LockManager interface compliance
func (lo *LockOrchestrator) Lock(ctx context.Context, project models.Project, workspace string, user models.User) (*models.ProjectLock, error)
func (lo *LockOrchestrator) Unlock(ctx context.Context, project models.Project, workspace string, user models.User) (*models.ProjectLock, error)
func (lo *LockOrchestrator) List(ctx context.Context) ([]*models.ProjectLock, error)
```

### Event Integration
```go
// Custom event handlers
type CustomEventHandler struct {
    webhook string
}

func (h *CustomEventHandler) HandleEvent(ctx context.Context, event *LockEventData) error {
    // Send to external system
    return sendWebhook(h.webhook, event)
}
```

### Metrics Integration
```go
// Prometheus integration
http.Handle("/metrics", promhttp.Handler())

// Custom metrics endpoints
http.HandleFunc("/metrics/atlantis", func(w http.ResponseWriter, r *http.Request) {
    metrics := orchestrator.GetMetrics()
    json.NewEncoder(w).Encode(metrics)
})
```

## Testing Strategy

### Unit Tests
- Component initialization and lifecycle
- Request processing and validation
- Event handling and metrics collection
- Error scenarios and recovery

### Integration Tests
- End-to-end lock operations
- Event system functionality
- Metrics accuracy and export
- Component failure and recovery

### Performance Tests
- Load testing with concurrent requests
- Memory and CPU profiling
- Event system throughput
- Metrics collection overhead

## Deployment Considerations

### Resource Requirements
- **Minimum**: 512MB RAM, 1 CPU core
- **Recommended**: 2GB RAM, 2 CPU cores
- **High-load**: 4GB+ RAM, 4+ CPU cores

### Monitoring Setup
```yaml
# Prometheus scraping
- job_name: 'atlantis-enhanced-locking'
  static_configs:
    - targets: ['atlantis:4141']
  metrics_path: '/metrics'
  scrape_interval: 30s
```

### Alerting Rules
```yaml
# High error rate
- alert: HighLockErrorRate
  expr: atlantis_lock_error_rate > 0.05
  labels:
    severity: warning

# Component health
- alert: ComponentUnhealthy
  expr: atlantis_component_health < 80
  labels:
    severity: critical
```

## Future Enhancements

### Planned Features
1. **Distributed Mode**: Multi-node orchestration
2. **Advanced Analytics**: ML-based performance optimization
3. **Custom Policies**: Rule-based lock management
4. **Integration APIs**: REST/GraphQL interfaces
5. **Dashboard UI**: Real-time monitoring interface

### Extensibility Points
- **Event Handlers**: Custom business logic
- **Metrics Exporters**: Additional formats and destinations
- **Backend Adapters**: New storage systems
- **Policy Engines**: Advanced decision making
- **Notification Systems**: Multi-channel alerting

## Conclusion

The Enhanced Locking Manager and Events System provides a robust, scalable foundation for advanced lock management in Atlantis. With comprehensive monitoring, event tracking, and orchestration capabilities, it enables organizations to manage infrastructure locks at scale while maintaining reliability and performance.

The system's modular architecture ensures easy maintenance and extensibility, while its backward compatibility guarantees seamless integration with existing Atlantis deployments.