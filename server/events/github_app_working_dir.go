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
func (g *GithubAppWorkingDir) Clone(log logging.SimpleLogging, headRepo models.Repo, p models.PullRequest, workspace string) (string, bool, error) {

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

	baseRepo := &p.BaseRepo

	// Realistically, this is a super brittle way of supporting clones using gh app installation tokens
	// This URL should be built during Repo creation and the struct should be immutable going forward.
	// Doing this requires a larger refactor however, and can probably be coupled with supporting > 1 installation
	authURL := fmt.Sprintf("://x-access-token:%s", token)
	baseRepo.CloneURL = strings.Replace(baseRepo.CloneURL, "://:", authURL, 1)
	baseRepo.SanitizedCloneURL = strings.Replace(baseRepo.SanitizedCloneURL, "://:", "://x-access-token:", 1)
	headRepo.CloneURL = strings.Replace(headRepo.CloneURL, "://:", authURL, 1)
	headRepo.SanitizedCloneURL = strings.Replace(baseRepo.SanitizedCloneURL, "://:", "://x-access-token:", 1)

	return g.WorkingDir.Clone(log, headRepo, p, workspace)
}
