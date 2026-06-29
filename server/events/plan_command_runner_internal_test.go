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
	slashfulRepo := models.Repo{FullName: "group/subgroup/repo"}
	slashfulPull := models.PullRequest{BaseRepo: slashfulRepo, Num: pull.Num}
	slashfulOnPlanProject := command.ProjectContext{
		BaseRepo:      slashfulRepo,
		RepoRelDir:    "terraform",
		ProjectName:   "prod",
		Workspace:     "default",
		RepoLocksMode: valid.RepoLocksOnPlanMode,
	}
	onPlanKey := "owner/repo/terraform/default/prod"
	onApplyKey := keyForProject(onApplyProject)
	disabledKey := keyForProject(disabledProject)
	secondOnPlanKey := keyForProject(secondOnPlanProject)
	sameWorkspaceOnPlanKey := "owner/repo/terraform/default/stage"
	slashfulOnPlanKey := "group/subgroup/repo/terraform/default/prod"

	tests := []struct {
		name                             string
		ctxPull                          models.PullRequest
		projectCmds                      []command.ProjectContext
		locksByKey                       map[string]models.ProjectLock
		expectedUnlockIfOwnedByPullCalls []unlockIfOwnedByPullCall
		expectedDeletedKeys              []string
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
			expectedUnlockIfOwnedByPullCalls: []unlockIfOwnedByPullCall{expectedUnlockCall(onPlanProject, pull.Num)},
			expectedDeletedKeys:              []string{onPlanKey},
		},
		{
			name:                             "selected on_plan lock owned by another pull is not unlocked",
			projectCmds:                      []command.ProjectContext{onPlanProject},
			locksByKey:                       map[string]models.ProjectLock{onPlanKey: lockForPull(repo, 2)},
			expectedUnlockIfOwnedByPullCalls: []unlockIfOwnedByPullCall{expectedUnlockCall(onPlanProject, pull.Num)},
		},
		{
			name:                             "selected on_plan lock missing is ignored",
			projectCmds:                      []command.ProjectContext{onPlanProject},
			expectedUnlockIfOwnedByPullCalls: []unlockIfOwnedByPullCall{expectedUnlockCall(onPlanProject, pull.Num)},
		},
		{
			name:        "mixed mode unlocks only current-pull selected on_plan project",
			projectCmds: []command.ProjectContext{onPlanProject, onApplyProject, disabledProject},
			locksByKey: map[string]models.ProjectLock{
				onPlanKey:   lockForPull(repo, pull.Num),
				onApplyKey:  lockForPull(repo, pull.Num),
				disabledKey: lockForPull(repo, pull.Num),
			},
			expectedUnlockIfOwnedByPullCalls: []unlockIfOwnedByPullCall{expectedUnlockCall(onPlanProject, pull.Num)},
			expectedDeletedKeys:              []string{onPlanKey},
		},
		{
			name:        "multiple selected on_plan keys are unlocked",
			projectCmds: []command.ProjectContext{onPlanProject, secondOnPlanProject},
			locksByKey: map[string]models.ProjectLock{
				onPlanKey:       lockForPull(repo, pull.Num),
				secondOnPlanKey: lockForPull(repo, pull.Num),
			},
			expectedUnlockIfOwnedByPullCalls: []unlockIfOwnedByPullCall{
				expectedUnlockCall(onPlanProject, pull.Num),
				expectedUnlockCall(secondOnPlanProject, pull.Num),
			},
			expectedDeletedKeys: []string{onPlanKey, secondOnPlanKey},
		},
		{
			name:        "duplicate selected on_plan lock key unlocks once",
			projectCmds: []command.ProjectContext{onPlanProject, onPlanProject},
			locksByKey: map[string]models.ProjectLock{
				onPlanKey: lockForPull(repo, pull.Num),
			},
			expectedUnlockIfOwnedByPullCalls: []unlockIfOwnedByPullCall{expectedUnlockCall(onPlanProject, pull.Num)},
			expectedDeletedKeys:              []string{onPlanKey},
		},
		{
			name:        "same dir and workspace with different project names unlocks both keys",
			projectCmds: []command.ProjectContext{onPlanProject, sameWorkspaceOnPlanProject},
			locksByKey: map[string]models.ProjectLock{
				onPlanKey:              lockForPull(repo, pull.Num),
				sameWorkspaceOnPlanKey: lockForPull(repo, pull.Num),
			},
			expectedUnlockIfOwnedByPullCalls: []unlockIfOwnedByPullCall{
				{
					key:       "owner/repo/terraform/default/prod",
					project:   models.NewProject("owner/repo", "terraform", "prod"),
					workspace: "default",
					pullNum:   pull.Num,
				},
				{
					key:       "owner/repo/terraform/default/stage",
					project:   models.NewProject("owner/repo", "terraform", "stage"),
					workspace: "default",
					pullNum:   pull.Num,
				},
			},
			expectedDeletedKeys: []string{
				"owner/repo/terraform/default/prod",
				"owner/repo/terraform/default/stage",
			},
		},
		{
			name:        "slashful repo name unlocks selected project without parsing lock key",
			ctxPull:     slashfulPull,
			projectCmds: []command.ProjectContext{slashfulOnPlanProject},
			locksByKey: map[string]models.ProjectLock{
				slashfulOnPlanKey: lockForPull(slashfulRepo, slashfulPull.Num),
			},
			expectedUnlockIfOwnedByPullCalls: []unlockIfOwnedByPullCall{
				{
					key:       "group/subgroup/repo/terraform/default/prod",
					project:   models.NewProject("group/subgroup/repo", "terraform", "prod"),
					workspace: "default",
					pullNum:   slashfulPull.Num,
				},
			},
			expectedDeletedKeys: []string{"group/subgroup/repo/terraform/default/prod"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctxPull := pull
			if tt.ctxPull.Num != 0 || tt.ctxPull.BaseRepo.FullName != "" {
				ctxPull = tt.ctxPull
			}
			finder := &recordingPendingPlanFinder{}
			locker := &recordingPlanCleanupLocker{locksByKey: tt.locksByKey}
			runner := &PlanCommandRunner{
				workingDir:        &planCleanupWorkingDir{pullDir: pullDir},
				pendingPlanFinder: finder,
				lockingLocker:     locker,
			}
			ctx := &command.Context{
				Log:  logging.NewNoopLogger(t),
				Pull: ctxPull,
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
			if !equalUnlockIfOwnedByPullCalls(locker.unlockIfOwnedByPullCalls, tt.expectedUnlockIfOwnedByPullCalls) {
				t.Fatalf("expected UnlockIfOwnedByPull calls %#v, got %#v", tt.expectedUnlockIfOwnedByPullCalls, locker.unlockIfOwnedByPullCalls)
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
	key       string
	project   models.Project
	workspace string
	pullNum   int
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

func (r *recordingPlanCleanupLocker) UnlockIfOwnedByPull(project models.Project, workspace string, pullNum int) (*models.ProjectLock, error) {
	key := models.GenerateLockKey(project, workspace)
	r.unlockIfOwnedByPullCalls = append(r.unlockIfOwnedByPullCalls, unlockIfOwnedByPullCall{key: key, project: project, workspace: workspace, pullNum: pullNum})
	lock, ok := r.locksByKey[key]
	if !ok || lock.Pull.Num != pullNum {
		return nil, nil
	}
	delete(r.locksByKey, key)
	r.deletedKeys = append(r.deletedKeys, key)
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

func expectedUnlockCall(project command.ProjectContext, pullNum int) unlockIfOwnedByPullCall {
	modelProject := models.NewProject(project.BaseRepo.FullName, project.RepoRelDir, project.ProjectName)
	return unlockIfOwnedByPullCall{
		key:       models.GenerateLockKey(modelProject, project.Workspace),
		project:   modelProject,
		workspace: project.Workspace,
		pullNum:   pullNum,
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

func equalUnlockIfOwnedByPullCalls(calls []unlockIfOwnedByPullCall, expected []unlockIfOwnedByPullCall) bool {
	if len(calls) != len(expected) {
		return false
	}
	for i := range calls {
		if calls[i] != expected[i] {
			return false
		}
	}
	return true
}
