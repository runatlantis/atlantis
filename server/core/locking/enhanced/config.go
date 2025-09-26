// Configuration management for enhanced locking system
package enhanced

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// ConfigBuilder helps build enhanced locking configuration from various sources
type ConfigBuilder struct {
	config *Config
}

// NewConfigBuilder creates a new configuration builder with defaults
func NewConfigBuilder() *ConfigBuilder {
	return &ConfigBuilder{
		config: DefaultConfig(),
	}
}

// FromEnvironment sets configuration values from environment variables
func (cb *ConfigBuilder) FromEnvironment() *ConfigBuilder {
	// Backend configuration
	if backend := os.Getenv("ATLANTIS_ENHANCED_LOCKING_BACKEND"); backend != "" {
		cb.config.Backend = backend
	}

	// Feature flags
	if enabled := os.Getenv("ATLANTIS_ENHANCED_LOCKING_ENABLED"); enabled != "" {
		if val, err := strconv.ParseBool(enabled); err == nil {
			cb.config.EnablePriorityQueue = val
		}
	}

	if priorityQueue := os.Getenv("ATLANTIS_ENHANCED_LOCKING_PRIORITY_QUEUE"); priorityQueue != "" {
		if val, err := strconv.ParseBool(priorityQueue); err == nil {
			cb.config.EnablePriorityQueue = val
		}
	}

	if retries := os.Getenv("ATLANTIS_ENHANCED_LOCKING_RETRIES"); retries != "" {
		if val, err := strconv.ParseBool(retries); err == nil {
			cb.config.EnableRetries = val
		}
	}

	if metrics := os.Getenv("ATLANTIS_ENHANCED_LOCKING_METRICS"); metrics != "" {
		if val, err := strconv.ParseBool(metrics); err == nil {
			cb.config.EnableMetrics = val
		}
	}

	// Timeout configuration
	if timeout := os.Getenv("ATLANTIS_ENHANCED_LOCKING_DEFAULT_TIMEOUT"); timeout != "" {
		if val, err := time.ParseDuration(timeout); err == nil {
			cb.config.DefaultTimeout = val
		}
	}

	if maxTimeout := os.Getenv("ATLANTIS_ENHANCED_LOCKING_MAX_TIMEOUT"); maxTimeout != "" {
		if val, err := time.ParseDuration(maxTimeout); err == nil {
			cb.config.MaxTimeout = val
		}
	}

	// Queue configuration
	if queueSize := os.Getenv("ATLANTIS_ENHANCED_LOCKING_MAX_QUEUE_SIZE"); queueSize != "" {
		if val, err := strconv.Atoi(queueSize); err == nil {
			cb.config.MaxQueueSize = val
		}
	}

	// Anti-starvation configuration
	if threshold := os.Getenv("ATLANTIS_ENHANCED_LOCKING_STARVATION_THRESHOLD"); threshold != "" {
		if val, err := time.ParseDuration(threshold); err == nil {
			cb.config.StarvationThreshold = val
		}
	}

	if boost := os.Getenv("ATLANTIS_ENHANCED_LOCKING_MAX_PRIORITY_BOOST"); boost != "" {
		if val, err := strconv.Atoi(boost); err == nil {
			cb.config.MaxPriorityBoost = val
		}
	}

	// Redis configuration
	cb.configureRedisFromEnvironment()

	return cb
}

// configureRedisFromEnvironment sets Redis configuration from environment variables
func (cb *ConfigBuilder) configureRedisFromEnvironment() {
	if addresses := os.Getenv("ATLANTIS_ENHANCED_LOCKING_REDIS_ADDRESSES"); addresses != "" {
		cb.config.RedisConfig.Addresses = strings.Split(addresses, ",")
	}

	if password := os.Getenv("ATLANTIS_ENHANCED_LOCKING_REDIS_PASSWORD"); password != "" {
		cb.config.RedisConfig.Password = password
	}

	if db := os.Getenv("ATLANTIS_ENHANCED_LOCKING_REDIS_DB"); db != "" {
		if val, err := strconv.Atoi(db); err == nil {
			cb.config.RedisConfig.DB = val
		}
	}

	if poolSize := os.Getenv("ATLANTIS_ENHANCED_LOCKING_REDIS_POOL_SIZE"); poolSize != "" {
		if val, err := strconv.Atoi(poolSize); err == nil {
			cb.config.RedisConfig.PoolSize = val
		}
	}

	if keyPrefix := os.Getenv("ATLANTIS_ENHANCED_LOCKING_REDIS_KEY_PREFIX"); keyPrefix != "" {
		cb.config.RedisConfig.KeyPrefix = keyPrefix
	}

	if ttl := os.Getenv("ATLANTIS_ENHANCED_LOCKING_REDIS_TTL"); ttl != "" {
		if val, err := time.ParseDuration(ttl); err == nil {
			cb.config.RedisConfig.DefaultTTL = val
		}
	}

	if connTimeout := os.Getenv("ATLANTIS_ENHANCED_LOCKING_REDIS_CONN_TIMEOUT"); connTimeout != "" {
		if val, err := time.ParseDuration(connTimeout); err == nil {
			cb.config.RedisConfig.ConnectionTimeout = val
		}
	}

	if readTimeout := os.Getenv("ATLANTIS_ENHANCED_LOCKING_REDIS_READ_TIMEOUT"); readTimeout != "" {
		if val, err := time.ParseDuration(readTimeout); err == nil {
			cb.config.RedisConfig.ReadTimeout = val
		}
	}

	if writeTimeout := os.Getenv("ATLANTIS_ENHANCED_LOCKING_REDIS_WRITE_TIMEOUT"); writeTimeout != "" {
		if val, err := time.ParseDuration(writeTimeout); err == nil {
			cb.config.RedisConfig.WriteTimeout = val
		}
	}

	if clusterMode := os.Getenv("ATLANTIS_ENHANCED_LOCKING_REDIS_CLUSTER"); clusterMode != "" {
		if val, err := strconv.ParseBool(clusterMode); err == nil {
			cb.config.RedisConfig.ClusterMode = val
		}
	}
}

// WithBackend sets the backend type
func (cb *ConfigBuilder) WithBackend(backend string) *ConfigBuilder {
	cb.config.Backend = backend
	return cb
}

// WithPriorityQueue enables or disables priority queue
func (cb *ConfigBuilder) WithPriorityQueue(enabled bool) *ConfigBuilder {
	cb.config.EnablePriorityQueue = enabled
	return cb
}

// WithRetries enables or disables retry mechanism
func (cb *ConfigBuilder) WithRetries(enabled bool) *ConfigBuilder {
	cb.config.EnableRetries = enabled
	return cb
}

// WithMetrics enables or disables metrics collection
func (cb *ConfigBuilder) WithMetrics(enabled bool) *ConfigBuilder {
	cb.config.EnableMetrics = enabled
	return cb
}

// WithTimeouts sets timeout configuration
func (cb *ConfigBuilder) WithTimeouts(defaultTimeout, maxTimeout time.Duration) *ConfigBuilder {
	cb.config.DefaultTimeout = defaultTimeout
	cb.config.MaxTimeout = maxTimeout
	return cb
}

// WithQueueConfig sets queue configuration
func (cb *ConfigBuilder) WithQueueConfig(maxSize int, starvationThreshold time.Duration, maxBoost int) *ConfigBuilder {
	cb.config.MaxQueueSize = maxSize
	cb.config.StarvationThreshold = starvationThreshold
	cb.config.MaxPriorityBoost = maxBoost
	return cb
}

// WithRedisConfig sets Redis configuration
func (cb *ConfigBuilder) WithRedisConfig(config RedisConfig) *ConfigBuilder {
	cb.config.RedisConfig = config
	return cb
}

// Build builds and validates the configuration
func (cb *ConfigBuilder) Build() (*Config, error) {
	if err := cb.config.Validate(); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}
	return cb.config, nil
}

// MustBuild builds the configuration and panics on validation errors
func (cb *ConfigBuilder) MustBuild() *Config {
	config, err := cb.Build()
	if err != nil {
		panic(fmt.Sprintf("Failed to build enhanced locking configuration: %v", err))
	}
	return config
}

// GetConfig returns the current configuration (without validation)
func (cb *ConfigBuilder) GetConfig() *Config {
	return cb.config
}

// FeatureFlag represents a feature flag configuration
type FeatureFlag struct {
	Name        string
	Description string
	Default     bool
	Enabled     bool
}

// FeatureFlags manages feature flags for enhanced locking
type FeatureFlags struct {
	flags map[string]*FeatureFlag
}

// NewFeatureFlags creates a new feature flags manager with defaults
func NewFeatureFlags() *FeatureFlags {
	ff := &FeatureFlags{
		flags: make(map[string]*FeatureFlag),
	}

	// Register default feature flags
	ff.Register("enhanced_locking", "Enable enhanced locking system", false)
	ff.Register("priority_queue", "Enable priority-based queue", false)
	ff.Register("redis_backend", "Enable Redis backend", false)
	ff.Register("retry_mechanism", "Enable automatic retry mechanism", false)
	ff.Register("deadlock_detection", "Enable deadlock detection", false)
	ff.Register("event_streaming", "Enable lock event streaming", false)
	ff.Register("metrics_collection", "Enable metrics collection", true)
	ff.Register("legacy_compatibility", "Maintain compatibility with legacy system", true)

	return ff
}

// Register registers a new feature flag
func (ff *FeatureFlags) Register(name, description string, defaultValue bool) {
	ff.flags[name] = &FeatureFlag{
		Name:        name,
		Description: description,
		Default:     defaultValue,
		Enabled:     defaultValue,
	}
}

// Enable enables a feature flag
func (ff *FeatureFlags) Enable(name string) error {
	flag, exists := ff.flags[name]
	if !exists {
		return fmt.Errorf("feature flag '%s' not found", name)
	}
	flag.Enabled = true
	return nil
}

// Disable disables a feature flag
func (ff *FeatureFlags) Disable(name string) error {
	flag, exists := ff.flags[name]
	if !exists {
		return fmt.Errorf("feature flag '%s' not found", name)
	}
	flag.Enabled = false
	return nil
}

// IsEnabled checks if a feature flag is enabled
func (ff *FeatureFlags) IsEnabled(name string) bool {
	flag, exists := ff.flags[name]
	if !exists {
		return false
	}
	return flag.Enabled
}

// LoadFromEnvironment loads feature flag values from environment variables
func (ff *FeatureFlags) LoadFromEnvironment() {
	for name, flag := range ff.flags {
		envVar := "ATLANTIS_FEATURE_" + strings.ToUpper(strings.ReplaceAll(name, "_", ""))
		if value := os.Getenv(envVar); value != "" {
			if enabled, err := strconv.ParseBool(value); err == nil {
				flag.Enabled = enabled
			}
		}
	}
}

// GetAll returns all feature flags
func (ff *FeatureFlags) GetAll() map[string]*FeatureFlag {
	result := make(map[string]*FeatureFlag)
	for name, flag := range ff.flags {
		result[name] = &FeatureFlag{
			Name:        flag.Name,
			Description: flag.Description,
			Default:     flag.Default,
			Enabled:     flag.Enabled,
		}
	}
	return result
}

// GlobalConfig represents global configuration for enhanced locking that can be
// injected into Atlantis's existing configuration system
type GlobalConfig struct {
	// Enhanced locking configuration
	Enhanced *Config `mapstructure:"enhanced" json:"enhanced"`

	// Feature flags
	Features *FeatureFlags `mapstructure:"-" json:"-"`

	// Migration settings
	Migration MigrationConfig `mapstructure:"migration" json:"migration"`
}

// MigrationConfig controls the migration from legacy to enhanced locking
type MigrationConfig struct {
	// Enable gradual migration
	Enabled bool `mapstructure:"enabled" json:"enabled"`

	// Percentage of requests to route to enhanced system (0-100)
	Percentage int `mapstructure:"percentage" json:"percentage"`

	// Rollback on error
	AutoRollback bool `mapstructure:"auto_rollback" json:"auto_rollback"`

	// Rollback threshold (percentage of failures that triggers rollback)
	RollbackThreshold float64 `mapstructure:"rollback_threshold" json:"rollback_threshold"`

	// Validation mode (compare results between legacy and enhanced)
	ValidationMode bool `mapstructure:"validation_mode" json:"validation_mode"`

	// Log differences between systems
	LogDifferences bool `mapstructure:"log_differences" json:"log_differences"`
}

// DefaultGlobalConfig returns the default global configuration
func DefaultGlobalConfig() *GlobalConfig {
	return &GlobalConfig{
		Enhanced: DefaultConfig(),
		Features: NewFeatureFlags(),
		Migration: MigrationConfig{
			Enabled:           false,
			Percentage:        0,
			AutoRollback:      true,
			RollbackThreshold: 5.0, // 5% failure rate triggers rollback
			ValidationMode:    false,
			LogDifferences:    true,
		},
	}
}

// Validate validates the global configuration
func (gc *GlobalConfig) Validate() error {
	if gc.Enhanced != nil {
		if err := gc.Enhanced.Validate(); err != nil {
			return fmt.Errorf("enhanced config validation failed: %w", err)
		}
	}

	if gc.Migration.Percentage < 0 || gc.Migration.Percentage > 100 {
		return fmt.Errorf("migration percentage must be between 0 and 100, got %d", gc.Migration.Percentage)
	}

	if gc.Migration.RollbackThreshold < 0 || gc.Migration.RollbackThreshold > 100 {
		return fmt.Errorf("rollback threshold must be between 0 and 100, got %f", gc.Migration.RollbackThreshold)
	}

	return nil
}

// IsEnhancedEnabled returns true if enhanced locking should be used
func (gc *GlobalConfig) IsEnhancedEnabled() bool {
	return gc.Features.IsEnabled("enhanced_locking")
}

// IsMigrationEnabled returns true if migration is enabled
func (gc *GlobalConfig) IsMigrationEnabled() bool {
	return gc.Migration.Enabled
}

// ShouldUseEnhanced determines if a request should use enhanced locking based on migration percentage
func (gc *GlobalConfig) ShouldUseEnhanced(requestID string) bool {
	if !gc.IsEnhancedEnabled() {
		return false
	}

	if !gc.IsMigrationEnabled() {
		return true
	}

	// Use a simple hash-based approach for consistent routing
	hash := simpleHash(requestID)
	return (hash % 100) < gc.Migration.Percentage
}

// simpleHash provides a simple hash function for request routing
func simpleHash(s string) int {
	hash := 0
	for _, c := range s {
		hash = hash*31 + int(c)
	}
	if hash < 0 {
		hash = -hash
	}
	return hash
}

// ConfigProvider interface for dependency injection
type ConfigProvider interface {
	GetEnhancedConfig() *Config
	GetFeatureFlags() *FeatureFlags
	GetMigrationConfig() MigrationConfig
	IsEnhancedEnabled() bool
	ShouldUseEnhanced(requestID string) bool
}

// DefaultConfigProvider implements ConfigProvider with global config
type DefaultConfigProvider struct {
	global *GlobalConfig
}

// NewDefaultConfigProvider creates a new default config provider
func NewDefaultConfigProvider(global *GlobalConfig) *DefaultConfigProvider {
	return &DefaultConfigProvider{global: global}
}

// GetEnhancedConfig returns the enhanced locking configuration
func (p *DefaultConfigProvider) GetEnhancedConfig() *Config {
	return p.global.Enhanced
}

// GetFeatureFlags returns the feature flags
func (p *DefaultConfigProvider) GetFeatureFlags() *FeatureFlags {
	return p.global.Features
}

// GetMigrationConfig returns the migration configuration
func (p *DefaultConfigProvider) GetMigrationConfig() MigrationConfig {
	return p.global.Migration
}

// IsEnhancedEnabled returns true if enhanced locking is enabled
func (p *DefaultConfigProvider) IsEnhancedEnabled() bool {
	return p.global.IsEnhancedEnabled()
}

// ShouldUseEnhanced determines if enhanced locking should be used for a request
func (p *DefaultConfigProvider) ShouldUseEnhanced(requestID string) bool {
	return p.global.ShouldUseEnhanced(requestID)
}