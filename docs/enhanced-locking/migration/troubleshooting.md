# Enhanced Locking System - Troubleshooting Guide

## Overview

This troubleshooting guide provides solutions for common issues encountered when deploying, migrating to, or operating the Enhanced Locking System in production environments.

## Quick Diagnostic Commands

### System Health Check

```bash
# Check overall system health
curl -s http://atlantis:4141/api/enhanced-locks/health | jq .

# Get basic metrics
curl -s http://atlantis:4141/api/enhanced-locks/metrics | jq .

# Check configuration status
atlantis config validate enhanced-locking

# Test basic lock operations
atlantis locks test --basic
```

### Log Analysis

```bash
# Check recent enhanced locking logs
journalctl -u atlantis --since="1 hour ago" | grep "enhanced-locking"

# Monitor real-time logs
tail -f /var/log/atlantis/atlantis.log | grep -E "(enhanced-locking|deadlock|priority-queue)"

# Check error patterns
grep -E "(ERROR|FATAL)" /var/log/atlantis/atlantis.log | grep enhanced-locking | tail -20
```

## Common Issues and Solutions

### 1. Enhanced Locking System Won't Start

#### Symptoms
- Atlantis starts but enhanced locking features are not available
- Error: "enhanced locking system disabled"
- Health check fails

#### Possible Causes & Solutions

**Configuration Error**
```bash
# Check configuration syntax
atlantis config validate

# Common fix: Enable the system
atlantis config set enhanced-locking.enabled=true

# Restart Atlantis
systemctl restart atlantis
```

**Missing Dependencies**
```bash
# Check Redis connectivity (if using Redis backend)
redis-cli -h redis-server ping

# Test Redis authentication
redis-cli -h redis-server -a $REDIS_PASSWORD ping

# Check BoltDB permissions (if using BoltDB)
ls -la /var/lib/atlantis/atlantis.db
chown atlantis:atlantis /var/lib/atlantis/atlantis.db
```

**Resource Constraints**
```bash
# Check memory usage
free -h
ps aux | grep atlantis

# Check disk space
df -h /var/lib/atlantis

# Increase memory limits
export ATLANTIS_GOMAXPROCS=4
export ATLANTIS_GOMAXMEMORY=4GB
```

### 2. Redis Connection Issues

#### Symptoms
- "Redis connection failed" errors
- Locks timing out
- Inconsistent lock state

#### Solutions

**Connection Pool Issues**
```yaml
# Increase connection pool size
enhanced-locking:
  redis:
    pool-size: 50
    max-retries: 3
    retry-delay: 1s
    dial-timeout: 5s
    read-timeout: 3s
    write-timeout: 3s
```

**Cluster Configuration**
```bash
# Verify cluster status
redis-cli --cluster check redis-1:6379

# Check cluster nodes
redis-cli --cluster nodes redis-1:6379

# Test failover
redis-cli --cluster failover redis-1:6379
```

**Authentication Issues**
```bash
# Test authentication
redis-cli -h redis-1 -p 6379 -a $REDIS_PASSWORD ping

# Check password environment variable
echo $REDIS_PASSWORD

# Update configuration
atlantis config set enhanced-locking.redis.password="$REDIS_PASSWORD"
```

**Network Issues**
```bash
# Test connectivity
telnet redis-1 6379

# Check firewall rules
iptables -L | grep 6379

# Test from Atlantis host
curl -v telnet://redis-1:6379
```

### 3. Performance Issues

#### Symptoms
- Slow lock acquisition
- High memory usage
- Timeout errors

#### Solutions

**Lock Acquisition Timeouts**
```yaml
# Increase timeouts
enhanced-locking:
  lock-timeout: 30s
  queue-timeout: 60s

  # Optimize queue processing
  priority-queue: true
  queue-config:
    processing-interval: 10ms
    batch-size: 10
```

**Memory Usage Issues**
```bash
# Monitor memory usage
top -p $(pgrep atlantis)

# Check for memory leaks
valgrind --tool=memcheck atlantis server

# Optimize garbage collection
export GOGC=80
export GOMEMLIMIT=2GiB
```

**High CPU Usage**
```yaml
# Reduce check intervals
enhanced-locking:
  deadlock-detection:
    check-interval: 60s

  # Optimize priority queue
  priority-queue:
    heap-optimization: true
    lazy-evaluation: true
```

### 4. Deadlock Detection Issues

#### Symptoms
- False positive deadlock alerts
- Deadlocks not being detected
- Resolution failures

#### Solutions

**False Positives**
```yaml
# Adjust detection sensitivity
enhanced-locking:
  deadlock-detection:
    check-interval: 45s
    min-cycle-length: 3
    confidence-threshold: 0.8
```

**Detection Not Working**
```bash
# Enable debug logging
export ATLANTIS_LOG_LEVEL=debug

# Check detector status
curl -s http://atlantis:4141/api/enhanced-locks/deadlock/status

# Manual deadlock test
atlantis locks deadlock-test --users=5 --resources=3
```

**Resolution Failures**
```yaml
# Configure multiple resolution policies
enhanced-locking:
  deadlock-detection:
    resolution-policies:
      - "lowest_priority"
      - "youngest_first"
      - "random"
    max-resolution-attempts: 5
```

### 5. Migration Issues

#### Symptoms
- Data inconsistencies between backends
- Migration stuck or failing
- Lock state corruption

#### Solutions

**Data Synchronization Issues**
```bash
# Check data consistency
atlantis locks verify --source=boltdb --target=redis

# Force resynchronization
atlantis locks sync --source=boltdb --target=redis --force

# Repair inconsistencies
atlantis locks repair --backend=redis --dry-run
atlantis locks repair --backend=redis --execute
```

**Migration Stuck**
```bash
# Check migration status
curl -s http://atlantis:4141/api/enhanced-locks/migration/status

# Reset migration state
atlantis migration reset --confirm

# Manual failover
atlantis config set enhanced-locking.migration.active-system=legacy
```

**Data Corruption**
```bash
# Restore from backup
systemctl stop atlantis
cp /var/lib/atlantis/atlantis.db.backup.* /var/lib/atlantis/atlantis.db
systemctl start atlantis

# Verify data integrity
atlantis locks verify --check-integrity
```

### 6. Queue Management Issues

#### Symptoms
- Queue depth growing continuously
- Requests timing out in queue
- Unfair queue processing

#### Solutions

**Queue Overflow**
```yaml
# Increase queue limits
enhanced-locking:
  priority-queue: true
  queue-config:
    max-queue-size: 1000
    overflow-policy: "reject_oldest"

  # Enable batch processing
  batch-processing: true
  batch-size: 20
```

**Queue Starvation**
```yaml
# Enable anti-starvation
enhanced-locking:
  priority-queue:
    anti-starvation: true
    priority-boost-interval: 30s
    max-wait-time: 300s
```

**Processing Bottlenecks**
```bash
# Monitor queue metrics
watch -n 1 'curl -s http://atlantis:4141/api/enhanced-locks/queue/metrics'

# Increase processing threads
export ATLANTIS_QUEUE_WORKERS=8

# Enable parallel processing
atlantis config set enhanced-locking.queue-config.parallel-processing=true
```

### 7. Monitoring and Alerting Issues

#### Symptoms
- Missing metrics
- Alerts not firing
- Dashboard showing no data

#### Solutions

**Metrics Not Available**
```yaml
# Enable metrics collection
enhanced-locking:
  metrics: true
  enhanced-metrics: true

  # Configure Prometheus
  metrics-config:
    endpoint: "/metrics"
    interval: 30s
```

**Grafana Dashboard Issues**
```bash
# Import dashboard
curl -X POST http://grafana:3000/api/dashboards/db \
  -H "Content-Type: application/json" \
  -d @enhanced-locking-dashboard.json

# Check data source
curl -s http://grafana:3000/api/datasources

# Test Prometheus connection
curl -s http://prometheus:9090/api/v1/query?query=atlantis_locks_total
```

**Alert Configuration**
```yaml
# Prometheus alerts
groups:
- name: atlantis-enhanced-locking
  rules:
  - alert: HighErrorRate
    expr: rate(atlantis_lock_errors_total[5m]) > 0.05
    for: 2m

  - alert: QueueBacklog
    expr: atlantis_queue_size > 100
    for: 5m
```

## Error Code Reference

### Error Codes and Meanings

| Code | Description | Typical Cause | Solution |
|------|-------------|---------------|----------|
| EL001 | Backend connection failed | Redis/BoltDB unavailable | Check backend health |
| EL002 | Lock acquisition timeout | High contention | Increase timeout or optimize |
| EL003 | Invalid configuration | Configuration syntax error | Validate configuration |
| EL004 | Deadlock detected | Circular wait condition | Check deadlock resolution |
| EL005 | Queue overflow | Too many pending requests | Increase queue size |
| EL006 | Migration failure | Data inconsistency | Check migration status |
| EL007 | Permission denied | File system permissions | Fix file ownership |
| EL008 | Resource exhaustion | Memory/CPU limits | Increase resources |

### Log Pattern Analysis

**Critical Errors**
```bash
# Database corruption
grep "database corruption detected" /var/log/atlantis/atlantis.log

# Memory exhaustion
grep "out of memory" /var/log/atlantis/atlantis.log

# Deadlock cascade
grep "cascade resolution failed" /var/log/atlantis/atlantis.log
```

**Warning Patterns**
```bash
# Performance degradation
grep "slow operation detected" /var/log/atlantis/atlantis.log

# Connection issues
grep "connection retry" /var/log/atlantis/atlantis.log

# Queue warnings
grep "queue approaching limit" /var/log/atlantis/atlantis.log
```

## Emergency Procedures

### Emergency Rollback

**Immediate Rollback**
```bash
# Disable enhanced locking immediately
atlantis config set enhanced-locking.enabled=false
systemctl restart atlantis

# Verify rollback
curl -s http://atlantis:4141/api/locks/health
```

**Graceful Rollback**
```bash
# Reduce traffic gradually
atlantis config set enhanced-locking.migration.traffic-percentage=0
sleep 30

# Disable system
atlantis config set enhanced-locking.enabled=false
systemctl reload atlantis
```

### Data Recovery

**Restore from Backup**
```bash
# Stop service
systemctl stop atlantis

# Restore database
cp /var/lib/atlantis/atlantis.db.backup.latest /var/lib/atlantis/atlantis.db

# Restore configuration
cp /etc/atlantis/atlantis.yaml.backup.latest /etc/atlantis/atlantis.yaml

# Start service
systemctl start atlantis

# Verify integrity
atlantis locks verify --full-check
```

### Performance Emergency

**High Load Response**
```bash
# Increase resource limits immediately
echo 'vm.max_map_count=262144' >> /etc/sysctl.conf
sysctl -p

# Scale horizontally (if supported)
kubectl scale deployment atlantis --replicas=3

# Reduce non-essential features
atlantis config set enhanced-locking.deadlock-detection.enabled=false
atlantis config set enhanced-locking.enhanced-metrics=false
```

## Preventive Measures

### Regular Health Checks

**Daily Checks**
```bash
#!/bin/bash
# daily-health-check.sh

# Basic health
curl -f http://atlantis:4141/api/enhanced-locks/health

# Performance check
LATENCY=$(curl -s http://atlantis:4141/api/enhanced-locks/metrics | jq .avg_latency_ms)
if [ "$LATENCY" -gt 1000 ]; then
    echo "WARNING: High latency detected: ${LATENCY}ms"
fi

# Queue depth check
QUEUE_SIZE=$(curl -s http://atlantis:4141/api/enhanced-locks/queue/status | jq .depth)
if [ "$QUEUE_SIZE" -gt 50 ]; then
    echo "WARNING: Queue backlog: $QUEUE_SIZE items"
fi
```

**Weekly Maintenance**
```bash
#!/bin/bash
# weekly-maintenance.sh

# Backup database
cp /var/lib/atlantis/atlantis.db "/backup/atlantis.db.$(date +%Y%m%d)"

# Rotate logs
logrotate /etc/logrotate.d/atlantis

# Performance report
atlantis locks benchmark --duration=60s > "/reports/performance-$(date +%Y%m%d).txt"

# Configuration validation
atlantis config validate > "/reports/config-validation-$(date +%Y%m%d).txt"
```

### Monitoring Setup

**Key Metrics to Monitor**
```yaml
# Essential metrics
- atlantis_lock_duration_seconds
- atlantis_queue_size
- atlantis_deadlock_detected_total
- atlantis_lock_errors_total
- atlantis_backend_health_score

# Resource metrics
- process_memory_usage
- process_cpu_usage
- disk_usage_bytes
- network_connections_active
```

**Alert Thresholds**
```yaml
# Production thresholds
error_rate: 1%
queue_depth: 100
latency_p95: 1000ms
memory_usage: 80%
deadlock_rate: 10/hour
```

## Getting Help

### Diagnostic Information to Collect

When seeking support, gather this information:

```bash
# System information
uname -a
atlantis version
cat /etc/atlantis/atlantis.yaml

# Health status
curl -s http://atlantis:4141/api/enhanced-locks/health
curl -s http://atlantis:4141/api/enhanced-locks/metrics

# Recent logs
tail -100 /var/log/atlantis/atlantis.log

# Configuration validation
atlantis config validate

# Performance metrics
atlantis locks benchmark --duration=30s
```

### Support Channels

1. **Documentation**: [Enhanced Locking Documentation](../README.md)
2. **GitHub Issues**: File detailed bug reports
3. **Community Forums**: Ask questions and share solutions
4. **Emergency Support**: Contact maintainers for critical issues

### Creating Effective Bug Reports

Include this information:

1. **Environment Details**:
   - Atlantis version
   - Operating system
   - Backend type (Redis/BoltDB)
   - Configuration (redacted)

2. **Problem Description**:
   - Expected behavior
   - Actual behavior
   - Steps to reproduce

3. **Diagnostic Information**:
   - Error messages
   - Log snippets
   - Health check outputs
   - Performance metrics

4. **Impact Assessment**:
   - Number of users affected
   - Business impact
   - Urgency level

---

Remember: Most issues can be resolved by following the systematic approach outlined in this guide. When in doubt, start with the basic health checks and work through the solutions methodically.