# VCS Service Architecture

## Overview

The VCS Service is a central component of the new VCS plugin architecture for Atlantis, designed to address [issue #5574](https://github.com/runatlantis/atlantis/issues/5574) by providing a unified interface for VCS operations while supporting provider-specific features through a plugin system.

## Architecture Components

### 1. Service Layer (`internal/infrastructure/vcs/service.go`)

The VCS Service acts as the primary orchestrator for all VCS operations:

```go
type Service struct {
    registry         vcs.VCSRegistry
    validator        *vcs.FeatureValidator
    deprecationMgr   *DeprecationManager
    migrationCLI     *MigrationCLI
    defaultProvider  ProviderType
    enabledFeatures  map[string]bool
    logger           logging.SimpleLogging
}
```

**Key Responsibilities:**
- Plugin registration and lifecycle management
- Feature capability validation
- Provider-specific operation routing
- Migration assistance and deprecation warnings
- Health monitoring and feature matrix reporting

### 2. Domain Layer (`internal/domain/vcs/`)

**Plugin Interface (`plugin.go`):**
```go
type VCSPlugin interface {
    GetRepository(ctx context.Context, owner, name string) (*Repository, error)
    GetPullRequest(ctx context.Context, repo Repository, number int) (*PullRequest, error)
    CreateCommitStatus(ctx context.Context, repo Repository, sha string, status CommitStatus) error
    
    // Capability detection
    Capabilities() VCSCapabilities
    
    // Feature implementations
    CheckMergeableBypass(ctx context.Context, pr *PullRequest) (bool, error)
    ValidateTeamMembership(ctx context.Context, user string, teams []string) (bool, error)
    ValidateGroupMembership(ctx context.Context, user string, groups []string) (bool, error)
}
```

**Capabilities System:**
```go
type VCSCapabilities struct {
    SupportsMergeableBypass bool
    SupportsTeamAllowlist   bool
    SupportsGroupAllowlist  bool
    SupportsCustomFields    bool
    MaxPageSize            int
}
```

### 3. Migration Support

**Deprecation Manager (`deprecation_manager.go`):**
- Tracks deprecated VCS flags and configuration
- Provides migration warnings and guidance
- Supports gradual migration strategies

**Migration CLI (`migration_cli.go`):**
- Command-line tools for migration planning
- Configuration validation and compatibility checking
- Step-by-step migration instructions

## Addressing Issue #5574

### Problem Statement

Issue #5574 highlighted the need for VCS-specific features while maintaining clean abstractions. The key challenges were:

1. **Feature Coupling**: VCS-specific features were tightly coupled to implementation
2. **Provider Lock-in**: Difficult to support multiple VCS providers with different capabilities
3. **Configuration Complexity**: VCS-specific flags scattered throughout the codebase
4. **Testing Challenges**: Hard to test VCS interactions in isolation

### Solution Architecture

#### 1. Plugin-Based Design

**Before:**
```go
// Tightly coupled to GitHub
if config.GithubAllowMergeableBypass {
    // GitHub-specific logic mixed with business logic
}
```

**After:**
```go
// Provider-agnostic with capability detection
if service.SupportsFeature(providerType, "mergeable_bypass") && 
   service.IsFeatureEnabled("mergeable_bypass") {
    result, err := service.CheckMergeableBypass(ctx, providerType, pr)
}
```

#### 2. Capability-Based Feature Detection

Instead of provider-specific flags, the system uses capability detection:

```go
// Check if provider supports the feature
capabilities := plugin.Capabilities()
if !capabilities.SupportsMergeableBypass {
    return false, errors.Errorf("provider %s does not support mergeable bypass", providerType)
}
```

#### 3. Feature Flag System

Global feature enablement separate from provider capabilities:

```go
enabledFeatures := map[string]bool{
    "mergeable_bypass": true,
    "team_allowlist":   true,
    "group_allowlist":  false,
}
```

#### 4. Migration Strategy

**Gradual Migration Path:**
1. **Phase 1**: Plugin system runs alongside existing code
2. **Phase 2**: Feature flags control which system to use
3. **Phase 3**: Deprecation warnings for old flags
4. **Phase 4**: Complete migration to plugin system

## Usage Examples

### 1. Basic Service Setup

```go
config := &ServiceConfig{
    DefaultProvider: GitHub,
    EnabledFeatures: map[string]bool{
        "mergeable_bypass": true,
        "team_allowlist":   true,
    },
    UserConfig: userConfig,
    Logger:     logger,
}

service, err := NewService(config)
if err != nil {
    return err
}
```

### 2. Plugin Registration

```go
githubPlugin := github.NewPlugin(githubConfig)
err := service.RegisterPlugin("github", githubPlugin)
if err != nil {
    return err
}
```

### 3. Feature Usage

```go
// Check mergeable bypass
canBypass, err := service.CheckMergeableBypass(ctx, GitHub, pullRequest)
if err != nil {
    return err
}

// Validate team membership
isValidTeam, err := service.ValidateTeamMembership(ctx, GitHub, user, teams)
if err != nil {
    return err
}
```

### 4. Migration Support

```go
// Check for deprecated flags
migrationCLI := service.GetMigrationCLI()
err := migrationCLI.RunMigrationCommand("check", []string{})

// Generate migration plan
err = migrationCLI.RunMigrationCommand("plan", []string{})

// Validate current configuration
err = migrationCLI.RunMigrationCommand("validate", []string{})
```

## Benefits

### 1. Provider Flexibility

- **Multi-Provider Support**: Easy to add new VCS providers
- **Feature Parity**: Clear visibility into provider capabilities
- **Gradual Adoption**: Can migrate providers one at a time

### 2. Clean Architecture

- **Separation of Concerns**: Domain logic separate from VCS implementation
- **Testability**: Easy to mock VCS interactions
- **Maintainability**: Clear boundaries between components

### 3. Migration Safety

- **Backwards Compatibility**: Existing configurations continue to work
- **Deprecation Warnings**: Clear guidance on migration path
- **Validation Tools**: CLI tools to validate migration steps

### 4. Operational Benefits

- **Health Monitoring**: Built-in health checks for all providers
- **Feature Matrix**: Clear view of what each provider supports
- **Migration Planning**: Automated migration planning and validation

## Future Enhancements

### 1. Dynamic Plugin Loading

```go
// Load plugins from external modules
pluginLoader := NewPluginLoader()
plugin, err := pluginLoader.LoadPlugin("github-enterprise", pluginConfig)
```

### 2. Advanced Capability Detection

```go
// Version-specific capabilities
type VCSCapabilities struct {
    Features map[string]FeatureCapability
}

type FeatureCapability struct {
    Supported      bool
    MinVersion     string
    Configuration  map[string]interface{}
    Limitations    []string
}
```

### 3. Configuration Management

```go
// Centralized VCS configuration
type VCSConfig struct {
    Providers map[ProviderType]ProviderConfig
    Features  FeatureConfig
    Migration MigrationConfig
}
```

## Integration with Existing Codebase

### 1. Server Integration

The service integrates with the existing Atlantis server:

```go
// In server initialization
vcsService, err := vcs.NewService(&vcs.ServiceConfig{
    DefaultProvider: config.VCSProvider,
    EnabledFeatures: config.VCSFeatures,
    UserConfig:      config,
    Logger:          logger,
})

server := &Server{
    VCSService: vcsService,
    // ... other components
}
```

### 2. Event Handler Integration

VCS operations are routed through the service:

```go
// In event handlers
func (s *Server) handlePullRequest(ctx context.Context, event PullRequestEvent) error {
    // Get repository information
    repo, err := s.VCSService.GetRepository(ctx, event.ProviderType, event.Owner, event.Repo)
    if err != nil {
        return err
    }
    
    // Check mergeable bypass if supported
    if s.VCSService.SupportsFeature(event.ProviderType, "mergeable_bypass") {
        canBypass, err := s.VCSService.CheckMergeableBypass(ctx, event.ProviderType, event.PullRequest)
        // ... handle result
    }
}
```

### 3. Testing Integration

The mock plugin system enables comprehensive testing:

```go
func TestPullRequestHandling(t *testing.T) {
    mockPlugin := vcs.NewMockVCSPlugin(vcs.VCSCapabilities{
        SupportsMergeableBypass: true,
        SupportsTeamAllowlist:   true,
    })
    
    service := createTestService(t)
    service.RegisterPlugin("github", mockPlugin)
    
    // Test scenarios...
}
```

## Conclusion

The VCS Service architecture successfully addresses issue #5574 by providing:

1. **Clean Abstraction**: Unified interface for VCS operations
2. **Provider Flexibility**: Support for multiple VCS providers with different capabilities
3. **Feature Safety**: Capability-based feature detection prevents runtime errors
4. **Migration Path**: Tools and strategies for gradual migration
5. **Operational Excellence**: Health monitoring, feature matrices, and validation tools

This architecture sets the foundation for a more maintainable, testable, and extensible VCS integration system in Atlantis. 