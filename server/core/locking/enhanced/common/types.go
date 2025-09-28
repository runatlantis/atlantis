package common

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

// LockEvent represents events in the locking system
type LockEvent struct {
	Type      string             `json:"type"`
	LockID    string             `json:"lock_id"`
	Resource  ResourceIdentifier `json:"resource"`
	Owner     string             `json:"owner"`
	Timestamp time.Time          `json:"timestamp"`
	Metadata  map[string]string  `json:"metadata,omitempty"`
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
