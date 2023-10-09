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
	RegisterMockTestingT(t)
	l := lockmocks.NewMockLocker()
	When(l.Unlock("id")).ThenReturn(nil, errors.New("err"))
	dlc := events.DefaultDeleteLockCommand{
		Locker: l,
		Logger: logging.NewNoopLogger(t),
	}
	_, err := dlc.DeleteLock("id")
	ErrEquals(t, "err", err)
}

func TestDeleteLock_None(t *testing.T) {
	t.Log("If there is no lock at that ID we return nil")
	RegisterMockTestingT(t)
	l := lockmocks.NewMockLocker()
	When(l.Unlock("id")).ThenReturn(nil, nil)
	dlc := events.DefaultDeleteLockCommand{
		Locker: l,
		Logger: logging.NewNoopLogger(t),
	}
	lock, err := dlc.DeleteLock("id")
	Ok(t, err)
	Assert(t, lock == nil, "lock was not nil")
}

func TestDeleteLock_Success(t *testing.T) {
	t.Log("Delete lock deletes successfully the plan file")
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
		Logger:           logging.NewNoopLogger(t),
		Backend:          db,
		WorkingDirLocker: workingDirLocker,
		WorkingDir:       workingDir,
	}
	lock, err := dlc.DeleteLock("id")
	Ok(t, err)
	Assert(t, lock != nil, "lock was nil")
	workingDir.VerifyWasCalledOnce().DeletePlan(pull.BaseRepo, pull, workspace, path, projectName)
}

func TestDeleteLocksByPull_LockerErr(t *testing.T) {
	t.Log("If there is an error retrieving the lock, returned a failed status")
	repoName := "reponame"
	pullNum := 2
	RegisterMockTestingT(t)
	l := lockmocks.NewMockLocker()
	workingDir := events.NewMockWorkingDir()
	When(l.UnlockByPull(repoName, pullNum)).ThenReturn(nil, errors.New("err"))
	dlc := events.DefaultDeleteLockCommand{
		Locker:     l,
		Logger:     logging.NewNoopLogger(t),
		WorkingDir: workingDir,
	}
	_, err := dlc.DeleteLocksByPull(repoName, pullNum)
	ErrEquals(t, "err", err)
	workingDir.VerifyWasCalled(Never()).DeletePlan(Any[models.Repo](), Any[models.PullRequest](), Any[string](), Any[string](), Any[string]())
}

func TestDeleteLocksByPull_None(t *testing.T) {
	t.Log("If there is no lock at that ID there is no error")
	repoName := "reponame"
	pullNum := 2
	RegisterMockTestingT(t)
	l := lockmocks.NewMockLocker()
	workingDir := events.NewMockWorkingDir()
	When(l.UnlockByPull(repoName, pullNum)).ThenReturn([]models.ProjectLock{}, nil)
	dlc := events.DefaultDeleteLockCommand{
		Locker:     l,
		Logger:     logging.NewNoopLogger(t),
		WorkingDir: workingDir,
	}
	_, err := dlc.DeleteLocksByPull(repoName, pullNum)
	Ok(t, err)
	workingDir.VerifyWasCalled(Never()).DeletePlan(Any[models.Repo](), Any[models.PullRequest](), Any[string](), Any[string](), Any[string]())
}

func TestDeleteLocksByPull_SingleSuccess(t *testing.T) {
	t.Log("If a single lock is successfully deleted")
	repoName := "reponame"
	pullNum := 2
	path := "."
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
				Path:         path,
				RepoFullName: pull.BaseRepo.FullName,
			},
		},
	}, nil,
	)
	dlc := events.DefaultDeleteLockCommand{
		Locker:     l,
		Logger:     logging.NewNoopLogger(t),
		WorkingDir: workingDir,
	}
	_, err := dlc.DeleteLocksByPull(repoName, pullNum)
	Ok(t, err)
	workingDir.VerifyWasCalled(Once()).DeletePlan(pull.BaseRepo, pull, workspace, path, projectName)
}

func TestDeleteLocksByPull_MultipleSuccess(t *testing.T) {
	t.Log("If multiple locks are successfully deleted")
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
		Logger:     logging.NewNoopLogger(t),
		WorkingDir: workingDir,
	}
	_, err := dlc.DeleteLocksByPull(repoName, pullNum)
	Ok(t, err)
	workingDir.VerifyWasCalled(Once()).DeletePlan(pull.BaseRepo, pull, workspace, path1, projectName)
	workingDir.VerifyWasCalled(Once()).DeletePlan(pull.BaseRepo, pull, workspace, path2, projectName)
}
