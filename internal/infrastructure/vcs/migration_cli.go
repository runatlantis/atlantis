package vcs

import (
	"fmt"
	"os"
	"strings"

	"github.com/runatlantis/atlantis/server"
	"github.com/runatlantis/atlantis/server/logging"
)

// MigrationCLI provides command-line tools for VCS plugin migration
type MigrationCLI struct {
	deprecationManager *DeprecationManager
	logger             logging.SimpleLogging
}

// NewMigrationCLI creates a new migration CLI
func NewMigrationCLI(userConfig server.UserConfig, logger logging.SimpleLogging) *MigrationCLI {
	return &MigrationCLI{
		deprecationManager: NewDeprecationManager(logger, userConfig),
		logger:             logger,
	}
}

// RunMigrationCommand runs a migration command
func (cli *MigrationCLI) RunMigrationCommand(command string, args []string) error {
	switch command {
	case "check":
		return cli.checkDeprecatedFlags()
	case "plan":
		return cli.generateMigrationPlan()
	case "validate":
		return cli.validateConfiguration()
	case "status":
		return cli.showMigrationStatus()
	case "help":
		return cli.showHelp()
	default:
		return fmt.Errorf("unknown migration command: %s. Use 'help' for available commands", command)
	}
}

// checkDeprecatedFlags checks for deprecated flags
func (cli *MigrationCLI) checkDeprecatedFlags() error {
	cli.logger.Info("🔍 Checking for deprecated VCS flags...")
	cli.logger.Info("")
	
	deprecatedFlags := cli.deprecationManager.CheckDeprecatedFlags()
	
	if len(deprecatedFlags) == 0 {
		cli.logger.Info("✅ No deprecated VCS flags found!")
		cli.logger.Info("   You're already using the modern VCS plugin system.")
		return nil
	}
	
	cli.logger.Info("⚠️  Found %d deprecated VCS flag(s):", len(deprecatedFlags))
	cli.logger.Info("")
	
	for i, flag := range deprecatedFlags {
		cli.logger.Info("%d. Flag: --%s", i+1, flag.Name)
		cli.logger.Info("   Value: %v", flag.Value)
		cli.logger.Info("   VCS Provider: %s", flag.VCSProvider)
		cli.logger.Info("   Deprecated in: %s", flag.DeprecatedIn)
		cli.logger.Info("   Will be removed in: %s", flag.RemovalVersion)
		cli.logger.Info("   Replacement: %s", flag.Replacement)
		cli.logger.Info("")
	}
	
	cli.logger.Info("💡 Next steps:")
	cli.logger.Info("   • Run 'atlantis migrate plan' to generate a migration plan")
	cli.logger.Info("   • See: https://github.com/runatlantis/atlantis/issues/5574")
	
	return nil
}

// generateMigrationPlan generates and displays a migration plan
func (cli *MigrationCLI) generateMigrationPlan() error {
	cli.logger.Info("📋 Generating VCS plugin migration plan...")
	cli.logger.Info("")
	
	plan := cli.deprecationManager.GenerateMigrationPlan()
	cli.deprecationManager.PrintMigrationPlan(plan)
	
	return nil
}

// validateConfiguration validates the current configuration
func (cli *MigrationCLI) validateConfiguration() error {
	cli.logger.Info("✅ Validating VCS configuration...")
	cli.logger.Info("")
	
	err := cli.deprecationManager.ValidateConfiguration()
	if err != nil {
		cli.logger.Warn("❌ Configuration validation failed:")
		cli.logger.Warn("   %s", err.Error())
		return err
	}
	
	cli.logger.Info("✅ Configuration validation passed!")
	
	// Check environment variables
	cli.checkEnvironmentVariables()
	
	return nil
}

// showMigrationStatus shows the current migration status
func (cli *MigrationCLI) showMigrationStatus() error {
	cli.logger.Info("📊 VCS Plugin Migration Status")
	cli.logger.Info("═══════════════════════════════")
	cli.logger.Info("")
	
	status := cli.deprecationManager.GetDeprecationStatus()
	
	// Overall status
	if !status.HasDeprecatedFlags {
		cli.logger.Info("✅ Status: READY")
		cli.logger.Info("   No deprecated flags found")
		cli.logger.Info("   VCS plugin system can be enabled")
	} else {
		cli.logger.Info("⚠️  Status: MIGRATION REQUIRED")
		cli.logger.Info("   Found %d deprecated flag(s)", len(status.DeprecatedFlags))
		if status.BreakingChanges {
			cli.logger.Info("   ⚠️  Breaking changes coming in next major version!")
		}
	}
	
	cli.logger.Info("")
	
	// Environment status
	cli.logger.Info("🌍 Environment Configuration:")
	vcsPluginsEnabled := os.Getenv("ATLANTIS_VCS_PLUGINS_ENABLED") == "true"
	migrationMode := os.Getenv("ATLANTIS_VCS_PLUGINS_MIGRATION_MODE")
	fallbackEnabled := os.Getenv("ATLANTIS_VCS_PLUGINS_FALLBACK") != "false"
	
	cli.logger.Info("   VCS Plugins Enabled: %v", vcsPluginsEnabled)
	cli.logger.Info("   Migration Mode: %s", getOrDefault(migrationMode, "disabled"))
	cli.logger.Info("   Fallback Enabled: %v", fallbackEnabled)
	cli.logger.Info("")
	
	// Recommendations
	cli.logger.Info("💡 Recommendations:")
	if status.HasDeprecatedFlags && !vcsPluginsEnabled {
		cli.logger.Info("   1. Run 'atlantis migrate plan' to see migration steps")
		cli.logger.Info("   2. Enable VCS plugins: export ATLANTIS_VCS_PLUGINS_ENABLED=true")
		cli.logger.Info("   3. Set gradual migration: export ATLANTIS_VCS_PLUGINS_MIGRATION_MODE=gradual")
	} else if status.HasDeprecatedFlags && vcsPluginsEnabled {
		cli.logger.Info("   1. Test VCS operations with plugin system")
		cli.logger.Info("   2. Remove deprecated flags from configuration")
		cli.logger.Info("   3. Set strict mode: export ATLANTIS_VCS_PLUGINS_MIGRATION_MODE=strict")
	} else if !status.HasDeprecatedFlags && !vcsPluginsEnabled {
		cli.logger.Info("   1. Enable VCS plugins: export ATLANTIS_VCS_PLUGINS_ENABLED=true")
		cli.logger.Info("   2. Set migration mode: export ATLANTIS_VCS_PLUGINS_MIGRATION_MODE=gradual")
	} else {
		cli.logger.Info("   ✅ Configuration looks good!")
		cli.logger.Info("   Consider setting strict mode: export ATLANTIS_VCS_PLUGINS_MIGRATION_MODE=strict")
	}
	
	return nil
}

// checkEnvironmentVariables checks relevant environment variables
func (cli *MigrationCLI) checkEnvironmentVariables() {
	cli.logger.Info("🌍 Environment Variables:")
	
	envVars := []string{
		"ATLANTIS_VCS_PLUGINS_ENABLED",
		"ATLANTIS_VCS_PLUGINS_MIGRATION_MODE",
		"ATLANTIS_VCS_PLUGINS_FALLBACK",
		"ATLANTIS_VCS_PLUGINS_LOG_DEPRECATION",
		"ATLANTIS_VCS_PLUGINS_VALIDATION",
	}
	
	for _, envVar := range envVars {
		value := os.Getenv(envVar)
		if value != "" {
			cli.logger.Info("   %s=%s", envVar, value)
		} else {
			cli.logger.Info("   %s=(not set)", envVar)
		}
	}
	
	cli.logger.Info("")
}

// showHelp shows help information
func (cli *MigrationCLI) showHelp() error {
	cli.logger.Info("🔧 Atlantis VCS Plugin Migration Tool")
	cli.logger.Info("═══════════════════════════════════")
	cli.logger.Info("")
	cli.logger.Info("USAGE:")
	cli.logger.Info("   atlantis migrate <command>")
	cli.logger.Info("")
	cli.logger.Info("COMMANDS:")
	cli.logger.Info("   check      Check for deprecated VCS flags")
	cli.logger.Info("   plan       Generate a step-by-step migration plan")
	cli.logger.Info("   validate   Validate current VCS configuration")
	cli.logger.Info("   status     Show current migration status")
	cli.logger.Info("   help       Show this help message")
	cli.logger.Info("")
	cli.logger.Info("EXAMPLES:")
	cli.logger.Info("   atlantis migrate check")
	cli.logger.Info("   atlantis migrate plan")
	cli.logger.Info("   atlantis migrate status")
	cli.logger.Info("")
	cli.logger.Info("ENVIRONMENT VARIABLES:")
	cli.logger.Info("   ATLANTIS_VCS_PLUGINS_ENABLED=true")
	cli.logger.Info("   ATLANTIS_VCS_PLUGINS_MIGRATION_MODE=gradual")
	cli.logger.Info("   ATLANTIS_VCS_PLUGINS_FALLBACK=true")
	cli.logger.Info("")
	cli.logger.Info("MIGRATION MODES:")
	cli.logger.Info("   disabled   Use legacy VCS clients only (default)")
	cli.logger.Info("   opt-in     Enable specific VCS plugin operations")
	cli.logger.Info("   gradual    Use plugins with fallback to legacy")
	cli.logger.Info("   strict     Use plugins only, no fallback")
	cli.logger.Info("")
	cli.logger.Info("RESOURCES:")
	cli.logger.Info("   • GitHub Issue: https://github.com/runatlantis/atlantis/issues/5574")
	cli.logger.Info("   • Documentation: docs/VCS_PLUGIN_ARCHITECTURE.md")
	cli.logger.Info("   • Migration Guide: docs/VCS_PLUGIN_MIGRATION_GUIDE.md")
	
	return nil
}

// getOrDefault returns the value if not empty, otherwise returns the default
func getOrDefault(value, defaultValue string) string {
	if value == "" {
		return defaultValue
	}
	return value
}

// GenerateConfigurationExample generates example configuration
func (cli *MigrationCLI) GenerateConfigurationExample() string {
	var sb strings.Builder
	
	sb.WriteString("# VCS Plugin Migration Configuration Example\n")
	sb.WriteString("# ==========================================\n\n")
	
	sb.WriteString("# Phase 1: Enable VCS plugins with fallback\n")
	sb.WriteString("export ATLANTIS_VCS_PLUGINS_ENABLED=true\n")
	sb.WriteString("export ATLANTIS_VCS_PLUGINS_MIGRATION_MODE=gradual\n")
	sb.WriteString("export ATLANTIS_VCS_PLUGINS_FALLBACK=true\n")
	sb.WriteString("export ATLANTIS_VCS_PLUGINS_LOG_DEPRECATION=true\n\n")
	
	sb.WriteString("# Phase 2: Test specific operations\n")
	sb.WriteString("# export ATLANTIS_VCS_PLUGINS_ENABLE_FILE_OPS=true\n")
	sb.WriteString("# export ATLANTIS_VCS_PLUGINS_ENABLE_COMMENT_OPS=true\n\n")
	
	sb.WriteString("# Phase 3: Full migration (remove legacy flags)\n")
	sb.WriteString("# export ATLANTIS_VCS_PLUGINS_MIGRATION_MODE=strict\n")
	sb.WriteString("# export ATLANTIS_VCS_PLUGINS_FALLBACK=false\n\n")
	
	sb.WriteString("# Remove these deprecated flags:\n")
	deprecatedFlags := cli.deprecationManager.CheckDeprecatedFlags()
	for _, flag := range deprecatedFlags {
		sb.WriteString(fmt.Sprintf("# --%s (deprecated)\n", flag.Name))
	}
	
	return sb.String()
}

// ExportMigrationReport exports a detailed migration report
func (cli *MigrationCLI) ExportMigrationReport(filename string) error {
	var sb strings.Builder
	
	sb.WriteString("# Atlantis VCS Plugin Migration Report\n")
	sb.WriteString("# Generated: " + fmt.Sprintf("%v", os.Getenv("DATE")) + "\n\n")
	
	// Current status
	status := cli.deprecationManager.GetDeprecationStatus()
	sb.WriteString("## Current Status\n\n")
	sb.WriteString(fmt.Sprintf("- Deprecated flags found: %d\n", len(status.DeprecatedFlags)))
	sb.WriteString(fmt.Sprintf("- Migration required: %v\n", status.MigrationRequired))
	sb.WriteString(fmt.Sprintf("- Breaking changes coming: %v\n\n", status.BreakingChanges))
	
	// Deprecated flags
	if len(status.DeprecatedFlags) > 0 {
		sb.WriteString("## Deprecated Flags\n\n")
		for _, flag := range status.DeprecatedFlags {
			sb.WriteString(fmt.Sprintf("### --%s\n", flag.Name))
			sb.WriteString(fmt.Sprintf("- **Value**: %v\n", flag.Value))
			sb.WriteString(fmt.Sprintf("- **VCS Provider**: %s\n", flag.VCSProvider))
			sb.WriteString(fmt.Sprintf("- **Deprecated in**: %s\n", flag.DeprecatedIn))
			sb.WriteString(fmt.Sprintf("- **Removal version**: %s\n", flag.RemovalVersion))
			sb.WriteString(fmt.Sprintf("- **Replacement**: %s\n", flag.Replacement))
			sb.WriteString(fmt.Sprintf("- **Migration guide**: %s\n\n", flag.MigrationGuide))
		}
	}
	
	// Migration plan
	plan := cli.deprecationManager.GenerateMigrationPlan()
	if plan.TotalFlags > 0 {
		sb.WriteString("## Migration Plan\n\n")
		sb.WriteString(fmt.Sprintf("- **Total flags**: %d\n", plan.TotalFlags))
		sb.WriteString(fmt.Sprintf("- **GitHub flags**: %d\n", plan.GitHubFlags))
		sb.WriteString(fmt.Sprintf("- **GitLab flags**: %d\n", plan.GitLabFlags))
		sb.WriteString(fmt.Sprintf("- **Estimated effort**: %s\n\n", plan.EstimatedEffort))
		
		sb.WriteString("### Migration Steps\n\n")
		for _, step := range plan.MigrationSteps {
			sb.WriteString(fmt.Sprintf("#### Step %d: %s\n", step.Order, step.Title))
			sb.WriteString(fmt.Sprintf("%s\n\n", step.Description))
			
			sb.WriteString("**Commands:**\n")
			for _, cmd := range step.Commands {
				sb.WriteString(fmt.Sprintf("```bash\n%s\n```\n", cmd))
			}
			
			sb.WriteString(fmt.Sprintf("**Validation:** %s\n\n", step.Validation))
		}
	}
	
	// Configuration example
	sb.WriteString("## Configuration Example\n\n")
	sb.WriteString("```bash\n")
	sb.WriteString(cli.GenerateConfigurationExample())
	sb.WriteString("```\n\n")
	
	// Resources
	sb.WriteString("## Resources\n\n")
	sb.WriteString("- [GitHub Issue #5574](https://github.com/runatlantis/atlantis/issues/5574)\n")
	sb.WriteString("- [VCS Plugin Architecture](docs/VCS_PLUGIN_ARCHITECTURE.md)\n")
	sb.WriteString("- [Migration Guide](docs/VCS_PLUGIN_MIGRATION_GUIDE.md)\n")
	
	// Write to file
	return os.WriteFile(filename, []byte(sb.String()), 0644)
} 