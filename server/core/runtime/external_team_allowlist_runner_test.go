// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package runtime_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/runatlantis/atlantis/server/core/runtime"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/logging"
	. "github.com/runatlantis/atlantis/testing"
)

func TestDefaultExternalTeamAllowlistRunner_Run(t *testing.T) {
	logger := logging.NewNoopLogger(t)

	baseCtx := models.TeamAllowlistCheckerContext{
		BaseRepo: models.Repo{
			Name:  "atlantis",
			Owner: "runatlantis",
		},
		HeadRepo: models.Repo{
			Name:  "atlantis",
			Owner: "contrib-user",
		},
		Pull: models.PullRequest{
			Num:        1,
			HeadBranch: "feature-branch",
			HeadCommit: "abc123",
			BaseBranch: "main",
			Author:     "acme-user",
			URL:        "https://github.com/runatlantis/atlantis/pull/1",
		},
		User: models.User{
			Username: "acme-user",
		},
		Log:                logger,
		CommandName:        "plan",
		ProjectName:        "myproject",
		RepoDir:            "/tmp/repo",
		RepoRelDir:         "modules/vpc",
		EscapedCommentArgs: []string{"-target=resource"},
		Workspace:          "staging",
		API:                false,
		Verbose:            false,
	}

	t.Run("passes all environment variables to command", func(t *testing.T) {
		runner := runtime.DefaultExternalTeamAllowlistRunner{}
		// Print all the env vars we care about
		cmd := "echo " +
			"BASE_BRANCH_NAME=$BASE_BRANCH_NAME " +
			"BASE_REPO_NAME=$BASE_REPO_NAME " +
			"BASE_REPO_OWNER=$BASE_REPO_OWNER " +
			"HEAD_BRANCH_NAME=$HEAD_BRANCH_NAME " +
			"HEAD_COMMIT=$HEAD_COMMIT " +
			"HEAD_REPO_NAME=$HEAD_REPO_NAME " +
			"HEAD_REPO_OWNER=$HEAD_REPO_OWNER " +
			"PULL_AUTHOR=$PULL_AUTHOR " +
			"PULL_NUM=$PULL_NUM " +
			"PULL_URL=$PULL_URL " +
			"USER_NAME=$USER_NAME " +
			"COMMAND_NAME=$COMMAND_NAME " +
			"PROJECT_NAME=$PROJECT_NAME " +
			"REPO_ROOT=$REPO_ROOT " +
			"REPO_REL_PATH=$REPO_REL_PATH " +
			"WORKSPACE=$WORKSPACE " +
			"API=$API " +
			"VERBOSE=$VERBOSE"

		out, err := runner.Run(baseCtx, "sh", "-c", cmd)
		Ok(t, err)

		expected := "BASE_BRANCH_NAME=main " +
			"BASE_REPO_NAME=atlantis " +
			"BASE_REPO_OWNER=runatlantis " +
			"HEAD_BRANCH_NAME=feature-branch " +
			"HEAD_COMMIT=abc123 " +
			"HEAD_REPO_NAME=atlantis " +
			"HEAD_REPO_OWNER=contrib-user " +
			"PULL_AUTHOR=acme-user " +
			"PULL_NUM=1 " +
			"PULL_URL=https://github.com/runatlantis/atlantis/pull/1 " +
			"USER_NAME=acme-user " +
			"COMMAND_NAME=plan " +
			"PROJECT_NAME=myproject " +
			"REPO_ROOT=/tmp/repo " +
			"REPO_REL_PATH=modules/vpc " +
			"WORKSPACE=staging " +
			"API=false " +
			"VERBOSE=false"

		Equals(t, expected, out)
	})

	t.Run("WORKSPACE env var is set correctly", func(t *testing.T) {
		runner := runtime.DefaultExternalTeamAllowlistRunner{}
		ctx := baseCtx
		ctx.Workspace = "production"

		out, err := runner.Run(ctx, "sh", "-c", "echo $WORKSPACE")
		Ok(t, err)
		Equals(t, "production", out)
	})

	t.Run("API env var is true when API is set", func(t *testing.T) {
		runner := runtime.DefaultExternalTeamAllowlistRunner{}
		ctx := baseCtx
		ctx.API = true

		out, err := runner.Run(ctx, "sh", "-c", "echo $API")
		Ok(t, err)
		Equals(t, "true", out)
	})

	t.Run("API env var is false when API is not set", func(t *testing.T) {
		runner := runtime.DefaultExternalTeamAllowlistRunner{}
		ctx := baseCtx
		ctx.API = false

		out, err := runner.Run(ctx, "sh", "-c", "echo $API")
		Ok(t, err)
		Equals(t, "false", out)
	})

	t.Run("VERBOSE env var is true when Verbose is set", func(t *testing.T) {
		runner := runtime.DefaultExternalTeamAllowlistRunner{}
		ctx := baseCtx
		ctx.Verbose = true

		out, err := runner.Run(ctx, "sh", "-c", "echo $VERBOSE")
		Ok(t, err)
		Equals(t, "true", out)
	})

	t.Run("VERBOSE env var is false when Verbose is not set", func(t *testing.T) {
		runner := runtime.DefaultExternalTeamAllowlistRunner{}
		ctx := baseCtx
		ctx.Verbose = false

		out, err := runner.Run(ctx, "sh", "-c", "echo $VERBOSE")
		Ok(t, err)
		Equals(t, "false", out)
	})

	t.Run("COMMENT_ARGS are comma separated", func(t *testing.T) {
		runner := runtime.DefaultExternalTeamAllowlistRunner{}
		ctx := baseCtx
		ctx.EscapedCommentArgs = []string{"-target=foo", "-var=bar"}

		out, err := runner.Run(ctx, "sh", "-c", "echo $COMMENT_ARGS")
		Ok(t, err)
		Equals(t, "-target=foo,-var=bar", out)
	})

	t.Run("returns error on command failure", func(t *testing.T) {
		runner := runtime.DefaultExternalTeamAllowlistRunner{}
		_, err := runner.Run(baseCtx, "sh", "-c", "exit 1")
		Assert(t, err != nil, "expected error")
		Assert(t, strings.Contains(err.Error(), "exit status 1"), "expected exit status in error, got: %s", err)
	})

	t.Run("returns command output on error", func(t *testing.T) {
		runner := runtime.DefaultExternalTeamAllowlistRunner{}
		out, err := runner.Run(baseCtx, "sh", "-c", "echo denied && exit 1")
		Assert(t, err != nil, "expected error")
		Assert(t, strings.Contains(out, "denied"), "expected output to contain 'denied', got: %s", out)
	})

	t.Run("inherits OS environment", func(t *testing.T) {
		runner := runtime.DefaultExternalTeamAllowlistRunner{}
		envKey := "ATLANTIS_TEST_INHERIT_ENV"
		envVal := "inherited_value"
		t.Setenv(envKey, envVal)

		out, err := runner.Run(baseCtx, "sh", "-c", fmt.Sprintf("echo $%s", envKey))
		Ok(t, err)
		Equals(t, envVal, out)
	})

	t.Run("custom env vars override OS environment", func(t *testing.T) {
		runner := runtime.DefaultExternalTeamAllowlistRunner{}
		// Set an OS env var with the same name as a custom one
		t.Setenv("WORKSPACE", "os-value")

		ctx := baseCtx
		ctx.Workspace = "custom-value"

		out, err := runner.Run(ctx, "sh", "-c", "echo $WORKSPACE")
		Ok(t, err)
		// The custom env var should win because it's appended after os.Environ()
		Equals(t, "custom-value", out)
	})

	t.Run("trims trailing whitespace from output", func(t *testing.T) {
		runner := runtime.DefaultExternalTeamAllowlistRunner{}
		out, err := runner.Run(baseCtx, "sh", "-c", "printf 'allowed  \n  '")
		Ok(t, err)
		Equals(t, "allowed", out)
	})

	t.Run("empty workspace defaults to empty string", func(t *testing.T) {
		runner := runtime.DefaultExternalTeamAllowlistRunner{}
		ctx := baseCtx
		ctx.Workspace = ""

		// Use ${WORKSPACE+set} to verify WORKSPACE is actually set (even if empty),
		// not just absent from the environment
		out, err := runner.Run(ctx, "sh", "-c", "echo workspace_is_${WORKSPACE+set}")
		Ok(t, err)
		Equals(t, "workspace_is_set", out)
	})

	// Verify env var is not leaked from previous test
	t.Run("does not leak env between runs", func(t *testing.T) {
		runner := runtime.DefaultExternalTeamAllowlistRunner{}
		// t.Setenv in "inherits OS environment" subtest auto-restores on cleanup,
		// so ATLANTIS_TEST_INHERIT_ENV should not be set here.

		out, err := runner.Run(baseCtx, "sh", "-c", "echo val_is_${ATLANTIS_TEST_INHERIT_ENV}end")
		Ok(t, err)
		Equals(t, "val_is_end", out)
	})
}
