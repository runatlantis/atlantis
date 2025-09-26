# Enhanced Locking Migration Test Execution Runbook

## Overview

This runbook provides comprehensive guidance for executing tests during the enhanced locking system migration. It covers all phases of migration testing, from initial validation to production rollout.

## Prerequisites

### Environment Setup

1. **Test Environment Requirements**
   - Go 1.21 or later
   - Docker for Redis testing (optional)
   - Access to test Atlantis instance
   - Sufficient test data and repositories

2. **Dependencies**
   ```bash
   go mod tidy
   go install github.com/stretchr/testify/assert
   go install github.com/redis/go-redis/v9  # For Redis backend tests
   ```

3. **Test Data Setup**
   ```bash
   # Create test repositories
   mkdir -p test-repos/{basic,priority,deadlock}

   # Setup test configuration
   export ATLANTIS_TEST_MODE=true
   export REDIS_URL=redis://localhost:6379  # If testing Redis backend
   ```

## Test Execution Phases

### Phase 1: Basic Enhanced Locking Functionality

**Objective**: Validate core enhanced locking features without advanced capabilities.

#### Pre-execution Checklist
- [ ] Legacy locking system is functional
- [ ] Test environment is clean
- [ ] No active locks exist
- [ ] System resources are adequate

#### Execution Steps

1. **Run Phase 1 Test Suite**
   ```bash
   cd /Users/pepe.amengual/github/atlantis
   go test -v ./tests -run TestPhase1EnhancedLocking -timeout 10m
   ```

2. **Monitor Test Output**
   - Look for any FAILED test cases
   - Monitor resource usage during tests
   - Check for memory leaks or performance degradation

3. **Validate Results**
   - All basic locking operations succeed
   - No regression in performance
   - Error handling works correctly
   - Configuration validation passes

#### Success Criteria
- ✅ All tests pass (100% success rate)
- ✅ No critical errors or panics
- ✅ Performance within acceptable range (< 1s average operation time)
- ✅ Memory usage stable

#### Failure Response
If any tests fail:
1. **Stop immediately** - Do not proceed to Phase 2
2. Review test logs for failure details
3. Fix issues in enhanced locking implementation
4. Re-run Phase 1 tests until all pass
5. Document lessons learned

### Phase 2: Backwards Compatibility Validation

**Objective**: Ensure enhanced locking maintains full compatibility with existing Atlantis interfaces.

#### Pre-execution Checklist
- [ ] Phase 1 tests pass completely
- [ ] Legacy system backup available
- [ ] Test cases cover all legacy interfaces
- [ ] Fallback mechanisms ready

#### Execution Steps

1. **Run Phase 2 Test Suite**
   ```bash
   go test -v ./tests -run TestPhase2BackwardsCompatibility -timeout 15m
   ```

2. **Execute Compatibility Validation**
   ```bash
   # Test specific compatibility scenarios
   go test -v ./tests -run TestLegacyTryLockInterface -timeout 5m
   go test -v ./tests -run TestLegacyListInterface -timeout 5m
   go test -v ./tests -run TestLegacyUnlockByPullInterface -timeout 5m
   ```

3. **Concurrent Compatibility Testing**
   ```bash
   # Test legacy operations under concurrent load
   go test -v ./tests -run TestConcurrentLegacyOperations -timeout 10m
   ```

#### Success Criteria
- ✅ All legacy interfaces work identically
- ✅ No breaking changes in API responses
- ✅ Lock format preservation verified
- ✅ Fallback mechanisms functional
- ✅ Performance parity with legacy system

#### Failure Response
If compatibility tests fail:
1. **Critical**: Legacy compatibility is mandatory
2. Identify specific interface breaking changes
3. Fix adapter layer or enhanced implementation
4. Verify fixes don't break Phase 1 functionality
5. Re-run both Phase 1 and Phase 2 tests

### Phase 3: Advanced Features Testing

**Objective**: Validate advanced enhanced locking capabilities (priority queuing, deadlock detection, etc.).

#### Pre-execution Checklist
- [ ] Phase 1 and 2 tests pass completely
- [ ] Redis instance available (if testing Redis backend)
- [ ] Advanced features configuration ready
- [ ] Performance baseline established

#### Execution Steps

1. **Run Phase 3 Test Suite**
   ```bash
   go test -v ./tests -run TestPhase3AdvancedFeatures -timeout 20m
   ```

2. **Priority Queue Testing**
   ```bash
   go test -v ./tests -run TestPriorityQueueing -timeout 10m
   ```

3. **Deadlock Detection Testing**
   ```bash
   go test -v ./tests -run TestDeadlockDetection -timeout 10m
   ```

4. **Redis Backend Testing** (if available)
   ```bash
   # Ensure Redis is running
   docker run -d --name redis-test -p 6379:6379 redis:latest

   go test -v ./tests -run TestRedisBackendIntegration -timeout 15m

   # Cleanup
   docker stop redis-test && docker rm redis-test
   ```

5. **Performance Under Load Testing**
   ```bash
   go test -v ./tests -run TestAdvancedPerformanceUnderLoad -timeout 30m
   ```

#### Success Criteria
- ✅ Priority queuing works correctly
- ✅ Deadlock detection prevents circular waits
- ✅ Redis backend integration functional
- ✅ Event system operates without errors
- ✅ Performance acceptable under advanced feature load

#### Failure Response
Advanced feature failures may be acceptable for initial rollout:
1. **Assess criticality** of failed features
2. Consider deploying without specific advanced features
3. Document known limitations
4. Plan fixes for subsequent releases

## Migration Validation Testing

### Full Migration Validation

**Objective**: Validate complete migration path from legacy to advanced enhanced locking.

#### Execution Steps

1. **Run Migration Validation Suite**
   ```bash
   go test -v ./tests -run TestMigrationValidation -timeout 45m
   ```

2. **Monitor Migration Phases**
   - Watch for phase transition success/failure
   - Monitor system health during transitions
   - Verify data integrity at each phase

3. **Review Migration Report**
   ```bash
   # Report generated at: /Users/pepe.amengual/github/atlantis/docs/migration_validation_report.json
   cat docs/migration_validation_report.json | jq '.executive_summary'
   cat docs/migration_validation_report.json | jq '.safety_level'
   cat docs/migration_validation_report.json | jq '.recommended_action'
   ```

#### Success Criteria
- ✅ All migration phases complete successfully
- ✅ Safety level is "safe" or "caution"
- ✅ Recommended action is "Proceed with migration"
- ✅ No critical failures reported

## Rollback Testing

### Rollback Validation

**Objective**: Ensure safe rollback capability at each migration phase.

#### Execution Steps

1. **Run Rollback Test Suite**
   ```bash
   go test -v ./tests -run TestRollbackProcedures -timeout 30m
   ```

2. **Test Specific Rollback Scenarios**
   ```bash
   # Test critical rollback paths
   go test -v ./tests -run TestRollbackScenarios/AdvancedToLegacy -timeout 15m
   go test -v ./tests -run TestRollbackScenarios/BasicToLegacy -timeout 10m
   ```

3. **Safety Gate Validation**
   ```bash
   go test -v ./tests -run TestSafetyGates -timeout 5m
   ```

4. **Review Rollback Report**
   ```bash
   cat docs/rollback_test_report.json | jq '.safety_assessment.overall_risk_level'
   cat docs/rollback_test_report.json | jq '.recommendations'
   ```

#### Success Criteria
- ✅ All rollback scenarios succeed
- ✅ Overall risk level is "low" or "medium"
- ✅ Safety gates pass validation
- ✅ Data integrity maintained during rollback

## Production Readiness Checklist

Before deploying to production, ensure:

### Testing Completeness
- [ ] Phase 1 tests: 100% pass rate
- [ ] Phase 2 tests: 100% pass rate
- [ ] Phase 3 tests: ≥90% pass rate (acceptable failures documented)
- [ ] Migration validation: "safe" or "caution" level
- [ ] Rollback testing: All critical scenarios pass

### Performance Validation
- [ ] No significant performance regression (≤10% degradation acceptable)
- [ ] Memory usage within acceptable bounds
- [ ] Concurrent operation handling validated
- [ ] Load testing under realistic conditions

### Operational Readiness
- [ ] Monitoring and alerting configured
- [ ] Rollback procedures documented and tested
- [ ] Team trained on new system capabilities
- [ ] Emergency contacts and procedures ready

### Documentation and Communication
- [ ] Migration plan reviewed and approved
- [ ] Stakeholders notified of changes
- [ ] User-facing documentation updated
- [ ] Known limitations documented

## Troubleshooting Guide

### Common Test Failures

#### Test Timeouts
**Symptoms**: Tests fail with timeout errors
**Causes**:
- System under heavy load
- Deadlock in test code
- Resource exhaustion

**Solutions**:
1. Increase test timeout values
2. Run tests on dedicated environment
3. Check for resource leaks

#### Lock State Inconsistency
**Symptoms**: Tests fail with "lock not found" or "unexpected lock state"
**Causes**:
- Race conditions in test code
- Improper cleanup between tests
- Backend state corruption

**Solutions**:
1. Add proper test cleanup
2. Use test isolation patterns
3. Reset backend state between tests

#### Redis Connection Failures
**Symptoms**: Redis backend tests fail with connection errors
**Causes**:
- Redis not running
- Network connectivity issues
- Wrong Redis configuration

**Solutions**:
1. Start Redis with `docker run -d -p 6379:6379 redis:latest`
2. Check Redis connectivity with `redis-cli ping`
3. Verify Redis URL configuration

#### Memory Leaks
**Symptoms**: Tests pass but memory usage grows continuously
**Causes**:
- Goroutine leaks in enhanced locking system
- Unclosed channels or connections
- Improper cleanup in test teardown

**Solutions**:
1. Use `go test -race` to detect race conditions
2. Add proper resource cleanup in tests
3. Monitor goroutine counts during tests

### Performance Issues

#### Slow Test Execution
**Symptoms**: Tests take much longer than expected
**Investigation**:
1. Profile test execution: `go test -cpuprofile=cpu.prof -memprofile=mem.prof`
2. Check for excessive allocations
3. Look for inefficient algorithms

**Solutions**:
1. Optimize hot paths in enhanced locking
2. Reduce test iteration counts for development
3. Use parallel test execution where safe

#### High Resource Usage
**Symptoms**: Tests consume excessive CPU or memory
**Solutions**:
1. Implement proper connection pooling
2. Add resource limits to test configuration
3. Use mock backends for unit tests

## Emergency Procedures

### Critical Test Failures

If critical tests fail during production migration:

1. **STOP IMMEDIATELY** - Do not proceed with migration
2. **Assess Impact** - Determine if system is in safe state
3. **Execute Rollback** - Use tested rollback procedures
4. **Notify Stakeholders** - Inform team of migration halt
5. **Root Cause Analysis** - Identify and fix underlying issues
6. **Re-validate** - Run full test suite before retry

### Production Issues After Deployment

If issues arise in production after deployment:

1. **Monitor System Health** - Check for lock corruption or system instability
2. **Assess Rollback Necessity** - Use rollback test results to guide decision
3. **Execute Rollback if Needed** - Follow tested rollback procedures
4. **Collect Diagnostics** - Gather logs and system state information
5. **Post-Incident Review** - Document what went wrong and improve tests

## Test Report Analysis

### Interpreting Test Results

#### Migration Validation Report
```json
{
  "safety_level": "safe|caution|danger",
  "recommended_action": "Proceed with migration|Proceed with caution|Do not proceed",
  "readiness_assessment": {
    "ready_for_next_phase": true/false,
    "blocking_issues": [...],
    "warning_issues": [...]
  }
}
```

**Actions Based on Safety Level**:
- **Safe**: Proceed with confidence, normal monitoring
- **Caution**: Proceed with enhanced monitoring, rapid rollback capability
- **Danger**: Do not proceed, fix critical issues first

#### Rollback Test Report
```json
{
  "safety_assessment": {
    "overall_risk_level": "low|medium|high",
    "data_loss_risk": "none|low|medium|high",
    "service_disruption_risk": "none|low|medium|high"
  }
}
```

**Risk Level Guidelines**:
- **Low**: Safe to rollback anytime
- **Medium**: Plan rollback during maintenance window
- **High**: Emergency-only rollback, significant risks

## Continuous Integration Integration

### CI Pipeline Integration

Add to your CI/CD pipeline:

```yaml
# Example GitHub Actions workflow
name: Enhanced Locking Tests
on: [push, pull_request]

jobs:
  enhanced-locking-tests:
    runs-on: ubuntu-latest
    services:
      redis:
        image: redis:latest
        ports:
          - 6379:6379

    steps:
    - uses: actions/checkout@v3
    - uses: actions/setup-go@v4
      with:
        go-version: '1.21'

    - name: Run Phase 1 Tests
      run: go test -v ./tests -run TestPhase1EnhancedLocking -timeout 10m

    - name: Run Phase 2 Tests
      run: go test -v ./tests -run TestPhase2BackwardsCompatibility -timeout 15m

    - name: Run Phase 3 Tests
      run: go test -v ./tests -run TestPhase3AdvancedFeatures -timeout 20m

    - name: Run Migration Validation
      run: go test -v ./tests -run TestMigrationValidation -timeout 45m

    - name: Run Rollback Tests
      run: go test -v ./tests -run TestRollbackProcedures -timeout 30m

    - name: Upload Test Reports
      uses: actions/upload-artifact@v3
      if: always()
      with:
        name: test-reports
        path: docs/*_report.json
```

## Conclusion

This runbook provides comprehensive guidance for testing the enhanced locking migration. Follow the phases sequentially, validate success criteria at each step, and be prepared to rollback if issues arise.

Remember:
- **Safety First**: Never proceed with failed critical tests
- **Document Everything**: Record all issues and resolutions
- **Test Thoroughly**: Better to catch issues in testing than production
- **Plan for Rollback**: Always have a tested rollback path ready

For questions or issues not covered in this runbook, consult the development team or create an issue in the repository.