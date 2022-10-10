package events

import (
	"fmt"

	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/models"
)

// ProjectOutputWrapper is a decorator that creates a new PR status check per project.
// The status contains a url that outputs current progress of the terraform plan/apply command.
type ProjectOutputWrapper struct {
	ProjectCommandRunner

	ProjectStatusUpdater command.StatusUpdater
}

func (p *ProjectOutputWrapper) Plan(ctx command.ProjectContext) command.ProjectResult {
	statusID, err := p.ProjectStatusUpdater.UpdateProjectStatus(ctx, models.PendingCommitStatus)
	if err != nil {
		ctx.Log.ErrorContext(ctx.RequestCtx, fmt.Sprintf("updating project PR status %v", err))
	}

	// Write the statusID to project context which is used by the command runners when making consecutive status updates
	// Noop when checks is not enabled
	ctx.StatusID = statusID

	result := p.ProjectCommandRunner.Plan(ctx)
	if result.Error != nil || result.Failure != "" {
		if _, err := p.ProjectStatusUpdater.UpdateProjectStatus(ctx, models.FailedCommitStatus); err != nil {
			ctx.Log.ErrorContext(ctx.RequestCtx, fmt.Sprintf("updating project PR status %v", err))
		}

		return result
	}

	if _, err := p.ProjectStatusUpdater.UpdateProjectStatus(ctx, models.SuccessCommitStatus); err != nil {
		ctx.Log.ErrorContext(ctx.RequestCtx, fmt.Sprintf("updating project PR status %v", err))
	}
	return result
}

func (p *ProjectOutputWrapper) Apply(ctx command.ProjectContext) command.ProjectResult {
	statusID, err := p.ProjectStatusUpdater.UpdateProjectStatus(ctx, models.PendingCommitStatus)
	if err != nil {
		ctx.Log.ErrorContext(ctx.RequestCtx, fmt.Sprintf("updating project PR status %v", err))
	}

	// Write the statusID to project context which is used by the command runners when making consecutive status updates
	// Noop when checks is not enabled
	ctx.StatusID = statusID

	result := p.ProjectCommandRunner.Apply(ctx)
	if result.Error != nil || result.Failure != "" {
		if _, err := p.ProjectStatusUpdater.UpdateProjectStatus(ctx, models.FailedCommitStatus); err != nil {
			ctx.Log.ErrorContext(ctx.RequestCtx, fmt.Sprintf("updating project PR status %v", err))
		}

		return result
	}

	if _, err := p.ProjectStatusUpdater.UpdateProjectStatus(ctx, models.SuccessCommitStatus); err != nil {
		ctx.Log.ErrorContext(ctx.RequestCtx, fmt.Sprintf("updating project PR status %v", err))
	}

	return result
}
