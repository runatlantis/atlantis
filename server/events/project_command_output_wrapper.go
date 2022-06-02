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
	JobURLSetter JobURLSetter
	JobCloser    JobCloser
}

func (p *ProjectOutputWrapper) Plan(ctx command.ProjectContext) command.ProjectResult {
	result := p.updateProjectPRStatus(command.Plan, ctx, p.ProjectCommandRunner.Plan)

	// A job needs to be closed if a log streaming job has been registered in our log handler which happens in the
	// step runner when the first tf output for the operation is streamed. However, there's no way to tell is the operation reached
	// the step runner and a job has been registered. So, we just assume that a job exists for now and close it.
	// If the job does not exist, we log an error.
	p.JobCloser.CloseJob(ctx.JobID, ctx.BaseRepo)
	return result
}

func (p *ProjectOutputWrapper) Apply(ctx command.ProjectContext) command.ProjectResult {
	result := p.updateProjectPRStatus(command.Apply, ctx, p.ProjectCommandRunner.Apply)

	// A job needs to be closed if a log streaming job has been registered in our log handler which happens in the
	// step runner when the first tf output for the operation is streamed. However, there's no way to tell is the operation reached
	// the step runner and a job has been registered. So, we just assume that a job exists for now and close it.
	// If the job does not exist, we log an error.
	p.JobCloser.CloseJob(ctx.JobID, ctx.BaseRepo)
	return result
}

func (p *ProjectOutputWrapper) updateProjectPRStatus(commandName command.Name, ctx command.ProjectContext, execute func(ctx command.ProjectContext) command.ProjectResult) command.ProjectResult {
	// Create a PR status to track project's plan status. The status will
	// include a link to view the progress of atlantis plan command in real
	// time
	if err := p.JobURLSetter.SetJobURLWithStatus(ctx, commandName, models.PendingCommitStatus); err != nil {
		ctx.Log.ErrorContext(ctx.RequestCtx, fmt.Sprintf("updating project PR status %v", err))
	}

	// ensures we are differentiating between project level command and overall command
	result := execute(ctx)

	if result.Error != nil || result.Failure != "" {
		if err := p.JobURLSetter.SetJobURLWithStatus(ctx, commandName, models.FailedCommitStatus); err != nil {
			ctx.Log.ErrorContext(ctx.RequestCtx, fmt.Sprintf("updating project PR status %v", err))
		}

		return result
	}

	if err := p.JobURLSetter.SetJobURLWithStatus(ctx, commandName, models.SuccessCommitStatus); err != nil {
		ctx.Log.ErrorContext(ctx.RequestCtx, fmt.Sprintf("updating project PR status %v", err))
	}

	return result
}
