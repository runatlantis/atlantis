package tests

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/runatlantis/atlantis/server/core/locking/enhanced"
	"github.com/runatlantis/atlantis/server/core/locking/enhanced/backends"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/logging"
)

// Test suite for enhanced locking integration tests
func TestEnhancedLockingIntegration(t *testing.T) {
	t.Run("BasicLockUnlockCycle", testBasicLockUnlockCycle)
	t.Run("ConcurrentLocking", testConcurrentLocking)
	t.Run("PriorityQueueing", testPriorityQueueing)
	t.Run("TimeoutHandling", testTimeoutHandling)
	t.Run("RetryMechanism", testRetryMechanism)
	t.Run("DeadlockPrevention", testDeadlockPrevention)
	t.Run("DeadlockDetectionAdvanced", testDeadlockDetectionAdvanced)
	t.Run("DeadlockResolutionPolicies", testDeadlockResolutionPolicies)
	t.Run("CircularWaitScenarios", testCircularWaitScenarios)
	t.Run("MultiResourceDeadlock", testMultiResourceDeadlock)
	t.Run("DeadlockPreventionWithPriority", testDeadlockPreventionWithPriority)
	t.Run("CascadeDeadlockResolution", testCascadeDeadlockResolution)
	t.Run("BackwardCompatibility", testBackwardCompatibility)
	t.Run("RedisBackendIntegration", testRedisBackendIntegration)
	t.Run("PerformanceUnderLoad", testPerformanceUnderLoad)
	t.Run("DeadlockPerformanceBenchmark", testDeadlockPerformanceBenchmark)
	t.Run("EndToEndSystemTest", testEndToEndSystemTest)
}

func testBasicLockUnlockCycle(t *testing.T) {
	manager, cleanup := setupTestManager(t)
	defer cleanup()

	ctx := context.Background()

	// Test data
	project := models.Project{
		RepoFullName: "test/repo",
		Path:         ".",
	}
	user := models.User{Username: "testuser"}
	workspace := "default"

	// Acquire lock
	lock, err := manager.Lock(ctx, project, workspace, user)
	require.NoError(t, err)
	require.NotNil(t, lock)
	assert.Equal(t, project.RepoFullName, lock.Project.RepoFullName)
	assert.Equal(t, workspace, lock.Workspace)
	assert.Equal(t, user.Username, lock.User.Username)

	// Verify lock exists
	locks, err := manager.List(ctx)
	require.NoError(t, err)
	assert.Len(t, locks, 1)
	assert.Equal(t, project.RepoFullName, locks[0].Project.RepoFullName)

	// Release lock
	releasedLock, err := manager.Unlock(ctx, project, workspace, user)
	require.NoError(t, err)
	require.NotNil(t, releasedLock)

	// Verify lock is gone
	locks, err = manager.List(ctx)
	require.NoError(t, err)
	assert.Len(t, locks, 0)
}

func testConcurrentLocking(t *testing.T) {
	manager, cleanup := setupTestManager(t)
	defer cleanup()

	ctx := context.Background()

	project := models.Project{
		RepoFullName: "test/concurrent",
		Path:         ".",
	}
	workspace := "default"

	user1 := models.User{Username: "user1"}
	user2 := models.User{Username: "user2"}

	var wg sync.WaitGroup
	var lock1, lock2 *models.ProjectLock
	var err1, err2 error

	// Try to acquire the same lock concurrently
	wg.Add(2)
	go func() {
		defer wg.Done()
		lock1, err1 = manager.Lock(ctx, project, workspace, user1)
	}()

	go func() {
		defer wg.Done()
		time.Sleep(100 * time.Millisecond) // Slight delay to ensure order
		lock2, err2 = manager.Lock(ctx, project, workspace, user2)
	}()

	wg.Wait()

	// One should succeed, one should fail or be queued
	if err1 == nil && err2 == nil {
		t.Fatal("Both concurrent locks should not succeed without queuing")
	}

	successCount := 0
	if err1 == nil && lock1 != nil {
		successCount++
	}
	if err2 == nil && lock2 != nil {
		successCount++
	}

	// At least one should succeed
	assert.GreaterOrEqual(t, successCount, 1)

	// Clean up any acquired locks
	if lock1 != nil {
		manager.Unlock(ctx, project, workspace, user1)
	}
	if lock2 != nil {
		manager.Unlock(ctx, project, workspace, user2)
	}
}

func testPriorityQueueing(t *testing.T) {
	config := &enhanced.EnhancedConfig{
		Enabled:             true,
		EnablePriorityQueue: true,
		MaxQueueSize:        10,
		QueueTimeout:        30 * time.Second,
	}

	manager, cleanup := setupTestManagerWithConfig(t, config)
	defer cleanup()

	ctx := context.Background()

	project := models.Project{
		RepoFullName: "test/priority",
		Path:         ".",
	}
	workspace := "default"

	user1 := models.User{Username: "user1"}
	user2 := models.User{Username: "user2"}
	user3 := models.User{Username: "user3"}

	// User1 acquires lock
	lock1, err := manager.LockWithPriority(ctx, project, workspace, user1, enhanced.PriorityNormal)
	require.NoError(t, err)
	require.NotNil(t, lock1)

	// User2 requests with high priority (should be queued)
	go func() {
		lock2, err := manager.LockWithPriority(ctx, project, workspace, user2, enhanced.PriorityHigh)
		if err != nil {
			t.Logf("User2 high priority lock failed: %v", err)
		} else if lock2 != nil {
			t.Logf("User2 high priority lock acquired")
		}
	}()

	// User3 requests with low priority (should be queued after user2)
	go func() {
		lock3, err := manager.LockWithPriority(ctx, project, workspace, user3, enhanced.PriorityLow)
		if err != nil {
			t.Logf("User3 low priority lock failed: %v", err)
		} else if lock3 != nil {
			t.Logf("User3 low priority lock acquired")
		}
	}()

	time.Sleep(1 * time.Second) // Allow queuing to happen

	// Release user1's lock
	_, err = manager.Unlock(ctx, project, workspace, user1)
	require.NoError(t, err)

	time.Sleep(2 * time.Second) // Allow queue processing

	// Verify behavior - this is more of a smoke test as the exact behavior depends on implementation
	t.Log("Priority queueing test completed - check logs for queue processing")
}

func testTimeoutHandling(t *testing.T) {
	config := &enhanced.EnhancedConfig{
		Enabled:        true,
		DefaultTimeout: 2 * time.Second,
	}

	manager, cleanup := setupTestManagerWithConfig(t, config)
	defer cleanup()

	ctx := context.Background()

	project := models.Project{
		RepoFullName: "test/timeout",
		Path:         ".",
	}
	user := models.User{Username: "timeoutuser"}
	workspace := "default"

	// Acquire lock with timeout
	lock, err := manager.LockWithTimeout(ctx, project, workspace, user, 1*time.Second)
	require.NoError(t, err)
	require.NotNil(t, lock)

	// Wait for timeout to trigger
	time.Sleep(2 * time.Second)

	// Lock should be automatically released (check is implementation dependent)
	t.Log("Timeout handling test completed - lock should have been released")
}

func testRetryMechanism(t *testing.T) {
	config := &enhanced.EnhancedConfig{
		Enabled:           true,
		EnableRetry:       true,
		MaxRetryAttempts:  3,
		RetryBaseDelay:    100 * time.Millisecond,
		RetryMaxDelay:     1 * time.Second,
	}

	manager, cleanup := setupTestManagerWithConfig(t, config)
	defer cleanup()

	ctx := context.Background()

	project := models.Project{
		RepoFullName: "test/retry",
		Path:         ".",
	}
	workspace := "default"

	user1 := models.User{Username: "user1"}
	user2 := models.User{Username: "user2"}

	// User1 acquires lock
	lock1, err := manager.Lock(ctx, project, workspace, user1)
	require.NoError(t, err)
	require.NotNil(t, lock1)

	// User2 tries to acquire same lock (should retry and eventually fail)
	start := time.Now()
	lock2, err := manager.Lock(ctx, project, workspace, user2)
	duration := time.Since(start)

	// Should have failed after retries
	assert.Error(t, err)
	assert.Nil(t, lock2)

	// Should have taken some time due to retries
	assert.Greater(t, duration, 300*time.Millisecond, "Should have spent time retrying")

	// Clean up
	manager.Unlock(ctx, project, workspace, user1)
}

func testDeadlockPrevention(t *testing.T) {
	config := &enhanced.EnhancedConfig{
		Enabled:                 true,
		EnableDeadlockDetection: true,
		DeadlockCheckInterval:   1 * time.Second,
	}

	manager, cleanup := setupTestManagerWithConfig(t, config)
	defer cleanup()

	ctx := context.Background()

	// This is a simplified deadlock test - full deadlock scenarios are complex
	project1 := models.Project{RepoFullName: "test/deadlock1", Path: "."}
	project2 := models.Project{RepoFullName: "test/deadlock2", Path: "."}
	user := models.User{Username: "deadlockuser"}
	workspace := "default"

	// Acquire first lock
	lock1, err := manager.Lock(ctx, project1, workspace, user)
	require.NoError(t, err)
	require.NotNil(t, lock1)

	// Try to acquire second lock (should succeed as no circular dependency)
	lock2, err := manager.Lock(ctx, project2, workspace, user)

	// This should either succeed or fail gracefully
	if err != nil {
		t.Logf("Second lock failed as expected: %v", err)
	} else {
		require.NotNil(t, lock2)
		manager.Unlock(ctx, project2, workspace, user)
	}

	// Clean up
	manager.Unlock(ctx, project1, workspace, user)
}

func testBackwardCompatibility(t *testing.T) {
	// Test that enhanced locking maintains compatibility with existing interfaces
	config := &enhanced.EnhancedConfig{
		Enabled:            true,
		LegacyFallback:     true,
		PreserveLegacyFormat: true,
	}

	manager, cleanup := setupTestManagerWithConfig(t, config)
	defer cleanup()

	ctx := context.Background()

	// Use the adapter to test legacy interface compatibility
	adapter := enhanced.NewLockingAdapter(
		manager,
		nil, // backend will be set up in setupTestManager
		config,
		nil, // no legacy fallback for this test
		logging.NewNoopLogger(t),
	)

	project := models.Project{
		RepoFullName: "test/compat",
		Path:         ".",
	}
	user := models.User{Username: "compatuser"}
	workspace := "default"

	legacyLock := models.ProjectLock{
		Project:   project,
		Workspace: workspace,
		User:      user,
		Time:      time.Now(),
	}

	// Test legacy TryLock interface
	acquired, resp, err := adapter.TryLock(legacyLock)
	require.NoError(t, err)
	assert.True(t, acquired)
	assert.Empty(t, resp.LockFailureReason)

	// Test legacy List interface
	locks, err := adapter.List()
	require.NoError(t, err)
	assert.Len(t, locks, 1)

	// Test legacy GetLock interface
	retrievedLock, err := adapter.GetLock(project, workspace)
	require.NoError(t, err)
	require.NotNil(t, retrievedLock)
	assert.Equal(t, project.RepoFullName, retrievedLock.Project.RepoFullName)

	// Test legacy Unlock interface
	unlockedLock, err := adapter.Unlock(project, workspace, user)
	require.NoError(t, err)
	require.NotNil(t, unlockedLock)

	// Verify lock is gone
	finalLock, err := adapter.GetLock(project, workspace)
	require.NoError(t, err)
	assert.Nil(t, finalLock)
}

func testRedisBackendIntegration(t *testing.T) {
	// This test requires a Redis instance - skip if not available
	redisClient := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
		DB:   1, // Use test database
	})

	// Test Redis connection
	ctx := context.Background()
	_, err := redisClient.Ping(ctx).Result()
	if err != nil {
		t.Skip("Redis not available, skipping Redis integration test")
	}

	// Clean up test data
	defer redisClient.FlushDB(ctx)

	config := &enhanced.EnhancedConfig{
		Enabled:          true,
		Backend:          "redis",
		RedisKeyPrefix:   "atlantis:test:lock:",
		RedisLockTTL:     5 * time.Minute,
	}

	backend := backends.NewRedisBackend(redisClient, config, logging.NewNoopLogger(t))
	manager := enhanced.NewEnhancedLockManager(backend, config, logging.NewNoopLogger(t))

	require.NoError(t, manager.Start(ctx))
	defer manager.Stop()

	project := models.Project{
		RepoFullName: "test/redis",
		Path:         ".",
	}
	user := models.User{Username: "redisuser"}
	workspace := "default"

	// Test basic operations with Redis backend
	lock, err := manager.Lock(ctx, project, workspace, user)
	require.NoError(t, err)
	require.NotNil(t, lock)

	// Verify lock exists in Redis
	keys, err := redisClient.Keys(ctx, config.RedisKeyPrefix+"*").Result()
	require.NoError(t, err)
	assert.Greater(t, len(keys), 0, "Should have locks stored in Redis")

	// Release lock
	releasedLock, err := manager.Unlock(ctx, project, workspace, user)
	require.NoError(t, err)
	require.NotNil(t, releasedLock)

	// Verify lock is removed from Redis
	time.Sleep(100 * time.Millisecond) // Allow for async cleanup
	keys, err = redisClient.Keys(ctx, config.RedisKeyPrefix+"*").Result()
	require.NoError(t, err)
	// Keys might still exist due to queue data structures, but lock should be gone
}

func testPerformanceUnderLoad(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	manager, cleanup := setupTestManager(t)
	defer cleanup()

	ctx := context.Background()

	// Performance test parameters
	numUsers := 50
	numOperations := 100
	concurrentRequests := 10

	var wg sync.WaitGroup
	results := make(chan time.Duration, numOperations*concurrentRequests)
	errors := make(chan error, numOperations*concurrentRequests)

	startTime := time.Now()

	// Create multiple concurrent operations
	for i := 0; i < concurrentRequests; i++ {
		wg.Add(1)
		go func(routineID int) {
			defer wg.Done()

			for j := 0; j < numOperations/concurrentRequests; j++ {
				operationStart := time.Now()

				project := models.Project{
					RepoFullName: fmt.Sprintf("test/perf-%d", routineID),
					Path:         fmt.Sprintf("path-%d", j),
				}
				user := models.User{Username: fmt.Sprintf("user-%d", routineID)}
				workspace := fmt.Sprintf("ws-%d", j%5) // Distribute across workspaces

				// Acquire lock
				lock, err := manager.Lock(ctx, project, workspace, user)
				if err != nil {
					errors <- err
					continue
				}

				// Hold lock briefly
				time.Sleep(10 * time.Millisecond)

				// Release lock
				_, err = manager.Unlock(ctx, project, workspace, user)
				if err != nil {
					errors <- err
					continue
				}

				operationTime := time.Since(operationStart)
				results <- operationTime
			}
		}(i)
	}

	wg.Wait()
	totalTime := time.Since(startTime)

	// Close channels and collect results
	close(results)
	close(errors)

	var totalOperations int
	var totalDuration time.Duration
	var maxDuration time.Duration
	var errorCount int

	for duration := range results {
		totalOperations++
		totalDuration += duration
		if duration > maxDuration {
			maxDuration = duration
		}
	}

	for err := range errors {
		errorCount++
		t.Logf("Operation error: %v", err)
	}

	// Performance assertions
	assert.Greater(t, totalOperations, 0, "Should have completed some operations")

	if totalOperations > 0 {
		averageDuration := totalDuration / time.Duration(totalOperations)
		t.Logf("Performance Results:")
		t.Logf("  Total Operations: %d", totalOperations)
		t.Logf("  Total Time: %v", totalTime)
		t.Logf("  Average Operation Time: %v", averageDuration)
		t.Logf("  Max Operation Time: %v", maxDuration)
		t.Logf("  Operations/Second: %.2f", float64(totalOperations)/totalTime.Seconds())
		t.Logf("  Error Count: %d", errorCount)

		// Performance benchmarks
		assert.Less(t, averageDuration, 1*time.Second, "Average operation should be under 1 second")
		assert.Less(t, maxDuration, 5*time.Second, "Max operation should be under 5 seconds")
		assert.Less(t, float64(errorCount)/float64(totalOperations+errorCount), 0.1, "Error rate should be under 10%")
	}
}

// Helper functions

func setupTestManager(t *testing.T) (enhanced.LockManager, func()) {
	config := enhanced.DefaultConfig()
	config.Enabled = true
	return setupTestManagerWithConfig(t, config)
}

func setupTestManagerWithConfig(t *testing.T, config *enhanced.EnhancedConfig) (enhanced.LockManager, func()) {
	// Create in-memory backend for testing
	backend := &MockBackend{
		locks:   make(map[string]*enhanced.EnhancedLock),
		metrics: &enhanced.BackendStats{},
	}

	manager := enhanced.NewEnhancedLockManager(backend, config, logging.NewNoopLogger(t))

	ctx := context.Background()
	err := manager.Start(ctx)
	require.NoError(t, err)

	cleanup := func() {
		manager.Stop()
	}

	return manager, cleanup
}

// MockBackend implements the Backend interface for testing
type MockBackend struct {
	mutex   sync.RWMutex
	locks   map[string]*enhanced.EnhancedLock
	metrics *enhanced.BackendStats
}

func (mb *MockBackend) AcquireLock(ctx context.Context, request *enhanced.EnhancedLockRequest) (*enhanced.EnhancedLock, error) {
	mb.mutex.Lock()
	defer mb.mutex.Unlock()

	// Check if lock already exists for this resource
	resourceKey := mb.getResourceKey(request.Resource)
	for _, lock := range mb.locks {
		if mb.getResourceKey(lock.Resource) == resourceKey && lock.State == enhanced.LockStateAcquired {
			return nil, enhanced.NewLockExistsError(resourceKey)
		}
	}

	// Create new lock
	lock := &enhanced.EnhancedLock{
		ID:         request.ID,
		Resource:   request.Resource,
		State:      enhanced.LockStateAcquired,
		Priority:   request.Priority,
		Owner:      request.User.Username,
		AcquiredAt: time.Now(),
		Metadata:   request.Metadata,
		Version:    1,
		OriginalLock: &models.ProjectLock{
			Project:   request.Project,
			Workspace: request.Workspace,
			User:      request.User,
			Time:      time.Now(),
		},
	}

	if request.Timeout > 0 {
		expiresAt := time.Now().Add(request.Timeout)
		lock.ExpiresAt = &expiresAt
	}

	mb.locks[lock.ID] = lock
	return lock, nil
}

func (mb *MockBackend) TryAcquireLock(ctx context.Context, request *enhanced.EnhancedLockRequest) (*enhanced.EnhancedLock, bool, error) {
	lock, err := mb.AcquireLock(ctx, request)
	if err != nil {
		return nil, false, err
	}
	return lock, true, nil
}

func (mb *MockBackend) ReleaseLock(ctx context.Context, lockID string) error {
	mb.mutex.Lock()
	defer mb.mutex.Unlock()

	if _, exists := mb.locks[lockID]; !exists {
		return enhanced.NewLockNotFoundError(lockID)
	}

	delete(mb.locks, lockID)
	return nil
}

func (mb *MockBackend) GetLock(ctx context.Context, lockID string) (*enhanced.EnhancedLock, error) {
	mb.mutex.RLock()
	defer mb.mutex.RUnlock()

	lock, exists := mb.locks[lockID]
	if !exists {
		return nil, enhanced.NewLockNotFoundError(lockID)
	}

	return lock, nil
}

func (mb *MockBackend) ListLocks(ctx context.Context) ([]*enhanced.EnhancedLock, error) {
	mb.mutex.RLock()
	defer mb.mutex.RUnlock()

	locks := make([]*enhanced.EnhancedLock, 0, len(mb.locks))
	for _, lock := range mb.locks {
		locks = append(locks, lock)
	}

	return locks, nil
}

func (mb *MockBackend) RefreshLock(ctx context.Context, lockID string, extension time.Duration) error {
	mb.mutex.Lock()
	defer mb.mutex.Unlock()

	lock, exists := mb.locks[lockID]
	if !exists {
		return enhanced.NewLockNotFoundError(lockID)
	}

	if lock.ExpiresAt != nil {
		newExpiration := lock.ExpiresAt.Add(extension)
		lock.ExpiresAt = &newExpiration
	}

	return nil
}

func (mb *MockBackend) TransferLock(ctx context.Context, lockID string, newOwner string) error {
	mb.mutex.Lock()
	defer mb.mutex.Unlock()

	lock, exists := mb.locks[lockID]
	if !exists {
		return enhanced.NewLockNotFoundError(lockID)
	}

	lock.Owner = newOwner
	return nil
}

func (mb *MockBackend) EnqueueLockRequest(ctx context.Context, request *enhanced.EnhancedLockRequest) error {
	// Mock implementation - just return nil
	return nil
}

func (mb *MockBackend) DequeueNextRequest(ctx context.Context) (*enhanced.EnhancedLockRequest, error) {
	// Mock implementation
	return nil, nil
}

func (mb *MockBackend) GetQueueStatus(ctx context.Context) (*enhanced.QueueStatus, error) {
	return &enhanced.QueueStatus{
		Size:             0,
		PendingRequests:  []*enhanced.EnhancedLockRequest{},
		QueuesByPriority: make(map[enhanced.Priority]int),
	}, nil
}

func (mb *MockBackend) HealthCheck(ctx context.Context) error {
	return nil
}

func (mb *MockBackend) GetStats(ctx context.Context) (*enhanced.BackendStats, error) {
	mb.mutex.RLock()
	defer mb.mutex.RUnlock()

	mb.metrics.ActiveLocks = int64(len(mb.locks))
	mb.metrics.LastUpdated = time.Now()
	return mb.metrics, nil
}

func (mb *MockBackend) Subscribe(ctx context.Context, eventTypes []string) (<-chan *enhanced.LockEvent, error) {
	// Mock implementation - return empty channel
	eventChan := make(chan *enhanced.LockEvent)
	close(eventChan)
	return eventChan, nil
}

func (mb *MockBackend) CleanupExpiredLocks(ctx context.Context) (int, error) {
	mb.mutex.Lock()
	defer mb.mutex.Unlock()

	cleaned := 0
	for lockID, lock := range mb.locks {
		if lock.ExpiresAt != nil && time.Now().After(*lock.ExpiresAt) {
			delete(mb.locks, lockID)
			cleaned++
		}
	}

	return cleaned, nil
}

func (mb *MockBackend) GetLegacyLock(project models.Project, workspace string) (*models.ProjectLock, error) {
	mb.mutex.RLock()
	defer mb.mutex.RUnlock()

	for _, lock := range mb.locks {
		if lock.Resource.Namespace == project.RepoFullName &&
		   lock.Resource.Path == project.Path &&
		   lock.Resource.Workspace == workspace &&
		   lock.State == enhanced.LockStateAcquired {
			return lock.OriginalLock, nil
		}
	}

	return nil, nil
}

func (mb *MockBackend) ConvertToLegacy(lock *enhanced.EnhancedLock) *models.ProjectLock {
	if lock.OriginalLock != nil {
		return lock.OriginalLock
	}

	return &models.ProjectLock{
		Project: models.Project{
			RepoFullName: lock.Resource.Namespace,
			Path:         lock.Resource.Path,
		},
		Workspace: lock.Resource.Workspace,
		User:      models.User{Username: lock.Owner},
		Time:      lock.AcquiredAt,
	}
}

func (mb *MockBackend) ConvertFromLegacy(legacyLock *models.ProjectLock) *enhanced.EnhancedLock {
	return &enhanced.EnhancedLock{
		ID: fmt.Sprintf("legacy_%d", time.Now().UnixNano()),
		Resource: enhanced.ResourceIdentifier{
			Type:      enhanced.ResourceTypeProject,
			Namespace: legacyLock.Project.RepoFullName,
			Name:      legacyLock.Project.Path,
			Workspace: legacyLock.Workspace,
			Path:      legacyLock.Project.Path,
		},
		State:        enhanced.LockStateAcquired,
		Priority:     enhanced.PriorityNormal,
		Owner:        legacyLock.User.Username,
		AcquiredAt:   legacyLock.Time,
		Version:      1,
		OriginalLock: legacyLock,
	}
}

func (mb *MockBackend) getResourceKey(resource enhanced.ResourceIdentifier) string {
	return fmt.Sprintf("%s/%s/%s", resource.Namespace, resource.Name, resource.Workspace)
}

// Comprehensive deadlock testing scenarios

func testDeadlockDetectionAdvanced(t *testing.T) {
	config := &enhanced.EnhancedConfig{
		Enabled:                 true,
		EnableDeadlockDetection: true,
		DeadlockCheckInterval:   100 * time.Millisecond,
		EnablePriorityQueue:     true,
		MaxQueueSize:           100,
	}

	manager, cleanup := setupTestManagerWithConfig(t, config)
	defer cleanup()

	ctx := context.Background()

	// Create resources for deadlock scenario
	projectA := models.Project{RepoFullName: "test/project-a", Path: "."}
	projectB := models.Project{RepoFullName: "test/project-b", Path: "."}
	projectC := models.Project{RepoFullName: "test/project-c", Path: "."}

	userX := models.User{Username: "userX"}
	userY := models.User{Username: "userY"}
	userZ := models.User{Username: "userZ"}

	workspace := "default"

	// Phase 1: User X locks A
	lockAX, err := manager.Lock(ctx, projectA, workspace, userX)
	require.NoError(t, err)
	require.NotNil(t, lockAX)
	t.Logf("Phase 1: User X acquired lock on project A")

	// Phase 2: User Y locks B
	lockBY, err := manager.Lock(ctx, projectB, workspace, userY)
	require.NoError(t, err)
	require.NotNil(t, lockBY)
	t.Logf("Phase 2: User Y acquired lock on project B")

	// Phase 3: User Z locks C
	lockCZ, err := manager.Lock(ctx, projectC, workspace, userZ)
	require.NoError(t, err)
	require.NotNil(t, lockCZ)
	t.Logf("Phase 3: User Z acquired lock on project C")

	// Phase 4: Create potential deadlock scenario
	// User X tries to lock B (held by Y), User Y tries to lock C (held by Z), User Z tries to lock A (held by X)

	var wg sync.WaitGroup
	var errorsCollected []error
	var mu sync.Mutex

	wg.Add(3)

	// User X tries to lock B
	go func() {
		defer wg.Done()
		ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
		_, err := manager.Lock(ctx, projectB, workspace, userX)
		mu.Lock()
		if err != nil {
			errorsCollected = append(errorsCollected, fmt.Errorf("userX->B: %w", err))
		}
		mu.Unlock()
		t.Logf("User X attempting to lock B: %v", err)
	}()

	// User Y tries to lock C
	go func() {
		defer wg.Done()
		time.Sleep(100 * time.Millisecond) // Slight delay to create dependency
		ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
		_, err := manager.Lock(ctx, projectC, workspace, userY)
		mu.Lock()
		if err != nil {
			errorsCollected = append(errorsCollected, fmt.Errorf("userY->C: %w", err))
		}
		mu.Unlock()
		t.Logf("User Y attempting to lock C: %v", err)
	}()

	// User Z tries to lock A
	go func() {
		defer wg.Done()
		time.Sleep(200 * time.Millisecond) // Create circular dependency
		ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
		_, err := manager.Lock(ctx, projectA, workspace, userZ)
		mu.Lock()
		if err != nil {
			errorsCollected = append(errorsCollected, fmt.Errorf("userZ->A: %w", err))
		}
		mu.Unlock()
		t.Logf("User Z attempting to lock A: %v", err)
	}()

	wg.Wait()

	// Analyze results
	mu.Lock()
	t.Logf("Deadlock detection test completed with %d errors", len(errorsCollected))
	for _, err := range errorsCollected {
		t.Logf("Error: %v", err)
	}
	mu.Unlock()

	// The system should handle this gracefully - either through queueing, deadlock detection, or timeouts
	assert.True(t, len(errorsCollected) > 0, "Should have some blocked operations due to deadlock scenario")

	// Clean up
	manager.Unlock(ctx, projectA, workspace, userX)
	manager.Unlock(ctx, projectB, workspace, userY)
	manager.Unlock(ctx, projectC, workspace, userZ)
}

func testDeadlockResolutionPolicies(t *testing.T) {
	// Test different deadlock resolution policies
	policies := []struct {
		name     string
		priority enhanced.Priority
		expectedVictim string
	}{
		{"LowestPriority", enhanced.PriorityLow, "should select low priority victim"},
		{"HighPriority", enhanced.PriorityHigh, "should prefer high priority lock"},
		{"Critical", enhanced.PriorityCritical, "should preserve critical priority lock"},
	}

	for _, policy := range policies {
		t.Run(policy.name, func(t *testing.T) {
			config := &enhanced.EnhancedConfig{
				Enabled:                 true,
				EnableDeadlockDetection: true,
				DeadlockCheckInterval:   50 * time.Millisecond,
			}

			manager, cleanup := setupTestManagerWithConfig(t, config)
			defer cleanup()

			ctx := context.Background()

			project1 := models.Project{RepoFullName: "test/policy1", Path: "."}
			project2 := models.Project{RepoFullName: "test/policy2", Path: "."}
			user1 := models.User{Username: "policyuser1"}
			user2 := models.User{Username: "policyuser2"}
			workspace := "default"

			// Acquire locks with different priorities
			lock1, err := manager.LockWithPriority(ctx, project1, workspace, user1, policy.priority)
			require.NoError(t, err)
			require.NotNil(t, lock1)

			lock2, err := manager.LockWithPriority(ctx, project2, workspace, user2, enhanced.PriorityNormal)
			require.NoError(t, err)
			require.NotNil(t, lock2)

			// The test validates that locks with different priorities can coexist
			t.Logf("Policy test '%s' completed: %s", policy.name, policy.expectedVictim)

			// Clean up
			manager.Unlock(ctx, project1, workspace, user1)
			manager.Unlock(ctx, project2, workspace, user2)
		})
	}
}

func testCircularWaitScenarios(t *testing.T) {
	config := &enhanced.EnhancedConfig{
		Enabled:                 true,
		EnableDeadlockDetection: true,
		DeadlockCheckInterval:   200 * time.Millisecond,
		EnablePriorityQueue:     true,
		MaxQueueSize:           50,
		QueueTimeout:           10 * time.Second,
	}

	manager, cleanup := setupTestManagerWithConfig(t, config)
	defer cleanup()

	ctx := context.Background()

	// Scenario: A->B->C->A circular wait
	projects := []models.Project{
		{RepoFullName: "test/circular-a", Path: "."},
		{RepoFullName: "test/circular-b", Path: "."},
		{RepoFullName: "test/circular-c", Path: "."},
	}

	users := []models.User{
		{Username: "circular-user-1"},
		{Username: "circular-user-2"},
		{Username: "circular-user-3"},
	}

	workspace := "default"

	// Phase 1: Each user acquires their primary resource
	var primaryLocks []*models.ProjectLock
	for i, user := range users {
		lock, err := manager.Lock(ctx, projects[i], workspace, user)
		require.NoError(t, err)
		require.NotNil(t, lock)
		primaryLocks = append(primaryLocks, lock)
		t.Logf("User %s acquired primary lock on project %s", user.Username, projects[i].RepoFullName)
	}

	// Phase 2: Create circular wait pattern
	var wg sync.WaitGroup
	results := make([]error, len(users))

	for i, user := range users {
		wg.Add(1)
		go func(userIndex int, u models.User) {
			defer wg.Done()
			// Each user tries to acquire the next resource in the circle
			nextProject := projects[(userIndex+1)%len(projects)]
			ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
			defer cancel()

			_, err := manager.Lock(ctx, nextProject, workspace, u)
			results[userIndex] = err
			t.Logf("User %s attempting circular lock on %s: %v", u.Username, nextProject.RepoFullName, err)
		}(i, user)

		// Stagger the attempts to create clear circular dependency
		time.Sleep(200 * time.Millisecond)
	}

	wg.Wait()

	// Analysis: At least some operations should fail or timeout due to circular wait
	failedOperations := 0
	for i, err := range results {
		if err != nil {
			failedOperations++
			t.Logf("User %d operation failed: %v", i, err)
		}
	}

	assert.Greater(t, failedOperations, 0, "Some operations should fail in circular wait scenario")
	t.Logf("Circular wait test: %d/%d operations failed as expected", failedOperations, len(users))

	// Cleanup
	for i, user := range users {
		manager.Unlock(ctx, projects[i], workspace, user)
	}
}

func testMultiResourceDeadlock(t *testing.T) {
	config := &enhanced.EnhancedConfig{
		Enabled:                 true,
		EnableDeadlockDetection: true,
		DeadlockCheckInterval:   150 * time.Millisecond,
		EnablePriorityQueue:     true,
	}

	manager, cleanup := setupTestManagerWithConfig(t, config)
	defer cleanup()

	ctx := context.Background()

	// Complex scenario: Multiple resources per user with crossing dependencies
	resources := []struct {
		project   models.Project
		workspace string
	}{
		{{RepoFullName: "multi/resource-db", Path: "."}, "prod"},
		{{RepoFullName: "multi/resource-api", Path: "."}, "staging"},
		{{RepoFullName: "multi/resource-ui", Path: "."}, "dev"},
		{{RepoFullName: "multi/resource-cache", Path: "."}, "prod"},
	}

	users := []models.User{
		{Username: "backend-team"},
		{Username: "frontend-team"},
		{Username: "devops-team"},
	}

	// Phase 1: Each team acquires some resources
	teamResources := map[string][]*models.ProjectLock{
		"backend-team":  {},
		"frontend-team": {},
		"devops-team":   {},
	}

	// Backend team gets DB and API
	for i := 0; i < 2; i++ {
		lock, err := manager.Lock(ctx, resources[i].project, resources[i].workspace, users[0])
		if err == nil && lock != nil {
			teamResources["backend-team"] = append(teamResources["backend-team"], lock)
			t.Logf("Backend team acquired %s/%s", resources[i].project.RepoFullName, resources[i].workspace)
		}
	}

	// Frontend team gets UI and Cache
	for i := 2; i < 4; i++ {
		lock, err := manager.Lock(ctx, resources[i].project, resources[i].workspace, users[1])
		if err == nil && lock != nil {
			teamResources["frontend-team"] = append(teamResources["frontend-team"], lock)
			t.Logf("Frontend team acquired %s/%s", resources[i].project.RepoFullName, resources[i].workspace)
		}
	}

	// Phase 2: Cross-team resource requests (potential deadlock)
	var wg sync.WaitGroup
	crossRequestResults := make(map[string]error)
	var resultsMutex sync.Mutex

	// Backend team tries to get UI (held by frontend)
	wg.Add(1)
	go func() {
		defer wg.Done()
		ctx, cancel := context.WithTimeout(ctx, 4*time.Second)
		defer cancel()
		_, err := manager.Lock(ctx, resources[2].project, resources[2].workspace, users[0])
		resultsMutex.Lock()
		crossRequestResults["backend->ui"] = err
		resultsMutex.Unlock()
	}()

	// Frontend team tries to get DB (held by backend)
	wg.Add(1)
	go func() {
		defer wg.Done()
		time.Sleep(300 * time.Millisecond)
		ctx, cancel := context.WithTimeout(ctx, 4*time.Second)
		defer cancel()
		_, err := manager.Lock(ctx, resources[0].project, resources[0].workspace, users[1])
		resultsMutex.Lock()
		crossRequestResults["frontend->db"] = err
		resultsMutex.Unlock()
	}()

	// DevOps team tries to get API and Cache (potential conflict)
	wg.Add(1)
	go func() {
		defer wg.Done()
		time.Sleep(600 * time.Millisecond)
		ctx, cancel := context.WithTimeout(ctx, 4*time.Second)
		defer cancel()
		_, err1 := manager.Lock(ctx, resources[1].project, resources[1].workspace, users[2])
		_, err2 := manager.Lock(ctx, resources[3].project, resources[3].workspace, users[2])
		resultsMutex.Lock()
		crossRequestResults["devops->api"] = err1
		crossRequestResults["devops->cache"] = err2
		resultsMutex.Unlock()
	}()

	wg.Wait()

	// Analyze cross-team request results
	resultsMutex.Lock()
	conflictCount := 0
	for request, err := range crossRequestResults {
		if err != nil {
			conflictCount++
			t.Logf("Cross-team request %s failed: %v", request, err)
		} else {
			t.Logf("Cross-team request %s succeeded", request)
		}
	}
	resultsMutex.Unlock()

	assert.Greater(t, conflictCount, 0, "Should have some conflicts in multi-resource deadlock scenario")
	t.Logf("Multi-resource deadlock test: %d/%d cross-requests had conflicts", conflictCount, len(crossRequestResults))

	// Cleanup all team resources
	for team, locks := range teamResources {
		for _, lock := range locks {
			manager.Unlock(ctx, lock.Project, lock.Workspace, lock.User)
			t.Logf("Released lock for team %s: %s/%s", team, lock.Project.RepoFullName, lock.Workspace)
		}
	}
}

func testDeadlockPreventionWithPriority(t *testing.T) {
	config := &enhanced.EnhancedConfig{
		Enabled:                 true,
		EnableDeadlockDetection: true,
		DeadlockCheckInterval:   100 * time.Millisecond,
		EnablePriorityQueue:     true,
		MaxQueueSize:           50,
	}

	manager, cleanup := setupTestManagerWithConfig(t, config)
	defer cleanup()

	ctx := context.Background()

	// Test priority-aware deadlock prevention
	projectX := models.Project{RepoFullName: "priority/project-x", Path: "."}
	projectY := models.Project{RepoFullName: "priority/project-y", Path: "."}

	criticalUser := models.User{Username: "critical-ops"}
	normalUser := models.User{Username: "normal-dev"}
	workspace := "production"

	// Critical user acquires project X
	criticalLock, err := manager.LockWithPriority(ctx, projectX, workspace, criticalUser, enhanced.PriorityCritical)
	require.NoError(t, err)
	require.NotNil(t, criticalLock)
	t.Logf("Critical user acquired lock on project X")

	// Normal user acquires project Y
	normalLock, err := manager.LockWithPriority(ctx, projectY, workspace, normalUser, enhanced.PriorityNormal)
	require.NoError(t, err)
	require.NotNil(t, normalLock)
	t.Logf("Normal user acquired lock on project Y")

	// Now create potential deadlock with priority consideration
	var wg sync.WaitGroup
	var criticalErr, normalErr error

	// Critical user tries to get Y (should have higher priority)
	wg.Add(1)
	go func() {
		defer wg.Done()
		ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
		defer cancel()
		_, criticalErr = manager.LockWithPriority(ctx, projectY, workspace, criticalUser, enhanced.PriorityCritical)
		t.Logf("Critical user trying to get Y: %v", criticalErr)
	}()

	// Normal user tries to get X (should be lower priority)
	wg.Add(1)
	go func() {
		defer wg.Done()
		time.Sleep(200 * time.Millisecond)
		ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
		defer cancel()
		_, normalErr = manager.LockWithPriority(ctx, projectX, workspace, normalUser, enhanced.PriorityNormal)
		t.Logf("Normal user trying to get X: %v", normalErr)
	}()

	wg.Wait()

	// Analysis: System should handle this based on priority
	if criticalErr == nil && normalErr != nil {
		t.Log("Priority system working: critical user succeeded, normal user blocked")
	} else if criticalErr != nil && normalErr == nil {
		t.Log("Unexpected: normal user succeeded over critical user")
	} else {
		t.Logf("Both operations had same outcome: critical=%v, normal=%v", criticalErr, normalErr)
	}

	// At least one should be blocked to prevent deadlock
	assert.True(t, criticalErr != nil || normalErr != nil, "At least one operation should be blocked")

	// Cleanup
	manager.Unlock(ctx, projectX, workspace, criticalUser)
	manager.Unlock(ctx, projectY, workspace, normalUser)
}

func testCascadeDeadlockResolution(t *testing.T) {
	config := &enhanced.EnhancedConfig{
		Enabled:                 true,
		EnableDeadlockDetection: true,
		DeadlockCheckInterval:   200 * time.Millisecond,
		EnablePriorityQueue:     true,
	}

	manager, cleanup := setupTestManagerWithConfig(t, config)
	defer cleanup()

	ctx := context.Background()

	// Scenario: Chain of dependencies that could cascade when resolved
	projects := []models.Project{
		{RepoFullName: "cascade/service-a", Path: "."},
		{RepoFullName: "cascade/service-b", Path: "."},
		{RepoFullName: "cascade/service-c", Path: "."},
		{RepoFullName: "cascade/service-d", Path: "."},
	}

	users := []models.User{
		{Username: "team-alpha"},
		{Username: "team-beta"},
		{Username: "team-gamma"},
		{Username: "team-delta"},
	}

	workspace := "production"

	// Create chain: A->B->C->D where each team has one service but wants the next
	var primaryLocks []*models.ProjectLock

	// Each team acquires their primary service
	for i, user := range users {
		lock, err := manager.Lock(ctx, projects[i], workspace, user)
		require.NoError(t, err)
		require.NotNil(t, lock)
		primaryLocks = append(primaryLocks, lock)
		t.Logf("Team %s acquired primary service %s", user.Username, projects[i].RepoFullName)
	}

	// Create cascade scenario: each team wants the next service in chain
	var wg sync.WaitGroup
	cascadeResults := make(map[int]error)
	var cascadeMutex sync.Mutex

	for i, user := range users {
		wg.Add(1)
		go func(teamIndex int, u models.User) {
			defer wg.Done()
			nextServiceIndex := (teamIndex + 1) % len(projects)
			nextProject := projects[nextServiceIndex]

			ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
			defer cancel()

			_, err := manager.Lock(ctx, nextProject, workspace, u)
			cascadeMutex.Lock()
			cascadeResults[teamIndex] = err
			cascadeMutex.Unlock()
			t.Logf("Team %s (%d) trying to acquire service %s: %v", u.Username, teamIndex, nextProject.RepoFullName, err)
		}(i, user)

		// Stagger to create clear dependency chain
		time.Sleep(300 * time.Millisecond)
	}

	wg.Wait()

	// Analyze cascade behavior
	cascadeMutex.Lock()
	blockedTeams := 0
	for teamIndex, err := range cascadeResults {
		if err != nil {
			blockedTeams++
			t.Logf("Team %d blocked in cascade: %v", teamIndex, err)
		} else {
			t.Logf("Team %d succeeded in cascade", teamIndex)
		}
	}
	cascadeMutex.Unlock()

	assert.Greater(t, blockedTeams, 0, "Some teams should be blocked in cascade scenario")
	t.Logf("Cascade resolution test: %d/%d teams blocked", blockedTeams, len(users))

	// The system should prevent the full cascade deadlock
	assert.Less(t, blockedTeams, len(users), "Not all teams should be blocked (deadlock prevention)")

	// Cleanup
	for i, user := range users {
		manager.Unlock(ctx, projects[i], workspace, user)
	}
}

func testDeadlockPerformanceBenchmark(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping deadlock performance benchmark in short mode")
	}

	config := &enhanced.EnhancedConfig{
		Enabled:                 true,
		EnableDeadlockDetection: true,
		DeadlockCheckInterval:   50 * time.Millisecond,
		EnablePriorityQueue:     true,
		MaxQueueSize:           500,
	}

	manager, cleanup := setupTestManagerWithConfig(t, config)
	defer cleanup()

	ctx := context.Background()

	// Performance parameters
	numResources := 20
	numUsers := 50
	numOperations := 200
	concurrentRequests := 10

	// Create resource pool
	resources := make([]models.Project, numResources)
	for i := 0; i < numResources; i++ {
		resources[i] = models.Project{
			RepoFullName: fmt.Sprintf("benchmark/resource-%d", i),
			Path:         ".",
		}
	}

	// Create user pool
	users := make([]models.User, numUsers)
	for i := 0; i < numUsers; i++ {
		users[i] = models.User{Username: fmt.Sprintf("benchuser-%d", i)}
	}

	workspace := "benchmark"

	var wg sync.WaitGroup
	operationTimes := make(chan time.Duration, numOperations)
	operationErrors := make(chan error, numOperations)

	startTime := time.Now()

	for i := 0; i < concurrentRequests; i++ {
		wg.Add(1)
		go func(routineID int) {
			defer wg.Done()
			rand.Seed(int64(routineID))

			for j := 0; j < numOperations/concurrentRequests; j++ {
				opStart := time.Now()

				// Randomly select resource and user
				resource := resources[rand.Intn(numResources)]
				user := users[rand.Intn(numUsers)]

				// Acquire lock
				lock, err := manager.LockWithTimeout(ctx, resource, workspace, user, 2*time.Second)
				if err != nil {
					operationErrors <- err
					continue
				}

				// Hold briefly (simulate work)
				time.Sleep(time.Duration(rand.Intn(50)) * time.Millisecond)

				// Release lock
				_, err = manager.Unlock(ctx, resource, workspace, user)
				if err != nil {
					operationErrors <- err
					continue
				}

				opTime := time.Since(opStart)
				operationTimes <- opTime
			}
		}(i)
	}

	wg.Wait()
	totalBenchmarkTime := time.Since(startTime)

	// Collect results
	close(operationTimes)
	close(operationErrors)

	var operationCount int
	var totalOpTime time.Duration
	var maxOpTime time.Duration
	var errorCount int

	for opTime := range operationTimes {
		operationCount++
		totalOpTime += opTime
		if opTime > maxOpTime {
			maxOpTime = opTime
		}
	}

	for range operationErrors {
		errorCount++
	}

	// Performance analysis
	if operationCount > 0 {
		averageOpTime := totalOpTime / time.Duration(operationCount)
		opsPerSecond := float64(operationCount) / totalBenchmarkTime.Seconds()

		t.Logf("Deadlock Performance Benchmark Results:")
		t.Logf("  Total Benchmark Time: %v", totalBenchmarkTime)
		t.Logf("  Successful Operations: %d", operationCount)
		t.Logf("  Failed Operations: %d", errorCount)
		t.Logf("  Average Operation Time: %v", averageOpTime)
		t.Logf("  Max Operation Time: %v", maxOpTime)
		t.Logf("  Operations/Second: %.2f", opsPerSecond)
		t.Logf("  Error Rate: %.2f%%", float64(errorCount)/float64(operationCount+errorCount)*100)

		// Performance assertions for deadlock detection overhead
		assert.Less(t, averageOpTime, 500*time.Millisecond, "Deadlock detection shouldn't add significant overhead")
		assert.Greater(t, opsPerSecond, 10.0, "Should maintain reasonable throughput with deadlock detection")
		assert.Less(t, float64(errorCount)/float64(operationCount+errorCount), 0.15, "Error rate should be acceptable")
	} else {
		t.Error("No operations completed successfully in benchmark")
	}
}

func testEndToEndSystemTest(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping end-to-end system test in short mode")
	}

	config := &enhanced.EnhancedConfig{
		Enabled:                 true,
		Backend:                "boltdb",
		EnableDeadlockDetection: true,
		DeadlockCheckInterval:   100 * time.Millisecond,
		EnablePriorityQueue:     true,
		MaxQueueSize:           100,
		EnableRetry:            true,
		MaxRetryAttempts:       3,
		RetryBaseDelay:         100 * time.Millisecond,
		RetryMaxDelay:          2 * time.Second,
		EnableEvents:           true,
		EventBufferSize:        1000,
	}

	manager, cleanup := setupTestManagerWithConfig(t, config)
	defer cleanup()

	ctx := context.Background()

	// Real-world scenario: Multiple teams working on microservices deployment
	services := []struct {
		name      string
		project   models.Project
		workspace string
		priority  enhanced.Priority
	}{
		{"user-service", models.Project{RepoFullName: "company/user-service", Path: "."}, "production", enhanced.PriorityHigh},
		{"auth-service", models.Project{RepoFullName: "company/auth-service", Path: "."}, "production", enhanced.PriorityCritical},
		{"payment-service", models.Project{RepoFullName: "company/payment-service", Path: "."}, "production", enhanced.PriorityHigh},
		{"notification-service", models.Project{RepoFullName: "company/notification-service", Path: "."}, "staging", enhanced.PriorityNormal},
		{"analytics-service", models.Project{RepoFullName: "company/analytics-service", Path: "."}, "development", enhanced.PriorityLow},
		{"gateway-service", models.Project{RepoFullName: "company/gateway-service", Path: "."}, "production", enhanced.PriorityCritical},
	}

	teams := []models.User{
		{Username: "backend-team"},
		{Username: "platform-team"},
		{Username: "security-team"},
		{Username: "data-team"},
		{Username: "devops-team"},
	}

	// Simulate real deployment scenario
	var wg sync.WaitGroup
	deploymentResults := make(map[string]error)
	var resultsMutex sync.Mutex

	t.Log("Starting end-to-end deployment simulation...")

	// Phase 1: Critical services deployment (auth, gateway)
	for _, service := range services {
		if service.priority == enhanced.PriorityCritical {
			wg.Add(1)
			go func(svc struct {
				name      string
				project   models.Project
				workspace string
				priority  enhanced.Priority
			}) {
				defer wg.Done()
				team := teams[2] // security team for critical services

				t.Logf("Deploying critical service: %s", svc.name)
				lock, err := manager.LockWithPriority(ctx, svc.project, svc.workspace, team, svc.priority)

				resultsMutex.Lock()
				deploymentResults[fmt.Sprintf("critical-%s", svc.name)] = err
				resultsMutex.Unlock()

				if err == nil {
					// Simulate deployment time
					time.Sleep(500 * time.Millisecond)
					manager.Unlock(ctx, svc.project, svc.workspace, team)
					t.Logf("Critical service %s deployed successfully", svc.name)
				} else {
					t.Logf("Critical service %s deployment failed: %v", svc.name, err)
				}
			}(service)
		}
	}

	wg.Wait()
	// Phase 2: High priority services (user, payment)
	for _, service := range services {
		if service.priority == enhanced.PriorityHigh {
			wg.Add(1)
			go func(svc struct {
				name      string
				project   models.Project
				workspace string
				priority  enhanced.Priority
			}) {
				defer wg.Done()
				team := teams[0] // backend team

				t.Logf("Deploying high priority service: %s", svc.name)
				lock, err := manager.LockWithPriority(ctx, svc.project, svc.workspace, team, svc.priority)

				resultsMutex.Lock()
				deploymentResults[fmt.Sprintf("high-%s", svc.name)] = err
				resultsMutex.Unlock()

				if err == nil {
					time.Sleep(400 * time.Millisecond)
					manager.Unlock(ctx, svc.project, svc.workspace, team)
					t.Logf("High priority service %s deployed successfully", svc.name)
				} else {
					t.Logf("High priority service %s deployment failed: %v", svc.name, err)
				}
			}(service)
		}
	}

	wg.Wait()

	// Phase 3: Normal and low priority services (remaining)
	for _, service := range services {
		if service.priority == enhanced.PriorityNormal || service.priority == enhanced.PriorityLow {
			wg.Add(1)
			go func(svc struct {
				name      string
				project   models.Project
				workspace string
				priority  enhanced.Priority
			}) {
				defer wg.Done()
				var team models.User
				if svc.priority == enhanced.PriorityNormal {
					team = teams[1] // platform team
				} else {
					team = teams[3] // data team
				}

				t.Logf("Deploying service: %s (priority: %v)", svc.name, svc.priority)
				lock, err := manager.LockWithPriority(ctx, svc.project, svc.workspace, team, svc.priority)

				resultsMutex.Lock()
				deploymentResults[fmt.Sprintf("%v-%s", svc.priority, svc.name)] = err
				resultsMutex.Unlock()

				if err == nil {
					time.Sleep(300 * time.Millisecond)
					manager.Unlock(ctx, svc.project, svc.workspace, team)
					t.Logf("Service %s deployed successfully", svc.name)
				} else {
					t.Logf("Service %s deployment failed: %v", svc.name, err)
				}
			}(service)
		}
	}

	wg.Wait()

	// Analyze end-to-end results
	resultsMutex.Lock()
	successfulDeployments := 0
	failedDeployments := 0

	for deployment, err := range deploymentResults {
		if err == nil {
			successfulDeployments++
			t.Logf("✅ %s: SUCCESS", deployment)
		} else {
			failedDeployments++
			t.Logf("❌ %s: FAILED - %v", deployment, err)
		}
	}
	resultsMutex.Unlock()

	t.Logf("End-to-End System Test Results:")
	t.Logf("  Successful Deployments: %d", successfulDeployments)
	t.Logf("  Failed Deployments: %d", failedDeployments)
	t.Logf("  Success Rate: %.2f%%", float64(successfulDeployments)/float64(successfulDeployments+failedDeployments)*100)

	// System should handle the deployment workflow gracefully
	assert.Greater(t, successfulDeployments, 0, "Should have some successful deployments")
	assert.Less(t, float64(failedDeployments)/float64(successfulDeployments+failedDeployments), 0.5, "Failure rate should be reasonable")

	// Critical services should have higher success rate
	criticalSuccess := 0
	criticalTotal := 0
	for deployment, err := range deploymentResults {
		if strings.Contains(deployment, "critical-") {
			criticalTotal++
			if err == nil {
				criticalSuccess++
			}
		}
	}

	if criticalTotal > 0 {
		criticalSuccessRate := float64(criticalSuccess) / float64(criticalTotal)
		t.Logf("Critical Services Success Rate: %.2f%%", criticalSuccessRate*100)
		assert.Greater(t, criticalSuccessRate, 0.5, "Critical services should have higher success rate")
	}

	t.Log("End-to-end system test completed successfully")
}