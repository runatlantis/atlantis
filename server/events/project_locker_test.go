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

package events_test

import (
	"fmt"
	"testing"

	. "github.com/petergtz/pegomock/v4"
	"github.com/runatlantis/atlantis/server/core/locking"
	"github.com/runatlantis/atlantis/server/core/locking/mocks"
	"github.com/runatlantis/atlantis/server/events"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/events/vcs"
	"github.com/runatlantis/atlantis/server/logging"
	. "github.com/runatlantis/atlantis/testing"
)

func TestDefaultProjectLocker_TryLockWhenLocked(t *testing.T) {
	var githubClient *vcs.GithubClient
	mockClient := vcs.NewClientProxy(githubClient, nil, nil, nil, nil)
	mockLocker := mocks.NewMockLocker()
	locker := events.DefaultProjectLocker{
		Locker:    mockLocker,
		VCSClient: mockClient,
	}
	expProject := models.Project{}
	expWorkspace := "default"
	expPull := models.PullRequest{}
	expUser := models.User{}

	lockingPull := models.PullRequest{
		Num: 2,
	}
	When(mockLocker.TryLock(expProject, expWorkspace, expPull, expUser)).ThenReturn(
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

func TestDefaultProjectLocker_TryLockWhenLockedAndEnqueued(t *testing.T) {
	var githubClient *vcs.GithubClient
	mockClient := vcs.NewClientProxy(githubClient, nil, nil, nil, nil)
	mockLocker := mocks.NewMockLocker()
	locker := events.DefaultProjectLocker{
		Locker:    mockLocker,
		VCSClient: mockClient,
	}
	expProject := models.Project{}
	expWorkspace := "default"
	expPull := models.PullRequest{}
	expUser := models.User{}

	lockingPull := models.PullRequest{
		Num: 2,
	}
	When(mockLocker.TryLock(expProject, expWorkspace, expPull, expUser)).ThenReturn(
		locking.TryLockResponse{
			LockAcquired: false,
			CurrLock: models.ProjectLock{
				Pull: lockingPull,
			},
			EnqueueStatus: &models.EnqueueStatus{
				Status:     models.Enqueued,
				QueueDepth: 1,
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
		LockFailureReason: fmt.Sprintf("This project is currently locked by an unapplied plan from pull %s. This PR entered the waiting queue :clock130:, number of PRs ahead of you: **%d**.", link, 1),
	}, res)
}

func TestDefaultProjectLocker_TryLockWhenLockedAndAlreadyInQueue(t *testing.T) {
	var githubClient *vcs.GithubClient
	mockClient := vcs.NewClientProxy(githubClient, nil, nil, nil, nil)
	mockLocker := mocks.NewMockLocker()
	locker := events.DefaultProjectLocker{
		Locker:    mockLocker,
		VCSClient: mockClient,
	}
	expProject := models.Project{}
	expWorkspace := "default"
	expPull := models.PullRequest{}
	expUser := models.User{}

	lockingPull := models.PullRequest{
		Num: 2,
	}
	When(mockLocker.TryLock(expProject, expWorkspace, expPull, expUser)).ThenReturn(
		locking.TryLockResponse{
			LockAcquired: false,
			CurrLock: models.ProjectLock{
				Pull: lockingPull,
			},
			EnqueueStatus: &models.EnqueueStatus{
				Status:     models.AlreadyInTheQueue,
				QueueDepth: 1,
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
		LockFailureReason: fmt.Sprintf("This project is currently locked by an unapplied plan from pull %s. This PR is already in the queue :clock130:, number of PRs ahead of you: **%d**.", link, 1),
	}, res)
}

func TestDefaultProjectLocker_TryLockWhenLockedSamePull(t *testing.T) {
	RegisterMockTestingT(t)
	var githubClient *vcs.GithubClient
	mockClient := vcs.NewClientProxy(githubClient, nil, nil, nil, nil)
	mockLocker := mocks.NewMockLocker()
	locker := events.DefaultProjectLocker{
		Locker:    mockLocker,
		VCSClient: mockClient,
	}
	expProject := models.Project{}
	expWorkspace := "default"
	expPull := models.PullRequest{Num: 2}
	expUser := models.User{}

	lockingPull := models.PullRequest{
		Num: 2,
	}
	lockKey := "key"
	When(mockLocker.TryLock(expProject, expWorkspace, expPull, expUser)).ThenReturn(
		locking.TryLockResponse{
			LockAcquired: false,
			CurrLock: models.ProjectLock{
				Pull: lockingPull,
			},
			LockKey: lockKey,
		},
		nil,
	)
	res, err := locker.TryLock(logging.NewNoopLogger(t), expPull, expUser, expWorkspace, expProject, true)
	Ok(t, err)
	Equals(t, true, res.LockAcquired)

	// UnlockFn should work.
	mockLocker.VerifyWasCalled(Never()).Unlock(lockKey)
	err = res.UnlockFn()
	Ok(t, err)
	mockLocker.VerifyWasCalledOnce().Unlock(lockKey)
}

func TestDefaultProjectLocker_TryLockUnlocked(t *testing.T) {
	RegisterMockTestingT(t)
	var githubClient *vcs.GithubClient
	mockClient := vcs.NewClientProxy(githubClient, nil, nil, nil, nil)
	mockLocker := mocks.NewMockLocker()
	locker := events.DefaultProjectLocker{
		Locker:    mockLocker,
		VCSClient: mockClient,
	}
	expProject := models.Project{}
	expWorkspace := "default"
	expPull := models.PullRequest{Num: 2}
	expUser := models.User{}

	lockingPull := models.PullRequest{
		Num: 2,
	}
	lockKey := "key"
	When(mockLocker.TryLock(expProject, expWorkspace, expPull, expUser)).ThenReturn(
		locking.TryLockResponse{
			LockAcquired: true,
			CurrLock: models.ProjectLock{
				Pull: lockingPull,
			},
			LockKey: lockKey,
		},
		nil,
	)
	res, err := locker.TryLock(logging.NewNoopLogger(t), expPull, expUser, expWorkspace, expProject, true)
	Ok(t, err)
	Equals(t, true, res.LockAcquired)

	// UnlockFn should work.
	mockLocker.VerifyWasCalled(Never()).Unlock(lockKey)
	err = res.UnlockFn()
	Ok(t, err)
	mockLocker.VerifyWasCalledOnce().Unlock(lockKey)
}

func TestDefaultProjectLocker_RepoLocking(t *testing.T) {
	var githubClient *vcs.GithubClient
	mockClient := vcs.NewClientProxy(githubClient, nil, nil, nil, nil)
	expProject := models.Project{}
	expWorkspace := "default"
	expPull := models.PullRequest{Num: 2}
	expUser := models.User{}
	lockKey := "key"

	tests := []struct {
		name        string
		repoLocking bool
		setup       func(locker *mocks.MockLocker, noOpLocker *mocks.MockLocker)
		verify      func(locker *mocks.MockLocker, noOpLocker *mocks.MockLocker)
	}{
		{
			"enable repo locking",
			true,
			func(locker *mocks.MockLocker, noOpLocker *mocks.MockLocker) {
				When(locker.TryLock(expProject, expWorkspace, expPull, expUser)).ThenReturn(
					locking.TryLockResponse{
						LockAcquired: true,
						CurrLock:     models.ProjectLock{},
						LockKey:      lockKey,
					},
					nil,
				)
			},
			func(locker *mocks.MockLocker, noOpLocker *mocks.MockLocker) {
				locker.VerifyWasCalledOnce().TryLock(expProject, expWorkspace, expPull, expUser)
				noOpLocker.VerifyWasCalled(Never()).TryLock(expProject, expWorkspace, expPull, expUser)
			},
		},
		{
			"disable repo locking",
			false,
			func(locker *mocks.MockLocker, noOpLocker *mocks.MockLocker) {
				When(noOpLocker.TryLock(expProject, expWorkspace, expPull, expUser)).ThenReturn(
					locking.TryLockResponse{
						LockAcquired: true,
						CurrLock:     models.ProjectLock{},
						LockKey:      lockKey,
					},
					nil,
				)
			},
			func(locker *mocks.MockLocker, noOpLocker *mocks.MockLocker) {
				locker.VerifyWasCalled(Never()).TryLock(expProject, expWorkspace, expPull, expUser)
				noOpLocker.VerifyWasCalledOnce().TryLock(expProject, expWorkspace, expPull, expUser)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			RegisterMockTestingT(t)
			mockLocker := mocks.NewMockLocker()
			mockNoOpLocker := mocks.NewMockLocker()
			locker := events.DefaultProjectLocker{
				Locker:     mockLocker,
				NoOpLocker: mockNoOpLocker,
				VCSClient:  mockClient,
			}
			tt.setup(mockLocker, mockNoOpLocker)
			res, err := locker.TryLock(logging.NewNoopLogger(t), expPull, expUser, expWorkspace, expProject, tt.repoLocking)
			Ok(t, err)
			Equals(t, true, res.LockAcquired)
			tt.verify(mockLocker, mockNoOpLocker)
		})
	}
}
