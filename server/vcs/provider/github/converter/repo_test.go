package converter_test

import (
	"testing"

	"github.com/google/go-github/v45/github"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/vcs/provider/github/converter"
	"github.com/stretchr/testify/assert"
)

var Repo = github.Repository{
	FullName: github.String("owner/repo"),
	Owner:    &github.User{Login: github.String("owner")},
	Name:     github.String("repo"),
	CloneURL: github.String("https://github.com/owner/repo.git"),
}

func TestParseGithubRepo(t *testing.T) {
	repoConverter := converter.RepoConverter{
		GithubUser:  "github-user",
		GithubToken: "github-token",
	}
	r, err := repoConverter.Convert(&Repo)
	assert.NoError(t, err)
	assert.Equal(t, models.Repo{
		Owner:             "owner",
		FullName:          "owner/repo",
		CloneURL:          "https://github-user:github-token@github.com/owner/repo.git",
		SanitizedCloneURL: "https://github-user:<redacted>@github.com/owner/repo.git",
		Name:              "repo",
		VCSHost: models.VCSHost{
			Hostname: "github.com",
			Type:     models.Github,
		},
	}, r)
}
