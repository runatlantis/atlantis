package enhanced

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPriorityString(t *testing.T) {
	tests := []struct {
		priority Priority
		expected string
	}{
		{PriorityLow, "low"},
		{PriorityNormal, "normal"},
		{PriorityHigh, "high"},
		{PriorityCritical, "critical"},
		{Priority(999), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.priority.String())
		})
	}
}

func TestParsePriority(t *testing.T) {
	tests := []struct {
		input       string
		expected    Priority
		expectError bool
	}{
		{"low", PriorityLow, false},
		{"normal", PriorityNormal, false},
		{"high", PriorityHigh, false},
		{"critical", PriorityCritical, false},
		{"invalid", PriorityNormal, true},
		{"", PriorityNormal, true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			priority, err := ParsePriority(tt.input)
			if tt.expectError {
				assert.Error(t, err)
				assert.IsType(t, &ConfigError{}, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, priority)
			}
		})
	}
}

func TestIsValidPriority(t *testing.T) {
	tests := []struct {
		priority Priority
		expected bool
	}{
		{PriorityLow, true},
		{PriorityNormal, true},
		{PriorityHigh, true},
		{PriorityCritical, true},
		{Priority(-1), false},
		{Priority(999), false},
	}

	for _, tt := range tests {
		t.Run(tt.priority.String(), func(t *testing.T) {
			assert.Equal(t, tt.expected, IsValidPriority(tt.priority))
		})
	}
}

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	assert.NotNil(t, config)
	assert.Equal(t, "boltdb", config.Backend)
	assert.False(t, config.EnablePriorityQueue)
	assert.False(t, config.EnableRetries)
	assert.True(t, config.EnableMetrics)
	assert.Equal(t, 30*time.Second, config.DefaultTimeout)
	assert.Equal(t, 5*time.Minute, config.MaxTimeout)
	assert.Equal(t, 1000, config.MaxQueueSize)
	assert.Equal(t, 2*time.Minute, config.StarvationThreshold)
	assert.Equal(t, 3, config.MaxPriorityBoost)
}

func TestConfigValidation(t *testing.T) {
	t.Run("valid default config", func(t *testing.T) {
		config := DefaultConfig()
		err := config.Validate()
		assert.NoError(t, err)
	})

	t.Run("invalid backend", func(t *testing.T) {
		config := DefaultConfig()
		config.Backend = "invalid"
		err := config.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "backend")
	})

	t.Run("invalid timeout", func(t *testing.T) {
		config := DefaultConfig()
		config.DefaultTimeout = -1
		err := config.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "default_timeout")
	})

	t.Run("max timeout less than default", func(t *testing.T) {
		config := DefaultConfig()
		config.MaxTimeout = config.DefaultTimeout - 1
		err := config.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "max_timeout")
	})

	t.Run("invalid queue size", func(t *testing.T) {
		config := DefaultConfig()
		config.MaxQueueSize = 0
		err := config.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "max_queue_size")
	})
}

func TestRedisConfigValidation(t *testing.T) {
	t.Run("empty addresses", func(t *testing.T) {
		config := &RedisConfig{}
		err := config.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "addresses")
	})

	t.Run("invalid pool size", func(t *testing.T) {
		config := &RedisConfig{
			Addresses: []string{"localhost:6379"},
			PoolSize:  0,
		}
		err := config.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "pool_size")
	})

	t.Run("invalid TTL", func(t *testing.T) {
		config := &RedisConfig{
			Addresses:  []string{"localhost:6379"},
			PoolSize:   10,
			DefaultTTL: -1,
		}
		err := config.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "default_ttl")
	})

	t.Run("valid redis config", func(t *testing.T) {
		config := &RedisConfig{
			Addresses:  []string{"localhost:6379"},
			PoolSize:   10,
			DefaultTTL: time.Hour,
		}
		err := config.Validate()
		assert.NoError(t, err)
	})
}

func TestConfigError(t *testing.T) {
	err := &ConfigError{
		Field: "test_field",
		Value: "test_value",
		Msg:   "test message",
	}

	assert.Contains(t, err.Error(), "test_field")
	assert.Contains(t, err.Error(), "test message")
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