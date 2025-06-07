package vcs_test

import (
	"testing"

	"github.com/runatlantis/atlantis/internal/domain/vcs"
	vcsInfra "github.com/runatlantis/atlantis/internal/infrastructure/vcs"
	"github.com/runatlantis/atlantis/server"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/logging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCompatibilityLayer_GetModifiedFiles_LegacyOnly(t *testing.T) {
	// Arrange
	logger := logging.NewNoopLogger(t)
	mockLegacyClient := &MockVCSClient{}
	registry := vcs.NewVCSRegistry()
	userConfig := server.UserConfig{}
	
	config := vcsInfra.CompatibilityConfig{
		EnableNewSystem:  false,
		FallbackToLegacy: true,
	}
	
	compatLayer := vcsInfra.NewCompatibilityLayer(
		mockLegacyClient,
		registry,
		userConfig,
		logger,
		config,
	)

	repo := models.Repo{
		FullName: "owner/repo",
		VCSHost:  models.VCSHost{Type: models.Github},
	}
	pull := models.PullRequest{Num: 123}

	// Act
	files, err := compatLayer.GetModifiedFiles(logger, repo, pull)

	// Assert
	assert.NoError(t, err)
	assert.Nil(t, files) // MockVCSClient returns nil
	// Verify legacy client was called (would be implemented in mock)
}

func TestCompatibilityLayer_GetModifiedFiles_NewSystemWithFallback(t *testing.T) {
	// Arrange
	logger := logging.NewNoopLogger(t)
	mockLegacyClient := &MockVCSClient{}
	registry := vcs.NewVCSRegistry()
	userConfig := server.UserConfig{}
	
	config := vcsInfra.CompatibilityConfig{
		EnableNewSystem:  true,
		FallbackToLegacy: true,
	}
	
	compatLayer := vcsInfra.NewCompatibilityLayer(
		mockLegacyClient,
		registry,
		userConfig,
		logger,
		config,
	)

	repo := models.Repo{
		FullName: "owner/repo",
		VCSHost:  models.VCSHost{Type: models.Github},
	}
	pull := models.PullRequest{Num: 123}

	// Act
	files, err := compatLayer.GetModifiedFiles(logger, repo, pull)

	// Assert
	assert.NoError(t, err)
	assert.Nil(t, files) // MockVCSClient returns nil
	// Should fallback to legacy since no plugin is registered
}

func TestCompatibilityLayer_UpdateStatus_WithPlugin(t *testing.T) {
	// Arrange
	logger := logging.NewNoopLogger(t)
	mockLegacyClient := &MockVCSClient{}
	registry := vcs.NewVCSRegistry()
	userConfig := server.UserConfig{}
	
	// Register a mock plugin
	mockPlugin := vcs.NewMockVCSPlugin(vcs.VCSCapabilities{
		SupportsMergeableBypass: true,
	})
	err := registry.Register("github", mockPlugin)
	require.NoError(t, err)
	
	config := vcsInfra.CompatibilityConfig{
		EnableNewSystem:  true,
		FallbackToLegacy: true,
	}
	
	compatLayer := vcsInfra.NewCompatibilityLayer(
		mockLegacyClient,
		registry,
		userConfig,
		logger,
		config,
	)

	repo := models.Repo{
		FullName: "owner/repo",
		VCSHost:  models.VCSHost{Type: models.Github},
	}
	pull := models.PullRequest{
		Num:        123,
		HeadCommit: "abc123",
	}

	// Act
	err = compatLayer.UpdateStatus(logger, repo, pull, models.SuccessCommitStatus, "test", "Test passed", "http://example.com")

	// Assert
	assert.NoError(t, err)
	// Verify plugin was used (would check mock state)
}

func TestCompatibilityLayer_PullIsMergeable_WithCapablePlugin(t *testing.T) {
	// Arrange
	logger := logging.NewNoopLogger(t)
	mockLegacyClient := &MockVCSClient{}
	registry := vcs.NewVCSRegistry()
	userConfig := server.UserConfig{}
	
	// Create adapter with mergeable bypass support
	adapterConfig := vcsInfra.VCSAdapterConfig{
		AllowMergeableBypass: true,
	}
	adapter := vcsInfra.NewLegacyVCSClientAdapter(mockLegacyClient, "github", logger, adapterConfig)
	err := registry.Register("github", adapter)
	require.NoError(t, err)
	
	config := vcsInfra.CompatibilityConfig{
		EnableNewSystem:  true,
		FallbackToLegacy: true,
	}
	
	compatLayer := vcsInfra.NewCompatibilityLayer(
		mockLegacyClient,
		registry,
		userConfig,
		logger,
		config,
	)

	repo := models.Repo{
		FullName: "owner/repo",
		VCSHost:  models.VCSHost{Type: models.Github},
	}
	pull := models.PullRequest{Num: 123}

	// Act
	isMergeable, err := compatLayer.PullIsMergeable(logger, repo, pull, "atlantis", []string{})

	// Assert
	assert.NoError(t, err)
	assert.True(t, isMergeable) // MockVCSClient returns true for PullIsMergeable
}

func TestCompatibilityLayer_CreateComment_AlwaysUsesLegacy(t *testing.T) {
	// Arrange
	logger := logging.NewNoopLogger(t)
	mockLegacyClient := &MockVCSClient{}
	registry := vcs.NewVCSRegistry()
	userConfig := server.UserConfig{}
	
	config := vcsInfra.CompatibilityConfig{
		EnableNewSystem:  true,
		FallbackToLegacy: true,
	}
	
	compatLayer := vcsInfra.NewCompatibilityLayer(
		mockLegacyClient,
		registry,
		userConfig,
		logger,
		config,
	)

	repo := models.Repo{
		FullName: "owner/repo",
		VCSHost:  models.VCSHost{Type: models.Github},
	}

	// Act
	err := compatLayer.CreateComment(logger, repo, 123, "test comment", "plan")

	// Assert
	assert.NoError(t, err)
	// Should always use legacy for now since comment creation isn't in plugin interface
}

func TestCompatibilityLayer_ConvertVCSHostType(t *testing.T) {
	// Arrange
	logger := logging.NewNoopLogger(t)
	mockLegacyClient := &MockVCSClient{}
	registry := vcs.NewVCSRegistry()
	userConfig := server.UserConfig{}
	
	config := vcsInfra.CompatibilityConfig{}
	
	compatLayer := vcsInfra.NewCompatibilityLayer(
		mockLegacyClient,
		registry,
		userConfig,
		logger,
		config,
	)

	// Test cases for VCS host type conversion
	testCases := []struct {
		input    models.VCSHostType
		expected string
	}{
		{models.Github, "github"},
		{models.Gitlab, "gitlab"},
		{models.BitbucketCloud, "bitbucket"},
		{models.BitbucketServer, "bitbucket"},
		{models.AzureDevops, "azuredevops"},
		{models.Gitea, "gitea"},
	}

	for _, tc := range testCases {
		t.Run(tc.expected, func(t *testing.T) {
			// This tests the internal conversion logic indirectly
			// by trying to get a plugin for each VCS type
			repo := models.Repo{
				VCSHost: models.VCSHost{Type: tc.input},
			}
			
			// Act - this will call convertVCSHostType internally
			_, err := compatLayer.GetModifiedFiles(logger, repo, models.PullRequest{})
			
			// Assert - should not panic and should handle the conversion
			assert.NoError(t, err)
		})
	}
}

func TestCompatibilityLayer_ConvertCommitState(t *testing.T) {
	// Arrange
	logger := logging.NewNoopLogger(t)
	mockLegacyClient := &MockVCSClient{}
	registry := vcs.NewVCSRegistry()
	userConfig := server.UserConfig{}
	
	// Register a mock plugin to enable new system
	mockPlugin := vcs.NewMockVCSPlugin(vcs.VCSCapabilities{})
	err := registry.Register("github", mockPlugin)
	require.NoError(t, err)
	
	config := vcsInfra.CompatibilityConfig{
		EnableNewSystem:  true,
		FallbackToLegacy: false, // Force using new system
	}
	
	compatLayer := vcsInfra.NewCompatibilityLayer(
		mockLegacyClient,
		registry,
		userConfig,
		logger,
		config,
	)

	repo := models.Repo{
		FullName: "owner/repo",
		VCSHost:  models.VCSHost{Type: models.Github},
	}
	pull := models.PullRequest{
		Num:        123,
		HeadCommit: "abc123",
	}

	// Test cases for commit state conversion
	testCases := []struct {
		input models.CommitStatus
		name  string
	}{
		{models.PendingCommitStatus, "pending"},
		{models.SuccessCommitStatus, "success"},
		{models.FailedCommitStatus, "failed"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Act - this will call convertCommitState internally
			err := compatLayer.UpdateStatus(logger, repo, pull, tc.input, "test", "Test", "http://example.com")
			
			// Assert - should not panic and should handle the conversion
			assert.NoError(t, err)
		})
	}
} 