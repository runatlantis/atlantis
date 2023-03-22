package converter

import (
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/vcs/provider/github"
)

// RepoConverter converts a github repository to our internal model.
type RepoConverter struct {
	GithubUser  string
	GithubToken string
}

// ParseGithubRepo parses the response from the GitHub API endpoint that
// returns a repo into the Atlantis model.
func (c RepoConverter) Convert(ghRepo github.ExternalRepo) (models.Repo, error) {
	repo, err := models.NewRepo(models.Github, ghRepo.GetFullName(), ghRepo.GetCloneURL(), c.GithubUser, c.GithubToken)

	if err != nil {
		return repo, err
	}

	// set the default branch here since it's annoying to do so within models since a lot of other code depends on it.
	// We should be moving all that code here in the future anyways
	repo.DefaultBranch = ghRepo.GetDefaultBranch()
	return repo, nil
}
