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
//
package events_test

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/go-github/github"
	"github.com/lkysow/go-gitlab"
	"github.com/mohae/deepcopy"
	"github.com/runatlantis/atlantis/server/events"
	"github.com/runatlantis/atlantis/server/events/models"
	. "github.com/runatlantis/atlantis/server/events/vcs/fixtures"
	. "github.com/runatlantis/atlantis/testing"
)

var parser = events.EventParser{
	GithubUser:         "github-user",
	GithubToken:        "github-token",
	GitlabUser:         "gitlab-user",
	GitlabToken:        "gitlab-token",
	BitbucketUser:      "bitbucket-user",
	BitbucketToken:     "bitbucket-token",
	BitbucketServerURL: "http://mycorp.com:7490",
}

func TestParseGithubRepo(t *testing.T) {
	r, err := parser.ParseGithubRepo(&Repo)
	Ok(t, err)
	Equals(t, models.Repo{
		Owner:             "owner",
		FullName:          "owner/repo",
		CloneURL:          "https://github-user:github-token@github.com/owner/repo.git",
		SanitizedCloneURL: Repo.GetCloneURL(),
		Name:              "repo",
		VCSHost: models.VCSHost{
			Hostname: "github.com",
			Type:     models.Github,
		},
	}, r)
}

func TestParseGithubIssueCommentEvent(t *testing.T) {
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
	_, _, _, err := parser.ParseGithubIssueCommentEvent(&testComment)
	ErrEquals(t, "comment.user.login is null", err)

	testComment = deepcopy.Copy(comment).(github.IssueCommentEvent)
	testComment.Comment.User = nil
	_, _, _, err = parser.ParseGithubIssueCommentEvent(&testComment)
	ErrEquals(t, "comment.user.login is null", err)

	testComment = deepcopy.Copy(comment).(github.IssueCommentEvent)
	testComment.Comment.User.Login = nil
	_, _, _, err = parser.ParseGithubIssueCommentEvent(&testComment)
	ErrEquals(t, "comment.user.login is null", err)

	testComment = deepcopy.Copy(comment).(github.IssueCommentEvent)
	testComment.Issue = nil
	_, _, _, err = parser.ParseGithubIssueCommentEvent(&testComment)
	ErrEquals(t, "issue.number is null", err)

	// this should be successful
	repo, user, pullNum, err := parser.ParseGithubIssueCommentEvent(&comment)
	Ok(t, err)
	Equals(t, models.Repo{
		Owner:             *comment.Repo.Owner.Login,
		FullName:          *comment.Repo.FullName,
		CloneURL:          "https://github-user:github-token@github.com/owner/repo.git",
		SanitizedCloneURL: *comment.Repo.CloneURL,
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
	_, _, _, _, _, err := parser.ParseGithubPullEvent(&github.PullRequestEvent{})
	ErrEquals(t, "pull_request is null", err)

	testEvent := deepcopy.Copy(PullEvent).(github.PullRequestEvent)
	testEvent.PullRequest.HTMLURL = nil
	_, _, _, _, _, err = parser.ParseGithubPullEvent(&testEvent)
	ErrEquals(t, "html_url is null", err)

	testEvent = deepcopy.Copy(PullEvent).(github.PullRequestEvent)
	testEvent.Sender = nil
	_, _, _, _, _, err = parser.ParseGithubPullEvent(&testEvent)
	ErrEquals(t, "sender is null", err)

	testEvent = deepcopy.Copy(PullEvent).(github.PullRequestEvent)
	testEvent.Sender.Login = nil
	_, _, _, _, _, err = parser.ParseGithubPullEvent(&testEvent)
	ErrEquals(t, "sender.login is null", err)

	actPull, evType, actBaseRepo, actHeadRepo, actUser, err := parser.ParseGithubPullEvent(&PullEvent)
	Ok(t, err)
	expBaseRepo := models.Repo{
		Owner:             "owner",
		FullName:          "owner/repo",
		CloneURL:          "https://github-user:github-token@github.com/owner/repo.git",
		SanitizedCloneURL: Repo.GetCloneURL(),
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
		Branch:     Pull.Head.GetRef(),
		HeadCommit: Pull.Head.GetSHA(),
		Num:        Pull.GetNumber(),
		State:      models.OpenPullState,
		BaseRepo:   expBaseRepo,
	}, actPull)
	Equals(t, models.OpenedPullEvent, evType)
	Equals(t, models.User{Username: "user"}, actUser)
}

func TestParseGithubPullEvent_EventType(t *testing.T) {
	cases := []struct {
		action string
		exp    models.PullRequestEventType
	}{
		{
			action: "synchronize",
			exp:    models.UpdatedPullEvent,
		},
		{
			action: "unassigned",
			exp:    models.OtherPullEvent,
		},
		{
			action: "review_requested",
			exp:    models.OtherPullEvent,
		},
		{
			action: "review_request_removed",
			exp:    models.OtherPullEvent,
		},
		{
			action: "labeled",
			exp:    models.OtherPullEvent,
		},
		{
			action: "unlabeled",
			exp:    models.OtherPullEvent,
		},
		{
			action: "opened",
			exp:    models.OpenedPullEvent,
		},
		{
			action: "edited",
			exp:    models.OtherPullEvent,
		},
		{
			action: "closed",
			exp:    models.ClosedPullEvent,
		},
		{
			action: "reopened",
			exp:    models.OtherPullEvent,
		},
	}

	for _, c := range cases {
		t.Run(c.action, func(t *testing.T) {
			event := deepcopy.Copy(PullEvent).(github.PullRequestEvent)
			event.Action = &c.action
			_, actType, _, _, _, err := parser.ParseGithubPullEvent(&event)
			Ok(t, err)
			Equals(t, c.exp, actType)
		})
	}
}

func TestParseGithubPull(t *testing.T) {
	testPull := deepcopy.Copy(Pull).(github.PullRequest)
	testPull.Head.SHA = nil
	_, _, _, err := parser.ParseGithubPull(&testPull)
	ErrEquals(t, "head.sha is null", err)

	testPull = deepcopy.Copy(Pull).(github.PullRequest)
	testPull.HTMLURL = nil
	_, _, _, err = parser.ParseGithubPull(&testPull)
	ErrEquals(t, "html_url is null", err)

	testPull = deepcopy.Copy(Pull).(github.PullRequest)
	testPull.Head.Ref = nil
	_, _, _, err = parser.ParseGithubPull(&testPull)
	ErrEquals(t, "head.ref is null", err)

	testPull = deepcopy.Copy(Pull).(github.PullRequest)
	testPull.User.Login = nil
	_, _, _, err = parser.ParseGithubPull(&testPull)
	ErrEquals(t, "user.login is null", err)

	testPull = deepcopy.Copy(Pull).(github.PullRequest)
	testPull.Number = nil
	_, _, _, err = parser.ParseGithubPull(&testPull)
	ErrEquals(t, "number is null", err)

	pullRes, actBaseRepo, actHeadRepo, err := parser.ParseGithubPull(&Pull)
	Ok(t, err)
	expBaseRepo := models.Repo{
		Owner:             "owner",
		FullName:          "owner/repo",
		CloneURL:          "https://github-user:github-token@github.com/owner/repo.git",
		SanitizedCloneURL: Repo.GetCloneURL(),
		Name:              "repo",
		VCSHost: models.VCSHost{
			Hostname: "github.com",
			Type:     models.Github,
		},
	}
	Equals(t, models.PullRequest{
		URL:        Pull.GetHTMLURL(),
		Author:     Pull.User.GetLogin(),
		Branch:     Pull.Head.GetRef(),
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
	var event *gitlab.MergeEvent
	err := json.Unmarshal([]byte(mergeEventJSON), &event)
	Ok(t, err)
	pull, evType, actBaseRepo, actHeadRepo, actUser, err := parser.ParseGitlabMergeEvent(*event)
	Ok(t, err)

	expBaseRepo := models.Repo{
		FullName:          "gitlabhq/gitlab-test",
		Name:              "gitlab-test",
		SanitizedCloneURL: "https://example.com/gitlabhq/gitlab-test.git",
		Owner:             "gitlabhq",
		CloneURL:          "https://gitlab-user:gitlab-token@example.com/gitlabhq/gitlab-test.git",
		VCSHost: models.VCSHost{
			Hostname: "example.com",
			Type:     models.Gitlab,
		},
	}

	Equals(t, models.PullRequest{
		URL:        "http://example.com/diaspora/merge_requests/1",
		Author:     "root",
		Num:        1,
		HeadCommit: "da1560886d4f094c3e6c9ef40349f7d38b5d27d7",
		Branch:     "ms-viewport",
		State:      models.OpenPullState,
		BaseRepo:   expBaseRepo,
	}, pull)
	Equals(t, models.OpenedPullEvent, evType)

	Equals(t, expBaseRepo, actBaseRepo)
	Equals(t, models.Repo{
		FullName:          "awesome_space/awesome_project",
		Name:              "awesome_project",
		SanitizedCloneURL: "http://example.com/awesome_space/awesome_project.git",
		Owner:             "awesome_space",
		CloneURL:          "http://gitlab-user:gitlab-token@example.com/awesome_space/awesome_project.git",
		VCSHost: models.VCSHost{
			Hostname: "example.com",
			Type:     models.Gitlab,
		},
	}, actHeadRepo)
	Equals(t, models.User{Username: "root"}, actUser)

	t.Log("If the state is closed, should set field correctly.")
	event.ObjectAttributes.State = "closed"
	pull, _, _, _, _, err = parser.ParseGitlabMergeEvent(*event)
	Ok(t, err)
	Equals(t, models.ClosedPullState, pull.State)
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
			action: "update",
			exp:    models.UpdatedPullEvent,
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

	for _, c := range cases {
		t.Run(c.action, func(t *testing.T) {
			var event *gitlab.MergeEvent
			eventJSON := strings.Replace(mergeEventJSON, `"action": "open"`, fmt.Sprintf(`"action": %q`, c.action), 1)
			err := json.Unmarshal([]byte(eventJSON), &event)
			Ok(t, err)
			_, evType, _, _, _, err := parser.ParseGitlabMergeEvent(*event)
			Ok(t, err)
			Equals(t, c.exp, evType)
		})
	}
}

func TestParseGitlabMergeRequest(t *testing.T) {
	t.Log("should properly parse a gitlab merge request")
	var event *gitlab.MergeRequest
	err := json.Unmarshal([]byte(mergeRequestJSON), &event)
	Ok(t, err)
	repo := models.Repo{
		FullName:          "gitlabhq/gitlab-test",
		Name:              "gitlab-test",
		SanitizedCloneURL: "https://example.com/gitlabhq/gitlab-test.git",
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
		Branch:     "abc",
		State:      models.OpenPullState,
		BaseRepo:   repo,
	}, pull)

	t.Log("If the state is closed, should set field correctly.")
	event.State = "closed"
	pull = parser.ParseGitlabMergeRequest(event, repo)
	Equals(t, models.ClosedPullState, pull.State)
}

func TestParseGitlabMergeCommentEvent(t *testing.T) {
	t.Log("should properly parse a gitlab merge comment event")
	var event *gitlab.MergeCommentEvent
	err := json.Unmarshal([]byte(mergeCommentEventJSON), &event)
	Ok(t, err)
	baseRepo, headRepo, user, err := parser.ParseGitlabMergeCommentEvent(*event)
	Ok(t, err)
	Equals(t, models.Repo{
		FullName:          "gitlabhq/gitlab-test",
		Name:              "gitlab-test",
		SanitizedCloneURL: "https://example.com/gitlabhq/gitlab-test.git",
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
		SanitizedCloneURL: "https://example.com/gitlab-org/gitlab-test.git",
		Owner:             "gitlab-org",
		CloneURL:          "https://gitlab-user:gitlab-token@example.com/gitlab-org/gitlab-test.git",
		VCSHost: models.VCSHost{
			Hostname: "example.com",
			Type:     models.Gitlab,
		},
	}, headRepo)
	Equals(t, models.User{
		Username: "root",
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
			cmd := events.NewCommentCommand(c.RepoRelDir, nil, events.PlanCommand, false, "workspace", "")
			Equals(t, c.ExpDir, cmd.RepoRelDir)
		})
	}
}

func TestNewCommand_EmptyDirWorkspaceProject(t *testing.T) {
	cmd := events.NewCommentCommand("", nil, events.PlanCommand, false, "", "")
	Equals(t, events.CommentCommand{
		RepoRelDir:  "",
		Flags:       nil,
		Name:        events.PlanCommand,
		Verbose:     false,
		Workspace:   "",
		ProjectName: "",
	}, *cmd)
}

func TestNewCommand_AllFieldsSet(t *testing.T) {
	cmd := events.NewCommentCommand("dir", []string{"a", "b"}, events.PlanCommand, true, "workspace", "project")
	Equals(t, events.CommentCommand{
		Workspace:   "workspace",
		RepoRelDir:  "dir",
		Verbose:     true,
		Flags:       []string{"a", "b"},
		Name:        events.PlanCommand,
		ProjectName: "project",
	}, *cmd)
}

func TestAutoplanCommand_CommandName(t *testing.T) {
	Equals(t, events.PlanCommand, (events.AutoplanCommand{}).CommandName())
}

func TestAutoplanCommand_IsVerbose(t *testing.T) {
	Equals(t, false, (events.AutoplanCommand{}).IsVerbose())
}

func TestAutoplanCommand_IsAutoplan(t *testing.T) {
	Equals(t, true, (events.AutoplanCommand{}).IsAutoplan())
}

func TestCommentCommand_CommandName(t *testing.T) {
	Equals(t, events.PlanCommand, (events.CommentCommand{
		Name: events.PlanCommand,
	}).CommandName())
	Equals(t, events.ApplyCommand, (events.CommentCommand{
		Name: events.ApplyCommand,
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
	exp := `command="plan" verbose=true dir="mydir" workspace="myworkspace" project="myproject" flags="flag1,flag2"`
	Equals(t, exp, (events.CommentCommand{
		RepoRelDir:  "mydir",
		Flags:       []string{"flag1", "flag2"},
		Name:        events.PlanCommand,
		Verbose:     true,
		Workspace:   "myworkspace",
		ProjectName: "myproject",
	}).String())
}

func TestParseBitbucketCloudCommentEvent_EmptyString(t *testing.T) {
	_, _, _, _, _, err := parser.ParseBitbucketCloudCommentEvent([]byte(""))
	ErrEquals(t, "parsing json: unexpected end of JSON input", err)
}

func TestParseBitbucketCloudCommentEvent_EmptyObject(t *testing.T) {
	_, _, _, _, _, err := parser.ParseBitbucketCloudCommentEvent([]byte("{}"))
	ErrContains(t, "Key: 'CommentEvent.CommonEventData.Actor' Error:Field validation for 'Actor' failed on the 'required' tag\nKey: 'CommentEvent.CommonEventData.Repository' Error:Field validation for 'Repository' failed on the 'required' tag\nKey: 'CommentEvent.CommonEventData.PullRequest' Error:Field validation for 'PullRequest' failed on the 'required' tag\nKey: 'CommentEvent.Comment' Error:Field validation for 'Comment' failed on the 'required' tag", err)
}

func TestParseBitbucketCloudCommentEvent_CommitHashMissing(t *testing.T) {
	path := filepath.Join("testdata", "bitbucket-cloud-comment-event.json")
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		Ok(t, err)
	}
	emptyCommitHash := strings.Replace(string(bytes), `        "hash": "e0624da46d3a",`, "", -1)
	_, _, _, _, _, err = parser.ParseBitbucketCloudCommentEvent([]byte(emptyCommitHash))
	ErrContains(t, "Key: 'CommentEvent.CommonEventData.PullRequest.Source.Commit.Hash' Error:Field validation for 'Hash' failed on the 'required' tag", err)
}

func TestParseBitbucketCloudCommentEvent_ValidEvent(t *testing.T) {
	path := filepath.Join("testdata", "bitbucket-cloud-comment-event.json")
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		Ok(t, err)
	}
	pull, baseRepo, headRepo, user, comment, err := parser.ParseBitbucketCloudCommentEvent(bytes)
	Ok(t, err)
	expBaseRepo := models.Repo{
		FullName:          "lkysow/atlantis-example",
		Owner:             "lkysow",
		Name:              "atlantis-example",
		CloneURL:          "https://bitbucket-user:bitbucket-token@bitbucket.org/lkysow/atlantis-example.git",
		SanitizedCloneURL: "https://bitbucket.org/lkysow/atlantis-example.git",
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
		Branch:     "lkysow/maintf-edited-online-with-bitbucket-1532029690581",
		Author:     "lkysow",
		State:      models.ClosedPullState,
		BaseRepo:   expBaseRepo,
	}, pull)
	Equals(t, models.Repo{
		FullName:          "lkysow-fork/atlantis-example",
		Owner:             "lkysow-fork",
		Name:              "atlantis-example",
		CloneURL:          "https://bitbucket-user:bitbucket-token@bitbucket.org/lkysow-fork/atlantis-example.git",
		SanitizedCloneURL: "https://bitbucket.org/lkysow-fork/atlantis-example.git",
		VCSHost: models.VCSHost{
			Hostname: "bitbucket.org",
			Type:     models.BitbucketCloud,
		},
	}, headRepo)
	Equals(t, models.User{
		Username: "lkysow",
	}, user)
	Equals(t, "my comment", comment)
}

func TestParseBitbucketCloudCommentEvent_MultipleStates(t *testing.T) {
	path := filepath.Join("testdata", "bitbucket-cloud-comment-event.json")
	bytes, err := ioutil.ReadFile(path)
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
			"DECLINE",
			models.ClosedPullState,
		},
	}

	for _, c := range cases {
		t.Run(c.pullState, func(t *testing.T) {
			withState := strings.Replace(string(bytes), `"state": "MERGED"`, fmt.Sprintf(`"state": "%s"`, c.pullState), -1)
			pull, _, _, _, _, err := parser.ParseBitbucketCloudCommentEvent([]byte(withState))
			Ok(t, err)
			Equals(t, c.exp, pull.State)
		})
	}
}

func TestParseBitbucketCloudPullEvent_ValidEvent(t *testing.T) {
	path := filepath.Join("testdata", "bitbucket-cloud-pull-event-fulfilled.json")
	bytes, err := ioutil.ReadFile(path)
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
		SanitizedCloneURL: "https://bitbucket.org/lkysow/atlantis-example.git",
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
		Branch:     "lkysow/maintf-edited-online-with-bitbucket-1532029690581",
		Author:     "lkysow",
		State:      models.ClosedPullState,
		BaseRepo:   expBaseRepo,
	}, pull)
	Equals(t, models.Repo{
		FullName:          "lkysow-fork/atlantis-example",
		Owner:             "lkysow-fork",
		Name:              "atlantis-example",
		CloneURL:          "https://bitbucket-user:bitbucket-token@bitbucket.org/lkysow-fork/atlantis-example.git",
		SanitizedCloneURL: "https://bitbucket.org/lkysow-fork/atlantis-example.git",
		VCSHost: models.VCSHost{
			Hostname: "bitbucket.org",
			Type:     models.BitbucketCloud,
		},
	}, headRepo)
	Equals(t, models.User{
		Username: "lkysow",
	}, user)
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
			act := parser.GetBitbucketCloudEventType(c.header)
			Equals(t, c.exp, act)
		})
	}
}

func TestParseBitbucketServerCommentEvent_EmptyString(t *testing.T) {
	_, _, _, _, _, err := parser.ParseBitbucketServerCommentEvent([]byte(""))
	ErrEquals(t, "parsing json: unexpected end of JSON input", err)
}

func TestParseBitbucketServerCommentEvent_EmptyObject(t *testing.T) {
	_, _, _, _, _, err := parser.ParseBitbucketServerCommentEvent([]byte("{}"))
	ErrContains(t, `API response "{}" was missing fields: Key: 'CommentEvent.CommonEventData.Actor' Error:Field validation for 'Actor' failed on the 'required' tag`, err)
}

func TestParseBitbucketServerCommentEvent_CommitHashMissing(t *testing.T) {
	path := filepath.Join("testdata", "bitbucket-server-comment-event.json")
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		Ok(t, err)
	}
	emptyCommitHash := strings.Replace(string(bytes), `"latestCommit": "bfb1af1ba9c2a2fa84cd61af67e6e1b60a22e060",`, "", -1)
	_, _, _, _, _, err = parser.ParseBitbucketServerCommentEvent([]byte(emptyCommitHash))
	ErrContains(t, "Key: 'CommentEvent.CommonEventData.PullRequest.FromRef.LatestCommit' Error:Field validation for 'LatestCommit' failed on the 'required' tag", err)
}

func TestParseBitbucketServerCommentEvent_ValidEvent(t *testing.T) {
	path := filepath.Join("testdata", "bitbucket-server-comment-event.json")
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		Ok(t, err)
	}
	pull, baseRepo, headRepo, user, comment, err := parser.ParseBitbucketServerCommentEvent(bytes)
	Ok(t, err)
	expBaseRepo := models.Repo{
		FullName:          "atlantis/atlantis-example",
		Owner:             "atlantis",
		Name:              "atlantis-example",
		CloneURL:          "http://bitbucket-user:bitbucket-token@mycorp.com:7490/scm/at/atlantis-example.git",
		SanitizedCloneURL: "http://mycorp.com:7490/scm/at/atlantis-example.git",
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
		Branch:     "branch",
		Author:     "lkysow",
		State:      models.OpenPullState,
		BaseRepo:   expBaseRepo,
	}, pull)
	Equals(t, models.Repo{
		FullName:          "atlantis-fork/atlantis-example",
		Owner:             "atlantis-fork",
		Name:              "atlantis-example",
		CloneURL:          "http://bitbucket-user:bitbucket-token@mycorp.com:7490/scm/fk/atlantis-example.git",
		SanitizedCloneURL: "http://mycorp.com:7490/scm/fk/atlantis-example.git",
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
	bytes, err := ioutil.ReadFile(path)
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
			pull, _, _, _, _, err := parser.ParseBitbucketServerCommentEvent([]byte(withState))
			Ok(t, err)
			Equals(t, c.exp, pull.State)
		})
	}
}

func TestParseBitbucketServerPullEvent_ValidEvent(t *testing.T) {
	path := filepath.Join("testdata", "bitbucket-server-pull-event-merged.json")
	bytes, err := ioutil.ReadFile(path)
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
		SanitizedCloneURL: "http://mycorp.com:7490/scm/at/atlantis-example.git",
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
		Branch:     "branch",
		Author:     "lkysow",
		State:      models.ClosedPullState,
		BaseRepo:   expBaseRepo,
	}, pull)
	Equals(t, models.Repo{
		FullName:          "atlantis-fork/atlantis-example",
		Owner:             "atlantis-fork",
		Name:              "atlantis-example",
		CloneURL:          "http://bitbucket-user:bitbucket-token@mycorp.com:7490/scm/fk/atlantis-example.git",
		SanitizedCloneURL: "http://mycorp.com:7490/scm/fk/atlantis-example.git",
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
			header: "random",
			exp:    models.OtherPullEvent,
		},
	}
	for _, c := range cases {
		t.Run(c.header, func(t *testing.T) {
			act := parser.GetBitbucketServerEventType(c.header)
			Equals(t, c.exp, act)
		})
	}
}

var mergeEventJSON = `{
  "object_kind": "merge_request",
  "user": {
    "name": "Administrator",
    "username": "root",
    "avatar_url": "http://www.gravatar.com/avatar/e64c7d89f26bd1972efa854d13d7dd61?s=40\u0026d=identicon"
  },
  "project": {
    "id": 1,
    "name":"Gitlab Test",
    "description":"Aut reprehenderit ut est.",
    "web_url":"http://example.com/gitlabhq/gitlab-test",
    "avatar_url":null,
    "git_ssh_url":"git@example.com:gitlabhq/gitlab-test.git",
    "git_http_url":"https://example.com/gitlabhq/gitlab-test.git",
    "namespace":"GitlabHQ",
    "visibility_level":20,
    "path_with_namespace":"gitlabhq/gitlab-test",
    "default_branch":"master",
    "homepage":"http://example.com/gitlabhq/gitlab-test",
    "url":"https://example.com/gitlabhq/gitlab-test.git",
    "ssh_url":"git@example.com:gitlabhq/gitlab-test.git",
    "http_url":"https://example.com/gitlabhq/gitlab-test.git"
  },
  "repository": {
    "name": "Gitlab Test",
    "url": "https://example.com/gitlabhq/gitlab-test.git",
    "description": "Aut reprehenderit ut est.",
    "homepage": "http://example.com/gitlabhq/gitlab-test"
  },
  "object_attributes": {
    "id": 99,
    "target_branch": "master",
    "source_branch": "ms-viewport",
    "source_project_id": 14,
    "author_id": 51,
    "assignee_id": 6,
    "title": "MS-Viewport",
    "created_at": "2013-12-03T17:23:34Z",
    "updated_at": "2013-12-03T17:23:34Z",
    "st_commits": null,
    "st_diffs": null,
    "milestone_id": null,
    "state": "opened",
    "merge_status": "unchecked",
    "target_project_id": 14,
    "iid": 1,
    "description": "",
    "source": {
      "name":"Awesome Project",
      "description":"Aut reprehenderit ut est.",
      "web_url":"http://example.com/awesome_space/awesome_project",
      "avatar_url":null,
      "git_ssh_url":"git@example.com:awesome_space/awesome_project.git",
      "git_http_url":"http://example.com/awesome_space/awesome_project.git",
      "namespace":"Awesome Space",
      "visibility_level":20,
      "path_with_namespace":"awesome_space/awesome_project",
      "default_branch":"master",
      "homepage":"http://example.com/awesome_space/awesome_project",
      "url":"http://example.com/awesome_space/awesome_project.git",
      "ssh_url":"git@example.com:awesome_space/awesome_project.git",
      "http_url":"http://example.com/awesome_space/awesome_project.git"
    },
    "target": {
      "name":"Awesome Project",
      "description":"Aut reprehenderit ut est.",
      "web_url":"http://example.com/awesome_space/awesome_project",
      "avatar_url":null,
      "git_ssh_url":"git@example.com:awesome_space/awesome_project.git",
      "git_http_url":"http://example.com/awesome_space/awesome_project.git",
      "namespace":"Awesome Space",
      "visibility_level":20,
      "path_with_namespace":"awesome_space/awesome_project",
      "default_branch":"master",
      "homepage":"http://example.com/awesome_space/awesome_project",
      "url":"http://example.com/awesome_space/awesome_project.git",
      "ssh_url":"git@example.com:awesome_space/awesome_project.git",
      "http_url":"http://example.com/awesome_space/awesome_project.git"
    },
    "last_commit": {
      "id": "da1560886d4f094c3e6c9ef40349f7d38b5d27d7",
      "message": "fixed readme",
      "timestamp": "2012-01-03T23:36:29+02:00",
      "url": "http://example.com/awesome_space/awesome_project/commits/da1560886d4f094c3e6c9ef40349f7d38b5d27d7",
      "author": {
        "name": "GitLab dev user",
        "email": "gitlabdev@dv6700.(none)"
      }
    },
    "work_in_progress": false,
    "url": "http://example.com/diaspora/merge_requests/1",
    "action": "open",
    "assignee": {
      "name": "User1",
      "username": "user1",
      "avatar_url": "http://www.gravatar.com/avatar/e64c7d89f26bd1972efa854d13d7dd61?s=40\u0026d=identicon"
    }
  },
  "labels": [{
    "id": 206,
    "title": "API",
    "color": "#ffffff",
    "project_id": 14,
    "created_at": "2013-12-03T17:15:43Z",
    "updated_at": "2013-12-03T17:15:43Z",
    "template": false,
    "description": "API related issues",
    "type": "ProjectLabel",
    "group_id": 41
  }],
  "changes": {
    "updated_by_id": [null, 1],
    "updated_at": ["2017-09-15 16:50:55 UTC", "2017-09-15 16:52:00 UTC"],
    "labels": {
      "previous": [{
        "id": 206,
        "title": "API",
        "color": "#ffffff",
        "project_id": 14,
        "created_at": "2013-12-03T17:15:43Z",
        "updated_at": "2013-12-03T17:15:43Z",
        "template": false,
        "description": "API related issues",
        "type": "ProjectLabel",
        "group_id": 41
      }],
      "current": [{
        "id": 205,
        "title": "Platform",
        "color": "#123123",
        "project_id": 14,
        "created_at": "2013-12-03T17:15:43Z",
        "updated_at": "2013-12-03T17:15:43Z",
        "template": false,
        "description": "Platform related issues",
        "type": "ProjectLabel",
        "group_id": 41
      }]
    }
  }
}`

var mergeCommentEventJSON = `{
  "object_kind": "note",
  "user": {
    "name": "Administrator",
    "username": "root",
    "avatar_url": "http://www.gravatar.com/avatar/e64c7d89f26bd1972efa854d13d7dd61?s=40\u0026d=identicon"
  },
  "project_id": 5,
  "project":{
    "id": 5,
    "name":"Gitlab Test",
    "description":"Aut reprehenderit ut est.",
    "web_url":"http://example.com/gitlabhq/gitlab-test",
    "avatar_url":null,
    "git_ssh_url":"git@example.com:gitlabhq/gitlab-test.git",
    "git_http_url":"https://example.com/gitlabhq/gitlab-test.git",
    "namespace":"Gitlab Org",
    "visibility_level":10,
    "path_with_namespace":"gitlabhq/gitlab-test",
    "default_branch":"master",
    "homepage":"http://example.com/gitlabhq/gitlab-test",
    "url":"https://example.com/gitlabhq/gitlab-test.git",
    "ssh_url":"git@example.com:gitlabhq/gitlab-test.git",
    "http_url":"https://example.com/gitlabhq/gitlab-test.git"
  },
  "repository":{
    "name": "Gitlab Test",
    "url": "http://localhost/gitlab-org/gitlab-test.git",
    "description": "Aut reprehenderit ut est.",
    "homepage": "http://example.com/gitlab-org/gitlab-test"
  },
  "object_attributes": {
    "id": 1244,
    "note": "This MR needs work.",
    "noteable_type": "MergeRequest",
    "author_id": 1,
    "created_at": "2015-05-17",
    "updated_at": "2015-05-17",
    "project_id": 5,
    "attachment": null,
    "line_code": null,
    "commit_id": "",
    "noteable_id": 7,
    "system": false,
    "st_diff": null,
    "url": "http://example.com/gitlab-org/gitlab-test/merge_requests/1#note_1244"
  },
  "merge_request": {
    "id": 7,
    "target_branch": "markdown",
    "source_branch": "master",
    "source_project_id": 5,
    "author_id": 8,
    "assignee_id": 28,
    "title": "Tempora et eos debitis quae laborum et.",
    "created_at": "2015-03-01 20:12:53 UTC",
    "updated_at": "2015-03-21 18:27:27 UTC",
    "milestone_id": 11,
    "state": "opened",
    "merge_status": "cannot_be_merged",
    "target_project_id": 5,
    "iid": 1,
    "description": "Et voluptas corrupti assumenda temporibus. Architecto cum animi eveniet amet asperiores. Vitae numquam voluptate est natus sit et ad id.",
    "position": 0,
    "source":{
      "name":"Gitlab Test",
      "description":"Aut reprehenderit ut est.",
      "web_url":"http://example.com/gitlab-org/gitlab-test",
      "avatar_url":null,
      "git_ssh_url":"git@example.com:gitlab-org/gitlab-test.git",
      "git_http_url":"https://example.com/gitlab-org/gitlab-test.git",
      "namespace":"Gitlab Org",
      "visibility_level":10,
      "path_with_namespace":"gitlab-org/gitlab-test",
      "default_branch":"master",
      "homepage":"http://example.com/gitlab-org/gitlab-test",
      "url":"https://example.com/gitlab-org/gitlab-test.git",
      "ssh_url":"git@example.com:gitlab-org/gitlab-test.git",
      "http_url":"https://example.com/gitlab-org/gitlab-test.git",
      "git_http_url":"https://example.com/gitlab-org/gitlab-test.git"
    },
    "target": {
      "name":"Gitlab Test",
      "description":"Aut reprehenderit ut est.",
      "web_url":"http://example.com/gitlabhq/gitlab-test",
      "avatar_url":null,
      "git_ssh_url":"git@example.com:gitlabhq/gitlab-test.git",
      "git_http_url":"https://example.com/gitlabhq/gitlab-test.git",
      "namespace":"Gitlab Org",
      "visibility_level":10,
      "path_with_namespace":"gitlabhq/gitlab-test",
      "default_branch":"master",
      "homepage":"http://example.com/gitlabhq/gitlab-test",
      "url":"https://example.com/gitlabhq/gitlab-test.git",
      "ssh_url":"git@example.com:gitlabhq/gitlab-test.git",
      "http_url":"https://example.com/gitlabhq/gitlab-test.git"
    },
    "last_commit": {
      "id": "562e173be03b8ff2efb05345d12df18815438a4b",
      "message": "Merge branch 'another-branch' into 'master'\n\nCheck in this test\n",
      "timestamp": "2002-10-02T10:00:00-05:00",
      "url": "http://example.com/gitlab-org/gitlab-test/commit/562e173be03b8ff2efb05345d12df18815438a4b",
      "author": {
        "name": "John Smith",
        "email": "john@example.com"
      }
    },
    "work_in_progress": false,
    "assignee": {
      "name": "User1",
      "username": "user1",
      "avatar_url": "http://www.gravatar.com/avatar/e64c7d89f26bd1972efa854d13d7dd61?s=40\u0026d=identicon"
    }
  }
}`

var mergeRequestJSON = `{
  "id":6056811,
  "iid":8,
  "project_id":4580910,
  "title":"Update main.tf",
  "description":"",
  "state":"opened",
  "created_at":"2017-11-13T19:33:42.704Z",
  "updated_at":"2017-11-13T23:35:26.200Z",
  "target_branch":"master",
  "source_branch":"abc",
  "upvotes":0,
  "downvotes":0,
  "author":{
	 "id":1755902,
	 "name":"Luke Kysow",
	 "username":"lkysow",
	 "state":"active",
	 "avatar_url":"https://secure.gravatar.com/avatar/25fd57e71590fe28736624ff24d41c5f?s=80\u0026d=identicon",
	 "web_url":"https://gitlab.com/lkysow"
  },
  "assignee":null,
  "source_project_id":4580910,
  "target_project_id":4580910,
  "labels":[

  ],
  "work_in_progress":false,
  "milestone":null,
  "merge_when_pipeline_succeeds":false,
  "merge_status":"can_be_merged",
  "sha":"0b4ac85ea3063ad5f2974d10cd68dd1f937aaac2",
  "merge_commit_sha":null,
  "user_notes_count":10,
  "approvals_before_merge":null,
  "discussion_locked":null,
  "should_remove_source_branch":null,
  "force_remove_source_branch":false,
  "squash":false,
  "web_url":"https://gitlab.com/lkysow/atlantis-example/merge_requests/8",
  "time_stats":{
	 "time_estimate":0,
	 "total_time_spent":0,
	 "human_time_estimate":null,
	 "human_total_time_spent":null
  }
}`
