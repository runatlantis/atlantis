package events

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/runatlantis/atlantis/server/core/locking"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/events/vcs"
	"github.com/runatlantis/atlantis/server/logging"
)

// EnhancedLockingSystem provides a robust locking mechanism that addresses race conditions
type EnhancedLockingSystem struct {
	locker           locking.Locker
	backend          locking.Backend
	vcsClient        vcs.Client
	logger           logging.SimpleLogging
	queueManager     models.PlanQueueManager
	enableQueue      bool
	enableRetry      bool
	maxRetryAttempts int
	retryDelay       time.Duration

	// In-memory locks to prevent race conditions
	memoryLocks      map[string]*MemoryLock
	memoryLocksMutex sync.RWMutex

	// Working directory protection
	workingDirLocks map[string]*WorkingDirLock
	workingDirMutex sync.RWMutex
}

// MemoryLock represents an in-memory lock to prevent race conditions
type MemoryLock struct {
	Project   models.Project
	Workspace string
	Pull      models.PullRequest
	User      models.User
	Time      time.Time
	Context   context.Context
	Cancel    context.CancelFunc
}

// WorkingDirLock protects working directories from premature deletion
type WorkingDirLock struct {
	RepoFullName string
	PullNum      int
	Workspace    string
	Time         time.Time
	Context      context.Context
	Cancel       context.CancelFunc
}

// NewEnhancedLockingSystem creates a new enhanced locking system
func NewEnhancedLockingSystem(
	locker locking.Locker,
	backend locking.Backend,
	vcsClient vcs.Client,
	logger logging.SimpleLogging,
	queueManager models.PlanQueueManager,
	enableQueue bool,
	enableRetry bool,
	maxRetryAttempts int,
	retryDelay int,
) *EnhancedLockingSystem {
	return &EnhancedLockingSystem{
		locker:           locker,
		backend:          backend,
		vcsClient:        vcsClient,
		logger:           logger,
		queueManager:     queueManager,
		enableQueue:      enableQueue,
		enableRetry:      enableRetry,
		maxRetryAttempts: maxRetryAttempts,
		retryDelay:       time.Duration(retryDelay) * time.Second,
		memoryLocks:      make(map[string]*MemoryLock),
		workingDirLocks:  make(map[string]*WorkingDirLock),
	}
}

// TryLockWithRetry attempts to acquire a lock with retry logic
func (e *EnhancedLockingSystem) TryLockWithRetry(
	project models.Project,
	workspace string,
	pull models.PullRequest,
	user models.User,
) (*TryLockResponse, error) {

	// First, try to acquire memory lock to prevent race conditions
	memoryLockKey := e.memoryLockKey(project, workspace)
	if !e.tryAcquireMemoryLock(memoryLockKey, project, workspace, pull, user) {
		return &TryLockResponse{
			LockAcquired:      false,
			LockFailureReason: "Another operation is in progress for this project/workspace",
		}, nil
	}

	// Try to acquire the actual lock
	var lastErr error
	for attempt := 1; attempt <= e.maxRetryAttempts; attempt++ {
		lockAttempt, err := e.locker.TryLock(project, workspace, pull, user)
		if err != nil {
			lastErr = err
			e.logger.Warn("Lock attempt %d failed: %s", attempt, err)
			continue
		}

		if lockAttempt.LockAcquired {
			e.logger.Info("Successfully acquired lock on attempt %d", attempt)
			return &TryLockResponse{
				LockAcquired: true,
				UnlockFn: func() error {
					// Release memory lock when unlocking
					e.releaseMemoryLock(memoryLockKey)
					return e.unlockWithCleanup(project, workspace, lockAttempt.LockKey)
				},
				LockKey: lockAttempt.LockKey,
			}, nil
		}

		// If lock not acquired and it's not our own lock
		if lockAttempt.CurrLock.Pull.Num != pull.Num {
			if e.enableQueue {
				return e.handleQueueLogic(project, workspace, pull, user, lockAttempt)
			}

			// Return failure with retry information
			if e.enableRetry && attempt < e.maxRetryAttempts {
				e.logger.Info("Lock busy, retrying in %v (attempt %d/%d)", e.retryDelay, attempt, e.maxRetryAttempts)
				time.Sleep(e.retryDelay)
				continue
			}

			link, err := e.vcsClient.MarkdownPullLink(lockAttempt.CurrLock.Pull)
			if err != nil {
				// Release memory lock on error
				e.releaseMemoryLock(memoryLockKey)
				return nil, err
			}

			failureMsg := fmt.Sprintf(
				"This project is currently locked by an unapplied plan from pull %s. To continue, delete the lock from %s or apply that plan and merge the pull request.\n\n"+
					"Once the lock is released, comment `atlantis plan` here to re-plan.",
				link, link)

			// Release memory lock on failure
			e.releaseMemoryLock(memoryLockKey)
			return &TryLockResponse{
				LockAcquired:      false,
				LockFailureReason: failureMsg,
			}, nil
		}

		// If it's our own lock, return success
		return &TryLockResponse{
			LockAcquired: true,
			UnlockFn: func() error {
				// Release memory lock when unlocking
				e.releaseMemoryLock(memoryLockKey)
				return e.unlockWithCleanup(project, workspace, lockAttempt.LockKey)
			},
			LockKey: lockAttempt.LockKey,
		}, nil
	}

	// Release memory lock on final failure
	e.releaseMemoryLock(memoryLockKey)
	return nil, fmt.Errorf("failed to acquire lock after %d attempts: %w", e.maxRetryAttempts, lastErr)
}

// handleQueueLogic handles the queue logic when lock is not available
func (e *EnhancedLockingSystem) handleQueueLogic(
	project models.Project,
	workspace string,
	pull models.PullRequest,
	user models.User,
	lockAttempt locking.TryLockResponse,
) (*TryLockResponse, error) {

	// Check if we're already in the queue
	inQueue, err := e.queueManager.IsInQueue(project, workspace, pull.Num)
	if err != nil {
		e.logger.Warn("Error checking queue status: %s", err)
	}

	if !inQueue {
		// Add to queue
		queueEntry := models.PlanQueueEntry{
			ID:        fmt.Sprintf("%s-%s-%d", project.String(), workspace, pull.Num),
			Project:   project,
			Workspace: workspace,
			Pull:      pull,
			User:      user,
			Time:      time.Now(),
			Priority:  0,
			Command:   "plan",
		}

		if err := e.queueManager.AddToQueue(queueEntry); err != nil {
			e.logger.Warn("Failed to add to queue: %s", err)
		} else {
			e.logger.Info("Added PR %d to plan queue for project %s, workspace %s", pull.Num, project.String(), workspace)
		}
	}

	// Create queue-aware failure message
	link, err := e.vcsClient.MarkdownPullLink(lockAttempt.CurrLock.Pull)
	if err != nil {
		return nil, err
	}

	failureMsg := fmt.Sprintf(
		"This project is currently locked by an unapplied plan from pull %s. To continue, delete the lock from %s or apply that plan and merge the pull request.\n\n"+
			"Your plan request has been added to the queue. You will be notified when it's your turn to plan.\n\n"+
			"Once the lock is released, comment `atlantis plan` here to re-plan.",
		link, link)

	return &TryLockResponse{
		LockAcquired:      false,
		LockFailureReason: failureMsg,
	}, nil
}

// unlockWithCleanup unlocks and handles cleanup
func (e *EnhancedLockingSystem) unlockWithCleanup(project models.Project, workspace, lockKey string) error {
	// Unlock the actual lock
	_, err := e.locker.Unlock(lockKey)
	if err != nil {
		return fmt.Errorf("unlocking: %w", err)
	}

	// Try to transfer lock to next person in queue
	if e.enableQueue && e.queueManager != nil {
		if transferErr := e.queueManager.TransferLock(project, workspace); transferErr != nil {
			e.logger.Warn("Failed to transfer lock to next person in queue: %s", transferErr)
		}
	}

	return nil
}

// ProtectWorkingDir protects a working directory from deletion
func (e *EnhancedLockingSystem) ProtectWorkingDir(repoFullName string, pullNum int, workspace string) context.CancelFunc {
	e.workingDirMutex.Lock()
	defer e.workingDirMutex.Unlock()

	key := fmt.Sprintf("%s:%d:%s", repoFullName, pullNum, workspace)
	ctx, cancel := context.WithCancel(context.Background())

	e.workingDirLocks[key] = &WorkingDirLock{
		RepoFullName: repoFullName,
		PullNum:      pullNum,
		Workspace:    workspace,
		Time:         time.Now(),
		Context:      ctx,
		Cancel:       cancel,
	}

	e.logger.Debug("Protected working directory: %s", key)
	return cancel
}

// IsWorkingDirProtected checks if a working directory is protected
func (e *EnhancedLockingSystem) IsWorkingDirProtected(repoFullName string, pullNum int, workspace string) bool {
	e.workingDirMutex.RLock()
	defer e.workingDirMutex.RUnlock()

	key := fmt.Sprintf("%s:%d:%s", repoFullName, pullNum, workspace)
	_, exists := e.workingDirLocks[key]
	return exists
}

// CleanupWorkingDirProtection removes protection for a working directory
func (e *EnhancedLockingSystem) CleanupWorkingDirProtection(repoFullName string, pullNum int, workspace string) {
	e.workingDirMutex.Lock()
	defer e.workingDirMutex.Unlock()

	key := fmt.Sprintf("%s:%d:%s", repoFullName, pullNum, workspace)
	if lock, exists := e.workingDirLocks[key]; exists {
		lock.Cancel()
		delete(e.workingDirLocks, key)
		e.logger.Debug("Removed working directory protection: %s", key)
	}
}

// tryAcquireMemoryLock attempts to acquire a memory lock
func (e *EnhancedLockingSystem) tryAcquireMemoryLock(key string, project models.Project, workspace string, pull models.PullRequest, user models.User) bool {
	e.memoryLocksMutex.Lock()
	defer e.memoryLocksMutex.Unlock()

	if _, exists := e.memoryLocks[key]; exists {
		return false
	}

	ctx, cancel := context.WithCancel(context.Background())
	e.memoryLocks[key] = &MemoryLock{
		Project:   project,
		Workspace: workspace,
		Pull:      pull,
		User:      user,
		Time:      time.Now(),
		Context:   ctx,
		Cancel:    cancel,
	}

	return true
}

// releaseMemoryLock releases a memory lock
func (e *EnhancedLockingSystem) releaseMemoryLock(key string) {
	e.memoryLocksMutex.Lock()
	defer e.memoryLocksMutex.Unlock()

	if lock, exists := e.memoryLocks[key]; exists {
		lock.Cancel()
		delete(e.memoryLocks, key)
	}
}

// memoryLockKey generates a key for memory locks
func (e *EnhancedLockingSystem) memoryLockKey(project models.Project, workspace string) string {
	return fmt.Sprintf("memory:%s:%s:%s", project.RepoFullName, project.Path, workspace)
}

// CleanupAllLocks cleans up all locks for a pull request
func (e *EnhancedLockingSystem) CleanupAllLocks(repoFullName string, pullNum int) error {
	// Clean up queue entries
	if e.enableQueue && e.queueManager != nil {
		if err := e.queueManager.CleanupQueue(repoFullName, pullNum); err != nil {
			e.logger.Warn("Failed to cleanup queue: %s", err)
		}
	}

	// Clean up working directory protection
	e.workingDirMutex.Lock()
	defer e.workingDirMutex.Unlock()

	keysToRemove := []string{}
	for key, lock := range e.workingDirLocks {
		if lock.RepoFullName == repoFullName && lock.PullNum == pullNum {
			keysToRemove = append(keysToRemove, key)
		}
	}

	for _, key := range keysToRemove {
		if lock, exists := e.workingDirLocks[key]; exists {
			lock.Cancel()
			delete(e.workingDirLocks, key)
			e.logger.Debug("Cleaned up working directory protection: %s", key)
		}
	}

	return nil
}
