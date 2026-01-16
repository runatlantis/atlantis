// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package events

import (
	"strings"

	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/events/vcs/github"
	"github.com/runatlantis/atlantis/server/logging"
)

const redactedReplacement = "://:<redacted>@"

// GithubAppWorkingDir implements WorkingDir.
// It acts as a proxy to an instance of WorkingDir that refreshes the app's token
// before every clone, given Github App tokens expire quickly
type GithubAppWorkingDir struct {
	WorkingDir
	Credentials    github.Credentials
	GithubHostname string
}

// Clone writes a fresh token for Github App authentication
func (g *GithubAppWorkingDir) Clone(logger logging.SimpleLogging, headRepo models.Repo, p models.PullRequest, workspace string) (string, error) {
	g.fixReposURL(&p, &headRepo)
	return g.WorkingDir.Clone(logger, headRepo, p, workspace)
}

func (g *GithubAppWorkingDir) MergeAgain(logger logging.SimpleLogging, headRepo models.Repo, p models.PullRequest, workspace string) (bool, error) {
	g.fixReposURL(&p, &headRepo)
	return g.WorkingDir.MergeAgain(logger, headRepo, p, workspace)
}

func (g *GithubAppWorkingDir) fixReposURL(p *models.PullRequest, headRepo *models.Repo) {
	// Realistically, this is a super brittle way of supporting clones using gh app installation tokens
	// This URL should be built during Repo creation and the struct should be immutable going forward.
	// Doing this requires a larger refactor however, and can probably be coupled with supporting > 1 installation

	// This removes the credential part from the url and leaves us with the raw http url
	// git will then pick up credentials from the credential store which is set in vcs.WriteGitCreds.
	// Git credentials will then be rotated by vcs.GitCredsTokenRotator
	replacement := "://"
	p.BaseRepo.CloneURL = strings.Replace(p.BaseRepo.CloneURL, "://:@", replacement, 1)
	p.BaseRepo.SanitizedCloneURL = strings.Replace(p.BaseRepo.SanitizedCloneURL, redactedReplacement, replacement, 1)
	headRepo.CloneURL = strings.Replace(headRepo.CloneURL, "://:@", replacement, 1)
	headRepo.SanitizedCloneURL = strings.Replace(p.BaseRepo.SanitizedCloneURL, redactedReplacement, replacement, 1)

}
