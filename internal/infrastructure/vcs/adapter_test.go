package vcs_test

import (
	"context"
	"testing"

	"github.com/runatlantis/atlantis/internal/domain/vcs"
	vcsInfra "github.com/runatlantis/atlantis/internal/infrastructure/vcs"
	"github.com/runatlantis/atlantis/server/events/models"
	vcsClient "github.com/runatlantis/atlantis/server/events/vcs"
	"github.com/runatlantis/atlantis/server/logging"
	"github.com/stretchr/testify/assert"
)

func TestLegacyVCSClientAdapter_GetRepository_Success(t *testing.T) {
	// Arrange
	logger := logging.NewNoopLogger(t)
	mockClient := &vcsClient.NotConfiguredVCSClient{}
	config := vcsInfra.VCSAdapterConfig{}
	adapter := vcsInfra.NewLegacyVCSClientAdapter(mockClient, "github", logger, config)

	// Act
	repo, err := adapter.GetRepository(context.Background(), "owner", "repo")

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, "owner/repo", repo.FullName)
	assert.Equal(t, "owner", repo.Owner)
	assert.Equal(t, "repo", repo.Name)
	assert.Contains(t, repo.HTMLURL, "github.com/owner/repo")
}

func TestLegacyVCSClientAdapter_GetPullRequest_Success(t *testing.T) {
	// Arrange
	logger := logging.NewNoopLogger(t)
	mockClient := &vcsClient.NotConfiguredVCSClient{}
	config := vcsInfra.VCSAdapterConfig{}
	adapter := vcsInfra.NewLegacyVCSClientAdapter(mockClient, "github", logger, config)

	repo := vcs.Repository{
		FullName: "owner/repo",
		HTMLURL:  "https://github.com/owner/repo",
	}

	// Act
	pr, err := adapter.GetPullRequest(context.Background(), repo, 123)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, 123, pr.Number)
	assert.Equal(t, vcs.PullRequestOpen, pr.State)
	assert.Contains(t, pr.URL, "/pull/123")
}

func TestLegacyVCSClientAdapter_Capabilities_GitHub(t *testing.T) {
	// Arrange
	logger := logging.NewNoopLogger(t)
	mockClient := &vcsClient.NotConfiguredVCSClient{}
	config := vcsInfra.VCSAdapterConfig{}
	adapter := vcsInfra.NewLegacyVCSClientAdapter(mockClient, "github", logger, config)

	// Act
	capabilities := adapter.Capabilities()

	// Assert
	assert.True(t, capabilities.SupportsMergeableBypass)
	assert.True(t, capabilities.SupportsTeamAllowlist)
	assert.False(t, capabilities.SupportsGroupAllowlist)
	assert.Equal(t, 100, capabilities.MaxPageSize)
}

func TestLegacyVCSClientAdapter_Capabilities_GitLab(t *testing.T) {
	// Arrange
	logger := logging.NewNoopLogger(t)
	mockClient := &vcsClient.NotConfiguredVCSClient{}
	config := vcsInfra.VCSAdapterConfig{}
	adapter := vcsInfra.NewLegacyVCSClientAdapter(mockClient, "gitlab", logger, config)

	// Act
	capabilities := adapter.Capabilities()

	// Assert
	assert.True(t, capabilities.SupportsMergeableBypass)
	assert.False(t, capabilities.SupportsTeamAllowlist)
	assert.True(t, capabilities.SupportsGroupAllowlist)
	assert.Equal(t, 100, capabilities.MaxPageSize)
}

func TestLegacyVCSClientAdapter_Capabilities_AzureDevOps(t *testing.T) {
	// Arrange
	logger := logging.NewNoopLogger(t)
	mockClient := &vcsClient.NotConfiguredVCSClient{}
	config := vcsInfra.VCSAdapterConfig{}
	adapter := vcsInfra.NewLegacyVCSClientAdapter(mockClient, "azuredevops", logger, config)

	// Act
	capabilities := adapter.Capabilities()

	// Assert
	assert.False(t, capabilities.SupportsMergeableBypass)
	assert.False(t, capabilities.SupportsTeamAllowlist)
	assert.False(t, capabilities.SupportsGroupAllowlist)
	assert.True(t, capabilities.SupportsCustomFields)
	assert.Equal(t, 100, capabilities.MaxPageSize)
}

func TestLegacyVCSClientAdapter_CheckMergeableBypass_Disabled(t *testing.T) {
	// Arrange
	logger := logging.NewNoopLogger(t)
	mockClient := &vcsClient.NotConfiguredVCSClient{}
	config := vcsInfra.VCSAdapterConfig{
		AllowMergeableBypass: false,
	}
	adapter := vcsInfra.NewLegacyVCSClientAdapter(mockClient, "github", logger, config)

	pr := &vcs.PullRequest{Number: 123}

	// Act
	canBypass, err := adapter.CheckMergeableBypass(context.Background(), pr)

	// Assert
	assert.NoError(t, err)
	assert.False(t, canBypass)
}

func TestLegacyVCSClientAdapter_CheckMergeableBypass_NonGitHub(t *testing.T) {
	// Arrange
	logger := logging.NewNoopLogger(t)
	mockClient := &vcsClient.NotConfiguredVCSClient{}
	config := vcsInfra.VCSAdapterConfig{
		AllowMergeableBypass: true,
	}
	adapter := vcsInfra.NewLegacyVCSClientAdapter(mockClient, "gitlab", logger, config)

	pr := &vcs.PullRequest{Number: 123}

	// Act
	canBypass, err := adapter.CheckMergeableBypass(context.Background(), pr)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "mergeable bypass only supported for GitHub")
	assert.False(t, canBypass)
}

func TestLegacyVCSClientAdapter_ValidateTeamMembership_NoRestriction(t *testing.T) {
	// Arrange
	logger := logging.NewNoopLogger(t)
	mockClient := &vcsClient.NotConfiguredVCSClient{}
	config := vcsInfra.VCSAdapterConfig{
		TeamAllowlist: nil, // No restrictions
	}
	adapter := vcsInfra.NewLegacyVCSClientAdapter(mockClient, "github", logger, config)

	// Act
	isValid, err := adapter.ValidateTeamMembership(context.Background(), "user123", []string{"team1"})

	// Assert
	assert.NoError(t, err)
	assert.True(t, isValid) // Should pass when no restrictions
}

func TestLegacyVCSClientAdapter_ValidateGroupMembership_NonGitLab(t *testing.T) {
	// Arrange
	logger := logging.NewNoopLogger(t)
	mockClient := &vcsClient.NotConfiguredVCSClient{}
	config := vcsInfra.VCSAdapterConfig{
		GroupAllowlist: []string{"group1"},
	}
	adapter := vcsInfra.NewLegacyVCSClientAdapter(mockClient, "github", logger, config)

	// Act
	isValid, err := adapter.ValidateGroupMembership(context.Background(), "user123", []string{"group1"})

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "group membership validation only supported for GitLab")
	assert.False(t, isValid)
}

func TestLegacyVCSClientAdapter_ValidateGroupMembership_NoRestriction(t *testing.T) {
	// Arrange
	logger := logging.NewNoopLogger(t)
	mockClient := &vcsClient.NotConfiguredVCSClient{}
	config := vcsInfra.VCSAdapterConfig{
		GroupAllowlist: nil, // No restrictions
	}
	adapter := vcsInfra.NewLegacyVCSClientAdapter(mockClient, "gitlab", logger, config)

	// Act
	isValid, err := adapter.ValidateGroupMembership(context.Background(), "user123", []string{"group1"})

	// Assert
	assert.NoError(t, err)
	assert.True(t, isValid) // Should pass when no restrictions
}

func TestLegacyVCSClientAdapter_CreateCommitStatus_Success(t *testing.T) {
	// Arrange
	logger := logging.NewNoopLogger(t)
	mockClient := &MockVCSClient{}
	config := vcsInfra.VCSAdapterConfig{}
	adapter := vcsInfra.NewLegacyVCSClientAdapter(mockClient, "github", logger, config)

	repo := vcs.Repository{
		FullName: "owner/repo",
		Owner:    "owner",
		Name:     "repo",
		CloneURL: "https://github.com/owner/repo.git",
	}

	status := vcs.CommitStatus{
		State:       vcs.CommitSuccess,
		Description: "Tests passed",
		Context:     "ci/tests",
		TargetURL:   "https://ci.example.com/build/123",
	}

	// Act
	err := adapter.CreateCommitStatus(context.Background(), repo, "abc123", status)

	// Assert
	assert.NoError(t, err)
	assert.True(t, mockClient.UpdateStatusCalled)
}

// MockVCSClient is a mock implementation for testing
type MockVCSClient struct {
	UpdateStatusCalled bool
}

func (m *MockVCSClient) GetModifiedFiles(logger logging.SimpleLogging, repo models.Repo, pull models.PullRequest) ([]string, error) {
	return nil, nil
}

func (m *MockVCSClient) CreateComment(logger logging.SimpleLogging, repo models.Repo, pullNum int, comment string, command string) error {
	return nil
}

func (m *MockVCSClient) ReactToComment(logger logging.SimpleLogging, repo models.Repo, pullNum int, commentID int64, reaction string) error {
	return nil
}

func (m *MockVCSClient) HidePrevCommandComments(logger logging.SimpleLogging, repo models.Repo, pullNum int, command string, dir string) error {
	return nil
}

func (m *MockVCSClient) PullIsApproved(logger logging.SimpleLogging, repo models.Repo, pull models.PullRequest) (models.ApprovalStatus, error) {
	return models.ApprovalStatus{}, nil
}

func (m *MockVCSClient) PullIsMergeable(logger logging.SimpleLogging, repo models.Repo, pull models.PullRequest, vcsstatusname string, ignoreVCSStatusNames []string) (bool, error) {
	return true, nil
}

func (m *MockVCSClient) UpdateStatus(logger logging.SimpleLogging, repo models.Repo, pull models.PullRequest, state models.CommitStatus, src string, description string, url string) error {
	m.UpdateStatusCalled = true
	return nil
}

func (m *MockVCSClient) DiscardReviews(logger logging.SimpleLogging, repo models.Repo, pull models.PullRequest) error {
	return nil
}

func (m *MockVCSClient) MergePull(logger logging.SimpleLogging, pull models.PullRequest, pullOptions models.PullRequestOptions) error {
	return nil
}

func (m *MockVCSClient) MarkdownPullLink(pull models.PullRequest) (string, error) {
	return "", nil
}

func (m *MockVCSClient) GetTeamNamesForUser(logger logging.SimpleLogging, repo models.Repo, user models.User) ([]string, error) {
	return []string{"team1", "team2"}, nil
}

func (m *MockVCSClient) SupportsSingleFileDownload(repo models.Repo) bool {
	return false
}

func (m *MockVCSClient) GetFileContent(logger logging.SimpleLogging, pull models.PullRequest, fileName string) (bool, []byte, error) {
	return false, nil, nil
}

func (m *MockVCSClient) GetCloneURL(logger logging.SimpleLogging, vcsHostType models.VCSHostType, repo string) (string, error) {
	return "", nil
}

func (m *MockVCSClient) GetPullLabels(logger logging.SimpleLogging, repo models.Repo, pull models.PullRequest) ([]string, error) {
	return nil, nil
} 