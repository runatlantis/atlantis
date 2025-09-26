# Redis Backend for Enhanced Locking System

## Overview

The Redis backend provides a distributed, high-performance foundation for Atlantis's enhanced locking system. It offers advanced features including clustering support, atomic operations, health monitoring, and priority queuing.

## Architecture

### Core Components

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│  Redis Backend  │    │ Script Manager  │    │ Health Monitor  │
│                 │────│                 │────│                 │
│ - Lock Ops      │    │ - Lua Scripts   │    │ - Circuit Breaker│
│ - Queue Mgmt    │    │ - Atomicity     │    │ - Metrics       │
│ - TTL Mgmt      │    │ - Performance   │    │ - Auto-Recovery │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         └───────────────────────┼───────────────────────┘
                                 │
                    ┌─────────────────┐
                    │ Cluster Manager │
                    │                 │
                    │ - Leader Election│
                    │ - Node Discovery│
                    │ - Consensus     │
                    └─────────────────┘
```

### Key Features

1. **Atomic Operations**: Lua scripts ensure consistency across distributed operations
2. **Health Monitoring**: Circuit breaker pattern with automatic failure recovery
3. **Cluster Support**: Leader election and distributed consensus mechanisms
4. **Priority Queuing**: Fair scheduling with anti-starvation protection
5. **Performance Optimization**: Connection pooling and batch operations
6. **TTL Management**: Automatic cleanup of expired locks

## Configuration

### Basic Configuration

```yaml
enhanced_locking:
  enabled: true
  backend: "redis"
  redis:
    addresses: ["localhost:6379"]
    password: ""
    db: 0
    pool_size: 20
    key_prefix: "atlantis:enhanced:lock:"
    default_ttl: "1h"
    connection_timeout: "5s"
    read_timeout: "3s"
    write_timeout: "3s"
    cluster_mode: false
```

### Cluster Configuration

```yaml
enhanced_locking:
  backend: "redis"
  redis:
    addresses:
      - "redis-cluster-0:6379"
      - "redis-cluster-1:6379"
      - "redis-cluster-2:6379"
    cluster_mode: true
    pool_size: 50
    key_prefix: "atlantis:enhanced:lock:"
    default_ttl: "2h"
```

### Advanced Configuration

```yaml
enhanced_locking:
  backend: "redis"
  enable_priority_queue: true
  enable_retries: true
  enable_metrics: true
  default_timeout: "30s"
  max_timeout: "5m"
  max_queue_size: 1000
  starvation_threshold: "2m"
  max_priority_boost: 3

  redis:
    addresses: ["redis:6379"]
    pool_size: 30
    key_prefix: "atlantis:enhanced:lock:"
    default_ttl: "1h"
    cluster_mode: false

    # Health monitoring
    health_check_interval: "10s"
    circuit_breaker:
      failure_threshold: 5
      reset_timeout: "30s"
      half_open_max_requests: 3
```

## Deployment

### Docker Compose

```yaml
version: '3.8'
services:
  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"
    command: redis-server --appendonly yes
    volumes:
      - redis_data:/data
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 10s
      timeout: 5s
      retries: 3

  atlantis:
    image: runatlantis/atlantis:latest
    environment:
      ATLANTIS_ENHANCED_LOCKING_ENABLED: "true"
      ATLANTIS_ENHANCED_LOCKING_BACKEND: "redis"
      ATLANTIS_ENHANCED_LOCKING_REDIS_ADDRESSES: "redis:6379"
      ATLANTIS_ENHANCED_LOCKING_ENABLE_PRIORITY_QUEUE: "true"
    depends_on:
      redis:
        condition: service_healthy
    volumes:
      - atlantis_data:/atlantis-data

volumes:
  redis_data:
  atlantis_data:
```

### Redis Cluster with Docker

```yaml
version: '3.8'
services:
  redis-cluster:
    image: redis:7-alpine
    command: redis-cli --cluster create
      redis-node-1:6379 redis-node-2:6379 redis-node-3:6379
      --cluster-replicas 0 --cluster-yes
    depends_on:
      - redis-node-1
      - redis-node-2
      - redis-node-3

  redis-node-1:
    image: redis:7-alpine
    ports:
      - "7001:6379"
    command: redis-server --cluster-enabled yes --cluster-config-file nodes.conf
    volumes:
      - redis_node_1:/data

  redis-node-2:
    image: redis:7-alpine
    ports:
      - "7002:6379"
    command: redis-server --cluster-enabled yes --cluster-config-file nodes.conf
    volumes:
      - redis_node_2:/data

  redis-node-3:
    image: redis:7-alpine
    ports:
      - "7003:6379"
    command: redis-server --cluster-enabled yes --cluster-config-file nodes.conf
    volumes:
      - redis_node_3:/data

volumes:
  redis_node_1:
  redis_node_2:
  redis_node_3:
```

### Kubernetes Deployment

```yaml
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: redis-cluster
spec:
  serviceName: redis-cluster
  replicas: 3
  selector:
    matchLabels:
      app: redis-cluster
  template:
    metadata:
      labels:
        app: redis-cluster
    spec:
      containers:
      - name: redis
        image: redis:7-alpine
        ports:
        - containerPort: 6379
        command:
        - redis-server
        - --cluster-enabled
        - "yes"
        - --cluster-config-file
        - nodes.conf
        - --appendonly
        - "yes"
        volumeMounts:
        - name: redis-data
          mountPath: /data
        readinessProbe:
          exec:
            command: ["redis-cli", "ping"]
          initialDelaySeconds: 5
          timeoutSeconds: 5
        livenessProbe:
          exec:
            command: ["redis-cli", "ping"]
          initialDelaySeconds: 30
          timeoutSeconds: 5
  volumeClaimTemplates:
  - metadata:
      name: redis-data
    spec:
      accessModes: ["ReadWriteOnce"]
      resources:
        requests:
          storage: 1Gi
---
apiVersion: v1
kind: Service
metadata:
  name: redis-cluster
spec:
  clusterIP: None
  selector:
    app: redis-cluster
  ports:
  - port: 6379
    targetPort: 6379
```

## Monitoring and Observability

### Health Metrics

The Redis backend exposes comprehensive health metrics:

```go
type HealthMetrics struct {
    ConnectionsActive   int64
    ConnectionsTotal    int64
    ResponseTime        time.Duration
    MemoryUsage         int64
    SlowQueries         int64
    LastHealthCheck     time.Time
    CircuitBreakerState string
}
```

### Performance Metrics

```go
type PerformanceMetrics struct {
    OperationsPerSecond float64
    AverageLatency      time.Duration
    P95Latency          time.Duration
    P99Latency          time.Duration
    ErrorRate           float64
    QueueDepth          int
    ActiveLocks         int64
}
```

### Circuit Breaker States

- **Closed**: Normal operation, all requests allowed
- **Open**: Failure threshold exceeded, requests rejected
- **Half-Open**: Testing recovery, limited requests allowed

## Operations

### Lock Operations

```bash
# Check active locks
atlantis locks list

# Force unlock (admin only)
atlantis locks unlock --project=myorg/myrepo --workspace=default --force

# Queue status
atlantis locks queue --project=myorg/myrepo --workspace=default
```

### Health Checks

```bash
# Backend health
curl http://atlantis:4141/api/health/locking

# Redis connectivity
redis-cli -h redis ping

# Cluster status
redis-cli -h redis cluster info
```

### Performance Monitoring

```bash
# Redis performance stats
redis-cli -h redis info stats

# Connection pool stats
redis-cli -h redis info clients

# Memory usage
redis-cli -h redis info memory
```

## Troubleshooting

### Common Issues

#### Connection Failures

**Symptoms**:
- "connection refused" errors
- Circuit breaker in open state

**Solutions**:
1. Check Redis service status
2. Verify network connectivity
3. Check Redis authentication
4. Review connection pool settings

```bash
# Test connectivity
redis-cli -h <redis-host> -p <redis-port> ping

# Check logs
docker logs <redis-container>
```

#### Performance Issues

**Symptoms**:
- High lock acquisition latency
- Queue timeouts
- Memory warnings

**Solutions**:
1. Increase connection pool size
2. Enable Redis pipelining
3. Optimize Lua script performance
4. Monitor Redis memory usage

```bash
# Monitor performance
redis-cli -h redis --latency-history

# Check slow queries
redis-cli -h redis slowlog get 10
```

#### Cluster Split-Brain

**Symptoms**:
- Inconsistent lock states
- Leader election failures
- Node communication errors

**Solutions**:
1. Check network partitions
2. Verify cluster configuration
3. Restart cluster nodes
4. Review leader election logs

```bash
# Check cluster status
redis-cli -h redis cluster nodes

# Force cluster reconfiguration
redis-cli -h redis cluster reset
```

### Debugging

#### Enable Debug Logging

```yaml
# atlantis.yaml
log-level: debug

# Environment variable
ATLANTIS_LOG_LEVEL=debug
```

#### Redis Debugging

```bash
# Monitor Redis commands
redis-cli -h redis monitor

# Check Redis logs
redis-cli -h redis config get "*log*"

# Analyze memory usage
redis-cli -h redis --bigkeys
```

#### Lua Script Debugging

```bash
# Enable Lua debugging
redis-cli -h redis config set lua-time-limit 5000

# Check script cache
redis-cli -h redis script exists <sha1>

# Flush script cache
redis-cli -h redis script flush
```

## Security Considerations

### Authentication

```yaml
redis:
  password: "${REDIS_PASSWORD}"  # Use environment variables
  auth_method: "password"        # or "acl"
```

### Network Security

```yaml
redis:
  tls_enabled: true
  tls_cert_file: "/path/to/cert.pem"
  tls_key_file: "/path/to/key.pem"
  tls_ca_file: "/path/to/ca.pem"
```

### Access Control

```bash
# Redis ACL configuration
redis-cli ACL SETUSER atlantis +@all ~atlantis:enhanced:lock:* &*
```

## Performance Tuning

### Redis Configuration

```conf
# redis.conf
maxmemory 2gb
maxmemory-policy allkeys-lru
tcp-keepalive 300
timeout 0
tcp-backlog 511
databases 1
save ""  # Disable RDB for better performance
appendonly yes
appendfsync everysec
```

### Connection Pool Tuning

```yaml
redis:
  pool_size: 50              # Increase for high concurrency
  max_retries: 3
  min_retry_backoff: "8ms"
  max_retry_backoff: "512ms"
  dial_timeout: "5s"
  read_timeout: "3s"
  write_timeout: "3s"
  pool_timeout: "4s"
  idle_timeout: "5m"
  idle_check_frequency: "1m"
```

### Lua Script Optimization

- Keep scripts small and focused
- Use Redis data structures efficiently
- Minimize cross-slot operations in cluster mode
- Cache frequently used values

## Migration Guide

### From BoltDB to Redis

1. **Preparation**:
   ```bash
   # Backup existing locks
   atlantis locks export --format=json > locks_backup.json
   ```

2. **Configuration**:
   ```yaml
   # Update atlantis.yaml
   enhanced_locking:
     backend: "redis"
     legacy_fallback: true  # Enable during migration
   ```

3. **Migration**:
   ```bash
   # Import locks to Redis
   atlantis locks import --format=json locks_backup.json

   # Verify migration
   atlantis locks list --backend=redis
   ```

4. **Validation**:
   ```bash
   # Test lock operations
   atlantis locks test --backend=redis

   # Disable fallback
   atlantis config set enhanced_locking.legacy_fallback false
   ```

## Best Practices

### Production Deployment

1. **High Availability**: Use Redis Sentinel or Cluster
2. **Monitoring**: Implement comprehensive metrics collection
3. **Backup**: Regular snapshots and point-in-time recovery
4. **Security**: Enable authentication and encryption
5. **Performance**: Monitor and tune based on usage patterns

### Scaling Considerations

1. **Horizontal Scaling**: Use Redis Cluster for large deployments
2. **Vertical Scaling**: Increase memory and CPU for single instance
3. **Connection Management**: Tune pool sizes based on load
4. **Queue Management**: Monitor queue depths and adjust timeouts

### Maintenance

1. **Regular Updates**: Keep Redis version current
2. **Memory Management**: Monitor and configure eviction policies
3. **Log Analysis**: Regular review of error patterns
4. **Performance Testing**: Periodic load testing

## API Reference

### Backend Interface

```go
type Backend interface {
    // Core operations
    TryLock(ctx context.Context, request LockRequest) (*LockResponse, error)
    Unlock(ctx context.Context, lockKey string) error
    GetLock(ctx context.Context, lockKey string) (*EnhancedLock, error)
    ListLocks(ctx context.Context) ([]EnhancedLock, error)

    // Queue operations
    GetQueueStatus(ctx context.Context, lockKey string) (*QueueStatus, error)
    GetUserQueue(ctx context.Context, user string) ([]QueueEntry, error)

    // Advanced features
    RefreshLock(ctx context.Context, lockKey string, ttl time.Duration) error
    TransferLock(ctx context.Context, lockKey string, newUser models.User) error

    // Health and monitoring
    HealthCheck(ctx context.Context) error
    GetMetrics(ctx context.Context) (*Metrics, error)
}
```

### Configuration Schema

```json
{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "type": "object",
  "properties": {
    "enhanced_locking": {
      "type": "object",
      "properties": {
        "backend": {
          "type": "string",
          "enum": ["boltdb", "redis"]
        },
        "redis": {
          "type": "object",
          "properties": {
            "addresses": {
              "type": "array",
              "items": {"type": "string"}
            },
            "cluster_mode": {"type": "boolean"},
            "pool_size": {"type": "integer"},
            "default_ttl": {"type": "string"}
          }
        }
      }
    }
  }
}
```

## Conclusion

The Redis backend provides a robust, scalable foundation for Atlantis's enhanced locking system. With features like atomic operations, health monitoring, and cluster support, it enables reliable distributed locking for large-scale Terraform operations.

For additional support and advanced configurations, refer to the [Enhanced Locking System Overview](enhanced-locking-overview.md) and [Integration Guide](locking-integration-guide.md).