// Copyright 2026 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package events_test

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/alicebob/miniredis/v2"
	. "github.com/petergtz/pegomock/v4"
	"github.com/runatlantis/atlantis/server/core/boltdb"
	"github.com/runatlantis/atlantis/server/core/db"
	dbmocks "github.com/runatlantis/atlantis/server/core/db/mocks"
	"github.com/runatlantis/atlantis/server/core/locking"
	lockmocks "github.com/runatlantis/atlantis/server/core/locking/mocks"
	"github.com/runatlantis/atlantis/server/core/planstore"
	"github.com/runatlantis/atlantis/server/core/redis"
	"github.com/runatlantis/atlantis/server/core/runtime"
	"github.com/runatlantis/atlantis/server/events"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/logging"
	. "github.com/runatlantis/atlantis/testing"
	"go.uber.org/mock/gomock"
)

// Exercise real locks, saved plan files and project statuses, including named
// projects sharing a directory/workspace and identically named projects elsewhere.
func TestDeleteLocksByProject_Isolation(t *testing.T) {
	for _, backend := range []string{"bolt", "redis"} {
		for _, separatePlans := range []bool{false, true} {
			name := "plans in checkout"
			if separatePlans {
				name = "separate plan directory"
			}
			t.Run(backend+"/"+name, func(t *testing.T) {
				logger := logging.NewNoopLogger(t)
				var database db.Database
				var err error
				if backend == "bolt" {
					database, err = boltdb.New(t.TempDir())
				} else {
					s := miniredis.RunT(t)
					database, err = redis.New(s.Host(), s.Server().Addr().Port, "", false, false, 0)
				}
				Ok(t, err)
				t.Cleanup(func() { database.Close() })
				locker := locking.NewClient(database)
				workingDir := &events.FileWorkspace{DataDir: t.TempDir()}
				planRoot := workingDir.DataDir
				if separatePlans {
					planRoot = t.TempDir()
					workingDir.LocalSharePlanDir = planRoot
				}
				store := &projectUnlockPlanStore{}
				deleter := &events.DefaultDeleteLockCommand{Locker: locker, WorkingDir: workingDir, Database: database, PlanStore: store}
				repo := models.Repo{FullName: "owner/repo", Owner: "owner", Name: "repo"}
				pull := models.PullRequest{Num: 42, BaseRepo: repo}
				otherRepo := models.Repo{FullName: "owner/other", Owner: "owner", Name: "other"}
				locks := []models.ProjectLock{
					{Pull: pull, Workspace: "default", Project: models.Project{RepoFullName: repo.FullName, Path: ".", ProjectName: "keep"}},
					{Pull: pull, Workspace: "default", Project: models.Project{RepoFullName: repo.FullName, Path: ".", ProjectName: "target"}},
					{Pull: pull, Workspace: "staging", Project: models.Project{RepoFullName: repo.FullName, Path: "nested", ProjectName: "other"}},
					{Pull: models.PullRequest{Num: 43, BaseRepo: repo}, Workspace: "production", Project: models.Project{RepoFullName: repo.FullName, Path: ".", ProjectName: "target"}},
					{Pull: models.PullRequest{Num: 42, BaseRepo: otherRepo}, Workspace: "default", Project: models.Project{RepoFullName: otherRepo.FullName, Path: ".", ProjectName: "target"}},
				}
				paths := make([]string, len(locks))
				for i, lock := range locks {
					response, err := locker.TryLock(lock.Project, lock.Workspace, lock.Pull, models.User{})
					Ok(t, err)
					Assert(t, response.LockAcquired, "expected lock acquisition")
					paths[i] = filepath.Join(planstore.PullDir(planRoot, lock.Project.RepoFullName, lock.Pull.Num), lock.Workspace, lock.Project.Path, runtime.GetPlanFilename(lock.Workspace, lock.Project.ProjectName))
					Ok(t, os.MkdirAll(filepath.Dir(paths[i]), 0700))
					Ok(t, os.WriteFile(paths[i], []byte("saved plan"), 0600))
					_, err = database.UpdatePullWithResults(lock.Pull, []command.ProjectResult{{
						Command: command.Plan, Workspace: lock.Workspace, RepoRelDir: lock.Project.Path, ProjectName: lock.Project.ProjectName,
						ProjectCommandOutput: command.ProjectCommandOutput{PlanSuccess: &models.PlanSuccess{}},
					}})
					Ok(t, err)
				}

				// Unknown projects must leave every lock and plan intact.
				count, err := deleter.DeleteLocksByProject(logger, repo.FullName, pull.Num, "missing")
				Equals(t, 0, count)
				ErrEquals(t, `no Atlantis lock found for project "missing" in this PR`, err)
				allLocks, err := locker.List()
				Ok(t, err)
				Equals(t, len(locks), len(allLocks))
				for _, path := range paths {
					_, err := os.Stat(path)
					Ok(t, err)
				}
				Equals(t, 0, len(store.deleted))

				count, err = deleter.DeleteLocksByProject(logger, repo.FullName, pull.Num, "target")
				Ok(t, err)
				Equals(t, 1, count)
				Equals(t, []string{"owner/repo/42/default/./target"}, store.deleted)
				Equals(t, 0, store.deletedPulls)
				for i, lock := range locks {
					remaining, err := locker.GetLock(models.GenerateLockKey(lock.Project, lock.Workspace))
					Ok(t, err)
					_, statErr := os.Stat(paths[i])
					status, err := database.GetPullStatus(lock.Pull)
					Ok(t, err)
					Assert(t, status != nil, "expected pull status")
					for _, project := range status.Projects {
						if project.ProjectName == lock.Project.ProjectName {
							if i == 1 {
								Equals(t, models.DiscardedPlanStatus, project.Status)
							} else {
								Assert(t, project.Status != models.DiscardedPlanStatus, "other project status changed")
							}
						}
					}
					if i == 1 {
						Assert(t, remaining == nil, "target lock remains")
						Assert(t, os.IsNotExist(statErr), "target plan remains")
					} else {
						Assert(t, remaining != nil, "unrelated lock removed")
						Ok(t, statErr)
					}
				}
				// A repeat unlock cannot fall back to deleting the other projects.
				count, err = deleter.DeleteLocksByProject(logger, repo.FullName, pull.Num, "target")
				Equals(t, 0, count)
				ErrEquals(t, `no Atlantis lock found for project "target" in this PR`, err)
				allLocks, err = locker.List()
				Ok(t, err)
				Equals(t, len(locks)-1, len(allLocks))
			})
		}
	}
}

func TestDeleteLocksByProject_Failures(t *testing.T) {
	for _, stage := range []string{"empty name", "list", "unlock", "ownership changed", "local plan", "status"} {
		t.Run(stage, func(t *testing.T) {
			RegisterMockTestingT(t)
			ctrl := gomock.NewController(t)
			locker := lockmocks.NewMockLocker(ctrl)
			database := dbmocks.NewMockDatabase(ctrl)
			workingDir := events.NewMockWorkingDir()
			store := &projectUnlockPlanStore{}
			logger := logging.NewNoopLogger(t)
			lock := models.ProjectLock{
				Pull:      models.PullRequest{Num: 42, BaseRepo: models.Repo{FullName: "owner/repo"}},
				Workspace: "default", Project: models.Project{RepoFullName: "owner/repo", Path: ".", ProjectName: "target"},
			}
			failure := errors.New("test error")
			target := "target"
			expectedCount := 0
			expectedError := ""
			switch stage {
			case "empty name":
				target = ""
				expectedError = "project name must not be empty"
			case "list":
				locker.EXPECT().List().Return(nil, failure)
				expectedError = "listing project locks: test error"
			default:
				locker.EXPECT().List().Return(map[string]models.ProjectLock{"id": lock}, nil)
				unlock := locker.EXPECT().UnlockIfOwnedByPull(lock.Project, lock.Workspace, 42)
				switch stage {
				case "unlock":
					unlock.Return(nil, failure)
					expectedError = `unlocking project "target": test error`
				case "ownership changed":
					unlock.Return(nil, nil)
					expectedError = `no Atlantis lock found for project "target" in this PR`
				default:
					unlock.Return(&lock, nil)
					expectedCount = 1
					if stage == "local plan" {
						When(workingDir.DeletePlan(Any[logging.SimpleLogging](), Eq(lock.Pull.BaseRepo), Eq(lock.Pull), Eq(lock.Workspace), Eq("."), Eq("target"))).ThenReturn(failure)
						expectedError = `discarding plan for project "target": test error`
					} else {
						database.EXPECT().UpdateProjectStatus(lock.Pull, lock.Workspace, ".", "target", models.DiscardedPlanStatus).Return(failure)
						expectedError = `updating status for project "target": test error`
					}
				}
			}
			deleter := events.DefaultDeleteLockCommand{Locker: locker, WorkingDir: workingDir, Database: database, PlanStore: store}
			count, err := deleter.DeleteLocksByProject(logger, "owner/repo", 42, target)
			Equals(t, expectedCount, count)
			ErrEquals(t, expectedError, err)
			if expectedCount == 0 {
				workingDir.VerifyWasCalled(Never()).DeletePlan(Any[logging.SimpleLogging](), Any[models.Repo](), Any[models.PullRequest](), Any[string](), Any[string](), Any[string]())
			}
			if stage != "status" {
				Equals(t, 0, len(store.deleted))
			}
			Equals(t, 0, store.deletedPulls)
		})
	}
}

// Embed the local implementation so only deletion needs instrumentation.
type projectUnlockPlanStore struct {
	planstore.LocalPlanStore
	deleted      []string
	deletedPulls int
}

func (s *projectUnlockPlanStore) DeletePlanForProject(owner, repo string, pullNum int, workspace, repoRelDir, projectName string) error {
	s.deleted = append(s.deleted, fmt.Sprintf("%s/%s/%d/%s/%s/%s", owner, repo, pullNum, workspace, repoRelDir, projectName))
	return nil
}

func (s *projectUnlockPlanStore) DeleteForPull(string, string, int) error {
	s.deletedPulls++
	return nil
}
