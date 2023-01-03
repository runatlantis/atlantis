package converter_test

import (
	"testing"

	"github.com/google/go-github/v45/github"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/neptune/gateway/event"
	"github.com/runatlantis/atlantis/server/vcs/provider/github/converter"
	"github.com/stretchr/testify/assert"
)

func TestConvert_PRReviewEvent(t *testing.T) {
	subject := converter.PullRequestReviewEvent{
		RepoConverter: converter.RepoConverter{
			GithubUser:  "user",
			GithubToken: "token",
		},
	}

	repoFullName := "owner/repo"
	repoOwner := "owner"
	repoName := "repo"
	cloneURL := "https://github.com/owner/repo.git"
	commitID := "123"
	login := "nish"
	user := models.User{Username: login}
	action := "Approved"
	prNum := 10
	installationID := int64(456)

	repo := &github.Repository{
		FullName:      github.String(repoFullName),
		Owner:         &github.User{Login: github.String(repoOwner)},
		Name:          github.String(repoName),
		CloneURL:      github.String(cloneURL),
		DefaultBranch: github.String("main"),
	}

	expectedResult := event.PullRequestReview{
		User:              user,
		PRNum:             prNum,
		Action:            action,
		Ref:               commitID,
		InstallationToken: installationID,
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
			DefaultBranch: "main",
		},
	}

	result, err := subject.Convert(
		&github.PullRequestReviewEvent{
			Repo: repo,
			Review: &github.PullRequestReview{
				CommitID: github.String(commitID),
			},
			Action: github.String(action),
			Sender: &github.User{
				Login: github.String(login),
			},
			PullRequest: &github.PullRequest{
				Number: github.Int(prNum),
			},
			Installation: &github.Installation{
				ID: github.Int64(installationID),
			},
		},
	)

	assert.NoError(t, err)
	assert.Equal(t, expectedResult, result)
}
