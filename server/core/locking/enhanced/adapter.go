package enhanced

import (
	"context"
	"fmt"
	"time"

	"github.com/runatlantis/atlantis/server/core/locking"
	"github.com/runatlantis/atlantis/server/core/locking/types"
	"github.com/runatlantis/atlantis/server/events"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/logging"
)

// LockingAdapter provides backward compatibility with existing Atlantis locking interfaces
type LockingAdapter struct {
	manager        LockManager
	backend        Backend
	config         *EnhancedConfig
	log            logging.SimpleLogging
	legacyFallback locking.Backend // Original backend for fallback
}

// NewLockingAdapter creates a new adapter that implements the legacy locking.Backend interface
func NewLockingAdapter(manager LockManager, backend Backend, config *EnhancedConfig, legacyFallback locking.Backend, log logging.SimpleLogging) *LockingAdapter {
	return &LockingAdapter{
		manager:        manager,
		backend:        backend,
		config:         config,
		log:            log,
		legacyFallback: legacyFallback,
	}
}

// TryLock attempts to acquire a lock (implements locking.Backend interface)
func (la *LockingAdapter) TryLock(lock models.ProjectLock) (bool, events.TryLockResponse, error) {
	if !la.config.Enabled {
		// Fall back to legacy backend
		acquired, _, err := la.legacyFallback.TryLock(lock)
		if err != nil {
			return false, events.TryLockResponse{}, err
		}
		return acquired, events.TryLockResponse{
			LockAcquired: acquired,
		}, nil
	}

	ctx := context.Background()

	// Convert legacy lock to enhanced request
	request := la.convertToEnhancedRequest(lock)

	// Try to acquire the lock
	enhancedLock, acquired, err := la.backend.TryAcquireLock(ctx, request)
	if err != nil {
		la.log.Err("Failed to acquire enhanced lock: %v", err)

		// Fall back to legacy if enabled
		if la.config.LegacyFallback {
			la.log.Info("Falling back to legacy locking system")
			acquired, _, err := la.legacyFallback.TryLock(lock)
			if err != nil {
				return false, events.TryLockResponse{}, err
			}
			return acquired, events.TryLockResponse{
				LockAcquired: acquired,
			}, nil
		}

		return false, events.TryLockResponse{}, err
	}

	if !acquired {
		// Lock is held by someone else
		if enhancedLock != nil && enhancedLock.State == "pending" {
			// Request is queued - return empty lock
			return false, events.TryLockResponse{
				LockAcquired: false,
			}, nil
		}

		// Find who holds the lock
		existingLocks, err := la.backend.ListLocks(ctx)
		if err != nil {
			return false, events.TryLockResponse{}, err
		}

		for _, existingLock := range existingLocks {
			if la.sameResource(request.Resource, existingLock.Resource) {
				// Build failure reason message
				failureReason := fmt.Sprintf("This project is currently locked by %s", existingLock.Owner)
				return false, events.TryLockResponse{
					LockAcquired:      false,
					LockFailureReason: failureReason,
				}, nil
			}
		}

		return false, events.TryLockResponse{
			LockAcquired: false,
		}, nil
	}

	// Successfully acquired
	la.log.Info("Enhanced lock acquired: %s for project %s/%s", enhancedLock.ID, lock.Project.RepoFullName, lock.Workspace)
	return true, events.TryLockResponse{
		LockAcquired: true,
	}, nil
}

// Unlock releases a lock (implements locking.Backend interface)
func (la *LockingAdapter) Unlock(project models.Project, workspace string) (*models.ProjectLock, error) {
	if !la.config.Enabled {
		// Fall back to legacy backend
		return la.legacyFallback.Unlock(project, workspace)
	}

	ctx := context.Background()

	// Find the lock to unlock
	locks, err := la.backend.ListLocks(ctx)
	if err != nil {
		la.log.Err("Failed to list locks for unlock: %v", err)

		// Fall back to legacy if enabled
		if la.config.LegacyFallback {
			return la.legacyFallback.Unlock(project, workspace)
		}

		return nil, err
	}

	var targetLock *EnhancedLock
	for _, lock := range locks {
		if lock.Resource.Namespace == project.RepoFullName &&
			lock.Resource.Path == project.Path &&
			lock.Resource.Workspace == workspace &&
			lock.Resource.Type == "project" {
			targetLock = lock
			break
		}
	}

	if targetLock == nil {
		// No lock found - this is expected behavior for some Atlantis operations
		return nil, nil
	}

	// Release the lock
	err = la.backend.ReleaseLock(ctx, targetLock.ID)
	if err != nil {
		la.log.Err("Failed to release enhanced lock: %v", err)

		// Fall back to legacy if enabled
		if la.config.LegacyFallback {
			return la.legacyFallback.Unlock(project, workspace)
		}

		return nil, err
	}

	// Return the legacy format
	legacyLock := la.convertToLegacyLock(targetLock)
	la.log.Info("Enhanced lock released: %s for project %s/%s", targetLock.ID, project.RepoFullName, workspace)
	return legacyLock, nil
}

// List returns all current locks (implements locking.Backend interface)
func (la *LockingAdapter) List() ([]models.ProjectLock, error) {
	if !la.config.Enabled {
		// Fall back to legacy backend
		return la.legacyFallback.List()
	}

	ctx := context.Background()

	// Get enhanced locks
	locks, err := la.backend.ListLocks(ctx)
	if err != nil {
		la.log.Err("Failed to list enhanced locks: %v", err)

		// Fall back to legacy if enabled
		if la.config.LegacyFallback {
			return la.legacyFallback.List()
		}

		return nil, err
	}

	// Convert to legacy format
	var lockList []models.ProjectLock
	for _, lock := range locks {
		if lock.State == "acquired" {
			legacyLock := la.convertToLegacyLock(lock)
			if legacyLock != nil {
				lockList = append(lockList, *legacyLock)
			}
		}
	}

	return lockList, nil
}

// UnlockByPull releases all locks associated with a pull request (implements locking.Backend interface)
func (la *LockingAdapter) UnlockByPull(repoFullName string, pullNum int) ([]models.ProjectLock, error) {
	if !la.config.Enabled {
		// Fall back to legacy backend
		return la.legacyFallback.UnlockByPull(repoFullName, pullNum)
	}

	ctx := context.Background()

	// Get all locks for this repository
	locks, err := la.backend.ListLocks(ctx)
	if err != nil {
		la.log.Err("Failed to list locks for UnlockByPull: %v", err)

		// Fall back to legacy if enabled
		if la.config.LegacyFallback {
			return la.legacyFallback.UnlockByPull(repoFullName, pullNum)
		}

		return nil, err
	}

	var unlockedLocks []models.ProjectLock

	// Find locks that match the repository
	for _, lock := range locks {
		if lock.Resource.Namespace == repoFullName && lock.State == LockStateAcquired {
			// Release the lock
			err := la.backend.ReleaseLock(ctx, lock.ID)
			if err != nil {
				la.log.Warn("Failed to release lock %s during UnlockByPull: %v", lock.ID, err)
				continue
			}

			// Add to unlocked list
			legacyLock := la.backend.ConvertToLegacy(lock)
			if legacyLock != nil {
				unlockedLocks = append(unlockedLocks, *legacyLock)
			}
		}
	}

	la.log.Info("Unlocked %d locks for pull request %s#%d", len(unlockedLocks), repoFullName, pullNum)
	return unlockedLocks, nil
}

// GetLock returns a specific lock (implements locking.Backend interface)
func (la *LockingAdapter) GetLock(project models.Project, workspace string) (*models.ProjectLock, error) {
	if !la.config.Enabled {
		// Fall back to legacy backend
		return la.legacyFallback.GetLock(project, workspace)
	}

	// Use the backend's legacy compatibility method
	return la.backend.GetLegacyLock(project, workspace)
}

// Conversion helpers

func (la *LockingAdapter) convertToEnhancedRequest(lock models.ProjectLock) *types.EnhancedLockRequest {
	return &types.EnhancedLockRequest{
		ID: generateRequestID(),
		Resource: types.ResourceIdentifier{
			Type:      types.ResourceTypeProject,
			Namespace: lock.Project.RepoFullName,
			Name:      lock.Project.Path,
			Workspace: lock.Workspace,
			Path:      lock.Project.Path,
		},
		Priority:    types.PriorityNormal,
		Timeout:     la.config.DefaultTimeout,
		Metadata:    make(map[string]string),
		Context:     context.Background(),
		RequestedAt: time.Now(),
		Project:     lock.Project,
		Workspace:   lock.Workspace,
		User:        lock.User,
	}
}

func (la *LockingAdapter) sameResource(r1, r2 types.ResourceIdentifier) bool {
	return r1.Namespace == r2.Namespace &&
		r1.Name == r2.Name &&
		r1.Workspace == r2.Workspace &&
		r1.Path == r2.Path
}

// convertToLegacyLock converts an enhanced lock to a legacy ProjectLock
func (la *LockingAdapter) convertToLegacyLock(enhancedLock *types.EnhancedLock) *models.ProjectLock {
	if enhancedLock == nil {
		return nil
	}

	return &models.ProjectLock{
		Project: models.Project{
			RepoFullName: enhancedLock.Resource.Namespace,
			Path:         enhancedLock.Resource.Path,
		},
		Workspace: enhancedLock.Resource.Workspace,
		User: models.User{
			Username: enhancedLock.Owner,
		},
		Time: enhancedLock.AcquiredAt,
	}
}

// Enhanced methods (additional capabilities beyond legacy interface)

// LockWithPriority acquires a lock with specific priority
func (la *LockingAdapter) LockWithPriority(ctx context.Context, project models.Project, workspace string, user models.User, priority Priority) (*models.ProjectLock, error) {
	if !la.config.Enabled {
		// Fall back to regular lock
		return la.manager.Lock(ctx, project, workspace, user)
	}

	lock, err := la.manager.LockWithPriority(ctx, project, workspace, user, priority)
	if err != nil && la.config.LegacyFallback {
		// Fall back to legacy
		legacyLock := models.ProjectLock{
			Project:   project,
			Workspace: workspace,
			User:      user,
			Time:      time.Now(),
		}
		acquired, _, legacyErr := la.legacyFallback.TryLock(legacyLock)
		if legacyErr != nil {
			return nil, legacyErr
		}
		if !acquired {
			return nil, NewLockExistsError(fmt.Sprintf("%s/%s", project.RepoFullName, workspace))
		}
		return &legacyLock, nil
	}

	return lock, err
}

// LockWithTimeout acquires a lock with a specific timeout
func (la *LockingAdapter) LockWithTimeout(ctx context.Context, project models.Project, workspace string, user models.User, timeout time.Duration) (*models.ProjectLock, error) {
	if !la.config.Enabled {
		// Fall back to regular lock
		return la.manager.Lock(ctx, project, workspace, user)
	}

	lock, err := la.manager.LockWithTimeout(ctx, project, workspace, user, timeout)
	if err != nil && la.config.LegacyFallback {
		// Fall back to legacy
		legacyLock := models.ProjectLock{
			Project:   project,
			Workspace: workspace,
			User:      user,
			Time:      time.Now(),
		}
		acquired, _, legacyErr := la.legacyFallback.TryLock(legacyLock)
		if legacyErr != nil {
			return nil, legacyErr
		}
		if !acquired {
			return nil, NewLockExistsError(fmt.Sprintf("%s/%s", project.RepoFullName, workspace))
		}
		return &legacyLock, nil
	}

	return lock, err
}

// GetQueuePosition returns the position of a request in the queue
func (la *LockingAdapter) GetQueuePosition(ctx context.Context, project models.Project, workspace string) (int, error) {
	if !la.config.Enabled {
		return -1, fmt.Errorf("enhanced locking is not enabled")
	}

	return la.manager.GetQueuePosition(ctx, project, workspace)
}

// GetEnhancedStats returns enhanced locking statistics
func (la *LockingAdapter) GetEnhancedStats(ctx context.Context) (*BackendStats, error) {
	if !la.config.Enabled {
		return nil, fmt.Errorf("enhanced locking is not enabled")
	}

	return la.backend.GetStats(ctx)
}

// HealthCheck verifies the enhanced locking system is working
func (la *LockingAdapter) HealthCheck(ctx context.Context) error {
	if !la.config.Enabled {
		// Check legacy backend
		if la.legacyFallback != nil {
			// Legacy backends may not implement health checks
			return nil
		}
		return nil
	}

	return la.backend.HealthCheck(ctx)
}

// Configuration and management

// IsEnhancedModeEnabled returns whether enhanced locking is enabled
func (la *LockingAdapter) IsEnhancedModeEnabled() bool {
	return la.config.Enabled
}

// GetConfiguration returns the current enhanced locking configuration
func (la *LockingAdapter) GetConfiguration() *EnhancedConfig {
	return la.config
}

// UpdateConfiguration updates the enhanced locking configuration (runtime changes)
func (la *LockingAdapter) UpdateConfiguration(config *EnhancedConfig) error {
	// Validate configuration
	if err := la.validateConfig(config); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}

	la.config = config
	la.log.Info("Enhanced locking configuration updated")
	return nil
}

func (la *LockingAdapter) validateConfig(config *EnhancedConfig) error {
	if config.DefaultTimeout <= 0 {
		return fmt.Errorf("default timeout must be positive")
	}
	if config.MaxTimeout < config.DefaultTimeout {
		return fmt.Errorf("max timeout must be >= default timeout")
	}
	if config.MaxQueueSize < 0 {
		return fmt.Errorf("max queue size cannot be negative")
	}
	if config.MaxRetryAttempts < 0 {
		return fmt.Errorf("max retry attempts cannot be negative")
	}
	return nil
}

// generateRequestID generates a unique request ID
func generateRequestID() string {
	return fmt.Sprintf("req_%d_%d", time.Now().UnixNano(), time.Now().Nanosecond()%1000)
}

// CompatibilityChecker helps verify enhanced locking compatibility
type CompatibilityChecker struct {
	adapter *LockingAdapter
	log     logging.SimpleLogging
}

// NewCompatibilityChecker creates a new compatibility checker
func NewCompatibilityChecker(adapter *LockingAdapter, log logging.SimpleLogging) *CompatibilityChecker {
	return &CompatibilityChecker{
		adapter: adapter,
		log:     log,
	}
}

// VerifyBackwardCompatibility tests that enhanced locking maintains compatibility
func (cc *CompatibilityChecker) VerifyBackwardCompatibility(ctx context.Context) error {
	cc.log.Info("Verifying enhanced locking backward compatibility...")

	// Test basic lock/unlock cycle
	testProject := models.Project{
		RepoFullName: "test/compatibility-check",
		Path:         ".",
	}
	testUser := models.User{Username: "atlantis-test"}
	testWorkspace := "default"

	// Test TryLock
	acquired, _, err := cc.adapter.TryLock(models.ProjectLock{
		Project:   testProject,
		Workspace: testWorkspace,
		User:      testUser,
		Time:      time.Now(),
	})

	if err != nil {
		return fmt.Errorf("TryLock compatibility test failed: %w", err)
	}

	if !acquired {
		return fmt.Errorf("TryLock compatibility test failed: lock not acquired")
	}

	// Test List
	locks, err := cc.adapter.List()
	if err != nil {
		return fmt.Errorf("List compatibility test failed: %w", err)
	}

	found := false
	for key, lock := range locks {
		if lock.Project.RepoFullName == testProject.RepoFullName &&
			lock.Workspace == testWorkspace {
			found = true
			cc.log.Info("Found test lock in list: %s", key)
			break
		}
	}

	if !found {
		return fmt.Errorf("List compatibility test failed: test lock not found in results")
	}

	// Test GetLock
	retrievedLock, err := cc.adapter.GetLock(testProject, testWorkspace)
	if err != nil {
		return fmt.Errorf("GetLock compatibility test failed: %w", err)
	}

	if retrievedLock == nil {
		return fmt.Errorf("GetLock compatibility test failed: lock not found")
	}

	// Test Unlock
	unlockedLock, err := cc.adapter.Unlock(testProject, testWorkspace)
	if err != nil {
		return fmt.Errorf("Unlock compatibility test failed: %w", err)
	}

	if unlockedLock == nil {
		return fmt.Errorf("Unlock compatibility test failed: no lock returned")
	}

	// Verify lock is gone
	finalLock, err := cc.adapter.GetLock(testProject, testWorkspace)
	if err != nil {
		return fmt.Errorf("Final GetLock compatibility test failed: %w", err)
	}

	if finalLock != nil {
		return fmt.Errorf("Final GetLock compatibility test failed: lock still exists after unlock")
	}

	cc.log.Info("Enhanced locking backward compatibility verified successfully")
	return nil
}

// RunCompatibilityTest runs a comprehensive compatibility test
func (cc *CompatibilityChecker) RunCompatibilityTest(ctx context.Context) (*CompatibilityReport, error) {
	report := &CompatibilityReport{
		StartTime: time.Now(),
		Tests:     make([]CompatibilityTest, 0),
	}

	// Test enhanced vs legacy behavior
	tests := []struct {
		name string
		test func(context.Context) error
	}{
		{"BasicLockUnlock", cc.testBasicLockUnlock},
		{"ConcurrentAccess", cc.testConcurrentAccess},
		{"UnlockByPull", cc.testUnlockByPull},
		{"ListConsistency", cc.testListConsistency},
	}

	for _, test := range tests {
		testResult := CompatibilityTest{
			Name:      test.name,
			StartTime: time.Now(),
		}

		err := test.test(ctx)
		testResult.EndTime = time.Now()
		testResult.Duration = testResult.EndTime.Sub(testResult.StartTime)

		if err != nil {
			testResult.Success = false
			testResult.Error = err.Error()
		} else {
			testResult.Success = true
		}

		report.Tests = append(report.Tests, testResult)
		cc.log.Info("Compatibility test %s: %v (took %v)", test.name, testResult.Success, testResult.Duration)
	}

	report.EndTime = time.Now()
	report.Duration = report.EndTime.Sub(report.StartTime)

	// Calculate overall success
	report.Success = true
	for _, test := range report.Tests {
		if !test.Success {
			report.Success = false
			break
		}
	}

	return report, nil
}

func (cc *CompatibilityChecker) testBasicLockUnlock(ctx context.Context) error {
	return cc.VerifyBackwardCompatibility(ctx)
}

func (cc *CompatibilityChecker) testConcurrentAccess(ctx context.Context) error {
	// Test that concurrent lock attempts behave correctly
	testProject := models.Project{
		RepoFullName: "test/concurrent-check",
		Path:         ".",
	}
	testUser1 := models.User{Username: "user1"}
	testUser2 := models.User{Username: "user2"}
	testWorkspace := "default"

	// User1 acquires lock
	acquired1, _, err := cc.adapter.TryLock(models.ProjectLock{
		Project: testProject, Workspace: testWorkspace, User: testUser1, Time: time.Now(),
	})
	if err != nil {
		return err
	}
	if !acquired1 {
		return fmt.Errorf("first lock not acquired")
	}

	// User2 should not be able to acquire same lock
	acquired2, _, err := cc.adapter.TryLock(models.ProjectLock{
		Project: testProject, Workspace: testWorkspace, User: testUser2, Time: time.Now(),
	})
	if err != nil {
		return err
	}
	if acquired2 {
		return fmt.Errorf("second lock should not have been acquired")
	}
	// Note: Legacy behavior - lock failure reasons are not available in simplified interface

	// Clean up
	cc.adapter.Unlock(testProject, testWorkspace)
	return nil
}

func (cc *CompatibilityChecker) testUnlockByPull(ctx context.Context) error {
	// Test UnlockByPull functionality
	testRepo := "test/unlock-by-pull-check"
	testProject := models.Project{RepoFullName: testRepo, Path: "."}
	testUser := models.User{Username: "test-user"}

	// Create multiple locks for the same repo
	workspaces := []string{"ws1", "ws2", "ws3"}
	for _, ws := range workspaces {
		acquired, _, err := cc.adapter.TryLock(models.ProjectLock{
			Project: testProject, Workspace: ws, User: testUser, Time: time.Now(),
		})
		if err != nil {
			return err
		}
		if !acquired {
			return fmt.Errorf("failed to acquire lock for workspace %s", ws)
		}
	}

	// Unlock all locks for the repository
	unlockedLocks, err := cc.adapter.UnlockByPull(testRepo, 123)
	if err != nil {
		return err
	}

	if len(unlockedLocks) != len(workspaces) {
		return fmt.Errorf("expected %d unlocked locks, got %d", len(workspaces), len(unlockedLocks))
	}

	return nil
}

func (cc *CompatibilityChecker) testListConsistency(ctx context.Context) error {
	// Test that List returns consistent results
	initialLocks, err := cc.adapter.List()
	if err != nil {
		return err
	}

	testProject := models.Project{RepoFullName: "test/list-consistency", Path: "."}
	testUser := models.User{Username: "test-user"}
	testWorkspace := "default"

	// Add a lock
	acquired, _, err := cc.adapter.TryLock(models.ProjectLock{
		Project: testProject, Workspace: testWorkspace, User: testUser, Time: time.Now(),
	})
	if err != nil {
		return err
	}
	if !acquired {
		return fmt.Errorf("failed to acquire test lock")
	}

	// Check that List includes the new lock
	updatedLocks, err := cc.adapter.List()
	if err != nil {
		return err
	}

	if len(updatedLocks) != len(initialLocks)+1 {
		return fmt.Errorf("expected %d locks after adding one, got %d", len(initialLocks)+1, len(updatedLocks))
	}

	// Remove the lock
	_, err = cc.adapter.Unlock(testProject, testWorkspace)
	if err != nil {
		return err
	}

	// Check that List no longer includes the lock
	finalLocks, err := cc.adapter.List()
	if err != nil {
		return err
	}

	if len(finalLocks) != len(initialLocks) {
		return fmt.Errorf("expected %d locks after removal, got %d", len(initialLocks), len(finalLocks))
	}

	return nil
}

// CompatibilityReport contains the results of compatibility testing
type CompatibilityReport struct {
	Success   bool                `json:"success"`
	StartTime time.Time           `json:"start_time"`
	EndTime   time.Time           `json:"end_time"`
	Duration  time.Duration       `json:"duration"`
	Tests     []CompatibilityTest `json:"tests"`
}

// CompatibilityTest represents a single compatibility test result
type CompatibilityTest struct {
	Name      string        `json:"name"`
	Success   bool          `json:"success"`
	Error     string        `json:"error,omitempty"`
	StartTime time.Time     `json:"start_time"`
	EndTime   time.Time     `json:"end_time"`
	Duration  time.Duration `json:"duration"`
}
