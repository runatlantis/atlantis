// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package runtime_test

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	version "github.com/hashicorp/go-version"
	. "github.com/petergtz/pegomock/v4"

	"github.com/runatlantis/atlantis/server/core/runtime"
	tf "github.com/runatlantis/atlantis/server/core/terraform"
	"github.com/runatlantis/atlantis/server/core/terraform/mocks"
	"github.com/runatlantis/atlantis/server/core/terraform/tfclient"
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

// Test that when the caller (e.g. an atlantis.yaml `env` step) already set
// TF_CLI_CONFIG_FILE, Run honors that explicit choice instead of silently
// overwriting it with the provider cache proxy's own generated config -
// even though that means this one init runs without the provider cache.
func TestRun_ProviderCache_HonorsExplicitTFCLIConfigFile(t *testing.T) {
	RegisterMockTestingT(t)
	terraform := tfclientmocks.NewMockClient()
	logger := logging.NewNoopLogger(t)
	mockDownloader := mocks.NewMockDownloader()
	tfDistribution := tf.NewDistributionTerraformWithDownloader(mockDownloader)
	tfVersion, _ := version.NewVersion("1.14.0")

	iso := runtime.InitStepRunner{
		TerraformExecutor:     terraform,
		DefaultTFDistribution: tfDistribution,
		DefaultTFVersion:      tfVersion,
		ProviderCache: &tfclient.ProviderCacheConfig{
			MirrorBaseURL: "http://127.0.0.1:8080/",
			RegistryHosts: []string{"registry.terraform.io"},
			MirrorDir:     t.TempDir(),
		},
	}
	When(terraform.RunCommandWithVersion(Any[command.ProjectContext](), Any[string](), Any[[]string](), Any[map[string]string](), Any[tf.Distribution](), Any[*version.Version](), Any[string]())).
		ThenReturn("output", nil)

	path := t.TempDir()
	explicitEnvs := map[string]string{"TF_CLI_CONFIG_FILE": "/custom/path/.terraformrc"}
	output, err := iso.Run(command.ProjectContext{Workspace: "workspace", RepoRelDir: ".", Log: logger}, nil, path, explicitEnvs)
	Ok(t, err)
	Equals(t, "", output)

	// The plain single-shot path ran with envs untouched, not the
	// provider-cache-rewritten TF_CLI_CONFIG_FILE.
	terraform.VerifyWasCalledOnce().RunCommandWithVersion(Any[command.ProjectContext](), Any[string](), Any[[]string](), Eq(explicitEnvs), Any[tf.Distribution](), Any[*version.Version](), Any[string]())

	_, err = os.Stat(filepath.Join(path, ".terraformrc-provider-cache-discovery"))
	Assert(t, os.IsNotExist(err), "provider cache CLI config should not have been written when TF_CLI_CONFIG_FILE was already explicitly set")
}

// Test that when phase 1 (discovery, routed through the proxy via a host
// block) succeeds outright - e.g. nothing needed installing - Run returns
// immediately without ever running phase 2 against the mirror.
func TestRun_ProviderCache_SucceedsOnDiscoveryPhaseAlone(t *testing.T) {
	RegisterMockTestingT(t)
	terraform := tfclientmocks.NewMockClient()
	logger := logging.NewNoopLogger(t)
	mockDownloader := mocks.NewMockDownloader()
	tfDistribution := tf.NewDistributionTerraformWithDownloader(mockDownloader)
	tfVersion, _ := version.NewVersion("1.14.0")

	iso := runtime.InitStepRunner{
		TerraformExecutor:     terraform,
		DefaultTFDistribution: tfDistribution,
		DefaultTFVersion:      tfVersion,
		ProviderCache: &tfclient.ProviderCacheConfig{
			MirrorBaseURL: "http://127.0.0.1:8080/",
			RegistryHosts: []string{"registry.terraform.io"},
			MirrorDir:     t.TempDir(),
		},
	}
	When(terraform.RunCommandWithVersion(Any[command.ProjectContext](), Any[string](), Any[[]string](), Any[map[string]string](), Any[tf.Distribution](), Any[*version.Version](), Any[string]())).
		ThenReturn("output", nil)

	path := t.TempDir()
	output, err := iso.Run(command.ProjectContext{Workspace: "workspace", RepoRelDir: ".", Log: logger}, nil, path, nil)
	Ok(t, err)
	Equals(t, "", output)

	terraform.VerifyWasCalledOnce().RunCommandWithVersion(Any[command.ProjectContext](), Any[string](), Any[[]string](), Any[map[string]string](), Any[tf.Distribution](), Any[*version.Version](), Any[string]())

	discoveryFile, err := os.ReadFile(filepath.Join(path, ".terraformrc-provider-cache-discovery"))
	Ok(t, err)
	Assert(t, strings.Contains(string(discoveryFile), `host "registry.terraform.io"`), "discovery CLI config missing host block: %s", discoveryFile)
	_, err = os.Stat(filepath.Join(path, ".terraformrc-provider-cache-mirror"))
	Assert(t, os.IsNotExist(err), "mirror CLI config should not have been written when phase 1 succeeded")
}

// Test that when phase 1 fails for a reason that names one of the configured
// registry hosts (the proxy is still installing a needed provider), Run
// retries against the filesystem mirror and returns success once it reports
// the provider is there.
func TestRun_ProviderCache_RetriesAgainstMirrorUntilReady(t *testing.T) {
	RegisterMockTestingT(t)
	terraform := tfclientmocks.NewMockClient()
	logger := logging.NewNoopLogger(t)
	mockDownloader := mocks.NewMockDownloader()
	tfDistribution := tf.NewDistributionTerraformWithDownloader(mockDownloader)
	tfVersion, _ := version.NewVersion("1.14.0")

	iso := runtime.InitStepRunner{
		TerraformExecutor:     terraform,
		DefaultTFDistribution: tfDistribution,
		DefaultTFVersion:      tfVersion,
		ProviderCache: &tfclient.ProviderCacheConfig{
			MirrorBaseURL: "http://127.0.0.1:8080/",
			RegistryHosts: []string{"registry.terraform.io"},
			MirrorDir:     t.TempDir(),
		},
	}
	// 1st call: initial discovery, fails - the proxy is (by design) installing
	// the provider in the background. 2nd call: mirror check, not ready yet.
	// 3rd call: re-triggered discovery (the retry loop asks the proxy again
	// each cycle), still not ready. 4th call: mirror check, ready.
	When(terraform.RunCommandWithVersion(Any[command.ProjectContext](), Any[string](), Any[[]string](), Any[map[string]string](), Any[tf.Distribution](), Any[*version.Version](), Any[string]())).
		ThenReturn(
			"Error: failed to install provider from registry.terraform.io: 423 Locked", errors.New("exit status 1"),
		).ThenReturn(
		"Error: no available releases match the given constraints (registry.terraform.io)", errors.New("exit status 1"),
	).ThenReturn(
		"Error: failed to install provider from registry.terraform.io: 423 Locked", errors.New("exit status 1"),
	).ThenReturn("", nil)

	path := t.TempDir()
	output, err := iso.Run(command.ProjectContext{Workspace: "workspace", RepoRelDir: ".", Log: logger}, nil, path, nil)
	Ok(t, err)
	Equals(t, "", output)

	terraform.VerifyWasCalled(Times(4)).RunCommandWithVersion(Any[command.ProjectContext](), Any[string](), Any[[]string](), Any[map[string]string](), Any[tf.Distribution](), Any[*version.Version](), Any[string]())

	mirrorFile, err := os.ReadFile(filepath.Join(path, ".terraformrc-provider-cache-mirror"))
	Ok(t, err)
	Assert(t, strings.Contains(string(mirrorFile), "filesystem_mirror"), "mirror CLI config missing filesystem_mirror block: %s", mirrorFile)
}

// Test that a phase-1 failure unrelated to the provider cache (its output
// names none of the configured registry hosts) is surfaced immediately,
// without wasting the mirror retry budget on a failure retrying can't fix.
func TestRun_ProviderCache_UnrelatedFailureSkipsMirrorPhase(t *testing.T) {
	RegisterMockTestingT(t)
	terraform := tfclientmocks.NewMockClient()
	logger := logging.NewNoopLogger(t)
	mockDownloader := mocks.NewMockDownloader()
	tfDistribution := tf.NewDistributionTerraformWithDownloader(mockDownloader)
	tfVersion, _ := version.NewVersion("1.14.0")

	iso := runtime.InitStepRunner{
		TerraformExecutor:     terraform,
		DefaultTFDistribution: tfDistribution,
		DefaultTFVersion:      tfVersion,
		ProviderCache: &tfclient.ProviderCacheConfig{
			MirrorBaseURL: "http://127.0.0.1:8080/",
			RegistryHosts: []string{"registry.terraform.io"},
			MirrorDir:     t.TempDir(),
		},
	}
	When(terraform.RunCommandWithVersion(Any[command.ProjectContext](), Any[string](), Any[[]string](), Any[map[string]string](), Any[tf.Distribution](), Any[*version.Version](), Any[string]())).
		ThenReturn("Error: Unsupported argument", errors.New("exit status 1"))

	path := t.TempDir()
	output, err := iso.Run(command.ProjectContext{Workspace: "workspace", RepoRelDir: ".", Log: logger}, nil, path, nil)
	ErrEquals(t, "exit status 1", err)
	Equals(t, "Error: Unsupported argument", output)

	terraform.VerifyWasCalledOnce().RunCommandWithVersion(Any[command.ProjectContext](), Any[string](), Any[[]string](), Any[map[string]string](), Any[tf.Distribution](), Any[*version.Version](), Any[string]())
}

// Test that once the mirror retry budget is exhausted, Run surfaces the last
// attempt's real error rather than retrying forever.
func TestRun_ProviderCache_MirrorPhaseTimesOutAndSurfacesError(t *testing.T) {
	RegisterMockTestingT(t)
	terraform := tfclientmocks.NewMockClient()
	logger := logging.NewNoopLogger(t)
	mockDownloader := mocks.NewMockDownloader()
	tfDistribution := tf.NewDistributionTerraformWithDownloader(mockDownloader)
	tfVersion, _ := version.NewVersion("1.14.0")

	iso := runtime.InitStepRunner{
		TerraformExecutor:     terraform,
		DefaultTFDistribution: tfDistribution,
		DefaultTFVersion:      tfVersion,
		ProviderCache: &tfclient.ProviderCacheConfig{
			MirrorBaseURL: "http://127.0.0.1:8080/",
			RegistryHosts: []string{"registry.terraform.io"},
			MirrorDir:     t.TempDir(),
		},
		// Shorter than the first backoff step, so the retry loop gives up
		// after its one phase-2 attempt instead of sleeping.
		ProviderCacheMirrorWaitTimeout: 1 * time.Millisecond,
	}
	When(terraform.RunCommandWithVersion(Any[command.ProjectContext](), Any[string](), Any[[]string](), Any[map[string]string](), Any[tf.Distribution](), Any[*version.Version](), Any[string]())).
		ThenReturn("Error: failed to install provider from registry.terraform.io: 423 Locked", errors.New("exit status 1"))

	path := t.TempDir()
	output, err := iso.Run(command.ProjectContext{Workspace: "workspace", RepoRelDir: ".", Log: logger}, nil, path, nil)
	ErrEquals(t, "exit status 1", err)
	Equals(t, "Error: failed to install provider from registry.terraform.io: 423 Locked", output)

	// Phase 1 plus at least one phase-2 attempt before the budget ran out.
	// With a near-instant mocked RunCommandWithVersion, exactly how many
	// phase-2 attempts fit inside a 1ms budget before the deadline check
	// fires is a wall-clock race, not something worth pinning down exactly.
	terraform.VerifyWasCalled(AtLeast(2)).RunCommandWithVersion(Any[command.ProjectContext](), Any[string](), Any[[]string](), Any[map[string]string](), Any[tf.Distribution](), Any[*version.Version](), Any[string]())
}

// Test that once the mirror retry budget is exhausted, Run queries the
// provider cache proxy directly (not through Terraform) for the real cause
// and appends it to the output, instead of only ever surfacing Terraform's
// generic "provider not found" error.
func TestRun_ProviderCache_SurfacesInstallerErrorOnTimeout(t *testing.T) {
	RegisterMockTestingT(t)
	terraform := tfclientmocks.NewMockClient()
	logger := logging.NewNoopLogger(t)
	mockDownloader := mocks.NewMockDownloader()
	tfDistribution := tf.NewDistributionTerraformWithDownloader(mockDownloader)
	tfVersion, _ := version.NewVersion("1.14.0")

	proxy := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/status/registry.terraform.io" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"errors":["registry.terraform.io/hashicorp/null/3.2.1/linux/amd64: refusing to install an unverifiable package"]}`)
	}))
	t.Cleanup(proxy.Close)

	iso := runtime.InitStepRunner{
		TerraformExecutor:     terraform,
		DefaultTFDistribution: tfDistribution,
		DefaultTFVersion:      tfVersion,
		ProviderCache: &tfclient.ProviderCacheConfig{
			MirrorBaseURL: proxy.URL + "/",
			RegistryHosts: []string{"registry.terraform.io"},
			MirrorDir:     t.TempDir(),
		},
		ProviderCacheMirrorWaitTimeout: 1 * time.Millisecond,
	}
	When(terraform.RunCommandWithVersion(Any[command.ProjectContext](), Any[string](), Any[[]string](), Any[map[string]string](), Any[tf.Distribution](), Any[*version.Version](), Any[string]())).
		ThenReturn("Error: failed to install provider from registry.terraform.io: 423 Locked", errors.New("exit status 1"))

	path := t.TempDir()
	output, err := iso.Run(command.ProjectContext{Workspace: "workspace", RepoRelDir: ".", Log: logger}, nil, path, nil)
	ErrEquals(t, "exit status 1", err)
	Assert(t, strings.Contains(output, "Error: failed to install provider"), "expected Terraform's own error to still be present, got %q", output)
	Assert(t, strings.Contains(output, "refusing to install an unverifiable package"), "expected the provider cache proxy's real error to be appended, got %q", output)
}

// Test that Run gives up well before ProviderCacheMirrorWaitTimeout when the
// provider cache proxy reports the exact same install failure on two
// consecutive retry cycles - a stuck (not merely still-transient) install -
// instead of waiting out the full (here: default, several-minute) budget.
func TestRun_ProviderCache_ShortCircuitsOnRepeatedInstallerError(t *testing.T) {
	RegisterMockTestingT(t)
	terraform := tfclientmocks.NewMockClient()
	logger := logging.NewNoopLogger(t)
	mockDownloader := mocks.NewMockDownloader()
	tfDistribution := tf.NewDistributionTerraformWithDownloader(mockDownloader)
	tfVersion, _ := version.NewVersion("1.14.0")

	// Always reports the same, stable error.
	proxy := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/status/registry.terraform.io" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"errors":["registry.terraform.io/hashicorp/null/3.2.1/linux/amd64: refusing to install an unverifiable package"]}`)
	}))
	t.Cleanup(proxy.Close)

	iso := runtime.InitStepRunner{
		TerraformExecutor:     terraform,
		DefaultTFDistribution: tfDistribution,
		DefaultTFVersion:      tfVersion,
		ProviderCache: &tfclient.ProviderCacheConfig{
			MirrorBaseURL: proxy.URL + "/",
			RegistryHosts: []string{"registry.terraform.io"},
			MirrorDir:     t.TempDir(),
		},
		// Left at the (multi-minute) default: if this test passes quickly,
		// it's because the repeated-error short-circuit fired, not because
		// the budget ran out.
	}
	When(terraform.RunCommandWithVersion(Any[command.ProjectContext](), Any[string](), Any[[]string](), Any[map[string]string](), Any[tf.Distribution](), Any[*version.Version](), Any[string]())).
		ThenReturn("Error: failed to install provider from registry.terraform.io: 423 Locked", errors.New("exit status 1"))

	path := t.TempDir()
	start := time.Now()
	output, err := iso.Run(command.ProjectContext{Workspace: "workspace", RepoRelDir: ".", Log: logger}, nil, path, nil)
	elapsed := time.Since(start)

	ErrEquals(t, "exit status 1", err)
	Assert(t, strings.Contains(output, "refusing to install an unverifiable package"), "expected the provider cache proxy's real error to be appended, got %q", output)
	// Two backoff cycles (1s + 2s) plus overhead, well under the multi-minute
	// default budget the short-circuit is meant to avoid waiting out.
	Assert(t, elapsed < 30*time.Second, "took %s; the repeated-error short-circuit does not appear to have fired", elapsed)

	// Discovery, mirror(fail), discovery(retrigger), mirror(fail),
	// discovery(retrigger) - short-circuits on the second repeat, without a
	// third mirror check.
	terraform.VerifyWasCalled(Times(5)).RunCommandWithVersion(Any[command.ProjectContext](), Any[string](), Any[[]string](), Any[map[string]string](), Any[tf.Distribution](), Any[*version.Version](), Any[string]())
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
