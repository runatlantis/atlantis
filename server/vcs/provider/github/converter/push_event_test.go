package converter_test

import (
	"testing"

	"github.com/google/go-github/v45/github"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/neptune/gateway/event"
	"github.com/runatlantis/atlantis/server/vcs"
	"github.com/runatlantis/atlantis/server/vcs/provider/github/converter"
	"github.com/stretchr/testify/assert"
)

func TestConvert_PushEvent(t *testing.T) {
	subject := converter.PushEvent{
		RepoConverter: converter.RepoConverter{
			GithubUser:  "user",
			GithubToken: "token",
		},
	}

	repoFullName := "owner/repo"
	repoOwner := "owner"
	repoName := "repo"
	cloneURL := "https://github.com/owner/repo.git"
	ref := "refs/heads/main"
	sha := "1234"
	user := "nish"
	installationToken := 1234

	pushEventRepo := &github.PushEventRepository{
		FullName: github.String(repoFullName),
		Owner:    &github.User{Login: github.String(repoOwner)},
		Name:     github.String(repoName),
		CloneURL: github.String(cloneURL),
	}

	expectedResult := event.Push{
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
		},
		Ref: vcs.Ref{
			Type: vcs.BranchRef,
			Name: "main",
		},
		Sha: sha,
		Sender: vcs.User{
			Login: "nish",
		},
		InstallationToken: int64(installationToken),
	}

	result, err := subject.Convert(&github.PushEvent{
		Repo: pushEventRepo,
		Ref:  github.String(ref),
		HeadCommit: &github.HeadCommit{
			SHA: github.String(sha),
		},
		Sender: &github.User{
			Login: github.String(user),
		},
		Installation: &github.Installation{
			ID: github.Int64(int64(installationToken)),
		},
	})

	assert.NoError(t, err)
	assert.Equal(t, expectedResult, result)
}
