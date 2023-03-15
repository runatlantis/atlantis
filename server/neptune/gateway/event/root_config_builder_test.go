package event_test

import (
	"context"
	"errors"
	"github.com/runatlantis/atlantis/server/vcs/provider/github"
	"github.com/uber-go/tally/v4"
	"testing"

	"github.com/runatlantis/atlantis/server/core/config/valid"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/logging"
	"github.com/runatlantis/atlantis/server/neptune/gateway/event"
	"github.com/stretchr/testify/assert"
)

var rcb event.RootConfigBuilder
var globalCfg valid.GlobalCfg
var expectedErr = errors.New("some error") //nolint:revive // error name is fine for testing purposes

func setupTesting(t *testing.T) {
	globalCfg = valid.NewGlobalCfg("somedir")
	// creates a default PCB to used in each test; individual tests mutate a specific field to test certain functionalities
	rcb = event.RootConfigBuilder{
		RepoFetcher:     &mockRepoFetcher{},
		HooksRunner:     &mockHooksRunner{},
		ParserValidator: &mockParserValidator{},
		Strategy: &event.ModifiedRootsStrategy{
			RootFinder:  &mockRootFinder{},
			FileFetcher: &mockFileFetcher{},
		},
		GlobalCfg: globalCfg,
		Logger:    logging.NewNoopCtxLogger(t),
		Scope:     tally.NewTestScope("test", map[string]string{}),
	}
}

func TestRootConfigBuilder_Success(t *testing.T) {
	repo := models.Repo{
		FullName: "nish/repo",
	}

	commit := &event.RepoCommit{
		Repo: repo,
		Sha:  "1234",
	}
	setupTesting(t)
	projects := []valid.Project{
		{
			Name: &commit.Repo.FullName,
		},
	}
	rcb.Strategy.RootFinder = &mockRootFinder{
		ConfigProjects: projects,
	}
	projCfg := globalCfg.MergeProjectCfg(commit.Repo.ID(), projects[0], valid.RepoCfg{})
	expProjectConfigs := []*valid.MergedProjectCfg{
		&projCfg,
	}

	projectConfigs, err := rcb.Build(context.Background(), commit, 2)
	assert.NoError(t, err)
	assert.Equal(t, expProjectConfigs, projectConfigs)
}

func TestRootConfigBuilder_Success_explicitRoots(t *testing.T) {
	repo := models.Repo{
		FullName: "nish/repo",
	}

	commit := &event.RepoCommit{
		Repo: repo,
		Sha:  "1234",
	}
	setupTesting(t)
	root := testRoot
	projects := []valid.Project{
		{
			Name: &root,
		},
	}
	rcb.ParserValidator = &mockParserValidator{
		repoCfg: valid.RepoCfg{
			Projects: projects,
		},
	}
	rootFinder := &mockRootFinder{
		ConfigProjects: projects,
	}

	filefetcher := &mockFileFetcher{}
	rcb.Strategy.RootFinder = rootFinder
	rcb.Strategy.FileFetcher = filefetcher

	projCfg := globalCfg.MergeProjectCfg(commit.Repo.ID(), projects[0], valid.RepoCfg{})
	expProjectConfigs := []*valid.MergedProjectCfg{
		&projCfg,
	}

	projectConfigs, err := rcb.Build(context.Background(), commit, 2, event.BuilderOptions{
		RootNames: []string{root},
	})
	assert.NoError(t, err)
	assert.Equal(t, expProjectConfigs, projectConfigs)
	assert.False(t, rootFinder.called)
	assert.False(t, filefetcher.called)
}

func TestRootConfigBuilder_Success_explicitRoots_invalid(t *testing.T) {
	repo := models.Repo{
		FullName: "nish/repo",
	}

	commit := &event.RepoCommit{
		Repo: repo,
		Sha:  "1234",
	}
	setupTesting(t)
	root := testRoot
	projects := []valid.Project{
		{
			Name: &root,
		},
	}
	rcb.ParserValidator = &mockParserValidator{
		repoCfg: valid.RepoCfg{
			Projects: projects,
		},
	}

	_, err := rcb.Build(context.Background(), commit, 2, event.BuilderOptions{
		RootNames: []string{"another root"},
	})
	assert.Error(t, err)
}

func TestRootConfigBuilder_DetermineRootsError(t *testing.T) {
	setupTesting(t)
	repo := models.Repo{
		FullName: "nish/repo",
	}

	commit := &event.RepoCommit{
		Repo: repo,
		Sha:  "1234",
	}
	mockRootFinder := &mockRootFinder{
		error: expectedErr,
	}
	rcb.Strategy.RootFinder = mockRootFinder
	projectConfigs, err := rcb.Build(context.Background(), commit, 2)
	assert.Error(t, err)
	assert.Empty(t, projectConfigs)

}

func TestRootConfigBuilder_ParserValidatorParseError(t *testing.T) {
	setupTesting(t)
	repo := models.Repo{
		FullName: "nish/repo",
	}

	commit := &event.RepoCommit{
		Repo: repo,
		Sha:  "1234",
	}
	mockParserValidator := &mockParserValidator{
		error: expectedErr,
	}
	rcb.ParserValidator = mockParserValidator
	projectConfigs, err := rcb.Build(context.Background(), commit, 2)
	assert.Error(t, err)
	assert.Empty(t, projectConfigs)

}

func TestRootConfigBuilder_GetModifiedFilesError(t *testing.T) {
	setupTesting(t)
	repo := models.Repo{
		FullName: "nish/repo",
	}

	commit := &event.RepoCommit{
		Repo: repo,
		Sha:  "1234",
	}
	rcb.Strategy.FileFetcher = &mockFileFetcher{
		error: expectedErr,
	}
	projectConfigs, err := rcb.Build(context.Background(), commit, 2)
	assert.Error(t, err)
	assert.Empty(t, projectConfigs)
}

func TestRootConfigBuilder_CloneError(t *testing.T) {
	setupTesting(t)
	repo := models.Repo{
		FullName: "nish/repo",
	}

	commit := &event.RepoCommit{
		Repo: repo,
		Sha:  "1234",
	}
	rcb.RepoFetcher = &mockRepoFetcher{
		cloneError: expectedErr,
	}
	projectConfigs, err := rcb.Build(context.Background(), commit, 2)
	assert.Error(t, err)
	assert.Empty(t, projectConfigs)

}

func TestRootConfigBuilder_HooksRunnerError(t *testing.T) {
	setupTesting(t)
	repo := models.Repo{
		FullName: "nish/repo",
	}

	commit := &event.RepoCommit{
		Repo: repo,
		Sha:  "1234",
	}
	mockHooksRunner := &mockHooksRunner{
		error: expectedErr,
	}
	rcb.HooksRunner = mockHooksRunner
	projectConfigs, err := rcb.Build(context.Background(), commit, 2)
	assert.Error(t, err)
	assert.Empty(t, projectConfigs)

}

// Mock implementations

type mockRepoFetcher struct {
	cloneError error
}

func (r *mockRepoFetcher) Fetch(_ context.Context, _ models.Repo, _ string, _ string, _ github.RepoFetcherOptions) (string, func(ctx context.Context, filePath string), error) {
	return "", func(ctx context.Context, filePath string) {}, r.cloneError
}

type mockHooksRunner struct {
	error error
}

func (h *mockHooksRunner) Run(_ context.Context, _ models.Repo, _ string) error {
	return h.error
}

type mockFileFetcher struct {
	called bool
	error  error
}

func (f *mockFileFetcher) GetModifiedFiles(_ context.Context, _ models.Repo, _ int64, _ github.FileFetcherOptions) ([]string, error) {
	f.called = true
	return nil, f.error
}

type mockRootFinder struct {
	called         bool
	ConfigProjects []valid.Project
	error          error
}

func (m *mockRootFinder) FindRoots(_ context.Context, _ valid.RepoCfg, _ string, _ []string) ([]valid.Project, error) {
	m.called = true
	return m.ConfigProjects, m.error
}

type mockParserValidator struct {
	repoCfg valid.RepoCfg
	error   error
}

func (v *mockParserValidator) ParseRepoCfg(_ string, _ string) (valid.RepoCfg, error) {
	return v.repoCfg, v.error
}
