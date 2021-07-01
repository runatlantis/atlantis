package locking

import (
	"errors"
	"time"

	"github.com/runatlantis/atlantis/server/events/models"
)

//go:generate pegomock generate -m --use-experimental-model-gen --package mocks -o mocks/mock_apply_lock_checker.go ApplyLockChecker

// ApplyLockChecker is an implementation of the global apply lock retrieval.
// It returns an object that contains information about apply locks status.
type ApplyLockChecker interface {
	CheckApplyLock() (ApplyCommandLock, error)
}

//go:generate pegomock generate -m --use-experimental-model-gen --package mocks -o mocks/mock_apply_locker.go ApplyLocker

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
	// Either by using DisableApply flag or creating a global ApplyCommandLock
	// DisableApply lock take precedence when set
	Locked  bool
	Time    time.Time
	Failure string
}

type ApplyClient struct {
	backend          Backend
	disableApplyFlag bool
}

func NewApplyClient(backend Backend, disableApplyFlag bool) ApplyLocker {
	return &ApplyClient{
		backend:          backend,
		disableApplyFlag: disableApplyFlag,
	}
}

// LockApply acquires global apply lock.
// DisableApplyFlag takes presedence to any existing locks, if it is set to true
// this function returns an error
func (c *ApplyClient) LockApply() (ApplyCommandLock, error) {
	response := ApplyCommandLock{}

	if c.disableApplyFlag {
		return response, errors.New("DisableApplyFlag is set; Apply commands are locked globally until flag is unset")
	}

	applyCmdLock, err := c.backend.LockCommand(models.ApplyCommand, time.Now())
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
// DisableApplyFlag takes presedence to any existing locks, if it is set to true
// this function returns an error
func (c *ApplyClient) UnlockApply() error {
	if c.disableApplyFlag {
		return errors.New("apply commands are disabled until DisableApply flag is unset")
	}

	err := c.backend.UnlockCommand(models.ApplyCommand)
	if err != nil {
		return err
	}

	return nil
}

// CheckApplyLock retrieves an apply command lock if present.
// If DisableApplyFlag is set it will always return a lock.
func (c *ApplyClient) CheckApplyLock() (ApplyCommandLock, error) {
	response := ApplyCommandLock{}

	if c.disableApplyFlag {
		return ApplyCommandLock{
			Locked: true,
		}, nil
	}

	applyCmdLock, err := c.backend.CheckCommandLock(models.ApplyCommand)
	if err != nil {
		return response, err
	}

	if applyCmdLock != nil {
		response.Locked = true
		response.Time = applyCmdLock.LockTime()
	}

	return response, nil
}
