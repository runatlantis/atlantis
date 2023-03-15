package vcs

import (
	"time"

	"github.com/mitchellh/go-homedir"
	"github.com/pkg/errors"

	"github.com/runatlantis/atlantis/server/logging"
)

// GitCredsTokenRotator continuously tries to rotate the github app access token every 30 seconds and writes the ~/.git-credentials file
type GitCredsTokenRotator interface {
	Start() error
	Stop()
	rotate() error
}

type githubAppTokenRotator struct {
	stop              chan struct{}
	log               logging.SimpleLogging
	githubCredentials GithubCredentials
	githubHostname    string
}

func NewGithubAppTokenRotator(
	logger logging.SimpleLogging,
	githubCredentials GithubCredentials,
	githubHostname string) GitCredsTokenRotator {
	return &githubAppTokenRotator{
		stop:              make(chan struct{}),
		log:               logger,
		githubCredentials: githubCredentials,
		githubHostname:    githubHostname,
	}
}

// make sure interface is implemented correctly
var _ GitCredsTokenRotator = (*githubAppTokenRotator)(nil)

func (r *githubAppTokenRotator) Start() error {
	ticker := time.NewTicker(30 * time.Second)

	go func() {
		for {
			select {
			case <-ticker.C:
				r.rotate()
			case <-r.stop:
				ticker.Stop()
				return
			}
		}
	}()

	return r.rotate()
}

func (r *githubAppTokenRotator) Stop() {
	close(r.stop)
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
