package command

import (
	"time"
)

// LockMetadata contains additional data provided to the lock
type LockMetadata struct {
	UnixTime int64
}

// Lock represents a global lock for an atlantis command (plan, apply, policy_check).
// It is used to prevent commands from being executed
type Lock struct {
	// Time is the time at which the lock was first created.
	LockMetadata LockMetadata
	CommandName  Name
}

func (l *Lock) LockTime() time.Time {
	return time.Unix(l.LockMetadata.UnixTime, 0)
}

func (l *Lock) IsLocked() bool {
	return !l.LockTime().IsZero()
}
