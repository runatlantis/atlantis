# Enhanced Plan Queue and Locking System

This document describes the enhanced plan queue and locking system that addresses the long-standing issues with Atlantis locking and workspace management.

## Overview

The enhanced system provides:

1. **Plan Queue Functionality** - Queues plan requests when locks are unavailable
2. **Lock Retry Logic** - Automatically retries lock acquisition with configurable delays
3. **Race Condition Prevention** - In-memory locks prevent concurrent operations on the same project/workspace
4. **Working Directory Protection** - Prevents premature deletion of working directories
5. **Automatic Lock Transfer** - Transfers locks to the next person in queue when available

## Features

### 1. Plan Queue System

When a plan request comes in and the project is already locked by another PR, instead of immediately failing, the PR gets queued. The system will:

-  Add the request to a queue for that specific project/workspace
-  Notify the user about their position in the queue
-  Automatically transfer the lock to the next person when the current lock is released
-  Notify users when it's their turn to plan

**Configuration:**

```bash
--enable-plan-queue=true
```

### 2. Lock Retry Logic

Instead of immediately failing when a lock is busy, the system can automatically retry lock acquisition with configurable delays.

**Configuration:**

```bash
--enable-lock-retry=true
--lock-retry-max-attempts=3
--lock-retry-delay=5
```

### 3. Race Condition Prevention

The system uses in-memory locks to prevent race conditions between concurrent operations on the same project/workspace. This addresses issues like:

-  Multiple plan requests for the same project/workspace
-  Race conditions between post-workflow hooks and automerge pull cleanup
-  Lock creation for no apparent reason

### 4. Working Directory Protection

The system protects working directories from premature deletion by:

-  Tracking which working directories are in use
-  Preventing deletion while operations are in progress
-  Automatic cleanup when operations complete

### 5. Automatic Lock Transfer

When a lock is released, the system automatically:

-  Checks if there are queued requests for that project/workspace
-  Transfers the lock to the next person in the queue
-  Notifies the user that they now have the lock

## Configuration

### Command Line Flags

```bash
# Enable plan queue functionality
--enable-plan-queue=true

# Enable lock retry functionality
--enable-lock-retry=true

# Maximum number of retry attempts (default: 3)
--lock-retry-max-attempts=3

# Delay between retry attempts in seconds (default: 5)
--lock-retry-delay=5
```

### Environment Variables

```bash
# Enable plan queue functionality
ATLANTIS_ENABLE_PLAN_QUEUE=true

# Enable lock retry functionality
ATLANTIS_ENABLE_LOCK_RETRY=true

# Maximum number of retry attempts
ATLANTIS_LOCK_RETRY_MAX_ATTEMPTS=3

# Delay between retry attempts in seconds
ATLANTIS_LOCK_RETRY_DELAY=5
```

## Architecture

### Components

1. **EnhancedLockingSystem** - Core locking system with retry and queue support
2. **PlanQueueManager** - Manages plan queues for projects/workspaces
3. **EnhancedProjectLocker** - Enhanced project locker with queue integration
4. **Memory Locks** - In-memory locks to prevent race conditions
5. **Working Directory Protection** - Protects working directories from deletion

### Data Flow

1. **Plan Request** → EnhancedProjectLocker
2. **Lock Check** → EnhancedLockingSystem
3. **If Lock Available** → Acquire lock and proceed
4. **If Lock Busy** → Add to queue or retry (based on configuration)
5. **Lock Release** → Transfer to next person in queue
6. **Working Directory** → Protected during operations

## Benefits

### For Users

-  **No More Manual Retries** - System automatically retries lock acquisition
-  **Queue Awareness** - Users know their position in the queue
-  **Automatic Notifications** - Users are notified when it's their turn
-  **Reduced Interruptions** - Fewer failed plan requests due to busy locks

### For Operators

-  **Reduced Support Load** - Fewer issues with locks and workspaces
-  **Better Resource Utilization** - Queues ensure efficient use of resources
-  **Improved Reliability** - Race conditions and workspace issues are prevented
-  **Better Monitoring** - Queue status and lock transfers are logged

### For the System

-  **Improved Stability** - Race conditions are eliminated
-  **Better Resource Management** - Working directories are properly protected
-  **Scalability** - Queue system handles high concurrency better
-  **Maintainability** - Cleaner separation of concerns

## Migration

### From Default Locking

The enhanced system is backward compatible. To migrate:

1. **Enable features gradually** - Start with lock retry, then add queue functionality
2. **Monitor logs** - Watch for any issues during migration
3. **Adjust configuration** - Tune retry attempts and delays based on your environment

### Configuration Examples

**Conservative Migration:**

```bash
--enable-lock-retry=true
--lock-retry-max-attempts=2
--lock-retry-delay=10
--enable-plan-queue=false
```

**Full Feature Set:**

```bash
--enable-lock-retry=true
--lock-retry-max-attempts=3
--lock-retry-delay=5
--enable-plan-queue=true
```

## Troubleshooting

### Common Issues

1. **Queue Not Working**

   -  Check if `--enable-plan-queue=true` is set
   -  Verify queue manager is properly initialized
   -  Check logs for queue-related errors

2. **Retry Not Working**

   -  Check if `--enable-lock-retry=true` is set
   -  Verify retry configuration values
   -  Check logs for retry attempts

3. **Working Directory Issues**
   -  Check if working directory protection is enabled
   -  Verify cleanup is happening properly
   -  Check logs for protection-related messages

### Log Messages

The system provides detailed logging for:

-  Queue operations (add, remove, transfer)
-  Retry attempts and results
-  Working directory protection
-  Lock transfers
-  Memory lock operations

### Monitoring

Key metrics to monitor:

-  Queue length per project/workspace
-  Retry success/failure rates
-  Lock transfer success rates
-  Working directory protection status

## Future Enhancements

Potential future improvements:

1. **Priority Queues** - Allow users to set priority for their requests
2. **Queue Timeouts** - Automatically remove stale queue entries
3. **Queue Persistence** - Store queues in backend for persistence across restarts
4. **Advanced Notifications** - Slack/email notifications for queue updates
5. **Queue Analytics** - Metrics and dashboards for queue performance

## References

-  [ADR #3345 - Project Locks](https://github.com/runatlantis/atlantis/pull/3345)
-  [PR #4997 - Lock Retry Logic](https://github.com/runatlantis/atlantis/pull/4997)
-  [Issue #1914 - Workspace lock creation](https://github.com/runatlantis/atlantis/issues/1914)
-  [Issue #2200 - Lock creation for no apparent reason](https://github.com/runatlantis/atlantis/issues/2200)
-  [Issue #3336 - Race condition between hooks and cleanup](https://github.com/runatlantis/atlantis/issues/3336)
