package request_test

import (
	"context"
	"testing"

	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/neptune/gateway/api/middleware"
	"github.com/runatlantis/atlantis/server/neptune/gateway/api/request"
	"github.com/runatlantis/atlantis/server/neptune/gateway/api/request/external"
	"github.com/runatlantis/atlantis/server/vcs/provider/github"
	"github.com/stretchr/testify/assert"
)

func TestDeployCoverter_Success(t *testing.T) {
	var token int64 = 1
	owner := "nish"
	repo := "repo"
	branch := "main"
	revision := "123"
	username := "user"

	expectedRepo := models.Repo{
		Name:          repo,
		Owner:         owner,
		DefaultBranch: branch,
	}

	expectedResult := request.Deploy{
		RootNames:         []string{"root1"},
		Repo:              expectedRepo,
		Branch:            branch,
		Revision:          revision,
		InstallationToken: token,
		User: models.User{
			Username: username,
		},
	}

	expectedBranch := github.Branch{
		Name:     branch,
		Revision: revision,
	}
	rRetriever := &repoRetriever{
		expectedT:     t,
		expectedToken: token,
		expectedOwner: owner,
		expectedRepo:  repo,
		resultRepo:    expectedRepo,
	}

	bRetriever := &branchRetriever{
		expectedT:      t,
		expectedToken:  token,
		expectedOwner:  owner,
		expectedRepo:   repo,
		expectedBranch: branch,
		resultBranch:   expectedBranch,
	}

	iRetriever := &installationRetriever{
		expectedT:   t,
		expectedOrg: owner,
		resultInstallation: github.Installation{
			Token: token,
		},
	}

	subject := &request.DeployConverter{
		InstallationRetriever: iRetriever,
		RepoRetriever:         rRetriever,
		BranchRetriever:       bRetriever,
	}

	result, err := subject.Convert(context.WithValue(context.Background(), middleware.UsernameContextKey, username), external.DeployRequest{
		Roots: []string{
			"root1",
		},
		Repo: external.Repo{
			Owner: owner,
			Name:  repo,
		},
	})

	assert.NoError(t, err)
	assert.Equal(t, expectedResult, result)
}

func TestDeployCoverter_UsernameMissing(t *testing.T) {
	owner := "nish"
	repo := "repo"

	rRetriever := &repoRetriever{}
	bRetriever := &branchRetriever{}
	iRetriever := &installationRetriever{}

	subject := &request.DeployConverter{
		InstallationRetriever: iRetriever,
		RepoRetriever:         rRetriever,
		BranchRetriever:       bRetriever,
	}

	_, err := subject.Convert(context.Background(), external.DeployRequest{
		Roots: []string{
			"root1",
		},
		Repo: external.Repo{
			Owner: owner,
			Name:  repo,
		},
	})

	assert.Error(t, err)
}

func TestDeployCoverter_DefaultBranchMissing(t *testing.T) {
	var token int64 = 1
	owner := "nish"
	repo := "repo"
	username := "user"

	expectedRepo := models.Repo{
		Name:  repo,
		Owner: owner,
	}

	rRetriever := &repoRetriever{
		expectedT:     t,
		expectedToken: token,
		expectedOwner: owner,
		expectedRepo:  repo,
		resultRepo:    expectedRepo,
	}

	bRetriever := &branchRetriever{}
	iRetriever := &installationRetriever{
		expectedT:   t,
		expectedOrg: owner,
		resultInstallation: github.Installation{
			Token: token,
		},
	}

	subject := &request.DeployConverter{
		InstallationRetriever: iRetriever,
		RepoRetriever:         rRetriever,
		BranchRetriever:       bRetriever,
	}

	_, err := subject.Convert(context.WithValue(context.Background(), middleware.UsernameContextKey, username), external.DeployRequest{
		Roots: []string{
			"root1",
		},
		Repo: external.Repo{
			Owner: owner,
			Name:  repo,
		},
	})

	assert.Error(t, err)
}

type repoRetriever struct {
	expectedT     *testing.T
	expectedToken int64
	expectedOwner string
	expectedRepo  string

	resultRepo models.Repo
}

func (r *repoRetriever) Get(ctx context.Context, installationToken int64, owner, repo string) (models.Repo, error) {
	assert.Equal(r.expectedT, r.expectedToken, installationToken)
	assert.Equal(r.expectedT, r.expectedOwner, owner)
	assert.Equal(r.expectedT, r.expectedRepo, repo)

	return r.resultRepo, nil
}

type branchRetriever struct {
	expectedT      *testing.T
	expectedToken  int64
	expectedOwner  string
	expectedRepo   string
	expectedBranch string

	resultBranch github.Branch
}

func (r *branchRetriever) GetBranch(ctx context.Context, installationToken int64, owner, repo, branch string, followRedirects bool) (github.Branch, error) {
	assert.Equal(r.expectedT, r.expectedToken, installationToken)
	assert.Equal(r.expectedT, r.expectedOwner, owner)
	assert.Equal(r.expectedT, r.expectedRepo, repo)
	assert.Equal(r.expectedT, r.expectedBranch, branch)

	return r.resultBranch, nil

}

type installationRetriever struct {
	expectedT   *testing.T
	expectedOrg string

	resultInstallation github.Installation
}

func (r *installationRetriever) FindOrganizationInstallation(ctx context.Context, org string) (github.Installation, error) {
	assert.Equal(r.expectedT, r.expectedOrg, org)

	return r.resultInstallation, nil
}
