// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package tfclient

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	version "github.com/hashicorp/go-version"
	. "github.com/petergtz/pegomock/v4"
	runtimemodels "github.com/runatlantis/atlantis/server/core/runtime/models"
	"github.com/runatlantis/atlantis/server/core/terraform"
	terraform_mocks "github.com/runatlantis/atlantis/server/core/terraform/mocks"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/models"
	jobmocks "github.com/runatlantis/atlantis/server/jobs/mocks"
	"github.com/runatlantis/atlantis/server/logging"
	logmocks "github.com/runatlantis/atlantis/server/logging/mocks"
	. "github.com/runatlantis/atlantis/testing"
)

// Test that we write the file as expected
func TestGenerateRCFile_WritesFile(t *testing.T) {
	tmp := t.TempDir()

	err := generateRCFile("token", "hostname", nil, tmp)
	Ok(t, err)

	expContents := `credentials "hostname" {
  token = "token"
}`
	actContents, err := os.ReadFile(filepath.Join(tmp, ".terraformrc"))
	Ok(t, err)
	Equals(t, expContents, string(actContents))
}

// Test that when the provider cache proxy is configured we write a host block
// per registry, alongside the credentials block.
func TestGenerateRCFile_WritesProviderCacheHostBlocks(t *testing.T) {
	tmp := t.TempDir()

	pc := &ProviderCacheConfig{
		MirrorBaseURL: "http://127.0.0.1:8080/",
		RegistryHosts: []string{"registry.terraform.io", "registry.opentofu.org"},
	}
	err := generateRCFile("token", "hostname", pc, tmp)
	Ok(t, err)

	expContents := `credentials "hostname" {
  token = "token"
}

host "registry.terraform.io" {
  services = {
    "providers.v1" = "http://127.0.0.1:8080/registry.terraform.io/v1/providers/"
  }
}

host "registry.opentofu.org" {
  services = {
    "providers.v1" = "http://127.0.0.1:8080/registry.opentofu.org/v1/providers/"
  }
}`
	actContents, err := os.ReadFile(filepath.Join(tmp, ".terraformrc"))
	Ok(t, err)
	Equals(t, expContents, string(actContents))
}

// Test that with no TFE token and no provider cache there is nothing to write.
func TestGenerateRCFile_NoopWhenNothingConfigured(t *testing.T) {
	tmp := t.TempDir()

	err := generateRCFile("", "hostname", nil, tmp)
	Ok(t, err)

	_, err = os.Stat(filepath.Join(tmp, ".terraformrc"))
	Assert(t, os.IsNotExist(err), "expected no .terraformrc to be written")
}

// Test that the provider cache host block can be written without a TFE token.
func TestGenerateRCFile_ProviderCacheOnly(t *testing.T) {
	tmp := t.TempDir()

	pc := &ProviderCacheConfig{
		MirrorBaseURL: "http://127.0.0.1:9999/",
		RegistryHosts: []string{"registry.terraform.io"},
	}
	err := generateRCFile("", "hostname", pc, tmp)
	Ok(t, err)

	expContents := `host "registry.terraform.io" {
  services = {
    "providers.v1" = "http://127.0.0.1:9999/registry.terraform.io/v1/providers/"
  }
}`
	actContents, err := os.ReadFile(filepath.Join(tmp, ".terraformrc"))
	Ok(t, err)
	Equals(t, expContents, string(actContents))
}

// Test that if the file already exists and its contents will be modified if
// we write our config that we error out.
func TestGenerateRCFile_WillNotOverwrite(t *testing.T) {
	tmp := t.TempDir()

	rcFile := filepath.Join(tmp, ".terraformrc")
	err := os.WriteFile(rcFile, []byte("contents"), 0600)
	Ok(t, err)

	actErr := generateRCFile("token", "hostname", nil, tmp)
	expErr := fmt.Sprintf("can't write Terraform CLI config to %s because that file has contents that would be overwritten", tmp+"/.terraformrc")
	ErrEquals(t, expErr, actErr)
}

// Test that if the file already exists and its contents will NOT be modified if
// we write our config that we don't error.
func TestGenerateRCFile_NoErrIfContentsSame(t *testing.T) {
	tmp := t.TempDir()

	rcFile := filepath.Join(tmp, ".terraformrc")
	contents := `credentials "app.terraform.io" {
  token = "token"
}`
	err := os.WriteFile(rcFile, []byte(contents), 0600)
	Ok(t, err)

	err = generateRCFile("token", "app.terraform.io", nil, tmp)
	Ok(t, err)
}

// Test that if we can't read the existing file to see if the contents will be
// the same that we just error out.
func TestGenerateRCFile_ErrIfCannotRead(t *testing.T) {
	tmp := t.TempDir()

	rcFile := filepath.Join(tmp, ".terraformrc")
	err := os.WriteFile(rcFile, []byte("can't see me!"), 0000)
	Ok(t, err)

	expErr := fmt.Sprintf("trying to read %s to ensure we're not overwriting it: open %s: permission denied", rcFile, rcFile)
	actErr := generateRCFile("token", "hostname", nil, tmp)
	ErrEquals(t, expErr, actErr)
}

// Test that if we can't write, we error out.
func TestGenerateRCFile_ErrIfCannotWrite(t *testing.T) {
	rcFile := "/this/dir/does/not/exist/.terraformrc"
	expErr := fmt.Sprintf("writing generated .terraformrc file to %s: open %s: no such file or directory", rcFile, rcFile)
	actErr := generateRCFile("token", "hostname", nil, "/this/dir/does/not/exist")
	ErrEquals(t, expErr, actErr)
}

// Test that it executes with the expected env vars.
func TestDefaultClient_RunCommandWithVersion_EnvVars(t *testing.T) {
	v, err := version.NewVersion("0.11.11")
	Ok(t, err)
	tmp := t.TempDir()
	logger := logging.NewNoopLogger(t)
	projectCmdOutputHandler := jobmocks.NewMockProjectCommandOutputHandler()

	ctx := command.ProjectContext{
		Log:                logger,
		Workspace:          "default",
		RepoRelDir:         ".",
		User:               models.User{Username: "username"},
		EscapedCommentArgs: []string{"comment", "args"},
		ProjectName:        "projectname",
		Pull: models.PullRequest{
			Num: 2,
		},
	}
	client := &DefaultClient{
		defaultVersion:          v,
		terraformPluginCacheDir: tmp,
		overrideTF:              "echo",
		usePluginCache:          true,
		projectCmdOutputHandler: projectCmdOutputHandler,
	}

	args := []string{
		"TF_IN_AUTOMATION=$TF_IN_AUTOMATION",
		"TF_PLUGIN_CACHE_DIR=$TF_PLUGIN_CACHE_DIR",
		"WORKSPACE=$WORKSPACE",
		"ATLANTIS_TERRAFORM_VERSION=$ATLANTIS_TERRAFORM_VERSION",
		"DIR=$DIR",
	}
	// These stand in for operator-configured extra_args, which are the only
	// arguments whose variable references are expanded.
	ctx.ExpandableArgs = args
	customEnvVars := map[string]string{}
	mockDownloader := terraform_mocks.NewMockDownloader()
	distribution := terraform.NewDistributionTerraformWithDownloader(mockDownloader)
	out, err := client.RunCommandWithVersion(ctx, tmp, args, customEnvVars, distribution, nil, "workspace")
	Ok(t, err)
	exp := fmt.Sprintf("TF_IN_AUTOMATION=true TF_PLUGIN_CACHE_DIR=%s WORKSPACE=workspace ATLANTIS_TERRAFORM_VERSION=0.11.11 DIR=%s\n", tmp, tmp)
	Equals(t, exp, out)
}

// Test that it returns an error on error.
func TestDefaultClient_RunCommandWithVersion_Error(t *testing.T) {
	v, err := version.NewVersion("0.11.11")
	Ok(t, err)
	tmp := t.TempDir()
	logger := logging.NewNoopLogger(t)
	projectCmdOutputHandler := jobmocks.NewMockProjectCommandOutputHandler()

	ctx := command.ProjectContext{
		Log:                logger,
		Workspace:          "default",
		RepoRelDir:         ".",
		User:               models.User{Username: "username"},
		EscapedCommentArgs: []string{"comment", "args"},
		ProjectName:        "projectname",
		Pull: models.PullRequest{
			Num: 2,
		},
		BaseRepo: models.Repo{
			FullName: "owner/repo",
			Owner:    "owner",
			Name:     "repo",
		},
	}
	client := &DefaultClient{
		defaultVersion:          v,
		terraformPluginCacheDir: tmp,
		overrideTF:              "sh",
		projectCmdOutputHandler: projectCmdOutputHandler,
	}

	// Arguments are passed to the process as a vector, so a command that needs
	// shell operators has to ask for a shell explicitly.
	args := []string{
		"-c",
		"echo dying; exit 1",
	}
	mockDownloader := terraform_mocks.NewMockDownloader()
	distribution := terraform.NewDistributionTerraformWithDownloader(mockDownloader)
	out, err := client.RunCommandWithVersion(ctx, tmp, args, map[string]string{}, distribution, nil, "workspace")
	ErrEquals(t, fmt.Sprintf(`running 'sh -c echo dying; exit 1' in '%s': exit status 1`, tmp), err)
	// Test that we still get our output.
	Equals(t, "dying\n", out)
}

func TestDefaultClient_RunCommandAsync_Success(t *testing.T) {
	RegisterMockTestingT(t)
	v, err := version.NewVersion("0.11.11")
	Ok(t, err)
	tmp := t.TempDir()
	logger := logmocks.NewMockSimpleLogging()
	When(logger.With(Any[string](), Any[any]())).ThenReturn(logger)
	projectCmdOutputHandler := jobmocks.NewMockProjectCommandOutputHandler()

	ctx := command.ProjectContext{
		Log:                logger,
		Workspace:          "default",
		RepoRelDir:         ".",
		User:               models.User{Username: "username"},
		EscapedCommentArgs: []string{"comment", "args"},
		ProjectName:        "projectname",
		Pull: models.PullRequest{
			Num: 2,
		},
		BaseRepo: models.Repo{
			FullName: "owner/repo",
			Owner:    "owner",
			Name:     "repo",
		},
	}
	client := &DefaultClient{
		defaultVersion:          v,
		terraformPluginCacheDir: tmp,
		overrideTF:              "echo",
		usePluginCache:          true,
		projectCmdOutputHandler: projectCmdOutputHandler,
	}

	args := []string{
		"TF_IN_AUTOMATION=$TF_IN_AUTOMATION",
		"TF_PLUGIN_CACHE_DIR=$TF_PLUGIN_CACHE_DIR",
		"WORKSPACE=$WORKSPACE",
		"ATLANTIS_TERRAFORM_VERSION=$ATLANTIS_TERRAFORM_VERSION",
		"DIR=$DIR",
	}
	// These stand in for operator-configured extra_args.
	ctx.ExpandableArgs = args
	mockDownloader := terraform_mocks.NewMockDownloader()
	distribution := terraform.NewDistributionTerraformWithDownloader(mockDownloader)
	_, outCh := client.RunCommandAsync(ctx, tmp, args, map[string]string{}, distribution, nil, "workspace")

	out, err := waitCh(outCh)
	Ok(t, err)
	exp := fmt.Sprintf("TF_IN_AUTOMATION=true TF_PLUGIN_CACHE_DIR=%s WORKSPACE=workspace ATLANTIS_TERRAFORM_VERSION=0.11.11 DIR=%s", tmp, tmp)
	Equals(t, exp, out)

	logger.VerifyWasCalledOnce().With(Eq("duration"), Any[any]())
}

func TestDefaultClient_RunCommandAsyncSuppressesProjectOutputHandler(t *testing.T) {
	RegisterMockTestingT(t)
	v, err := version.NewVersion("0.11.11")
	Ok(t, err)
	tmp := t.TempDir()
	logger := logmocks.NewMockSimpleLogging()
	When(logger.With(Any[string](), Any[any]())).ThenReturn(logger)
	projectCmdOutputHandler := jobmocks.NewMockProjectCommandOutputHandler()
	ctx := command.ProjectContext{
		Log:               logger,
		Workspace:         "default",
		RepoRelDir:        ".",
		SuppressJobOutput: true,
	}
	client := &DefaultClient{
		defaultVersion:          v,
		terraformPluginCacheDir: tmp,
		overrideTF:              "echo",
		projectCmdOutputHandler: projectCmdOutputHandler,
	}
	mockDownloader := terraform_mocks.NewMockDownloader()
	distribution := terraform.NewDistributionTerraformWithDownloader(mockDownloader)

	_, outCh := client.RunCommandAsync(ctx, tmp, []string{"hello"}, map[string]string{}, distribution, nil, "default")
	out, err := waitCh(outCh)
	Ok(t, err)
	Equals(t, "hello", out)
	projectCmdOutputHandler.VerifyWasCalled(Never()).Send(Any[command.ProjectContext](), Any[string](), Any[bool]())
	logger.VerifyWasCalledOnce().With(Eq("duration"), Any[any]())
}

func TestDefaultClient_RunCommandAsync_BigOutput(t *testing.T) {
	RegisterMockTestingT(t)
	v, err := version.NewVersion("0.11.11")
	Ok(t, err)
	tmp := t.TempDir()
	logger := logmocks.NewMockSimpleLogging()
	When(logger.With(Any[string](), Any[any]())).ThenReturn(logger)
	projectCmdOutputHandler := jobmocks.NewMockProjectCommandOutputHandler()

	ctx := command.ProjectContext{
		Log:                logger,
		Workspace:          "default",
		RepoRelDir:         ".",
		User:               models.User{Username: "username"},
		EscapedCommentArgs: []string{"comment", "args"},
		ProjectName:        "projectname",
		Pull: models.PullRequest{
			Num: 2,
		},
		BaseRepo: models.Repo{
			FullName: "owner/repo",
			Owner:    "owner",
			Name:     "repo",
		},
	}
	client := &DefaultClient{
		defaultVersion:          v,
		terraformPluginCacheDir: tmp,
		overrideTF:              "cat",
		projectCmdOutputHandler: projectCmdOutputHandler,
	}
	filename := filepath.Join(tmp, "data")
	f, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	Ok(t, err)

	var exp strings.Builder
	for range 1024 {
		s := strings.Repeat("0", 10) + "\n"
		exp.WriteString(s)
		_, err = f.WriteString(s)
		Ok(t, err)
	}
	mockDownloader := terraform_mocks.NewMockDownloader()
	distribution := terraform.NewDistributionTerraformWithDownloader(mockDownloader)
	_, outCh := client.RunCommandAsync(ctx, tmp, []string{filename}, map[string]string{}, distribution, nil, "workspace")

	out, err := waitCh(outCh)
	Ok(t, err)
	Equals(t, strings.TrimRight(exp.String(), "\n"), out)

	logger.VerifyWasCalledOnce().With(Eq("duration"), Any[any]())
}

func TestDefaultClient_RunCommandAsync_StderrOutput(t *testing.T) {
	RegisterMockTestingT(t)
	v, err := version.NewVersion("0.11.11")
	Ok(t, err)
	tmp := t.TempDir()
	logger := logmocks.NewMockSimpleLogging()
	When(logger.With(Any[string](), Any[any]())).ThenReturn(logger)
	projectCmdOutputHandler := jobmocks.NewMockProjectCommandOutputHandler()

	ctx := command.ProjectContext{
		Log:                logger,
		Workspace:          "default",
		RepoRelDir:         ".",
		User:               models.User{Username: "username"},
		EscapedCommentArgs: []string{"comment", "args"},
		ProjectName:        "projectname",
		Pull: models.PullRequest{
			Num: 2,
		},
		BaseRepo: models.Repo{
			FullName: "owner/repo",
			Owner:    "owner",
			Name:     "repo",
		},
	}
	client := &DefaultClient{
		defaultVersion:          v,
		terraformPluginCacheDir: tmp,
		overrideTF:              "sh",
		projectCmdOutputHandler: projectCmdOutputHandler,
	}
	mockDownloader := terraform_mocks.NewMockDownloader()
	distribution := terraform.NewDistributionTerraformWithDownloader(mockDownloader)
	_, outCh := client.RunCommandAsync(ctx, tmp, []string{"-c", "echo stderr >&2"}, map[string]string{}, distribution, nil, "workspace")

	out, err := waitCh(outCh)
	Ok(t, err)
	Equals(t, "stderr", out)

	logger.VerifyWasCalledOnce().With(Eq("duration"), Any[any]())
}

func TestDefaultClient_RunCommandAsync_ExitOne(t *testing.T) {
	RegisterMockTestingT(t)
	v, err := version.NewVersion("0.11.11")
	Ok(t, err)
	tmp := t.TempDir()
	logger := logmocks.NewMockSimpleLogging()
	When(logger.With(Any[string](), Any[any]())).ThenReturn(logger)
	projectCmdOutputHandler := jobmocks.NewMockProjectCommandOutputHandler()

	ctx := command.ProjectContext{
		Log:                logger,
		Workspace:          "default",
		RepoRelDir:         ".",
		User:               models.User{Username: "username"},
		EscapedCommentArgs: []string{"comment", "args"},
		ProjectName:        "projectname",
		Pull: models.PullRequest{
			Num: 2,
		},
		BaseRepo: models.Repo{
			FullName: "owner/repo",
			Owner:    "owner",
			Name:     "repo",
		},
	}
	client := &DefaultClient{
		defaultVersion:          v,
		terraformPluginCacheDir: tmp,
		overrideTF:              "sh",
		projectCmdOutputHandler: projectCmdOutputHandler,
	}
	mockDownloader := terraform_mocks.NewMockDownloader()
	distribution := terraform.NewDistributionTerraformWithDownloader(mockDownloader)
	_, outCh := client.RunCommandAsync(ctx, tmp, []string{"-c", "echo dying; exit 1"}, map[string]string{}, distribution, nil, "workspace")

	out, err := waitCh(outCh)
	ErrEquals(t, fmt.Sprintf(`running 'sh -c echo dying; exit 1' in '%s': exit status 1`, tmp), err)
	// Test that we still get our output.
	Equals(t, "dying", out)

	logger.VerifyWasCalledOnce().With(Eq("duration"), Any[any]())
}

func TestDefaultClient_RunCommandAsync_Input(t *testing.T) {
	RegisterMockTestingT(t)
	v, err := version.NewVersion("0.11.11")
	Ok(t, err)
	tmp := t.TempDir()
	logger := logmocks.NewMockSimpleLogging()
	When(logger.With(Any[string](), Any[any]())).ThenReturn(logger)
	projectCmdOutputHandler := jobmocks.NewMockProjectCommandOutputHandler()

	ctx := command.ProjectContext{
		Log:                logger,
		Workspace:          "default",
		RepoRelDir:         ".",
		User:               models.User{Username: "username"},
		EscapedCommentArgs: []string{"comment", "args"},
		ProjectName:        "projectname",
		Pull: models.PullRequest{
			Num: 2,
		},
		BaseRepo: models.Repo{
			FullName: "owner/repo",
			Owner:    "owner",
			Name:     "repo",
		},
	}
	client := &DefaultClient{
		defaultVersion:          v,
		terraformPluginCacheDir: tmp,
		overrideTF:              "sh",
		projectCmdOutputHandler: projectCmdOutputHandler,
	}

	mockDownloader := terraform_mocks.NewMockDownloader()
	distribution := terraform.NewDistributionTerraformWithDownloader(mockDownloader)
	inCh, outCh := client.RunCommandAsync(ctx, tmp, []string{"-c", "head -n 1"}, map[string]string{}, distribution, nil, "workspace")
	inCh <- "echo me\n"

	out, err := waitCh(outCh)
	Ok(t, err)
	Equals(t, "echo me", out)

	logger.VerifyWasCalledOnce().With(Eq("duration"), Any[any]())
}

func waitCh(ch <-chan runtimemodels.Line) (string, error) {
	var ls []string
	for line := range ch {
		if line.Err != nil {
			return strings.Join(ls, "\n"), line.Err
		}
		ls = append(ls, line.Line)
	}
	return strings.Join(ls, "\n"), nil
}

// Terraform arguments must reach the process as an argument vector. The
// workspace name in particular is derived from pull-request controlled input,
// and joining arguments into a string for `sh -c` made every shell
// metacharacter in them significant.
func TestDefaultClient_RunCommandWithVersion_DoesNotInterpretArgsAsShell(t *testing.T) {
	v, err := version.NewVersion("0.11.11")
	Ok(t, err)
	tmp := t.TempDir()
	logger := logging.NewNoopLogger(t)
	projectCmdOutputHandler := jobmocks.NewMockProjectCommandOutputHandler()

	ctx := command.ProjectContext{
		Log:        logger,
		Workspace:  "default",
		RepoRelDir: ".",
		User:       models.User{Username: "username"},
		Pull:       models.PullRequest{Num: 2},
	}
	client := &DefaultClient{
		defaultVersion:          v,
		terraformPluginCacheDir: tmp,
		overrideTF:              "echo",
		projectCmdOutputHandler: projectCmdOutputHandler,
	}

	mockDownloader := terraform_mocks.NewMockDownloader()
	distribution := terraform.NewDistributionTerraformWithDownloader(mockDownloader)

	// "workspace" is not log-streaming eligible, so this exercises the
	// synchronous path, which is the one the workspace step runner uses.
	out, err := client.RunCommandWithVersion(ctx, tmp, []string{"workspace", "select", "poc;id;false;#"},
		map[string]string{}, distribution, nil, "default")

	Ok(t, err)
	Equals(t, "workspace select poc;id;false;#\n", out)
	Assert(t, !strings.Contains(out, "uid="), "arguments were interpreted by a shell, output was %q", out)
}

func TestDefaultClient_RunCommandAsync_DoesNotInterpretArgsAsShell(t *testing.T) {
	RegisterMockTestingT(t)
	v, err := version.NewVersion("0.11.11")
	Ok(t, err)
	tmp := t.TempDir()
	logger := logmocks.NewMockSimpleLogging()
	When(logger.With(Any[string](), Any[any]())).ThenReturn(logger)
	projectCmdOutputHandler := jobmocks.NewMockProjectCommandOutputHandler()

	ctx := command.ProjectContext{
		Log:        logger,
		Workspace:  "default",
		RepoRelDir: ".",
		User:       models.User{Username: "username"},
		Pull:       models.PullRequest{Num: 2},
		BaseRepo:   models.Repo{FullName: "owner/repo", Owner: "owner", Name: "repo"},
	}
	client := &DefaultClient{
		defaultVersion:          v,
		terraformPluginCacheDir: tmp,
		overrideTF:              "echo",
		projectCmdOutputHandler: projectCmdOutputHandler,
	}

	mockDownloader := terraform_mocks.NewMockDownloader()
	distribution := terraform.NewDistributionTerraformWithDownloader(mockDownloader)

	// "plan" is log-streaming eligible, so this exercises the streaming path.
	_, outCh := client.RunCommandAsync(ctx, tmp, []string{"plan", "-var", "x=a;id"},
		map[string]string{}, distribution, nil, "default")

	out, err := waitCh(outCh)
	Ok(t, err)
	Equals(t, "plan -var x=a;id", out)
	Assert(t, !strings.Contains(out, "uid="), "arguments were interpreted by a shell, output was %q", out)
}

// extra_args may refer to environment variables, including ones a workflow set
// via an `env` step, which reach the client as customEnvVars. Expansion is done
// by Atlantis rather than by a shell, so it has to see those too.
func TestDefaultClient_RunCommandWithVersion_ExpandsCustomEnvVars(t *testing.T) {
	v, err := version.NewVersion("0.11.11")
	Ok(t, err)
	tmp := t.TempDir()
	logger := logging.NewNoopLogger(t)
	projectCmdOutputHandler := jobmocks.NewMockProjectCommandOutputHandler()

	ctx := command.ProjectContext{
		Log:        logger,
		Workspace:  "default",
		RepoRelDir: ".",
		User:       models.User{Username: "username"},
		Pull:       models.PullRequest{Num: 2},
	}
	client := &DefaultClient{
		defaultVersion:          v,
		terraformPluginCacheDir: tmp,
		overrideTF:              "echo",
		projectCmdOutputHandler: projectCmdOutputHandler,
	}

	mockDownloader := terraform_mocks.NewMockDownloader()
	distribution := terraform.NewDistributionTerraformWithDownloader(mockDownloader)

	ctx.ExpandableArgs = []string{"-var-file", "$CUSTOM_TFVARS"}
	out, err := client.RunCommandWithVersion(ctx, tmp, []string{"-var-file", "$CUSTOM_TFVARS"},
		map[string]string{"CUSTOM_TFVARS": "staging.tfvars"}, distribution, nil, "default")

	Ok(t, err)
	Equals(t, "-var-file staging.tfvars\n", out)
}

func TestDefaultClient_RunCommandAsync_ExpandsCustomEnvVars(t *testing.T) {
	RegisterMockTestingT(t)
	v, err := version.NewVersion("0.11.11")
	Ok(t, err)
	tmp := t.TempDir()
	logger := logmocks.NewMockSimpleLogging()
	When(logger.With(Any[string](), Any[any]())).ThenReturn(logger)
	projectCmdOutputHandler := jobmocks.NewMockProjectCommandOutputHandler()

	ctx := command.ProjectContext{
		Log:        logger,
		Workspace:  "default",
		RepoRelDir: ".",
		User:       models.User{Username: "username"},
		Pull:       models.PullRequest{Num: 2},
		BaseRepo:   models.Repo{FullName: "owner/repo", Owner: "owner", Name: "repo"},
	}
	client := &DefaultClient{
		defaultVersion:          v,
		terraformPluginCacheDir: tmp,
		overrideTF:              "echo",
		projectCmdOutputHandler: projectCmdOutputHandler,
	}

	mockDownloader := terraform_mocks.NewMockDownloader()
	distribution := terraform.NewDistributionTerraformWithDownloader(mockDownloader)

	ctx.ExpandableArgs = []string{"-var-file", "$CUSTOM_TFVARS"}
	_, outCh := client.RunCommandAsync(ctx, tmp, []string{"plan", "-var-file", "$CUSTOM_TFVARS"},
		map[string]string{"CUSTOM_TFVARS": "staging.tfvars"}, distribution, nil, "default")

	out, err := waitCh(outCh)
	Ok(t, err)
	Equals(t, "plan -var-file staging.tfvars", out)
}

// The argv Atlantis executes has environment variables expanded, but the string
// Atlantis logs and puts in failing pull request comments must not: an
// extra_args entry may reference a variable holding a credential.
func TestDefaultClient_PrepCmd_DisplayKeepsVariablesUnexpanded(t *testing.T) {
	v, err := version.NewVersion("0.11.11")
	Ok(t, err)
	log := logging.NewNoopLogger(t)
	client := &DefaultClient{defaultVersion: v, overrideTF: "echo"}

	argv, display, _, err := client.prepCmd(log, nil, v, "default", "/tmp",
		[]string{"plan", "-var=token=$TF_SECRET"},
		map[string]string{"TF_SECRET": "s3cr3t-value"},
		[]string{"-var=token=$TF_SECRET"}, nil)
	Ok(t, err)

	// The process still receives the expanded value.
	Assert(t, strings.Contains(strings.Join(argv, " "), "token=s3cr3t-value"),
		"argv should carry the expanded value, got %q", argv)

	// The display string must not.
	Assert(t, !strings.Contains(display, "s3cr3t-value"),
		"display string leaked the expanded value: %q", display)
	Assert(t, strings.Contains(display, "$TF_SECRET"),
		"display string should show the unexpanded reference, got %q", display)
}

// The plan path used to be wrapped in %q because a Bitbucket Server repo owner
// can contain a space and the command line was shell source. With an argument
// vector the path is one argument regardless of what it contains, and quoting
// it would make the quotes part of the filename.
func TestDefaultClient_PrepCmd_PathWithSpaceIsASingleArgument(t *testing.T) {
	v, err := version.NewVersion("0.11.11")
	Ok(t, err)
	log := logging.NewNoopLogger(t)
	client := &DefaultClient{defaultVersion: v, overrideTF: "echo"}

	planPath := "/data/repos/bitbucket owner/repo/2/default/default.tfplan"

	argv, _, _, err := client.prepCmd(log, nil, v, "default", "/tmp",
		[]string{"plan", "-out", planPath}, nil, nil, nil)

	Ok(t, err)
	Equals(t, []string{"echo", "plan", "-out", planPath}, argv)
}

// A pull request comment must never have environment variables expanded into
// it. Terraform echoes an invalid flag value back in its error output, and
// Atlantis renders that into a pull request comment, so expanding a comment
// argument would let anyone who can comment read the Atlantis process
// environment.
func TestDefaultClient_PrepCmd_DoesNotExpandCommentArgs(t *testing.T) {
	t.Setenv("ATLANTIS_GH_TOKEN", "ghp_CANARY_SECRET_VALUE")

	v, err := version.NewVersion("0.11.11")
	Ok(t, err)
	log := logging.NewNoopLogger(t)
	client := &DefaultClient{defaultVersion: v, overrideTF: "echo"}

	commentArgs := []string{"-parallelism=$ATLANTIS_GH_TOKEN"}
	args := append([]string{"plan"}, commentArgs...)

	// No expandable args: everything is literal.
	argv, display, _, err := client.prepCmd(log, nil, v, "default", "/tmp", args, nil, nil, commentArgs)
	Ok(t, err)
	Assert(t, !strings.Contains(strings.Join(argv, " "), "ghp_CANARY_SECRET_VALUE"),
		"comment arg was expanded into the argv: %q", argv)
	Assert(t, !strings.Contains(display, "ghp_CANARY_SECRET_VALUE"),
		"comment arg was expanded into the display string: %q", display)

	// Even when an operator's extra_args happens to be the same string, the
	// comment copy is not expanded.
	argv, _, _, err = client.prepCmd(log, nil, v, "default", "/tmp", args, nil, commentArgs, commentArgs)
	Ok(t, err)
	Assert(t, !strings.Contains(strings.Join(argv, " "), "ghp_CANARY_SECRET_VALUE"),
		"comment arg was expanded when replayed as an extra_arg: %q", argv)
}

// Operator-configured extra_args keep their expansion.
func TestDefaultClient_PrepCmd_ExpandsConfiguredExtraArgs(t *testing.T) {
	v, err := version.NewVersion("0.11.11")
	Ok(t, err)
	log := logging.NewNoopLogger(t)
	client := &DefaultClient{defaultVersion: v, overrideTF: "echo"}

	extraArgs := []string{"-var-file=$WORKSPACE.tfvars"}
	args := append([]string{"plan"}, extraArgs...)

	argv, display, _, err := client.prepCmd(log, nil, v, "staging", "/tmp", args, nil, extraArgs, nil)

	Ok(t, err)
	Equals(t, []string{"echo", "plan", "-var-file=staging.tfvars"}, argv)
	// The display keeps the reference, so a variable holding a credential does
	// not reach the logs or a pull request comment.
	Assert(t, strings.Contains(display, "$WORKSPACE"),
		"display should keep the unexpanded reference, got %q", display)
}
