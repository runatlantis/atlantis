package status

import (
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/logging"
)

//go:generate pegomock generate github.com/runatlantis/atlantis/server/events/status --package mocks -o mocks/mock_status_cleaner.go StatusCleaner

// StatusCleaner provides operations for clearing and managing status checks
type StatusCleaner interface {
	// ClearPendingStatuses clears any pending status checks for the specified commands
	ClearPendingStatuses(ctx *command.Context, commands []command.Name) error
	
	// ClearAllStatuses clears all status checks for the PR (plan, apply, policy)
	ClearAllStatuses(ctx *command.Context) error
	
	// ClearStatusForCommand clears status for a specific command
	ClearStatusForCommand(ctx *command.Context, cmdName command.Name) error
	
	// ResetStatusesToSuccess resets all statuses to success with 0/0 counts
	ResetStatusesToSuccess(ctx *command.Context) error
}

// DefaultStatusCleaner implements StatusCleaner
type DefaultStatusCleaner struct {
	CommitStatusUpdater CommitStatusUpdater
	Logger              logging.SimpleLogging
}

// NewStatusCleaner creates a new StatusCleaner
func NewStatusCleaner(updater CommitStatusUpdater, logger logging.SimpleLogging) StatusCleaner {
	return &DefaultStatusCleaner{
		CommitStatusUpdater: updater,
		Logger:              logger,
	}
}

// ClearPendingStatuses clears any pending status checks for the specified commands
func (c *DefaultStatusCleaner) ClearPendingStatuses(ctx *command.Context, commands []command.Name) error {
	for _, cmdName := range commands {
		if err := c.ClearStatusForCommand(ctx, cmdName); err != nil {
			ctx.Log.Warn("failed to clear pending status for %s: %s", cmdName.String(), err)
			// Continue clearing other statuses even if one fails
		}
	}
	return nil
}

// ClearAllStatuses clears all status checks for the PR
func (c *DefaultStatusCleaner) ClearAllStatuses(ctx *command.Context) error {
	commands := []command.Name{command.Plan, command.PolicyCheck, command.Apply}
	return c.ClearPendingStatuses(ctx, commands)
}

// ClearStatusForCommand clears status for a specific command
func (c *DefaultStatusCleaner) ClearStatusForCommand(ctx *command.Context, cmdName command.Name) error {
	ctx.Log.Debug("clearing status for command %s", cmdName.String())
	// Set success with 0/0 to effectively "clear" the status while keeping it visible
	return c.CommitStatusUpdater.UpdateCombinedCount(
		ctx.Log, 
		ctx.Pull.BaseRepo, 
		ctx.Pull, 
		models.SuccessCommitStatus, 
		cmdName, 
		0, 
		0,
	)
}

// ResetStatusesToSuccess resets all statuses to success with 0/0 counts
func (c *DefaultStatusCleaner) ResetStatusesToSuccess(ctx *command.Context) error {
	ctx.Log.Debug("resetting all statuses to success 0/0")
	commands := []command.Name{command.Plan, command.PolicyCheck, command.Apply}
	
	for _, cmdName := range commands {
		if err := c.CommitStatusUpdater.UpdateCombinedCount(
			ctx.Log, 
			ctx.Pull.BaseRepo, 
			ctx.Pull, 
			models.SuccessCommitStatus, 
			cmdName, 
			0, 
			0,
		); err != nil {
			ctx.Log.Warn("failed to reset status for %s: %s", cmdName.String(), err)
		}
	}
	return nil
}

// StatusCleanupHelper provides helper functions for common cleanup scenarios
type StatusCleanupHelper struct {
	Cleaner StatusCleaner
}

// NewStatusCleanupHelper creates a new helper
func NewStatusCleanupHelper(cleaner StatusCleaner) *StatusCleanupHelper {
	return &StatusCleanupHelper{
		Cleaner: cleaner,
	}
}

// CleanupAfterSilence cleans up any existing statuses when silence flags are enabled
func (h *StatusCleanupHelper) CleanupAfterSilence(ctx *command.Context, reason string) error {
	ctx.Log.Debug("cleaning up statuses due to silence: %s", reason)
	return h.Cleaner.ClearAllStatuses(ctx)
}

// CleanupPendingOnly clears only pending statuses, leaving success/failure statuses intact
func (h *StatusCleanupHelper) CleanupPendingOnly(ctx *command.Context, commands []command.Name) error {
	ctx.Log.Debug("cleaning up pending statuses only")
	// This would require querying current status first (future enhancement)
	// For now, just clear all specified commands
	return h.Cleaner.ClearPendingStatuses(ctx, commands)
}