package vcs

import (
	"github.com/atlantisnorth/atlantis/server/events/models"
)

//go:generate pegomock generate -m --use-experimental-model-gen --package mocks -o mocks/mock_client.go Client

// Client is used to make API calls to a VCS host like GitHub or GitLab.
type Client interface {
	GetModifiedFiles(repo models.Repo, pull models.PullRequest) ([]string, error)
	CreateComment(repo models.Repo, pull models.PullRequest, comment string) error
	PullIsApproved(repo models.Repo, pull models.PullRequest) (bool, error)
	UpdateStatus(repo models.Repo, pull models.PullRequest, state CommitStatus, description string) error
}
