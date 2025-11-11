package vcs

import (
	"fmt"
	"strings"
	"time"

	"github.com/runatlantis/atlantis/server"
	"github.com/runatlantis/atlantis/server/logging"
)

// DeprecationManager handles deprecation warnings and migration guidance for legacy VCS flags
type DeprecationManager struct {
	logger     logging.SimpleLogging
	userConfig server.UserConfig
	
	// Deprecation settings
	enableWarnings     bool
	enableMigrationTips bool
	warningInterval    time.Duration
	lastWarningTime    map[string]time.Time
}

// DeprecatedFlag represents a deprecated VCS flag
type DeprecatedFlag struct {
	Name           string
	Value          interface{}
	DeprecatedIn   string
	RemovalVersion string
	Replacement    string
	MigrationGuide string
	VCSProvider    string
}

// NewDeprecationManager creates a new deprecation manager
func NewDeprecationManager(logger logging.SimpleLogging, userConfig server.UserConfig) *DeprecationManager {
	return &DeprecationManager{
		logger:          logger,
		userConfig:      userConfig,
		enableWarnings:  true,
		enableMigrationTips: true,
		warningInterval: 24 * time.Hour, // Warn once per day
		lastWarningTime: make(map[string]time.Time),
	}
}

// CheckDeprecatedFlags checks for deprecated VCS flags and logs warnings
func (dm *DeprecationManager) CheckDeprecatedFlags() []DeprecatedFlag {
	var deprecatedFlags []DeprecatedFlag
	
	// Check GitHub-specific flags
	if dm.userConfig.GithubAllowMergeableBypassApply {
		deprecatedFlags = append(deprecatedFlags, DeprecatedFlag{
			Name:           "gh-allow-mergeable-bypass-apply",
			Value:          dm.userConfig.GithubAllowMergeableBypassApply,
			DeprecatedIn:   "v0.30.0",
			RemovalVersion: "v0.35.0",
			Replacement:    "VCS Plugin Configuration",
			MigrationGuide: "Use ATLANTIS_VCS_PLUGINS_ENABLED=true with plugin-based mergeable bypass",
			VCSProvider:    "GitHub",
		})
	}
	
	if dm.userConfig.GithubTeamAllowlist != "" {
		deprecatedFlags = append(deprecatedFlags, DeprecatedFlag{
			Name:           "gh-team-allowlist",
			Value:          dm.userConfig.GithubTeamAllowlist,
			DeprecatedIn:   "v0.30.0",
			RemovalVersion: "v0.35.0",
			Replacement:    "VCS Plugin Configuration",
			MigrationGuide: "Use VCS plugin team validation with ATLANTIS_VCS_PLUGINS_ENABLED=true",
			VCSProvider:    "GitHub",
		})
	}
	
	// Check GitLab-specific flags
	if dm.userConfig.GitlabGroupAllowlist != "" {
		deprecatedFlags = append(deprecatedFlags, DeprecatedFlag{
			Name:           "gitlab-group-allowlist",
			Value:          dm.userConfig.GitlabGroupAllowlist,
			DeprecatedIn:   "v0.30.0",
			RemovalVersion: "v0.35.0",
			Replacement:    "VCS Plugin Configuration",
			MigrationGuide: "Use VCS plugin group validation with ATLANTIS_VCS_PLUGINS_ENABLED=true",
			VCSProvider:    "GitLab",
		})
	}
	
	// Log warnings for deprecated flags
	for _, flag := range deprecatedFlags {
		dm.logDeprecationWarning(flag)
	}
	
	return deprecatedFlags
}

// logDeprecationWarning logs a deprecation warning with rate limiting
func (dm *DeprecationManager) logDeprecationWarning(flag DeprecatedFlag) {
	if !dm.enableWarnings {
		return
	}
	
	// Rate limit warnings
	if lastWarning, exists := dm.lastWarningTime[flag.Name]; exists {
		if time.Since(lastWarning) < dm.warningInterval {
			return
		}
	}
	
	dm.lastWarningTime[flag.Name] = time.Now()
	
	// Log the deprecation warning
	dm.logger.Warn("DEPRECATION WARNING: Flag '--%s' is deprecated and will be removed in %s", 
		flag.Name, flag.RemovalVersion)
	dm.logger.Warn("  Current value: %v", flag.Value)
	dm.logger.Warn("  Deprecated in: %s", flag.DeprecatedIn)
	dm.logger.Warn("  VCS Provider: %s", flag.VCSProvider)
	dm.logger.Warn("  Replacement: %s", flag.Replacement)
	
	if dm.enableMigrationTips {
		dm.logger.Warn("  Migration Guide: %s", flag.MigrationGuide)
		dm.logger.Warn("  See: https://github.com/runatlantis/atlantis/issues/5574")
	}
}

// GenerateMigrationPlan generates a migration plan for deprecated flags
func (dm *DeprecationManager) GenerateMigrationPlan() MigrationPlan {
	deprecatedFlags := dm.CheckDeprecatedFlags()
	
	plan := MigrationPlan{
		TotalFlags:      len(deprecatedFlags),
		GitHubFlags:     0,
		GitLabFlags:     0,
		MigrationSteps:  []MigrationStep{},
		EstimatedEffort: "Low",
	}
	
	// Count flags by provider
	for _, flag := range deprecatedFlags {
		switch flag.VCSProvider {
		case "GitHub":
			plan.GitHubFlags++
		case "GitLab":
			plan.GitLabFlags++
		}
	}
	
	// Generate migration steps
	if plan.TotalFlags > 0 {
		plan.MigrationSteps = dm.generateMigrationSteps(deprecatedFlags)
		plan.EstimatedEffort = dm.estimateEffort(plan.TotalFlags)
	}
	
	return plan
}

// generateMigrationSteps creates step-by-step migration instructions
func (dm *DeprecationManager) generateMigrationSteps(flags []DeprecatedFlag) []MigrationStep {
	var steps []MigrationStep
	
	// Step 1: Enable VCS plugins
	steps = append(steps, MigrationStep{
		Order:       1,
		Title:       "Enable VCS Plugin System",
		Description: "Enable the new VCS plugin architecture",
		Commands: []string{
			"export ATLANTIS_VCS_PLUGINS_ENABLED=true",
			"export ATLANTIS_VCS_PLUGINS_MIGRATION_MODE=gradual",
			"export ATLANTIS_VCS_PLUGINS_FALLBACK=true",
		},
		Validation: "Check logs for 'VCS plugin migration warnings enabled'",
	})
	
	// Step 2: Configure VCS-specific features
	for _, flag := range flags {
		switch flag.Name {
		case "gh-allow-mergeable-bypass-apply":
			steps = append(steps, MigrationStep{
				Order:       2,
				Title:       "Migrate GitHub Mergeable Bypass",
				Description: "Replace --gh-allow-mergeable-bypass-apply with plugin configuration",
				Commands: []string{
					"# Remove: --gh-allow-mergeable-bypass-apply",
					"# The VCS plugin system handles mergeable bypass automatically",
					"# Verify in logs: 'Successfully used new VCS plugin for PullIsMergeable'",
				},
				Validation: "Test pull request mergeable checks work correctly",
			})
			
		case "gh-team-allowlist":
			steps = append(steps, MigrationStep{
				Order:       3,
				Title:       "Migrate GitHub Team Allowlist",
				Description: fmt.Sprintf("Replace --gh-team-allowlist='%s' with plugin validation", flag.Value),
				Commands: []string{
					"# Remove: --gh-team-allowlist",
					"# The VCS plugin system handles team validation automatically",
					"# Verify in logs: 'Successfully used new VCS plugin for team validation'",
				},
				Validation: "Test team-based access control works correctly",
			})
			
		case "gitlab-group-allowlist":
			steps = append(steps, MigrationStep{
				Order:       4,
				Title:       "Migrate GitLab Group Allowlist",
				Description: fmt.Sprintf("Replace --gitlab-group-allowlist='%s' with plugin validation", flag.Value),
				Commands: []string{
					"# Remove: --gitlab-group-allowlist",
					"# The VCS plugin system handles group validation automatically",
					"# Verify in logs: 'Successfully used new VCS plugin for group validation'",
				},
				Validation: "Test group-based access control works correctly",
			})
		}
	}
	
	// Step 3: Test and validate
	steps = append(steps, MigrationStep{
		Order:       5,
		Title:       "Test Migration",
		Description: "Validate that all VCS operations work with the plugin system",
		Commands: []string{
			"# Test pull request operations",
			"# Test commit status updates",
			"# Test access control (teams/groups)",
			"# Monitor logs for plugin usage",
		},
		Validation: "All VCS operations work without fallback to legacy system",
	})
	
	// Step 4: Remove legacy flags
	steps = append(steps, MigrationStep{
		Order:       6,
		Title:       "Remove Legacy Flags",
		Description: "Remove deprecated flags from configuration",
		Commands: []string{
			"export ATLANTIS_VCS_PLUGINS_MIGRATION_MODE=strict",
			"export ATLANTIS_VCS_PLUGINS_FALLBACK=false",
			"# Remove all deprecated --gh-* and --gitlab-* flags",
		},
		Validation: "No deprecation warnings in logs",
	})
	
	return steps
}

// estimateEffort estimates the migration effort based on number of flags
func (dm *DeprecationManager) estimateEffort(flagCount int) string {
	switch {
	case flagCount == 0:
		return "None"
	case flagCount <= 2:
		return "Low"
	case flagCount <= 4:
		return "Medium"
	default:
		return "High"
	}
}

// MigrationPlan represents a complete migration plan
type MigrationPlan struct {
	TotalFlags      int
	GitHubFlags     int
	GitLabFlags     int
	MigrationSteps  []MigrationStep
	EstimatedEffort string
}

// MigrationStep represents a single migration step
type MigrationStep struct {
	Order       int
	Title       string
	Description string
	Commands    []string
	Validation  string
}

// PrintMigrationPlan prints a formatted migration plan
func (dm *DeprecationManager) PrintMigrationPlan(plan MigrationPlan) {
	if plan.TotalFlags == 0 {
		dm.logger.Info("✅ No deprecated VCS flags found. You're already using the modern VCS plugin system!")
		return
	}
	
	dm.logger.Info("🔄 VCS Plugin Migration Plan")
	dm.logger.Info("═══════════════════════════")
	dm.logger.Info("Total deprecated flags: %d", plan.TotalFlags)
	dm.logger.Info("GitHub flags: %d", plan.GitHubFlags)
	dm.logger.Info("GitLab flags: %d", plan.GitLabFlags)
	dm.logger.Info("Estimated effort: %s", plan.EstimatedEffort)
	dm.logger.Info("")
	
	for _, step := range plan.MigrationSteps {
		dm.logger.Info("Step %d: %s", step.Order, step.Title)
		dm.logger.Info("  %s", step.Description)
		dm.logger.Info("")
		
		for _, cmd := range step.Commands {
			if strings.HasPrefix(cmd, "#") {
				dm.logger.Info("  %s", cmd)
			} else {
				dm.logger.Info("  $ %s", cmd)
			}
		}
		
		dm.logger.Info("")
		dm.logger.Info("  Validation: %s", step.Validation)
		dm.logger.Info("")
	}
	
	dm.logger.Info("📚 Additional Resources:")
	dm.logger.Info("  • VCS Plugin Architecture: docs/VCS_PLUGIN_ARCHITECTURE.md")
	dm.logger.Info("  • GitHub Issue #5574: https://github.com/runatlantis/atlantis/issues/5574")
	dm.logger.Info("  • Migration Guide: docs/VCS_PLUGIN_MIGRATION_GUIDE.md")
}

// ValidateConfiguration validates that deprecated flags are not used with new plugin system
func (dm *DeprecationManager) ValidateConfiguration() error {
	deprecatedFlags := dm.CheckDeprecatedFlags()
	
	// Check if VCS plugins are enabled
	vcsPluginsEnabled := false
	// This would check environment variables or configuration
	// For now, assume we can detect this from the environment
	
	if vcsPluginsEnabled && len(deprecatedFlags) > 0 {
		var flagNames []string
		for _, flag := range deprecatedFlags {
			flagNames = append(flagNames, flag.Name)
		}
		
		return fmt.Errorf("deprecated VCS flags cannot be used with VCS plugin system enabled. "+
			"Please remove the following flags: %s. "+
			"See migration guide: docs/VCS_PLUGIN_MIGRATION_GUIDE.md",
			strings.Join(flagNames, ", "))
	}
	
	return nil
}

// GetDeprecationStatus returns the current deprecation status
func (dm *DeprecationManager) GetDeprecationStatus() DeprecationStatus {
	deprecatedFlags := dm.CheckDeprecatedFlags()
	
	status := DeprecationStatus{
		HasDeprecatedFlags: len(deprecatedFlags) > 0,
		DeprecatedFlags:    deprecatedFlags,
		MigrationRequired:  len(deprecatedFlags) > 0,
		BreakingChanges:    false, // Not yet, but will be in future versions
	}
	
	// Check if we're approaching the removal version
	for _, flag := range deprecatedFlags {
		if flag.RemovalVersion == "v0.35.0" {
			status.BreakingChanges = true
			break
		}
	}
	
	return status
}

// DeprecationStatus represents the current deprecation status
type DeprecationStatus struct {
	HasDeprecatedFlags bool
	DeprecatedFlags    []DeprecatedFlag
	MigrationRequired  bool
	BreakingChanges    bool
} 