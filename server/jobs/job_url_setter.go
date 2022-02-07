package jobs

import "github.com/runatlantis/atlantis/server/events/models"

//go:generate pegomock generate -m --use-experimental-model-gen --package mocks -o mocks/mock_project_job_url_generator.go ProjectJobURLGenerator

// ProjectJobURLGenerator generates urls to view project's progress.
type ProjectJobURLGenerator interface {
	GenerateProjectJobURL(p models.ProjectCommandContext) (string, error)
}

//go:generate pegomock generate -m --use-experimental-model-gen --package mocks -o mocks/mock_project_status_updater.go ProjectStatusUpdater

type ProjectStatusUpdater interface {
	// UpdateProject sets the commit status for the project represented by
	// ctx.
	UpdateProject(ctx models.ProjectCommandContext, cmdName models.CommandName, status models.CommitStatus, url string) error
}

type JobURLSetter struct {
	projectJobURLGenerator ProjectJobURLGenerator
	projectStatusUpdater   ProjectStatusUpdater
}

func NewJobURLSetter(projectJobURLGenerator ProjectJobURLGenerator, projectStatusUpdater ProjectStatusUpdater) *JobURLSetter {
	return &JobURLSetter{
		projectJobURLGenerator: projectJobURLGenerator,
		projectStatusUpdater:   projectStatusUpdater,
	}
}

func (j *JobURLSetter) SetJobURLWithStatus(ctx models.ProjectCommandContext, cmdName models.CommandName, status models.CommitStatus) error {
	url, err := j.projectJobURLGenerator.GenerateProjectJobURL(ctx)

	if err != nil {
		return err
	}
	return j.projectStatusUpdater.UpdateProject(ctx, cmdName, status, url)
}
