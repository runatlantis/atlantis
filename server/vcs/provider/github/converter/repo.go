package converter

import (
	"github.com/google/go-github/v45/github"
	"github.com/runatlantis/atlantis/server/events/models"
)

// RepoConverter converts a github repository to our internal model.
type RepoConverter struct {
	GithubUser  string
	GithubToken string
}

// ParseGithubRepo parses the response from the GitHub API endpoint that
// returns a repo into the Atlantis model.
func (c *RepoConverter) Convert(ghRepo *github.Repository) (models.Repo, error) {
	return models.NewRepo(models.Github, ghRepo.GetFullName(), ghRepo.GetCloneURL(), c.GithubUser, c.GithubToken)
}
