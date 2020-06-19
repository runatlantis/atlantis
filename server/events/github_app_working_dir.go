package events

import (
	"fmt"
	"strings"

	"github.com/mitchellh/go-homedir"
	"github.com/pkg/errors"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/events/vcs"
	"github.com/runatlantis/atlantis/server/logging"
)

// GithubAppWorkingDir implements WorkingDir.
// It acts as a proxy to an instance of WorkingDir that refreshes the app's token
// before every clone, given Github App tokens expire quickly
type GithubAppWorkingDir struct {
	WorkingDir
	Credentials    vcs.GithubCredentials
	GithubHostname string
}

// Clone writes a fresh token for Github App authentication
func (g *GithubAppWorkingDir) Clone(log *logging.SimpleLogger, baseRepo models.Repo, headRepo models.Repo, p models.PullRequest, workspace string) (string, bool, error) {

	log.Info("Refreshing git tokens for Github App")

	token, err := g.Credentials.GetToken()
	if err != nil {
		return "", false, errors.Wrap(err, "getting github token")
	}

	home, err := homedir.Dir()
	if err != nil {
		return "", false, errors.Wrap(err, "getting home dir to write ~/.git-credentials file")
	}

	// https://developer.github.com/apps/building-github-apps/authenticating-with-github-apps/#http-based-git-access-by-an-installation
	if err := WriteGitCreds("x-access-token", token, g.GithubHostname, home, log, true); err != nil {
		return "", false, err
	}

	authURL := fmt.Sprintf("://x-access-token:%s", token)
	baseRepo.CloneURL = strings.Replace(baseRepo.CloneURL, "://:", authURL, 1)
	headRepo.CloneURL = strings.Replace(headRepo.CloneURL, "://:", authURL, 1)
	return g.WorkingDir.Clone(log, baseRepo, headRepo, p, workspace)
}
