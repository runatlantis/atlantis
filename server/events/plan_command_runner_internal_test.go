// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package events

import (
	"testing"

	"github.com/runatlantis/atlantis/server/core/config/valid"
	"github.com/runatlantis/atlantis/server/core/locking"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/logging"
)

func TestPlanCommandRunner_DeletePlansAndPlanLocksByRepoLockMode(t *testing.T) {
	repo := models.Repo{FullName: "owner/repo"}
	pull := models.PullRequest{BaseRepo: repo, Num: 1}
	pullDir := "/tmp/pull-dir"

	onPlanProject := command.ProjectContext{
		BaseRepo:      repo,
		RepoRelDir:    "terraform",
		ProjectName:   "prod",
		Workspace:     "default",
		RepoLocksMode: valid.RepoLocksOnPlanMode,
	}
	onApplyProject := command.ProjectContext{
		BaseRepo:      repo,
		RepoRelDir:    "terraform",
		ProjectName:   "prod",
		Workspace:     "default",
		RepoLocksMode: valid.RepoLocksOnApplyMode,
	}
	disabledProject := command.ProjectContext{
		BaseRepo:      repo,
		RepoRelDir:    "terraform",
		ProjectName:   "prod",
		Workspace:     "default",
		RepoLocksMode: valid.RepoLocksDisabledMode,
	}

	tests := []struct {
		name                     string
		projectCmds              []command.ProjectContext
		expectUnlockByPull       bool
		expectedUnlockByPullRepo string
		expectedUnlockByPullNum  int
		expectedUnlockKeys       []string
	}{
		{
			name:        "on_apply deletes plans without deleting pull locks",
			projectCmds: []command.ProjectContext{onApplyProject},
		},
		{
			name:        "disabled deletes plans without deleting pull locks",
			projectCmds: []command.ProjectContext{disabledProject},
		},
		{
			name:                     "on_plan keeps pull-wide cleanup",
			projectCmds:              []command.ProjectContext{onPlanProject},
			expectUnlockByPull:       true,
			expectedUnlockByPullRepo: repo.FullName,
			expectedUnlockByPullNum:  pull.Num,
		},
		{
			name:        "mixed mode unlocks only on_plan project",
			projectCmds: []command.ProjectContext{onPlanProject, onApplyProject},
			expectedUnlockKeys: []string{
				"owner/repo/terraform/default/prod",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			finder := &recordingPendingPlanFinder{}
			locker := &recordingPlanCleanupLocker{}
			runner := &PlanCommandRunner{
				workingDir:        &planCleanupWorkingDir{pullDir: pullDir},
				pendingPlanFinder: finder,
				lockingLocker:     locker,
			}
			ctx := &command.Context{
				Log:  logging.NewNoopLogger(t),
				Pull: pull,
			}

			runner.deletePlansAndPlanLocks(ctx, tt.projectCmds)

			if len(finder.deletedPullDirs) != 1 || finder.deletedPullDirs[0] != pullDir {
				t.Fatalf("expected DeletePlans(%q), got %#v", pullDir, finder.deletedPullDirs)
			}
			if tt.expectUnlockByPull {
				if len(locker.unlockByPullCalls) != 1 {
					t.Fatalf("expected one UnlockByPull call, got %#v", locker.unlockByPullCalls)
				}
				call := locker.unlockByPullCalls[0]
				if call.repoFullName != tt.expectedUnlockByPullRepo || call.pullNum != tt.expectedUnlockByPullNum {
					t.Fatalf("expected UnlockByPull(%q, %d), got UnlockByPull(%q, %d)", tt.expectedUnlockByPullRepo, tt.expectedUnlockByPullNum, call.repoFullName, call.pullNum)
				}
			} else if len(locker.unlockByPullCalls) != 0 {
				t.Fatalf("expected no UnlockByPull calls, got %#v", locker.unlockByPullCalls)
			}
			if !equalStringSlices(locker.unlockKeys, tt.expectedUnlockKeys) {
				t.Fatalf("expected Unlock keys %#v, got %#v", tt.expectedUnlockKeys, locker.unlockKeys)
			}
		})
	}
}

type recordingPendingPlanFinder struct {
	PendingPlanFinder
	deletedPullDirs []string
}

func (r *recordingPendingPlanFinder) DeletePlans(pullDir string) error {
	r.deletedPullDirs = append(r.deletedPullDirs, pullDir)
	return nil
}

type unlockByPullCall struct {
	repoFullName string
	pullNum      int
}

type recordingPlanCleanupLocker struct {
	locking.Locker
	unlockKeys        []string
	unlockByPullCalls []unlockByPullCall
}

func (r *recordingPlanCleanupLocker) Unlock(key string) (*models.ProjectLock, error) {
	r.unlockKeys = append(r.unlockKeys, key)
	return nil, nil
}

func (r *recordingPlanCleanupLocker) UnlockByPull(repoFullName string, pullNum int) ([]models.ProjectLock, error) {
	r.unlockByPullCalls = append(r.unlockByPullCalls, unlockByPullCall{repoFullName: repoFullName, pullNum: pullNum})
	return nil, nil
}

type planCleanupWorkingDir struct {
	WorkingDir
	pullDir string
}

func (p *planCleanupWorkingDir) GetPullDir(models.Repo, models.PullRequest) (string, error) {
	return p.pullDir, nil
}

func equalStringSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
