package events

import "github.com/runatlantis/atlantis/server/events/models"

//go:generate pegomock generate -m --use-experimental-model-gen --package mocks -o mocks/mock_pull_status_fetcher.go PullStatusFetcher

// PullStatusFetcher fetches our internal model of a pull requests status
type PullStatusFetcher interface {
	GetPullStatus(pull models.PullRequest) (*models.PullStatus, error)
}
