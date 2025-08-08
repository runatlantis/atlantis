package events

import (
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/logging"
	tally "github.com/uber-go/tally/v4"
)

const (
	baseCloneDirName = ".base"
)

// baseLocks provides per-PR locking for base clone operations
var baseLocks sync.Map

// WorkspaceCopyOptimizer wraps FileWorkspace with copy-based optimization
type WorkspaceCopyOptimizer struct {
	*FileWorkspace
	enabled bool
	scope   tally.Scope
}

// NewWorkspaceCopyOptimizer creates a workspace with copy optimization
func NewWorkspaceCopyOptimizer(base *FileWorkspace, enabled bool, scope tally.Scope) *WorkspaceCopyOptimizer {
	if scope == nil {
		scope = tally.NoopScope
	}
	return &WorkspaceCopyOptimizer{
		FileWorkspace: base,
		enabled:       enabled,
		scope:         scope.SubScope("workspace_copy_optimization"),
	}
}

// Clone implements optimized cloning with copy strategy and fallback
func (w *WorkspaceCopyOptimizer) Clone(logger logging.SimpleLogging, headRepo models.Repo, p models.PullRequest, workspace string) (string, error) {
	timer := w.scope.Timer("clone_duration").Start()
	defer timer.Stop()

	// Fall back to original implementation if optimization disabled
	if !w.enabled {
		w.scope.Counter("fallback_disabled").Inc(1)
		return w.FileWorkspace.Clone(logger, headRepo, p, workspace)
	}

	workspaceDir := w.cloneDir(p.BaseRepo, p, workspace)
	baseDir := w.baseCloneDir(p.BaseRepo, p)

	startTime := time.Now()
	defer func() {
		w.scope.Timer("total_clone_time").Record(time.Since(startTime))
	}()

	// Step 1: Ensure base clone exists and is current
	baseTimer := w.scope.Timer("base_clone_duration").Start()
	err := w.ensureBaseClone(logger, headRepo, p, baseDir)
	baseTimer.Stop()

	if err != nil {
		logger.Warn("Base clone failed, falling back to direct clone: %v", err)
		w.scope.Counter("fallback_base_clone_error").Inc(1)
		return w.FileWorkspace.Clone(logger, headRepo, p, workspace)
	}

	// Step 2: Create or update workspace from base clone
	copyTimer := w.scope.Timer("workspace_copy_duration").Start()
	err = w.ensureWorkspaceFromBase(logger, baseDir, workspaceDir, p)
	copyTimer.Stop()

	if err != nil {
		logger.Warn("Workspace copy failed, falling back to direct clone: %v", err)
		w.scope.Counter("fallback_copy_error").Inc(1)
		return w.FileWorkspace.Clone(logger, headRepo, p, workspace)
	}

	w.scope.Counter("successful_copy_optimization").Inc(1)
	logger.Info("Successfully created workspace %s using copy optimization", workspace)
	return workspaceDir, nil
}

// baseCloneDir returns the path to the base clone for a PR
func (w *WorkspaceCopyOptimizer) baseCloneDir(r models.Repo, p models.PullRequest) string {
	return filepath.Join(w.repoPullDir(r, p), baseCloneDirName)
}

// ensureBaseClone ensures the base clone exists and is at the correct commit
func (w *WorkspaceCopyOptimizer) ensureBaseClone(logger logging.SimpleLogging, headRepo models.Repo, p models.PullRequest, baseDir string) error {
	// Use per-PR locking to prevent concurrent base clone operations
	lockKey := w.repoPullDir(p.BaseRepo, p)
	value, _ := baseLocks.LoadOrStore(lockKey, new(sync.Mutex))
	mutex := value.(*sync.Mutex)
	mutex.Lock()
	defer mutex.Unlock()

	// Check if base clone exists and is at correct commit
	if w.isBaseCloneValid(logger, baseDir, p) {
		logger.Debug("Base clone at %s is valid, skipping clone", baseDir)
		w.scope.Counter("base_clone_cache_hit").Inc(1)
		return nil
	}

	logger.Info("Creating/updating base clone at %s", baseDir)
	w.scope.Counter("base_clone_cache_miss").Inc(1)

	// Remove existing base clone if it exists
	if err := os.RemoveAll(baseDir); err != nil {
		return errors.Wrapf(err, "removing existing base clone at %s", baseDir)
	}

	// Create base clone using existing forceClone logic
	c := wrappedGitContext{baseDir, headRepo, p}
	return w.forceClone(logger, c)
}

// isBaseCloneValid checks if base clone exists and is at correct commit
func (w *WorkspaceCopyOptimizer) isBaseCloneValid(logger logging.SimpleLogging, baseDir string, p models.PullRequest) bool {
	// Check if directory exists
	if _, err := os.Stat(baseDir); err != nil {
		return false
	}

	// Use same logic as original Clone method to verify commit
	pullHead := "HEAD"
	if w.CheckoutMerge && p.Num > 0 {
		pullHead = "HEAD^2"
	}

	// Execute git rev-parse to get current commit
	revParseCmd := exec.Command("git", "rev-parse", pullHead) // #nosec
	revParseCmd.Dir = baseDir
	outputRevParseCmd, err := revParseCmd.CombinedOutput()
	if err != nil {
		logger.Debug("Base clone commit verification failed: %s", err)
		return false
	}

	currCommit := strings.TrimSpace(string(outputRevParseCmd))
	// We're prefix matching here because BitBucket doesn't give us the full commit
	if strings.HasPrefix(currCommit, p.HeadCommit) {
		logger.Debug("Base clone is at correct commit %q", p.HeadCommit)
		return true
	}

	logger.Debug("Base clone commit mismatch, wanted %q got %q", p.HeadCommit, currCommit)
	return false
}

// ensureWorkspaceFromBase creates workspace by copying from base clone
func (w *WorkspaceCopyOptimizer) ensureWorkspaceFromBase(logger logging.SimpleLogging, baseDir, workspaceDir string, p models.PullRequest) error {
	// Check if workspace already exists and is current
	if w.isWorkspaceCurrent(logger, workspaceDir, p) {
		logger.Debug("Workspace %s is current, skipping copy", workspaceDir)
		w.scope.Counter("workspace_cache_hit").Inc(1)
		return nil
	}

	logger.Info("Creating workspace %s by copying from %s", workspaceDir, baseDir)
	w.scope.Counter("workspace_cache_miss").Inc(1)

	// Remove existing workspace
	if err := os.RemoveAll(workspaceDir); err != nil {
		return errors.Wrapf(err, "removing existing workspace %s", workspaceDir)
	}

	// Copy base to workspace
	copyStartTime := time.Now()
	if err := w.copyDirectory(baseDir, workspaceDir); err != nil {
		return errors.Wrapf(err, "copying %s to %s", baseDir, workspaceDir)
	}
	w.scope.Timer("filesystem_copy_time").Record(time.Since(copyStartTime))

	return nil
}

// isWorkspaceCurrent checks if workspace exists and is at correct commit
func (w *WorkspaceCopyOptimizer) isWorkspaceCurrent(logger logging.SimpleLogging, workspaceDir string, p models.PullRequest) bool {
	// Check if directory exists
	if _, err := os.Stat(workspaceDir); err != nil {
		return false
	}

	// Use same commit verification logic as base clone
	pullHead := "HEAD"
	if w.CheckoutMerge && p.Num > 0 {
		pullHead = "HEAD^2"
	}

	// Execute git rev-parse to get current commit
	revParseCmd := exec.Command("git", "rev-parse", pullHead) // #nosec
	revParseCmd.Dir = workspaceDir
	outputRevParseCmd, err := revParseCmd.CombinedOutput()
	if err != nil {
		return false
	}

	currCommit := strings.TrimSpace(string(outputRevParseCmd))
	return strings.HasPrefix(currCommit, p.HeadCommit)
}

// copyDirectory recursively copies a directory and all its contents with optimizations
func (w *WorkspaceCopyOptimizer) copyDirectory(src, dst string) error {
	// Get source directory info
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	// Create destination directory with same permissions
	if err := os.MkdirAll(dst, srcInfo.Mode()); err != nil {
		return err
	}

	// Read source directory contents
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	// Copy each entry
	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			// Recursively copy subdirectory
			if err := w.copyDirectory(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			// Copy file with optimizations
			if err := w.copyFile(srcPath, dstPath); err != nil {
				return err
			}
		}
	}

	return nil
}

// copyFile copies a single file preserving permissions
func (w *WorkspaceCopyOptimizer) copyFile(src, dst string) error {
	// Open source file
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	// Get source file info for permissions
	srcInfo, err := srcFile.Stat()
	if err != nil {
		return err
	}

	// Create destination file with same permissions
	dstFile, err := os.OpenFile(dst, os.O_RDWR|os.O_CREATE|os.O_TRUNC, srcInfo.Mode())
	if err != nil {
		return err
	}
	defer dstFile.Close()

	// Copy
	_, err = io.Copy(dstFile, srcFile)
	return err
}


// Delete extends the base implementation to clean up base clone
func (w *WorkspaceCopyOptimizer) Delete(logger logging.SimpleLogging, r models.Repo, p models.PullRequest) error {
	// Delete entire PR directory including base clone
	repoPullDir := w.repoPullDir(r, p)
	logger.Info("Deleting repo pull directory with copy optimization: " + repoPullDir)
	w.scope.Counter("full_pr_delete").Inc(1)
	return os.RemoveAll(repoPullDir)
}

// DeleteForWorkspace extends the base implementation
func (w *WorkspaceCopyOptimizer) DeleteForWorkspace(logger logging.SimpleLogging, r models.Repo, p models.PullRequest, workspace string) error {
	workspaceDir := w.cloneDir(r, p, workspace)
	logger.Info("Deleting workspace directory with copy optimization: " + workspaceDir)
	w.scope.Counter("workspace_delete").Inc(1)
	return os.RemoveAll(workspaceDir)
}

// GetStats returns optimization statistics
func (w *WorkspaceCopyOptimizer) GetStats() map[string]interface{} {
	return map[string]interface{}{
		"enabled":              w.enabled,
		"active_base_locks":    w.getBaseLockCount(),
		"implementation":       "copy_optimization_v1",
		"base_clone_dir_name":  baseCloneDirName,
		"features": map[string]bool{
			"efficient_copy":     true,
			"commit_validation":  true,
			"cache_validation":   true,
			"fallback_support":   true,
			"metrics_enabled":    w.scope != tally.NoopScope,
		},
	}
}

// getBaseLockCount returns the number of active base clone locks
func (w *WorkspaceCopyOptimizer) getBaseLockCount() int {
	count := 0
	baseLocks.Range(func(key, value interface{}) bool {
		count++
		return true
	})
	return count
}

// IsOptimizationEnabled returns whether copy optimization is enabled
func (w *WorkspaceCopyOptimizer) IsOptimizationEnabled() bool {
	return w.enabled
}

// SetOptimizationEnabled allows toggling optimization at runtime
func (w *WorkspaceCopyOptimizer) SetOptimizationEnabled(enabled bool) {
	w.enabled = enabled
	if enabled {
		w.scope.Counter("optimization_enabled").Inc(1)
	} else {
		w.scope.Counter("optimization_disabled").Inc(1)
	}
}
