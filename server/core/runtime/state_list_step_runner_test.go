// Copyright 2026 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package runtime_test

import (
	"testing"

	"github.com/hashicorp/go-version"
	. "github.com/petergtz/pegomock/v4"
	"github.com/runatlantis/atlantis/server/core/runtime"
	tf "github.com/runatlantis/atlantis/server/core/terraform"
	"github.com/runatlantis/atlantis/server/core/terraform/tfclient/mocks"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/logging"
	. "github.com/runatlantis/atlantis/testing"
)

func TestStateListStepRunner_Run(t *testing.T) {
	t.Run("all resources", func(t *testing.T) {
		assertStateListStepRunnerRunsCommand(t, []string{}, []string{"state", "list"})
	})
	t.Run("ID filter", func(t *testing.T) {
		assertStateListStepRunnerRunsCommand(t, []string{"-id=i-123"}, []string{"state", "list", "-id=i-123"})
	})
}

func TestStateListStepRunner_ConfiguredExtraArgs(t *testing.T) {
	assertStateListStepRunnerRunsCommand(t, []string{"module.app"}, []string{"state", "list", "-id=$RESOURCE_ID", "module.app"}, "-id=$RESOURCE_ID")
}

func TestStateListStepRunner_UsesDefaultVersionAndDistribution(t *testing.T) {
	assertStateListStepRunnerUsesDefaults(t, []string{}, []string{"state", "list"})
}

func assertStateListStepRunnerRunsCommand(t *testing.T, args, want []string, extraArgs ...string) {
	t.Helper()
	RegisterMockTestingT(t)
	executor := mocks.NewMockClient()
	dir := t.TempDir()
	def := version.Must(version.NewVersion("1.10.0"))
	project := version.Must(version.NewVersion("1.11.1"))
	distributionName := "opentofu"
	distribution := tf.NewDistribution(distributionName)
	ctx := command.ProjectContext{Log: logging.NewNoopLogger(t), CommandName: command.StateList, Workspace: "production", CommentArgs: args, TerraformVersion: project, TerraformDistribution: &distributionName}
	When(executor.RunCommandWithVersion(Any[command.ProjectContext](), Any[string](), Eq([]string{"workspace", "show"}), Any[map[string]string](), Any[tf.Distribution](), Any[*version.Version](), Any[string]())).ThenReturn("production\n", nil)
	When(executor.RunCommandWithVersion(Any[command.ProjectContext](), Any[string](), Eq(want), Any[map[string]string](), Any[tf.Distribution](), Any[*version.Version](), Any[string]())).ThenReturn("result", nil)
	runner := runtime.NewStateListStepRunner(executor, tf.NewDistribution("terraform"), def)
	out, err := runner.Run(ctx, extraArgs, dir, nil)
	Ok(t, err)
	Equals(t, "result", out)
	expectedCtx := ctx
	if len(extraArgs) > 0 {
		expectedCtx.ExpandableArgs = extraArgs
	}
	executor.VerifyWasCalledOnce().RunCommandWithVersion(expectedCtx, dir, want, map[string]string(nil), distribution, project, "production")
}

func assertStateListStepRunnerUsesDefaults(t *testing.T, args, want []string) {
	t.Helper()
	RegisterMockTestingT(t)
	executor := mocks.NewMockClient()
	dir := t.TempDir()
	tfVersion := version.Must(version.NewVersion("1.11.1"))
	distribution := tf.NewDistribution("terraform")
	ctx := command.ProjectContext{Log: logging.NewNoopLogger(t), CommandName: command.StateList, Workspace: "default", CommentArgs: args}
	When(executor.RunCommandWithVersion(Any[command.ProjectContext](), Any[string](), Eq([]string{"workspace", "show"}), Any[map[string]string](), Any[tf.Distribution](), Any[*version.Version](), Any[string]())).ThenReturn("default", nil)
	When(executor.RunCommandWithVersion(Any[command.ProjectContext](), Any[string](), Eq(want), Any[map[string]string](), Any[tf.Distribution](), Any[*version.Version](), Any[string]())).ThenReturn("result", nil)
	out, err := runtime.NewStateListStepRunner(executor, distribution, tfVersion).Run(ctx, nil, dir, nil)
	Ok(t, err)
	Equals(t, "result", out)
	executor.VerifyWasCalledOnce().RunCommandWithVersion(ctx, dir, want, map[string]string(nil), distribution, tfVersion, "default")
}
