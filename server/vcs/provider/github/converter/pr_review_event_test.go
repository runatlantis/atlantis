package converter_test

import (
	"github.com/runatlantis/atlantis/server/vcs"
	"testing"
	"time"

	"github.com/google/go-github/v45/github"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/neptune/gateway/event"
	"github.com/runatlantis/atlantis/server/vcs/provider/github/converter"
	"github.com/stretchr/testify/assert"
)

func TestConvert_PRReviewEvent(t *testing.T) {
	repoConverter := converter.RepoConverter{
		GithubUser:  "user",
		GithubToken: "token",
	}
	subject := converter.PullRequestReviewEvent{
		RepoConverter: repoConverter,
		PullConverter: converter.PullConverter{
			RepoConverter: repoConverter,
		},
	}

	repoFullName := "owner/repo"
	repoOwner := "owner"
	repoName := "repo"
	cloneURL := "https://github.com/owner/repo.git"
	commitID := "123"
	login := "nish"
	user := models.User{Username: login}
	state := "approved"
	prNum := 10
	installationID := int64(456)
	time := time.Now()
	url := "test.com"

	repo := &github.Repository{
		FullName:      github.String(repoFullName),
		Owner:         &github.User{Login: github.String(repoOwner)},
		Name:          github.String(repoName),
		CloneURL:      github.String(cloneURL),
		DefaultBranch: github.String("main"),
	}

	pr := &github.PullRequest{
		Number: github.Int(prNum),
		Head: &github.PullRequestBranch{
			SHA:  github.String(commitID),
			Ref:  github.String(commitID),
			Repo: repo,
		},
		Base: &github.PullRequestBranch{
			SHA:  github.String(commitID),
			Ref:  github.String(commitID),
			Repo: repo,
		},
		User: &github.User{
			Login: github.String(login),
		},
		HTMLURL: github.String(url),
	}

	expectedRepo := models.Repo{
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
	}

	expectedPull := models.PullRequest{
		Num:        prNum,
		HeadCommit: commitID,
		URL:        url,
		HeadBranch: commitID,
		BaseBranch: commitID,
		Author:     user.Username,
		BaseRepo:   expectedRepo,
		HeadRepo:   expectedRepo,
		State:      models.ClosedPullState,
		HeadRef: vcs.Ref{
			Type: vcs.BranchRef,
			Name: commitID,
		},
	}

	expectedResult := event.PullRequestReview{
		InstallationToken: installationID,
		Repo:              expectedRepo,
		User:              user,
		State:             state,
		Ref:               commitID,
		Timestamp:         time,
		Pull:              expectedPull,
	}

	result, err := subject.Convert(
		&github.PullRequestReviewEvent{
			Repo: repo,
			Review: &github.PullRequestReview{
				CommitID:    github.String(commitID),
				SubmittedAt: &time,
				State:       github.String(state),
			},
			Sender: &github.User{
				Login: github.String(login),
			},
			PullRequest: pr,
			Installation: &github.Installation{
				ID: github.Int64(installationID),
			},
		},
	)

	assert.NoError(t, err)
	assert.Equal(t, expectedResult, result)
}
