package runtime

import (
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/runatlantis/atlantis/server/events/command"
)

func NewPlanTypeStepRunnerDelegate(defaultRunner Runner, remotePlanRunner Runner) Runner {
	return &planTypeStepRunnerDelegate{
		defaultRunner:    defaultRunner,
		remotePlanRunner: remotePlanRunner,
	}
}

// planTypeStepRunnerDelegate delegates based on the type of plan, ie. remote backend which doesn't support certain functions
type planTypeStepRunnerDelegate struct {
	defaultRunner    Runner
	remotePlanRunner Runner
}

func (p *planTypeStepRunnerDelegate) isRemotePlan(planFile string) (bool, error) {
	data, err := os.ReadFile(planFile)

	if err != nil {
		return false, errors.Wrapf(err, "unable to read %s", planFile)
	}

	return IsRemotePlan(data), nil
}

func (p *planTypeStepRunnerDelegate) Run(ctx command.ProjectContext, extraArgs []string, path string, envs map[string]string) (string, error) {
	planFile := filepath.Join(path, GetPlanFilename(ctx.Workspace, ctx.ProjectName))
	remotePlan, err := p.isRemotePlan(planFile)

	if err != nil {
		return "", err
	}

	if remotePlan {
		return p.remotePlanRunner.Run(ctx, extraArgs, path, envs)
	}

	return p.defaultRunner.Run(ctx, extraArgs, path, envs)
}
