package events_test

import (
	"errors"
	"testing"

	. "github.com/petergtz/pegomock/v4"
	"github.com/runatlantis/atlantis/server/core/db"
	lockmocks "github.com/runatlantis/atlantis/server/core/locking/mocks"
	"github.com/runatlantis/atlantis/server/events"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/logging"
	. "github.com/runatlantis/atlantis/testing"
)

func TestDeleteLock_LockerErr(t *testing.T) {
	t.Log("If there is an error retrieving the lock, we return the error")
	logger := logging.NewNoopLogger(t)
	RegisterMockTestingT(t)
	l := lockmocks.NewMockLocker()
	When(l.Unlock("id")).ThenReturn(nil, errors.New("err"))
	dlc := events.DefaultDeleteLockCommand{Locker: l}
	_, err := dlc.DeleteLock(logger, "id")
	ErrEquals(t, "err", err)
}

func TestDeleteLock_None(t *testing.T) {
	t.Log("If there is no lock at that ID we return nil")
	logger := logging.NewNoopLogger(t)
	RegisterMockTestingT(t)
	l := lockmocks.NewMockLocker()
	When(l.Unlock("id")).ThenReturn(nil, nil)
	dlc := events.DefaultDeleteLockCommand{Locker: l}
	lock, err := dlc.DeleteLock(logger, "id")
	Ok(t, err)
	Assert(t, lock == nil, "lock was not nil")
}

func TestDeleteLock_Success(t *testing.T) {
	t.Log("Delete lock deletes successfully the plan file")
	logger := logging.NewNoopLogger(t)
	RegisterMockTestingT(t)
	l := lockmocks.NewMockLocker()
	When(l.Unlock("id")).ThenReturn(&models.ProjectLock{}, nil)
	workingDir := events.NewMockWorkingDir()
	workingDirLocker := events.NewDefaultWorkingDirLocker()
	workspace := "workspace"
	path := "path"
	projectName := ""
	pull := models.PullRequest{
		BaseRepo: models.Repo{FullName: "owner/repo"},
	}
	When(l.Unlock("id")).ThenReturn(&models.ProjectLock{
		Pull:      pull,
		Workspace: workspace,
		Project: models.Project{
			Path:         path,
			RepoFullName: pull.BaseRepo.FullName,
		},
	}, nil)
	tmp := t.TempDir()
	db, err := db.New(tmp)
	Ok(t, err)
	dlc := events.DefaultDeleteLockCommand{
		Locker:           l,
		Backend:          db,
		WorkingDirLocker: workingDirLocker,
		WorkingDir:       workingDir,
	}
	lock, err := dlc.DeleteLock(logger, "id")
	Ok(t, err)
	Assert(t, lock != nil, "lock was nil")
	workingDir.VerifyWasCalledOnce().DeletePlan(Any[logging.SimpleLogging](), Eq(pull.BaseRepo), Eq(pull), Eq(workspace),
		Eq(path), Eq(projectName))
}

func TestDeleteLocksByPull_LockerErr(t *testing.T) {
	t.Log("If there is an error retrieving the lock, returned a failed status")
	logger := logging.NewNoopLogger(t)
	repoName := "reponame"
	pullNum := 2
	RegisterMockTestingT(t)
	l := lockmocks.NewMockLocker()
	workingDir := events.NewMockWorkingDir()
	When(l.UnlockByPull(repoName, pullNum)).ThenReturn(nil, errors.New("err"))
	dlc := events.DefaultDeleteLockCommand{
		Locker:     l,
		WorkingDir: workingDir,
	}
	_, err := dlc.DeleteLocksByPull(logger, repoName, pullNum)
	ErrEquals(t, "err", err)
	workingDir.VerifyWasCalled(Never()).DeletePlan(Any[logging.SimpleLogging](), Any[models.Repo](), Any[models.PullRequest](),
		Any[string](), Any[string](), Any[string]())
}

func TestDeleteLocksByPull_None(t *testing.T) {
	t.Log("If there is no lock at that ID there is no error")
	logger := logging.NewNoopLogger(t)
	repoName := "reponame"
	pullNum := 2
	RegisterMockTestingT(t)
	l := lockmocks.NewMockLocker()
	workingDir := events.NewMockWorkingDir()
	When(l.UnlockByPull(repoName, pullNum)).ThenReturn([]models.ProjectLock{}, nil)
	dlc := events.DefaultDeleteLockCommand{
		Locker:     l,
		WorkingDir: workingDir,
	}
	_, err := dlc.DeleteLocksByPull(logger, repoName, pullNum)
	Ok(t, err)
	workingDir.VerifyWasCalled(Never()).DeletePlan(Any[logging.SimpleLogging](), Any[models.Repo](), Any[models.PullRequest](),
		Any[string](), Any[string](), Any[string]())
}

func TestDeleteLocksByPull_SingleSuccess(t *testing.T) {
	t.Log("If a single lock is successfully deleted")
	logger := logging.NewNoopLogger(t)
	repoName := "reponame"
	pullNum := 2
	path := "."
	workspace := "default"
	projectName := "projectname"

	RegisterMockTestingT(t)
	l := lockmocks.NewMockLocker()
	workingDir := events.NewMockWorkingDir()
	pull := models.PullRequest{
		BaseRepo: models.Repo{FullName: repoName},
		Num:      pullNum,
	}
	When(l.UnlockByPull(repoName, pullNum)).ThenReturn([]models.ProjectLock{
		{
			Pull:      pull,
			Workspace: workspace,
			Project: models.Project{
				Path:         path,
				RepoFullName: pull.BaseRepo.FullName,
				ProjectName:  projectName,
			},
		},
	}, nil,
	)
	dlc := events.DefaultDeleteLockCommand{
		Locker:     l,
		WorkingDir: workingDir,
	}
	_, err := dlc.DeleteLocksByPull(logger, repoName, pullNum)
	Ok(t, err)
	workingDir.VerifyWasCalled(Once()).DeletePlan(Any[logging.SimpleLogging](), Eq(pull.BaseRepo), Eq(pull), Eq(workspace),
		Eq(path), Eq(projectName))
}

func TestDeleteLocksByPull_MultipleSuccess(t *testing.T) {
	t.Log("If multiple locks are successfully deleted")
	logger := logging.NewNoopLogger(t)
	repoName := "reponame"
	pullNum := 2
	path1 := "path1"
	path2 := "path2"
	workspace := "default"
	projectName := ""

	RegisterMockTestingT(t)
	l := lockmocks.NewMockLocker()
	workingDir := events.NewMockWorkingDir()
	pull := models.PullRequest{
		BaseRepo: models.Repo{FullName: repoName},
		Num:      pullNum,
	}
	When(l.UnlockByPull(repoName, pullNum)).ThenReturn([]models.ProjectLock{
		{
			Pull:      pull,
			Workspace: workspace,
			Project: models.Project{
				Path:         path1,
				RepoFullName: pull.BaseRepo.FullName,
			},
		},
		{
			Pull:      pull,
			Workspace: workspace,
			Project: models.Project{
				Path:         path2,
				RepoFullName: pull.BaseRepo.FullName,
			},
		},
	}, nil,
	)
	dlc := events.DefaultDeleteLockCommand{
		Locker:     l,
		WorkingDir: workingDir,
	}
	_, err := dlc.DeleteLocksByPull(logger, repoName, pullNum)
	Ok(t, err)
	workingDir.VerifyWasCalled(Once()).DeletePlan(logger, pull.BaseRepo, pull, workspace, path1, projectName)
	workingDir.VerifyWasCalled(Once()).DeletePlan(logger, pull.BaseRepo, pull, workspace, path2, projectName)
}
