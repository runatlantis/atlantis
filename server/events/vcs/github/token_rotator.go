// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package github

import (
	"fmt"
	"time"

	"github.com/runatlantis/atlantis/server/events/vcs/common"
	"github.com/runatlantis/atlantis/server/logging"
	"github.com/runatlantis/atlantis/server/scheduled"
)

// GithubTokenRotator continuously tries to rotate the github app access token every 30 seconds and writes the ~/.git-credentials file
type TokenRotator interface {
	Run()
	GenerateJob() (scheduled.JobDefinition, error)
}

type tokenRotator struct {
	log               logging.SimpleLogging
	githubCredentials Credentials
	githubHostname    string
	gitUser           string
	homeDirPath       string
}

func NewTokenRotator(
	log logging.SimpleLogging,
	githubCredentials Credentials,
	githubHostname string,
	gitUser string,
	homeDirPath string) TokenRotator {

	return &tokenRotator{
		log:               log,
		githubCredentials: githubCredentials,
		githubHostname:    githubHostname,
		gitUser:           gitUser,
		homeDirPath:       homeDirPath,
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
		// at least log the error message here, as we want to notify the that user that the key rotation wasn't successful
		r.log.Err(err.Error())
	}
}

func (r *tokenRotator) rotate() error {
	r.log.Debug("Refreshing Github tokens for .git-credentials")

	token, err := r.githubCredentials.GetToken()
	if err != nil {
		return fmt.Errorf("getting github token: %w", err)
	}
	r.log.Debug("Token successfully refreshed")

	// https://developer.github.com/apps/building-github-apps/authenticating-with-github-apps/#http-based-git-access-by-an-installation
	if err := common.WriteGitCreds(r.gitUser, token, r.githubHostname, r.homeDirPath, r.log, true); err != nil {
		return fmt.Errorf("writing ~/.git-credentials file: %w", err)
	}
	return nil
}
