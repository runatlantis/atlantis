package enhanced

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	assert.NotNil(t, config)
	assert.Equal(t, "boltdb", config.Backend)
	assert.False(t, config.EnablePriorityQueue)
	assert.False(t, config.EnableRetries)
	assert.True(t, config.EnableMetrics)
	assert.Equal(t, 30*time.Second, config.DefaultTimeout)
	assert.Equal(t, 5*time.Minute, config.MaxTimeout)
}

func TestLockErrorCreation(t *testing.T) {
	t.Run("lock exists error", func(t *testing.T) {
		err := NewLockExistsError("test-resource")
		assert.Equal(t, ErrCodeLockExists, err.Code)
		assert.Contains(t, err.Message, "test-resource")
	})

	t.Run("lock not found error", func(t *testing.T) {
		err := NewLockNotFoundError("test-lock-id")
		assert.Equal(t, ErrCodeLockNotFound, err.Code)
		assert.Contains(t, err.Message, "test-lock-id")
	})

	t.Run("timeout error", func(t *testing.T) {
		timeout := 30 * time.Second
		err := NewTimeoutError(timeout)
		assert.Equal(t, ErrCodeTimeout, err.Code)
		assert.Contains(t, err.Message, timeout.String())
	})
}

func TestPriorityConstants(t *testing.T) {
	// Test that priority constants are defined
	assert.Equal(t, Priority(0), PriorityLow)
	assert.Equal(t, Priority(1), PriorityNormal)
	assert.Equal(t, Priority(2), PriorityHigh)
	assert.Equal(t, Priority(3), PriorityCritical)
}

func TestLockStateConstants(t *testing.T) {
	// Test that lock state constants are defined
	assert.Equal(t, LockState("acquired"), LockStateAcquired)
	assert.Equal(t, LockState("pending"), LockStatePending)
	assert.Equal(t, LockState("expired"), LockStateExpired)
	assert.Equal(t, LockState("released"), LockStateReleased)
}

func TestResourceTypeConstants(t *testing.T) {
	// Test that resource type constants are defined
	assert.Equal(t, ResourceType("project"), ResourceTypeProject)
	assert.Equal(t, ResourceType("workspace"), ResourceTypeWorkspace)
	assert.Equal(t, ResourceType("global"), ResourceTypeGlobal)
	assert.Equal(t, ResourceType("custom"), ResourceTypeCustom)
}

func TestLockErrorInterface(t *testing.T) {
	err := &LockError{
		Type:    "TestError",
		Message: "test message",
		Code:    "TEST_CODE",
	}

	errStr := err.Error()
	assert.Contains(t, errStr, "TEST_CODE")
	assert.Contains(t, errStr, "test message")
}

// Placeholder tests for future implementation
func TestEnhancedLockingIntegration(t *testing.T) {
	t.Skip("Enhanced locking integration tests - to be implemented in future PRs")
}

func TestRedisBackendIntegration(t *testing.T) {
	t.Skip("Redis backend integration tests - to be implemented in Redis backend PR")
}

func TestPriorityQueueIntegration(t *testing.T) {
	t.Skip("Priority queue integration tests - to be implemented in priority queue PR")
}

func TestEventStreamingIntegration(t *testing.T) {
	t.Skip("Event streaming integration tests - to be implemented in event streaming PR")
}

func TestDeadlockDetectionIntegration(t *testing.T) {
	t.Skip("Deadlock detection integration tests - to be implemented in deadlock detection PR")
}

func TestMigrationIntegration(t *testing.T) {
	t.Skip("Migration integration tests - to be implemented in migration PR")
}

func TestPerformanceBenchmarks(t *testing.T) {
	t.Skip("Performance benchmarks - to be implemented in optimization PRs")
}