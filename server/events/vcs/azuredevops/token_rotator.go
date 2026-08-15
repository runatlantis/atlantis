// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package azuredevops

import (
	"fmt"
	"time"

	"github.com/runatlantis/atlantis/server/events/vcs/common"
	"github.com/runatlantis/atlantis/server/logging"
	"github.com/runatlantis/atlantis/server/scheduled"
)

// TokenRotator continuously refreshes the Azure DevOps token and rewrites the
// ~/.git-credentials file so git operations always use a current token. This is
// used together with --azuredevops-token-file so that an externally-rotated
// token is picked up without restarting Atlantis.
type TokenRotator interface {
	Run()
	GenerateJob() (scheduled.JobDefinition, error)
}

type tokenRotator struct {
	log         logging.SimpleLogging
	credentials Credentials
	hostname    string
	gitUser     string
	homeDirPath string
}

// NewTokenRotator returns a TokenRotator that refreshes ~/.git-credentials for
// Azure DevOps using the provided credentials.
func NewTokenRotator(
	log logging.SimpleLogging,
	credentials Credentials,
	hostname string,
	gitUser string,
	homeDirPath string) TokenRotator {

	return &tokenRotator{
		log:         log,
		credentials: credentials,
		hostname:    hostname,
		gitUser:     gitUser,
		homeDirPath: homeDirPath,
	}
}

// make sure interface is implemented correctly
var _ TokenRotator = (*tokenRotator)(nil)

func (r *tokenRotator) GenerateJob() (scheduled.JobDefinition, error) {

	return scheduled.JobDefinition{
		Job:    r,
		Period: 30 * time.Second,
	}, r.rotate()
}

func (r *tokenRotator) Run() {
	err := r.rotate()
	if err != nil {
		// at least log the error message here, as we want to notify the user that the token rotation wasn't successful
		r.log.Err("%s", err.Error())
	}
}

func (r *tokenRotator) rotate() error {
	r.log.Debug("refreshing azure devops token for .git-credentials")

	token, err := r.credentials.GetToken()
	if err != nil {
		return fmt.Errorf("getting azure devops token: %w", err)
	}

	// Guard against a transiently empty token (e.g. the backing secret was
	// briefly empty mid-rotation or was accidentally wiped). Overwriting the
	// last-good credential with an empty token would break git operations until
	// the next cycle, so skip this refresh and keep the working credential.
	if token == "" {
		r.log.Warn("azure devops token is empty; keeping the last-good .git-credentials")
		return nil
	}
	r.log.Debug("token successfully refreshed")

	// Replace the existing line in place (keyed on user + hostname) so rotated
	// tokens don't accumulate stale entries.
	if err := common.WriteGitCreds(r.gitUser, token, r.hostname, r.homeDirPath, r.log, true); err != nil {
		return fmt.Errorf("writing ~/.git-credentials file: %w", err)
	}
	return nil
}
