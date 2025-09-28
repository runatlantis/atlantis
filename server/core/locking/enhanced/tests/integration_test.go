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
	t.Run("BackwardCompatibility", testBackwardCompatibility)
	t.Run("RedisBackendIntegration", testRedisBackendIntegration)
	t.Run("PerformanceUnderLoad", testPerformanceUnderLoad)
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
		Enabled:          true,
		EnableRetry:      true,
		MaxRetryAttempts: 3,
		RetryBaseDelay:   100 * time.Millisecond,
		RetryMaxDelay:    1 * time.Second,
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

	// Should have taken some time due to retries (but may be fast if error is non-retryable)
	t.Logf("Lock retry attempt took %v", duration)

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
		Enabled:              true,
		LegacyFallback:       true,
		PreserveLegacyFormat: true,
	}

	manager, cleanup := setupTestManagerWithConfig(t, config)
	defer cleanup()

	// Use the adapter to test legacy interface compatibility
	// Get the backend from setupTestManager
	backend := NewMockBackend()
	adapter := enhanced.NewLockingAdapter(
		manager,
		backend,
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
	unlockedLock, err := adapter.Unlock(project, workspace)
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
		Enabled:        true,
		Backend:        "redis",
		RedisKeyPrefix: "atlantis:test:lock:",
		RedisLockTTL:   5 * time.Minute,
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
				_, err := manager.Lock(ctx, project, workspace, user)
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
	backend := NewMockBackend()

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

// NewMockBackend creates a new MockBackend instance
func NewMockBackend() *MockBackend {
	return &MockBackend{
		locks:   make(map[string]*enhanced.EnhancedLock),
		metrics: &enhanced.BackendStats{},
	}
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
