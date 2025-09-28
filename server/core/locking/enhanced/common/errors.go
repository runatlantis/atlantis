package common

import (
	"fmt"
	"time"
)

// LockError represents errors in the enhanced locking system
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
