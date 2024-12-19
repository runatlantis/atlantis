package runtime

import (
	"os"
	"path/filepath"

	version "github.com/hashicorp/go-version"
	"github.com/runatlantis/atlantis/server/core/terraform"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/utils"
)

type importStepRunner struct {
	terraformExecutor     TerraformExec
	defaultTFDistribution terraform.Distribution
	defaultTFVersion      *version.Version
}

func NewImportStepRunner(terraformExecutor TerraformExec, defaultTfDistribution terraform.Distribution, defaultTfVersion *version.Version) Runner {
	runner := &importStepRunner{
		terraformExecutor:     terraformExecutor,
		defaultTFDistribution: defaultTfDistribution,
		defaultTFVersion:      defaultTfVersion,
	}
	return NewWorkspaceStepRunnerDelegate(terraformExecutor, defaultTfDistribution, defaultTfVersion, runner)
}

func (p *importStepRunner) Run(ctx command.ProjectContext, extraArgs []string, path string, envs map[string]string) (string, error) {
	tfDistribution := p.defaultTFDistribution
	tfVersion := p.defaultTFVersion
	if ctx.TerraformDistribution != nil {
		tfDistribution = terraform.NewDistribution(*ctx.TerraformDistribution)
	}
	if ctx.TerraformVersion != nil {
		tfVersion = ctx.TerraformVersion
	}

	importCmd := []string{"import"}
	importCmd = append(importCmd, extraArgs...)
	importCmd = append(importCmd, ctx.EscapedCommentArgs...)
	out, err := p.terraformExecutor.RunCommandWithVersion(ctx, filepath.Clean(path), importCmd, envs, tfDistribution, tfVersion, ctx.Workspace)

	// If the import was successful and a plan file exists, delete the plan.
	planPath := filepath.Join(path, GetPlanFilename(ctx.Workspace, ctx.ProjectName))
	if err == nil {
		if _, planPathErr := os.Stat(planPath); !os.IsNotExist(planPathErr) {
			ctx.Log.Info("import successful, deleting planfile")
			if removeErr := utils.RemoveIgnoreNonExistent(planPath); removeErr != nil {
				ctx.Log.Warn("failed to delete planfile after successful import: %s", removeErr)
			}
		}
	}
	return out, err
}
