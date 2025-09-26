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

	"github.com/runatlantis/atlantis/server/core/locking"
	"github.com/runatlantis/atlantis/server/core/locking/enhanced"
	"github.com/runatlantis/atlantis/server/core/locking/enhanced/backends"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/logging"
)

// BackwardCompatibilityTestSuite ensures enhanced locking maintains full compatibility with legacy interfaces
type BackwardCompatibilityTestSuite struct {
	legacyBackend    locking.Backend
	enhancedManager  enhanced.LockManager
	adapter          *enhanced.LockingAdapter
	redisClient      redis.UniversalClient
	config           *enhanced.EnhancedConfig
	logger           logging.SimpleLogging
	cleanup          func()
}

// SetupCompatibilityTestSuite initializes the test environment
func SetupCompatibilityTestSuite(t *testing.T) *BackwardCompatibilityTestSuite {
	// Setup Redis client (use test database)
	redisClient := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
		DB:   15, // Use dedicated test database
	})

	// Test Redis connection
	ctx := context.Background()
	if err := redisClient.Ping(ctx).Err(); err != nil {
		t.Skip("Redis not available for compatibility testing")
	}

	// Clean test database
	redisClient.FlushDB(ctx)

	config := &enhanced.EnhancedConfig{
		Enabled:                 true,
		Backend:                 "redis",
		DefaultTimeout:          30 * time.Minute,
		MaxTimeout:              2 * time.Hour,
		EnablePriorityQueue:     true,
		MaxQueueSize:           100,
		QueueTimeout:           10 * time.Minute,
		EnableRetry:            true,
		MaxRetryAttempts:       3,
		RetryBaseDelay:         time.Second,
		RetryMaxDelay:          30 * time.Second,
		EnableDeadlockDetection: true,
		DeadlockCheckInterval:   5 * time.Second,
		EnableEvents:           true,
		EventBufferSize:        100,
		RedisClusterMode:       false,
		RedisKeyPrefix:         "atlantis:test:lock:",
		RedisLockTTL:           time.Hour,
		LegacyFallback:         true,
		PreserveLegacyFormat:   true,
	}

	logger := logging.NewNoopLogger(t)

	// Create enhanced backend
	redisBackend := backends.NewRedisBackend(redisClient, config, logger)
	enhancedManager := enhanced.NewEnhancedLockManager(redisBackend, config, logger)

	// Create legacy backend (mock for testing)
	legacyBackend := NewMockLegacyBackend()

	// Create adapter
	adapter := enhanced.NewLockingAdapter(enhancedManager, redisBackend, config, legacyBackend, logger)

	// Start manager
	require.NoError(t, enhancedManager.Start(ctx))

	cleanup := func() {
		enhancedManager.Stop()
		redisClient.FlushDB(ctx)
		redisClient.Close()
	}

	return &BackwardCompatibilityTestSuite{
		legacyBackend:   legacyBackend,
		enhancedManager: enhancedManager,
		adapter:         adapter,
		redisClient:     redisClient,
		config:          config,
		logger:          logger,
		cleanup:         cleanup,
	}
}

// TestFullBackwardCompatibility runs comprehensive backward compatibility tests
func TestFullBackwardCompatibility(t *testing.T) {
	suite := SetupCompatibilityTestSuite(t)
	defer suite.cleanup()

	t.Run("InterfaceCompatibility", suite.TestInterfaceCompatibility)
	t.Run("DataFormatCompatibility", suite.TestDataFormatCompatibility)
	t.Run("BehaviorCompatibility", suite.TestBehaviorCompatibility)
	t.Run("ErrorHandlingCompatibility", suite.TestErrorHandlingCompatibility)
	t.Run("ConcurrencyCompatibility", suite.TestConcurrencyCompatibility)
	t.Run("PerformanceCompatibility", suite.TestPerformanceCompatibility)
	t.Run("FallbackCompatibility", suite.TestFallbackCompatibility)
}

// TestInterfaceCompatibility verifies all legacy interface methods work correctly
func (s *BackwardCompatibilityTestSuite) TestInterfaceCompatibility(t *testing.T) {
	// Test data
	project := models.Project{
		RepoFullName: "test/compat-interface",
		Path:         ".",
	}
	user := models.User{Username: "testuser"}
	workspace := "default"
	pull := models.PullRequest{
		Num:  1,
		Author: user.Username,
	}

	lock := models.ProjectLock{
		Project:   project,
		Workspace: workspace,
		User:      user,
		Pull:      pull,
		Time:      time.Now(),
	}

	ctx := context.Background()

	// Test TryLock method
	t.Run("TryLock", func(t *testing.T) {
		acquired, resp, err := s.adapter.TryLock(lock)
		require.NoError(t, err)
		assert.True(t, acquired)
		assert.Empty(t, resp.LockFailureReason)

		// Try to acquire same lock again
		acquired2, resp2, err2 := s.adapter.TryLock(lock)
		require.NoError(t, err2)
		assert.False(t, acquired2)
		assert.NotEmpty(t, resp2.LockFailureReason)
		assert.Equal(t, project.RepoFullName, resp2.CurrLock.Project.RepoFullName)
	})

	// Test List method
	t.Run("List", func(t *testing.T) {
		locks, err := s.adapter.List()
		require.NoError(t, err)
		assert.Len(t, locks, 1)

		key := fmt.Sprintf("%s/%s/%s", project.RepoFullName, project.Path, workspace)
		foundLock, exists := locks[key]
		assert.True(t, exists)
		assert.Equal(t, project.RepoFullName, foundLock.Project.RepoFullName)
		assert.Equal(t, workspace, foundLock.Workspace)
		assert.Equal(t, user.Username, foundLock.User.Username)
	})

	// Test GetLock method
	t.Run("GetLock", func(t *testing.T) {
		retrievedLock, err := s.adapter.GetLock(project, workspace)
		require.NoError(t, err)
		require.NotNil(t, retrievedLock)
		assert.Equal(t, project.RepoFullName, retrievedLock.Project.RepoFullName)
		assert.Equal(t, workspace, retrievedLock.Workspace)
		assert.Equal(t, user.Username, retrievedLock.User.Username)
	})

	// Test Unlock method
	t.Run("Unlock", func(t *testing.T) {
		unlockedLock, err := s.adapter.Unlock(project, workspace, user)
		require.NoError(t, err)
		require.NotNil(t, unlockedLock)
		assert.Equal(t, project.RepoFullName, unlockedLock.Project.RepoFullName)

		// Verify lock is gone
		finalLock, err := s.adapter.GetLock(project, workspace)
		require.NoError(t, err)
		assert.Nil(t, finalLock)

		// Verify list is empty
		locks, err := s.adapter.List()
		require.NoError(t, err)
		assert.Empty(t, locks)
	})

	// Test UnlockByPull method
	t.Run("UnlockByPull", func(t *testing.T) {
		// Create multiple locks for the same repository
		workspaces := []string{"ws1", "ws2", "ws3"}
		for _, ws := range workspaces {
			lockForWs := lock
			lockForWs.Workspace = ws
			acquired, _, err := s.adapter.TryLock(lockForWs)
			require.NoError(t, err)
			require.True(t, acquired)
		}

		// Unlock by pull request
		unlockedLocks, err := s.adapter.UnlockByPull(project.RepoFullName, pull.Num)
		require.NoError(t, err)
		assert.Len(t, unlockedLocks, len(workspaces))

		// Verify all locks are gone
		for _, ws := range workspaces {
			finalLock, err := s.adapter.GetLock(project, ws)
			require.NoError(t, err)
			assert.Nil(t, finalLock)
		}
	})
}

// TestDataFormatCompatibility ensures data formats remain compatible
func (s *BackwardCompatibilityTestSuite) TestDataFormatCompatibility(t *testing.T) {
	// Create legacy lock data
	legacyLock := &models.ProjectLock{
		Project: models.Project{
			RepoFullName: "test/data-compat",
			Path:         "terraform/",
		},
		Workspace: "production",
		User:      models.User{Username: "engineer"},
		Pull: models.PullRequest{
			Num:    42,
			Author: "engineer",
		},
		Time: time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
	}

	// Convert to enhanced format
	enhancedLock := s.adapter.(*enhanced.LockingAdapter).ConvertFromLegacy(legacyLock)

	// Verify conversion
	assert.Equal(t, legacyLock.Project.RepoFullName, enhancedLock.Resource.Namespace)
	assert.Equal(t, legacyLock.Project.Path, enhancedLock.Resource.Path)
	assert.Equal(t, legacyLock.Workspace, enhancedLock.Resource.Workspace)
	assert.Equal(t, legacyLock.User.Username, enhancedLock.Owner)
	assert.Equal(t, legacyLock.Time, enhancedLock.AcquiredAt)

	// Convert back to legacy format
	reconvertedLock := s.adapter.(*enhanced.LockingAdapter).ConvertToLegacy(enhancedLock)

	// Verify round-trip conversion preserves data
	assert.Equal(t, legacyLock.Project.RepoFullName, reconvertedLock.Project.RepoFullName)
	assert.Equal(t, legacyLock.Project.Path, reconvertedLock.Project.Path)
	assert.Equal(t, legacyLock.Workspace, reconvertedLock.Workspace)
	assert.Equal(t, legacyLock.User.Username, reconvertedLock.User.Username)
	assert.Equal(t, legacyLock.Time.Unix(), reconvertedLock.Time.Unix()) // Compare timestamps
}

// TestBehaviorCompatibility ensures enhanced system behaves like legacy system
func (s *BackwardCompatibilityTestSuite) TestBehaviorCompatibility(t *testing.T) {
	project := models.Project{RepoFullName: "test/behavior", Path: "."}
	user1 := models.User{Username: "user1"}
	user2 := models.User{Username: "user2"}
	workspace := "default"
	pull := models.PullRequest{Num: 1}

	lock1 := models.ProjectLock{Project: project, Workspace: workspace, User: user1, Pull: pull, Time: time.Now()}
	lock2 := models.ProjectLock{Project: project, Workspace: workspace, User: user2, Pull: pull, Time: time.Now()}

	// Test exclusive locking behavior
	t.Run("ExclusiveLocking", func(t *testing.T) {
		// First user acquires lock
		acquired1, _, err := s.adapter.TryLock(lock1)
		require.NoError(t, err)
		assert.True(t, acquired1)

		// Second user cannot acquire same lock
		acquired2, resp, err := s.adapter.TryLock(lock2)
		require.NoError(t, err)
		assert.False(t, acquired2)
		assert.NotEmpty(t, resp.LockFailureReason)
		assert.Equal(t, user1.Username, resp.CurrLock.User.Username)

		// First user releases lock
		_, err = s.adapter.Unlock(project, workspace, user1)
		require.NoError(t, err)

		// Second user can now acquire lock
		acquired3, _, err := s.adapter.TryLock(lock2)
		require.NoError(t, err)
		assert.True(t, acquired3)

		// Cleanup
		s.adapter.Unlock(project, workspace, user2)
	})

	// Test workspace isolation
	t.Run("WorkspaceIsolation", func(t *testing.T) {
		lock1.Workspace = "staging"
		lock2.Workspace = "production"

		// Both users can acquire locks in different workspaces
		acquired1, _, err := s.adapter.TryLock(lock1)
		require.NoError(t, err)
		assert.True(t, acquired1)

		acquired2, _, err := s.adapter.TryLock(lock2)
		require.NoError(t, err)
		assert.True(t, acquired2)

		// Cleanup
		s.adapter.Unlock(project, "staging", user1)
		s.adapter.Unlock(project, "production", user2)
	})

	// Test unlocking non-existent locks
	t.Run("UnlockNonExistent", func(t *testing.T) {
		nonExistentProject := models.Project{RepoFullName: "test/nonexistent", Path: "."}

		// Should not error when unlocking non-existent lock
		unlockedLock, err := s.adapter.Unlock(nonExistentProject, "default", user1)
		require.NoError(t, err)
		assert.Nil(t, unlockedLock) // Should return nil for non-existent locks
	})
}

// TestErrorHandlingCompatibility ensures error handling matches legacy system
func (s *BackwardCompatibilityTestSuite) TestErrorHandlingCompatibility(t *testing.T) {
	project := models.Project{RepoFullName: "test/errors", Path: "."}
	user := models.User{Username: "testuser"}
	workspace := "default"
	lock := models.ProjectLock{Project: project, Workspace: workspace, User: user, Time: time.Now()}

	// Test error conditions that should behave like legacy system
	t.Run("DuplicateLockError", func(t *testing.T) {
		// Acquire lock first
		acquired, _, err := s.adapter.TryLock(lock)
		require.NoError(t, err)
		require.True(t, acquired)

		// Try to acquire again - should fail gracefully
		acquired2, resp, err := s.adapter.TryLock(lock)
		require.NoError(t, err) // Should not return error, just indication of failure
		assert.False(t, acquired2)
		assert.NotEmpty(t, resp.LockFailureReason)
		assert.Equal(t, user.Username, resp.CurrLock.User.Username)

		// Cleanup
		s.adapter.Unlock(project, workspace, user)
	})

	// Test error handling during Redis failures (fallback behavior)
	t.Run("BackendFailureHandling", func(t *testing.T) {
		if !s.config.LegacyFallback {
			t.Skip("Legacy fallback not enabled")
		}

		// Simulate Redis failure by using invalid configuration
		originalConfig := s.config.RedisKeyPrefix
		s.config.RedisKeyPrefix = "" // Invalid prefix to cause errors

		// Operations should fall back to legacy backend
		acquired, _, err := s.adapter.TryLock(lock)

		// Restore configuration
		s.config.RedisKeyPrefix = originalConfig

		// Should either succeed via fallback or fail gracefully
		if err != nil {
			assert.Contains(t, err.Error(), "fallback") // Error should mention fallback
		} else {
			assert.True(t, acquired) // Or succeed via fallback
		}
	})
}

// TestConcurrencyCompatibility ensures concurrent behavior matches expectations
func (s *BackwardCompatibilityTestSuite) TestConcurrencyCompatibility(t *testing.T) {
	numConcurrentClients := 20
	numOperationsPerClient := 10
	project := models.Project{RepoFullName: "test/concurrency", Path: "."}
	workspace := "default"

	var wg sync.WaitGroup
	results := make(chan bool, numConcurrentClients*numOperationsPerClient)
	errors := make(chan error, numConcurrentClients*numOperationsPerClient)

	// Create concurrent lock attempts
	for i := 0; i < numConcurrentClients; i++ {
		wg.Add(1)
		go func(clientID int) {
			defer wg.Done()
			user := models.User{Username: fmt.Sprintf("user%d", clientID)}

			for j := 0; j < numOperationsPerClient; j++ {
				lock := models.ProjectLock{
					Project:   project,
					Workspace: workspace,
					User:      user,
					Time:      time.Now(),
				}

				// Try to acquire lock
				acquired, _, err := s.adapter.TryLock(lock)
				if err != nil {
					errors <- err
					continue
				}

				if acquired {
					// Hold lock briefly
					time.Sleep(10 * time.Millisecond)

					// Release lock
					_, err := s.adapter.Unlock(project, workspace, user)
					if err != nil {
						errors <- err
						continue
					}
				}

				results <- acquired
			}
		}(i)
	}

	wg.Wait()
	close(results)
	close(errors)

	// Analyze results
	successCount := 0
	for acquired := range results {
		if acquired {
			successCount++
		}
	}

	errorCount := 0
	for err := range errors {
		t.Logf("Concurrent operation error: %v", err)
		errorCount++
	}

	// Should have reasonable success rate and low error rate
	totalOperations := numConcurrentClients * numOperationsPerClient
	successRate := float64(successCount) / float64(totalOperations-errorCount)
	errorRate := float64(errorCount) / float64(totalOperations)

	assert.Greater(t, successRate, 0.1, "Should have reasonable success rate")
	assert.Less(t, errorRate, 0.1, "Should have low error rate")

	t.Logf("Concurrent test results: %d successes, %d errors, %.2f%% success rate, %.2f%% error rate",
		successCount, errorCount, successRate*100, errorRate*100)
}

// TestPerformanceCompatibility ensures performance is acceptable
func (s *BackwardCompatibilityTestSuite) TestPerformanceCompatibility(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	project := models.Project{RepoFullName: "test/performance", Path: "."}
	user := models.User{Username: "perfuser"}
	workspace := "default"

	numOperations := 100
	operationTimes := make([]time.Duration, 0, numOperations)

	for i := 0; i < numOperations; i++ {
		lock := models.ProjectLock{
			Project:   project,
			Workspace: fmt.Sprintf("ws%d", i), // Different workspace each time
			User:      user,
			Time:      time.Now(),
		}

		startTime := time.Now()

		// Perform lock/unlock cycle
		acquired, _, err := s.adapter.TryLock(lock)
		require.NoError(t, err)
		require.True(t, acquired)

		_, err = s.adapter.Unlock(project, lock.Workspace, user)
		require.NoError(t, err)

		operationTime := time.Since(startTime)
		operationTimes = append(operationTimes, operationTime)
	}

	// Calculate performance metrics
	var totalTime time.Duration
	maxTime := time.Duration(0)
	minTime := time.Hour // Initialize to large value

	for _, t := range operationTimes {
		totalTime += t
		if t > maxTime {
			maxTime = t
		}
		if t < minTime {
			minTime = t
		}
	}

	avgTime := totalTime / time.Duration(len(operationTimes))

	t.Logf("Performance metrics for %d operations:", numOperations)
	t.Logf("  Average: %v", avgTime)
	t.Logf("  Min: %v", minTime)
	t.Logf("  Max: %v", maxTime)

	// Performance assertions
	assert.Less(t, avgTime, 100*time.Millisecond, "Average operation should be under 100ms")
	assert.Less(t, maxTime, 1*time.Second, "Max operation should be under 1 second")
}

// TestFallbackCompatibility tests fallback to legacy system
func (s *BackwardCompatibilityTestSuite) TestFallbackCompatibility(t *testing.T) {
	if !s.config.LegacyFallback {
		t.Skip("Legacy fallback not enabled")
	}

	project := models.Project{RepoFullName: "test/fallback", Path: "."}
	user := models.User{Username: "fallbackuser"}
	workspace := "default"
	lock := models.ProjectLock{Project: project, Workspace: workspace, User: user, Time: time.Now()}

	// Temporarily disable enhanced system
	originalEnabled := s.config.Enabled
	s.config.Enabled = false

	// Operations should use legacy backend
	acquired, _, err := s.adapter.TryLock(lock)
	require.NoError(t, err)
	assert.True(t, acquired)

	// Verify lock exists
	retrievedLock, err := s.adapter.GetLock(project, workspace)
	require.NoError(t, err)
	require.NotNil(t, retrievedLock)

	// Release lock
	unlockedLock, err := s.adapter.Unlock(project, workspace, user)
	require.NoError(t, err)
	require.NotNil(t, unlockedLock)

	// Restore enhanced system
	s.config.Enabled = originalEnabled

	t.Log("Fallback compatibility test completed successfully")
}

// MockLegacyBackend implements locking.Backend for testing fallback behavior
type MockLegacyBackend struct {
	locks   map[string]models.ProjectLock
	mutex   sync.RWMutex
}

func NewMockLegacyBackend() *MockLegacyBackend {
	return &MockLegacyBackend{
		locks: make(map[string]models.ProjectLock),
	}
}

func (m *MockLegacyBackend) TryLock(lock models.ProjectLock) (bool, locking.TryLockResponse, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	key := fmt.Sprintf("%s/%s/%s", lock.Project.RepoFullName, lock.Project.Path, lock.Workspace)

	if existingLock, exists := m.locks[key]; exists {
		return false, locking.TryLockResponse{
			CurrLock:          existingLock,
			LockFailureReason: fmt.Sprintf("Lock held by %s", existingLock.User.Username),
		}, nil
	}

	m.locks[key] = lock
	return true, locking.TryLockResponse{}, nil
}

func (m *MockLegacyBackend) Unlock(project models.Project, workspace string, user models.User) (*models.ProjectLock, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	key := fmt.Sprintf("%s/%s/%s", project.RepoFullName, project.Path, workspace)

	if lock, exists := m.locks[key]; exists {
		if lock.User.Username == user.Username {
			delete(m.locks, key)
			return &lock, nil
		}
	}

	return nil, nil
}

func (m *MockLegacyBackend) List() (map[string]models.ProjectLock, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	result := make(map[string]models.ProjectLock)
	for key, lock := range m.locks {
		result[key] = lock
	}
	return result, nil
}

func (m *MockLegacyBackend) UnlockByPull(repoFullName string, pullNum int) ([]models.ProjectLock, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	var unlockedLocks []models.ProjectLock
	keysToDelete := make([]string, 0)

	for key, lock := range m.locks {
		if lock.Project.RepoFullName == repoFullName {
			unlockedLocks = append(unlockedLocks, lock)
			keysToDelete = append(keysToDelete, key)
		}
	}

	for _, key := range keysToDelete {
		delete(m.locks, key)
	}

	return unlockedLocks, nil
}

func (m *MockLegacyBackend) GetLock(project models.Project, workspace string) (*models.ProjectLock, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	key := fmt.Sprintf("%s/%s/%s", project.RepoFullName, project.Path, workspace)
	if lock, exists := m.locks[key]; exists {
		return &lock, nil
	}
	return nil, nil
}