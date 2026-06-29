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
		RepoRelDir:    "terraform/apply",
		ProjectName:   "apply-prod",
		Workspace:     "default",
		RepoLocksMode: valid.RepoLocksOnApplyMode,
	}
	disabledProject := command.ProjectContext{
		BaseRepo:      repo,
		RepoRelDir:    "terraform/disabled",
		ProjectName:   "disabled-prod",
		Workspace:     "default",
		RepoLocksMode: valid.RepoLocksDisabledMode,
	}
	secondOnPlanProject := command.ProjectContext{
		BaseRepo:      repo,
		RepoRelDir:    "terraform",
		ProjectName:   "stage",
		Workspace:     "staging",
		RepoLocksMode: valid.RepoLocksOnPlanMode,
	}
	onPlanKey := keyForProject(onPlanProject)
	onApplyKey := keyForProject(onApplyProject)
	disabledKey := keyForProject(disabledProject)
	secondOnPlanKey := keyForProject(secondOnPlanProject)

	tests := []struct {
		name                string
		projectCmds         []command.ProjectContext
		locksByKey          map[string]models.ProjectLock
		expectedGetLockKeys []string
		expectedUnlockKeys  []string
	}{
		{
			name: "empty selection deletes plans without deleting locks",
		},
		{
			name:        "on_apply deletes plans without deleting locks",
			projectCmds: []command.ProjectContext{onApplyProject},
			locksByKey: map[string]models.ProjectLock{
				onApplyKey: lockForPull(repo, pull.Num),
			},
		},
		{
			name:        "disabled deletes plans without deleting locks",
			projectCmds: []command.ProjectContext{disabledProject},
			locksByKey: map[string]models.ProjectLock{
				disabledKey: lockForPull(repo, pull.Num),
			},
		},
		{
			name:        "selected on_plan lock owned by current pull is unlocked",
			projectCmds: []command.ProjectContext{onPlanProject},
			locksByKey: map[string]models.ProjectLock{
				onPlanKey: lockForPull(repo, pull.Num),
			},
			expectedGetLockKeys: []string{onPlanKey},
			expectedUnlockKeys:  []string{onPlanKey},
		},
		{
			name:                "selected on_plan lock owned by another pull is not unlocked",
			projectCmds:         []command.ProjectContext{onPlanProject},
			locksByKey:          map[string]models.ProjectLock{onPlanKey: lockForPull(repo, 2)},
			expectedGetLockKeys: []string{onPlanKey},
		},
		{
			name:                "selected on_plan lock missing is ignored",
			projectCmds:         []command.ProjectContext{onPlanProject},
			expectedGetLockKeys: []string{onPlanKey},
		},
		{
			name:        "mixed mode unlocks only current-pull selected on_plan project",
			projectCmds: []command.ProjectContext{onPlanProject, onApplyProject, disabledProject},
			locksByKey: map[string]models.ProjectLock{
				onPlanKey:   lockForPull(repo, pull.Num),
				onApplyKey:  lockForPull(repo, pull.Num),
				disabledKey: lockForPull(repo, pull.Num),
			},
			expectedGetLockKeys: []string{onPlanKey},
			expectedUnlockKeys:  []string{onPlanKey},
		},
		{
			name:        "multiple selected on_plan keys are unlocked",
			projectCmds: []command.ProjectContext{onPlanProject, secondOnPlanProject},
			locksByKey: map[string]models.ProjectLock{
				onPlanKey:       lockForPull(repo, pull.Num),
				secondOnPlanKey: lockForPull(repo, pull.Num),
			},
			expectedGetLockKeys: []string{onPlanKey, secondOnPlanKey},
			expectedUnlockKeys:  []string{onPlanKey, secondOnPlanKey},
		},
		{
			name:        "duplicate selected on_plan lock key unlocks once",
			projectCmds: []command.ProjectContext{onPlanProject, onPlanProject},
			locksByKey: map[string]models.ProjectLock{
				onPlanKey: lockForPull(repo, pull.Num),
			},
			expectedGetLockKeys: []string{onPlanKey},
			expectedUnlockKeys:  []string{onPlanKey},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			finder := &recordingPendingPlanFinder{}
			locker := &recordingPlanCleanupLocker{locksByKey: tt.locksByKey}
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
			if len(locker.unlockByPullCalls) != 0 {
				t.Fatalf("expected no UnlockByPull calls, got %#v", locker.unlockByPullCalls)
			}
			if !equalStringSlices(locker.getLockKeys, tt.expectedGetLockKeys) {
				t.Fatalf("expected GetLock keys %#v, got %#v", tt.expectedGetLockKeys, locker.getLockKeys)
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
	locksByKey        map[string]models.ProjectLock
	getLockKeys       []string
	unlockKeys        []string
	unlockByPullCalls []unlockByPullCall
}

func (r *recordingPlanCleanupLocker) GetLock(key string) (*models.ProjectLock, error) {
	r.getLockKeys = append(r.getLockKeys, key)
	lock, ok := r.locksByKey[key]
	if !ok {
		return nil, nil
	}
	return &lock, nil
}

func (r *recordingPlanCleanupLocker) Unlock(key string) (*models.ProjectLock, error) {
	r.unlockKeys = append(r.unlockKeys, key)
	delete(r.locksByKey, key)
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

func keyForProject(project command.ProjectContext) string {
	return GenerateLockID(project)
}

func lockForPull(repo models.Repo, pullNum int) models.ProjectLock {
	return models.ProjectLock{
		Pull: models.PullRequest{
			BaseRepo: repo,
			Num:      pullNum,
		},
	}
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
