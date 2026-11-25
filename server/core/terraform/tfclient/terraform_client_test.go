// Copyright 2017 HootSuite Media Inc.
// SPDX-License-Identifier: Apache-2.0
// Modified hereafter by contributors to runatlantis/atlantis.

package tfclient_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	version "github.com/hashicorp/go-version"
	. "github.com/petergtz/pegomock/v4"
	"github.com/runatlantis/atlantis/cmd"
	"github.com/runatlantis/atlantis/server/core/terraform"
	"github.com/runatlantis/atlantis/server/core/terraform/mocks"
	"github.com/runatlantis/atlantis/server/core/terraform/tfclient"
	"github.com/runatlantis/atlantis/server/events/command"
	jobmocks "github.com/runatlantis/atlantis/server/jobs/mocks"
	"github.com/runatlantis/atlantis/server/logging"
	. "github.com/runatlantis/atlantis/testing"
)

func TestMustConstraint_PanicsOnBadConstraint(t *testing.T) {
	t.Log("MustConstraint should panic on a bad constraint")
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("The code did not panic")
		}
	}()

	tfclient.MustConstraint("invalid constraint")
}

func TestMustConstraint(t *testing.T) {
	t.Log("MustConstraint should return the constrain")
	c := tfclient.MustConstraint(">0.1")
	expectedConstraint, err := version.NewConstraint(">0.1")
	Ok(t, err)
	Equals(t, expectedConstraint.String(), c.String())
}

// Test that if terraform is in path and we're not setting the default-tf flag,
// that we use that version as our default version.
func TestNewClient_LocalTFOnly(t *testing.T) {
	fakeBinOut := `Terraform v0.11.10

Your version of Terraform is out of date! The latest version
is 0.11.13. You can update by downloading from developer.hashicorp.com/terraform/downloads
`
	tmp, binDir, cacheDir := mkSubDirs(t)
	projectCmdOutputHandler := jobmocks.NewMockProjectCommandOutputHandler()
	ctx := command.ProjectContext{
		Log:        logging.NewNoopLogger(t),
		Workspace:  "default",
		RepoRelDir: ".",
	}

	logger := logging.NewNoopLogger(t)

	// We're testing this by adding our own "fake" terraform binary to path that
	// outputs what would normally come from terraform version.
	err := os.WriteFile(filepath.Join(tmp, "terraform"), fmt.Appendf(nil, "#!/bin/sh\necho '%s'", fakeBinOut), 0700) // #nosec G306
	Ok(t, err)
	defer tempSetEnv(t, "PATH", fmt.Sprintf("%s:%s", tmp, os.Getenv("PATH")))()

	mockDownloader := mocks.NewMockDownloader()
	distribution := terraform.NewDistributionTerraformWithDownloader(mockDownloader)

	c, err := tfclient.NewClient(logger, distribution, binDir, cacheDir, "", "", "", cmd.DefaultTFVersionFlag, cmd.DefaultTFDownloadURL, true, true, projectCmdOutputHandler)
	Ok(t, err)

	Ok(t, err)
	Equals(t, "0.11.10", c.DefaultVersion().String())

	output, err := c.RunCommandWithVersion(ctx, tmp, []string{"terraform", "init"}, map[string]string{"test": "123"}, distribution, nil, "")
	Ok(t, err)
	Equals(t, fakeBinOut+"\n", output)
}

// Test that if terraform is in path and the default-tf flag is set to the
// same version that we don't download anything.
func TestNewClient_LocalTFMatchesFlag(t *testing.T) {
	fakeBinOut := `Terraform v0.11.10

Your version of Terraform is out of date! The latest version
is 0.11.13. You can update by downloading from developer.hashicorp.com/terraform/downloads
`
	logger := logging.NewNoopLogger(t)
	tmp, binDir, cacheDir := mkSubDirs(t)
	projectCmdOutputHandler := jobmocks.NewMockProjectCommandOutputHandler()
	ctx := command.ProjectContext{
		Log:        logging.NewNoopLogger(t),
		Workspace:  "default",
		RepoRelDir: ".",
	}

	// We're testing this by adding our own "fake" terraform binary to path that
	// outputs what would normally come from terraform version.
	err := os.WriteFile(filepath.Join(tmp, "terraform"), fmt.Appendf(nil, "#!/bin/sh\necho '%s'", fakeBinOut), 0700) // #nosec G306
	Ok(t, err)
	defer tempSetEnv(t, "PATH", fmt.Sprintf("%s:%s", tmp, os.Getenv("PATH")))()

	mockDownloader := mocks.NewMockDownloader()
	distribution := terraform.NewDistributionTerraformWithDownloader(mockDownloader)

	c, err := tfclient.NewClient(logger, distribution, binDir, cacheDir, "", "", "0.11.10", cmd.DefaultTFVersionFlag, cmd.DefaultTFDownloadURL, true, true, projectCmdOutputHandler)
	Ok(t, err)

	Ok(t, err)
	Equals(t, "0.11.10", c.DefaultVersion().String())

	output, err := c.RunCommandWithVersion(ctx, tmp, []string{"terraform", "init"}, map[string]string{}, distribution, nil, "")
	Ok(t, err)
	Equals(t, fakeBinOut+"\n", output)
}

// Test that if terraform is not in PATH and we didn't set the default-tf flag
// that we error.
func TestNewClient_NoTF(t *testing.T) {
	logger := logging.NewNoopLogger(t)
	tmp, binDir, cacheDir := mkSubDirs(t)
	projectCmdOutputHandler := jobmocks.NewMockProjectCommandOutputHandler()

	// Set PATH to only include our empty directory.
	defer tempSetEnv(t, "PATH", tmp)()

	mockDownloader := mocks.NewMockDownloader()
	distribution := terraform.NewDistributionTerraformWithDownloader(mockDownloader)

	_, err := tfclient.NewClient(logger, distribution, binDir, cacheDir, "", "", "", cmd.DefaultTFVersionFlag, cmd.DefaultTFDownloadURL, true, true, projectCmdOutputHandler)
	ErrEquals(t, "terraform not found in $PATH. Set --default-tf-version or download terraform from https://developer.hashicorp.com/terraform/downloads", err)
}

// Test that if the default-tf flag is set and that binary is in our PATH
// that we use it.
func TestNewClient_DefaultTFFlagInPath(t *testing.T) {
	fakeBinOut := "Terraform v0.11.10\n"
	logger := logging.NewNoopLogger(t)
	tmp, binDir, cacheDir := mkSubDirs(t)
	projectCmdOutputHandler := jobmocks.NewMockProjectCommandOutputHandler()
	ctx := command.ProjectContext{
		Log:        logging.NewNoopLogger(t),
		Workspace:  "default",
		RepoRelDir: ".",
	}

	// We're testing this by adding our own "fake" terraform binary to path that
	// outputs what would normally come from terraform version.
	err := os.WriteFile(filepath.Join(tmp, "terraform0.11.10"), fmt.Appendf(nil, "#!/bin/sh\necho '%s'", fakeBinOut), 0700) // #nosec G306
	Ok(t, err)
	defer tempSetEnv(t, "PATH", fmt.Sprintf("%s:%s", tmp, os.Getenv("PATH")))()

	mockDownloader := mocks.NewMockDownloader()
	distribution := terraform.NewDistributionTerraformWithDownloader(mockDownloader)

	c, err := tfclient.NewClient(logger, distribution, binDir, cacheDir, "", "", "0.11.10", cmd.DefaultTFVersionFlag, cmd.DefaultTFDownloadURL, false, true, projectCmdOutputHandler)
	Ok(t, err)

	Ok(t, err)
	Equals(t, "0.11.10", c.DefaultVersion().String())

	output, err := c.RunCommandWithVersion(ctx, tmp, []string{"terraform", "init"}, map[string]string{}, distribution, nil, "")
	Ok(t, err)
	Equals(t, fakeBinOut+"\n", output)
}

// Test that if the default-tf flag is set and that binary is in our download
// bin dir that we use it.
func TestNewClient_DefaultTFFlagInBinDir(t *testing.T) {
	fakeBinOut := "Terraform v0.11.10\n"
	tmp, binDir, cacheDir := mkSubDirs(t)
	projectCmdOutputHandler := jobmocks.NewMockProjectCommandOutputHandler()
	ctx := command.ProjectContext{
		Log:        logging.NewNoopLogger(t),
		Workspace:  "default",
		RepoRelDir: ".",
	}

	// Add our fake binary to {datadir}/bin/terraform{version}.
	err := os.WriteFile(filepath.Join(binDir, "terraform0.11.10"), fmt.Appendf(nil, "#!/bin/sh\necho '%s'", fakeBinOut), 0700) // #nosec G306
	Ok(t, err)
	defer tempSetEnv(t, "PATH", fmt.Sprintf("%s:%s", tmp, os.Getenv("PATH")))()

	mockDownloader := mocks.NewMockDownloader()
	distribution := terraform.NewDistributionTerraformWithDownloader(mockDownloader)

	c, err := tfclient.NewClient(logging.NewNoopLogger(t), distribution, binDir, cacheDir, "", "", "0.11.10", cmd.DefaultTFVersionFlag, cmd.DefaultTFDownloadURL, true, true, projectCmdOutputHandler)
	Ok(t, err)

	Ok(t, err)
	Equals(t, "0.11.10", c.DefaultVersion().String())

	output, err := c.RunCommandWithVersion(ctx, tmp, []string{"terraform", "init"}, map[string]string{}, distribution, nil, "")
	Ok(t, err)
	Equals(t, fakeBinOut+"\n", output)
}

// Test that if we don't have that version of TF that we download it.
func TestNewClient_DefaultTFFlagDownload(t *testing.T) {
	RegisterMockTestingT(t)
	logger := logging.NewNoopLogger(t)
	tmp, binDir, cacheDir := mkSubDirs(t)
	projectCmdOutputHandler := jobmocks.NewMockProjectCommandOutputHandler()
	ctx := command.ProjectContext{
		Log:        logging.NewNoopLogger(t),
		Workspace:  "default",
		RepoRelDir: ".",
	}

	// Set PATH to empty so there's no TF available.
	orig := os.Getenv("PATH")
	defer tempSetEnv(t, "PATH", "")()

	mockDownloader := mocks.NewMockDownloader()
	When(mockDownloader.Install(Any[context.Context](), Any[string](), Any[string](), Any[*version.Version]())).Then(func(params []Param) ReturnValues {
		binPath := filepath.Join(params[1].(string), "terraform0.11.10")
		err := os.WriteFile(binPath, []byte("#!/bin/sh\necho '\nTerraform v0.11.10\n'"), 0700) // #nosec G306
		return []ReturnValue{binPath, err}
	})
	distribution := terraform.NewDistributionTerraformWithDownloader(mockDownloader)
	c, err := tfclient.NewClient(logger, distribution, binDir, cacheDir, "", "", "0.11.10", cmd.DefaultTFVersionFlag, cmd.DefaultTFDownloadURL, true, true, projectCmdOutputHandler)
	Ok(t, err)

	Ok(t, err)
	Equals(t, "0.11.10", c.DefaultVersion().String())

	mockDownloader.VerifyWasCalledEventually(Once(), 2*time.Second).Install(context.Background(), binDir, cmd.DefaultTFDownloadURL, version.Must(version.NewVersion("0.11.10")))

	// Reset PATH so that it has sh.
	Ok(t, os.Setenv("PATH", orig))

	output, err := c.RunCommandWithVersion(ctx, tmp, []string{"terraform", "init"}, map[string]string{}, distribution, nil, "")
	Ok(t, err)
	Equals(t, "\nTerraform v0.11.10\n\n", output)
}

// Test that we get an error if the terraform version flag is malformed.
func TestNewClient_BadVersion(t *testing.T) {
	logger := logging.NewNoopLogger(t)
	_, binDir, cacheDir := mkSubDirs(t)
	projectCmdOutputHandler := jobmocks.NewMockProjectCommandOutputHandler()
	mockDownloader := mocks.NewMockDownloader()
	distribution := terraform.NewDistributionTerraformWithDownloader(mockDownloader)
	_, err := tfclient.NewClient(logger, distribution, binDir, cacheDir, "", "", "malformed", cmd.DefaultTFVersionFlag, cmd.DefaultTFDownloadURL, true, true, projectCmdOutputHandler)
	ErrEquals(t, "malformed version: malformed", err)
}

// Test that if we run a command with a version we don't have, we download it.
func TestRunCommandWithVersion_DLsTF(t *testing.T) {
	logger := logging.NewNoopLogger(t)
	RegisterMockTestingT(t)
	tmp, binDir, cacheDir := mkSubDirs(t)
	projectCmdOutputHandler := jobmocks.NewMockProjectCommandOutputHandler()
	ctx := command.ProjectContext{
		Log:        logging.NewNoopLogger(t),
		Workspace:  "default",
		RepoRelDir: ".",
	}

	v, err := version.NewVersion("99.99.99")
	Ok(t, err)

	mockDownloader := mocks.NewMockDownloader()
	distribution := terraform.NewDistributionTerraformWithDownloader(mockDownloader)
	// Set up our mock downloader to write a fake tf binary when it's called.
	When(mockDownloader.Install(context.Background(), binDir, cmd.DefaultTFDownloadURL, v)).Then(func(params []Param) ReturnValues {
		binPath := filepath.Join(params[1].(string), "terraform99.99.99")
		err := os.WriteFile(binPath, []byte("#!/bin/sh\necho '\nTerraform v99.99.99\n'"), 0700) // #nosec G306
		return []ReturnValue{binPath, err}
	})

	c, err := tfclient.NewClient(logger, distribution, binDir, cacheDir, "", "", "0.11.10", cmd.DefaultTFVersionFlag, cmd.DefaultTFDownloadURL, true, true, projectCmdOutputHandler)
	Ok(t, err)
	Equals(t, "0.11.10", c.DefaultVersion().String())

	output, err := c.RunCommandWithVersion(ctx, tmp, []string{"terraform", "init"}, map[string]string{}, distribution, v, "")

	Assert(t, err == nil, "err: %s: %s", err, output)
	Equals(t, "\nTerraform v99.99.99\n\n", output)
}

func TestRunCommandWithVersion_RedownloadsBrokenManagedBinary(t *testing.T) {
	logger := logging.NewNoopLogger(t)
	RegisterMockTestingT(t)
	tmp, binDir, cacheDir := mkSubDirs(t)
	projectCmdOutputHandler := jobmocks.NewMockProjectCommandOutputHandler()
	ctx := command.ProjectContext{
		Log:        logging.NewNoopLogger(t),
		Workspace:  "default",
		RepoRelDir: ".",
	}

	v, err := version.NewVersion("99.99.99")
	Ok(t, err)

	pathDir := filepath.Join(tmp, "path")
	Ok(t, os.MkdirAll(pathDir, 0700))
	Ok(t, writeExecutable(filepath.Join(pathDir, "terraform"), "echo '\nTerraform v0.11.10\n'"))
	defer tempSetEnv(t, "PATH", fmt.Sprintf("%s:%s", pathDir, os.Getenv("PATH")))()

	binPath := filepath.Join(binDir, "terraform99.99.99")
	Ok(t, writeExecutable(binPath, "exit 1"))

	mockDownloader := mocks.NewMockDownloader()
	distribution := terraform.NewDistributionTerraformWithDownloader(mockDownloader)
	When(mockDownloader.Install(context.Background(), binDir, cmd.DefaultTFDownloadURL, v)).Then(func(params []Param) ReturnValues {
		binPath := filepath.Join(params[1].(string), "terraform99.99.99")
		err := writeExecutable(binPath, "echo '\nTerraform v99.99.99\n'")
		return []ReturnValue{binPath, err}
	})

	c, err := tfclient.NewClient(logger, distribution, binDir, cacheDir, "", "", "", cmd.DefaultTFVersionFlag, cmd.DefaultTFDownloadURL, true, true, projectCmdOutputHandler)
	Ok(t, err)

	output, err := c.RunCommandWithVersion(ctx, tmp, []string{"terraform", "init"}, map[string]string{}, distribution, v, "")

	Assert(t, err == nil, "err: %s: %s", err, output)
	Equals(t, "\nTerraform v99.99.99\n\n", output)
	mockDownloader.VerifyWasCalledEventually(Once(), 2*time.Second).Install(context.Background(), binDir, cmd.DefaultTFDownloadURL, v)
}

func TestRunCommandWithVersion_UsesClientDistributionWhenArgNil(t *testing.T) {
	logger := logging.NewNoopLogger(t)
	tmp, binDir, cacheDir := mkSubDirs(t)
	projectCmdOutputHandler := jobmocks.NewMockProjectCommandOutputHandler()
	ctx := command.ProjectContext{
		Log:        logging.NewNoopLogger(t),
		Workspace:  "default",
		RepoRelDir: ".",
	}

	pathDir := filepath.Join(tmp, "path")
	Ok(t, os.MkdirAll(pathDir, 0700))
	Ok(t, writeExecutable(filepath.Join(pathDir, "terraform"), "echo '\nTerraform v0.11.10\n'"))
	defer tempSetEnv(t, "PATH", fmt.Sprintf("%s:%s", pathDir, os.Getenv("PATH")))()

	v, err := version.NewVersion("99.99.99")
	Ok(t, err)
	Ok(t, writeExecutable(filepath.Join(binDir, "terraform99.99.99"), "echo '\nTerraform v99.99.99\n'"))

	distribution := terraform.NewDistributionTerraformWithDownloader(mocks.NewMockDownloader())
	c, err := tfclient.NewClient(logger, distribution, binDir, cacheDir, "", "", "", cmd.DefaultTFVersionFlag, cmd.DefaultTFDownloadURL, false, true, projectCmdOutputHandler)
	Ok(t, err)

	output, err := c.RunCommandWithVersion(ctx, tmp, []string{"version"}, map[string]string{}, nil, v, "")

	Assert(t, err == nil, "err: %s: %s", err, output)
	Equals(t, "\nTerraform v99.99.99\n\n", output)
}

func TestRunCommandWithVersion_ValidatesVersionWithoutInheritedCLIArgs(t *testing.T) {
	logger := logging.NewNoopLogger(t)
	RegisterMockTestingT(t)
	tmp, binDir, cacheDir := mkSubDirs(t)
	projectCmdOutputHandler := jobmocks.NewMockProjectCommandOutputHandler()
	ctx := command.ProjectContext{
		Log:        logging.NewNoopLogger(t),
		Workspace:  "default",
		RepoRelDir: ".",
	}

	pathDir := filepath.Join(tmp, "path")
	Ok(t, os.MkdirAll(pathDir, 0700))
	Ok(t, writeExecutable(filepath.Join(pathDir, "terraform"), "echo 'Terraform v0.11.10'"))
	defer tempSetEnv(t, "PATH", fmt.Sprintf("%s:%s", pathDir, os.Getenv("PATH")))()
	defer tempSetEnv(t, "TF_CLI_ARGS", "-json")()
	defer tempSetEnv(t, "TF_CLI_ARGS_version", "-json")()

	v, err := version.NewVersion("99.99.99")
	Ok(t, err)
	Ok(t, writeExecutable(filepath.Join(binDir, "terraform99.99.99"), `if [ "$1" = "version" ]; then
	if [ -n "${TF_CLI_ARGS:-}" ] || [ -n "${TF_CLI_ARGS_version:-}" ]; then
		echo "poisoned cli args"
		exit 1
	fi
	echo 'Terraform v99.99.99'
	exit 0
fi
echo ran "$1"`))

	mockDownloader := mocks.NewMockDownloader()
	distribution := terraform.NewDistributionTerraformWithDownloader(mockDownloader)
	c, err := tfclient.NewClient(logger, distribution, binDir, cacheDir, "", "", "", cmd.DefaultTFVersionFlag, cmd.DefaultTFDownloadURL, false, true, projectCmdOutputHandler)
	Ok(t, err)

	output, err := c.RunCommandWithVersion(ctx, tmp, []string{"plan"}, map[string]string{}, distribution, v, "")

	Assert(t, err == nil, "err: %s: %s", err, output)
	Equals(t, "ran plan\n", output)
	mockDownloader.VerifyWasCalled(Never())
}

func TestRunCommandWithVersion_DoesNotHoldVersionLockDuringValidation(t *testing.T) {
	logger := logging.NewNoopLogger(t)
	tmp, binDir, cacheDir := mkSubDirs(t)
	projectCmdOutputHandler := jobmocks.NewMockProjectCommandOutputHandler()
	ctx := command.ProjectContext{
		Log:        logging.NewNoopLogger(t),
		Workspace:  "default",
		RepoRelDir: ".",
	}

	pathDir := filepath.Join(tmp, "path")
	Ok(t, os.MkdirAll(pathDir, 0700))
	Ok(t, writeExecutable(filepath.Join(pathDir, "terraform"), "echo '\nTerraform v0.11.10\n'"))
	defer tempSetEnv(t, "PATH", fmt.Sprintf("%s:%s", pathDir, os.Getenv("PATH")))()

	slowVersion, err := version.NewVersion("99.99.99")
	Ok(t, err)
	fastVersion, err := version.NewVersion("98.98.98")
	Ok(t, err)

	startedFile := filepath.Join(tmp, "slow-validation-started")
	Ok(t, writeExecutable(
		filepath.Join(binDir, "terraform99.99.99"),
		fmt.Sprintf("printf started > %q\nsleep 2\nexit 1", startedFile),
	))
	Ok(t, writeExecutable(filepath.Join(binDir, "terraform98.98.98"), "echo '\nTerraform v98.98.98\n'"))

	mockDownloader := mocks.NewMockDownloader()
	distribution := terraform.NewDistributionTerraformWithDownloader(mockDownloader)
	c, err := tfclient.NewClient(logger, distribution, binDir, cacheDir, "", "", "", cmd.DefaultTFVersionFlag, cmd.DefaultTFDownloadURL, false, true, projectCmdOutputHandler)
	Ok(t, err)

	slowDone := make(chan error, 1)
	go func() {
		_, err := c.RunCommandWithVersion(ctx, tmp, []string{"version"}, map[string]string{}, distribution, slowVersion, "")
		slowDone <- err
	}()
	waitForFile(t, startedFile, time.Second)

	fastDone := make(chan error, 1)
	go func() {
		output, err := c.RunCommandWithVersion(ctx, tmp, []string{"version"}, map[string]string{}, distribution, fastVersion, "")
		if err != nil {
			fastDone <- err
			return
		}
		if output != "\nTerraform v98.98.98\n\n" {
			fastDone <- fmt.Errorf("unexpected output: %q", output)
			return
		}
		fastDone <- nil
	}()

	var fastErr error
	timedOut := false
	select {
	case fastErr = <-fastDone:
	case <-time.After(750 * time.Millisecond):
		timedOut = true
	}

	var slowErr error
	select {
	case slowErr = <-slowDone:
	case <-time.After(3 * time.Second):
		t.Fatal("slow validation did not finish")
	}
	ErrContains(t, "failed to execute", slowErr)

	if timedOut {
		select {
		case fastErr = <-fastDone:
		case <-time.After(time.Second):
		}
		t.Fatalf("second Terraform command blocked behind slow version validation: %v", fastErr)
	}
	Ok(t, fastErr)
}

func TestRunCommandWithVersion_SerializesConcurrentInstallsForSameVersion(t *testing.T) {
	logger := logging.NewNoopLogger(t)
	tmp, binDir, cacheDir := mkSubDirs(t)
	projectCmdOutputHandler := jobmocks.NewMockProjectCommandOutputHandler()
	ctx := command.ProjectContext{
		Log:        logging.NewNoopLogger(t),
		Workspace:  "default",
		RepoRelDir: ".",
	}

	pathDir := filepath.Join(tmp, "path")
	Ok(t, os.MkdirAll(pathDir, 0700))
	Ok(t, writeExecutable(filepath.Join(pathDir, "terraform"), "echo '\nTerraform v0.11.10\n'"))
	defer tempSetEnv(t, "PATH", fmt.Sprintf("%s:%s", pathDir, os.Getenv("PATH")))()

	v, err := version.NewVersion("99.99.99")
	Ok(t, err)

	downloader := &trackingDownloader{delay: 200 * time.Millisecond}
	distribution := terraform.NewDistributionTerraformWithDownloader(downloader)
	c, err := tfclient.NewClient(logger, distribution, binDir, cacheDir, "", "", "", cmd.DefaultTFVersionFlag, cmd.DefaultTFDownloadURL, true, true, projectCmdOutputHandler)
	Ok(t, err)

	errCh := make(chan error, 2)
	for range 2 {
		go func() {
			output, err := c.RunCommandWithVersion(ctx, tmp, []string{"version"}, map[string]string{}, distribution, v, "")
			if err != nil {
				errCh <- err
				return
			}
			if output != "\nTerraform v99.99.99\n\n" {
				errCh <- fmt.Errorf("unexpected output: %q", output)
				return
			}
			errCh <- nil
		}()
	}

	for range 2 {
		Ok(t, <-errCh)
	}
	Equals(t, int32(1), downloader.maxConcurrent.Load())
	Equals(t, int32(1), downloader.calls.Load())
}

func TestRunCommandWithVersion_SerializesConcurrentDownloadsAcrossVersions(t *testing.T) {
	logger := logging.NewNoopLogger(t)
	tmp, binDir, cacheDir := mkSubDirs(t)
	projectCmdOutputHandler := jobmocks.NewMockProjectCommandOutputHandler()
	ctx := command.ProjectContext{
		Log:        logging.NewNoopLogger(t),
		Workspace:  "default",
		RepoRelDir: ".",
	}

	pathDir := filepath.Join(tmp, "path")
	Ok(t, os.MkdirAll(pathDir, 0700))
	Ok(t, writeExecutable(filepath.Join(pathDir, "terraform"), "echo '\nTerraform v0.11.10\n'"))
	defer tempSetEnv(t, "PATH", fmt.Sprintf("%s:%s", pathDir, os.Getenv("PATH")))()

	firstVersion, err := version.NewVersion("99.99.99")
	Ok(t, err)
	secondVersion, err := version.NewVersion("98.98.98")
	Ok(t, err)

	downloader := &trackingDownloader{delay: 500 * time.Millisecond}
	distribution := terraform.NewDistributionTerraformWithDownloader(downloader)
	c, err := tfclient.NewClient(logger, distribution, binDir, cacheDir, "", "", "", cmd.DefaultTFVersionFlag, cmd.DefaultTFDownloadURL, true, true, projectCmdOutputHandler)
	Ok(t, err)

	runVersion := func(v *version.Version) <-chan error {
		errCh := make(chan error, 1)
		go func() {
			output, err := c.RunCommandWithVersion(ctx, tmp, []string{"version"}, map[string]string{}, distribution, v, "")
			if err != nil {
				errCh <- err
				return
			}
			expectedOutput := fmt.Sprintf("\nTerraform v%s\n\n", v.String())
			if output != expectedOutput {
				errCh <- fmt.Errorf("unexpected output: %q", output)
				return
			}
			errCh <- nil
		}()
		return errCh
	}

	firstErr := runVersion(firstVersion)
	waitForAtomicAtLeast(t, &downloader.active, 1, time.Second)
	secondErr := runVersion(secondVersion)

	Ok(t, <-firstErr)
	Ok(t, <-secondErr)
	Equals(t, int32(1), downloader.maxConcurrent.Load())
	Equals(t, int32(2), downloader.calls.Load())
}

func TestEnsureVersion_ReusesConcurrentManagedBinaryRepair(t *testing.T) {
	logger := logging.NewNoopLogger(t)
	tmp, binDir, cacheDir := mkSubDirs(t)
	projectCmdOutputHandler := jobmocks.NewMockProjectCommandOutputHandler()

	pathDir := filepath.Join(tmp, "path")
	Ok(t, os.MkdirAll(pathDir, 0700))
	Ok(t, writeExecutable(filepath.Join(pathDir, "terraform"), "echo '\nTerraform v0.11.10\n'"))
	defer tempSetEnv(t, "PATH", fmt.Sprintf("%s:%s", pathDir, os.Getenv("PATH")))()

	v, err := version.NewVersion("99.99.99")
	Ok(t, err)

	startedFile := filepath.Join(tmp, "broken-validation-started")
	releaseFile := filepath.Join(tmp, "release-broken-validation")
	Ok(t, writeExecutable(
		filepath.Join(binDir, "terraform99.99.99"),
		fmt.Sprintf("printf x >> %q\nwhile [ ! -f %q ]; do sleep 0.01; done\nexit 1", startedFile, releaseFile),
	))

	downloader := &trackingDownloader{}
	distribution := terraform.NewDistributionTerraformWithDownloader(downloader)
	c, err := tfclient.NewClient(logger, distribution, binDir, cacheDir, "", "", "", cmd.DefaultTFVersionFlag, cmd.DefaultTFDownloadURL, true, true, projectCmdOutputHandler)
	Ok(t, err)

	errCh := make(chan error, 2)
	for range 2 {
		go func() {
			errCh <- c.EnsureVersion(logger, distribution, v)
		}()
	}

	waitForFileSize(t, startedFile, 2, time.Second)
	Ok(t, os.WriteFile(releaseFile, []byte("release"), 0600))

	for range 2 {
		Ok(t, <-errCh)
	}
	Equals(t, int32(1), downloader.calls.Load())
}

// Test that EnsureVersion downloads terraform.
func TestEnsureVersion_downloaded(t *testing.T) {
	logger := logging.NewNoopLogger(t)
	RegisterMockTestingT(t)
	_, binDir, cacheDir := mkSubDirs(t)
	projectCmdOutputHandler := jobmocks.NewMockProjectCommandOutputHandler()

	mockDownloader := mocks.NewMockDownloader()
	distribution := terraform.NewDistributionTerraformWithDownloader(mockDownloader)

	downloadsAllowed := true
	c, err := tfclient.NewTestClient(logger, distribution, binDir, cacheDir, "", "", "0.11.10", cmd.DefaultTFVersionFlag, cmd.DefaultTFDownloadURL, downloadsAllowed, true, projectCmdOutputHandler)
	Ok(t, err)

	Equals(t, "0.11.10", c.DefaultVersion().String())

	v, err := version.NewVersion("99.99.99")
	Ok(t, err)

	When(mockDownloader.Install(context.Background(), binDir, cmd.DefaultTFDownloadURL, v)).Then(func(params []Param) ReturnValues {
		binPath := filepath.Join(params[1].(string), "terraform99.99.99")
		err := os.WriteFile(binPath, []byte("#!/bin/sh\necho '\nTerraform v99.99.99\n'"), 0700) // #nosec G306
		return []ReturnValue{binPath, err}
	})

	err = c.EnsureVersion(logger, distribution, v)

	Ok(t, err)

	mockDownloader.VerifyWasCalledEventually(Once(), 2*time.Second).Install(context.Background(), binDir, cmd.DefaultTFDownloadURL, v)
}

// Test that EnsureVersion fails if the thing it downloads fails to run.
func TestEnsureVersion_ErrsWhenDownloadedBinaryCannotExecute(t *testing.T) {
	logger := logging.NewNoopLogger(t)
	RegisterMockTestingT(t)
	_, binDir, cacheDir := mkSubDirs(t)
	projectCmdOutputHandler := jobmocks.NewMockProjectCommandOutputHandler()

	mockDownloader := mocks.NewMockDownloader()
	distribution := terraform.NewDistributionTerraformWithDownloader(mockDownloader)

	downloadsAllowed := true
	c, err := tfclient.NewTestClient(logger, distribution, binDir, cacheDir, "", "", "0.11.10", cmd.DefaultTFVersionFlag, cmd.DefaultTFDownloadURL, downloadsAllowed, true, projectCmdOutputHandler)
	Ok(t, err)

	Equals(t, "0.11.10", c.DefaultVersion().String())

	v, err := version.NewVersion("99.99.99")
	Ok(t, err)

	When(mockDownloader.Install(context.Background(), binDir, cmd.DefaultTFDownloadURL, v)).Then(func(params []Param) ReturnValues {
		binPath := filepath.Join(params[1].(string), "terraform99.99.99")
		err := writeExecutable(binPath, "exit 1")
		return []ReturnValue{binPath, err}
	})

	err = c.EnsureVersion(logger, distribution, v)

	ErrContains(t, "failed to execute", err)

	mockDownloader.VerifyWasCalledEventually(Twice(), 2*time.Second).Install(context.Background(), binDir, cmd.DefaultTFDownloadURL, v)
}

// Test that EnsureVersion fixes a broken binary.
func TestEnsureVersion_RedownloadsBrokenManagedBinary(t *testing.T) {
	logger := logging.NewNoopLogger(t)
	RegisterMockTestingT(t)
	tmp, binDir, cacheDir := mkSubDirs(t)
	projectCmdOutputHandler := jobmocks.NewMockProjectCommandOutputHandler()
	ctx := command.ProjectContext{
		Log:        logging.NewNoopLogger(t),
		Workspace:  "default",
		RepoRelDir: ".",
	}

	mockDownloader := mocks.NewMockDownloader()
	distribution := terraform.NewDistributionTerraformWithDownloader(mockDownloader)

	downloadsAllowed := true
	c, err := tfclient.NewTestClient(logger, distribution, binDir, cacheDir, "", "", "0.11.10", cmd.DefaultTFVersionFlag, cmd.DefaultTFDownloadURL, downloadsAllowed, true, projectCmdOutputHandler)
	Ok(t, err)

	Equals(t, "0.11.10", c.DefaultVersion().String())

	v, err := version.NewVersion("99.99.99")
	Ok(t, err)

	binPath := filepath.Join(binDir, "terraform99.99.99")
	Ok(t, writeExecutable(binPath, "exit 1"))

	When(mockDownloader.Install(context.Background(), binDir, cmd.DefaultTFDownloadURL, v)).Then(func(params []Param) ReturnValues {
		binPath := filepath.Join(params[1].(string), "terraform99.99.99")
		err := writeExecutable(binPath, "echo '\nTerraform v99.99.99\n'")
		return []ReturnValue{binPath, err}
	})

	err = c.EnsureVersion(logger, distribution, v)

	Ok(t, err)

	output, err := c.RunCommandWithVersion(ctx, tmp, []string{"terraform", "init"}, map[string]string{}, distribution, v, "")
	Assert(t, err == nil, "err: %s: %s", err, output)
	Equals(t, "\nTerraform v99.99.99\n\n", output)

	mockDownloader.VerifyWasCalledEventually(Once(), 2*time.Second).Install(context.Background(), binDir, cmd.DefaultTFDownloadURL, v)
}

func TestEnsureVersion_RedownloadsWithoutRemovingPathBinary(t *testing.T) {
	logger := logging.NewNoopLogger(t)
	RegisterMockTestingT(t)
	tmp, binDir, cacheDir := mkSubDirs(t)
	projectCmdOutputHandler := jobmocks.NewMockProjectCommandOutputHandler()
	ctx := command.ProjectContext{
		Log:        logging.NewNoopLogger(t),
		Workspace:  "default",
		RepoRelDir: ".",
	}

	pathDir := filepath.Join(tmp, "path")
	Ok(t, os.MkdirAll(pathDir, 0700))
	pathBinary := filepath.Join(pathDir, "terraform99.99.99")
	Ok(t, writeExecutable(pathBinary, "exit 1"))
	defer tempSetEnv(t, "PATH", fmt.Sprintf("%s:%s", pathDir, os.Getenv("PATH")))()

	mockDownloader := mocks.NewMockDownloader()
	distribution := terraform.NewDistributionTerraformWithDownloader(mockDownloader)

	downloadsAllowed := true
	c, err := tfclient.NewTestClient(logger, distribution, binDir, cacheDir, "", "", "0.11.10", cmd.DefaultTFVersionFlag, cmd.DefaultTFDownloadURL, downloadsAllowed, true, projectCmdOutputHandler)
	Ok(t, err)

	Equals(t, "0.11.10", c.DefaultVersion().String())

	v, err := version.NewVersion("99.99.99")
	Ok(t, err)

	When(mockDownloader.Install(context.Background(), binDir, cmd.DefaultTFDownloadURL, v)).Then(func(params []Param) ReturnValues {
		binPath := filepath.Join(params[1].(string), "terraform99.99.99")
		err := writeExecutable(binPath, "echo '\nTerraform v99.99.99\n'")
		return []ReturnValue{binPath, err}
	})

	err = c.EnsureVersion(logger, distribution, v)
	Ok(t, err)

	_, err = os.Stat(pathBinary)
	Ok(t, err)

	output, err := c.RunCommandWithVersion(ctx, tmp, []string{"terraform", "init"}, map[string]string{}, distribution, v, "")
	Assert(t, err == nil, "err: %s: %s", err, output)
	Equals(t, "\nTerraform v99.99.99\n\n", output)

	mockDownloader.VerifyWasCalledEventually(Once(), 2*time.Second).Install(context.Background(), binDir, cmd.DefaultTFDownloadURL, v)
}

// Test that EnsureVersion downloads terraform from a custom URL.
func TestEnsureVersion_downloaded_customURL(t *testing.T) {
	logger := logging.NewNoopLogger(t)
	RegisterMockTestingT(t)
	_, binDir, cacheDir := mkSubDirs(t)
	projectCmdOutputHandler := jobmocks.NewMockProjectCommandOutputHandler()

	mockDownloader := mocks.NewMockDownloader()
	distribution := terraform.NewDistributionTerraformWithDownloader(mockDownloader)
	downloadsAllowed := true
	customURL := "http://releases.example.com"

	c, err := tfclient.NewTestClient(logger, distribution, binDir, cacheDir, "", "", "0.11.10", cmd.DefaultTFVersionFlag, customURL, downloadsAllowed, true, projectCmdOutputHandler)
	Ok(t, err)

	Equals(t, "0.11.10", c.DefaultVersion().String())

	v, err := version.NewVersion("99.99.99")
	Ok(t, err)

	When(mockDownloader.Install(context.Background(), binDir, customURL, v)).Then(func(params []Param) ReturnValues {
		binPath := filepath.Join(params[1].(string), "terraform99.99.99")
		err := os.WriteFile(binPath, []byte("#!/bin/sh\necho '\nTerraform v99.99.99\n'"), 0700) // #nosec G306
		return []ReturnValue{binPath, err}
	})

	err = c.EnsureVersion(logger, distribution, v)

	Ok(t, err)

	mockDownloader.VerifyWasCalledEventually(Once(), 2*time.Second).Install(context.Background(), binDir, customURL, v)
}

// Test that EnsureVersion throws an error when downloads are disabled
func TestEnsureVersion_downloaded_downloadingDisabled(t *testing.T) {
	logger := logging.NewNoopLogger(t)
	RegisterMockTestingT(t)
	_, binDir, cacheDir := mkSubDirs(t)
	projectCmdOutputHandler := jobmocks.NewMockProjectCommandOutputHandler()

	mockDownloader := mocks.NewMockDownloader()
	distribution := terraform.NewDistributionTerraformWithDownloader(mockDownloader)

	downloadsAllowed := false
	c, err := tfclient.NewTestClient(logger, distribution, binDir, cacheDir, "", "", "0.11.10", cmd.DefaultTFVersionFlag, cmd.DefaultTFDownloadURL, downloadsAllowed, true, projectCmdOutputHandler)
	Ok(t, err)

	Equals(t, "0.11.10", c.DefaultVersion().String())

	v, err := version.NewVersion("99.99.99")
	Ok(t, err)

	err = c.EnsureVersion(logger, distribution, v)
	ErrContains(t, "could not find terraform version", err)
	ErrContains(t, "downloads are disabled", err)
	mockDownloader.VerifyWasCalled(Never())
}

// tempSetEnv sets env var key to value. It returns a function that when called
// will reset the env var to its original value.
func tempSetEnv(t *testing.T, key string, value string) func() {
	orig := os.Getenv(key)
	Ok(t, os.Setenv(key, value))
	return func() { os.Setenv(key, orig) }
}

func writeExecutable(path string, body string) error {
	return os.WriteFile(path, []byte("#!/bin/sh\n"+body), 0700) // #nosec G306
}

func waitForFile(t *testing.T, path string, timeout time.Duration) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if _, err := os.Stat(path); err == nil {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("timed out waiting for %s", path)
}

func waitForFileSize(t *testing.T, path string, minSize int64, timeout time.Duration) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if stat, err := os.Stat(path); err == nil && stat.Size() >= minSize {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("timed out waiting for %s to reach size %d", path, minSize)
}

func waitForAtomicAtLeast(t *testing.T, counter *atomic.Int32, min int32, timeout time.Duration) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if counter.Load() >= min {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("timed out waiting for counter to reach %d", min)
}

type trackingDownloader struct {
	active        atomic.Int32
	calls         atomic.Int32
	delay         time.Duration
	maxConcurrent atomic.Int32
}

func (d *trackingDownloader) Install(_ context.Context, dir string, _ string, v *version.Version) (string, error) {
	active := d.active.Add(1)
	defer d.active.Add(-1)
	d.calls.Add(1)
	d.recordMaxConcurrent(active)
	time.Sleep(d.delay)

	binPath := filepath.Join(dir, "terraform"+v.String())
	return binPath, writeExecutable(binPath, fmt.Sprintf("echo '\nTerraform v%s\n'", v.String()))
}

func (d *trackingDownloader) recordMaxConcurrent(active int32) {
	for {
		current := d.maxConcurrent.Load()
		if active <= current || d.maxConcurrent.CompareAndSwap(current, active) {
			return
		}
	}
}

// returns parent, bindir, cachedir
func mkSubDirs(t *testing.T) (string, string, string) {
	tmp := t.TempDir()
	binDir := filepath.Join(tmp, "bin")
	err := os.MkdirAll(binDir, 0700)
	Ok(t, err)

	cachedir := filepath.Join(tmp, "plugin-cache")
	err = os.MkdirAll(cachedir, 0700)
	Ok(t, err)

	return tmp, binDir, cachedir
}

// If TF downloads are disabled, test that terraform version is used when specified in terraform configuration only if an exact version
func TestDefaultProjectCommandBuilder_TerraformVersion(t *testing.T) {
	// For the following tests:
	// If terraform configuration is used, result should be `0.12.8`.
	// If project configuration is used, result should be `0.12.6`.
	// If an inexact version is used, the result should be `nil`
	// If default is to be used, result should be `nil`.

	baseVersionConfig := `
terraform {
  required_version = "%s"
}
`
	// Depending on when the tests are run, the > and >= matching versions will have to be increased.
	// It's probably not worth testing the terraform-switcher version here so we only test <, <=, and ~>.
	// One way to test this in the future is to mock tfswitcher.GetTFList() to return the highest
	// version of 1.3.5.
	expectedVersions := map[string]string{
		"= 0.12.8":  "0.12.8",
		"< 0.12.8":  "0.12.7",
		"<= 0.12.8": "0.12.8",
		"~> 0.12.8": "0.12.31",

		"= 1.0.0":  "1.0.0",
		"< 1.0.0":  "0.15.5",
		"<= 1.0.0": "1.0.0",
		"~> 1.0.0": "1.0.11",

		"= 1.0":  "1.0.0",
		"< 1.0":  "0.15.5",
		"<= 1.0": "1.0.0",
		// cannot use ~> 1.3 or ~> 1.0 since that is a moving target since it will always
		// resolve to the latest terraform 1.x
		"~> 1.3.0": "1.3.10",
	}

	type testCase struct {
		DirStructure map[string]any
		Exp          map[string]string
		IsExact      bool
	}

	testCases := make(map[string]testCase)
	for version, expected := range expectedVersions {
		testCases[fmt.Sprintf("version using \"%s\"", version)] = testCase{
			DirStructure: map[string]any{
				"project1": map[string]any{
					"main.tf": fmt.Sprintf(baseVersionConfig, version),
				},
			},
			Exp: map[string]string{
				"project1": expected,
			},
			IsExact: version[0] == "="[0],
		}
	}

	testCases["no version specified"] = testCase{
		DirStructure: map[string]any{
			"project1": map[string]any{
				"main.tf": nil,
			},
		},
		Exp: map[string]string{
			"project1": "",
		},
		IsExact: true,
	}

	testCases["projects with different terraform versions"] = testCase{
		DirStructure: map[string]any{
			"project1": map[string]any{
				"main.tf": fmt.Sprintf(baseVersionConfig, "= 0.12.8"),
			},
			"project2": map[string]any{
				"main.tf": strings.ReplaceAll(fmt.Sprintf(baseVersionConfig, "= 0.12.8"), "0.12.8", "0.12.9"),
			},
		},
		Exp: map[string]string{
			"project1": "0.12.8",
			"project2": "0.12.9",
		},
		IsExact: true,
	}

	runDetectVersionTestCase := func(t *testing.T, name string, testCase testCase, downloadsAllowed bool) bool {
		return t.Run(name, func(t *testing.T) {
			RegisterMockTestingT(t)

			logger := logging.NewNoopLogger(t)
			RegisterMockTestingT(t)
			_, binDir, cacheDir := mkSubDirs(t)
			projectCmdOutputHandler := jobmocks.NewMockProjectCommandOutputHandler()
			mockDownloader := mocks.NewMockDownloader()
			distribution := terraform.NewDistributionTerraformWithDownloader(mockDownloader)

			c, err := tfclient.NewTestClient(
				logger,
				distribution,
				binDir,
				cacheDir,
				"",
				"",
				"",
				cmd.DefaultTFVersionFlag,
				cmd.DefaultTFDownloadURL,
				downloadsAllowed,
				true,
				projectCmdOutputHandler)
			Ok(t, err)

			tmpDir := DirStructure(t, testCase.DirStructure)

			for project, expectedVersion := range testCase.Exp {
				detectedVersion := c.DetectVersion(logger, nil, filepath.Join(tmpDir, project))

				expectNil := expectedVersion == "" || (!testCase.IsExact && !downloadsAllowed)
				if expectNil {
					Assert(t, detectedVersion == nil, "TerraformVersion is supposed to be nil.")
				} else {
					Assert(t, detectedVersion != nil, "TerraformVersion is nil.")
					Ok(t, err)
					Equals(t, expectedVersion, detectedVersion.String())
				}
			}

		})
	}

	for name, testCase := range testCases {
		runDetectVersionTestCase(t, name+": Downloads Allowed", testCase, true)
		runDetectVersionTestCase(t, name+": Downloads Disabled", testCase, false)
	}
}

type constraintResolvingDistribution struct {
	binName         string
	resolvedVersion string
	constraints     []string
}

func (d *constraintResolvingDistribution) BinName() string {
	return d.binName
}

func (*constraintResolvingDistribution) Downloader() terraform.Downloader {
	return nil
}

func (d *constraintResolvingDistribution) ResolveConstraint(_ context.Context, constraintStr string) (*version.Version, error) {
	d.constraints = append(d.constraints, constraintStr)
	return version.NewVersion(d.resolvedVersion)
}

func TestDetectVersion_UsesDistributionOverride(t *testing.T) {
	logger := logging.NewNoopLogger(t)
	tmp, binDir, cacheDir := mkSubDirs(t)
	projectCmdOutputHandler := jobmocks.NewMockProjectCommandOutputHandler()
	Ok(t, writeExecutable(filepath.Join(tmp, "terraform"), "echo '\nTerraform v0.11.10\n'"))
	defer tempSetEnv(t, "PATH", fmt.Sprintf("%s:%s", tmp, os.Getenv("PATH")))()
	defaultDistribution := &constraintResolvingDistribution{
		binName:         "terraform",
		resolvedVersion: "1.15.7",
	}
	opentofuDistribution := &constraintResolvingDistribution{
		binName:         "tofu",
		resolvedVersion: "1.12.3",
	}
	c, err := tfclient.NewTestClient(
		logger,
		defaultDistribution,
		binDir,
		cacheDir,
		"",
		"",
		"",
		cmd.DefaultTFVersionFlag,
		cmd.DefaultTFDownloadURL,
		true,
		true,
		projectCmdOutputHandler,
	)
	Ok(t, err)

	tmpDir := DirStructure(t, map[string]any{
		"project1": map[string]any{
			"main.tf": `terraform {
  required_version = ">= 1.5"
}
`,
		},
	})

	detectedDefault := c.DetectVersion(logger, nil, filepath.Join(tmpDir, "project1"))
	Assert(t, detectedDefault != nil, "TerraformVersion is nil.")
	Equals(t, "1.15.7", detectedDefault.String())
	Equals(t, []string{">= 1.5"}, defaultDistribution.constraints)

	detectedOpenTofu := c.DetectVersion(logger, opentofuDistribution, filepath.Join(tmpDir, "project1"))
	Assert(t, detectedOpenTofu != nil, "TerraformVersion is nil.")
	Equals(t, "1.12.3", detectedOpenTofu.String())
	Equals(t, []string{">= 1.5"}, opentofuDistribution.constraints)
}

func TestExtractExactRegex(t *testing.T) {
	logger := logging.NewNoopLogger(t)
	RegisterMockTestingT(t)
	_, binDir, cacheDir := mkSubDirs(t)
	projectCmdOutputHandler := jobmocks.NewMockProjectCommandOutputHandler()
	mockDownloader := mocks.NewMockDownloader()
	distribution := terraform.NewDistributionTerraformWithDownloader(mockDownloader)

	c, err := tfclient.NewTestClient(logger, distribution, binDir, cacheDir, "", "", "0.11.10", cmd.DefaultTFVersionFlag, cmd.DefaultTFDownloadURL, true, true, projectCmdOutputHandler)
	Ok(t, err)

	tests := []struct {
		version string
		want    []string
	}{
		{"= 1.2.3", []string{"1.2.3"}},
		{"=1.2.3", []string{"1.2.3"}},
		{"1.2.3", []string{"1.2.3"}},
		{"v1.2.3", nil},
		{">= 1.2.3", nil},
		{">=1.2.3", nil},
		{"<= 1.2.3", nil},
		{"<=1.2.3", nil},
		{"~> 1.2.3", nil},
	}

	for _, tt := range tests {
		t.Run(tt.version, func(t *testing.T) {
			if got := c.ExtractExactRegex(logger, tt.version); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ExtractExactRegex() = %v, want %v", got, tt.want)
			}
		})
	}
}
