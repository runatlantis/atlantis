// Phase 1 Testing Strategy: Basic Enhanced Locking Functionality
// This phase tests the core enhanced locking system without advanced features

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

	"github.com/runatlantis/atlantis/server/core/locking/enhanced"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/logging"
)

// Phase1TestSuite validates basic enhanced locking functionality
type Phase1TestSuite struct {
	suite.Suite
	manager enhanced.LockManager
	config  *enhanced.EnhancedConfig
	cleanup func()
	ctx     context.Context
}

// SetupSuite initializes the test environment for Phase 1
func (s *Phase1TestSuite) SetupSuite() {
	s.ctx = context.Background()

	// Phase 1 configuration: basic features only
	s.config = &enhanced.EnhancedConfig{
		Enabled:                 true,
		Backend:                 "boltdb", // Start with simple backend
		DefaultTimeout:          30 * time.Minute,
		MaxTimeout:              2 * time.Hour,

		// Disable advanced features in Phase 1
		EnablePriorityQueue:     false,
		EnableRetry:            false,
		EnableDeadlockDetection: false,
		EnableEvents:           false,

		// Enable backward compatibility
		LegacyFallback:         true,
		PreserveLegacyFormat:   true,
	}

	s.manager, s.cleanup = s.setupTestManager()
}

func (s *Phase1TestSuite) TearDownSuite() {
	if s.cleanup != nil {
		s.cleanup()
	}
}

// Test Core Lock Acquisition and Release
func (s *Phase1TestSuite) TestBasicLockAcquisition() {
	project := models.Project{
		RepoFullName: "test/phase1-basic",
		Path:         ".",
	}
	user := models.User{Username: "phase1-user"}
	workspace := "default"

	// Test 1: Acquire lock
	lock, err := s.manager.Lock(s.ctx, project, workspace, user)
	require.NoError(s.T(), err, "Should successfully acquire lock")
	require.NotNil(s.T(), lock, "Lock should not be nil")

	// Validate lock properties
	assert.Equal(s.T(), project.RepoFullName, lock.Project.RepoFullName)
	assert.Equal(s.T(), workspace, lock.Workspace)
	assert.Equal(s.T(), user.Username, lock.User.Username)
	assert.WithinDuration(s.T(), time.Now(), lock.Time, time.Second)

	// Test 2: List locks shows our lock
	locks, err := s.manager.List(s.ctx)
	require.NoError(s.T(), err, "Should list locks without error")
	assert.Len(s.T(), locks, 1, "Should have exactly one lock")

	foundLock := locks[0]
	assert.Equal(s.T(), project.RepoFullName, foundLock.Project.RepoFullName)
	assert.Equal(s.T(), workspace, foundLock.Workspace)
	assert.Equal(s.T(), user.Username, foundLock.User.Username)

	// Test 3: Release lock
	releasedLock, err := s.manager.Unlock(s.ctx, project, workspace, user)
	require.NoError(s.T(), err, "Should release lock without error")
	require.NotNil(s.T(), releasedLock, "Released lock should not be nil")

	// Test 4: Verify lock is gone
	locks, err = s.manager.List(s.ctx)
	require.NoError(s.T(), err, "Should list locks after unlock")
	assert.Len(s.T(), locks, 0, "Should have no locks after unlock")
}

// Test Lock Conflict Detection
func (s *Phase1TestSuite) TestLockConflictDetection() {
	project := models.Project{
		RepoFullName: "test/phase1-conflict",
		Path:         ".",
	}
	user1 := models.User{Username: "user1"}
	user2 := models.User{Username: "user2"}
	workspace := "default"

	// User1 acquires lock
	lock1, err := s.manager.Lock(s.ctx, project, workspace, user1)
	require.NoError(s.T(), err, "User1 should acquire lock")
	require.NotNil(s.T(), lock1, "Lock1 should not be nil")

	// User2 tries to acquire same lock - should fail
	lock2, err := s.manager.Lock(s.ctx, project, workspace, user2)
	assert.Error(s.T(), err, "User2 should not acquire conflicting lock")
	assert.Nil(s.T(), lock2, "Lock2 should be nil due to conflict")

	// Verify error is lock exists error
	lockErr, ok := err.(*enhanced.LockError)
	if assert.True(s.T(), ok, "Error should be LockError type") {
		assert.Equal(s.T(), enhanced.ErrCodeLockExists, lockErr.Code)
	}

	// Only user1's lock should exist
	locks, err := s.manager.List(s.ctx)
	require.NoError(s.T(), err)
	assert.Len(s.T(), locks, 1)
	assert.Equal(s.T(), user1.Username, locks[0].User.Username)

	// Cleanup
	s.manager.Unlock(s.ctx, project, workspace, user1)
}

// Test Multiple Workspaces
func (s *Phase1TestSuite) TestMultipleWorkspaces() {
	project := models.Project{
		RepoFullName: "test/phase1-multi-ws",
		Path:         ".",
	}
	user := models.User{Username: "multi-ws-user"}
	workspaces := []string{"dev", "staging", "prod"}

	// Acquire locks for different workspaces
	var locks []*models.ProjectLock
	for _, ws := range workspaces {
		lock, err := s.manager.Lock(s.ctx, project, ws, user)
		require.NoError(s.T(), err, "Should acquire lock for workspace %s", ws)
		require.NotNil(s.T(), lock, "Lock should not be nil for workspace %s", ws)
		locks = append(locks, lock)
	}

	// Verify all locks exist
	allLocks, err := s.manager.List(s.ctx)
	require.NoError(s.T(), err)
	assert.Len(s.T(), allLocks, len(workspaces), "Should have locks for all workspaces")

	// Verify each workspace has its lock
	locksByWs := make(map[string]*models.ProjectLock)
	for _, lock := range allLocks {
		locksByWs[lock.Workspace] = lock
	}

	for _, ws := range workspaces {
		assert.Contains(s.T(), locksByWs, ws, "Should have lock for workspace %s", ws)
		assert.Equal(s.T(), project.RepoFullName, locksByWs[ws].Project.RepoFullName)
		assert.Equal(s.T(), user.Username, locksByWs[ws].User.Username)
	}

	// Release all locks
	for _, ws := range workspaces {
		releasedLock, err := s.manager.Unlock(s.ctx, project, ws, user)
		require.NoError(s.T(), err, "Should release lock for workspace %s", ws)
		require.NotNil(s.T(), releasedLock, "Released lock should not be nil for workspace %s", ws)
	}

	// Verify all locks are gone
	finalLocks, err := s.manager.List(s.ctx)
	require.NoError(s.T(), err)
	assert.Len(s.T(), finalLocks, 0, "Should have no locks after releasing all")
}

// Test Lock Timeout Configuration
func (s *Phase1TestSuite) TestLockTimeoutConfiguration() {
	project := models.Project{
		RepoFullName: "test/phase1-timeout",
		Path:         ".",
	}
	user := models.User{Username: "timeout-user"}
	workspace := "default"

	// Test with custom timeout
	customTimeout := 5 * time.Second
	lock, err := s.manager.LockWithTimeout(s.ctx, project, workspace, user, customTimeout)
	require.NoError(s.T(), err, "Should acquire lock with custom timeout")
	require.NotNil(s.T(), lock, "Lock should not be nil")

	// Verify lock exists
	locks, err := s.manager.List(s.ctx)
	require.NoError(s.T(), err)
	assert.Len(s.T(), locks, 1, "Should have one lock")

	// Note: Timeout behavior testing depends on implementation
	// In Phase 1, we mainly validate timeout can be set without errors
	s.T().Logf("Lock acquired with %v timeout", customTimeout)

	// Cleanup
	s.manager.Unlock(s.ctx, project, workspace, user)
}

// Test Concurrent Access Without Advanced Features
func (s *Phase1TestSuite) TestConcurrentBasicOperations() {
	numOperations := 10
	var wg sync.WaitGroup
	results := make(chan operationResult, numOperations)

	for i := 0; i < numOperations; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			project := models.Project{
				RepoFullName: fmt.Sprintf("test/concurrent-%d", id),
				Path:         ".",
			}
			user := models.User{Username: fmt.Sprintf("user-%d", id)}
			workspace := "default"

			start := time.Now()

			// Acquire lock
			lock, err := s.manager.Lock(s.ctx, project, workspace, user)
			if err != nil {
				results <- operationResult{ID: id, Success: false, Error: err, Duration: time.Since(start)}
				return
			}

			// Hold lock briefly
			time.Sleep(10 * time.Millisecond)

			// Release lock
			_, err = s.manager.Unlock(s.ctx, project, workspace, user)
			results <- operationResult{
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
			s.T().Logf("Operation %d failed: %v", result.ID, result.Error)
		}
	}

	// All operations should succeed in Phase 1 basic scenario
	assert.Equal(s.T(), numOperations, successful, "All concurrent operations should succeed")
	assert.Equal(s.T(), 0, failed, "No operations should fail")

	avgDuration := totalDuration / time.Duration(numOperations)
	s.T().Logf("Average operation duration: %v", avgDuration)

	// Performance expectation for basic operations
	assert.Less(s.T(), avgDuration, time.Second, "Average operation should be under 1 second")
}

// Test Error Handling
func (s *Phase1TestSuite) TestErrorHandling() {
	// Test unlocking non-existent lock
	project := models.Project{
		RepoFullName: "test/non-existent",
		Path:         ".",
	}
	user := models.User{Username: "test-user"}
	workspace := "default"

	// Should return nil without error (Atlantis behavior)
	releasedLock, err := s.manager.Unlock(s.ctx, project, workspace, user)
	assert.NoError(s.T(), err, "Unlocking non-existent lock should not error")
	assert.Nil(s.T(), releasedLock, "Should return nil for non-existent lock")

	// Test invalid project/workspace combinations
	invalidProject := models.Project{
		RepoFullName: "",
		Path:         "",
	}

	lock, err := s.manager.Lock(s.ctx, invalidProject, workspace, user)
	// Implementation may handle this differently, but should not panic
	s.T().Logf("Lock attempt with invalid project: err=%v, lock=%v", err, lock)
}

// Helper types and functions

type operationResult struct {
	ID       int
	Success  bool
	Error    error
	Duration time.Duration
}

func (s *Phase1TestSuite) setupTestManager() (enhanced.LockManager, func()) {
	backend := &MockBackend{
		locks:   make(map[string]*enhanced.EnhancedLock),
		metrics: &enhanced.BackendStats{},
	}

	manager := enhanced.NewEnhancedLockManager(backend, s.config, logging.NewNoopLogger(s.T()))

	err := manager.Start(s.ctx)
	require.NoError(s.T(), err)

	cleanup := func() {
		manager.Stop()
	}

	return manager, cleanup
}

// MockBackend - simplified version for Phase 1 testing
type MockBackend struct {
	mutex   sync.RWMutex
	locks   map[string]*enhanced.EnhancedLock
	metrics *enhanced.BackendStats
}

func (mb *MockBackend) AcquireLock(ctx context.Context, request *enhanced.EnhancedLockRequest) (*enhanced.EnhancedLock, error) {
	mb.mutex.Lock()
	defer mb.mutex.Unlock()

	resourceKey := fmt.Sprintf("%s/%s/%s", request.Resource.Namespace, request.Resource.Name, request.Resource.Workspace)

	// Check for existing lock
	for _, lock := range mb.locks {
		if fmt.Sprintf("%s/%s/%s", lock.Resource.Namespace, lock.Resource.Name, lock.Resource.Workspace) == resourceKey &&
		   lock.State == enhanced.LockStateAcquired {
			return nil, enhanced.NewLockExistsError(resourceKey)
		}
	}

	// Create new lock
	lock := &enhanced.EnhancedLock{
		ID:         request.ID,
		Resource:   request.Resource,
		State:      enhanced.LockStateAcquired,
		Priority:   request.Priority,
		Owner:      request.User.Username,
		AcquiredAt: time.Now(),
		Metadata:   request.Metadata,
		Version:    1,
		OriginalLock: &models.ProjectLock{
			Project:   request.Project,
			Workspace: request.Workspace,
			User:      request.User,
			Time:      time.Now(),
		},
	}

	if request.Timeout > 0 {
		expiresAt := time.Now().Add(request.Timeout)
		lock.ExpiresAt = &expiresAt
	}

	mb.locks[lock.ID] = lock
	mb.metrics.TotalRequests++
	mb.metrics.SuccessfulAcquires++
	mb.metrics.ActiveLocks++

	return lock, nil
}

func (mb *MockBackend) TryAcquireLock(ctx context.Context, request *enhanced.EnhancedLockRequest) (*enhanced.EnhancedLock, bool, error) {
	lock, err := mb.AcquireLock(ctx, request)
	if err != nil {
		return nil, false, err
	}
	return lock, true, nil
}

func (mb *MockBackend) ReleaseLock(ctx context.Context, lockID string) error {
	mb.mutex.Lock()
	defer mb.mutex.Unlock()

	if _, exists := mb.locks[lockID]; !exists {
		return enhanced.NewLockNotFoundError(lockID)
	}

	delete(mb.locks, lockID)
	mb.metrics.ActiveLocks--
	return nil
}

func (mb *MockBackend) GetLock(ctx context.Context, lockID string) (*enhanced.EnhancedLock, error) {
	mb.mutex.RLock()
	defer mb.mutex.RUnlock()

	lock, exists := mb.locks[lockID]
	if !exists {
		return nil, enhanced.NewLockNotFoundError(lockID)
	}
	return lock, nil
}

func (mb *MockBackend) ListLocks(ctx context.Context) ([]*enhanced.EnhancedLock, error) {
	mb.mutex.RLock()
	defer mb.mutex.RUnlock()

	locks := make([]*enhanced.EnhancedLock, 0, len(mb.locks))
	for _, lock := range mb.locks {
		if lock.State == enhanced.LockStateAcquired {
			locks = append(locks, lock)
		}
	}
	return locks, nil
}

// Implement remaining Backend interface methods with basic functionality
func (mb *MockBackend) RefreshLock(ctx context.Context, lockID string, extension time.Duration) error {
	return nil // Basic implementation
}

func (mb *MockBackend) TransferLock(ctx context.Context, lockID string, newOwner string) error {
	return nil // Basic implementation
}

func (mb *MockBackend) EnqueueLockRequest(ctx context.Context, request *enhanced.EnhancedLockRequest) error {
	return nil // Not used in Phase 1
}

func (mb *MockBackend) DequeueNextRequest(ctx context.Context) (*enhanced.EnhancedLockRequest, error) {
	return nil, nil // Not used in Phase 1
}

func (mb *MockBackend) GetQueueStatus(ctx context.Context) (*enhanced.QueueStatus, error) {
	return &enhanced.QueueStatus{Size: 0, PendingRequests: []*enhanced.EnhancedLockRequest{}}, nil
}

func (mb *MockBackend) HealthCheck(ctx context.Context) error {
	return nil
}

func (mb *MockBackend) GetStats(ctx context.Context) (*enhanced.BackendStats, error) {
	mb.mutex.RLock()
	defer mb.mutex.RUnlock()

	mb.metrics.ActiveLocks = int64(len(mb.locks))
	mb.metrics.LastUpdated = time.Now()
	return mb.metrics, nil
}

func (mb *MockBackend) Subscribe(ctx context.Context, eventTypes []string) (<-chan *enhanced.LockEvent, error) {
	eventChan := make(chan *enhanced.LockEvent)
	close(eventChan)
	return eventChan, nil
}

func (mb *MockBackend) CleanupExpiredLocks(ctx context.Context) (int, error) {
	return 0, nil
}

func (mb *MockBackend) GetLegacyLock(project models.Project, workspace string) (*models.ProjectLock, error) {
	mb.mutex.RLock()
	defer mb.mutex.RUnlock()

	for _, lock := range mb.locks {
		if lock.Resource.Namespace == project.RepoFullName &&
		   lock.Resource.Path == project.Path &&
		   lock.Resource.Workspace == workspace &&
		   lock.State == enhanced.LockStateAcquired {
			return lock.OriginalLock, nil
		}
	}
	return nil, nil
}

func (mb *MockBackend) ConvertToLegacy(lock *enhanced.EnhancedLock) *models.ProjectLock {
	if lock.OriginalLock != nil {
		return lock.OriginalLock
	}
	return &models.ProjectLock{
		Project: models.Project{
			RepoFullName: lock.Resource.Namespace,
			Path:         lock.Resource.Path,
		},
		Workspace: lock.Resource.Workspace,
		User:      models.User{Username: lock.Owner},
		Time:      lock.AcquiredAt,
	}
}

func (mb *MockBackend) ConvertFromLegacy(lock *models.ProjectLock) *enhanced.EnhancedLock {
	return &enhanced.EnhancedLock{
		ID: fmt.Sprintf("legacy_%d", time.Now().UnixNano()),
		Resource: enhanced.ResourceIdentifier{
			Type:      enhanced.ResourceTypeProject,
			Namespace: lock.Project.RepoFullName,
			Name:      lock.Project.Path,
			Workspace: lock.Workspace,
			Path:      lock.Project.Path,
		},
		State:        enhanced.LockStateAcquired,
		Priority:     enhanced.PriorityNormal,
		Owner:        lock.User.Username,
		AcquiredAt:   lock.Time,
		Version:      1,
		OriginalLock: lock,
	}
}

// Test Suite Runner
func TestPhase1EnhancedLocking(t *testing.T) {
	suite.Run(t, new(Phase1TestSuite))
}