// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package runtime

import (
	"os"
	"path/filepath"

	version "github.com/hashicorp/go-version"
	"github.com/runatlantis/atlantis/server/core/terraform"
	"github.com/runatlantis/atlantis/server/events/command"
)

type importStepRunner struct {
	terraformExecutor     TerraformExec
	defaultTFDistribution terraform.Distribution
	defaultTFVersion      *version.Version
	planStore             PlanStore
}

func NewImportStepRunner(terraformExecutor TerraformExec, defaultTfDistribution terraform.Distribution, defaultTfVersion *version.Version, planStore PlanStore) Runner {
	runner := &importStepRunner{
		terraformExecutor:     terraformExecutor,
		defaultTFDistribution: defaultTfDistribution,
		defaultTFVersion:      defaultTfVersion,
		planStore:             planStore,
	}
	return NewWorkspaceStepRunnerDelegate(terraformExecutor, defaultTfDistribution, defaultTfVersion, runner)
}

func (p *importStepRunner) Run(ctx command.ProjectContext, extraArgs []string, path string, envs map[string]string) (string, error) {
	// extra_args comes from configuration, so environment variable references
	// in it may be expanded. Marked on this copy of the context; everything
	// else, including comment args, stays literal.
	if len(extraArgs) > 0 {
		ctx.ExpandableArgs = extraArgs
	}
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
	importCmd = append(importCmd, ctx.CommentArgs...)
	out, err := p.terraformExecutor.RunCommandWithVersion(ctx, filepath.Clean(path), importCmd, envs, tfDistribution, tfVersion, ctx.Workspace)

	// If the import was successful and a plan file exists, delete the plan.
	planPath := GetPlanFilePath(ctx, path)
	if err == nil {
		if _, planPathErr := os.Stat(planPath); !os.IsNotExist(planPathErr) {
			ctx.Log.Info("import successful, deleting planfile")
			if removeErr := p.planStore.Remove(ctx, planPath); removeErr != nil {
				ctx.Log.Warn("failed to delete planfile after successful import: %s", removeErr)
			}
		}
	}
	return out, err
}
