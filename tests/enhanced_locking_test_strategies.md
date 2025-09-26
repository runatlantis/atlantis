# Enhanced Locking System - Comprehensive Testing Strategy

## Overview

This document outlines comprehensive testing strategies for migrating Atlantis from the legacy BoltDB-based locking system to the enhanced Redis-based locking system with priority queuing, retry mechanisms, and deadlock prevention.

## Architecture Analysis

### Current Enhanced System Components
- **Enhanced Manager**: Priority queuing, retry logic, deadlock detection
- **Redis Backend**: Atomic operations with Lua scripts, TTL support, pub/sub events
- **BoltDB Adapter**: Backward compatibility layer
- **Priority Queue**: Multi-level resource-based queuing
- **Timeout Manager**: Lock expiration and heartbeat mechanisms
- **Deadlock Detector**: Cycle detection and resolution

### Migration Phases Identified
1. **Phase 1**: Enable enhanced system alongside legacy (dual mode)
2. **Phase 2**: Route new requests to enhanced system, legacy reads remain
3. **Phase 3**: Full migration with legacy fallback
4. **Phase 4**: Remove legacy system entirely

## Testing Framework Architecture

### Test Categories
- **Unit Tests**: Individual component testing
- **Integration Tests**: System-level behavior validation
- **Migration Tests**: Phase transition validation
- **Load Tests**: Performance and scalability validation
- **Compatibility Tests**: Backward compatibility verification
- **Rollback Tests**: Safe fallback scenario validation

## Phase-Specific Test Strategies

### Phase 1: Dual Mode Testing
**Objective**: Verify enhanced system works alongside legacy without interference

**Test Scenarios**:
```go
// Dual Mode Operation Test
func TestDualModeOperation(t *testing.T) {
    // Both systems should handle different resources independently
    legacyLock := acquireLegacyLock("project1", "workspace1")
    enhancedLock := acquireEnhancedLock("project2", "workspace2")

    // Verify no interference
    verifyLockIsolation(legacyLock, enhancedLock)

    // Test configuration switching
    testConfigurationSwitching()

    // Verify metrics separation
    verifyMetricsIsolation()
}
```

**Validation Criteria**:
- ✅ Legacy locks unaffected by enhanced system
- ✅ Enhanced features work independently
- ✅ Configuration changes apply correctly
- ✅ No resource leaks or conflicts
- ✅ Metrics are properly segmented

### Phase 2: Routing Migration Testing
**Objective**: Validate selective routing to enhanced system

**Test Scenarios**:
```go
// Selective Routing Test
func TestSelectiveRouting(t *testing.T) {
    // Configure routing rules
    setRoutingRule("priority_projects", ENHANCED_BACKEND)
    setRoutingRule("legacy_projects", LEGACY_BACKEND)

    // Test routing decisions
    verifyRoutingDecision("priority/repo", ENHANCED_BACKEND)
    verifyRoutingDecision("legacy/repo", LEGACY_BACKEND)

    // Test migration of existing locks
    migrateExistingLock("project1", "workspace1")
    verifyLockMigration("project1", "workspace1")
}
```

**Validation Criteria**:
- ✅ Routing rules applied correctly
- ✅ Existing locks migrate safely
- ✅ No data loss during migration
- ✅ Performance maintained during transition
- ✅ Error handling for migration failures

### Phase 3: Full Migration with Fallback Testing
**Objective**: Validate complete migration with safety fallback

**Test Scenarios**:
```go
// Full Migration Test
func TestFullMigrationWithFallback(t *testing.T) {
    // Enable full enhanced mode with fallback
    config := EnhancedConfig{
        Enabled: true,
        LegacyFallback: true,
        PreserveLegacyFormat: true
    }

    // Test normal operations
    testEnhancedOperations(config)

    // Test fallback scenarios
    simulateRedisFailure()
    verifyFallbackToLegacy()

    // Test recovery scenarios
    restoreRedisService()
    verifyRecoveryToEnhanced()
}
```

**Validation Criteria**:
- ✅ All operations use enhanced system by default
- ✅ Fallback triggers correctly on failures
- ✅ Data consistency maintained during fallback
- ✅ Recovery works seamlessly
- ✅ No lock loss or corruption

### Phase 4: Legacy System Removal Testing
**Objective**: Validate safe removal of legacy components

**Test Scenarios**:
```go
// Legacy Removal Test
func TestLegacySystemRemoval(t *testing.T) {
    // Ensure no legacy dependencies
    verifyNoLegacyReferences()

    // Test enhanced-only operations
    testPureBandedOperations()

    // Verify cleanup procedures
    testLegacyDataCleanup()

    // Performance validation
    measurePerformanceImprovement()
}
```

**Validation Criteria**:
- ✅ No references to legacy backend
- ✅ All features work without legacy fallback
- ✅ Legacy data properly cleaned up
- ✅ Performance improvements realized
- ✅ System stability maintained

## Backward Compatibility Test Suite

### Interface Compatibility Tests
```go
func TestBackwardCompatibilityInterface(t *testing.T) {
    adapter := NewLockingAdapter(enhancedManager, redisBackend, config, nil, logger)

    // Test all legacy interface methods
    testTryLockInterface(adapter)
    testUnlockInterface(adapter)
    testListInterface(adapter)
    testUnlockByPullInterface(adapter)
    testGetLockInterface(adapter)

    // Verify response formats match exactly
    verifyResponseFormatCompatibility(adapter)
}
```

### Data Format Compatibility Tests
```go
func TestDataFormatCompatibility(t *testing.T) {
    // Create lock in legacy format
    legacyLock := createLegacyLock()

    // Convert to enhanced format
    enhancedLock := convertToEnhanced(legacyLock)

    // Convert back to legacy format
    reconvertedLock := convertToLegacy(enhancedLock)

    // Verify data integrity
    verifyDataIntegrity(legacyLock, reconvertedLock)
}
```

## Redis Backend Integration Testing

### Connection and Clustering Tests
```go
func TestRedisIntegration(t *testing.T) {
    // Single Redis instance
    testSingleRedisInstance()

    // Redis cluster mode
    testRedisClusterMode()

    // Redis Sentinel mode
    testRedisSentinelMode()

    // Connection pooling
    testConnectionPooling()

    // Failover scenarios
    testRedisFailover()
}
```

### Atomic Operations Testing
```go
func TestRedisAtomicOperations(t *testing.T) {
    // Lua script execution
    testLuaScriptAtomicity()

    // Concurrent modifications
    testConcurrentAtomicOperations()

    // Transaction rollback
    testTransactionRollback()

    // Pipeline operations
    testPipelineConsistency()
}
```

### Redis Performance Testing
```go
func TestRedisPerformance(t *testing.T) {
    // Latency measurements
    measureRedisLatency()

    // Throughput testing
    measureRedisThroughput()

    // Memory usage
    measureRedisMemoryUsage()

    // Connection overhead
    measureConnectionOverhead()
}
```

## Load Testing Framework

### Concurrent Operations Testing
```go
func TestConcurrentOperations(t *testing.T) {
    numClients := 100
    numOperations := 1000

    // Create concurrent clients
    clients := createConcurrentClients(numClients)

    // Execute concurrent operations
    results := executeConcurrentOperations(clients, numOperations, []Operation{
        LOCK_ACQUIRE,
        LOCK_RELEASE,
        LOCK_QUERY,
        QUEUE_OPERATIONS,
    })

    // Analyze results
    analyzePerformanceResults(results)
    verifyNoDeadlocks(results)
    verifyDataConsistency(results)
}
```

### Scalability Testing
```go
func TestScalabilityLimits(t *testing.T) {
    // Test increasing load patterns
    testLoadPatterns := []int{10, 50, 100, 500, 1000, 5000}

    for _, load := range testLoadPatterns {
        result := runLoadTest(load)
        validatePerformanceMetrics(result)

        if result.ErrorRate > 0.01 { // 1% error threshold
            t.Errorf("Error rate too high at load %d: %f", load, result.ErrorRate)
        }
    }
}
```

### Priority Queue Load Testing
```go
func TestPriorityQueuePerformance(t *testing.T) {
    // Create mixed priority workload
    workload := createMixedPriorityWorkload(1000, map[Priority]float64{
        PriorityCritical: 0.1,
        PriorityHigh:     0.2,
        PriorityNormal:   0.5,
        PriorityLow:      0.2,
    })

    // Execute and measure priority ordering
    results := executeWorkload(workload)
    verifyPriorityOrdering(results)
    measureQueueLatency(results)
}
```

## Rollback Testing Protocols

### Safe Rollback Procedures
```go
func TestSafeRollbackProcedures(t *testing.T) {
    // Each phase should have rollback capability
    phases := []MigrationPhase{PHASE_1, PHASE_2, PHASE_3, PHASE_4}

    for _, phase := range phases {
        // Progress to phase
        migrateToPhase(phase)

        // Create test state
        createTestState(phase)

        // Execute rollback
        rollbackResult := executeRollback(phase)

        // Verify rollback success
        verifyRollbackIntegrity(rollbackResult)
        verifyNoDataLoss(rollbackResult)
        verifyServiceContinuity(rollbackResult)
    }
}
```

### Emergency Rollback Testing
```go
func TestEmergencyRollback(t *testing.T) {
    // Simulate critical failures during migration
    failures := []FailureScenario{
        REDIS_CLUSTER_FAILURE,
        DATA_CORRUPTION,
        NETWORK_PARTITION,
        MEMORY_EXHAUSTION,
    }

    for _, failure := range failures {
        // Induce failure during migration
        simulateFailure(failure)

        // Trigger emergency rollback
        emergencyRollback()

        // Verify system recovery
        verifySystemRecovery()
        verifyDataIntegrity()
    }
}
```

## Migration Validation Test Matrix

| Test Scenario | Phase 1 | Phase 2 | Phase 3 | Phase 4 |
|---------------|---------|---------|---------|---------|
| **Basic Lock Operations** | ✅ | ✅ | ✅ | ✅ |
| **Concurrent Access** | ✅ | ✅ | ✅ | ✅ |
| **Priority Queuing** | ❌ | ✅ | ✅ | ✅ |
| **Timeout Handling** | ❌ | ✅ | ✅ | ✅ |
| **Deadlock Prevention** | ❌ | ✅ | ✅ | ✅ |
| **Retry Mechanisms** | ❌ | ✅ | ✅ | ✅ |
| **Legacy Fallback** | ✅ | ✅ | ✅ | ❌ |
| **Data Migration** | ❌ | ✅ | ✅ | ✅ |
| **Performance Monitoring** | ✅ | ✅ | ✅ | ✅ |
| **Rollback Capability** | ✅ | ✅ | ✅ | ❌ |

## Advanced Testing Scenarios

### Chaos Engineering Tests
```go
func TestChaosScenarios(t *testing.T) {
    chaosScenarios := []ChaosTest{
        // Network failures
        {"NetworkPartition", simulateNetworkPartition, 30 * time.Second},
        {"PacketLoss", simulatePacketLoss, 60 * time.Second},

        // Resource exhaustion
        {"MemoryExhaustion", simulateMemoryExhaustion, 45 * time.Second},
        {"CPUStarvation", simulateCPUStarvation, 30 * time.Second},

        // Service failures
        {"RedisFailure", simulateRedisFailure, 120 * time.Second},
        {"AtlantisRestart", simulateAtlantisRestart, 90 * time.Second},
    }

    for _, scenario := range chaosScenarios {
        runChaosTest(scenario)
    }
}
```

### Multi-Instance Testing
```go
func TestMultiInstanceBehavior(t *testing.T) {
    // Test behavior with multiple Atlantis instances
    instances := createMultipleInstances(3)

    // Test distributed locking behavior
    testDistributedLocking(instances)

    // Test leader election (if applicable)
    testLeaderElection(instances)

    // Test split-brain scenarios
    testSplitBrainPrevention(instances)
}
```

### Edge Case Testing
```go
func TestEdgeCases(t *testing.T) {
    // Test boundary conditions
    testMaximumLockCount()
    testMaximumQueueSize()
    testMaximumTimeout()

    // Test unusual timing scenarios
    testRapidLockUnlockCycles()
    testNearTimeoutOperations()
    testClockSkewScenarios()

    // Test resource exhaustion
    testRedisMemoryLimits()
    testConnectionPoolExhaustion()
    testQueueOverflow()
}
```

## Test Execution Framework

### Automated Test Pipeline
```yaml
# .github/workflows/enhanced-locking-tests.yml
name: Enhanced Locking System Tests

on:
  pull_request:
    paths:
      - 'server/core/locking/enhanced/**'
      - 'server/core/locking/**'

jobs:
  unit-tests:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v3
      - run: go test ./server/core/locking/enhanced/...

  integration-tests:
    runs-on: ubuntu-latest
    services:
      redis:
        image: redis:7-alpine
        options: >-
          --health-cmd "redis-cli ping"
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v3
      - run: go test -tags integration ./server/core/locking/enhanced/tests/...

  load-tests:
    runs-on: ubuntu-latest
    if: github.event.pull_request.base.ref == 'main'
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v3
      - run: go test -tags load ./server/core/locking/enhanced/tests/...
```

### Test Environment Configuration
```go
// Test configuration for different environments
type TestConfig struct {
    RedisURL          string
    RedisClusterNodes []string
    TestDuration      time.Duration
    ConcurrentClients int
    Operations        int
    ChaosEnabled      bool
}

var TestEnvironments = map[string]TestConfig{
    "unit": {
        RedisURL:          "redis://localhost:6379",
        TestDuration:      30 * time.Second,
        ConcurrentClients: 10,
        Operations:        100,
        ChaosEnabled:      false,
    },
    "integration": {
        RedisURL:          "redis://redis:6379",
        TestDuration:      5 * time.Minute,
        ConcurrentClients: 50,
        Operations:        1000,
        ChaosEnabled:      false,
    },
    "load": {
        RedisClusterNodes: []string{
            "redis-1:6379", "redis-2:6379", "redis-3:6379",
        },
        TestDuration:      15 * time.Minute,
        ConcurrentClients: 500,
        Operations:        10000,
        ChaosEnabled:      false,
    },
    "chaos": {
        RedisClusterNodes: []string{
            "redis-1:6379", "redis-2:6379", "redis-3:6379",
        },
        TestDuration:      30 * time.Minute,
        ConcurrentClients: 100,
        Operations:        5000,
        ChaosEnabled:      true,
    },
}
```

## Success Criteria and Metrics

### Performance Benchmarks
- **Lock Acquisition Latency**: < 10ms P95, < 50ms P99
- **Throughput**: > 1000 operations/second
- **Queue Processing**: < 1s average wait time for normal priority
- **Memory Usage**: < 2x current BoltDB memory usage
- **Error Rate**: < 0.1% under normal load

### Reliability Metrics
- **Availability**: 99.9% uptime during migration
- **Data Consistency**: Zero lock corruption or loss
- **Rollback Success Rate**: 100% successful rollbacks
- **Recovery Time**: < 30s after Redis failover

### Compatibility Metrics
- **API Compatibility**: 100% existing API compatibility
- **Data Format Compatibility**: Seamless legacy data handling
- **Configuration Compatibility**: All existing configs supported

This comprehensive testing strategy ensures safe migration from the legacy locking system to the enhanced Redis-based system while maintaining backward compatibility and system reliability.