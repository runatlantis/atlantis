package types

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

// String returns the string representation of Priority
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

// ParsePriority parses a string into a Priority
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
			Msg:   "invalid priority value",
		}
	}
}

// IsValidPriority checks if a priority value is valid
func IsValidPriority(p Priority) bool {
	return p >= PriorityLow && p <= PriorityCritical
}

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

// GetID returns the request ID (interface method)
func (r *EnhancedLockRequest) GetID() string {
	return r.ID
}

// GetOwner returns the request owner (interface method)
func (r *EnhancedLockRequest) GetOwner() string {
	return r.User.Username
}

// GetPriority returns the request priority (interface method)
func (r *EnhancedLockRequest) GetPriority() Priority {
	return r.Priority
}

// GetResource returns the resource identifier (interface method)
func (r *EnhancedLockRequest) GetResource() ResourceIdentifier {
	return r.Resource
}

// GetRequestedAt returns when the request was made (interface method)
func (r *EnhancedLockRequest) GetRequestedAt() time.Time {
	return r.RequestedAt
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

// GetID returns the lock ID (interface method)
func (l *EnhancedLock) GetID() string {
	return l.ID
}

// GetOwner returns the lock owner (interface method)
func (l *EnhancedLock) GetOwner() string {
	return l.Owner
}

// GetPriority returns the lock priority (interface method)
func (l *EnhancedLock) GetPriority() Priority {
	return l.Priority
}

// LockError represents errors in the locking system
type LockError struct {
	Type    string `json:"type"`
	Message string `json:"message"`
	Code    string `json:"code"`
}

func (e *LockError) Error() string {
	return fmt.Sprintf("lock error [%s]: %s", e.Code, e.Message)
}

// ConfigError represents configuration validation errors
type ConfigError struct {
	Field string `json:"field"`
	Value string `json:"value"`
	Msg   string `json:"message"`
}

func (e *ConfigError) Error() string {
	return fmt.Sprintf("config error in field '%s' with value '%s': %s", e.Field, e.Value, e.Msg)
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