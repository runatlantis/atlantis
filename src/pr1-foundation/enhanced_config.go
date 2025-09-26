// PR #1: Enhanced Locking Foundation - Configuration Infrastructure
// This file implements the enhanced locking configuration system
// that will be integrated into server/user_config.go

package server

import (
	"errors"
	"fmt"
	"time"

	"github.com/runatlantis/atlantis/server/core/config/valid"
)

// EnhancedLockingConfig represents the configuration for the enhanced locking system
type EnhancedLockingConfig struct {
	// Global enable/disable flag for enhanced locking
	Enabled bool `mapstructure:"enabled" json:"enabled"`
	
	// Backend type: "boltdb" (legacy) or "redis" (enhanced)
	Backend string `mapstructure:"backend" json:"backend"`
	
	// Redis-specific configuration
	Redis RedisLockingConfig `mapstructure:"redis" json:"redis"`
	
	// Advanced features configuration
	Features LockingFeaturesConfig `mapstructure:"features" json:"features"`
	
	// Fallback and compatibility settings
	Fallback FallbackConfig `mapstructure:"fallback" json:"fallback"`
	
	// Performance and monitoring settings
	Performance PerformanceConfig `mapstructure:"performance" json:"performance"`
}

// RedisLockingConfig contains Redis-specific settings
type RedisLockingConfig struct {
	// Redis server endpoints
	Addresses []string `mapstructure:"addresses" json:"addresses"`
	
	// Redis password (optional)
	Password string `mapstructure:"password" json:"password"`
	
	// Database number
	DB int `mapstructure:"db" json:"db"`
	
	// Connection pool settings
	PoolSize int `mapstructure:"pool_size" json:"pool_size"`
	
	// Key prefix for all lock keys
	KeyPrefix string `mapstructure:"key_prefix" json:"key_prefix"`
	
	// Default lock TTL
	LockTTL time.Duration `mapstructure:"lock_ttl" json:"lock_ttl"`
	
	// Connection timeout
	ConnTimeout time.Duration `mapstructure:"conn_timeout" json:"conn_timeout"`
	
	// Read timeout
	ReadTimeout time.Duration `mapstructure:"read_timeout" json:"read_timeout"`
	
	// Write timeout
	WriteTimeout time.Duration `mapstructure:"write_timeout" json:"write_timeout"`
	
	// Enable cluster mode
	ClusterMode bool `mapstructure:"cluster_mode" json:"cluster_mode"`
	
	// TLS configuration
	TLS TLSConfig `mapstructure:"tls" json:"tls"`
}

// LockingFeaturesConfig controls advanced features
type LockingFeaturesConfig struct {
	// Enable priority queue for lock acquisition
	PriorityQueue bool `mapstructure:"priority_queue" json:"priority_queue"`
	
	// Enable deadlock detection
	DeadlockDetection bool `mapstructure:"deadlock_detection" json:"deadlock_detection"`
	
	// Enable automatic retry mechanism
	RetryMechanism bool `mapstructure:"retry_mechanism" json:"retry_mechanism"`
	
	// Enable lock queue monitoring
	QueueMonitoring bool `mapstructure:"queue_monitoring" json:"queue_monitoring"`
	
	// Enable event streaming
	EventStreaming bool `mapstructure:"event_streaming" json:"event_streaming"`
	
	// Enable distributed tracing
	DistributedTracing bool `mapstructure:"distributed_tracing" json:"distributed_tracing"`
}

// FallbackConfig controls legacy compatibility
type FallbackConfig struct {
	// Keep legacy system as fallback
	LegacyEnabled bool `mapstructure:"legacy_enabled" json:"legacy_enabled"`
	
	// Preserve legacy data format
	PreserveFormat bool `mapstructure:"preserve_format" json:"preserve_format"`
	
	// Auto-fallback on Redis failure
	AutoFallback bool `mapstructure:"auto_fallback" json:"auto_fallback"`
	
	// Fallback timeout
	FallbackTimeout time.Duration `mapstructure:"fallback_timeout" json:"fallback_timeout"`
}

// PerformanceConfig controls performance-related settings
type PerformanceConfig struct {
	// Maximum concurrent locks
	MaxConcurrentLocks int `mapstructure:"max_concurrent_locks" json:"max_concurrent_locks"`
	
	// Queue processing batch size
	QueueBatchSize int `mapstructure:"queue_batch_size" json:"queue_batch_size"`
	
	// Lock acquisition timeout
	AcquisitionTimeout time.Duration `mapstructure:"acquisition_timeout" json:"acquisition_timeout"`
	
	// Health check interval
	HealthCheckInterval time.Duration `mapstructure:"health_check_interval" json:"health_check_interval"`
	
	// Metrics collection interval
	MetricsInterval time.Duration `mapstructure:"metrics_interval" json:"metrics_interval"`
}

// TLSConfig for Redis connections
type TLSConfig struct {
	Enabled bool `mapstructure:"enabled" json:"enabled"`
	CertFile string `mapstructure:"cert_file" json:"cert_file"`
	KeyFile string `mapstructure:"key_file" json:"key_file"`
	CAFile string `mapstructure:"ca_file" json:"ca_file"`
	SkipVerify bool `mapstructure:"skip_verify" json:"skip_verify"`
}

// DefaultEnhancedLockingConfig returns the default configuration
func DefaultEnhancedLockingConfig() EnhancedLockingConfig {
	return EnhancedLockingConfig{
		Enabled: false, // Start disabled for safety
		Backend: "boltdb", // Default to legacy backend
		Redis: RedisLockingConfig{
			Addresses: []string{"localhost:6379"},
			Password: "",
			DB: 0,
			PoolSize: 10,
			KeyPrefix: "atlantis:enhanced:lock:",
			LockTTL: time.Hour,
			ConnTimeout: 5 * time.Second,
			ReadTimeout: 3 * time.Second,
			WriteTimeout: 3 * time.Second,
			ClusterMode: false,
		},
		Features: LockingFeaturesConfig{
			PriorityQueue: false,
			DeadlockDetection: false,
			RetryMechanism: false,
			QueueMonitoring: false,
			EventStreaming: false,
			DistributedTracing: false,
		},
		Fallback: FallbackConfig{
			LegacyEnabled: true,
			PreserveFormat: true,
			AutoFallback: true,
			FallbackTimeout: 10 * time.Second,
		},
		Performance: PerformanceConfig{
			MaxConcurrentLocks: 1000,
			QueueBatchSize: 100,
			AcquisitionTimeout: 30 * time.Second,
			HealthCheckInterval: 30 * time.Second,
			MetricsInterval: 15 * time.Second,
		},
	}
}

// Validate validates the enhanced locking configuration
func (c *EnhancedLockingConfig) Validate() error {
	if c.Backend != "boltdb" && c.Backend != "redis" {
		return fmt.Errorf("invalid backend: %s. Must be 'boltdb' or 'redis'", c.Backend)
	}
	
	if c.Backend == "redis" {
		if len(c.Redis.Addresses) == 0 {
			return errors.New("redis addresses cannot be empty when using redis backend")
		}
		
		if c.Redis.PoolSize <= 0 {
			return errors.New("redis pool size must be positive")
		}
		
		if c.Redis.LockTTL <= 0 {
			return errors.New("redis lock TTL must be positive")
		}
	}
	
	if c.Performance.MaxConcurrentLocks <= 0 {
		return errors.New("max concurrent locks must be positive")
	}
	
	if c.Performance.AcquisitionTimeout <= 0 {
		return errors.New("acquisition timeout must be positive")
	}
	
	return nil
}

// ToUserConfig integrates with existing UserConfig structure
func (c *EnhancedLockingConfig) ToUserConfig() *valid.LockingConfig {
	// Convert enhanced config to legacy format for backward compatibility
	return &valid.LockingConfig{
		// Map enhanced settings to legacy structure
	}
}

// Feature flag helpers
func (c *EnhancedLockingConfig) IsEnhancedEnabled() bool {
	return c.Enabled
}

func (c *EnhancedLockingConfig) IsRedisBackend() bool {
	return c.Backend == "redis"
}

func (c *EnhancedLockingConfig) ShouldUsePriorityQueue() bool {
	return c.Features.PriorityQueue
}

func (c *EnhancedLockingConfig) ShouldDetectDeadlocks() bool {
	return c.Features.DeadlockDetection
}

func (c *EnhancedLockingConfig) ShouldEnableRetry() bool {
	return c.Features.RetryMechanism
}
