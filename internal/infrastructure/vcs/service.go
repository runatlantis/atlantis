package vcs

import (
	"context"
	"strings"

	"github.com/pkg/errors"
	"github.com/runatlantis/atlantis/internal/domain/vcs"
	"github.com/runatlantis/atlantis/server"
	"github.com/runatlantis/atlantis/server/logging"
)

// ProviderType represents a VCS provider type
type ProviderType string

const (
	GitHub     ProviderType = "github"
	GitLab     ProviderType = "gitlab"
	Bitbucket  ProviderType = "bitbucket"
	AzureRepos ProviderType = "azurerepos"
)

// Service provides VCS operations through a plugin-based architecture
type Service struct {
	registry         vcs.VCSRegistry
	validator        *vcs.FeatureValidator
	deprecationMgr   *DeprecationManager
	migrationCLI     *MigrationCLI
	defaultProvider  ProviderType
	enabledFeatures  map[string]bool
	logger           logging.SimpleLogging
}

// ServiceConfig contains configuration options for the VCS service
type ServiceConfig struct {
	DefaultProvider   ProviderType
	EnabledFeatures   map[string]bool
	UserConfig        server.UserConfig
	Logger            logging.SimpleLogging
}

// SimpleRegistry implements VCSRegistry interface
type SimpleRegistry struct {
	plugins map[string]vcs.VCSPlugin
}

// NewSimpleRegistry creates a new simple registry
func NewSimpleRegistry() *SimpleRegistry {
	return &SimpleRegistry{
		plugins: make(map[string]vcs.VCSPlugin),
	}
}

// Register registers a plugin
func (r *SimpleRegistry) Register(name string, plugin vcs.VCSPlugin) error {
	if plugin == nil {
		return errors.New("plugin cannot be nil")
	}
	r.plugins[name] = plugin
	return nil
}

// Get retrieves a plugin by name
func (r *SimpleRegistry) Get(name string) (vcs.VCSPlugin, error) {
	plugin, exists := r.plugins[name]
	if !exists {
		return nil, errors.Errorf("plugin %s not found", name)
	}
	return plugin, nil
}

// List returns all registered plugin names
func (r *SimpleRegistry) List() []string {
	var names []string
	for name := range r.plugins {
		names = append(names, name)
	}
	return names
}

// NewService creates a new VCS service with the provided configuration
func NewService(config *ServiceConfig) (*Service, error) {
	if config == nil {
		return nil, errors.New("service config cannot be nil")
	}

	registry := NewSimpleRegistry()
	validator := &vcs.FeatureValidator{}
	
	deprecationMgr := NewDeprecationManager(config.Logger, config.UserConfig)
	migrationCLI := NewMigrationCLI(config.UserConfig, config.Logger)

	return &Service{
		registry:        registry,
		validator:       validator,
		deprecationMgr:  deprecationMgr,
		migrationCLI:    migrationCLI,
		defaultProvider: config.DefaultProvider,
		enabledFeatures: config.EnabledFeatures,
		logger:          config.Logger,
	}, nil
}

// RegisterPlugin registers a VCS plugin with the service
func (s *Service) RegisterPlugin(name string, plugin vcs.VCSPlugin) error {
	return s.registry.Register(name, plugin)
}

// GetPlugin retrieves a plugin by provider type
func (s *Service) GetPlugin(providerType ProviderType) (vcs.VCSPlugin, error) {
	plugin, err := s.registry.Get(string(providerType))
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get plugin for provider %s", providerType)
	}

	return plugin, nil
}

// GetDefaultPlugin returns the plugin for the default provider
func (s *Service) GetDefaultPlugin() (vcs.VCSPlugin, error) {
	return s.GetPlugin(s.defaultProvider)
}

// SupportsFeature checks if a provider supports a specific feature
func (s *Service) SupportsFeature(providerType ProviderType, feature string) (bool, error) {
	plugin, err := s.GetPlugin(providerType)
	if err != nil {
		return false, err
	}

	capabilities := plugin.Capabilities()
	switch feature {
	case "mergeable_bypass":
		return capabilities.SupportsMergeableBypass, nil
	case "team_allowlist":
		return capabilities.SupportsTeamAllowlist, nil
	case "group_allowlist":
		return capabilities.SupportsGroupAllowlist, nil
	case "custom_fields":
		return capabilities.SupportsCustomFields, nil
	default:
		return false, nil
	}
}

// ValidateConfiguration validates VCS configuration for a provider
func (s *Service) ValidateConfiguration(providerType ProviderType, config vcs.FeatureConfig) error {
	return s.validator.ValidateFeatures(string(providerType), config)
}

// ListProviders returns all registered provider types
func (s *Service) ListProviders() []ProviderType {
	names := s.registry.List()
	var providers []ProviderType
	for _, name := range names {
		providers = append(providers, ProviderType(name))
	}
	return providers
}

// IsFeatureEnabled checks if a feature is globally enabled
func (s *Service) IsFeatureEnabled(feature string) bool {
	enabled, exists := s.enabledFeatures[feature]
	return exists && enabled
}

// GetRepository retrieves repository information
func (s *Service) GetRepository(ctx context.Context, providerType ProviderType, owner, name string) (*vcs.Repository, error) {
	plugin, err := s.GetPlugin(providerType)
	if err != nil {
		return nil, err
	}

	return plugin.GetRepository(ctx, owner, name)
}

// GetPullRequest retrieves pull request information
func (s *Service) GetPullRequest(ctx context.Context, providerType ProviderType, repo vcs.Repository, number int) (*vcs.PullRequest, error) {
	plugin, err := s.GetPlugin(providerType)
	if err != nil {
		return nil, err
	}

	return plugin.GetPullRequest(ctx, repo, number)
}

// UpdateCommitStatus updates the commit status
func (s *Service) UpdateCommitStatus(ctx context.Context, providerType ProviderType, repo vcs.Repository, sha string, status vcs.CommitStatus) error {
	plugin, err := s.GetPlugin(providerType)
	if err != nil {
		return err
	}

	return plugin.CreateCommitStatus(ctx, repo, sha, status)
}

// CheckMergeableBypass checks if mergeable bypass is allowed
func (s *Service) CheckMergeableBypass(ctx context.Context, providerType ProviderType, pr *vcs.PullRequest) (bool, error) {
	plugin, err := s.GetPlugin(providerType)
	if err != nil {
		return false, err
	}

	capabilities := plugin.Capabilities()
	if !capabilities.SupportsMergeableBypass {
		return false, errors.Errorf("provider %s does not support mergeable bypass", providerType)
	}

	if !s.IsFeatureEnabled("mergeable_bypass") {
		return false, errors.New("mergeable bypass feature is disabled")
	}

	return plugin.CheckMergeableBypass(ctx, pr)
}

// ValidateTeamMembership validates team membership
func (s *Service) ValidateTeamMembership(ctx context.Context, providerType ProviderType, user string, teams []string) (bool, error) {
	plugin, err := s.GetPlugin(providerType)
	if err != nil {
		return false, err
	}

	capabilities := plugin.Capabilities()
	if !capabilities.SupportsTeamAllowlist {
		return false, errors.Errorf("provider %s does not support team allowlist", providerType)
	}

	if !s.IsFeatureEnabled("team_allowlist") {
		return false, errors.New("team allowlist feature is disabled")
	}

	return plugin.ValidateTeamMembership(ctx, user, teams)
}

// ValidateGroupMembership validates group membership
func (s *Service) ValidateGroupMembership(ctx context.Context, providerType ProviderType, user string, groups []string) (bool, error) {
	plugin, err := s.GetPlugin(providerType)
	if err != nil {
		return false, err
	}

	capabilities := plugin.Capabilities()
	if !capabilities.SupportsGroupAllowlist {
		return false, errors.Errorf("provider %s does not support group allowlist", providerType)
	}

	if !s.IsFeatureEnabled("group_allowlist") {
		return false, errors.New("group allowlist feature is disabled")
	}

	return plugin.ValidateGroupMembership(ctx, user, groups)
}

// GetMigrationCLI returns the migration CLI for provider transitions
func (s *Service) GetMigrationCLI() *MigrationCLI {
	return s.migrationCLI
}

// GetDeprecationManager returns the deprecation manager
func (s *Service) GetDeprecationManager() *DeprecationManager {
	return s.deprecationMgr
}

// HealthCheck performs a health check on all registered plugins
func (s *Service) HealthCheck(ctx context.Context) map[ProviderType]error {
	providers := s.ListProviders()
	results := make(map[ProviderType]error)

	for _, providerType := range providers {
		plugin, err := s.GetPlugin(providerType)
		if err != nil {
			results[providerType] = err
			continue
		}

		// Try to get capabilities as a basic health check
		capabilities := plugin.Capabilities()
		if capabilities.MaxPageSize <= 0 {
			results[providerType] = errors.New("plugin has invalid configuration")
			continue
		}

		results[providerType] = nil // Success
	}

	return results
}

// GetFeatureMatrix returns a matrix of features supported by each provider
func (s *Service) GetFeatureMatrix() map[ProviderType]map[string]bool {
	providers := s.ListProviders()
	matrix := make(map[ProviderType]map[string]bool)

	for _, providerType := range providers {
		plugin, err := s.GetPlugin(providerType)
		if err != nil {
			continue
		}

		capabilities := plugin.Capabilities()
		features := map[string]bool{
			"mergeable_bypass": capabilities.SupportsMergeableBypass,
			"team_allowlist":   capabilities.SupportsTeamAllowlist,
			"group_allowlist":  capabilities.SupportsGroupAllowlist,
			"custom_fields":    capabilities.SupportsCustomFields,
		}
		matrix[providerType] = features
	}

	return matrix
}

// ValidateProviderMigration validates if migration between providers is possible
func (s *Service) ValidateProviderMigration(from, to ProviderType, features []string) error {
	fromPlugin, err := s.GetPlugin(from)
	if err != nil {
		return errors.Wrapf(err, "source provider %s not available", from)
	}

	toPlugin, err := s.GetPlugin(to)
	if err != nil {
		return errors.Wrapf(err, "target provider %s not available", to)
	}

	fromCapabilities := fromPlugin.Capabilities()
	toCapabilities := toPlugin.Capabilities()

	var unsupportedFeatures []string
	for _, feature := range features {
		var fromSupports, toSupports bool
		
		switch feature {
		case "mergeable_bypass":
			fromSupports = fromCapabilities.SupportsMergeableBypass
			toSupports = toCapabilities.SupportsMergeableBypass
		case "team_allowlist":
			fromSupports = fromCapabilities.SupportsTeamAllowlist
			toSupports = toCapabilities.SupportsTeamAllowlist
		case "group_allowlist":
			fromSupports = fromCapabilities.SupportsGroupAllowlist
			toSupports = toCapabilities.SupportsGroupAllowlist
		case "custom_fields":
			fromSupports = fromCapabilities.SupportsCustomFields
			toSupports = toCapabilities.SupportsCustomFields
		}
		
		if fromSupports && !toSupports {
			unsupportedFeatures = append(unsupportedFeatures, feature)
		}
	}

	if len(unsupportedFeatures) > 0 {
		return errors.Errorf(
			"migration from %s to %s would lose support for features: %s",
			from, to, strings.Join(unsupportedFeatures, ", "),
		)
	}

	return nil
} 