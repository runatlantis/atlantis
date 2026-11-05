// Copyright 2026 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package runtime

import (
	"path/filepath"

	"github.com/hashicorp/go-version"
	"github.com/runatlantis/atlantis/server/core/terraform"
	"github.com/runatlantis/atlantis/server/events/command"
)

type outputStepRunner struct {
	terraformExecutor     TerraformExec
	defaultTFDistribution terraform.Distribution
	defaultTFVersion      *version.Version
}

func NewOutputStepRunner(terraformExecutor TerraformExec, defaultTfDistribution terraform.Distribution, defaultTfVersion *version.Version) Runner {
	runner := &outputStepRunner{
		terraformExecutor:     terraformExecutor,
		defaultTFDistribution: defaultTfDistribution,
		defaultTFVersion:      defaultTfVersion,
	}
	return &workspaceStepRunnerDelegate{
		terraformExecutor:     terraformExecutor,
		defaultTfDistribution: defaultTfDistribution,
		defaultTfVersion:      defaultTfVersion,
		delegate:              runner,
		existingWorkspaceOnly: true,
	}
}

func (p *outputStepRunner) Run(ctx command.ProjectContext, extraArgs []string, path string, envs map[string]string) (string, error) {
	if err := command.ValidateInspectionArgs(command.Output, ctx.CommentArgs); err != nil {
		return "", err
	}
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

	outputCmd := []string{"output"}
	outputCmd = append(outputCmd, extraArgs...)
	outputCmd = append(outputCmd, ctx.CommentArgs...)
	return p.terraformExecutor.RunCommandWithVersion(ctx, filepath.Clean(path), outputCmd, envs, tfDistribution, tfVersion, ctx.Workspace)
}
