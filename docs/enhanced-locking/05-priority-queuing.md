# Priority Queuing and Timeouts - Enhanced Locking System

## Overview

This document describes the Priority Queuing and Timeout Management system, which provides advanced queue processing capabilities with adaptive retry logic and comprehensive timeout handling for the Atlantis enhanced locking system.

## 🎯 Key Features

### Priority-Based Queuing
- **Heap-based Priority Queue**: Efficient O(log n) insertion and removal
- **Five Priority Levels**: Emergency, Critical, High, Normal, Low
- **FIFO Within Priority**: Fair processing within same priority level
- **Dynamic Priority Adjustment**: Runtime priority changes
- **Resource Isolation**: Separate queues per resource to prevent head-of-line blocking

### Advanced Timeout Management
- **Adaptive Timeouts**: Automatically adjust based on historical performance
- **Multiple Timeout Types**: Lock acquisition, queue wait, renewal, connection, operation
- **Timeout Escalation**: Progressive timeout increases for persistent failures
- **Circuit Breaker Integration**: Prevent cascade failures during high contention

### Batch Processing
- **Configurable Batch Size**: Process multiple requests simultaneously
- **Load Balancing**: Intelligent distribution across queues
- **Round-Robin Processing**: Fair queue processing
- **Parallel Execution**: Concurrent batch processing for improved throughput

### Comprehensive Metrics
- **Real-time Monitoring**: Live queue status and performance metrics
- **Historical Analysis**: Track performance trends over time
- **Adaptive Learning**: System learns from past performance
- **Alerting Integration**: Proactive notifications for queue issues

## 🏗️ Architecture

### Component Structure

```
server/core/locking/enhanced/
├── queue/
│   └── priority_queue.go        # Heap-based priority queue implementation
├── timeout/
│   ├── manager.go              # Timeout management and adaptive logic
│   └── retry.go                # Enhanced retry logic with circuit breakers
└── docs/enhanced-locking/
    └── 05-priority-queuing.md  # This documentation
```

### Priority Queue Architecture

```go
type PriorityQueue struct {
    mu       sync.RWMutex
    items    []*QueueItem              // Heap-ordered items
    lookup   map[string]*QueueItem     // Fast O(1) lookups
    metrics  *QueueMetrics             // Performance tracking
    config   QueueConfig               // Configuration
    // Background processing
    ctx      context.Context
    cancel   context.CancelFunc
    wg       sync.WaitGroup
}
```

### Queue Item Structure

```go
type QueueItem struct {
    // Core lock information
    Project   models.Project
    Workspace string
    Pull      models.PullRequest
    User      models.User

    // Queue metadata
    Priority       Priority
    Timestamp      time.Time
    RequestID      string

    // Async communication
    ResultChan     chan QueueResult
    Ctx            context.Context

    // Performance tracking
    QueueEntryTime time.Time
    RetryCount     int
}
```

## 🔧 Configuration

### Queue Configuration

```go
type QueueConfig struct {
    MaxQueueSize        int                    // Maximum items in queue (default: 1000)
    BatchProcessingSize int                    // Items per batch (default: 10)
    ProcessingInterval  time.Duration          // Batch interval (default: 100ms)
    MaxWaitTime         time.Duration          // Max queue wait (default: 10min)
    MaxRetries          int                    // Max retry attempts (default: 3)
    PriorityWeights     map[Priority]int       // Priority weights
    EnableBatching      bool                   // Batch processing (default: true)
    EnableMetrics       bool                   // Metrics collection (default: true)
}
```

### Timeout Configuration

```go
type TimeoutConfig struct {
    // Base timeouts
    BaseLockTimeout       time.Duration        // Default: 30s
    BaseQueueTimeout      time.Duration        // Default: 5min
    BaseRenewalTimeout    time.Duration        // Default: 10s
    BaseConnectionTimeout time.Duration        // Default: 5s
    BaseOperationTimeout  time.Duration        // Default: 60s

    // Adaptive behavior
    EnableAdaptiveTimeout bool                 // Default: true
    MinTimeoutMultiplier  float64             // Default: 0.5
    MaxTimeoutMultiplier  float64             // Default: 3.0
    AdaptationFactor      float64             // Default: 0.1
    HistoryWindow         time.Duration        // Default: 24h

    // Escalation
    EnableEscalation      bool                 // Default: true
    EscalationThreshold   int                  // Default: 3
    EscalationMultiplier  float64             // Default: 1.5
    MaxEscalationLevel    int                  // Default: 5
}
```

## 📊 Priority Levels

### Priority Definitions

| Priority | Value | Use Case | Example |
|----------|-------|----------|---------|
| Emergency | 4 | Critical hotfixes, security patches | Rollback production outage |
| Critical | 3 | Production deployments | Release to production |
| High | 2 | Important features, urgent fixes | Customer-impacting bug fix |
| Normal | 1 | Standard development work | Feature development |
| Low | 0 | Background tasks, cleanup | Documentation updates |

### Priority Assignment Rules

```go
func DeterminePriority(project models.Project, pull models.PullRequest) Priority {
    // Emergency: Security fixes, hotfixes
    if strings.Contains(pull.Title, "HOTFIX") ||
       strings.Contains(pull.Title, "SECURITY") {
        return PriorityEmergency
    }

    // Critical: Production branches
    if pull.BaseBranch == "main" || pull.BaseBranch == "production" {
        return PriorityCritical
    }

    // High: Release branches
    if strings.HasPrefix(pull.BaseBranch, "release/") {
        return PriorityHigh
    }

    // Normal: Default for development
    return PriorityNormal
}
```

## ⏱️ Timeout Management

### Adaptive Timeout Calculation

The system automatically adjusts timeouts based on historical performance:

```go
func (tm *TimeoutManager) calculateAdaptiveTimeout(timeoutType TimeoutType, resourceID string) time.Duration {
    baseTimeout := tm.getBaseTimeout(timeoutType)

    if !tm.config.EnableAdaptiveTimeout {
        return baseTimeout
    }

    history := tm.getTimeoutHistory(resourceID)
    if history == nil {
        return baseTimeout
    }

    // Calculate success rate
    successRate := float64(history.SuccessCount) / float64(history.SuccessCount + history.TimeoutCount)

    // Adjust multiplier based on success rate
    var multiplier float64
    if successRate > 0.9 {
        // High success rate - reduce timeout
        multiplier = tm.config.MinTimeoutMultiplier +
                    (1.0-tm.config.MinTimeoutMultiplier)*(successRate-0.9)*10
    } else if successRate < 0.5 {
        // Low success rate - increase timeout
        multiplier = 1.0 + (tm.config.MaxTimeoutMultiplier-1.0)*(0.5-successRate)*2
    } else {
        // Moderate success rate - slight adjustment
        multiplier = 1.0 + tm.config.AdaptationFactor*(0.8-successRate)
    }

    return time.Duration(float64(baseTimeout) * multiplier)
}
```

### Timeout Escalation

When operations consistently fail, the system escalates timeouts:

```go
func (tm *TimeoutManager) calculateRetryDelay(timeout *TimeoutContext) time.Duration {
    baseDelay := tm.getBaseTimeout(timeout.Type) / 4 // 25% of base timeout

    escalationLevel := timeout.RetryCount
    if escalationLevel > tm.config.MaxEscalationLevel {
        escalationLevel = tm.config.MaxEscalationLevel
    }

    multiplier := 1.0
    for i := 0; i < escalationLevel; i++ {
        multiplier *= tm.config.EscalationMultiplier
    }

    return time.Duration(float64(baseDelay) * multiplier)
}
```

## 🔄 Retry Logic

### Retry Strategies

The system supports multiple retry strategies:

1. **Fixed Delay**: Constant delay between retries
2. **Exponential Backoff**: Exponentially increasing delays
3. **Linear Backoff**: Linearly increasing delays
4. **Jittered Exponential**: Exponential backoff with random jitter
5. **Adaptive Backoff**: Learns from historical performance

### Enhanced Retry Policies

```go
// Default policy for most operations
func DefaultRetryPolicy() RetryPolicy {
    return RetryPolicy{
        MaxAttempts:       3,
        Strategy:          JitteredExponential,
        BaseDelay:         1 * time.Second,
        MaxDelay:          30 * time.Second,
        BackoffMultiplier: 2.0,
        JitterPercent:     0.1,
        CircuitBreaker:    DefaultCircuitBreakerConfig(),
        RateLimiter:       DefaultRateLimiterConfig(),
    }
}

// Aggressive policy for critical operations
func AggressiveRetryPolicy() RetryPolicy {
    return RetryPolicy{
        MaxAttempts:       5,
        Strategy:          AdaptiveBackoff,
        BaseDelay:         500 * time.Millisecond,
        MaxDelay:          60 * time.Second,
        BackoffMultiplier: 1.5,
        JitterPercent:     0.2,
    }
}
```

### Circuit Breaker Integration

Circuit breakers prevent cascade failures during high contention:

```go
type CircuitBreakerConfig struct {
    FailureThreshold int           // Failures before opening (default: 5)
    SuccessThreshold int           // Successes to close (default: 3)
    Timeout          time.Duration // Open duration (default: 30s)
    ResetTimeout     time.Duration // Reset period (default: 60s)
}
```

## 📈 Metrics and Monitoring

### Queue Metrics

```go
type QueueMetrics struct {
    TotalEnqueued         int64                    // Total items added
    TotalProcessed        int64                    // Total items processed
    TotalTimeouts         int64                    // Total timeout events
    TotalErrors           int64                    // Total error events
    CurrentQueueSize      int64                    // Current queue depth
    AverageWaitTime       time.Duration            // Average queue wait time
    PriorityDistribution  map[Priority]int64       // Items by priority
    BatchesProcessed      int64                    // Total batches processed
    LastProcessingTime    time.Time               // Last processing timestamp
    ProcessingDuration    time.Duration            // Last processing duration
}
```

### Timeout Metrics

```go
type TimeoutMetrics struct {
    TotalTimeouts         int64                    // Total timeout events
    TimeoutsByType        map[TimeoutType]int64    // Timeouts by type
    AdaptiveAdjustments   int64                    // Adaptive adjustments made
    EscalationEvents      int64                    // Escalation events
    AverageResolution     time.Duration            // Average resolution time
    CurrentActiveTimeouts int64                    // Active timeout count
}
```

### Retry Metrics

```go
type EnhancedRetryMetrics struct {
    TotalRetries          int64                    // Total retry attempts
    SuccessfulRetries     int64                    // Successful retries
    FailedRetries         int64                    // Failed retries
    CircuitBreakerTrips   int64                    // Circuit breaker activations
    RateLimitedRetries    int64                    // Rate limited attempts
    AdaptiveAdjustments   int64                    // Adaptive strategy changes
    StrategyDistribution  map[RetryStrategy]int64  // Usage by strategy
    AverageRetryLatency   time.Duration            // Average retry duration
    MaxRetryAttempts      int                      // Highest retry count seen
}
```

## 🚀 Usage Examples

### Basic Queue Usage

```go
// Create priority queue
config := QueueConfig{
    MaxQueueSize:        1000,
    BatchProcessingSize: 10,
    ProcessingInterval:  100 * time.Millisecond,
    MaxWaitTime:        10 * time.Minute,
    EnableBatching:     true,
    EnableMetrics:      true,
}

queue := NewPriorityQueue(config, logger)

// Enqueue a high-priority lock request
ctx := context.Background()
item, err := queue.Enqueue(ctx, project, workspace, pull, user, PriorityHigh)
if err != nil {
    return fmt.Errorf("failed to enqueue: %w", err)
}

// Wait for result
select {
case result := <-item.ResultChan:
    if result.Success {
        log.Info("Lock acquired after %v wait", result.WaitTime)
    } else {
        log.Error("Lock acquisition failed: %v", result.Error)
    }
case <-ctx.Done():
    return ctx.Err()
}
```

### Timeout Management

```go
// Create timeout manager
config := TimeoutConfig{
    EnableAdaptiveTimeout: true,
    EnableEscalation:     true,
    BaseLockTimeout:      30 * time.Second,
    MaxTimeoutMultiplier: 3.0,
}

timeoutMgr := NewTimeoutManager(config, logger)

// Create adaptive timeout
timeout, err := timeoutMgr.CreateTimeout(
    ctx,
    LockAcquisitionTimeout,
    resourceID,
    userID,
    nil, // custom context
)
if err != nil {
    return fmt.Errorf("failed to create timeout: %w", err)
}

// Wait for timeout or completion
select {
case result := <-timeout.ResultChan:
    if result.Success {
        log.Info("Operation completed in %v", result.Duration)
    } else {
        log.Error("Operation timed out: %v", result.Error)
    }
case <-ctx.Done():
    timeoutMgr.CancelTimeout(timeout.ID)
    return ctx.Err()
}
```

### Enhanced Retry Logic

```go
// Create retry manager with adaptive backoff
policy := RetryPolicy{
    MaxAttempts:       3,
    Strategy:          AdaptiveBackoff,
    BaseDelay:         1 * time.Second,
    MaxDelay:          30 * time.Second,
    BackoffMultiplier: 2.0,
    JitterPercent:     0.1,
}

retryMgr := NewEnhancedRetryManager(policy, logger)

// Execute operation with retry
err := retryMgr.ExecuteWithRetry(ctx, "lock-acquisition", func(ctx context.Context, attempt int) error {
    // Attempt lock acquisition
    return lockBackend.TryLock(lock)
})

if err != nil {
    return fmt.Errorf("lock acquisition failed after retries: %w", err)
}
```

## 🔍 Monitoring and Alerting

### Key Metrics to Monitor

1. **Queue Depth**: Monitor queue size to detect bottlenecks
2. **Wait Times**: Track average and P95 wait times
3. **Timeout Rate**: Monitor adaptive timeout effectiveness
4. **Retry Success Rate**: Track retry policy effectiveness
5. **Circuit Breaker State**: Monitor for cascade failure prevention

### Alerting Thresholds

```yaml
alerts:
  queue_depth:
    warning: 100 items
    critical: 500 items
  average_wait_time:
    warning: 30 seconds
    critical: 2 minutes
  timeout_rate:
    warning: 5%
    critical: 15%
  retry_failure_rate:
    warning: 10%
    critical: 25%
```

### Grafana Dashboard Queries

```promql
# Queue depth over time
atlantis_queue_size{queue_type="priority"}

# Average wait time by priority
rate(atlantis_queue_wait_time_sum[5m]) / rate(atlantis_queue_wait_time_count[5m])

# Timeout rate by type
rate(atlantis_timeouts_total[5m]) / rate(atlantis_operations_total[5m])

# Retry success rate
rate(atlantis_retry_success_total[5m]) / rate(atlantis_retry_total[5m])
```

## 🛡️ Fault Tolerance

### Circuit Breaker Protection

The system includes circuit breakers to prevent cascade failures:

- **Closed State**: Normal operation, all requests pass through
- **Open State**: Circuit breaker trips, requests fail fast
- **Half-Open State**: Limited requests allowed to test recovery

### Rate Limiting

Token bucket rate limiting prevents resource exhaustion:

```go
rateLimiter := NewRateLimiter(
    10.0,  // maxTokens: burst capacity
    1.0,   // refillRate: tokens per second
)

if !rateLimiter.Allow() {
    return fmt.Errorf("rate limited")
}
```

### Graceful Degradation

When queues become overwhelmed:

1. **Priority Shedding**: Drop low-priority requests first
2. **Queue Limits**: Enforce maximum queue sizes
3. **Timeout Reduction**: Reduce timeouts under high load
4. **Circuit Breaking**: Fail fast to prevent cascade failures

## 🧪 Testing

### Unit Tests

```bash
# Test priority queue functionality
go test ./server/core/locking/enhanced/queue -v

# Test timeout management
go test ./server/core/locking/enhanced/timeout -v

# Test retry logic
go test ./server/core/locking/enhanced/timeout -run TestRetry -v
```

### Integration Tests

```bash
# Test complete priority queuing system
go test ./server/core/locking/enhanced/tests -run TestPriorityQueue -v

# Test timeout integration
go test ./server/core/locking/enhanced/tests -run TestTimeout -v

# Test batch processing
go test ./server/core/locking/enhanced/tests -run TestBatch -v
```

### Load Testing

```bash
# Simulate high queue load
go test ./server/core/locking/enhanced/tests -run TestQueueLoad -v

# Test timeout performance under load
go test ./server/core/locking/enhanced/tests -run TestTimeoutLoad -v

# Test retry effectiveness
go test ./server/core/locking/enhanced/tests -run TestRetryLoad -v
```

## 🔧 Performance Tuning

### Queue Optimization

1. **Batch Size**: Tune based on system capacity and latency requirements
2. **Processing Interval**: Balance between responsiveness and efficiency
3. **Queue Size**: Set based on expected peak load

### Timeout Tuning

1. **Base Timeouts**: Set based on normal operation latency
2. **Adaptation Factor**: Control how aggressively timeouts adapt
3. **History Window**: Balance between responsiveness and stability

### Retry Optimization

1. **Max Attempts**: Balance between resilience and latency
2. **Backoff Strategy**: Choose based on failure patterns
3. **Jitter**: Reduce thundering herd effects

## 📝 Best Practices

### Queue Management

1. **Monitor Queue Depth**: Set up alerting for queue buildup
2. **Tune Batch Size**: Optimize for your workload characteristics
3. **Use Appropriate Priorities**: Don't overuse high priorities
4. **Clean Up Stale Requests**: Implement request expiration

### Timeout Strategy

1. **Enable Adaptive Timeouts**: Let the system learn optimal values
2. **Monitor Success Rates**: Adjust base timeouts based on trends
3. **Use Escalation Judiciously**: Don't let timeouts grow unbounded
4. **Consider Circuit Breakers**: Prevent cascade failures

### Retry Policy

1. **Choose Appropriate Strategy**: Match strategy to failure patterns
2. **Add Jitter**: Prevent thundering herd effects
3. **Use Circuit Breakers**: Fail fast during outages
4. **Monitor Retry Metrics**: Ensure policies are effective

## 🚨 Troubleshooting

### Common Issues

#### Queue Buildup

**Symptoms**: Increasing queue depth, longer wait times
**Causes**: Insufficient processing capacity, stuck operations
**Solutions**:
- Increase batch size
- Add more processing threads
- Implement request timeouts
- Scale horizontally

#### Timeout Thrashing

**Symptoms**: Frequent timeout adjustments, unstable performance
**Causes**: Highly variable operation latency, incorrect base timeouts
**Solutions**:
- Increase adaptation factor
- Extend history window
- Adjust base timeout values
- Implement request classification

#### Retry Storms

**Symptoms**: High retry rates, circuit breaker trips
**Causes**: Systematic failures, inadequate retry policies
**Solutions**:
- Implement exponential backoff
- Add circuit breakers
- Use jittered retries
- Classify error types

### Debugging Commands

```bash
# Check queue status
curl -s http://atlantis:4141/api/queue/status | jq .

# View timeout metrics
curl -s http://atlantis:4141/api/timeout/metrics | jq .

# Check retry statistics
curl -s http://atlantis:4141/api/retry/stats | jq .

# Get adaptive data
curl -s http://atlantis:4141/api/adaptive/data | jq .
```

## 🔮 Future Enhancements

### Planned Features

1. **Machine Learning Integration**: Use ML for timeout prediction
2. **Dynamic Priority Adjustment**: Automatically adjust priorities based on context
3. **Cross-Queue Load Balancing**: Distribute load across multiple queue instances
4. **Advanced Circuit Breaker Patterns**: Implement bulkhead and timeout patterns
5. **Queue Persistence**: Survive restarts with persistent queue state

### Performance Improvements

1. **Lock-Free Data Structures**: Reduce contention in high-throughput scenarios
2. **NUMA Awareness**: Optimize for multi-socket systems
3. **Compression**: Compress queue state for memory efficiency
4. **Sharding**: Distribute queues across multiple nodes

---

This priority queuing and timeout system provides robust, scalable queue management with intelligent timeout handling and comprehensive retry logic for the enhanced Atlantis locking system.