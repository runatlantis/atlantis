package events

import (
	"context"

	"github.com/runatlantis/atlantis/server/core/locking"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/events/vcs"
	"github.com/runatlantis/atlantis/server/logging"
)

// EnhancedProjectLocker extends the default project locker with enhanced features
type EnhancedProjectLocker struct {
	*DefaultProjectLocker
	enhancedLocking *EnhancedLockingSystem
	enableQueue     bool
	enableRetry     bool
}

// NewEnhancedProjectLocker creates a new enhanced project locker
func NewEnhancedProjectLocker(
	locker locking.Locker,
	noOpLocker locking.Locker,
	vcsClient vcs.Client,
	enhancedLocking *EnhancedLockingSystem,
	enableQueue bool,
	enableRetry bool,
) *EnhancedProjectLocker {
	return &EnhancedProjectLocker{
		DefaultProjectLocker: &DefaultProjectLocker{
			Locker:     locker,
			NoOpLocker: noOpLocker,
			VCSClient:  vcsClient,
		},
		enhancedLocking: enhancedLocking,
		enableQueue:     enableQueue,
		enableRetry:     enableRetry,
	}
}

// TryLock attempts to acquire a lock with enhanced features
func (e *EnhancedProjectLocker) TryLock(log logging.SimpleLogging, pull models.PullRequest, user models.User, workspace string, project models.Project, repoLocking bool) (*TryLockResponse, error) {

	// If enhanced locking is enabled, use it
	if e.enhancedLocking != nil {
		// Use enhanced locking system
		response, err := e.enhancedLocking.TryLockWithRetry(project, workspace, pull, user)
		if err != nil {
			return nil, err
		}

		if response.LockAcquired {
			log.Info("Acquired lock with id '%s' using enhanced locking", response.LockKey)
		}

		return response, nil
	}

	// Fall back to default behavior
	return e.DefaultProjectLocker.TryLock(log, pull, user, workspace, project, repoLocking)
}

// ProtectWorkingDir protects a working directory from deletion
func (e *EnhancedProjectLocker) ProtectWorkingDir(repoFullName string, pullNum int, workspace string) context.CancelFunc {
	if e.enhancedLocking != nil {
		return e.enhancedLocking.ProtectWorkingDir(repoFullName, pullNum, workspace)
	}

	// Return a no-op cancel function if enhanced locking is not available
	return func() {}
}

// IsWorkingDirProtected checks if a working directory is protected
func (e *EnhancedProjectLocker) IsWorkingDirProtected(repoFullName string, pullNum int, workspace string) bool {
	if e.enhancedLocking != nil {
		return e.enhancedLocking.IsWorkingDirProtected(repoFullName, pullNum, workspace)
	}

	return false
}
