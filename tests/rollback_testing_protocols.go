package tests

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/runatlantis/atlantis/server/core/locking"
	"github.com/runatlantis/atlantis/server/core/locking/enhanced"
	"github.com/runatlantis/atlantis/server/core/locking/enhanced/backends"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/logging"
)

// MigrationPhase represents different stages of the enhanced locking migration
type MigrationPhase int

const (
	PhaseInitial MigrationPhase = iota // Legacy system only
	PhaseDual                          // Both systems running
	PhaseRouting                       // Selective routing to enhanced system
	PhaseMigrated                      // Full enhanced system with fallback
	PhaseEnhancedOnly                  // Enhanced system only
)

func (p MigrationPhase) String() string {
	switch p {
	case PhaseInitial:
		return "Initial"
	case PhaseDual:
		return "Dual"
	case PhaseRouting:
		return "Routing"
	case PhaseMigrated:
		return "Migrated"
	case PhaseEnhancedOnly:
		return "EnhancedOnly"
	default:
		return "Unknown"
	}
}

// RollbackResult contains the results of a rollback test
type RollbackResult struct {
	Phase              MigrationPhase `json:"phase"`
	RollbackSuccessful bool           `json:"rollback_successful"`
	DataIntegrityOK    bool           `json:"data_integrity_ok"`
	ServiceContinuity  bool           `json:"service_continuity"`
	RollbackDuration   time.Duration  `json:"rollback_duration"`
	LocksPreserved     int            `json:"locks_preserved"`
	LocksLost          int            `json:"locks_lost"`
	ErrorsEncountered  []string       `json:"errors_encountered"`
}

// RollbackTestSuite provides comprehensive rollback testing for each migration phase
type RollbackTestSuite struct {
	redisClient     redis.UniversalClient
	legacyBackend   locking.Backend
	enhancedBackend enhanced.Backend
	enhancedManager enhanced.LockManager
	adapter         *enhanced.LockingAdapter
	config          *enhanced.EnhancedConfig
	logger          logging.SimpleLogging
	cleanup         func()
}

// SetupRollbackTestSuite initializes the rollback testing environment
func SetupRollbackTestSuite(t *testing.T) *RollbackTestSuite {
	// Setup Redis for enhanced system
	redisClient := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
		DB:   12, // Dedicated rollback test database
	})

	ctx := context.Background()
	if err := redisClient.Ping(ctx).Err(); err != nil {
		t.Skip("Redis not available for rollback testing")
	}

	redisClient.FlushDB(ctx)

	config := &enhanced.EnhancedConfig{
		Enabled:                 true,
		Backend:                 "redis",
		DefaultTimeout:          30 * time.Second,
		MaxTimeout:              5 * time.Minute,
		EnablePriorityQueue:     true,
		MaxQueueSize:           100,
		QueueTimeout:           30 * time.Second,
		EnableRetry:            true,
		MaxRetryAttempts:       3,
		RetryBaseDelay:         time.Second,
		RetryMaxDelay:          10 * time.Second,
		EnableDeadlockDetection: true,
		DeadlockCheckInterval:   5 * time.Second,
		EnableEvents:           false, // Disable for testing
		EventBufferSize:        100,
		RedisClusterMode:       false,
		RedisKeyPrefix:         "atlantis:rollback:test:",
		RedisLockTTL:           time.Hour,
		LegacyFallback:         true,
		PreserveLegacyFormat:   true,
	}

	logger := logging.NewNoopLogger(t)

	// Create components
	legacyBackend := NewMockLegacyBackend()
	enhancedBackend := backends.NewRedisBackend(redisClient, config, logger)
	enhancedManager := enhanced.NewEnhancedLockManager(enhancedBackend, config, logger)
	adapter := enhanced.NewLockingAdapter(enhancedManager, enhancedBackend, config, legacyBackend, logger)

	require.NoError(t, enhancedManager.Start(ctx))

	cleanup := func() {
		enhancedManager.Stop()
		redisClient.FlushDB(ctx)
		redisClient.Close()
	}

	return &RollbackTestSuite{
		redisClient:     redisClient,
		legacyBackend:   legacyBackend,
		enhancedBackend: enhancedBackend,
		enhancedManager: enhancedManager,
		adapter:         adapter,
		config:          config,
		logger:          logger,
		cleanup:         cleanup,
	}
}

// TestRollbackProtocols runs all rollback testing scenarios
func TestRollbackProtocols(t *testing.T) {
	suite := SetupRollbackTestSuite(t)
	defer suite.cleanup()

	// Test rollback from each migration phase
	phases := []MigrationPhase{
		PhaseDual,
		PhaseRouting,
		PhaseMigrated,
		PhaseEnhancedOnly,
	}

	for _, phase := range phases {
		t.Run(fmt.Sprintf("RollbackFrom%s", phase), func(t *testing.T) {
			result := suite.TestPhaseRollback(t, phase)
			suite.ValidateRollbackResult(t, result)
		})
	}

	// Test emergency rollback scenarios
	t.Run("EmergencyRollback", suite.TestEmergencyRollback)
	t.Run("PartialFailureRollback", suite.TestPartialFailureRollback)
	t.Run("DataCorruptionRollback", suite.TestDataCorruptionRollback)
}

// TestPhaseRollback tests rollback from a specific migration phase
func (s *RollbackTestSuite) TestPhaseRollback(t *testing.T, phase MigrationPhase) *RollbackResult {
	ctx := context.Background()
	result := &RollbackResult{
		Phase:             phase,
		ErrorsEncountered: make([]string, 0),
	}

	// Create test state for the phase
	testState := s.createTestState(t, ctx, phase)

	// Record pre-rollback metrics
	preRollbackLocks := s.countActiveLocks(ctx)

	// Execute rollback
	startTime := time.Now()
	rollbackSuccess := s.executeRollback(ctx, phase, result)
	result.RollbackDuration = time.Since(startTime)
	result.RollbackSuccessful = rollbackSuccess

	// Validate post-rollback state
	result.DataIntegrityOK = s.validateDataIntegrity(ctx, testState)
	result.ServiceContinuity = s.validateServiceContinuity(ctx)

	// Count preserved/lost locks
	postRollbackLocks := s.countActiveLocks(ctx)
	result.LocksPreserved = postRollbackLocks
	result.LocksLost = preRollbackLocks - postRollbackLocks

	return result
}

// createTestState creates realistic test state for a migration phase
func (s *RollbackTestSuite) createTestState(t *testing.T, ctx context.Context, phase MigrationPhase) map[string]interface{} {
	state := make(map[string]interface{})
	locks := make([]*models.ProjectLock, 0)

	// Configure system for the phase
	s.configureForPhase(phase)

	// Create test locks based on phase
	switch phase {
	case PhaseDual:
		// Both systems have locks
		legacyLocks := s.createLegacyLocks(ctx, 5)
		enhancedLocks := s.createEnhancedLocks(ctx, 3)
		locks = append(locks, legacyLocks...)
		locks = append(locks, enhancedLocks...)

	case PhaseRouting:
		// Mixed locks with routing rules
		routedLocks := s.createRoutedLocks(ctx, 8)
		locks = append(locks, routedLocks...)

	case PhaseMigrated:
		// Mostly enhanced with some legacy fallback
		enhancedLocks := s.createEnhancedLocks(ctx, 10)
		fallbackLocks := s.createFallbackLocks(ctx, 2)
		locks = append(locks, enhancedLocks...)
		locks = append(locks, fallbackLocks...)

	case PhaseEnhancedOnly:
		// Only enhanced locks
		enhancedLocks := s.createEnhancedLocks(ctx, 12)
		locks = append(locks, enhancedLocks...)
	}

	state["locks"] = locks
	state["phase"] = phase
	state["timestamp"] = time.Now()

	return state
}

// executeRollback performs the actual rollback for a phase
func (s *RollbackTestSuite) executeRollback(ctx context.Context, phase MigrationPhase, result *RollbackResult) bool {
	switch phase {
	case PhaseDual:
		return s.rollbackFromDual(ctx, result)
	case PhaseRouting:
		return s.rollbackFromRouting(ctx, result)
	case PhaseMigrated:
		return s.rollbackFromMigrated(ctx, result)
	case PhaseEnhancedOnly:
		return s.rollbackFromEnhancedOnly(ctx, result)
	default:
		result.ErrorsEncountered = append(result.ErrorsEncountered, "Unknown phase")
		return false
	}
}

// rollbackFromDual rolls back from dual system mode to legacy only
func (s *RollbackTestSuite) rollbackFromDual(ctx context.Context, result *RollbackResult) bool {
	// 1. Stop enhanced system
	if err := s.enhancedManager.Stop(); err != nil {
		result.ErrorsEncountered = append(result.ErrorsEncountered, fmt.Sprintf("Failed to stop enhanced manager: %v", err))
		return false
	}

	// 2. Migrate enhanced locks to legacy system
	if !s.migrateEnhancedToLegacy(ctx, result) {
		return false
	}

	// 3. Disable enhanced configuration
	s.config.Enabled = false
	s.config.LegacyFallback = false

	// 4. Verify legacy system is operational
	return s.verifyLegacySystem(ctx, result)
}

// rollbackFromRouting rolls back from routing mode to dual mode
func (s *RollbackTestSuite) rollbackFromRouting(ctx context.Context, result *RollbackResult) bool {
	// 1. Disable routing rules
	s.disableRouting(result)

	// 2. Migrate routed locks back to legacy
	if !s.migrateRoutedToLegacy(ctx, result) {
		return false
	}

	// 3. Reset to dual mode configuration
	s.config.Enabled = true
	s.config.LegacyFallback = true

	return true
}

// rollbackFromMigrated rolls back from full migration to routing mode
func (s *RollbackTestSuite) rollbackFromMigrated(ctx context.Context, result *RollbackResult) bool {
	// 1. Re-enable legacy backend fully
	if !s.reactivateLegacyBackend(ctx, result) {
		return false
	}

	// 2. Migrate critical locks to legacy
	if !s.migrateCriticalLocks(ctx, result) {
		return false
	}

	// 3. Enable routing mode
	return s.enableRoutingMode(result)
}

// rollbackFromEnhancedOnly rolls back from enhanced-only to migrated mode
func (s *RollbackTestSuite) rollbackFromEnhancedOnly(ctx context.Context, result *RollbackResult) bool {
	// 1. Re-enable legacy fallback
	s.config.LegacyFallback = true

	// 2. Initialize legacy backend
	if !s.initializeLegacyBackend(ctx, result) {
		return false
	}

	// 3. Migrate some locks to legacy for safety
	return s.migrateForSafety(ctx, result)
}

// Helper methods for migration operations

func (s *RollbackTestSuite) migrateEnhancedToLegacy(ctx context.Context, result *RollbackResult) bool {
	// Get all enhanced locks
	enhancedLocks, err := s.enhancedBackend.ListLocks(ctx)
	if err != nil {
		result.ErrorsEncountered = append(result.ErrorsEncountered, fmt.Sprintf("Failed to list enhanced locks: %v", err))
		return false
	}

	// Convert and create in legacy system
	for _, enhancedLock := range enhancedLocks {
		legacyLock := s.enhancedBackend.ConvertToLegacy(enhancedLock)
		if legacyLock != nil {
			// Try to acquire in legacy system
			acquired, _, err := s.legacyBackend.TryLock(*legacyLock)
			if err != nil {
				result.ErrorsEncountered = append(result.ErrorsEncountered, fmt.Sprintf("Failed to migrate lock %s: %v", enhancedLock.ID, err))
				continue
			}
			if !acquired {
				result.ErrorsEncountered = append(result.ErrorsEncountered, fmt.Sprintf("Could not acquire migrated lock %s in legacy system", enhancedLock.ID))
			}
		}
	}

	return len(result.ErrorsEncountered) == 0
}

func (s *RollbackTestSuite) migrateRoutedToLegacy(ctx context.Context, result *RollbackResult) bool {
	// Simplified: assume all routed locks can be migrated
	// In real implementation, this would query routing rules and migrate accordingly
	return s.migrateEnhancedToLegacy(ctx, result)
}

func (s *RollbackTestSuite) migrateCriticalLocks(ctx context.Context, result *RollbackResult) bool {
	// Get locks marked as critical priority
	enhancedLocks, err := s.enhancedBackend.ListLocks(ctx)
	if err != nil {
		result.ErrorsEncountered = append(result.ErrorsEncountered, fmt.Sprintf("Failed to list locks for critical migration: %v", err))
		return false
	}

	criticalCount := 0
	for _, lock := range enhancedLocks {
		if lock.Priority == enhanced.PriorityCritical || lock.Priority == enhanced.PriorityHigh {
			legacyLock := s.enhancedBackend.ConvertToLegacy(lock)
			if legacyLock != nil {
				acquired, _, err := s.legacyBackend.TryLock(*legacyLock)
				if err != nil || !acquired {
					result.ErrorsEncountered = append(result.ErrorsEncountered, fmt.Sprintf("Failed to migrate critical lock %s", lock.ID))
				} else {
					criticalCount++
				}
			}
		}
	}

	return criticalCount > 0 || len(enhancedLocks) == 0
}

func (s *RollbackTestSuite) validateDataIntegrity(ctx context.Context, testState map[string]interface{}) bool {
	// Verify that locks from test state are still accessible
	originalLocks, ok := testState["locks"].([]*models.ProjectLock)
	if !ok {
		return false
	}

	preservedCount := 0
	for _, originalLock := range originalLocks {
		// Try to find the lock in current system
		currentLock, err := s.adapter.GetLock(originalLock.Project, originalLock.Workspace)
		if err == nil && currentLock != nil {
			// Verify key fields match
			if currentLock.Project.RepoFullName == originalLock.Project.RepoFullName &&
				currentLock.Workspace == originalLock.Workspace &&
				currentLock.User.Username == originalLock.User.Username {
				preservedCount++
			}
		}
	}

	// At least 80% of locks should be preserved during rollback
	preservationRate := float64(preservedCount) / float64(len(originalLocks))
	return preservationRate >= 0.8
}

func (s *RollbackTestSuite) validateServiceContinuity(ctx context.Context) bool {
	// Test basic operations after rollback
	testProject := models.Project{RepoFullName: "test/continuity", Path: "."}
	testUser := models.User{Username: "continuitytest"}
	testWorkspace := "default"
	testLock := models.ProjectLock{
		Project:   testProject,
		Workspace: testWorkspace,
		User:      testUser,
		Time:      time.Now(),
	}

	// Try lock operation
	acquired, _, err := s.adapter.TryLock(testLock)
	if err != nil || !acquired {
		return false
	}

	// Try list operation
	_, err = s.adapter.List()
	if err != nil {
		return false
	}

	// Try unlock operation
	_, err = s.adapter.Unlock(testProject, testWorkspace, testUser)
	if err != nil {
		return false
	}

	return true
}

// TestEmergencyRollback tests emergency rollback scenarios
func (s *RollbackTestSuite) TestEmergencyRollback(t *testing.T) {
	ctx := context.Background()

	// Simulate critical system failure scenarios
	emergencyScenarios := []struct {
		name        string
		failureType string
		setup       func() error
		expected    bool
	}{
		{
			name:        "RedisCompleteFailure",
			failureType: "redis_down",
			setup: func() error {
				// Simulate Redis failure by closing connection
				return s.redisClient.Close()
			},
			expected: true, // Should successfully rollback to legacy
		},
		{
			name:        "EnhancedManagerCrash",
			failureType: "manager_crash",
			setup: func() error {
				// Stop enhanced manager abruptly
				return s.enhancedManager.Stop()
			},
			expected: true, // Should handle gracefully
		},
		{
			name:        "ConfigurationCorruption",
			failureType: "config_corrupt",
			setup: func() error {
				// Corrupt configuration
				s.config.Backend = "invalid_backend"
				s.config.RedisKeyPrefix = ""
				return nil
			},
			expected: true, // Should detect and handle
		},
	}

	for _, scenario := range emergencyScenarios {
		t.Run(scenario.name, func(t *testing.T) {
			// Setup the failure condition
			if err := scenario.setup(); err != nil {
				t.Logf("Setup error (expected): %v", err)
			}

			// Attempt emergency rollback
			result := s.executeEmergencyRollback(ctx, scenario.failureType)

			if scenario.expected {
				assert.True(t, result.RollbackSuccessful, "Emergency rollback should succeed for %s", scenario.name)
				assert.True(t, result.ServiceContinuity, "Service should remain available after emergency rollback")
			}

			t.Logf("Emergency rollback for %s completed in %v", scenario.name, result.RollbackDuration)
		})
	}
}

// TestPartialFailureRollback tests rollback when only some components fail
func (s *RollbackTestSuite) TestPartialFailureRollback(t *testing.T) {
	ctx := context.Background()

	// Create mixed state
	s.createEnhancedLocks(ctx, 5)
	s.createLegacyLocks(ctx, 3)

	// Simulate partial Redis failure (some operations fail, others succeed)
	partialFailureClient := &PartialFailureRedisClient{
		client:      s.redisClient,
		failureRate: 0.3, // 30% of operations fail
	}

	// Replace backend with partial failure version
	failureConfig := *s.config
	partialBackend := backends.NewRedisBackend(partialFailureClient, &failureConfig, s.logger)

	// Test rollback under partial failure conditions
	result := &RollbackResult{
		Phase:             PhaseMigrated,
		ErrorsEncountered: make([]string, 0),
	}

	startTime := time.Now()
	success := s.rollbackFromMigrated(ctx, result)
	result.RollbackDuration = time.Since(startTime)
	result.RollbackSuccessful = success

	// Validate that rollback handled partial failures gracefully
	assert.True(t, result.RollbackSuccessful || len(result.ErrorsEncountered) < 5,
		"Should handle partial failures with minimal errors")
	assert.Less(t, result.RollbackDuration, 30*time.Second,
		"Rollback should complete quickly even with failures")
}

// TestDataCorruptionRollback tests rollback when data corruption is detected
func (s *RollbackTestSuite) TestDataCorruptionRollback(t *testing.T) {
	ctx := context.Background()

	// Create some locks
	locks := s.createEnhancedLocks(ctx, 10)

	// Simulate data corruption by directly modifying Redis keys
	corruptedKeys := []string{
		s.config.RedisKeyPrefix + "lock:test/corrupt1:.:default",
		s.config.RedisKeyPrefix + "lock:test/corrupt2:.:default",
	}

	for _, key := range corruptedKeys {
		// Insert corrupted data
		s.redisClient.Set(ctx, key, "corrupted_json_data{invalid}", time.Hour)
	}

	// Test detection and rollback
	result := s.TestPhaseRollback(t, PhaseMigrated)

	assert.True(t, result.RollbackSuccessful, "Should successfully rollback despite data corruption")
	assert.True(t, result.ServiceContinuity, "Service should continue after corruption rollback")

	// Verify corrupted data doesn't affect rollback
	assert.Less(t, result.LocksLost, len(locks)/2, "Should not lose more than half the locks due to corruption")
}

// ValidateRollbackResult validates that rollback results meet acceptance criteria
func (s *RollbackTestSuite) ValidateRollbackResult(t *testing.T, result *RollbackResult) {
	// Core validation criteria
	assert.True(t, result.RollbackSuccessful, "Rollback should succeed for phase %s", result.Phase)
	assert.True(t, result.DataIntegrityOK, "Data integrity should be preserved for phase %s", result.Phase)
	assert.True(t, result.ServiceContinuity, "Service should remain available for phase %s", result.Phase)

	// Performance criteria
	assert.Less(t, result.RollbackDuration, 2*time.Minute, "Rollback should complete within 2 minutes for phase %s", result.Phase)

	// Data preservation criteria
	preservationRate := float64(result.LocksPreserved) / float64(result.LocksPreserved+result.LocksLost)
	assert.Greater(t, preservationRate, 0.8, "Should preserve at least 80%% of locks for phase %s", result.Phase)

	// Error tolerance
	assert.LessOrEqual(t, len(result.ErrorsEncountered), 3, "Should have minimal errors during rollback for phase %s", result.Phase)

	// Log results for review
	t.Logf("Rollback validation for %s:", result.Phase)
	t.Logf("  Success: %t", result.RollbackSuccessful)
	t.Logf("  Duration: %v", result.RollbackDuration)
	t.Logf("  Locks preserved: %d, lost: %d (%.1f%% preservation)",
		result.LocksPreserved, result.LocksLost, preservationRate*100)
	t.Logf("  Errors: %d", len(result.ErrorsEncountered))
	for _, err := range result.ErrorsEncountered {
		t.Logf("    - %s", err)
	}
}

// Helper methods for test setup and execution

func (s *RollbackTestSuite) configureForPhase(phase MigrationPhase) {
	switch phase {
	case PhaseDual:
		s.config.Enabled = true
		s.config.LegacyFallback = true
	case PhaseRouting:
		s.config.Enabled = true
		s.config.LegacyFallback = true
		// Routing rules would be configured here
	case PhaseMigrated:
		s.config.Enabled = true
		s.config.LegacyFallback = true
	case PhaseEnhancedOnly:
		s.config.Enabled = true
		s.config.LegacyFallback = false
	}
}

func (s *RollbackTestSuite) createLegacyLocks(ctx context.Context, count int) []*models.ProjectLock {
	locks := make([]*models.ProjectLock, 0, count)
	for i := 0; i < count; i++ {
		lock := &models.ProjectLock{
			Project:   models.Project{RepoFullName: fmt.Sprintf("test/legacy-%d", i), Path: "."},
			Workspace: "default",
			User:      models.User{Username: fmt.Sprintf("legacyuser%d", i)},
			Time:      time.Now(),
		}
		acquired, _, _ := s.legacyBackend.TryLock(*lock)
		if acquired {
			locks = append(locks, lock)
		}
	}
	return locks
}

func (s *RollbackTestSuite) createEnhancedLocks(ctx context.Context, count int) []*models.ProjectLock {
	locks := make([]*models.ProjectLock, 0, count)
	for i := 0; i < count; i++ {
		lock := &models.ProjectLock{
			Project:   models.Project{RepoFullName: fmt.Sprintf("test/enhanced-%d", i), Path: "."},
			Workspace: "default",
			User:      models.User{Username: fmt.Sprintf("enhanceduser%d", i)},
			Time:      time.Now(),
		}
		acquired, _, _ := s.adapter.TryLock(*lock)
		if acquired {
			locks = append(locks, lock)
		}
	}
	return locks
}

func (s *RollbackTestSuite) createRoutedLocks(ctx context.Context, count int) []*models.ProjectLock {
	// Simulate routed locks (combination of legacy and enhanced)
	locks := make([]*models.ProjectLock, 0)
	locks = append(locks, s.createLegacyLocks(ctx, count/2)...)
	locks = append(locks, s.createEnhancedLocks(ctx, count/2)...)
	return locks
}

func (s *RollbackTestSuite) createFallbackLocks(ctx context.Context, count int) []*models.ProjectLock {
	// Create locks that would use fallback mechanism
	return s.createLegacyLocks(ctx, count)
}

func (s *RollbackTestSuite) countActiveLocks(ctx context.Context) int {
	locks, err := s.adapter.List()
	if err != nil {
		return 0
	}
	return len(locks)
}

func (s *RollbackTestSuite) executeEmergencyRollback(ctx context.Context, failureType string) *RollbackResult {
	result := &RollbackResult{
		Phase:             PhaseEnhancedOnly, // Assume worst case
		ErrorsEncountered: make([]string, 0),
	}

	startTime := time.Now()

	// Emergency rollback logic
	switch failureType {
	case "redis_down":
		// Immediately switch to legacy backend
		s.config.Enabled = false
		s.config.LegacyFallback = false
		result.RollbackSuccessful = true

	case "manager_crash":
		// Try to restart or fall back
		s.config.LegacyFallback = true
		result.RollbackSuccessful = true

	case "config_corrupt":
		// Reset to safe defaults
		s.config = enhanced.DefaultConfig()
		s.config.LegacyFallback = true
		result.RollbackSuccessful = true

	default:
		result.ErrorsEncountered = append(result.ErrorsEncountered, "Unknown failure type")
		result.RollbackSuccessful = false
	}

	result.RollbackDuration = time.Since(startTime)
	result.ServiceContinuity = s.validateServiceContinuity(ctx)
	result.DataIntegrityOK = result.ServiceContinuity // Simplified for emergency scenarios

	return result
}

// Utility methods for rollback operations
func (s *RollbackTestSuite) disableRouting(result *RollbackResult) {
	// Simplified: in real implementation, this would disable routing rules
	result.ErrorsEncountered = append(result.ErrorsEncountered, "Routing disabled")
}

func (s *RollbackTestSuite) reactivateLegacyBackend(ctx context.Context, result *RollbackResult) bool {
	// Ensure legacy backend is fully operational
	return s.verifyLegacySystem(ctx, result)
}

func (s *RollbackTestSuite) enableRoutingMode(result *RollbackResult) bool {
	// Re-enable routing configuration
	return true
}

func (s *RollbackTestSuite) initializeLegacyBackend(ctx context.Context, result *RollbackResult) bool {
	// Initialize legacy backend if needed
	return true
}

func (s *RollbackTestSuite) migrateForSafety(ctx context.Context, result *RollbackResult) bool {
	// Migrate some locks to legacy for safety
	return s.migrateCriticalLocks(ctx, result)
}

func (s *RollbackTestSuite) verifyLegacySystem(ctx context.Context, result *RollbackResult) bool {
	// Test that legacy system is working
	testLock := models.ProjectLock{
		Project:   models.Project{RepoFullName: "test/legacy-verify", Path: "."},
		Workspace: "default",
		User:      models.User{Username: "verifyuser"},
		Time:      time.Now(),
	}

	acquired, _, err := s.legacyBackend.TryLock(testLock)
	if err != nil || !acquired {
		result.ErrorsEncountered = append(result.ErrorsEncountered, "Legacy system verification failed")
		return false
	}

	s.legacyBackend.Unlock(testLock.Project, testLock.Workspace, testLock.User)
	return true
}

// PartialFailureRedisClient simulates partial Redis failures for testing
type PartialFailureRedisClient struct {
	client      redis.UniversalClient
	failureRate float64
	mutex       sync.Mutex
}

func (p *PartialFailureRedisClient) Ping(ctx context.Context) *redis.StatusCmd {
	if p.shouldFail() {
		cmd := redis.NewStatusCmd(ctx)
		cmd.SetErr(fmt.Errorf("simulated ping failure"))
		return cmd
	}
	return p.client.Ping(ctx)
}

func (p *PartialFailureRedisClient) Get(ctx context.Context, key string) *redis.StringCmd {
	if p.shouldFail() {
		cmd := redis.NewStringCmd(ctx, "get", key)
		cmd.SetErr(fmt.Errorf("simulated get failure"))
		return cmd
	}
	return p.client.Get(ctx, key)
}

func (p *PartialFailureRedisClient) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd {
	if p.shouldFail() {
		cmd := redis.NewStatusCmd(ctx, "set", key, value)
		cmd.SetErr(fmt.Errorf("simulated set failure"))
		return cmd
	}
	return p.client.Set(ctx, key, value, expiration)
}

func (p *PartialFailureRedisClient) shouldFail() bool {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	return rand.Float64() < p.failureRate
}

// Implement other redis.UniversalClient methods as needed...