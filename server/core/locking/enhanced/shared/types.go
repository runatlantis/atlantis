package shared

import (
	"context"
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
	ID          string             `json:"id"`
	Resource    ResourceIdentifier `json:"resource"`
	Priority    Priority           `json:"priority"`
	Timeout     time.Duration      `json:"timeout"`
	Metadata    map[string]string  `json:"metadata,omitempty"`
	Context     context.Context    `json:"-"`
	RequestedAt time.Time          `json:"requested_at"`

	// Backward compatibility fields
	Project   models.Project `json:"project"`
	Workspace string         `json:"workspace"`
	User      models.User    `json:"user"`
}

// GetID returns the request ID
func (r *EnhancedLockRequest) GetID() string {
	return r.ID
}

// GetOwner returns the request owner
func (r *EnhancedLockRequest) GetOwner() string {
	return r.User.Username
}

// GetPriority returns the request priority
func (r *EnhancedLockRequest) GetPriority() Priority {
	return r.Priority
}

// GetResource returns the resource identifier
func (r *EnhancedLockRequest) GetResource() ResourceIdentifier {
	return r.Resource
}

// GetRequestedAt returns when the request was made
func (r *EnhancedLockRequest) GetRequestedAt() time.Time {
	return r.RequestedAt
}

// EnhancedLock represents an acquired lock with enhanced capabilities
type EnhancedLock struct {
	ID         string             `json:"id"`
	Resource   ResourceIdentifier `json:"resource"`
	State      LockState          `json:"state"`
	Priority   Priority           `json:"priority"`
	Owner      string             `json:"owner"`
	AcquiredAt time.Time          `json:"acquired_at"`
	ExpiresAt  *time.Time         `json:"expires_at,omitempty"`
	Metadata   map[string]string  `json:"metadata,omitempty"`
	Version    int64              `json:"version"` // For optimistic locking

	// Backward compatibility - embed original lock
	OriginalLock *models.ProjectLock `json:"original_lock,omitempty"`
}

// GetID returns the lock ID
func (l *EnhancedLock) GetID() string {
	return l.ID
}

// GetOwner returns the lock owner
func (l *EnhancedLock) GetOwner() string {
	return l.Owner
}

// GetPriority returns the lock priority
func (l *EnhancedLock) GetPriority() Priority {
	return l.Priority
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
