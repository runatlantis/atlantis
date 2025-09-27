# Redis Backend Foundation - Enhanced Locking System

## Overview

This document describes the Redis backend implementation for Atlantis Enhanced Locking System, providing distributed storage with clustering support, health monitoring, and high-performance atomic operations.

## Architecture

### Core Components

1. **RedisBackend** - Main backend implementation
2. **ScriptManager** - Lua script management and execution
3. **HealthMonitor** - Comprehensive health monitoring with circuit breaker
4. **ClusterManager** - Redis cluster coordination and consensus

### Distributed Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Atlantis      │    │   Atlantis      │    │   Atlantis      │
│   Node 1        │    │   Node 2        │    │   Node 3        │
│                 │    │                 │    │                 │
│ ┌─────────────┐ │    │ ┌─────────────┐ │    │ ┌─────────────┐ │
│ │ Enhanced    │ │    │ │ Enhanced    │ │    │ │ Enhanced    │ │
│ │ Lock Mgr    │ │    │ │ Lock Mgr    │ │    │ │ Lock Mgr    │ │
│ └─────────────┘ │    │ └─────────────┘ │    │ └─────────────┘ │
│ ┌─────────────┐ │    │ ┌─────────────┐ │    │ ┌─────────────┐ │
│ │ Redis       │ │    │ │ Redis       │ │    │ │ Redis       │ │
│ │ Backend     │ │◄───┤ │ Backend     │ │◄───┤ │ Backend     │ │
│ └─────────────┘ │    │ └─────────────┘ │    │ └─────────────┘ │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         └───────────────────────┼───────────────────────┘
                                 │
                                 ▼
                    ┌─────────────────────────┐
                    │    Redis Cluster        │
                    │                         │
                    │  ┌─────┐ ┌─────┐ ┌─────┐│
                    │  │Node1│ │Node2│ │Node3││
                    │  └─────┘ └─────┘ └─────┘│
                    │                         │
                    │  • Atomic Operations    │
                    │  • Leader Election      │
                    │  • Consensus Mechanism  │
                    │  • Health Monitoring    │
                    └─────────────────────────┘
```

## Configuration

### Basic Redis Configuration

```yaml
enhanced_locking:
  enabled: true
  backend: "redis"
  redis:
    addresses:
      - "redis-1:6379"
      - "redis-2:6379"
      - "redis-3:6379"
    password: "${REDIS_PASSWORD}"
    db: 0
    pool_size: 20
    key_prefix: "atlantis:enhanced:lock:"
    default_ttl: "1h"
    connection_timeout: "5s"
    read_timeout: "3s"
    write_timeout: "3s"
    cluster_mode: true
```

### Clustering Configuration

```yaml
enhanced_locking:
  redis:
    cluster_mode: true
    cluster_config:
      node_id: "${ATLANTIS_NODE_ID}"
      heartbeat_interval: "30s"
      node_timeout: "90s"
      leader_election_timeout: "10s"
      consensus_threshold: 2
      enable_leader_election: true
      enable_consensus: true
      max_cluster_size: 7
      replication_factor: 3
```

### Health Monitoring Configuration

```yaml
enhanced_locking:
  redis:
    health_monitoring:
      check_interval: "30s"
      timeout_threshold: "5s"
      error_threshold: 5
      recovery_threshold: 3
      circuit_breaker_enabled: true
      performance_monitoring: true
      alerting_enabled: true
      max_slow_queries: 10
      memory_threshold_mb: 1024
```

## Features

### Atomic Operations with Lua Scripts

The Redis backend uses Lua scripts for atomic operations:

- **acquire_lock** - Atomic lock acquisition with queue support
- **release_lock** - Atomic lock release with queue processing
- **refresh_lock** - TTL extension
- **transfer_lock** - Ownership transfer
- **cleanup_expired** - Batch cleanup of expired locks
- **distributed_acquire** - Cluster consensus for distributed locks

### Health Monitoring

- **Circuit Breaker** - Automatic failure detection and recovery
- **Performance Metrics** - Response time tracking and bottleneck detection
- **Memory Monitoring** - Redis memory usage and threshold alerts
- **Connection Pool Stats** - Connection health and performance
- **Slow Query Detection** - Identification of performance issues

### Clustering Support

- **Leader Election** - Automatic leader selection using Redis consensus
- **Node Discovery** - Dynamic cluster membership management
- **Failure Detection** - Heartbeat-based node health monitoring
- **Split-Brain Prevention** - Consensus mechanisms to prevent conflicts
- **Load Balancing** - Intelligent distribution of lock operations

### TTL Management

- **Automatic Expiration** - Configurable lock timeouts
- **Advanced Cleanup** - Batch processing for expired locks
- **TTL Extension** - Dynamic timeout adjustment
- **Memory Optimization** - Efficient cleanup routines

## Deployment Guide

### Redis Cluster Setup

#### Option 1: Docker Compose

```yaml
version: '3.8'
services:
  redis-1:
    image: redis:7-alpine
    ports:
      - "7001:6379"
    command: redis-server --cluster-enabled yes --cluster-config-file nodes.conf --cluster-node-timeout 5000 --appendonly yes
    volumes:
      - redis-1-data:/data

  redis-2:
    image: redis:7-alpine
    ports:
      - "7002:6379"
    command: redis-server --cluster-enabled yes --cluster-config-file nodes.conf --cluster-node-timeout 5000 --appendonly yes
    volumes:
      - redis-2-data:/data

  redis-3:
    image: redis:7-alpine
    ports:
      - "7003:6379"
    command: redis-server --cluster-enabled yes --cluster-config-file nodes.conf --cluster-node-timeout 5000 --appendonly yes
    volumes:
      - redis-3-data:/data

volumes:
  redis-1-data:
  redis-2-data:
  redis-3-data:
```

#### Option 2: Kubernetes

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: redis-cluster-config
data:
  redis.conf: |
    cluster-enabled yes
    cluster-config-file nodes.conf
    cluster-node-timeout 5000
    appendonly yes
    dir /data
    port 6379

---
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
        - /etc/redis/redis.conf
        volumeMounts:
        - name: data
          mountPath: /data
        - name: config
          mountPath: /etc/redis
      volumes:
      - name: config
        configMap:
          name: redis-cluster-config
  volumeClaimTemplates:
  - metadata:
      name: data
    spec:
      accessModes: [ "ReadWriteOnce" ]
      resources:
        requests:
          storage: 10Gi
```

### Environment Variables

```bash
# Redis Configuration
export REDIS_ADDRESSES="redis-1:6379,redis-2:6379,redis-3:6379"
export REDIS_PASSWORD="your-secure-password"
export REDIS_DB=0

# Cluster Configuration
export ATLANTIS_NODE_ID="$(hostname)-$(date +%s)"
export REDIS_CLUSTER_MODE=true

# Health Monitoring
export REDIS_HEALTH_CHECK_INTERVAL=30s
export REDIS_CIRCUIT_BREAKER_ENABLED=true
```

### Atlantis Configuration

```yaml
# atlantis.yaml
enhanced-locking:
  enabled: true
  backend: redis
  redis:
    addresses:
      - ${REDIS_ADDRESSES}
    password: ${REDIS_PASSWORD}
    cluster_mode: ${REDIS_CLUSTER_MODE}
    pool_size: 20
    key_prefix: "atlantis:enhanced:lock:"
    default_ttl: 1h

  # Enable advanced features
  enable_priority_queue: true
  enable_retries: true
  enable_metrics: true

  # Timeouts
  default_timeout: 30s
  max_timeout: 5m
  max_queue_size: 1000
```

## Performance Optimization

### Connection Pooling

The Redis backend includes connection pool optimization:

```go
// Warm up connections on startup
backend.WarmupConnections(ctx, 20)

// Monitor pool statistics
stats := backend.GetConnectionPoolStats()
fmt.Printf("Pool hits: %d, misses: %d", stats["hits"], stats["misses"])
```

### Lua Script Optimization

Scripts are preloaded and cached for optimal performance:

- Uses EVALSHA for faster execution
- Automatic fallback to EVAL if script not cached
- Batch operations to reduce round trips

### Performance Benchmarking

Built-in benchmarking framework:

```go
benchmark, err := backend.BenchmarkBasicOperations(ctx, 1000)
if err == nil {
    fmt.Printf("Operations per second: %.2f", benchmark.OpsPerSecond)
    fmt.Printf("Average latency: %v", benchmark.GetLatency)
}
```

## Monitoring and Alerts

### Health Check Endpoints

```bash
# Basic health check
curl http://atlantis:4141/api/enhanced-locks/health

# Detailed health report
curl http://atlantis:4141/api/enhanced-locks/health/detailed

# Cluster status
curl http://atlantis:4141/api/enhanced-locks/cluster/status
```

### Metrics Collection

The system exposes various metrics:

- Lock acquisition rates
- Queue depths
- Response times
- Error rates
- Circuit breaker state
- Cluster health

### Alerting

Configure alerts based on:

- Circuit breaker state changes
- High error rates
- Memory threshold breaches
- Slow query detection
- Node failures

## Troubleshooting

### Common Issues

#### Connection Pool Exhaustion

```bash
# Check pool statistics
curl http://atlantis:4141/api/enhanced-locks/redis/pool-stats

# Solution: Increase pool size
export REDIS_POOL_SIZE=50
```

#### High Memory Usage

```bash
# Check Redis memory usage
redis-cli INFO memory

# Clean up expired locks
curl -X POST http://atlantis:4141/api/enhanced-locks/cleanup
```

#### Cluster Split-Brain

```bash
# Check cluster state
curl http://atlantis:4141/api/enhanced-locks/cluster/nodes

# Check leader election
redis-cli GET atlantis:cluster:leader
```

### Debug Commands

```bash
# Enable debug logging
export ATLANTIS_LOG_LEVEL=debug

# Monitor lock operations
redis-cli MONITOR

# Check script cache
redis-cli SCRIPT EXISTS <script-sha>

# View cluster nodes
redis-cli CLUSTER NODES
```

## Security Considerations

### Authentication

- Use strong Redis passwords
- Enable Redis AUTH
- Configure TLS/SSL for production

### Network Security

- Use private networks for Redis cluster
- Implement firewall rules
- Consider Redis Sentinel for high availability

### Access Control

- Implement Redis ACLs
- Use dedicated Redis users for Atlantis
- Regularly rotate credentials

## Performance Tuning

### Redis Configuration

```conf
# Memory optimization
maxmemory 2gb
maxmemory-policy allkeys-lru

# Network optimization
tcp-keepalive 300
timeout 0

# Persistence optimization
save 900 1
save 300 10
save 60 10000
```

### Connection Optimization

- Tune connection pool sizes based on load
- Monitor connection usage patterns
- Use connection multiplexing where possible

### Lua Script Optimization

- Keep scripts atomic and efficient
- Minimize Redis calls within scripts
- Use appropriate data structures

## Migration Guide

### From BoltDB to Redis

1. **Preparation**
   ```bash
   # Backup existing locks
   atlantis locks export > locks-backup.json
   ```

2. **Configuration Update**
   ```yaml
   enhanced_locking:
     backend: redis  # Change from boltdb
   ```

3. **Data Migration**
   ```bash
   # Import locks to Redis backend
   atlantis locks import locks-backup.json
   ```

### Cluster Migration

For migrating to clustered Redis:

1. Set up Redis cluster
2. Update configuration
3. Restart Atlantis nodes one by one
4. Verify cluster health

## API Reference

### Health Check API

```bash
GET /api/enhanced-locks/health
GET /api/enhanced-locks/health/detailed
GET /api/enhanced-locks/metrics
```

### Cluster Management API

```bash
GET /api/enhanced-locks/cluster/status
GET /api/enhanced-locks/cluster/nodes
POST /api/enhanced-locks/cluster/leader/election
```

### Performance API

```bash
POST /api/enhanced-locks/benchmark
GET /api/enhanced-locks/pool-stats
POST /api/enhanced-locks/connections/warmup
```

## Conclusion

The Redis backend provides a robust, scalable foundation for the Enhanced Locking System with:

- **High Performance** - Lua scripts and connection pooling
- **High Availability** - Clustering and health monitoring
- **Operational Excellence** - Comprehensive monitoring and debugging tools
- **Security** - Authentication and access control features

This foundation supports enterprise-grade deployments with automatic failover, performance optimization, and operational visibility.