# Enhanced Locking System - Deployment Runbook

## Overview

This runbook provides detailed procedures for deploying and operating the Enhanced Locking System in production environments. It covers deployment strategies, monitoring, maintenance, and incident response.

## Pre-Deployment Checklist

### Infrastructure Requirements

#### Minimum Requirements
- **CPU**: 2 cores
- **Memory**: 4GB RAM
- **Disk**: 20GB SSD
- **Network**: 1Gbps, low latency to Redis
- **OS**: Linux (Ubuntu 20.04+ or RHEL 8+)

#### Recommended Production Requirements
- **CPU**: 4-8 cores
- **Memory**: 8-16GB RAM
- **Disk**: 50GB SSD with IOPS > 3000
- **Network**: 10Gbps, sub-millisecond latency to Redis
- **Load Balancer**: HAProxy or similar
- **Monitoring**: Prometheus, Grafana

#### Redis Cluster Requirements
- **Nodes**: 3-7 nodes (odd number)
- **Memory**: 4GB per node minimum
- **Disk**: SSD with persistence
- **Network**: Dedicated VLAN, low latency
- **Replication**: 2+ replicas per master

### Security Requirements

```bash
# Firewall rules
ufw allow from 10.0.0.0/8 to any port 6379  # Redis
ufw allow from 10.0.0.0/8 to any port 4141  # Atlantis
ufw deny 6379  # Block external Redis access
ufw deny 4141  # Block external Atlantis access

# SSL/TLS certificates
openssl req -x509 -newkey rsa:4096 -keyout atlantis.key -out atlantis.crt -days 365

# Redis AUTH password
export REDIS_PASSWORD=$(openssl rand -base64 32)
```

### Environment Setup

```bash
# Create atlantis user
useradd -r -s /bin/false atlantis

# Create directories
mkdir -p /opt/atlantis/{bin,config,data,logs}
mkdir -p /var/lib/atlantis
mkdir -p /var/log/atlantis

# Set permissions
chown -R atlantis:atlantis /opt/atlantis
chown -R atlantis:atlantis /var/lib/atlantis
chown -R atlantis:atlantis /var/log/atlantis
```

## Deployment Strategies

### Strategy 1: Blue-Green Deployment

#### Phase 1: Prepare Green Environment

```bash
# Set up green environment
ansible-playbook -i inventory/production deploy-green.yml

# Verify green environment
curl -f http://atlantis-green:4141/healthz
```

#### Phase 2: Deploy Enhanced Locking

```yaml
# green-config.yaml
enhanced-locking:
  enabled: true
  backend: redis
  redis:
    addresses:
      - "redis-green-1:6379"
      - "redis-green-2:6379"
      - "redis-green-3:6379"
    password: "${REDIS_PASSWORD}"
    cluster-mode: true
    pool-size: 20

  # Conservative initial settings
  priority-queue: false
  deadlock-detection: false
  metrics: true
  health-checks: true
```

#### Phase 3: Switch Traffic

```bash
# Update load balancer
curl -X POST http://lb-admin:8080/api/upstream/atlantis \
  -d '{"servers": ["atlantis-green-1:4141", "atlantis-green-2:4141"]}'

# Monitor for 5 minutes
for i in {1..30}; do
  curl -s http://atlantis:4141/api/enhanced-locks/health | jq .healthy
  sleep 10
done
```

#### Phase 4: Rollback if Needed

```bash
# Emergency rollback
curl -X POST http://lb-admin:8080/api/upstream/atlantis \
  -d '{"servers": ["atlantis-blue-1:4141", "atlantis-blue-2:4141"]}'
```

### Strategy 2: Rolling Deployment

#### Phase 1: Deploy Node by Node

```bash
# Deploy to first node
ansible-playbook -i inventory/production deploy.yml --limit atlantis-1

# Verify health
curl -f http://atlantis-1:4141/healthz

# Remove from load balancer, deploy to second node
curl -X DELETE http://lb-admin:8080/api/upstream/atlantis/atlantis-2:4141
ansible-playbook -i inventory/production deploy.yml --limit atlantis-2

# Add back to load balancer
curl -X POST http://lb-admin:8080/api/upstream/atlantis \
  -d '{"server": "atlantis-2:4141"}'
```

#### Phase 2: Gradual Feature Enablement

```bash
# Enable enhanced locking on 25% of nodes
ansible-playbook -i inventory/production enable-enhanced.yml --limit atlantis-1

# Monitor for 1 hour
sleep 3600

# Enable on 50% of nodes
ansible-playbook -i inventory/production enable-enhanced.yml --limit atlantis-1,atlantis-2

# Monitor and continue
```

### Strategy 3: Canary Deployment

#### Phase 1: Deploy Canary

```yaml
# canary-config.yaml
enhanced-locking:
  enabled: true
  migration:
    mode: "canary"
    traffic-percentage: 5  # 5% of traffic
    canary-criteria:
      - workspace: "staging"
      - project: "test-*"

  backend: redis
  fallback-enabled: true
  validation-enabled: true
```

#### Phase 2: Monitor Canary

```bash
# Monitor canary metrics
while true; do
  curl -s http://atlantis:4141/api/enhanced-locks/canary/metrics | \
    jq '{success_rate, error_rate, latency_p95}'
  sleep 30
done
```

#### Phase 3: Promote or Rollback

```bash
# If successful, promote to full deployment
ansible-playbook -i inventory/production promote-canary.yml

# If issues, rollback canary
ansible-playbook -i inventory/production rollback-canary.yml
```

## Configuration Management

### Production Configuration Template

```yaml
# /opt/atlantis/config/atlantis.yaml
# Enhanced Locking Production Configuration

# Basic Atlantis settings
atlantis-url: "https://atlantis.company.com"
gh-user: "atlantis-bot"
gh-token-file: "/opt/atlantis/config/github-token"
repo-allowlist: "github.com/company/*"

# Enhanced Locking Configuration
enhanced-locking:
  enabled: true
  backend: "redis"

  # Redis cluster configuration
  redis:
    addresses:
      - "redis-cluster-1.internal:6379"
      - "redis-cluster-2.internal:6379"
      - "redis-cluster-3.internal:6379"
    password-file: "/opt/atlantis/config/redis-password"
    db: 0
    cluster-mode: true
    pool-size: 50
    key-prefix: "atlantis:prod:lock:"
    default-ttl: "1h"

    # Connection settings
    connection-timeout: "5s"
    read-timeout: "3s"
    write-timeout: "3s"
    max-retries: 3

    # Health monitoring
    health-monitoring:
      enabled: true
      check-interval: "30s"
      circuit-breaker-enabled: true
      performance-monitoring: true
      alerting-enabled: true

  # Feature configuration
  priority-queue: true
  deadlock-detection: true
  retries: true
  metrics: true
  event-streaming: true

  # Performance settings
  queue-config:
    max-queue-size: 1000
    batch-processing-size: 20
    processing-interval: "100ms"
    max-wait-time: "10m"
    enable-batching: true

  timeout-config:
    enable-adaptive-timeout: true
    base-lock-timeout: "30s"
    base-queue-timeout: "5m"
    max-timeout-multiplier: 3.0
    adaptation-factor: 0.1

  # Deadlock detection
  deadlock-config:
    check-interval: "30s"
    resolution-policy: "lowest_priority"
    enable-prevention: true
    max-resolution-attempts: 3

  # Monitoring and observability
  monitoring:
    metrics-enabled: true
    detailed-metrics: true
    export-prometheus: true
    log-level: "info"

  # Security settings
  security:
    enable-auth: true
    tls-enabled: true
    cert-file: "/opt/atlantis/config/tls.crt"
    key-file: "/opt/atlantis/config/tls.key"
```

### Environment-Specific Configurations

#### Staging Environment
```yaml
enhanced-locking:
  enabled: true
  backend: "boltdb"  # Simpler for staging
  priority-queue: true
  deadlock-detection: true

  # More aggressive timeouts for faster testing
  timeout-config:
    base-lock-timeout: "10s"
    base-queue-timeout: "2m"

  # Debug logging
  monitoring:
    log-level: "debug"
    detailed-metrics: true
```

#### Development Environment
```yaml
enhanced-locking:
  enabled: true
  backend: "memory"  # In-memory for development
  priority-queue: false
  deadlock-detection: false

  # Fast timeouts for development
  timeout-config:
    base-lock-timeout: "5s"
    base-queue-timeout: "30s"

  # Verbose logging
  monitoring:
    log-level: "trace"
```

## Redis Cluster Deployment

### Redis Cluster Setup

```bash
# Deploy Redis nodes
for i in {1..6}; do
  docker run -d \
    --name redis-node-$i \
    --net redis-cluster \
    -p $((7000 + i)):6379 \
    redis:7-alpine \
    redis-server --cluster-enabled yes \
                 --cluster-config-file nodes.conf \
                 --cluster-node-timeout 5000 \
                 --appendonly yes \
                 --appendfsync everysec \
                 --maxmemory 4gb \
                 --maxmemory-policy allkeys-lru
done

# Create cluster
redis-cli --cluster create \
  redis-node-1:6379 redis-node-2:6379 redis-node-3:6379 \
  redis-node-4:6379 redis-node-5:6379 redis-node-6:6379 \
  --cluster-replicas 1 --cluster-yes

# Verify cluster
redis-cli --cluster check redis-node-1:6379
```

### Redis Configuration

```conf
# /etc/redis/redis.conf
# Production Redis configuration

# Network
bind 0.0.0.0
port 6379
tcp-backlog 511
tcp-keepalive 300

# Memory
maxmemory 4gb
maxmemory-policy allkeys-lru

# Persistence
save 900 1
save 300 10
save 60 10000
appendonly yes
appendfsync everysec

# Cluster
cluster-enabled yes
cluster-config-file nodes.conf
cluster-node-timeout 5000
cluster-announce-ip 10.0.0.x
cluster-announce-port 6379

# Security
requirepass ${REDIS_PASSWORD}
```

### Redis Monitoring

```bash
# Monitor Redis performance
redis-cli --latency -h redis-cluster-1
redis-cli --stat -h redis-cluster-1

# Check cluster health
redis-cli --cluster info redis-cluster-1:6379
redis-cli --cluster check redis-cluster-1:6379
```

## Monitoring and Alerting

### Prometheus Configuration

```yaml
# prometheus.yml
global:
  scrape_interval: 15s

scrape_configs:
  - job_name: 'atlantis-enhanced-locking'
    static_configs:
      - targets: ['atlantis-1:4141', 'atlantis-2:4141', 'atlantis-3:4141']
    metrics_path: '/api/enhanced-locks/metrics'
    scrape_interval: 10s

  - job_name: 'redis-cluster'
    static_configs:
      - targets: ['redis-1:6379', 'redis-2:6379', 'redis-3:6379']
    metrics_path: '/metrics'
```

### Grafana Dashboard

Import the enhanced locking dashboard:

```json
{
  "dashboard": {
    "title": "Atlantis Enhanced Locking",
    "panels": [
      {
        "title": "Lock Operations/sec",
        "type": "graph",
        "targets": [
          {
            "expr": "rate(atlantis_lock_operations_total[5m])",
            "legendFormat": "{{instance}}"
          }
        ]
      },
      {
        "title": "Queue Depth",
        "type": "graph",
        "targets": [
          {
            "expr": "atlantis_queue_size",
            "legendFormat": "{{instance}}"
          }
        ]
      }
    ]
  }
}
```

### Alert Rules

```yaml
# alerts.yml
groups:
- name: atlantis-enhanced-locking
  rules:
  - alert: EnhancedLockingHighErrorRate
    expr: rate(atlantis_lock_errors_total[5m]) > 0.05
    for: 2m
    labels:
      severity: critical
    annotations:
      summary: "High error rate in enhanced locking system"

  - alert: EnhancedLockingHighLatency
    expr: histogram_quantile(0.95, rate(atlantis_lock_duration_seconds_bucket[5m])) > 1.0
    for: 5m
    labels:
      severity: warning
    annotations:
      summary: "High latency in lock operations"

  - alert: EnhancedLockingQueueBacklog
    expr: atlantis_queue_size > 100
    for: 3m
    labels:
      severity: warning
    annotations:
      summary: "Lock queue backlog detected"

  - alert: RedisClusterDown
    expr: up{job="redis-cluster"} < 0.5
    for: 1m
    labels:
      severity: critical
    annotations:
      summary: "Redis cluster majority down"
```

### Log Management

```bash
# Configure log rotation
cat > /etc/logrotate.d/atlantis << EOF
/var/log/atlantis/*.log {
    daily
    rotate 30
    compress
    delaycompress
    missingok
    notifempty
    postrotate
        systemctl reload atlantis
    endscript
}
EOF

# Configure rsyslog for centralized logging
echo "*.* @@logserver.company.com:514" >> /etc/rsyslog.conf
systemctl restart rsyslog
```

## Operations Procedures

### Daily Operations

#### Health Checks

```bash
#!/bin/bash
# daily-health-check.sh

echo "=== Daily Enhanced Locking Health Check ==="
echo "Date: $(date)"

# Basic health
echo "1. Basic Health Check:"
curl -s http://atlantis:4141/api/enhanced-locks/health | jq .

# Performance metrics
echo "2. Performance Metrics:"
curl -s http://atlantis:4141/api/enhanced-locks/metrics | \
  jq '{
    active_locks: .active_locks,
    queue_depth: .queue_depth,
    avg_latency: .avg_latency_ms,
    error_rate: .error_rate
  }'

# Redis health
echo "3. Redis Cluster Health:"
redis-cli --cluster check redis-1:6379

# Queue status
echo "4. Queue Status:"
curl -s http://atlantis:4141/api/enhanced-locks/queue/status | jq .

# Recent errors
echo "5. Recent Errors (last hour):"
journalctl -u atlantis --since "1 hour ago" | grep ERROR | tail -10

echo "=== Health Check Complete ==="
```

#### Performance Monitoring

```bash
#!/bin/bash
# performance-monitor.sh

# Collect performance metrics
curl -s http://atlantis:4141/api/enhanced-locks/metrics > /tmp/metrics.json

# Check for performance anomalies
LATENCY_P95=$(jq -r '.latency_p95_ms' /tmp/metrics.json)
ERROR_RATE=$(jq -r '.error_rate' /tmp/metrics.json)
QUEUE_DEPTH=$(jq -r '.queue_depth' /tmp/metrics.json)

# Alert if thresholds exceeded
if (( $(echo "$LATENCY_P95 > 1000" | bc -l) )); then
    echo "ALERT: High latency detected: ${LATENCY_P95}ms"
    # Send alert
fi

if (( $(echo "$ERROR_RATE > 0.01" | bc -l) )); then
    echo "ALERT: High error rate detected: ${ERROR_RATE}"
    # Send alert
fi

if [ "$QUEUE_DEPTH" -gt 50 ]; then
    echo "ALERT: High queue depth detected: ${QUEUE_DEPTH}"
    # Send alert
fi
```

### Weekly Operations

#### Performance Review

```bash
#!/bin/bash
# weekly-performance-review.sh

echo "=== Weekly Performance Review ==="

# Generate performance report
curl -s "http://prometheus:9090/api/v1/query_range?query=rate(atlantis_lock_duration_seconds_sum[5m])&start=$(date -d '7 days ago' +%s)&end=$(date +%s)&step=3600" | \
  jq '.data.result' > weekly-performance.json

# Check for trends
python3 analyze-trends.py weekly-performance.json

# Capacity planning
echo "Current utilization:"
curl -s http://atlantis:4141/api/enhanced-locks/metrics | \
  jq '{
    cpu_usage: .cpu_usage_percent,
    memory_usage: .memory_usage_percent,
    queue_utilization: (.queue_depth / .max_queue_size * 100)
  }'
```

#### Backup Verification

```bash
#!/bin/bash
# verify-backups.sh

echo "=== Backup Verification ==="

# Verify Redis backups
for node in redis-1 redis-2 redis-3; do
    echo "Checking backup for $node..."
    if [ -f "/backup/redis/$node/dump.rdb" ]; then
        redis-check-rdb "/backup/redis/$node/dump.rdb"
        echo "$node backup: OK"
    else
        echo "$node backup: MISSING"
    fi
done

# Verify configuration backups
if [ -f "/backup/atlantis/atlantis.yaml" ]; then
    echo "Configuration backup: OK"
else
    echo "Configuration backup: MISSING"
fi
```

### Monthly Operations

#### Capacity Planning

```bash
#!/bin/bash
# monthly-capacity-review.sh

echo "=== Monthly Capacity Review ==="

# Analyze growth trends
curl -s "http://prometheus:9090/api/v1/query_range?query=atlantis_lock_operations_total&start=$(date -d '30 days ago' +%s)&end=$(date +%s)&step=86400" | \
  jq '.data.result' > monthly-usage.json

# Generate capacity recommendations
python3 capacity-planner.py monthly-usage.json

# Review resource utilization
echo "Resource utilization trends:"
curl -s "http://prometheus:9090/api/v1/query?query=avg_over_time(atlantis_memory_usage_percent[30d])" | \
  jq '.data.result[0].value[1]'
```

#### Security Review

```bash
#!/bin/bash
# monthly-security-review.sh

echo "=== Monthly Security Review ==="

# Check certificate expiration
openssl x509 -in /opt/atlantis/config/tls.crt -noout -dates

# Audit access logs
grep "enhanced-locks" /var/log/atlantis/access.log | \
  awk '{print $1}' | sort | uniq -c | sort -nr

# Check for security updates
apt list --upgradable | grep -E "(redis|atlantis)"

# Rotate Redis password if needed
if [ "$ROTATE_REDIS_PASSWORD" = "true" ]; then
    ./rotate-redis-password.sh
fi
```

## Incident Response

### Incident Classification

#### Severity 1 (Critical)
- Complete system outage
- Data corruption
- Security breach
- >50% error rate

#### Severity 2 (High)
- Partial outage
- Performance degradation >100%
- Queue backlog >500 items
- Redis cluster issues

#### Severity 3 (Medium)
- Minor performance issues
- Non-critical errors
- Configuration problems

### Response Procedures

#### Severity 1 Response

```bash
#!/bin/bash
# sev1-response.sh

echo "SEVERITY 1 INCIDENT RESPONSE"
echo "Incident start time: $(date)"

# Step 1: Immediate assessment
echo "1. System Assessment:"
curl -s http://atlantis:4141/healthz
curl -s http://atlantis:4141/api/enhanced-locks/health

# Step 2: Emergency rollback if needed
if [ "$EMERGENCY_ROLLBACK" = "true" ]; then
    echo "2. Emergency Rollback:"
    curl -X POST http://atlantis:4141/api/enhanced-locks/emergency/disable
    systemctl restart atlantis
fi

# Step 3: Collect diagnostics
echo "3. Collecting Diagnostics:"
journalctl -u atlantis --since "1 hour ago" > incident-logs.txt
curl -s http://atlantis:4141/api/enhanced-locks/debug/dump > debug-dump.json

# Step 4: Notify stakeholders
echo "4. Notification:"
# Send incident notification
curl -X POST http://pagerduty-api/incidents \
  -d '{"incident": {"type": "incident", "title": "Atlantis Enhanced Locking Sev1"}}'
```

#### Common Incident Scenarios

##### Redis Cluster Failure

```bash
# Diagnose Redis issues
redis-cli --cluster check redis-1:6379
redis-cli --cluster info redis-1:6379

# Check Redis logs
journalctl -u redis --since "1 hour ago"

# Failover to backup or rollback
if [ "$REDIS_CLUSTER_FAILED" = "true" ]; then
    # Switch to BoltDB backend
    curl -X POST http://atlantis:4141/api/enhanced-locks/config/backend \
      -d '{"backend": "boltdb"}'
fi
```

##### High Memory Usage

```bash
# Check memory usage
free -h
ps aux --sort=-%mem | head -10

# Check for memory leaks
curl -s http://atlantis:4141/api/enhanced-locks/debug/memory

# Restart if necessary
if [ "$MEMORY_USAGE" -gt 90 ]; then
    systemctl restart atlantis
fi
```

##### Queue Backlog

```bash
# Analyze queue backlog
curl -s http://atlantis:4141/api/enhanced-locks/queue/status

# Increase processing capacity
curl -X POST http://atlantis:4141/api/enhanced-locks/config/queue \
  -d '{"batch_size": 50, "processing_interval": "50ms"}'

# Clear stale queue items if needed
curl -X POST http://atlantis:4141/api/enhanced-locks/queue/cleanup
```

## Maintenance Procedures

### Scheduled Maintenance

#### Redis Maintenance

```bash
#!/bin/bash
# redis-maintenance.sh

echo "=== Redis Maintenance ==="

# 1. Create backup before maintenance
for node in redis-1 redis-2 redis-3; do
    redis-cli -h $node BGSAVE
    while [ "$(redis-cli -h $node LASTSAVE)" = "$(redis-cli -h $node LASTSAVE)" ]; do
        sleep 1
    done
done

# 2. Update Redis nodes one by one
for node in redis-1 redis-2 redis-3; do
    echo "Maintaining $node..."

    # Failover if master
    if redis-cli -h $node INFO replication | grep "role:master"; then
        redis-cli --cluster failover $node:6379
        sleep 30
    fi

    # Update node
    systemctl stop redis@$node
    apt update && apt upgrade redis-server
    systemctl start redis@$node

    # Wait for node to rejoin cluster
    sleep 60
    redis-cli --cluster check redis-1:6379
done

echo "Redis maintenance complete"
```

#### Atlantis Maintenance

```bash
#!/bin/bash
# atlantis-maintenance.sh

echo "=== Atlantis Maintenance ==="

# 1. Backup configuration and data
cp /opt/atlantis/config/atlantis.yaml /backup/atlantis/atlantis.yaml.$(date +%Y%m%d)
cp /var/lib/atlantis/atlantis.db /backup/atlantis/atlantis.db.$(date +%Y%m%d)

# 2. Update Atlantis binary
systemctl stop atlantis
cp atlantis-new /opt/atlantis/bin/atlantis
chmod +x /opt/atlantis/bin/atlantis

# 3. Validate configuration
/opt/atlantis/bin/atlantis config validate

# 4. Start with health check
systemctl start atlantis
sleep 30

# 5. Verify health
curl -f http://localhost:4141/healthz
curl -f http://localhost:4141/api/enhanced-locks/health

echo "Atlantis maintenance complete"
```

### Configuration Updates

```bash
#!/bin/bash
# update-config.sh

# 1. Backup current configuration
cp /opt/atlantis/config/atlantis.yaml /opt/atlantis/config/atlantis.yaml.backup

# 2. Validate new configuration
atlantis config validate --config new-atlantis.yaml

# 3. Apply configuration
cp new-atlantis.yaml /opt/atlantis/config/atlantis.yaml

# 4. Reload configuration
systemctl reload atlantis

# 5. Verify changes
curl -s http://atlantis:4141/api/enhanced-locks/config | jq .
```

## Performance Tuning

### System Level Tuning

```bash
# Kernel parameters for high performance
echo 'net.core.somaxconn = 65535' >> /etc/sysctl.conf
echo 'net.ipv4.tcp_max_syn_backlog = 65535' >> /etc/sysctl.conf
echo 'vm.overcommit_memory = 1' >> /etc/sysctl.conf
sysctl -p

# File descriptor limits
echo 'atlantis soft nofile 65535' >> /etc/security/limits.conf
echo 'atlantis hard nofile 65535' >> /etc/security/limits.conf

# CPU governor
echo performance > /sys/devices/system/cpu/cpu*/cpufreq/scaling_governor
```

### Application Level Tuning

```yaml
# Performance optimized configuration
enhanced-locking:
  # Increase connection pools
  redis:
    pool-size: 100
    max-idle-connections: 50

  # Optimize queue processing
  queue-config:
    batch-processing-size: 50
    processing-interval: "50ms"
    enable-parallel-processing: true

  # Tune timeouts for performance
  timeout-config:
    base-lock-timeout: "15s"  # Reduced from 30s
    enable-adaptive-timeout: true
    adaptation-factor: 0.2  # More aggressive adaptation
```

### Redis Tuning

```conf
# High performance Redis configuration
tcp-backlog 511
tcp-keepalive 60
timeout 0

# Memory optimization
maxmemory-policy allkeys-lru
hash-max-ziplist-entries 512
hash-max-ziplist-value 64

# Network optimization
tcp-nodelay yes
```

## Backup and Recovery

### Backup Procedures

```bash
#!/bin/bash
# backup.sh

BACKUP_DIR="/backup/atlantis/$(date +%Y%m%d_%H%M%S)"
mkdir -p "$BACKUP_DIR"

# Backup Redis data
for node in redis-1 redis-2 redis-3; do
    redis-cli -h $node BGSAVE
    scp $node:/var/lib/redis/dump.rdb "$BACKUP_DIR/redis-$node-dump.rdb"
done

# Backup Atlantis configuration
cp /opt/atlantis/config/atlantis.yaml "$BACKUP_DIR/"

# Backup SSL certificates
cp /opt/atlantis/config/tls.* "$BACKUP_DIR/"

# Create backup manifest
cat > "$BACKUP_DIR/manifest.json" << EOF
{
  "backup_time": "$(date -Iseconds)",
  "atlantis_version": "$(atlantis version)",
  "redis_version": "$(redis-cli --version)",
  "files": [
    "atlantis.yaml",
    "redis-redis-1-dump.rdb",
    "redis-redis-2-dump.rdb",
    "redis-redis-3-dump.rdb",
    "tls.crt",
    "tls.key"
  ]
}
EOF

# Compress backup
tar -czf "$BACKUP_DIR.tar.gz" -C "$(dirname $BACKUP_DIR)" "$(basename $BACKUP_DIR)"
rm -rf "$BACKUP_DIR"

echo "Backup completed: $BACKUP_DIR.tar.gz"
```

### Recovery Procedures

```bash
#!/bin/bash
# recovery.sh

BACKUP_FILE="$1"
if [ -z "$BACKUP_FILE" ]; then
    echo "Usage: $0 <backup-file.tar.gz>"
    exit 1
fi

echo "=== RECOVERY PROCEDURE ==="
echo "Backup file: $BACKUP_FILE"

# Stop services
systemctl stop atlantis
systemctl stop redis

# Extract backup
TEMP_DIR="/tmp/recovery-$$"
mkdir -p "$TEMP_DIR"
tar -xzf "$BACKUP_FILE" -C "$TEMP_DIR"

# Restore configuration
cp "$TEMP_DIR"/atlantis.yaml /opt/atlantis/config/
cp "$TEMP_DIR"/tls.* /opt/atlantis/config/

# Restore Redis data
for node in redis-1 redis-2 redis-3; do
    if [ -f "$TEMP_DIR/redis-$node-dump.rdb" ]; then
        scp "$TEMP_DIR/redis-$node-dump.rdb" $node:/var/lib/redis/dump.rdb
    fi
done

# Start services
systemctl start redis
sleep 30
systemctl start atlantis

# Verify recovery
sleep 30
curl -f http://localhost:4141/healthz
curl -f http://localhost:4141/api/enhanced-locks/health

# Cleanup
rm -rf "$TEMP_DIR"

echo "Recovery completed successfully"
```

## Contact Information

### Escalation Matrix

| Severity | Contact | Response Time |
|----------|---------|---------------|
| Sev 1 | On-call engineer | 15 minutes |
| Sev 2 | Platform team | 2 hours |
| Sev 3 | Platform team | Next business day |

### Emergency Contacts

- **On-call Engineer**: +1-555-0123
- **Platform Team Lead**: +1-555-0124
- **Redis DBA**: +1-555-0125
- **Security Team**: +1-555-0126

### Communication Channels

- **Incident Channel**: #incident-response
- **Platform Team**: #platform-engineering
- **Atlantis Users**: #atlantis-support

---

This runbook should be reviewed and updated quarterly to ensure all procedures remain current and accurate.