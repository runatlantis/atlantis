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
	sameWorkspaceOnPlanProject := command.ProjectContext{
		BaseRepo:      repo,
		RepoRelDir:    "terraform",
		ProjectName:   "stage",
		Workspace:     "default",
		RepoLocksMode: valid.RepoLocksOnPlanMode,
	}
	onPlanKey := "owner/repo/terraform/default/prod"
	onApplyKey := keyForProject(onApplyProject)
	disabledKey := keyForProject(disabledProject)
	secondOnPlanKey := keyForProject(secondOnPlanProject)
	sameWorkspaceOnPlanKey := "owner/repo/terraform/default/stage"

	tests := []struct {
		name                            string
		projectCmds                     []command.ProjectContext
		locksByKey                      map[string]models.ProjectLock
		expectedUnlockIfOwnedByPullKeys []string
		expectedDeletedKeys             []string
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
			expectedUnlockIfOwnedByPullKeys: []string{onPlanKey},
			expectedDeletedKeys:             []string{onPlanKey},
		},
		{
			name:                            "selected on_plan lock owned by another pull is not unlocked",
			projectCmds:                     []command.ProjectContext{onPlanProject},
			locksByKey:                      map[string]models.ProjectLock{onPlanKey: lockForPull(repo, 2)},
			expectedUnlockIfOwnedByPullKeys: []string{onPlanKey},
		},
		{
			name:                            "selected on_plan lock missing is ignored",
			projectCmds:                     []command.ProjectContext{onPlanProject},
			expectedUnlockIfOwnedByPullKeys: []string{onPlanKey},
		},
		{
			name:        "mixed mode unlocks only current-pull selected on_plan project",
			projectCmds: []command.ProjectContext{onPlanProject, onApplyProject, disabledProject},
			locksByKey: map[string]models.ProjectLock{
				onPlanKey:   lockForPull(repo, pull.Num),
				onApplyKey:  lockForPull(repo, pull.Num),
				disabledKey: lockForPull(repo, pull.Num),
			},
			expectedUnlockIfOwnedByPullKeys: []string{onPlanKey},
			expectedDeletedKeys:             []string{onPlanKey},
		},
		{
			name:        "multiple selected on_plan keys are unlocked",
			projectCmds: []command.ProjectContext{onPlanProject, secondOnPlanProject},
			locksByKey: map[string]models.ProjectLock{
				onPlanKey:       lockForPull(repo, pull.Num),
				secondOnPlanKey: lockForPull(repo, pull.Num),
			},
			expectedUnlockIfOwnedByPullKeys: []string{onPlanKey, secondOnPlanKey},
			expectedDeletedKeys:             []string{onPlanKey, secondOnPlanKey},
		},
		{
			name:        "duplicate selected on_plan lock key unlocks once",
			projectCmds: []command.ProjectContext{onPlanProject, onPlanProject},
			locksByKey: map[string]models.ProjectLock{
				onPlanKey: lockForPull(repo, pull.Num),
			},
			expectedUnlockIfOwnedByPullKeys: []string{onPlanKey},
			expectedDeletedKeys:             []string{onPlanKey},
		},
		{
			name:        "same dir and workspace with different project names unlocks both keys",
			projectCmds: []command.ProjectContext{onPlanProject, sameWorkspaceOnPlanProject},
			locksByKey: map[string]models.ProjectLock{
				onPlanKey:              lockForPull(repo, pull.Num),
				sameWorkspaceOnPlanKey: lockForPull(repo, pull.Num),
			},
			expectedUnlockIfOwnedByPullKeys: []string{
				"owner/repo/terraform/default/prod",
				"owner/repo/terraform/default/stage",
			},
			expectedDeletedKeys: []string{
				"owner/repo/terraform/default/prod",
				"owner/repo/terraform/default/stage",
			},
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
			if len(locker.getLockKeys) != 0 {
				t.Fatalf("expected no GetLock calls, got %#v", locker.getLockKeys)
			}
			if len(locker.unlockKeys) != 0 {
				t.Fatalf("expected no Unlock calls, got %#v", locker.unlockKeys)
			}
			if !equalUnlockIfOwnedByPullCalls(locker.unlockIfOwnedByPullCalls, tt.expectedUnlockIfOwnedByPullKeys, repo.FullName, pull.Num) {
				t.Fatalf("expected UnlockIfOwnedByPull keys %#v for repo %q pull %d, got %#v", tt.expectedUnlockIfOwnedByPullKeys, repo.FullName, pull.Num, locker.unlockIfOwnedByPullCalls)
			}
			if !equalStringSlices(locker.deletedKeys, tt.expectedDeletedKeys) {
				t.Fatalf("expected deleted keys %#v, got %#v", tt.expectedDeletedKeys, locker.deletedKeys)
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

type unlockIfOwnedByPullCall struct {
	lockKey      string
	repoFullName string
	pullNum      int
}

type recordingPlanCleanupLocker struct {
	locking.Locker
	locksByKey               map[string]models.ProjectLock
	unlockIfOwnedByPullCalls []unlockIfOwnedByPullCall
	deletedKeys              []string
	getLockKeys              []string
	unlockKeys               []string
	unlockByPullCalls        []unlockByPullCall
}

func (r *recordingPlanCleanupLocker) GetLock(key string) (*models.ProjectLock, error) {
	r.getLockKeys = append(r.getLockKeys, key)
	return nil, nil
}

func (r *recordingPlanCleanupLocker) Unlock(key string) (*models.ProjectLock, error) {
	r.unlockKeys = append(r.unlockKeys, key)
	return nil, nil
}

func (r *recordingPlanCleanupLocker) UnlockIfOwnedByPull(lockKey string, repoFullName string, pullNum int) (*models.ProjectLock, error) {
	r.unlockIfOwnedByPullCalls = append(r.unlockIfOwnedByPullCalls, unlockIfOwnedByPullCall{lockKey: lockKey, repoFullName: repoFullName, pullNum: pullNum})
	lock, ok := r.locksByKey[lockKey]
	if !ok || lock.Pull.Num != pullNum {
		return nil, nil
	}
	delete(r.locksByKey, lockKey)
	r.deletedKeys = append(r.deletedKeys, lockKey)
	return &lock, nil
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

func equalUnlockIfOwnedByPullCalls(calls []unlockIfOwnedByPullCall, keys []string, repoFullName string, pullNum int) bool {
	if len(calls) != len(keys) {
		return false
	}
	for i := range calls {
		if calls[i].lockKey != keys[i] || calls[i].repoFullName != repoFullName || calls[i].pullNum != pullNum {
			return false
		}
	}
	return true
}
