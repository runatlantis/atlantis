package event

import (
	"context"
	"github.com/pkg/errors"
	"github.com/runatlantis/atlantis/server/core/config/valid"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/logging"
	"github.com/runatlantis/atlantis/server/lyft/feature"
	"github.com/runatlantis/atlantis/server/neptune/gateway/sync"
	"github.com/runatlantis/atlantis/server/neptune/workflows"
	"github.com/runatlantis/atlantis/server/vcs"
	"go.temporal.io/sdk/client"
)

type PushAction string

const (
	DeletedAction PushAction = "deleted"
	CreatedAction PushAction = "created"
	UpdatedAction PushAction = "updated"
)

type Push struct {
	Repo              models.Repo
	Ref               vcs.Ref
	Sha               string
	Sender            vcs.User
	InstallationToken int64
	Action            PushAction
}

type signaler interface {
	SignalWithStartWorkflow(ctx context.Context, workflowID string, signalName string, signalArg interface{},
		options client.StartWorkflowOptions, workflow interface{}, workflowArgs ...interface{}) (client.WorkflowRun, error)
}

type scheduler interface {
	Schedule(ctx context.Context, f sync.Executor) error
}

const defaultWorkspace = "default"

type PushHandler struct {
	Allocator      feature.Allocator
	Scheduler      scheduler
	TemporalClient signaler
	Logger         logging.Logger
	GlobalCfg      valid.GlobalCfg
}

func (p *PushHandler) Handle(ctx context.Context, event Push) error {
	shouldAllocate, err := p.Allocator.ShouldAllocate(feature.PlatformMode, feature.FeatureContext{
		RepoName: event.Repo.FullName,
	})

	if err != nil {
		p.Logger.ErrorContext(ctx, "unable to allocate platformmode")
		return nil
	}

	if !shouldAllocate {
		p.Logger.DebugContext(ctx, "handler not configured for allocation")
		return nil
	}

	if event.Ref.Type != vcs.BranchRef || event.Ref.Name != event.Repo.DefaultBranch {
		p.Logger.DebugContext(ctx, "dropping event for unexpected ref")
		return nil
	}

	if event.Action == DeletedAction {
		p.Logger.WarnContext(ctx, "ref was deleted, resources might still exist")
		return nil
	}

	return p.Scheduler.Schedule(ctx, func(ctx context.Context) error {
		return p.handle(ctx, event)
	})
}

func (p *PushHandler) handle(ctx context.Context, event Push) error {
	options := client.StartWorkflowOptions{TaskQueue: workflows.DeployTaskQueue}

	// TODO: clone and build project config

	run, err := p.TemporalClient.SignalWithStartWorkflow(
		ctx,

		// TODO: name should include root name as well
		event.Repo.FullName,
		workflows.DeployNewRevisionSignalID,
		workflows.DeployNewRevisionSignalRequest{
			Revision: event.Sha,
		},
		options,
		workflows.Deploy,

		// TODO: add other request params as we support them
		workflows.DeployRequest{
			Repository: workflows.Repo{
				URL:      event.Repo.CloneURL,
				FullName: event.Repo.FullName,
				Name:     event.Repo.Name,
				Owner:    event.Repo.Owner,
				Credentials: workflows.AppCredentials{
					InstallationToken: event.InstallationToken,
				},
			},
			Root: workflows.Root{
				Name: "TODO",
				Plan: workflows.Job{
					Steps: p.generatePlanSteps(event.Repo.ID()),
				},
				Apply: workflows.Job{
					Steps: p.generateApplySteps(event.Repo.ID()),
				},
			},
		},
	)

	if err != nil {
		return errors.Wrap(err, "signalling workflow")
	}

	p.Logger.InfoContext(ctx, "Signaled workflow.", map[string]interface{}{
		"workflow-id": run.GetID(), "run-id": run.GetRunID(),
	})

	return nil
}

func (p *PushHandler) generatePlanSteps(repoID string) []workflows.Step {
	// NOTE: for deployment workflows, we won't support command level user requests for log level output verbosity
	var workflowSteps []workflows.Step

	// TODO: replace example use of DefaultProjCfg with a project config generator that handles default vs. merged cfg case
	projectConfig := p.GlobalCfg.DefaultProjCfg(p.Logger, repoID, "path", defaultWorkspace)
	steps := projectConfig.Workflow.Plan.Steps
	for _, step := range steps {
		workflowSteps = append(workflowSteps, workflows.Step{
			StepName:    step.StepName,
			ExtraArgs:   step.ExtraArgs,
			RunCommand:  step.RunCommand,
			EnvVarName:  step.EnvVarName,
			EnvVarValue: step.EnvVarValue,
		})
	}
	return workflowSteps
}

func (p *PushHandler) generateApplySteps(repoID string) []workflows.Step {
	var workflowSteps []workflows.Step
	projectConfig := p.GlobalCfg.DefaultProjCfg(p.Logger, repoID, "path", defaultWorkspace)
	steps := projectConfig.Workflow.Apply.Steps
	for _, step := range steps {
		workflowSteps = append(workflowSteps, workflows.Step{
			StepName:    step.StepName,
			ExtraArgs:   step.ExtraArgs,
			RunCommand:  step.RunCommand,
			EnvVarName:  step.EnvVarName,
			EnvVarValue: step.EnvVarValue,
		})
	}
	return workflowSteps
}
