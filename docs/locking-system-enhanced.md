# Enhanced Locking System Architecture

## Table of Contents
1. [Overview](#overview)
2. [Architecture Components](#architecture-components)
3. [Lock ID Format Evolution](#lock-id-format-evolution)
4. [Enhanced Features](#enhanced-features)
5. [Backend Abstraction](#backend-abstraction)
6. [Queue System](#queue-system)
7. [Resource Management](#resource-management)
8. [System Diagrams](#system-diagrams)
9. [Technical Specifications](#technical-specifications)
10. [Migration Guide](#migration-guide)

## Overview

The Enhanced Locking System represents a significant evolution of Atlantis's original locking mechanism, introducing enterprise-grade features while maintaining full backward compatibility. This system provides priority-based queuing, deadlock detection, distributed storage, and sophisticated timeout management.

### Key Improvements
- **Priority-Based Queuing**: 4-tier priority system (Low, Normal, High, Critical)
- **Deadlock Prevention & Resolution**: Wait-for graph based detection with configurable resolution policies
- **Distributed Backend Support**: Redis cluster support with atomic operations
- **Resource Isolation**: Prevents head-of-line blocking with resource-based queuing
- **Fault Tolerance**: Circuit breakers, retry mechanisms, and adaptive timeouts
- **Event-Driven Architecture**: Real-time notifications and monitoring
- **Backward Compatibility**: Seamless integration with existing Atlantis deployments

## Architecture Components

### Core Components Hierarchy

```
Enhanced Locking System
├── EnhancedLockManager (Orchestrator)
├── LockingAdapter (Legacy Compatibility)
├── Backend Abstraction Layer
│   ├── RedisBackend (Distributed)
│   └── BoltDBBackend (Local - Legacy)
├── Advanced Subsystems
│   ├── PriorityQueue System
│   ├── TimeoutManager
│   ├── DeadlockDetector
│   └── RetryManager
└── Supporting Components
    ├── Circuit Breakers
    ├── Rate Limiters
    └── Event Publishers
```

### 1. EnhancedLockManager

The central orchestrator that coordinates all locking operations with advanced capabilities.

**Key Responsibilities:**
- Lock lifecycle management
- Priority-based request handling
- Deadlock prevention and resolution
- Timeout and retry coordination
- Event emission and callback handling
- Metrics collection and health monitoring

**Interface Methods:**
```go
type LockManager interface {
    // Core operations (backward compatible)
    Lock(ctx context.Context, project models.Project, workspace string, user models.User) (*models.ProjectLock, error)
    Unlock(ctx context.Context, project models.Project, workspace string, user models.User) (*models.ProjectLock, error)
    List(ctx context.Context) ([]*models.ProjectLock, error)

    // Enhanced operations
    LockWithPriority(ctx context.Context, project models.Project, workspace string, user models.User, priority Priority) (*models.ProjectLock, error)
    LockWithTimeout(ctx context.Context, project models.Project, workspace string, user models.User, timeout time.Duration) (*models.ProjectLock, error)

    // Queue management
    GetQueuePosition(ctx context.Context, project models.Project, workspace string) (int, error)
    CancelQueuedRequest(ctx context.Context, project models.Project, workspace string, user models.User) error
}
```

### 2. Backend Abstraction

The enhanced system supports multiple storage backends through a unified interface.

**Backend Interface:**
```go
type Backend interface {
    // Core locking operations
    AcquireLock(ctx context.Context, request *EnhancedLockRequest) (*EnhancedLock, error)
    ReleaseLock(ctx context.Context, lockID string) error
    GetLock(ctx context.Context, lockID string) (*EnhancedLock, error)
    ListLocks(ctx context.Context) ([]*EnhancedLock, error)

    // Enhanced operations
    TryAcquireLock(ctx context.Context, request *EnhancedLockRequest) (*EnhancedLock, bool, error)
    RefreshLock(ctx context.Context, lockID string, extension time.Duration) error
    TransferLock(ctx context.Context, lockID string, newOwner string) error

    // Queue operations
    EnqueueLockRequest(ctx context.Context, request *EnhancedLockRequest) error
    DequeueNextRequest(ctx context.Context) (*EnhancedLockRequest, error)
}
```

### 3. LockingAdapter (Backward Compatibility)

Ensures seamless integration with existing Atlantis installations.

**Compatibility Features:**
- Implements original `locking.Backend` interface
- Automatic fallback to legacy systems
- Lock format conversion (enhanced ↔ legacy)
- Configuration-driven behavior
- Zero-downtime migration support

## Lock ID Format Evolution

### Legacy Format
```
Format: {repository}/{path}/{workspace}
Example: "runatlantis/atlantis/server/core/default"

Components:
- repository: GitHub repository full name
- path: Project path within repository
- workspace: Terraform workspace name

Limitations:
- Not globally unique
- No timestamp information
- Limited metadata capacity
- Single-backend assumption
```

### Enhanced Format
```
Format: {type}_{timestamp}_{owner}_{hash}
Example: "lock_1638360000000000000_alice_a1b2c3d4e5f6"

Components:
- type: Always "lock" for lock IDs
- timestamp: Nanosecond Unix timestamp
- owner: Username of lock owner
- hash: 8-character hex hash for uniqueness

Advantages:
- Globally unique across distributed systems
- Temporal ordering capability
- Owner information embedded
- Hash collision prevention
- Backend-agnostic design
```

### Resource Identification

Enhanced locks use structured resource identification:

```go
type ResourceIdentifier struct {
    Type      ResourceType `json:"type"`      // project, workspace, global, custom
    Namespace string       `json:"namespace"` // Repository namespace
    Name      string       `json:"name"`      // Resource name
    Workspace string       `json:"workspace"` // Terraform workspace
    Path      string       `json:"path"`      // Project path
}
```

**Examples:**
```json
{
  "type": "project",
  "namespace": "runatlantis/atlantis",
  "name": "server/core",
  "workspace": "production",
  "path": "server/core"
}
```

## Enhanced Features

### 1. Priority System

Four-tier priority system ensures critical operations get precedence:

```go
type Priority int

const (
    PriorityLow      Priority = 0  // Background operations
    PriorityNormal   Priority = 1  // Standard operations (default)
    PriorityHigh     Priority = 2  // Important operations
    PriorityCritical Priority = 3  // Emergency operations
)
```

**Usage Examples:**
- `PriorityLow`: Cleanup operations, background maintenance
- `PriorityNormal`: Regular plan/apply operations
- `PriorityHigh`: Hotfix deployments, urgent changes
- `PriorityCritical`: Emergency rollbacks, security patches

### 2. Queue Management

Resource-based queuing prevents head-of-line blocking:

```
Global Queue (Traditional):
Request1 (Low) → Request2 (High) → Request3 (Normal)
↑ Blocks all subsequent requests

Resource-Based Queues (Enhanced):
Resource A: Request1 (Low) → Request3 (Normal)
Resource B: Request2 (High) ← Can proceed independently
Resource C: Request4 (Critical) ← Highest priority
```

### 3. Deadlock Detection

Implements wait-for graph based deadlock detection:

**Detection Algorithm:**
1. Maintain wait-for graph of lock dependencies
2. Periodic cycle detection using depth-first search
3. Configurable resolution policies:
   - `LIFO`: Abort newest request
   - `FIFO`: Abort oldest request
   - `lowest_priority`: Abort lowest priority request
   - `random`: Random victim selection

**Prevention:**
- Pre-acquisition deadlock simulation
- Cycle detection before granting locks
- Configurable prevention thresholds

### 4. Timeout Management

Sophisticated timeout handling with multiple strategies:

**Standard Timeouts:**
- Per-request timeout configuration
- Automatic cleanup of expired locks
- Configurable default and maximum timeouts

**Adaptive Timeouts:**
- System load-based adjustment
- Success rate consideration
- Historical latency analysis
- Dynamic threshold adaptation

**Circuit Breakers:**
- Failure threshold detection
- Automatic request blocking during outages
- Gradual recovery with half-open state
- Configurable reset policies

## Backend Abstraction

### RedisBackend Implementation

**Key Features:**
- Atomic operations via Lua scripts
- Cluster mode support
- Pub/Sub event notifications
- TTL-based expiration
- Queue persistence

**Redis Key Schema:**
```
Lock Keys:    atlantis:enhanced:lock:{namespace}:{path}:{workspace}
Queue Keys:   atlantis:enhanced:queue:{namespace}:{path}:{workspace}
Event Keys:   atlantis:lock:{event_type}
```

**Atomic Lock Acquisition (Lua Script):**
```lua
-- Check if lock exists
local existing = redis.call('GET', lockKey)
if existing then
    if queueEnabled then
        -- Add to priority queue
        local score = (4 - priority) * 1000000 + redis.call('TIME')[1]
        redis.call('ZADD', queueKey, score, lockData)
        return {false, "queued"}
    end
    return {false, "exists"}
end

-- Acquire the lock
redis.call('SETEX', lockKey, ttl, lockData)
redis.call('PUBLISH', 'atlantis:lock:acquired', lockKey)
return {true, "acquired"}
```

### BoltDB Backend (Legacy)

Maintains compatibility with existing single-node deployments:
- File-based storage
- ACID transactions
- Embedded database
- No external dependencies

## Queue System

### Priority Queue Implementation

Uses Go's heap interface for efficient priority ordering:

```go
type PriorityQueue struct {
    items    []*QueueItem
    mutex    sync.RWMutex
    maxSize  int
    notEmpty chan struct{}
}

type QueueItem struct {
    Request   *EnhancedLockRequest
    Priority  Priority
    Timestamp time.Time
    Index     int // heap index
}
```

**Ordering Logic:**
1. Higher priority requests processed first
2. Within same priority, FIFO ordering by timestamp
3. Configurable queue size limits per resource

### Resource-Based Queuing

Prevents global blocking through resource isolation:

```go
type ResourceBasedQueue struct {
    queues   map[string]*PriorityQueue  // Resource key → Queue
    mutex    sync.RWMutex
    maxSize  int
}
```

**Benefits:**
- Independent processing per resource
- Eliminates head-of-line blocking
- Improved concurrent throughput
- Better resource utilization

## Resource Management

### Resource Hierarchy

```
Global Resources
├── Repository Resources (org/repo)
│   ├── Project Resources (path)
│   │   ├── Workspace Resources (workspace)
│   │   └── Custom Resources (user-defined)
│   └── Branch Resources (branch)
└── Custom Namespaces (user-defined)
```

### Lock States

```go
type LockState string

const (
    LockStateAcquired LockState = "acquired"  // Lock is held
    LockStatePending  LockState = "pending"   // Request is queued
    LockStateExpired  LockState = "expired"   // Lock has expired
    LockStateReleased LockState = "released"  // Lock was released
)
```

### Enhanced Lock Structure

```go
type EnhancedLock struct {
    ID          string             `json:"id"`
    Resource    ResourceIdentifier `json:"resource"`
    State       LockState          `json:"state"`
    Priority    Priority           `json:"priority"`
    Owner       string             `json:"owner"`
    AcquiredAt  time.Time          `json:"acquired_at"`
    ExpiresAt   *time.Time         `json:"expires_at,omitempty"`
    Metadata    map[string]string  `json:"metadata,omitempty"`
    Version     int64              `json:"version"` // Optimistic locking
}
```

## System Diagrams

### Lock Lifecycle Diagram

```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│   Lock Request  │───▶│  Priority Check  │───▶│  Resource Check │
└─────────────────┘    └──────────────────┘    └─────────────────┘
                                                         │
                       ┌──────────────────┐              ▼
                       │    Queue         │    ┌─────────────────┐
                   ┌───│   (if blocked)   │◀───│  Deadlock Check │
                   │   └──────────────────┘    └─────────────────┘
                   ▼                                      │
          ┌─────────────────┐                             ▼
          │  Wait in Queue  │                   ┌─────────────────┐
          │  (by priority)  │                   │  Acquire Lock   │
          └─────────────────┘                   └─────────────────┘
                   │                                      │
                   ▼                                      ▼
          ┌─────────────────┐                   ┌─────────────────┐
          │ Timeout/Cancel  │                   │   Lock Held     │
          └─────────────────┘                   └─────────────────┘
                   │                                      │
                   ▼                                      ▼
          ┌─────────────────┐                   ┌─────────────────┐
          │    Cleanup      │◀──────────────────│  Release Lock   │
          └─────────────────┘                   └─────────────────┘
```

### Queue Priority Management

```
Resource A Queue:
┌─────────────────────────────────────────────────────────┐
│  Critical (3) │  High (2)  │  Normal (1) │   Low (0)   │
│  ┌─────────┐   │  ┌───────┐ │  ┌───────┐  │  ┌───────┐  │
│  │Request1 │   │  │Req2   │ │  │Req3   │  │  │Req4   │  │
│  │t1       │   │  │t2     │ │  │t3     │  │  │t4     │  │
│  └─────────┘   │  └───────┘ │  └───────┘  │  └───────┘  │
└─────────────────────────────────────────────────────────┘
Processing Order: Req1 → Req2 → Req3 → Req4
```

### Backend Storage Architecture

```
┌─────────────────────────────────────────────────────────┐
│                    Enhanced Lock Manager                │
├─────────────────────────────────────────────────────────┤
│                     Backend Interface                   │
├────────────────────┬────────────────────────────────────┤
│    Redis Backend   │         BoltDB Backend             │
│  ┌──────────────┐  │       ┌──────────────────┐         │
│  │ Redis Cluster│  │       │   Local Storage  │         │
│  │              │  │       │                  │         │
│  │ ┌──────────┐ │  │       │ ┌──────────────┐ │         │
│  │ │ Node 1   │ │  │       │ │   BoltDB     │ │         │
│  │ │ Node 2   │ │  │       │ │   File       │ │         │
│  │ │ Node 3   │ │  │       │ └──────────────┘ │         │
│  │ └──────────┘ │  │       └──────────────────┘         │
│  │              │  │                                    │
│  │ Pub/Sub      │  │       Transaction Log              │
│  │ Lua Scripts  │  │       ACID Guarantees              │
│  └──────────────┘  │                                    │
└────────────────────┴────────────────────────────────────┘
```

## Technical Specifications

### Configuration Options

```go
type EnhancedConfig struct {
    // Core configuration
    Enabled                bool          `mapstructure:"enabled"`
    Backend                string        `mapstructure:"backend"`
    DefaultTimeout         time.Duration `mapstructure:"default_timeout"`
    MaxTimeout             time.Duration `mapstructure:"max_timeout"`

    // Queue configuration
    EnablePriorityQueue    bool          `mapstructure:"enable_priority_queue"`
    MaxQueueSize          int           `mapstructure:"max_queue_size"`
    QueueTimeout          time.Duration `mapstructure:"queue_timeout"`

    // Retry configuration
    EnableRetry           bool          `mapstructure:"enable_retry"`
    MaxRetryAttempts      int           `mapstructure:"max_retry_attempts"`
    RetryBaseDelay        time.Duration `mapstructure:"retry_base_delay"`
    RetryMaxDelay         time.Duration `mapstructure:"retry_max_delay"`

    // Deadlock detection
    EnableDeadlockDetection bool        `mapstructure:"enable_deadlock_detection"`
    DeadlockCheckInterval   time.Duration `mapstructure:"deadlock_check_interval"`

    // Redis specific
    RedisClusterMode      bool          `mapstructure:"redis_cluster_mode"`
    RedisKeyPrefix        string        `mapstructure:"redis_key_prefix"`
    RedisLockTTL          time.Duration `mapstructure:"redis_lock_ttl"`

    // Backward compatibility
    LegacyFallback        bool          `mapstructure:"legacy_fallback"`
    PreserveLegacyFormat  bool          `mapstructure:"preserve_legacy_format"`
}
```

### Default Configuration

```go
func DefaultConfig() *EnhancedConfig {
    return &EnhancedConfig{
        Enabled:                 false, // Opt-in for backward compatibility
        Backend:                 "boltdb",
        DefaultTimeout:          30 * time.Minute,
        MaxTimeout:              2 * time.Hour,
        EnablePriorityQueue:     false,
        MaxQueueSize:           1000,
        QueueTimeout:           10 * time.Minute,
        EnableRetry:            false,
        MaxRetryAttempts:       3,
        RetryBaseDelay:         time.Second,
        RetryMaxDelay:          30 * time.Second,
        EnableDeadlockDetection: false,
        DeadlockCheckInterval:   30 * time.Second,
        RedisKeyPrefix:         "atlantis:enhanced:lock:",
        RedisLockTTL:           time.Hour,
        LegacyFallback:         true,
        PreserveLegacyFormat:   true,
    }
}
```

### Performance Characteristics

**Lock Acquisition Latency:**
- Local (BoltDB): < 1ms
- Redis Single Node: 1-5ms
- Redis Cluster: 5-15ms
- With Queuing: +10-50ms (depending on queue depth)

**Throughput:**
- BoltDB: ~1,000 ops/sec
- Redis Single: ~10,000 ops/sec
- Redis Cluster: ~50,000 ops/sec (distributed)

**Memory Usage:**
- Base Manager: ~10MB
- Per Active Lock: ~1KB
- Queue Overhead: ~500B per queued request
- Deadlock Detector: ~100KB + (nodes × edges × 50B)

### Error Handling

**Error Types:**
```go
const (
    ErrCodeLockExists       = "LOCK_EXISTS"
    ErrCodeLockNotFound     = "LOCK_NOT_FOUND"
    ErrCodeLockExpired      = "LOCK_EXPIRED"
    ErrCodeTimeout          = "TIMEOUT"
    ErrCodeQueueFull        = "QUEUE_FULL"
    ErrCodeDeadlock         = "DEADLOCK"
    ErrCodeBackendError     = "BACKEND_ERROR"
    ErrCodeInvalidRequest   = "INVALID_REQUEST"
    ErrCodePermissionDenied = "PERMISSION_DENIED"
)
```

**Retry Policies:**
- Exponential backoff with jitter
- Configurable max attempts (default: 3)
- Circuit breaker integration
- Error-specific retry decisions

## Migration Guide

### Phase 1: Enable Enhanced System

```yaml
# atlantis.yaml
enhanced_locking:
  enabled: true
  backend: "boltdb"  # Start with existing backend
  legacy_fallback: true
  preserve_legacy_format: true
```

### Phase 2: Enable Advanced Features

```yaml
enhanced_locking:
  enabled: true
  backend: "boltdb"
  enable_priority_queue: true
  enable_retry: true
  enable_deadlock_detection: true
  legacy_fallback: true
```

### Phase 3: Migrate to Distributed Backend

```yaml
enhanced_locking:
  enabled: true
  backend: "redis"
  redis_cluster_mode: true
  redis_key_prefix: "atlantis:enhanced:lock:"
  legacy_fallback: false  # Pure enhanced mode
```

### Compatibility Matrix

| Feature | Legacy Mode | Enhanced Mode | Notes |
|---------|-------------|---------------|-------|
| Basic Lock/Unlock | ✅ | ✅ | Full compatibility |
| Lock Listing | ✅ | ✅ | Format preserved |
| UnlockByPull | ✅ | ✅ | Behavior identical |
| Priority Locks | ❌ | ✅ | Enhanced only |
| Queue Management | ❌ | ✅ | Enhanced only |
| Deadlock Detection | ❌ | ✅ | Enhanced only |
| Distributed Storage | ❌ | ✅ | Enhanced only |

### Testing Strategy

**Unit Tests:**
- Component isolation testing
- Mock backend implementations
- Edge case validation
- Performance benchmarks

**Integration Tests:**
- End-to-end workflows
- Backend compatibility
- Migration scenarios
- Failure recovery

**Load Tests:**
- Concurrent access patterns
- Queue depth management
- Deadlock scenarios
- Performance degradation

### Monitoring and Observability

**Key Metrics:**
- Lock acquisition rate and latency
- Queue depth and wait times
- Deadlock detection and resolution
- Backend health and performance
- Error rates by type

**Health Checks:**
- Backend connectivity
- Queue status
- Deadlock detector status
- Memory usage
- Active lock counts

**Alerts:**
- High queue depths
- Deadlock detection failures
- Backend connectivity issues
- Memory pressure
- Unusual error rates

This enhanced locking system provides a robust foundation for enterprise-scale Terraform automation while maintaining the simplicity and reliability that makes Atlantis effective.