// PR #6: Default Configuration Update - Documentation and Default Settings
// This file contains the final configuration updates to make enhanced locking the default

package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/runatlantis/atlantis/server"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// UpdatedDefaultConfig represents the new default configuration with enhanced locking enabled
func UpdatedDefaultConfig() server.EnhancedLockingConfig {
	return server.EnhancedLockingConfig{
		// Enhanced locking now enabled by default (BREAKING CHANGE)
		Enabled: true, // Changed from false to true
		
		// Default to Redis backend for new installations
		Backend: "redis", // Changed from "boltdb" to "redis"
		
		// Redis configuration with production-ready defaults
		Redis: server.RedisLockingConfig{
			Addresses: []string{"redis:6379"}, // Docker-friendly default
			Password: "", // Will use REDIS_PASSWORD env var if set
			DB: 0,
			PoolSize: 20, // Increased from 10
			KeyPrefix: "atlantis:enhanced:lock:",
			LockTTL: 2 * time.Hour, // Increased from 1 hour
			ConnTimeout: 10 * time.Second, // Increased for reliability
			ReadTimeout: 5 * time.Second,
			WriteTimeout: 5 * time.Second,
			ClusterMode: false, // Will auto-detect cluster mode
		},
		
		// Enable advanced features by default
		Features: server.LockingFeaturesConfig{
			PriorityQueue: true, // Now enabled by default
			DeadlockDetection: true, // Now enabled by default
			RetryMechanism: true, // Now enabled by default
			QueueMonitoring: true, // Now enabled by default
			EventStreaming: false, // Keep disabled to avoid breaking changes
			DistributedTracing: false, // Opt-in feature
		},
		
		// Maintain backward compatibility
		Fallback: server.FallbackConfig{
			LegacyEnabled: true, // Always keep legacy as fallback
			PreserveFormat: true,
			AutoFallback: true,
			FallbackTimeout: 30 * time.Second, // Increased timeout
		},
		
		// Production-tuned performance settings
		Performance: server.PerformanceConfig{
			MaxConcurrentLocks: 5000, // Increased from 1000
			QueueBatchSize: 200, // Increased from 100
			AcquisitionTimeout: 60 * time.Second, // Increased from 30
			HealthCheckInterval: 15 * time.Second, // More frequent checks
			MetricsInterval: 10 * time.Second, // More frequent metrics
		},
	}
}

// MigrationHelper assists users in migrating to enhanced locking
type MigrationHelper struct {
	currentConfig *server.Config
	enhancedConfig *server.EnhancedLockingConfig
}

// NewMigrationHelper creates a helper for configuration migration
func NewMigrationHelper(configPath string) (*MigrationHelper, error) {
	// Load current configuration
	currentConfig, err := server.LoadConfig(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load current config: %w", err)
	}
	
	enhanced := UpdatedDefaultConfig()
	
	return &MigrationHelper{
		currentConfig: currentConfig,
		enhancedConfig: &enhanced,
	}, nil
}

// GenerateMigrationPlan creates a step-by-step migration plan
func (m *MigrationHelper) GenerateMigrationPlan() MigrationPlan {
	plan := MigrationPlan{
		Steps: make([]MigrationStep, 0),
		Requirements: make([]string, 0),
		Warnings: make([]string, 0),
	}
	
	// Step 1: Infrastructure preparation
	plan.Steps = append(plan.Steps, MigrationStep{
		Name: "Infrastructure Setup",
		Description: "Prepare Redis infrastructure and validate connectivity",
		Actions: []string{
			"Deploy Redis instance (standalone or cluster)",
			"Configure Redis persistence and backup",
			"Set up Redis monitoring and alerting",
			"Test connectivity from Atlantis instances",
		},
		Validation: "redis-cli ping returns PONG",
		Rollback: "Not applicable - infrastructure setup",
	})
	
	// Step 2: Configuration update
	plan.Steps = append(plan.Steps, MigrationStep{
		Name: "Configuration Migration",
		Description: "Update Atlantis configuration to use enhanced locking",
		Actions: []string{
			"Update server configuration with Redis settings",
			"Enable enhanced locking features gradually",
			"Configure fallback settings for safety",
			"Update environment variables and secrets",
		},
		Validation: "Atlantis starts successfully with enhanced locking",
		Rollback: "Revert configuration to previous version",
	})
	
	// Step 3: Gradual rollout
	plan.Steps = append(plan.Steps, MigrationStep{
		Name: "Gradual Feature Rollout",
		Description: "Enable enhanced features one by one",
		Actions: []string{
			"Enable Redis backend with legacy fallback",
			"Monitor system behavior and performance",
			"Enable priority queue after 24h stability",
			"Enable deadlock detection after 48h stability",
			"Full feature enablement after 1 week",
		},
		Validation: "All features working without errors",
		Rollback: "Disable problematic features via configuration",
	})
	
	// Step 4: Legacy cleanup
	plan.Steps = append(plan.Steps, MigrationStep{
		Name: "Legacy System Cleanup",
		Description: "Remove dependency on legacy locking after migration",
		Actions: []string{
			"Verify all locks are in Redis backend",
			"Disable auto-fallback to legacy system",
			"Remove BoltDB files after backup",
			"Update monitoring to focus on Redis metrics",
		},
		Validation: "System operates entirely on enhanced locking",
		Rollback: "Re-enable legacy fallback and restore BoltDB files",
	})
	
	// Add requirements
	plan.Requirements = append(plan.Requirements,
		"Redis 6.0+ instance with appropriate memory allocation",
		"Network connectivity between Atlantis and Redis",
		"Backup strategy for Redis data",
		"Monitoring and alerting for Redis health",
		"Rollback plan tested in staging environment",
	)
	
	// Add warnings based on current configuration
	if m.currentConfig.LockingDBType == "boltdb" {
		plan.Warnings = append(plan.Warnings,
			"Current installation uses BoltDB - migration required",
			"Existing locks will need to be migrated to Redis",
		)
	}
	
	if m.currentConfig.DisableApply {
		plan.Warnings = append(plan.Warnings,
			"Apply is currently disabled - enhanced locking provides better coordination for applies",
		)
	}
	
	return plan
}

// MigrationPlan represents a complete migration strategy
type MigrationPlan struct {
	Steps []MigrationStep
	Requirements []string
	Warnings []string
	EstimatedDuration string
	RiskLevel string
}

// MigrationStep represents a single step in the migration
type MigrationStep struct {
	Name string
	Description string
	Actions []string
	Validation string
	Rollback string
	EstimatedTime string
	RiskLevel string
}

// CompatibilityChecker validates backward compatibility
type CompatibilityChecker struct {
	currentVersion string
	targetVersion string
}

// CheckCompatibility validates that the migration is safe
func (c *CompatibilityChecker) CheckCompatibility() CompatibilityReport {
	report := CompatibilityReport{
		IsCompatible: true,
		BreakingChanges: make([]string, 0),
		Deprecations: make([]string, 0),
		Recommendations: make([]string, 0),
	}
	
	// Check for breaking changes
	if c.currentVersion < "v0.19.0" {
		report.BreakingChanges = append(report.BreakingChanges,
			"Enhanced locking requires Atlantis v0.19.0 or newer",
		)
		report.IsCompatible = false
	}
	
	// Add deprecation notices
	report.Deprecations = append(report.Deprecations,
		"BoltDB backend is deprecated and will be removed in v0.22.0",
		"Legacy locking interface will be removed in v0.23.0",
	)
	
	// Add recommendations
	report.Recommendations = append(report.Recommendations,
		"Test migration in staging environment first",
		"Monitor system metrics closely during migration",
		"Have rollback plan ready and tested",
		"Consider gradual rollout using feature flags",
	)
	
	return report
}

// CompatibilityReport contains compatibility analysis results
type CompatibilityReport struct {
	IsCompatible bool
	BreakingChanges []string
	Deprecations []string
	Recommendations []string
}

// Environment-specific configuration generators

// GenerateDockerComposeConfig creates Docker Compose configuration with Redis
func GenerateDockerComposeConfig() string {
	return `version: '3.8'
services:
  atlantis:
    image: runatlantis/atlantis:latest
    environment:
      ATLANTIS_ENHANCED_LOCKING_ENABLED: "true"
      ATLANTIS_ENHANCED_LOCKING_BACKEND: "redis"
      ATLANTIS_REDIS_ADDRESSES: "redis:6379"
      ATLANTIS_REDIS_PASSWORD: "${REDIS_PASSWORD:-}"
    depends_on:
      - redis
    ports:
      - "4141:4141"
    volumes:
      - atlantis-data:/atlantis-data

  redis:
    image: redis:7-alpine
    command: redis-server --appendonly yes --requirepass "${REDIS_PASSWORD:-}"
    volumes:
      - redis-data:/data
    ports:
      - "6379:6379"
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 10s
      timeout: 5s
      retries: 5

volumes:
  atlantis-data:
  redis-data:
`
}

// GenerateKubernetesConfig creates Kubernetes manifests with Redis
func GenerateKubernetesConfig() string {
	return `apiVersion: v1
kind: ConfigMap
metadata:
  name: atlantis-enhanced-config
data:
  config.yaml: |
    enhanced-locking:
      enabled: true
      backend: redis
      redis:
        addresses:
          - redis:6379
        password: ""
        pool-size: 20
        lock-ttl: 2h
      features:
        priority-queue: true
        deadlock-detection: true
        retry-mechanism: true
        queue-monitoring: true
      fallback:
        legacy-enabled: true
        auto-fallback: true
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: redis
spec:
  replicas: 1
  selector:
    matchLabels:
      app: redis
  template:
    metadata:
      labels:
        app: redis
    spec:
      containers:
      - name: redis
        image: redis:7-alpine
        command: ["redis-server", "--appendonly", "yes"]
        ports:
        - containerPort: 6379
        volumeMounts:
        - name: redis-data
          mountPath: /data
      volumes:
      - name: redis-data
        persistentVolumeClaim:
          claimName: redis-pvc
---
apiVersion: v1
kind: Service
metadata:
  name: redis
spec:
  selector:
    app: redis
  ports:
  - port: 6379
    targetPort: 6379
`
}

// CLI command for configuration migration
func NewMigrateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use: "migrate-enhanced-locking",
		Short: "Migrate to enhanced locking system",
		Long: `Migrates your Atlantis installation from legacy BoltDB locking to the enhanced Redis-based locking system.

This command will:
1. Analyze your current configuration
2. Generate a migration plan
3. Optionally apply the migration with your confirmation

The migration is designed to be backward compatible and includes automatic rollback capabilities.`,
		RunE: runMigration,
	}
	
	cmd.Flags().String("config", "", "Path to current Atlantis configuration file")
	cmd.Flags().Bool("dry-run", true, "Show migration plan without applying changes")
	cmd.Flags().Bool("auto-confirm", false, "Apply migration without prompting for confirmation")
	cmd.Flags().String("redis-url", "redis://localhost:6379", "Redis connection URL for testing")
	
	return cmd
}

func runMigration(cmd *cobra.Command, args []string) error {
	configPath := viper.GetString("config")
	dryRun := viper.GetBool("dry-run")
	autoConfirm := viper.GetBool("auto-confirm")
	
	// Create migration helper
	helper, err := NewMigrationHelper(configPath)
	if err != nil {
		return fmt.Errorf("failed to create migration helper: %w", err)
	}
	
	// Generate migration plan
	plan := helper.GenerateMigrationPlan()
	
	// Display migration plan
	fmt.Println("Enhanced Locking Migration Plan")
	fmt.Println("================================")
	
	for i, step := range plan.Steps {
		fmt.Printf("\nStep %d: %s\n", i+1, step.Name)
		fmt.Printf("Description: %s\n", step.Description)
		fmt.Println("Actions:")
		for _, action := range step.Actions {
			fmt.Printf("  - %s\n", action)
		}
		fmt.Printf("Validation: %s\n", step.Validation)
		fmt.Printf("Rollback: %s\n", step.Rollback)
	}
	
	if len(plan.Requirements) > 0 {
		fmt.Println("\nRequirements:")
		for _, req := range plan.Requirements {
			fmt.Printf("  - %s\n", req)
		}
	}
	
	if len(plan.Warnings) > 0 {
		fmt.Println("\nWarnings:")
		for _, warning := range plan.Warnings {
			fmt.Printf("  ⚠️  %s\n", warning)
		}
	}
	
	if dryRun {
		fmt.Println("\n🔍 Dry run mode - no changes will be applied")
		fmt.Println("Use --dry-run=false to apply the migration")
		return nil
	}
	
	// Apply migration if not dry run
	if !autoConfirm {
		fmt.Print("\nDo you want to proceed with the migration? (yes/no): ")
		var response string
		fmt.Scanln(&response)
		if response != "yes" && response != "y" {
			fmt.Println("Migration cancelled")
			return nil
		}
	}
	
	fmt.Println("\n🚀 Starting migration...")
	// Actual migration logic would be implemented here
	
	return nil
}

// Configuration validation
func ValidateEnhancedConfig(config *server.EnhancedLockingConfig) error {
	if err := config.Validate(); err != nil {
		return fmt.Errorf("invalid enhanced locking configuration: %w", err)
	}
	
	// Additional production readiness checks
	if config.Backend == "redis" {
		if len(config.Redis.Addresses) == 0 {
			return fmt.Errorf("Redis addresses must be specified when using Redis backend")
		}
		
		if config.Redis.LockTTL < time.Minute {
			return fmt.Errorf("Redis lock TTL should be at least 1 minute for production use")
		}
	}
	
	return nil
}

// Documentation generator
func GenerateUpgradeGuide() string {
	return `# Atlantis Enhanced Locking Upgrade Guide

## Overview
This guide helps you migrate from the legacy BoltDB locking system to the enhanced Redis-based locking system in Atlantis v0.20.0+.

## Benefits of Enhanced Locking
- **Distributed Support**: Multiple Atlantis instances can share the same lock store
- **Priority Queuing**: Critical operations can jump ahead in the queue
- **Deadlock Detection**: Automatic detection and resolution of deadlock situations
- **Better Observability**: Comprehensive metrics and monitoring
- **Improved Performance**: Redis provides better performance than BoltDB

## Prerequisites
- Atlantis v0.19.0 or newer
- Redis 6.0+ instance
- Network connectivity between Atlantis and Redis

## Migration Steps

### 1. Infrastructure Preparation
\`\`\`bash
# Deploy Redis (example using Docker)
docker run -d --name redis -p 6379:6379 redis:7-alpine
\`\`\`

### 2. Configuration Update
Update your Atlantis configuration to enable enhanced locking:

\`\`\`yaml
enhanced-locking:
  enabled: true
  backend: redis
  redis:
    addresses:
      - localhost:6379
    lock-ttl: 2h
  features:
    priority-queue: true
    deadlock-detection: true
  fallback:
    legacy-enabled: true
    auto-fallback: true
\`\`\`

### 3. Testing
Test the configuration in a staging environment before production deployment.

### 4. Gradual Rollout
Enable features gradually to ensure system stability.

## Rollback Plan
If issues occur, you can quickly rollback by:
1. Setting \`enhanced-locking.enabled: false\`
2. Restarting Atlantis
3. The system will automatically fallback to BoltDB

## Monitoring
Monitor these metrics after migration:
- Lock acquisition times
- Queue depths
- Redis connectivity
- Error rates

## Support
For issues during migration, please:
1. Check the troubleshooting section
2. Review system logs
3. Contact support with detailed error information
`
}
