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
	testPull.Base.Ref = nil
	_, _, _, err = parser.ParseGithubPull(&testPull)
	ErrEquals(t, "base.ref is null", err)

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
	bytes, err := ioutil.ReadFile(path)
	Ok(t, err)
	var event *gitlab.MergeEvent
	err = json.Unmarshal(bytes, &event)
	Ok(t, err)
	pull, evType, actBaseRepo, actHeadRepo, actUser, err := parser.ParseGitlabMergeRequestEvent(*event)
	Ok(t, err)

	expBaseRepo := models.Repo{
		FullName:          "lkysow/atlantis-example",
		Name:              "atlantis-example",
		SanitizedCloneURL: "https://gitlab.com/lkysow/atlantis-example.git",
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
		BaseBranch: "master",
		State:      models.OpenPullState,
		BaseRepo:   expBaseRepo,
	}, pull)
	Equals(t, models.OpenedPullEvent, evType)

	Equals(t, expBaseRepo, actBaseRepo)
	Equals(t, models.Repo{
		FullName:          "sourceorg/atlantis-example",
		Name:              "atlantis-example",
		SanitizedCloneURL: "https://gitlab.com/sourceorg/atlantis-example.git",
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

// Should be able to parse a merge event from a repo that is in a subgroup,
// i.e. instead of under an owner/repo it's under an owner/group/subgroup/repo.
func TestParseGitlabMergeEvent_Subgroup(t *testing.T) {
	path := filepath.Join("testdata", "gitlab-merge-request-event-subgroup.json")
	bytes, err := ioutil.ReadFile(path)
	Ok(t, err)
	var event *gitlab.MergeEvent
	err = json.Unmarshal(bytes, &event)
	Ok(t, err)
	pull, evType, actBaseRepo, actHeadRepo, actUser, err := parser.ParseGitlabMergeRequestEvent(*event)
	Ok(t, err)

	expBaseRepo := models.Repo{
		FullName:          "lkysow-test/subgroup/sub-subgroup/atlantis-example",
		Name:              "atlantis-example",
		SanitizedCloneURL: "https://gitlab.com/lkysow-test/subgroup/sub-subgroup/atlantis-example.git",
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
		BaseBranch: "master",
		State:      models.OpenPullState,
		BaseRepo:   expBaseRepo,
	}, pull)
	Equals(t, models.OpenedPullEvent, evType)

	Equals(t, expBaseRepo, actBaseRepo)
	Equals(t, models.Repo{
		FullName:          "lkysow-test/subgroup/sub-subgroup/atlantis-example",
		Name:              "atlantis-example",
		SanitizedCloneURL: "https://gitlab.com/lkysow-test/subgroup/sub-subgroup/atlantis-example.git",
		Owner:             "lkysow-test/subgroup/sub-subgroup",
		CloneURL:          "https://gitlab-user:gitlab-token@gitlab.com/lkysow-test/subgroup/sub-subgroup/atlantis-example.git",
		VCSHost: models.VCSHost{
			Hostname: "gitlab.com",
			Type:     models.Gitlab,
		},
	}, actHeadRepo)
	Equals(t, models.User{Username: "lkysow"}, actUser)
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

	path := filepath.Join("testdata", "gitlab-merge-request-event.json")
	bytes, err := ioutil.ReadFile(path)
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
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		Ok(t, err)
	}
	var event *gitlab.MergeRequest
	err = json.Unmarshal(bytes, &event)
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
		HeadBranch: "abc",
		BaseBranch: "master",
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
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		Ok(t, err)
	}
	var event *gitlab.MergeRequest
	err = json.Unmarshal(bytes, &event)
	Ok(t, err)

	repo := models.Repo{
		FullName:          "lkysow-test/subgroup/sub-subgroup/atlantis-example",
		Name:              "atlantis-example",
		SanitizedCloneURL: "https://gitlab.com/lkysow-test/subgroup/sub-subgroup/atlantis-example.git",
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
		BaseBranch: "master",
		State:      models.OpenPullState,
		BaseRepo:   repo,
	}, pull)
}

func TestParseGitlabMergeCommentEvent(t *testing.T) {
	t.Log("should properly parse a gitlab merge comment event")
	path := filepath.Join("testdata", "gitlab-merge-request-comment-event.json")
	bytes, err := ioutil.ReadFile(path)
	Ok(t, err)
	var event *gitlab.MergeCommentEvent
	err = json.Unmarshal(bytes, &event)
	Ok(t, err)
	baseRepo, headRepo, user, err := parser.ParseGitlabMergeRequestCommentEvent(*event)
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

// Should properly parse a gitlab merge comment event from a subgroup repo.
func TestParseGitlabMergeCommentEvent_Subgroup(t *testing.T) {
	path := filepath.Join("testdata", "gitlab-merge-request-comment-event-subgroup.json")
	bytes, err := ioutil.ReadFile(path)
	Ok(t, err)
	var event *gitlab.MergeCommentEvent
	err = json.Unmarshal(bytes, &event)
	Ok(t, err)
	baseRepo, headRepo, user, err := parser.ParseGitlabMergeRequestCommentEvent(*event)
	Ok(t, err)

	Equals(t, models.Repo{
		FullName:          "lkysow-test/subgroup/sub-subgroup/atlantis-example",
		Name:              "atlantis-example",
		SanitizedCloneURL: "https://gitlab.com/lkysow-test/subgroup/sub-subgroup/atlantis-example.git",
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
		SanitizedCloneURL: "https://gitlab.com/lkysow-test/subgroup/sub-subgroup/atlantis-example.git",
		Owner:             "lkysow-test/subgroup/sub-subgroup",
		CloneURL:          "https://gitlab-user:gitlab-token@gitlab.com/lkysow-test/subgroup/sub-subgroup/atlantis-example.git",
		VCSHost: models.VCSHost{
			Hostname: "gitlab.com",
			Type:     models.Gitlab,
		},
	}, headRepo)
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
	_, _, _, _, _, err := parser.ParseBitbucketCloudPullCommentEvent([]byte(""))
	ErrEquals(t, "parsing json: unexpected end of JSON input", err)
}

func TestParseBitbucketCloudCommentEvent_EmptyObject(t *testing.T) {
	_, _, _, _, _, err := parser.ParseBitbucketCloudPullCommentEvent([]byte("{}"))
	ErrContains(t, "Key: 'CommentEvent.CommonEventData.Actor' Error:Field validation for 'Actor' failed on the 'required' tag\nKey: 'CommentEvent.CommonEventData.Repository' Error:Field validation for 'Repository' failed on the 'required' tag\nKey: 'CommentEvent.CommonEventData.PullRequest' Error:Field validation for 'PullRequest' failed on the 'required' tag\nKey: 'CommentEvent.Comment' Error:Field validation for 'Comment' failed on the 'required' tag", err)
}

func TestParseBitbucketCloudCommentEvent_CommitHashMissing(t *testing.T) {
	path := filepath.Join("testdata", "bitbucket-cloud-comment-event.json")
	bytes, err := ioutil.ReadFile(path)
	Ok(t, err)
	emptyCommitHash := strings.Replace(string(bytes), `        "hash": "e0624da46d3a",`, "", -1)
	_, _, _, _, _, err = parser.ParseBitbucketCloudPullCommentEvent([]byte(emptyCommitHash))
	ErrContains(t, "Key: 'CommentEvent.CommonEventData.PullRequest.Source.Commit.Hash' Error:Field validation for 'Hash' failed on the 'required' tag", err)
}

func TestParseBitbucketCloudCommentEvent_ValidEvent(t *testing.T) {
	path := filepath.Join("testdata", "bitbucket-cloud-comment-event.json")
	bytes, err := ioutil.ReadFile(path)
	Ok(t, err)
	pull, baseRepo, headRepo, user, comment, err := parser.ParseBitbucketCloudPullCommentEvent(bytes)
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
		HeadBranch: "lkysow/maintf-edited-online-with-bitbucket-1532029690581",
		BaseBranch: "master",
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
			pull, _, _, _, _, err := parser.ParseBitbucketCloudPullCommentEvent([]byte(withState))
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
		HeadBranch: "lkysow/maintf-edited-online-with-bitbucket-1532029690581",
		BaseBranch: "master",
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
			act := parser.GetBitbucketCloudPullEventType(c.header)
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
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		Ok(t, err)
	}
	emptyCommitHash := strings.Replace(string(bytes), `"latestCommit": "bfb1af1ba9c2a2fa84cd61af67e6e1b60a22e060",`, "", -1)
	_, _, _, _, _, err = parser.ParseBitbucketServerPullCommentEvent([]byte(emptyCommitHash))
	ErrContains(t, "Key: 'CommentEvent.CommonEventData.PullRequest.FromRef.LatestCommit' Error:Field validation for 'LatestCommit' failed on the 'required' tag", err)
}

func TestParseBitbucketServerCommentEvent_ValidEvent(t *testing.T) {
	path := filepath.Join("testdata", "bitbucket-server-comment-event.json")
	bytes, err := ioutil.ReadFile(path)
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
		HeadBranch: "branch",
		BaseBranch: "master",
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
			pull, _, _, _, _, err := parser.ParseBitbucketServerPullCommentEvent([]byte(withState))
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
		HeadBranch: "branch",
		BaseBranch: "master",
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
			act := parser.GetBitbucketServerPullEventType(c.header)
			Equals(t, c.exp, act)
		})
	}
}
