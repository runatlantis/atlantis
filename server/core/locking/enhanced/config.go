package enhanced

import (
	"fmt"
	"time"
)

// EnhancedConfig holds configuration for enhanced locking
type EnhancedConfig struct {
	Enabled        bool          `mapstructure:"enabled"`
	Backend        string        `mapstructure:"backend"` // "redis", "boltdb"
	DefaultTimeout time.Duration `mapstructure:"default_timeout"`
	MaxTimeout     time.Duration `mapstructure:"max_timeout"`

	// Future feature flags (disabled by default)
	EnablePriorityQueue bool `mapstructure:"enable_priority_queue"`
	EnableRetries       bool `mapstructure:"enable_retries"`
	EnableMetrics       bool `mapstructure:"enable_metrics"`
	EnableEvents        bool `mapstructure:"enable_events"`

	// Queue and retry settings
	MaxQueueSize        int           `mapstructure:"max_queue_size"`
	MaxRetryAttempts    int           `mapstructure:"max_retry_attempts"`
	QueueTimeout        time.Duration `mapstructure:"queue_timeout"`
	EnableRetry         bool          `mapstructure:"enable_retry"`
	RetryBaseDelay      time.Duration `mapstructure:"retry_base_delay"`
	RetryMaxDelay       time.Duration `mapstructure:"retry_max_delay"`
	StarvationThreshold time.Duration `mapstructure:"starvation_threshold"`
	MaxPriorityBoost    int           `mapstructure:"max_priority_boost"`

	// Deadlock detection settings
	EnableDeadlockDetection bool          `mapstructure:"enable_deadlock_detection"`
	DeadlockCheckInterval   time.Duration `mapstructure:"deadlock_check_interval"`

	// Redis-specific settings
	RedisKeyPrefix  string        `mapstructure:"redis_key_prefix"`
	RedisLockTTL    time.Duration `mapstructure:"redis_lock_ttl"`
	EventBufferSize int           `mapstructure:"event_buffer_size"`

	// Backward compatibility
	LegacyFallback       bool `mapstructure:"legacy_fallback"`
	PreserveLegacyFormat bool `mapstructure:"preserve_legacy_format"`
}

// DefaultConfig returns default configuration for enhanced locking
func DefaultConfig() *EnhancedConfig {
	return &EnhancedConfig{
		Enabled:                 false, // Opt-in for backward compatibility
		Backend:                 "boltdb",
		DefaultTimeout:          30 * time.Second,
		MaxTimeout:              5 * time.Minute,
		EnablePriorityQueue:     false,
		EnableRetries:           false,
		EnableMetrics:           true,
		EnableEvents:            false,
		MaxQueueSize:            1000,
		MaxRetryAttempts:        3,
		QueueTimeout:            5 * time.Minute,
		EnableRetry:             false,
		RetryBaseDelay:          1 * time.Second,
		RetryMaxDelay:           30 * time.Second,
		StarvationThreshold:     2 * time.Minute,
		MaxPriorityBoost:        3,
		EnableDeadlockDetection: false,
		DeadlockCheckInterval:   30 * time.Second,
		RedisKeyPrefix:          "atlantis:locks:",
		RedisLockTTL:            24 * time.Hour,
		EventBufferSize:         100,
		LegacyFallback:          true,
		PreserveLegacyFormat:    true,
	}
}

// Validate validates the configuration
func (c *EnhancedConfig) Validate() error {
	// Validate backend
	if c.Backend != "boltdb" && c.Backend != "redis" {
		return &ConfigError{
			Field: "backend",
			Value: c.Backend,
			Msg:   "must be 'boltdb' or 'redis'",
		}
	}

	// Validate timeouts
	if c.DefaultTimeout <= 0 {
		return &ConfigError{
			Field: "default_timeout",
			Value: c.DefaultTimeout.String(),
			Msg:   "must be positive",
		}
	}

	if c.MaxTimeout <= 0 {
		return &ConfigError{
			Field: "max_timeout",
			Value: c.MaxTimeout.String(),
			Msg:   "must be positive",
		}
	}

	if c.MaxTimeout < c.DefaultTimeout {
		return &ConfigError{
			Field: "max_timeout",
			Value: c.MaxTimeout.String(),
			Msg:   "must be greater than or equal to default_timeout",
		}
	}

	// Validate queue settings
	if c.MaxQueueSize <= 0 {
		return &ConfigError{
			Field: "max_queue_size",
			Value: fmt.Sprintf("%d", c.MaxQueueSize),
			Msg:   "must be positive",
		}
	}

	return nil
}

// RedisConfig holds Redis-specific configuration
type RedisConfig struct {
	Addresses  []string      `mapstructure:"addresses"`
	PoolSize   int           `mapstructure:"pool_size"`
	DefaultTTL time.Duration `mapstructure:"default_ttl"`
	Username   string        `mapstructure:"username"`
	Password   string        `mapstructure:"password"`
	Database   int           `mapstructure:"database"`
}

// Validate validates the Redis configuration
func (c *RedisConfig) Validate() error {
	if len(c.Addresses) == 0 {
		return &ConfigError{
			Field: "addresses",
			Value: "",
			Msg:   "at least one Redis address is required",
		}
	}

	if c.PoolSize <= 0 {
		return &ConfigError{
			Field: "pool_size",
			Value: fmt.Sprintf("%d", c.PoolSize),
			Msg:   "must be positive",
		}
	}

	if c.DefaultTTL <= 0 {
		return &ConfigError{
			Field: "default_ttl",
			Value: c.DefaultTTL.String(),
			Msg:   "must be positive",
		}
	}

	return nil
}
