package converter_test

import (
	"testing"
	"time"

	"github.com/google/go-github/v45/github"
	"github.com/mohae/deepcopy"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/vcs/provider/github/converter"
	"github.com/stretchr/testify/assert"
)

var timestamp = time.Now()

var Pull = github.PullRequest{
	Head: &github.PullRequestBranch{
		SHA:  github.String("sha256"),
		Ref:  github.String("ref"),
		Repo: &Repo,
	},
	Base: &github.PullRequestBranch{
		SHA:  github.String("sha256"),
		Repo: &Repo,
		Ref:  github.String("basebranch"),
	},
	HTMLURL: github.String("html-url"),
	User: &github.User{
		Login: github.String("user"),
	},
	Number:    github.Int(1),
	State:     github.String("open"),
	UpdatedAt: &timestamp,
}

func TestParseGithubPull(t *testing.T) {
	repoConverter := converter.RepoConverter{
		GithubUser:  "github-user",
		GithubToken: "github-token",
	}
	pullConverter := converter.PullConverter{RepoConverter: repoConverter}
	testPull := deepcopy.Copy(Pull).(github.PullRequest)
	testPull.Head.SHA = nil
	_, err := pullConverter.Convert(&testPull)
	assert.EqualError(t, err, "head.sha is null")

	testPull = deepcopy.Copy(Pull).(github.PullRequest)
	testPull.HTMLURL = nil
	_, err = pullConverter.Convert(&testPull)
	assert.EqualError(t, err, "html_url is null")

	testPull = deepcopy.Copy(Pull).(github.PullRequest)
	testPull.Head.Ref = nil
	_, err = pullConverter.Convert(&testPull)
	assert.EqualError(t, err, "head.ref is null")

	testPull = deepcopy.Copy(Pull).(github.PullRequest)
	testPull.Base.Ref = nil
	_, err = pullConverter.Convert(&testPull)
	assert.EqualError(t, err, "base.ref is null")

	testPull = deepcopy.Copy(Pull).(github.PullRequest)
	testPull.User.Login = nil
	_, err = pullConverter.Convert(&testPull)
	assert.EqualError(t, err, "user.login is null")

	testPull = deepcopy.Copy(Pull).(github.PullRequest)
	testPull.Number = nil
	_, err = pullConverter.Convert(&testPull)
	assert.EqualError(t, err, "number is null")

	pullRes, err := pullConverter.Convert(&Pull)
	assert.NoError(t, err)

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
	assert.Equal(t, models.PullRequest{
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
	}, pullRes)
}
