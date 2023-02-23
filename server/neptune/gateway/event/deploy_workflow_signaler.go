package event

import (
	"context"
	"fmt"

	"github.com/runatlantis/atlantis/server/core/config/valid"
	"github.com/runatlantis/atlantis/server/neptune/workflows"
	"go.temporal.io/sdk/client"
)

type signaler interface {
	SignalWithStartWorkflow(ctx context.Context, workflowID string, signalName string, signalArg interface{},
		options client.StartWorkflowOptions, workflow interface{}, workflowArgs ...interface{}) (client.WorkflowRun, error)
	SignalWorkflow(ctx context.Context, workflowID string, runID string, signalName string, arg interface{}) error
}

const (
	Deprecated = "deprecated"
	Destroy    = "-destroy"
)

type DeployWorkflowSignaler struct {
	TemporalClient signaler
}

func (d *DeployWorkflowSignaler) SignalWithStartWorkflow(ctx context.Context, rootCfg *valid.MergedProjectCfg, rootDeployOptions RootDeployOptions) (client.WorkflowRun, error) {
	options := client.StartWorkflowOptions{
		TaskQueue: workflows.DeployTaskQueue,
		SearchAttributes: map[string]interface{}{
			"atlantis_repository": rootDeployOptions.Repo.FullName,
			"atlantis_root":       rootCfg.Name,
		},
	}

	repo := rootDeployOptions.Repo
	var tfVersion string
	if rootCfg.TerraformVersion != nil {
		tfVersion = rootCfg.TerraformVersion.String()
	}

	run, err := d.TemporalClient.SignalWithStartWorkflow(
		ctx,
		buildDeployWorkflowID(repo.FullName, rootCfg.Name),
		workflows.DeployNewRevisionSignalID,
		workflows.DeployNewRevisionSignalRequest{
			Revision: rootDeployOptions.Revision,
			InitiatingUser: workflows.User{
				Name: rootDeployOptions.Sender.Username,
			},
			Root: workflows.Root{
				Name: rootCfg.Name,
				Plan: workflows.Job{
					Steps: d.generateSteps(rootCfg.DeploymentWorkflow.Plan.Steps),
				},
				Apply: workflows.Job{
					Steps: d.generateSteps(rootCfg.DeploymentWorkflow.Apply.Steps),
				},
				RepoRelPath:  rootCfg.RepoRelDir,
				WhenModified: rootCfg.WhenModified,
				TfVersion:    tfVersion,
				PlanMode:     d.generatePlanMode(rootCfg),
				Trigger:      rootDeployOptions.Trigger,
				Rerun:        rootDeployOptions.Rerun,
			},
			Repo: workflows.Repo{
				URL:      repo.CloneURL,
				FullName: repo.FullName,
				Name:     repo.Name,
				Owner:    repo.Owner,
				Credentials: workflows.AppCredentials{
					InstallationToken: rootDeployOptions.InstallationToken,
				},
			},
			Tags: rootCfg.Tags,
		},
		options,
		workflows.Deploy,
		workflows.DeployRequest{
			Repo: workflows.DeployRequestRepo{
				FullName: repo.FullName,
			},
			Root: workflows.DeployRequestRoot{
				Name: rootCfg.Name,
			},
		},
	)
	return run, err
}

func buildDeployWorkflowID(repoName string, rootName string) string {
	return fmt.Sprintf("%s||%s", repoName, rootName)
}

func (d *DeployWorkflowSignaler) generateSteps(steps []valid.Step) []workflows.Step {
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

func (d *DeployWorkflowSignaler) generatePlanMode(cfg *valid.MergedProjectCfg) workflows.PlanMode {
	t, ok := cfg.Tags[Deprecated]
	if ok && t == Destroy {
		return workflows.DestroyPlanMode
	}

	return workflows.NormalPlanMode
}

func (d *DeployWorkflowSignaler) SignalWorkflow(ctx context.Context, workflowID string, runID string, signalName string, args interface{}) error {
	return d.TemporalClient.SignalWorkflow(ctx, workflowID, runID, signalName, args)
}
