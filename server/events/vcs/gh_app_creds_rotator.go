package vcs

import (
	"time"

	"github.com/mitchellh/go-homedir"
	"github.com/pkg/errors"

	"github.com/runatlantis/atlantis/server/logging"
	"github.com/runatlantis/atlantis/server/scheduled"
)

// GitCredsTokenRotator continuously tries to rotate the github app access token every 30 seconds and writes the ~/.git-credentials file
type GitCredsTokenRotator interface {
	Run()
	rotate() error
	GenerateJob() (scheduled.JobDefinition, error)
}

type githubAppTokenRotator struct {
	log               logging.SimpleLogging
	githubCredentials GithubCredentials
	githubHostname    string
}

func NewGithubAppTokenRotator(
	logger logging.SimpleLogging,
	githubCredentials GithubCredentials,
	githubHostname string) GitCredsTokenRotator {
	return &githubAppTokenRotator{
		log:               logger,
		githubCredentials: githubCredentials,
		githubHostname:    githubHostname,
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
	r.rotate()
}

func (r *githubAppTokenRotator) rotate() error {
	r.log.Debug("Refreshing git tokens for Github App")
	token, err := r.githubCredentials.GetToken()
	if err != nil {
		return errors.Wrap(err, "getting github token")
	}
	r.log.Debug("token %s", token)
	home, err := homedir.Dir()
	if err != nil {
		return errors.Wrap(err, "getting home dir to write ~/.git-credentials file")
	}

	// https://developer.github.com/apps/building-github-apps/authenticating-with-github-apps/#http-based-git-access-by-an-installation
	if err := WriteGitCreds("x-access-token", token, r.githubHostname, home, r.log, true); err != nil {
		return errors.Wrap(err, "writing ~/.git-credentials file")
	}
	return nil
}
