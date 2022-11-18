package event

import (
	"context"
	"github.com/pkg/errors"
	"github.com/runatlantis/atlantis/server/core/config/valid"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/logging"
	contextInternal "github.com/runatlantis/atlantis/server/neptune/context"
	"github.com/runatlantis/atlantis/server/neptune/workflows"
	"go.temporal.io/sdk/client"
)

type deploySignaler interface {
	SignalWithStartWorkflow(ctx context.Context, rootCfg *valid.MergedProjectCfg, rootDeployOptions RootDeployOptions) (client.WorkflowRun, error)
	SignalWorkflow(ctx context.Context, workflowID string, runID string, signalName string, arg interface{}) error
}

type rootConfigBuilder interface {
	Build(ctx context.Context, repo models.Repo, branch string, sha string, installationToken int64, builderOptions BuilderOptions) ([]*valid.MergedProjectCfg, error)
}

type RootDeployer struct {
	Logger            logging.Logger
	RootConfigBuilder rootConfigBuilder
	DeploySignaler    deploySignaler
}

type RootDeployOptions struct {
	Repo              models.Repo
	Branch            string
	Revision          string
	Sender            models.User
	InstallationToken int64
	BuilderOptions    BuilderOptions
	Trigger           workflows.Trigger
	Rerun             bool
}

func (d *RootDeployer) Deploy(ctx context.Context, deployOptions RootDeployOptions) error {
	rootCfgs, err := d.RootConfigBuilder.Build(ctx, deployOptions.Repo, deployOptions.Branch, deployOptions.Revision, deployOptions.InstallationToken, deployOptions.BuilderOptions)
	if err != nil {
		return errors.Wrap(err, "generating roots")
	}
	for _, rootCfg := range rootCfgs {
		c := context.WithValue(ctx, contextInternal.ProjectKey, rootCfg.Name)
		if rootCfg.WorkflowMode != valid.PlatformWorkflowMode {
			d.Logger.DebugContext(c, "root is not configured for platform mode, skipping...")
			continue
		}
		run, err := d.DeploySignaler.SignalWithStartWorkflow(c, rootCfg, deployOptions)
		if err != nil {
			return errors.Wrap(err, "signalling workflow")
		}

		d.Logger.InfoContext(c, "Signaled workflow.", map[string]interface{}{
			"workflow-id": run.GetID(), "run-id": run.GetRunID(),
		})
	}
	return nil
}
