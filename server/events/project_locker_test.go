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

func TestDefaultProjectLocker_TryLockWhenLockedSamePull(t *testing.T) {
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
				Locker:     mockLocker,
				NoOpLocker: mockNoOpLocker,
				VCSClient:  mockClient,
			}
			tt.setup(mockLocker, mockNoOpLocker)
			res, err := locker.TryLock(logging.NewNoopLogger(t), expPull, expUser, expWorkspace, expProject, tt.repoLocking)
			Ok(t, err)
			Equals(t, true, res.LockAcquired)
		})
	}
}
