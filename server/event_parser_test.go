package server_test

import (
	"testing"

	"github.com/google/go-github/github"
	"github.com/hootsuite/atlantis/server"
	. "github.com/hootsuite/atlantis/testing_util"
	"strings"
	"errors"
	"github.com/hootsuite/atlantis/models"
	"github.com/mohae/deepcopy"
)

var ghRepo = github.Repository{
	FullName: github.String("owner/repo"),
	Owner: &github.User{Login: github.String("owner")},
	Name: github.String("repo"),
	CloneURL: github.String("https://github.com/lkysow/atlantis-example.git"),
}
var parser = server.EventParser{"user", "token"}

func TestDetermineCommandNoBody(t *testing.T) {
	_, err := parser.DetermineCommand(&github.IssueCommentEvent{})
	Equals(t, errors.New("comment.body is null"), err)
}

func TestDetermineCommandInvalid(t *testing.T) {
	t.Log("given a comment that does not match the regex should return an error")
	comments := []string{
		// just the executable, no command
		"run",
		"atlantis",
		"@user",
		// invalid command
		"run slkjd",
		"atlantis slkjd",
		"@user slkjd",
		"atlantis plans",
		// misc
		"related comment mentioning atlantis",
	}
	for _, c := range comments {
		_, e := parser.DetermineCommand(buildComment(c))
		Assert(t, e != nil, "expected error for comment: "+c)
	}
}

func TestDetermineCommandHelp(t *testing.T) {
	t.Log("given a help comment, should match")
	comments := []string{
		"run help",
		"atlantis help",
		"@user help",
		"atlantis help --verbose",
	}
	for _, c := range comments {
		command, e := parser.DetermineCommand(buildComment(c))
		Ok(t, e)
		Equals(t, server.Help, command.Name)
	}
}

func TestDetermineCommandPermutations(t *testing.T) {
	execNames := []string{"run", "atlantis", "@user"}
	commandNames := []server.CommandName{server.Plan, server.Apply}
	envs := []string{"", "default", "env", "env-dash", "env_underscore", "camelEnv"}
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
			for _, env := range envs {
				for _, flags := range flagCases {
					// If github comments end in a newline they get \r\n appended.
					// Ensure that we parse commands properly either way.
					for _, lineEnding := range []string{"", "\r\n"} {
						comment := strings.Join(append([]string{exec, name.String(), env}, flags...), " ") + lineEnding
						t.Log("testing comment: " + comment)
						c, err := parser.DetermineCommand(buildComment(comment))
						Ok(t, err)
						Equals(t, name, c.Name)
						if env == "" {
							Equals(t, "default", c.Environment)
						} else {
							Equals(t, env, c.Environment)
						}
						Equals(t, stringInSlice("--verbose", flags), c.Verbose)

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

func TestExtractRepoData(t *testing.T) {
	testRepo := ghRepo
	testRepo.FullName = nil
	_, err := parser.ExtractRepoData(&testRepo)
	Equals(t, errors.New("repository.full_name is null"), err)

	testRepo = ghRepo
	testRepo.Owner = nil
	_, err = parser.ExtractRepoData(&testRepo)
	Equals(t, errors.New("repository.owner.login is null"), err)

	testRepo = ghRepo
	testRepo.Name = nil
	_, err = parser.ExtractRepoData(&testRepo)
	Equals(t, errors.New("repository.name is null"), err)

	testRepo = ghRepo
	testRepo.CloneURL = nil
	_, err = parser.ExtractRepoData(&testRepo)
	Equals(t, errors.New("repository.clone_url is null"), err)

	t.Log("should replace https clone with user/pass")
	{
		r, err := parser.ExtractRepoData(&ghRepo)
		Ok(t, err)
		Equals(t, models.Repo{
			Owner: "owner",
			FullName: "owner/repo",
			CloneURL: "https://user:token@github.com/lkysow/atlantis-example.git",
			SanitizedCloneURL: ghRepo.GetCloneURL(),
			Name: "repo",
		}, r)
	}
}

func TestExtractCommentData(t *testing.T) {
	comment := github.IssueCommentEvent{
		Repo: &ghRepo,
		Issue: &github.Issue{
			Number: github.Int(1),
			User: &github.User{Login: github.String("issue_user")},
			HTMLURL: github.String("https://github.com/hootsuite/atlantis/issues/1"),
		},
		Comment: &github.IssueComment{
			User: &github.User{Login: github.String("comment_user")},
		},
	}
	ctx := server.CommandContext{}

	testComment := deepcopy.Copy(comment).(github.IssueCommentEvent)
	testComment.Repo = nil
	err := parser.ExtractCommentData(&testComment, &ctx)
	Equals(t, errors.New("repository.full_name is null"), err)

	testComment = deepcopy.Copy(comment).(github.IssueCommentEvent)
	testComment.Issue = nil
	err = parser.ExtractCommentData(&testComment, &ctx)
	Equals(t, errors.New("issue.number is null"), err)

	testComment = deepcopy.Copy(comment).(github.IssueCommentEvent)
	testComment.Issue.User = nil
	err = parser.ExtractCommentData(&testComment, &ctx)
	Equals(t, errors.New("issue.user.login is null"), err)

	testComment = deepcopy.Copy(comment).(github.IssueCommentEvent)
	testComment.Issue.HTMLURL = nil
	err = parser.ExtractCommentData(&testComment, &ctx)
	Equals(t, errors.New("issue.html_url is null"), err)

	testComment = deepcopy.Copy(comment).(github.IssueCommentEvent)
	testComment.Comment.User.Login = nil
	err = parser.ExtractCommentData(&testComment, &ctx)
	Equals(t, errors.New("comment.user.login is null"), err)

	// this should be successful
	err = parser.ExtractCommentData(&comment, &ctx)
	Ok(t, err)
	Equals(t, models.Repo{
		Owner: *comment.Repo.Owner.Login,
		FullName: *comment.Repo.FullName,
		CloneURL: "https://user:token@github.com/lkysow/atlantis-example.git",
		SanitizedCloneURL: *comment.Repo.CloneURL,
		Name: "repo",
	}, ctx.BaseRepo)
	Equals(t, models.User{
		Username: *comment.Comment.User.Login,
	}, ctx.User)
	Equals(t, models.PullRequest{
		Num: *comment.Issue.Number,
	}, ctx.Pull)
}

func TestExtractPullData(t *testing.T) {
	pull := github.PullRequest{
		Head: &github.PullRequestBranch{
			SHA: github.String("sha256"),
			Ref: github.String("ref"),
			Repo: &ghRepo,
		},
		Base: &github.PullRequestBranch{
			SHA: github.String("sha256"),
		},
		HTMLURL: github.String("html-url"),
		User: &github.User{
			Login: github.String("user"),
		},
		Number: github.Int(1),
	}

	testPull := deepcopy.Copy(pull).(github.PullRequest)
	testPull.Head.SHA = nil
	_, _, err := parser.ExtractPullData(&testPull)
	Equals(t, errors.New("head.sha is null"), err)

	testPull = deepcopy.Copy(pull).(github.PullRequest)
	testPull.Base.SHA = nil
	_, _, err = parser.ExtractPullData(&testPull)
	Equals(t, errors.New("base.sha is null"), err)

	testPull = deepcopy.Copy(pull).(github.PullRequest)
	testPull.HTMLURL = nil
	_, _, err = parser.ExtractPullData(&testPull)
	Equals(t, errors.New("html_url is null"), err)

	testPull = deepcopy.Copy(pull).(github.PullRequest)
	testPull.Head.Ref = nil
	_, _, err = parser.ExtractPullData(&testPull)
	Equals(t, errors.New("head.ref is null"), err)

	testPull = deepcopy.Copy(pull).(github.PullRequest)
	testPull.User.Login = nil
	_, _, err = parser.ExtractPullData(&testPull)
	Equals(t, errors.New("user.login is null"), err)

	testPull = deepcopy.Copy(pull).(github.PullRequest)
	testPull.Number = nil
	_, _, err = parser.ExtractPullData(&testPull)
	Equals(t, errors.New("number is null"), err)

	testPull = deepcopy.Copy(pull).(github.PullRequest)
	testPull.Head.Repo = nil
	_, _, err = parser.ExtractPullData(&testPull)
	Equals(t, errors.New("repository.full_name is null"), err)

	pullRes, repoRes, err := parser.ExtractPullData(&pull)
	Equals(t, models.PullRequest{
		BaseCommit: pull.Base.GetSHA(),
		URL: pull.GetHTMLURL(),
		Author: pull.User.GetLogin(),
		Branch: pull.Head.GetRef(),
		HeadCommit: pull.Head.GetSHA(),
		Num: pull.GetNumber(),
	}, pullRes)

	Equals(t,models.Repo{
		Owner: "owner",
		FullName: "owner/repo",
		CloneURL: "https://user:token@github.com/lkysow/atlantis-example.git",
		SanitizedCloneURL: ghRepo.GetCloneURL(),
		Name: "repo",
	}, repoRes)
}

func buildComment(c string) *github.IssueCommentEvent {
	return &github.IssueCommentEvent{
		Comment: &github.IssueComment{
			Body: github.String(c),
		},
	}
}

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

