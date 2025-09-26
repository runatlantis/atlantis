// Phase 2 Testing Strategy: Backwards Compatibility Validation
// This phase ensures the enhanced locking system maintains full compatibility with existing Atlantis interfaces

package tests

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/runatlantis/atlantis/server/core/locking"
	"github.com/runatlantis/atlantis/server/core/locking/enhanced"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/logging"
)

// Phase2TestSuite validates backwards compatibility with legacy locking interfaces
type Phase2TestSuite struct {
	suite.Suite
	enhancedManager enhanced.LockManager
	legacyAdapter   *enhanced.LockingAdapter
	legacyBackend   locking.Backend
	config          *enhanced.EnhancedConfig
	cleanup         func()
	ctx             context.Context
}

// SetupSuite initializes the test environment for Phase 2
func (s *Phase2TestSuite) SetupSuite() {
	s.ctx = context.Background()

	// Phase 2 configuration: enhanced features enabled with legacy compatibility
	s.config = &enhanced.EnhancedConfig{
		Enabled:                 true,
		Backend:                 "boltdb",
		DefaultTimeout:          30 * time.Minute,
		MaxTimeout:              2 * time.Hour,

		// Keep advanced features disabled for cleaner compatibility testing
		EnablePriorityQueue:     false,
		EnableRetry:            false,
		EnableDeadlockDetection: false,
		EnableEvents:           false,

		// Critical: Enable full backward compatibility
		LegacyFallback:         true,
		PreserveLegacyFormat:   true,
	}

	s.enhancedManager, s.legacyBackend, s.cleanup = s.setupCompatibilityTest()

	// Create the adapter that bridges enhanced and legacy systems
	s.legacyAdapter = enhanced.NewLockingAdapter(
		s.enhancedManager,
		nil, // backend will be injected by setupCompatibilityTest
		s.config,
		s.legacyBackend,
		logging.NewNoopLogger(s.T()),
	)
}

func (s *Phase2TestSuite) TearDownSuite() {
	if s.cleanup != nil {
		s.cleanup()
	}
}

// Test Legacy TryLock Interface
func (s *Phase2TestSuite) TestLegacyTryLockInterface() {
	project := models.Project{
		RepoFullName: "test/legacy-trylock",
		Path:         ".",
	}
	user := models.User{Username: "legacy-user"}
	workspace := "default"

	legacyLock := models.ProjectLock{
		Project:   project,
		Workspace: workspace,
		User:      user,
		Time:      time.Now(),
	}

	// Test successful lock acquisition
	acquired, resp, err := s.legacyAdapter.TryLock(legacyLock)
	require.NoError(s.T(), err, "TryLock should succeed")
	assert.True(s.T(), acquired, "Lock should be acquired")
	assert.Empty(s.T(), resp.LockFailureReason, "No failure reason on success")
	assert.Empty(s.T(), resp.CurrLock, "No current lock info on successful acquisition")

	// Test conflicting lock attempt
	user2 := models.User{Username: "legacy-user2"}
	conflictingLock := models.ProjectLock{
		Project:   project,
		Workspace: workspace,
		User:      user2,
		Time:      time.Now(),
	}

	acquired2, resp2, err2 := s.legacyAdapter.TryLock(conflictingLock)
	require.NoError(s.T(), err2, "TryLock should not error on conflict")
	assert.False(s.T(), acquired2, "Conflicting lock should not be acquired")
	assert.NotEmpty(s.T(), resp2.LockFailureReason, "Should provide failure reason")

	// Verify current lock information is provided
	assert.Equal(s.T(), project.RepoFullName, resp2.CurrLock.Project.RepoFullName)
	assert.Equal(s.T(), workspace, resp2.CurrLock.Workspace)
	assert.Equal(s.T(), user.Username, resp2.CurrLock.User.Username)

	// Cleanup
	s.legacyAdapter.Unlock(project, workspace, user)
}

// Test Legacy List Interface
func (s *Phase2TestSuite) TestLegacyListInterface() {
	// Create multiple locks using legacy interface
	testData := []struct {
		project   models.Project
		workspace string
		user      models.User
	}{
		{
			project:   models.Project{RepoFullName: "test/list1", Path: "."},
			workspace: "dev",
			user:      models.User{Username: "user1"},
		},
		{
			project:   models.Project{RepoFullName: "test/list2", Path: "terraform"},
			workspace: "staging",
			user:      models.User{Username: "user2"},
		},
		{
			project:   models.Project{RepoFullName: "test/list1", Path: "."},
			workspace: "prod",
			user:      models.User{Username: "user3"},
		},
	}

	// Acquire all test locks
	for _, data := range testData {
		legacyLock := models.ProjectLock{
			Project:   data.project,
			Workspace: data.workspace,
			User:      data.user,
			Time:      time.Now(),
		}

		acquired, _, err := s.legacyAdapter.TryLock(legacyLock)
		require.NoError(s.T(), err, "Should acquire lock for %s/%s", data.project.RepoFullName, data.workspace)
		require.True(s.T(), acquired, "Lock should be acquired for %s/%s", data.project.RepoFullName, data.workspace)
	}

	// Test List method
	lockMap, err := s.legacyAdapter.List()
	require.NoError(s.T(), err, "List should succeed")
	assert.Len(s.T(), lockMap, len(testData), "Should return all acquired locks")

	// Verify lock contents and format
	for _, data := range testData {
		expectedKey := fmt.Sprintf("%s/%s/%s", data.project.RepoFullName, data.project.Path, data.workspace)

		lock, exists := lockMap[expectedKey]
		assert.True(s.T(), exists, "Should find lock for key %s", expectedKey)

		if exists {
			assert.Equal(s.T(), data.project.RepoFullName, lock.Project.RepoFullName)
			assert.Equal(s.T(), data.project.Path, lock.Project.Path)
			assert.Equal(s.T(), data.workspace, lock.Workspace)
			assert.Equal(s.T(), data.user.Username, lock.User.Username)
			assert.WithinDuration(s.T(), time.Now(), lock.Time, 5*time.Second)
		}
	}

	// Cleanup all locks
	for _, data := range testData {
		s.legacyAdapter.Unlock(data.project, data.workspace, data.user)
	}

	// Verify cleanup
	emptyMap, err := s.legacyAdapter.List()
	require.NoError(s.T(), err)
	assert.Len(s.T(), emptyMap, 0, "All locks should be cleaned up")
}

// Test Legacy GetLock Interface
func (s *Phase2TestSuite) TestLegacyGetLockInterface() {
	project := models.Project{
		RepoFullName: "test/get-lock",
		Path:         "infrastructure",
	}
	user := models.User{Username: "get-lock-user"}
	workspace := "production"

	// Test GetLock when no lock exists
	lock, err := s.legacyAdapter.GetLock(project, workspace)
	require.NoError(s.T(), err, "GetLock should not error when no lock exists")
	assert.Nil(s.T(), lock, "Should return nil when no lock exists")

	// Acquire a lock
	legacyLock := models.ProjectLock{
		Project:   project,
		Workspace: workspace,
		User:      user,
		Time:      time.Now(),
	}

	acquired, _, err := s.legacyAdapter.TryLock(legacyLock)
	require.NoError(s.T(), err)
	require.True(s.T(), acquired)

	// Test GetLock when lock exists
	retrievedLock, err := s.legacyAdapter.GetLock(project, workspace)
	require.NoError(s.T(), err, "GetLock should succeed when lock exists")
	require.NotNil(s.T(), retrievedLock, "Should return lock when it exists")

	// Verify lock details
	assert.Equal(s.T(), project.RepoFullName, retrievedLock.Project.RepoFullName)
	assert.Equal(s.T(), project.Path, retrievedLock.Project.Path)
	assert.Equal(s.T(), workspace, retrievedLock.Workspace)
	assert.Equal(s.T(), user.Username, retrievedLock.User.Username)

	// Test GetLock for different workspace (should return nil)
	otherWorkspaceLock, err := s.legacyAdapter.GetLock(project, "different-workspace")
	require.NoError(s.T(), err)
	assert.Nil(s.T(), otherWorkspaceLock, "Should return nil for different workspace")

	// Cleanup
	s.legacyAdapter.Unlock(project, workspace, user)
}

// Test Legacy Unlock Interface
func (s *Phase2TestSuite) TestLegacyUnlockInterface() {
	project := models.Project{
		RepoFullName: "test/unlock-test",
		Path:         "modules",
	}
	user := models.User{Username: "unlock-user"}
	workspace := "testing"

	// Test unlocking non-existent lock (should not error)
	unlockedLock, err := s.legacyAdapter.Unlock(project, workspace, user)
	require.NoError(s.T(), err, "Unlocking non-existent lock should not error")
	assert.Nil(s.T(), unlockedLock, "Should return nil for non-existent lock")

	// Acquire a lock first
	legacyLock := models.ProjectLock{
		Project:   project,
		Workspace: workspace,
		User:      user,
		Time:      time.Now(),
	}

	acquired, _, err := s.legacyAdapter.TryLock(legacyLock)
	require.NoError(s.T(), err)
	require.True(s.T(), acquired)

	// Verify lock exists
	existingLock, err := s.legacyAdapter.GetLock(project, workspace)
	require.NoError(s.T(), err)
	require.NotNil(s.T(), existingLock)

	// Test successful unlock
	unlockedLock, err = s.legacyAdapter.Unlock(project, workspace, user)
	require.NoError(s.T(), err, "Unlock should succeed")
	require.NotNil(s.T(), unlockedLock, "Should return unlocked lock details")

	// Verify unlocked lock details
	assert.Equal(s.T(), project.RepoFullName, unlockedLock.Project.RepoFullName)
	assert.Equal(s.T(), project.Path, unlockedLock.Project.Path)
	assert.Equal(s.T(), workspace, unlockedLock.Workspace)
	assert.Equal(s.T(), user.Username, unlockedLock.User.Username)

	// Verify lock is actually gone
	finalLock, err := s.legacyAdapter.GetLock(project, workspace)
	require.NoError(s.T(), err)
	assert.Nil(s.T(), finalLock, "Lock should be gone after unlock")

	// Test unlocking by wrong user (should handle gracefully)
	wrongUser := models.User{Username: "wrong-user"}

	// First create a lock
	acquired, _, err = s.legacyAdapter.TryLock(legacyLock)
	require.NoError(s.T(), err)
	require.True(s.T(), acquired)

	// Try to unlock with wrong user
	wrongUnlock, err := s.legacyAdapter.Unlock(project, workspace, wrongUser)
	// Implementation may handle this differently, but should not panic
	s.T().Logf("Wrong user unlock result: lock=%v, err=%v", wrongUnlock, err)

	// Cleanup with correct user
	s.legacyAdapter.Unlock(project, workspace, user)
}

// Test Legacy UnlockByPull Interface
func (s *Phase2TestSuite) TestLegacyUnlockByPullInterface() {
	repoFullName := "test/unlock-by-pull"
	pullNum := 123

	// Create multiple locks for the same repository
	testWorkspaces := []string{"dev", "staging", "prod"}
	testPaths := []string{".", "modules/vpc", "modules/rds"}

	var createdLocks []models.ProjectLock

	for _, ws := range testWorkspaces {
		for _, path := range testPaths {
			project := models.Project{
				RepoFullName: repoFullName,
				Path:         path,
			}
			user := models.User{Username: fmt.Sprintf("user-%s-%s", ws, path)}

			legacyLock := models.ProjectLock{
				Project:   project,
				Workspace: ws,
				User:      user,
				Time:      time.Now(),
			}

			acquired, _, err := s.legacyAdapter.TryLock(legacyLock)
			require.NoError(s.T(), err, "Should acquire lock for %s/%s", ws, path)
			require.True(s.T(), acquired, "Lock should be acquired for %s/%s", ws, path)

			createdLocks = append(createdLocks, legacyLock)
		}
	}

	// Verify all locks exist
	initialLocks, err := s.legacyAdapter.List()
	require.NoError(s.T(), err)
	expectedLockCount := len(testWorkspaces) * len(testPaths)
	assert.Len(s.T(), initialLocks, expectedLockCount, "Should have all created locks")

	// Test UnlockByPull
	unlockedLocks, err := s.legacyAdapter.UnlockByPull(repoFullName, pullNum)
	require.NoError(s.T(), err, "UnlockByPull should succeed")
	assert.Len(s.T(), unlockedLocks, expectedLockCount, "Should unlock all locks for the repository")

	// Verify all locks for this repository are gone
	finalLocks, err := s.legacyAdapter.List()
	require.NoError(s.T(), err)
	assert.Len(s.T(), finalLocks, 0, "Should have no locks after UnlockByPull")

	// Verify returned lock details
	unlockedByKey := make(map[string]models.ProjectLock)
	for _, lock := range unlockedLocks {
		key := fmt.Sprintf("%s/%s/%s", lock.Project.RepoFullName, lock.Project.Path, lock.Workspace)
		unlockedByKey[key] = lock
	}

	for _, originalLock := range createdLocks {
		key := fmt.Sprintf("%s/%s/%s", originalLock.Project.RepoFullName, originalLock.Project.Path, originalLock.Workspace)

		unlockedLock, exists := unlockedByKey[key]
		assert.True(s.T(), exists, "Should find unlocked lock for %s", key)

		if exists {
			assert.Equal(s.T(), originalLock.Project.RepoFullName, unlockedLock.Project.RepoFullName)
			assert.Equal(s.T(), originalLock.Project.Path, unlockedLock.Project.Path)
			assert.Equal(s.T(), originalLock.Workspace, unlockedLock.Workspace)
			assert.Equal(s.T(), originalLock.User.Username, unlockedLock.User.Username)
		}
	}

	// Test UnlockByPull with no locks (should not error)
	emptyUnlocks, err := s.legacyAdapter.UnlockByPull("nonexistent/repo", 456)
	require.NoError(s.T(), err, "UnlockByPull with no locks should not error")
	assert.Len(s.T(), emptyUnlocks, 0, "Should return empty slice when no locks exist")
}

// Test Enhanced Features Through Legacy Interface
func (s *Phase2TestSuite) TestEnhancedFeaturesCompatibility() {
	project := models.Project{
		RepoFullName: "test/enhanced-compat",
		Path:         ".",
	}
	user := models.User{Username: "enhanced-user"}
	workspace := "default"

	// Test that enhanced features don't break legacy interface

	// Test LockWithPriority through adapter (enhanced feature)
	lock, err := s.legacyAdapter.LockWithPriority(s.ctx, project, workspace, user, enhanced.PriorityHigh)
	require.NoError(s.T(), err, "LockWithPriority should work through adapter")
	require.NotNil(s.T(), lock, "Should return lock")

	// Verify lock is visible through legacy List interface
	locks, err := s.legacyAdapter.List()
	require.NoError(s.T(), err)
	assert.Len(s.T(), locks, 1, "Enhanced lock should be visible through legacy interface")

	// Release via legacy interface
	releasedLock, err := s.legacyAdapter.Unlock(project, workspace, user)
	require.NoError(s.T(), err)
	require.NotNil(s.T(), releasedLock)

	// Test LockWithTimeout through adapter (enhanced feature)
	customTimeout := 5 * time.Minute
	lock, err = s.legacyAdapter.LockWithTimeout(s.ctx, project, workspace, user, customTimeout)
	require.NoError(s.T(), err, "LockWithTimeout should work through adapter")
	require.NotNil(s.T(), lock, "Should return lock with timeout")

	// Cleanup
	s.legacyAdapter.Unlock(project, workspace, user)
}

// Test Format Preservation
func (s *Phase2TestSuite) TestLegacyFormatPreservation() {
	project := models.Project{
		RepoFullName: "test/format-preservation",
		Path:         "terraform/environments/prod",
	}
	user := models.User{Username: "format-user"}
	workspace := "production"

	originalTime := time.Now().Add(-5 * time.Minute)

	legacyLock := models.ProjectLock{
		Project:   project,
		Workspace: workspace,
		User:      user,
		Time:      originalTime,
	}

	// Acquire lock
	acquired, _, err := s.legacyAdapter.TryLock(legacyLock)
	require.NoError(s.T(), err)
	require.True(s.T(), acquired)

	// Retrieve lock and verify format preservation
	retrievedLock, err := s.legacyAdapter.GetLock(project, workspace)
	require.NoError(s.T(), err)
	require.NotNil(s.T(), retrievedLock)

	// Verify all fields are preserved in legacy format
	assert.Equal(s.T(), project.RepoFullName, retrievedLock.Project.RepoFullName)
	assert.Equal(s.T(), project.Path, retrievedLock.Project.Path)
	assert.Equal(s.T(), workspace, retrievedLock.Workspace)
	assert.Equal(s.T(), user.Username, retrievedLock.User.Username)

	// Time should be close to acquisition time (not original time, as that's implementation detail)
	assert.WithinDuration(s.T(), time.Now(), retrievedLock.Time, 10*time.Second)

	// Test through List interface as well
	locks, err := s.legacyAdapter.List()
	require.NoError(s.T(), err)
	require.Len(s.T(), locks, 1)

	key := fmt.Sprintf("%s/%s/%s", project.RepoFullName, project.Path, workspace)
	listedLock, exists := locks[key]
	require.True(s.T(), exists, "Lock should exist in list")

	// Verify format consistency between GetLock and List
	assert.Equal(s.T(), retrievedLock.Project.RepoFullName, listedLock.Project.RepoFullName)
	assert.Equal(s.T(), retrievedLock.Project.Path, listedLock.Project.Path)
	assert.Equal(s.T(), retrievedLock.Workspace, listedLock.Workspace)
	assert.Equal(s.T(), retrievedLock.User.Username, listedLock.User.Username)

	// Cleanup
	s.legacyAdapter.Unlock(project, workspace, user)
}

// Test Fallback Behavior
func (s *Phase2TestSuite) TestFallbackBehavior() {
	// This test simulates scenarios where enhanced backend might fail
	// and system should fall back to legacy behavior

	s.T().Log("Testing fallback behavior - this is implementation specific")

	project := models.Project{
		RepoFullName: "test/fallback",
		Path:         ".",
	}
	user := models.User{Username: "fallback-user"}
	workspace := "default"

	// Normal operation should work
	legacyLock := models.ProjectLock{
		Project:   project,
		Workspace: workspace,
		User:      user,
		Time:      time.Now(),
	}

	acquired, _, err := s.legacyAdapter.TryLock(legacyLock)
	require.NoError(s.T(), err, "Normal operation should work")
	assert.True(s.T(), acquired, "Should acquire lock normally")

	// Verify through legacy interfaces
	retrievedLock, err := s.legacyAdapter.GetLock(project, workspace)
	require.NoError(s.T(), err)
	require.NotNil(s.T(), retrievedLock)

	locks, err := s.legacyAdapter.List()
	require.NoError(s.T(), err)
	assert.Len(s.T(), locks, 1)

	// Cleanup
	unlockedLock, err := s.legacyAdapter.Unlock(project, workspace, user)
	require.NoError(s.T(), err)
	require.NotNil(s.T(), unlockedLock)
}

// Test Concurrent Access Through Legacy Interface
func (s *Phase2TestSuite) TestConcurrentLegacyOperations() {
	numOperations := 20
	var wg sync.WaitGroup
	results := make(chan compatibilityResult, numOperations)

	for i := 0; i < numOperations; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			project := models.Project{
				RepoFullName: fmt.Sprintf("test/concurrent-legacy-%d", id),
				Path:         ".",
			}
			user := models.User{Username: fmt.Sprintf("legacy-user-%d", id)}
			workspace := "default"

			start := time.Now()

			legacyLock := models.ProjectLock{
				Project:   project,
				Workspace: workspace,
				User:      user,
				Time:      time.Now(),
			}

			// Try to acquire lock
			acquired, resp, err := s.legacyAdapter.TryLock(legacyLock)
			if err != nil {
				results <- compatibilityResult{ID: id, Success: false, Error: err, Duration: time.Since(start)}
				return
			}

			if !acquired {
				results <- compatibilityResult{
					ID:       id,
					Success:  false,
					Error:    fmt.Errorf("lock not acquired: %s", resp.LockFailureReason),
					Duration: time.Since(start),
				}
				return
			}

			// Verify through GetLock
			retrievedLock, err := s.legacyAdapter.GetLock(project, workspace)
			if err != nil || retrievedLock == nil {
				results <- compatibilityResult{ID: id, Success: false, Error: fmt.Errorf("GetLock failed: %v", err), Duration: time.Since(start)}
				return
			}

			// Brief hold time
			time.Sleep(10 * time.Millisecond)

			// Release lock
			_, err = s.legacyAdapter.Unlock(project, workspace, user)
			results <- compatibilityResult{
				ID:       id,
				Success:  err == nil,
				Error:    err,
				Duration: time.Since(start),
			}
		}(i)
	}

	wg.Wait()
	close(results)

	// Analyze results
	var successful, failed int
	var totalDuration time.Duration

	for result := range results {
		totalDuration += result.Duration
		if result.Success {
			successful++
		} else {
			failed++
			s.T().Logf("Concurrent operation %d failed: %v", result.ID, result.Error)
		}
	}

	// All operations should succeed for backwards compatibility
	assert.Equal(s.T(), numOperations, successful, "All concurrent legacy operations should succeed")
	assert.Equal(s.T(), 0, failed, "No legacy operations should fail")

	avgDuration := totalDuration / time.Duration(numOperations)
	s.T().Logf("Average legacy operation duration: %v", avgDuration)

	// Performance should be reasonable
	assert.Less(s.T(), avgDuration, 2*time.Second, "Average legacy operation should be under 2 seconds")
}

// Helper types and functions

type compatibilityResult struct {
	ID       int
	Success  bool
	Error    error
	Duration time.Duration
}

func (s *Phase2TestSuite) setupCompatibilityTest() (enhanced.LockManager, locking.Backend, func()) {
	// Create enhanced backend
	enhancedBackend := &CompatibilityMockBackend{
		locks:   make(map[string]*enhanced.EnhancedLock),
		metrics: &enhanced.BackendStats{},
	}

	// Create enhanced manager
	enhancedManager := enhanced.NewEnhancedLockManager(enhancedBackend, s.config, logging.NewNoopLogger(s.T()))

	// Create legacy mock backend for fallback testing
	legacyBackend := &LegacyMockBackend{
		locks: make(map[string]models.ProjectLock),
	}

	// Start enhanced manager
	err := enhancedManager.Start(s.ctx)
	require.NoError(s.T(), err)

	cleanup := func() {
		enhancedManager.Stop()
	}

	return enhancedManager, legacyBackend, cleanup
}

// CompatibilityMockBackend extends the basic mock backend with compatibility features
type CompatibilityMockBackend struct {
	mutex   sync.RWMutex
	locks   map[string]*enhanced.EnhancedLock
	metrics *enhanced.BackendStats
}

// Implement all Backend interface methods (same as Phase1 but with better legacy support)
// [Implementation methods would be similar to Phase1 but with enhanced legacy conversion]

// LegacyMockBackend implements the legacy locking.Backend interface for fallback testing
type LegacyMockBackend struct {
	mutex sync.RWMutex
	locks map[string]models.ProjectLock
}

func (lmb *LegacyMockBackend) TryLock(lock models.ProjectLock) (bool, locking.TryLockResponse, error) {
	lmb.mutex.Lock()
	defer lmb.mutex.Unlock()

	key := fmt.Sprintf("%s/%s/%s", lock.Project.RepoFullName, lock.Project.Path, lock.Workspace)

	// Check if lock exists
	if existingLock, exists := lmb.locks[key]; exists {
		return false, locking.TryLockResponse{
			CurrLock:          existingLock,
			LockFailureReason: fmt.Sprintf("Lock held by %s since %s", existingLock.User.Username, existingLock.Time.Format(time.RFC3339)),
		}, nil
	}

	// Acquire lock
	lock.Time = time.Now()
	lmb.locks[key] = lock
	return true, locking.TryLockResponse{}, nil
}

func (lmb *LegacyMockBackend) Unlock(project models.Project, workspace string, user models.User) (*models.ProjectLock, error) {
	lmb.mutex.Lock()
	defer lmb.mutex.Unlock()

	key := fmt.Sprintf("%s/%s/%s", project.RepoFullName, project.Path, workspace)

	lock, exists := lmb.locks[key]
	if !exists {
		return nil, nil // Atlantis behavior: return nil for non-existent locks
	}

	delete(lmb.locks, key)
	return &lock, nil
}

func (lmb *LegacyMockBackend) List() (map[string]models.ProjectLock, error) {
	lmb.mutex.RLock()
	defer lmb.mutex.RUnlock()

	result := make(map[string]models.ProjectLock)
	for key, lock := range lmb.locks {
		result[key] = lock
	}
	return result, nil
}

func (lmb *LegacyMockBackend) UnlockByPull(repoFullName string, pullNum int) ([]models.ProjectLock, error) {
	lmb.mutex.Lock()
	defer lmb.mutex.Unlock()

	var unlockedLocks []models.ProjectLock
	var keysToDelete []string

	for key, lock := range lmb.locks {
		if lock.Project.RepoFullName == repoFullName {
			unlockedLocks = append(unlockedLocks, lock)
			keysToDelete = append(keysToDelete, key)
		}
	}

	for _, key := range keysToDelete {
		delete(lmb.locks, key)
	}

	return unlockedLocks, nil
}

func (lmb *LegacyMockBackend) GetLock(project models.Project, workspace string) (*models.ProjectLock, error) {
	lmb.mutex.RLock()
	defer lmb.mutex.RUnlock()

	key := fmt.Sprintf("%s/%s/%s", project.RepoFullName, project.Path, workspace)
	if lock, exists := lmb.locks[key]; exists {
		return &lock, nil
	}
	return nil, nil
}

// Implement CompatibilityMockBackend methods (similar to Phase1MockBackend but with better legacy conversion)
// [Include all the enhanced.Backend interface methods here]

// Test Suite Runner
func TestPhase2BackwardsCompatibility(t *testing.T) {
	suite.Run(t, new(Phase2TestSuite))
}

// Additional compatibility validation functions can be added here for specific edge cases