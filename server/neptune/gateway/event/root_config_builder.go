package event

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"github.com/runatlantis/atlantis/server/core/config"
	"github.com/runatlantis/atlantis/server/core/config/valid"
	"github.com/runatlantis/atlantis/server/events/metrics"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/logging"
	"github.com/runatlantis/atlantis/server/vcs/provider/github"
	"github.com/uber-go/tally/v4"
)

// repoFetcher manages a cloned repo's workspace on disk for running commands.
type repoFetcher interface {
	Fetch(ctx context.Context, baseRepo models.Repo, branch string, sha string, options github.RepoFetcherOptions) (string, func(ctx context.Context, filePath string), error)
}

// hooksRunner runs preworkflow hooks for a given repository/commit
type hooksRunner interface {
	Run(repo models.Repo, repoDir string) error
}

// fileFetcher handles being able to identify and fetch the changed files per individual commit
type fileFetcher interface {
	GetModifiedFiles(ctx context.Context, repo models.Repo, installationToken int64, fileFetcherOptions github.FileFetcherOptions) ([]string, error)
}

// rootFinder determines which roots were modified in a given event.
type rootFinder interface {
	// FindRoots returns the list of roots that were modified
	// based on modifiedFiles and the repo's config.
	FindRoots(ctx context.Context, config valid.RepoCfg, absRepoDir string, modifiedFiles []string) ([]valid.Project, error)
}

// parserValidator config builds repo specific configurations
type parserValidator interface {
	ParseRepoCfg(absRepoDir string, repoID string) (valid.RepoCfg, error)
}

type RootConfigBuilder struct {
	RepoFetcher     repoFetcher
	HooksRunner     hooksRunner
	ParserValidator parserValidator
	RootFinder      rootFinder
	FileFetcher     fileFetcher
	GlobalCfg       valid.GlobalCfg
	Logger          logging.Logger
	Scope           tally.Scope
}

type BuilderOptions struct {
	RepoFetcherOptions github.RepoFetcherOptions
	FileFetcherOptions github.FileFetcherOptions
}

func (b *RootConfigBuilder) Build(ctx context.Context, repo models.Repo, branch string, sha string, installationToken int64, builderOptions BuilderOptions) ([]*valid.MergedProjectCfg, error) {
	mergedRootCfgs, err := b.build(ctx, repo, branch, sha, installationToken, builderOptions)
	if err != nil {
		b.Scope.Counter(metrics.FilterErrorMetric).Inc(1)
		return nil, err
	}
	if len(mergedRootCfgs) == 0 {
		b.Scope.Counter(metrics.FilterAbsentMetric).Inc(1)
		return mergedRootCfgs, nil
	}
	b.Scope.Counter(metrics.FilterPresentMetric).Inc(1)
	return mergedRootCfgs, nil
}

func (b *RootConfigBuilder) build(ctx context.Context, repo models.Repo, branch string, sha string, installationToken int64, builderOptions BuilderOptions) ([]*valid.MergedProjectCfg, error) {
	// Generate a new filepath location and clone repo into it
	repoDir, cleanup, err := b.RepoFetcher.Fetch(ctx, repo, branch, sha, builderOptions.RepoFetcherOptions)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("creating temporary clone at path: %s", repoDir))
	}
	defer cleanup(ctx, repoDir)

	// Run pre-workflow hooks
	err = b.HooksRunner.Run(repo, repoDir)
	if err != nil {
		return nil, errors.Wrap(err, "running pre-workflow hooks")
	}

	// Fetch files modified in commit
	modifiedFiles, err := b.FileFetcher.GetModifiedFiles(ctx, repo, installationToken, builderOptions.FileFetcherOptions)
	if err != nil {
		return nil, errors.Wrapf(err, "finding modified files: %s", modifiedFiles)
	}

	// Parse repo configs into specific root configs (i.e. roots)
	// TODO: rename project to roots
	var mergedRootCfgs []*valid.MergedProjectCfg
	repoCfg, err := b.ParserValidator.ParseRepoCfg(repoDir, repo.ID())
	if err != nil {
		return nil, errors.Wrapf(err, "parsing %s", config.AtlantisYAMLFilename)
	}
	matchingRoots, err := b.RootFinder.FindRoots(ctx, repoCfg, repoDir, modifiedFiles)
	if err != nil {
		return nil, errors.Wrap(err, "determining roots")
	}
	for _, mr := range matchingRoots {
		mergedRootCfg := b.GlobalCfg.MergeProjectCfg(repo.ID(), mr, repoCfg)
		mergedRootCfgs = append(mergedRootCfgs, &mergedRootCfg)
	}
	return mergedRootCfgs, nil
}
