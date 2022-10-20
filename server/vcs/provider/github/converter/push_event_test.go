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
		FullName:      github.String(repoFullName),
		Owner:         &github.User{Login: github.String(repoOwner)},
		Name:          github.String(repoName),
		CloneURL:      github.String(cloneURL),
		DefaultBranch: github.String("main"),
	}

	t.Run("created branch ref", func(t *testing.T) {
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
				DefaultBranch: "main",
			},
			Ref: vcs.Ref{
				Type: vcs.BranchRef,
				Name: "main",
			},
			Sha: sha,
			Sender: models.User{
				Username: "nish",
			},
			InstallationToken: int64(installationToken),
			Action:            event.CreatedAction,
		}

		result, err := subject.Convert(&github.PushEvent{
			Repo: pushEventRepo,
			Ref:  github.String(ref),
			HeadCommit: &github.HeadCommit{
				ID: github.String(sha),
			},
			Sender: &github.User{
				Login: github.String(user),
			},
			Installation: &github.Installation{
				ID: github.Int64(int64(installationToken)),
			},
			Created: github.Bool(true),
		})

		assert.NoError(t, err)
		assert.Equal(t, expectedResult, result)
	})

	t.Run("deleted branch ref", func(t *testing.T) {
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
				DefaultBranch: "main",
			},
			Ref: vcs.Ref{
				Type: vcs.BranchRef,
				Name: "main",
			},
			Sha: sha,
			Sender: models.User{
				Username: "nish",
			},
			InstallationToken: int64(installationToken),
			Action:            event.DeletedAction,
		}

		result, err := subject.Convert(&github.PushEvent{
			Repo: pushEventRepo,
			Ref:  github.String(ref),
			HeadCommit: &github.HeadCommit{
				ID: github.String(sha),
			},
			Sender: &github.User{
				Login: github.String(user),
			},
			Installation: &github.Installation{
				ID: github.Int64(int64(installationToken)),
			},
			Deleted: github.Bool(true),
		})

		assert.NoError(t, err)
		assert.Equal(t, expectedResult, result)
	})
}
