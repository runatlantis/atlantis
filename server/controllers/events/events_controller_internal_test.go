package events

import (
	"testing"
	"time"

	coreevents "github.com/runatlantis/atlantis/server/events"
	"github.com/runatlantis/atlantis/server/events/models"
)

type blockingAutoplanRunner struct {
	started chan models.PullRequest
	release chan struct{}
	done    chan models.PullRequest
}

func (r *blockingAutoplanRunner) RunCommentCommand(models.Repo, *models.Repo, *models.PullRequest, models.User, int, *coreevents.CommentCommand) {
}

func (r *blockingAutoplanRunner) RunAutoplanCommand(_ models.Repo, _ models.Repo, pull models.PullRequest, _ models.User) {
	r.started <- pull
	if pull.HeadCommit == "first" {
		<-r.release
	}
	r.done <- pull
}

func TestHandlePullRequestEventCoalescesAndSerializesAutoplans(t *testing.T) {
	allowlist, err := coreevents.NewRepoAllowlistChecker("*")
	if err != nil {
		t.Fatal(err)
	}
	runner := &blockingAutoplanRunner{
		started: make(chan models.PullRequest, 3),
		release: make(chan struct{}),
		done:    make(chan models.PullRequest, 3),
	}
	controller := VCSEventsController{
		CommandRunner:        runner,
		RepoAllowlistChecker: allowlist,
		TestingMode:          false,
		AutoplanRuns:         NewAutoplanRunCoordinator(),
	}
	baseRepo := models.Repo{FullName: "owner/repo", VCSHost: models.VCSHost{Type: models.Github, Hostname: "github.com"}}
	first := models.PullRequest{BaseRepo: baseRepo, Num: 1, HeadCommit: "first"}
	second := models.PullRequest{BaseRepo: baseRepo, Num: 1, HeadCommit: "second"}
	third := models.PullRequest{BaseRepo: baseRepo, Num: 1, HeadCommit: "third"}

	controller.handlePullRequestEvent(nil, baseRepo, baseRepo, first, models.User{}, models.OpenedPullEvent)
	expectAutoplan(t, runner.started, "first")
	controller.handlePullRequestEvent(nil, baseRepo, baseRepo, first, models.User{}, models.UpdatedPullEvent)
	controller.handlePullRequestEvent(nil, baseRepo, baseRepo, second, models.User{}, models.UpdatedPullEvent)
	controller.handlePullRequestEvent(nil, baseRepo, baseRepo, third, models.User{}, models.UpdatedPullEvent)
	expectNoAutoplan(t, runner.started)

	close(runner.release)
	expectAutoplan(t, runner.done, "first")
	expectAutoplan(t, runner.started, "third")
	expectAutoplan(t, runner.done, "third")

	controller.handlePullRequestEvent(nil, baseRepo, baseRepo, third, models.User{}, models.UpdatedPullEvent)
	expectAutoplan(t, runner.started, "third")
}

func TestAutoplanRunCoordinatorSeparatesVCSHosts(t *testing.T) {
	coordinator := NewAutoplanRunCoordinator()
	github := autoplanRequest{
		baseRepo: models.Repo{FullName: "owner/repo", VCSHost: models.VCSHost{Type: models.Github, Hostname: "github.com"}},
		pull:     models.PullRequest{Num: 1, HeadCommit: "same-commit"},
	}
	githubEnterprise := autoplanRequest{
		baseRepo: models.Repo{FullName: "owner/repo", VCSHost: models.VCSHost{Type: models.Github, Hostname: "github.example.com"}},
		pull:     models.PullRequest{Num: 1, HeadCommit: "same-commit"},
	}

	if _, started := coordinator.start(github); !started {
		t.Fatal("expected github.com autoplan to start")
	}
	if _, started := coordinator.start(githubEnterprise); !started {
		t.Fatal("expected GitHub Enterprise autoplan to start")
	}
}

func expectAutoplan(t *testing.T, runs <-chan models.PullRequest, commit string) {
	t.Helper()
	select {
	case pull := <-runs:
		if pull.HeadCommit != commit {
			t.Fatalf("got %q, want %q", pull.HeadCommit, commit)
		}
	case <-time.After(time.Second):
		t.Fatalf("timed out waiting for %q", commit)
	}
}

func expectNoAutoplan(t *testing.T, runs <-chan models.PullRequest) {
	t.Helper()
	select {
	case pull := <-runs:
		t.Fatalf("unexpected autoplan for %q", pull.HeadCommit)
	default:
	}
}
