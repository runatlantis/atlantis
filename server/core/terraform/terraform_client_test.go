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
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	version "github.com/hashicorp/go-version"
	"github.com/petergtz/pegomock"
	. "github.com/petergtz/pegomock"
	"github.com/runatlantis/atlantis/cmd"
	"github.com/runatlantis/atlantis/server/core/terraform"
	"github.com/runatlantis/atlantis/server/core/terraform/mocks"
	"github.com/runatlantis/atlantis/server/events/models"
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
is 0.11.13. You can update by downloading from www.terraform.io/downloads.html
`
	tmp, binDir, cacheDir, cleanup := mkSubDirs(t)
	projectCmdOutputHandler := jobmocks.NewMockProjectCommandOutputHandler()
	ctx := models.ProjectCommandContext{
		Log:        logging.NewNoopLogger(t),
		Workspace:  "default",
		RepoRelDir: ".",
	}
	defer cleanup()

	logger := logging.NewNoopLogger(t)

	// We're testing this by adding our own "fake" terraform binary to path that
	// outputs what would normally come from terraform version.
	err := os.WriteFile(filepath.Join(tmp, "terraform"), []byte(fmt.Sprintf("#!/bin/sh\necho '%s'", fakeBinOut)), 0700) // #nosec G306
	Ok(t, err)
	defer tempSetEnv(t, "PATH", fmt.Sprintf("%s:%s", tmp, os.Getenv("PATH")))()

	c, err := terraform.NewClient(logger, binDir, cacheDir, "", "", "", cmd.DefaultTFVersionFlag, cmd.DefaultTFDownloadURL, nil, true, projectCmdOutputHandler)
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
is 0.11.13. You can update by downloading from www.terraform.io/downloads.html
`
	logger := logging.NewNoopLogger(t)
	tmp, binDir, cacheDir, cleanup := mkSubDirs(t)
	projectCmdOutputHandler := jobmocks.NewMockProjectCommandOutputHandler()
	ctx := models.ProjectCommandContext{
		Log:        logging.NewNoopLogger(t),
		Workspace:  "default",
		RepoRelDir: ".",
	}
	defer cleanup()

	// We're testing this by adding our own "fake" terraform binary to path that
	// outputs what would normally come from terraform version.
	err := os.WriteFile(filepath.Join(tmp, "terraform"), []byte(fmt.Sprintf("#!/bin/sh\necho '%s'", fakeBinOut)), 0700) // #nosec G306
	Ok(t, err)
	defer tempSetEnv(t, "PATH", fmt.Sprintf("%s:%s", tmp, os.Getenv("PATH")))()

	c, err := terraform.NewClient(logger, binDir, cacheDir, "", "", "0.11.10", cmd.DefaultTFVersionFlag, cmd.DefaultTFDownloadURL, nil, true, projectCmdOutputHandler)
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
	tmp, binDir, cacheDir, cleanup := mkSubDirs(t)
	projectCmdOutputHandler := jobmocks.NewMockProjectCommandOutputHandler()
	defer cleanup()

	// Set PATH to only include our empty directory.
	defer tempSetEnv(t, "PATH", tmp)()

	_, err := terraform.NewClient(logger, binDir, cacheDir, "", "", "", cmd.DefaultTFVersionFlag, cmd.DefaultTFDownloadURL, nil, true, projectCmdOutputHandler)
	ErrEquals(t, "terraform not found in $PATH. Set --default-tf-version or download terraform from https://www.terraform.io/downloads.html", err)
}

// Test that if the default-tf flag is set and that binary is in our PATH
// that we use it.
func TestNewClient_DefaultTFFlagInPath(t *testing.T) {
	fakeBinOut := "Terraform v0.11.10\n"
	logger := logging.NewNoopLogger(t)
	tmp, binDir, cacheDir, cleanup := mkSubDirs(t)
	projectCmdOutputHandler := jobmocks.NewMockProjectCommandOutputHandler()
	ctx := models.ProjectCommandContext{
		Log:        logging.NewNoopLogger(t),
		Workspace:  "default",
		RepoRelDir: ".",
	}
	defer cleanup()

	// We're testing this by adding our own "fake" terraform binary to path that
	// outputs what would normally come from terraform version.
	err := os.WriteFile(filepath.Join(tmp, "terraform0.11.10"), []byte(fmt.Sprintf("#!/bin/sh\necho '%s'", fakeBinOut)), 0700) // #nosec G306
	Ok(t, err)
	defer tempSetEnv(t, "PATH", fmt.Sprintf("%s:%s", tmp, os.Getenv("PATH")))()

	c, err := terraform.NewClient(logger, binDir, cacheDir, "", "", "0.11.10", cmd.DefaultTFVersionFlag, cmd.DefaultTFDownloadURL, nil, true, projectCmdOutputHandler)
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
	tmp, binDir, cacheDir, cleanup := mkSubDirs(t)
	projectCmdOutputHandler := jobmocks.NewMockProjectCommandOutputHandler()
	ctx := models.ProjectCommandContext{
		Log:        logging.NewNoopLogger(t),
		Workspace:  "default",
		RepoRelDir: ".",
	}
	defer cleanup()

	// Add our fake binary to {datadir}/bin/terraform{version}.
	err := os.WriteFile(filepath.Join(binDir, "terraform0.11.10"), []byte(fmt.Sprintf("#!/bin/sh\necho '%s'", fakeBinOut)), 0700) // #nosec G306
	Ok(t, err)
	defer tempSetEnv(t, "PATH", fmt.Sprintf("%s:%s", tmp, os.Getenv("PATH")))()

	c, err := terraform.NewClient(logging.NewNoopLogger(t), binDir, cacheDir, "", "", "0.11.10", cmd.DefaultTFVersionFlag, cmd.DefaultTFDownloadURL, nil, true, projectCmdOutputHandler)
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
	tmp, binDir, cacheDir, cleanup := mkSubDirs(t)
	projectCmdOutputHandler := jobmocks.NewMockProjectCommandOutputHandler()
	ctx := models.ProjectCommandContext{
		Log:        logging.NewNoopLogger(t),
		Workspace:  "default",
		RepoRelDir: ".",
	}
	defer cleanup()

	// Set PATH to empty so there's no TF available.
	orig := os.Getenv("PATH")
	defer tempSetEnv(t, "PATH", "")()

	mockDownloader := mocks.NewMockDownloader()
	When(mockDownloader.GetFile(AnyString(), AnyString())).Then(func(params []pegomock.Param) pegomock.ReturnValues {
		err := os.WriteFile(params[0].(string), []byte("#!/bin/sh\necho '\nTerraform v0.11.10\n'"), 0700) // #nosec G306
		return []pegomock.ReturnValue{err}
	})
	c, err := terraform.NewClient(logger, binDir, cacheDir, "", "", "0.11.10", cmd.DefaultTFVersionFlag, "https://my-mirror.releases.mycompany.com", mockDownloader, true, projectCmdOutputHandler)
	Ok(t, err)

	Ok(t, err)
	Equals(t, "0.11.10", c.DefaultVersion().String())
	baseURL := "https://my-mirror.releases.mycompany.com/terraform/0.11.10"
	expURL := fmt.Sprintf("%s/terraform_0.11.10_%s_%s.zip?checksum=file:%s/terraform_0.11.10_SHA256SUMS",
		baseURL,
		runtime.GOOS,
		runtime.GOARCH,
		baseURL)
	mockDownloader.VerifyWasCalledEventually(Once(), 2*time.Second).GetFile(filepath.Join(tmp, "bin", "terraform0.11.10"), expURL)

	// Reset PATH so that it has sh.
	Ok(t, os.Setenv("PATH", orig))

	output, err := c.RunCommandWithVersion(ctx, tmp, []string{"terraform", "init"}, map[string]string{}, nil, "")
	Ok(t, err)
	Equals(t, "\nTerraform v0.11.10\n\n", output)
}

// Test that we get an error if the terraform version flag is malformed.
func TestNewClient_BadVersion(t *testing.T) {
	logger := logging.NewNoopLogger(t)
	_, binDir, cacheDir, cleanup := mkSubDirs(t)
	projectCmdOutputHandler := jobmocks.NewMockProjectCommandOutputHandler()
	defer cleanup()
	_, err := terraform.NewClient(logger, binDir, cacheDir, "", "", "malformed", cmd.DefaultTFVersionFlag, cmd.DefaultTFDownloadURL, nil, true, projectCmdOutputHandler)
	ErrEquals(t, "Malformed version: malformed", err)
}

// Test that if we run a command with a version we don't have, we download it.
func TestRunCommandWithVersion_DLsTF(t *testing.T) {
	logger := logging.NewNoopLogger(t)
	RegisterMockTestingT(t)
	tmp, binDir, cacheDir, cleanup := mkSubDirs(t)
	projectCmdOutputHandler := jobmocks.NewMockProjectCommandOutputHandler()
	ctx := models.ProjectCommandContext{
		Log:        logging.NewNoopLogger(t),
		Workspace:  "default",
		RepoRelDir: ".",
	}
	defer cleanup()

	mockDownloader := mocks.NewMockDownloader()
	// Set up our mock downloader to write a fake tf binary when it's called.
	baseURL := fmt.Sprintf("%s/terraform/99.99.99", cmd.DefaultTFDownloadURL)
	expURL := fmt.Sprintf("%s/terraform_99.99.99_%s_%s.zip?checksum=file:%s/terraform_99.99.99_SHA256SUMS",
		baseURL,
		runtime.GOOS,
		runtime.GOARCH,
		baseURL)
	When(mockDownloader.GetFile(filepath.Join(tmp, "bin", "terraform99.99.99"), expURL)).Then(func(params []pegomock.Param) pegomock.ReturnValues {
		err := os.WriteFile(params[0].(string), []byte("#!/bin/sh\necho '\nTerraform v99.99.99\n'"), 0700) // #nosec G306
		return []pegomock.ReturnValue{err}
	})

	c, err := terraform.NewClient(logger, binDir, cacheDir, "", "", "0.11.10", cmd.DefaultTFVersionFlag, cmd.DefaultTFDownloadURL, mockDownloader, true, projectCmdOutputHandler)
	Ok(t, err)
	Equals(t, "0.11.10", c.DefaultVersion().String())

	v, err := version.NewVersion("99.99.99")
	Ok(t, err)

	output, err := c.RunCommandWithVersion(ctx, tmp, []string{"terraform", "init"}, map[string]string{}, v, "")

	Assert(t, err == nil, "err: %s: %s", err, output)
	Equals(t, "\nTerraform v99.99.99\n\n", output)
}

// Test the EnsureVersion downloads terraform.
func TestEnsureVersion_downloaded(t *testing.T) {
	logger := logging.NewNoopLogger(t)
	RegisterMockTestingT(t)
	tmp, binDir, cacheDir, cleanup := mkSubDirs(t)
	projectCmdOutputHandler := jobmocks.NewMockProjectCommandOutputHandler()
	defer cleanup()

	mockDownloader := mocks.NewMockDownloader()

	c, err := terraform.NewTestClient(logger, binDir, cacheDir, "", "", "0.11.10", cmd.DefaultTFVersionFlag, cmd.DefaultTFDownloadURL, mockDownloader, true, projectCmdOutputHandler)
	Ok(t, err)

	Equals(t, "0.11.10", c.DefaultVersion().String())

	v, err := version.NewVersion("99.99.99")
	Ok(t, err)

	err = c.EnsureVersion(logger, v)

	Ok(t, err)

	baseURL := fmt.Sprintf("%s/terraform/99.99.99", cmd.DefaultTFDownloadURL)
	expURL := fmt.Sprintf("%s/terraform_99.99.99_%s_%s.zip?checksum=file:%s/terraform_99.99.99_SHA256SUMS",
		baseURL,
		runtime.GOOS,
		runtime.GOARCH,
		baseURL)
	mockDownloader.VerifyWasCalledEventually(Once(), 2*time.Second).GetFile(filepath.Join(tmp, "bin", "terraform99.99.99"), expURL)
}

// tempSetEnv sets env var key to value. It returns a function that when called
// will reset the env var to its original value.
func tempSetEnv(t *testing.T, key string, value string) func() {
	orig := os.Getenv(key)
	Ok(t, os.Setenv(key, value))
	return func() { os.Setenv(key, orig) }
}

// returns parent, bindir, cachedir, cleanup func
func mkSubDirs(t *testing.T) (string, string, string, func()) {
	tmp, cleanup := TempDir(t)
	binDir := filepath.Join(tmp, "bin")
	err := os.MkdirAll(binDir, 0700)
	Ok(t, err)

	cachedir := filepath.Join(tmp, "plugin-cache")
	err = os.MkdirAll(cachedir, 0700)
	Ok(t, err)

	return tmp, binDir, cachedir, cleanup
}
