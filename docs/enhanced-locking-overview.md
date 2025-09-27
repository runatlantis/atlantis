# Atlantis Enhanced Locking System - Documentation Overview

This comprehensive documentation suite covers the complete migration from Atlantis legacy locking to the enhanced system, including system architecture, migration procedures, configuration, troubleshooting, and best practices.

## 📚 Documentation Structure

### Quick Navigation

| Document | Purpose | Audience |
|----------|---------|----------|
| [🚀 Quick Start](#quick-start-5-minute-setup) | Get enhanced locking running in 5 minutes | Everyone |
| [📋 Migration Guide](./locking-migration-guide.md) | Complete step-by-step migration process | Platform Teams, DevOps |
| [🏗️ System Architecture](./locking-system-diagrams.md) | Visual system documentation and diagrams | Architects, Engineers |
| [⚙️ Enhanced System Docs](../server/core/locking/enhanced/README.md) | Technical implementation details | Developers |

### Complete Documentation Index

#### 1. Getting Started Documents
- **[Quick Start Guide](#quick-start-5-minute-setup)** (This document)
  - 5-minute setup for trying enhanced locking
  - Essential configuration examples
  - Basic testing procedures

#### 2. Migration and Deployment
- **[Migration Guide](./locking-migration-guide.md)** - *Comprehensive migration procedures*
  - Pre-migration assessment and planning
  - Phased migration approach (4-6 weeks)
  - Configuration migration examples
  - Testing and validation procedures
  - Rollback procedures and emergency responses
  - Monitoring and alerting setup

#### 3. System Architecture and Design
- **[System Diagrams](./locking-system-diagrams.md)** - *Visual system documentation*
  - Legacy vs Enhanced system architecture
  - Lock lifecycle diagrams
  - Queue management and priority systems
  - Resource hierarchy and isolation
  - Database schema layouts
  - Migration deployment flows

#### 4. Technical Implementation
- **[Enhanced System README](../server/core/locking/enhanced/README.md)** - *Developer documentation*
  - Detailed API reference
  - Component architecture
  - Configuration options
  - Performance characteristics
  - Troubleshooting guide
  - Development setup

#### 5. Legacy System Reference
- **[Legacy System Documentation](./locking-system-legacy.md)** - *Legacy system reference*
  - Current system architecture
  - Existing limitations and challenges
  - Compatibility requirements
  - Migration preparation

#### 6. Configuration and Integration
- **[Configuration Guide](./locking-configuration.md)** - *Configuration reference*
  - Server-level configuration
  - atlantis.yaml integration
  - Environment-specific settings
  - Security configurations

- **[Integration Guide](./locking-integration-guide.md)** - *CI/CD and tool integration*
  - GitHub Actions integration
  - Jenkins pipeline setup
  - Terraform Cloud compatibility
  - Custom webhook configuration

#### 7. Enhanced System Features
- **[Enhanced System Overview](./locking-system-enhanced.md)** - *Feature overview*
  - Priority-based queuing
  - Deadlock detection
  - Horizontal scaling
  - Advanced monitoring

## 🎯 Reading Recommendations

### For First-Time Users
1. **Start here**: [Quick Start Guide](#quick-start-5-minute-setup)
2. **Understand the system**: [System Diagrams](./locking-system-diagrams.md)
3. **Plan migration**: [Migration Guide](./locking-migration-guide.md)

### For Platform Engineers
1. **Migration planning**: [Migration Guide](./locking-migration-guide.md)
2. **Architecture deep-dive**: [System Diagrams](./locking-system-diagrams.md)
3. **Configuration details**: [Enhanced README](../server/core/locking/enhanced/README.md)
4. **Monitoring setup**: Migration Guide sections 10-11

### For Developers
1. **Technical implementation**: [Enhanced README](../server/core/locking/enhanced/README.md)
2. **API reference**: Enhanced README section "API Reference"
3. **Integration examples**: [Integration Guide](./locking-integration-guide.md)
4. **Troubleshooting**: Migration Guide section 9

### For Decision Makers
1. **System overview**: [Enhanced System Overview](./locking-system-enhanced.md)
2. **Migration benefits**: [Migration Guide](./locking-migration-guide.md) sections 1-2
3. **Cost analysis**: Migration Guide section "Cost and Sizing Questions"

## 🚀 Quick Start (5-Minute Setup)

Get enhanced locking running quickly in a test environment to evaluate the system.

### Prerequisites

- Atlantis version 0.19.0+ running
- Redis server available (local or remote)
- Basic Docker/Docker Compose knowledge

### Step 1: Start Redis (1 minute)

```bash
# Option A: Docker (quickest)
docker run -d --name atlantis-redis -p 6379:6379 redis:7-alpine

# Option B: Docker Compose (recommended for persistence)
cat > docker-compose.redis.yml << 'EOF'
version: '3.8'
services:
  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"
    command: redis-server --maxmemory 1gb --maxmemory-policy allkeys-lru
    volumes:
      - redis_data:/data
volumes:
  redis_data:
EOF

docker-compose -f docker-compose.redis.yml up -d
```

### Step 2: Configure Enhanced Locking (2 minutes)

Add enhanced locking configuration to your Atlantis server configuration:

```yaml
# server.yaml (or environment variables)
enhanced-locking:
  enabled: true
  backend: redis
  legacy-fallback: true  # Safe fallback during testing

# Redis configuration
redis:
  addr: "localhost:6379"
  password: ""  # Set if your Redis has authentication
  db: 0
  key-prefix: "atlantis:enhanced:"

# Conservative initial settings
default-timeout: "30m"
max-timeout: "2h"

# Start with basic features (add more later)
enable-priority-queue: false
enable-retry: false
enable-deadlock-detection: false
```

### Step 3: Restart Atlantis (30 seconds)

```bash
# Restart your Atlantis server to load the new configuration
# Method depends on your deployment:

# Docker:
docker restart atlantis

# Systemd:
sudo systemctl restart atlantis

# Kubernetes:
kubectl rollout restart deployment atlantis
```

### Step 4: Verify Enhanced System (1 minute)

```bash
# Check Atlantis logs for enhanced locking initialization
grep -i "enhanced locking" /var/log/atlantis/atlantis.log

# Expected log entries:
# INFO enhanced locking system initialized successfully
# INFO redis backend health check passed
# INFO enhanced locking enabled with legacy fallback

# Test Redis connectivity
redis-cli ping
# Expected: PONG
```

### Step 5: Test Basic Operations (30 seconds)

Create a test PR to verify locking works:

```bash
# In your test repository, create a simple change
echo "test_var = \"enhanced_locking_test\"" >> variables.tf
git add variables.tf
git commit -m "Test enhanced locking system"
git push origin test-enhanced-locking

# Create PR and run atlantis plan
# Check Atlantis logs for enhanced locking operations:
grep -i "lock acquired\|enhanced" /var/log/atlantis/atlantis.log
```

**🎉 Success Indicators:**
- Atlantis starts without errors
- Redis connection established
- Lock operations complete successfully
- Legacy fallback available if needed

## ⚙️ Essential Configuration Examples

### Development Environment

```yaml
# Minimal configuration for local development
enhanced-locking:
  enabled: true
  backend: redis
  legacy-fallback: true
  default-timeout: "15m"  # Shorter timeouts for development

redis:
  addr: "localhost:6379"
  password: ""
  db: 1  # Use different DB from production
  key-prefix: "atlantis:dev:"

# Development-friendly settings
enable-priority-queue: false
enable-retry: true
max-retry-attempts: 2
```

### Production Environment

```yaml
# Production-ready configuration with enhanced features
enhanced-locking:
  enabled: true
  backend: redis
  legacy-fallback: true  # Keep during migration
  default-timeout: "30m"
  max-timeout: "2h"

redis:
  addr: "redis-cluster.internal:6379"
  password: "${REDIS_PASSWORD}"
  db: 0
  cluster-mode: true  # Use cluster for high availability
  key-prefix: "atlantis:prod:"
  lock-ttl: "1h"

# Production features
enable-priority-queue: true
max-queue-size: 1000
queue-timeout: "10m"

enable-retry: true
max-retry-attempts: 3
retry-base-delay: "1s"

enable-deadlock-detection: true
deadlock-check-interval: "30s"

# Monitoring
enable-events: true
event-buffer-size: 5000
```

### High-Traffic Environment

```yaml
# Optimized for high concurrency (1000+ concurrent users)
enhanced-locking:
  enabled: true
  backend: redis
  default-timeout: "20m"  # Shorter for faster turnover

redis:
  cluster-mode: true
  addrs:
    - "redis-1.cluster:6379"
    - "redis-2.cluster:6379"
    - "redis-3.cluster:6379"
    - "redis-4.cluster:6379"
    - "redis-5.cluster:6379"
    - "redis-6.cluster:6379"

  # Connection pool optimization
  pool-size: 100
  min-idle-conns: 20
  max-conn-age: "30m"

# High-throughput settings
enable-priority-queue: true
max-queue-size: 5000
queue-batch-size: 50
queue-processing-interval: "500ms"

# Automatic cleanup for performance
enable-automatic-cleanup: true
cleanup-interval: "2m"
cleanup-batch-size: 100
```

## 🧪 Testing Procedures

### Basic Health Check

```bash
# 1. Check Redis connectivity
redis-cli -h redis-host -p 6379 ping

# 2. Check Atlantis logs
tail -f /var/log/atlantis/atlantis.log | grep -i enhanced

# 3. Test lock acquisition (create a test PR)
# Expected behavior:
# - Lock acquired message in logs
# - Redis keys created: KEYS atlantis:enhanced:*
# - Lock released after operation completes
```

### Load Testing (Optional)

```bash
# Simple load test with concurrent operations
for i in {1..10}; do
  (
    # Create test branch
    git checkout -b test-load-$i
    echo "test_$i = \"value\"" >> variables.tf
    git add variables.tf
    git commit -m "Load test $i"
    git push origin test-load-$i
    # Create PR via API or UI
  ) &
done
wait

# Monitor system during load:
# - Check queue depth: redis-cli ZCARD atlantis:enhanced:queue:*
# - Monitor latency: grep "lock acquired" /var/log/atlantis/atlantis.log
# - Check error rates: grep -i error /var/log/atlantis/atlantis.log
```

### Performance Verification

```bash
# Check key performance metrics
echo "=== Enhanced Locking Performance ==="

# Average lock acquisition time (should be <1s)
grep "lock acquired" /var/log/atlantis/atlantis.log | tail -100 | \
  awk '/duration:/ {sum+=$NF; count++} END {print "Average acquisition time:", sum/count "ms"}'

# Queue depth (should be <100 normally)
redis-cli ZCARD "atlantis:enhanced:queue:*" 2>/dev/null || echo "No queues active"

# Active locks (monitor for leaks)
redis-cli KEYS "atlantis:enhanced:lock:*" | wc -l

# Error rate (should be <1%)
TOTAL=$(grep -c "lock operation" /var/log/atlantis/atlantis.log)
ERRORS=$(grep -c "lock.*error" /var/log/atlantis/atlantis.log)
echo "Error rate: $(echo "scale=2; $ERRORS * 100 / $TOTAL" | bc)%"
```

## 🚨 Common Issues and Quick Fixes

### Issue 1: Redis Connection Failed

```bash
# Symptoms: "redis connection refused" in logs
# Quick fix:
docker ps | grep redis  # Check if Redis is running
redis-cli ping          # Test connectivity

# If Redis is down:
docker run -d --name atlantis-redis -p 6379:6379 redis:7-alpine
```

### Issue 2: Lock Acquisition Slow

```bash
# Symptoms: Lock operations taking >5 seconds
# Quick diagnosis:
redis-cli --latency-history -i 1  # Check Redis latency

# Quick fixes:
# 1. Reduce timeout in configuration
# 2. Clear any stuck locks
redis-cli KEYS "atlantis:enhanced:lock:*" | xargs redis-cli DEL

# 3. Restart with optimized settings
```

### Issue 3: Enhanced System Not Loading

```bash
# Check configuration syntax
atlantis server --config=/path/to/config --validate-config-only

# Check Redis permissions
redis-cli SET test_key test_value
redis-cli GET test_key
redis-cli DEL test_key
```

## 📊 Monitoring Dashboard Setup

### Quick Grafana Dashboard

```json
{
  "title": "Atlantis Enhanced Locking - Quick Start",
  "panels": [
    {
      "title": "System Health",
      "type": "stat",
      "targets": [
        {"expr": "atlantis_enhanced_locking_backend_health", "legendFormat": "Backend Health"}
      ]
    },
    {
      "title": "Active Locks",
      "type": "stat",
      "targets": [
        {"expr": "atlantis_enhanced_locking_active_locks", "legendFormat": "Active Locks"}
      ]
    },
    {
      "title": "Queue Depth",
      "type": "graph",
      "targets": [
        {"expr": "atlantis_enhanced_locking_queue_depth", "legendFormat": "Queue Depth"}
      ]
    }
  ]
}
```

### Basic Alerting

```yaml
# Alert when system is unhealthy
- alert: EnhancedLockingDown
  expr: atlantis_enhanced_locking_backend_health == 0
  for: 30s
  annotations:
    summary: "Enhanced locking system is unhealthy"

# Alert when queue gets too deep
- alert: HighQueueDepth
  expr: atlantis_enhanced_locking_queue_depth > 50
  for: 5m
  annotations:
    summary: "Lock queue depth is high"
```

## ➡️ Next Steps

After successfully running the quick start:

1. **Evaluate Performance**
   - Run for 24-48 hours with normal traffic
   - Monitor key metrics and system behavior
   - Compare with legacy system performance

2. **Plan Full Migration**
   - Review the complete [Migration Guide](./locking-migration-guide.md)
   - Set up comprehensive monitoring
   - Plan rollback procedures

3. **Enable Advanced Features**
   - Priority queues for production workloads
   - Deadlock detection for complex dependencies
   - Redis clustering for high availability

4. **Team Training**
   - Share documentation with development teams
   - Set up troubleshooting procedures
   - Establish operational runbooks

## 📞 Getting Help

### Documentation Resources
- **Migration Issues**: [Migration Guide](./locking-migration-guide.md) - Section "Troubleshooting Common Issues"
- **Configuration Help**: [Enhanced README](../server/core/locking/enhanced/README.md) - Section "Configuration"
- **Performance Issues**: [Migration Guide](./locking-migration-guide.md) - Section "Performance Optimization"

### Community Support
- **GitHub Issues**: Report bugs and request features
- **Atlantis Slack**: Community discussion and support
- **Documentation**: Comprehensive troubleshooting guides available

### Professional Support
- **Enterprise Support**: Available for production deployments
- **Migration Assistance**: Professional services for complex migrations
- **Custom Development**: Feature development and integration support

---

**Ready to enhance your Atlantis deployment!** 🚀 Start with the quick setup above, then dive into the full migration guide when you're ready to deploy to production.

## 📋 Documentation Maintenance

This documentation suite is actively maintained and updated. Last comprehensive review: 2024.

### Version History
- **v1.0** (2024): Initial enhanced locking system documentation
- **v1.1** (2024): Added comprehensive migration guide
- **v1.2** (2024): Added visual system diagrams
- **v1.3** (2024): Added quick start guide and consolidated overview

### Contributing to Documentation
- Documentation follows the same contribution process as code
- All examples should be tested and verified
- Screenshots and diagrams should be kept up-to-date
- Performance numbers should be updated with each major release

---

*This documentation overview serves as the central hub for all enhanced locking system documentation. All documents are designed to work together as a comprehensive learning and reference system.*