package enhanced

import (
	"context"
	"testing"
	"time"

	"github.com/runatlantis/atlantis/server/core/locking"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/logging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockLegacyBackend implements the legacy locking.Backend interface for testing
type MockLegacyBackend struct {
	locks map[string]models.ProjectLock
	calls []string // Track method calls for verification
}

func NewMockLegacyBackend() *MockLegacyBackend {
	return &MockLegacyBackend{
		locks: make(map[string]models.ProjectLock),
		calls: make([]string, 0),
	}
}

func (m *MockLegacyBackend) TryLock(lock models.ProjectLock) (bool, models.ProjectLock, error) {
	m.calls = append(m.calls, "TryLock")
	key := models.GenerateLockKey(lock.Project, lock.Workspace)

	if existing, exists := m.locks[key]; exists {
		return false, existing, nil
	}

	m.locks[key] = lock
	return true, models.ProjectLock{}, nil
}

func (m *MockLegacyBackend) Unlock(project models.Project, workspace string) (*models.ProjectLock, error) {
	m.calls = append(m.calls, "Unlock")
	key := models.GenerateLockKey(project, workspace)

	if lock, exists := m.locks[key]; exists {
		delete(m.locks, key)
		return &lock, nil
	}

	return nil, nil
}

func (m *MockLegacyBackend) List() ([]models.ProjectLock, error) {
	m.calls = append(m.calls, "List")
	var locks []models.ProjectLock
	for _, lock := range m.locks {
		locks = append(locks, lock)
	}
	return locks, nil
}

func (m *MockLegacyBackend) GetLock(project models.Project, workspace string) (*models.ProjectLock, error) {
	m.calls = append(m.calls, "GetLock")
	key := models.GenerateLockKey(project, workspace)

	if lock, exists := m.locks[key]; exists {
		return &lock, nil
	}

	return nil, nil
}

func (m *MockLegacyBackend) UnlockByPull(repoFullName string, pullNum int) ([]models.ProjectLock, error) {
	m.calls = append(m.calls, "UnlockByPull")
	var unlockedLocks []models.ProjectLock

	for key, lock := range m.locks {
		if lock.Project.RepoFullName == repoFullName {
			unlockedLocks = append(unlockedLocks, lock)
			delete(m.locks, key)
		}
	}

	return unlockedLocks, nil
}

// Additional legacy backend methods for full interface compliance
func (m *MockLegacyBackend) UpdateProjectStatus(pull models.PullRequest, workspace string, repoRelDir string, newStatus models.ProjectPlanStatus) error {
	m.calls = append(m.calls, "UpdateProjectStatus")
	return nil
}

func (m *MockLegacyBackend) GetPullStatus(pull models.PullRequest) (*models.PullStatus, error) {
	m.calls = append(m.calls, "GetPullStatus")
	return nil, nil
}

func (m *MockLegacyBackend) DeletePullStatus(pull models.PullRequest) error {
	m.calls = append(m.calls, "DeletePullStatus")
	return nil
}

func (m *MockLegacyBackend) UpdatePullWithResults(pull models.PullRequest, newResults []interface{}) (models.PullStatus, error) {
	m.calls = append(m.calls, "UpdatePullWithResults")
	return models.PullStatus{}, nil
}

func (m *MockLegacyBackend) LockCommand(cmdName interface{}, lockTime time.Time) (interface{}, error) {
	m.calls = append(m.calls, "LockCommand")
	return nil, nil
}

func (m *MockLegacyBackend) UnlockCommand(cmdName interface{}) error {
	m.calls = append(m.calls, "UnlockCommand")
	return nil
}

func (m *MockLegacyBackend) CheckCommandLock(cmdName interface{}) (interface{}, error) {
	m.calls = append(m.calls, "CheckCommandLock")
	return nil, nil
}

// MockEnhancedBackend implements the enhanced Backend interface for testing
type MockEnhancedBackend struct {
	locks     map[string]*EnhancedLock
	calls     []string
	healthy   bool
	shouldFail bool
}

func NewMockEnhancedBackend() *MockEnhancedBackend {
	return &MockEnhancedBackend{
		locks:   make(map[string]*EnhancedLock),
		calls:   make([]string, 0),
		healthy: true,
	}
}

func (m *MockEnhancedBackend) AcquireLock(ctx context.Context, request *EnhancedLockRequest) (*EnhancedLock, error) {
	m.calls = append(m.calls, "AcquireLock")
	if m.shouldFail {
		return nil, assert.AnError
	}

	key := models.GenerateLockKey(request.Project, request.Workspace)

	if _, exists := m.locks[key]; exists {
		return nil, NewLockExistsError(key)
	}

	lock := &EnhancedLock{
		ID:       request.ID,
		Resource: request.Resource,
		State:    LockStateAcquired,
		Priority: request.Priority,
		Owner:    request.User.Username,
		AcquiredAt: time.Now(),
		Metadata: request.Metadata,
		OriginalLock: &models.ProjectLock{
			Project:   request.Project,
			Workspace: request.Workspace,
			User:      request.User,
			Time:      time.Now(),
		},
	}

	m.locks[key] = lock
	return lock, nil
}

func (m *MockEnhancedBackend) TryAcquireLock(ctx context.Context, request *EnhancedLockRequest) (*EnhancedLock, bool, error) {
	m.calls = append(m.calls, "TryAcquireLock")
	if m.shouldFail {
		return nil, false, assert.AnError
	}

	lock, err := m.AcquireLock(ctx, request)
	if err != nil {
		if _, ok := err.(*LockError); ok {
			return nil, false, nil // Lock exists, not an error
		}
		return nil, false, err
	}

	return lock, true, nil
}

func (m *MockEnhancedBackend) ReleaseLock(ctx context.Context, lockID string) error {
	m.calls = append(m.calls, "ReleaseLock")
	if m.shouldFail {
		return assert.AnError
	}

	for key, lock := range m.locks {
		if lock.ID == lockID {
			delete(m.locks, key)
			return nil
		}
	}

	return NewLockNotFoundError(lockID)
}

func (m *MockEnhancedBackend) GetLock(ctx context.Context, lockID string) (*EnhancedLock, error) {
	m.calls = append(m.calls, "GetLock")
	if m.shouldFail {
		return nil, assert.AnError
	}

	for _, lock := range m.locks {
		if lock.ID == lockID {
			return lock, nil
		}
	}

	return nil, NewLockNotFoundError(lockID)
}

func (m *MockEnhancedBackend) ListLocks(ctx context.Context) ([]*EnhancedLock, error) {
	m.calls = append(m.calls, "ListLocks")
	if m.shouldFail {
		return nil, assert.AnError
	}

	var locks []*EnhancedLock
	for _, lock := range m.locks {
		locks = append(locks, lock)
	}
	return locks, nil
}

func (m *MockEnhancedBackend) RefreshLock(ctx context.Context, lockID string, extension time.Duration) error {
	m.calls = append(m.calls, "RefreshLock")
	return nil
}

func (m *MockEnhancedBackend) TransferLock(ctx context.Context, lockID string, newOwner string) error {
	m.calls = append(m.calls, "TransferLock")
	return nil
}

func (m *MockEnhancedBackend) EnqueueLockRequest(ctx context.Context, request *EnhancedLockRequest) error {
	m.calls = append(m.calls, "EnqueueLockRequest")
	return nil
}

func (m *MockEnhancedBackend) DequeueNextRequest(ctx context.Context) (*EnhancedLockRequest, error) {
	m.calls = append(m.calls, "DequeueNextRequest")
	return nil, nil
}

func (m *MockEnhancedBackend) GetQueueStatus(ctx context.Context) (*QueueStatus, error) {
	m.calls = append(m.calls, "GetQueueStatus")
	return &QueueStatus{}, nil
}

func (m *MockEnhancedBackend) HealthCheck(ctx context.Context) error {
	m.calls = append(m.calls, "HealthCheck")
	if !m.healthy {
		return assert.AnError
	}
	return nil
}

func (m *MockEnhancedBackend) GetStats(ctx context.Context) (*BackendStats, error) {
	m.calls = append(m.calls, "GetStats")
	return &BackendStats{}, nil
}

func (m *MockEnhancedBackend) Subscribe(ctx context.Context, eventTypes []string) (<-chan *LockEvent, error) {
	m.calls = append(m.calls, "Subscribe")
	return nil, nil
}

func (m *MockEnhancedBackend) CleanupExpiredLocks(ctx context.Context) (int, error) {
	m.calls = append(m.calls, "CleanupExpiredLocks")
	return 0, nil
}

func (m *MockEnhancedBackend) GetLegacyLock(project models.Project, workspace string) (*models.ProjectLock, error) {
	m.calls = append(m.calls, "GetLegacyLock")
	key := models.GenerateLockKey(project, workspace)

	if lock, exists := m.locks[key]; exists {
		return lock.OriginalLock, nil
	}

	return nil, nil
}

func (m *MockEnhancedBackend) ConvertToLegacy(lock *EnhancedLock) *models.ProjectLock {
	return lock.OriginalLock
}

func (m *MockEnhancedBackend) ConvertFromLegacy(lock *models.ProjectLock) *EnhancedLock {
	return &EnhancedLock{
		ID:       generateRequestID(),
		Resource: ResourceIdentifier{
			Type:      ResourceTypeProject,
			Namespace: lock.Project.RepoFullName,
			Name:      lock.Project.Path,
			Workspace: lock.Workspace,
			Path:      lock.Project.Path,
		},
		State:        LockStateAcquired,
		Priority:     PriorityNormal,
		Owner:        lock.User.Username,
		AcquiredAt:   lock.Time,
		OriginalLock: lock,
	}
}

// Test fixtures
func createTestProject() models.Project {
	return models.Project{
		RepoFullName: "test/repo",
		Path:         ".",
	}
}

func createTestUser() models.User {
	return models.User{Username: "testuser"}
}

func createTestProjectLock() models.ProjectLock {
	return models.ProjectLock{
		Project:   createTestProject(),
		Workspace: "default",
		User:      createTestUser(),
		Time:      time.Now(),
	}
}

// Test cases

func TestCompatibilityLayer_StrictMode(t *testing.T) {
	legacyBackend := NewMockLegacyBackend()
	enhancedBackend := NewMockEnhancedBackend()
	config := DefaultConfig()
	log := logging.NewNoopLogger(t)

	cl := NewCompatibilityLayer(CompatibilityModeStrict, enhancedBackend, legacyBackend, config, log)

	testLock := createTestProjectLock()

	// Test TryLock in strict mode - should only use legacy
	acquired, _, err := cl.TryLock(testLock)
	require.NoError(t, err)
	assert.True(t, acquired)

	// Verify only legacy backend was called
	assert.Contains(t, legacyBackend.calls, "TryLock")
	assert.NotContains(t, enhancedBackend.calls, "TryAcquireLock")

	// Test Unlock in strict mode
	_, err = cl.Unlock(testLock.Project, testLock.Workspace)
	require.NoError(t, err)

	// Verify only legacy backend was called
	assert.Contains(t, legacyBackend.calls, "Unlock")
	assert.NotContains(t, enhancedBackend.calls, "ReleaseLock")
}

func TestCompatibilityLayer_NativeMode(t *testing.T) {
	legacyBackend := NewMockLegacyBackend()
	enhancedBackend := NewMockEnhancedBackend()
	config := DefaultConfig()
	log := logging.NewNoopLogger(t)

	cl := NewCompatibilityLayer(CompatibilityModeNative, enhancedBackend, legacyBackend, config, log)

	testLock := createTestProjectLock()

	// Test TryLock in native mode - should only use enhanced
	acquired, _, err := cl.TryLock(testLock)
	require.NoError(t, err)
	assert.True(t, acquired)

	// Verify only enhanced backend was called
	assert.Contains(t, enhancedBackend.calls, "TryAcquireLock")
	assert.NotContains(t, legacyBackend.calls, "TryLock")

	// Test Unlock in native mode
	_, err = cl.Unlock(testLock.Project, testLock.Workspace)
	require.NoError(t, err)

	// Verify only enhanced backend was called
	assert.Contains(t, enhancedBackend.calls, "ListLocks")
	assert.Contains(t, enhancedBackend.calls, "ReleaseLock")
	assert.NotContains(t, legacyBackend.calls, "Unlock")
}

func TestCompatibilityLayer_HybridMode_SuccessPath(t *testing.T) {
	legacyBackend := NewMockLegacyBackend()
	enhancedBackend := NewMockEnhancedBackend()
	config := DefaultConfig()
	log := logging.NewNoopLogger(t)

	cl := NewCompatibilityLayer(CompatibilityModeHybrid, enhancedBackend, legacyBackend, config, log)

	testLock := createTestProjectLock()

	// Test TryLock in hybrid mode - should try enhanced first
	acquired, _, err := cl.TryLock(testLock)
	require.NoError(t, err)
	assert.True(t, acquired)

	// Verify enhanced backend was called, not legacy
	assert.Contains(t, enhancedBackend.calls, "TryAcquireLock")
	assert.NotContains(t, legacyBackend.calls, "TryLock")
}

func TestCompatibilityLayer_HybridMode_FallbackPath(t *testing.T) {
	legacyBackend := NewMockLegacyBackend()
	enhancedBackend := NewMockEnhancedBackend()
	enhancedBackend.shouldFail = true // Force enhanced to fail
	config := DefaultConfig()
	log := logging.NewNoopLogger(t)

	cl := NewCompatibilityLayer(CompatibilityModeHybrid, enhancedBackend, legacyBackend, config, log)

	testLock := createTestProjectLock()

	// Test TryLock in hybrid mode with enhanced failure - should fallback to legacy
	acquired, _, err := cl.TryLock(testLock)
	require.NoError(t, err)
	assert.True(t, acquired)

	// Verify both backends were called (enhanced first, then fallback to legacy)
	assert.Contains(t, enhancedBackend.calls, "TryAcquireLock")
	assert.Contains(t, legacyBackend.calls, "TryLock")
}

func TestCompatibilityLayer_List(t *testing.T) {
	legacyBackend := NewMockLegacyBackend()
	enhancedBackend := NewMockEnhancedBackend()
	config := DefaultConfig()
	log := logging.NewNoopLogger(t)

	cl := NewCompatibilityLayer(CompatibilityModeHybrid, enhancedBackend, legacyBackend, config, log)

	testLock := createTestProjectLock()

	// Add a lock to enhanced backend
	ctx := context.Background()
	request := cl.convertToEnhancedRequest(testLock)
	enhancedLock, err := enhancedBackend.AcquireLock(ctx, request)
	require.NoError(t, err)

	// Add a lock to legacy backend
	testLock2 := createTestProjectLock()
	testLock2.Project.Path = "path2"
	legacyBackend.TryLock(testLock2)

	// Test List - should merge results from both backends
	locks, err := cl.List()
	require.NoError(t, err)

	// Should contain locks from both backends
	assert.Len(t, locks, 2)

	// Verify both backends were called
	assert.Contains(t, enhancedBackend.calls, "ListLocks")
	assert.Contains(t, legacyBackend.calls, "List")

	// Verify the enhanced lock was converted properly
	foundEnhancedLock := false
	for _, lock := range locks {
		if lock.Project.RepoFullName == enhancedLock.Resource.Namespace {
			foundEnhancedLock = true
			break
		}
	}
	assert.True(t, foundEnhancedLock, "Enhanced lock should be present in list")
}

func TestCompatibilityLayer_UnlockByPull(t *testing.T) {
	legacyBackend := NewMockLegacyBackend()
	enhancedBackend := NewMockEnhancedBackend()
	config := DefaultConfig()
	log := logging.NewNoopLogger(t)

	cl := NewCompatibilityLayer(CompatibilityModeHybrid, enhancedBackend, legacyBackend, config, log)

	repoName := "test/repo"
	pullNum := 123

	// Add locks to both backends
	testLock1 := createTestProjectLock()
	testLock1.Project.RepoFullName = repoName
	testLock1.Workspace = "ws1"

	testLock2 := createTestProjectLock()
	testLock2.Project.RepoFullName = repoName
	testLock2.Workspace = "ws2"

	// Add to enhanced backend
	ctx := context.Background()
	request1 := cl.convertToEnhancedRequest(testLock1)
	enhancedBackend.AcquireLock(ctx, request1)

	// Add to legacy backend
	legacyBackend.TryLock(testLock2)

	// Test UnlockByPull
	unlockedLocks, err := cl.UnlockByPull(repoName, pullNum)
	require.NoError(t, err)

	// Should have unlocked from both backends
	assert.Len(t, unlockedLocks, 2)

	// Verify both backends were called
	assert.Contains(t, enhancedBackend.calls, "ListLocks")
	assert.Contains(t, enhancedBackend.calls, "ReleaseLock")
	assert.Contains(t, legacyBackend.calls, "UnlockByPull")
}

func TestCompatibilityLayer_GetLock(t *testing.T) {
	legacyBackend := NewMockLegacyBackend()
	enhancedBackend := NewMockEnhancedBackend()
	config := DefaultConfig()
	log := logging.NewNoopLogger(t)

	cl := NewCompatibilityLayer(CompatibilityModeHybrid, enhancedBackend, legacyBackend, config, log)

	testLock := createTestProjectLock()

	// Add lock to enhanced backend
	ctx := context.Background()
	request := cl.convertToEnhancedRequest(testLock)
	enhancedBackend.AcquireLock(ctx, request)

	// Test GetLock
	retrievedLock, err := cl.GetLock(testLock.Project, testLock.Workspace)
	require.NoError(t, err)
	assert.NotNil(t, retrievedLock)
	assert.Equal(t, testLock.Project.RepoFullName, retrievedLock.Project.RepoFullName)
	assert.Equal(t, testLock.Workspace, retrievedLock.Workspace)

	// Verify enhanced backend was called
	assert.Contains(t, enhancedBackend.calls, "GetLegacyLock")
}

func TestCompatibilityLayer_Migration(t *testing.T) {
	legacyBackend := NewMockLegacyBackend()
	enhancedBackend := NewMockEnhancedBackend()
	config := DefaultConfig()
	log := logging.NewNoopLogger(t)

	cl := NewCompatibilityLayer(CompatibilityModeHybrid, enhancedBackend, legacyBackend, config, log)

	// Add locks to legacy backend
	testLock1 := createTestProjectLock()
	testLock1.Workspace = "ws1"
	testLock2 := createTestProjectLock()
	testLock2.Workspace = "ws2"

	legacyBackend.TryLock(testLock1)
	legacyBackend.TryLock(testLock2)

	// Start migration
	ctx := context.Background()
	err := cl.StartMigration(ctx)
	require.NoError(t, err)

	// Check migration status
	status := cl.GetMigrationStatus()
	assert.Equal(t, MigrationPhaseCompleted, status.Phase)
	assert.Equal(t, 2, status.LegacyLocksCount)
	assert.Equal(t, 2, status.MigratedCount)

	// Verify locks were migrated to enhanced backend
	enhancedLocks, err := enhancedBackend.ListLocks(ctx)
	require.NoError(t, err)
	assert.Len(t, enhancedLocks, 2)
}

func TestCompatibilityLayer_PerformanceStats(t *testing.T) {
	legacyBackend := NewMockLegacyBackend()
	enhancedBackend := NewMockEnhancedBackend()
	config := DefaultConfig()
	log := logging.NewNoopLogger(t)

	cl := NewCompatibilityLayer(CompatibilityModeHybrid, enhancedBackend, legacyBackend, config, log)

	testLock := createTestProjectLock()

	// Perform some operations
	cl.TryLock(testLock)
	cl.List()
	cl.GetLock(testLock.Project, testLock.Workspace)

	// Get performance stats
	stats := cl.GetPerformanceStats()
	assert.Equal(t, CompatibilityModeHybrid, stats.Mode)
	assert.Greater(t, stats.EnhancedOps, int64(0))
}

func TestCompatibilityLayer_HealthCheck(t *testing.T) {
	legacyBackend := NewMockLegacyBackend()
	enhancedBackend := NewMockEnhancedBackend()
	config := DefaultConfig()
	log := logging.NewNoopLogger(t)

	cl := NewCompatibilityLayer(CompatibilityModeHybrid, enhancedBackend, legacyBackend, config, log)

	// Test healthy state
	ctx := context.Background()
	err := cl.HealthCheck(ctx)
	assert.NoError(t, err)

	// Test unhealthy enhanced backend
	enhancedBackend.healthy = false
	err = cl.HealthCheck(ctx)
	assert.Error(t, err)
}

func TestCompatibilityLayer_RuntimeModeSwitch(t *testing.T) {
	legacyBackend := NewMockLegacyBackend()
	enhancedBackend := NewMockEnhancedBackend()
	config := DefaultConfig()
	log := logging.NewNoopLogger(t)

	cl := NewCompatibilityLayer(CompatibilityModeStrict, enhancedBackend, legacyBackend, config, log)

	// Initially in strict mode
	assert.Equal(t, CompatibilityModeStrict, cl.GetCompatibilityMode())

	// Switch to hybrid mode
	err := cl.SetCompatibilityMode(CompatibilityModeHybrid)
	require.NoError(t, err)
	assert.Equal(t, CompatibilityModeHybrid, cl.GetCompatibilityMode())

	// Switch to native mode
	err = cl.SetCompatibilityMode(CompatibilityModeNative)
	require.NoError(t, err)
	assert.Equal(t, CompatibilityModeNative, cl.GetCompatibilityMode())
}

func TestCompatibilityChecker_VerifyBackwardCompatibility(t *testing.T) {
	legacyBackend := NewMockLegacyBackend()
	enhancedBackend := NewMockEnhancedBackend()
	config := DefaultConfig()
	log := logging.NewNoopLogger(t)

	adapter := NewLockingAdapter(nil, enhancedBackend, config, legacyBackend, log)
	checker := NewCompatibilityChecker(adapter, log)

	ctx := context.Background()
	err := checker.VerifyBackwardCompatibility(ctx)
	assert.NoError(t, err)
}

func TestCompatibilityChecker_RunCompatibilityTest(t *testing.T) {
	legacyBackend := NewMockLegacyBackend()
	enhancedBackend := NewMockEnhancedBackend()
	config := DefaultConfig()
	log := logging.NewNoopLogger(t)

	adapter := NewLockingAdapter(nil, enhancedBackend, config, legacyBackend, log)
	checker := NewCompatibilityChecker(adapter, log)

	ctx := context.Background()
	report, err := checker.RunCompatibilityTest(ctx)
	require.NoError(t, err)
	assert.True(t, report.Success)
	assert.Len(t, report.Tests, 4) // BasicLockUnlock, ConcurrentAccess, UnlockByPull, ListConsistency

	// Verify all tests passed
	for _, test := range report.Tests {
		assert.True(t, test.Success, "Test %s should pass", test.Name)
	}
}

func TestFallbackSystem_CircuitBreaker(t *testing.T) {
	legacyBackend := NewMockLegacyBackend()
	enhancedBackend := NewMockEnhancedBackend()
	config := DefaultConfig()
	log := logging.NewNoopLogger(t)

	fs := NewFallbackSystem(enhancedBackend, legacyBackend, config, log)

	testLock := createTestProjectLock()

	// Initially circuit should be closed
	assert.Equal(t, CircuitStateClosed, fs.circuitBreaker.GetState())

	// Cause failures to open circuit
	enhancedBackend.shouldFail = true
	for i := 0; i < 6; i++ { // Threshold is 5
		fs.TryLock(testLock)
	}

	// Circuit should now be open
	assert.Equal(t, CircuitStateOpen, fs.circuitBreaker.GetState())

	// Reset enhanced backend
	enhancedBackend.shouldFail = false

	// Next call should still use fallback due to open circuit
	acquired, _, err := fs.TryLock(testLock)
	require.NoError(t, err)
	assert.True(t, acquired)

	// Should have used legacy backend due to open circuit
	assert.Contains(t, legacyBackend.calls, "TryLock")
}

func TestFallbackSystem_HealthCheck(t *testing.T) {
	legacyBackend := NewMockLegacyBackend()
	enhancedBackend := NewMockEnhancedBackend()
	config := DefaultConfig()
	log := logging.NewNoopLogger(t)

	fs := NewFallbackSystem(enhancedBackend, legacyBackend, config, log)

	ctx := context.Background()

	// Test healthy state
	health := fs.HealthCheck(ctx)
	assert.True(t, health.Overall)
	assert.True(t, health.Enhanced.Healthy)
	assert.True(t, health.Legacy.Healthy)

	// Test unhealthy enhanced backend
	enhancedBackend.healthy = false
	health = fs.HealthCheck(ctx)
	assert.True(t, health.Overall) // Still overall healthy due to legacy
	assert.False(t, health.Enhanced.Healthy)
	assert.True(t, health.Legacy.Healthy)
}

func TestFallbackSystem_Stats(t *testing.T) {
	legacyBackend := NewMockLegacyBackend()
	enhancedBackend := NewMockEnhancedBackend()
	config := DefaultConfig()
	log := logging.NewNoopLogger(t)

	fs := NewFallbackSystem(enhancedBackend, legacyBackend, config, log)

	testLock := createTestProjectLock()

	// Perform some operations
	fs.TryLock(testLock)
	fs.List()

	// Get stats
	stats := fs.GetStats()
	assert.Greater(t, stats.EnhancedAttempts, int64(0))
	assert.Equal(t, int64(0), stats.FallbackAttempts) // Should not fallback in normal case
}

func TestFallbackSystem_IsReadyForMigration(t *testing.T) {
	legacyBackend := NewMockLegacyBackend()
	enhancedBackend := NewMockEnhancedBackend()
	config := DefaultConfig()
	log := logging.NewNoopLogger(t)

	fs := NewFallbackSystem(enhancedBackend, legacyBackend, config, log)

	// Test ready state
	ready, reason := fs.IsReadyForMigration()
	assert.True(t, ready)
	assert.Equal(t, "System ready for migration", reason)

	// Test not ready state (unhealthy enhanced backend)
	enhancedBackend.healthy = false
	ready, reason = fs.IsReadyForMigration()
	assert.False(t, ready)
	assert.Contains(t, reason, "Enhanced backend is not healthy")
}

// Benchmark tests to ensure performance is maintained

func BenchmarkCompatibilityLayer_StrictMode_TryLock(b *testing.B) {
	legacyBackend := NewMockLegacyBackend()
	enhancedBackend := NewMockEnhancedBackend()
	config := DefaultConfig()
	log := logging.NewNoopLogger(nil)

	cl := NewCompatibilityLayer(CompatibilityModeStrict, enhancedBackend, legacyBackend, config, log)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		testLock := createTestProjectLock()
		testLock.Workspace = "ws" + string(rune(i%1000)) // Avoid conflicts
		cl.TryLock(testLock)
	}
}

func BenchmarkCompatibilityLayer_HybridMode_TryLock(b *testing.B) {
	legacyBackend := NewMockLegacyBackend()
	enhancedBackend := NewMockEnhancedBackend()
	config := DefaultConfig()
	log := logging.NewNoopLogger(nil)

	cl := NewCompatibilityLayer(CompatibilityModeHybrid, enhancedBackend, legacyBackend, config, log)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		testLock := createTestProjectLock()
		testLock.Workspace = "ws" + string(rune(i%1000)) // Avoid conflicts
		cl.TryLock(testLock)
	}
}

func BenchmarkCompatibilityLayer_NativeMode_TryLock(b *testing.B) {
	legacyBackend := NewMockLegacyBackend()
	enhancedBackend := NewMockEnhancedBackend()
	config := DefaultConfig()
	log := logging.NewNoopLogger(nil)

	cl := NewCompatibilityLayer(CompatibilityModeNative, enhancedBackend, legacyBackend, config, log)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		testLock := createTestProjectLock()
		testLock.Workspace = "ws" + string(rune(i%1000)) // Avoid conflicts
		cl.TryLock(testLock)
	}
}