package vcs

import (
	"fmt"

	"github.com/hootsuite/atlantis/server/events/models"
)

// NotConfiguredVCSClient is used as a placeholder when Atlantis isn't configured
// on startup to support a certain VCS host. For example, if there is no GitHub
// config then this client will be used which will error if it's ever called.
type NotConfiguredVCSClient struct {
	Host Host
}

func (a *NotConfiguredVCSClient) GetModifiedFiles(repo models.Repo, pull models.PullRequest) ([]string, error) {
	return nil, a.err()
}
func (a *NotConfiguredVCSClient) CreateComment(repo models.Repo, pull models.PullRequest, comment string) error {
	return a.err()
}
func (a *NotConfiguredVCSClient) PullIsApproved(repo models.Repo, pull models.PullRequest) (bool, error) {
	return false, a.err()
}
func (a *NotConfiguredVCSClient) UpdateStatus(repo models.Repo, pull models.PullRequest, state CommitStatus, description string) error {
	return a.err()
}
func (a *NotConfiguredVCSClient) err() error {
	//noinspection GoErrorStringFormat
	return fmt.Errorf("Atlantis was not configured to support repos from %s", a.Host.String())
}
