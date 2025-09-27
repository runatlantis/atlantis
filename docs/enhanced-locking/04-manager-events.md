# Enhanced Locking Manager and Events System

**Status:** Complete #5843 - Enhanced Manager and Events
**Dependencies:** #5842 (Foundation), #5836 (Compatibility), #5840 (Redis)
**Size:** ~350-400 lines total implementation

## Overview

This document describes the Enhanced Locking Manager and Events System, which provides centralized orchestration and comprehensive event tracking for the enhanced locking system.

## Components

### 1. Enhanced Lock Manager (`manager.go`)

The `EnhancedLockManager` serves as the central orchestrator for all locking operations, integrating priority queuing, deadlock detection, timeout management, and event emission.

#### Key Features

- **Unified Interface**: Implements the standard `LockManager` interface while providing enhanced capabilities
- **Component Integration**: Coordinates between backends, queues, timeout managers, and deadlock detectors
- **Event-Driven Architecture**: Emits comprehensive events for all lock lifecycle stages
- **Performance Monitoring**: Integrated metrics collection and health monitoring
- **Graceful Lifecycle**: Proper startup and shutdown sequences with worker management

### 2. Event Manager (`events.go`)

The `EventManager` provides comprehensive event tracking and subscription capabilities for the locking system.

#### Event Types

- **Lock Lifecycle Events**:
  - `lock_requested`: Lock acquisition initiated
  - `lock_acquired`: Lock successfully obtained
  - `lock_released`: Lock released by owner
  - `lock_expired`: Lock automatically expired
  - `lock_failed`: Lock acquisition failed

- **Queue Events**:
  - `lock_queued`: Request added to priority queue
  - `queued_lock_acquired`: Queued request processed successfully
  - `queue_timeout`: Queued request timed out

- **System Events**:
  - `deadlock_detected`: Deadlock situation identified
  - `deadlock_resolved`: Deadlock automatically resolved
  - `system_maintenance`: Maintenance operations performed
  - `health_change`: System health status changed

### 3. Metrics Collector (`metrics.go`)

The `MetricsCollector` provides comprehensive performance monitoring and analysis capabilities.

#### Metric Categories

- **Core Metrics**: Total requests, acquisitions, failures, releases
- **Performance Metrics**: Wait times (min, max, average), Hold times (min, max, average)
- **Priority-Based Metrics**: Per-priority performance tracking
- **Health Scoring**: System health score (0-100) based on error rates and performance

## Integration Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                Enhanced Lock Manager                        │
├─────────────────────────────────────────────────────────────┤
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────┐ │
│  │   Backend   │  │    Queue    │  │   Deadlock Detector │ │
│  │ Integration │  │  Manager    │  │                     │ │
│  └─────────────┘  └─────────────┘  └─────────────────────┘ │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────┐ │
│  │   Timeout   │  │   Retry     │  │    Event Manager    │ │
│  │  Manager    │  │  Manager    │  │                     │ │
│  └─────────────┘  └─────────────┘  └─────────────────────┘ │
│  ┌─────────────────────────────────────────────────────────┐ │
│  │               Metrics Collector                        │ │
│  └─────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────┘
```

## Usage Examples

### Basic Usage

```go
// Initialize manager with backend and config
manager := NewEnhancedLockManager(backend, config, logger)

// Start the manager
err := manager.Start(ctx)
if err != nil {
    return fmt.Errorf("failed to start manager: %w", err)
}
defer manager.Stop(ctx)

// Use standard interface
lock, err := manager.Lock(ctx, project, workspace, user)
```

### Advanced Usage with Events

```go
// Create event subscriber
subscriber := NewLoggingEventSubscriber("main", logger)

// Subscribe to events (if event manager is available)
if manager.eventManager != nil {
    manager.eventManager.Subscribe(subscriber)
}

// Use enhanced features
lock, err := manager.LockWithPriority(ctx, project, workspace, user, PriorityHigh)
```

## Dependencies

This implementation depends on:

1. **#5842**: Enhanced locking foundation and types
2. **#5836**: Backward compatibility adapter
3. **#5840**: Redis backend implementation

## File Structure

```
server/core/locking/enhanced/
├── manager.go          # ~370 lines - Central orchestration (enhanced)
├── events.go           # ~200 lines - Event system
├── metrics.go          # ~150 lines - Metrics collection
└── docs/enhanced-locking/
    └── 04-manager-events.md    # This documentation
```

## Summary

This PR provides the central orchestration layer for the enhanced locking system, with comprehensive event tracking and metrics collection. The manager integrates all components while maintaining backward compatibility and providing enhanced operational capabilities.