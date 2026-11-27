// Copyright 2017 HootSuite Media Inc.
// SPDX-License-Identifier: Apache-2.0
// Modified hereafter by contributors to runatlantis/atlantis.

package events

import (
	"fmt"
	"strings"

	"github.com/runatlantis/atlantis/server/core/locking"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/events/vcs"
	"github.com/runatlantis/atlantis/server/logging"
)

//go:generate go tool pegomock generate --package mocks -o mocks/mock_project_lock.go ProjectLocker

// ProjectLocker locks this project against other plans being run until this
// project is unlocked.
type ProjectLocker interface {
	// TryLock attempts to acquire the lock for this project. It returns true if the lock
	// was acquired. If it returns false, the lock was not acquired and the second
	// return value will be a string describing why the lock was not acquired.
	// The third return value is a function that can be called to unlock the
	// lock. It will only be set if the lock was acquired. Any errors will set
	// error.
	TryLock(log logging.SimpleLogging, pull models.PullRequest, user models.User, workspace string, project models.Project, repoLocking bool) (*TryLockResponse, error)
}

// DefaultProjectLocker implements ProjectLocker.
type DefaultProjectLocker struct {
	Locker         locking.Locker
	NoOpLocker     locking.Locker
	VCSClient      vcs.Client
	ExecutableName string
	// LockURLGenerator, if set, is used to link to the lock in the Atlantis UI
	// so it can be discarded with one click.
	LockURLGenerator LockURLGenerator
}

// TryLockResponse is the result of trying to lock a project.
type TryLockResponse struct {
	// LockAcquired is true if the lock was acquired.
	LockAcquired bool
	// LockFailureReason is the reason why the lock was not acquired. It will
	// only be set if LockAcquired is false.
	LockFailureReason string
	// UnlockFn will unlock the lock created by the caller. This might be called
	// if there is an error later and the caller doesn't want to continue to
	// hold the lock.
	UnlockFn func() error
	// LockKey is the key for the lock if the lock was acquired.
	LockKey string
}

// TryLock implements ProjectLocker.TryLock.
func (p *DefaultProjectLocker) TryLock(log logging.SimpleLogging, pull models.PullRequest, user models.User, workspace string, project models.Project, repoLocking bool) (*TryLockResponse, error) {
	locker := p.Locker
	if !repoLocking {
		locker = p.NoOpLocker
	}

	lockAttempt, err := locker.TryLock(project, workspace, pull, user)
	if err != nil {
		return nil, err
	}
	if !lockAttempt.LockAcquired && lockAttempt.CurrLock.Pull.Num != pull.Num {
		link, err := p.VCSClient.MarkdownPullLink(lockAttempt.CurrLock.Pull)
		if err != nil {
			return nil, err
		}
		failureMsg := p.lockedFailureMessage(link, lockAttempt.LockKey)
		return &TryLockResponse{
			LockAcquired:      false,
			LockFailureReason: failureMsg,
		}, nil
	}
	log.Info("Acquired lock with id '%s'", lockAttempt.LockKey)
	return &TryLockResponse{
		LockAcquired: true,
		UnlockFn: func() error {
			_, err := locker.UnlockIfOwnedByPull(project, workspace, pull.Num)
			return err
		},
		LockKey: lockAttempt.LockKey,
	}, nil
}

func (p *DefaultProjectLocker) lockedFailureMessage(pullLink string, lockKey string) string {
	var msg strings.Builder
	fmt.Fprintf(&msg, "This project is currently locked by an unapplied plan from pull %s.\n\n", pullLink)
	msg.WriteString("To continue, release the lock in one of these ways:\n")
	fmt.Fprintf(&msg, "* Comment `%s unlock` on pull %s to delete all of its plans and locks.\n", p.ExecutableName, pullLink)
	if p.LockURLGenerator != nil && lockKey != "" {
		fmt.Fprintf(&msg, "* [Delete only this project's plan and lock in the Atlantis UI](%s).\n", p.LockURLGenerator.GenerateLockURL(lockKey))
	}
	fmt.Fprintf(&msg, "* Apply that plan and merge pull %s.\n\n", pullLink)
	fmt.Fprintf(&msg, "Once the lock is released, comment `%s plan` here to re-plan.", p.ExecutableName)
	return msg.String()
}
