package event

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	"github.com/runatlantis/atlantis/server/core/config/valid"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/logging"
	"github.com/runatlantis/atlantis/server/lyft/feature"
	contextInternal "github.com/runatlantis/atlantis/server/neptune/gateway/context"
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

const (
	Deprecated = "deprecated"
	Destroy    = "-destroy"
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
	SignalWorkflow(ctx context.Context, workflowID string, runID string, signalName string, arg interface{}) error
}

type scheduler interface {
	Schedule(ctx context.Context, f sync.Executor) error
}

type rootConfigBuilder interface {
	Build(ctx context.Context, event Push) ([]*valid.MergedProjectCfg, error)
}

type PushHandler struct {
	Allocator         feature.Allocator
	Scheduler         scheduler
	TemporalClient    signaler
	Logger            logging.Logger
	RootConfigBuilder rootConfigBuilder
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
	rootCfgs, err := p.RootConfigBuilder.Build(ctx, event)
	if err != nil {
		return errors.Wrap(err, "generating roots")
	}
	for _, rootCfg := range rootCfgs {
		ctx = context.WithValue(ctx, contextInternal.ProjectKey, rootCfg.Name)
		run, err := p.startWorkflow(ctx, event, rootCfg)
		if err != nil {
			return errors.Wrap(err, "signalling workflow")
		}

		p.Logger.InfoContext(ctx, "Signaled workflow.", map[string]interface{}{
			"workflow-id": run.GetID(), "run-id": run.GetRunID(),
		})
	}
	return nil
}

func (p *PushHandler) startWorkflow(ctx context.Context, event Push, rootCfg *valid.MergedProjectCfg) (client.WorkflowRun, error) {
	options := client.StartWorkflowOptions{TaskQueue: workflows.DeployTaskQueue}

	var tfVersion string
	if rootCfg.TerraformVersion != nil {
		tfVersion = rootCfg.TerraformVersion.String()
	}

	run, err := p.TemporalClient.SignalWithStartWorkflow(
		ctx,
		fmt.Sprintf("%s||%s", event.Repo.FullName, rootCfg.Name),
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
				HeadCommit: workflows.HeadCommit{
					Ref: workflows.Ref{
						Name: event.Ref.Name,
						Type: string(event.Ref.Type),
					},
				},
			},
			Root: workflows.Root{
				Name: rootCfg.Name,
				Plan: workflows.Job{
					Steps: p.generateSteps(rootCfg.DeploymentWorkflow.Plan.Steps),
				},
				Apply: workflows.Job{
					Steps: p.generateSteps(rootCfg.DeploymentWorkflow.Apply.Steps),
				},
				RepoRelPath: rootCfg.RepoRelDir,
				TfVersion:   tfVersion,
				PlanMode:    p.generatePlanMode(rootCfg),
			},
		},
	)
	return run, err
}

func (p *PushHandler) generatePlanMode(cfg *valid.MergedProjectCfg) workflows.PlanMode {
	t, ok := cfg.Tags[Deprecated]
	if ok && t == Destroy {
		return workflows.DestroyPlanMode
	}

	return workflows.NormalPlanMode
}

func (p *PushHandler) generateSteps(steps []valid.Step) []workflows.Step {
	// NOTE: for deployment workflows, we won't support command level user requests for log level output verbosity
	var workflowSteps []workflows.Step
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
