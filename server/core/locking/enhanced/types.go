// Package enhanced provides the enhanced locking system for Atlantis
// This package implements advanced locking features including Redis backend,
// priority queuing, and distributed coordination.
package enhanced

import (
	"context"
	"time"

	"github.com/runatlantis/atlantis/server/events/models"
)

// Priority levels for lock requests
type Priority int

const (
	PriorityLow Priority = iota
	PriorityNormal
	PriorityHigh
	PriorityCritical
)

// LockRequest represents a request for a project lock with enhanced features
type LockRequest struct {
	// Core lock information
	Project   models.Project
	Workspace string
	User      models.User
	Pull      models.PullRequest

	// Enhanced features
	Priority     Priority
	Timeout      time.Duration
	Tags         map[string]string
	RequestID    string
	RequestTime  time.Time

	// Retry configuration
	MaxRetries   int
	RetryBackoff time.Duration
}

// LockResponse represents the response to a lock request
type LockResponse struct {
	// Success information
	Acquired     bool
	Lock         *EnhancedLock
	LockKey      string

	// Queue information
	QueuePosition int
	EstimatedWait time.Duration

	// Error information
	Error        error
	ErrorCode    string
	Reason       string
}

// EnhancedLock represents an enhanced project lock with additional metadata
type EnhancedLock struct {
	// Legacy compatibility
	models.ProjectLock

	// Enhanced features
	ID           string
	Priority     Priority
	Tags         map[string]string
	TTL          time.Duration
	ExpiresAt    time.Time

	// Queue information
	QueueTime    time.Duration
	AcquireTime  time.Time

	// Distributed features
	NodeID       string
	Version      int64
	Checksum     string
}

// QueueEntry represents an entry in the lock queue
type QueueEntry struct {
	Request      LockRequest
	Position     int
	EnqueueTime  time.Time
	EstimatedWait time.Duration

	// Priority boosting for anti-starvation
	OriginalPriority Priority
	CurrentPriority  Priority
	BoostCount       int
}

// LockEvent represents events in the locking system
type LockEvent struct {
	Type        EventType
	LockID      string
	LockKey     string
	User        string
	Timestamp   time.Time
	Details     map[string]interface{}
}

// EventType represents the type of lock event
type EventType string

const (
	EventLockRequested  EventType = "lock_requested"
	EventLockAcquired   EventType = "lock_acquired"
	EventLockReleased   EventType = "lock_released"
	EventLockExpired    EventType = "lock_expired"
	EventLockQueued     EventType = "lock_queued"
	EventLockTimeout    EventType = "lock_timeout"
	EventLockError      EventType = "lock_error"
)

// Backend represents the enhanced locking backend interface
type Backend interface {
	// Core locking operations
	TryLock(ctx context.Context, request LockRequest) (*LockResponse, error)
	Unlock(ctx context.Context, lockKey string) error
	GetLock(ctx context.Context, lockKey string) (*EnhancedLock, error)
	ListLocks(ctx context.Context) ([]EnhancedLock, error)

	// Queue operations
	GetQueueStatus(ctx context.Context, lockKey string) (*QueueStatus, error)
	GetUserQueue(ctx context.Context, user string) ([]QueueEntry, error)

	// Advanced features
	RefreshLock(ctx context.Context, lockKey string, ttl time.Duration) error
	TransferLock(ctx context.Context, lockKey string, newUser models.User) error

	// Batch operations
	BatchUnlock(ctx context.Context, lockKeys []string) error
	BatchGetLocks(ctx context.Context, lockKeys []string) ([]EnhancedLock, error)

	// Monitoring and health
	HealthCheck(ctx context.Context) error
	GetMetrics(ctx context.Context) (*Metrics, error)
}

// QueueStatus represents the status of the lock queue
type QueueStatus struct {
	LockKey       string
	QueueLength   int
	AverageWait   time.Duration
	CurrentHolder *EnhancedLock
	NextInQueue   *QueueEntry
}

// Metrics represents locking system metrics
type Metrics struct {
	TotalLocks        int64
	ActiveLocks       int64
	QueuedRequests    int64
	AverageWaitTime   time.Duration
	LockSuccessRate   float64
	BackendLatency    time.Duration
	LastUpdated       time.Time
}

// EnhancedLocker is the main interface for the enhanced locking system
type EnhancedLocker interface {
	// Core operations
	Lock(ctx context.Context, request LockRequest) (*LockResponse, error)
	Unlock(ctx context.Context, lockKey string) error
	GetLock(ctx context.Context, lockKey string) (*EnhancedLock, error)

	// Queue management
	GetQueueStatus(ctx context.Context, lockKey string) (*QueueStatus, error)
	CancelQueuedRequest(ctx context.Context, requestID string) error

	// Batch operations
	UnlockByPull(ctx context.Context, repoFullName string, pullNum int) ([]EnhancedLock, error)

	// Administrative operations
	ForceUnlock(ctx context.Context, lockKey string, reason string) error
	GetMetrics(ctx context.Context) (*Metrics, error)
	HealthCheck(ctx context.Context) error
}

// Config represents the enhanced locking configuration
type Config struct {
	// Backend configuration
	Backend     string        `mapstructure:"backend" json:"backend"`
	RedisConfig RedisConfig   `mapstructure:"redis" json:"redis"`

	// Feature flags
	EnablePriorityQueue bool `mapstructure:"enable_priority_queue" json:"enable_priority_queue"`
	EnableRetries       bool `mapstructure:"enable_retries" json:"enable_retries"`
	EnableMetrics       bool `mapstructure:"enable_metrics" json:"enable_metrics"`

	// Timeouts and limits
	DefaultTimeout      time.Duration `mapstructure:"default_timeout" json:"default_timeout"`
	MaxTimeout          time.Duration `mapstructure:"max_timeout" json:"max_timeout"`
	MaxQueueSize        int           `mapstructure:"max_queue_size" json:"max_queue_size"`

	// Anti-starvation
	StarvationThreshold time.Duration `mapstructure:"starvation_threshold" json:"starvation_threshold"`
	MaxPriorityBoost    int           `mapstructure:"max_priority_boost" json:"max_priority_boost"`
}

// RedisConfig represents Redis-specific configuration
type RedisConfig struct {
	Addresses        []string      `mapstructure:"addresses" json:"addresses"`
	Password         string        `mapstructure:"password" json:"password"`
	DB               int           `mapstructure:"db" json:"db"`
	PoolSize         int           `mapstructure:"pool_size" json:"pool_size"`
	KeyPrefix        string        `mapstructure:"key_prefix" json:"key_prefix"`
	DefaultTTL       time.Duration `mapstructure:"default_ttl" json:"default_ttl"`
	ConnectionTimeout time.Duration `mapstructure:"connection_timeout" json:"connection_timeout"`
	ReadTimeout      time.Duration `mapstructure:"read_timeout" json:"read_timeout"`
	WriteTimeout     time.Duration `mapstructure:"write_timeout" json:"write_timeout"`
	ClusterMode      bool          `mapstructure:"cluster_mode" json:"cluster_mode"`
}

// EventHandler handles lock events
type EventHandler interface {
	HandleEvent(ctx context.Context, event LockEvent) error
}

// LockEventStream provides real-time lock events
type LockEventStream interface {
	Subscribe(ctx context.Context, filter EventFilter) (<-chan LockEvent, error)
	Unsubscribe(subscriptionID string) error
}

// EventFilter filters lock events
type EventFilter struct {
	EventTypes []EventType
	Users      []string
	Projects   []string
	LockKeys   []string
}

// DefaultConfig returns the default enhanced locking configuration
func DefaultConfig() *Config {
	return &Config{
		Backend: "boltdb", // Start with legacy backend for safety
		RedisConfig: RedisConfig{
			Addresses:         []string{"localhost:6379"},
			Password:          "",
			DB:                0,
			PoolSize:          10,
			KeyPrefix:         "atlantis:enhanced:lock:",
			DefaultTTL:        time.Hour,
			ConnectionTimeout: 5 * time.Second,
			ReadTimeout:       3 * time.Second,
			WriteTimeout:      3 * time.Second,
			ClusterMode:       false,
		},
		EnablePriorityQueue:  false,
		EnableRetries:        false,
		EnableMetrics:        true,
		DefaultTimeout:       30 * time.Second,
		MaxTimeout:           5 * time.Minute,
		MaxQueueSize:         1000,
		StarvationThreshold:  2 * time.Minute,
		MaxPriorityBoost:     3,
	}
}

// Validate validates the enhanced locking configuration
func (c *Config) Validate() error {
	if c.Backend != "boltdb" && c.Backend != "redis" {
		return &ConfigError{
			Field: "backend",
			Value: c.Backend,
			Msg:   "must be 'boltdb' or 'redis'",
		}
	}

	if c.DefaultTimeout <= 0 {
		return &ConfigError{
			Field: "default_timeout",
			Value: c.DefaultTimeout,
			Msg:   "must be positive",
		}
	}

	if c.MaxTimeout < c.DefaultTimeout {
		return &ConfigError{
			Field: "max_timeout",
			Value: c.MaxTimeout,
			Msg:   "must be >= default_timeout",
		}
	}

	if c.MaxQueueSize <= 0 {
		return &ConfigError{
			Field: "max_queue_size",
			Value: c.MaxQueueSize,
			Msg:   "must be positive",
		}
	}

	// Validate Redis config if Redis backend is selected
	if c.Backend == "redis" {
		return c.RedisConfig.Validate()
	}

	return nil
}

// Validate validates Redis configuration
func (r *RedisConfig) Validate() error {
	if len(r.Addresses) == 0 {
		return &ConfigError{
			Field: "redis.addresses",
			Value: r.Addresses,
			Msg:   "cannot be empty",
		}
	}

	if r.PoolSize <= 0 {
		return &ConfigError{
			Field: "redis.pool_size",
			Value: r.PoolSize,
			Msg:   "must be positive",
		}
	}

	if r.DefaultTTL <= 0 {
		return &ConfigError{
			Field: "redis.default_ttl",
			Value: r.DefaultTTL,
			Msg:   "must be positive",
		}
	}

	return nil
}

// ConfigError represents a configuration validation error
type ConfigError struct {
	Field string
	Value interface{}
	Msg   string
}

func (e *ConfigError) Error() string {
	return "config error in field " + e.Field + ": " + e.Msg
}

// GenerateLockKey generates a consistent lock key for a project and workspace
func GenerateLockKey(project models.Project, workspace string) string {
	return models.GenerateLockKey(project, workspace)
}

// IsValidPriority checks if a priority value is valid
func IsValidPriority(p Priority) bool {
	return p >= PriorityLow && p <= PriorityCritical
}

// PriorityString returns the string representation of a priority
func (p Priority) String() string {
	switch p {
	case PriorityLow:
		return "low"
	case PriorityNormal:
		return "normal"
	case PriorityHigh:
		return "high"
	case PriorityCritical:
		return "critical"
	default:
		return "unknown"
	}
}

// ParsePriority parses a priority string
func ParsePriority(s string) (Priority, error) {
	switch s {
	case "low":
		return PriorityLow, nil
	case "normal":
		return PriorityNormal, nil
	case "high":
		return PriorityHigh, nil
	case "critical":
		return PriorityCritical, nil
	default:
		return PriorityNormal, &ConfigError{
			Field: "priority",
			Value: s,
			Msg:   "must be one of: low, normal, high, critical",
		}
	}
}