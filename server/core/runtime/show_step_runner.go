package runtime

import (
	"context"
	"os"
	"path/filepath"

	"github.com/hashicorp/go-version"
	"github.com/pkg/errors"
	"github.com/runatlantis/atlantis/server/events/command"
)

const minimumShowTfVersion string = "0.12.0"

func NewShowStepRunner(executor TerraformExec, defaultTFVersion *version.Version) (Runner, error) {
	runner := &PlanTypeStepRunnerDelegate{
		defaultRunner: &ShowStepRunner{
			TerraformExecutor: executor,
			DefaultTFVersion:  defaultTFVersion,
		},
		remotePlanRunner: NullRunner{},
	}

	return NewMinimumVersionStepRunnerDelegate(minimumShowTfVersion, defaultTFVersion, runner)
}

// ShowStepRunner runs terraform show on an existing plan file and outputs it to a json file
type ShowStepRunner struct {
	TerraformExecutor TerraformExec
	DefaultTFVersion  *version.Version
}

func (p *ShowStepRunner) Run(ctx context.Context, prjCtx command.ProjectContext, extraArgs []string, path string, envs map[string]string) (string, error) {
	tfVersion := p.DefaultTFVersion
	if prjCtx.TerraformVersion != nil {
		tfVersion = prjCtx.TerraformVersion
	}

	planFile := filepath.Join(path, GetPlanFilename(prjCtx.Workspace, prjCtx.ProjectName))
	showResultFile := filepath.Join(path, prjCtx.GetShowResultFileName())

	output, err := p.TerraformExecutor.RunCommandWithVersion(
		ctx,
		prjCtx,
		path,
		[]string{"show", "-json", filepath.Clean(planFile)},
		envs,
		tfVersion,
		prjCtx.Workspace,
	)

	if err != nil {
		return output, errors.Wrap(err, "running terraform show")
	}

	if err := os.WriteFile(showResultFile, []byte(output), os.ModePerm); err != nil {
		return "", errors.Wrap(err, "writing terraform show result")
	}

	// don't return the output if it's successful since this is too large
	return "", nil
}
