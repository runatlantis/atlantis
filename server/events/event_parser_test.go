package events_test

import (
	"testing"

	"errors"
	"strings"

	"encoding/json"

	"github.com/google/go-github/github"
	"github.com/lkysow/go-gitlab"
	"github.com/mohae/deepcopy"
	"github.com/runatlantis/atlantis/server/events"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/events/vcs"
	. "github.com/runatlantis/atlantis/server/events/vcs/fixtures"
	. "github.com/runatlantis/atlantis/testing"
)

var parser = events.EventParser{
	GithubUser:  "github-user",
	GithubToken: "github-token",
	GitlabUser:  "gitlab-user",
	GitlabToken: "gitlab-token",
}

func TestDetermineCommandInvalid(t *testing.T) {
	t.Log("given a comment that does not match the regex should return an error")
	comments := []string{
		// just the executable, no command
		"run",
		"atlantis",
		"@github-user",
		// invalid command
		"run slkjd",
		"atlantis slkjd",
		"@github-user slkjd",
		"atlantis plans",
		// misc
		"related comment mentioning atlantis",
	}
	for _, c := range comments {
		_, e := parser.DetermineCommand(c, vcs.Github)
		Assert(t, e != nil, "expected error for comment: "+c)
	}
}

func TestDetermineCommandHelp(t *testing.T) {
	t.Log("given a help comment, should match")
	comments := []string{
		"run help",
		"atlantis help",
		"@github-user help",
		"atlantis help --verbose",
	}
	for _, c := range comments {
		command, e := parser.DetermineCommand(c, vcs.Github)
		Ok(t, e)
		Equals(t, events.Help, command.Name)
	}
}

// nolint: gocyclo
func TestDetermineCommandPermutations(t *testing.T) {
	execNames := []string{"run", "atlantis", "@github-user", "@gitlab-user"}
	commandNames := []events.CommandName{events.Plan, events.Apply}
	workspaces := []string{"", "default", "workspace", "workspace-dash", "workspace_underscore", "camelWorkspace"}
	flagCases := [][]string{
		{},
		{"--verbose"},
		{"-key=value"},
		{"-key", "value"},
		{"-key1=value1", "-key2=value2"},
		{"-key1=value1", "-key2", "value2"},
		{"-key1", "value1", "-key2=value2"},
		{"--verbose", "key2=value2"},
		{"-key1=value1", "--verbose"},
	}

	// test all permutations
	for _, exec := range execNames {
		for _, name := range commandNames {
			for _, workspace := range workspaces {
				for _, flags := range flagCases {
					// If github comments end in a newline they get \r\n appended.
					// Ensure that we parse commands properly either way.
					for _, lineEnding := range []string{"", "\r\n"} {
						comment := strings.Join(append([]string{exec, name.String(), workspace}, flags...), " ") + lineEnding
						t.Log("testing comment: " + comment)

						// In order to test gitlab without fully refactoring this test
						// we're just detecting if we're using the gitlab user as the
						// exec name.
						vcsHost := vcs.Github
						if exec == "@gitlab-user" {
							vcsHost = vcs.Gitlab
						}
						c, err := parser.DetermineCommand(comment, vcsHost)
						Ok(t, err)
						Equals(t, name, c.Name)
						if workspace == "" {
							Equals(t, "default", c.Workspace)
						} else {
							Equals(t, workspace, c.Workspace)
						}
						Equals(t, containsVerbose(flags), c.Verbose)

						// ensure --verbose never shows up in flags
						for _, f := range c.Flags {
							Assert(t, f != "--verbose", "Should not pass on the --verbose flag: %v", flags)
						}

						// check all flags are present
						for _, f := range flags {
							if f != "--verbose" {
								Contains(t, f, c.Flags)
							}
						}
					}
				}
			}
		}
	}
}

func TestParseGithubRepo(t *testing.T) {
	testRepo := Repo
	testRepo.FullName = nil
	_, err := parser.ParseGithubRepo(&testRepo)
	Equals(t, errors.New("repository.full_name is null"), err)

	testRepo = Repo
	testRepo.Owner = nil
	_, err = parser.ParseGithubRepo(&testRepo)
	Equals(t, errors.New("repository.owner.login is null"), err)

	testRepo = Repo
	testRepo.Name = nil
	_, err = parser.ParseGithubRepo(&testRepo)
	Equals(t, errors.New("repository.name is null"), err)

	testRepo = Repo
	testRepo.CloneURL = nil
	_, err = parser.ParseGithubRepo(&testRepo)
	Equals(t, errors.New("repository.clone_url is null"), err)

	t.Log("should replace https clone with user/pass")
	{
		r, err := parser.ParseGithubRepo(&Repo)
		Ok(t, err)
		Equals(t, models.Repo{
			Owner:             "owner",
			FullName:          "owner/repo",
			CloneURL:          "https://github-user:github-token@github.com/lkysow/atlantis-example.git",
			SanitizedCloneURL: Repo.GetCloneURL(),
			Name:              "repo",
		}, r)
	}
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
	testComment.Repo = nil
	_, _, _, err := parser.ParseGithubIssueCommentEvent(&testComment)
	Equals(t, errors.New("repository.full_name is null"), err)

	testComment = deepcopy.Copy(comment).(github.IssueCommentEvent)
	testComment.Comment = nil
	_, _, _, err = parser.ParseGithubIssueCommentEvent(&testComment)
	Equals(t, errors.New("comment.user.login is null"), err)

	testComment = deepcopy.Copy(comment).(github.IssueCommentEvent)
	testComment.Comment.User = nil
	_, _, _, err = parser.ParseGithubIssueCommentEvent(&testComment)
	Equals(t, errors.New("comment.user.login is null"), err)

	testComment = deepcopy.Copy(comment).(github.IssueCommentEvent)
	testComment.Comment.User.Login = nil
	_, _, _, err = parser.ParseGithubIssueCommentEvent(&testComment)
	Equals(t, errors.New("comment.user.login is null"), err)

	testComment = deepcopy.Copy(comment).(github.IssueCommentEvent)
	testComment.Issue = nil
	_, _, _, err = parser.ParseGithubIssueCommentEvent(&testComment)
	Equals(t, errors.New("issue.number is null"), err)

	// this should be successful
	repo, user, pullNum, err := parser.ParseGithubIssueCommentEvent(&comment)
	Ok(t, err)
	Equals(t, models.Repo{
		Owner:             *comment.Repo.Owner.Login,
		FullName:          *comment.Repo.FullName,
		CloneURL:          "https://github-user:github-token@github.com/lkysow/atlantis-example.git",
		SanitizedCloneURL: *comment.Repo.CloneURL,
		Name:              "repo",
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
	Equals(t, errors.New("head.sha is null"), err)

	testPull = deepcopy.Copy(Pull).(github.PullRequest)
	testPull.HTMLURL = nil
	_, _, err = parser.ParseGithubPull(&testPull)
	Equals(t, errors.New("html_url is null"), err)

	testPull = deepcopy.Copy(Pull).(github.PullRequest)
	testPull.Head.Ref = nil
	_, _, err = parser.ParseGithubPull(&testPull)
	Equals(t, errors.New("head.ref is null"), err)

	testPull = deepcopy.Copy(Pull).(github.PullRequest)
	testPull.User.Login = nil
	_, _, err = parser.ParseGithubPull(&testPull)
	Equals(t, errors.New("user.login is null"), err)

	testPull = deepcopy.Copy(Pull).(github.PullRequest)
	testPull.Number = nil
	_, _, err = parser.ParseGithubPull(&testPull)
	Equals(t, errors.New("number is null"), err)

	testPull = deepcopy.Copy(Pull).(github.PullRequest)
	testPull.Head.Repo = nil
	_, _, err = parser.ParseGithubPull(&testPull)
	Equals(t, errors.New("repository.full_name is null"), err)

	pullRes, repoRes, err := parser.ParseGithubPull(&Pull)
	Ok(t, err)
	Equals(t, models.PullRequest{
		URL:        Pull.GetHTMLURL(),
		Author:     Pull.User.GetLogin(),
		Branch:     Pull.Head.GetRef(),
		HeadCommit: Pull.Head.GetSHA(),
		Num:        Pull.GetNumber(),
		State:      models.Open,
	}, pullRes)

	Equals(t, models.Repo{
		Owner:             "owner",
		FullName:          "owner/repo",
		CloneURL:          "https://github-user:github-token@github.com/lkysow/atlantis-example.git",
		SanitizedCloneURL: Repo.GetCloneURL(),
		Name:              "repo",
	}, repoRes)
}

func TestParseGitlabMergeEvent(t *testing.T) {
	t.Log("should properly parse a gitlab merge event")
	var event *gitlab.MergeEvent
	err := json.Unmarshal([]byte(mergeEventJSON), &event)
	Ok(t, err)
	pull, repo := parser.ParseGitlabMergeEvent(*event)
	Equals(t, models.PullRequest{
		URL:        "http://example.com/diaspora/merge_requests/1",
		Author:     "root",
		Num:        1,
		HeadCommit: "da1560886d4f094c3e6c9ef40349f7d38b5d27d7",
		Branch:     "ms-viewport",
		State:      models.Open,
	}, pull)

	Equals(t, models.Repo{
		FullName:          "gitlabhq/gitlab-test",
		Name:              "gitlab-test",
		SanitizedCloneURL: "https://example.com/gitlabhq/gitlab-test.git",
		Owner:             "gitlabhq",
		CloneURL:          "https://gitlab-user:gitlab-token@example.com/gitlabhq/gitlab-test.git",
	}, repo)

	t.Log("If the state is closed, should set field correctly.")
	event.ObjectAttributes.State = "closed"
	pull, _ = parser.ParseGitlabMergeEvent(*event)
	Equals(t, models.Closed, pull.State)
}

func TestParseGitlabMergeRequest(t *testing.T) {
	t.Log("should properly parse a gitlab merge request")
	var event *gitlab.MergeRequest
	err := json.Unmarshal([]byte(mergeRequestJSON), &event)
	Ok(t, err)
	pull := parser.ParseGitlabMergeRequest(event)
	Equals(t, models.PullRequest{
		URL:        "https://gitlab.com/lkysow/atlantis-example/merge_requests/8",
		Author:     "lkysow",
		Num:        8,
		HeadCommit: "0b4ac85ea3063ad5f2974d10cd68dd1f937aaac2",
		Branch:     "abc",
		State:      models.Open,
	}, pull)

	t.Log("If the state is closed, should set field correctly.")
	event.State = "closed"
	pull = parser.ParseGitlabMergeRequest(event)
	Equals(t, models.Closed, pull.State)
}

func TestParseGitlabMergeCommentEvent(t *testing.T) {
	t.Log("should properly parse a gitlab merge comment event")
	var event *gitlab.MergeCommentEvent
	err := json.Unmarshal([]byte(mergeCommentEventJSON), &event)
	Ok(t, err)
	baseRepo, headRepo, user := parser.ParseGitlabMergeCommentEvent(*event)
	Equals(t, models.Repo{
		FullName:          "gitlabhq/gitlab-test",
		Name:              "gitlab-test",
		SanitizedCloneURL: "https://example.com/gitlabhq/gitlab-test.git",
		Owner:             "gitlabhq",
		CloneURL:          "https://gitlab-user:gitlab-token@example.com/gitlabhq/gitlab-test.git",
	}, baseRepo)
	Equals(t, models.Repo{
		FullName:          "gitlab-org/gitlab-test",
		Name:              "gitlab-test",
		SanitizedCloneURL: "https://example.com/gitlab-org/gitlab-test.git",
		Owner:             "gitlab-org",
		CloneURL:          "https://gitlab-user:gitlab-token@example.com/gitlab-org/gitlab-test.git",
	}, headRepo)
	Equals(t, models.User{
		Username: "root",
	}, user)
}

func containsVerbose(list []string) bool {
	for _, b := range list {
		if b == "--verbose" {
			return true
		}
	}
	return false
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
