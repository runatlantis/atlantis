package enhanced

import (
	"context"
	"time"

	"github.com/runatlantis/atlantis/server/core/locking/types"
	"github.com/runatlantis/atlantis/server/events/models"
)

// Type aliases for backward compatibility and shared types
type Priority = types.Priority
type LockState = types.LockState
type ResourceType = types.ResourceType
type ResourceIdentifier = types.ResourceIdentifier
type EnhancedLockRequest = types.EnhancedLockRequest
type EnhancedLock = types.EnhancedLock
type LockError = types.LockError
type ConfigError = types.ConfigError

// Re-export constants
const (
	PriorityLow      = types.PriorityLow
	PriorityNormal   = types.PriorityNormal
	PriorityHigh     = types.PriorityHigh
	PriorityCritical = types.PriorityCritical

	LockStateAcquired = types.LockStateAcquired
	LockStatePending  = types.LockStatePending
	LockStateExpired  = types.LockStateExpired
	LockStateReleased = types.LockStateReleased

	ResourceTypeProject   = types.ResourceTypeProject
	ResourceTypeWorkspace = types.ResourceTypeWorkspace
	ResourceTypeGlobal    = types.ResourceTypeGlobal
	ResourceTypeCustom    = types.ResourceTypeCustom

	ErrCodeLockExists       = types.ErrCodeLockExists
	ErrCodeLockNotFound     = types.ErrCodeLockNotFound
	ErrCodeLockExpired      = types.ErrCodeLockExpired
	ErrCodeTimeout          = types.ErrCodeTimeout
	ErrCodeQueueFull        = types.ErrCodeQueueFull
	ErrCodeDeadlock         = types.ErrCodeDeadlock
	ErrCodeBackendError     = types.ErrCodeBackendError
	ErrCodeInvalidRequest   = types.ErrCodeInvalidRequest
	ErrCodePermissionDenied = types.ErrCodePermissionDenied
)

// Note: Interface methods (GetID, GetOwner, GetPriority) are defined in the shared types package

// LockEvent represents events in the locking system
type LockEvent struct {
	Type      string             `json:"type"`
	LockID    string             `json:"lock_id"`
	Resource  ResourceIdentifier `json:"resource"`
	Owner     string             `json:"owner"`
	Timestamp time.Time          `json:"timestamp"`
	Metadata  map[string]string  `json:"metadata,omitempty"`
}

// EnhancedConfig and DefaultConfig are defined in config.go to avoid duplication

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
	Size             int                    `json:"size"`
	PendingRequests  []*EnhancedLockRequest `json:"pending_requests"`
	OldestRequest    *time.Time             `json:"oldest_request,omitempty"`
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

// Re-export helper functions from shared types
var (
	NewLockExistsError   = types.NewLockExistsError
	NewLockNotFoundError = types.NewLockNotFoundError
	NewTimeoutError      = types.NewTimeoutError
	ParsePriority        = types.ParsePriority
	IsValidPriority      = types.IsValidPriority
)
