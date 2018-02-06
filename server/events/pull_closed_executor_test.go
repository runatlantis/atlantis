package events_test

import (
	"errors"
	"testing"

	"github.com/atlantisnorth/atlantis/server/events"
	lockmocks "github.com/atlantisnorth/atlantis/server/events/locking/mocks"
	"github.com/atlantisnorth/atlantis/server/events/mocks"
	"github.com/atlantisnorth/atlantis/server/events/mocks/matchers"
	"github.com/atlantisnorth/atlantis/server/events/models"
	"github.com/atlantisnorth/atlantis/server/events/models/fixtures"
	"github.com/atlantisnorth/atlantis/server/events/vcs"
	vcsmocks "github.com/atlantisnorth/atlantis/server/events/vcs/mocks"
	. "github.com/atlantisnorth/atlantis/testing"
	. "github.com/petergtz/pegomock"
)

func TestCleanUpPullWorkspaceErr(t *testing.T) {
	t.Log("when workspace.Delete returns an error, we return it")
	RegisterMockTestingT(t)
	w := mocks.NewMockAtlantisWorkspace()
	pce := events.PullClosedExecutor{
		Workspace: w,
	}
	err := errors.New("err")
	When(w.Delete(fixtures.Repo, fixtures.Pull)).ThenReturn(err)
	actualErr := pce.CleanUpPull(fixtures.Repo, fixtures.Pull, vcs.Github)
	Equals(t, "cleaning workspace: err", actualErr.Error())
}

func TestCleanUpPullUnlockErr(t *testing.T) {
	t.Log("when locker.UnlockByPull returns an error, we return it")
	RegisterMockTestingT(t)
	w := mocks.NewMockAtlantisWorkspace()
	l := lockmocks.NewMockLocker()
	pce := events.PullClosedExecutor{
		Locker:    l,
		Workspace: w,
	}
	err := errors.New("err")
	When(l.UnlockByPull(fixtures.Repo.FullName, fixtures.Pull.Num)).ThenReturn(nil, err)
	actualErr := pce.CleanUpPull(fixtures.Repo, fixtures.Pull, vcs.Github)
	Equals(t, "cleaning up locks: err", actualErr.Error())
}

func TestCleanUpPullNoLocks(t *testing.T) {
	t.Log("when there are no locks to clean up, we don't comment")
	RegisterMockTestingT(t)
	w := mocks.NewMockAtlantisWorkspace()
	l := lockmocks.NewMockLocker()
	cp := vcsmocks.NewMockClientProxy()
	pce := events.PullClosedExecutor{
		Locker:    l,
		VCSClient: cp,
		Workspace: w,
	}
	When(l.UnlockByPull(fixtures.Repo.FullName, fixtures.Pull.Num)).ThenReturn(nil, nil)
	err := pce.CleanUpPull(fixtures.Repo, fixtures.Pull, vcs.Github)
	Ok(t, err)
	cp.VerifyWasCalled(Never()).CreateComment(matchers.AnyModelsRepo(), matchers.AnyModelsPullRequest(), AnyString(), matchers.AnyVcsHost())
}

func TestCleanUpPullComments(t *testing.T) {
	t.Log("should comment correctly")
	RegisterMockTestingT(t)
	cases := []struct {
		Description string
		Locks       []models.ProjectLock
		Exp         string
	}{
		{
			"single lock, empty path",
			[]models.ProjectLock{
				{
					Project:   models.NewProject("owner/repo", ""),
					Workspace: "default",
				},
			},
			"- path: `owner/repo/.` workspace: `default`",
		},
		{
			"single lock, non-empty path",
			[]models.ProjectLock{
				{
					Project:   models.NewProject("owner/repo", "path"),
					Workspace: "default",
				},
			},
			"- path: `owner/repo/path` workspace: `default`",
		},
		{
			"single path, multiple workspaces",
			[]models.ProjectLock{
				{
					Project:   models.NewProject("owner/repo", "path"),
					Workspace: "workspace1",
				},
				{
					Project:   models.NewProject("owner/repo", "path"),
					Workspace: "workspace2",
				},
			},
			"- path: `owner/repo/path` workspaces: `workspace1`, `workspace2`",
		},
		{
			"multiple paths, multiple workspaces",
			[]models.ProjectLock{
				{
					Project:   models.NewProject("owner/repo", "path"),
					Workspace: "workspace1",
				},
				{
					Project:   models.NewProject("owner/repo", "path"),
					Workspace: "workspace2",
				},
				{
					Project:   models.NewProject("owner/repo", "path2"),
					Workspace: "workspace1",
				},
				{
					Project:   models.NewProject("owner/repo", "path2"),
					Workspace: "workspace2",
				},
			},
			"- path: `owner/repo/path` workspaces: `workspace1`, `workspace2`\n- path: `owner/repo/path2` workspaces: `workspace1`, `workspace2`",
		},
	}
	for _, c := range cases {
		w := mocks.NewMockAtlantisWorkspace()
		cp := vcsmocks.NewMockClientProxy()
		l := lockmocks.NewMockLocker()
		pce := events.PullClosedExecutor{
			Locker:    l,
			VCSClient: cp,
			Workspace: w,
		}
		t.Log("testing: " + c.Description)
		When(l.UnlockByPull(fixtures.Repo.FullName, fixtures.Pull.Num)).ThenReturn(c.Locks, nil)
		err := pce.CleanUpPull(fixtures.Repo, fixtures.Pull, vcs.Github)
		Ok(t, err)
		_, _, comment, _ := cp.VerifyWasCalledOnce().CreateComment(matchers.AnyModelsRepo(), matchers.AnyModelsPullRequest(), AnyString(), matchers.AnyVcsHost()).GetCapturedArguments()

		expected := "Locks and plans deleted for the projects and workspaces modified in this pull request:\n\n" + c.Exp
		Equals(t, expected, comment)
	}
}
