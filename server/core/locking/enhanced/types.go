package enhanced

import (
	"context"
	"fmt"
	"time"

	"github.com/runatlantis/atlantis/server/core/locking"
	"github.com/runatlantis/atlantis/server/events/models"
)

// Priority levels for enhanced locking
type Priority int

const (
	PriorityLow Priority = iota
	PriorityNormal
	PriorityHigh
	PriorityCritical
)

// LockState represents the current state of a lock
type LockState string

const (
	LockStateAcquired LockState = "acquired"
	LockStatePending  LockState = "pending"
	LockStateExpired  LockState = "expired"
	LockStateReleased LockState = "released"
)

// ResourceType defines the type of resource being locked
type ResourceType string

const (
	ResourceTypeProject   ResourceType = "project"
	ResourceTypeWorkspace ResourceType = "workspace"
	ResourceTypeGlobal    ResourceType = "global"
	ResourceTypeCustom    ResourceType = "custom"
)

// EnhancedLockRequest represents a request for an enhanced lock
type EnhancedLockRequest struct {
	ID          string                 `json:"id"`
	Resource    ResourceIdentifier     `json:"resource"`
	Priority    Priority               `json:"priority"`
	Timeout     time.Duration          `json:"timeout"`
	Metadata    map[string]string      `json:"metadata,omitempty"`
	Context     context.Context        `json:"-"`
	RequestedAt time.Time              `json:"requested_at"`

	// Backward compatibility fields
	Project   models.Project `json:"project"`
	Workspace string         `json:"workspace"`
	User      models.User    `json:"user"`
}

// ResourceIdentifier uniquely identifies a lockable resource
type ResourceIdentifier struct {
	Type      ResourceType `json:"type"`
	Namespace string       `json:"namespace"` // Repository namespace
	Name      string       `json:"name"`      // Resource name
	Workspace string       `json:"workspace,omitempty"`
	Path      string       `json:"path,omitempty"`
}

// EnhancedLock represents an acquired lock with enhanced capabilities
type EnhancedLock struct {
	ID          string             `json:"id"`
	Resource    ResourceIdentifier `json:"resource"`
	State       LockState          `json:"state"`
	Priority    Priority           `json:"priority"`
	Owner       string             `json:"owner"`
	AcquiredAt  time.Time          `json:"acquired_at"`
	ExpiresAt   *time.Time         `json:"expires_at,omitempty"`
	Metadata    map[string]string  `json:"metadata,omitempty"`
	Version     int64              `json:"version"` // For optimistic locking

	// Backward compatibility - embed original lock
	OriginalLock *models.ProjectLock `json:"original_lock,omitempty"`
}

// LockEvent represents events in the locking system
type LockEvent struct {
	Type      string             `json:"type"`
	LockID    string             `json:"lock_id"`
	Resource  ResourceIdentifier `json:"resource"`
	Owner     string             `json:"owner"`
	Timestamp time.Time          `json:"timestamp"`
	Metadata  map[string]string  `json:"metadata,omitempty"`
}

// Config holds configuration for enhanced locking
type EnhancedConfig struct {
	// Core configuration
	Enabled                bool          `mapstructure:"enabled"`
	Backend                string        `mapstructure:"backend"` // "redis", "boltdb"
	DefaultTimeout         time.Duration `mapstructure:"default_timeout"`
	MaxTimeout             time.Duration `mapstructure:"max_timeout"`

	// Queue configuration
	EnablePriorityQueue    bool          `mapstructure:"enable_priority_queue"`
	MaxQueueSize          int           `mapstructure:"max_queue_size"`
	QueueTimeout          time.Duration `mapstructure:"queue_timeout"`

	// Retry configuration
	EnableRetry           bool          `mapstructure:"enable_retry"`
	MaxRetryAttempts      int           `mapstructure:"max_retry_attempts"`
	RetryBaseDelay        time.Duration `mapstructure:"retry_base_delay"`
	RetryMaxDelay         time.Duration `mapstructure:"retry_max_delay"`

	// Deadlock detection
	EnableDeadlockDetection bool        `mapstructure:"enable_deadlock_detection"`
	DeadlockCheckInterval   time.Duration `mapstructure:"deadlock_check_interval"`

	// Events
	EnableEvents          bool          `mapstructure:"enable_events"`
	EventBufferSize       int           `mapstructure:"event_buffer_size"`

	// Redis specific
	RedisClusterMode      bool          `mapstructure:"redis_cluster_mode"`
	RedisKeyPrefix        string        `mapstructure:"redis_key_prefix"`
	RedisLockTTL          time.Duration `mapstructure:"redis_lock_ttl"`

	// Backward compatibility
	LegacyFallback        bool          `mapstructure:"legacy_fallback"`
	PreserveLegacyFormat  bool          `mapstructure:"preserve_legacy_format"`
}

// DefaultConfig returns default configuration for enhanced locking
func DefaultConfig() *EnhancedConfig {
	return &EnhancedConfig{
		Enabled:                 false, // Opt-in for backward compatibility
		Backend:                 "boltdb",
		DefaultTimeout:          30 * time.Minute,
		MaxTimeout:              2 * time.Hour,
		EnablePriorityQueue:     false,
		MaxQueueSize:           1000,
		QueueTimeout:           10 * time.Minute,
		EnableRetry:            false,
		MaxRetryAttempts:       3,
		RetryBaseDelay:         time.Second,
		RetryMaxDelay:          30 * time.Second,
		EnableDeadlockDetection: false,
		DeadlockCheckInterval:   30 * time.Second,
		EnableEvents:           false,
		EventBufferSize:        1000,
		RedisClusterMode:       false,
		RedisKeyPrefix:         "atlantis:enhanced:lock:",
		RedisLockTTL:           time.Hour,
		LegacyFallback:         true,
		PreserveLegacyFormat:   true,
	}
}

// Backend interface for enhanced locking backends
type Backend interface {
	// Core locking operations
	AcquireLock(ctx context.Context, request *EnhancedLockRequest) (*EnhancedLock, error)
	ReleaseLock(ctx context.Context, lockID string) error
	GetLock(ctx context.Context, lockID string) (*EnhancedLock, error)
	ListLocks(ctx context.Context) ([]*EnhancedLock, error)

	// Enhanced operations
	TryAcquireLock(ctx context.Context, request *EnhancedLockRequest) (*EnhancedLock, bool, error)
	RefreshLock(ctx context.Context, lockID string, extension time.Duration) error
	TransferLock(ctx context.Context, lockID string, newOwner string) error

	// Queue operations
	EnqueueLockRequest(ctx context.Context, request *EnhancedLockRequest) error
	DequeueNextRequest(ctx context.Context) (*EnhancedLockRequest, error)
	GetQueueStatus(ctx context.Context) (*QueueStatus, error)

	// Health and metrics
	HealthCheck(ctx context.Context) error
	GetStats(ctx context.Context) (*BackendStats, error)

	// Event subscription
	Subscribe(ctx context.Context, eventTypes []string) (<-chan *LockEvent, error)

	// Cleanup and maintenance
	CleanupExpiredLocks(ctx context.Context) (int, error)

	// Backward compatibility
	GetLegacyLock(project models.Project, workspace string) (*models.ProjectLock, error)
	ConvertToLegacy(lock *EnhancedLock) *models.ProjectLock
	ConvertFromLegacy(lock *models.ProjectLock) *EnhancedLock
}

// QueueStatus provides information about the current queue state
type QueueStatus struct {
	Size            int                     `json:"size"`
	PendingRequests []*EnhancedLockRequest  `json:"pending_requests"`
	OldestRequest   *time.Time              `json:"oldest_request,omitempty"`
	QueuesByPriority map[Priority]int       `json:"queues_by_priority"`
}

// BackendStats provides performance and operational statistics
type BackendStats struct {
	ActiveLocks        int64         `json:"active_locks"`
	TotalRequests      int64         `json:"total_requests"`
	SuccessfulAcquires int64         `json:"successful_acquires"`
	FailedAcquires     int64         `json:"failed_acquires"`
	AverageWaitTime    time.Duration `json:"average_wait_time"`
	AverageHoldTime    time.Duration `json:"average_hold_time"`
	QueueDepth         int           `json:"queue_depth"`
	HealthScore        int           `json:"health_score"` // 0-100
	LastUpdated        time.Time     `json:"last_updated"`
}

// LockManager interface for enhanced lock management
type LockManager interface {
	// Core operations
	Lock(ctx context.Context, project models.Project, workspace string, user models.User) (*models.ProjectLock, error)
	Unlock(ctx context.Context, project models.Project, workspace string, user models.User) (*models.ProjectLock, error)
	List(ctx context.Context) ([]*models.ProjectLock, error)

	// Enhanced operations with priority
	LockWithPriority(ctx context.Context, project models.Project, workspace string, user models.User, priority Priority) (*models.ProjectLock, error)
	LockWithTimeout(ctx context.Context, project models.Project, workspace string, user models.User, timeout time.Duration) (*models.ProjectLock, error)

	// Queue management
	GetQueuePosition(ctx context.Context, project models.Project, workspace string) (int, error)
	CancelQueuedRequest(ctx context.Context, project models.Project, workspace string, user models.User) error

	// Statistics and monitoring
	GetStats(ctx context.Context) (*BackendStats, error)
	GetHealth(ctx context.Context) error
}

// Error types for enhanced locking
type LockError struct {
	Type    string `json:"type"`
	Message string `json:"message"`
	Code    string `json:"code"`
}

func (e *LockError) Error() string {
	return fmt.Sprintf("lock error [%s]: %s", e.Code, e.Message)
}

// Common error codes
const (
	ErrCodeLockExists       = "LOCK_EXISTS"
	ErrCodeLockNotFound     = "LOCK_NOT_FOUND"
	ErrCodeLockExpired      = "LOCK_EXPIRED"
	ErrCodeTimeout          = "TIMEOUT"
	ErrCodeQueueFull        = "QUEUE_FULL"
	ErrCodeDeadlock         = "DEADLOCK"
	ErrCodeBackendError     = "BACKEND_ERROR"
	ErrCodeInvalidRequest   = "INVALID_REQUEST"
	ErrCodePermissionDenied = "PERMISSION_DENIED"
)

// Helper functions for creating common errors
func NewLockExistsError(resource string) *LockError {
	return &LockError{
		Type:    "LockExists",
		Message: fmt.Sprintf("lock already exists for resource: %s", resource),
		Code:    ErrCodeLockExists,
	}
}

func NewLockNotFoundError(lockID string) *LockError {
	return &LockError{
		Type:    "LockNotFound",
		Message: fmt.Sprintf("lock not found: %s", lockID),
		Code:    ErrCodeLockNotFound,
	}
}

func NewTimeoutError(timeout time.Duration) *LockError {
	return &LockError{
		Type:    "Timeout",
		Message: fmt.Sprintf("operation timed out after %v", timeout),
		Code:    ErrCodeTimeout,
	}
}