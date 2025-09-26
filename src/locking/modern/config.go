package modern

import (
	"fmt"
	"time"

	"github.com/runatlantis/atlantis/server/core/locking/enhanced"
)

// ModernLockingConfig extends enhanced config with additional modern features
type ModernLockingConfig struct {
	*enhanced.EnhancedConfig

	// Advanced Redis Configuration
	Redis ModernRedisConfig `mapstructure:"redis"`

	// Fair Scheduling Configuration
	FairScheduling FairSchedulingConfig `mapstructure:"fair_scheduling"`

	// Observability Configuration
	Observability ObservabilityConfig `mapstructure:"observability"`

	// Migration Configuration
	Migration MigrationConfig `mapstructure:"migration"`

	// Security Configuration
	Security SecurityConfig `mapstructure:"security"`
}

// ModernRedisConfig provides advanced Redis configuration
type ModernRedisConfig struct {
	// Clustering Configuration
	EnableClustering     bool     `mapstructure:"enable_clustering"`
	ClusterNodes         []string `mapstructure:"cluster_nodes"`
	ClusterSlots         int      `mapstructure:"cluster_slots"`

	// High Availability Configuration
	EnableHA             bool              `mapstructure:"enable_ha"`
	SentinelNodes        []string          `mapstructure:"sentinel_nodes"`
	SentinelMasterName   string            `mapstructure:"sentinel_master_name"`

	// Connection Pooling
	MaxIdleConns         int               `mapstructure:"max_idle_conns"`
	MaxActiveConns       int               `mapstructure:"max_active_conns"`
	ConnTimeout          time.Duration     `mapstructure:"conn_timeout"`
	ReadTimeout          time.Duration     `mapstructure:"read_timeout"`
	WriteTimeout         time.Duration     `mapstructure:"write_timeout"`

	// Advanced Features
	EnablePipelining     bool              `mapstructure:"enable_pipelining"`
	PipelineSize         int               `mapstructure:"pipeline_size"`
	EnableCompression    bool              `mapstructure:"enable_compression"`
	CompressionThreshold int               `mapstructure:"compression_threshold"`

	// Consistency Configuration
	ConsistencyLevel     string            `mapstructure:"consistency_level"` // "eventual", "strong"
	ReadPreference       string            `mapstructure:"read_preference"`   // "primary", "secondary", "nearest"
}

// FairSchedulingConfig configures advanced fair scheduling
type FairSchedulingConfig struct {
	// Algorithm Configuration
	Algorithm            string            `mapstructure:"algorithm"`          // "priority", "weighted_round_robin", "lottery", "cfs"

	// Weight Configuration
	PriorityWeights      map[string]int    `mapstructure:"priority_weights"`
	UserWeights          map[string]int    `mapstructure:"user_weights"`
	ProjectWeights       map[string]int    `mapstructure:"project_weights"`

	// Time-based Configuration
	EnableTimeSlicing    bool              `mapstructure:"enable_time_slicing"`
	TimeSliceDuration    time.Duration     `mapstructure:"time_slice_duration"`
	MaxContinuousTime    time.Duration     `mapstructure:"max_continuous_time"`

	// Anti-starvation Configuration
	StarvationThreshold  time.Duration     `mapstructure:"starvation_threshold"`
	StarvationBoost      float64           `mapstructure:"starvation_boost"`

	// Load Balancing
	EnableLoadBalancing  bool              `mapstructure:"enable_load_balancing"`
	LoadBalanceWindow    time.Duration     `mapstructure:"load_balance_window"`
}

// ObservabilityConfig configures metrics and monitoring
type ObservabilityConfig struct {
	// Metrics Configuration
	EnableMetrics        bool              `mapstructure:"enable_metrics"`
	MetricsBackend       string            `mapstructure:"metrics_backend"`    // "prometheus", "statsd", "datadog"
	MetricsPrefix        string            `mapstructure:"metrics_prefix"`
	MetricsInterval      time.Duration     `mapstructure:"metrics_interval"`

	// Tracing Configuration
	EnableTracing        bool              `mapstructure:"enable_tracing"`
	TracingBackend       string            `mapstructure:"tracing_backend"`    // "jaeger", "zipkin", "datadog"
	SampleRate           float64           `mapstructure:"sample_rate"`

	// Logging Configuration
	EnableStructuredLogs bool              `mapstructure:"enable_structured_logs"`
	LogLevel             string            `mapstructure:"log_level"`
	LogFormat            string            `mapstructure:"log_format"`         // "json", "text"

	// Health Checks
	EnableHealthChecks   bool              `mapstructure:"enable_health_checks"`
	HealthCheckInterval  time.Duration     `mapstructure:"health_check_interval"`
	HealthCheckTimeout   time.Duration     `mapstructure:"health_check_timeout"`

	// Performance Monitoring
	EnableProfiling      bool              `mapstructure:"enable_profiling"`
	ProfilingPort        int               `mapstructure:"profiling_port"`
}

// MigrationConfig configures migration from legacy systems
type MigrationConfig struct {
	// Migration Strategy
	Strategy             string            `mapstructure:"strategy"`           // "gradual", "canary", "blue_green"
	MigrationPercentage  int               `mapstructure:"migration_percentage"`

	// Validation Configuration
	EnableValidation     bool              `mapstructure:"enable_validation"`
	ValidationInterval   time.Duration     `mapstructure:"validation_interval"`
	ValidationThreshold  float64           `mapstructure:"validation_threshold"`

	// Rollback Configuration
	EnableAutoRollback   bool              `mapstructure:"enable_auto_rollback"`
	RollbackThreshold    float64           `mapstructure:"rollback_threshold"`
	RollbackCooldown     time.Duration     `mapstructure:"rollback_cooldown"`

	// Data Migration
	EnableDataMigration  bool              `mapstructure:"enable_data_migration"`
	MigrationBatchSize   int               `mapstructure:"migration_batch_size"`
	MigrationDelay       time.Duration     `mapstructure:"migration_delay"`
}

// SecurityConfig configures security features
type SecurityConfig struct {
	// Authentication
	EnableAuthentication bool              `mapstructure:"enable_authentication"`
	AuthenticationMode   string            `mapstructure:"authentication_mode"` // "none", "basic", "jwt", "oauth"

	// Authorization
	EnableAuthorization  bool              `mapstructure:"enable_authorization"`
	AuthorizationRules   []AuthRule        `mapstructure:"authorization_rules"`

	// Encryption
	EnableEncryption     bool              `mapstructure:"enable_encryption"`
	EncryptionAlgorithm  string            `mapstructure:"encryption_algorithm"`
	EncryptionKey        string            `mapstructure:"encryption_key"`

	// Audit Logging
	EnableAuditLogging   bool              `mapstructure:"enable_audit_logging"`
	AuditLogBackend      string            `mapstructure:"audit_log_backend"`
	AuditLogLevel        string            `mapstructure:"audit_log_level"`

	// Rate Limiting
	EnableRateLimiting   bool              `mapstructure:"enable_rate_limiting"`
	RateLimitAlgorithm   string            `mapstructure:"rate_limit_algorithm"`
	RateLimitRules       []RateLimitRule   `mapstructure:"rate_limit_rules"`
}

// AuthRule defines an authorization rule
type AuthRule struct {
	Resource   string   `mapstructure:"resource"`
	Actions    []string `mapstructure:"actions"`
	Users      []string `mapstructure:"users"`
	Groups     []string `mapstructure:"groups"`
	Conditions string   `mapstructure:"conditions"`
}

// RateLimitRule defines a rate limiting rule
type RateLimitRule struct {
	Resource    string        `mapstructure:"resource"`
	Limit       int           `mapstructure:"limit"`
	Window      time.Duration `mapstructure:"window"`
	Users       []string      `mapstructure:"users"`
	Groups      []string      `mapstructure:"groups"`
}

// DefaultModernConfig returns a default modern locking configuration
func DefaultModernConfig() *ModernLockingConfig {
	base := enhanced.DefaultConfig()

	return &ModernLockingConfig{
		EnhancedConfig: base,
		Redis: ModernRedisConfig{
			EnableClustering:     false,
			ClusterSlots:         16384,
			EnableHA:             false,
			MaxIdleConns:         10,
			MaxActiveConns:       100,
			ConnTimeout:          5 * time.Second,
			ReadTimeout:          3 * time.Second,
			WriteTimeout:         3 * time.Second,
			EnablePipelining:     false,
			PipelineSize:         10,
			EnableCompression:    false,
			CompressionThreshold: 1024,
			ConsistencyLevel:     "strong",
			ReadPreference:       "primary",
		},
		FairScheduling: FairSchedulingConfig{
			Algorithm: "priority",
			PriorityWeights: map[string]int{
				"critical": 1000,
				"high":     100,
				"normal":   10,
				"low":      1,
			},
			UserWeights:          make(map[string]int),
			ProjectWeights:       make(map[string]int),
			EnableTimeSlicing:    false,
			TimeSliceDuration:    30 * time.Second,
			MaxContinuousTime:    5 * time.Minute,
			StarvationThreshold:  2 * time.Minute,
			StarvationBoost:      2.0,
			EnableLoadBalancing:  false,
			LoadBalanceWindow:    time.Minute,
		},
		Observability: ObservabilityConfig{
			EnableMetrics:        true,
			MetricsBackend:       "prometheus",
			MetricsPrefix:        "atlantis_modern_locking",
			MetricsInterval:      30 * time.Second,
			EnableTracing:        false,
			TracingBackend:       "jaeger",
			SampleRate:           0.1,
			EnableStructuredLogs: true,
			LogLevel:             "info",
			LogFormat:            "json",
			EnableHealthChecks:   true,
			HealthCheckInterval:  30 * time.Second,
			HealthCheckTimeout:   5 * time.Second,
			EnableProfiling:      false,
			ProfilingPort:        6060,
		},
		Migration: MigrationConfig{
			Strategy:            "gradual",
			MigrationPercentage: 10,
			EnableValidation:    true,
			ValidationInterval:  time.Minute,
			ValidationThreshold: 0.95,
			EnableAutoRollback:  true,
			RollbackThreshold:   0.90,
			RollbackCooldown:    5 * time.Minute,
			EnableDataMigration: false,
			MigrationBatchSize:  100,
			MigrationDelay:      100 * time.Millisecond,
		},
		Security: SecurityConfig{
			EnableAuthentication: false,
			AuthenticationMode:   "none",
			EnableAuthorization:  false,
			EnableEncryption:     false,
			EncryptionAlgorithm:  "AES-256-GCM",
			EnableAuditLogging:   false,
			AuditLogBackend:      "file",
			AuditLogLevel:        "info",
			EnableRateLimiting:   false,
			RateLimitAlgorithm:   "token_bucket",
		},
	}
}

// Validate validates the modern configuration
func (c *ModernLockingConfig) Validate() error {
	// Validate base configuration
	if err := c.validateEnhancedConfig(); err != nil {
		return fmt.Errorf("enhanced config validation failed: %w", err)
	}

	// Validate Redis configuration
	if err := c.validateRedisConfig(); err != nil {
		return fmt.Errorf("redis config validation failed: %w", err)
	}

	// Validate fair scheduling configuration
	if err := c.validateFairSchedulingConfig(); err != nil {
		return fmt.Errorf("fair scheduling config validation failed: %w", err)
	}

	// Validate observability configuration
	if err := c.validateObservabilityConfig(); err != nil {
		return fmt.Errorf("observability config validation failed: %w", err)
	}

	// Validate migration configuration
	if err := c.validateMigrationConfig(); err != nil {
		return fmt.Errorf("migration config validation failed: %w", err)
	}

	// Validate security configuration
	if err := c.validateSecurityConfig(); err != nil {
		return fmt.Errorf("security config validation failed: %w", err)
	}

	return nil
}

func (c *ModernLockingConfig) validateEnhancedConfig() error {
	if c.EnhancedConfig == nil {
		return fmt.Errorf("enhanced config cannot be nil")
	}
	if c.DefaultTimeout <= 0 {
		return fmt.Errorf("default timeout must be positive")
	}
	if c.MaxTimeout < c.DefaultTimeout {
		return fmt.Errorf("max timeout must be >= default timeout")
	}
	return nil
}

func (c *ModernLockingConfig) validateRedisConfig() error {
	if c.Redis.EnableClustering && len(c.Redis.ClusterNodes) == 0 {
		return fmt.Errorf("cluster nodes must be specified when clustering is enabled")
	}
	if c.Redis.EnableHA && len(c.Redis.SentinelNodes) == 0 {
		return fmt.Errorf("sentinel nodes must be specified when HA is enabled")
	}
	if c.Redis.MaxActiveConns <= 0 {
		return fmt.Errorf("max active connections must be positive")
	}
	if c.Redis.ConnTimeout <= 0 {
		return fmt.Errorf("connection timeout must be positive")
	}
	return nil
}

func (c *ModernLockingConfig) validateFairSchedulingConfig() error {
	validAlgorithms := map[string]bool{
		"priority": true, "weighted_round_robin": true, "lottery": true, "cfs": true,
	}
	if !validAlgorithms[c.FairScheduling.Algorithm] {
		return fmt.Errorf("invalid scheduling algorithm: %s", c.FairScheduling.Algorithm)
	}
	if c.FairScheduling.StarvationThreshold <= 0 {
		return fmt.Errorf("starvation threshold must be positive")
	}
	if c.FairScheduling.StarvationBoost <= 1.0 {
		return fmt.Errorf("starvation boost must be > 1.0")
	}
	return nil
}

func (c *ModernLockingConfig) validateObservabilityConfig() error {
	validMetricsBackends := map[string]bool{
		"prometheus": true, "statsd": true, "datadog": true,
	}
	if c.Observability.EnableMetrics && !validMetricsBackends[c.Observability.MetricsBackend] {
		return fmt.Errorf("invalid metrics backend: %s", c.Observability.MetricsBackend)
	}
	if c.Observability.SampleRate < 0 || c.Observability.SampleRate > 1 {
		return fmt.Errorf("sample rate must be between 0 and 1")
	}
	return nil
}

func (c *ModernLockingConfig) validateMigrationConfig() error {
	validStrategies := map[string]bool{
		"gradual": true, "canary": true, "blue_green": true,
	}
	if !validStrategies[c.Migration.Strategy] {
		return fmt.Errorf("invalid migration strategy: %s", c.Migration.Strategy)
	}
	if c.Migration.MigrationPercentage < 0 || c.Migration.MigrationPercentage > 100 {
		return fmt.Errorf("migration percentage must be between 0 and 100")
	}
	return nil
}

func (c *ModernLockingConfig) validateSecurityConfig() error {
	validAuthModes := map[string]bool{
		"none": true, "basic": true, "jwt": true, "oauth": true,
	}
	if !validAuthModes[c.Security.AuthenticationMode] {
		return fmt.Errorf("invalid authentication mode: %s", c.Security.AuthenticationMode)
	}
	return nil
}

// Clone creates a deep copy of the configuration
func (c *ModernLockingConfig) Clone() *ModernLockingConfig {
	// This is a simplified clone - in production, use a proper deep copy library
	clone := &ModernLockingConfig{}
	*clone = *c
	return clone
}

// Merge merges another configuration into this one
func (c *ModernLockingConfig) Merge(other *ModernLockingConfig) error {
	if other == nil {
		return nil
	}

	// Simple merge - in production, implement proper deep merge logic
	if other.Enabled {
		c.Enabled = other.Enabled
	}

	// Merge specific sections as needed
	if other.Redis.EnableClustering {
		c.Redis = other.Redis
	}

	return nil
}