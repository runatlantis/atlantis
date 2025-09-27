# Atlantis Locking Configuration Reference

A comprehensive guide to configuring both legacy and enhanced locking systems in Atlantis, with detailed `atlantis.yaml` integration patterns and real-world deployment scenarios.

## Overview

Atlantis provides two locking systems:

1. **Legacy Locking**: BoltDB-based locking with simple directory/workspace isolation
2. **Enhanced Locking**: Redis-based distributed locking with priority queues, deadlock detection, and advanced features

Both systems integrate seamlessly with `atlantis.yaml` project configurations and parallel execution settings.

## Locking System Comparison

| Feature | Legacy Locking | Enhanced Locking |
|---------|----------------|------------------|
| Storage Backend | BoltDB (local file) | Redis (distributed) |
| Horizontal Scaling | Single instance only | Redis Cluster support |
| Priority Queuing | No | Yes (4 priority levels) |
| Deadlock Detection | No | Yes (with multiple resolution policies) |
| Timeout Management | Basic | Advanced with TTL and retry logic |
| Performance | ~1000 ops/sec | ~50,000 ops/sec |
| Multi-tenancy | Manual isolation | Built-in workspace isolation |
| Disaster Recovery | File backups | Redis persistence + clustering |

## Server Configuration

### Legacy Locking (Default)

```yaml
# server.yaml or environment variables
locking-db-type: "boltdb"  # Default
data-dir: "~/.atlantis"    # Where BoltDB file is stored

# Optional: Disable repository locking entirely
disable-repo-locking: false  # Default: false
```

Environment variables:
```bash
ATLANTIS_LOCKING_DB_TYPE=boltdb
ATLANTIS_DATA_DIR=/var/atlantis-data
ATLANTIS_DISABLE_REPO_LOCKING=false
```

### Enhanced Locking with Redis

```yaml
# server.yaml
locking-db-type: "redis"
redis-host: "localhost"
redis-port: 6379
redis-password: "your-redis-password"
redis-db: 0
redis-tls-enabled: true
redis-insecure-skip-verify: false

# Enhanced locking specific configuration
enhanced-locking:
  enabled: true
  backend: "redis"
  default-timeout: "30m"
  max-timeout: "2h"

  # Priority queue configuration
  priority-queue:
    enabled: true
    max-queue-size: 1000
    queue-timeout: "10m"

  # Retry mechanism
  retry:
    enabled: true
    max-attempts: 3
    base-delay: "1s"
    max-delay: "30s"

  # Deadlock detection
  deadlock-detection:
    enabled: true
    check-interval: "30s"
    resolution-policy: "lowest_priority"  # or "oldest_lock", "random", "abort"

  # Redis-specific settings
  redis:
    cluster-mode: false
    key-prefix: "atlantis:lock:"
    lock-ttl: "1h"

  # Backward compatibility
  backward-compatibility:
    legacy-fallback: true
    preserve-legacy-format: true
```

Environment variables:
```bash
ATLANTIS_LOCKING_DB_TYPE=redis
ATLANTIS_REDIS_HOST=redis.example.com
ATLANTIS_REDIS_PORT=6379
ATLANTIS_REDIS_PASSWORD=secure-password
ATLANTIS_REDIS_TLS_ENABLED=true
```

## atlantis.yaml Integration

### Basic Project Locking Configuration

```yaml
version: 3
projects:
  - name: frontend
    dir: frontend
    workspace: production
    # Legacy repo_locking (deprecated)
    repo_locking: true

    # Enhanced repo_locks configuration
    repo_locks:
      mode: on_plan  # Options: disabled, on_plan, on_apply

  - name: backend
    dir: backend
    workspace: production
    repo_locks:
      mode: on_apply  # Only lock during apply operations
```

### Locking Modes Explained

| Mode | Description | When Lock is Acquired | When Lock is Released |
|------|-------------|----------------------|---------------------|
| `disabled` | No locking | Never | N/A |
| `on_plan` | Lock on plan (default) | During `atlantis plan` | PR merge/close or manual unlock |
| `on_apply` | Lock on apply only | During `atlantis apply` | After apply completion |

### Parallel Execution with Locking

```yaml
version: 3
# Global parallel settings
parallel_plan: true
parallel_apply: true
parallel-pool-size: 15  # Server setting

projects:
  # These will run in parallel within their execution order group
  - name: shared-infrastructure
    dir: infrastructure/shared
    execution_order_group: 1  # Runs first
    repo_locks:
      mode: on_plan

  - name: vpc-dev
    dir: infrastructure/vpc
    workspace: development
    execution_order_group: 2  # Runs after group 1
    repo_locks:
      mode: on_plan

  - name: vpc-prod
    dir: infrastructure/vpc
    workspace: production
    execution_order_group: 2  # Runs in parallel with vpc-dev
    repo_locks:
      mode: on_apply  # Only lock during apply

  - name: app-dev
    dir: applications/web
    workspace: development
    execution_order_group: 3
    depends_on: ["vpc-dev"]  # Explicit dependency
    repo_locks:
      mode: on_plan

  - name: app-prod
    dir: applications/web
    workspace: production
    execution_order_group: 3
    depends_on: ["vpc-prod"]
    repo_locks:
      mode: on_apply
```

### Multi-Workspace Project Locking

```yaml
version: 3
projects:
  # Each workspace gets its own lock
  - name: database-dev
    dir: database
    workspace: development
    repo_locks:
      mode: on_plan

  - name: database-staging
    dir: database
    workspace: staging
    repo_locks:
      mode: on_plan

  - name: database-prod
    dir: database
    workspace: production
    repo_locks:
      mode: on_apply  # Production requires apply-time locking
    plan_requirements: [approved]
    apply_requirements: [approved]
```

## Environment-Specific Configurations

### Development Environment

Focus on fast iteration with minimal locking overhead:

```yaml
# atlantis.yaml for development
version: 3
parallel_plan: true
parallel_apply: false  # Safer for dev environment

projects:
  - name: dev-api
    dir: api
    workspace: development
    repo_locks:
      mode: disabled  # No locking in dev for speed

  - name: dev-frontend
    dir: frontend
    workspace: development
    repo_locks:
      mode: on_plan  # Light locking
```

Server configuration:
```yaml
# Lightweight locking for dev
enhanced-locking:
  enabled: false  # Use legacy for simplicity
```

### Staging Environment

Balance between speed and safety:

```yaml
# atlantis.yaml for staging
version: 3
parallel_plan: true
parallel_apply: true
abort_on_execution_order_fail: true  # Fail fast

projects:
  - name: staging-infrastructure
    dir: infrastructure
    workspace: staging
    execution_order_group: 1
    repo_locks:
      mode: on_plan

  - name: staging-applications
    dir: applications
    workspace: staging
    execution_order_group: 2
    depends_on: ["staging-infrastructure"]
    repo_locks:
      mode: on_plan
```

Server configuration:
```yaml
enhanced-locking:
  enabled: true
  priority-queue:
    enabled: false  # Keep simple
  deadlock-detection:
    enabled: true   # Safety net
```

### Production Environment

Maximum safety with advanced locking features:

```yaml
# atlantis.yaml for production
version: 3
parallel_plan: false   # Conservative planning
parallel_apply: false  # Sequential applies for safety
abort_on_execution_order_fail: true

projects:
  - name: prod-networking
    dir: infrastructure/networking
    workspace: production
    execution_order_group: 1
    repo_locks:
      mode: on_apply  # Lock only during destructive operations
    plan_requirements: [approved, mergeable]
    apply_requirements: [approved, mergeable]

  - name: prod-security
    dir: infrastructure/security
    workspace: production
    execution_order_group: 1
    repo_locks:
      mode: on_apply
    plan_requirements: [approved, mergeable]
    apply_requirements: [approved, mergeable]

  - name: prod-compute
    dir: infrastructure/compute
    workspace: production
    execution_order_group: 2
    depends_on: ["prod-networking", "prod-security"]
    repo_locks:
      mode: on_apply
    plan_requirements: [approved, mergeable]
    apply_requirements: [approved, mergeable]

  - name: prod-applications
    dir: applications
    workspace: production
    execution_order_group: 3
    depends_on: ["prod-compute"]
    repo_locks:
      mode: on_apply
    plan_requirements: [approved, mergeable]
    apply_requirements: [approved, mergeable]
```

Server configuration:
```yaml
enhanced-locking:
  enabled: true
  default-timeout: "45m"  # Longer timeouts for production

  priority-queue:
    enabled: true
    max-queue-size: 100   # Conservative queue size

  retry:
    enabled: true
    max-attempts: 5       # More retries for reliability

  deadlock-detection:
    enabled: true
    resolution-policy: "abort"  # Conservative resolution

  redis:
    cluster-mode: true    # High availability
    lock-ttl: "2h"        # Longer TTL for production processes
```

## Advanced Locking Patterns

### Cross-Environment Dependencies

```yaml
version: 3
projects:
  # Shared resources that affect multiple environments
  - name: shared-dns
    dir: shared/dns
    workspace: global
    execution_order_group: 1
    repo_locks:
      mode: on_apply  # Critical shared resource

  # Development environment depends on shared resources
  - name: dev-app
    dir: applications/api
    workspace: development
    execution_order_group: 2
    # Note: depends_on only works within same repo
    # Use manual coordination for cross-repo dependencies
    repo_locks:
      mode: on_plan

  # Production environment with strict dependencies
  - name: prod-app
    dir: applications/api
    workspace: production
    execution_order_group: 3
    repo_locks:
      mode: on_apply
    plan_requirements: [approved, mergeable]
    apply_requirements: [approved, mergeable]
```

### High-Throughput Scenarios

For environments with many concurrent operations:

```yaml
version: 3
# Enable full parallelization
parallel_plan: true
parallel_apply: true

projects:
  # Microservices with independent locking
  - name: service-a-dev
    dir: services/service-a
    workspace: development
    repo_locks:
      mode: on_plan

  - name: service-a-prod
    dir: services/service-a
    workspace: production
    repo_locks:
      mode: on_apply

  - name: service-b-dev
    dir: services/service-b
    workspace: development
    repo_locks:
      mode: on_plan

  - name: service-b-prod
    dir: services/service-b
    workspace: production
    repo_locks:
      mode: on_apply
```

Server configuration for high throughput:
```yaml
parallel-pool-size: 50  # Increase from default of 15

enhanced-locking:
  enabled: true

  priority-queue:
    enabled: true
    max-queue-size: 5000  # Large queue for high volume
    queue-timeout: "5m"   # Shorter timeout for fast turnover

  retry:
    enabled: true
    max-attempts: 2       # Fewer retries for speed
    base-delay: "500ms"   # Faster retries

  redis:
    cluster-mode: true    # Scale horizontally
    key-prefix: "atlantis:ht:lock:"  # Separate namespace
```

### Disaster Recovery Configuration

```yaml
# Primary site configuration
enhanced-locking:
  enabled: true

  redis:
    cluster-mode: true
    # Primary Redis cluster
    nodes:
      - "redis-1.primary.com:6379"
      - "redis-2.primary.com:6379"
      - "redis-3.primary.com:6379"

  backup:
    enabled: true
    interval: "5m"
    retention: "7d"
    destination: "s3://atlantis-backups/locks/"

  replication:
    enabled: true
    # Async replication to DR site
    targets:
      - "redis-cluster.dr.com:6379"
```

atlantis.yaml for DR scenarios:
```yaml
version: 3
# Shorter timeouts for faster failover
projects:
  - name: critical-infrastructure
    dir: infrastructure/critical
    workspace: production
    repo_locks:
      mode: on_apply
    # Shorter timeout for faster DR recovery
    timeout: "20m"  # Custom workflow timeout

workflows:
  production:
    plan:
      steps:
        - init
        - plan
    apply:
      steps:
        # Add health checks before apply
        - run: ./scripts/pre-apply-health-check.sh
        - apply
        # Verify after apply
        - run: ./scripts/post-apply-health-check.sh
```

## Troubleshooting Common Issues

### Lock Contention

When multiple teams/PRs compete for the same resources:

```yaml
# Use enhanced locking with priority queues
version: 3
projects:
  - name: high-priority-security-update
    dir: security
    workspace: production
    # In enhanced locking, this would get PriorityHigh
    repo_locks:
      mode: on_apply
    # Add custom workflow with priority hints
    workflow: high-priority

workflows:
  high-priority:
    plan:
      steps:
        - run: echo "HIGH PRIORITY - Security Update"
        - init
        - plan
```

### Deadlock Prevention

```yaml
version: 3
# Use consistent ordering to prevent deadlocks
projects:
  - name: network-prod
    dir: network
    workspace: production
    execution_order_group: 1  # Always first

  - name: compute-prod
    dir: compute
    workspace: production
    execution_order_group: 2  # Always second
    depends_on: ["network-prod"]

  - name: app-prod
    dir: applications
    workspace: production
    execution_order_group: 3  # Always third
    depends_on: ["compute-prod"]
```

### Lock Timeout Configuration

```yaml
version: 3
projects:
  - name: long-running-migration
    dir: database/migrations
    workspace: production
    repo_locks:
      mode: on_apply
    # Use custom workflow with longer timeout
    workflow: long-migration

workflows:
  long-migration:
    apply:
      steps:
        - init
        - run: echo "Starting long migration - may take 60+ minutes"
        - apply
        - run: echo "Migration completed successfully"
```

## Migration Strategies

### Legacy to Enhanced Migration

Phase 1: Enable enhanced locking alongside legacy
```yaml
enhanced-locking:
  enabled: true
  backward-compatibility:
    legacy-fallback: true    # Fall back if enhanced fails
    preserve-legacy-format: true
```

Phase 2: Test with shadow mode
```yaml
enhanced-locking:
  enabled: true
  shadow-mode: true  # Test enhanced without using it
```

Phase 3: Full migration
```yaml
enhanced-locking:
  enabled: true
  backward-compatibility:
    legacy-fallback: false   # Pure enhanced mode
```

### Zero-Downtime Migration atlantis.yaml

```yaml
version: 3
# During migration, be conservative
parallel_plan: false
parallel_apply: false

projects:
  - name: migration-test
    dir: test
    workspace: development
    repo_locks:
      mode: on_plan  # Test enhanced locking first

  # Keep production conservative during migration
  - name: prod-critical
    dir: production
    workspace: production
    repo_locks:
      mode: on_apply
    plan_requirements: [approved]
    apply_requirements: [approved]
```

## Performance Tuning

### Memory-Optimized Configuration

```yaml
enhanced-locking:
  redis:
    lock-ttl: "15m"       # Shorter TTL for faster cleanup
    key-prefix: "atl:l:"  # Shorter prefix saves memory

  priority-queue:
    max-queue-size: 100   # Smaller queue

  deadlock-detection:
    check-interval: "60s" # Less frequent checks
```

### CPU-Optimized Configuration

```yaml
enhanced-locking:
  retry:
    enabled: false        # Disable retries to reduce CPU

  deadlock-detection:
    enabled: false        # Disable for CPU savings

  priority-queue:
    enabled: false        # Simple FIFO queue
```

### Network-Optimized Configuration

```yaml
enhanced-locking:
  redis:
    cluster-mode: false   # Single Redis for lower latency
    compression: true     # Enable if available

  batch-operations: true  # Batch lock operations
```

## Security Considerations

### Redis Security

```yaml
enhanced-locking:
  redis:
    tls-enabled: true
    cert-file: "/etc/atlantis/redis-client.crt"
    key-file: "/etc/atlantis/redis-client.key"
    ca-file: "/etc/atlantis/redis-ca.crt"
    username: "atlantis"
    password-file: "/etc/atlantis/redis-password"
```

### Access Control Integration

```yaml
version: 3
projects:
  - name: security-infrastructure
    dir: security
    workspace: production
    repo_locks:
      mode: on_apply
    # Require team-specific approval
    plan_requirements: [approved]
    apply_requirements: [approved]
    # Limit to security team members only
    # (configured in server-side repo config)
```

This comprehensive configuration reference provides the foundation for implementing both legacy and enhanced locking systems in any Atlantis deployment, from simple single-tenant setups to complex multi-tenant, high-availability environments.