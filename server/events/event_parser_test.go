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
	GithubUser:  "github-user",
	GithubToken: "github-token",
	GitlabUser:  "gitlab-user",
	GitlabToken: "gitlab-token",
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

func TestParseGithubPull(t *testing.T) {
	testPull := deepcopy.Copy(Pull).(github.PullRequest)
	testPull.Head.SHA = nil
	_, _, err := parser.ParseGithubPull(&testPull)
	ErrEquals(t, "head.sha is null", err)

	testPull = deepcopy.Copy(Pull).(github.PullRequest)
	testPull.HTMLURL = nil
	_, _, err = parser.ParseGithubPull(&testPull)
	ErrEquals(t, "html_url is null", err)

	testPull = deepcopy.Copy(Pull).(github.PullRequest)
	testPull.Head.Ref = nil
	_, _, err = parser.ParseGithubPull(&testPull)
	ErrEquals(t, "head.ref is null", err)

	testPull = deepcopy.Copy(Pull).(github.PullRequest)
	testPull.User.Login = nil
	_, _, err = parser.ParseGithubPull(&testPull)
	ErrEquals(t, "user.login is null", err)

	testPull = deepcopy.Copy(Pull).(github.PullRequest)
	testPull.Number = nil
	_, _, err = parser.ParseGithubPull(&testPull)
	ErrEquals(t, "number is null", err)

	pullRes, _, err := parser.ParseGithubPull(&Pull)
	Ok(t, err)
	Equals(t, models.PullRequest{
		URL:        Pull.GetHTMLURL(),
		Author:     Pull.User.GetLogin(),
		Branch:     Pull.Head.GetRef(),
		HeadCommit: Pull.Head.GetSHA(),
		Num:        Pull.GetNumber(),
		State:      models.Open,
		BaseRepo: models.Repo{
			Owner:             "owner",
			FullName:          "owner/repo",
			CloneURL:          "https://github-user:github-token@github.com/owner/repo.git",
			SanitizedCloneURL: Repo.GetCloneURL(),
			Name:              "repo",
			VCSHost: models.VCSHost{
				Hostname: "github.com",
				Type:     models.Github,
			},
		},
	}, pullRes)
}

func TestParseGitlabMergeEvent(t *testing.T) {
	t.Log("should properly parse a gitlab merge event")
	var event *gitlab.MergeEvent
	err := json.Unmarshal([]byte(mergeEventJSON), &event)
	Ok(t, err)
	pull, repo, err := parser.ParseGitlabMergeEvent(*event)
	Ok(t, err)

	expRepo := models.Repo{
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
		State:      models.Open,
		BaseRepo:   expRepo,
	}, pull)

	Equals(t, expRepo, repo)

	t.Log("If the state is closed, should set field correctly.")
	event.ObjectAttributes.State = "closed"
	pull, _, err = parser.ParseGitlabMergeEvent(*event)
	Ok(t, err)
	Equals(t, models.Closed, pull.State)
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
		State:      models.Open,
		BaseRepo:   repo,
	}, pull)

	t.Log("If the state is closed, should set field correctly.")
	event.State = "closed"
	pull = parser.ParseGitlabMergeRequest(event, repo)
	Equals(t, models.Closed, pull.State)
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
		Dir    string
		ExpDir string
	}{
		{
			"",
			".",
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
		t.Run(c.Dir, func(t *testing.T) {
			cmd := events.NewCommand(c.Dir, nil, events.Plan, false, "workspace")
			Equals(t, c.ExpDir, cmd.Dir)
		})
	}
}

func TestNewCommand_EmptyWorkspace(t *testing.T) {
	cmd := events.NewCommand("dir", nil, events.Plan, false, "")
	Equals(t, "default", cmd.Workspace)
}

func TestNewCommand_AllFieldsSet(t *testing.T) {
	cmd := events.NewCommand("dir", []string{"a", "b"}, events.Plan, true, "workspace")
	Equals(t, events.Command{
		Workspace: "workspace",
		Dir:       "dir",
		Verbose:   true,
		Flags:     []string{"a", "b"},
		Name:      events.Plan,
	}, *cmd)
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
