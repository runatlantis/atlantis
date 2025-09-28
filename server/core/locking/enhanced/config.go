package enhanced

import (
	"time"
)

// EnhancedConfig holds configuration for enhanced locking
type EnhancedConfig struct {
	Enabled        bool          `mapstructure:"enabled"`
	Backend        string        `mapstructure:"backend"` // "redis", "boltdb"
	DefaultTimeout time.Duration `mapstructure:"default_timeout"`
	MaxTimeout     time.Duration `mapstructure:"max_timeout"`

	// Future feature flags (disabled by default)
	EnablePriorityQueue     bool `mapstructure:"enable_priority_queue"`
	EnableRetries          bool `mapstructure:"enable_retries"`
	EnableMetrics          bool `mapstructure:"enable_metrics"`

	// Backward compatibility
	LegacyFallback       bool `mapstructure:"legacy_fallback"`
	PreserveLegacyFormat bool `mapstructure:"preserve_legacy_format"`
}

// DefaultConfig returns default configuration for enhanced locking
func DefaultConfig() *EnhancedConfig {
	return &EnhancedConfig{
		Enabled:              false, // Opt-in for backward compatibility
		Backend:              "boltdb",
		DefaultTimeout:       30 * time.Second,
		MaxTimeout:           5 * time.Minute,
		EnablePriorityQueue:  false,
		EnableRetries:        false,
		EnableMetrics:        true,
		LegacyFallback:       true,
		PreserveLegacyFormat: true,
	}
}