package converter_test

import (
	"testing"

	"github.com/google/go-github/v45/github"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/neptune/gateway/event"
	"github.com/runatlantis/atlantis/server/vcs/provider/github/converter"
	"github.com/stretchr/testify/assert"
)

func TestConvert_CheckRunEvent(t *testing.T) {
	subject := converter.CheckRunEvent{
		RepoConverter: converter.RepoConverter{
			GithubUser:  "user",
			GithubToken: "token",
		},
	}

	repoFullName := "owner/repo"
	repoOwner := "owner"
	repoName := "repo"
	cloneURL := "https://github.com/owner/repo.git"
	requestedActionsID := "some_id"
	externalID := "external id"
	checkRunName := "some name"
	login := "nish"
	user := models.User{Username: login}

	repo := &github.Repository{
		FullName:      github.String(repoFullName),
		Owner:         &github.User{Login: github.String(repoOwner)},
		Name:          github.String(repoName),
		CloneURL:      github.String(cloneURL),
		DefaultBranch: github.String("main"),
	}

	t.Run("success with requested actions", func(t *testing.T) {
		expectedResult := event.CheckRun{
			Name: checkRunName,
			Action: event.RequestedActionChecksAction{
				Identifier: requestedActionsID,
			},
			ExternalID: externalID,
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
			User: user,
		}

		result, err := subject.Convert(
			&github.CheckRunEvent{
				Repo:   repo,
				Action: github.String("requested_action"),
				RequestedAction: &github.RequestedAction{
					Identifier: requestedActionsID,
				},
				CheckRun: &github.CheckRun{
					ExternalID: github.String(externalID),
					Name:       github.String(checkRunName),
				},
				Sender: &github.User{
					Login: github.String(login),
				},
			},
		)

		assert.NoError(t, err)
		assert.Equal(t, expectedResult, result)
	})

	t.Run("success with requested actions", func(t *testing.T) {
		expectedResult := event.CheckRun{
			Name:       checkRunName,
			Action:     event.WrappedCheckRunAction("created"),
			ExternalID: externalID,
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
			User: user,
		}

		result, err := subject.Convert(
			&github.CheckRunEvent{
				Repo:   repo,
				Action: github.String("created"),
				RequestedAction: &github.RequestedAction{
					Identifier: requestedActionsID,
				},
				CheckRun: &github.CheckRun{
					ExternalID: github.String(externalID),
					Name:       github.String(checkRunName),
				},
				Sender: &github.User{
					Login: github.String(login),
				},
			},
		)

		assert.NoError(t, err)
		assert.Equal(t, expectedResult, result)
	})

}
