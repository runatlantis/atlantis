# VCS Plugin Compatibility - Phase 3

## Overview

Phase 3 of the VCS Plugin Architecture implementation focuses on **Backward compatibility layer**. This phase creates a compatibility layer that allows both old and new VCS systems to work together during migration, providing a smooth transition path with feature flags and gradual rollout capabilities.

## Implemented Components

### 1. Compatibility Layer (`internal/infrastructure/vcs/compatibility_layer.go`)

The `CompatibilityLayer` provides seamless backward compatibility during VCS plugin migration:

- **Dual System Support**: Supports both legacy VCS clients and new plugin system
- **Intelligent Fallback**: Falls back to legacy system when plugins fail
- **Deprecation Warnings**: Logs warnings to help users migrate
- **Feature Detection**: Uses plugin capabilities to determine available features

#### Key Features:

```go
// Try new plugin system first, fallback to legacy
func (c *CompatibilityLayer) UpdateStatus(...) error {
    if c.enableNewSystem {
        plugin, err := c.getVCSPlugin(repo.VCSHost.Type)
        if err == nil {
            // Use new plugin system
            return plugin.CreateCommitStatus(...)
        }
        
        if !c.fallbackToLegacy {
            return fmt.Errorf("new VCS plugin failed and fallback disabled")
        }
        
        c.logger.Warn("New VCS plugin failed, falling back to legacy client")
    }
    
    // Use legacy system
    c.logDeprecationWarning("UpdateStatus", "Use VCS plugin CreateCommitStatus method")
    return c.legacyClient.UpdateStatus(...)
}
```

### 2. Feature Flag System (`internal/infrastructure/vcs/feature_flags.go`)

Comprehensive feature flag system for controlling VCS plugin rollout:

#### Migration Modes:

| Mode | Description | Plugin Enabled | Fallback | Use Case |
|------|-------------|----------------|----------|----------|
| `disabled` | Legacy only | ❌ | ✅ | Default, safe rollback |
| `opt-in` | Selective enablement | ✅ | ✅ | Testing specific features |
| `gradual` | Plugin with fallback | ✅ | ✅ | Production rollout |
| `strict` | Plugin only | ✅ | ❌ | Full migration |

#### Environment Variables:

```bash
# Enable VCS plugins
export ATLANTIS_VCS_PLUGINS_ENABLED=true

# Enable fallback to legacy system
export ATLANTIS_VCS_PLUGINS_FALLBACK=true

# Log deprecation warnings
export ATLANTIS_VCS_PLUGINS_LOG_DEPRECATION=true

# Set migration mode
export ATLANTIS_VCS_PLUGINS_MIGRATION_MODE=gradual

# Enable specific operation groups
export ATLANTIS_VCS_PLUGINS_ENABLE_FILE_OPS=false
export ATLANTIS_VCS_PLUGINS_ENABLE_COMMENT_OPS=false
export ATLANTIS_VCS_PLUGINS_ENABLE_REVIEW_OPS=false
export ATLANTIS_VCS_PLUGINS_ENABLE_MERGE_OPS=false
```

### 3. Operation-Level Control

Fine-grained control over which operations use the new plugin system:

```go
type VCSPluginEnabledOperations struct {
    // Repository operations (enabled by default)
    GetRepository    bool
    GetPullRequest   bool
    
    // Status operations (enabled by default)
    CreateCommitStatus bool
    UpdateStatus       bool
    
    // Validation operations (enabled by default)
    CheckMergeableBypass    bool
    ValidateTeamMembership  bool
    ValidateGroupMembership bool
    
    // Complex operations (disabled by default)
    GetModifiedFiles bool
    CreateComment    bool
    MergePull        bool
}
```

## Migration Strategy

### Phase 3A: Safe Rollout (Default)

```bash
# Start with plugins disabled
export ATLANTIS_VCS_PLUGINS_MIGRATION_MODE=disabled
```

- All operations use legacy VCS clients
- No risk of breaking existing functionality
- Deprecation warnings help identify usage patterns

### Phase 3B: Gradual Testing

```bash
# Enable plugins with fallback
export ATLANTIS_VCS_PLUGINS_MIGRATION_MODE=gradual
export ATLANTIS_VCS_PLUGINS_ENABLED=true
export ATLANTIS_VCS_PLUGINS_FALLBACK=true
```

- New plugin system tries first
- Falls back to legacy on any failure
- Logs show which operations succeed with plugins

### Phase 3C: Selective Enablement

```bash
# Enable specific operation groups
export ATLANTIS_VCS_PLUGINS_MIGRATION_MODE=opt-in
export ATLANTIS_VCS_PLUGINS_ENABLE_FILE_OPS=true
```

- Only enabled operations use plugins
- Granular control over migration
- Reduce risk by enabling operations incrementally

### Phase 3D: Full Migration

```bash
# Plugin-only mode
export ATLANTIS_VCS_PLUGINS_MIGRATION_MODE=strict
export ATLANTIS_VCS_PLUGINS_FALLBACK=false
```

- All operations must use plugins
- No fallback to legacy system
- Complete migration to new architecture

## Compatibility Matrix

### Supported Operations

| Operation | Legacy | Plugin | Fallback | Notes |
|-----------|--------|--------|----------|-------|
| `GetRepository` | ✅ | ✅ | ✅ | Full compatibility |
| `GetPullRequest` | ✅ | ✅ | ✅ | Full compatibility |
| `CreateCommitStatus` | ✅ | ✅ | ✅ | Full compatibility |
| `UpdateStatus` | ✅ | ✅ | ✅ | Full compatibility |
| `CheckMergeableBypass` | ✅ | ✅ | ✅ | GitHub only |
| `ValidateTeamMembership` | ✅ | ✅ | ✅ | GitHub only |
| `ValidateGroupMembership` | ✅ | ✅ | ✅ | GitLab only |
| `GetModifiedFiles` | ✅ | 🚧 | ✅ | Plugin implementation pending |
| `CreateComment` | ✅ | 🚧 | ✅ | Plugin implementation pending |
| `MergePull` | ✅ | 🚧 | ✅ | Plugin implementation pending |

### VCS Provider Support

| Provider | Legacy | Plugin | Capabilities |
|----------|--------|--------|--------------|
| GitHub | ✅ | ✅ | Mergeable bypass, team allowlist |
| GitLab | ✅ | ✅ | Group allowlist, merge requests |
| Azure DevOps | ✅ | ✅ | Custom fields, pull requests |
| Bitbucket | ✅ | ✅ | Basic operations |
| Gitea | ✅ | ✅ | Basic operations |

## Usage Examples

### Basic Setup

```go
// Load feature flags from environment
flags := vcs.LoadVCSFeatureFlagsFromEnv()

// Create compatibility layer
compatLayer := vcs.NewCompatibilityLayer(
    legacyVCSClient,
    pluginRegistry,
    userConfig,
    logger,
    flags.ToCompatibilityConfig(),
)

// Use compatibility layer as drop-in replacement
files, err := compatLayer.GetModifiedFiles(logger, repo, pull)
```

### Custom Configuration

```go
// Custom feature flags
flags := vcs.VCSFeatureFlags{
    EnableVCSPlugins:       true,
    FallbackToLegacy:       true,
    LogDeprecationWarnings: true,
    VCSPluginMigrationMode: vcs.MigrationModeGradual,
}

// Create compatibility layer
compatLayer := vcs.NewCompatibilityLayer(
    legacyVCSClient,
    pluginRegistry,
    userConfig,
    logger,
    flags.ToCompatibilityConfig(),
)
```

### Monitoring Migration

```go
// Enable detailed logging
flags := vcs.LoadVCSFeatureFlagsFromEnv()
if flags.ShouldLogDeprecationWarnings() {
    logger.Info("VCS plugin migration warnings enabled")
}

// Check migration mode
switch flags.GetMigrationMode() {
case vcs.MigrationModeDisabled:
    logger.Info("VCS plugins disabled, using legacy only")
case vcs.MigrationModeGradual:
    logger.Info("VCS plugins enabled with fallback")
case vcs.MigrationModeStrict:
    logger.Info("VCS plugins only, no fallback")
}
```

## Benefits

### 1. **Zero-Downtime Migration**
- No breaking changes to existing functionality
- Gradual rollout with immediate rollback capability
- Feature flags allow instant disable if issues arise

### 2. **Risk Mitigation**
- Fallback to proven legacy system
- Operation-level control reduces blast radius
- Comprehensive logging for troubleshooting

### 3. **Flexible Deployment**
- Environment-based configuration
- Per-operation enablement
- Multiple migration strategies

### 4. **Observability**
- Deprecation warnings guide migration
- Success/failure metrics for each system
- Clear visibility into plugin usage

## Testing

Run the compatibility layer tests:

```bash
go test ./internal/infrastructure/vcs/...
```

### Test Coverage

- ✅ Legacy-only mode
- ✅ Plugin-with-fallback mode
- ✅ Plugin-only mode
- ✅ VCS host type conversion
- ✅ Commit state conversion
- ✅ Error handling and fallback logic

## Next Steps

Phase 3 enables:

- **Phase 4**: Deprecation of legacy flags with migration warnings
- **Production rollout**: Gradual migration in production environments
- **Monitoring**: Metrics and observability for migration progress

## Related Documentation

- [VCS Plugin Architecture Overview](VCS_PLUGIN_ARCHITECTURE.md)
- [VCS Plugin Integration Phase 2](VCS_PLUGIN_INTEGRATION_PHASE2.md)
- [GitHub Issue #5574](https://github.com/runatlantis/atlantis/issues/5574) - Original problem statement 