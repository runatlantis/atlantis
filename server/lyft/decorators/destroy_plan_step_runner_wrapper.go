package decorators

import (
	"github.com/runatlantis/atlantis/server/events"
	"github.com/runatlantis/atlantis/server/events/models"
)

const Deprecated = "deprecated"
const Destroy = "-destroy"

type DestroyPlanStepRunnerWrapper struct {
	events.StepRunner
}

func (d *DestroyPlanStepRunnerWrapper) Run(ctx models.ProjectCommandContext, extraArgs []string, path string, envs map[string]string) (string, error) {
	// DestroyPlan tag is true when the Terraform client should construct a destroy plan given a repo config.
	if ctx.Tags[Deprecated] == Destroy {
		extraArgs = append(extraArgs, Destroy)
	}
	return d.StepRunner.Run(ctx, extraArgs, path, envs)
}
