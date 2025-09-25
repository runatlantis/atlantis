# Atlantis Enhanced Locking Migration Guide

This comprehensive guide covers migrating from the legacy Atlantis locking system to the enhanced system, including testing strategies, rollback procedures, and production deployment best practices.

## Table of Contents

- [Migration Overview](#migration-overview)
- [Pre-Migration Assessment](#pre-migration-assessment)
- [Migration Phases](#migration-phases)
- [Configuration Migration](#configuration-migration)
- [Testing and Validation](#testing-and-validation)
- [Rollback Procedures](#rollback-procedures)
- [Monitoring and Observability](#monitoring-and-observability)
- [Troubleshooting Common Issues](#troubleshooting-common-issues)
- [Performance Optimization](#performance-optimization)

## Migration Overview

### Migration Strategy

The enhanced locking system migration follows a **zero-downtime, gradual rollout** approach:

```
Legacy Only → Shadow Mode → Hybrid Mode → Enhanced Only
    ↓             ↓            ↓             ↓
 Week 1-2      Week 3-4     Week 5-6     Week 7+
```

### Key Benefits of Enhanced System

| Feature | Legacy System | Enhanced System |
|---------|---------------|----------------|
| **Throughput** | ~1,000 ops/sec | 50,000+ ops/sec |
| **Latency** | ~10ms | <1ms |
| **Scaling** | Single instance | Redis Cluster |
| **Queuing** | None | Priority-based |
| **Deadlock Detection** | None | Automatic |
| **Fault Tolerance** | Basic | Circuit breakers |
| **Observability** | Limited | Comprehensive |

### Migration Prerequisites

- [ ] Atlantis version 0.19.0+
- [ ] Redis server/cluster available
- [ ] Monitoring infrastructure ready
- [ ] Backup and rollback procedures tested
- [ ] Team training completed

## Pre-Migration Assessment

### 1. Current System Analysis

Run the assessment tool to analyze your current locking patterns:

```bash
# Analyze current lock usage patterns
atlantis locking analyze --period=30d --output=json > lock-analysis.json

# Review common lock contention points
atlantis locking contention --top=20
```

### 2. Environment Inventory

Document your current setup:

```yaml
# assessment-config.yaml
assessment:
  atlantis_version: "0.28.5"
  backend_type: "boltdb"  # or "redis"
  concurrent_users: 25
  repositories_count: 150
  average_locks_per_hour: 500
  peak_locks_per_hour: 2000
  longest_held_lock: "45m"
  lock_contention_rate: "5%"
```

### 3. Resource Requirements Planning

Based on your assessment, plan resources:

```yaml
# resource-planning.yaml
redis_cluster:
  # Small deployment (< 1000 locks/hour)
  small:
    instances: 1
    memory: "2GB"
    cpu: "1 core"

  # Medium deployment (1000-10000 locks/hour)
  medium:
    instances: 3
    memory: "4GB each"
    cpu: "2 cores each"

  # Large deployment (10000+ locks/hour)
  large:
    instances: 6
    memory: "8GB each"
    cpu: "4 cores each"
    replication: true
```

## Migration Phases

### Phase 1: Preparation and Testing (Week 1-2)

#### 1.1 Deploy Enhanced System (Disabled)

```yaml
# server.yaml - Initial deployment
locking-db-type: boltdb  # Keep existing backend

# Add enhanced locking configuration (disabled)
enhanced-locking:
  enabled: false  # CRITICAL: Start disabled
  backend: redis
  legacy-fallback: true
  preserve-legacy-format: true

  # Redis configuration
  redis:
    addr: "localhost:6379"
    password: "${REDIS_PASSWORD}"
    db: 0
    cluster-mode: false

  # Conservative initial settings
  default-timeout: "30m"
  max-timeout: "2h"
  enable-priority-queue: false
  enable-retry: false
  enable-deadlock-detection: false
```

#### 1.2 Setup Monitoring

```yaml
# monitoring.yaml
monitoring:
  metrics:
    enabled: true
    endpoint: "/metrics"

  alerts:
    - name: "locking-system-health"
      condition: "lock_backend_health == 0"
      severity: "critical"

    - name: "high-lock-contention"
      condition: "lock_contention_rate > 0.1"
      severity: "warning"

    - name: "lock-timeout-rate"
      condition: "lock_timeout_rate > 0.05"
      severity: "warning"
```

#### 1.3 Run Compatibility Tests

```bash
# Test compatibility with existing locks
atlantis enhanced-locking test-compatibility \
  --legacy-backend=boltdb \
  --test-duration=1h \
  --concurrent-operations=100

# Expected output:
# ✓ Legacy lock format compatibility: PASS
# ✓ Key generation compatibility: PASS
# ✓ Concurrent access patterns: PASS
# ✓ Unlock behavior consistency: PASS
# ✓ List operations compatibility: PASS
```

### Phase 2: Shadow Mode (Week 3-4)

#### 2.1 Enable Enhanced System (Shadow Mode)

```yaml
# server.yaml - Enable shadow mode
enhanced-locking:
  enabled: true              # Enable enhanced system
  legacy-fallback: true      # CRITICAL: Keep fallback enabled
  monitor-only: true          # Shadow mode - don't process locks

  # Enable basic monitoring
  enable-events: true
  event-buffer-size: 1000
```

#### 2.2 Monitor Shadow Mode

```bash
# Monitor enhanced system health while legacy handles traffic
atlantis enhanced-locking status --watch

# Expected metrics:
# Enhanced Backend Health: OK
# Legacy Backend Health: OK
# Traffic Split: Legacy 100%, Enhanced 0%
# Compatibility Score: 100%
```

#### 2.3 Validate Data Consistency

```bash
# Compare lock states between systems
atlantis enhanced-locking compare-backends \
  --legacy=boltdb \
  --enhanced=redis \
  --check-interval=10m

# Monitor for discrepancies
atlantis enhanced-locking validate-consistency --duration=24h
```

### Phase 3: Hybrid Mode - Gradual Migration (Week 5-6)

#### 3.1 Enable Gradual Traffic Split

```yaml
# server.yaml - Start gradual migration
enhanced-locking:
  enabled: true
  legacy-fallback: true
  monitor-only: false         # Disable shadow mode
  traffic-split: 10          # Start with 10% traffic

  # Begin enabling enhanced features
  enable-priority-queue: false  # Still conservative
  enable-retry: false
  enable-deadlock-detection: false
```

#### 3.2 Weekly Traffic Increases

Week 5: 10% Enhanced, 90% Legacy
```bash
atlantis config set enhanced-locking.traffic-split 10
```

Week 6: 50% Enhanced, 50% Legacy
```bash
atlantis config set enhanced-locking.traffic-split 50
```

Week 7: 90% Enhanced, 10% Legacy
```bash
atlantis config set enhanced-locking.traffic-split 90
```

#### 3.3 Monitor Performance During Migration

```bash
# Monitor key metrics during traffic increase
atlantis metrics watch \
  --metrics="lock_acquisition_time,lock_success_rate,queue_depth" \
  --alert-thresholds="acquisition_time>100ms,success_rate<0.95"
```

### Phase 4: Full Enhanced System (Week 7+)

#### 4.1 Complete Migration

```yaml
# server.yaml - Full enhanced system
enhanced-locking:
  enabled: true
  legacy-fallback: false      # Disable fallback
  traffic-split: 100         # 100% enhanced traffic

  # Enable advanced features gradually
  enable-priority-queue: true
  max-queue-size: 1000
  queue-timeout: "10m"

  enable-retry: true
  max-retry-attempts: 3
  retry-base-delay: "1s"

  enable-deadlock-detection: true
  deadlock-check-interval: "30s"
```

#### 4.2 Remove Legacy Components

```yaml
# server.yaml - Clean configuration
locking-db-type: enhanced-redis  # Remove old backend config

# Remove legacy backend entirely (after validation period)
# enhanced-locking.legacy-fallback: false  # Already disabled
```

## Configuration Migration

### Legacy BoltDB to Enhanced Redis

#### Before (Legacy BoltDB):
```yaml
# legacy-config.yaml
locking-db-type: boltdb
boltdb:
  path: "/var/lib/atlantis/atlantis.db"
  bucket: "runLocks"
```

#### After (Enhanced Redis):
```yaml
# enhanced-config.yaml
locking-db-type: enhanced-redis

enhanced-locking:
  enabled: true
  backend: redis

  redis:
    addr: "redis-cluster:6379"
    password: "${REDIS_PASSWORD}"
    db: 0
    cluster-mode: true
    key-prefix: "atlantis:lock:"
    lock-ttl: "1h"

  default-timeout: "30m"
  max-timeout: "2h"

  # Advanced features
  enable-priority-queue: true
  enable-retry: true
  enable-deadlock-detection: true
  enable-events: true
```

### Legacy Redis to Enhanced Redis

#### Before (Legacy Redis):
```yaml
# legacy-redis.yaml
locking-db-type: redis
redis:
  host: "localhost"
  port: 6379
  password: "${REDIS_PASSWORD}"
  db: 0
```

#### After (Enhanced Redis):
```yaml
# enhanced-redis.yaml
locking-db-type: enhanced-redis

enhanced-locking:
  enabled: true
  backend: redis

  redis:
    addr: "localhost:6379"
    password: "${REDIS_PASSWORD}"
    db: 1  # Use different DB to avoid conflicts
    cluster-mode: false
    key-prefix: "atlantis:enhanced:"

  # Reuse existing Redis instance
  redis-migration:
    migrate-existing-locks: true
    legacy-key-pattern: "pr:*"
    cleanup-legacy-keys: false  # Keep for rollback
```

### atlantis.yaml Integration Updates

#### Legacy atlantis.yaml:
```yaml
# legacy-atlantis.yaml
version: 3
projects:
  - name: infrastructure
    dir: .
    workspace: production
    # Basic locking (no advanced options)
```

#### Enhanced atlantis.yaml:
```yaml
# enhanced-atlantis.yaml
version: 3
projects:
  - name: infrastructure
    dir: .
    workspace: production
    repo_locks:
      mode: on_apply
      priority: high          # Enhanced: priority levels
      timeout: "45m"          # Enhanced: custom timeouts
      resource_isolation: strict  # Enhanced: isolation levels

  - name: development
    dir: .
    workspace: dev
    repo_locks:
      mode: on_plan
      priority: normal
      timeout: "15m"
      resource_isolation: standard

  # Enhanced: Complex dependencies
  - name: database
    dir: database
    workspace: production
    execution_order_group: 1
    repo_locks:
      mode: on_apply
      priority: critical

  - name: applications
    dir: apps
    workspace: production
    execution_order_group: 2
    depends_on:
      - database
    repo_locks:
      mode: on_apply
      priority: high
```

## Testing and Validation

### 1. Automated Testing Suite

```bash
# Run comprehensive test suite
atlantis enhanced-locking test-suite \
  --include=compatibility,performance,reliability \
  --duration=2h \
  --concurrent-users=50

# Test specific scenarios
atlantis enhanced-locking test-scenario \
  --scenario=high-contention \
  --concurrent-locks=100 \
  --duration=30m

atlantis enhanced-locking test-scenario \
  --scenario=queue-fairness \
  --priority-mix="critical:10%,high:20%,normal:60%,low:10%" \
  --duration=15m
```

### 2. Load Testing

```yaml
# load-test-config.yaml
load_test:
  name: "enhanced-locking-load-test"
  duration: "1h"

  scenarios:
    - name: "normal-operations"
      weight: 70
      operations:
        - type: "plan"
          frequency: "30/min"
          timeout: "5m"
        - type: "apply"
          frequency: "10/min"
          timeout: "15m"

    - name: "high-priority-emergency"
      weight: 10
      operations:
        - type: "apply"
          frequency: "5/min"
          timeout: "30m"
          priority: "critical"

    - name: "concurrent-development"
      weight: 20
      operations:
        - type: "plan"
          frequency: "60/min"
          timeout: "2m"
          priority: "normal"
```

### 3. Compatibility Validation

```bash
# Validate lock format compatibility
atlantis enhanced-locking validate-formats \
  --sample-size=1000 \
  --legacy-backend=boltdb \
  --enhanced-backend=redis

# Test migration scenarios
atlantis enhanced-locking test-migration \
  --source=boltdb \
  --target=enhanced-redis \
  --validate-data-integrity=true
```

## Rollback Procedures

### Emergency Rollback (< 5 minutes)

```bash
# Immediate rollback to legacy system
atlantis config set enhanced-locking.enabled false
atlantis config set enhanced-locking.legacy-fallback true

# Restart Atlantis to ensure clean state
atlantis restart --wait-for-healthy

# Verify rollback success
atlantis locking status
# Expected: "Backend: legacy, Status: healthy"
```

### Planned Rollback

#### Step 1: Prepare Rollback
```bash
# Export current enhanced locks for potential recovery
atlantis enhanced-locking export \
  --output=/backup/enhanced-locks-$(date +%Y%m%d).json

# Validate legacy backend health
atlantis locking validate --backend=legacy
```

#### Step 2: Gradual Traffic Reduction
```yaml
# Gradually reduce enhanced traffic
enhanced-locking:
  traffic-split: 50    # Reduce from 100% to 50%
  legacy-fallback: true
```

Wait 30 minutes, monitor for issues, then:
```yaml
enhanced-locking:
  traffic-split: 0     # All traffic to legacy
  legacy-fallback: true
```

#### Step 3: Complete Rollback
```yaml
# Final rollback configuration
enhanced-locking:
  enabled: false
  legacy-fallback: true

# Revert to original backend
locking-db-type: boltdb  # or redis
```

### Rollback Validation

```bash
# Validate rollback success
atlantis locking test \
  --backend=legacy \
  --operations=100 \
  --concurrent=10

# Check for any data inconsistencies
atlantis locking compare-state \
  --before=/backup/pre-migration-state.json \
  --after=current \
  --ignore-timestamps=true
```

## Monitoring and Observability

### Key Metrics to Monitor

#### System Health Metrics
```yaml
metrics:
  backend_health:
    description: "Backend health status (0-1)"
    type: gauge
    alerts:
      critical: "< 1"

  lock_success_rate:
    description: "Lock acquisition success rate"
    type: gauge
    alerts:
      critical: "< 0.95"
      warning: "< 0.98"

  average_lock_time:
    description: "Average lock acquisition time"
    type: histogram
    alerts:
      critical: "> 1000ms"
      warning: "> 100ms"
```

#### Performance Metrics
```yaml
metrics:
  lock_throughput:
    description: "Locks acquired per second"
    type: gauge
    target: "> 1000/sec"

  queue_depth:
    description: "Current queue depth"
    type: gauge
    alerts:
      critical: "> 1000"
      warning: "> 100"

  deadlocks_detected:
    description: "Number of deadlocks detected"
    type: counter
    alerts:
      warning: "> 10/hour"
```

### Monitoring Dashboard

```yaml
# grafana-dashboard.yaml
dashboard:
  title: "Atlantis Enhanced Locking"

  panels:
    - title: "System Health"
      metrics:
        - backend_health
        - lock_success_rate
        - error_rate

    - title: "Performance"
      metrics:
        - lock_throughput
        - average_lock_time
        - queue_depth

    - title: "Advanced Features"
      metrics:
        - priority_queue_usage
        - deadlocks_detected
        - circuit_breaker_state

    - title: "Resource Usage"
      metrics:
        - redis_memory_usage
        - redis_cpu_usage
        - connection_pool_usage
```

### Alerting Rules

```yaml
# alerting-rules.yaml
alerts:
  - name: "EnhancedLockingDown"
    condition: "backend_health == 0"
    severity: "critical"
    message: "Enhanced locking backend is unhealthy"
    actions:
      - page_oncall
      - trigger_rollback_runbook

  - name: "HighLockContention"
    condition: "queue_depth > 100"
    severity: "warning"
    message: "High lock contention detected"
    actions:
      - slack_devops_channel

  - name: "DeadlockSpike"
    condition: "rate(deadlocks_detected[5m]) > 10"
    severity: "warning"
    message: "Unusually high deadlock detection rate"
    actions:
      - investigate_automation
```

## Troubleshooting Common Issues

### Issue 1: Lock Acquisition Failures

**Symptoms:**
- Increased error rate in lock acquisition
- Users reporting "lock already exists" errors
- Queue depth increasing
- PR commands timing out with "failed to acquire lock" messages

**Investigation:**
```bash
# Check backend health
atlantis enhanced-locking health-check

# Examine failed lock attempts
atlantis enhanced-locking logs --filter="level=error,component=lock-acquisition" --lines=100

# Check for stuck locks
atlantis enhanced-locking list-locks --filter="state=expired" --verbose

# Identify lock contention hotspots
atlantis enhanced-locking analyze-contention --period=24h --top=10

# Check lock holder details
atlantis enhanced-locking list-locks --filter="held_duration>30m" --show-holders
```

**Common Causes & Solutions:**

1. **Expired but uncleaned locks:**
```bash
# Clean up expired locks immediately
atlantis enhanced-locking cleanup --expired-only --force

# Enable automatic cleanup (if not already enabled)
atlantis config set enhanced-locking.enable-automatic-cleanup true
atlantis config set enhanced-locking.cleanup-interval "5m"
```

2. **Backend connectivity issues:**
```bash
# Restart backend connection pool
atlantis enhanced-locking restart-backend --graceful

# Check Redis connectivity and health
redis-cli -h ${REDIS_HOST} -p ${REDIS_PORT} ping
redis-cli -h ${REDIS_HOST} -p ${REDIS_PORT} info replication
```

3. **Configuration issues:**
```bash
# Increase timeout values for heavy workloads
atlantis config set enhanced-locking.default-timeout "45m"
atlantis config set enhanced-locking.max-timeout "2h"

# Increase queue capacity
atlantis config set enhanced-locking.max-queue-size 2000
```

**Prevention:**
```yaml
# server.yaml - Proactive configuration
enhanced-locking:
  # Automatic cleanup prevents stuck locks
  enable-automatic-cleanup: true
  cleanup-interval: "3m"

  # Reasonable timeouts prevent indefinite holds
  default-timeout: "30m"
  max-timeout: "2h"

  # Health checks ensure backend reliability
  health-check-interval: "30s"

  # Circuit breaker prevents cascade failures
  enable-circuit-breaker: true
```

### Issue 2: Redis Connection Issues

**Symptoms:**
- Intermittent connection errors: "connection refused" or "timeout"
- Timeout errors in logs: "redis: connection pool timeout"
- Fallback to legacy system triggered frequently
- Connection pool exhaustion warnings

**Investigation:**
```bash
# Test Redis connectivity comprehensively
redis-cli -h ${REDIS_HOST} -p ${REDIS_PORT} ping
redis-cli -h ${REDIS_HOST} -p ${REDIS_PORT} info replication
redis-cli -h ${REDIS_HOST} -p ${REDIS_PORT} info memory
redis-cli -h ${REDIS_HOST} -p ${REDIS_PORT} info clients

# Check connection pool status
atlantis enhanced-locking redis-status --verbose

# Monitor connection patterns
atlantis enhanced-locking monitor-connections --duration=5m --interval=10s

# Check network connectivity
ping ${REDIS_HOST}
telnet ${REDIS_HOST} ${REDIS_PORT}
```

**Detailed Log Analysis:**
```bash
# Look for specific Redis error patterns
atlantis logs --filter="redis" --since="1h" | grep -E "(timeout|connection|pool)"

# Check for connection leaks
atlantis enhanced-locking connection-report --show-leaks
```

**Resolution Steps:**

1. **Connection Pool Optimization:**
```bash
# Increase connection pool size
atlantis config set enhanced-locking.redis.pool-size 50
atlantis config set enhanced-locking.redis.min-idle-conns 10

# Optimize pool timeouts
atlantis config set enhanced-locking.redis.pool-timeout "10s"
atlantis config set enhanced-locking.redis.idle-timeout "15m"
atlantis config set enhanced-locking.redis.max-conn-age "1h"
```

2. **Network Timeout Optimization:**
```bash
# Increase Redis operation timeouts
atlantis config set enhanced-locking.redis.read-timeout "10s"
atlantis config set enhanced-locking.redis.write-timeout "10s"
atlantis config set enhanced-locking.redis.dial-timeout "5s"

# Enable keep-alive and retry
atlantis config set enhanced-locking.redis.keep-alive true
atlantis config set enhanced-locking.redis.retry-on-failure true
atlantis config set enhanced-locking.redis.max-retries 3
```

3. **Redis Server Configuration:**
```bash
# Optimize Redis server settings
redis-cli config set tcp-keepalive 60
redis-cli config set timeout 300
redis-cli config set maxclients 10000

# Monitor Redis performance
redis-cli --latency-history -i 1
redis-cli --stat
```

**Advanced Troubleshooting:**

For Redis Cluster deployments:
```bash
# Check cluster health
redis-cli -c cluster info
redis-cli -c cluster nodes

# Test each cluster node
for node in redis-node-1 redis-node-2 redis-node-3; do
  echo "Testing $node:"
  redis-cli -h $node ping
  redis-cli -h $node info replication
done
```

**Prevention Configuration:**
```yaml
# server.yaml - Robust Redis configuration
enhanced-locking:
  redis:
    # Connection pool settings
    pool-size: 50
    min-idle-conns: 10
    max-conn-age: "1h"
    pool-timeout: "10s"
    idle-timeout: "15m"

    # Network timeouts
    read-timeout: "10s"
    write-timeout: "10s"
    dial-timeout: "5s"

    # Reliability features
    keep-alive: true
    retry-on-failure: true
    max-retries: 3

    # Health monitoring
    health-check-period: "30s"

  # Fallback configuration
  fallback-enabled: true
  fallback-threshold: 5  # failures before fallback
  fallback-cooldown: "2m"
```

### Issue 3: Priority Queue Not Working

**Symptoms:**
- High-priority requests waiting behind low-priority tasks
- Queue processing appears random or FIFO
- Critical production deployments delayed by development work
- Priority metrics showing incorrect distributions

**Investigation:**
```bash
# Check queue configuration and status
atlantis enhanced-locking queue-status --verbose

# Examine specific resource queue
atlantis enhanced-locking queue-inspect --resource="company/infra/./prod" --show-priorities

# Analyze priority distribution
atlantis enhanced-locking priority-analysis --period=24h

# Check atlantis.yaml priority assignments
atlantis config validate --check=priority-assignments --verbose

# Monitor queue processing order
atlantis enhanced-locking queue-monitor --resource="company/infra" --duration=10m
```

**Common Issues & Fixes:**

1. **Priority Queue Disabled:**
```bash
# Verify priority queue is enabled
atlantis config get enhanced-locking.enable-priority-queue
atlantis config set enhanced-locking.enable-priority-queue true

# Restart to ensure configuration is loaded
atlantis restart --wait-for-healthy
```

2. **Invalid Priority Assignments:**
```bash
# Check for invalid priority values in atlantis.yaml
atlantis config validate --check=priority-assignments --fix-invalid

# Common invalid values to fix:
# - priority: "high" (should be: priority: high)
# - priority: 5 (should be one of: critical, high, normal, low)
```

3. **Queue Configuration Issues:**
```bash
# Reset corrupted queue state
atlantis enhanced-locking queue-reset --resource="problematic/resource" --preserve-pending

# Rebuild priority indexes
atlantis enhanced-locking rebuild-priority-index --verify

# Clear and restart queue processing
atlantis enhanced-locking queue-restart --graceful
```

**Detailed Debugging:**
```bash
# Enable priority queue debugging
atlantis config set enhanced-locking.debug-priority-queue true

# Watch real-time queue operations
atlantis enhanced-locking queue-tail --filter="priority=critical" --follow

# Generate priority queue report
atlantis enhanced-locking priority-report --output=json > priority-debug.json
```

**Verification Steps:**
```bash
# Test priority ordering with sample operations
atlantis enhanced-locking test-priority --simulate-workload=mixed --duration=5m

# Expected output should show:
# Critical priority: processed first (0-2s avg wait)
# High priority: processed second (2-10s avg wait)
# Normal priority: processed third (10-30s avg wait)
# Low priority: processed last (30s+ avg wait)
```

**Prevention Configuration:**
```yaml
# server.yaml - Proper priority queue setup
enhanced-locking:
  enable-priority-queue: true

  # Queue processing configuration
  queue-processing:
    batch-size: 10
    processing-interval: "1s"
    priority-boost-factor: 2.0  # Multiplier for higher priorities

  # Priority timeout overrides
  priority-timeouts:
    critical: "60m"
    high: "45m"
    normal: "30m"
    low: "15m"

  # Queue monitoring
  enable-queue-metrics: true
  queue-stats-interval: "30s"
```

### Issue 4: Performance Degradation

**Symptoms:**
- Slower lock acquisition times (>5s instead of <1s)
- Increased CPU/memory usage on Atlantis server
- Users reporting operation timeouts
- Redis memory usage growing without cleanup
- High latency in lock operations

**Investigation:**

1. **System Performance Analysis:**
```bash
# Profile lock manager performance
atlantis enhanced-locking profile --duration=5m --include-stack-traces

# Check system resource usage
atlantis system-status --include-performance

# Monitor lock acquisition timing
atlantis enhanced-locking timing-analysis --period=1h --percentiles=50,90,95,99
```

2. **Redis Performance Analysis:**
```bash
# Check Redis performance metrics
redis-cli --latency-history -i 1
redis-cli --stat -i 1

# Analyze slow operations
redis-cli slowlog get 50
atlantis enhanced-locking slow-log --threshold=50ms --analyze

# Check Redis memory usage
redis-cli info memory
redis-cli memory stats
```

3. **Application-Level Analysis:**
```bash
# Examine lock operation patterns
atlantis enhanced-locking operation-stats --breakdown=by-operation --period=2h

# Check for lock leaks
atlantis enhanced-locking leak-detection --scan-depth=deep

# Analyze queue performance
atlantis enhanced-locking queue-performance --metrics=depth,wait-time,throughput
```

**Resolution Strategies:**

1. **Redis Optimization:**
```bash
# Optimize Redis memory usage
redis-cli config set maxmemory-policy allkeys-lru
redis-cli config set maxmemory 4gb

# Disable persistence for better performance (if acceptable)
redis-cli config set save ""
redis-cli config set appendonly no

# Optimize network settings
redis-cli config set tcp-keepalive 60
redis-cli config set timeout 300
```

2. **Connection Pool Tuning:**
```bash
# Scale connection pool based on load
atlantis config set enhanced-locking.redis.pool-size 100
atlantis config set enhanced-locking.redis.min-idle-conns 20

# Optimize connection lifecycle
atlantis config set enhanced-locking.redis.max-conn-age "30m"
atlantis config set enhanced-locking.redis.idle-check-frequency "1m"
```

3. **Application Configuration Tuning:**
```bash
# Optimize queue processing
atlantis config set enhanced-locking.max-queue-size 5000
atlantis config set enhanced-locking.queue-batch-size 50
atlantis config set enhanced-locking.queue-processing-interval "500ms"

# Improve cleanup efficiency
atlantis config set enhanced-locking.cleanup-batch-size 100
atlantis config set enhanced-locking.cleanup-interval "2m"
```

4. **Redis Scaling (if needed):**
```bash
# Scale to Redis cluster for high-load environments
atlantis enhanced-locking redis-scale --nodes=6 --replicas=1

# Configure read replicas for better performance
atlantis config set enhanced-locking.redis.read-preference "slave"
atlantis config set enhanced-locking.redis.route-by-latency true
```

**Performance Monitoring Setup:**
```yaml
# server.yaml - Performance monitoring configuration
enhanced-locking:
  # Comprehensive metrics collection
  enable-performance-metrics: true
  metrics-collection-interval: "10s"

  # Performance alerting thresholds
  performance-alerts:
    lock-acquisition-time:
      warning: "100ms"
      critical: "1000ms"

    queue-depth:
      warning: 50
      critical: 200

    memory-usage:
      warning: "1GB"
      critical: "2GB"

  # Automatic performance optimization
  auto-tuning:
    enabled: true
    adjustment-interval: "5m"
    max-pool-size: 200
    min-pool-size: 10
```

### Issue 5: Deadlock Detection and Resolution

**Symptoms:**
- Operations hanging indefinitely
- Circular waiting patterns in logs
- Multiple resources locked in conflicting order
- Deadlock detection alerts firing

**Investigation:**
```bash
# Check for active deadlocks
atlantis enhanced-locking deadlock-status --show-chains

# Analyze deadlock patterns
atlantis enhanced-locking deadlock-analysis --period=24h --show-patterns

# Examine lock dependency chains
atlantis enhanced-locking dependency-graph --resource="company/infra" --depth=5

# Check for circular dependencies in atlantis.yaml
atlantis config validate --check=circular-dependencies
```

**Resolution:**
```bash
# Break active deadlocks (use with caution)
atlantis enhanced-locking break-deadlock --deadlock-id="dl-123" --victim-strategy=youngest

# Enable automatic deadlock resolution
atlantis config set enhanced-locking.enable-automatic-deadlock-resolution true
atlantis config set enhanced-locking.deadlock-resolution-strategy "victim-selection"

# Adjust deadlock detection sensitivity
atlantis config set enhanced-locking.deadlock-check-interval "15s"
atlantis config set enhanced-locking.deadlock-timeout "2m"
```

### Issue 6: Memory Leaks and Resource Exhaustion

**Symptoms:**
- Atlantis server memory usage continuously growing
- Redis memory usage not stabilizing
- System becoming unresponsive over time
- "Out of memory" errors

**Investigation:**
```bash
# Monitor memory usage patterns
atlantis enhanced-locking memory-profile --duration=30m --interval=30s

# Check for lock leaks
atlantis enhanced-locking leak-detection --full-scan

# Analyze object retention
atlantis enhanced-locking memory-analysis --show-allocations

# Check Redis memory details
redis-cli memory doctor
redis-cli memory stats
```

**Resolution:**
```bash
# Force cleanup of stale objects
atlantis enhanced-locking cleanup --include-stale --force

# Adjust garbage collection settings
atlantis config set enhanced-locking.gc-interval "1m"
atlantis config set enhanced-locking.object-retention-period "10m"

# Configure Redis memory limits
redis-cli config set maxmemory 4gb
redis-cli config set maxmemory-policy allkeys-lru

# Restart with memory optimization
atlantis restart --optimize-memory
```

### Issue 7: Event System and Monitoring Failures

**Symptoms:**
- Missing event notifications
- Monitoring dashboards showing stale data
- Webhook delivery failures
- Metrics collection gaps

**Investigation:**
```bash
# Check event system health
atlantis enhanced-locking event-status --verbose

# Test event delivery
atlantis enhanced-locking test-events --include-webhooks

# Check metrics collection
atlantis metrics status --show-collectors

# Analyze event queue
atlantis enhanced-locking event-queue-status --show-pending
```

**Resolution:**
```bash
# Restart event system
atlantis enhanced-locking restart-events

# Clear event backlog
atlantis enhanced-locking event-queue-clear --preserve-recent=1h

# Reconfigure event delivery
atlantis config set enhanced-locking.event-buffer-size 10000
atlantis config set enhanced-locking.event-batch-size 100
```

## Monitoring and Alerting

### Key Metrics to Monitor

#### System Health Metrics
```yaml
metrics:
  backend_health:
    description: "Backend health status (0=unhealthy, 1=healthy)"
    type: gauge
    query: "atlantis_enhanced_locking_backend_health"
    alerts:
      critical: "< 1"
      duration: "30s"

  system_availability:
    description: "Overall system availability percentage"
    type: gauge
    query: "atlantis_enhanced_locking_availability"
    alerts:
      critical: "< 0.99"
      warning: "< 0.995"

  lock_success_rate:
    description: "Lock acquisition success rate"
    type: gauge
    query: "rate(atlantis_enhanced_locking_acquisitions_success_total[5m]) / rate(atlantis_enhanced_locking_acquisitions_total[5m])"
    alerts:
      critical: "< 0.95"
      warning: "< 0.98"
      duration: "5m"

  error_rate:
    description: "Overall error rate for locking operations"
    type: gauge
    query: "rate(atlantis_enhanced_locking_errors_total[5m])"
    alerts:
      critical: "> 0.05"  # 5% error rate
      warning: "> 0.02"   # 2% error rate
```

#### Performance Metrics
```yaml
metrics:
  lock_acquisition_latency:
    description: "Lock acquisition time percentiles"
    type: histogram
    query: "histogram_quantile(0.95, atlantis_enhanced_locking_acquisition_duration_seconds_bucket)"
    alerts:
      critical: "> 5"     # 5 seconds
      warning: "> 1"      # 1 second

  queue_depth:
    description: "Current queue depth across all resources"
    type: gauge
    query: "sum(atlantis_enhanced_locking_queue_depth)"
    alerts:
      critical: "> 1000"
      warning: "> 100"

  queue_wait_time:
    description: "Average time requests spend in queue"
    type: gauge
    query: "histogram_quantile(0.90, atlantis_enhanced_locking_queue_wait_duration_seconds_bucket)"
    alerts:
      critical: "> 600"   # 10 minutes
      warning: "> 300"    # 5 minutes

  throughput:
    description: "Operations per second"
    type: gauge
    query: "rate(atlantis_enhanced_locking_operations_total[1m])"
    target: "> 1000"  # Target throughput

  lock_hold_duration:
    description: "How long locks are typically held"
    type: histogram
    query: "histogram_quantile(0.95, atlantis_enhanced_locking_hold_duration_seconds_bucket)"
    alerts:
      warning: "> 3600"  # 1 hour
      critical: "> 7200" # 2 hours
```

#### Advanced Feature Metrics
```yaml
metrics:
  deadlocks_detected:
    description: "Number of deadlocks detected"
    type: counter
    query: "increase(atlantis_enhanced_locking_deadlocks_detected_total[1h])"
    alerts:
      warning: "> 5"      # More than 5 deadlocks per hour
      critical: "> 20"    # More than 20 deadlocks per hour

  priority_queue_effectiveness:
    description: "Priority queue working correctly"
    type: gauge
    query: "atlantis_enhanced_locking_priority_queue_inversion_ratio"
    alerts:
      warning: "> 0.1"    # 10% priority inversions
      critical: "> 0.2"   # 20% priority inversions

  circuit_breaker_state:
    description: "Circuit breaker state (0=closed, 1=open)"
    type: gauge
    query: "atlantis_enhanced_locking_circuit_breaker_state"
    alerts:
      warning: "== 1"     # Circuit breaker open

  retry_rate:
    description: "Operations requiring retries"
    type: gauge
    query: "rate(atlantis_enhanced_locking_retries_total[5m]) / rate(atlantis_enhanced_locking_operations_total[5m])"
    alerts:
      warning: "> 0.1"    # 10% of operations need retry
      critical: "> 0.3"   # 30% of operations need retry
```

#### Resource Usage Metrics
```yaml
metrics:
  redis_memory_usage:
    description: "Redis memory usage in bytes"
    type: gauge
    query: "atlantis_enhanced_locking_redis_memory_used_bytes"
    alerts:
      warning: "> 2000000000"   # 2GB
      critical: "> 4000000000"  # 4GB

  redis_cpu_usage:
    description: "Redis CPU usage percentage"
    type: gauge
    query: "atlantis_enhanced_locking_redis_cpu_percent"
    alerts:
      warning: "> 70"
      critical: "> 90"

  connection_pool_usage:
    description: "Connection pool utilization"
    type: gauge
    query: "atlantis_enhanced_locking_redis_pool_active_connections / atlantis_enhanced_locking_redis_pool_max_connections"
    alerts:
      warning: "> 0.8"     # 80% pool utilization
      critical: "> 0.95"   # 95% pool utilization

  atlantis_memory_usage:
    description: "Atlantis server memory usage"
    type: gauge
    query: "process_resident_memory_bytes{job='atlantis'}"
    alerts:
      warning: "> 2000000000"   # 2GB
      critical: "> 4000000000"  # 4GB
```

### Monitoring Dashboard

#### Grafana Dashboard Configuration
```json
{
  "dashboard": {
    "title": "Atlantis Enhanced Locking System",
    "tags": ["atlantis", "locking", "infrastructure"],
    "time": {"from": "now-1h", "to": "now"},
    "refresh": "30s",

    "panels": [
      {
        "title": "System Health Overview",
        "type": "stat",
        "targets": [
          {
            "expr": "atlantis_enhanced_locking_backend_health",
            "legendFormat": "Backend Health"
          },
          {
            "expr": "atlantis_enhanced_locking_availability",
            "legendFormat": "System Availability"
          }
        ]
      },

      {
        "title": "Lock Operations",
        "type": "timeseries",
        "targets": [
          {
            "expr": "rate(atlantis_enhanced_locking_acquisitions_success_total[5m])",
            "legendFormat": "Successful Acquisitions/sec"
          },
          {
            "expr": "rate(atlantis_enhanced_locking_acquisitions_failed_total[5m])",
            "legendFormat": "Failed Acquisitions/sec"
          },
          {
            "expr": "rate(atlantis_enhanced_locking_releases_total[5m])",
            "legendFormat": "Releases/sec"
          }
        ]
      },

      {
        "title": "Performance Metrics",
        "type": "timeseries",
        "targets": [
          {
            "expr": "histogram_quantile(0.50, atlantis_enhanced_locking_acquisition_duration_seconds_bucket)",
            "legendFormat": "50th percentile"
          },
          {
            "expr": "histogram_quantile(0.95, atlantis_enhanced_locking_acquisition_duration_seconds_bucket)",
            "legendFormat": "95th percentile"
          },
          {
            "expr": "histogram_quantile(0.99, atlantis_enhanced_locking_acquisition_duration_seconds_bucket)",
            "legendFormat": "99th percentile"
          }
        ]
      },

      {
        "title": "Queue Status",
        "type": "timeseries",
        "targets": [
          {
            "expr": "sum(atlantis_enhanced_locking_queue_depth) by (priority)",
            "legendFormat": "Queue Depth - {{priority}}"
          },
          {
            "expr": "histogram_quantile(0.90, atlantis_enhanced_locking_queue_wait_duration_seconds_bucket)",
            "legendFormat": "90th percentile wait time"
          }
        ]
      },

      {
        "title": "Resource Usage",
        "type": "timeseries",
        "targets": [
          {
            "expr": "atlantis_enhanced_locking_redis_memory_used_bytes / 1024 / 1024 / 1024",
            "legendFormat": "Redis Memory (GB)"
          },
          {
            "expr": "process_resident_memory_bytes{job='atlantis'} / 1024 / 1024 / 1024",
            "legendFormat": "Atlantis Memory (GB)"
          },
          {
            "expr": "atlantis_enhanced_locking_redis_pool_active_connections",
            "legendFormat": "Active Connections"
          }
        ]
      },

      {
        "title": "Advanced Features",
        "type": "timeseries",
        "targets": [
          {
            "expr": "increase(atlantis_enhanced_locking_deadlocks_detected_total[1h])",
            "legendFormat": "Deadlocks Detected (1h)"
          },
          {
            "expr": "atlantis_enhanced_locking_circuit_breaker_state",
            "legendFormat": "Circuit Breaker State"
          },
          {
            "expr": "rate(atlantis_enhanced_locking_retries_total[5m])",
            "legendFormat": "Retries/sec"
          }
        ]
      }
    ]
  }
}
```

### Alert Thresholds and Conditions

#### Critical Alerts (Page-worthy)
```yaml
alert_rules:
  - name: "EnhancedLockingSystemDown"
    condition: "atlantis_enhanced_locking_backend_health == 0"
    duration: "30s"
    severity: "critical"
    message: "Enhanced locking backend is completely unhealthy"
    actions:
      - page_oncall
      - trigger_emergency_runbook
      - auto_fallback_to_legacy

  - name: "HighLockFailureRate"
    condition: "rate(atlantis_enhanced_locking_acquisitions_success_total[5m]) / rate(atlantis_enhanced_locking_acquisitions_total[5m]) < 0.95"
    duration: "5m"
    severity: "critical"
    message: "Lock success rate below 95% for 5 minutes"
    actions:
      - page_oncall
      - escalate_to_platform_team

  - name: "ExtremelyHighLatency"
    condition: "histogram_quantile(0.95, atlantis_enhanced_locking_acquisition_duration_seconds_bucket) > 10"
    duration: "2m"
    severity: "critical"
    message: "95th percentile lock acquisition time above 10 seconds"
    actions:
      - page_oncall
      - trigger_performance_runbook

  - name: "MassiveQueueBacklog"
    condition: "sum(atlantis_enhanced_locking_queue_depth) > 1000"
    duration: "5m"
    severity: "critical"
    message: "Total queue depth exceeds 1000 requests"
    actions:
      - page_oncall
      - trigger_capacity_scaling
```

#### Warning Alerts (Notification-worthy)
```yaml
alert_rules:
  - name: "ModeratePerformanceDegradation"
    condition: "histogram_quantile(0.95, atlantis_enhanced_locking_acquisition_duration_seconds_bucket) > 1"
    duration: "10m"
    severity: "warning"
    message: "95th percentile lock acquisition time above 1 second"
    actions:
      - slack_devops_channel
      - trigger_performance_investigation

  - name: "HighQueueDepth"
    condition: "sum(atlantis_enhanced_locking_queue_depth) > 100"
    duration: "5m"
    severity: "warning"
    message: "Queue depth consistently above 100"
    actions:
      - slack_devops_channel

  - name: "FrequentDeadlocks"
    condition: "increase(atlantis_enhanced_locking_deadlocks_detected_total[1h]) > 10"
    duration: "1m"
    severity: "warning"
    message: "More than 10 deadlocks detected in the past hour"
    actions:
      - slack_devops_channel
      - trigger_deadlock_analysis

  - name: "CircuitBreakerActive"
    condition: "atlantis_enhanced_locking_circuit_breaker_state == 1"
    duration: "1m"
    severity: "warning"
    message: "Circuit breaker is open - failing fast to prevent cascading failures"
    actions:
      - slack_devops_channel

  - name: "HighResourceUsage"
    condition: "atlantis_enhanced_locking_redis_memory_used_bytes > 2000000000"  # 2GB
    duration: "5m"
    severity: "warning"
    message: "Redis memory usage above 2GB"
    actions:
      - slack_devops_channel
      - trigger_cleanup_job
```

### Dashboard Recommendations

#### Executive/Management Dashboard
```yaml
executive_dashboard:
  title: "Atlantis Locking - Executive Overview"
  panels:
    - system_availability_sla
    - deployment_throughput_trend
    - user_satisfaction_score
    - cost_efficiency_metrics
    - migration_progress_status
```

#### Operations Dashboard
```yaml
operations_dashboard:
  title: "Atlantis Locking - Operations"
  panels:
    - current_system_health
    - active_alerts_summary
    - performance_trends_24h
    - queue_status_by_priority
    - resource_utilization
    - error_rates_breakdown
```

#### Development/Debugging Dashboard
```yaml
debug_dashboard:
  title: "Atlantis Locking - Development & Debug"
  panels:
    - individual_lock_timeline
    - deadlock_detection_details
    - redis_operation_breakdown
    - queue_processing_details
    - event_system_health
    - configuration_drift_detection
```

### Log Analysis Patterns

#### Critical Error Patterns to Monitor
```bash
# Lock acquisition failures
grep -E "failed to acquire lock|lock acquisition timeout|lock already exists" atlantis.log

# Redis connectivity issues
grep -E "redis.*connection|redis.*timeout|redis.*failed" atlantis.log

# Deadlock detection
grep -E "deadlock detected|circular dependency|breaking deadlock" atlantis.log

# Performance degradation
grep -E "slow operation|high latency|performance warning" atlantis.log

# Configuration issues
grep -E "invalid configuration|config validation failed|feature disabled" atlantis.log
```

#### Log Analysis Commands
```bash
# Analyze error patterns over time
awk '/ERROR/ {print $1 " " $2}' atlantis.log | uniq -c | sort -nr

# Find most problematic resources
grep "lock acquisition failed" atlantis.log | awk '{print $NF}' | sort | uniq -c | sort -nr | head -20

# Performance analysis from logs
grep "lock acquired" atlantis.log | awk '{print $(NF-1)}' | awk -F'ms' '{sum+=$1; count++} END {print "Average:", sum/count "ms"}'

# Queue depth analysis
grep "queue depth" atlantis.log | tail -100 | awk '{print $NF}' | sort -n | awk '{a[NR]=$1} END {print "Median:", a[int(NR/2)]}'
```

## Performance Optimization

### Redis Configuration Optimization

#### Memory Optimization
```bash
# Redis configuration for optimal memory usage
redis.conf:
# Memory management
maxmemory 8gb                    # Set based on available system memory
maxmemory-policy allkeys-lru     # Evict least recently used keys when memory full
maxmemory-samples 5              # Number of keys to sample for LRU

# Disable persistence if data loss is acceptable (much faster)
save ""                          # Disable RDB snapshots
appendonly no                    # Disable AOF
appendfsync no                   # Don't force fsync

# Or configure minimal persistence
save 900 1                       # Save after 900s if at least 1 key changed
save 300 10                      # Save after 300s if at least 10 keys changed
save 60 10000                    # Save after 60s if at least 10000 keys changed

# Memory optimization
hash-max-ziplist-entries 512
hash-max-ziplist-value 64
list-max-ziplist-entries 512
list-max-ziplist-value 64
set-max-intset-entries 512
zset-max-ziplist-entries 128
zset-max-ziplist-value 64
```

#### Network and Connection Optimization
```bash
# Redis network configuration
tcp-keepalive 60                 # Enable TCP keepalive
timeout 300                      # Client idle timeout (5 minutes)
tcp-backlog 511                  # TCP listen backlog

# Connection limits
maxclients 10000                 # Maximum concurrent clients

# Buffer sizes for optimal throughput
client-output-buffer-limit normal 0 0 0
client-output-buffer-limit replica 256mb 64mb 60
client-output-buffer-limit pubsub 32mb 8mb 60

# Disable slow operations logging if not needed
slowlog-log-slower-than -1       # Disable slow log for performance
```

#### Redis Cluster Optimization
```bash
# For Redis Cluster deployments
redis-cluster.conf:
cluster-enabled yes
cluster-config-file nodes.conf
cluster-node-timeout 15000       # 15 second timeout
cluster-announce-port 6379
cluster-announce-bus-port 16379

# Performance tuning
cluster-migration-barrier 1      # Minimum replicas for migration
cluster-require-full-coverage no # Allow partial coverage for availability
```

### Application-Level Optimization

#### Connection Pool Tuning
```yaml
# server.yaml - Optimized connection pool settings
enhanced-locking:
  redis:
    # Connection pool configuration
    pool-size: 100                # Increase for high-throughput environments
    min-idle-conns: 20           # Keep minimum connections warm
    max-conn-age: "30m"          # Rotate connections periodically
    pool-timeout: "10s"          # Connection acquisition timeout
    idle-timeout: "15m"          # Close idle connections after 15 minutes
    idle-check-frequency: "1m"   # Check for idle connections every minute

    # Network timeouts
    read-timeout: "10s"
    write-timeout: "10s"
    dial-timeout: "5s"

    # Reliability features
    keep-alive: true
    retry-on-failure: true
    max-retries: 3
    retry-delay: "100ms"
```

#### Queue Processing Optimization
```yaml
# server.yaml - Queue processing tuning
enhanced-locking:
  # Queue management
  max-queue-size: 5000           # Higher limit for burst traffic
  queue-timeout: "10m"           # Reasonable timeout for queued requests

  # Processing efficiency
  queue-batch-size: 50           # Process requests in batches
  queue-processing-interval: "500ms"  # Fast processing interval
  queue-worker-count: 10         # Multiple workers for parallel processing

  # Priority queue optimization
  enable-priority-queue: true
  priority-boost-factor: 2.0     # Multiplier for higher priority processing

  # Resource-based queues for better isolation
  enable-resource-based-queues: true
  max-resource-queues: 1000
```

#### Lock Management Optimization
```yaml
# server.yaml - Lock management tuning
enhanced-locking:
  # Timeout configuration
  default-timeout: "20m"         # Shorter default for faster turnover
  max-timeout: "2h"             # Reasonable maximum
  timeout-check-interval: "30s" # Check for expired locks frequently

  # Cleanup optimization
  enable-automatic-cleanup: true
  cleanup-interval: "2m"        # Frequent cleanup
  cleanup-batch-size: 100       # Process cleanups in batches
  cleanup-worker-count: 3       # Multiple cleanup workers

  # Lock efficiency
  enable-lock-caching: true     # Cache lock state locally
  lock-cache-size: 10000        # Large cache for frequently accessed locks
  lock-cache-ttl: "5m"         # Cache entries for 5 minutes
```

#### Event System Optimization
```yaml
# server.yaml - Event system tuning
enhanced-locking:
  # Event processing
  enable-events: true
  event-buffer-size: 10000       # Large buffer for event bursts
  event-batch-size: 200         # Process events in large batches
  event-processing-interval: "1s"  # Frequent processing
  event-worker-count: 5         # Multiple event workers

  # Event filtering
  event-filters:
    - type: "lock-acquired"
      priority: "high"
    - type: "lock-released"
      priority: "normal"
    - type: "queue-depth"
      priority: "low"
```

### Performance Monitoring and Auto-Tuning

#### Automatic Performance Adjustment
```yaml
# server.yaml - Auto-tuning configuration
enhanced-locking:
  auto-tuning:
    enabled: true
    adjustment-interval: "5m"    # Check and adjust every 5 minutes

    # Pool size auto-adjustment
    pool-size-auto-adjust: true
    min-pool-size: 10
    max-pool-size: 200
    target-pool-utilization: 0.7  # Target 70% utilization

    # Queue size auto-adjustment
    queue-size-auto-adjust: true
    min-queue-size: 100
    max-queue-size: 10000
    target-queue-depth: 50       # Target queue depth

    # Timeout auto-adjustment
    timeout-auto-adjust: true
    min-timeout: "10m"
    max-timeout: "4h"
    target-success-rate: 0.98    # Adjust timeouts to maintain 98% success rate
```

#### Performance Metrics Collection
```yaml
# server.yaml - Comprehensive metrics
enhanced-locking:
  metrics:
    enabled: true
    collection-interval: "10s"   # High-frequency metrics collection

    # Performance metrics
    enable-timing-histograms: true
    histogram-buckets: [0.001, 0.005, 0.01, 0.05, 0.1, 0.5, 1, 5, 10, 30]

    # Resource metrics
    enable-resource-metrics: true
    enable-queue-metrics: true
    enable-connection-metrics: true

    # Advanced metrics
    enable-operation-tracing: true  # For detailed performance analysis
    trace-sample-rate: 0.1         # Sample 10% of operations
```

### Hardware and Infrastructure Optimization

#### Redis Hardware Recommendations
```yaml
# Production Redis hardware specs
redis_hardware:
  # Small deployment (< 1000 ops/sec)
  small:
    cpu: "2 cores, 3+ GHz"
    memory: "4GB RAM"
    storage: "SSD, 50GB+"
    network: "1 Gbps"

  # Medium deployment (1000-10000 ops/sec)
  medium:
    cpu: "4 cores, 3+ GHz"
    memory: "8GB RAM"
    storage: "NVMe SSD, 100GB+"
    network: "10 Gbps"

  # Large deployment (10000+ ops/sec)
  large:
    cpu: "8+ cores, 3+ GHz"
    memory: "16GB+ RAM"
    storage: "NVMe SSD, 200GB+"
    network: "10+ Gbps"
    redis_cluster: true

  # High availability requirements
  high_availability:
    replicas: 2                  # At least 2 replicas per master
    sentinel: true               # Use Redis Sentinel for failover
    data_persistence: true       # Enable persistence despite performance cost
```

#### Network Optimization
```bash
# Linux kernel network tuning for Redis
sysctl.conf:
# TCP settings
net.core.somaxconn = 65535
net.core.netdev_max_backlog = 5000
net.ipv4.tcp_max_syn_backlog = 65535
net.ipv4.tcp_keepalive_time = 600
net.ipv4.tcp_keepalive_intvl = 60
net.ipv4.tcp_keepalive_probes = 3

# Memory settings
vm.overcommit_memory = 1
vm.swappiness = 1                # Minimize swapping

# File descriptor limits
fs.file-max = 2097152
```

#### Container Optimization (Docker/Kubernetes)
```yaml
# docker-compose.yaml - Optimized Redis container
version: '3.8'
services:
  redis:
    image: redis:7-alpine
    command: redis-server --maxmemory 4gb --maxmemory-policy allkeys-lru
    deploy:
      resources:
        limits:
          memory: 4G
          cpus: '2'
        reservations:
          memory: 2G
          cpus: '1'
    sysctls:
      - net.core.somaxconn=65535
    ulimits:
      memlock: -1
      nofile: 65535
    volumes:
      - redis_data:/data
      - ./redis.conf:/usr/local/etc/redis/redis.conf
    ports:
      - "6379:6379"

volumes:
  redis_data:
```

```yaml
# kubernetes.yaml - Optimized Redis deployment
apiVersion: apps/v1
kind: Deployment
metadata:
  name: redis
spec:
  replicas: 1
  selector:
    matchLabels:
      app: redis
  template:
    metadata:
      labels:
        app: redis
    spec:
      containers:
      - name: redis
        image: redis:7-alpine
        command: ["redis-server", "/etc/redis/redis.conf"]
        resources:
          requests:
            memory: "2Gi"
            cpu: "1"
          limits:
            memory: "4Gi"
            cpu: "2"
        ports:
        - containerPort: 6379
        volumeMounts:
        - name: redis-config
          mountPath: /etc/redis
        - name: redis-data
          mountPath: /data
      volumes:
      - name: redis-config
        configMap:
          name: redis-config
      - name: redis-data
        persistentVolumeClaim:
          claimName: redis-pvc
```

### Load Testing and Benchmarking

#### Performance Benchmarking Suite
```bash
# Built-in performance testing
atlantis enhanced-locking benchmark --help

# Comprehensive benchmark
atlantis enhanced-locking benchmark \
  --duration=30m \
  --concurrent-users=100 \
  --operations-per-second=1000 \
  --priority-distribution="critical:10%,high:20%,normal:60%,low:10%" \
  --resource-count=50 \
  --report-format=json \
  --output=benchmark-results.json

# Redis-specific benchmarking
redis-benchmark -h redis-server -p 6379 -c 100 -n 100000 -t set,get -d 1024

# Load testing with realistic scenarios
atlantis enhanced-locking load-test \
  --scenario=realistic-workload \
  --duration=1h \
  --ramp-up=10m \
  --users=500 \
  --think-time=5s
```

#### Custom Load Testing Script
```bash
#!/bin/bash
# load-test-atlantis.sh

echo "Starting Atlantis Enhanced Locking Load Test"

# Test configuration
DURATION="30m"
CONCURRENT_USERS=50
BASE_URL="http://atlantis-server"
RESULTS_DIR="load-test-results-$(date +%Y%m%d-%H%M%S)"

mkdir -p "$RESULTS_DIR"

# Monitor system resources
monitor_resources() {
    while true; do
        echo "$(date),$(free -m | grep Mem | awk '{print $3}'),$(top -bn1 | grep "Cpu(s)" | awk '{print $2}' | sed 's/%us,//')" >> "$RESULTS_DIR/system-metrics.csv"
        sleep 10
    done
}

# Monitor Redis performance
monitor_redis() {
    while true; do
        redis-cli --latency-history -i 10 >> "$RESULTS_DIR/redis-latency.log"
        redis-cli info stats | grep -E "instantaneous_ops_per_sec|used_memory_human" >> "$RESULTS_DIR/redis-stats.log"
        sleep 10
    done
}

# Start monitoring
monitor_resources &
MONITOR_PID=$!

monitor_redis &
REDIS_MONITOR_PID=$!

# Run load test
echo "Running load test for $DURATION with $CONCURRENT_USERS concurrent users"
atlantis enhanced-locking load-test \
  --duration=$DURATION \
  --concurrent-users=$CONCURRENT_USERS \
  --base-url=$BASE_URL \
  --output="$RESULTS_DIR/load-test-results.json" \
  --verbose

# Stop monitoring
kill $MONITOR_PID $REDIS_MONITOR_PID

echo "Load test completed. Results saved to $RESULTS_DIR"
echo "Summary:"
jq '.summary' "$RESULTS_DIR/load-test-results.json"
```

### Performance Troubleshooting

#### Performance Analysis Commands
```bash
# Real-time performance monitoring
atlantis enhanced-locking monitor \
  --metrics=latency,throughput,queue-depth,error-rate \
  --interval=5s \
  --duration=10m

# Detailed performance profiling
atlantis enhanced-locking profile \
  --type=cpu \
  --duration=5m \
  --output=cpu-profile.prof

# Memory profiling
atlantis enhanced-locking profile \
  --type=memory \
  --duration=5m \
  --output=memory-profile.prof

# Lock contention analysis
atlantis enhanced-locking analyze-contention \
  --period=24h \
  --top=20 \
  --show-patterns \
  --output=contention-analysis.json
```

#### Performance Optimization Checklist
```yaml
performance_checklist:
  redis_optimization:
    - memory_policy_configured: true
    - persistence_disabled_if_acceptable: true
    - network_timeouts_optimized: true
    - connection_pool_sized_correctly: true

  application_optimization:
    - queue_batch_processing_enabled: true
    - automatic_cleanup_configured: true
    - event_system_tuned: true
    - metrics_collection_optimized: true

  infrastructure_optimization:
    - adequate_hardware_provisioned: true
    - network_tuning_applied: true
    - container_resources_allocated: true
    - monitoring_dashboards_configured: true

  load_testing:
    - baseline_performance_measured: true
    - realistic_load_tested: true
    - performance_regression_tests_automated: true
    - capacity_planning_completed: true
```

## Frequently Asked Questions (FAQ)

### Migration Questions

#### Q: Do I need to change my atlantis.yaml files when migrating?
**A:** No, you don't need to make any changes to your `atlantis.yaml` files for basic migration. The enhanced locking system maintains full backward compatibility with existing configurations. You can optionally add enhanced features like priorities and custom timeouts later.

#### Q: How long does the migration take?
**A:** The complete migration typically takes 4-6 weeks:
- Week 1-2: Preparation and testing (enhanced system disabled)
- Week 3-4: Shadow mode (monitoring only, no traffic processing)
- Week 5-6: Gradual traffic migration (10% → 50% → 90% → 100%)

For urgent deployments, a fast-track migration can be completed in 1-2 weeks with increased risk.

#### Q: Can I rollback if something goes wrong?
**A:** Yes, the migration supports both emergency rollback (<5 minutes) and planned rollback procedures. The enhanced system includes a fallback mechanism that automatically switches to the legacy system if critical issues are detected.

#### Q: Will there be any downtime during migration?
**A:** No, the migration is designed to be zero-downtime. Users will not experience any service interruption. The worst case is a brief fallback to the legacy system if issues occur.

#### Q: What happens to existing locks during migration?
**A:** Existing locks are preserved throughout the migration. The enhanced system can read and manage locks created by the legacy system. During the migration phases, locks created by either system are accessible by both.

### Performance Questions

#### Q: What performance improvements can I expect?
**A:** The enhanced locking system provides significant performance improvements:
- **50x throughput increase**: From ~1,000 ops/sec to 50,000+ ops/sec
- **10x latency reduction**: From ~10ms to <1ms average lock acquisition time
- **Better scalability**: Supports Redis clustering for horizontal scaling
- **Reduced contention**: Priority queues and intelligent scheduling

#### Q: How much Redis memory will I need?
**A:** Memory requirements depend on your lock patterns:
- **Small deployment** (<1000 locks/hour): 1-2GB Redis memory
- **Medium deployment** (1000-10000 locks/hour): 2-4GB Redis memory
- **Large deployment** (10000+ locks/hour): 4-8GB+ Redis memory

Each active lock uses approximately 1-2KB of memory.

#### Q: Can the enhanced system handle the same load as my legacy system?
**A:** Yes, and much more. The enhanced system is designed to handle 50x the throughput of the legacy system. If you're currently running at capacity with the legacy system, you should see significant performance improvements.

#### Q: What if Redis goes down?
**A:** The enhanced system includes several failover mechanisms:
- **Automatic fallback**: Falls back to legacy system if Redis becomes unavailable
- **Circuit breaker**: Prevents cascading failures
- **Redis clustering**: High availability through Redis Cluster
- **Connection pooling**: Resilient connection management

### Feature Comparison Questions

#### Q: What features will I lose by migrating?
**A:** You won't lose any existing functionality. The enhanced system maintains full compatibility with all legacy locking features while adding new capabilities.

#### Q: What new features do I get with the enhanced system?
**A:** The enhanced system adds many new features:

| Feature Category | Enhanced System Features |
|-----------------|-------------------------|
| **Performance** | Priority queues, batched operations, connection pooling |
| **Reliability** | Deadlock detection, circuit breakers, automatic cleanup |
| **Scalability** | Redis clustering, horizontal scaling, load balancing |
| **Observability** | Comprehensive metrics, event system, detailed logging |
| **Configuration** | Priority levels, custom timeouts, resource isolation |

#### Q: How does priority queuing work?
**A:** Priority queuing processes lock requests based on configured priority levels:
1. **Critical**: Emergency fixes, production incidents (processed immediately)
2. **High**: Important features, security updates (processed within seconds)
3. **Normal**: Regular development work (processed within minutes)
4. **Low**: Documentation, cleanup tasks (processed when capacity available)

Priority is configured in your `atlantis.yaml` project configuration.

#### Q: What happens if I don't configure priorities?
**A:** All operations default to "normal" priority and are processed in order. The system works exactly like the legacy system until you add priority configurations.

### Configuration Questions

#### Q: How do I configure Redis for optimal performance?
**A:** Key Redis optimizations for Atlantis:

```bash
# Memory optimization
maxmemory 4gb
maxmemory-policy allkeys-lru

# Performance optimization (disable persistence if acceptable)
save ""
appendonly no

# Network optimization
tcp-keepalive 60
timeout 300
```

See the Performance Optimization section for complete configuration examples.

#### Q: Should I use Redis Cluster?
**A:** Use Redis Cluster for:
- **High throughput environments** (>10,000 locks/hour)
- **High availability requirements**
- **Large-scale deployments** with many concurrent users

Single Redis instance is sufficient for most deployments.

#### Q: How do I tune connection pool settings?
**A:** Connection pool settings depend on your load:

```yaml
# Light load (<1000 ops/sec)
enhanced-locking:
  redis:
    pool-size: 20
    min-idle-conns: 5

# Medium load (1000-5000 ops/sec)
enhanced-locking:
  redis:
    pool-size: 50
    min-idle-conns: 10

# Heavy load (>5000 ops/sec)
enhanced-locking:
  redis:
    pool-size: 100
    min-idle-conns: 20
```

### Troubleshooting Questions

#### Q: What should I do if lock acquisition is slow?
**A:** Check these common causes:
1. **Redis performance**: Use `redis-cli --latency` to check Redis response times
2. **Connection pool exhaustion**: Monitor pool utilization metrics
3. **Queue congestion**: Check queue depth metrics
4. **Network latency**: Verify network connectivity to Redis

See the Performance Degradation troubleshooting section for detailed steps.

#### Q: How do I identify deadlocks?
**A:** The enhanced system automatically detects deadlocks:

```bash
# Check for active deadlocks
atlantis enhanced-locking deadlock-status --show-chains

# Analyze deadlock patterns
atlantis enhanced-locking deadlock-analysis --period=24h

# Enable automatic deadlock resolution
atlantis config set enhanced-locking.enable-automatic-deadlock-resolution true
```

#### Q: What logs should I monitor?
**A:** Key log patterns to monitor:

```bash
# Critical errors
grep -E "failed to acquire lock|redis.*connection|deadlock detected" atlantis.log

# Performance issues
grep -E "slow operation|high latency|timeout" atlantis.log

# Configuration problems
grep -E "invalid configuration|feature disabled" atlantis.log
```

#### Q: How do I clean up stuck locks?
**A:** Use the built-in cleanup tools:

```bash
# Clean up expired locks
atlantis enhanced-locking cleanup --expired-only

# Force cleanup of specific resources
atlantis enhanced-locking cleanup --resource="company/repo/path" --force

# Enable automatic cleanup
atlantis config set enhanced-locking.enable-automatic-cleanup true
```

### Monitoring Questions

#### Q: What metrics should I monitor?
**A:** Key metrics for system health:

**Critical Metrics:**
- Lock success rate (should be >98%)
- Lock acquisition latency (should be <1s p95)
- Queue depth (should be <100)
- Redis memory usage
- Error rate (should be <2%)

**Performance Metrics:**
- Throughput (operations per second)
- Queue wait times
- Connection pool utilization
- Deadlock detection rate

See the Monitoring and Alerting section for complete metric definitions.

#### Q: How do I set up alerts?
**A:** Configure alerts for critical thresholds:

```yaml
# Critical alerts (page-worthy)
- lock_success_rate < 95%
- lock_acquisition_latency_p95 > 5s
- total_queue_depth > 1000
- redis_memory_usage > 4GB

# Warning alerts (notification-worthy)
- lock_acquisition_latency_p95 > 1s
- queue_depth > 100
- deadlocks_per_hour > 10
```

#### Q: What dashboards should I create?
**A:** Recommended dashboard hierarchy:
1. **Executive Dashboard**: SLA metrics, availability, cost efficiency
2. **Operations Dashboard**: Current health, active alerts, performance trends
3. **Debug Dashboard**: Detailed lock timelines, deadlock details, configuration drift

### Cost and Sizing Questions

#### Q: What are the infrastructure costs?
**A:** Cost comparison (monthly estimates for medium deployment):

| Component | Legacy System | Enhanced System | Savings |
|-----------|---------------|-----------------|---------|
| **Compute** | $200 | $150 | -25% |
| **Storage** | $50 | $30 | -40% |
| **Redis** | $0 | $100 | +$100 |
| **Monitoring** | $20 | $30 | +$10 |
| **Total** | $270 | $310 | +$40/month |

**ROI Benefits:**
- 50x performance improvement
- Reduced operational overhead
- Better developer productivity
- Reduced downtime incidents

#### Q: How do I size my Redis instance?
**A:** Sizing guidelines based on usage patterns:

```yaml
# Calculate based on peak concurrent locks
redis_memory_calculation:
  locks_per_hour: 5000
  average_lock_duration: "20m"
  peak_concurrent_locks: 1667  # (5000 * 20) / 60
  memory_per_lock: "2KB"
  base_redis_overhead: "500MB"
  total_memory_needed: "4GB"  # (1667 * 2KB) + 500MB + buffer
```

Use 2x the calculated memory for safety buffer.

### Integration Questions

#### Q: How does this integrate with CI/CD pipelines?
**A:** The enhanced system integrates seamlessly with existing CI/CD:
- **Same API**: All existing webhook and API integrations work unchanged
- **Better performance**: Faster lock acquisition improves pipeline execution times
- **Priority support**: Mark production deployments as high/critical priority
- **Better observability**: More detailed metrics for pipeline monitoring

#### Q: Can I use this with multiple Atlantis instances?
**A:** Yes, multiple Atlantis instances can share the same Redis backend:

```yaml
# Each Atlantis instance configuration
enhanced-locking:
  redis:
    addr: "shared-redis-cluster:6379"
    key-prefix: "atlantis:locks:"  # Shared namespace

  # Instance-specific settings
  instance-id: "atlantis-prod-1"   # Unique identifier
```

#### Q: How does this work with Terraform Cloud/Enterprise?
**A:** The enhanced locking system is independent of Terraform backends:
- Works with any Terraform backend (local, S3, Terraform Cloud, etc.)
- Provides Atlantis-level locking regardless of Terraform's locking mechanism
- Can be configured to disable Terraform's built-in locking to avoid conflicts

#### Q: What about compliance and audit requirements?
**A:** The enhanced system provides comprehensive audit capabilities:
- **Detailed logging**: All lock operations logged with timestamps and user context
- **Event system**: Real-time events for compliance monitoring
- **Metrics retention**: Long-term metrics storage for audit purposes
- **Access control**: Integration with existing authentication and authorization

### Best Practices Questions

#### Q: What are the recommended migration practices?
**A:** Key migration best practices:

1. **Start conservative**: Begin with basic configuration, add features gradually
2. **Monitor extensively**: Set up comprehensive monitoring before migration
3. **Test thoroughly**: Run load tests and compatibility validation
4. **Plan rollbacks**: Always have a tested rollback procedure
5. **Train teams**: Ensure teams understand new features and troubleshooting

#### Q: How should I organize atlantis.yaml configurations?
**A:** Recommended patterns for enhanced system:

```yaml
# Organize by environment and criticality
projects:
  # Production: High priority, strict settings
  - name: prod-infrastructure
    workspace: production
    repo_locks:
      mode: on_apply
      priority: critical
      timeout: "60m"

  # Staging: Standard settings
  - name: staging-infrastructure
    workspace: staging
    repo_locks:
      mode: on_plan
      priority: normal
      timeout: "30m"

  # Development: Relaxed settings
  - name: dev-infrastructure
    workspace: development
    repo_locks:
      mode: on_plan
      priority: low
      timeout: "15m"
      allow_concurrent: true
```

#### Q: When should I use different priority levels?
**A:** Priority level guidelines:

| Priority | Use Cases | Examples |
|----------|-----------|----------|
| **Critical** | Production incidents, security patches | Hotfixes, emergency deployments |
| **High** | Important features, scheduled maintenance | Release deployments, infrastructure updates |
| **Normal** | Regular development work | Feature development, routine updates |
| **Low** | Cleanup, documentation, experiments | Terraform fmt, documentation updates |

#### Q: How do I optimize for my specific workload?
**A:** Workload-specific optimizations:

**High-frequency, short-duration locks:**
```yaml
enhanced-locking:
  default-timeout: "5m"
  queue-batch-size: 100
  cleanup-interval: "1m"
```

**Low-frequency, long-duration locks:**
```yaml
enhanced-locking:
  default-timeout: "2h"
  queue-timeout: "30m"
  cleanup-interval: "10m"
```

**Mixed workload:**
```yaml
enhanced-locking:
  enable-adaptive-timeouts: true
  enable-priority-queue: true
  auto-tuning:
    enabled: true
```

#### Memory Optimization
```bash
# Redis configuration for optimal performance
redis.conf:
maxmemory 4gb
maxmemory-policy allkeys-lru
save ""  # Disable persistence for better performance
appendonly no
tcp-keepalive 60
timeout 300
```

#### Connection Pool Optimization
```yaml
# server.yaml - Redis connection optimization
enhanced-locking:
  redis:
    pool-size: 20
    min-idle-conns: 5
    max-conn-age: "30m"
    pool-timeout: "5s"
    idle-timeout: "10m"
    idle-check-frequency: "1m"
```

#### Cluster Configuration
```yaml
# Redis cluster optimization
redis-cluster:
  nodes:
    - "redis-1:6379"
    - "redis-2:6379"
    - "redis-3:6379"
    - "redis-4:6379"
    - "redis-5:6379"
    - "redis-6:6379"

  read-preference: "slave"  # Read from slaves when possible
  route-by-latency: true
  max-redirects: 3
```

### Application-Level Optimization

#### Queue Management
```yaml
# Optimize queue processing
enhanced-locking:
  # Separate queues per resource prevent blocking
  enable-resource-based-queues: true

  # Reasonable queue limits
  max-queue-size: 1000
  queue-timeout: "10m"

  # Efficient cleanup
  queue-cleanup-interval: "5m"
  expired-lock-cleanup-interval: "1m"
```

#### Timeout Configuration
```yaml
# Balanced timeout configuration
enhanced-locking:
  default-timeout: "20m"  # Shorter than legacy 30m
  max-timeout: "1h"       # Prevent very long locks

  # Adaptive timeouts based on load
  enable-adaptive-timeouts: true
  adaptive-timeout-factor: 1.5
```

#### Feature Tuning
```yaml
# Performance-focused feature configuration
enhanced-locking:
  # Deadlock detection with reasonable frequency
  enable-deadlock-detection: true
  deadlock-check-interval: "1m"  # More frequent for faster resolution

  # Circuit breaker to prevent cascade failures
  enable-circuit-breaker: true
  circuit-breaker:
    failure-threshold: 10
    success-threshold: 5
    timeout: "30s"

  # Event system optimization
  enable-events: true
  event-buffer-size: 5000  # Larger buffer for high throughput
  event-batch-size: 100    # Batch events for efficiency
```

---

## atlantis.yaml Integration Patterns

This section covers how `atlantis.yaml` project configurations integrate with both legacy and enhanced locking systems, focusing on lock scope, granularity, and resource isolation patterns.

### Understanding Lock Scope and Granularity

#### Legacy System Lock Key Format

The legacy locking system uses a simple key format: `{repoFullName}/{path}/{workspace}`

```bash
# Example legacy lock keys:
company/infrastructure/./production
company/infrastructure/database/production
company/infrastructure/applications/staging
company/infrastructure/modules/shared
```

#### Enhanced System Resource Identification

The enhanced system uses structured resource identifiers with additional metadata:

```go
type ResourceIdentifier struct {
    Type      ResourceType // project, workspace, global, custom
    Namespace string       // Repository namespace (repoFullName)
    Name      string       // Resource name (project path)
    Workspace string       // Workspace name
    Path      string       // Directory path
}
```

### Project Configuration Impact on Locking

#### 1. Single Directory, Multiple Projects

**Configuration:**
```yaml
# atlantis.yaml
version: 3
projects:
  - name: production-app
    dir: .
    workspace: production
    repo_locks:
      mode: on_apply

  - name: staging-app
    dir: .
    workspace: staging
    repo_locks:
      mode: on_plan
```

**Legacy Lock Behavior:**
- Creates separate locks: `company/app/./production` and `company/app/./staging`
- Each workspace locked independently
- Same directory, different workspaces = different locks

**Enhanced Lock Behavior:**
```yaml
# Lock Resource 1:
namespace: "company/app"
name: "."
workspace: "production"
type: "project"

# Lock Resource 2:
namespace: "company/app"
name: "."
workspace: "staging"
type: "project"
```

#### 2. Multiple Directories, Different Lock Modes

**Configuration:**
```yaml
# atlantis.yaml
version: 3
projects:
  - name: database
    dir: database
    workspace: production
    repo_locks:
      mode: on_apply

  - name: api
    dir: applications/api
    workspace: production
    repo_locks:
      mode: on_plan

  - name: frontend
    dir: applications/frontend
    workspace: production
    repo_locks:
      mode: disabled
```

**Lock Granularity Comparison:**

| Project | Legacy Lock Key | Enhanced Resource ID | Lock Trigger |
|---------|----------------|---------------------|--------------|
| database | `company/infra/database/production` | `{namespace: "company/infra", name: "database", workspace: "production"}` | Apply only |
| api | `company/infra/applications/api/production` | `{namespace: "company/infra", name: "applications/api", workspace: "production"}` | Plan and Apply |
| frontend | No lock | No lock | Never |

### Enhanced System Integration Features

#### 1. Priority-Based Locking with atlantis.yaml

The enhanced system extends `atlantis.yaml` with priority settings:

```yaml
# Enhanced atlantis.yaml (backward compatible)
version: 3
projects:
  - name: critical-infrastructure
    dir: infrastructure
    workspace: production
    repo_locks:
      mode: on_apply
      priority: critical        # Enhanced: priority levels
      timeout: "45m"           # Enhanced: custom timeouts
      resource_isolation: strict  # Enhanced: isolation levels

  - name: development-environment
    dir: infrastructure
    workspace: dev
    repo_locks:
      mode: on_plan
      priority: normal
      timeout: "15m"
      resource_isolation: standard
```

**Priority Mapping:**
```yaml
# Priority levels in enhanced system
priority_mapping:
  critical: 3    # Emergency fixes, production issues
  high: 2        # Important features, security updates
  normal: 1      # Regular development work
  low: 0         # Documentation, cleanup tasks
```

#### 2. Queue Isolation Based on Project Configuration

**Resource-Based Queue Isolation:**
```yaml
# atlantis.yaml with enhanced queue configuration
version: 3
projects:
  - name: shared-modules
    dir: modules
    workspace: shared
    execution_order_group: 0
    repo_locks:
      mode: on_apply
      priority: high
      # Enhanced: This creates isolated queue for shared resources
      queue_isolation: resource_type

  - name: app-1
    dir: applications/app1
    workspace: production
    execution_order_group: 1
    depends_on: [shared-modules]
    repo_locks:
      mode: on_apply
      priority: normal
      queue_isolation: workspace  # Separate queue per workspace
```

**Queue Behavior:**
- **Legacy**: Single global queue, first-come-first-served
- **Enhanced**: Multiple queues based on configuration
  - Resource-type queues: Shared modules vs applications
  - Workspace queues: Production vs staging vs development
  - Priority queues: Critical > High > Normal > Low

#### 3. Execution Order Groups and Locking

**Configuration:**
```yaml
# atlantis.yaml with execution ordering
version: 3
projects:
  - name: vpc
    dir: infrastructure/vpc
    workspace: production
    execution_order_group: 1
    repo_locks:
      mode: on_apply
      priority: high

  - name: database
    dir: infrastructure/database
    workspace: production
    execution_order_group: 2
    depends_on: [vpc]
    repo_locks:
      mode: on_apply
      priority: high

  - name: applications
    dir: applications
    workspace: production
    execution_order_group: 3
    depends_on: [database]
    repo_locks:
      mode: on_apply
      priority: normal
```

**Enhanced System Dependency Lock Behavior:**
1. **Sequential Execution**: Higher order groups wait for lower groups
2. **Dependency Tracking**: Enhanced system tracks cross-project dependencies
3. **Deadlock Prevention**: Automatically prevents circular dependencies
4. **Automatic Queuing**: Dependent projects automatically queued until dependencies complete

### Advanced Configuration Patterns

#### 1. Workspace-Specific Lock Behavior

```yaml
# Complex workspace configuration
version: 3
projects:
  # Production: Strict locking, high priority
  - name: prod-infrastructure
    dir: .
    workspace: production
    repo_locks:
      mode: on_apply
      priority: critical
      timeout: "60m"
      resource_isolation: strict
      retry_policy:
        max_attempts: 3
        base_delay: "30s"

  # Staging: Moderate locking, normal priority
  - name: staging-infrastructure
    dir: .
    workspace: staging
    repo_locks:
      mode: on_plan
      priority: normal
      timeout: "30m"
      resource_isolation: standard

  # Development: Minimal locking, low priority
  - name: dev-infrastructure
    dir: .
    workspace: development
    repo_locks:
      mode: on_plan
      priority: low
      timeout: "15m"
      resource_isolation: relaxed
      allow_concurrent: true  # Enhanced: Allow concurrent dev locks
```

#### 2. Custom Workflow Lock Integration

```yaml
# workflows with locking integration
version: 3
workflows:
  production-deployment:
    plan:
      steps:
      - run: |
          echo "Acquiring high-priority lock for production deployment"
          # Enhanced system automatically applies project priority
      - init
      - plan:
          extra_args: ["-lock=false"]  # Terraform-level lock disabled, using Atlantis locks
    apply:
      steps:
      - run: |
          echo "Production deployment with critical priority lock"
      - apply:
          extra_args: ["-lock=false"]

  development-workflow:
    plan:
      steps:
      - run: |
          echo "Development work with relaxed locking"
      - init
      - plan
    apply:
      steps:
      - apply

projects:
  - name: prod-app
    dir: .
    workspace: production
    workflow: production-deployment
    repo_locks:
      mode: on_apply
      priority: critical

  - name: dev-app
    dir: .
    workspace: development
    workflow: development-workflow
    repo_locks:
      mode: on_plan
      priority: low
```

### Migration Considerations for atlantis.yaml

#### Backward Compatibility Guarantees

✅ **Fully Compatible (No Changes Required):**
```yaml
# Existing atlantis.yaml - works unchanged
version: 3
projects:
  - name: my-project
    dir: .
    workspace: production
    repo_locks:
      mode: on_apply
# Enhanced system automatically applies:
# - priority: normal (default)
# - timeout: 30m (default)
# - resource_isolation: standard (default)
```

✅ **Enhanced Features (Optional):**
```yaml
# Extended atlantis.yaml - optional enhancements
version: 3
projects:
  - name: my-project
    dir: .
    workspace: production
    repo_locks:
      mode: on_apply
      # Optional enhanced features:
      priority: high              # New in enhanced system
      timeout: "45m"              # New in enhanced system
      resource_isolation: strict  # New in enhanced system
```

#### Required Changes (None for Basic Migration)

**No configuration changes are required** for migration from legacy to enhanced locking system. The enhanced system:

1. **Preserves existing behavior**: All existing `repo_locks` configurations work unchanged
2. **Maintains lock keys**: Same resource identification logic for compatibility
3. **Supports gradual enhancement**: Add enhanced features incrementally

#### Migration Path Examples

**Phase 1: Basic Migration (No atlantis.yaml changes)**
```yaml
# Before and After: Identical configuration
version: 3
projects:
  - name: infrastructure
    dir: .
    workspace: production
    repo_locks:
      mode: on_apply
# Works identically in both systems
```

**Phase 2: Add Enhanced Features**
```yaml
# Enhanced configuration (optional upgrade)
version: 3
projects:
  - name: infrastructure
    dir: .
    workspace: production
    repo_locks:
      mode: on_apply
      priority: high           # Add priority
      timeout: "45m"           # Custom timeout

  - name: development
    dir: .
    workspace: dev
    repo_locks:
      mode: on_plan
      priority: normal         # Different priority
      timeout: "15m"           # Shorter timeout for dev
```

**Phase 3: Advanced Features**
```yaml
# Full enhanced system features
version: 3
projects:
  - name: shared-infrastructure
    dir: shared
    workspace: production
    execution_order_group: 1
    repo_locks:
      mode: on_apply
      priority: critical
      timeout: "60m"
      resource_isolation: strict
      retry_policy:
        max_attempts: 3
        base_delay: "30s"
      deadlock_detection: true

  - name: application
    dir: app
    workspace: production
    execution_order_group: 2
    depends_on: [shared-infrastructure]
    repo_locks:
      mode: on_apply
      priority: high
      timeout: "30m"
      queue_timeout: "10m"
```

### Lock Behavior Examples

#### Example 1: Multi-Directory Repository

**Repository Structure:**
```
company/infrastructure/
├── atlantis.yaml
├── networking/
│   ├── vpc/
│   └── security-groups/
├── compute/
│   ├── eks/
│   └── ec2/
└── databases/
    ├── rds/
    └── elasticache/
```

**atlantis.yaml Configuration:**
```yaml
version: 3
projects:
  # Networking - Deploy first
  - name: vpc
    dir: networking/vpc
    workspace: production
    execution_order_group: 1
    repo_locks:
      mode: on_apply
      priority: high

  - name: security-groups
    dir: networking/security-groups
    workspace: production
    execution_order_group: 1
    repo_locks:
      mode: on_apply
      priority: high

  # Databases - Deploy after networking
  - name: rds
    dir: databases/rds
    workspace: production
    execution_order_group: 2
    depends_on: [vpc, security-groups]
    repo_locks:
      mode: on_apply
      priority: normal

  # Compute - Deploy after databases
  - name: eks
    dir: compute/eks
    workspace: production
    execution_order_group: 3
    depends_on: [rds]
    repo_locks:
      mode: on_apply
      priority: normal
```

**Lock Behavior Comparison:**

| System | Lock Keys/Resources | Concurrency | Dependencies |
|--------|-------------------|-------------|--------------|
| **Legacy** | `company/infrastructure/networking/vpc/production`<br/>`company/infrastructure/networking/security-groups/production`<br/>`company/infrastructure/databases/rds/production`<br/>`company/infrastructure/compute/eks/production` | Independent locks<br/>No dependency enforcement | Manual coordination required |
| **Enhanced** | Structured resources with metadata<br/>Priority-based queuing<br/>Dependency tracking | Automatic dependency ordering<br/>Priority-based execution<br/>Deadlock prevention | Automatic dependency enforcement<br/>Sequential execution groups |

#### Example 2: Multi-Workspace Configuration

**atlantis.yaml:**
```yaml
version: 3
projects:
  # Production: High security, strict locking
  - name: prod-app
    dir: .
    workspace: production
    repo_locks:
      mode: on_apply
      priority: critical
      timeout: "60m"
      resource_isolation: strict

  # Staging: Moderate security, standard locking
  - name: staging-app
    dir: .
    workspace: staging
    repo_locks:
      mode: on_plan
      priority: normal
      timeout: "30m"
      resource_isolation: standard

  # Development: Relaxed locking, multiple concurrent allowed
  - name: dev-app-east
    dir: .
    workspace: dev-east
    repo_locks:
      mode: on_plan
      priority: low
      timeout: "15m"
      resource_isolation: relaxed
      allow_concurrent: true

  - name: dev-app-west
    dir: .
    workspace: dev-west
    repo_locks:
      mode: on_plan
      priority: low
      timeout: "15m"
      resource_isolation: relaxed
      allow_concurrent: true
```

**Enhanced System Behavior:**
- **Production**: Gets priority in queue, strict isolation
- **Staging**: Standard queuing, moderate isolation
- **Development**: Can run concurrently, lower priority

**Queue Processing Order:**
1. Critical priority (production) - processed first
2. Normal priority (staging) - processed second
3. Low priority (development) - processed last, but can run concurrently

---

## Conclusion

This migration guide provides a comprehensive, step-by-step approach to migrating from Atlantis legacy locking to the enhanced system. Key success factors:

1. **Gradual Migration**: Phased approach minimizes risk
2. **Comprehensive Testing**: Validate each phase thoroughly
3. **Rollback Readiness**: Always maintain ability to rollback
4. **Monitoring**: Observe system health throughout migration
5. **Performance Tuning**: Optimize for your specific workload
6. **atlantis.yaml Compatibility**: Full backward compatibility with optional enhancements

The enhanced locking system provides significant improvements in performance, scalability, and features while maintaining full backward compatibility during the migration process. The `atlantis.yaml` integration allows for granular control over lock behavior, priorities, and resource isolation without requiring any configuration changes for basic migration.