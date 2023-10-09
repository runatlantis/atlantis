// Copyright 2017 HootSuite Media Inc.
//
// Licensed under the Apache License, Version 2.0 (the License);
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//    http://www.apache.org/licenses/LICENSE-2.0
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an AS IS BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
// Modified hereafter by contributors to runatlantis/atlantis.

package events

import (
	"fmt"

	"github.com/runatlantis/atlantis/server/core/locking"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/events/vcs"
	"github.com/runatlantis/atlantis/server/logging"
)

//go:generate pegomock generate --package mocks -o mocks/mock_project_lock.go ProjectLocker

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
	Locker     locking.Locker
	NoOpLocker locking.Locker
	VCSClient  vcs.Client
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
		var failureMsg string
		if lockAttempt.EnqueueStatus == nil {
			failureMsg = fmt.Sprintf(
				"This project is currently locked by an unapplied plan from pull %s. To continue, delete the lock from %s or apply that plan and merge the pull request.\n\nOnce the lock is released, comment `atlantis plan` here to re-plan.",
				link,
				link)
		} else {
			failureMsg = generateMessageWithQueueStatus(link, lockAttempt)
		}
		return &TryLockResponse{
			LockAcquired:      false,
			LockFailureReason: failureMsg,
		}, nil
	}
	log.Info("acquired lock with id %q", lockAttempt.LockKey)
	return &TryLockResponse{
		LockAcquired: true,
		UnlockFn: func() error {
			// TODO(Ghais): based on a feature flag, either call this function below or
			//  omit UnlockFn altogether (caller doesn't lose the lock)
			_, dequeuedLock, err := p.Locker.Unlock(lockAttempt.LockKey)
			if dequeuedLock != nil {
				// TODO(Ghais): add test
				return p.commentOnDequeuedPullRequest(log, *dequeuedLock)
			}
			return err
		},
		LockKey: lockAttempt.LockKey,
	}, nil
}

func generateMessageWithQueueStatus(link string, lockAttempt locking.TryLockResponse) string {
	failureMsg := fmt.Sprintf("This project is currently locked by an unapplied plan from pull %s.", link)
	switch lockAttempt.EnqueueStatus.Status {
	case models.Enqueued:
		failureMsg = fmt.Sprintf("%s This PR entered the waiting queue :clock130:, number of PRs ahead of you: **%d**.", failureMsg, lockAttempt.EnqueueStatus.QueueDepth)
	case models.AlreadyInTheQueue:
		failureMsg = fmt.Sprintf("%s This PR is already in the queue :clock130:, number of PRs ahead of you: **%d**.", failureMsg, lockAttempt.EnqueueStatus.QueueDepth)
	}
	return failureMsg
}

func (p *DefaultProjectLocker) commentOnDequeuedPullRequest(log logging.SimpleLogging, dequeuedLock models.ProjectLock) error {
	planVcsMessage := models.BuildCommentOnDequeuedPullRequest([]models.ProjectLock{dequeuedLock})
	if commentErr := p.VCSClient.CreateComment(dequeuedLock.Pull.BaseRepo, dequeuedLock.Pull.Num, planVcsMessage, ""); commentErr != nil {
		log.Err("unable to comment on PR %d: %s", dequeuedLock.Pull.Num, commentErr)
		return commentErr
	}
	return nil
}
