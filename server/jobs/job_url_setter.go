package jobs

import (
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/models"
)

//go:generate pegomock generate --package mocks -o mocks/mock_project_job_url_generator.go ProjectJobURLGenerator

// ProjectJobURLGenerator generates urls to view project's progress.
type ProjectJobURLGenerator interface {
	GenerateProjectJobURL(p command.ProjectContext) (string, error)
}

//go:generate pegomock generate --package mocks -o mocks/mock_project_status_updater.go ProjectStatusUpdater

type ProjectStatusUpdater interface {
	// UpdateProject sets the commit status for the project represented by
	// ctx.
	UpdateProject(ctx command.ProjectContext, cmdName command.Name, status models.CommitStatus, url string, result *command.ProjectResult) error
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

func (j *JobURLSetter) SetJobURLWithStatus(ctx command.ProjectContext, cmdName command.Name, status models.CommitStatus, result *command.ProjectResult) error {
	url, err := j.projectJobURLGenerator.GenerateProjectJobURL(ctx)

	if err != nil {
		return err
	}
	return j.projectStatusUpdater.UpdateProject(ctx, cmdName, status, url, result)
}
