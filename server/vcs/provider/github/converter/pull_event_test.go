package converter_test

import (
	"github.com/runatlantis/atlantis/server/vcs"
	"testing"

	"github.com/google/go-github/v45/github"
	"github.com/mohae/deepcopy"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/vcs/provider/github/converter"
	. "github.com/runatlantis/atlantis/testing"
)

var PullEvent = github.PullRequestEvent{
	Sender: &github.User{
		Login: github.String("user"),
	},
	Repo:        &Repo,
	PullRequest: &Pull,
	Action:      github.String("opened"),
	Installation: &github.Installation{
		ID: github.Int64(1),
	},
}

func TestConvert_PullRequestEvent(t *testing.T) {
	repoConverter := converter.RepoConverter{
		GithubUser:  "github-user",
		GithubToken: "github-token",
	}
	subject := converter.PullEventConverter{
		PullConverter: converter.PullConverter{RepoConverter: repoConverter},
	}
	_, err := subject.Convert(&github.PullRequestEvent{})
	ErrEquals(t, "pull_request is null", err)

	testEvent := deepcopy.Copy(PullEvent).(github.PullRequestEvent)
	testEvent.PullRequest.HTMLURL = nil
	_, err = subject.Convert(&testEvent)
	ErrEquals(t, "html_url is null", err)

	testEvent = deepcopy.Copy(PullEvent).(github.PullRequestEvent)
	testEvent.Sender = nil
	_, err = subject.Convert(&testEvent)
	ErrEquals(t, "sender is null", err)

	testEvent = deepcopy.Copy(PullEvent).(github.PullRequestEvent)
	testEvent.Sender.Login = nil
	_, err = subject.Convert(&testEvent)
	ErrEquals(t, "sender.login is null", err)

	actPull, err := subject.Convert(&PullEvent)
	Ok(t, err)
	expBaseRepo := models.Repo{
		Owner:             "owner",
		FullName:          "owner/repo",
		CloneURL:          "https://github-user:github-token@github.com/owner/repo.git",
		SanitizedCloneURL: "https://github-user:<redacted>@github.com/owner/repo.git",
		Name:              "repo",
		VCSHost: models.VCSHost{
			Hostname: "github.com",
			Type:     models.Github,
		},
		DefaultBranch: "main",
	}
	Equals(t, models.PullRequest{
		URL:        Pull.GetHTMLURL(),
		Author:     Pull.User.GetLogin(),
		HeadBranch: Pull.Head.GetRef(),
		BaseBranch: Pull.Base.GetRef(),
		HeadCommit: Pull.Head.GetSHA(),
		Num:        Pull.GetNumber(),
		State:      models.OpenPullState,
		BaseRepo:   expBaseRepo,
		HeadRepo:   expBaseRepo,
		UpdatedAt:  timestamp,
		HeadRef: vcs.Ref{
			Type: "branch",
			Name: "ref",
		},
	}, actPull.Pull)
	Equals(t, models.OpenedPullEvent, actPull.EventType)
	Equals(t, models.User{Username: "user"}, actPull.User)
	Equals(t, int64(1), actPull.InstallationToken)
}

func TestConvert_PullRequestEvent_Draft(t *testing.T) {
	repoConverter := converter.RepoConverter{
		GithubUser:  "github-user",
		GithubToken: "github-token",
	}
	subject := converter.PullEventConverter{
		PullConverter: converter.PullConverter{RepoConverter: repoConverter},
	}

	// verify that draft PRs are treated as 'other' events by default
	testEvent := deepcopy.Copy(PullEvent).(github.PullRequestEvent)
	draftPR := true
	testEvent.PullRequest.Draft = &draftPR
	pull, err := subject.Convert(&testEvent)
	Ok(t, err)
	Equals(t, models.OtherPullEvent, pull.EventType)
	// verify that drafts are planned if requested
	subject.AllowDraftPRs = true
	defer func() { subject.AllowDraftPRs = false }()
	pull, err = subject.Convert(&testEvent)
	Ok(t, err)
	Equals(t, models.OpenedPullEvent, pull.EventType)
	Equals(t, int64(1), pull.InstallationToken)
}

func TestConvert_PullRequestEvent_EventType(t *testing.T) {
	repoConverter := converter.RepoConverter{
		GithubUser:  "github-user",
		GithubToken: "github-token",
	}
	subject := converter.PullEventConverter{
		PullConverter: converter.PullConverter{RepoConverter: repoConverter},
	}
	cases := []struct {
		action   string
		exp      models.PullRequestEventType
		draftExp models.PullRequestEventType
	}{
		{
			action:   "synchronize",
			exp:      models.UpdatedPullEvent,
			draftExp: models.OtherPullEvent,
		},
		{
			action:   "unassigned",
			exp:      models.OtherPullEvent,
			draftExp: models.OtherPullEvent,
		},
		{
			action:   "review_requested",
			exp:      models.OtherPullEvent,
			draftExp: models.OtherPullEvent,
		},
		{
			action:   "review_request_removed",
			exp:      models.OtherPullEvent,
			draftExp: models.OtherPullEvent,
		},
		{
			action:   "labeled",
			exp:      models.OtherPullEvent,
			draftExp: models.OtherPullEvent,
		},
		{
			action:   "unlabeled",
			exp:      models.OtherPullEvent,
			draftExp: models.OtherPullEvent,
		},
		{
			action:   "opened",
			exp:      models.OpenedPullEvent,
			draftExp: models.OtherPullEvent,
		},
		{
			action:   "edited",
			exp:      models.OtherPullEvent,
			draftExp: models.OtherPullEvent,
		},
		{
			action:   "closed",
			exp:      models.ClosedPullEvent,
			draftExp: models.ClosedPullEvent,
		},
		{
			action:   "reopened",
			exp:      models.OtherPullEvent,
			draftExp: models.OtherPullEvent,
		},
		{
			action:   "ready_for_review",
			exp:      models.OpenedPullEvent,
			draftExp: models.OtherPullEvent,
		},
	}

	for _, c := range cases {
		t.Run(c.action, func(t *testing.T) {
			// Test normal parsing
			event := deepcopy.Copy(PullEvent).(github.PullRequestEvent)
			event.Action = &c.action
			pull, err := subject.Convert(&event)
			Ok(t, err)
			Equals(t, c.exp, pull.EventType)
			// Test draft parsing when draft PRs disabled
			draftPR := true
			event.PullRequest.Draft = &draftPR
			pull, err = subject.Convert(&event)
			Ok(t, err)
			Equals(t, c.draftExp, pull.EventType)

			subjectDraft := converter.PullEventConverter{
				PullConverter: converter.PullConverter{RepoConverter: repoConverter},
			}
			// Test draft parsing when draft PRs are enabled.
			subjectDraft.AllowDraftPRs = true
			pull, err = subjectDraft.Convert(&event)
			Ok(t, err)
			Equals(t, int64(1), pull.InstallationToken)
			Equals(t, c.exp, pull.EventType)
		})
	}
}
