# Enhanced Locking System - Configuration Examples

## Overview

This document provides comprehensive configuration examples for the Enhanced Locking System, covering various deployment scenarios, feature combinations, and performance tuning options.

## Basic Configurations

### 1. Minimal Setup (New Installation)

```yaml
# atlantis.yaml - Basic enhanced locking
enhanced-locking:
  enabled: true
  backend: boltdb

  # Safe defaults for new installations
  priority-queue: false
  deadlock-detection: false
  retries: false

  # Essential features
  metrics: true
  health-checks: true
```

### 2. Development Environment

```yaml
# atlantis.yaml - Development setup
enhanced-locking:
  enabled: true
  backend: boltdb

  # Enable features for testing
  priority-queue: true
  deadlock-detection: true
  retries: true
  enhanced-metrics: true

  # Development-friendly settings
  lock-timeout: 30s
  queue-config:
    max-queue-size: 50
    processing-interval: 100ms

  # Enable debug features
  debug-mode: true
  verbose-logging: true
```

### 3. Production-Ready Configuration

```yaml
# atlantis.yaml - Production setup
enhanced-locking:
  enabled: true
  backend: redis

  # Production features
  priority-queue: true
  deadlock-detection: true
  retries: true
  enhanced-metrics: true

  # Conservative timeouts
  lock-timeout: 300s
  queue-timeout: 600s

  # Redis configuration
  redis:
    addresses: ["redis-1:6379", "redis-2:6379", "redis-3:6379"]
    password: "${REDIS_PASSWORD}"
    cluster-mode: true
    pool-size: 50
    max-retries: 3
    retry-delay: 100ms

  # Queue optimization
  queue-config:
    max-queue-size: 1000
    processing-interval: 50ms
    batch-processing: true
    batch-size: 10

  # Deadlock detection
  deadlock-detection:
    check-interval: 30s
    resolution-policy: "lowest_priority"
    max-resolution-attempts: 3

  # Monitoring
  metrics: true
  health-checks: true
  event-streaming: true
```

## Backend-Specific Configurations

### Redis Cluster Configuration

```yaml
# High-availability Redis setup
enhanced-locking:
  enabled: true
  backend: redis

  redis:
    # Redis Cluster nodes
    addresses:
      - "redis-1:6379"
      - "redis-2:6379"
      - "redis-3:6379"
      - "redis-4:6379"
      - "redis-5:6379"
      - "redis-6:6379"

    # Authentication
    password: "${REDIS_PASSWORD}"
    username: "${REDIS_USERNAME}"  # Redis 6+ ACL

    # Cluster settings
    cluster-mode: true
    read-only-slaves: true
    route-by-latency: true
    route-randomly: false

    # Connection pool
    pool-size: 100
    min-idle-connections: 10
    max-connection-age: 30m
    pool-timeout: 4s
    idle-timeout: 5m
    idle-check-frequency: 1m

    # Timeouts
    dial-timeout: 5s
    read-timeout: 3s
    write-timeout: 3s

    # Retry configuration
    max-retries: 3
    min-retry-backoff: 8ms
    max-retry-backoff: 512ms

    # Health checks
    health-check-interval: 30s
    health-check-timeout: 10s
```

### Redis Sentinel Configuration

```yaml
# Redis with Sentinel for failover
enhanced-locking:
  enabled: true
  backend: redis

  redis:
    # Sentinel configuration
    sentinel:
      master-name: "atlantis-master"
      addresses:
        - "sentinel-1:26379"
        - "sentinel-2:26379"
        - "sentinel-3:26379"
      password: "${SENTINEL_PASSWORD}"

    # Redis settings
    password: "${REDIS_PASSWORD}"
    db: 1  # Use specific database

    # Connection settings
    pool-size: 50
    dial-timeout: 5s
    read-timeout: 3s
    write-timeout: 3s

    # Failover settings
    max-retries: 5
    retry-delay: 200ms
    failover-timeout: 10s
```

### BoltDB Optimized Configuration

```yaml
# BoltDB with performance tuning
enhanced-locking:
  enabled: true
  backend: boltdb

  boltdb:
    # File settings
    path: "/var/lib/atlantis/enhanced-locks.db"
    file-mode: 0600

    # Performance tuning
    page-size: 4096
    no-grow-sync: false
    no-free-list-sync: false
    free-list-type: "map"  # or "array"

    # Timeouts
    timeout: 10s

    # Backup settings
    auto-backup: true
    backup-interval: 1h
    backup-retention: 72h
```

## Migration Configurations

### Shadow Mode Migration

```yaml
# Run enhanced system alongside legacy for validation
enhanced-locking:
  enabled: true

  # Migration settings
  migration:
    mode: "shadow"
    validation-enabled: true
    comparison-logging: true
    traffic-percentage: 0  # No actual traffic routing

    # Monitoring shadow operations
    shadow-metrics: true
    performance-comparison: true

  # Start with existing backend
  backend: "boltdb"
  compatibility-mode: true

  # Conservative feature set
  priority-queue: false
  deadlock-detection: false
  retries: false
  metrics: true
```

### Gradual Migration

```yaml
# Gradually route traffic to enhanced system
enhanced-locking:
  enabled: true

  migration:
    mode: "gradual"
    traffic-percentage: 25  # Start with 25% of traffic

    # Canary criteria - start with dev workspaces
    canary-criteria:
      - workspace: "development"
      - workspace: "staging"

    # Safety features
    fallback-enabled: true
    validation-enabled: true
    automatic-rollback: true
    rollback-threshold: 0.05  # 5% error rate

    # Monitoring
    migration-metrics: true
    comparison-tracking: true

  # Target configuration
  backend: "redis"
  redis:
    addresses: ["redis-1:6379", "redis-2:6379"]
    password: "${REDIS_PASSWORD}"
```

### Blue-Green Migration

```yaml
# Blue-green deployment for instant rollback
enhanced-locking:
  enabled: true

  migration:
    mode: "blue-green"
    active-system: "legacy"  # Start with legacy active

    # Health monitoring
    health-check-enabled: true
    health-check-interval: 10s

    # Automatic rollback
    automatic-rollback: true
    rollback-threshold: 0.05
    rollback-cooldown: 5m

    # Switch criteria
    switch-criteria:
      min-validation-time: 24h
      min-successful-operations: 1000
      max-error-rate: 0.01
```

## Feature-Specific Configurations

### Priority Queue Configuration

```yaml
# Advanced priority queue setup
enhanced-locking:
  enabled: true
  priority-queue: true

  queue-config:
    # Queue sizing
    max-queue-size: 1000
    overflow-policy: "reject_oldest"  # or "reject_newest", "block"

    # Processing optimization
    processing-interval: 50ms
    batch-processing: true
    batch-size: 20
    max-batch-wait: 100ms

    # Priority levels
    priority-levels: 5
    default-priority: 3

    # Anti-starvation
    anti-starvation: true
    priority-boost-interval: 60s
    max-wait-time: 300s

    # Adaptive timeouts
    adaptive-timeouts: true
    timeout-adjustment-factor: 1.5
    min-timeout: 10s
    max-timeout: 600s

    # Parallel processing
    parallel-processing: true
    worker-count: 4
    worker-queue-size: 100
```

### Deadlock Detection Configuration

```yaml
# Comprehensive deadlock detection
enhanced-locking:
  enabled: true
  deadlock-detection: true

  deadlock-config:
    # Detection settings
    enabled: true
    check-interval: 30s
    detection-algorithm: "dfs"  # or "tarjan"

    # Graph analysis
    max-graph-size: 10000
    graph-cleanup-interval: 5m
    node-timeout: 600s

    # Resolution policies
    resolution-policies:
      - "lowest_priority"
      - "youngest_first"
      - "random"
    adaptive-policy-selection: true

    # Resolution settings
    auto-resolve: true
    max-resolution-attempts: 5
    resolution-timeout: 30s

    # Anti-starvation
    victim-history-ttl: 300s
    max-preemptions-per-hour: 10
    cooldown-period: 60s

    # Monitoring
    metrics-enabled: true
    detailed-logging: true
    resolution-hooks: true
```

### Retry Configuration

```yaml
# Intelligent retry system
enhanced-locking:
  enabled: true
  retries: true

  retry-config:
    # Basic retry settings
    enabled: true
    max-attempts: 5
    base-delay: 100ms
    max-delay: 10s

    # Backoff strategies
    backoff-strategy: "exponential"  # or "linear", "fixed"
    exponential-base: 2
    jitter: true
    jitter-factor: 0.1

    # Retry conditions
    retry-on-timeout: true
    retry-on-connection-error: true
    retry-on-temporary-failure: true

    # Circuit breaker integration
    circuit-breaker: true
    failure-threshold: 5
    success-threshold: 3
    timeout: 60s
```

## Performance Tuning Examples

### High-Throughput Configuration

```yaml
# Optimized for maximum throughput
enhanced-locking:
  enabled: true
  backend: redis

  # Aggressive timeouts
  lock-timeout: 60s
  queue-timeout: 120s

  # Large connection pools
  redis:
    pool-size: 200
    min-idle-connections: 50
    max-connection-age: 10m

  # Queue optimization
  queue-config:
    max-queue-size: 5000
    processing-interval: 10ms
    batch-processing: true
    batch-size: 50
    parallel-processing: true
    worker-count: 16

  # Minimal safety checks for speed
  deadlock-detection:
    check-interval: 120s

  # Reduced metrics overhead
  enhanced-metrics: false
  event-streaming: false
```

### Low-Latency Configuration

```yaml
# Optimized for minimal latency
enhanced-locking:
  enabled: true
  backend: redis

  # Aggressive timeouts
  lock-timeout: 10s

  # Low-latency Redis settings
  redis:
    addresses: ["redis-local:6379"]  # Single local instance
    pool-size: 20
    dial-timeout: 1s
    read-timeout: 500ms
    write-timeout: 500ms

  # Fast queue processing
  queue-config:
    processing-interval: 1ms
    batch-processing: false  # Disable for lowest latency

  # Frequent deadlock checks
  deadlock-detection:
    check-interval: 5s

  # Minimal retry delays
  retry-config:
    base-delay: 10ms
    max-delay: 1s
    backoff-strategy: "linear"
```

### Memory-Optimized Configuration

```yaml
# Optimized for minimal memory usage
enhanced-locking:
  enabled: true
  backend: boltdb  # Lower memory than Redis

  # Conservative queue sizes
  queue-config:
    max-queue-size: 100
    batch-size: 5
    worker-count: 2

  # Limited graph size
  deadlock-detection:
    max-graph-size: 1000
    graph-cleanup-interval: 1m

  # Minimal metrics
  metrics: true
  enhanced-metrics: false
  event-streaming: false

  # Conservative retry limits
  retry-config:
    max-attempts: 3
```

## Monitoring and Observability

### Comprehensive Monitoring Setup

```yaml
# Full observability configuration
enhanced-locking:
  enabled: true

  # Metrics configuration
  metrics: true
  enhanced-metrics: true
  metrics-config:
    endpoint: "/metrics"
    interval: 15s
    detailed-histograms: true
    export-format: "prometheus"

  # Event streaming
  event-streaming: true
  event-config:
    buffer-size: 1000
    flush-interval: 5s
    destinations:
      - type: "webhook"
        url: "http://event-collector:8080/events"
      - type: "file"
        path: "/var/log/atlantis/events.json"

  # Health checks
  health-checks: true
  health-config:
    endpoint: "/health"
    detailed: true
    check-interval: 30s

  # Debug features
  debug-mode: false
  profiling: true
  profiling-config:
    cpu-profile: false
    memory-profile: true
    trace: false
```

### Prometheus Integration

```yaml
# Prometheus-specific configuration
enhanced-locking:
  enabled: true
  metrics: true

  # Prometheus metrics
  prometheus:
    enabled: true
    endpoint: "/metrics"
    namespace: "atlantis"
    subsystem: "enhanced_locking"

    # Custom labels
    labels:
      environment: "production"
      cluster: "main"
      version: "v1.0.0"

    # Histogram buckets
    latency-buckets: [0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0, 2.5, 5.0, 10.0]
    size-buckets: [10, 25, 50, 100, 250, 500, 1000, 2500, 5000, 10000]
```

## Security Configurations

### TLS and Authentication

```yaml
# Security-hardened configuration
enhanced-locking:
  enabled: true
  backend: redis

  # Redis security
  redis:
    # TLS configuration
    tls-enabled: true
    tls-cert-file: "/etc/ssl/certs/redis-client.crt"
    tls-key-file: "/etc/ssl/private/redis-client.key"
    tls-ca-file: "/etc/ssl/certs/redis-ca.crt"
    tls-skip-verify: false

    # Authentication
    username: "${REDIS_USERNAME}"
    password: "${REDIS_PASSWORD}"

    # Connection security
    require-auth: true

  # API security
  api-security:
    authentication-required: true
    rate-limiting: true
    rate-limit: 100  # requests per minute

  # Audit logging
  audit:
    enabled: true
    log-file: "/var/log/atlantis/enhanced-locking-audit.log"
    log-level: "info"
    include-payloads: false  # Security: don't log sensitive data
```

## Environment-Specific Examples

### Kubernetes Deployment

```yaml
# Kubernetes-optimized configuration
enhanced-locking:
  enabled: true
  backend: redis

  # Service discovery
  redis:
    addresses: ["redis-headless.atlantis:6379"]
    cluster-mode: false

  # Pod-aware settings
  queue-config:
    worker-count: ${ATLANTIS_QUEUE_WORKERS:-2}  # CPU-based scaling

  # Health checks for k8s
  health-checks: true
  health-config:
    endpoint: "/health"
    liveness-check: true
    readiness-check: true

  # Graceful shutdown
  shutdown:
    grace-period: 30s
    drain-timeout: 60s
```

### Docker Compose Setup

```yaml
# Docker Compose configuration
enhanced-locking:
  enabled: true
  backend: redis

  # Docker networking
  redis:
    addresses: ["redis:6379"]

  # Container-friendly settings
  boltdb:
    path: "/data/enhanced-locks.db"  # Volume mount

  # Logging for containers
  logging:
    level: "info"
    format: "json"
    output: "stdout"
```

### Multi-Region Setup

```yaml
# Multi-region configuration
enhanced-locking:
  enabled: true
  backend: redis

  # Regional Redis clusters
  redis:
    # Primary region
    addresses: ["redis-us-east-1:6379", "redis-us-east-2:6379"]

    # Cross-region replication
    replication:
      enabled: true
      read-preference: "nearest"
      write-concern: "majority"

  # Regional awareness
  region-config:
    current-region: "us-east-1"
    regions: ["us-east-1", "us-west-2", "eu-west-1"]
    cross-region-timeout: 10s
```

## Testing Configurations

### Load Testing Setup

```yaml
# Configuration for load testing
enhanced-locking:
  enabled: true
  backend: redis

  # Stress test settings
  queue-config:
    max-queue-size: 10000
    processing-interval: 1ms
    batch-processing: true
    batch-size: 100
    worker-count: 32

  # Aggressive timeouts for testing
  lock-timeout: 5s

  # Enhanced monitoring during tests
  metrics: true
  enhanced-metrics: true
  debug-mode: true

  # Test-specific features
  test-mode: true
  test-config:
    synthetic-load: true
    chaos-testing: true
    fault-injection: true
```

### Integration Testing

```yaml
# Integration test configuration
enhanced-locking:
  enabled: true
  backend: boltdb  # In-memory for fast tests

  boltdb:
    path: ":memory:"  # In-memory database

  # Fast timeouts for quick tests
  lock-timeout: 1s
  queue-timeout: 5s

  # Minimal features for testing
  priority-queue: true
  deadlock-detection: false
  retries: false

  # Test hooks
  test-hooks:
    enabled: true
    hook-timeout: 100ms
```

---

## Configuration Validation

Always validate your configuration before deploying:

```bash
# Validate configuration syntax
atlantis config validate

# Test configuration in dry-run mode
atlantis config test --dry-run

# Check feature compatibility
atlantis config check-features
```

For more configuration options and details, refer to the [Enhanced Locking Documentation](../README.md).