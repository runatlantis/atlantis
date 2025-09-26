// PR #1: Enhanced Locking Foundation - Dependency Injection Setup
// This file shows how to wire the enhanced locking system into the existing server
// To be integrated into cmd/server.go and server/server.go

package server

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	"github.com/runatlantis/atlantis/server/core/locking"
	"github.com/runatlantis/atlantis/server/core/locking/enhanced"
	"github.com/runatlantis/atlantis/server/events"
)

// EnhancedLockingComponents holds all enhanced locking dependencies
type EnhancedLockingComponents struct {
	Config   *EnhancedLockingConfig
	Adapter  *enhanced.LockingAdapter
	Backend  locking.Backend
	Manager  *enhanced.Manager
	Reporter *enhanced.HealthReporter
}

// InitializeEnhancedLocking sets up the enhanced locking system
// This function will be called from server initialization
func InitializeEnhancedLocking(
	ctx context.Context,
	config *EnhancedLockingConfig,
	legacyBackend locking.Backend,
) (*EnhancedLockingComponents, error) {
	
	// Validate configuration first
	if err := config.Validate(); err != nil {
		return nil, errors.Wrap(err, "invalid enhanced locking configuration")
	}
	
	components := &EnhancedLockingComponents{
		Config: config,
	}
	
	// Initialize based on configuration
	if !config.IsEnhancedEnabled() {
		// Enhanced system disabled - use legacy backend directly
		components.Backend = legacyBackend
		return components, nil
	}
	
	// Enhanced system enabled - set up components
	var err error
	
	// Initialize the appropriate backend
	if config.IsRedisBackend() {
		components.Backend, err = initializeRedisBackend(ctx, &config.Redis)
		if err != nil {
			if config.Fallback.AutoFallback {
				// Fall back to legacy backend on Redis failure
				components.Backend = legacyBackend
			} else {
				return nil, errors.Wrap(err, "failed to initialize Redis backend")
			}
		}
	} else {
		// Use legacy backend (BoltDB)
		components.Backend = legacyBackend
	}
	
	// Initialize the enhanced manager
	components.Manager, err = enhanced.NewManager(
		enhanced.ManagerConfig{
			Backend: components.Backend,
			Features: enhanced.Features{
				PriorityQueue: config.ShouldUsePriorityQueue(),
				DeadlockDetection: config.ShouldDetectDeadlocks(),
				RetryMechanism: config.ShouldEnableRetry(),
				QueueMonitoring: config.Features.QueueMonitoring,
				EventStreaming: config.Features.EventStreaming,
			},
			Performance: enhanced.PerformanceConfig{
				MaxConcurrentLocks: config.Performance.MaxConcurrentLocks,
				QueueBatchSize: config.Performance.QueueBatchSize,
				AcquisitionTimeout: config.Performance.AcquisitionTimeout,
			},
		},
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to initialize enhanced manager")
	}
	
	// Initialize the adapter layer
	components.Adapter = enhanced.NewLockingAdapter(
		enhanced.AdapterConfig{
			Enhanced: components.Manager,
			Legacy: legacyBackend,
			FallbackEnabled: config.Fallback.AutoFallback,
			PreserveFormat: config.Fallback.PreserveFormat,
		},
	)
	
	// Initialize health reporter
	components.Reporter = enhanced.NewHealthReporter(
		enhanced.HealthReporterConfig{
			Manager: components.Manager,
			Interval: config.Performance.HealthCheckInterval,
		},
	)
	
	// Start background services
	if err := components.startServices(ctx); err != nil {
		return nil, errors.Wrap(err, "failed to start enhanced locking services")
	}
	
	return components, nil
}

// initializeRedisBackend creates and configures a Redis backend
func initializeRedisBackend(ctx context.Context, config *RedisLockingConfig) (locking.Backend, error) {
	// This will be implemented in PR #2
	// For now, return an error to indicate Redis is not yet available
	return nil, fmt.Errorf("Redis backend not yet implemented - will be available in PR #2")
}

// startServices starts background services for the enhanced locking system
func (c *EnhancedLockingComponents) startServices(ctx context.Context) error {
	if c.Manager != nil {
		if err := c.Manager.Start(ctx); err != nil {
			return errors.Wrap(err, "failed to start enhanced manager")
		}
	}
	
	if c.Reporter != nil {
		if err := c.Reporter.Start(ctx); err != nil {
			return errors.Wrap(err, "failed to start health reporter")
		}
	}
	
	return nil
}

// Shutdown gracefully shuts down the enhanced locking system
func (c *EnhancedLockingComponents) Shutdown(ctx context.Context) error {
	var errors []error
	
	if c.Reporter != nil {
		if err := c.Reporter.Shutdown(ctx); err != nil {
			errors = append(errors, err)
		}
	}
	
	if c.Manager != nil {
		if err := c.Manager.Shutdown(ctx); err != nil {
			errors = append(errors, err)
		}
	}
	
	if len(errors) > 0 {
		return fmt.Errorf("shutdown errors: %v", errors)
	}
	
	return nil
}

// GetLockingBackend returns the appropriate locking backend based on configuration
func (c *EnhancedLockingComponents) GetLockingBackend() locking.Backend {
	if c.Config.IsEnhancedEnabled() && c.Adapter != nil {
		return c.Adapter
	}
	return c.Backend
}

// ServerIntegration shows how to integrate with existing server structure
type ServerIntegration struct {
	// Existing server fields...
	
	// Enhanced locking components
	EnhancedLocking *EnhancedLockingComponents
	
	// Modified fields
	ProjectLocker *events.ProjectLocker // Will use enhanced backend
	WorkingDirLocker events.WorkingDirLocker // Will use enhanced backend
}

// NewServerIntegration shows how to modify server initialization
func NewServerIntegration(
	ctx context.Context,
	enhancedConfig *EnhancedLockingConfig,
	legacyBackend locking.Backend,
	// ... other existing parameters
) (*ServerIntegration, error) {
	
	// Initialize enhanced locking
	enhanced, err := InitializeEnhancedLocking(ctx, enhancedConfig, legacyBackend)
	if err != nil {
		return nil, errors.Wrap(err, "failed to initialize enhanced locking")
	}
	
	// Get the backend to use (enhanced or legacy)
	lockingBackend := enhanced.GetLockingBackend()
	
	// Create server with enhanced backend
	server := &ServerIntegration{
		EnhancedLocking: enhanced,
		// Initialize existing components with enhanced backend
		ProjectLocker: events.NewProjectLocker(lockingBackend),
		WorkingDirLocker: events.NewWorkingDirLocker(),
		// ... other existing initializations
	}
	
	return server, nil
}

// Configuration validation example for cmd/server.go integration
func ValidateEnhancedLockingFlags() error {
	// This function will validate command-line flags
	// and environment variables for enhanced locking
	return nil
}

// Feature flag helpers for gradual rollout
type FeatureFlags struct {
	EnhancedLockingEnabled bool
	RedisBackendEnabled bool
	PriorityQueueEnabled bool
	DeadlockDetectionEnabled bool
}

func GetFeatureFlags() FeatureFlags {
	// This will read from environment variables or config files
	// for gradual feature rollout
	return FeatureFlags{
		EnhancedLockingEnabled: false, // Start disabled
		RedisBackendEnabled: false,
		PriorityQueueEnabled: false,
		DeadlockDetectionEnabled: false,
	}
}
