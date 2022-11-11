package feature

import (
	"context"
	"time"

	"github.com/pkg/errors"
	"github.com/runatlantis/atlantis/server/events/vcs"
	"github.com/runatlantis/atlantis/server/logging"
	ffclient "github.com/thomaspoignant/go-feature-flag"
	"github.com/thomaspoignant/go-feature-flag/ffuser"
)

// Add more attributes as needed to determine eligibility of a feature
type FeatureContext struct { //nolint:revive // avoiding refactor while adding linter action
	RepoName         string
	PullCreationTime time.Time
}

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
//
//go:generate pegomock generate -m --use-experimental-model-gen --package mocks -o mocks/mock_allocator.go Allocator
type Allocator interface {
	ShouldAllocate(featureID Name, featureCtx FeatureContext) (bool, error)
}

func NewGHSourcedAllocator(repoConfig RepoConfig, githubClient vcs.IGithubClient, logger logging.Logger) (Allocator, error) {
	// fail if no github client is provided
	if githubClient == nil {
		return nil, errors.New("no github client provided")
	}

	ff, err := ffclient.New(
		ffclient.Config{
			Context:   context.Background(),
			Retriever: &CustomGithubRetriever{client: githubClient, cfg: repoConfig},
		},
	)

	if err != nil {
		return nil, errors.Wrapf(err, "initializing feature allocator")
	}

	return &PercentageBasedAllocator{logger: logger, featureFlag: ff}, err

}

// NewStringSourcedAllocator uses a string constant for the feature configuration
// This is used for e2e testing.  This should reflect the configuration we intend to test.
//
// External configuration shouldn't be dialed up to 100 unless we've tested that scenario here.
func NewStringSourcedAllocator(logger logging.Logger) (*PercentageBasedAllocator, error) {
	return NewStringSourcedAllocatorWithRetriever(logger, Configuration)
}

func NewStringSourcedAllocatorWithRetriever(logger logging.Logger, retriever ffclient.Retriever) (*PercentageBasedAllocator, error) {
	ff, err := ffclient.New(
		ffclient.Config{
			Context:   context.Background(),
			Retriever: retriever,
		},
	)

	if err != nil {
		return nil, errors.Wrapf(err, "initializing feature allocator")
	}

	return &PercentageBasedAllocator{logger: logger, featureFlag: ff}, err
}

// PercentageBasedAllocator allocates features based on a percentage of the total repositories
// at the time of writing, go-feature-flag is primarily used as the backend for this allocator:
// https://thomaspoignant.github.io/go-feature-flag/
type PercentageBasedAllocator struct {
	logger      logging.Logger
	featureFlag *ffclient.GoFeatureFlag
}

func (r *PercentageBasedAllocator) ShouldAllocate(featureID Name, featureContext FeatureContext) (bool, error) {
	// rule defintion used by this ff definition is not smart enough to understand different time formats
	// so we use the Unix() time in seconds to evaluate if this feature should be allocated
	repo := ffuser.NewUserBuilder(featureContext.RepoName).
		AddCustom("prCreationTime", featureContext.PullCreationTime.Unix()).
		Build()

	shouldAllocate, err := r.featureFlag.BoolVariation(string(featureID), repo, false)

	// if we error out we shouldn't enable the feature, could be risky
	// Note: if the feature doesn't exist, the library returns the default value.
	if err != nil {
		return false, errors.Wrapf(err, "getting feature %s", featureID)
	}

	return shouldAllocate, nil
}

func (r *PercentageBasedAllocator) Close() {
	r.featureFlag.Close()
}
