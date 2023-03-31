package events

import (
	"strings"

	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/events/vcs"
	"github.com/runatlantis/atlantis/server/logging"
)

const redactedReplacement = "://:<redacted>@"

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
	baseRepo := &p.BaseRepo

	// Realistically, this is a super brittle way of supporting clones using gh app installation tokens
	// This URL should be built during Repo creation and the struct should be immutable going forward.
	// Doing this requires a larger refactor however, and can probably be coupled with supporting > 1 installation

	// This removes the credential part from the url and leaves us with the raw http url
	// git will then pick up credentials from the credential store which is set in vcs.WriteGitCreds.
	// Git credentials will then be rotated by vcs.GitCredsTokenRotator
	replacement := "://"
	baseRepo.CloneURL = strings.Replace(baseRepo.CloneURL, "://:@", replacement, 1)
	baseRepo.SanitizedCloneURL = strings.Replace(baseRepo.SanitizedCloneURL, redactedReplacement, replacement, 1)
	headRepo.CloneURL = strings.Replace(headRepo.CloneURL, "://:@", replacement, 1)
	headRepo.SanitizedCloneURL = strings.Replace(baseRepo.SanitizedCloneURL, redactedReplacement, replacement, 1)

	return g.WorkingDir.Clone(log, headRepo, p, workspace)
}
