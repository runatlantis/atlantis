# Enhanced Locking System - Migration Guide

## Overview

This guide provides step-by-step procedures for migrating from the legacy locking system to the Enhanced Locking System. The migration is designed to be safe, gradual, and reversible at any stage.

## Pre-Migration Checklist

### 1. Environment Assessment

Before starting migration, assess your current environment:

```bash
# Check current Atlantis version
atlantis version

# Analyze current lock usage
atlantis locks list --format=json > current-locks.json

# Check configuration
atlantis config show > current-config.yaml

# Review system resources
free -h
df -h
```

### 2. Backup Current State

**Critical: Always backup before migration**

```bash
# Backup BoltDB locks database
cp /var/lib/atlantis/atlantis.db /var/lib/atlantis/atlantis.db.backup.$(date +%Y%m%d_%H%M%S)

# Backup configuration
cp /etc/atlantis/atlantis.yaml /etc/atlantis/atlantis.yaml.backup.$(date +%Y%m%d_%H%M%S)

# Export current locks
atlantis locks export --output=/tmp/locks-backup-$(date +%Y%m%d_%H%M%S).json
```

### 3. Prerequisites Check

Verify system requirements:

```bash
# Check Go version (1.19+ required)
go version

# Check available memory (minimum 2GB recommended)
free -g

# Check disk space (additional 1GB recommended)
df -h /var/lib/atlantis

# Check network connectivity (for Redis)
ping redis-server.example.com
```

### 4. Staging Environment Setup

**Mandatory: Test migration in staging first**

```bash
# Set up staging environment with production data copy
rsync -av production:/var/lib/atlantis/ staging:/var/lib/atlantis/

# Configure staging environment
cp production-config.yaml staging-config.yaml
```

## Migration Strategies

### Strategy 1: Shadow Mode Migration (Recommended)

This is the safest approach - runs enhanced system alongside legacy for validation.

#### Phase 1: Enable Shadow Mode

```yaml
# atlantis.yaml
enhanced-locking:
  enabled: true
  migration:
    mode: "shadow"
    validation-enabled: true
    comparison-logging: true
    traffic-percentage: 0  # No actual traffic routing

  # Use existing backend initially
  backend: "boltdb"
  compatibility-mode: true

  # Enable safe monitoring features only
  metrics: true
  health-checks: true
```

```bash
# Apply configuration
atlantis config reload

# Monitor shadow mode operation
tail -f /var/log/atlantis/atlantis.log | grep "enhanced-locking"
```

#### Phase 2: Validation and Monitoring

Monitor for 24-48 hours to ensure stability:

```bash
# Check shadow mode metrics
curl -s http://atlantis:4141/api/enhanced-locks/shadow/metrics | jq .

# Compare performance
curl -s http://atlantis:4141/api/enhanced-locks/comparison/performance

# Review validation results
curl -s http://atlantis:4141/api/enhanced-locks/validation/results
```

Expected metrics:
- `shadow_operations_total` > 1000 (sufficient test coverage)
- `shadow_validation_success_rate` > 99.5%
- `shadow_performance_ratio` between 0.8-1.2

#### Phase 3: Gradual Traffic Routing

Once shadow mode validates successfully:

```yaml
# atlantis.yaml
enhanced-locking:
  enabled: true
  migration:
    mode: "gradual"
    traffic-percentage: 5  # Start with 5%
    canary-criteria:
      - workspace: "development"  # Start with dev workspaces
    fallback-enabled: true
    validation-enabled: true
```

Gradually increase traffic percentage:
- Week 1: 5% traffic
- Week 2: 25% traffic
- Week 3: 50% traffic
- Week 4: 100% traffic

### Strategy 2: Blue-Green Migration

For environments requiring instant rollback capability:

#### Phase 1: Prepare Green Environment

```yaml
# green-config.yaml
enhanced-locking:
  enabled: true
  migration:
    mode: "blue-green"
    active-system: "legacy"  # Start with legacy active
    health-check-enabled: true
    automatic-rollback: true
    rollback-threshold: 0.05  # 5% error rate triggers rollback
```

#### Phase 2: Switch to Green

```bash
# Switch to enhanced system
atlantis config set enhanced-locking.migration.active-system=enhanced

# Monitor for issues
while true; do
  health=$(curl -s http://atlantis:4141/api/enhanced-locks/health | jq .healthy)
  echo "Health: $health"
  sleep 10
done
```

#### Phase 3: Rollback if Needed

```bash
# Emergency rollback
atlantis config set enhanced-locking.migration.active-system=legacy

# Or disable entirely
atlantis config set enhanced-locking.enabled=false
```

## Backend Migration

### From BoltDB to Redis

#### Phase 1: Redis Setup

```bash
# Set up Redis cluster (3 nodes minimum)
docker-compose up -d redis-cluster

# Verify cluster health
redis-cli --cluster check redis-1:6379
```

#### Phase 2: Dual Backend Configuration

```yaml
# atlantis.yaml - Run both backends temporarily
enhanced-locking:
  enabled: true
  backend: "redis"

  # Redis configuration
  redis:
    addresses: ["redis-1:6379", "redis-2:6379", "redis-3:6379"]
    password: "${REDIS_PASSWORD}"
    cluster-mode: true
    pool-size: 20

  # Migration settings
  migration:
    mode: "backend-migration"
    source-backend: "boltdb"
    target-backend: "redis"
    sync-enabled: true
```

#### Phase 3: Data Migration

```bash
# Export locks from BoltDB
atlantis locks export --backend=boltdb --output=boltdb-locks.json

# Import to Redis
atlantis locks import --backend=redis --input=boltdb-locks.json

# Verify migration
atlantis locks verify --source=boltdb --target=redis
```

#### Phase 4: Switch to Redis

```yaml
# atlantis.yaml - Redis only
enhanced-locking:
  enabled: true
  backend: "redis"
  redis:
    addresses: ["redis-1:6379", "redis-2:6379", "redis-3:6379"]
    password: "${REDIS_PASSWORD}"
    cluster-mode: true
```

## Feature Enablement

### Enable Features Gradually

#### Week 1: Foundation
```yaml
enhanced-locking:
  enabled: true
  backend: "boltdb"
  metrics: true
  health-checks: true
```

#### Week 2: Enhanced Metrics
```yaml
enhanced-locking:
  enabled: true
  enhanced-metrics: true
  event-streaming: true
```

#### Week 3: Priority Queuing
```yaml
enhanced-locking:
  enabled: true
  priority-queue: true
  queue-config:
    max-queue-size: 100  # Start conservative
    batch-processing: false  # Enable later
```

#### Week 4: Advanced Features
```yaml
enhanced-locking:
  enabled: true
  priority-queue: true
  deadlock-detection: true
  retries: true

  # Advanced queue features
  queue-config:
    batch-processing: true
    adaptive-timeouts: true
```

## Migration Validation

### Automated Validation

Create validation scripts:

```bash
#!/bin/bash
# validate-migration.sh

echo "Running migration validation..."

# Test basic lock operations
echo "Testing basic operations..."
atlantis locks test --operations=100 --concurrent=10

# Compare performance
echo "Comparing performance..."
LEGACY_TIME=$(atlantis locks benchmark --backend=legacy --iterations=100 | grep "avg" | awk '{print $2}')
ENHANCED_TIME=$(atlantis locks benchmark --backend=enhanced --iterations=100 | grep "avg" | awk '{print $2}')

RATIO=$(echo "scale=2; $ENHANCED_TIME / $LEGACY_TIME" | bc)
echo "Performance ratio (enhanced/legacy): $RATIO"

if (( $(echo "$RATIO > 1.5" | bc -l) )); then
    echo "ERROR: Performance degradation > 50%"
    exit 1
fi

# Test concurrency
echo "Testing concurrency..."
atlantis locks stress-test --duration=60s --concurrent-users=50

echo "Validation completed successfully"
```

### Manual Validation

#### Test Scenarios

1. **Basic Operations**
```bash
# Test lock acquisition
atlantis lock test-project development

# Test lock release
atlantis unlock test-project development

# Test lock listing
atlantis locks list
```

2. **Concurrency Testing**
```bash
# Simulate multiple users
for i in {1..10}; do
  atlantis lock test-project workspace-$i &
done
wait

# Check for deadlocks or errors
atlantis locks list --show-errors
```

3. **Error Handling**
```bash
# Test Redis failure scenario
docker stop redis-1

# Verify fallback behavior
atlantis lock test-project development

# Restore Redis
docker start redis-1
```

## Monitoring During Migration

### Key Metrics to Watch

```bash
# Monitor continuously during migration
watch -n 5 'curl -s http://atlantis:4141/api/enhanced-locks/metrics | jq "{
  locks_active: .locks_active,
  queue_depth: .queue_depth,
  error_rate: .error_rate,
  avg_latency: .avg_latency_ms,
  backend_health: .backend_health
}"'
```

### Alert Thresholds

Set up alerts for:

| Metric | Warning | Critical |
|--------|---------|----------|
| Error Rate | > 1% | > 5% |
| Queue Depth | > 50 | > 100 |
| Average Latency | > 500ms | > 1000ms |
| Memory Usage | > 80% | > 90% |
| Backend Health | < 95% | < 90% |

### Grafana Dashboard

Import enhanced locking dashboard:

```bash
# Import dashboard
curl -X POST http://grafana:3000/api/dashboards/db \
  -H "Content-Type: application/json" \
  -d @enhanced-locking-dashboard.json
```

## Rollback Procedures

### Automatic Rollback

Configure automatic rollback triggers:

```yaml
enhanced-locking:
  migration:
    rollback:
      enabled: true
      triggers:
        error-rate: 0.05  # 5%
        latency-increase: 2.0  # 2x increase
        health-score: 0.8  # Below 80%
      cooldown: "5m"  # Wait before re-enabling
```

### Manual Rollback

#### Emergency Rollback (Immediate)

```bash
# Disable enhanced locking immediately
atlantis config set enhanced-locking.enabled=false
atlantis restart

# Verify rollback
curl -s http://atlantis:4141/api/locks/health
```

#### Graceful Rollback

```bash
# Reduce traffic gradually
atlantis config set enhanced-locking.migration.traffic-percentage=50
sleep 60
atlantis config set enhanced-locking.migration.traffic-percentage=25
sleep 60
atlantis config set enhanced-locking.migration.traffic-percentage=0
sleep 60

# Disable enhanced system
atlantis config set enhanced-locking.enabled=false
```

#### Data Recovery

If data issues occur:

```bash
# Restore from backup
systemctl stop atlantis
cp /var/lib/atlantis/atlantis.db.backup.* /var/lib/atlantis/atlantis.db
systemctl start atlantis

# Verify data integrity
atlantis locks verify --check-integrity
```

## Post-Migration Tasks

### 1. Performance Optimization

After successful migration:

```yaml
# Optimize configuration
enhanced-locking:
  enabled: true
  backend: "redis"

  # Enable performance features
  priority-queue: true
  queue-config:
    batch-processing: true
    processing-interval: "50ms"

  # Enable advanced features
  deadlock-detection: true
  adaptive-timeouts: true
```

### 2. Cleanup Legacy Components

```bash
# Remove legacy configuration
sed -i '/^# Legacy locking/,/^$/d' /etc/atlantis/atlantis.yaml

# Archive legacy data
tar -czf legacy-locks-archive-$(date +%Y%m%d).tar.gz /var/lib/atlantis/atlantis.db.backup.*
```

### 3. Documentation Update

Update runbooks and documentation:

```bash
# Update operational procedures
vim /docs/operations/locking-procedures.md

# Update monitoring runbooks
vim /docs/monitoring/enhanced-locking-alerts.md

# Update backup procedures
vim /docs/backup/enhanced-locking-backup.md
```

### 4. Team Training

Conduct training sessions covering:
- New monitoring dashboards
- Enhanced troubleshooting procedures
- New configuration options
- Performance tuning guidelines

## Troubleshooting Migration Issues

### Common Issues and Solutions

#### Issue: Performance Degradation

**Symptoms:**
- Slower lock acquisition
- Higher memory usage
- Timeout errors

**Solution:**
```bash
# Check resource usage
top -p $(pgrep atlantis)

# Increase memory allocation
export ATLANTIS_GOMAXPROCS=4
export ATLANTIS_GOMAXMEMORY=2GB

# Optimize Redis connection pool
atlantis config set enhanced-locking.redis.pool-size=50
```

#### Issue: Configuration Errors

**Symptoms:**
- Startup failures
- Invalid configuration warnings
- Feature not working

**Solution:**
```bash
# Validate configuration
atlantis config validate

# Check configuration syntax
yamllint /etc/atlantis/atlantis.yaml

# Use configuration wizard
atlantis config wizard --enhanced-locking
```

#### Issue: Data Synchronization

**Symptoms:**
- Lock state inconsistencies
- Missing locks
- Duplicate locks

**Solution:**
```bash
# Run data synchronization
atlantis locks sync --source=boltdb --target=redis

# Repair inconsistencies
atlantis locks repair --dry-run
atlantis locks repair --execute

# Verify data integrity
atlantis locks verify --full-check
```

## Migration Timeline Template

### Week 1: Preparation
- [ ] Environment assessment
- [ ] Backup current state
- [ ] Staging environment setup
- [ ] Team training on new system

### Week 2: Shadow Mode
- [ ] Enable shadow mode
- [ ] Monitor for 48 hours
- [ ] Validate performance parity
- [ ] Address any issues found

### Week 3: Gradual Rollout
- [ ] Start with 5% traffic
- [ ] Increase to 25% traffic
- [ ] Monitor error rates and performance
- [ ] Increase to 50% traffic

### Week 4: Full Migration
- [ ] Complete traffic migration (100%)
- [ ] Monitor for 72 hours
- [ ] Enable advanced features
- [ ] Complete post-migration tasks

### Week 5: Optimization
- [ ] Performance tuning
- [ ] Enable all desired features
- [ ] Cleanup legacy components
- [ ] Update documentation

## Support and Resources

For migration support:

- **Documentation**: [Enhanced Locking Documentation](../README.md)
- **Troubleshooting**: [Troubleshooting Guide](troubleshooting.md)
- **Configuration**: [Configuration Examples](../examples/configuration-examples.md)
- **Community**: Atlantis community forums
- **Emergency**: Contact Atlantis maintainers

---

Remember: **Migration should always be gradual, monitored, and reversible.** Never rush the process, and always have a rollback plan ready.