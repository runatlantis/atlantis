package vcs

import (
	"context"
	"testing"

	"github.com/runatlantis/atlantis/internal/domain/vcs"
	"github.com/runatlantis/atlantis/server"
	"github.com/runatlantis/atlantis/server/logging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewService(t *testing.T) {
	tests := []struct {
		name      string
		config    *ServiceConfig
		wantError bool
	}{
		{
			name:      "nil config should return error",
			config:    nil,
			wantError: true,
		},
		{
			name: "valid config should create service",
			config: &ServiceConfig{
				DefaultProvider: GitHub,
				EnabledFeatures: map[string]bool{
					"mergeable_bypass": true,
					"team_allowlist":   true,
				},
				UserConfig: server.UserConfig{},
				Logger:     logging.NewNoopLogger(t),
			},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service, err := NewService(tt.config)
			
			if tt.wantError {
				assert.Error(t, err)
				assert.Nil(t, service)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, service)
				assert.Equal(t, tt.config.DefaultProvider, service.defaultProvider)
				assert.Equal(t, tt.config.EnabledFeatures, service.enabledFeatures)
			}
		})
	}
}

func TestService_RegisterPlugin(t *testing.T) {
	service := createTestService(t)
	mockPlugin := vcs.NewMockVCSPlugin(vcs.VCSCapabilities{
		SupportsMergeableBypass: true,
		SupportsTeamAllowlist:   true,
		SupportsGroupAllowlist:  false,
		SupportsCustomFields:    true,
		MaxPageSize:            100,
	})

	tests := []struct {
		name      string
		pluginName string
		plugin    vcs.VCSPlugin
		wantError bool
	}{
		{
			name:       "valid plugin should register successfully",
			pluginName: "github",
			plugin:     mockPlugin,
			wantError:  false,
		},
		{
			name:       "nil plugin should return error",
			pluginName: "gitlab",
			plugin:     nil,
			wantError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.RegisterPlugin(tt.pluginName, tt.plugin)
			
			if tt.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				
				// Verify plugin was registered
				retrievedPlugin, err := service.GetPlugin(ProviderType(tt.pluginName))
				assert.NoError(t, err)
				assert.Equal(t, tt.plugin, retrievedPlugin)
			}
		})
	}
}

func TestService_GetPlugin(t *testing.T) {
	service := createTestService(t)
	mockPlugin := vcs.NewMockVCSPlugin(vcs.VCSCapabilities{
		SupportsMergeableBypass: true,
		SupportsTeamAllowlist:   true,
		MaxPageSize:            100,
	})

	// Register a plugin
	err := service.RegisterPlugin("github", mockPlugin)
	require.NoError(t, err)

	tests := []struct {
		name         string
		providerType ProviderType
		wantPlugin   vcs.VCSPlugin
		wantError    bool
	}{
		{
			name:         "existing provider should return plugin",
			providerType: GitHub,
			wantPlugin:   mockPlugin,
			wantError:    false,
		},
		{
			name:         "non-existing provider should return error",
			providerType: GitLab,
			wantPlugin:   nil,
			wantError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			plugin, err := service.GetPlugin(tt.providerType)
			
			if tt.wantError {
				assert.Error(t, err)
				assert.Nil(t, plugin)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantPlugin, plugin)
			}
		})
	}
}

func TestService_SupportsFeature(t *testing.T) {
	service := createTestService(t)
	mockPlugin := vcs.NewMockVCSPlugin(vcs.VCSCapabilities{
		SupportsMergeableBypass: true,
		SupportsTeamAllowlist:   true,
		SupportsGroupAllowlist:  false,
		SupportsCustomFields:    true,
		MaxPageSize:            100,
	})

	err := service.RegisterPlugin("github", mockPlugin)
	require.NoError(t, err)

	tests := []struct {
		name         string
		providerType ProviderType
		feature      string
		wantSupports bool
		wantError    bool
	}{
		{
			name:         "supported feature should return true",
			providerType: GitHub,
			feature:      "mergeable_bypass",
			wantSupports: true,
			wantError:    false,
		},
		{
			name:         "unsupported feature should return false",
			providerType: GitHub,
			feature:      "group_allowlist",
			wantSupports: false,
			wantError:    false,
		},
		{
			name:         "unknown feature should return false",
			providerType: GitHub,
			feature:      "unknown_feature",
			wantSupports: false,
			wantError:    false,
		},
		{
			name:         "non-existing provider should return error",
			providerType: GitLab,
			feature:      "mergeable_bypass",
			wantSupports: false,
			wantError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			supports, err := service.SupportsFeature(tt.providerType, tt.feature)
			
			if tt.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantSupports, supports)
			}
		})
	}
}

func TestService_IsFeatureEnabled(t *testing.T) {
	service := createTestService(t)

	tests := []struct {
		name        string
		feature     string
		wantEnabled bool
	}{
		{
			name:        "enabled feature should return true",
			feature:     "mergeable_bypass",
			wantEnabled: true,
		},
		{
			name:        "disabled feature should return false",
			feature:     "team_allowlist",
			wantEnabled: false,
		},
		{
			name:        "unknown feature should return false",
			feature:     "unknown_feature",
			wantEnabled: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			enabled := service.IsFeatureEnabled(tt.feature)
			assert.Equal(t, tt.wantEnabled, enabled)
		})
	}
}

func TestService_CheckMergeableBypass(t *testing.T) {
	service := createTestService(t)
	mockPlugin := vcs.NewMockVCSPlugin(vcs.VCSCapabilities{
		SupportsMergeableBypass: true,
		MaxPageSize:            100,
	})

	// Set up the mock to return true for mergeable bypass
	mergeable := true
	mockPlugin.SetMergeableBypassFunc(func(pr *vcs.PullRequest) bool {
		return true
	})

	err := service.RegisterPlugin("github", mockPlugin)
	require.NoError(t, err)

	ctx := context.Background()
	pr := &vcs.PullRequest{Number: 1, Title: "Test PR", Mergeable: &mergeable}

	tests := []struct {
		name         string
		providerType ProviderType
		wantResult   bool
		wantError    bool
	}{
		{
			name:         "should check mergeable bypass successfully",
			providerType: GitHub,
			wantResult:   true,
			wantError:    false,
		},
		{
			name:         "non-existing provider should return error",
			providerType: GitLab,
			wantResult:   false,
			wantError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := service.CheckMergeableBypass(ctx, tt.providerType, pr)
			
			if tt.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantResult, result)
			}
		})
	}
}

func TestService_ValidateTeamMembership(t *testing.T) {
	service := createTestService(t)
	mockPlugin := vcs.NewMockVCSPlugin(vcs.VCSCapabilities{
		SupportsTeamAllowlist: true,
		MaxPageSize:          100,
	})

	// Set up team membership
	user := "testuser"
	teams := []string{"team1", "team2"}
	mockPlugin.AddTeamMembership(user, teams)

	err := service.RegisterPlugin("github", mockPlugin)
	require.NoError(t, err)

	ctx := context.Background()

	// Enable the feature for this test
	service.enabledFeatures["team_allowlist"] = true

	result, err := service.ValidateTeamMembership(ctx, GitHub, user, teams)
	assert.NoError(t, err)
	assert.True(t, result)
}

func TestService_ValidateGroupMembership(t *testing.T) {
	service := createTestService(t)
	mockPlugin := vcs.NewMockVCSPlugin(vcs.VCSCapabilities{
		SupportsGroupAllowlist: true,
		MaxPageSize:           100,
	})

	// Set up group membership
	user := "testuser"
	groups := []string{"group1", "group2"}
	mockPlugin.AddGroupMembership(user, groups)

	err := service.RegisterPlugin("gitlab", mockPlugin)
	require.NoError(t, err)

	ctx := context.Background()

	// Enable the feature for this test
	service.enabledFeatures["group_allowlist"] = true

	result, err := service.ValidateGroupMembership(ctx, GitLab, user, groups)
	assert.NoError(t, err)
	assert.True(t, result)
}

func TestService_HealthCheck(t *testing.T) {
	service := createTestService(t)
	
	healthyPlugin := vcs.NewMockVCSPlugin(vcs.VCSCapabilities{
		SupportsMergeableBypass: true,
		MaxPageSize:            100,
	})
	
	unhealthyPlugin := vcs.NewMockVCSPlugin(vcs.VCSCapabilities{
		SupportsMergeableBypass: true,
		MaxPageSize:            0, // Invalid configuration
	})

	err := service.RegisterPlugin("github", healthyPlugin)
	require.NoError(t, err)
	
	err = service.RegisterPlugin("gitlab", unhealthyPlugin)
	require.NoError(t, err)

	ctx := context.Background()
	results := service.HealthCheck(ctx)

	assert.Len(t, results, 2)
	assert.NoError(t, results[GitHub])
	assert.Error(t, results[GitLab])
	assert.Contains(t, results[GitLab].Error(), "invalid configuration")
}

func TestService_GetFeatureMatrix(t *testing.T) {
	service := createTestService(t)
	
	githubPlugin := vcs.NewMockVCSPlugin(vcs.VCSCapabilities{
		SupportsMergeableBypass: true,
		SupportsTeamAllowlist:   true,
		SupportsGroupAllowlist:  false,
		SupportsCustomFields:    true,
		MaxPageSize:            100,
	})
	
	gitlabPlugin := vcs.NewMockVCSPlugin(vcs.VCSCapabilities{
		SupportsMergeableBypass: false,
		SupportsTeamAllowlist:   false,
		SupportsGroupAllowlist:  true,
		SupportsCustomFields:    false,
		MaxPageSize:            50,
	})

	err := service.RegisterPlugin("github", githubPlugin)
	require.NoError(t, err)
	
	err = service.RegisterPlugin("gitlab", gitlabPlugin)
	require.NoError(t, err)

	matrix := service.GetFeatureMatrix()

	assert.Len(t, matrix, 2)
	
	// Check GitHub features
	githubFeatures := matrix[GitHub]
	assert.True(t, githubFeatures["mergeable_bypass"])
	assert.True(t, githubFeatures["team_allowlist"])
	assert.False(t, githubFeatures["group_allowlist"])
	assert.True(t, githubFeatures["custom_fields"])
	
	// Check GitLab features
	gitlabFeatures := matrix[GitLab]
	assert.False(t, gitlabFeatures["mergeable_bypass"])
	assert.False(t, gitlabFeatures["team_allowlist"])
	assert.True(t, gitlabFeatures["group_allowlist"])
	assert.False(t, gitlabFeatures["custom_fields"])
}

func TestService_ValidateProviderMigration(t *testing.T) {
	service := createTestService(t)
	
	githubPlugin := vcs.NewMockVCSPlugin(vcs.VCSCapabilities{
		SupportsMergeableBypass: true,
		SupportsTeamAllowlist:   true,
		SupportsGroupAllowlist:  false,
		SupportsCustomFields:    true,
		MaxPageSize:            100,
	})
	
	gitlabPlugin := vcs.NewMockVCSPlugin(vcs.VCSCapabilities{
		SupportsMergeableBypass: false,
		SupportsTeamAllowlist:   false,
		SupportsGroupAllowlist:  true,
		SupportsCustomFields:    false,
		MaxPageSize:            50,
	})

	err := service.RegisterPlugin("github", githubPlugin)
	require.NoError(t, err)
	
	err = service.RegisterPlugin("gitlab", gitlabPlugin)
	require.NoError(t, err)

	tests := []struct {
		name      string
		from      ProviderType
		to        ProviderType
		features  []string
		wantError bool
	}{
		{
			name:      "migration with no feature loss should succeed",
			from:      GitLab,
			to:        GitHub,
			features:  []string{"group_allowlist"},
			wantError: false,
		},
		{
			name:      "migration with feature loss should fail",
			from:      GitHub,
			to:        GitLab,
			features:  []string{"mergeable_bypass", "team_allowlist"},
			wantError: true,
		},
		{
			name:      "migration with compatible features should succeed",
			from:      GitHub,
			to:        GitLab,
			features:  []string{"group_allowlist"}, // GitHub doesn't support, GitLab does
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.ValidateProviderMigration(tt.from, tt.to, tt.features)
			
			if tt.wantError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "would lose support for features")
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestService_ListProviders(t *testing.T) {
	service := createTestService(t)
	
	mockPlugin := vcs.NewMockVCSPlugin(vcs.VCSCapabilities{
		MaxPageSize: 100,
	})

	err := service.RegisterPlugin("github", mockPlugin)
	require.NoError(t, err)
	
	err = service.RegisterPlugin("gitlab", mockPlugin)
	require.NoError(t, err)

	providers := service.ListProviders()
	
	assert.Len(t, providers, 2)
	assert.Contains(t, providers, GitHub)
	assert.Contains(t, providers, GitLab)
}

// createTestService creates a service instance for testing
func createTestService(t *testing.T) *Service {
	config := &ServiceConfig{
		DefaultProvider: GitHub,
		EnabledFeatures: map[string]bool{
			"mergeable_bypass": true,
			"team_allowlist":   false,
		},
		UserConfig: server.UserConfig{},
		Logger:     logging.NewNoopLogger(t),
	}

	service, err := NewService(config)
	require.NoError(t, err)
	
	return service
} 