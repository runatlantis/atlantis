# VCS Features Migration Guide

This guide explains how to migrate from VCS-specific flags to the new plugin-based architecture addressing [issue #5574](https://github.com/runatlantis/atlantis/issues/5574).

## Problem Statement

Currently, Atlantis has VCS-specific flags like:
- `--gh-allow-mergeable-bypass-apply`
- `--gh-team-allowlist`
- `--gitlab-group-allowlist`
- `--gitea-page-size`

These flags create several issues:
1. **Tight Coupling**: Features are tied to specific VCS implementations
2. **Inconsistent Interface**: Each VCS has different flag patterns
3. **Limited Extensibility**: Hard to add features to new VCS providers
4. **Configuration Confusion**: Unclear separation between VCS config and features

## Solution: Plugin-Based Architecture

### Phase 1: Define Interfaces (Week 1)

Create clear abstractions for VCS capabilities:

```go
// Before: VCS-specific flags everywhere
if atlantisConfig.GHAllowMergeableBypass && vcs.Type == "github" {
    // GitHub-specific logic
}

// After: Plugin-based with capability detection
if vcsPlugin.Capabilities().SupportsMergeableBypass && config.AllowMergeableBypass {
    allowed, err := vcsPlugin.CheckMergeableBypass(ctx, pr)
}
```

### Phase 2: Implement Feature Validation (Week 2)

Add runtime validation for unsupported features:

```go
validator := vcs.NewFeatureValidator(vcsRegistry)
if err := validator.ValidateFeatures("gitlab", config); err != nil {
    log.Warn("Feature not supported: %v", err)
}
```

### Phase 3: Create VCS Plugins (Weeks 3-4)

Implement each VCS as a plugin:

```go
// GitHub Plugin
type GitHubPlugin struct {
    client *github.Client
}

func (g *GitHubPlugin) Capabilities() vcs.VCSCapabilities {
    return vcs.VCSCapabilities{
        SupportsMergeableBypass: true,
        SupportsTeamAllowlist:   true,
        SupportsGroupAllowlist:  false, // GitHub uses teams, not groups
        MaxPageSize:            100,
    }
}

// GitLab Plugin  
type GitLabPlugin struct {
    client *gitlab.Client
}

func (g *GitLabPlugin) Capabilities() vcs.VCSCapabilities {
    return vcs.VCSCapabilities{
        SupportsMergeableBypass: true,
        SupportsTeamAllowlist:   false, // GitLab uses groups, not teams
        SupportsGroupAllowlist:  true,
        MaxPageSize:            100,
    }
}
```

### Phase 4: Update Configuration (Week 5)

Transform configuration from VCS-specific to feature-based:

```yaml
# OLD: VCS-specific flags
server_config:
  gh_allow_mergeable_bypass: true
  gh_team_allowlist: ["devops"]
  gitlab_group_allowlist: ["administrators"]

# NEW: Separated concerns
vcs_config:
  github:
    hostname: "github.com"
    token: "${GITHUB_TOKEN}"
  gitlab:
    url: "https://gitlab.com"
    token: "${GITLAB_TOKEN}"

feature_config:
  allow_mergeable_bypass: true
  team_allowlist: ["devops"]        # Works with GitHub
  group_allowlist: ["administrators"] # Works with GitLab
```

### Phase 5: Migration Strategy (Weeks 6-8)

#### Backward Compatibility

Keep old flags working with deprecation warnings:

```go
// Migration layer
if config.GHAllowMergeableBypass {
    log.Warn("--gh-allow-mergeable-bypass is deprecated, use --allow-mergeable-bypass")
    newConfig.AllowMergeableBypass = true
}
```

#### Feature Detection

Warn users about unsupported features:

```bash
# Example output
WARN: team_allowlist is not supported by GitLab, consider using group_allowlist instead
WARN: group_allowlist is not supported by GitHub, teams are used automatically
```

## Benefits

1. **Consistency**: All VCS providers use the same feature interface
2. **Extensibility**: Easy to add new VCS providers or features  
3. **Testability**: Mock plugins for unit testing
4. **Clarity**: Clear separation between VCS configuration and feature flags
5. **Future-Proof**: Plugin architecture allows for external VCS providers

## Testing Strategy

```go
// Test with different VCS capabilities
func TestFeatureValidation(t *testing.T) {
    // GitHub supports teams but not groups
    githubPlugin := vcs.NewMockVCSPlugin(vcs.VCSCapabilities{
        SupportsTeamAllowlist: true,
        SupportsGroupAllowlist: false,
    })
    
    // GitLab supports groups but not teams  
    gitlabPlugin := vcs.NewMockVCSPlugin(vcs.VCSCapabilities{
        SupportsTeamAllowlist: false,
        SupportsGroupAllowlist: true,
    })
    
    config := vcs.FeatureConfig{
        TeamAllowlist: []string{"devops"},
    }
    
    // Should pass for GitHub
    assert.NoError(t, validator.ValidateFeatures("github", config))
    
    // Should warn for GitLab  
    assert.Error(t, validator.ValidateFeatures("gitlab", config))
}
```

## Implementation Checklist

- [ ] Define VCS plugin interfaces
- [ ] Create capability detection system
- [ ] Implement feature validation
- [ ] Create mock plugins for testing
- [ ] Implement GitHub plugin
- [ ] Implement GitLab plugin  
- [ ] Implement Azure DevOps plugin
- [ ] Add configuration migration
- [ ] Add deprecation warnings
- [ ] Update documentation
- [ ] Add integration tests
- [ ] Remove deprecated flags (future release)

## Migration Timeline

- **Week 1**: Interface design and validation
- **Week 2**: Core plugin infrastructure
- **Weeks 3-4**: VCS plugin implementations
- **Week 5**: Configuration updates
- **Weeks 6-8**: Migration and testing
- **Week 9**: Documentation and rollout 