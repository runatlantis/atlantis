package runtime

import (
	"os"
	"path/filepath"

	"github.com/hashicorp/go-version"
	"github.com/pkg/errors"
	"github.com/runatlantis/atlantis/server/events/command"
)

const minimumShowTfVersion string = "0.12.0"

func NewShowStepRunner(executor TerraformExec, defaultTFVersion *version.Version) (Runner, error) {
	showStepRunner := &showStepRunner{
		terraformExecutor: executor,
		defaultTFVersion:  defaultTFVersion,
	}
	remotePlanRunner := NullRunner{}
	runner := NewPlanTypeStepRunnerDelegate(showStepRunner, remotePlanRunner)
	return NewMinimumVersionStepRunnerDelegate(minimumShowTfVersion, defaultTFVersion, runner)
}

// showStepRunner runs terraform show on an existing plan file and outputs it to a json file
type showStepRunner struct {
	terraformExecutor TerraformExec
	defaultTFVersion  *version.Version
}

func (p *showStepRunner) Run(ctx command.ProjectContext, extraArgs []string, path string, envs map[string]string) (string, error) {
	tfVersion := p.defaultTFVersion
	if ctx.TerraformVersion != nil {
		tfVersion = ctx.TerraformVersion
	}

	planFile := filepath.Join(path, GetPlanFilename(ctx.Workspace, ctx.ProjectName))
	showResultFile := filepath.Join(path, ctx.GetShowResultFileName())

	output, err := p.terraformExecutor.RunCommandWithVersion(
		ctx,
		path,
		[]string{"show", "-json", filepath.Clean(planFile)},
		envs,
		tfVersion,
		ctx.Workspace,
	)

	if err != nil {
		return "", errors.Wrap(err, "running terraform show")
	}

	if err := os.WriteFile(showResultFile, []byte(output), os.ModePerm); err != nil {
		return "", errors.Wrap(err, "writing terraform show result")
	}

	return output, nil
}
