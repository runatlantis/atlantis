package events_test

import (
	"testing"

	. "github.com/petergtz/pegomock"
	"github.com/runatlantis/atlantis/server/events"
	eventMocks "github.com/runatlantis/atlantis/server/events/mocks"
	"github.com/runatlantis/atlantis/server/events/models"
	vcsMocks "github.com/runatlantis/atlantis/server/events/vcs/mocks"
	"github.com/runatlantis/atlantis/server/logging"
	. "github.com/runatlantis/atlantis/testing"
)

func TestClone_GithubAppSetsCorrectUrl(t *testing.T) {
	workingDir := eventMocks.NewMockWorkingDir()

	credentials := vcsMocks.NewMockGithubCredentials()

	ghAppWorkingDir := events.GithubAppWorkingDir{
		WorkingDir:     workingDir,
		Credentials:    credentials,
		GithubHostname: "some-host",
	}

	baseRepo, _ := models.NewRepo(
		models.Github,
		"runatlantis/atlantis",
		"https://github.com/runatlantis/atlantis.git",

		// user and token have to be blank otherwise this proxy wouldn't be invoked to begin with
		"",
		"",
	)

	logger := logging.NewNoopCtxLogger(t)

	headRepo := baseRepo

	modifiedBaseRepo := baseRepo
	modifiedBaseRepo.CloneURL = "https://x-access-token:token@github.com/runatlantis/atlantis.git"
	modifiedBaseRepo.SanitizedCloneURL = "https://x-access-token:<redacted>@github.com/runatlantis/atlantis.git"

	When(credentials.GetToken()).ThenReturn("token", nil)
	When(workingDir.Clone(logger, modifiedBaseRepo, models.PullRequest{BaseRepo: modifiedBaseRepo}, "default")).ThenReturn(
		"", true, nil,
	)

	_, success, _ := ghAppWorkingDir.Clone(logger, headRepo, models.PullRequest{BaseRepo: baseRepo}, "default")

	Assert(t, success == true, "clone url mutation error")
}
