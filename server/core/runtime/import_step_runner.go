// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package runtime

import (
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
	return out, err
}
