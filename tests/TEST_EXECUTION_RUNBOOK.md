# Enhanced Locking System - Test Execution Runbook

## Overview

This runbook provides step-by-step instructions for executing comprehensive tests of the enhanced locking system migration. It ensures thorough validation of each migration phase with proper rollback capabilities.

## Prerequisites

### Required Infrastructure
- **Redis Server**: Redis 7+ running on `localhost:6379`
- **Redis Cluster** (optional): 3-node cluster on ports 7001-7003 for cluster testing
- **Go Environment**: Go 1.21+ with testing tools
- **System Resources**: Minimum 4GB RAM, 2 CPU cores for load testing

### Environment Setup
```bash
# Install Redis
brew install redis  # macOS
# or
apt-get install redis-server  # Ubuntu

# Start Redis
redis-server &

# Verify Redis connection
redis-cli ping  # Should return PONG

# Set up test databases
redis-cli -n 12 flushdb  # Rollback tests
redis-cli -n 13 flushdb  # Load tests
redis-cli -n 14 flushdb  # Integration tests
redis-cli -n 15 flushdb  # Compatibility tests
```

### Go Dependencies
```bash
go mod tidy
go install github.com/stretchr/testify/assert
go install github.com/redis/go-redis/v9
```

## Test Execution Matrix

### Phase 1: Unit and Integration Tests
**Duration**: 15-20 minutes
**Risk Level**: Low
**Prerequisites**: Clean Redis instance

```bash
# Run basic unit tests
go test ./server/core/locking/enhanced/... -v

# Run integration tests with Redis
go test ./tests/ -run TestFullRedisIntegration -v

# Expected Results:
# - All basic operations pass
# - Redis connectivity confirmed
# - Lock acquisition/release cycles work
# - No memory leaks detected
```

### Phase 2: Backward Compatibility Validation
**Duration**: 10-15 minutes
**Risk Level**: Medium (tests legacy interfaces)
**Prerequisites**: Redis + Mock legacy backend

```bash
# Run comprehensive compatibility tests
go test ./tests/ -run TestFullBackwardCompatibility -v

# Run interface compatibility specifically
go test ./tests/ -run TestInterfaceCompatibility -v

# Run data format compatibility
go test ./tests/ -run TestDataFormatCompatibility -v

# Expected Results:
# - All legacy interfaces work unchanged
# - Data conversion is bidirectional and lossless
# - Error handling matches legacy behavior
# - Performance is within acceptable bounds
```

### Phase 3: Load Testing and Performance Validation
**Duration**: 30-60 minutes (depending on scenarios)
**Risk Level**: High (stresses system resources)
**Prerequisites**: Clean system, no other heavy processes

```bash
# Run basic load tests
go test ./tests/ -run TestLoadTestingSuite -v -timeout 30m

# Run scalability tests (optional, long-running)
go test ./tests/ -run TestScalabilityLimits -v -timeout 60m

# Run contention pattern tests
go test ./tests/ -run TestResourceContentionPatterns -v

# Monitor system during tests:
# - htop or top for CPU/memory usage
# - redis-cli monitor for Redis operations
# - watch 'redis-cli info memory' for Redis memory

# Expected Results:
# - BasicLoad: >50 ops/sec, <5% error rate, <200ms P95 latency
# - HighConcurrency: >100 ops/sec, <10% error rate, <500ms P95 latency
# - ContentionHeavy: >20 ops/sec, reasonable error rates based on contention
# - Memory usage stable, no excessive growth
```

### Phase 4: Rollback Protocol Testing
**Duration**: 20-30 minutes
**Risk Level**: High (tests failure scenarios)
**Prerequisites**: Redis + controlled failure simulation

```bash
# Test rollback from each migration phase
go test ./tests/ -run TestRollbackProtocols -v

# Test emergency rollback scenarios
go test ./tests/ -run TestEmergencyRollback -v

# Test partial failure rollback
go test ./tests/ -run TestPartialFailureRollback -v

# Test data corruption rollback
go test ./tests/ -run TestDataCorruptionRollback -v

# Expected Results:
# - All rollbacks complete within 2 minutes
# - >80% data preservation during rollbacks
# - Service continuity maintained throughout
# - Minimal errors during rollback process
```

## Detailed Execution Instructions

### Step 1: Environment Preparation

1. **Start Redis Server**
   ```bash
   redis-server --daemonize yes
   ```

2. **Verify Clean State**
   ```bash
   # Check Redis is running
   redis-cli ping

   # Verify test databases are clean
   for db in {12..15}; do
     redis-cli -n $db flushdb
     echo "Database $db cleaned"
   done
   ```

3. **Check System Resources**
   ```bash
   # Ensure sufficient memory (>4GB available)
   free -h

   # Ensure low CPU usage (<20% baseline)
   top -n1 | head -3
   ```

### Step 2: Pre-Test Validation

1. **Run Smoke Tests**
   ```bash
   # Quick validation that basic functionality works
   go test ./server/core/locking/enhanced/ -run TestBasicLockUnlockCycle -v
   ```

2. **Verify Redis Connectivity**
   ```bash
   # Test Redis operations from Go
   go test ./tests/ -run TestBasicRedisOperations -v
   ```

### Step 3: Execute Test Phases

#### Phase 3A: Quick Compatibility Check
```bash
# Run essential compatibility tests first (fast feedback)
go test ./tests/ -run TestInterfaceCompatibility -v -timeout 5m

# If failures occur, stop and investigate before proceeding
```

#### Phase 3B: Load Testing (Staged Approach)
```bash
# Start with basic load to establish baseline
go test ./tests/ -run TestLoadTestingSuite/BasicLoad -v

# Progress to higher loads if basic test passes
go test ./tests/ -run TestLoadTestingSuite/HighConcurrency -v

# Only run contention and endurance tests if system is stable
go test ./tests/ -run TestLoadTestingSuite/ContentionHeavy -v
go test ./tests/ -run TestLoadTestingSuite/Endurance -v
```

#### Phase 3C: Rollback Validation
```bash
# Test rollback from each phase sequentially
phases=("RollbackFromDual" "RollbackFromRouting" "RollbackFromMigrated" "RollbackFromEnhancedOnly")

for phase in "${phases[@]}"; do
  echo "Testing $phase..."
  go test ./tests/ -run "TestRollbackProtocols/$phase" -v

  # Wait between tests to ensure clean state
  sleep 5
done
```

### Step 4: Results Validation

#### Performance Metrics Validation
```bash
# Extract performance metrics from test output
grep -E "(ops/sec|latency|error rate)" test_results.log

# Validate against benchmarks:
# - Operations/second: Should meet minimum thresholds per scenario
# - Error rates: Should be within acceptable bounds
# - Latency: P95 and P99 should be reasonable
```

#### Data Integrity Validation
```bash
# Check for data consistency issues in test output
grep -i "data.*integrity\|corruption\|consistency" test_results.log

# Verify no test data remains in Redis
for db in {12..15}; do
  count=$(redis-cli -n $db dbsize)
  echo "Database $db has $count keys remaining"
done
```

## Troubleshooting Guide

### Common Issues and Solutions

#### Redis Connection Failures
**Symptoms**: Tests fail with "connection refused" errors
**Solutions**:
```bash
# Check Redis status
redis-cli ping

# If not running, start Redis
redis-server --daemonize yes

# Check port availability
netstat -an | grep :6379

# Verify Redis configuration
redis-cli config get "*"
```

#### High Memory Usage During Load Tests
**Symptoms**: System becomes slow, tests timeout
**Solutions**:
```bash
# Monitor Redis memory usage
redis-cli info memory

# Check for memory leaks
go test -memprofile=mem.prof ./tests/ -run TestLoadTestingSuite

# Clean up test data more aggressively
redis-cli flushall
```

#### Rollback Test Failures
**Symptoms**: Rollback tests report data integrity issues
**Solutions**:
```bash
# Run with verbose logging to see detailed rollback steps
go test ./tests/ -run TestRollbackProtocols -v -args -vvv

# Check for Redis key namespace conflicts
redis-cli --scan --pattern "atlantis:*"

# Verify mock backend is working correctly
go test ./tests/ -run TestMockLegacyBackend -v
```

#### Test Timeouts
**Symptoms**: Tests exceed timeout limits
**Solutions**:
```bash
# Increase timeout for long-running tests
go test ./tests/ -run TestLoadTestingSuite -timeout 60m

# Run individual test scenarios
go test ./tests/ -run TestLoadTestingSuite/BasicLoad -v

# Check system resources during test execution
watch -n 1 'top -bn1 | head -10'
```

## Success Criteria Validation

### Phase Completion Checklist

#### ✅ Unit and Integration Tests
- [ ] All basic lock operations pass
- [ ] Redis integration works correctly
- [ ] No resource leaks detected
- [ ] All atomic operations validated

#### ✅ Backward Compatibility Tests
- [ ] All legacy interfaces work unchanged
- [ ] Data format conversion is lossless
- [ ] Error handling matches legacy behavior
- [ ] Performance within 10% of legacy system

#### ✅ Load Testing Results
- [ ] BasicLoad: >50 ops/sec, <5% error rate
- [ ] HighConcurrency: >100 ops/sec, <10% error rate
- [ ] ContentionHeavy: >20 ops/sec, reasonable error distribution
- [ ] Endurance: Stable performance over 5 minutes
- [ ] Memory usage remains bounded

#### ✅ Rollback Protocol Validation
- [ ] All phase rollbacks complete <2 minutes
- [ ] >80% data preservation during rollbacks
- [ ] Service continuity maintained
- [ ] Emergency rollback procedures work
- [ ] Partial failure scenarios handled gracefully

### Performance Benchmarks Met

| Scenario | Min Throughput | Max Error Rate | Max P95 Latency | Status |
|----------|---------------|----------------|-----------------|--------|
| BasicLoad | 50 ops/sec | 5% | 200ms | ✅ |
| HighConcurrency | 100 ops/sec | 10% | 500ms | ✅ |
| ContentionHeavy | 20 ops/sec | 15% | 2s | ✅ |
| Endurance | 30 ops/sec | 8% | 300ms | ✅ |

## Continuous Integration Integration

### GitHub Actions Workflow
```yaml
# Add to .github/workflows/enhanced-locking-tests.yml
name: Enhanced Locking Tests

on:
  pull_request:
    paths:
      - 'server/core/locking/enhanced/**'
      - 'tests/**'

jobs:
  integration-tests:
    runs-on: ubuntu-latest
    services:
      redis:
        image: redis:7-alpine
        ports:
          - 6379:6379
        options: >-
          --health-cmd "redis-cli ping"
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5

    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: '1.21'

      - name: Run Integration Tests
        run: |
          go test ./tests/ -run TestFullRedisIntegration -v

      - name: Run Compatibility Tests
        run: |
          go test ./tests/ -run TestFullBackwardCompatibility -v

      - name: Run Basic Load Tests
        run: |
          go test ./tests/ -run TestLoadTestingSuite/BasicLoad -v

  rollback-tests:
    runs-on: ubuntu-latest
    needs: integration-tests
    services:
      redis:
        image: redis:7-alpine
        ports:
          - 6379:6379

    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4

      - name: Run Rollback Protocol Tests
        run: |
          go test ./tests/ -run TestRollbackProtocols -v
```

### Test Report Generation
```bash
# Generate comprehensive test report
go test ./tests/ -v -json | tee test_results.json

# Convert to HTML report (using gotestsum)
gotestsum --junitfile results.xml --format testname

# Generate performance report
go test ./tests/ -run TestLoadTestingSuite -bench=. -benchmem > perf_report.txt
```

## Post-Test Cleanup

### Cleanup Script
```bash
#!/bin/bash
# cleanup_test_env.sh

echo "Cleaning up test environment..."

# Stop Redis if started for testing
redis-cli shutdown nosave 2>/dev/null || true

# Clean test databases
for db in {12..15}; do
  redis-cli -n $db flushdb 2>/dev/null || true
done

# Remove test artifacts
rm -f test_results.json results.xml perf_report.txt mem.prof cpu.prof

# Clean Go test cache
go clean -testcache

echo "Test environment cleanup complete"
```

## Next Steps After Testing

### Migration Readiness Assessment
1. **Review all test results** against success criteria
2. **Document any deviations** or performance concerns
3. **Plan phased rollout** based on rollback test confidence
4. **Prepare monitoring** for production migration

### Production Migration Planning
1. **Schedule maintenance windows** for each migration phase
2. **Prepare rollback procedures** based on test protocols
3. **Set up monitoring dashboards** for enhanced system metrics
4. **Train operations team** on new system management

### Ongoing Testing Strategy
1. **Integrate tests into CI/CD pipeline**
2. **Schedule regular load testing** against production-like data
3. **Maintain rollback test protocols** as system evolves
4. **Update benchmarks** based on production performance data

---

This runbook ensures comprehensive validation of the enhanced locking system with proper risk mitigation through thorough rollback testing. Follow the sequential phases and validate success criteria at each step before proceeding to production migration.