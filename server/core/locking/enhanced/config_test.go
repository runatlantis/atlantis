package enhanced

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewConfigBuilder(t *testing.T) {
	builder := NewConfigBuilder()
	assert.NotNil(t, builder)
	assert.NotNil(t, builder.config)
	assert.Equal(t, DefaultConfig(), builder.config)
}

func TestConfigBuilderFromEnvironment(t *testing.T) {
	// Set test environment variables
	testEnvVars := map[string]string{
		"ATLANTIS_ENHANCED_LOCKING_BACKEND":             "redis",
		"ATLANTIS_ENHANCED_LOCKING_ENABLED":             "true",
		"ATLANTIS_ENHANCED_LOCKING_PRIORITY_QUEUE":      "true",
		"ATLANTIS_ENHANCED_LOCKING_RETRIES":             "true",
		"ATLANTIS_ENHANCED_LOCKING_METRICS":             "false",
		"ATLANTIS_ENHANCED_LOCKING_DEFAULT_TIMEOUT":     "45s",
		"ATLANTIS_ENHANCED_LOCKING_MAX_TIMEOUT":         "10m",
		"ATLANTIS_ENHANCED_LOCKING_MAX_QUEUE_SIZE":      "500",
		"ATLANTIS_ENHANCED_LOCKING_STARVATION_THRESHOLD": "3m",
		"ATLANTIS_ENHANCED_LOCKING_MAX_PRIORITY_BOOST":  "5",
		"ATLANTIS_ENHANCED_LOCKING_REDIS_ADDRESSES":     "redis1:6379,redis2:6379",
		"ATLANTIS_ENHANCED_LOCKING_REDIS_PASSWORD":      "secret",
		"ATLANTIS_ENHANCED_LOCKING_REDIS_DB":            "1",
		"ATLANTIS_ENHANCED_LOCKING_REDIS_POOL_SIZE":     "20",
		"ATLANTIS_ENHANCED_LOCKING_REDIS_KEY_PREFIX":    "test:lock:",
		"ATLANTIS_ENHANCED_LOCKING_REDIS_TTL":           "2h",
		"ATLANTIS_ENHANCED_LOCKING_REDIS_CONN_TIMEOUT":  "10s",
		"ATLANTIS_ENHANCED_LOCKING_REDIS_READ_TIMEOUT":  "5s",
		"ATLANTIS_ENHANCED_LOCKING_REDIS_WRITE_TIMEOUT": "5s",
		"ATLANTIS_ENHANCED_LOCKING_REDIS_CLUSTER":       "true",
	}

	// Set environment variables
	for key, value := range testEnvVars {
		os.Setenv(key, value)
	}

	// Clean up after test
	defer func() {
		for key := range testEnvVars {
			os.Unsetenv(key)
		}
	}()

	builder := NewConfigBuilder().FromEnvironment()
	config := builder.GetConfig()

	assert.Equal(t, "redis", config.Backend)
	assert.True(t, config.EnablePriorityQueue)
	assert.True(t, config.EnableRetries)
	assert.False(t, config.EnableMetrics)
	assert.Equal(t, 45*time.Second, config.DefaultTimeout)
	assert.Equal(t, 10*time.Minute, config.MaxTimeout)
	assert.Equal(t, 500, config.MaxQueueSize)
	assert.Equal(t, 3*time.Minute, config.StarvationThreshold)
	assert.Equal(t, 5, config.MaxPriorityBoost)

	// Redis config
	assert.Equal(t, []string{"redis1:6379", "redis2:6379"}, config.RedisConfig.Addresses)
	assert.Equal(t, "secret", config.RedisConfig.Password)
	assert.Equal(t, 1, config.RedisConfig.DB)
	assert.Equal(t, 20, config.RedisConfig.PoolSize)
	assert.Equal(t, "test:lock:", config.RedisConfig.KeyPrefix)
	assert.Equal(t, 2*time.Hour, config.RedisConfig.DefaultTTL)
	assert.Equal(t, 10*time.Second, config.RedisConfig.ConnectionTimeout)
	assert.Equal(t, 5*time.Second, config.RedisConfig.ReadTimeout)
	assert.Equal(t, 5*time.Second, config.RedisConfig.WriteTimeout)
	assert.True(t, config.RedisConfig.ClusterMode)
}

func TestConfigBuilderChaining(t *testing.T) {
	redisConfig := RedisConfig{
		Addresses:  []string{"localhost:6379"},
		Password:   "test",
		DB:         2,
		PoolSize:   15,
		DefaultTTL: 30 * time.Minute,
	}

	config, err := NewConfigBuilder().
		WithBackend("redis").
		WithPriorityQueue(true).
		WithRetries(true).
		WithMetrics(false).
		WithTimeouts(60*time.Second, 15*time.Minute).
		WithQueueConfig(2000, 5*time.Minute, 7).
		WithRedisConfig(redisConfig).
		Build()

	require.NoError(t, err)
	assert.Equal(t, "redis", config.Backend)
	assert.True(t, config.EnablePriorityQueue)
	assert.True(t, config.EnableRetries)
	assert.False(t, config.EnableMetrics)
	assert.Equal(t, 60*time.Second, config.DefaultTimeout)
	assert.Equal(t, 15*time.Minute, config.MaxTimeout)
	assert.Equal(t, 2000, config.MaxQueueSize)
	assert.Equal(t, 5*time.Minute, config.StarvationThreshold)
	assert.Equal(t, 7, config.MaxPriorityBoost)
	assert.Equal(t, redisConfig, config.RedisConfig)
}

func TestConfigBuilderValidationError(t *testing.T) {
	_, err := NewConfigBuilder().
		WithBackend("invalid").
		Build()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "configuration validation failed")
}

func TestMustBuildPanic(t *testing.T) {
	assert.Panics(t, func() {
		NewConfigBuilder().
			WithBackend("invalid").
			MustBuild()
	})
}

func TestNewFeatureFlags(t *testing.T) {
	flags := NewFeatureFlags()
	assert.NotNil(t, flags)
	assert.NotNil(t, flags.flags)

	// Check default flags are registered
	assert.False(t, flags.IsEnabled("enhanced_locking"))
	assert.False(t, flags.IsEnabled("priority_queue"))
	assert.False(t, flags.IsEnabled("redis_backend"))
	assert.True(t, flags.IsEnabled("metrics_collection"))
	assert.True(t, flags.IsEnabled("legacy_compatibility"))
}

func TestFeatureFlagsOperations(t *testing.T) {
	flags := NewFeatureFlags()

	// Test enable
	err := flags.Enable("enhanced_locking")
	assert.NoError(t, err)
	assert.True(t, flags.IsEnabled("enhanced_locking"))

	// Test disable
	err = flags.Disable("enhanced_locking")
	assert.NoError(t, err)
	assert.False(t, flags.IsEnabled("enhanced_locking"))

	// Test unknown flag
	err = flags.Enable("unknown_flag")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")

	// Test IsEnabled with unknown flag
	assert.False(t, flags.IsEnabled("unknown_flag"))
}

func TestFeatureFlagsFromEnvironment(t *testing.T) {
	// Set test environment variables
	os.Setenv("ATLANTIS_FEATURE_ENHANCED_LOCKING", "true")
	os.Setenv("ATLANTIS_FEATURE_PRIORITY_QUEUE", "true")
	os.Setenv("ATLANTIS_FEATURE_REDIS_BACKEND", "false")

	defer func() {
		os.Unsetenv("ATLANTIS_FEATURE_ENHANCED_LOCKING")
		os.Unsetenv("ATLANTIS_FEATURE_PRIORITY_QUEUE")
		os.Unsetenv("ATLANTIS_FEATURE_REDIS_BACKEND")
	}()

	flags := NewFeatureFlags()
	flags.LoadFromEnvironment()

	assert.True(t, flags.IsEnabled("enhanced_locking"))
	assert.True(t, flags.IsEnabled("priority_queue"))
	assert.False(t, flags.IsEnabled("redis_backend"))
}

func TestGetAllFeatureFlags(t *testing.T) {
	flags := NewFeatureFlags()
	flags.Enable("enhanced_locking")

	allFlags := flags.GetAll()
	assert.NotEmpty(t, allFlags)

	enhancedFlag, exists := allFlags["enhanced_locking"]
	assert.True(t, exists)
	assert.True(t, enhancedFlag.Enabled)
	assert.Equal(t, "enhanced_locking", enhancedFlag.Name)
}

func TestDefaultGlobalConfig(t *testing.T) {
	config := DefaultGlobalConfig()
	assert.NotNil(t, config)
	assert.NotNil(t, config.Enhanced)
	assert.NotNil(t, config.Features)
	assert.False(t, config.Migration.Enabled)
	assert.Equal(t, 0, config.Migration.Percentage)
	assert.True(t, config.Migration.AutoRollback)
	assert.Equal(t, 5.0, config.Migration.RollbackThreshold)
}

func TestGlobalConfigValidation(t *testing.T) {
	t.Run("valid default config", func(t *testing.T) {
		config := DefaultGlobalConfig()
		err := config.Validate()
		assert.NoError(t, err)
	})

	t.Run("invalid migration percentage", func(t *testing.T) {
		config := DefaultGlobalConfig()
		config.Migration.Percentage = 150
		err := config.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "migration percentage")
	})

	t.Run("invalid rollback threshold", func(t *testing.T) {
		config := DefaultGlobalConfig()
		config.Migration.RollbackThreshold = -10
		err := config.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "rollback threshold")
	})
}

func TestGlobalConfigMethods(t *testing.T) {
	config := DefaultGlobalConfig()

	// Test IsEnhancedEnabled
	assert.False(t, config.IsEnhancedEnabled())
	config.Features.Enable("enhanced_locking")
	assert.True(t, config.IsEnhancedEnabled())

	// Test IsMigrationEnabled
	assert.False(t, config.IsMigrationEnabled())
	config.Migration.Enabled = true
	assert.True(t, config.IsMigrationEnabled())

	// Test ShouldUseEnhanced
	config.Features.Enable("enhanced_locking")
	config.Migration.Enabled = true
	config.Migration.Percentage = 50

	// Test consistent routing for same request ID
	requestID := "test-request-123"
	result1 := config.ShouldUseEnhanced(requestID)
	result2 := config.ShouldUseEnhanced(requestID)
	assert.Equal(t, result1, result2)
}

func TestSimpleHash(t *testing.T) {
	// Test that hash is consistent
	input := "test-string"
	hash1 := simpleHash(input)
	hash2 := simpleHash(input)
	assert.Equal(t, hash1, hash2)

	// Test that hash is non-negative
	assert.GreaterOrEqual(t, hash1, 0)

	// Test that different inputs produce different hashes (most of the time)
	hash3 := simpleHash("different-string")
	assert.NotEqual(t, hash1, hash3)
}

func TestDefaultConfigProvider(t *testing.T) {
	globalConfig := DefaultGlobalConfig()
	provider := NewDefaultConfigProvider(globalConfig)

	assert.Equal(t, globalConfig.Enhanced, provider.GetEnhancedConfig())
	assert.Equal(t, globalConfig.Features, provider.GetFeatureFlags())
	assert.Equal(t, globalConfig.Migration, provider.GetMigrationConfig())
	assert.Equal(t, globalConfig.IsEnhancedEnabled(), provider.IsEnhancedEnabled())

	requestID := "test-request"
	assert.Equal(t, globalConfig.ShouldUseEnhanced(requestID), provider.ShouldUseEnhanced(requestID))
}

// Placeholder tests for future implementation
func TestAdvancedConfigurationIntegration(t *testing.T) {
	t.Skip("Advanced configuration integration tests - to be implemented in future PRs")
}

func TestConfigurationHotReload(t *testing.T) {
	t.Skip("Configuration hot reload tests - to be implemented in configuration management PR")
}

func TestSecretManagementIntegration(t *testing.T) {
	t.Skip("Secret management integration tests - to be implemented in security PR")
}