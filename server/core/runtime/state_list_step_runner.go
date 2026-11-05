// Copyright 2026 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package runtime

import (
	"path/filepath"

	"github.com/hashicorp/go-version"
	"github.com/runatlantis/atlantis/server/core/terraform"
	"github.com/runatlantis/atlantis/server/events/command"
)

type stateListStepRunner struct {
	terraformExecutor     TerraformExec
	defaultTFDistribution terraform.Distribution
	defaultTFVersion      *version.Version
}

func NewStateListStepRunner(terraformExecutor TerraformExec, defaultTfDistribution terraform.Distribution, defaultTfVersion *version.Version) Runner {
	runner := &stateListStepRunner{
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

func (p *stateListStepRunner) Run(ctx command.ProjectContext, extraArgs []string, path string, envs map[string]string) (string, error) {
	if err := command.ValidateInspectionArgs(command.StateList, ctx.CommentArgs); err != nil {
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

	stateListCmd := []string{"state", "list"}
	stateListCmd = append(stateListCmd, extraArgs...)
	stateListCmd = append(stateListCmd, ctx.CommentArgs...)
	return p.terraformExecutor.RunCommandWithVersion(ctx, filepath.Clean(path), stateListCmd, envs, tfDistribution, tfVersion, ctx.Workspace)
}
