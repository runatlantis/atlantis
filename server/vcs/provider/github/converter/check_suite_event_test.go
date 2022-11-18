package converter_test

import (
	"github.com/google/go-github/v45/github"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/neptune/gateway/event"
	"github.com/runatlantis/atlantis/server/vcs/provider/github/converter"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestConvert_CheckSuiteEvent(t *testing.T) {
	subject := converter.CheckSuiteEvent{
		RepoConverter: converter.RepoConverter{
			GithubUser:  "user",
			GithubToken: "token",
		},
	}

	repoFullName := "owner/repo"
	repoOwner := "owner"
	repoName := "repo"
	cloneURL := "https://github.com/owner/repo.git"
	login := "nish"
	user := models.User{Username: login}
	headBranch := "main"
	headSha := "123"

	repo := &github.Repository{
		FullName:      github.String(repoFullName),
		Owner:         &github.User{Login: github.String(repoOwner)},
		Name:          github.String(repoName),
		CloneURL:      github.String(cloneURL),
		DefaultBranch: github.String(headBranch),
	}
	expectedResult := event.CheckSuite{
		Action: event.WrappedCheckRunAction("rerequested_action"),
		Repo: models.Repo{
			FullName:          repoFullName,
			Owner:             repoOwner,
			Name:              repoName,
			CloneURL:          "https://user:token@github.com/owner/repo.git",
			SanitizedCloneURL: "https://user:<redacted>@github.com/owner/repo.git",
			VCSHost: models.VCSHost{
				Type:     models.Github,
				Hostname: "github.com",
			},
			DefaultBranch: headBranch,
		},
		Sender:  user,
		HeadSha: headSha,
		Branch:  headBranch,
	}

	result, err := subject.Convert(
		&github.CheckSuiteEvent{
			Repo:   repo,
			Action: github.String("rerequested_action"),
			CheckSuite: &github.CheckSuite{
				HeadSHA:    github.String(headSha),
				HeadBranch: github.String(headBranch),
			},
			Sender: &github.User{
				Login: github.String(login),
			},
		},
	)

	assert.NoError(t, err)
	assert.Equal(t, expectedResult, result)
}
