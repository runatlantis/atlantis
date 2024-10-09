// Copyright 2017 HootSuite Media Inc.
//
// Licensed under the Apache License, Version 2.0 (the License);
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//    http://www.apache.org/licenses/LICENSE-2.0
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an AS IS BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
// Modified hereafter by contributors to runatlantis/atlantis.

package terraform_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	version "github.com/hashicorp/go-version"
	. "github.com/petergtz/pegomock/v4"
	"github.com/runatlantis/atlantis/cmd"
	"github.com/runatlantis/atlantis/server/core/terraform"
	"github.com/runatlantis/atlantis/server/core/terraform/mocks"
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

	terraform.MustConstraint("invalid constraint")
}

func TestMustConstraint(t *testing.T) {
	t.Log("MustConstraint should return the constrain")
	c := terraform.MustConstraint(">0.1")
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
	err := os.WriteFile(filepath.Join(tmp, "terraform"), []byte(fmt.Sprintf("#!/bin/sh\necho '%s'", fakeBinOut)), 0700) // #nosec G306
	Ok(t, err)
	defer tempSetEnv(t, "PATH", fmt.Sprintf("%s:%s", tmp, os.Getenv("PATH")))()

	mockDownloader := mocks.NewMockDownloader()
	distibution := terraform.NewDistributionTerraformWithDownloader(mockDownloader)

	c, err := terraform.NewClient(logger, distibution, binDir, cacheDir, "", "", "", cmd.DefaultTFVersionFlag, cmd.DefaultTFDownloadURL, true, true, projectCmdOutputHandler)
	Ok(t, err)

	Ok(t, err)
	Equals(t, "0.11.10", c.DefaultVersion().String())

	output, err := c.RunCommandWithVersion(ctx, tmp, []string{"terraform", "init"}, map[string]string{"test": "123"}, nil, "")
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
	err := os.WriteFile(filepath.Join(tmp, "terraform"), []byte(fmt.Sprintf("#!/bin/sh\necho '%s'", fakeBinOut)), 0700) // #nosec G306
	Ok(t, err)
	defer tempSetEnv(t, "PATH", fmt.Sprintf("%s:%s", tmp, os.Getenv("PATH")))()

	mockDownloader := mocks.NewMockDownloader()
	distribution := terraform.NewDistributionTerraformWithDownloader(mockDownloader)

	c, err := terraform.NewClient(logger, distribution, binDir, cacheDir, "", "", "0.11.10", cmd.DefaultTFVersionFlag, cmd.DefaultTFDownloadURL, true, true, projectCmdOutputHandler)
	Ok(t, err)

	Ok(t, err)
	Equals(t, "0.11.10", c.DefaultVersion().String())

	output, err := c.RunCommandWithVersion(ctx, tmp, []string{"terraform", "init"}, map[string]string{}, nil, "")
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

	_, err := terraform.NewClient(logger, distribution, binDir, cacheDir, "", "", "", cmd.DefaultTFVersionFlag, cmd.DefaultTFDownloadURL, true, true, projectCmdOutputHandler)
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
	err := os.WriteFile(filepath.Join(tmp, "terraform0.11.10"), []byte(fmt.Sprintf("#!/bin/sh\necho '%s'", fakeBinOut)), 0700) // #nosec G306
	Ok(t, err)
	defer tempSetEnv(t, "PATH", fmt.Sprintf("%s:%s", tmp, os.Getenv("PATH")))()

	mockDownloader := mocks.NewMockDownloader()
	distribution := terraform.NewDistributionTerraformWithDownloader(mockDownloader)

	c, err := terraform.NewClient(logger, distribution, binDir, cacheDir, "", "", "0.11.10", cmd.DefaultTFVersionFlag, cmd.DefaultTFDownloadURL, false, true, projectCmdOutputHandler)
	Ok(t, err)

	Ok(t, err)
	Equals(t, "0.11.10", c.DefaultVersion().String())

	output, err := c.RunCommandWithVersion(ctx, tmp, []string{"terraform", "init"}, map[string]string{}, nil, "")
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
	err := os.WriteFile(filepath.Join(binDir, "terraform0.11.10"), []byte(fmt.Sprintf("#!/bin/sh\necho '%s'", fakeBinOut)), 0700) // #nosec G306
	Ok(t, err)
	defer tempSetEnv(t, "PATH", fmt.Sprintf("%s:%s", tmp, os.Getenv("PATH")))()

	mockDownloader := mocks.NewMockDownloader()
	distribution := terraform.NewDistributionTerraformWithDownloader(mockDownloader)

	c, err := terraform.NewClient(logging.NewNoopLogger(t), distribution, binDir, cacheDir, "", "", "0.11.10", cmd.DefaultTFVersionFlag, cmd.DefaultTFDownloadURL, true, true, projectCmdOutputHandler)
	Ok(t, err)

	Ok(t, err)
	Equals(t, "0.11.10", c.DefaultVersion().String())

	output, err := c.RunCommandWithVersion(ctx, tmp, []string{"terraform", "init"}, map[string]string{}, nil, "")
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
	c, err := terraform.NewClient(logger, distribution, binDir, cacheDir, "", "", "0.11.10", cmd.DefaultTFVersionFlag, cmd.DefaultTFDownloadURL, true, true, projectCmdOutputHandler)
	Ok(t, err)

	Ok(t, err)
	Equals(t, "0.11.10", c.DefaultVersion().String())

	mockDownloader.VerifyWasCalledEventually(Once(), 2*time.Second).Install(context.Background(), binDir, cmd.DefaultTFDownloadURL, version.Must(version.NewVersion("0.11.10")))

	// Reset PATH so that it has sh.
	Ok(t, os.Setenv("PATH", orig))

	output, err := c.RunCommandWithVersion(ctx, tmp, []string{"terraform", "init"}, map[string]string{}, nil, "")
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
	_, err := terraform.NewClient(logger, distribution, binDir, cacheDir, "", "", "malformed", cmd.DefaultTFVersionFlag, cmd.DefaultTFDownloadURL, true, true, projectCmdOutputHandler)
	ErrEquals(t, "Malformed version: malformed", err)
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

	c, err := terraform.NewClient(logger, distribution, binDir, cacheDir, "", "", "0.11.10", cmd.DefaultTFVersionFlag, cmd.DefaultTFDownloadURL, true, true, projectCmdOutputHandler)
	Ok(t, err)
	Equals(t, "0.11.10", c.DefaultVersion().String())

	output, err := c.RunCommandWithVersion(ctx, tmp, []string{"terraform", "init"}, map[string]string{}, v, "")

	Assert(t, err == nil, "err: %s: %s", err, output)
	Equals(t, "\nTerraform v99.99.99\n\n", output)
}

// Test that EnsureVersion downloads terraform.
func TestEnsureVersion_downloaded(t *testing.T) {
	logger := logging.NewNoopLogger(t)
	RegisterMockTestingT(t)
	_, binDir, cacheDir := mkSubDirs(t)
	projectCmdOutputHandler := jobmocks.NewMockProjectCommandOutputHandler()

	mockDownloader := mocks.NewMockDownloader()
	distibution := terraform.NewDistributionTerraformWithDownloader(mockDownloader)

	downloadsAllowed := true
	c, err := terraform.NewTestClient(logger, distibution, binDir, cacheDir, "", "", "0.11.10", cmd.DefaultTFVersionFlag, cmd.DefaultTFDownloadURL, downloadsAllowed, true, projectCmdOutputHandler)
	Ok(t, err)

	Equals(t, "0.11.10", c.DefaultVersion().String())

	v, err := version.NewVersion("99.99.99")
	Ok(t, err)

	When(mockDownloader.Install(context.Background(), binDir, cmd.DefaultTFDownloadURL, v)).Then(func(params []Param) ReturnValues {
		binPath := filepath.Join(params[1].(string), "terraform99.99.99")
		err := os.WriteFile(binPath, []byte("#!/bin/sh\necho '\nTerraform v99.99.99\n'"), 0700) // #nosec G306
		return []ReturnValue{binPath, err}
	})

	err = c.EnsureVersion(logger, v)

	Ok(t, err)

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

	c, err := terraform.NewTestClient(logger, distribution, binDir, cacheDir, "", "", "0.11.10", cmd.DefaultTFVersionFlag, customURL, downloadsAllowed, true, projectCmdOutputHandler)
	Ok(t, err)

	Equals(t, "0.11.10", c.DefaultVersion().String())

	v, err := version.NewVersion("99.99.99")
	Ok(t, err)

	When(mockDownloader.Install(context.Background(), binDir, customURL, v)).Then(func(params []Param) ReturnValues {
		binPath := filepath.Join(params[1].(string), "terraform99.99.99")
		err := os.WriteFile(binPath, []byte("#!/bin/sh\necho '\nTerraform v99.99.99\n'"), 0700) // #nosec G306
		return []ReturnValue{binPath, err}
	})

	err = c.EnsureVersion(logger, v)

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
	c, err := terraform.NewTestClient(logger, distribution, binDir, cacheDir, "", "", "0.11.10", cmd.DefaultTFVersionFlag, cmd.DefaultTFDownloadURL, downloadsAllowed, true, projectCmdOutputHandler)
	Ok(t, err)

	Equals(t, "0.11.10", c.DefaultVersion().String())

	v, err := version.NewVersion("99.99.99")
	Ok(t, err)

	err = c.EnsureVersion(logger, v)
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
		DirStructure map[string]interface{}
		Exp          map[string]string
		IsExact      bool
	}

	testCases := make(map[string]testCase)
	for version, expected := range expectedVersions {
		testCases[fmt.Sprintf("version using \"%s\"", version)] = testCase{
			DirStructure: map[string]interface{}{
				"project1": map[string]interface{}{
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
		DirStructure: map[string]interface{}{
			"project1": map[string]interface{}{
				"main.tf": nil,
			},
		},
		Exp: map[string]string{
			"project1": "",
		},
		IsExact: true,
	}

	testCases["projects with different terraform versions"] = testCase{
		DirStructure: map[string]interface{}{
			"project1": map[string]interface{}{
				"main.tf": fmt.Sprintf(baseVersionConfig, "= 0.12.8"),
			},
			"project2": map[string]interface{}{
				"main.tf": strings.Replace(fmt.Sprintf(baseVersionConfig, "= 0.12.8"), "0.12.8", "0.12.9", -1),
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

			c, err := terraform.NewTestClient(
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
				detectedVersion := c.DetectVersion(logger, filepath.Join(tmpDir, project))

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

func TestExtractExactRegex(t *testing.T) {
	logger := logging.NewNoopLogger(t)
	RegisterMockTestingT(t)
	_, binDir, cacheDir := mkSubDirs(t)
	projectCmdOutputHandler := jobmocks.NewMockProjectCommandOutputHandler()
	mockDownloader := mocks.NewMockDownloader()
	distribution := terraform.NewDistributionTerraformWithDownloader(mockDownloader)

	c, err := terraform.NewTestClient(logger, distribution, binDir, cacheDir, "", "", "0.11.10", cmd.DefaultTFVersionFlag, cmd.DefaultTFDownloadURL, true, true, projectCmdOutputHandler)
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
