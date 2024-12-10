// Copyright 2017 HootSuite Media Inc.
//
// Licensed under the Apache License, Version 2.0 (the License);
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//    http://www.apache.org/licenses/LICENSE-2.0
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an AS IS BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
// Modified hereafter by contributors to runatlantis/atlantis.

package events_test

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/go-github/v66/github"
	"github.com/mcdafydd/go-azuredevops/azuredevops"
	"github.com/mohae/deepcopy"
	"github.com/runatlantis/atlantis/server/events"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/models"
	. "github.com/runatlantis/atlantis/server/events/vcs/testdata"
	"github.com/runatlantis/atlantis/server/logging"
	. "github.com/runatlantis/atlantis/testing"
	gitlab "github.com/xanzy/go-gitlab"
)

var parser = events.EventParser{
	GithubUser:         "github-user",
	GithubToken:        "github-token",
	GithubTokenFile:    "",
	GitlabUser:         "gitlab-user",
	GitlabToken:        "gitlab-token",
	AllowDraftPRs:      false,
	BitbucketUser:      "bitbucket-user",
	BitbucketToken:     "bitbucket-token",
	BitbucketServerURL: "http://mycorp.com:7490",
	AzureDevopsUser:    "azuredevops-user",
	AzureDevopsToken:   "azuredevops-token",
}

func TestParseGithubRepo(t *testing.T) {
	r, err := parser.ParseGithubRepo(&Repo)
	Ok(t, err)
	Equals(t, models.Repo{
		Owner:             "owner",
		FullName:          "owner/repo",
		CloneURL:          "https://github-user:github-token@github.com/owner/repo.git",
		SanitizedCloneURL: "https://github-user:<redacted>@github.com/owner/repo.git",
		Name:              "repo",
		VCSHost: models.VCSHost{
			Hostname: "github.com",
			Type:     models.Github,
		},
	}, r)
}

func TestParseGithubIssueCommentEvent(t *testing.T) {
	logger := logging.NewNoopLogger(t)
	comment := github.IssueCommentEvent{
		Repo: &Repo,
		Issue: &github.Issue{
			Number:  github.Int(1),
			User:    &github.User{Login: github.String("issue_user")},
			HTMLURL: github.String("https://github.com/runatlantis/atlantis/issues/1"),
		},
		Comment: &github.IssueComment{
			User: &github.User{Login: github.String("comment_user")},
		},
	}

	testComment := deepcopy.Copy(comment).(github.IssueCommentEvent)
	testComment.Comment = nil
	_, _, _, err := parser.ParseGithubIssueCommentEvent(logger, &testComment)
	ErrEquals(t, "comment.user.login is null", err)

	testComment = deepcopy.Copy(comment).(github.IssueCommentEvent)
	testComment.Comment.User = nil
	_, _, _, err = parser.ParseGithubIssueCommentEvent(logger, &testComment)
	ErrEquals(t, "comment.user.login is null", err)

	testComment = deepcopy.Copy(comment).(github.IssueCommentEvent)
	testComment.Comment.User.Login = nil
	_, _, _, err = parser.ParseGithubIssueCommentEvent(logger, &testComment)
	ErrEquals(t, "comment.user.login is null", err)

	testComment = deepcopy.Copy(comment).(github.IssueCommentEvent)
	testComment.Issue = nil
	_, _, _, err = parser.ParseGithubIssueCommentEvent(logger, &testComment)
	ErrEquals(t, "issue.number is null", err)

	// this should be successful
	repo, user, pullNum, err := parser.ParseGithubIssueCommentEvent(logger, &comment)
	Ok(t, err)
	Equals(t, models.Repo{
		Owner:             *comment.Repo.Owner.Login,
		FullName:          *comment.Repo.FullName,
		CloneURL:          "https://github-user:github-token@github.com/owner/repo.git",
		SanitizedCloneURL: "https://github-user:<redacted>@github.com/owner/repo.git",
		Name:              "repo",
		VCSHost: models.VCSHost{
			Hostname: "github.com",
			Type:     models.Github,
		},
	}, repo)
	Equals(t, models.User{
		Username: *comment.Comment.User.Login,
	}, user)
	Equals(t, *comment.Issue.Number, pullNum)
}

func TestParseGithubPullEvent(t *testing.T) {
	logger := logging.NewNoopLogger(t)
	_, _, _, _, _, err := parser.ParseGithubPullEvent(logger, &github.PullRequestEvent{})
	ErrEquals(t, "pull_request is null", err)

	testEvent := deepcopy.Copy(PullEvent).(github.PullRequestEvent)
	testEvent.PullRequest.HTMLURL = nil
	_, _, _, _, _, err = parser.ParseGithubPullEvent(logger, &testEvent)
	ErrEquals(t, "html_url is null", err)

	testEvent = deepcopy.Copy(PullEvent).(github.PullRequestEvent)
	testEvent.Sender = nil
	_, _, _, _, _, err = parser.ParseGithubPullEvent(logger, &testEvent)
	ErrEquals(t, "sender is null", err)

	testEvent = deepcopy.Copy(PullEvent).(github.PullRequestEvent)
	testEvent.Sender.Login = nil
	_, _, _, _, _, err = parser.ParseGithubPullEvent(logger, &testEvent)
	ErrEquals(t, "sender.login is null", err)

	actPull, evType, actBaseRepo, actHeadRepo, actUser, err := parser.ParseGithubPullEvent(logger, &PullEvent)
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
	}
	Equals(t, expBaseRepo, actBaseRepo)
	Equals(t, expBaseRepo, actHeadRepo)
	Equals(t, models.PullRequest{
		URL:        Pull.GetHTMLURL(),
		Author:     Pull.User.GetLogin(),
		HeadBranch: Pull.Head.GetRef(),
		BaseBranch: Pull.Base.GetRef(),
		HeadCommit: Pull.Head.GetSHA(),
		Num:        Pull.GetNumber(),
		State:      models.OpenPullState,
		BaseRepo:   expBaseRepo,
	}, actPull)
	Equals(t, models.OpenedPullEvent, evType)
	Equals(t, models.User{Username: "user"}, actUser)
}

func TestParseGithubPullEventFromDraft(t *testing.T) {
	logger := logging.NewNoopLogger(t)
	// verify that close event treated as 'close' events by default
	closeEvent := deepcopy.Copy(PullEvent).(github.PullRequestEvent)
	closeEvent.Action = github.String("closed")
	closeEvent.PullRequest.Draft = github.Bool(true)

	_, evType, _, _, _, err := parser.ParseGithubPullEvent(logger, &closeEvent)
	Ok(t, err)
	Equals(t, models.ClosedPullEvent, evType)

	// verify that draft PRs are treated as 'other' events by default
	testEvent := deepcopy.Copy(PullEvent).(github.PullRequestEvent)
	testEvent.PullRequest.Draft = github.Bool(true)
	_, evType, _, _, _, err = parser.ParseGithubPullEvent(logger, &testEvent)
	Ok(t, err)
	Equals(t, models.OtherPullEvent, evType)
	// verify that drafts are planned if requested
	parser.AllowDraftPRs = true
	defer func() { parser.AllowDraftPRs = false }()
	_, evType, _, _, _, err = parser.ParseGithubPullEvent(logger, &testEvent)
	Ok(t, err)
	Equals(t, models.OpenedPullEvent, evType)
}

func TestParseGithubPullEvent_EventType(t *testing.T) {
	logger := logging.NewNoopLogger(t)
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
			action := c.action
			event.Action = &action
			_, actType, _, _, _, err := parser.ParseGithubPullEvent(logger, &event)
			Ok(t, err)
			Equals(t, c.exp, actType)
			// Test draft parsing when draft PRs disabled
			draftPR := true
			event.PullRequest.Draft = &draftPR
			_, draftEvType, _, _, _, err := parser.ParseGithubPullEvent(logger, &event)
			Ok(t, err)
			Equals(t, c.draftExp, draftEvType)
			// Test draft parsing when draft PRs are enabled.
			draftParser := parser
			draftParser.AllowDraftPRs = true
			_, draftEvType, _, _, _, err = draftParser.ParseGithubPullEvent(logger, &event)
			Ok(t, err)
			Equals(t, c.exp, draftEvType)
		})
	}
}

func TestParseGithubPull(t *testing.T) {
	logger := logging.NewNoopLogger(t)
	testPull := deepcopy.Copy(Pull).(github.PullRequest)
	testPull.Head.SHA = nil
	_, _, _, err := parser.ParseGithubPull(logger, &testPull)
	ErrEquals(t, "head.sha is null", err)

	testPull = deepcopy.Copy(Pull).(github.PullRequest)
	testPull.HTMLURL = nil
	_, _, _, err = parser.ParseGithubPull(logger, &testPull)
	ErrEquals(t, "html_url is null", err)

	testPull = deepcopy.Copy(Pull).(github.PullRequest)
	testPull.Head.Ref = nil
	_, _, _, err = parser.ParseGithubPull(logger, &testPull)
	ErrEquals(t, "head.ref is null", err)

	testPull = deepcopy.Copy(Pull).(github.PullRequest)
	testPull.Base.Ref = nil
	_, _, _, err = parser.ParseGithubPull(logger, &testPull)
	ErrEquals(t, "base.ref is null", err)

	testPull = deepcopy.Copy(Pull).(github.PullRequest)
	testPull.User.Login = nil
	_, _, _, err = parser.ParseGithubPull(logger, &testPull)
	ErrEquals(t, "user.login is null", err)

	testPull = deepcopy.Copy(Pull).(github.PullRequest)
	testPull.Number = nil
	_, _, _, err = parser.ParseGithubPull(logger, &testPull)
	ErrEquals(t, "number is null", err)

	pullRes, actBaseRepo, actHeadRepo, err := parser.ParseGithubPull(logger, &Pull)
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
	}, pullRes)
	Equals(t, expBaseRepo, actBaseRepo)
	Equals(t, expBaseRepo, actHeadRepo)
}

func TestParseGitlabMergeEvent(t *testing.T) {
	t.Log("should properly parse a gitlab merge event")
	path := filepath.Join("testdata", "gitlab-merge-request-event.json")
	bytes, err := os.ReadFile(path)
	Ok(t, err)
	var event *gitlab.MergeEvent
	err = json.Unmarshal(bytes, &event)
	Ok(t, err)
	pull, evType, actBaseRepo, actHeadRepo, actUser, err := parser.ParseGitlabMergeRequestEvent(*event)
	Ok(t, err)

	expBaseRepo := models.Repo{
		FullName:          "lkysow/atlantis-example",
		Name:              "atlantis-example",
		SanitizedCloneURL: "https://gitlab-user:<redacted>@gitlab.com/lkysow/atlantis-example.git",
		Owner:             "lkysow",
		CloneURL:          "https://gitlab-user:gitlab-token@gitlab.com/lkysow/atlantis-example.git",
		VCSHost: models.VCSHost{
			Hostname: "gitlab.com",
			Type:     models.Gitlab,
		},
	}

	Equals(t, models.PullRequest{
		URL:        "https://gitlab.com/lkysow/atlantis-example/merge_requests/12",
		Author:     "lkysow",
		Num:        12,
		HeadCommit: "d2eae324ca26242abca45d7b49d582cddb2a4f15",
		HeadBranch: "patch-1",
		BaseBranch: "main",
		State:      models.OpenPullState,
		BaseRepo:   expBaseRepo,
	}, pull)
	Equals(t, models.OpenedPullEvent, evType)

	Equals(t, expBaseRepo, actBaseRepo)
	Equals(t, models.Repo{
		FullName:          "sourceorg/atlantis-example",
		Name:              "atlantis-example",
		SanitizedCloneURL: "https://gitlab-user:<redacted>@gitlab.com/sourceorg/atlantis-example.git",
		Owner:             "sourceorg",
		CloneURL:          "https://gitlab-user:gitlab-token@gitlab.com/sourceorg/atlantis-example.git",
		VCSHost: models.VCSHost{
			Hostname: "gitlab.com",
			Type:     models.Gitlab,
		},
	}, actHeadRepo)
	Equals(t, models.User{Username: "lkysow"}, actUser)

	t.Log("If the state is closed, should set field correctly.")
	event.ObjectAttributes.State = "closed"
	pull, _, _, _, _, err = parser.ParseGitlabMergeRequestEvent(*event)
	Ok(t, err)
	Equals(t, models.ClosedPullState, pull.State)
}

func TestParseGitlabMergeEventFromDraft(t *testing.T) {
	path := filepath.Join("testdata", "gitlab-merge-request-event.json")
	bytes, err := os.ReadFile(path)
	Ok(t, err)

	var event gitlab.MergeEvent
	err = json.Unmarshal(bytes, &event)
	Ok(t, err)

	testEvent := deepcopy.Copy(event).(gitlab.MergeEvent)
	testEvent.ObjectAttributes.WorkInProgress = true

	_, evType, _, _, _, err := parser.ParseGitlabMergeRequestEvent(testEvent)
	Ok(t, err)
	Equals(t, models.OtherPullEvent, evType)

	parser.AllowDraftPRs = true
	defer func() { parser.AllowDraftPRs = false }()
	_, evType, _, _, _, err = parser.ParseGitlabMergeRequestEvent(testEvent)
	Ok(t, err)
	Equals(t, models.OpenedPullEvent, evType)
}

// Should be able to parse a merge event from a repo that is in a subgroup,
// i.e. instead of under an owner/repo it's under an owner/group/subgroup/repo.
func TestParseGitlabMergeEvent_Subgroup(t *testing.T) {
	path := filepath.Join("testdata", "gitlab-merge-request-event-subgroup.json")
	bytes, err := os.ReadFile(path)
	Ok(t, err)
	var event *gitlab.MergeEvent
	err = json.Unmarshal(bytes, &event)
	Ok(t, err)
	pull, evType, actBaseRepo, actHeadRepo, actUser, err := parser.ParseGitlabMergeRequestEvent(*event)
	Ok(t, err)

	expBaseRepo := models.Repo{
		FullName:          "lkysow-test/subgroup/sub-subgroup/atlantis-example",
		Name:              "atlantis-example",
		SanitizedCloneURL: "https://gitlab-user:<redacted>@gitlab.com/lkysow-test/subgroup/sub-subgroup/atlantis-example.git",
		Owner:             "lkysow-test/subgroup/sub-subgroup",
		CloneURL:          "https://gitlab-user:gitlab-token@gitlab.com/lkysow-test/subgroup/sub-subgroup/atlantis-example.git",
		VCSHost: models.VCSHost{
			Hostname: "gitlab.com",
			Type:     models.Gitlab,
		},
	}

	Equals(t, models.PullRequest{
		URL:        "https://gitlab.com/lkysow-test/subgroup/sub-subgroup/atlantis-example/merge_requests/2",
		Author:     "lkysow",
		Num:        2,
		HeadCommit: "901d9770ef1a6862e2a73ec1bacc73590abb9aff",
		HeadBranch: "patch",
		BaseBranch: "main",
		State:      models.OpenPullState,
		BaseRepo:   expBaseRepo,
	}, pull)
	Equals(t, models.OpenedPullEvent, evType)

	Equals(t, expBaseRepo, actBaseRepo)
	Equals(t, models.Repo{
		FullName:          "lkysow-test/subgroup/sub-subgroup/atlantis-example",
		Name:              "atlantis-example",
		SanitizedCloneURL: "https://gitlab-user:<redacted>@gitlab.com/lkysow-test/subgroup/sub-subgroup/atlantis-example.git",
		Owner:             "lkysow-test/subgroup/sub-subgroup",
		CloneURL:          "https://gitlab-user:gitlab-token@gitlab.com/lkysow-test/subgroup/sub-subgroup/atlantis-example.git",
		VCSHost: models.VCSHost{
			Hostname: "gitlab.com",
			Type:     models.Gitlab,
		},
	}, actHeadRepo)
	Equals(t, models.User{Username: "lkysow"}, actUser)
}

func TestParseGitlabMergeEvent_Update_ActionType(t *testing.T) {
	cases := []struct {
		filename string
		exp      models.PullRequestEventType
	}{
		{
			filename: "gitlab-merge-request-event-update-title.json",
			exp:      models.OtherPullEvent,
		},
		{
			filename: "gitlab-merge-request-event-update-new-commit.json",
			exp:      models.UpdatedPullEvent,
		},
		{
			filename: "gitlab-merge-request-event-update-labels.json",
			exp:      models.OtherPullEvent,
		},
		{
			filename: "gitlab-merge-request-event-update-description.json",
			exp:      models.OtherPullEvent,
		},
		{
			filename: "gitlab-merge-request-event-update-assignee.json",
			exp:      models.OtherPullEvent,
		},
		{
			filename: "gitlab-merge-request-event-update-mixed.json",
			exp:      models.OtherPullEvent,
		},
		{
			filename: "gitlab-merge-request-event-update-target-branch.json",
			exp:      models.OtherPullEvent,
		},
		{
			filename: "gitlab-merge-request-event-update-reviewer.json",
			exp:      models.OtherPullEvent,
		},
		{
			filename: "gitlab-merge-request-event-update-milestone.json",
			exp:      models.OtherPullEvent,
		},
		{
			filename: "gitlab-merge-request-event-mark-as-ready.json",
			exp:      models.UpdatedPullEvent,
		},
	}

	for _, c := range cases {
		t.Run(c.filename, func(t *testing.T) {
			path := filepath.Join("testdata", c.filename)
			bytes, err := os.ReadFile(path)
			Ok(t, err)

			var event *gitlab.MergeEvent
			err = json.Unmarshal(bytes, &event)
			Ok(t, err)
			_, evType, _, _, _, err := parser.ParseGitlabMergeRequestEvent(*event)
			Ok(t, err)
			Equals(t, c.exp, evType)
		})
	}
}

func TestParseGitlabMergeEvent_ActionType(t *testing.T) {
	cases := []struct {
		action string
		exp    models.PullRequestEventType
	}{
		{
			action: "open",
			exp:    models.OpenedPullEvent,
		},
		{
			action: "merge",
			exp:    models.ClosedPullEvent,
		},
		{
			action: "close",
			exp:    models.ClosedPullEvent,
		},
		{
			action: "other",
			exp:    models.OtherPullEvent,
		},
	}

	path := filepath.Join("testdata", "gitlab-merge-request-event.json")
	bytes, err := os.ReadFile(path)
	Ok(t, err)
	mergeEventJSON := string(bytes)

	for _, c := range cases {
		t.Run(c.action, func(t *testing.T) {
			var event *gitlab.MergeEvent
			err = json.Unmarshal(bytes, &event)
			Ok(t, err)
			eventJSON := strings.Replace(mergeEventJSON, `"action": "open"`, fmt.Sprintf(`"action": %q`, c.action), 1)
			err := json.Unmarshal([]byte(eventJSON), &event)
			Ok(t, err)
			_, evType, _, _, _, err := parser.ParseGitlabMergeRequestEvent(*event)
			Ok(t, err)
			Equals(t, c.exp, evType)
		})
	}
}

func TestParseGitlabMergeRequest(t *testing.T) {
	t.Log("should properly parse a gitlab merge request")
	path := filepath.Join("testdata", "gitlab-get-merge-request.json")
	bytes, err := os.ReadFile(path)
	if err != nil {
		Ok(t, err)
	}
	var event *gitlab.MergeRequest
	err = json.Unmarshal(bytes, &event)
	Ok(t, err)
	repo := models.Repo{
		FullName:          "gitlabhq/gitlab-test",
		Name:              "gitlab-test",
		SanitizedCloneURL: "https://gitlab-user:<redacted>@example.com/gitlabhq/gitlab-test.git",
		Owner:             "gitlabhq",
		CloneURL:          "https://gitlab-user:gitlab-token@example.com/gitlabhq/gitlab-test.git",
		VCSHost: models.VCSHost{
			Hostname: "example.com",
			Type:     models.Gitlab,
		},
	}
	pull := parser.ParseGitlabMergeRequest(event, repo)
	Equals(t, models.PullRequest{
		URL:        "https://gitlab.com/lkysow/atlantis-example/merge_requests/8",
		Author:     "lkysow",
		Num:        8,
		HeadCommit: "0b4ac85ea3063ad5f2974d10cd68dd1f937aaac2",
		HeadBranch: "abc",
		BaseBranch: "main",
		State:      models.OpenPullState,
		BaseRepo:   repo,
	}, pull)

	t.Log("If the state is closed, should set field correctly.")
	event.State = "closed"
	pull = parser.ParseGitlabMergeRequest(event, repo)
	Equals(t, models.ClosedPullState, pull.State)
}

func TestParseGitlabMergeRequest_Subgroup(t *testing.T) {
	path := filepath.Join("testdata", "gitlab-get-merge-request-subgroup.json")
	bytes, err := os.ReadFile(path)
	if err != nil {
		Ok(t, err)
	}
	var event *gitlab.MergeRequest
	err = json.Unmarshal(bytes, &event)
	Ok(t, err)

	repo := models.Repo{
		FullName:          "lkysow-test/subgroup/sub-subgroup/atlantis-example",
		Name:              "atlantis-example",
		SanitizedCloneURL: "https://gitlab-user:<redacted>@gitlab.com/lkysow-test/subgroup/sub-subgroup/atlantis-example.git",
		Owner:             "lkysow-test/subgroup/sub-subgroup",
		CloneURL:          "https://gitlab-user:gitlab-token@gitlab.com/lkysow-test/subgroup/sub-subgroup/atlantis-example.git",
		VCSHost: models.VCSHost{
			Hostname: "gitlab.com",
			Type:     models.Gitlab,
		},
	}
	pull := parser.ParseGitlabMergeRequest(event, repo)
	Equals(t, models.PullRequest{
		URL:        "https://gitlab.com/lkysow-test/subgroup/sub-subgroup/atlantis-example/merge_requests/2",
		Author:     "lkysow",
		Num:        2,
		HeadCommit: "901d9770ef1a6862e2a73ec1bacc73590abb9aff",
		HeadBranch: "patch",
		BaseBranch: "main",
		State:      models.OpenPullState,
		BaseRepo:   repo,
	}, pull)
}

func TestParseGitlabMergeCommentEvent(t *testing.T) {
	t.Log("should properly parse a gitlab merge comment event")
	path := filepath.Join("testdata", "gitlab-merge-request-comment-event.json")
	bytes, err := os.ReadFile(path)
	Ok(t, err)
	var event *gitlab.MergeCommentEvent
	err = json.Unmarshal(bytes, &event)
	Ok(t, err)
	baseRepo, headRepo, commentID, user, err := parser.ParseGitlabMergeRequestCommentEvent(*event)
	Ok(t, err)
	Equals(t, models.Repo{
		FullName:          "gitlabhq/gitlab-test",
		Name:              "gitlab-test",
		SanitizedCloneURL: "https://gitlab-user:<redacted>@example.com/gitlabhq/gitlab-test.git",
		Owner:             "gitlabhq",
		CloneURL:          "https://gitlab-user:gitlab-token@example.com/gitlabhq/gitlab-test.git",
		VCSHost: models.VCSHost{
			Hostname: "example.com",
			Type:     models.Gitlab,
		},
	}, baseRepo)
	Equals(t, models.Repo{
		FullName:          "gitlab-org/gitlab-test",
		Name:              "gitlab-test",
		SanitizedCloneURL: "https://gitlab-user:<redacted>@example.com/gitlab-org/gitlab-test.git",
		Owner:             "gitlab-org",
		CloneURL:          "https://gitlab-user:gitlab-token@example.com/gitlab-org/gitlab-test.git",
		VCSHost: models.VCSHost{
			Hostname: "example.com",
			Type:     models.Gitlab,
		},
	}, headRepo)
	Equals(t, 1244, commentID)
	Equals(t, models.User{
		Username: "root",
	}, user)
}

// Should properly parse a gitlab merge comment event from a subgroup repo.
func TestParseGitlabMergeCommentEvent_Subgroup(t *testing.T) {
	path := filepath.Join("testdata", "gitlab-merge-request-comment-event-subgroup.json")
	bytes, err := os.ReadFile(path)
	Ok(t, err)
	var event *gitlab.MergeCommentEvent
	err = json.Unmarshal(bytes, &event)
	Ok(t, err)
	baseRepo, headRepo, commentID, user, err := parser.ParseGitlabMergeRequestCommentEvent(*event)
	Ok(t, err)

	Equals(t, models.Repo{
		FullName:          "lkysow-test/subgroup/sub-subgroup/atlantis-example",
		Name:              "atlantis-example",
		SanitizedCloneURL: "https://gitlab-user:<redacted>@gitlab.com/lkysow-test/subgroup/sub-subgroup/atlantis-example.git",
		Owner:             "lkysow-test/subgroup/sub-subgroup",
		CloneURL:          "https://gitlab-user:gitlab-token@gitlab.com/lkysow-test/subgroup/sub-subgroup/atlantis-example.git",
		VCSHost: models.VCSHost{
			Hostname: "gitlab.com",
			Type:     models.Gitlab,
		},
	}, baseRepo)
	Equals(t, models.Repo{
		FullName:          "lkysow-test/subgroup/sub-subgroup/atlantis-example",
		Name:              "atlantis-example",
		SanitizedCloneURL: "https://gitlab-user:<redacted>@gitlab.com/lkysow-test/subgroup/sub-subgroup/atlantis-example.git",
		Owner:             "lkysow-test/subgroup/sub-subgroup",
		CloneURL:          "https://gitlab-user:gitlab-token@gitlab.com/lkysow-test/subgroup/sub-subgroup/atlantis-example.git",
		VCSHost: models.VCSHost{
			Hostname: "gitlab.com",
			Type:     models.Gitlab,
		},
	}, headRepo)
	Equals(t, 96056916, commentID)
	Equals(t, models.User{
		Username: "lkysow",
	}, user)
}

func TestNewCommand_CleansDir(t *testing.T) {
	cases := []struct {
		RepoRelDir string
		ExpDir     string
	}{
		{
			"",
			"",
		},
		{
			"/",
			".",
		},
		{
			"./",
			".",
		},
		// We rely on our callers to not pass in relative dirs.
		{
			"..",
			"..",
		},
	}

	for _, c := range cases {
		t.Run(c.RepoRelDir, func(t *testing.T) {
			cmd := events.NewCommentCommand(c.RepoRelDir, nil, command.Plan, "", false, false, "", "workspace", "", "", false)
			Equals(t, c.ExpDir, cmd.RepoRelDir)
		})
	}
}

func TestNewCommand_EmptyDirWorkspaceProject(t *testing.T) {
	cmd := events.NewCommentCommand("", nil, command.Plan, "", false, false, "", "", "", "", false)
	Equals(t, events.CommentCommand{
		RepoRelDir:  "",
		Flags:       nil,
		Name:        command.Plan,
		Verbose:     false,
		Workspace:   "",
		ProjectName: "",
	}, *cmd)
}

func TestNewCommand_AllFieldsSet(t *testing.T) {
	cmd := events.NewCommentCommand("dir", []string{"a", "b"}, command.Plan, "", true, false, "", "workspace", "project", "policyset", false)
	Equals(t, events.CommentCommand{
		Workspace:   "workspace",
		RepoRelDir:  "dir",
		Verbose:     true,
		Flags:       []string{"a", "b"},
		Name:        command.Plan,
		ProjectName: "project",
		PolicySet:   "policyset",
	}, *cmd)
}

func TestAutoplanCommand_CommandName(t *testing.T) {
	Equals(t, command.Plan, (events.AutoplanCommand{}).CommandName())
}

func TestAutoplanCommand_IsVerbose(t *testing.T) {
	Equals(t, false, (events.AutoplanCommand{}).IsVerbose())
}

func TestAutoplanCommand_IsAutoplan(t *testing.T) {
	Equals(t, true, (events.AutoplanCommand{}).IsAutoplan())
}

func TestCommentCommand_CommandName(t *testing.T) {
	Equals(t, command.Plan, (events.CommentCommand{
		Name: command.Plan,
	}).CommandName())
	Equals(t, command.Apply, (events.CommentCommand{
		Name: command.Apply,
	}).CommandName())
}

func TestCommentCommand_IsVerbose(t *testing.T) {
	Equals(t, false, (events.CommentCommand{
		Verbose: false,
	}).IsVerbose())
	Equals(t, true, (events.CommentCommand{
		Verbose: true,
	}).IsVerbose())
}

func TestCommentCommand_IsAutoplan(t *testing.T) {
	Equals(t, false, (events.CommentCommand{}).IsAutoplan())
}

func TestCommentCommand_String(t *testing.T) {
	exp := `command="plan", verbose=true, dir="mydir", workspace="myworkspace", project="myproject", policyset="", auto-merge-disabled=false, auto-merge-method=, clear-policy-approval=false, flags="flag1,flag2"`
	Equals(t, exp, (events.CommentCommand{
		RepoRelDir:  "mydir",
		Flags:       []string{"flag1", "flag2"},
		Name:        command.Plan,
		Verbose:     true,
		Workspace:   "myworkspace",
		ProjectName: "myproject",
	}).String())
}

func TestParseBitbucketCloudCommentEvent_EmptyString(t *testing.T) {
	_, _, _, _, _, err := parser.ParseBitbucketCloudPullCommentEvent([]byte(""))
	ErrEquals(t, "parsing json: unexpected end of JSON input", err)
}

func TestParseBitbucketCloudCommentEvent_EmptyObject(t *testing.T) {
	_, _, _, _, _, err := parser.ParseBitbucketCloudPullCommentEvent([]byte("{}"))
	ErrContains(t, "Key: 'CommentEvent.CommonEventData.Actor' Error:Field validation for 'Actor' failed on the 'required' tag\nKey: 'CommentEvent.CommonEventData.Repository' Error:Field validation for 'Repository' failed on the 'required' tag\nKey: 'CommentEvent.CommonEventData.PullRequest' Error:Field validation for 'PullRequest' failed on the 'required' tag\nKey: 'CommentEvent.Comment' Error:Field validation for 'Comment' failed on the 'required' tag", err)
}

func TestParseBitbucketCloudCommentEvent_CommitHashMissing(t *testing.T) {
	path := filepath.Join("testdata", "bitbucket-cloud-comment-event.json")
	bytes, err := os.ReadFile(path)
	Ok(t, err)
	emptyCommitHash := strings.Replace(string(bytes), `        "hash": "e0624da46d3a",`, "", -1)
	_, _, _, _, _, err = parser.ParseBitbucketCloudPullCommentEvent([]byte(emptyCommitHash))
	ErrContains(t, "Key: 'CommentEvent.CommonEventData.PullRequest.Source.Commit.Hash' Error:Field validation for 'Hash' failed on the 'required' tag", err)
}

func TestParseBitbucketCloudCommentEvent_ValidEvent(t *testing.T) {
	path := filepath.Join("testdata", "bitbucket-cloud-comment-event.json")
	bytes, err := os.ReadFile(path)
	Ok(t, err)
	pull, baseRepo, headRepo, user, comment, err := parser.ParseBitbucketCloudPullCommentEvent(bytes)
	Ok(t, err)
	expBaseRepo := models.Repo{
		FullName:          "lkysow/atlantis-example",
		Owner:             "lkysow",
		Name:              "atlantis-example",
		CloneURL:          "https://bitbucket-user:bitbucket-token@bitbucket.org/lkysow/atlantis-example.git",
		SanitizedCloneURL: "https://bitbucket-user:<redacted>@bitbucket.org/lkysow/atlantis-example.git",
		VCSHost: models.VCSHost{
			Hostname: "bitbucket.org",
			Type:     models.BitbucketCloud,
		},
	}
	Equals(t, expBaseRepo, baseRepo)
	Equals(t, models.PullRequest{
		Num:        2,
		HeadCommit: "e0624da46d3a",
		URL:        "https://bitbucket.org/lkysow/atlantis-example/pull-requests/2",
		HeadBranch: "lkysow/maintf-edited-online-with-bitbucket-1532029690581",
		BaseBranch: "main",
		Author:     "557058:dc3817de-68b5-45cd-b81c-5c39d2560090",
		State:      models.ClosedPullState,
		BaseRepo:   expBaseRepo,
	}, pull)
	Equals(t, models.Repo{
		FullName:          "lkysow-fork/atlantis-example",
		Owner:             "lkysow-fork",
		Name:              "atlantis-example",
		CloneURL:          "https://bitbucket-user:bitbucket-token@bitbucket.org/lkysow-fork/atlantis-example.git",
		SanitizedCloneURL: "https://bitbucket-user:<redacted>@bitbucket.org/lkysow-fork/atlantis-example.git",
		VCSHost: models.VCSHost{
			Hostname: "bitbucket.org",
			Type:     models.BitbucketCloud,
		},
	}, headRepo)
	Equals(t, models.User{
		Username: "557058:dc3817de-68b5-45cd-b81c-5c39d2560090",
	}, user)
	Equals(t, "my comment", comment)
}

func TestParseBitbucketCloudCommentEvent_MultipleStates(t *testing.T) {
	path := filepath.Join("testdata", "bitbucket-cloud-comment-event.json")
	bytes, err := os.ReadFile(path)
	if err != nil {
		Ok(t, err)
	}

	cases := []struct {
		pullState string
		exp       models.PullRequestState
	}{
		{
			"OPEN",
			models.OpenPullState,
		},
		{
			"MERGED",
			models.ClosedPullState,
		},
		{
			"SUPERSEDED",
			models.ClosedPullState,
		},
		{
			"DECLINED",
			models.ClosedPullState,
		},
	}

	for _, c := range cases {
		t.Run(c.pullState, func(t *testing.T) {
			withState := strings.Replace(string(bytes), `"state": "MERGED"`, fmt.Sprintf(`"state": "%s"`, c.pullState), -1)
			pull, _, _, _, _, err := parser.ParseBitbucketCloudPullCommentEvent([]byte(withState))
			Ok(t, err)
			Equals(t, c.exp, pull.State)
		})
	}
}

func TestParseBitbucketCloudPullEvent_ValidEvent(t *testing.T) {
	path := filepath.Join("testdata", "bitbucket-cloud-pull-event-created.json")
	bytes, err := os.ReadFile(path)
	if err != nil {
		Ok(t, err)
	}
	pull, baseRepo, headRepo, user, err := parser.ParseBitbucketCloudPullEvent(bytes)
	Ok(t, err)
	expBaseRepo := models.Repo{
		FullName:          "lkysow/atlantis-example",
		Owner:             "lkysow",
		Name:              "atlantis-example",
		CloneURL:          "https://bitbucket-user:bitbucket-token@bitbucket.org/lkysow/atlantis-example.git",
		SanitizedCloneURL: "https://bitbucket-user:<redacted>@bitbucket.org/lkysow/atlantis-example.git",
		VCSHost: models.VCSHost{
			Hostname: "bitbucket.org",
			Type:     models.BitbucketCloud,
		},
	}
	Equals(t, expBaseRepo, baseRepo)
	Equals(t, models.PullRequest{
		Num:        16,
		HeadCommit: "1e69a602caef",
		URL:        "https://bitbucket.org/lkysow/atlantis-example/pull-requests/16",
		HeadBranch: "Luke/maintf-edited-online-with-bitbucket-1560433073473",
		BaseBranch: "main",
		Author:     "557058:dc3817de-68b5-45cd-b81c-5c39d2560090",
		State:      models.OpenPullState,
		BaseRepo:   expBaseRepo,
	}, pull)
	Equals(t, models.Repo{
		FullName:          "lkysow-fork/atlantis-example",
		Owner:             "lkysow-fork",
		Name:              "atlantis-example",
		CloneURL:          "https://bitbucket-user:bitbucket-token@bitbucket.org/lkysow-fork/atlantis-example.git",
		SanitizedCloneURL: "https://bitbucket-user:<redacted>@bitbucket.org/lkysow-fork/atlantis-example.git",
		VCSHost: models.VCSHost{
			Hostname: "bitbucket.org",
			Type:     models.BitbucketCloud,
		},
	}, headRepo)
	Equals(t, models.User{
		Username: "557058:dc3817de-68b5-45cd-b81c-5c39d2560090",
	}, user)
}

func TestParseBitbucketCloudPullEvent_States(t *testing.T) {
	for _, c := range []struct {
		JSON     string
		ExpState models.PullRequestState
	}{
		{
			JSON:     "bitbucket-cloud-pull-event-created.json",
			ExpState: models.OpenPullState,
		},
		{
			JSON:     "bitbucket-cloud-pull-event-fulfilled.json",
			ExpState: models.ClosedPullState,
		},
		{
			JSON:     "bitbucket-cloud-pull-event-rejected.json",
			ExpState: models.ClosedPullState,
		},
	} {
		path := filepath.Join("testdata", c.JSON)
		bytes, err := os.ReadFile(path)
		if err != nil {
			Ok(t, err)
		}
		pull, _, _, _, err := parser.ParseBitbucketCloudPullEvent(bytes)
		Ok(t, err)
		Equals(t, c.ExpState, pull.State)
	}
}

func TestBitBucketNonCodeChangesAreIgnored(t *testing.T) {
	// lets say a user opens a PR
	act := parser.GetBitbucketCloudPullEventType("pullrequest:created", "fakeSha", "https://github.com/fakeorg/fakerepo/pull/1")
	Equals(t, models.OpenedPullEvent, act)
	// Another update with same SHA should be ignored
	act = parser.GetBitbucketCloudPullEventType("pullrequest:updated", "fakeSha", "https://github.com/fakeorg/fakerepo/pull/1")
	Equals(t, models.OtherPullEvent, act)
	// Only if SHA changes do we act
	act = parser.GetBitbucketCloudPullEventType("pullrequest:updated", "fakeSha2", "https://github.com/fakeorg/fakerepo/pull/1")
	Equals(t, models.UpdatedPullEvent, act)

	// If sha changes in separate PR,
	act = parser.GetBitbucketCloudPullEventType("pullrequest:updated", "otherPRSha", "https://github.com/fakeorg/fakerepo/pull/2")
	Equals(t, models.UpdatedPullEvent, act)
	// We will still ignore same shas in first PR
	act = parser.GetBitbucketCloudPullEventType("pullrequest:updated", "fakeSha2", "https://github.com/fakeorg/fakerepo/pull/1")
	Equals(t, models.OtherPullEvent, act)
}

func TestBitbucketShaCacheExpires(t *testing.T) {
	// lets say a user opens a PR
	act := parser.GetBitbucketCloudPullEventType("pullrequest:created", "fakeSha", "https://github.com/fakeorg/fakerepo/pull/1")
	Equals(t, models.OpenedPullEvent, act)
	// Another update with same SHA should be ignored
	act = parser.GetBitbucketCloudPullEventType("pullrequest:updated", "fakeSha", "https://github.com/fakeorg/fakerepo/pull/1")
	Equals(t, models.OtherPullEvent, act)
	// But after 300 times, the cache should expire
	// this is so we don't have ever increasing memory usage
	for i := 0; i < 302; i++ {
		parser.GetBitbucketCloudPullEventType("pullrequest:updated", "fakeSha", fmt.Sprintf("https://github.com/fakeorg/fakerepo/pull/%d", i))
	}
	// and now SHA will seen as a change again
	act = parser.GetBitbucketCloudPullEventType("pullrequest:updated", "fakeSha", "https://github.com/fakeorg/fakerepo/pull/1")
	Equals(t, models.UpdatedPullEvent, act)
}

func TestGetBitbucketCloudEventType(t *testing.T) {
	cases := []struct {
		header string
		exp    models.PullRequestEventType
	}{
		{
			header: "pullrequest:created",
			exp:    models.OpenedPullEvent,
		},
		{
			header: "pullrequest:updated",
			exp:    models.UpdatedPullEvent,
		},
		{
			header: "pullrequest:fulfilled",
			exp:    models.ClosedPullEvent,
		},
		{
			header: "pullrequest:rejected",
			exp:    models.ClosedPullEvent,
		},
		{
			header: "random",
			exp:    models.OtherPullEvent,
		},
	}
	for _, c := range cases {
		t.Run(c.header, func(t *testing.T) {
			// we pass in the header as the SHA so the SHA changes each time
			// the code will ignore duplicate SHAS to avoid extra TF plans
			act := parser.GetBitbucketCloudPullEventType(c.header, c.header, "https://github.com/fakeorg/fakerepo/pull/1")
			Equals(t, c.exp, act)
		})
	}
}

func TestParseBitbucketServerCommentEvent_EmptyString(t *testing.T) {
	_, _, _, _, _, err := parser.ParseBitbucketServerPullCommentEvent([]byte(""))
	ErrEquals(t, "parsing json: unexpected end of JSON input", err)
}

func TestParseBitbucketServerCommentEvent_EmptyObject(t *testing.T) {
	_, _, _, _, _, err := parser.ParseBitbucketServerPullCommentEvent([]byte("{}"))
	ErrContains(t, `API response "{}" was missing fields: Key: 'CommentEvent.CommonEventData.Actor' Error:Field validation for 'Actor' failed on the 'required' tag`, err)
}

func TestParseBitbucketServerCommentEvent_CommitHashMissing(t *testing.T) {
	path := filepath.Join("testdata", "bitbucket-server-comment-event.json")
	bytes, err := os.ReadFile(path)
	if err != nil {
		Ok(t, err)
	}
	emptyCommitHash := strings.Replace(string(bytes), `"latestCommit": "bfb1af1ba9c2a2fa84cd61af67e6e1b60a22e060",`, "", -1)
	_, _, _, _, _, err = parser.ParseBitbucketServerPullCommentEvent([]byte(emptyCommitHash))
	ErrContains(t, "Key: 'CommentEvent.CommonEventData.PullRequest.FromRef.LatestCommit' Error:Field validation for 'LatestCommit' failed on the 'required' tag", err)
}

func TestParseBitbucketServerCommentEvent_ValidEvent(t *testing.T) {
	path := filepath.Join("testdata", "bitbucket-server-comment-event.json")
	bytes, err := os.ReadFile(path)
	if err != nil {
		Ok(t, err)
	}
	pull, baseRepo, headRepo, user, comment, err := parser.ParseBitbucketServerPullCommentEvent(bytes)
	Ok(t, err)
	expBaseRepo := models.Repo{
		FullName:          "atlantis/atlantis-example",
		Owner:             "atlantis",
		Name:              "atlantis-example",
		CloneURL:          "http://bitbucket-user:bitbucket-token@mycorp.com:7490/scm/at/atlantis-example.git",
		SanitizedCloneURL: "http://bitbucket-user:<redacted>@mycorp.com:7490/scm/at/atlantis-example.git",
		VCSHost: models.VCSHost{
			Hostname: "mycorp.com",
			Type:     models.BitbucketServer,
		},
	}
	Equals(t, expBaseRepo, baseRepo)
	Equals(t, models.PullRequest{
		Num:        1,
		HeadCommit: "bfb1af1ba9c2a2fa84cd61af67e6e1b60a22e060",
		URL:        "http://mycorp.com:7490/projects/AT/repos/atlantis-example/pull-requests/1",
		HeadBranch: "branch",
		BaseBranch: "main",
		Author:     "lkysow",
		State:      models.OpenPullState,
		BaseRepo:   expBaseRepo,
	}, pull)
	Equals(t, models.Repo{
		FullName:          "atlantis-fork/atlantis-example",
		Owner:             "atlantis-fork",
		Name:              "atlantis-example",
		CloneURL:          "http://bitbucket-user:bitbucket-token@mycorp.com:7490/scm/fk/atlantis-example.git",
		SanitizedCloneURL: "http://bitbucket-user:<redacted>@mycorp.com:7490/scm/fk/atlantis-example.git",
		VCSHost: models.VCSHost{
			Hostname: "mycorp.com",
			Type:     models.BitbucketServer,
		},
	}, headRepo)
	Equals(t, models.User{
		Username: "lkysow",
	}, user)
	Equals(t, "atlantis plan", comment)
}

func TestParseBitbucketServerCommentEvent_MultipleStates(t *testing.T) {
	path := filepath.Join("testdata", "bitbucket-server-comment-event.json")
	bytes, err := os.ReadFile(path)
	if err != nil {
		Ok(t, err)
	}

	cases := []struct {
		pullState string
		exp       models.PullRequestState
	}{
		{
			"OPEN",
			models.OpenPullState,
		},
		{
			"MERGED",
			models.ClosedPullState,
		},
		{
			"DECLINED",
			models.ClosedPullState,
		},
	}

	for _, c := range cases {
		t.Run(c.pullState, func(t *testing.T) {
			withState := strings.Replace(string(bytes), `"state": "OPEN"`, fmt.Sprintf(`"state": "%s"`, c.pullState), -1)
			pull, _, _, _, _, err := parser.ParseBitbucketServerPullCommentEvent([]byte(withState))
			Ok(t, err)
			Equals(t, c.exp, pull.State)
		})
	}
}

func TestParseBitbucketServerPullEvent_ValidEvent(t *testing.T) {
	path := filepath.Join("testdata", "bitbucket-server-pull-event-merged.json")
	bytes, err := os.ReadFile(path)
	if err != nil {
		Ok(t, err)
	}
	pull, baseRepo, headRepo, user, err := parser.ParseBitbucketServerPullEvent(bytes)
	Ok(t, err)
	expBaseRepo := models.Repo{
		FullName:          "atlantis/atlantis-example",
		Owner:             "atlantis",
		Name:              "atlantis-example",
		CloneURL:          "http://bitbucket-user:bitbucket-token@mycorp.com:7490/scm/at/atlantis-example.git",
		SanitizedCloneURL: "http://bitbucket-user:<redacted>@mycorp.com:7490/scm/at/atlantis-example.git",
		VCSHost: models.VCSHost{
			Hostname: "mycorp.com",
			Type:     models.BitbucketServer,
		},
	}
	Equals(t, expBaseRepo, baseRepo)
	Equals(t, models.PullRequest{
		Num:        2,
		HeadCommit: "86a574157f5a2dadaf595b9f06c70fdfdd039912",
		URL:        "http://mycorp.com:7490/projects/AT/repos/atlantis-example/pull-requests/2",
		HeadBranch: "branch",
		BaseBranch: "main",
		Author:     "lkysow",
		State:      models.ClosedPullState,
		BaseRepo:   expBaseRepo,
	}, pull)
	Equals(t, models.Repo{
		FullName:          "atlantis-fork/atlantis-example",
		Owner:             "atlantis-fork",
		Name:              "atlantis-example",
		CloneURL:          "http://bitbucket-user:bitbucket-token@mycorp.com:7490/scm/fk/atlantis-example.git",
		SanitizedCloneURL: "http://bitbucket-user:<redacted>@mycorp.com:7490/scm/fk/atlantis-example.git",
		VCSHost: models.VCSHost{
			Hostname: "mycorp.com",
			Type:     models.BitbucketServer,
		},
	}, headRepo)
	Equals(t, models.User{
		Username: "lkysow",
	}, user)
}

func TestGetBitbucketServerEventType(t *testing.T) {
	cases := []struct {
		header string
		exp    models.PullRequestEventType
	}{
		{
			header: "pr:opened",
			exp:    models.OpenedPullEvent,
		},
		{
			header: "pr:merged",
			exp:    models.ClosedPullEvent,
		},
		{
			header: "pr:declined",
			exp:    models.ClosedPullEvent,
		},
		{
			header: "pr:deleted",
			exp:    models.ClosedPullEvent,
		},
		{
			header: "random",
			exp:    models.OtherPullEvent,
		},
	}
	for _, c := range cases {
		t.Run(c.header, func(t *testing.T) {
			act := parser.GetBitbucketServerPullEventType(c.header)
			Equals(t, c.exp, act)
		})
	}
}

func TestParseAzureDevopsRepo(t *testing.T) {
	// this should be successful
	repo := ADRepo
	repo.ParentRepository = nil
	r, err := parser.ParseAzureDevopsRepo(&repo)
	Ok(t, err)
	Equals(t, models.Repo{
		Owner:             "owner/project",
		FullName:          "owner/project/repo",
		CloneURL:          "https://azuredevops-user:azuredevops-token@dev.azure.com/owner/project/_git/repo",
		SanitizedCloneURL: "https://azuredevops-user:<redacted>@dev.azure.com/owner/project/_git/repo",
		Name:              "repo",
		VCSHost: models.VCSHost{
			Hostname: "dev.azure.com",
			Type:     models.AzureDevops,
		},
	}, r)

	// this should be successful
	repo = ADRepo
	repo.WebURL = nil
	r, err = parser.ParseAzureDevopsRepo(&repo)
	Ok(t, err)
	Equals(t, models.Repo{
		Owner:             "owner/project",
		FullName:          "owner/project/repo",
		CloneURL:          "https://azuredevops-user:azuredevops-token@dev.azure.com/owner/project/_git/repo",
		SanitizedCloneURL: "https://azuredevops-user:<redacted>@dev.azure.com/owner/project/_git/repo",
		Name:              "repo",
		VCSHost: models.VCSHost{
			Hostname: "dev.azure.com",
			Type:     models.AzureDevops,
		},
	}, r)

	// this should be successful
	repo = ADRepo
	repo.WebURL = azuredevops.String("https://owner.visualstudio.com/project/_git/repo")
	r, err = parser.ParseAzureDevopsRepo(&repo)
	Ok(t, err)
	Equals(t, models.Repo{
		Owner:             "owner/project",
		FullName:          "owner/project/repo",
		CloneURL:          "https://azuredevops-user:azuredevops-token@owner.visualstudio.com/project/_git/repo",
		SanitizedCloneURL: "https://azuredevops-user:<redacted>@owner.visualstudio.com/project/_git/repo",
		Name:              "repo",
		VCSHost: models.VCSHost{
			Hostname: "owner.visualstudio.com",
			Type:     models.AzureDevops,
		},
	}, r)

	// this should be successful
	repo = ADRepo
	repo.WebURL = azuredevops.String("https://dev.azure.com/owner/project/_git/repo")
	r, err = parser.ParseAzureDevopsRepo(&repo)
	Ok(t, err)
	Equals(t, models.Repo{
		Owner:             "owner/project",
		FullName:          "owner/project/repo",
		CloneURL:          "https://azuredevops-user:azuredevops-token@dev.azure.com/owner/project/_git/repo",
		SanitizedCloneURL: "https://azuredevops-user:<redacted>@dev.azure.com/owner/project/_git/repo",
		Name:              "repo",
		VCSHost: models.VCSHost{
			Hostname: "dev.azure.com",
			Type:     models.AzureDevops,
		},
	}, r)
}

func TestParseAzureDevopsPullEvent(t *testing.T) {
	_, _, _, _, _, err := parser.ParseAzureDevopsPullEvent(ADPullEvent)
	Ok(t, err)

	testPull := deepcopy.Copy(ADPull).(azuredevops.GitPullRequest)
	testPull.LastMergeSourceCommit.CommitID = nil
	_, _, _, err = parser.ParseAzureDevopsPull(&testPull)
	ErrEquals(t, "lastMergeSourceCommit.commitID is null", err)

	testPull = deepcopy.Copy(ADPull).(azuredevops.GitPullRequest)
	testPull.URL = nil
	_, _, _, err = parser.ParseAzureDevopsPull(&testPull)
	ErrEquals(t, "url is null", err)
	testEvent := deepcopy.Copy(ADPullEvent).(azuredevops.Event)
	resource := deepcopy.Copy(testEvent.Resource).(*azuredevops.GitPullRequest)
	resource.CreatedBy = nil
	testEvent.Resource = resource
	_, _, _, _, _, err = parser.ParseAzureDevopsPullEvent(testEvent)
	ErrEquals(t, "CreatedBy is null", err)

	testEvent = deepcopy.Copy(ADPullEvent).(azuredevops.Event)
	resource = deepcopy.Copy(testEvent.Resource).(*azuredevops.GitPullRequest)
	resource.CreatedBy.UniqueName = azuredevops.String("")
	testEvent.Resource = resource
	_, _, _, _, _, err = parser.ParseAzureDevopsPullEvent(testEvent)
	ErrEquals(t, "CreatedBy.UniqueName is null", err)

	actPull, evType, actBaseRepo, actHeadRepo, actUser, err := parser.ParseAzureDevopsPullEvent(ADPullEvent)
	Ok(t, err)
	expBaseRepo := models.Repo{
		Owner:             "owner/project",
		FullName:          "owner/project/repo",
		CloneURL:          "https://azuredevops-user:azuredevops-token@dev.azure.com/owner/project/_git/repo",
		SanitizedCloneURL: "https://azuredevops-user:<redacted>@dev.azure.com/owner/project/_git/repo",
		Name:              "repo",
		VCSHost: models.VCSHost{
			Hostname: "dev.azure.com",
			Type:     models.AzureDevops,
		},
	}
	Equals(t, expBaseRepo, actBaseRepo)
	Equals(t, expBaseRepo, actHeadRepo)
	Equals(t, models.PullRequest{
		URL:        ADPull.GetURL(),
		Author:     ADPull.CreatedBy.GetUniqueName(),
		HeadBranch: "feature/sourceBranch",
		BaseBranch: "targetBranch",
		HeadCommit: ADPull.LastMergeSourceCommit.GetCommitID(),
		Num:        ADPull.GetPullRequestID(),
		State:      models.OpenPullState,
		BaseRepo:   expBaseRepo,
	}, actPull)
	Equals(t, models.OpenedPullEvent, evType)
	Equals(t, models.User{Username: "user@example.com"}, actUser)
}

func TestParseAzureDevopsPullEvent_EventType(t *testing.T) {
	cases := []struct {
		action string
		exp    models.PullRequestEventType
	}{
		{
			action: "git.pullrequest.updated",
			exp:    models.UpdatedPullEvent,
		},
		{
			action: "git.pullrequest.created",
			exp:    models.OpenedPullEvent,
		},
		{
			action: "git.pullrequest.updated",
			exp:    models.ClosedPullEvent,
		},
		{
			action: "anything_else",
			exp:    models.OtherPullEvent,
		},
	}

	for _, c := range cases {
		t.Run(c.action, func(t *testing.T) {
			event := deepcopy.Copy(ADPullEvent).(azuredevops.Event)
			if c.exp == models.ClosedPullEvent {
				event = deepcopy.Copy(ADPullClosedEvent).(azuredevops.Event)
			}
			event.EventType = c.action
			_, actType, _, _, _, err := parser.ParseAzureDevopsPullEvent(event)
			Ok(t, err)
			Equals(t, c.exp, actType)
		})
	}
}

func TestParseAzureDevopsPull(t *testing.T) {
	testPull := deepcopy.Copy(ADPull).(azuredevops.GitPullRequest)
	testPull.LastMergeSourceCommit.CommitID = nil
	_, _, _, err := parser.ParseAzureDevopsPull(&testPull)
	ErrEquals(t, "lastMergeSourceCommit.commitID is null", err)

	testPull = deepcopy.Copy(ADPull).(azuredevops.GitPullRequest)
	testPull.URL = nil
	_, _, _, err = parser.ParseAzureDevopsPull(&testPull)
	ErrEquals(t, "url is null", err)

	testPull = deepcopy.Copy(ADPull).(azuredevops.GitPullRequest)
	testPull.SourceRefName = nil
	_, _, _, err = parser.ParseAzureDevopsPull(&testPull)
	ErrEquals(t, "sourceRefName (branch name) is null", err)

	testPull = deepcopy.Copy(ADPull).(azuredevops.GitPullRequest)
	testPull.TargetRefName = nil
	_, _, _, err = parser.ParseAzureDevopsPull(&testPull)
	ErrEquals(t, "targetRefName (branch name) is null", err)

	testPull = deepcopy.Copy(ADPull).(azuredevops.GitPullRequest)
	testPull.CreatedBy = nil
	_, _, _, err = parser.ParseAzureDevopsPull(&testPull)
	ErrEquals(t, "CreatedBy is null", err)

	testPull = deepcopy.Copy(ADPull).(azuredevops.GitPullRequest)
	testPull.CreatedBy.UniqueName = nil
	_, _, _, err = parser.ParseAzureDevopsPull(&testPull)
	ErrEquals(t, "CreatedBy.UniqueName is null", err)

	testPull = deepcopy.Copy(ADPull).(azuredevops.GitPullRequest)
	testPull.PullRequestID = nil
	_, _, _, err = parser.ParseAzureDevopsPull(&testPull)
	ErrEquals(t, "pullRequestId is null", err)

	actPull, actBaseRepo, actHeadRepo, err := parser.ParseAzureDevopsPull(&ADPull)
	Ok(t, err)
	expBaseRepo := models.Repo{
		Owner:             "owner/project",
		FullName:          "owner/project/repo",
		CloneURL:          "https://azuredevops-user:azuredevops-token@dev.azure.com/owner/project/_git/repo",
		SanitizedCloneURL: "https://azuredevops-user:<redacted>@dev.azure.com/owner/project/_git/repo",
		Name:              "repo",
		VCSHost: models.VCSHost{
			Hostname: "dev.azure.com",
			Type:     models.AzureDevops,
		},
	}
	Equals(t, models.PullRequest{
		URL:        ADPull.GetURL(),
		Author:     ADPull.CreatedBy.GetUniqueName(),
		HeadBranch: "feature/sourceBranch",
		BaseBranch: "targetBranch",
		HeadCommit: ADPull.LastMergeSourceCommit.GetCommitID(),
		Num:        ADPull.GetPullRequestID(),
		State:      models.OpenPullState,
		BaseRepo:   expBaseRepo,
	}, actPull)
	Equals(t, expBaseRepo, actBaseRepo)
	Equals(t, expBaseRepo, actHeadRepo)
}

func TestParseAzureDevopsSelfHostedRepo(t *testing.T) {
	// this should be successful
	repo := ADSelfRepo
	repo.ParentRepository = nil
	r, err := parser.ParseAzureDevopsRepo(&repo)
	Ok(t, err)
	Equals(t, models.Repo{
		Owner:             "owner/project",
		FullName:          "owner/project/repo",
		CloneURL:          "https://azuredevops-user:azuredevops-token@devops.abc.com/owner/project/_git/repo",
		SanitizedCloneURL: "https://azuredevops-user:<redacted>@devops.abc.com/owner/project/_git/repo",
		Name:              "repo",
		VCSHost: models.VCSHost{
			Hostname: "devops.abc.com",
			Type:     models.AzureDevops,
		},
	}, r)

}

func TestParseAzureDevopsSelfHostedPullEvent(t *testing.T) {
	_, _, _, _, _, err := parser.ParseAzureDevopsPullEvent(ADSelfPullEvent)
	Ok(t, err)

	testPull := deepcopy.Copy(ADSelfPull).(azuredevops.GitPullRequest)
	testPull.LastMergeSourceCommit.CommitID = nil
	_, _, _, err = parser.ParseAzureDevopsPull(&testPull)
	ErrEquals(t, "lastMergeSourceCommit.commitID is null", err)

	testPull = deepcopy.Copy(ADSelfPull).(azuredevops.GitPullRequest)
	testPull.URL = nil
	_, _, _, err = parser.ParseAzureDevopsPull(&testPull)
	ErrEquals(t, "url is null", err)
	testEvent := deepcopy.Copy(ADSelfPullEvent).(azuredevops.Event)
	resource := deepcopy.Copy(testEvent.Resource).(*azuredevops.GitPullRequest)
	resource.CreatedBy = nil
	testEvent.Resource = resource
	_, _, _, _, _, err = parser.ParseAzureDevopsPullEvent(testEvent)
	ErrEquals(t, "CreatedBy is null", err)

	testEvent = deepcopy.Copy(ADSelfPullEvent).(azuredevops.Event)
	resource = deepcopy.Copy(testEvent.Resource).(*azuredevops.GitPullRequest)
	resource.CreatedBy.UniqueName = azuredevops.String("")
	testEvent.Resource = resource
	_, _, _, _, _, err = parser.ParseAzureDevopsPullEvent(testEvent)
	ErrEquals(t, "CreatedBy.UniqueName is null", err)

	actPull, evType, actBaseRepo, actHeadRepo, actUser, err := parser.ParseAzureDevopsPullEvent(ADSelfPullEvent)
	Ok(t, err)
	expBaseRepo := models.Repo{
		Owner:             "owner/project",
		FullName:          "owner/project/repo",
		CloneURL:          "https://azuredevops-user:azuredevops-token@devops.abc.com/owner/project/_git/repo",
		SanitizedCloneURL: "https://azuredevops-user:<redacted>@devops.abc.com/owner/project/_git/repo",
		Name:              "repo",
		VCSHost: models.VCSHost{
			Hostname: "devops.abc.com",
			Type:     models.AzureDevops,
		},
	}
	Equals(t, expBaseRepo, actBaseRepo)
	Equals(t, expBaseRepo, actHeadRepo)
	Equals(t, models.PullRequest{
		URL:        ADSelfPull.GetURL(),
		Author:     ADSelfPull.CreatedBy.GetUniqueName(),
		HeadBranch: "feature/sourceBranch",
		BaseBranch: "targetBranch",
		HeadCommit: ADSelfPull.LastMergeSourceCommit.GetCommitID(),
		Num:        ADSelfPull.GetPullRequestID(),
		State:      models.OpenPullState,
		BaseRepo:   expBaseRepo,
	}, actPull)
	Equals(t, models.OpenedPullEvent, evType)
	Equals(t, models.User{Username: "user@example.com"}, actUser)
}

func TestParseAzureDevopsSelfHostedPullEvent_EventType(t *testing.T) {
	cases := []struct {
		action string
		exp    models.PullRequestEventType
	}{
		{
			action: "git.pullrequest.updated",
			exp:    models.UpdatedPullEvent,
		},
		{
			action: "git.pullrequest.created",
			exp:    models.OpenedPullEvent,
		},
		{
			action: "git.pullrequest.updated",
			exp:    models.ClosedPullEvent,
		},
		{
			action: "anything_else",
			exp:    models.OtherPullEvent,
		},
	}

	for _, c := range cases {
		t.Run(c.action, func(t *testing.T) {
			event := deepcopy.Copy(ADSelfPullEvent).(azuredevops.Event)
			if c.exp == models.ClosedPullEvent {
				event = deepcopy.Copy(ADSelfPullClosedEvent).(azuredevops.Event)
			}
			event.EventType = c.action
			_, actType, _, _, _, err := parser.ParseAzureDevopsPullEvent(event)
			Ok(t, err)
			Equals(t, c.exp, actType)
		})
	}
}

func TestParseAzureSelfHostedDevopsPull(t *testing.T) {
	testPull := deepcopy.Copy(ADSelfPull).(azuredevops.GitPullRequest)
	testPull.LastMergeSourceCommit.CommitID = nil
	_, _, _, err := parser.ParseAzureDevopsPull(&testPull)
	ErrEquals(t, "lastMergeSourceCommit.commitID is null", err)

	testPull = deepcopy.Copy(ADSelfPull).(azuredevops.GitPullRequest)
	testPull.URL = nil
	_, _, _, err = parser.ParseAzureDevopsPull(&testPull)
	ErrEquals(t, "url is null", err)

	testPull = deepcopy.Copy(ADSelfPull).(azuredevops.GitPullRequest)
	testPull.SourceRefName = nil
	_, _, _, err = parser.ParseAzureDevopsPull(&testPull)
	ErrEquals(t, "sourceRefName (branch name) is null", err)

	testPull = deepcopy.Copy(ADSelfPull).(azuredevops.GitPullRequest)
	testPull.TargetRefName = nil
	_, _, _, err = parser.ParseAzureDevopsPull(&testPull)
	ErrEquals(t, "targetRefName (branch name) is null", err)

	testPull = deepcopy.Copy(ADSelfPull).(azuredevops.GitPullRequest)
	testPull.CreatedBy = nil
	_, _, _, err = parser.ParseAzureDevopsPull(&testPull)
	ErrEquals(t, "CreatedBy is null", err)

	testPull = deepcopy.Copy(ADSelfPull).(azuredevops.GitPullRequest)
	testPull.CreatedBy.UniqueName = nil
	_, _, _, err = parser.ParseAzureDevopsPull(&testPull)
	ErrEquals(t, "CreatedBy.UniqueName is null", err)

	testPull = deepcopy.Copy(ADSelfPull).(azuredevops.GitPullRequest)
	testPull.PullRequestID = nil
	_, _, _, err = parser.ParseAzureDevopsPull(&testPull)
	ErrEquals(t, "pullRequestId is null", err)

	actPull, actBaseRepo, actHeadRepo, err := parser.ParseAzureDevopsPull(&ADSelfPull)
	Ok(t, err)
	expBaseRepo := models.Repo{
		Owner:             "owner/project",
		FullName:          "owner/project/repo",
		CloneURL:          "https://azuredevops-user:azuredevops-token@devops.abc.com/owner/project/_git/repo",
		SanitizedCloneURL: "https://azuredevops-user:<redacted>@devops.abc.com/owner/project/_git/repo",
		Name:              "repo",
		VCSHost: models.VCSHost{
			Hostname: "devops.abc.com",
			Type:     models.AzureDevops,
		},
	}
	Equals(t, models.PullRequest{
		URL:        ADSelfPull.GetURL(),
		Author:     ADSelfPull.CreatedBy.GetUniqueName(),
		HeadBranch: "feature/sourceBranch",
		BaseBranch: "targetBranch",
		HeadCommit: ADSelfPull.LastMergeSourceCommit.GetCommitID(),
		Num:        ADSelfPull.GetPullRequestID(),
		State:      models.OpenPullState,
		BaseRepo:   expBaseRepo,
	}, actPull)
	Equals(t, expBaseRepo, actBaseRepo)
	Equals(t, expBaseRepo, actHeadRepo)
}
