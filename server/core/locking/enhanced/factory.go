// Factory and dependency injection for enhanced locking system
package enhanced

import (
	"fmt"

	"github.com/runatlantis/atlantis/server"
	"github.com/runatlantis/atlantis/server/core/locking"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/logging"
)

// Factory creates enhanced locking components
type Factory struct {
	logger logging.SimpleLogging
	config ConfigProvider
}

// NewFactory creates a new enhanced locking factory
func NewFactory(logger logging.SimpleLogging, config ConfigProvider) *Factory {
	return &Factory{
		logger: logger,
		config: config,
	}
}

// CreateLocker creates an enhanced locker based on configuration
func (f *Factory) CreateLocker() (locking.Locker, error) {
	if !f.config.IsEnhancedEnabled() {
		f.logger.Info("Enhanced locking disabled, falling back to legacy locker")
		return f.createLegacyLocker()
	}

	enhancedConfig := f.config.GetEnhancedConfig()

	switch enhancedConfig.Backend {
	case "redis":
		return f.createRedisLocker(enhancedConfig)
	case "boltdb":
		return f.createBoltDBLocker(enhancedConfig)
	default:
		return nil, fmt.Errorf("unsupported enhanced locking backend: %s", enhancedConfig.Backend)
	}
}

// CreateBackend creates an enhanced locking backend
func (f *Factory) CreateBackend() (Backend, error) {
	enhancedConfig := f.config.GetEnhancedConfig()

	switch enhancedConfig.Backend {
	case "redis":
		return f.createRedisBackend(enhancedConfig)
	case "boltdb":
		return f.createBoltDBBackend(enhancedConfig)
	default:
		return nil, fmt.Errorf("unsupported enhanced locking backend: %s", enhancedConfig.Backend)
	}
}

// createLegacyLocker creates a legacy locker for fallback
func (f *Factory) createLegacyLocker() (locking.Locker, error) {
	// This is a placeholder - in a real implementation, this would create
	// the existing BoltDB-based locker
	f.logger.Info("Creating legacy locker (placeholder)")
	return locking.NewNoOpLocker(), nil
}

// createRedisLocker creates a Redis-based enhanced locker
func (f *Factory) createRedisLocker(config *Config) (locking.Locker, error) {
	f.logger.Info("Creating Redis-based enhanced locker")

	// This is a placeholder - actual Redis implementation would go here
	// For now, we return a wrapper that can be implemented in future PRs
	return NewRedisLockerPlaceholder(config, f.logger), nil
}

// createBoltDBLocker creates a BoltDB-based enhanced locker
func (f *Factory) createBoltDBLocker(config *Config) (locking.Locker, error) {
	f.logger.Info("Creating BoltDB-based enhanced locker")

	// This is a placeholder - enhanced BoltDB implementation would go here
	return NewBoltDBLockerPlaceholder(config, f.logger), nil
}

// createRedisBackend creates a Redis backend
func (f *Factory) createRedisBackend(config *Config) (Backend, error) {
	f.logger.Info("Creating Redis backend")

	// Placeholder for Redis backend implementation
	return NewRedisBackendPlaceholder(config, f.logger), nil
}

// createBoltDBBackend creates a BoltDB backend
func (f *Factory) createBoltDBBackend(config *Config) (Backend, error) {
	f.logger.Info("Creating BoltDB backend")

	// Placeholder for enhanced BoltDB backend implementation
	return NewBoltDBBackendPlaceholder(config, f.logger), nil
}

// ConfigFromUserConfig converts server.UserConfig to enhanced locking Config
func ConfigFromUserConfig(userConfig server.UserConfig) *Config {
	config := DefaultConfig()

	// Map user config fields to enhanced config
	if userConfig.EnhancedLockingBackend != "" {
		config.Backend = userConfig.EnhancedLockingBackend
	}

	config.EnablePriorityQueue = userConfig.EnhancedLockingPriorityQueue
	config.EnableRetries = userConfig.EnhancedLockingRetries
	config.EnableMetrics = userConfig.EnhancedLockingMetrics

	// Configure Redis if using Redis backend
	if config.Backend == "redis" {
		if userConfig.EnhancedLockingRedisAddr != "" {
			config.RedisConfig.Addresses = []string{userConfig.EnhancedLockingRedisAddr}
		}
		config.RedisConfig.Password = userConfig.EnhancedLockingRedisPass
		config.RedisConfig.DB = userConfig.EnhancedLockingRedisDB
	}

	return config
}

// FeatureFlagsFromUserConfig creates feature flags from user config
func FeatureFlagsFromUserConfig(userConfig server.UserConfig) *FeatureFlags {
	flags := NewFeatureFlags()

	// Set enhanced locking flag from user config
	if userConfig.EnhancedLockingEnabled {
		flags.Enable("enhanced_locking")
	}

	if userConfig.EnhancedLockingPriorityQueue {
		flags.Enable("priority_queue")
	}

	if userConfig.EnhancedLockingBackend == "redis" {
		flags.Enable("redis_backend")
	}

	if userConfig.EnhancedLockingRetries {
		flags.Enable("retry_mechanism")
	}

	if userConfig.EnhancedLockingMetrics {
		flags.Enable("metrics_collection")
	}

	// Load additional flags from environment
	flags.LoadFromEnvironment()

	return flags
}

// GlobalConfigFromUserConfig creates a global config from user config
func GlobalConfigFromUserConfig(userConfig server.UserConfig) *GlobalConfig {
	return &GlobalConfig{
		Enhanced: ConfigFromUserConfig(userConfig),
		Features: FeatureFlagsFromUserConfig(userConfig),
		Migration: MigrationConfig{
			Enabled:           false, // Start with migration disabled
			Percentage:        0,
			AutoRollback:      true,
			RollbackThreshold: 5.0,
			ValidationMode:    false,
			LogDifferences:    true,
		},
	}
}

// ServiceLocator provides access to enhanced locking services
type ServiceLocator struct {
	factory        *Factory
	configProvider ConfigProvider
	backend        Backend
	locker         locking.Locker
	logger         logging.SimpleLogging
}

// NewServiceLocator creates a new service locator
func NewServiceLocator(logger logging.SimpleLogging, userConfig server.UserConfig) *ServiceLocator {
	globalConfig := GlobalConfigFromUserConfig(userConfig)
	configProvider := NewDefaultConfigProvider(globalConfig)
	factory := NewFactory(logger, configProvider)

	return &ServiceLocator{
		factory:        factory,
		configProvider: configProvider,
		logger:         logger,
	}
}

// GetLocker returns the configured locker (lazy initialization)
func (sl *ServiceLocator) GetLocker() (locking.Locker, error) {
	if sl.locker == nil {
		locker, err := sl.factory.CreateLocker()
		if err != nil {
			return nil, fmt.Errorf("failed to create locker: %w", err)
		}
		sl.locker = locker
	}
	return sl.locker, nil
}

// GetBackend returns the configured backend (lazy initialization)
func (sl *ServiceLocator) GetBackend() (Backend, error) {
	if sl.backend == nil {
		backend, err := sl.factory.CreateBackend()
		if err != nil {
			return nil, fmt.Errorf("failed to create backend: %w", err)
		}
		sl.backend = backend
	}
	return sl.backend, nil
}

// GetConfigProvider returns the config provider
func (sl *ServiceLocator) GetConfigProvider() ConfigProvider {
	return sl.configProvider
}

// IsEnhancedEnabled returns true if enhanced locking is enabled
func (sl *ServiceLocator) IsEnhancedEnabled() bool {
	return sl.configProvider.IsEnhancedEnabled()
}

// MigrationAwareLocker wraps a locker with migration logic
type MigrationAwareLocker struct {
	legacy        locking.Locker
	enhanced      locking.Locker
	config        ConfigProvider
	logger        logging.SimpleLogging
	requestCounter int
}

// NewMigrationAwareLocker creates a migration-aware locker
func NewMigrationAwareLocker(legacy, enhanced locking.Locker, config ConfigProvider, logger logging.SimpleLogging) *MigrationAwareLocker {
	return &MigrationAwareLocker{
		legacy:   legacy,
		enhanced: enhanced,
		config:   config,
		logger:   logger,
	}
}

// TryLock implements locking.Locker with migration support
func (m *MigrationAwareLocker) TryLock(project models.Project, workspace string, pull models.PullRequest, user models.User) (locking.TryLockResponse, error) {
	// Generate a request ID for migration routing
	requestID := fmt.Sprintf("%s-%s-%d", project.RepoFullName, workspace, pull.Num)
	m.requestCounter++

	if m.config.ShouldUseEnhanced(requestID) {
		m.logger.Debug("Using enhanced locker for request %s", requestID)
		return m.enhanced.TryLock(project, workspace, pull, user)
	}

	m.logger.Debug("Using legacy locker for request %s", requestID)
	return m.legacy.TryLock(project, workspace, pull, user)
}

// Unlock implements locking.Locker
func (m *MigrationAwareLocker) Unlock(key string) (*models.ProjectLock, error) {
	// For unlock operations, we need to check both systems
	// Start with enhanced if it's enabled
	if m.config.IsEnhancedEnabled() {
		lock, err := m.enhanced.Unlock(key)
		if err == nil && lock != nil {
			return lock, nil
		}
		m.logger.Debug("Enhanced unlock failed, trying legacy: %v", err)
	}

	return m.legacy.Unlock(key)
}

// List implements locking.Locker
func (m *MigrationAwareLocker) List() (map[string]models.ProjectLock, error) {
	legacyLocks, err := m.legacy.List()
	if err != nil {
		return nil, err
	}

	if !m.config.IsEnhancedEnabled() {
		return legacyLocks, nil
	}

	enhancedLocks, err := m.enhanced.List()
	if err != nil {
		m.logger.Warn("Failed to get enhanced locks, returning legacy only: %v", err)
		return legacyLocks, nil
	}

	// Merge locks from both systems
	merged := make(map[string]models.ProjectLock)
	for k, v := range legacyLocks {
		merged[k] = v
	}
	for k, v := range enhancedLocks {
		merged[k] = v
	}

	return merged, nil
}

// UnlockByPull implements locking.Locker
func (m *MigrationAwareLocker) UnlockByPull(repoFullName string, pullNum int) ([]models.ProjectLock, error) {
	var allLocks []models.ProjectLock

	// Unlock from legacy system
	legacyLocks, err := m.legacy.UnlockByPull(repoFullName, pullNum)
	if err != nil {
		m.logger.Warn("Failed to unlock from legacy system: %v", err)
	} else {
		allLocks = append(allLocks, legacyLocks...)
	}

	// Unlock from enhanced system if enabled
	if m.config.IsEnhancedEnabled() {
		enhancedLocks, err := m.enhanced.UnlockByPull(repoFullName, pullNum)
		if err != nil {
			m.logger.Warn("Failed to unlock from enhanced system: %v", err)
		} else {
			allLocks = append(allLocks, enhancedLocks...)
		}
	}

	return allLocks, nil
}

// GetLock implements locking.Locker
func (m *MigrationAwareLocker) GetLock(key string) (*models.ProjectLock, error) {
	// Try enhanced first if enabled
	if m.config.IsEnhancedEnabled() {
		lock, err := m.enhanced.GetLock(key)
		if err == nil && lock != nil {
			return lock, nil
		}
	}

	// Fall back to legacy
	return m.legacy.GetLock(key)
}