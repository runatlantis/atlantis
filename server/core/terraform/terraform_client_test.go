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
	"testing"

	"github.com/runatlantis/atlantis/cmd"
	"github.com/runatlantis/atlantis/server/core/terraform"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/models"
	jobmocks "github.com/runatlantis/atlantis/server/jobs/mocks"
	"github.com/runatlantis/atlantis/server/logging"
	. "github.com/runatlantis/atlantis/testing"
)

// Test that if terraform is not in PATH and we didn't set the default-tf flag
// that we error.
func TestNewClient_NoTF(t *testing.T) {
	tmp, binDir, cacheDir, cleanup := mkSubDirs(t)
	projectCmdOutputHandler := jobmocks.NewMockProjectCommandOutputHandler()
	defer cleanup()

	// Set PATH to only include our empty directory.
	defer tempSetEnv(t, "PATH", tmp)()

	_, err := terraform.NewClient(binDir, cacheDir, "", cmd.DefaultTFVersionFlag, cmd.DefaultTFDownloadURL, nil, true, projectCmdOutputHandler)
	ErrEquals(t, "getting default version: terraform not found in $PATH. Set --default-tf-version or download terraform from https://www.terraform.io/downloads.html", err)
}

// Test that if the default-tf flag is set and that binary is in our PATH
// that we use it.
func TestNewClient_DefaultTFFlagInPath(t *testing.T) {
	fakeBinOut := "Terraform v0.11.10\n"
	tmp, binDir, cacheDir, cleanup := mkSubDirs(t)
	projectCmdOutputHandler := jobmocks.NewMockProjectCommandOutputHandler()
	ctx := context.Background()
	prjCtx := command.ProjectContext{
		RequestCtx: context.TODO(),
		Log:        logging.NewNoopCtxLogger(t),
		Workspace:  "default",
		RepoRelDir: ".",
		BaseRepo:   models.Repo{FullName: "owner/repo"},
	}
	defer cleanup()

	// We're testing this by adding our own "fake" terraform binary to path that
	// outputs what would normally come from terraform version.
	err := os.WriteFile(filepath.Join(tmp, "terraform0.11.10"), []byte(fmt.Sprintf("#!/bin/sh\necho '%s'", fakeBinOut)), 0700) // #nosec G306
	Ok(t, err)
	defer tempSetEnv(t, "PATH", fmt.Sprintf("%s:%s", tmp, os.Getenv("PATH")))()

	c, err := terraform.NewClient(binDir, cacheDir, "0.11.10", cmd.DefaultTFVersionFlag, cmd.DefaultTFDownloadURL, nil, true, projectCmdOutputHandler)
	Ok(t, err)

	Ok(t, err)
	Equals(t, "0.11.10", c.DefaultVersion().String())

	output, err := c.RunCommandWithVersion(ctx, prjCtx, tmp, nil, map[string]string{}, nil, "")
	Ok(t, err)
	Equals(t, fakeBinOut+"\n", output)
}

// Test that if the default-tf flag is set and that binary is in our download
// bin dir that we use it.
func TestNewClient_DefaultTFFlagInBinDir(t *testing.T) {
	fakeBinOut := "Terraform v0.11.10\n"
	tmp, binDir, cacheDir, cleanup := mkSubDirs(t)
	projectCmdOutputHandler := jobmocks.NewMockProjectCommandOutputHandler()
	ctx := context.Background()
	prjCtx := command.ProjectContext{
		RequestCtx: context.TODO(),
		Log:        logging.NewNoopCtxLogger(t),
		Workspace:  "default",
		RepoRelDir: ".",
		BaseRepo:   models.Repo{FullName: "owner/repo"},
	}
	defer cleanup()

	// Add our fake binary to {datadir}/bin/terraform{version}.
	err := os.WriteFile(filepath.Join(binDir, "terraform0.11.10"), []byte(fmt.Sprintf("#!/bin/sh\necho '%s'", fakeBinOut)), 0700) // #nosec G306
	Ok(t, err)
	defer tempSetEnv(t, "PATH", fmt.Sprintf("%s:%s", tmp, os.Getenv("PATH")))()

	c, err := terraform.NewClient(binDir, cacheDir, "0.11.10", cmd.DefaultTFVersionFlag, cmd.DefaultTFDownloadURL, nil, true, projectCmdOutputHandler)
	Ok(t, err)

	Ok(t, err)
	Equals(t, "0.11.10", c.DefaultVersion().String())

	output, err := c.RunCommandWithVersion(ctx, prjCtx, tmp, nil, map[string]string{}, nil, "")
	Ok(t, err)
	Equals(t, fakeBinOut+"\n", output)
}

// Test that we get an error if the terraform version flag is malformed.
func TestNewClient_BadVersion(t *testing.T) {
	_, binDir, cacheDir, cleanup := mkSubDirs(t)
	projectCmdOutputHandler := jobmocks.NewMockProjectCommandOutputHandler()
	defer cleanup()

	_, err := terraform.NewClient(binDir, cacheDir, "malformed", cmd.DefaultTFVersionFlag, cmd.DefaultTFDownloadURL, nil, true, projectCmdOutputHandler)
	ErrEquals(t, "getting default version: parsing version malformed: Malformed version: malformed", err)
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
