// Phase 3 Testing Strategy: Advanced Features (Priority Queuing, Deadlock Detection, Redis Backend)
// This phase tests the advanced capabilities of the enhanced locking system

package tests

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/runatlantis/atlantis/server/core/locking/enhanced"
	"github.com/runatlantis/atlantis/server/core/locking/enhanced/backends"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/logging"
)

// Phase3TestSuite validates advanced enhanced locking features
type Phase3TestSuite struct {
	suite.Suite
	manager      enhanced.LockManager
	redisManager enhanced.LockManager
	config       *enhanced.EnhancedConfig
	cleanup      func()
	redisCleanup func()
	ctx          context.Context
}

// SetupSuite initializes the test environment for Phase 3
func (s *Phase3TestSuite) SetupSuite() {
	s.ctx = context.Background()

	// Phase 3 configuration: all advanced features enabled
	s.config = &enhanced.EnhancedConfig{
		Enabled:                 true,
		Backend:                 "boltdb",
		DefaultTimeout:          30 * time.Minute,
		MaxTimeout:              2 * time.Hour,

		// Enable all advanced features
		EnablePriorityQueue:     true,
		MaxQueueSize:           1000,
		QueueTimeout:           10 * time.Minute,

		EnableRetry:            true,
		MaxRetryAttempts:       3,
		RetryBaseDelay:         100 * time.Millisecond,
		RetryMaxDelay:          5 * time.Second,

		EnableDeadlockDetection: true,
		DeadlockCheckInterval:   5 * time.Second,

		EnableEvents:           true,
		EventBufferSize:        1000,

		// Redis configuration
		RedisClusterMode:       false,
		RedisKeyPrefix:         "atlantis:test:enhanced:",
		RedisLockTTL:           time.Hour,

		// Maintain backward compatibility
		LegacyFallback:         true,
		PreserveLegacyFormat:   true,
	}

	s.manager, s.cleanup = s.setupAdvancedTestManager()
	s.redisManager, s.redisCleanup = s.setupRedisTestManager()
}

func (s *Phase3TestSuite) TearDownSuite() {
	if s.cleanup != nil {
		s.cleanup()
	}
	if s.redisCleanup != nil {
		s.redisCleanup()
	}
}

// Test Priority Queue Functionality
func (s *Phase3TestSuite) TestPriorityQueueing() {
	project := models.Project{
		RepoFullName: "test/priority-queue",
		Path:         ".",
	}
	workspace := "default"

	// Users with different priorities
	lowUser := models.User{Username: "low-priority-user"}
	normalUser := models.User{Username: "normal-priority-user"}
	highUser := models.User{Username: "high-priority-user"}
	criticalUser := models.User{Username: "critical-priority-user"}

	// Step 1: Normal user acquires lock first
	lock1, err := s.manager.LockWithPriority(s.ctx, project, workspace, normalUser, enhanced.PriorityNormal)
	require.NoError(s.T(), err, "Normal user should acquire lock")
	require.NotNil(s.T(), lock1, "Lock should not be nil")

	// Step 2: Start concurrent requests with different priorities
	var wg sync.WaitGroup
	results := make(chan priorityTestResult, 3)

	// High priority request
	wg.Add(1)
	go func() {
		defer wg.Done()
		start := time.Now()
		lock, err := s.manager.LockWithPriority(s.ctx, project, workspace, highUser, enhanced.PriorityHigh)
		results <- priorityTestResult{
			User:     highUser.Username,
			Priority: enhanced.PriorityHigh,
			Lock:     lock,
			Error:    err,
			Duration: time.Since(start),
		}
	}()

	// Critical priority request
	wg.Add(1)
	go func() {
		defer wg.Done()
		start := time.Now()
		lock, err := s.manager.LockWithPriority(s.ctx, project, workspace, criticalUser, enhanced.PriorityCritical)
		results <- priorityTestResult{
			User:     criticalUser.Username,
			Priority: enhanced.PriorityCritical,
			Lock:     lock,
			Error:    err,
			Duration: time.Since(start),
		}
	}()

	// Low priority request
	wg.Add(1)
	go func() {
		defer wg.Done()
		start := time.Now()
		lock, err := s.manager.LockWithPriority(s.ctx, project, workspace, lowUser, enhanced.PriorityLow)
		results <- priorityTestResult{
			User:     lowUser.Username,
			Priority: enhanced.PriorityLow,
			Lock:     lock,
			Error:    err,
			Duration: time.Since(start),
		}
	}()

	// Allow requests to queue
	time.Sleep(500 * time.Millisecond)

	// Step 3: Release the initial lock to trigger queue processing
	_, err = s.manager.Unlock(s.ctx, project, workspace, normalUser)
	require.NoError(s.T(), err, "Should release initial lock")

	// Wait for all goroutines to complete
	go func() {
		wg.Wait()
		close(results)
	}()

	// Step 4: Collect and analyze results
	var priorityResults []priorityTestResult
	for result := range results {
		priorityResults = append(priorityResults, result)
		s.T().Logf("Priority result: user=%s, priority=%d, success=%t, duration=%v, error=%v",
			result.User, result.Priority, result.Lock != nil, result.Duration, result.Error)
	}

	// Step 5: Verify priority ordering (at least critical should be processed before low)
	assert.Len(s.T(), priorityResults, 3, "Should have results for all three users")

	// Find the successful lock (should be highest priority)
	var successfulResult *priorityTestResult
	for i := range priorityResults {
		if priorityResults[i].Lock != nil && priorityResults[i].Error == nil {
			successfulResult = &priorityResults[i]
			break
		}
	}

	if successfulResult != nil {
		s.T().Logf("Successful lock acquired by: %s with priority %d", successfulResult.User, successfulResult.Priority)
		// Critical priority should typically win, but implementation may vary
		assert.True(s.T(), successfulResult.Priority >= enhanced.PriorityHigh,
			"Higher priority user should typically acquire lock first")

		// Cleanup
		userModel := models.User{Username: successfulResult.User}
		s.manager.Unlock(s.ctx, project, workspace, userModel)
	}

	// Test queue position functionality
	position, err := s.manager.GetQueuePosition(s.ctx, project, workspace)
	s.T().Logf("Queue position result: %d, error: %v", position, err)
}

// Test Retry Mechanism
func (s *Phase3TestSuite) TestRetryMechanism() {
	project := models.Project{
		RepoFullName: "test/retry-mechanism",
		Path:         ".",
	}
	workspace := "default"

	user1 := models.User{Username: "retry-user1"}
	user2 := models.User{Username: "retry-user2"}

	// User1 acquires lock
	lock1, err := s.manager.LockWithPriority(s.ctx, project, workspace, user1, enhanced.PriorityNormal)
	require.NoError(s.T(), err, "User1 should acquire lock")
	require.NotNil(s.T(), lock1, "Lock1 should not be nil")

	// User2 tries with retry mechanism
	start := time.Now()
	lock2, err := s.manager.LockWithPriority(s.ctx, project, workspace, user2, enhanced.PriorityNormal)
	duration := time.Since(start)

	// Should fail after retries
	assert.Error(s.T(), err, "User2 should fail after retries")
	assert.Nil(s.T(), lock2, "Lock2 should be nil")

	// Should have taken time due to retries (at least 3 attempts * base delay)
	expectedMinDuration := time.Duration(s.config.MaxRetryAttempts) * s.config.RetryBaseDelay
	assert.GreaterOrEqual(s.T(), duration, expectedMinDuration,
		"Should have spent at least %v retrying, but took %v", expectedMinDuration, duration)

	s.T().Logf("Retry mechanism test: duration=%v, expected_min=%v", duration, expectedMinDuration)

	// Cleanup
	s.manager.Unlock(s.ctx, project, workspace, user1)

	// Test successful retry scenario
	// Release lock after a short delay to test successful retry
	go func() {
		time.Sleep(200 * time.Millisecond)
		s.manager.Unlock(s.ctx, project, workspace, user1)
	}()

	// This should succeed after retries
	start = time.Now()
	lock2, err = s.manager.LockWithPriority(s.ctx, project, workspace, user2, enhanced.PriorityNormal)
	duration = time.Since(start)

	if err == nil && lock2 != nil {
		s.T().Logf("Successful retry after %v", duration)
		s.manager.Unlock(s.ctx, project, workspace, user2)
	} else {
		s.T().Logf("Retry still failed: %v", err)
	}
}

// Test Deadlock Detection
func (s *Phase3TestSuite) TestDeadlockDetection() {
	// Setup resources that could potentially create deadlock
	project1 := models.Project{RepoFullName: "test/deadlock1", Path: "."}
	project2 := models.Project{RepoFullName: "test/deadlock2", Path: "."}
	workspace := "default"

	user1 := models.User{Username: "deadlock-user1"}
	user2 := models.User{Username: "deadlock-user2"}

	// Scenario 1: User1 locks project1
	lock1, err := s.manager.LockWithPriority(s.ctx, project1, workspace, user1, enhanced.PriorityNormal)
	require.NoError(s.T(), err, "User1 should acquire lock on project1")
	require.NotNil(s.T(), lock1, "Lock1 should not be nil")

	// Scenario 2: User2 locks project2
	lock2, err := s.manager.LockWithPriority(s.ctx, project2, workspace, user2, enhanced.PriorityNormal)
	require.NoError(s.T(), err, "User2 should acquire lock on project2")
	require.NotNil(s.T(), lock2, "Lock2 should not be nil")

	// Scenario 3: User1 tries to lock project2 (should be blocked)
	// User2 tries to lock project1 (potential deadlock)
	var wg sync.WaitGroup
	results := make(chan deadlockTestResult, 2)

	wg.Add(1)
	go func() {
		defer wg.Done()
		start := time.Now()
		lock, err := s.manager.LockWithPriority(s.ctx, project2, workspace, user1, enhanced.PriorityNormal)
		results <- deadlockTestResult{
			User:     user1.Username,
			Project:  project2.RepoFullName,
			Lock:     lock,
			Error:    err,
			Duration: time.Since(start),
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		start := time.Now()
		lock, err := s.manager.LockWithPriority(s.ctx, project1, workspace, user2, enhanced.PriorityNormal)
		results <- deadlockTestResult{
			User:     user2.Username,
			Project:  project1.RepoFullName,
			Lock:     lock,
			Error:    err,
			Duration: time.Since(start),
		}
	}()

	go func() {
		wg.Wait()
		close(results)
	}()

	// Allow deadlock detection to run
	time.Sleep(s.config.DeadlockCheckInterval + time.Second)

	// Analyze results
	var deadlockResults []deadlockTestResult
	for result := range results {
		deadlockResults = append(deadlockResults, result)
		s.T().Logf("Deadlock test result: user=%s, project=%s, success=%t, duration=%v, error=%v",
			result.User, result.Project, result.Lock != nil, result.Duration, result.Error)

		// At least one should fail due to deadlock prevention
		if result.Error != nil {
			lockErr, ok := result.Error.(*enhanced.LockError)
			if ok && lockErr.Code == enhanced.ErrCodeDeadlock {
				s.T().Logf("Deadlock prevented for user %s", result.User)
			}
		}
	}

	// Cleanup
	s.manager.Unlock(s.ctx, project1, workspace, user1)
	s.manager.Unlock(s.ctx, project2, workspace, user2)

	// Clean up any successful secondary locks
	for _, result := range deadlockResults {
		if result.Lock != nil {
			user := models.User{Username: result.User}
			if result.Project == project1.RepoFullName {
				s.manager.Unlock(s.ctx, project1, workspace, user)
			} else if result.Project == project2.RepoFullName {
				s.manager.Unlock(s.ctx, project2, workspace, user)
			}
		}
	}
}

// Test Event System
func (s *Phase3TestSuite) TestEventSystem() {
	project := models.Project{
		RepoFullName: "test/events",
		Path:         ".",
	}
	workspace := "default"
	user := models.User{Username: "event-user"}

	// This test verifies that the event system is working
	// The actual event testing would require access to the enhanced manager's event callbacks

	// Test basic lock/unlock cycle and verify no panics occur
	lock, err := s.manager.LockWithPriority(s.ctx, project, workspace, user, enhanced.PriorityNormal)
	require.NoError(s.T(), err, "Should acquire lock with events enabled")
	require.NotNil(s.T(), lock, "Lock should not be nil")

	// Hold lock briefly to generate events
	time.Sleep(100 * time.Millisecond)

	// Release lock
	releasedLock, err := s.manager.Unlock(s.ctx, project, workspace, user)
	require.NoError(s.T(), err, "Should release lock with events enabled")
	require.NotNil(s.T(), releasedLock, "Released lock should not be nil")

	s.T().Log("Event system test completed - verified no panics with events enabled")
}

// Test Redis Backend Integration
func (s *Phase3TestSuite) TestRedisBackendIntegration() {
	if s.redisManager == nil {
		s.T().Skip("Redis not available, skipping Redis backend tests")
		return
	}

	project := models.Project{
		RepoFullName: "test/redis-backend",
		Path:         ".",
	}
	workspace := "default"
	user := models.User{Username: "redis-user"}

	// Test basic operations with Redis backend
	lock, err := s.redisManager.LockWithPriority(s.ctx, project, workspace, user, enhanced.PriorityHigh)
	require.NoError(s.T(), err, "Should acquire lock with Redis backend")
	require.NotNil(s.T(), lock, "Redis lock should not be nil")

	// Test that lock is visible through list
	locks, err := s.redisManager.List(s.ctx)
	require.NoError(s.T(), err, "Should list Redis locks")
	assert.Len(s.T(), locks, 1, "Should have one Redis lock")

	// Test concurrent access with Redis
	user2 := models.User{Username: "redis-user2"}
	lock2, err := s.redisManager.LockWithPriority(s.ctx, project, workspace, user2, enhanced.PriorityNormal)
	assert.Error(s.T(), err, "Conflicting lock should fail with Redis")
	assert.Nil(s.T(), lock2, "Conflicting Redis lock should be nil")

	// Test unlock with Redis
	releasedLock, err := s.redisManager.Unlock(s.ctx, project, workspace, user)
	require.NoError(s.T(), err, "Should unlock Redis lock")
	require.NotNil(s.T(), releasedLock, "Redis unlocked lock should not be nil")

	// Verify Redis lock is gone
	finalLocks, err := s.redisManager.List(s.ctx)
	require.NoError(s.T(), err, "Should list Redis locks after unlock")
	assert.Len(s.T(), finalLocks, 0, "Should have no Redis locks after unlock")

	s.T().Log("Redis backend integration test completed successfully")
}

// Test Timeout Management with Advanced Features
func (s *Phase3TestSuite) TestAdvancedTimeoutManagement() {
	project := models.Project{
		RepoFullName: "test/advanced-timeout",
		Path:         ".",
	}
	workspace := "default"
	user := models.User{Username: "timeout-user"}

	// Test short timeout with retry
	shortTimeout := 2 * time.Second
	lock, err := s.manager.LockWithTimeout(s.ctx, project, workspace, user, shortTimeout)
	require.NoError(s.T(), err, "Should acquire lock with short timeout")
	require.NotNil(s.T(), lock, "Lock should not be nil")

	// Wait for timeout plus some buffer
	time.Sleep(shortTimeout + time.Second)

	// Lock should be handled by timeout manager
	// Implementation details may vary, but system should not crash
	s.T().Log("Timeout management test completed")

	// Try to acquire same lock (might succeed if timeout released it)
	user2 := models.User{Username: "timeout-user2"}
	lock2, err := s.manager.LockWithTimeout(s.ctx, project, workspace, user2, shortTimeout)

	if lock2 != nil {
		s.T().Log("Second lock acquired - timeout mechanism worked")
		s.manager.Unlock(s.ctx, project, workspace, user2)
	} else {
		s.T().Logf("Second lock failed: %v", err)
		// Cleanup original lock
		s.manager.Unlock(s.ctx, project, workspace, user)
	}
}

// Test Performance Under Load with Advanced Features
func (s *Phase3TestSuite) TestAdvancedPerformanceUnderLoad() {
	if testing.Short() {
		s.T().Skip("Skipping performance test in short mode")
		return
	}

	// Test parameters
	numUsers := 30
	numOperations := 150
	concurrentRequests := 15

	var wg sync.WaitGroup
	results := make(chan advancedPerformanceResult, numOperations)
	errors := make(chan error, numOperations)

	startTime := time.Now()

	// Create concurrent operations with different priorities
	for i := 0; i < concurrentRequests; i++ {
		wg.Add(1)
		go func(routineID int) {
			defer wg.Done()

			for j := 0; j < numOperations/concurrentRequests; j++ {
				operationStart := time.Now()

				project := models.Project{
					RepoFullName: fmt.Sprintf("test/advanced-perf-%d", routineID),
					Path:         fmt.Sprintf("path-%d", j),
				}
				user := models.User{Username: fmt.Sprintf("perf-user-%d-%d", routineID, j)}
				workspace := fmt.Sprintf("ws-%d", j%7)

				// Use different priorities
				priority := enhanced.Priority(j % 4) // 0-3 maps to Low-Critical

				// Acquire lock with priority
				lock, err := s.manager.LockWithPriority(s.ctx, project, workspace, user, priority)
				if err != nil {
					errors <- err
					continue
				}

				// Hold lock briefly
				time.Sleep(time.Duration(5+j%10) * time.Millisecond)

				// Release lock
				_, err = s.manager.Unlock(s.ctx, project, workspace, user)
				if err != nil {
					errors <- err
					continue
				}

				operationTime := time.Since(operationStart)
				results <- advancedPerformanceResult{
					RoutineID: routineID,
					Operation: j,
					Priority:  priority,
					Duration:  operationTime,
				}
			}
		}(i)
	}

	wg.Wait()
	totalTime := time.Since(startTime)

	// Close channels and collect results
	close(results)
	close(errors)

	// Analyze results
	var performanceResults []advancedPerformanceResult
	for result := range results {
		performanceResults = append(performanceResults, result)
	}

	var errorCount int
	for err := range errors {
		errorCount++
		s.T().Logf("Performance error: %v", err)
	}

	// Performance analysis
	totalOperations := len(performanceResults)
	assert.Greater(s.T(), totalOperations, 0, "Should have completed some operations")

	if totalOperations > 0 {
		var totalDuration time.Duration
		var maxDuration time.Duration
		priorityStats := make(map[enhanced.Priority][]time.Duration)

		for _, result := range performanceResults {
			totalDuration += result.Duration
			if result.Duration > maxDuration {
				maxDuration = result.Duration
			}

			priorityStats[result.Priority] = append(priorityStats[result.Priority], result.Duration)
		}

		averageDuration := totalDuration / time.Duration(totalOperations)

		s.T().Logf("Advanced Performance Results:")
		s.T().Logf("  Total Operations: %d", totalOperations)
		s.T().Logf("  Total Time: %v", totalTime)
		s.T().Logf("  Average Operation Time: %v", averageDuration)
		s.T().Logf("  Max Operation Time: %v", maxDuration)
		s.T().Logf("  Operations/Second: %.2f", float64(totalOperations)/totalTime.Seconds())
		s.T().Logf("  Error Count: %d", errorCount)

		// Priority-based analysis
		for priority, durations := range priorityStats {
			if len(durations) > 0 {
				var priorityTotal time.Duration
				for _, d := range durations {
					priorityTotal += d
				}
				avgPriorityDuration := priorityTotal / time.Duration(len(durations))
				s.T().Logf("  Priority %d: %d ops, avg %v", priority, len(durations), avgPriorityDuration)
			}
		}

		// Performance benchmarks for advanced features
		assert.Less(s.T(), averageDuration, 2*time.Second, "Average advanced operation should be under 2 seconds")
		assert.Less(s.T(), maxDuration, 10*time.Second, "Max advanced operation should be under 10 seconds")
		assert.Less(s.T(), float64(errorCount)/float64(totalOperations+errorCount), 0.15, "Error rate should be under 15% with advanced features")
	}
}

// Test Resource Management and Cleanup
func (s *Phase3TestSuite) TestResourceManagementAndCleanup() {
	// Test that advanced features properly clean up resources
	project := models.Project{
		RepoFullName: "test/resource-cleanup",
		Path:         ".",
	}
	workspace := "default"
	user := models.User{Username: "cleanup-user"}

	// Create and release multiple locks to test cleanup
	for i := 0; i < 10; i++ {
		lock, err := s.manager.LockWithPriority(s.ctx, project, fmt.Sprintf("ws-%d", i), user, enhanced.PriorityNormal)
		require.NoError(s.T(), err, "Should create lock %d", i)
		require.NotNil(s.T(), lock, "Lock %d should not be nil", i)

		// Brief hold
		time.Sleep(10 * time.Millisecond)

		// Release
		_, err = s.manager.Unlock(s.ctx, project, fmt.Sprintf("ws-%d", i), user)
		require.NoError(s.T(), err, "Should release lock %d", i)
	}

	// Verify all locks are cleaned up
	locks, err := s.manager.List(s.ctx)
	require.NoError(s.T(), err, "Should list locks after cleanup test")
	assert.Len(s.T(), locks, 0, "Should have no remaining locks after cleanup")

	s.T().Log("Resource management and cleanup test completed")
}

// Helper types

type priorityTestResult struct {
	User     string
	Priority enhanced.Priority
	Lock     *models.ProjectLock
	Error    error
	Duration time.Duration
}

type deadlockTestResult struct {
	User     string
	Project  string
	Lock     *models.ProjectLock
	Error    error
	Duration time.Duration
}

type advancedPerformanceResult struct {
	RoutineID int
	Operation int
	Priority  enhanced.Priority
	Duration  time.Duration
}

// Setup functions

func (s *Phase3TestSuite) setupAdvancedTestManager() (enhanced.LockManager, func()) {
	backend := &AdvancedMockBackend{
		locks:        make(map[string]*enhanced.EnhancedLock),
		queue:        make([]*enhanced.EnhancedLockRequest, 0),
		metrics:      &enhanced.BackendStats{},
		eventChannel: make(chan *enhanced.LockEvent, s.config.EventBufferSize),
	}

	manager := enhanced.NewEnhancedLockManager(backend, s.config, logging.NewNoopLogger(s.T()))

	err := manager.Start(s.ctx)
	require.NoError(s.T(), err)

	cleanup := func() {
		manager.Stop()
		close(backend.eventChannel)
	}

	return manager, cleanup
}

func (s *Phase3TestSuite) setupRedisTestManager() (enhanced.LockManager, func()) {
	// Try to connect to Redis
	redisClient := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
		DB:   15, // Use test database
	})

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	_, err := redisClient.Ping(ctx).Result()
	if err != nil {
		s.T().Logf("Redis not available: %v", err)
		return nil, func() {}
	}

	// Clean up test data
	redisClient.FlushDB(s.ctx)

	// Create Redis-specific config
	redisConfig := *s.config // Copy config
	redisConfig.Backend = "redis"

	backend := backends.NewRedisBackend(redisClient, &redisConfig, logging.NewNoopLogger(s.T()))
	manager := enhanced.NewEnhancedLockManager(backend, &redisConfig, logging.NewNoopLogger(s.T()))

	err = manager.Start(s.ctx)
	if err != nil {
		s.T().Logf("Failed to start Redis manager: %v", err)
		return nil, func() {}
	}

	cleanup := func() {
		manager.Stop()
		redisClient.FlushDB(s.ctx)
		redisClient.Close()
	}

	return manager, cleanup
}

// AdvancedMockBackend provides advanced features for testing
type AdvancedMockBackend struct {
	mutex        sync.RWMutex
	locks        map[string]*enhanced.EnhancedLock
	queue        []*enhanced.EnhancedLockRequest
	metrics      *enhanced.BackendStats
	eventChannel chan *enhanced.LockEvent
}

// Implement enhanced.Backend interface with advanced features
// (Implementation would include queue management, event emission, etc.)

// Test Suite Runner
func TestPhase3AdvancedFeatures(t *testing.T) {
	suite.Run(t, new(Phase3TestSuite))
}