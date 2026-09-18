package vcs

import (
	"fmt"
	"log"
)

// NewFeatureValidator creates a new feature validator
func NewFeatureValidator(registry VCSRegistry) *FeatureValidator {
	return &FeatureValidator{
		registry: registry,
	}
}

// ValidateAndWarn validates features and logs warnings for unsupported ones
func (fv *FeatureValidator) ValidateAndWarn(vcsType string, config FeatureConfig) error {
	plugin, err := fv.registry.Get(vcsType)
	if err != nil {
		return err
	}

	capabilities := plugin.Capabilities()
	var warnings []string

	// Check mergeable bypass support
	if config.AllowMergeableBypass && !capabilities.SupportsMergeableBypass {
		warnings = append(warnings, 
			fmt.Sprintf("mergeable-bypass is not supported by VCS provider '%s'", vcsType))
	}

	// Check team allowlist support
	if len(config.TeamAllowlist) > 0 && !capabilities.SupportsTeamAllowlist {
		if capabilities.SupportsGroupAllowlist {
			warnings = append(warnings, 
				fmt.Sprintf("team-allowlist is not supported by VCS provider '%s', consider using group-allowlist instead", vcsType))
		} else {
			warnings = append(warnings, 
				fmt.Sprintf("team-allowlist is not supported by VCS provider '%s'", vcsType))
		}
	}

	// Check group allowlist support
	if len(config.GroupAllowlist) > 0 && !capabilities.SupportsGroupAllowlist {
		if capabilities.SupportsTeamAllowlist {
			warnings = append(warnings, 
				fmt.Sprintf("group-allowlist is not supported by VCS provider '%s', teams are used automatically", vcsType))
		} else {
			warnings = append(warnings, 
				fmt.Sprintf("group-allowlist is not supported by VCS provider '%s'", vcsType))
		}
	}

	// Log all warnings
	for _, warning := range warnings {
		log.Printf("WARN: %s", warning)
	}

	// For now, we only warn but don't fail
	// In the future, this could be configurable (strict mode vs warning mode)
	return nil
}

// GetEffectiveConfig returns the effective configuration with unsupported features filtered out
func (fv *FeatureValidator) GetEffectiveConfig(vcsType string, config FeatureConfig) (FeatureConfig, error) {
	plugin, err := fv.registry.Get(vcsType)
	if err != nil {
		return config, err
	}

	capabilities := plugin.Capabilities()
	effectiveConfig := config

	// Filter out unsupported features
	if !capabilities.SupportsMergeableBypass {
		effectiveConfig.AllowMergeableBypass = false
	}

	if !capabilities.SupportsTeamAllowlist {
		effectiveConfig.TeamAllowlist = nil
	}

	if !capabilities.SupportsGroupAllowlist {
		effectiveConfig.GroupAllowlist = nil
	}

	return effectiveConfig, nil
}

// IsFeatureSupported checks if a specific feature is supported by the VCS
func (fv *FeatureValidator) IsFeatureSupported(vcsType, feature string) (bool, error) {
	plugin, err := fv.registry.Get(vcsType)
	if err != nil {
		return false, err
	}

	capabilities := plugin.Capabilities()

	switch feature {
	case "mergeable-bypass":
		return capabilities.SupportsMergeableBypass, nil
	case "team-allowlist":
		return capabilities.SupportsTeamAllowlist, nil
	case "group-allowlist":
		return capabilities.SupportsGroupAllowlist, nil
	case "custom-fields":
		return capabilities.SupportsCustomFields, nil
	default:
		return false, fmt.Errorf("unknown feature: %s", feature)
	}
} 