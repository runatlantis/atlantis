package vcs

import (
	"time"

	"github.com/pkg/errors"
	"github.com/runatlantis/atlantis/server/logging"
	"github.com/runatlantis/atlantis/server/scheduled"
)

// GitCredsTokenRotator continuously tries to rotate the github app access token every 30 seconds and writes the ~/.git-credentials file
type GitCredsTokenRotator interface {
	Run()
	GenerateJob() (scheduled.JobDefinition, error)
}

type githubAppTokenRotator struct {
	log               logging.SimpleLogging
	githubCredentials GithubCredentials
	githubHostname    string
	homeDirPath       string
}

func NewGithubAppTokenRotator(
	log logging.SimpleLogging,
	githubCredentials GithubCredentials,
	githubHostname string,
	homeDirPath string) GitCredsTokenRotator {

	return &githubAppTokenRotator{
		log:               log,
		githubCredentials: githubCredentials,
		githubHostname:    githubHostname,
		homeDirPath:       homeDirPath,
	}
}

// make sure interface is implemented correctly
var _ GitCredsTokenRotator = (*githubAppTokenRotator)(nil)

func (r *githubAppTokenRotator) GenerateJob() (scheduled.JobDefinition, error) {

	return scheduled.JobDefinition{
		Job:    r,
		Period: 30 * time.Second,
	}, r.rotate()
}

func (r *githubAppTokenRotator) Run() {
	err := r.rotate()
	if err != nil {
		// at least log the error message here, as we want to notify the that user that the key rotation wasn't successful
		r.log.Err(err.Error())
	}
}

func (r *githubAppTokenRotator) rotate() error {
	r.log.Debug("Refreshing git tokens for Github App")

	token, err := r.githubCredentials.GetToken()
	if err != nil {
		return errors.Wrap(err, "Getting github token")
	}
	r.log.Debug("Token successfully refreshed")

	// https://developer.github.com/apps/building-github-apps/authenticating-with-github-apps/#http-based-git-access-by-an-installation
	if err := WriteGitCreds("x-access-token", token, r.githubHostname, r.homeDirPath, r.log, true); err != nil {
		return errors.Wrap(err, "Writing ~/.git-credentials file")
	}
	return nil
}
