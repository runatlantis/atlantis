package runtime

import (
	"os"
	"path/filepath"

	"github.com/hashicorp/go-version"
	"github.com/pkg/errors"
	"github.com/runatlantis/atlantis/server/core/terraform"
	"github.com/runatlantis/atlantis/server/events/command"
)

const minimumShowTfVersion string = "0.12.0"

func NewShowStepRunner(executor TerraformExec, defaultTfDistribution terraform.Distribution, defaultTFVersion *version.Version) (Runner, error) {
	showStepRunner := &showStepRunner{
		terraformExecutor:     executor,
		defaultTfDistribution: defaultTfDistribution,
		defaultTFVersion:      defaultTFVersion,
	}
	remotePlanRunner := NullRunner{}
	runner := NewPlanTypeStepRunnerDelegate(showStepRunner, remotePlanRunner)
	return NewMinimumVersionStepRunnerDelegate(minimumShowTfVersion, defaultTFVersion, runner)
}

// showStepRunner runs terraform show on an existing plan file and outputs it to a json file
type showStepRunner struct {
	terraformExecutor     TerraformExec
	defaultTfDistribution terraform.Distribution
	defaultTFVersion      *version.Version
}

func (p *showStepRunner) Run(ctx command.ProjectContext, _ []string, path string, envs map[string]string) (string, error) {
	tfDistribution := p.defaultTfDistribution
	tfVersion := p.defaultTFVersion
	if ctx.TerraformDistribution != nil {
		tfDistribution = terraform.NewDistribution(*ctx.TerraformDistribution)
	}
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
		tfDistribution,
		tfVersion,
		ctx.Workspace,
	)

	if err != nil {
		return "", errors.Wrap(err, "running terraform show")
	}

	if err := os.WriteFile(showResultFile, []byte(output), 0600); err != nil {
		return "", errors.Wrap(err, "writing terraform show result")
	}

	return output, nil
}
