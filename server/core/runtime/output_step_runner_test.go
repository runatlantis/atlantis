// Copyright 2026 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package runtime_test

import (
	"errors"
	"os"
	"path/filepath"
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

func TestOutputStepRunner_Run(t *testing.T) {
	t.Run("all outputs", func(t *testing.T) {
		assertOutputStepRunnerPreservesPlan(t, []string{}, []string{"output"})
	})
	t.Run("raw named output", func(t *testing.T) {
		assertOutputStepRunnerPreservesPlan(t, []string{"-raw", "endpoint"}, []string{"output", "-raw", "endpoint"})
	})
}

func TestOutputStepRunner_DoesNotCreateMissingWorkspace(t *testing.T) {
	assertOutputStepRunnerDoesNotCreateWorkspace(t, []string{}, []string{"output"})
}

func TestOutputStepRunner_ConfiguredExtraArgs(t *testing.T) {
	assertOutputStepRunnerPreservesPlan(t, []string{"endpoint"}, []string{"output", "-json", "endpoint"}, "-json")
}

func TestOutputStepRunner_UsesDefaultVersionAndDistribution(t *testing.T) {
	assertOutputStepRunnerUsesDefaults(t, []string{}, []string{"output"})
}

func assertOutputStepRunnerPreservesPlan(t *testing.T, args, want []string, extraArgs ...string) {
	t.Helper()
	RegisterMockTestingT(t)
	executor := mocks.NewMockClient()
	dir := t.TempDir()
	plan := filepath.Join(dir, "production.tfplan")
	Ok(t, os.WriteFile(plan, []byte("saved plan"), 0600))
	def := version.Must(version.NewVersion("1.10.0"))
	project := version.Must(version.NewVersion("1.11.1"))
	distributionName := "opentofu"
	distribution := tf.NewDistribution(distributionName)
	ctx := command.ProjectContext{Log: logging.NewNoopLogger(t), CommandName: command.Output, Workspace: "production", CommentArgs: args, TerraformVersion: project, TerraformDistribution: &distributionName}
	When(executor.RunCommandWithVersion(Any[command.ProjectContext](), Any[string](), Eq([]string{"workspace", "show"}), Any[map[string]string](), Any[tf.Distribution](), Any[*version.Version](), Any[string]())).ThenReturn("production\n", nil)
	When(executor.RunCommandWithVersion(Any[command.ProjectContext](), Any[string](), Eq(want), Any[map[string]string](), Any[tf.Distribution](), Any[*version.Version](), Any[string]())).ThenReturn("result", nil)
	runner := runtime.NewOutputStepRunner(executor, tf.NewDistribution("terraform"), def)
	out, err := runner.Run(ctx, extraArgs, dir, nil)
	Ok(t, err)
	Equals(t, "result", out)
	expectedCtx := ctx
	if len(extraArgs) > 0 {
		expectedCtx.ExpandableArgs = extraArgs
	}
	executor.VerifyWasCalledOnce().RunCommandWithVersion(expectedCtx, dir, want, map[string]string(nil), distribution, project, "production")
	saved, err := os.ReadFile(plan)
	Ok(t, err)
	Equals(t, "saved plan", string(saved))
}

func assertOutputStepRunnerDoesNotCreateWorkspace(t *testing.T, args, want []string) {
	t.Helper()
	RegisterMockTestingT(t)
	executor := mocks.NewMockClient()
	tfVersion := version.Must(version.NewVersion("1.11.1"))
	distribution := tf.NewDistribution("terraform")
	ctx := command.ProjectContext{Log: logging.NewNoopLogger(t), CommandName: command.Output, CommentArgs: args, Workspace: "missing"}
	When(executor.RunCommandWithVersion(Any[command.ProjectContext](), Any[string](), Eq([]string{"workspace", "show"}), Any[map[string]string](), Any[tf.Distribution](), Any[*version.Version](), Any[string]())).ThenReturn("default", nil)
	When(executor.RunCommandWithVersion(Any[command.ProjectContext](), Any[string](), Eq([]string{"workspace", "select", "missing"}), Any[map[string]string](), Any[tf.Distribution](), Any[*version.Version](), Any[string]())).ThenReturn("", errors.New("workspace does not exist"))
	_, err := runtime.NewOutputStepRunner(executor, distribution, tfVersion).Run(ctx, nil, t.TempDir(), nil)
	Assert(t, err != nil, "expected workspace selection failure")
	executor.VerifyWasCalled(Never()).RunCommandWithVersion(Any[command.ProjectContext](), Any[string](), Eq([]string{"workspace", "new", "missing"}), Any[map[string]string](), Any[tf.Distribution](), Any[*version.Version](), Any[string]())
	executor.VerifyWasCalled(Never()).RunCommandWithVersion(Any[command.ProjectContext](), Any[string](), Eq(want), Any[map[string]string](), Any[tf.Distribution](), Any[*version.Version](), Any[string]())
}

func assertOutputStepRunnerUsesDefaults(t *testing.T, args, want []string) {
	t.Helper()
	RegisterMockTestingT(t)
	executor := mocks.NewMockClient()
	dir := t.TempDir()
	tfVersion := version.Must(version.NewVersion("1.11.1"))
	distribution := tf.NewDistribution("terraform")
	ctx := command.ProjectContext{Log: logging.NewNoopLogger(t), CommandName: command.Output, Workspace: "default", CommentArgs: args}
	When(executor.RunCommandWithVersion(Any[command.ProjectContext](), Any[string](), Eq([]string{"workspace", "show"}), Any[map[string]string](), Any[tf.Distribution](), Any[*version.Version](), Any[string]())).ThenReturn("default", nil)
	When(executor.RunCommandWithVersion(Any[command.ProjectContext](), Any[string](), Eq(want), Any[map[string]string](), Any[tf.Distribution](), Any[*version.Version](), Any[string]())).ThenReturn("result", nil)
	out, err := runtime.NewOutputStepRunner(executor, distribution, tfVersion).Run(ctx, nil, dir, nil)
	Ok(t, err)
	Equals(t, "result", out)
	executor.VerifyWasCalledOnce().RunCommandWithVersion(ctx, dir, want, map[string]string(nil), distribution, tfVersion, "default")
}
