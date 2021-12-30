package runtime

import (
	"os"
	"path/filepath"

	"github.com/hashicorp/go-version"
	"github.com/pkg/errors"
	"github.com/runatlantis/atlantis/server/events/models"
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

func (p *ShowStepRunner) Run(ctx models.ProjectCommandContext, extraArgs []string, path string, envs map[string]string) (string, error) {
	tfVersion := p.DefaultTFVersion
	if ctx.TerraformVersion != nil {
		tfVersion = ctx.TerraformVersion
	}

	planFile := filepath.Join(path, GetPlanFilename(ctx.Workspace, ctx.ProjectName))
	showResultFile := filepath.Join(path, ctx.GetShowResultFileName())

	output, err := p.TerraformExecutor.RunCommandWithVersion(
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
