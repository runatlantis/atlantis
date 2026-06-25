// Copyright 2017 HootSuite Media Inc.
// SPDX-License-Identifier: Apache-2.0
// Modified hereafter by contributors to runatlantis/atlantis.

package events_test

import (
	"fmt"
	"testing"
	"time"

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
		LockFailureReason: fmt.Sprintf("This project is currently locked by an unapplied plan from pull %s. To continue, delete the lock from %s or apply that plan and merge the pull request.\n\nOnce the lock is released, comment `atlantis plan` here to re-plan.", link, link),
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
		LockFailureReason: fmt.Sprintf("This project is currently locked by an unapplied plan from pull %s. To continue, delete the lock from %s or apply that plan and merge the pull request.\n\nOnce the lock is released, comment `%s plan` here to re-plan.", link, link, customExecutableName),
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
	// Unlock will be called once when UnlockFn is invoked
	mockLocker.EXPECT().Unlock(lockKey).Return(nil, nil)
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
	// Unlock will be called once when UnlockFn is invoked
	mockLocker.EXPECT().Unlock(lockKey).Return(nil, nil)
	res, err := locker.TryLock(logging.NewNoopLogger(t), expPull, expUser, expWorkspace, expProject, true)
	Ok(t, err)
	Equals(t, true, res.LockAcquired)

	// UnlockFn should work.
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

func TestDefaultProjectLocker_TryLockPopulatesLockTime(t *testing.T) {
	lockTime := time.Date(2025, 6, 15, 10, 30, 0, 0, time.UTC)

	tests := []struct {
		name         string
		lockAcquired bool
		description  string
	}{
		{
			name:         "fresh lock",
			lockAcquired: true,
		},
		{
			name:         "re-acquired from same PR",
			lockAcquired: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			var githubClient *github.Client
			mockClient := vcs.NewClientProxy(githubClient, nil, nil, nil, nil, nil)
			mockLocker := mocks.NewMockLocker(ctrl)
			locker := events.DefaultProjectLocker{
				Locker:    mockLocker,
				VCSClient: mockClient,
			}

			expProject := models.Project{}
			expWorkspace := "default"
			expPull := models.PullRequest{Num: 2}
			expUser := models.User{}
			lockKey := "key"

			mockLocker.EXPECT().TryLock(expProject, expWorkspace, expPull, expUser).Return(
				locking.TryLockResponse{
					LockAcquired: tt.lockAcquired,
					CurrLock: models.ProjectLock{
						Pull: expPull,
						Time: lockTime,
					},
					LockKey: lockKey,
				},
				nil,
			)
			if tt.lockAcquired {
				mockLocker.EXPECT().Unlock(lockKey).Return(nil, nil)
			} else {
				mockLocker.EXPECT().Unlock(lockKey).Return(nil, nil)
			}

			res, err := locker.TryLock(logging.NewNoopLogger(t), expPull, expUser, expWorkspace, expProject, true)
			Ok(t, err)
			Equals(t, true, res.LockAcquired)
			Equals(t, lockTime, res.LockTime)

			err = res.UnlockFn()
			Ok(t, err)
		})
	}
}
