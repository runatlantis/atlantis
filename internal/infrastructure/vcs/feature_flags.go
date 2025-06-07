package vcs

import (
	"os"
	"strconv"
	"strings"
)

// VCSFeatureFlags controls the rollout of VCS plugin features
type VCSFeatureFlags struct {
	// EnableVCSPlugins enables the new VCS plugin system
	EnableVCSPlugins bool
	
	// FallbackToLegacy enables fallback to legacy VCS clients
	FallbackToLegacy bool
	
	// LogDeprecationWarnings logs warnings for deprecated VCS operations
	LogDeprecationWarnings bool
	
	// EnableVCSPluginValidation enables VCS feature validation
	EnableVCSPluginValidation bool
	
	// VCSPluginMigrationMode controls migration behavior
	VCSPluginMigrationMode MigrationMode
}

// MigrationMode defines different migration strategies
type MigrationMode string

const (
	// MigrationModeDisabled - VCS plugins disabled, use legacy only
	MigrationModeDisabled MigrationMode = "disabled"
	
	// MigrationModeOptIn - VCS plugins enabled for specific operations
	MigrationModeOptIn MigrationMode = "opt-in"
	
	// MigrationModeGradual - VCS plugins enabled with fallback
	MigrationModeGradual MigrationMode = "gradual"
	
	// MigrationModeStrict - VCS plugins only, no fallback
	MigrationModeStrict MigrationMode = "strict"
)

// DefaultVCSFeatureFlags returns the default feature flag configuration
func DefaultVCSFeatureFlags() VCSFeatureFlags {
	return VCSFeatureFlags{
		EnableVCSPlugins:          false, // Disabled by default for safety
		FallbackToLegacy:          true,  // Always fallback initially
		LogDeprecationWarnings:    true,  // Help users migrate
		EnableVCSPluginValidation: true,  // Always validate
		VCSPluginMigrationMode:    MigrationModeDisabled,
	}
}

// LoadVCSFeatureFlagsFromEnv loads feature flags from environment variables
func LoadVCSFeatureFlagsFromEnv() VCSFeatureFlags {
	flags := DefaultVCSFeatureFlags()
	
	// ATLANTIS_VCS_PLUGINS_ENABLED
	if val := os.Getenv("ATLANTIS_VCS_PLUGINS_ENABLED"); val != "" {
		if enabled, err := strconv.ParseBool(val); err == nil {
			flags.EnableVCSPlugins = enabled
		}
	}
	
	// ATLANTIS_VCS_PLUGINS_FALLBACK
	if val := os.Getenv("ATLANTIS_VCS_PLUGINS_FALLBACK"); val != "" {
		if fallback, err := strconv.ParseBool(val); err == nil {
			flags.FallbackToLegacy = fallback
		}
	}
	
	// ATLANTIS_VCS_PLUGINS_LOG_DEPRECATION
	if val := os.Getenv("ATLANTIS_VCS_PLUGINS_LOG_DEPRECATION"); val != "" {
		if logWarnings, err := strconv.ParseBool(val); err == nil {
			flags.LogDeprecationWarnings = logWarnings
		}
	}
	
	// ATLANTIS_VCS_PLUGINS_VALIDATION
	if val := os.Getenv("ATLANTIS_VCS_PLUGINS_VALIDATION"); val != "" {
		if validation, err := strconv.ParseBool(val); err == nil {
			flags.EnableVCSPluginValidation = validation
		}
	}
	
	// ATLANTIS_VCS_PLUGINS_MIGRATION_MODE
	if val := os.Getenv("ATLANTIS_VCS_PLUGINS_MIGRATION_MODE"); val != "" {
		mode := MigrationMode(strings.ToLower(val))
		if isValidMigrationMode(mode) {
			flags.VCSPluginMigrationMode = mode
			
			// Adjust other flags based on migration mode
			switch mode {
			case MigrationModeDisabled:
				flags.EnableVCSPlugins = false
				flags.FallbackToLegacy = true
			case MigrationModeOptIn:
				flags.EnableVCSPlugins = true
				flags.FallbackToLegacy = true
			case MigrationModeGradual:
				flags.EnableVCSPlugins = true
				flags.FallbackToLegacy = true
			case MigrationModeStrict:
				flags.EnableVCSPlugins = true
				flags.FallbackToLegacy = false
			}
		}
	}
	
	return flags
}

// isValidMigrationMode checks if the migration mode is valid
func isValidMigrationMode(mode MigrationMode) bool {
	switch mode {
	case MigrationModeDisabled, MigrationModeOptIn, MigrationModeGradual, MigrationModeStrict:
		return true
	default:
		return false
	}
}

// ShouldUseVCSPlugins determines if VCS plugins should be used
func (f VCSFeatureFlags) ShouldUseVCSPlugins() bool {
	return f.EnableVCSPlugins && f.VCSPluginMigrationMode != MigrationModeDisabled
}

// ShouldFallbackToLegacy determines if fallback to legacy is allowed
func (f VCSFeatureFlags) ShouldFallbackToLegacy() bool {
	return f.FallbackToLegacy && f.VCSPluginMigrationMode != MigrationModeStrict
}

// ShouldLogDeprecationWarnings determines if deprecation warnings should be logged
func (f VCSFeatureFlags) ShouldLogDeprecationWarnings() bool {
	return f.LogDeprecationWarnings
}

// ShouldValidateVCSFeatures determines if VCS features should be validated
func (f VCSFeatureFlags) ShouldValidateVCSFeatures() bool {
	return f.EnableVCSPluginValidation
}

// GetMigrationMode returns the current migration mode
func (f VCSFeatureFlags) GetMigrationMode() MigrationMode {
	return f.VCSPluginMigrationMode
}

// ToCompatibilityConfig converts feature flags to compatibility layer config
func (f VCSFeatureFlags) ToCompatibilityConfig() CompatibilityConfig {
	return CompatibilityConfig{
		EnableNewSystem:      f.ShouldUseVCSPlugins(),
		FallbackToLegacy:     f.ShouldFallbackToLegacy(),
		LogMigrationWarnings: f.ShouldLogDeprecationWarnings(),
	}
}

// VCSPluginEnabledOperations defines which operations are enabled for VCS plugins
type VCSPluginEnabledOperations struct {
	// Repository operations
	GetRepository    bool
	GetPullRequest   bool
	
	// Status operations
	CreateCommitStatus bool
	UpdateStatus       bool
	
	// Validation operations
	CheckMergeableBypass    bool
	ValidateTeamMembership  bool
	ValidateGroupMembership bool
	
	// File operations (future)
	GetModifiedFiles bool
	GetFileContent   bool
	
	// Comment operations (future)
	CreateComment           bool
	ReactToComment          bool
	HidePrevCommandComments bool
	
	// Review operations (future)
	PullIsApproved bool
	DiscardReviews bool
	
	// Merge operations (future)
	MergePull bool
}

// DefaultVCSPluginEnabledOperations returns the default enabled operations
func DefaultVCSPluginEnabledOperations() VCSPluginEnabledOperations {
	return VCSPluginEnabledOperations{
		// Only enable basic operations initially
		GetRepository:           true,
		GetPullRequest:          true,
		CreateCommitStatus:      true,
		UpdateStatus:            true,
		CheckMergeableBypass:    true,
		ValidateTeamMembership:  true,
		ValidateGroupMembership: true,
		
		// Disable complex operations for now
		GetModifiedFiles:        false,
		GetFileContent:          false,
		CreateComment:           false,
		ReactToComment:          false,
		HidePrevCommandComments: false,
		PullIsApproved:          false,
		DiscardReviews:          false,
		MergePull:               false,
	}
}

// LoadVCSPluginEnabledOperationsFromEnv loads enabled operations from environment
func LoadVCSPluginEnabledOperationsFromEnv() VCSPluginEnabledOperations {
	ops := DefaultVCSPluginEnabledOperations()
	
	// Allow enabling specific operations via environment variables
	if val := os.Getenv("ATLANTIS_VCS_PLUGINS_ENABLE_FILE_OPS"); val != "" {
		if enabled, err := strconv.ParseBool(val); err == nil {
			ops.GetModifiedFiles = enabled
			ops.GetFileContent = enabled
		}
	}
	
	if val := os.Getenv("ATLANTIS_VCS_PLUGINS_ENABLE_COMMENT_OPS"); val != "" {
		if enabled, err := strconv.ParseBool(val); err == nil {
			ops.CreateComment = enabled
			ops.ReactToComment = enabled
			ops.HidePrevCommandComments = enabled
		}
	}
	
	if val := os.Getenv("ATLANTIS_VCS_PLUGINS_ENABLE_REVIEW_OPS"); val != "" {
		if enabled, err := strconv.ParseBool(val); err == nil {
			ops.PullIsApproved = enabled
			ops.DiscardReviews = enabled
		}
	}
	
	if val := os.Getenv("ATLANTIS_VCS_PLUGINS_ENABLE_MERGE_OPS"); val != "" {
		if enabled, err := strconv.ParseBool(val); err == nil {
			ops.MergePull = enabled
		}
	}
	
	return ops
} 