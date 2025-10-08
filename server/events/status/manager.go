package status

import (
	"errors"

	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/logging"
)

//go:generate pegomock generate github.com/runatlantis/atlantis/server/events/status --package mocks -o mocks/mock_status_manager.go StatusManager

// StatusManager provides high-level status management with policy-aware decisions
type StatusManager interface {
	// High-level operations that handle policy decisions internally
	HandleCommandStart(ctx *command.Context, cmdName command.Name) error
	HandleCommandEnd(ctx *command.Context, cmdName command.Name, result *command.Result) error
	HandleNoProjectsFound(ctx *command.Context, cmdName command.Name) error

	// Direct status operations (bypass policy)
	SetPending(ctx *command.Context, cmdName command.Name) error
	SetSuccess(ctx *command.Context, cmdName command.Name, numSuccess, numTotal int) error
	SetFailure(ctx *command.Context, cmdName command.Name, err error) error

	// Status clearing operations
	ClearAllStatuses(ctx *command.Context) error
	ClearStatusForCommand(ctx *command.Context, cmdName command.Name) error

	// Status querying (future enhancement)
	GetCurrentStatus(repo models.Repo, pull models.PullRequest) (*StatusState, error)
}

// DefaultStatusManager implements StatusManager with policy-aware status decisions
type DefaultStatusManager struct {
	CommitStatusUpdater CommitStatusUpdater
	Policy              StatusPolicy
	Logger              logging.SimpleLogging
}

// NewStatusManager creates a new StatusManager with the given policy and updater
func NewStatusManager(updater CommitStatusUpdater, policy StatusPolicy, logger logging.SimpleLogging) StatusManager {
	return &DefaultStatusManager{
		CommitStatusUpdater: updater,
		Policy:              policy,
		Logger:              logger,
	}
}

// HandleCommandStart handles the start of a command execution
func (s *DefaultStatusManager) HandleCommandStart(ctx *command.Context, cmdName command.Name) error {
	decision := s.Policy.DecideOnStart(ctx, cmdName)
	return s.executeDecision(ctx, cmdName, decision)
}

// HandleCommandEnd handles the completion of a command execution
func (s *DefaultStatusManager) HandleCommandEnd(ctx *command.Context, cmdName command.Name, result *command.Result) error {
	decision := s.Policy.DecideOnEnd(ctx, cmdName, result)
	return s.executeDecision(ctx, cmdName, decision)
}

// HandleNoProjectsFound handles the case when no projects are found
func (s *DefaultStatusManager) HandleNoProjectsFound(ctx *command.Context, cmdName command.Name) error {
	decision := s.Policy.DecideOnNoProjects(ctx, cmdName)
	return s.executeDecision(ctx, cmdName, decision)
}

// SetPending sets a pending status directly (bypasses policy)
func (s *DefaultStatusManager) SetPending(ctx *command.Context, cmdName command.Name) error {
	ctx.Log.Debug("setting pending status for %s", cmdName.String())
	return s.CommitStatusUpdater.UpdateCombined(ctx.Log, ctx.Pull.BaseRepo, ctx.Pull, models.PendingCommitStatus, cmdName)
}

// SetSuccess sets a success status directly (bypasses policy)
func (s *DefaultStatusManager) SetSuccess(ctx *command.Context, cmdName command.Name, numSuccess, numTotal int) error {
	ctx.Log.Debug("setting success status for %s (%d/%d)", cmdName.String(), numSuccess, numTotal)
	return s.CommitStatusUpdater.UpdateCombinedCount(ctx.Log, ctx.Pull.BaseRepo, ctx.Pull, models.SuccessCommitStatus, cmdName, numSuccess, numTotal)
}

// SetFailure sets a failure status directly (bypasses policy)
func (s *DefaultStatusManager) SetFailure(ctx *command.Context, cmdName command.Name, err error) error {
	ctx.Log.Debug("setting failure status for %s: %s", cmdName.String(), err.Error())
	return s.CommitStatusUpdater.UpdateCombined(ctx.Log, ctx.Pull.BaseRepo, ctx.Pull, models.FailedCommitStatus, cmdName)
}

// ClearAllStatuses clears all statuses for the PR
func (s *DefaultStatusManager) ClearAllStatuses(ctx *command.Context) error {
	ctx.Log.Debug("clearing all statuses")
	// Clear each command type individually
	commands := []command.Name{command.Plan, command.PolicyCheck, command.Apply}
	for _, cmd := range commands {
		if err := s.ClearStatusForCommand(ctx, cmd); err != nil {
			ctx.Log.Warn("failed to clear status for %s: %s", cmd.String(), err)
		}
	}
	return nil
}

// ClearStatusForCommand clears status for a specific command
func (s *DefaultStatusManager) ClearStatusForCommand(ctx *command.Context, cmdName command.Name) error {
	ctx.Log.Debug("clearing status for %s", cmdName.String())
	// Set success with 0/0 to effectively "clear" the status
	return s.CommitStatusUpdater.UpdateCombinedCount(ctx.Log, ctx.Pull.BaseRepo, ctx.Pull, models.SuccessCommitStatus, cmdName, 0, 0)
}

// GetCurrentStatus returns the current status state from VCS
// This would be used for status reconciliation and avoiding duplicate updates
func (s *DefaultStatusManager) GetCurrentStatus(repo models.Repo, pull models.PullRequest) (*StatusState, error) {
	// TODO: Query actual status from VCS provider (GitHub/GitLab/etc)
	// This should return current pending/success/failure state for each command type
	// Implementation would depend on VCS provider and might use existing VCS client
	return nil, errors.New("GetCurrentStatus not implemented - status querying from VCS not yet supported")
}

// executeDecision executes a status decision
func (s *DefaultStatusManager) executeDecision(ctx *command.Context, cmdName command.Name, decision StatusDecision) error {
	switch decision.Operation {
	case OperationSet:
		ctx.Log.Debug("status decision: set - %s", decision.Reason)
		return s.CommitStatusUpdater.UpdateCombinedCount(ctx.Log, ctx.Pull.BaseRepo, ctx.Pull, decision.Status, cmdName, decision.NumSuccess, decision.NumTotal)

	case OperationClear:
		ctx.Log.Debug("status decision: clear - %s", decision.Reason)
		return s.ClearStatusForCommand(ctx, cmdName)

	case OperationSilence:
		ctx.Log.Debug("status decision: silence - %s (%s)", decision.Reason, decision.SilenceType)
		// Do nothing - this is the silence behavior
		return nil

	default:
		ctx.Log.Warn("unknown status operation: %d", decision.Operation)
		return nil
	}
}
