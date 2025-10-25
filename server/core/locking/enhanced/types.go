package enhanced

import (
	"context"
	"fmt"
	"time"

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

// ResourceIdentifier uniquely identifies a lockable resource
type ResourceIdentifier struct {
	Type      ResourceType `json:"type"`
	Namespace string       `json:"namespace"` // Repository namespace
	Name      string       `json:"name"`      // Resource name
	Workspace string       `json:"workspace,omitempty"`
	Path      string       `json:"path,omitempty"`
}

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

// LockManager interface for enhanced lock management
// This is the foundation interface that must be implemented
type LockManager interface {
	// Core operations (backward compatible)
	Lock(ctx context.Context, project models.Project, workspace string, user models.User) (*models.ProjectLock, error)
	Unlock(ctx context.Context, project models.Project, workspace string, user models.User) (*models.ProjectLock, error)
	List(ctx context.Context) ([]*models.ProjectLock, error)
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