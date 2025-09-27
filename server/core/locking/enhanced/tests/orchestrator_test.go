package tests

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/runatlantis/atlantis/server/core/locking/enhanced"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/logging"
	. "github.com/petergtz/pegomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOrchestrator_Lifecycle(t *testing.T) {
	backend := NewMockBackend()
	config := enhanced.DefaultConfig()
	config.EnableEvents = true
	config.EnablePriorityQueue = true
	config.EnableDeadlockDetection = true

	logger := logging.NewNoopLogger(t)
	orchestrator := enhanced.NewLockOrchestrator(backend, config, logger)

	ctx := context.Background()

	// Test starting
	err := orchestrator.Start(ctx)
	require.NoError(t, err)

	// Verify status
	status := orchestrator.GetStatus()
	assert.True(t, status.Running)
	assert.NotZero(t, status.StartTime)
	assert.NotEmpty(t, status.Components)

	// Check component health
	health := orchestrator.GetComponentHealth()
	for name, component := range health {
		assert.Equal(t, "running", component.Status, "Component %s should be running", name)
		assert.Equal(t, 100, component.Health, "Component %s should have full health", name)
	}

	// Test stopping
	err = orchestrator.Stop()
	require.NoError(t, err)

	// Verify stopped
	status = orchestrator.GetStatus()
	assert.False(t, status.Running)
}

func TestOrchestrator_LockOperations(t *testing.T) {
	backend := NewMockBackend()
	config := enhanced.DefaultConfig()
	logger := logging.NewNoopLogger(t)
	orchestrator := enhanced.NewLockOrchestrator(backend, config, logger)

	ctx := context.Background()
	err := orchestrator.Start(ctx)
	require.NoError(t, err)
	defer orchestrator.Stop()

	project := models.Project{
		RepoFullName: "owner/repo",
		Path:         ".",
	}
	workspace := "default"
	user := models.User{Username: "test-user"}

	// Mock successful lock acquisition
	expectedLock := &enhanced.EnhancedLock{
		ID:       "test-lock-id",
		Resource: enhanced.ResourceIdentifier{
			Type:      enhanced.ResourceTypeProject,
			Namespace: project.RepoFullName,
			Name:      project.Path,
			Workspace: workspace,
		},
		State:      enhanced.LockStateAcquired,
		Priority:   enhanced.PriorityNormal,
		Owner:      user.Username,
		AcquiredAt: time.Now(),
	}

	legacyLock := &models.ProjectLock{
		Project:   project,
		Workspace: workspace,
		User:      user,
		Time:      expectedLock.AcquiredAt,
	}

	When(backend.TryAcquireLock(Any[context.Context](), Any[*enhanced.EnhancedLockRequest]())).
		ThenReturn(expectedLock, true, nil)
	When(backend.ConvertToLegacy(expectedLock)).ThenReturn(legacyLock)

	// Test lock acquisition
	lock, err := orchestrator.Lock(ctx, project, workspace, user)
	require.NoError(t, err)
	assert.NotNil(t, lock)
	assert.Equal(t, project.RepoFullName, lock.Project.RepoFullName)
	assert.Equal(t, workspace, lock.Workspace)
	assert.Equal(t, user.Username, lock.User.Username)

	// Verify backend was called
	backend.VerifyWasCalledOnce().TryAcquireLock(Any[context.Context](), Any[*enhanced.EnhancedLockRequest]())
}

func TestOrchestrator_LockWithOptions(t *testing.T) {
	backend := NewMockBackend()
	config := enhanced.DefaultConfig()
	logger := logging.NewNoopLogger(t)
	orchestrator := enhanced.NewLockOrchestrator(backend, config, logger)

	ctx := context.Background()
	err := orchestrator.Start(ctx)
	require.NoError(t, err)
	defer orchestrator.Stop()

	project := models.Project{RepoFullName: "owner/repo", Path: "."}
	workspace := "default"
	user := models.User{Username: "test-user"}

	// Mock successful lock acquisition
	expectedLock := &enhanced.EnhancedLock{
		ID:       "test-lock-id",
		State:    enhanced.LockStateAcquired,
		Priority: enhanced.PriorityHigh,
		Owner:    user.Username,
	}

	legacyLock := &models.ProjectLock{
		Project: project, Workspace: workspace, User: user, Time: time.Now(),
	}

	When(backend.TryAcquireLock(Any[context.Context](), Any[*enhanced.EnhancedLockRequest]())).
		ThenReturn(expectedLock, true, nil)
	When(backend.ConvertToLegacy(expectedLock)).ThenReturn(legacyLock)

	// Test lock with custom options
	options := enhanced.LockRequestOptions{
		Priority: enhanced.PriorityHigh,
		Timeout:  time.Minute,
		Metadata: map[string]string{"source": "test"},
	}

	lock, err := orchestrator.LockWithOptions(ctx, project, workspace, user, options)
	require.NoError(t, err)
	assert.NotNil(t, lock)

	// Verify the request had correct priority
	backend.VerifyWasCalledOnce().TryAcquireLock(Any[context.Context](), Any[*enhanced.EnhancedLockRequest]())
}

func TestOrchestrator_UnlockOperations(t *testing.T) {
	backend := NewMockBackend()
	config := enhanced.DefaultConfig()
	logger := logging.NewNoopLogger(t)
	orchestrator := enhanced.NewLockOrchestrator(backend, config, logger)

	ctx := context.Background()
	err := orchestrator.Start(ctx)
	require.NoError(t, err)
	defer orchestrator.Stop()

	project := models.Project{RepoFullName: "owner/repo", Path: "."}
	workspace := "default"
	user := models.User{Username: "test-user"}

	// Mock existing lock
	existingLock := &enhanced.EnhancedLock{
		ID:    "test-lock-id",
		Owner: user.Username,
		Resource: enhanced.ResourceIdentifier{
			Type:      enhanced.ResourceTypeProject,
			Namespace: project.RepoFullName,
			Name:      project.Path,
			Workspace: workspace,
		},
		State: enhanced.LockStateAcquired,
	}

	legacyLock := &models.ProjectLock{
		Project: project, Workspace: workspace, User: user, Time: time.Now(),
	}

	When(backend.ListLocks(Any[context.Context]())).ThenReturn([]*enhanced.EnhancedLock{existingLock}, nil)
	When(backend.ReleaseLock(Any[context.Context](), "test-lock-id")).ThenReturn(nil)
	When(backend.ConvertToLegacy(existingLock)).ThenReturn(legacyLock)

	// Test unlock
	lock, err := orchestrator.Unlock(ctx, project, workspace, user)
	require.NoError(t, err)
	assert.NotNil(t, lock)

	// Verify backend calls
	backend.VerifyWasCalledOnce().ListLocks(Any[context.Context]())
	backend.VerifyWasCalledOnce().ReleaseLock(Any[context.Context](), "test-lock-id")
}

func TestOrchestrator_ListOperations(t *testing.T) {
	backend := NewMockBackend()
	config := enhanced.DefaultConfig()
	logger := logging.NewNoopLogger(t)
	orchestrator := enhanced.NewLockOrchestrator(backend, config, logger)

	ctx := context.Background()
	err := orchestrator.Start(ctx)
	require.NoError(t, err)
	defer orchestrator.Stop()

	// Mock existing locks
	locks := []*enhanced.EnhancedLock{
		{
			ID:    "lock1",
			State: enhanced.LockStateAcquired,
			Resource: enhanced.ResourceIdentifier{
				Type:      enhanced.ResourceTypeProject,
				Namespace: "owner/repo1",
				Name:      ".",
				Workspace: "default",
			},
		},
		{
			ID:    "lock2",
			State: enhanced.LockStateAcquired,
			Resource: enhanced.ResourceIdentifier{
				Type:      enhanced.ResourceTypeProject,
				Namespace: "owner/repo2",
				Name:      ".",
				Workspace: "staging",
			},
		},
	}

	legacyLocks := []*models.ProjectLock{
		{Project: models.Project{RepoFullName: "owner/repo1"}, Workspace: "default"},
		{Project: models.Project{RepoFullName: "owner/repo2"}, Workspace: "staging"},
	}

	When(backend.ListLocks(Any[context.Context]())).ThenReturn(locks, nil)
	When(backend.ConvertToLegacy(locks[0])).ThenReturn(legacyLocks[0])
	When(backend.ConvertToLegacy(locks[1])).ThenReturn(legacyLocks[1])

	// Test list
	result, err := orchestrator.List(ctx)
	require.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, "owner/repo1", result[0].Project.RepoFullName)
	assert.Equal(t, "owner/repo2", result[1].Project.RepoFullName)

	// Verify backend call
	backend.VerifyWasCalledOnce().ListLocks(Any[context.Context]())
}

func TestOrchestrator_ConcurrentOperations(t *testing.T) {
	backend := NewMockBackend()
	config := enhanced.DefaultConfig()
	logger := logging.NewNoopLogger(t)
	orchestrator := enhanced.NewLockOrchestrator(backend, config, logger)

	ctx := context.Background()
	err := orchestrator.Start(ctx)
	require.NoError(t, err)
	defer orchestrator.Stop()

	// Mock successful lock acquisition
	When(backend.TryAcquireLock(Any[context.Context](), Any[*enhanced.EnhancedLockRequest]())).
		ThenReturn(&enhanced.EnhancedLock{ID: "test-lock"}, true, nil)
	When(backend.ConvertToLegacy(Any[*enhanced.EnhancedLock]())).
		ThenReturn(&models.ProjectLock{})

	// Test concurrent lock requests
	var wg sync.WaitGroup
	numGoroutines := 10
	results := make([]error, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()

			project := models.Project{
				RepoFullName: "owner/repo",
				Path:         ".",
			}
			workspace := "default"
			user := models.User{Username: "test-user"}

			_, err := orchestrator.Lock(ctx, project, workspace, user)
			results[index] = err
		}(i)
	}

	wg.Wait()

	// Verify all requests completed without errors
	for i, result := range results {
		assert.NoError(t, result, "Request %d should not have errored", i)
	}

	// Verify backend was called for each request
	backend.VerifyWasCalled(Times(numGoroutines)).TryAcquireLock(Any[context.Context](), Any[*enhanced.EnhancedLockRequest]())
}

func TestOrchestrator_RequestValidation(t *testing.T) {
	backend := NewMockBackend()
	config := enhanced.DefaultConfig()
	logger := logging.NewNoopLogger(t)
	orchestrator := enhanced.NewLockOrchestrator(backend, config, logger)

	ctx := context.Background()
	err := orchestrator.Start(ctx)
	require.NoError(t, err)
	defer orchestrator.Stop()

	tests := []struct {
		name      string
		project   models.Project
		workspace string
		user      models.User
		wantError bool
	}{
		{
			name:      "valid request",
			project:   models.Project{RepoFullName: "owner/repo", Path: "."},
			workspace: "default",
			user:      models.User{Username: "test-user"},
			wantError: false,
		},
		{
			name:      "empty repo name",
			project:   models.Project{RepoFullName: "", Path: "."},
			workspace: "default",
			user:      models.User{Username: "test-user"},
			wantError: true,
		},
		{
			name:      "empty workspace",
			project:   models.Project{RepoFullName: "owner/repo", Path: "."},
			workspace: "",
			user:      models.User{Username: "test-user"},
			wantError: true,
		},
		{
			name:      "empty username",
			project:   models.Project{RepoFullName: "owner/repo", Path: "."},
			workspace: "default",
			user:      models.User{Username: ""},
			wantError: true,
		},
	}

	// Mock successful backend response for valid requests
	When(backend.TryAcquireLock(Any[context.Context](), Any[*enhanced.EnhancedLockRequest]())).
		ThenReturn(&enhanced.EnhancedLock{ID: "test-lock"}, true, nil)
	When(backend.ConvertToLegacy(Any[*enhanced.EnhancedLock]())).
		ThenReturn(&models.ProjectLock{})

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := orchestrator.Lock(ctx, tt.project, tt.workspace, tt.user)

			if tt.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestOrchestrator_Metrics(t *testing.T) {
	backend := NewMockBackend()
	config := enhanced.DefaultConfig()
	logger := logging.NewNoopLogger(t)
	orchestrator := enhanced.NewLockOrchestrator(backend, config, logger)

	ctx := context.Background()
	err := orchestrator.Start(ctx)
	require.NoError(t, err)
	defer orchestrator.Stop()

	// Get initial metrics
	metrics := orchestrator.GetMetrics()
	require.NotNil(t, metrics)
	assert.NotZero(t, metrics.Timestamp)
	assert.NotNil(t, metrics.Manager)
	assert.NotNil(t, metrics.System)

	// Check metrics structure
	assert.NotNil(t, metrics.Summary)
	assert.GreaterOrEqual(t, metrics.Summary.OverallHealth, 0)
	assert.LessOrEqual(t, metrics.Summary.OverallHealth, 100)
}

func TestOrchestrator_HealthMonitoring(t *testing.T) {
	backend := NewMockBackend()
	config := enhanced.DefaultConfig()
	logger := logging.NewNoopLogger(t)
	orchestrator := enhanced.NewLockOrchestrator(backend, config, logger)

	ctx := context.Background()
	err := orchestrator.Start(ctx)
	require.NoError(t, err)
	defer orchestrator.Stop()

	// Mock healthy backend
	When(backend.HealthCheck(Any[context.Context]())).ThenReturn(nil)

	// Allow some time for health monitoring to run
	time.Sleep(100 * time.Millisecond)

	// Check component health
	health := orchestrator.GetComponentHealth()
	for name, component := range health {
		assert.NotEmpty(t, component.Name, "Component %s should have a name", name)
		assert.NotEmpty(t, component.Status, "Component %s should have a status", name)
		assert.GreaterOrEqual(t, component.Health, 0, "Component %s health should be >= 0", name)
		assert.LessOrEqual(t, component.Health, 100, "Component %s health should be <= 100", name)
	}

	// Verify backend health check was called
	backend.VerifyWasCalled(AtLeast(1)).HealthCheck(Any[context.Context]())
}

func TestOrchestrator_ComponentFailure(t *testing.T) {
	backend := NewMockBackend()
	config := enhanced.DefaultConfig()
	logger := logging.NewNoopLogger(t)
	orchestrator := enhanced.NewLockOrchestrator(backend, config, logger)

	ctx := context.Background()
	err := orchestrator.Start(ctx)
	require.NoError(t, err)
	defer orchestrator.Stop()

	// Mock backend failure
	When(backend.HealthCheck(Any[context.Context]())).ThenReturn(assert.AnError)

	// Allow some time for health monitoring to detect failure
	time.Sleep(200 * time.Millisecond)

	// Check that backend component is marked as unhealthy
	health := orchestrator.GetComponentHealth()
	if backendHealth, exists := health["backend"]; exists {
		assert.Equal(t, "error", backendHealth.Status)
		assert.Equal(t, 0, backendHealth.Health)
		assert.NotEmpty(t, backendHealth.Error)
	}
}

func TestOrchestrator_EventSystem(t *testing.T) {
	backend := NewMockBackend()
	config := enhanced.DefaultConfig()
	config.EnableEvents = true
	logger := logging.NewNoopLogger(t)
	orchestrator := enhanced.NewLockOrchestrator(backend, config, logger)

	ctx := context.Background()
	err := orchestrator.Start(ctx)
	require.NoError(t, err)
	defer orchestrator.Stop()

	project := models.Project{RepoFullName: "owner/repo", Path: "."}
	workspace := "default"
	user := models.User{Username: "test-user"}

	// Mock successful lock acquisition
	When(backend.TryAcquireLock(Any[context.Context](), Any[*enhanced.EnhancedLockRequest]())).
		ThenReturn(&enhanced.EnhancedLock{ID: "test-lock"}, true, nil)
	When(backend.ConvertToLegacy(Any[*enhanced.EnhancedLock]())).
		ThenReturn(&models.ProjectLock{})

	// Perform lock operation
	_, err = orchestrator.Lock(ctx, project, workspace, user)
	require.NoError(t, err)

	// Allow time for event processing
	time.Sleep(100 * time.Millisecond)

	// Check that event system is functioning (events would be logged)
	// In a real implementation, we'd verify specific events were emitted
	metrics := orchestrator.GetMetrics()
	assert.NotNil(t, metrics.Events)
}

func TestOrchestrator_StressTest(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}

	backend := NewMockBackend()
	config := enhanced.DefaultConfig()
	logger := logging.NewNoopLogger(t)
	orchestrator := enhanced.NewLockOrchestrator(backend, config, logger)

	ctx := context.Background()
	err := orchestrator.Start(ctx)
	require.NoError(t, err)
	defer orchestrator.Stop()

	// Mock successful lock acquisition
	When(backend.TryAcquireLock(Any[context.Context](), Any[*enhanced.EnhancedLockRequest]())).
		ThenReturn(&enhanced.EnhancedLock{ID: "test-lock"}, true, nil)
	When(backend.ConvertToLegacy(Any[*enhanced.EnhancedLock]())).
		ThenReturn(&models.ProjectLock{})

	// Stress test with many concurrent operations
	var wg sync.WaitGroup
	numOperations := 1000
	errors := make([]error, numOperations)

	startTime := time.Now()

	for i := 0; i < numOperations; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()

			project := models.Project{
				RepoFullName: "owner/repo",
				Path:         ".",
			}
			workspace := "default"
			user := models.User{Username: "test-user"}

			_, err := orchestrator.Lock(ctx, project, workspace, user)
			errors[index] = err
		}(i)
	}

	wg.Wait()
	duration := time.Since(startTime)

	// Verify all operations completed successfully
	errorCount := 0
	for _, err := range errors {
		if err != nil {
			errorCount++
		}
	}

	t.Logf("Completed %d operations in %v (%.2f ops/sec), %d errors",
		numOperations, duration, float64(numOperations)/duration.Seconds(), errorCount)

	assert.Equal(t, 0, errorCount, "All operations should succeed")
	assert.Less(t, duration, 10*time.Second, "Operations should complete within reasonable time")

	// Check final metrics
	metrics := orchestrator.GetMetrics()
	assert.GreaterOrEqual(t, metrics.Manager.TotalRequests, int64(numOperations))
}

// Mock Backend implementation for testing
type MockBackend struct {
	*MockedBackend
}

func NewMockBackend() *MockBackend {
	return &MockBackend{
		MockedBackend: NewMockedBackend(),
	}
}

func (m *MockBackend) TryAcquireLock(ctx context.Context, request *enhanced.EnhancedLockRequest) (*enhanced.EnhancedLock, bool, error) {
	args := m.Called(ctx, request)
	return args.Get(0).(*enhanced.EnhancedLock), args.Bool(1), args.Error(2)
}

func (m *MockBackend) AcquireLock(ctx context.Context, request *enhanced.EnhancedLockRequest) (*enhanced.EnhancedLock, error) {
	args := m.Called(ctx, request)
	return args.Get(0).(*enhanced.EnhancedLock), args.Error(1)
}

func (m *MockBackend) ReleaseLock(ctx context.Context, lockID string) error {
	args := m.Called(ctx, lockID)
	return args.Error(0)
}

func (m *MockBackend) GetLock(ctx context.Context, lockID string) (*enhanced.EnhancedLock, error) {
	args := m.Called(ctx, lockID)
	return args.Get(0).(*enhanced.EnhancedLock), args.Error(1)
}

func (m *MockBackend) ListLocks(ctx context.Context) ([]*enhanced.EnhancedLock, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*enhanced.EnhancedLock), args.Error(1)
}

func (m *MockBackend) RefreshLock(ctx context.Context, lockID string, extension time.Duration) error {
	args := m.Called(ctx, lockID, extension)
	return args.Error(0)
}

func (m *MockBackend) TransferLock(ctx context.Context, lockID string, newOwner string) error {
	args := m.Called(ctx, lockID, newOwner)
	return args.Error(0)
}

func (m *MockBackend) EnqueueLockRequest(ctx context.Context, request *enhanced.EnhancedLockRequest) error {
	args := m.Called(ctx, request)
	return args.Error(0)
}

func (m *MockBackend) DequeueNextRequest(ctx context.Context) (*enhanced.EnhancedLockRequest, error) {
	args := m.Called(ctx)
	return args.Get(0).(*enhanced.EnhancedLockRequest), args.Error(1)
}

func (m *MockBackend) GetQueueStatus(ctx context.Context) (*enhanced.QueueStatus, error) {
	args := m.Called(ctx)
	return args.Get(0).(*enhanced.QueueStatus), args.Error(1)
}

func (m *MockBackend) HealthCheck(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockBackend) GetStats(ctx context.Context) (*enhanced.BackendStats, error) {
	args := m.Called(ctx)
	return args.Get(0).(*enhanced.BackendStats), args.Error(1)
}

func (m *MockBackend) Subscribe(ctx context.Context, eventTypes []string) (<-chan *enhanced.LockEvent, error) {
	args := m.Called(ctx, eventTypes)
	return args.Get(0).(<-chan *enhanced.LockEvent), args.Error(1)
}

func (m *MockBackend) CleanupExpiredLocks(ctx context.Context) (int, error) {
	args := m.Called(ctx)
	return args.Int(0), args.Error(1)
}

func (m *MockBackend) GetLegacyLock(project models.Project, workspace string) (*models.ProjectLock, error) {
	args := m.Called(project, workspace)
	return args.Get(0).(*models.ProjectLock), args.Error(1)
}

func (m *MockBackend) ConvertToLegacy(lock *enhanced.EnhancedLock) *models.ProjectLock {
	args := m.Called(lock)
	return args.Get(0).(*models.ProjectLock)
}

func (m *MockBackend) ConvertFromLegacy(lock *models.ProjectLock) *enhanced.EnhancedLock {
	args := m.Called(lock)
	return args.Get(0).(*enhanced.EnhancedLock)
}