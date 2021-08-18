package feature

import (
	"context"

	"github.com/pkg/errors"
	"github.com/runatlantis/atlantis/server/events/vcs"
	"github.com/runatlantis/atlantis/server/logging"
	"github.com/thomaspoignant/go-feature-flag"
	"github.com/thomaspoignant/go-feature-flag/ffuser"
)

// If keys are missing, the default is respected, so we don't need to have
// real configuration here.
const Configuration StringRetriever = `some-key:
  true: true
  false: false
  default: false
  trackEvents: false`

// StringRetriever is used for testing purposes
type StringRetriever string

func (s StringRetriever) Retrieve(ctx context.Context) ([]byte, error) {
	return []byte(s), nil
}

type RepoConfig struct {
	Owner  string
	Repo   string
	Branch string
	Path   string
}

// CustomGithubRetriever uses Atlantis' internal client to retrieve the contents
// of the feature file.  This allows us to re-use GH credentials easily as opposed
// to the default ffclient.GithubRetriever.
type CustomGithubRetriever struct {
	client vcs.IGithubClient
	cfg    RepoConfig
}

func (c *CustomGithubRetriever) Retrieve(ctx context.Context) ([]byte, error) {
	return c.client.GetContents(c.cfg.Owner, c.cfg.Repo, c.cfg.Branch, c.cfg.Path)
}

// Allocator allocates features given a feature id and full repo name.
// Note: This means that if a feature is enabled for a repository, it will be enabled
// for all operations on that given repository regardless of the PR/Operation

// Additionally, implementations are assumed to provide deterministic results.
//go:generate pegomock generate -m --use-experimental-model-gen --package mocks -o mocks/mock_allocator.go Allocator
type Allocator interface {
	ShouldAllocate(featureID Name, fullRepoName string) (bool, error)
}

// PercentageBasedAllocator allocates features based on a percentage of the total repositories
// at the time of writing, go-feature-flag is primarily used as the backend for this allocator:
// https://thomaspoignant.github.io/go-feature-flag/
type PercentageBasedAllocator struct {
	logger logging.SimpleLogging
}

func NewGHSourcedAllocator(repoConfig RepoConfig, githubClient vcs.IGithubClient, logger logging.SimpleLogging) (Allocator, error) {

	// default to local config if no github client is provided.
	if githubClient == nil {
		logger.Warn("no github client provided, defaulting to local config for feature allocation.")
		return NoopAllocator{}, nil
	}

	err := ffclient.Init(
		ffclient.Config{
			Context:   context.Background(),
			Retriever: &CustomGithubRetriever{client: githubClient, cfg: repoConfig},
		},
	)

	if err != nil {
		return nil, errors.Wrapf(err, "initializing feature allocator")
	}

	return &PercentageBasedAllocator{logger: logger}, err

}

// NewStringSourcedAllocator uses a string constant for the feature configuration
// This is used for e2e testing.  This should reflect the configuration we intend to test.
//
// External configuration shouldn't be dialed up to 100 unless we've tested that scenario here.
func NewStringSourcedAllocator(logger logging.SimpleLogging) (Allocator, error) {
	err := ffclient.Init(
		ffclient.Config{
			Context:   context.Background(),
			Retriever: Configuration,
		},
	)

	if err != nil {
		return nil, errors.Wrapf(err, "initializing feature allocator")
	}

	return &PercentageBasedAllocator{logger: logger}, err
}

func (r *PercentageBasedAllocator) ShouldAllocate(featureID Name, fullRepoName string) (bool, error) {
	repo := ffuser.NewUser(fullRepoName)
	shouldAllocate, err := ffclient.BoolVariation(string(featureID), repo, false)

	r.logger.Debug("feature %s allocation: %t for repo: %s", featureID, shouldAllocate, fullRepoName)

	// if we error out we shouldn't enable the feature, could be risky
	// Note: if the feature doesn't exist, the library returns the default value.
	if err != nil {
		return false, errors.Wrapf(err, "getting feature %s", featureID)
	}

	return shouldAllocate, nil
}

// NoopAllocator is used in exceptional circumstances as a backup where all features
// are disabled by default for all repos. This is to ensure we don't fail to startup
// due to missing configuration since features are treated as secondary citizens.
type NoopAllocator struct{}

func (r NoopAllocator) ShouldAllocate(featureID Name, fullRepoName string) (bool, error) {
	return false, nil
}
