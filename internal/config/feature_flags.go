package config

import (
	"os"
	"strconv"
)

// FeatureFlags controls which features use the new architecture
type FeatureFlags struct {
	UseNewProjectManagement bool
	UseNewPlanCommand       bool
	UseNewApplyCommand      bool
	EnableMetrics          bool
}

// LoadFeatureFlags loads feature flags from environment variables
func LoadFeatureFlags() FeatureFlags {
	return FeatureFlags{
		UseNewProjectManagement: getBoolEnv("ATLANTIS_NEW_PROJECT_MANAGEMENT", false),
		UseNewPlanCommand:       getBoolEnv("ATLANTIS_NEW_PLAN_COMMAND", false),
		UseNewApplyCommand:      getBoolEnv("ATLANTIS_NEW_APPLY_COMMAND", false),
		EnableMetrics:          getBoolEnv("ATLANTIS_ENABLE_METRICS", true),
	}
}

func getBoolEnv(key string, defaultValue bool) bool {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return defaultValue
	}
	
	return parsed
} 