// Copyright 2017 HootSuite Media Inc.
// SPDX-License-Identifier: Apache-2.0
// Modified hereafter by contributors to runatlantis/atlantis.

package events_test

import (
	"fmt"
	"testing"

	"github.com/runatlantis/atlantis/server/core/locking"
	"github.com/runatlantis/atlantis/server/core/locking/mocks"
	"github.com/runatlantis/atlantis/server/events"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/events/vcs"
	"github.com/runatlantis/atlantis/server/events/vcs/github"
	"github.com/runatlantis/atlantis/server/logging"
	. "github.com/runatlantis/atlantis/testing"
	"go.uber.org/mock/gomock"
)

func TestDefaultProjectLocker_TryLockWhenLocked(t *testing.T) {
	ctrl := gomock.NewController(t)
	var githubClient *github.Client
	mockClient := vcs.NewClientProxy(githubClient, nil, nil, nil, nil, nil)
	mockLocker := mocks.NewMockLocker(ctrl)
	locker := events.DefaultProjectLocker{
		Locker:         mockLocker,
		VCSClient:      mockClient,
		ExecutableName: "atlantis",
	}
	expProject := models.Project{}
	expWorkspace := "default"
	expPull := models.PullRequest{}
	expUser := models.User{}

	lockingPull := models.PullRequest{
		Num: 2,
	}
	mockLocker.EXPECT().TryLock(expProject, expWorkspace, expPull, expUser).Return(
		locking.TryLockResponse{
			LockAcquired: false,
			CurrLock: models.ProjectLock{
				Pull: lockingPull,
			},
			LockKey: "",
		},
		nil,
	)
	res, err := locker.TryLock(logging.NewNoopLogger(t), expPull, expUser, expWorkspace, expProject, true)
	link, _ := mockClient.MarkdownPullLink(lockingPull)
	Ok(t, err)
	Equals(t, &events.TryLockResponse{
		LockAcquired:      false,
		LockFailureReason: fmt.Sprintf("This project is currently locked by an unapplied plan from pull %s.\n\nTo continue, release the lock in one of these ways:\n* Comment `atlantis unlock` on pull %s to delete all of its plans and locks.\n* Apply that plan and merge pull %s.\n\nOnce the lock is released, comment `atlantis plan` here to re-plan.", link, link, link),
	}, res)
}

func TestDefaultProjectLocker_TryLockWhenLockedCustomExecutableName(t *testing.T) {
	ctrl := gomock.NewController(t)
	var githubClient *github.Client
	mockClient := vcs.NewClientProxy(githubClient, nil, nil, nil, nil, nil)
	mockLocker := mocks.NewMockLocker(ctrl)
	customExecutableName := "atlantis-my-custom-name"
	locker := events.DefaultProjectLocker{
		Locker:         mockLocker,
		VCSClient:      mockClient,
		ExecutableName: customExecutableName,
	}
	expProject := models.Project{}
	expWorkspace := "default"
	expPull := models.PullRequest{}
	expUser := models.User{}

	lockingPull := models.PullRequest{
		Num: 2,
	}
	mockLocker.EXPECT().TryLock(expProject, expWorkspace, expPull, expUser).Return(
		locking.TryLockResponse{
			LockAcquired: false,
			CurrLock: models.ProjectLock{
				Pull: lockingPull,
			},
			LockKey: "",
		},
		nil,
	)
	res, err := locker.TryLock(logging.NewNoopLogger(t), expPull, expUser, expWorkspace, expProject, true)
	link, _ := mockClient.MarkdownPullLink(lockingPull)
	Ok(t, err)
	Equals(t, &events.TryLockResponse{
		LockAcquired:      false,
		LockFailureReason: fmt.Sprintf("This project is currently locked by an unapplied plan from pull %s.\n\nTo continue, release the lock in one of these ways:\n* Comment `%s unlock` on pull %s to delete all of its plans and locks.\n* Apply that plan and merge pull %s.\n\nOnce the lock is released, comment `%s plan` here to re-plan.", link, customExecutableName, link, link, customExecutableName),
	}, res)
}

func TestDefaultProjectLocker_TryLockWhenLockedSamePull(t *testing.T) {
	ctrl := gomock.NewController(t)
	var githubClient *github.Client
	mockClient := vcs.NewClientProxy(githubClient, nil, nil, nil, nil, nil)
	mockLocker := mocks.NewMockLocker(ctrl)
	locker := events.DefaultProjectLocker{
		Locker:         mockLocker,
		VCSClient:      mockClient,
		ExecutableName: "atlantis",
	}
	expProject := models.Project{}
	expWorkspace := "default"
	expPull := models.PullRequest{Num: 2}
	expUser := models.User{}

	lockingPull := models.PullRequest{
		Num: 2,
	}
	lockKey := "key"
	mockLocker.EXPECT().TryLock(expProject, expWorkspace, expPull, expUser).Return(
		locking.TryLockResponse{
			LockAcquired: false,
			CurrLock: models.ProjectLock{
				Pull: lockingPull,
			},
			LockKey: lockKey,
		},
		nil,
	)
	mockLocker.EXPECT().UnlockIfOwnedByPull(expProject, expWorkspace, expPull.Num).Return(nil, nil)
	res, err := locker.TryLock(logging.NewNoopLogger(t), expPull, expUser, expWorkspace, expProject, true)
	Ok(t, err)
	Equals(t, true, res.LockAcquired)

	// UnlockFn should work.
	err = res.UnlockFn()
	Ok(t, err)
}

func TestDefaultProjectLocker_TryLockUnlocked(t *testing.T) {
	ctrl := gomock.NewController(t)
	var githubClient *github.Client
	mockClient := vcs.NewClientProxy(githubClient, nil, nil, nil, nil, nil)
	mockLocker := mocks.NewMockLocker(ctrl)
	locker := events.DefaultProjectLocker{
		Locker:         mockLocker,
		VCSClient:      mockClient,
		ExecutableName: "atlantis",
	}
	expProject := models.Project{}
	expWorkspace := "default"
	expPull := models.PullRequest{Num: 2}
	expUser := models.User{}

	lockingPull := models.PullRequest{
		Num: 2,
	}
	lockKey := "key"
	mockLocker.EXPECT().TryLock(expProject, expWorkspace, expPull, expUser).Return(
		locking.TryLockResponse{
			LockAcquired: true,
			CurrLock: models.ProjectLock{
				Pull: lockingPull,
			},
			LockKey: lockKey,
		},
		nil,
	)
	mockLocker.EXPECT().UnlockIfOwnedByPull(expProject, expWorkspace, expPull.Num).Return(nil, nil)
	res, err := locker.TryLock(logging.NewNoopLogger(t), expPull, expUser, expWorkspace, expProject, true)
	Ok(t, err)
	Equals(t, true, res.LockAcquired)

	// UnlockFn should work.
	err = res.UnlockFn()
	Ok(t, err)
}

func TestDefaultProjectLocker_TryLockRepoLockingDisabledUnlockUsesNoOpLocker(t *testing.T) {
	ctrl := gomock.NewController(t)
	var githubClient *github.Client
	mockClient := vcs.NewClientProxy(githubClient, nil, nil, nil, nil, nil)
	mockLocker := mocks.NewMockLocker(ctrl)
	mockNoOpLocker := mocks.NewMockLocker(ctrl)
	locker := events.DefaultProjectLocker{
		Locker:         mockLocker,
		NoOpLocker:     mockNoOpLocker,
		VCSClient:      mockClient,
		ExecutableName: "atlantis",
	}
	expProject := models.Project{}
	expWorkspace := "default"
	expPull := models.PullRequest{Num: 2}
	expUser := models.User{}
	lockKey := "key"

	mockNoOpLocker.EXPECT().TryLock(expProject, expWorkspace, expPull, expUser).Return(
		locking.TryLockResponse{
			LockAcquired: true,
			CurrLock:     models.ProjectLock{},
			LockKey:      lockKey,
		},
		nil,
	)
	mockNoOpLocker.EXPECT().UnlockIfOwnedByPull(expProject, expWorkspace, expPull.Num).Return(nil, nil)

	res, err := locker.TryLock(logging.NewNoopLogger(t), expPull, expUser, expWorkspace, expProject, false)
	Ok(t, err)
	Equals(t, true, res.LockAcquired)

	err = res.UnlockFn()
	Ok(t, err)
}

func TestDefaultProjectLocker_RepoLocking(t *testing.T) {
	var githubClient *github.Client
	mockClient := vcs.NewClientProxy(githubClient, nil, nil, nil, nil, nil)
	expProject := models.Project{}
	expWorkspace := "default"
	expPull := models.PullRequest{Num: 2}
	expUser := models.User{}
	lockKey := "key"

	tests := []struct {
		name        string
		repoLocking bool
		setup       func(locker *mocks.MockLocker, noOpLocker *mocks.MockLocker)
	}{
		{
			"enable repo locking",
			true,
			func(locker *mocks.MockLocker, noOpLocker *mocks.MockLocker) {
				locker.EXPECT().TryLock(expProject, expWorkspace, expPull, expUser).Return(
					locking.TryLockResponse{
						LockAcquired: true,
						CurrLock:     models.ProjectLock{},
						LockKey:      lockKey,
					},
					nil,
				)
				// noOpLocker has no EXPECT — gomock will fail if it's called
			},
		},
		{
			"disable repo locking",
			false,
			func(locker *mocks.MockLocker, noOpLocker *mocks.MockLocker) {
				noOpLocker.EXPECT().TryLock(expProject, expWorkspace, expPull, expUser).Return(
					locking.TryLockResponse{
						LockAcquired: true,
						CurrLock:     models.ProjectLock{},
						LockKey:      lockKey,
					},
					nil,
				)
				// locker has no EXPECT — gomock will fail if it's called
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			mockLocker := mocks.NewMockLocker(ctrl)
			mockNoOpLocker := mocks.NewMockLocker(ctrl)
			locker := events.DefaultProjectLocker{
				Locker:         mockLocker,
				NoOpLocker:     mockNoOpLocker,
				VCSClient:      mockClient,
				ExecutableName: "atlantis",
			}
			tt.setup(mockLocker, mockNoOpLocker)
			res, err := locker.TryLock(logging.NewNoopLogger(t), expPull, expUser, expWorkspace, expProject, tt.repoLocking)
			Ok(t, err)
			Equals(t, true, res.LockAcquired)
		})
	}
}

type fakeLockURLGenerator struct{}

func (fakeLockURLGenerator) GenerateLockURL(lockID string) string {
	return "https://atlantis.example.com/lock?id=" + lockID
}

func TestDefaultProjectLocker_TryLockWhenLockedLinksToLockInUI(t *testing.T) {
	ctrl := gomock.NewController(t)
	var githubClient *github.Client
	mockClient := vcs.NewClientProxy(githubClient, nil, nil, nil, nil, nil)
	mockLocker := mocks.NewMockLocker(ctrl)
	locker := events.DefaultProjectLocker{
		Locker:           mockLocker,
		VCSClient:        mockClient,
		ExecutableName:   "atlantis-staging",
		LockURLGenerator: fakeLockURLGenerator{},
	}
	expProject := models.Project{}
	expWorkspace := "default"
	expPull := models.PullRequest{Num: 5}
	expUser := models.User{}

	lockingPull := models.PullRequest{Num: 2}
	mockLocker.EXPECT().TryLock(expProject, expWorkspace, expPull, expUser).Return(
		locking.TryLockResponse{
			LockAcquired: false,
			CurrLock:     models.ProjectLock{Pull: lockingPull},
			LockKey:      "repo/dir/default",
		},
		nil,
	)
	res, err := locker.TryLock(logging.NewNoopLogger(t), expPull, expUser, expWorkspace, expProject, true)
	link, _ := mockClient.MarkdownPullLink(lockingPull)
	Ok(t, err)
	Equals(t, false, res.LockAcquired)
	Equals(t, fmt.Sprintf("This project is currently locked by an unapplied plan from pull %s.\n\nTo continue, release the lock in one of these ways:\n* Comment `atlantis-staging unlock` on pull %s to delete all of its plans and locks.\n* [Delete only this project's plan and lock in the Atlantis UI](https://atlantis.example.com/lock?id=repo/dir/default).\n* Apply that plan and merge pull %s.\n\nOnce the lock is released, comment `atlantis-staging plan` here to re-plan.", link, link, link), res.LockFailureReason)
}
