package enhanced

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/runatlantis/atlantis/server/core/locking"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/logging"
)

// CompatibilityMode defines different compatibility levels
type CompatibilityMode string

const (
	CompatibilityModeStrict  CompatibilityMode = "strict"   // Exact legacy behavior
	CompatibilityModeHybrid CompatibilityMode = "hybrid"   // Enhanced with fallback
	CompatibilityModeNative CompatibilityMode = "native"   // Full enhanced mode
)

// CompatibilityLayer provides seamless migration from legacy to enhanced locking
type CompatibilityLayer struct {
	mode           CompatibilityMode
	enhanced       Backend
	legacy         locking.Backend
	config         *EnhancedConfig
	log            logging.SimpleLogging

	// Migration state
	migrationState *MigrationState
	migrationLock  sync.RWMutex

	// Performance tracking
	legacyOps      int64
	enhancedOps    int64
	fallbackOps    int64

	// Runtime switching
	enableSwitching bool
	lastHealthCheck time.Time
	healthCheckMu   sync.RWMutex
}

// MigrationState tracks the progress of migrating from legacy to enhanced locking
type MigrationState struct {
	Phase              MigrationPhase    `json:"phase"`
	StartTime          time.Time         `json:"start_time"`
	LastUpdate         time.Time         `json:"last_update"`
	LegacyLocksCount   int               `json:"legacy_locks_count"`
	EnhancedLocksCount int               `json:"enhanced_locks_count"`
	MigratedCount      int               `json:"migrated_count"`
	FailedMigrations   []string          `json:"failed_migrations"`
	EstimatedComplete  *time.Time        `json:"estimated_complete,omitempty"`
}

// MigrationPhase represents the current phase of migration
type MigrationPhase string

const (
	MigrationPhaseNotStarted  MigrationPhase = "not_started"
	MigrationPhasePreparation MigrationPhase = "preparation"
	MigrationPhaseActive      MigrationPhase = "active"
	MigrationPhaseCompleted   MigrationPhase = "completed"
	MigrationPhaseFailed      MigrationPhase = "failed"
)

// NewCompatibilityLayer creates a new compatibility layer
func NewCompatibilityLayer(mode CompatibilityMode, enhanced Backend, legacy locking.Backend, config *EnhancedConfig, log logging.SimpleLogging) *CompatibilityLayer {
	return &CompatibilityLayer{
		mode:     mode,
		enhanced: enhanced,
		legacy:   legacy,
		config:   config,
		log:      log,
		migrationState: &MigrationState{
			Phase:     MigrationPhaseNotStarted,
			StartTime: time.Now(),
		},
		enableSwitching: true,
	}
}

// TryLock attempts to acquire a lock with compatibility handling
func (cl *CompatibilityLayer) TryLock(lock models.ProjectLock) (bool, models.ProjectLock, error) {
	switch cl.mode {
	case CompatibilityModeStrict:
		return cl.tryLockLegacy(lock)
	case CompatibilityModeHybrid:
		return cl.tryLockHybrid(lock)
	case CompatibilityModeNative:
		return cl.tryLockEnhanced(lock)
	default:
		return cl.tryLockHybrid(lock) // Default to hybrid
	}
}

// tryLockLegacy uses only the legacy backend
func (cl *CompatibilityLayer) tryLockLegacy(lock models.ProjectLock) (bool, models.ProjectLock, error) {
	cl.legacyOps++
	cl.log.Debug("Using legacy locking for TryLock: %s/%s", lock.Project.RepoFullName, lock.Workspace)
	return cl.legacy.TryLock(lock)
}

// tryLockEnhanced uses only the enhanced backend
func (cl *CompatibilityLayer) tryLockEnhanced(lock models.ProjectLock) (bool, models.ProjectLock, error) {
	cl.enhancedOps++
	cl.log.Debug("Using enhanced locking for TryLock: %s/%s", lock.Project.RepoFullName, lock.Workspace)

	ctx := context.Background()
	request := cl.convertToEnhancedRequest(lock)

	enhancedLock, acquired, err := cl.enhanced.TryAcquireLock(ctx, request)
	if err != nil {
		return false, models.ProjectLock{}, err
	}

	if acquired && enhancedLock != nil {
		return true, lock, nil
	}

	// Lock not acquired - find who holds it
	existingLock, err := cl.findExistingLock(ctx, request.Resource)
	if err != nil {
		return false, models.ProjectLock{}, err
	}

	if existingLock != nil {
		legacyLock := cl.enhanced.ConvertToLegacy(existingLock)
		return false, *legacyLock, nil
	}

	return false, models.ProjectLock{}, nil
}

// tryLockHybrid tries enhanced first, falls back to legacy on failure
func (cl *CompatibilityLayer) tryLockHybrid(lock models.ProjectLock) (bool, models.ProjectLock, error) {
	// First, check if enhanced locking is healthy
	if cl.isEnhancedHealthy() {
		acquired, currLock, err := cl.tryLockEnhanced(lock)
		if err == nil {
			return acquired, currLock, nil
		}

		cl.log.Warn("Enhanced locking failed, falling back to legacy: %v", err)
		cl.fallbackOps++
	}

	// Fall back to legacy
	return cl.tryLockLegacy(lock)
}

// Unlock releases a lock with compatibility handling
func (cl *CompatibilityLayer) Unlock(project models.Project, workspace string) (*models.ProjectLock, error) {
	switch cl.mode {
	case CompatibilityModeStrict:
		return cl.unlockLegacy(project, workspace)
	case CompatibilityModeHybrid:
		return cl.unlockHybrid(project, workspace)
	case CompatibilityModeNative:
		return cl.unlockEnhanced(project, workspace)
	default:
		return cl.unlockHybrid(project, workspace)
	}
}

// unlockLegacy uses only the legacy backend
func (cl *CompatibilityLayer) unlockLegacy(project models.Project, workspace string) (*models.ProjectLock, error) {
	cl.legacyOps++
	return cl.legacy.Unlock(project, workspace)
}

// unlockEnhanced uses only the enhanced backend
func (cl *CompatibilityLayer) unlockEnhanced(project models.Project, workspace string) (*models.ProjectLock, error) {
	cl.enhancedOps++
	ctx := context.Background()

	// Find the lock to unlock
	locks, err := cl.enhanced.ListLocks(ctx)
	if err != nil {
		return nil, err
	}

	var targetLock *EnhancedLock
	for _, lock := range locks {
		if lock.Resource.Namespace == project.RepoFullName &&
		   lock.Resource.Path == project.Path &&
		   lock.Resource.Workspace == workspace {
			targetLock = lock
			break
		}
	}

	if targetLock == nil {
		return nil, nil // No lock found - this is normal
	}

	err = cl.enhanced.ReleaseLock(ctx, targetLock.ID)
	if err != nil {
		return nil, err
	}

	return cl.enhanced.ConvertToLegacy(targetLock), nil
}

// unlockHybrid tries enhanced first, falls back to legacy
func (cl *CompatibilityLayer) unlockHybrid(project models.Project, workspace string) (*models.ProjectLock, error) {
	// Try enhanced first
	if cl.isEnhancedHealthy() {
		lock, err := cl.unlockEnhanced(project, workspace)
		if err == nil {
			return lock, nil
		}

		cl.log.Warn("Enhanced unlock failed, falling back to legacy: %v", err)
		cl.fallbackOps++
	}

	// Fall back to legacy
	return cl.unlockLegacy(project, workspace)
}

// List returns all current locks with compatibility handling
func (cl *CompatibilityLayer) List() ([]models.ProjectLock, error) {
	switch cl.mode {
	case CompatibilityModeStrict:
		return cl.listLegacy()
	case CompatibilityModeHybrid:
		return cl.listHybrid()
	case CompatibilityModeNative:
		return cl.listEnhanced()
	default:
		return cl.listHybrid()
	}
}

// listLegacy uses only the legacy backend
func (cl *CompatibilityLayer) listLegacy() ([]models.ProjectLock, error) {
	cl.legacyOps++
	return cl.legacy.List()
}

// listEnhanced uses only the enhanced backend
func (cl *CompatibilityLayer) listEnhanced() ([]models.ProjectLock, error) {
	cl.enhancedOps++
	ctx := context.Background()

	locks, err := cl.enhanced.ListLocks(ctx)
	if err != nil {
		return nil, err
	}

	var legacyLocks []models.ProjectLock
	for _, lock := range locks {
		if lock.State == LockStateAcquired {
			legacyLock := cl.enhanced.ConvertToLegacy(lock)
			if legacyLock != nil {
				legacyLocks = append(legacyLocks, *legacyLock)
			}
		}
	}

	return legacyLocks, nil
}

// listHybrid tries to merge results from both backends
func (cl *CompatibilityLayer) listHybrid() ([]models.ProjectLock, error) {
	var allLocks []models.ProjectLock
	lockMap := make(map[string]models.ProjectLock) // Deduplicate by key

	// Get enhanced locks
	if cl.isEnhancedHealthy() {
		enhancedLocks, err := cl.listEnhanced()
		if err == nil {
			for _, lock := range enhancedLocks {
				key := fmt.Sprintf("%s/%s/%s", lock.Project.RepoFullName, lock.Project.Path, lock.Workspace)
				lockMap[key] = lock
			}
		} else {
			cl.log.Warn("Enhanced list failed: %v", err)
			cl.fallbackOps++
		}
	}

	// Get legacy locks (might overlap with enhanced)
	legacyLocks, err := cl.listLegacy()
	if err == nil {
		for _, lock := range legacyLocks {
			key := fmt.Sprintf("%s/%s/%s", lock.Project.RepoFullName, lock.Project.Path, lock.Workspace)
			if _, exists := lockMap[key]; !exists {
				lockMap[key] = lock
			}
		}
	}

	// Convert map back to slice
	for _, lock := range lockMap {
		allLocks = append(allLocks, lock)
	}

	return allLocks, nil
}

// GetLock retrieves a specific lock with compatibility handling
func (cl *CompatibilityLayer) GetLock(project models.Project, workspace string) (*models.ProjectLock, error) {
	switch cl.mode {
	case CompatibilityModeStrict:
		return cl.getLockLegacy(project, workspace)
	case CompatibilityModeHybrid:
		return cl.getLockHybrid(project, workspace)
	case CompatibilityModeNative:
		return cl.getLockEnhanced(project, workspace)
	default:
		return cl.getLockHybrid(project, workspace)
	}
}

// getLockLegacy uses only the legacy backend
func (cl *CompatibilityLayer) getLockLegacy(project models.Project, workspace string) (*models.ProjectLock, error) {
	cl.legacyOps++
	return cl.legacy.GetLock(project, workspace)
}

// getLockEnhanced uses only the enhanced backend
func (cl *CompatibilityLayer) getLockEnhanced(project models.Project, workspace string) (*models.ProjectLock, error) {
	cl.enhancedOps++
	return cl.enhanced.GetLegacyLock(project, workspace)
}

// getLockHybrid tries enhanced first, falls back to legacy
func (cl *CompatibilityLayer) getLockHybrid(project models.Project, workspace string) (*models.ProjectLock, error) {
	// Try enhanced first
	if cl.isEnhancedHealthy() {
		lock, err := cl.getLockEnhanced(project, workspace)
		if err == nil {
			return lock, nil
		}

		cl.log.Debug("Enhanced GetLock failed, trying legacy: %v", err)
		cl.fallbackOps++
	}

	// Fall back to legacy
	return cl.getLockLegacy(project, workspace)
}

// UnlockByPull releases all locks for a pull request with compatibility handling
func (cl *CompatibilityLayer) UnlockByPull(repoFullName string, pullNum int) ([]models.ProjectLock, error) {
	switch cl.mode {
	case CompatibilityModeStrict:
		return cl.unlockByPullLegacy(repoFullName, pullNum)
	case CompatibilityModeHybrid:
		return cl.unlockByPullHybrid(repoFullName, pullNum)
	case CompatibilityModeNative:
		return cl.unlockByPullEnhanced(repoFullName, pullNum)
	default:
		return cl.unlockByPullHybrid(repoFullName, pullNum)
	}
}

// unlockByPullLegacy uses only the legacy backend
func (cl *CompatibilityLayer) unlockByPullLegacy(repoFullName string, pullNum int) ([]models.ProjectLock, error) {
	cl.legacyOps++
	return cl.legacy.UnlockByPull(repoFullName, pullNum)
}

// unlockByPullEnhanced uses only the enhanced backend
func (cl *CompatibilityLayer) unlockByPullEnhanced(repoFullName string, pullNum int) ([]models.ProjectLock, error) {
	cl.enhancedOps++
	ctx := context.Background()

	// Get all locks for this repository
	locks, err := cl.enhanced.ListLocks(ctx)
	if err != nil {
		return nil, err
	}

	var unlockedLocks []models.ProjectLock

	// Find and release locks for this repository
	for _, lock := range locks {
		if lock.Resource.Namespace == repoFullName && lock.State == LockStateAcquired {
			err := cl.enhanced.ReleaseLock(ctx, lock.ID)
			if err != nil {
				cl.log.Warn("Failed to release lock %s during UnlockByPull: %v", lock.ID, err)
				continue
			}

			legacyLock := cl.enhanced.ConvertToLegacy(lock)
			if legacyLock != nil {
				unlockedLocks = append(unlockedLocks, *legacyLock)
			}
		}
	}

	return unlockedLocks, nil
}

// unlockByPullHybrid tries to unlock from both backends
func (cl *CompatibilityLayer) unlockByPullHybrid(repoFullName string, pullNum int) ([]models.ProjectLock, error) {
	var allUnlocked []models.ProjectLock
	unlockedMap := make(map[string]models.ProjectLock) // Deduplicate

	// Try enhanced first
	if cl.isEnhancedHealthy() {
		enhancedUnlocked, err := cl.unlockByPullEnhanced(repoFullName, pullNum)
		if err == nil {
			for _, lock := range enhancedUnlocked {
				key := fmt.Sprintf("%s/%s/%s", lock.Project.RepoFullName, lock.Project.Path, lock.Workspace)
				unlockedMap[key] = lock
			}
		} else {
			cl.log.Warn("Enhanced UnlockByPull failed: %v", err)
			cl.fallbackOps++
		}
	}

	// Also try legacy (for any locks not in enhanced system)
	legacyUnlocked, err := cl.unlockByPullLegacy(repoFullName, pullNum)
	if err == nil {
		for _, lock := range legacyUnlocked {
			key := fmt.Sprintf("%s/%s/%s", lock.Project.RepoFullName, lock.Project.Path, lock.Workspace)
			if _, exists := unlockedMap[key]; !exists {
				unlockedMap[key] = lock
			}
		}
	}

	// Convert map back to slice
	for _, lock := range unlockedMap {
		allUnlocked = append(allUnlocked, lock)
	}

	return allUnlocked, nil
}

// Migration management

// StartMigration begins the migration from legacy to enhanced locking
func (cl *CompatibilityLayer) StartMigration(ctx context.Context) error {
	cl.migrationLock.Lock()
	defer cl.migrationLock.Unlock()

	if cl.migrationState.Phase != MigrationPhaseNotStarted {
		return fmt.Errorf("migration already in progress or completed")
	}

	cl.migrationState.Phase = MigrationPhasePreparation
	cl.migrationState.LastUpdate = time.Now()
	cl.log.Info("Starting migration from legacy to enhanced locking")

	// Phase 1: Analyze current state
	legacyLocks, err := cl.legacy.List()
	if err != nil {
		cl.migrationState.Phase = MigrationPhaseFailed
		return fmt.Errorf("failed to list legacy locks: %w", err)
	}

	cl.migrationState.LegacyLocksCount = len(legacyLocks)
	cl.migrationState.Phase = MigrationPhaseActive

	// Phase 2: Migrate existing locks (if any)
	if len(legacyLocks) > 0 {
		cl.log.Info("Migrating %d existing legacy locks", len(legacyLocks))

		for _, legacyLock := range legacyLocks {
			err := cl.migrateSingleLock(ctx, legacyLock)
			if err != nil {
				cl.migrationState.FailedMigrations = append(cl.migrationState.FailedMigrations,
					fmt.Sprintf("%s/%s: %v", legacyLock.Project.RepoFullName, legacyLock.Workspace, err))
				cl.log.Warn("Failed to migrate lock %s/%s: %v", legacyLock.Project.RepoFullName, legacyLock.Workspace, err)
			} else {
				cl.migrationState.MigratedCount++
			}
		}
	}

	// Phase 3: Complete migration
	cl.migrationState.Phase = MigrationPhaseCompleted
	cl.migrationState.LastUpdate = time.Now()

	cl.log.Info("Migration completed: %d/%d locks migrated successfully",
		cl.migrationState.MigratedCount, cl.migrationState.LegacyLocksCount)

	return nil
}

// migrateSingleLock migrates a single lock from legacy to enhanced format
func (cl *CompatibilityLayer) migrateSingleLock(ctx context.Context, legacyLock models.ProjectLock) error {
	request := cl.convertToEnhancedRequest(legacyLock)

	// Try to acquire the lock in the enhanced system
	_, acquired, err := cl.enhanced.TryAcquireLock(ctx, request)
	if err != nil {
		return err
	}

	if !acquired {
		return fmt.Errorf("failed to acquire lock in enhanced system")
	}

	return nil
}

// GetMigrationStatus returns the current migration status
func (cl *CompatibilityLayer) GetMigrationStatus() *MigrationState {
	cl.migrationLock.RLock()
	defer cl.migrationLock.RUnlock()

	// Return a copy to prevent external modifications
	status := *cl.migrationState
	return &status
}

// SetCompatibilityMode changes the compatibility mode at runtime
func (cl *CompatibilityLayer) SetCompatibilityMode(mode CompatibilityMode) error {
	if !cl.enableSwitching {
		return fmt.Errorf("runtime switching is disabled")
	}

	cl.log.Info("Switching compatibility mode from %s to %s", cl.mode, mode)
	cl.mode = mode
	return nil
}

// GetCompatibilityMode returns the current compatibility mode
func (cl *CompatibilityLayer) GetCompatibilityMode() CompatibilityMode {
	return cl.mode
}

// Health checking and monitoring

// isEnhancedHealthy checks if the enhanced backend is functioning properly
func (cl *CompatibilityLayer) isEnhancedHealthy() bool {
	cl.healthCheckMu.RLock()
	if time.Since(cl.lastHealthCheck) < 30*time.Second {
		cl.healthCheckMu.RUnlock()
		return true // Assume healthy if recently checked
	}
	cl.healthCheckMu.RUnlock()

	cl.healthCheckMu.Lock()
	defer cl.healthCheckMu.Unlock()

	// Double-check after acquiring write lock
	if time.Since(cl.lastHealthCheck) < 30*time.Second {
		return true
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := cl.enhanced.HealthCheck(ctx)
	cl.lastHealthCheck = time.Now()

	if err != nil {
		cl.log.Warn("Enhanced backend health check failed: %v", err)
		return false
	}

	return true
}

// GetPerformanceStats returns performance statistics for the compatibility layer
func (cl *CompatibilityLayer) GetPerformanceStats() *CompatibilityStats {
	return &CompatibilityStats{
		Mode:         cl.mode,
		LegacyOps:    cl.legacyOps,
		EnhancedOps:  cl.enhancedOps,
		FallbackOps:  cl.fallbackOps,
		MigrationState: cl.GetMigrationStatus(),
	}
}

// CompatibilityStats provides performance and operational statistics
type CompatibilityStats struct {
	Mode           CompatibilityMode  `json:"mode"`
	LegacyOps      int64             `json:"legacy_ops"`
	EnhancedOps    int64             `json:"enhanced_ops"`
	FallbackOps    int64             `json:"fallback_ops"`
	MigrationState *MigrationState   `json:"migration_state"`
}

// Helper functions

// convertToEnhancedRequest converts a legacy lock to an enhanced request
func (cl *CompatibilityLayer) convertToEnhancedRequest(lock models.ProjectLock) *EnhancedLockRequest {
	return &EnhancedLockRequest{
		ID: generateRequestID(),
		Resource: ResourceIdentifier{
			Type:      ResourceTypeProject,
			Namespace: lock.Project.RepoFullName,
			Name:      lock.Project.Path,
			Workspace: lock.Workspace,
			Path:      lock.Project.Path,
		},
		Priority:    PriorityNormal,
		Timeout:     cl.config.DefaultTimeout,
		Metadata:    make(map[string]string),
		Context:     context.Background(),
		RequestedAt: lock.Time,
		Project:     lock.Project,
		Workspace:   lock.Workspace,
		User:        lock.User,
	}
}

// findExistingLock finds an existing lock for the given resource
func (cl *CompatibilityLayer) findExistingLock(ctx context.Context, resource ResourceIdentifier) (*EnhancedLock, error) {
	locks, err := cl.enhanced.ListLocks(ctx)
	if err != nil {
		return nil, err
	}

	for _, lock := range locks {
		if lock.Resource.Namespace == resource.Namespace &&
		   lock.Resource.Name == resource.Name &&
		   lock.Resource.Workspace == resource.Workspace &&
		   lock.State == LockStateAcquired {
			return lock, nil
		}
	}

	return nil, nil
}