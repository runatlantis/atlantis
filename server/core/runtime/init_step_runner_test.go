// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package runtime_test

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	version "github.com/hashicorp/go-version"
	. "github.com/petergtz/pegomock/v4"

	"github.com/runatlantis/atlantis/server/core/runtime"
	tf "github.com/runatlantis/atlantis/server/core/terraform"
	"github.com/runatlantis/atlantis/server/core/terraform/mocks"
	tfclientmocks "github.com/runatlantis/atlantis/server/core/terraform/tfclient/mocks"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/logging"
	. "github.com/runatlantis/atlantis/testing"
)

func TestRun_UsesGetOrInitForRightVersion(t *testing.T) {
	RegisterMockTestingT(t)
	mockDownloader := mocks.NewMockDownloader()
	tfDistribution := tf.NewDistributionTerraformWithDownloader(mockDownloader)
	cases := []struct {
		version string
		expCmd  string
	}{
		{
			"0.8.9",
			"get",
		},
		{
			"0.9.0",
			"init",
		},
		{
			"0.9.1",
			"init",
		},
		{
			"0.10.0",
			"init",
		},
	}

	for _, c := range cases {
		t.Run(c.version, func(t *testing.T) {
			terraform := tfclientmocks.NewMockClient()

			logger := logging.NewNoopLogger(t)
			ctx := command.ProjectContext{
				Workspace:  "workspace",
				RepoRelDir: ".",
				Log:        logger,
			}
			// extra_args is marked expandable on the context the client receives.
			ctx.ExpandableArgs = []string{"extra", "args"}

			tfVersion, _ := version.NewVersion(c.version)
			iso := runtime.InitStepRunner{
				TerraformExecutor:     terraform,
				DefaultTFDistribution: tfDistribution,
				DefaultTFVersion:      tfVersion,
			}
			When(terraform.RunCommandWithVersion(Any[command.ProjectContext](), Any[string](), Any[[]string](), Any[map[string]string](), Any[tf.Distribution](), Any[*version.Version](), Any[string]())).
				ThenReturn("output", nil)

			output, err := iso.Run(ctx, []string{"extra", "args"}, "/path", map[string]string(nil))
			Ok(t, err)
			// When there is no error, should not return init output to PR.
			Equals(t, "", output)

			// If using init then we specify -input=false but not for get.
			expArgs := []string{c.expCmd, "-input=false", "-upgrade", "extra", "args"}
			if c.expCmd == "get" {
				expArgs = []string{c.expCmd, "-upgrade", "extra", "args"}
			}
			terraform.VerifyWasCalledOnce().RunCommandWithVersion(ctx, "/path", expArgs, map[string]string(nil), tfDistribution, tfVersion, "workspace")
		})
	}
}

func TestInitStepRunner_IgnoresCommentArgsForExpansionPolicy(t *testing.T) {
	RegisterMockTestingT(t)
	terraform := tfclientmocks.NewMockClient()
	logger := logging.NewNoopLogger(t)
	tfVersion := version.Must(version.NewVersion("1.14.0"))
	mockDownloader := mocks.NewMockDownloader()
	tfDistribution := tf.NewDistributionTerraformWithDownloader(mockDownloader)

	const backendConfigArg = "-backend-config=$ATLANTIS_BACKEND_CONFIG"
	ctx := command.ProjectContext{
		Workspace:   "default",
		RepoRelDir:  ".",
		Log:         logger,
		CommentArgs: []string{backendConfigArg},
	}
	execCtx := ctx
	execCtx.ExpandableArgs = []string{backendConfigArg}
	execCtx.CommentArgs = nil

	runner := runtime.InitStepRunner{
		TerraformExecutor:     terraform,
		DefaultTFDistribution: tfDistribution,
		DefaultTFVersion:      tfVersion,
	}
	When(terraform.RunCommandWithVersion(Any[command.ProjectContext](), Any[string](), Any[[]string](), Any[map[string]string](), Any[tf.Distribution](), Any[*version.Version](), Any[string]())).
		ThenReturn("", nil)

	path := t.TempDir()
	output, err := runner.Run(ctx, []string{backendConfigArg}, path, nil)
	Ok(t, err)
	Equals(t, "", output)

	terraform.VerifyWasCalledOnce().RunCommandWithVersion(
		execCtx,
		path,
		[]string{"init", "-input=false", "-upgrade", backendConfigArg},
		map[string]string(nil),
		tfDistribution,
		tfVersion,
		"default",
	)
}

func TestInitStepRunner_TestRun_UsesConfiguredDistribution(t *testing.T) {
	RegisterMockTestingT(t)
	mockDownloader := mocks.NewMockDownloader()
	tfDistribution := tf.NewDistributionTerraformWithDownloader(mockDownloader)
	cases := []struct {
		version      string
		distribution string
		expCmd       string
	}{
		{
			"0.8.9",
			"opentofu",
			"get",
		},
		{
			"0.8.9",
			"terraform",
			"get",
		},
		{
			"0.9.0",
			"opentofu",
			"init",
		},
		{
			"0.9.1",
			"terraform",
			"init",
		},
	}

	for _, c := range cases {
		t.Run(c.version, func(t *testing.T) {
			terraform := tfclientmocks.NewMockClient()

			logger := logging.NewNoopLogger(t)
			ctx := command.ProjectContext{
				Workspace:             "workspace",
				RepoRelDir:            ".",
				Log:                   logger,
				TerraformDistribution: &c.distribution,
			}
			// extra_args is marked expandable on the context the client receives.
			ctx.ExpandableArgs = []string{"extra", "args"}

			tfVersion, _ := version.NewVersion(c.version)
			iso := runtime.InitStepRunner{
				TerraformExecutor:     terraform,
				DefaultTFDistribution: tfDistribution,
				DefaultTFVersion:      tfVersion,
			}
			When(terraform.RunCommandWithVersion(Any[command.ProjectContext](), Any[string](), Any[[]string](), Any[map[string]string](), Any[tf.Distribution](), Any[*version.Version](), Any[string]())).
				ThenReturn("output", nil)

			output, err := iso.Run(ctx, []string{"extra", "args"}, "/path", map[string]string(nil))
			Ok(t, err)
			// When there is no error, should not return init output to PR.
			Equals(t, "", output)

			// If using init then we specify -input=false but not for get.
			expArgs := []string{c.expCmd, "-input=false", "-upgrade", "extra", "args"}
			if c.expCmd == "get" {
				expArgs = []string{c.expCmd, "-upgrade", "extra", "args"}
			}
			terraform.VerifyWasCalledOnce().RunCommandWithVersion(Eq(ctx), Eq("/path"), Eq(expArgs), Eq(map[string]string(nil)), NotEq(tfDistribution), Eq(tfVersion), Eq("workspace"))
		})
	}
}

func TestRun_ShowInitOutputOnError(t *testing.T) {
	// If there was an error during init then we want the output to be returned.
	RegisterMockTestingT(t)
	tfClient := tfclientmocks.NewMockClient()
	logger := logging.NewNoopLogger(t)
	When(tfClient.RunCommandWithVersion(Any[command.ProjectContext](), Any[string](), Any[[]string](), Any[map[string]string](), Any[tf.Distribution](), Any[*version.Version](), Any[string]())).
		ThenReturn("output", errors.New("error"))
	mockDownloader := mocks.NewMockDownloader()
	tfDistribution := tf.NewDistributionTerraformWithDownloader(mockDownloader)
	tfVersion, _ := version.NewVersion("0.11.0")
	iso := runtime.InitStepRunner{
		TerraformExecutor:     tfClient,
		DefaultTFDistribution: tfDistribution,
		DefaultTFVersion:      tfVersion,
	}

	output, err := iso.Run(command.ProjectContext{
		Workspace:  "workspace",
		RepoRelDir: ".",
		Log:        logger,
	}, nil, "/path", map[string]string(nil))
	ErrEquals(t, "error", err)
	Equals(t, "output", output)
}

func TestRun_InitOmitsUpgradeFlagIfLockFileTracked(t *testing.T) {
	// Initialize the git repo.
	repoDir := initRepo(t)

	lockFilePath := filepath.Join(repoDir, ".terraform.lock.hcl")
	err := os.WriteFile(lockFilePath, nil, 0600)
	Ok(t, err)
	// commit lock file
	runCmd(t, repoDir, "git", "add", ".terraform.lock.hcl")
	runCmd(t, repoDir, "git", "commit", "-m", "add .terraform.lock.hcl")

	logger := logging.NewNoopLogger(t)
	ctx := command.ProjectContext{
		Workspace:  "workspace",
		RepoRelDir: ".",
		Log:        logger,
	}
	// extra_args is marked expandable on the context the client receives.
	ctx.ExpandableArgs = []string{"extra", "args"}

	RegisterMockTestingT(t)
	terraform := tfclientmocks.NewMockClient()
	mockDownloader := mocks.NewMockDownloader()
	tfDistribution := tf.NewDistributionTerraformWithDownloader(mockDownloader)
	tfVersion, _ := version.NewVersion("0.14.0")
	iso := runtime.InitStepRunner{
		TerraformExecutor:     terraform,
		DefaultTFDistribution: tfDistribution,
		DefaultTFVersion:      tfVersion,
	}
	When(terraform.RunCommandWithVersion(Any[command.ProjectContext](), Any[string](), Any[[]string](), Any[map[string]string](), Any[tf.Distribution](), Any[*version.Version](), Any[string]())).
		ThenReturn("output", nil)

	output, err := iso.Run(ctx, []string{"extra", "args"}, repoDir, map[string]string(nil))
	Ok(t, err)
	// When there is no error, should not return init output to PR.
	Equals(t, "", output)

	expectedArgs := []string{"init", "-input=false", "extra", "args"}
	terraform.VerifyWasCalledOnce().RunCommandWithVersion(ctx, repoDir, expectedArgs, map[string]string(nil), tfDistribution, tfVersion, "workspace")
}

func TestRun_InitKeepsUpgradeFlagIfLockFileNotPresent(t *testing.T) {
	tmpDir := t.TempDir()

	RegisterMockTestingT(t)
	terraform := tfclientmocks.NewMockClient()
	logger := logging.NewNoopLogger(t)
	ctx := command.ProjectContext{
		Workspace:  "workspace",
		RepoRelDir: ".",
		Log:        logger,
	}
	// extra_args is marked expandable on the context the client receives.
	ctx.ExpandableArgs = []string{"extra", "args"}
	mockDownloader := mocks.NewMockDownloader()
	tfDistribution := tf.NewDistributionTerraformWithDownloader(mockDownloader)
	tfVersion, _ := version.NewVersion("0.14.0")
	iso := runtime.InitStepRunner{
		TerraformExecutor:     terraform,
		DefaultTFDistribution: tfDistribution,
		DefaultTFVersion:      tfVersion,
	}
	When(terraform.RunCommandWithVersion(Any[command.ProjectContext](), Any[string](), Any[[]string](), Any[map[string]string](), Any[tf.Distribution](), Any[*version.Version](), Any[string]())).
		ThenReturn("output", nil)

	output, err := iso.Run(ctx, []string{"extra", "args"}, tmpDir, map[string]string(nil))
	Ok(t, err)
	// When there is no error, should not return init output to PR.
	Equals(t, "", output)

	expectedArgs := []string{"init", "-input=false", "-upgrade", "extra", "args"}
	terraform.VerifyWasCalledOnce().RunCommandWithVersion(ctx, tmpDir, expectedArgs, map[string]string(nil), tfDistribution, tfVersion, "workspace")
}

func TestRun_InitKeepUpgradeFlagIfLockFilePresentAndTFLessThanPoint14(t *testing.T) {
	tmpDir := t.TempDir()
	lockFilePath := filepath.Join(tmpDir, ".terraform.lock.hcl")
	err := os.WriteFile(lockFilePath, nil, 0600)
	Ok(t, err)

	RegisterMockTestingT(t)
	terraform := tfclientmocks.NewMockClient()

	logger := logging.NewNoopLogger(t)
	ctx := command.ProjectContext{
		Workspace:  "workspace",
		RepoRelDir: ".",
		Log:        logger,
	}
	// extra_args is marked expandable on the context the client receives.
	ctx.ExpandableArgs = []string{"extra", "args"}
	mockDownloader := mocks.NewMockDownloader()
	tfDistribution := tf.NewDistributionTerraformWithDownloader(mockDownloader)
	tfVersion, _ := version.NewVersion("0.13.0")
	iso := runtime.InitStepRunner{
		TerraformExecutor:     terraform,
		DefaultTFDistribution: tfDistribution,
		DefaultTFVersion:      tfVersion,
	}
	When(terraform.RunCommandWithVersion(Any[command.ProjectContext](), Any[string](), Any[[]string](), Any[map[string]string](), Any[tf.Distribution](), Any[*version.Version](), Any[string]())).
		ThenReturn("output", nil)

	output, err := iso.Run(ctx, []string{"extra", "args"}, tmpDir, map[string]string(nil))
	Ok(t, err)
	// When there is no error, should not return init output to PR.
	Equals(t, "", output)

	expectedArgs := []string{"init", "-input=false", "-upgrade", "extra", "args"}
	terraform.VerifyWasCalledOnce().RunCommandWithVersion(ctx, tmpDir, expectedArgs, map[string]string(nil), tfDistribution, tfVersion, "workspace")
}

func TestRun_InitExtraArgsDeDupe(t *testing.T) {
	RegisterMockTestingT(t)
	cases := []struct {
		description  string
		extraArgs    []string
		expectedArgs []string
	}{
		{
			"No extra args",
			[]string{},
			[]string{"init", "-input=false", "-upgrade"},
		},
		{
			"Override -upgrade",
			[]string{"-upgrade=false"},
			[]string{"init", "-input=false", "-upgrade=false"},
		},
		{
			"Override -input",
			[]string{"-input=true"},
			[]string{"init", "-input=true", "-upgrade"},
		},
		{
			"Override -input and -upgrade",
			[]string{"-input=true", "-upgrade=false"},
			[]string{"init", "-input=true", "-upgrade=false"},
		},
		{
			"Non duplicate extra args",
			[]string{"extra", "args"},
			[]string{"init", "-input=false", "-upgrade", "extra", "args"},
		},
		{
			"Override upgrade with extra args",
			[]string{"extra", "args", "-upgrade=false"},
			[]string{"init", "-input=false", "-upgrade=false", "extra", "args"},
		},
	}

	for _, c := range cases {
		t.Run(c.description, func(t *testing.T) {
			terraform := tfclientmocks.NewMockClient()

			logger := logging.NewNoopLogger(t)
			ctx := command.ProjectContext{
				Workspace:  "workspace",
				RepoRelDir: ".",
				Log:        logger,
			}
			// extra_args is marked expandable on the context the client receives.
			if len(c.extraArgs) > 0 {
				ctx.ExpandableArgs = c.extraArgs
			}
			mockDownloader := mocks.NewMockDownloader()
			tfDistribution := tf.NewDistributionTerraformWithDownloader(mockDownloader)
			tfVersion, _ := version.NewVersion("0.10.0")
			iso := runtime.InitStepRunner{
				TerraformExecutor:     terraform,
				DefaultTFDistribution: tfDistribution,
				DefaultTFVersion:      tfVersion,
			}
			When(terraform.RunCommandWithVersion(Any[command.ProjectContext](), Any[string](), Any[[]string](), Any[map[string]string](), Any[tf.Distribution](), Any[*version.Version](), Any[string]())).
				ThenReturn("output", nil)

			output, err := iso.Run(ctx, c.extraArgs, "/path", map[string]string(nil))
			Ok(t, err)
			// When there is no error, should not return init output to PR.
			Equals(t, "", output)

			terraform.VerifyWasCalledOnce().RunCommandWithVersion(ctx, "/path", c.expectedArgs, map[string]string(nil), tfDistribution, tfVersion, "workspace")
		})
	}
}

func TestRun_InitDeletesLockFileIfPresentAndNotTracked(t *testing.T) {
	// Initialize the git repo.
	repoDir := initRepo(t)

	lockFilePath := filepath.Join(repoDir, ".terraform.lock.hcl")
	err := os.WriteFile(lockFilePath, nil, 0600)
	Ok(t, err)

	RegisterMockTestingT(t)
	terraform := tfclientmocks.NewMockClient()

	logger := logging.NewNoopLogger(t)
	mockDownloader := mocks.NewMockDownloader()
	tfDistribution := tf.NewDistributionTerraformWithDownloader(mockDownloader)
	tfVersion, _ := version.NewVersion("0.14.0")

	iso := runtime.InitStepRunner{
		TerraformExecutor:     terraform,
		DefaultTFDistribution: tfDistribution,
		DefaultTFVersion:      tfVersion,
	}
	When(terraform.RunCommandWithVersion(Any[command.ProjectContext](), Any[string](), Any[[]string](), Any[map[string]string](), Any[tf.Distribution](), Any[*version.Version](), Any[string]())).
		ThenReturn("output", nil)

	ctx := command.ProjectContext{
		Workspace:  "workspace",
		RepoRelDir: ".",
		Log:        logger,
	}
	// extra_args is marked expandable on the context the client receives.
	ctx.ExpandableArgs = []string{"extra", "args"}
	output, err := iso.Run(ctx, []string{"extra", "args"}, repoDir, map[string]string(nil))
	Ok(t, err)
	// When there is no error, should not return init output to PR.
	Equals(t, "", output)

	expectedArgs := []string{"init", "-input=false", "-upgrade", "extra", "args"}
	terraform.VerifyWasCalledOnce().RunCommandWithVersion(ctx, repoDir, expectedArgs, map[string]string(nil), tfDistribution, tfVersion, "workspace")
}

// concurrencyTFExec records how many invocations ran concurrently and can
// block each invocation until released to force overlap between goroutines.
type concurrencyTFExec struct {
	// barrier blocks every invocation until release is closed.
	barrier bool
	release chan struct{}
	active  atomic.Int64
	max     atomic.Int64
	total   atomic.Int64
}

func (c *concurrencyTFExec) RunCommandWithVersion(ctx command.ProjectContext, path string, args []string, envs map[string]string, d tf.Distribution, v *version.Version, workspace string) (string, error) {
	cur := c.active.Add(1)
	c.total.Add(1)
	for {
		max := c.max.Load()
		if cur <= max || c.max.CompareAndSwap(max, cur) {
			break
		}
	}
	if c.barrier {
		<-c.release
	} else {
		time.Sleep(25 * time.Millisecond)
	}
	c.active.Add(-1)
	return "", nil
}

func (c *concurrencyTFExec) EnsureVersion(log logging.SimpleLogging, d tf.Distribution, v *version.Version) error {
	return nil
}

func TestRun_InitSerializedWhenPluginCacheEnabled(t *testing.T) {
	RegisterMockTestingT(t)
	exec := &concurrencyTFExec{release: make(chan struct{})}
	logger := logging.NewNoopLogger(t)
	tfVersion, _ := version.NewVersion("1.5.0")
	mockDownloader := mocks.NewMockDownloader()
	tfDistribution := tf.NewDistributionTerraformWithDownloader(mockDownloader)
	iso := runtime.InitStepRunner{
		TerraformExecutor:     exec,
		DefaultTFDistribution: tfDistribution,
		DefaultTFVersion:      tfVersion,
		PluginCache:           true,
	}

	const n = 5
	var wg sync.WaitGroup
	for range n {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := iso.Run(command.ProjectContext{
				Workspace:  "workspace",
				RepoRelDir: ".",
				Log:        logger,
			}, nil, "/path", map[string]string(nil))
			Ok(t, err)
		}()
	}
	wg.Wait()

	Equals(t, int64(n), exec.total.Load())
	Equals(t, int64(1), exec.max.Load())
}

func TestRun_InitParallelByDefault(t *testing.T) {
	RegisterMockTestingT(t)
	exec := &concurrencyTFExec{barrier: true, release: make(chan struct{})}
	logger := logging.NewNoopLogger(t)
	tfVersion, _ := version.NewVersion("1.5.0")
	mockDownloader := mocks.NewMockDownloader()
	tfDistribution := tf.NewDistributionTerraformWithDownloader(mockDownloader)
	iso := runtime.InitStepRunner{
		TerraformExecutor:     exec,
		DefaultTFDistribution: tfDistribution,
		DefaultTFVersion:      tfVersion,
	}

	const n = 5
	var wg sync.WaitGroup
	for range n {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := iso.Run(command.ProjectContext{
				Workspace:  "workspace",
				RepoRelDir: ".",
				Log:        logger,
			}, nil, "/path", map[string]string(nil))
			Ok(t, err)
		}()
	}

	// Wait until all runs have entered the executor. If they were serialized
	// this would time out because later runs only start after earlier ones.
	deadline := time.Now().Add(10 * time.Second)
	for exec.total.Load() < n && time.Now().Before(deadline) {
		time.Sleep(time.Millisecond)
	}

	Equals(t, int64(n), exec.max.Load())
	close(exec.release)
	wg.Wait()
	Equals(t, int64(n), exec.total.Load())
}

func runCmd(t *testing.T, dir string, name string, args ...string) string {
	t.Helper()
	cpCmd := exec.Command(name, args...)
	cpCmd.Dir = dir
	cpOut, err := cpCmd.CombinedOutput()
	Assert(t, err == nil, "err running %q: %s", strings.Join(append([]string{name}, args...), " "), cpOut)
	return string(cpOut)
}

func initRepo(t *testing.T) string {
	repoDir := t.TempDir()
	runCmd(t, repoDir, "git", "init")
	runCmd(t, repoDir, "touch", ".gitkeep")
	runCmd(t, repoDir, "git", "add", ".gitkeep")
	runCmd(t, repoDir, "git", "config", "--local", "user.email", "atlantisbot@runatlantis.io")
	runCmd(t, repoDir, "git", "config", "--local", "user.name", "atlantisbot")
	runCmd(t, repoDir, "git", "config", "--local", "commit.gpgsign", "false")
	runCmd(t, repoDir, "git", "commit", "-m", "initial commit")
	runCmd(t, repoDir, "git", "branch", "branch")
	return repoDir
}
