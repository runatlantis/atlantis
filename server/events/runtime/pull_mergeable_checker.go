package runtime

import (
	"github.com/runatlantis/atlantis/server/events/models"
)

//go:generate pegomock generate -m --use-experimental-model-gen --package mocks -o mocks/mock_pull_mergeable_checker.go PullMergeableChecker

type PullMergeableChecker interface {
	PullIsMergeable(baseRepo models.Repo, pull models.PullRequest) (bool, error)
}
