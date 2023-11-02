package locking

import (
	"errors"
	"time"

	"github.com/runatlantis/atlantis/server/events/command"
)

//go:generate pegomock generate --package mocks -o mocks/mock_apply_lock_checker.go ApplyLockChecker

// ApplyLockChecker is an implementation of the global apply lock retrieval.
// It returns an object that contains information about apply locks status.
type ApplyLockChecker interface {
	CheckApplyLock() (ApplyCommandLock, error)
}

//go:generate pegomock generate --package mocks -o mocks/mock_apply_locker.go ApplyLocker

// ApplyLocker interface that manages locks for apply command runner
type ApplyLocker interface {
	// LockApply creates a lock for ApplyCommand if lock already exists it will
	// return existing lock without any changes
	LockApply() (ApplyCommandLock, error)
	// UnlockApply deletes apply lock created by LockApply if present, otherwise
	// it is a no-op
	UnlockApply() error
	ApplyLockChecker
}

// ApplyCommandLock contains information about apply command lock status.
type ApplyCommandLock struct {
	// Locked is true is when apply commands are locked
	// Either by using omitting apply from AllowCommands or creating a global ApplyCommandLock
	// DisableApply lock take precedence when set
	Locked  bool
	Time    time.Time
	Failure string
}

type ApplyClient struct {
	backend      Backend
	disableApply bool
}

func NewApplyClient(backend Backend, disableApply bool) ApplyLocker {
	return &ApplyClient{
		backend:      backend,
		disableApply: disableApply,
	}
}

// LockApply acquires global apply lock.
// DisableApply takes presedence to any existing locks, if it is set to true
// this function returns an error
func (c *ApplyClient) LockApply() (ApplyCommandLock, error) {
	response := ApplyCommandLock{}

	if c.disableApply {
		return response, errors.New("apply is omitted from AllowCommands; Apply commands are locked globally until flag is updated")
	}

	applyCmdLock, err := c.backend.LockCommand(command.Apply, time.Now())
	if err != nil {
		return response, err
	}

	if applyCmdLock != nil {
		response.Locked = true
		response.Time = applyCmdLock.LockTime()
	}
	return response, nil
}

// UnlockApply releases a global apply lock.
// DisableApply takes presedence to any existing locks, if it is set to true
// this function returns an error
func (c *ApplyClient) UnlockApply() error {
	if c.disableApply {
		return errors.New("apply commands are disabled until AllowCommands flag is updated")
	}

	err := c.backend.UnlockCommand(command.Apply)
	if err != nil {
		return err
	}

	return nil
}

// CheckApplyLock retrieves an apply command lock if present.
// If DisableApply is set it will always return a lock.
func (c *ApplyClient) CheckApplyLock() (ApplyCommandLock, error) {
	response := ApplyCommandLock{}

	if c.disableApply {
		return ApplyCommandLock{
			Locked: true,
		}, nil
	}

	applyCmdLock, err := c.backend.CheckCommandLock(command.Apply)
	if err != nil {
		return response, err
	}

	if applyCmdLock != nil {
		response.Locked = true
		response.Time = applyCmdLock.LockTime()
	}

	return response, nil
}
