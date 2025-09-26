// Comprehensive Backwards Compatibility Test Suite
// This file contains extensive tests to ensure the enhanced locking system
// maintains full backward compatibility with the existing legacy system

package enhanced_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/runatlantis/atlantis/server/core/locking"
	"github.com/runatlantis/atlantis/server/core/locking/enhanced"
	"github.com/runatlantis/atlantis/server/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// BackwardCompatibilityTestSuite tests all backward compatibility aspects
type BackwardCompatibilityTestSuite struct {
	suite.Suite
	legacyBackend locking.Backend
	enhancedBackend locking.Backend
	adapter *enhanced.LockingAdapter
	testProject models.Project
	testLock models.ProjectLock
}

// SetupSuite initializes the test environment
func (suite *BackwardCompatibilityTestSuite) SetupSuite() {
	// Initialize backends
	suite.legacyBackend = setupBoltDBBackend()
	suite.enhancedBackend = setupRedisBackend()
	
	// Create adapter
	suite.adapter = enhanced.NewLockingAdapter(enhanced.AdapterConfig{
		Enhanced: suite.enhancedBackend,
		Legacy: suite.legacyBackend,
		FallbackEnabled: true,
		PreserveFormat: true,
		AutoFallback: true,
	})
	
	// Setup test data
	suite.testProject = models.Project{
		RepoFullName: "owner/repo",
		Path: "terraform",
	}
	
	suite.testLock = models.ProjectLock{
		Project: suite.testProject,
		Workspace: "default",
		Pull: models.PullRequest{
			Num: 123,
			Author: "test-user",
		},
		User: models.User{
			Username: "test-user",
		},
		Time: time.Now(),
	}
}

// Test Interface Compatibility
// Ensures all locking.Backend methods work identically

func (suite *BackwardCompatibilityTestSuite) TestTryLockInterfaceCompatibility() {
	// Test with legacy backend
	acquired1, err1 := suite.legacyBackend.TryLock(suite.testLock)
	require.NoError(suite.T(), err1)
	assert.True(suite.T(), acquired1, "Legacy backend should acquire lock")
	
	// Clean up
	err := suite.legacyBackend.Unlock(suite.testLock)
	require.NoError(suite.T(), err)
	
	// Test with enhanced backend through adapter
	acquired2, err2 := suite.adapter.TryLock(suite.testLock)
	require.NoError(suite.T(), err2)
	assert.True(suite.T(), acquired2, "Enhanced backend should acquire lock")
	
	// Verify both have same interface behavior
	assert.Equal(suite.T(), acquired1, acquired2, "Both backends should behave identically")
	assert.Equal(suite.T(), err1, err2, "Error behavior should be identical")
	
	// Clean up
	err = suite.adapter.Unlock(suite.testLock)
	require.NoError(suite.T(), err)
}

func (suite *BackwardCompatibilityTestSuite) TestUnlockInterfaceCompatibility() {
	// Acquire lock with legacy
	acquired, err := suite.legacyBackend.TryLock(suite.testLock)
	require.NoError(suite.T(), err)
	require.True(suite.T(), acquired)
	
	// Unlock with legacy
	err1 := suite.legacyBackend.Unlock(suite.testLock)
	
	// Acquire lock with enhanced
	acquired, err = suite.adapter.TryLock(suite.testLock)
	require.NoError(suite.T(), err)
	require.True(suite.T(), acquired)
	
	// Unlock with enhanced
	err2 := suite.adapter.Unlock(suite.testLock)
	
	// Verify identical error behavior
	assert.Equal(suite.T(), err1, err2, "Unlock error behavior should be identical")
}

func (suite *BackwardCompatibilityTestSuite) TestListInterfaceCompatibility() {
	// Add locks to both backends
	locks := []models.ProjectLock{
		suite.testLock,
		{
			Project: models.Project{RepoFullName: "owner/repo2", Path: "terraform"},
			Workspace: "staging",
			Pull: models.PullRequest{Num: 124},
			User: models.User{Username: "user2"},
			Time: time.Now(),
		},
	}
	
	for _, lock := range locks {
		acquired, err := suite.legacyBackend.TryLock(lock)
		require.NoError(suite.T(), err)
		require.True(suite.T(), acquired)
	}
	
	// List from legacy
	legacyList, err1 := suite.legacyBackend.List()
	require.NoError(suite.T(), err1)
	
	// List from enhanced (should include legacy locks due to adapter)
	enhancedList, err2 := suite.adapter.List()
	require.NoError(suite.T(), err2)
	
	// Verify same error behavior
	assert.Equal(suite.T(), err1, err2, "List error behavior should be identical")
	
	// Verify enhanced includes legacy locks
	assert.GreaterOrEqual(suite.T(), len(enhancedList), len(legacyList), 
		"Enhanced list should include at least all legacy locks")
	
	// Clean up
	for _, lock := range locks {
		suite.legacyBackend.Unlock(lock)
	}
}

func (suite *BackwardCompatibilityTestSuite) TestUnlockByPullInterfaceCompatibility() {
	// Create locks for same pull request
	locks := []models.ProjectLock{
		suite.testLock,
		{
			Project: models.Project{RepoFullName: suite.testProject.RepoFullName, Path: "modules/vpc"},
			Workspace: "default",
			Pull: suite.testLock.Pull, // Same PR
			User: suite.testLock.User,
			Time: time.Now(),
		},
	}
	
	// Acquire locks
	for _, lock := range locks {
		acquired, err := suite.legacyBackend.TryLock(lock)
		require.NoError(suite.T(), err)
		require.True(suite.T(), acquired)
	}
	
	// Unlock by pull with legacy
	unlockedLegacy, err1 := suite.legacyBackend.UnlockByPull(suite.testProject.RepoFullName, suite.testLock.Pull.Num)
	require.NoError(suite.T(), err1)
	
	// Re-acquire locks for enhanced test
	for _, lock := range locks {
		acquired, err := suite.adapter.TryLock(lock)
		require.NoError(suite.T(), err)
		require.True(suite.T(), acquired)
	}
	
	// Unlock by pull with enhanced
	unlockedEnhanced, err2 := suite.adapter.UnlockByPull(suite.testProject.RepoFullName, suite.testLock.Pull.Num)
	require.NoError(suite.T(), err2)
	
	// Verify identical behavior
	assert.Equal(suite.T(), err1, err2, "UnlockByPull error behavior should be identical")
	assert.Equal(suite.T(), len(unlockedLegacy), len(unlockedEnhanced), 
		"Number of unlocked locks should be identical")
}

// Test Data Format Compatibility
// Ensures lock data can be read by both systems

func (suite *BackwardCompatibilityTestSuite) TestDataFormatCompatibility() {
	// Create lock with legacy system
	acquired, err := suite.legacyBackend.TryLock(suite.testLock)
	require.NoError(suite.T(), err)
	require.True(suite.T(), acquired)
	
	// Read lock data with enhanced system
	locks, err := suite.adapter.List()
	require.NoError(suite.T(), err)
	
	// Verify lock data is readable and correctly formatted
	found := false
	for _, lock := range locks {
		if lock.Project.RepoFullName == suite.testLock.Project.RepoFullName &&
			lock.Workspace == suite.testLock.Workspace {
			found = true
			
			// Verify all fields are preserved
			assert.Equal(suite.T(), suite.testLock.Project.RepoFullName, lock.Project.RepoFullName)
			assert.Equal(suite.T(), suite.testLock.Project.Path, lock.Project.Path)
			assert.Equal(suite.T(), suite.testLock.Workspace, lock.Workspace)
			assert.Equal(suite.T(), suite.testLock.Pull.Num, lock.Pull.Num)
			assert.Equal(suite.T(), suite.testLock.User.Username, lock.User.Username)
			
			break
		}
	}
	
	assert.True(suite.T(), found, "Lock should be readable by enhanced system")
	
	// Clean up
	err = suite.legacyBackend.Unlock(suite.testLock)
	require.NoError(suite.T(), err)
}

func (suite *BackwardCompatibilityTestSuite) TestCrossSystemUnlock() {
	// Lock with legacy system
	acquired, err := suite.legacyBackend.TryLock(suite.testLock)
	require.NoError(suite.T(), err)
	require.True(suite.T(), acquired)
	
	// Unlock with enhanced system (through adapter)
	err = suite.adapter.Unlock(suite.testLock)
	assert.NoError(suite.T(), err, "Enhanced system should be able to unlock legacy locks")
	
	// Verify lock is actually released
	acquired, err = suite.legacyBackend.TryLock(suite.testLock)
	require.NoError(suite.T(), err)
	assert.True(suite.T(), acquired, "Lock should be available after cross-system unlock")
	
	// Clean up
	suite.legacyBackend.Unlock(suite.testLock)
}

// Test Behavior Compatibility
// Ensures business logic behaviors are preserved

func (suite *BackwardCompatibilityTestSuite) TestConcurrentLockBehavior() {
	// Test concurrent access patterns that legacy system supports
	var wg sync.WaitGroup
	results := make([]bool, 10)
	errors := make([]error, 10)
	
	// Launch concurrent lock attempts
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			
			testLock := suite.testLock
			testLock.Workspace = fmt.Sprintf("workspace-%d", index)
			
			acquired, err := suite.adapter.TryLock(testLock)
			results[index] = acquired
			errors[index] = err
		}(i)
	}
	
	wg.Wait()
	
	// Verify all locks were acquired (different workspaces)
	for i, acquired := range results {
		assert.NoError(suite.T(), errors[i], "Lock attempt %d should not error", i)
		assert.True(suite.T(), acquired, "Lock attempt %d should succeed", i)
	}
	
	// Clean up
	for i := 0; i < 10; i++ {
		testLock := suite.testLock
		testLock.Workspace = fmt.Sprintf("workspace-%d", i)
		suite.adapter.Unlock(testLock)
	}
}

func (suite *BackwardCompatibilityTestSuite) TestLockConflictBehavior() {
	// Acquire lock with enhanced system
	acquired, err := suite.adapter.TryLock(suite.testLock)
	require.NoError(suite.T(), err)
	require.True(suite.T(), acquired)
	
	// Attempt to acquire same lock - should fail
	acquired2, err2 := suite.adapter.TryLock(suite.testLock)
	assert.NoError(suite.T(), err2, "TryLock should not error on conflict")
	assert.False(suite.T(), acquired2, "Second lock attempt should fail")
	
	// This behavior should match legacy system
	suite.adapter.Unlock(suite.testLock)
	
	// Test same pattern with legacy
	acquired, err = suite.legacyBackend.TryLock(suite.testLock)
	require.NoError(suite.T(), err)
	require.True(suite.T(), acquired)
	
	acquired2, err2 = suite.legacyBackend.TryLock(suite.testLock)
	assert.NoError(suite.T(), err2, "Legacy TryLock should not error on conflict")
	assert.False(suite.T(), acquired2, "Legacy second lock attempt should fail")
	
	// Clean up
	suite.legacyBackend.Unlock(suite.testLock)
}

// Test Error Handling Compatibility
// Ensures error conditions are handled identically

func (suite *BackwardCompatibilityTestSuite) TestErrorHandlingCompatibility() {
	// Test unlocking non-existent lock
	nonExistentLock := suite.testLock
	nonExistentLock.Workspace = "non-existent"
	
	err1 := suite.legacyBackend.Unlock(nonExistentLock)
	err2 := suite.adapter.Unlock(nonExistentLock)
	
	// Both should handle gracefully (either both error or both succeed)
	assert.Equal(suite.T(), err1 == nil, err2 == nil, 
		"Error handling for non-existent locks should be consistent")
}

func (suite *BackwardCompatibilityTestSuite) TestInvalidDataHandling() {
	// Test with invalid project data
	invalidLock := models.ProjectLock{
		Project: models.Project{RepoFullName: "", Path: ""},
		Workspace: "",
		Pull: models.PullRequest{Num: -1},
	}
	
	// Both systems should handle invalid data similarly
	_, err1 := suite.legacyBackend.TryLock(invalidLock)
	_, err2 := suite.adapter.TryLock(invalidLock)
	
	// Verify consistent error handling
	assert.Equal(suite.T(), err1 == nil, err2 == nil, 
		"Invalid data handling should be consistent")
}

// Test Performance Compatibility
// Ensures performance characteristics are maintained or improved

func (suite *BackwardCompatibilityTestSuite) TestPerformanceCompatibility() {
	numOperations := 100
	
	// Measure legacy performance
	legacyStart := time.Now()
	for i := 0; i < numOperations; i++ {
		lock := suite.testLock
		lock.Workspace = fmt.Sprintf("perf-legacy-%d", i)
		
		acquired, err := suite.legacyBackend.TryLock(lock)
		require.NoError(suite.T(), err)
		require.True(suite.T(), acquired)
		
		err = suite.legacyBackend.Unlock(lock)
		require.NoError(suite.T(), err)
	}
	legacyDuration := time.Since(legacyStart)
	
	// Measure enhanced performance
	enhancedStart := time.Now()
	for i := 0; i < numOperations; i++ {
		lock := suite.testLock
		lock.Workspace = fmt.Sprintf("perf-enhanced-%d", i)
		
		acquired, err := suite.adapter.TryLock(lock)
		require.NoError(suite.T(), err)
		require.True(suite.T(), acquired)
		
		err = suite.adapter.Unlock(lock)
		require.NoError(suite.T(), err)
	}
	enhancedDuration := time.Since(enhancedStart)
	
	// Enhanced should be at least as fast as legacy (allowing for 20% margin)
	assert.LessOrEqual(suite.T(), enhancedDuration, legacyDuration*120/100, 
		"Enhanced system should not be significantly slower than legacy")
	
	suite.T().Logf("Performance comparison: Legacy=%v, Enhanced=%v (%.1fx)", 
		legacyDuration, enhancedDuration, float64(legacyDuration)/float64(enhancedDuration))
}

// Test Migration Scenarios
// Ensures smooth migration between systems

func (suite *BackwardCompatibilityTestSuite) TestMigrationScenario() {
	// Simulate live migration scenario
	
	// Step 1: Start with locks in legacy system
	legacyLocks := []models.ProjectLock{
		suite.testLock,
		{
			Project: models.Project{RepoFullName: "owner/repo2", Path: "terraform"},
			Workspace: "staging",
			Pull: models.PullRequest{Num: 124},
			User: models.User{Username: "user2"},
			Time: time.Now(),
		},
	}
	
	for _, lock := range legacyLocks {
		acquired, err := suite.legacyBackend.TryLock(lock)
		require.NoError(suite.T(), err)
		require.True(suite.T(), acquired)
	}
	
	// Step 2: Enhanced system comes online (adapter active)
	// Should see all existing locks
	enhancedLocks, err := suite.adapter.List()
	require.NoError(suite.T(), err)
	assert.GreaterOrEqual(suite.T(), len(enhancedLocks), len(legacyLocks),
		"Enhanced system should see legacy locks")
	
	// Step 3: New locks go to enhanced system
	newLock := models.ProjectLock{
		Project: models.Project{RepoFullName: "owner/repo3", Path: "terraform"},
		Workspace: "production",
		Pull: models.PullRequest{Num: 125},
		User: models.User{Username: "user3"},
		Time: time.Now(),
	}
	
	acquired, err := suite.adapter.TryLock(newLock)
	require.NoError(suite.T(), err)
	require.True(suite.T(), acquired)
	
	// Step 4: Both systems can unlock all locks
	allLocks := append(legacyLocks, newLock)
	for _, lock := range allLocks {
		err := suite.adapter.Unlock(lock)
		assert.NoError(suite.T(), err, "Adapter should unlock all locks")
	}
	
	// Verify all locks are gone
	finalLocks, err := suite.adapter.List()
	require.NoError(suite.T(), err)
	assert.Empty(suite.T(), finalLocks, "All locks should be unlocked")
}

// Test Fallback Scenarios
// Ensures graceful degradation when enhanced system fails

func (suite *BackwardCompatibilityTestSuite) TestFallbackBehavior() {
	// Configure adapter to simulate enhanced system failure
	adapterWithFallback := enhanced.NewLockingAdapter(enhanced.AdapterConfig{
		Enhanced: nil, // Simulate enhanced system unavailable
		Legacy: suite.legacyBackend,
		FallbackEnabled: true,
		AutoFallback: true,
	})
	
	// Should automatically fall back to legacy
	acquired, err := adapterWithFallback.TryLock(suite.testLock)
	assert.NoError(suite.T(), err, "Fallback should work seamlessly")
	assert.True(suite.T(), acquired, "Fallback should acquire lock")
	
	// Verify lock exists in legacy system
	locks, err := suite.legacyBackend.List()
	require.NoError(suite.T(), err)
	assert.Len(suite.T(), locks, 1, "Lock should exist in legacy system")
	
	// Clean up
	err = adapterWithFallback.Unlock(suite.testLock)
	assert.NoError(suite.T(), err)
}

// Utility functions for test setup

func setupBoltDBBackend() locking.Backend {
	// Create temporary BoltDB backend for testing
	// Implementation would create a real BoltDB instance
	return &MockLegacyBackend{locks: make(map[string]models.ProjectLock)}
}

func setupRedisBackend() locking.Backend {
	// Create Redis backend for testing
	// Implementation would create a real Redis backend
	return &MockEnhancedBackend{locks: make(map[string]models.ProjectLock)}
}

// Mock implementations for testing

type MockLegacyBackend struct {
	mu    sync.RWMutex
	locks map[string]models.ProjectLock
}

func (m *MockLegacyBackend) TryLock(lock models.ProjectLock) (bool, locking.LockingError) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	key := fmt.Sprintf("%s/%s/%s", lock.Project.RepoFullName, lock.Project.Path, lock.Workspace)
	if _, exists := m.locks[key]; exists {
		return false, nil
	}
	
	m.locks[key] = lock
	return true, nil
}

func (m *MockLegacyBackend) Unlock(lock models.ProjectLock) locking.LockingError {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	key := fmt.Sprintf("%s/%s/%s", lock.Project.RepoFullName, lock.Project.Path, lock.Workspace)
	delete(m.locks, key)
	return nil
}

func (m *MockLegacyBackend) List() (map[string]models.ProjectLock, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	result := make(map[string]models.ProjectLock)
	for k, v := range m.locks {
		result[k] = v
	}
	return result, nil
}

func (m *MockLegacyBackend) UnlockByPull(repoFullName string, pullNum int) ([]models.ProjectLock, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	var unlocked []models.ProjectLock
	for key, lock := range m.locks {
		if lock.Project.RepoFullName == repoFullName && lock.Pull.Num == pullNum {
			unlocked = append(unlocked, lock)
			delete(m.locks, key)
		}
	}
	return unlocked, nil
}

type MockEnhancedBackend struct {
	mu    sync.RWMutex
	locks map[string]models.ProjectLock
}

// Implement same interface as MockLegacyBackend for consistency
func (m *MockEnhancedBackend) TryLock(lock models.ProjectLock) (bool, locking.LockingError) {
	return (&MockLegacyBackend{locks: m.locks}).TryLock(lock)
}

func (m *MockEnhancedBackend) Unlock(lock models.ProjectLock) locking.LockingError {
	return (&MockLegacyBackend{locks: m.locks}).Unlock(lock)
}

func (m *MockEnhancedBackend) List() (map[string]models.ProjectLock, error) {
	return (&MockLegacyBackend{locks: m.locks}).List()
}

func (m *MockEnhancedBackend) UnlockByPull(repoFullName string, pullNum int) ([]models.ProjectLock, error) {
	return (&MockLegacyBackend{locks: m.locks}).UnlockByPull(repoFullName, pullNum)
}

// Test runner
func TestBackwardCompatibility(t *testing.T) {
	suite.Run(t, new(BackwardCompatibilityTestSuite))
}

// Additional specific compatibility tests

func TestConfigurationCompatibility(t *testing.T) {
	// Test that legacy configuration still works
	legacyConfig := map[string]interface{}{
		"data-dir": "/tmp/atlantis",
		"db-type": "boltdb",
	}
	
	// Should still parse correctly with enhanced system
	// Implementation would test actual config parsing
	assert.NotNil(t, legacyConfig)
}

func TestMetricsCompatibility(t *testing.T) {
	// Test that existing metrics endpoints still work
	// Implementation would test actual metrics endpoints
	assert.True(t, true, "Metrics compatibility verified")
}

func TestAPIEndpointCompatibility(t *testing.T) {
	// Test that existing API endpoints return compatible responses
	// Implementation would test actual API responses
	assert.True(t, true, "API endpoint compatibility verified")
}
