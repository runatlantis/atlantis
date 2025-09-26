# Enhanced Locking System: Deadlock Detection and Resolution

## Overview

The deadlock detection and resolution system provides automatic detection of circular dependencies in lock acquisition and implements configurable policies for resolving deadlocks when they occur. This system prevents the locking system from becoming permanently blocked due to circular wait conditions.

## Architecture

### Wait-For Graph Algorithm

The system uses a **wait-for graph** approach for deadlock detection:

- **Nodes**: Represent lock requests waiting for resources
- **Edges**: Represent wait relationships (Request A waits for Lock B)
- **Cycles**: Indicate deadlocks requiring resolution

```
Request A → Lock B (held by Request C)
Request C → Lock D (held by Request A)
         ↑________________↓
         Deadlock Cycle Detected
```

### Core Components

#### 1. Detector (`detector.go`)
- **Purpose**: Continuous monitoring and cycle detection
- **Algorithm**: Depth-First Search (DFS) with recursion stack
- **Features**:
  - Configurable scan intervals
  - Maximum cycle length limits
  - Thread-safe graph management
  - Comprehensive metrics collection

#### 2. Resolver (`resolver.go`)
- **Purpose**: Automatic deadlock resolution policies
- **Strategies**: Multiple resolution algorithms
- **Features**:
  - Cooldown periods to prevent rapid re-resolution
  - High-priority request preservation
  - Configurable resolution limits per cycle
  - Success/failure tracking

## Configuration

### Detection Configuration

```yaml
deadlock_detection:
  enabled: true
  scan_interval: "10s"          # How often to scan for deadlocks
  max_cycle_length: 10          # Maximum cycle size to detect
  timeout_threshold: "30s"      # When to consider waits suspicious
  resolution_policy: "youngest_first"  # Auto-resolution strategy
```

### Resolution Policies

#### 1. Youngest First (`youngest_first`)
**Strategy**: Abort the most recently created lock request
**Rationale**: Newer requests have done less work and are cheaper to restart
**Use Case**: General purpose, balances fairness with efficiency

```go
config.Policy = deadlock.YoungestFirst
```

#### 2. Lowest Priority (`lowest_priority`)
**Strategy**: Abort the request with the lowest priority level
**Rationale**: Preserve high-priority work (emergency, critical operations)
**Use Case**: Systems with clear priority hierarchies

```go
config.Policy = deadlock.LowestPriority
config.PreserveHighPriority = true  // Protect Emergency/Critical
```

#### 3. Maximum Retries (`max_retries`)
**Strategy**: Abort the request that has retried the most times
**Rationale**: Requests that keep failing may have underlying issues
**Use Case**: Systems where retry count indicates problem requests

```go
config.Policy = deadlock.MaxRetries
```

#### 4. Random Choice (`random`)
**Strategy**: Randomly select a request to abort
**Rationale**: Simple, unbiased selection
**Use Case**: When no clear preference exists

```go
config.Policy = deadlock.RandomChoice
```

#### 5. No Auto-Resolve (`no_auto_resolve`)
**Strategy**: Detect but don't automatically resolve deadlocks
**Rationale**: Require manual intervention for deadlock resolution
**Use Case**: Critical systems requiring human oversight

```go
config.Policy = deadlock.NoAutoResolve
```

### Resolution Configuration

```yaml
deadlock_resolution:
  policy: "youngest_first"
  max_resolutions_per_cycle: 3    # Limit simultaneous resolutions
  cooldown_period: "30s"          # Prevent rapid re-resolution
  notify_on_resolution: true      # Enable resolution callbacks
  preserve_high_priority: true    # Protect Emergency/Critical requests
```

## Integration

### Initialization

```go
import (
    "github.com/runatlantis/atlantis/server/core/locking/enhanced/deadlock"
    "github.com/runatlantis/atlantis/server/logging"
)

// Setup detection
detectionConfig := deadlock.DefaultDetectionConfig()
detectionConfig.ScanInterval = 10 * time.Second
detector := deadlock.NewDetector(detectionConfig, logger)

// Setup resolution with callback
resolutionConfig := deadlock.DefaultResolutionConfig()
resolutionConfig.Policy = deadlock.YoungestFirst

callback := func(result *deadlock.ResolutionResult) {
    logger.Warn("Deadlock resolved: %s aborted (reason: %s)", 
        result.AbortedRequestID, result.Reason)
}

resolver := deadlock.NewResolver(resolutionConfig, logger, callback)
```

### Runtime Operations

```go
// Start detection
ctx := context.Background()
err := detector.Start(ctx)
if err != nil {
    return fmt.Errorf("failed to start deadlock detector: %w", err)
}
defer detector.Stop()

// Add wait relationships as locks are requested
detector.AddWaitRelation("request-123", waitingForLock)

// Detect and resolve cycles
cycles, err := detector.DetectCycles()
if err != nil {
    return fmt.Errorf("cycle detection failed: %w", err)
}

if len(cycles) > 0 {
    results, err := resolver.ResolveCycles(ctx, cycles)
    if err != nil {
        return fmt.Errorf("cycle resolution failed: %w", err)
    }
    
    for _, result := range results {
        if result.Success {
            // Handle successful resolution
            handleResolution(result)
        } else {
            // Handle resolution failure
            logger.Error("Failed to resolve deadlock: %v", result.Error)
        }
    }
}

// Clean up when request completes
detector.RemoveWaitRelation("request-123")
```

## Monitoring and Metrics

### Detection Metrics

```go
metrics := detector.GetMetrics()
```

Available metrics:
- `cycles_detected`: Total number of deadlock cycles found
- `resolutions_applied`: Number of successful resolutions
- `active_wait_relations`: Current number of wait relationships
- `active_locks`: Number of locks with waiters
- `detection_enabled`: Whether detection is active
- `scan_interval`: Current scan frequency

### Resolution Metrics

```go
metrics := resolver.GetMetrics()
```

Available metrics:
- `resolutions_attempted`: Total resolution attempts
- `resolutions_succeeded`: Successful resolutions
- `resolutions_failed`: Failed resolution attempts
- `success_rate`: Resolution success percentage
- `policy`: Current resolution policy
- `cooldown_period`: Cooldown configuration
- `active_cooldowns`: Number of cycles in cooldown

### Wait-For Graph Debugging

```go
graph := detector.GetWaitForGraph()
```

Returns graph structure for debugging:
```json
{
  "edges": {
    "request-123": "namespace:project:workspace",
    "request-456": "namespace:other:workspace"
  },
  "waiters": {
    "namespace:project:workspace": ["request-456", "request-789"],
    "namespace:other:workspace": ["request-123"]
  }
}
```

## Performance Characteristics

### Detection Performance
- **Time Complexity**: O(V + E) for cycle detection using DFS
- **Space Complexity**: O(V) for visited/recursion tracking
- **Scan Frequency**: Configurable (default: 10 seconds)
- **Maximum Cycle Length**: Configurable limit (default: 10 nodes)

### Resolution Performance
- **Selection Time**: O(N) where N is cycle size
- **Cooldown Tracking**: O(1) lookup with periodic cleanup
- **Memory Usage**: Minimal overhead for tracking recent resolutions

### Benchmarks

Based on integration tests:
- **10-node cycle detection**: ~1ms per detection
- **Resolution selection**: ~100μs per cycle
- **Graph updates**: ~10μs per add/remove operation

## Error Handling

### Detection Errors
```go
cycles, err := detector.DetectCycles()
if err != nil {
    // Handle detection failure
    logger.Error("Deadlock detection failed: %v", err)
    // Continue with normal operations
}
```

### Resolution Errors
```go
results, err := resolver.ResolveCycles(ctx, cycles)
if err != nil {
    logger.Error("Resolution system failed: %v", err)
    // Fall back to manual intervention
    return err
}

for _, result := range results {
    if !result.Success {
        logger.Error("Failed to resolve cycle %s: %v", 
            result.CycleID, result.Error)
        // Handle individual resolution failure
    }
}
```

### Common Error Scenarios

1. **Graph Corruption**: If wait-for graph becomes inconsistent
   - **Recovery**: Restart detection system
   - **Prevention**: Proper cleanup on request completion

2. **Resolution Failure**: When selected victim cannot be aborted
   - **Recovery**: Try alternative resolution policy
   - **Escalation**: Manual intervention may be required

3. **Performance Degradation**: Large graphs causing slow detection
   - **Mitigation**: Reduce scan frequency or max cycle length
   - **Optimization**: Consider graph partitioning strategies

## Best Practices

### Configuration Tuning

1. **Scan Interval**:
   - **Fast systems**: 5-10 seconds
   - **High-load systems**: 15-30 seconds
   - **Low-priority systems**: 60+ seconds

2. **Resolution Policy**:
   - **Development environments**: `youngest_first`
   - **Production systems**: `lowest_priority` with high-priority preservation
   - **Critical systems**: `no_auto_resolve` with manual oversight

3. **Cooldown Period**:
   - **Normal operations**: 30-60 seconds
   - **High-contention systems**: 2-5 minutes
   - **Development/testing**: 10-30 seconds

### Operational Guidelines

1. **Monitoring Setup**:
   - Alert on detection of deadlocks
   - Track resolution success rates
   - Monitor detection performance metrics

2. **Incident Response**:
   - Log all deadlock detections for analysis
   - Review resolution decisions for correctness
   - Analyze patterns to prevent future deadlocks

3. **Maintenance Tasks**:
   - Periodic review of resolution policies
   - Performance tuning based on workload patterns
   - Regular testing of deadlock scenarios

## Integration with Enhanced Locking

The deadlock detection system integrates seamlessly with other enhanced locking features:

### Priority Queuing Integration
- Resolution policies respect priority levels
- High-priority requests can be protected from abortion
- Queue metrics influence resolution decisions

### Timeout Management Integration
- Detection timeout thresholds coordinate with retry systems
- Failed resolutions can trigger timeout escalation
- Adaptive timeout learning considers deadlock frequency

### Event System Integration
- Deadlock events published to monitoring systems
- Resolution events trigger notification callbacks
- Metrics collection for system health monitoring

## Testing and Validation

### Unit Tests
- Individual component testing (detector, resolver)
- Policy validation for all resolution strategies
- Error condition handling verification

### Integration Tests
- End-to-end deadlock detection and resolution
- Multi-node cycle scenarios
- Performance benchmarking under load

### Scenario Testing
- Complex deadlock patterns (chains, stars, meshes)
- High-contention environments
- Priority-based resolution validation
- Cooldown period effectiveness

## Migration and Rollout

### Phase 1: Detection Only
```yaml
deadlock_detection:
  enabled: true
  resolution_policy: "no_auto_resolve"  # Detect but don't resolve
```

### Phase 2: Conservative Resolution
```yaml
deadlock_detection:
  enabled: true
  resolution_policy: "youngest_first"
  max_resolutions_per_cycle: 1          # Limited resolution
  cooldown_period: "300s"               # Long cooldown
```

### Phase 3: Full Operation
```yaml
deadlock_detection:
  enabled: true
  resolution_policy: "lowest_priority"
  max_resolutions_per_cycle: 3
  cooldown_period: "30s"
  preserve_high_priority: true
```

This deadlock detection and resolution system provides robust protection against circular dependencies while maintaining high performance and configurability for various operational requirements.