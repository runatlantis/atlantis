package vcs_test

import (
	"testing"

	"github.com/runatlantis/atlantis/internal/domain/vcs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFeatureValidator_ValidateAndWarn_AllFeaturesSupported_NoWarnings(t *testing.T) {
	// Arrange
	registry := vcs.NewVCSRegistry()
	plugin := vcs.NewMockVCSPlugin(vcs.VCSCapabilities{
		SupportsMergeableBypass: true,
		SupportsTeamAllowlist:   true,
		SupportsGroupAllowlist:  true,
	})
	_ = registry.Register("github", plugin)
	
	validator := vcs.NewFeatureValidator(registry)
	config := vcs.FeatureConfig{
		AllowMergeableBypass: true,
		TeamAllowlist:        []string{"devops"},
		GroupAllowlist:       []string{"admin"},
	}

	// Act
	err := validator.ValidateAndWarn("github", config)

	// Assert
	assert.NoError(t, err)
}

func TestFeatureValidator_ValidateAndWarn_UnsupportedFeatures_LogsWarnings(t *testing.T) {
	// Arrange
	registry := vcs.NewVCSRegistry()
	plugin := vcs.NewMockVCSPlugin(vcs.VCSCapabilities{
		SupportsMergeableBypass: false,
		SupportsTeamAllowlist:   false,
		SupportsGroupAllowlist:  false,
	})
	_ = registry.Register("gitlab", plugin)
	
	validator := vcs.NewFeatureValidator(registry)
	config := vcs.FeatureConfig{
		AllowMergeableBypass: true,
		TeamAllowlist:        []string{"devops"},
		GroupAllowlist:       []string{"admin"},
	}

	// Act
	err := validator.ValidateAndWarn("gitlab", config)

	// Assert
	assert.NoError(t, err) // Warnings don't cause errors
}

func TestFeatureValidator_ValidateAndWarn_GitHubCapabilities_ProperSuggestions(t *testing.T) {
	// Arrange
	registry := vcs.NewVCSRegistry()
	plugin := vcs.NewMockVCSPlugin(vcs.VCSCapabilities{
		SupportsMergeableBypass: true,
		SupportsTeamAllowlist:   true,
		SupportsGroupAllowlist:  false, // GitHub doesn't support groups
	})
	_ = registry.Register("github", plugin)
	
	validator := vcs.NewFeatureValidator(registry)
	config := vcs.FeatureConfig{
		GroupAllowlist: []string{"admin"}, // Unsupported on GitHub
	}

	// Act
	err := validator.ValidateAndWarn("github", config)

	// Assert
	assert.NoError(t, err)
}

func TestFeatureValidator_ValidateAndWarn_GitLabCapabilities_ProperSuggestions(t *testing.T) {
	// Arrange
	registry := vcs.NewVCSRegistry()
	plugin := vcs.NewMockVCSPlugin(vcs.VCSCapabilities{
		SupportsMergeableBypass: true,
		SupportsTeamAllowlist:   false, // GitLab doesn't support teams
		SupportsGroupAllowlist:  true,
	})
	_ = registry.Register("gitlab", plugin)
	
	validator := vcs.NewFeatureValidator(registry)
	config := vcs.FeatureConfig{
		TeamAllowlist: []string{"devops"}, // Unsupported on GitLab
	}

	// Act
	err := validator.ValidateAndWarn("gitlab", config)

	// Assert
	assert.NoError(t, err)
}

func TestFeatureValidator_GetEffectiveConfig_FiltersUnsupportedFeatures(t *testing.T) {
	// Arrange
	registry := vcs.NewVCSRegistry()
	plugin := vcs.NewMockVCSPlugin(vcs.VCSCapabilities{
		SupportsMergeableBypass: true,
		SupportsTeamAllowlist:   false,
		SupportsGroupAllowlist:  false,
	})
	_ = registry.Register("vcs", plugin)
	
	validator := vcs.NewFeatureValidator(registry)
	config := vcs.FeatureConfig{
		AllowMergeableBypass: true,
		TeamAllowlist:        []string{"devops"},
		GroupAllowlist:       []string{"admin"},
	}

	// Act
	effectiveConfig, err := validator.GetEffectiveConfig("vcs", config)

	// Assert
	require.NoError(t, err)
	assert.True(t, effectiveConfig.AllowMergeableBypass)  // Supported
	assert.Nil(t, effectiveConfig.TeamAllowlist)         // Filtered out
	assert.Nil(t, effectiveConfig.GroupAllowlist)        // Filtered out
}

func TestFeatureValidator_IsFeatureSupported_KnownFeatures_ReturnsCorrectly(t *testing.T) {
	// Arrange
	registry := vcs.NewVCSRegistry()
	plugin := vcs.NewMockVCSPlugin(vcs.VCSCapabilities{
		SupportsMergeableBypass: true,
		SupportsTeamAllowlist:   false,
		SupportsGroupAllowlist:  true,
		SupportsCustomFields:    false,
	})
	_ = registry.Register("vcs", plugin)
	
	validator := vcs.NewFeatureValidator(registry)

	// Act & Assert
	supported, err := validator.IsFeatureSupported("vcs", "mergeable-bypass")
	assert.NoError(t, err)
	assert.True(t, supported)

	supported, err = validator.IsFeatureSupported("vcs", "team-allowlist")
	assert.NoError(t, err)
	assert.False(t, supported)

	supported, err = validator.IsFeatureSupported("vcs", "group-allowlist")
	assert.NoError(t, err)
	assert.True(t, supported)

	supported, err = validator.IsFeatureSupported("vcs", "custom-fields")
	assert.NoError(t, err)
	assert.False(t, supported)
}

func TestFeatureValidator_IsFeatureSupported_UnknownFeature_ReturnsError(t *testing.T) {
	// Arrange
	registry := vcs.NewVCSRegistry()
	plugin := vcs.NewMockVCSPlugin(vcs.VCSCapabilities{})
	_ = registry.Register("vcs", plugin)
	
	validator := vcs.NewFeatureValidator(registry)

	// Act
	supported, err := validator.IsFeatureSupported("vcs", "unknown-feature")

	// Assert
	assert.Error(t, err)
	assert.False(t, supported)
	assert.Contains(t, err.Error(), "unknown feature")
}

func TestFeatureValidator_ValidateAndWarn_NonExistentVCS_ReturnsError(t *testing.T) {
	// Arrange
	registry := vcs.NewVCSRegistry()
	validator := vcs.NewFeatureValidator(registry)
	config := vcs.FeatureConfig{}

	// Act
	err := validator.ValidateAndWarn("nonexistent", config)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestFeatureValidator_GetEffectiveConfig_NonExistentVCS_ReturnsError(t *testing.T) {
	// Arrange
	registry := vcs.NewVCSRegistry()
	validator := vcs.NewFeatureValidator(registry)
	config := vcs.FeatureConfig{}

	// Act
	_, err := validator.GetEffectiveConfig("nonexistent", config)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
} 