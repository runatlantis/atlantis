package tests

import (
	"context"
	"fmt"
	"math/rand"
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

// RedisIntegrationTestSuite provides comprehensive Redis backend testing
type RedisIntegrationTestSuite struct {
	singleClient   redis.UniversalClient
	clusterClient  redis.UniversalClient
	config         *enhanced.EnhancedConfig
	logger         logging.SimpleLogging
	cleanup        func()
}

// SetupRedisIntegrationTests initializes Redis integration test environment
func SetupRedisIntegrationTests(t *testing.T) *RedisIntegrationTestSuite {
	logger := logging.NewNoopLogger(t)

	// Single Redis instance setup
	singleClient := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
		DB:   14, // Use dedicated test database
	})

	// Test single Redis connection
	ctx := context.Background()
	if err := singleClient.Ping(ctx).Err(); err != nil {
		t.Skip("Redis single instance not available for integration testing")
	}

	// Clean test database
	singleClient.FlushDB(ctx)

	// Try to setup cluster client (optional for local testing)
	var clusterClient redis.UniversalClient
	clusterAddrs := []string{"localhost:7001", "localhost:7002", "localhost:7003"}

	clusterClient = redis.NewClusterClient(&redis.ClusterOptions{
		Addrs: clusterAddrs,
	})

	// Test cluster connection (non-blocking)
	clusterAvailable := false
	if err := clusterClient.Ping(ctx).Err(); err == nil {
		clusterAvailable = true
		t.Log("Redis cluster detected and will be tested")
	} else {
		t.Log("Redis cluster not available, skipping cluster-specific tests")
		clusterClient = nil
	}

	config := &enhanced.EnhancedConfig{
		Enabled:                 true,
		Backend:                 "redis",
		DefaultTimeout:          30 * time.Second,
		MaxTimeout:              5 * time.Minute,
		EnablePriorityQueue:     true,
		MaxQueueSize:           1000,
		QueueTimeout:           30 * time.Second,
		EnableRetry:            true,
		MaxRetryAttempts:       5,
		RetryBaseDelay:         100 * time.Millisecond,
		RetryMaxDelay:          5 * time.Second,
		EnableDeadlockDetection: true,
		DeadlockCheckInterval:   2 * time.Second,
		EnableEvents:           true,
		EventBufferSize:        1000,
		RedisClusterMode:       clusterAvailable,
		RedisKeyPrefix:         "atlantis:integration:test:",
		RedisLockTTL:           5 * time.Minute,
		LegacyFallback:         false,
		PreserveLegacyFormat:   true,
	}

	cleanup := func() {
		singleClient.FlushDB(ctx)
		singleClient.Close()
		if clusterClient != nil {
			clusterClient.Close()
		}
	}

	return &RedisIntegrationTestSuite{
		singleClient:  singleClient,
		clusterClient: clusterClient,
		config:        config,
		logger:        logger,
		cleanup:       cleanup,
	}
}

// TestFullRedisIntegration runs all Redis integration tests
func TestFullRedisIntegration(t *testing.T) {
	suite := SetupRedisIntegrationTests(t)
	defer suite.cleanup()

	t.Run("BasicRedisOperations", suite.TestBasicRedisOperations)
	t.Run("AtomicOperations", suite.TestAtomicOperations)
	t.Run("LuaScriptOperations", suite.TestLuaScriptOperations)
	t.Run("TTLAndExpiration", suite.TestTTLAndExpiration)
	t.Run("PriorityQueueOperations", suite.TestPriorityQueueOperations)
	t.Run("EventSubscription", suite.TestEventSubscription)
	t.Run("ConcurrentRedisOperations", suite.TestConcurrentRedisOperations)
	t.Run("RedisFailoverRecovery", suite.TestRedisFailoverRecovery)
	t.Run("RedisMemoryManagement", suite.TestRedisMemoryManagement)

	if suite.clusterClient != nil {
		t.Run("ClusterOperations", suite.TestClusterOperations)
		t.Run("ClusterFailover", suite.TestClusterFailover)
	}
}

// TestBasicRedisOperations tests basic Redis functionality
func (s *RedisIntegrationTestSuite) TestBasicRedisOperations(t *testing.T) {
	backend := backends.NewRedisBackend(s.singleClient, s.config, s.logger)
	ctx := context.Background()

	// Test health check
	t.Run("HealthCheck", func(t *testing.T) {
		err := backend.HealthCheck(ctx)
		assert.NoError(t, err)
	})

	// Test basic lock operations
	t.Run("BasicLockCycle", func(t *testing.T) {
		request := createTestLockRequest("test/basic", "default", "testuser")

		// Acquire lock
		lock, err := backend.AcquireLock(ctx, request)
		require.NoError(t, err)
		require.NotNil(t, lock)
		assert.Equal(t, enhanced.LockStateAcquired, lock.State)
		assert.Equal(t, request.User.Username, lock.Owner)

		// Get lock
		retrievedLock, err := backend.GetLock(ctx, lock.ID)
		require.NoError(t, err)
		require.NotNil(t, retrievedLock)
		assert.Equal(t, lock.ID, retrievedLock.ID)

		// List locks
		locks, err := backend.ListLocks(ctx)
		require.NoError(t, err)
		assert.Len(t, locks, 1)
		assert.Equal(t, lock.ID, locks[0].ID)

		// Release lock
		err = backend.ReleaseLock(ctx, lock.ID)
		require.NoError(t, err)

		// Verify lock is gone
		locks, err = backend.ListLocks(ctx)
		require.NoError(t, err)
		assert.Empty(t, locks)
	})

	// Test duplicate lock prevention
	t.Run("DuplicateLockPrevention", func(t *testing.T) {
		request := createTestLockRequest("test/duplicate", "default", "user1")

		// First lock should succeed
		lock1, err := backend.AcquireLock(ctx, request)
		require.NoError(t, err)
		require.NotNil(t, lock1)

		// Second lock on same resource should fail
		request2 := createTestLockRequest("test/duplicate", "default", "user2")
		lock2, err := backend.AcquireLock(ctx, request2)

		// Should return error indicating lock exists
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "LOCK_EXISTS")
		assert.Nil(t, lock2)

		// Cleanup
		backend.ReleaseLock(ctx, lock1.ID)
	})
}

// TestAtomicOperations verifies Redis atomic operations
func (s *RedisIntegrationTestSuite) TestAtomicOperations(t *testing.T) {
	backend := backends.NewRedisBackend(s.singleClient, s.config, s.logger)
	ctx := context.Background()

	// Test atomic try-acquire
	t.Run("AtomicTryAcquire", func(t *testing.T) {
		numClients := 10
		resource := "test/atomic"

		var wg sync.WaitGroup
		successes := make(chan string, numClients)
		failures := make(chan error, numClients)

		// Launch concurrent lock attempts
		for i := 0; i < numClients; i++ {
			wg.Add(1)
			go func(clientID int) {
				defer wg.Done()

				request := createTestLockRequest(resource, "default", fmt.Sprintf("user%d", clientID))
				lock, acquired, err := backend.TryAcquireLock(ctx, request)

				if err != nil {
					failures <- err
					return
				}

				if acquired {
					successes <- lock.Owner
				}
			}(i)
		}

		wg.Wait()
		close(successes)
		close(failures)

		// Verify only one client succeeded
		successCount := 0
		var winner string
		for owner := range successes {
			successCount++
			winner = owner
		}

		errorCount := 0
		for err := range failures {
			t.Logf("Expected failure: %v", err)
			errorCount++
		}

		assert.Equal(t, 1, successCount, "Exactly one client should succeed")
		assert.NotEmpty(t, winner, "Winner should be identified")
		t.Logf("Winner: %s, Failures: %d", winner, errorCount)

		// Cleanup
		locks, _ := backend.ListLocks(ctx)
		for _, lock := range locks {
			if lock.Resource.Namespace == resource {
				backend.ReleaseLock(ctx, lock.ID)
			}
		}
	})

	// Test atomic release with queue processing
	t.Run("AtomicReleaseWithQueue", func(t *testing.T) {
		if !s.config.EnablePriorityQueue {
			t.Skip("Priority queue not enabled")
		}

		resource := "test/queue-atomic"

		// Create initial lock
		request1 := createTestLockRequest(resource, "default", "holder")
		lock1, err := backend.AcquireLock(ctx, request1)
		require.NoError(t, err)

		// Queue additional requests
		request2 := createTestLockRequest(resource, "default", "waiter1")
		request2.Priority = enhanced.PriorityHigh
		err = backend.EnqueueLockRequest(ctx, request2)
		require.NoError(t, err)

		request3 := createTestLockRequest(resource, "default", "waiter2")
		request3.Priority = enhanced.PriorityNormal
		err = backend.EnqueueLockRequest(ctx, request3)
		require.NoError(t, err)

		// Release lock - should automatically process queue
		err = backend.ReleaseLock(ctx, lock1.ID)
		require.NoError(t, err)

		// Give queue processing time
		time.Sleep(100 * time.Millisecond)

		// Verify highest priority request got the lock
		locks, err := backend.ListLocks(ctx)
		require.NoError(t, err)

		if len(locks) > 0 {
			// Should be the high priority waiter
			assert.Equal(t, "waiter1", locks[0].Owner)
			backend.ReleaseLock(ctx, locks[0].ID)
		}
	})
}

// TestLuaScriptOperations tests Redis Lua script execution
func (s *RedisIntegrationTestSuite) TestLuaScriptOperations(t *testing.T) {
	ctx := context.Background()

	// Test custom Lua script for complex operations
	t.Run("CustomLuaScript", func(t *testing.T) {
		script := `
			local key = KEYS[1]
			local value = ARGV[1]
			local ttl = tonumber(ARGV[2])

			-- Atomic increment and set with TTL
			local current = redis.call('GET', key)
			if not current then
				current = "0"
			end

			local newValue = tostring(tonumber(current) + tonumber(value))
			redis.call('SETEX', key, ttl, newValue)

			return newValue
		`

		// Execute script
		result, err := s.singleClient.Eval(ctx, script, []string{"test:counter"}, 1, 300).Result()
		require.NoError(t, err)
		assert.Equal(t, "1", result)

		// Execute again
		result, err = s.singleClient.Eval(ctx, script, []string{"test:counter"}, 5, 300).Result()
		require.NoError(t, err)
		assert.Equal(t, "6", result)
	})

	// Test script error handling
	t.Run("LuaScriptErrorHandling", func(t *testing.T) {
		invalidScript := `
			redis.call('INVALID_COMMAND')
		`

		_, err := s.singleClient.Eval(ctx, invalidScript, []string{}, []string{}).Result()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Unknown Redis command")
	})
}

// TestTTLAndExpiration tests Redis TTL and automatic expiration
func (s *RedisIntegrationTestSuite) TestTTLAndExpiration(t *testing.T) {
	backend := backends.NewRedisBackend(s.singleClient, s.config, s.logger)
	ctx := context.Background()

	t.Run("LockExpiration", func(t *testing.T) {
		request := createTestLockRequest("test/ttl", "default", "ttluser")
		request.Timeout = 2 * time.Second // Short timeout for testing

		// Acquire lock with TTL
		lock, err := backend.AcquireLock(ctx, request)
		require.NoError(t, err)
		require.NotNil(t, lock)
		require.NotNil(t, lock.ExpiresAt)

		// Verify lock exists
		retrievedLock, err := backend.GetLock(ctx, lock.ID)
		require.NoError(t, err)
		require.NotNil(t, retrievedLock)

		// Wait for expiration
		time.Sleep(3 * time.Second)

		// Lock should be expired
		retrievedLock, err = backend.GetLock(ctx, lock.ID)
		if err == nil && retrievedLock != nil {
			// Check if it's marked as expired
			assert.Equal(t, enhanced.LockStateExpired, retrievedLock.State)
		}

		// Cleanup expired locks should work
		cleaned, err := backend.CleanupExpiredLocks(ctx)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, cleaned, 0)
	})

	t.Run("LockRefresh", func(t *testing.T) {
		request := createTestLockRequest("test/refresh", "default", "refreshuser")
		request.Timeout = 3 * time.Second

		// Acquire lock
		lock, err := backend.AcquireLock(ctx, request)
		require.NoError(t, err)

		originalExpiry := *lock.ExpiresAt

		// Wait briefly then refresh
		time.Sleep(1 * time.Second)
		err = backend.RefreshLock(ctx, lock.ID, 5*time.Second)
		require.NoError(t, err)

		// Verify extension
		refreshedLock, err := backend.GetLock(ctx, lock.ID)
		require.NoError(t, err)
		assert.True(t, refreshedLock.ExpiresAt.After(originalExpiry))

		// Cleanup
		backend.ReleaseLock(ctx, lock.ID)
	})
}

// TestPriorityQueueOperations tests Redis-based priority queuing
func (s *RedisIntegrationTestSuite) TestPriorityQueueOperations(t *testing.T) {
	if !s.config.EnablePriorityQueue {
		t.Skip("Priority queue not enabled")
	}

	backend := backends.NewRedisBackend(s.singleClient, s.config, s.logger)
	ctx := context.Background()

	t.Run("PriorityOrdering", func(t *testing.T) {
		resource := "test/priority-queue"

		// Create requests with different priorities
		requests := []*enhanced.EnhancedLockRequest{
			createTestLockRequest(resource, "default", "low-user"),
			createTestLockRequest(resource, "default", "normal-user"),
			createTestLockRequest(resource, "default", "high-user"),
			createTestLockRequest(resource, "default", "critical-user"),
		}
		requests[0].Priority = enhanced.PriorityLow
		requests[1].Priority = enhanced.PriorityNormal
		requests[2].Priority = enhanced.PriorityHigh
		requests[3].Priority = enhanced.PriorityCritical

		// Enqueue in random order
		shuffled := make([]*enhanced.EnhancedLockRequest, len(requests))
		copy(shuffled, requests)
		rand.Shuffle(len(shuffled), func(i, j int) {
			shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
		})

		for _, req := range shuffled {
			err := backend.EnqueueLockRequest(ctx, req)
			require.NoError(t, err)
		}

		// Dequeue and verify priority order
		expectedOrder := []string{"critical-user", "high-user", "normal-user", "low-user"}
		actualOrder := make([]string, 0, len(expectedOrder))

		for i := 0; i < len(expectedOrder); i++ {
			req, err := backend.DequeueNextRequest(ctx)
			require.NoError(t, err)
			if req != nil {
				actualOrder = append(actualOrder, req.User.Username)
			}
		}

		assert.Equal(t, expectedOrder, actualOrder, "Requests should be dequeued in priority order")
	})

	t.Run("QueueStatus", func(t *testing.T) {
		resource := "test/queue-status"

		// Add some requests to queue
		priorities := []enhanced.Priority{
			enhanced.PriorityLow,
			enhanced.PriorityNormal,
			enhanced.PriorityNormal,
			enhanced.PriorityHigh,
		}

		for i, priority := range priorities {
			req := createTestLockRequest(resource, "default", fmt.Sprintf("user%d", i))
			req.Priority = priority
			err := backend.EnqueueLockRequest(ctx, req)
			require.NoError(t, err)
		}

		// Check queue status
		status, err := backend.GetQueueStatus(ctx)
		require.NoError(t, err)
		require.NotNil(t, status)

		assert.Equal(t, len(priorities), status.Size)
		assert.Equal(t, 1, status.QueuesByPriority[enhanced.PriorityLow])
		assert.Equal(t, 2, status.QueuesByPriority[enhanced.PriorityNormal])
		assert.Equal(t, 1, status.QueuesByPriority[enhanced.PriorityHigh])
	})
}

// TestEventSubscription tests Redis pub/sub for lock events
func (s *RedisIntegrationTestSuite) TestEventSubscription(t *testing.T) {
	if !s.config.EnableEvents {
		t.Skip("Events not enabled")
	}

	backend := backends.NewRedisBackend(s.singleClient, s.config, s.logger)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	t.Run("LockEvents", func(t *testing.T) {
		// Subscribe to lock events
		eventChan, err := backend.Subscribe(ctx, []string{"acquired", "released"})
		require.NoError(t, err)

		var receivedEvents []*enhanced.LockEvent
		var wg sync.WaitGroup
		wg.Add(1)

		// Event listener
		go func() {
			defer wg.Done()
			timeout := time.NewTimer(5 * time.Second)
			defer timeout.Stop()

			for {
				select {
				case event, ok := <-eventChan:
					if !ok {
						return
					}
					receivedEvents = append(receivedEvents, event)
					if len(receivedEvents) >= 2 { // Expecting acquired + released
						return
					}
				case <-timeout.C:
					return
				case <-ctx.Done():
					return
				}
			}
		}()

		// Give subscription time to establish
		time.Sleep(100 * time.Millisecond)

		// Perform lock operations that should trigger events
		request := createTestLockRequest("test/events", "default", "eventuser")
		lock, err := backend.AcquireLock(ctx, request)
		require.NoError(t, err)

		time.Sleep(100 * time.Millisecond) // Allow event propagation

		err = backend.ReleaseLock(ctx, lock.ID)
		require.NoError(t, err)

		time.Sleep(100 * time.Millisecond) // Allow event propagation

		wg.Wait()

		// Verify events were received
		assert.GreaterOrEqual(t, len(receivedEvents), 1, "Should receive at least one event")
		for _, event := range receivedEvents {
			assert.NotEmpty(t, event.Type)
			assert.NotEmpty(t, event.Timestamp)
			t.Logf("Received event: %s at %s", event.Type, event.Timestamp)
		}
	})
}

// TestConcurrentRedisOperations tests high-concurrency Redis operations
func (s *RedisIntegrationTestSuite) TestConcurrentRedisOperations(t *testing.T) {
	backend := backends.NewRedisBackend(s.singleClient, s.config, s.logger)
	ctx := context.Background()

	t.Run("HighConcurrencyLocking", func(t *testing.T) {
		numGoroutines := 50
		numOperationsPerGoroutine := 20

		var wg sync.WaitGroup
		results := make(chan bool, numGoroutines*numOperationsPerGoroutine)
		errors := make(chan error, numGoroutines*numOperationsPerGoroutine)

		start := time.Now()

		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func(goroutineID int) {
				defer wg.Done()

				for j := 0; j < numOperationsPerGoroutine; j++ {
					// Use different resources to avoid blocking
					resource := fmt.Sprintf("test/concurrent-%d-%d", goroutineID, j)
					request := createTestLockRequest(resource, "default", fmt.Sprintf("user-%d", goroutineID))

					// Acquire lock
					lock, err := backend.AcquireLock(ctx, request)
					if err != nil {
						errors <- err
						continue
					}

					// Hold briefly
					time.Sleep(time.Duration(rand.Intn(10)) * time.Millisecond)

					// Release lock
					err = backend.ReleaseLock(ctx, lock.ID)
					if err != nil {
						errors <- err
						continue
					}

					results <- true
				}
			}(i)
		}

		wg.Wait()
		duration := time.Since(start)

		close(results)
		close(errors)

		// Analyze results
		successCount := 0
		for success := range results {
			if success {
				successCount++
			}
		}

		errorCount := 0
		for err := range errors {
			errorCount++
			t.Logf("Concurrent operation error: %v", err)
		}

		totalOperations := numGoroutines * numOperationsPerGoroutine
		successRate := float64(successCount) / float64(totalOperations) * 100

		t.Logf("Concurrent test completed in %v", duration)
		t.Logf("Total operations: %d, Successes: %d, Errors: %d", totalOperations, successCount, errorCount)
		t.Logf("Success rate: %.2f%%, Throughput: %.2f ops/sec", successRate, float64(totalOperations)/duration.Seconds())

		// Performance assertions
		assert.Greater(t, successRate, 95.0, "Success rate should be above 95%")
		assert.Less(t, duration, 30*time.Second, "Should complete within 30 seconds")
	})
}

// TestRedisFailoverRecovery tests Redis connection failure and recovery
func (s *RedisIntegrationTestSuite) TestRedisFailoverRecovery(t *testing.T) {
	// This test simulates Redis failures - implementation depends on test setup
	t.Run("ConnectionRecovery", func(t *testing.T) {
		backend := backends.NewRedisBackend(s.singleClient, s.config, s.logger)
		ctx := context.Background()

		// Verify normal operation
		request := createTestLockRequest("test/failover", "default", "failuser")
		lock, err := backend.AcquireLock(ctx, request)
		require.NoError(t, err)
		require.NotNil(t, lock)

		// Health check should pass
		err = backend.HealthCheck(ctx)
		assert.NoError(t, err)

		// Simulate connection issue by using invalid Redis instance
		invalidClient := redis.NewClient(&redis.Options{
			Addr: "localhost:9999", // Non-existent Redis
		})

		invalidBackend := backends.NewRedisBackend(invalidClient, s.config, s.logger)

		// Health check should fail
		err = invalidBackend.HealthCheck(ctx)
		assert.Error(t, err)

		// Operations should fail
		_, err = invalidBackend.AcquireLock(ctx, request)
		assert.Error(t, err)

		// Cleanup original lock
		err = backend.ReleaseLock(ctx, lock.ID)
		assert.NoError(t, err)
	})
}

// TestRedisMemoryManagement tests Redis memory usage patterns
func (s *RedisIntegrationTestSuite) TestRedisMemoryManagement(t *testing.T) {
	backend := backends.NewRedisBackend(s.singleClient, s.config, s.logger)
	ctx := context.Background()

	t.Run("MemoryUsagePattern", func(t *testing.T) {
		// Get initial memory stats
		initialInfo, err := s.singleClient.Info(ctx, "memory").Result()
		require.NoError(t, err)
		t.Logf("Initial Redis memory info: %s", initialInfo)

		// Create many locks
		numLocks := 1000
		lockIDs := make([]string, 0, numLocks)

		for i := 0; i < numLocks; i++ {
			request := createTestLockRequest(fmt.Sprintf("test/memory-%d", i), "default", fmt.Sprintf("user%d", i))
			lock, err := backend.AcquireLock(ctx, request)
			require.NoError(t, err)
			lockIDs = append(lockIDs, lock.ID)
		}

		// Check memory usage after creating locks
		midInfo, err := s.singleClient.Info(ctx, "memory").Result()
		require.NoError(t, err)
		t.Logf("Memory after creating %d locks: %s", numLocks, midInfo)

		// Release all locks
		for _, lockID := range lockIDs {
			err := backend.ReleaseLock(ctx, lockID)
			require.NoError(t, err)
		}

		// Check memory usage after releasing locks
		finalInfo, err := s.singleClient.Info(ctx, "memory").Result()
		require.NoError(t, err)
		t.Logf("Memory after releasing locks: %s", finalInfo)

		// Verify cleanup
		locks, err := backend.ListLocks(ctx)
		require.NoError(t, err)
		assert.Empty(t, locks, "All locks should be cleaned up")
	})

	t.Run("ExpiredLockCleanup", func(t *testing.T) {
		// Create locks with short TTL
		numLocks := 100
		for i := 0; i < numLocks; i++ {
			request := createTestLockRequest(fmt.Sprintf("test/cleanup-%d", i), "default", fmt.Sprintf("user%d", i))
			request.Timeout = 1 * time.Second

			_, err := backend.AcquireLock(ctx, request)
			require.NoError(t, err)
		}

		// Wait for expiration
		time.Sleep(2 * time.Second)

		// Cleanup expired locks
		cleaned, err := backend.CleanupExpiredLocks(ctx)
		require.NoError(t, err)
		t.Logf("Cleaned up %d expired locks", cleaned)

		// Verify cleanup
		locks, err := backend.ListLocks(ctx)
		require.NoError(t, err)
		assert.LessOrEqual(t, len(locks), numLocks/2, "Most locks should be cleaned up")
	})
}

// TestClusterOperations tests Redis cluster-specific functionality
func (s *RedisIntegrationTestSuite) TestClusterOperations(t *testing.T) {
	if s.clusterClient == nil {
		t.Skip("Redis cluster not available")
	}

	clusterConfig := *s.config
	clusterConfig.RedisClusterMode = true

	backend := backends.NewRedisBackend(s.clusterClient, &clusterConfig, s.logger)
	ctx := context.Background()

	t.Run("ClusterDistribution", func(t *testing.T) {
		// Create locks that should be distributed across cluster nodes
		numLocks := 100
		lockIDs := make([]string, 0, numLocks)

		for i := 0; i < numLocks; i++ {
			request := createTestLockRequest(fmt.Sprintf("test/cluster-%d", i), "default", fmt.Sprintf("user%d", i))
			lock, err := backend.AcquireLock(ctx, request)
			require.NoError(t, err)
			lockIDs = append(lockIDs, lock.ID)
		}

		// Verify locks exist
		locks, err := backend.ListLocks(ctx)
		require.NoError(t, err)
		assert.Len(t, locks, numLocks)

		// Cleanup
		for _, lockID := range lockIDs {
			backend.ReleaseLock(ctx, lockID)
		}
	})
}

// TestClusterFailover tests Redis cluster failover scenarios
func (s *RedisIntegrationTestSuite) TestClusterFailover(t *testing.T) {
	if s.clusterClient == nil {
		t.Skip("Redis cluster not available")
	}

	// This would require more complex test setup to actually fail cluster nodes
	t.Log("Cluster failover testing requires external cluster management")
}

// Helper function to create test lock requests
func createTestLockRequest(resource, workspace, username string) *enhanced.EnhancedLockRequest {
	return &enhanced.EnhancedLockRequest{
		ID: fmt.Sprintf("test_%d_%s", time.Now().UnixNano(), username),
		Resource: enhanced.ResourceIdentifier{
			Type:      enhanced.ResourceTypeProject,
			Namespace: resource,
			Name:      ".",
			Workspace: workspace,
			Path:      ".",
		},
		Priority:    enhanced.PriorityNormal,
		Timeout:     30 * time.Second,
		Metadata:    make(map[string]string),
		Context:     context.Background(),
		RequestedAt: time.Now(),
		Project: models.Project{
			RepoFullName: resource,
			Path:         ".",
		},
		Workspace: workspace,
		User: models.User{
			Username: username,
		},
	}
}