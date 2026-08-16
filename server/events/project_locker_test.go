// Copyright 2017 HootSuite Media Inc.
// SPDX-License-Identifier: Apache-2.0
// Modified hereafter by contributors to runatlantis/atlantis.

package events_test

import (
	"fmt"
	"testing"

	"github.com/runatlantis/atlantis/server/core/coordination"
	"github.com/runatlantis/atlantis/server/core/coordination/mocks"
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
		coordination.TryLockResponse{
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
	Equals(t, &events.ProjectLockResponse{
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
		coordination.TryLockResponse{
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
	Equals(t, &events.ProjectLockResponse{
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
		coordination.TryLockResponse{
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
		coordination.TryLockResponse{
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

func TestDefaultProjectLocker_TryLockRepoLockingDisabledSkipsLocker(t *testing.T) {
	ctrl := gomock.NewController(t)
	var githubClient *github.Client
	mockClient := vcs.NewClientProxy(githubClient, nil, nil, nil, nil, nil)
	// No EXPECT calls — gomock fails the test if the locker is touched.
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

	res, err := locker.TryLock(logging.NewNoopLogger(t), expPull, expUser, expWorkspace, expProject, false)
	Ok(t, err)
	Equals(t, true, res.LockAcquired)

	err = res.UnlockFn()
	Ok(t, err)
}

func TestDefaultProjectLocker_TryLockGlobalDisableSkipsLocker(t *testing.T) {
	ctrl := gomock.NewController(t)
	var githubClient *github.Client
	mockClient := vcs.NewClientProxy(githubClient, nil, nil, nil, nil, nil)
	// No EXPECT calls — gomock fails the test if the locker is touched.
	mockLocker := mocks.NewMockLocker(ctrl)
	locker := events.DefaultProjectLocker{
		Locker:             mockLocker,
		VCSClient:          mockClient,
		DisableRepoLocking: true,
		ExecutableName:     "atlantis",
	}

	res, err := locker.TryLock(logging.NewNoopLogger(t), models.PullRequest{Num: 2}, models.User{}, "default", models.Project{}, true)
	Ok(t, err)
	Equals(t, true, res.LockAcquired)
	Ok(t, res.UnlockFn())
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
		setup       func(locker *mocks.MockLocker)
	}{
		{
			"enable repo locking",
			true,
			func(locker *mocks.MockLocker) {
				locker.EXPECT().TryLock(expProject, expWorkspace, expPull, expUser).Return(
					coordination.TryLockResponse{
						LockAcquired: true,
						CurrLock:     models.ProjectLock{},
						LockKey:      lockKey,
					},
					nil,
				)
			},
		},
		{
			"disable repo locking",
			false,
			func(locker *mocks.MockLocker) {
				// no EXPECT — gomock will fail if the locker is called
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			mockLocker := mocks.NewMockLocker(ctrl)
			locker := events.DefaultProjectLocker{
				Locker:         mockLocker,
				VCSClient:      mockClient,
				ExecutableName: "atlantis",
			}
			tt.setup(mockLocker)
			res, err := locker.TryLock(logging.NewNoopLogger(t), expPull, expUser, expWorkspace, expProject, tt.repoLocking)
			Ok(t, err)
			Equals(t, true, res.LockAcquired)
		})
	}
}
