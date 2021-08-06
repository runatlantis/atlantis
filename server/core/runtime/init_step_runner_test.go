package runtime_test

import (
	"io/ioutil"
	"path/filepath"
	"testing"

	version "github.com/hashicorp/go-version"
	. "github.com/petergtz/pegomock"
	"github.com/pkg/errors"

	"github.com/runatlantis/atlantis/server/core/runtime"
	"github.com/runatlantis/atlantis/server/core/terraform/mocks"
	matchers2 "github.com/runatlantis/atlantis/server/core/terraform/mocks/matchers"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/logging"
	. "github.com/runatlantis/atlantis/testing"
)

func TestRun_UsesGetOrInitForRightVersion(t *testing.T) {
	RegisterMockTestingT(t)
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
			terraform := mocks.NewMockClient()

			logger := logging.NewNoopLogger(t)

			tfVersion, _ := version.NewVersion(c.version)
			iso := runtime.InitStepRunner{
				TerraformExecutor: terraform,
				DefaultTFVersion:  tfVersion,
			}
			When(terraform.RunCommandWithVersion(matchers2.AnyModelsProjectCommandContext(), AnyString(), AnyStringSlice(), matchers2.AnyMapOfStringToString(), matchers2.AnyPtrToGoVersionVersion(), matchers2.AnyModelsParallelCommand())).
				ThenReturn("output", nil)
			context := models.ProjectCommandContext{
				Workspace:  "workspace",
				RepoRelDir: ".",
				Log:        logger,
			}

			output, err := iso.Run(context, []string{"extra", "args"}, "/path", map[string]string(nil), models.NotParallel)
			Ok(t, err)
			// When there is no error, should not return init output to PR.
			Equals(t, "", output)

			// If using init then we specify -input=false but not for get.
			expArgs := []string{c.expCmd, "-input=false", "-no-color", "-upgrade", "extra", "args"}
			if c.expCmd == "get" {
				expArgs = []string{c.expCmd, "-no-color", "-upgrade", "extra", "args"}
			}
			terraform.VerifyWasCalledOnce().RunCommandWithVersion(context, "/path", expArgs, map[string]string(nil), tfVersion, models.NotParallel)
		})
	}
}

func TestRun_ShowInitOutputOnError(t *testing.T) {
	// If there was an error during init then we want the output to be returned.
	RegisterMockTestingT(t)
	tfClient := mocks.NewMockClient()
	logger := logging.NewNoopLogger(t)
	When(tfClient.RunCommandWithVersion(matchers2.AnyModelsProjectCommandContext(), AnyString(), AnyStringSlice(), matchers2.AnyMapOfStringToString(), matchers2.AnyPtrToGoVersionVersion(), matchers2.AnyModelsParallelCommand())).
		ThenReturn("output", errors.New("error"))

	tfVersion, _ := version.NewVersion("0.11.0")
	iso := runtime.InitStepRunner{
		TerraformExecutor: tfClient,
		DefaultTFVersion:  tfVersion,
	}

	output, err := iso.Run(models.ProjectCommandContext{
		Workspace:  "workspace",
		RepoRelDir: ".",
		Log:        logger,
	}, nil, "/path", map[string]string(nil), models.NotParallel)
	ErrEquals(t, "error", err)
	Equals(t, "output", output)
}

func TestRun_InitOmitsUpgradeFlagIfLockFilePresent(t *testing.T) {
	tmpDir, cleanup := TempDir(t)
	defer cleanup()
	lockFilePath := filepath.Join(tmpDir, ".terraform.lock.hcl")
	err := ioutil.WriteFile(lockFilePath, nil, 0600)
	Ok(t, err)

	RegisterMockTestingT(t)
	terraform := mocks.NewMockClient()

	logger := logging.NewNoopLogger(t)

	tfVersion, _ := version.NewVersion("0.14.0")
	iso := runtime.InitStepRunner{
		TerraformExecutor: terraform,
		DefaultTFVersion:  tfVersion,
	}
	When(terraform.RunCommandWithVersion(matchers2.AnyModelsProjectCommandContext(), AnyString(), AnyStringSlice(), matchers2.AnyMapOfStringToString(), matchers2.AnyPtrToGoVersionVersion(), matchers2.AnyModelsParallelCommand())).
		ThenReturn("output", nil)

	context := models.ProjectCommandContext{
		Workspace:  "workspace",
		RepoRelDir: ".",
		Log:        logger,
	}

	output, err := iso.Run(context, []string{"extra", "args"}, tmpDir, map[string]string(nil), models.NotParallel)
	Ok(t, err)
	// When there is no error, should not return init output to PR.
	Equals(t, "", output)

	expectedArgs := []string{"init", "-input=false", "-no-color", "extra", "args"}
	terraform.VerifyWasCalledOnce().RunCommandWithVersion(context, tmpDir, expectedArgs, map[string]string(nil), tfVersion, models.NotParallel)
}

func TestRun_InitKeepsUpgradeFlagIfLockFileNotPresent(t *testing.T) {
	tmpDir, cleanup := TempDir(t)
	defer cleanup()

	RegisterMockTestingT(t)
	terraform := mocks.NewMockClient()

	logger := logging.NewNoopLogger(t)

	tfVersion, _ := version.NewVersion("0.14.0")
	iso := runtime.InitStepRunner{
		TerraformExecutor: terraform,
		DefaultTFVersion:  tfVersion,
	}
	When(terraform.RunCommandWithVersion(matchers2.AnyModelsProjectCommandContext(), AnyString(), AnyStringSlice(), matchers2.AnyMapOfStringToString(), matchers2.AnyPtrToGoVersionVersion(), matchers2.AnyModelsParallelCommand())).
		ThenReturn("output", nil)

	context := models.ProjectCommandContext{
		Workspace:  "workspace",
		RepoRelDir: ".",
		Log:        logger,
	}

	output, err := iso.Run(context, []string{"extra", "args"}, tmpDir, map[string]string(nil), models.NotParallel)
	Ok(t, err)
	// When there is no error, should not return init output to PR.
	Equals(t, "", output)

	expectedArgs := []string{"init", "-input=false", "-no-color", "-upgrade", "extra", "args"}
	terraform.VerifyWasCalledOnce().RunCommandWithVersion(context, tmpDir, expectedArgs, map[string]string(nil), tfVersion, models.NotParallel)
}

func TestRun_InitKeepUpgradeFlagIfLockFilePresentAndTFLessThanPoint14(t *testing.T) {
	tmpDir, cleanup := TempDir(t)
	defer cleanup()
	lockFilePath := filepath.Join(tmpDir, ".terraform.lock.hcl")
	err := ioutil.WriteFile(lockFilePath, nil, 0600)
	Ok(t, err)

	RegisterMockTestingT(t)
	terraform := mocks.NewMockClient()

	logger := logging.NewNoopLogger(t)

	tfVersion, _ := version.NewVersion("0.13.0")
	iso := runtime.InitStepRunner{
		TerraformExecutor: terraform,
		DefaultTFVersion:  tfVersion,
	}
	When(terraform.RunCommandWithVersion(matchers2.AnyModelsProjectCommandContext(), AnyString(), AnyStringSlice(), matchers2.AnyMapOfStringToString(), matchers2.AnyPtrToGoVersionVersion(), matchers2.AnyModelsParallelCommand())).
		ThenReturn("output", nil)

	context := models.ProjectCommandContext{
		Workspace:  "workspace",
		RepoRelDir: ".",
		Log:        logger,
	}
	output, err := iso.Run(context, []string{"extra", "args"}, tmpDir, map[string]string(nil), models.NotParallel)
	Ok(t, err)
	// When there is no error, should not return init output to PR.
	Equals(t, "", output)

	expectedArgs := []string{"init", "-input=false", "-no-color", "-upgrade", "extra", "args"}
	terraform.VerifyWasCalledOnce().RunCommandWithVersion(context, tmpDir, expectedArgs, map[string]string(nil), tfVersion, models.NotParallel)
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
			[]string{"init", "-input=false", "-no-color", "-upgrade"},
		},
		{
			"Override -upgrade",
			[]string{"-upgrade=false"},
			[]string{"init", "-input=false", "-no-color", "-upgrade=false"},
		},
		{
			"Override -input",
			[]string{"-input=true"},
			[]string{"init", "-input=true", "-no-color", "-upgrade"},
		},
		{
			"Override -input and -upgrade",
			[]string{"-input=true", "-upgrade=false"},
			[]string{"init", "-input=true", "-no-color", "-upgrade=false"},
		},
		{
			"Non duplicate extra args",
			[]string{"extra", "args"},
			[]string{"init", "-input=false", "-no-color", "-upgrade", "extra", "args"},
		},
		{
			"Override upgrade with extra args",
			[]string{"extra", "args", "-upgrade=false"},
			[]string{"init", "-input=false", "-no-color", "-upgrade=false", "extra", "args"},
		},
	}

	for _, c := range cases {
		t.Run(c.description, func(t *testing.T) {
			terraform := mocks.NewMockClient()

			logger := logging.NewNoopLogger(t)

			tfVersion, _ := version.NewVersion("0.10.0")
			iso := runtime.InitStepRunner{
				TerraformExecutor: terraform,
				DefaultTFVersion:  tfVersion,
			}
			When(terraform.RunCommandWithVersion(matchers2.AnyModelsProjectCommandContext(), AnyString(), AnyStringSlice(), matchers2.AnyMapOfStringToString(), matchers2.AnyPtrToGoVersionVersion(), matchers2.AnyModelsParallelCommand())).
				ThenReturn("output", nil)

			context := models.ProjectCommandContext{
				Workspace:  "workspace",
				RepoRelDir: ".",
				Log:        logger,
			}
			output, err := iso.Run(context, c.extraArgs, "/path", map[string]string(nil), models.NotParallel)
			Ok(t, err)
			// When there is no error, should not return init output to PR.
			Equals(t, "", output)

			terraform.VerifyWasCalledOnce().RunCommandWithVersion(context, "/path", c.expectedArgs, map[string]string(nil), tfVersion, models.NotParallel)
		})
	}
}
