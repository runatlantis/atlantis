package event

import (
	"context"
	"fmt"
	"strings"

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
	Run(ctx context.Context, repo models.Repo, repoDir string) error
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

type ModifiedRootsStrategy struct {
	FileFetcher fileFetcher
	RootFinder  rootFinder
}

func (s *ModifiedRootsStrategy) FindMatches(ctx context.Context, config valid.RepoCfg, repo *LocalRepo, installationToken int64) ([]valid.Project, error) {
	// Fetch files modified in commit
	modifiedFiles, err := s.FileFetcher.GetModifiedFiles(ctx, repo.RepoCommit.Repo, installationToken, github.FileFetcherOptions{
		PRNum: repo.RepoCommit.OptionalPRNum,
		Sha:   repo.RepoCommit.Sha,
	})
	if err != nil {
		return nil, errors.Wrapf(err, "finding modified files: %s", modifiedFiles)
	}

	matchingRoots, err := s.RootFinder.FindRoots(ctx, config, repo.Dir, modifiedFiles)
	if err != nil {
		return nil, errors.Wrap(err, "determining roots")
	}

	return matchingRoots, nil
}

type RepoCommit struct {
	Repo          models.Repo
	Branch        string
	Sha           string
	OptionalPRNum int
}

type LocalRepo struct {
	*RepoCommit
	Dir string
}

type RootConfigBuilder struct {
	RepoFetcher     repoFetcher
	HooksRunner     hooksRunner
	ParserValidator parserValidator
	Strategy        *ModifiedRootsStrategy
	GlobalCfg       valid.GlobalCfg
	Logger          logging.Logger
	Scope           tally.Scope
}

type BuilderOptions struct {
	RepoFetcherOptions *github.RepoFetcherOptions
	RootNames          []string
}

func (b *RootConfigBuilder) Build(ctx context.Context, commit *RepoCommit, installationToken int64, opts ...BuilderOptions) ([]*valid.MergedProjectCfg, error) {
	mergedRootCfgs, err := b.build(ctx, commit, installationToken, opts...)
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

func (b *RootConfigBuilder) build(ctx context.Context, commit *RepoCommit, installationToken int64, opts ...BuilderOptions) ([]*valid.MergedProjectCfg, error) {
	var repoOptions github.RepoFetcherOptions
	var rootNames []string
	for _, o := range opts {
		if o.RepoFetcherOptions != nil {
			repoOptions = *o.RepoFetcherOptions
		}

		if len(o.RootNames) > 0 {
			rootNames = o.RootNames
		}
	}

	// Generate a new filepath location and clone repo into it
	repoDir, cleanup, err := b.RepoFetcher.Fetch(ctx, commit.Repo, commit.Branch, commit.Sha, repoOptions)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("creating temporary clone at path: %s", repoDir))
	}
	defer cleanup(ctx, repoDir)

	// ideally we should pass this around instead of repoDir separately
	localRepo := &LocalRepo{
		RepoCommit: commit,
		Dir:        repoDir,
	}

	// Run pre-workflow hooks
	err = b.HooksRunner.Run(ctx, localRepo.Repo, localRepo.Dir)
	if err != nil {
		return nil, errors.Wrap(err, "running pre-workflow hooks")
	}

	// Parse repo configs into specific root configs (i.e. roots)
	// TODO: rename project to roots
	var mergedRootCfgs []*valid.MergedProjectCfg
	repoCfg, err := b.ParserValidator.ParseRepoCfg(localRepo.Dir, localRepo.Repo.ID())
	if err != nil {
		return nil, errors.Wrapf(err, "parsing %s", config.AtlantisYAMLFilename)
	}

	matchingRoots, err := b.getMatchingRoots(ctx, repoCfg, localRepo, installationToken, rootNames)
	if err != nil {
		return nil, errors.Wrap(err, "getting matching roots")
	}

	for _, mr := range matchingRoots {
		mergedRootCfg := b.GlobalCfg.MergeProjectCfg(localRepo.Repo.ID(), mr, repoCfg)
		mergedRootCfgs = append(mergedRootCfgs, &mergedRootCfg)
	}
	return mergedRootCfgs, nil
}

func (b *RootConfigBuilder) getMatchingRoots(ctx context.Context, config valid.RepoCfg, repo *LocalRepo, installationToken int64, rootNames []string) ([]valid.Project, error) {
	if len(rootNames) > 0 {
		return b.validateAndGetRoots(config, rootNames)
	}

	return b.Strategy.FindMatches(ctx, config, repo, installationToken)
}

func (b *RootConfigBuilder) validateAndGetRoots(config valid.RepoCfg, rootNames []string) ([]valid.Project, error) {
	rootSet := make(map[string]valid.Project)

	for _, p := range config.Projects {

		// Project Name is not guaranteed in upstream atlantis but is guaranteed in ours so its OK to dereference
		rootSet[*p.Name] = p
	}

	var results []valid.Project
	var invalidRootNames []string
	for _, n := range rootNames {
		if p, ok := rootSet[n]; ok {
			results = append(results, p)
			continue
		}

		invalidRootNames = append(invalidRootNames, n)
	}

	if len(invalidRootNames) > 0 {
		return results, fmt.Errorf("%s are invalid root names", strings.Join(invalidRootNames, ", "))
	}

	return results, nil
}
