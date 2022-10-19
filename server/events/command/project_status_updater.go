package command

import (
	"context"
	"fmt"

	"github.com/runatlantis/atlantis/server/events/models"
)

//go:generate pegomock generate -m --use-experimental-model-gen --package mocks -o mocks/mock_job_closer.go JobCloser

// Job Closer closes a job by marking op complete and clearing up buffers if logs are successfully persisted
type JobCloser interface {
	CloseJob(ctx context.Context, jobID string, repo models.Repo)
}

//go:generate pegomock generate -m --use-experimental-model-gen --package mocks -o mocks/mock_project_job_url_generator.go ProjectJobURLGenerator

// JobURLGenerator generates urls to view project's progress.
type JobURLGenerator interface {
	GenerateProjectJobURL(jobID string) (string, error)
}

//go:generate pegomock generate -m --use-experimental-model-gen --package mocks -o mocks/mock_project_status_updater.go ProjectStatusUpdater

type projectVCSStatusUpdater interface {
	// UpdateProject sets the commit status for the project represented by
	// ctx.
	UpdateProject(ctx context.Context, projectCtx ProjectContext, cmdName fmt.Stringer, status models.VCSStatus, url string, statusID string) (string, error)
}

type ProjectStatusUpdater struct {
	ProjectJobURLGenerator  JobURLGenerator
	JobCloser               JobCloser
	ProjectVCSStatusUpdater projectVCSStatusUpdater
}

func (p ProjectStatusUpdater) UpdateProjectStatus(ctx ProjectContext, status models.VCSStatus) (string, error) {
	url, err := p.ProjectJobURLGenerator.GenerateProjectJobURL(ctx.JobID)
	if err != nil {
		ctx.Log.ErrorContext(ctx.RequestCtx, fmt.Sprintf("updating project PR status %v", err))
	}
	statusID, err := p.ProjectVCSStatusUpdater.UpdateProject(ctx.RequestCtx, ctx, ctx.CommandName, status, url, ctx.StatusID)

	// Close the Job if the operation is complete
	if status == models.SuccessVCSStatus || status == models.FailedVCSStatus {
		p.JobCloser.CloseJob(ctx.RequestCtx, ctx.JobID, ctx.BaseRepo)
	}
	return statusID, err
}
