package runtime_test

import (
	"io/ioutil"
	"path/filepath"
	"testing"

	version "github.com/hashicorp/go-version"
	. "github.com/petergtz/pegomock"
	"github.com/pkg/errors"

	"github.com/runatlantis/atlantis/server/events/mocks/matchers"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/events/runtime"
	"github.com/runatlantis/atlantis/server/events/terraform/mocks"
	matchers2 "github.com/runatlantis/atlantis/server/events/terraform/mocks/matchers"
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
			ctx := models.ProjectCommandContext{
				Workspace:  "workspace",
				RepoRelDir: ".",
				Log:        logger,
			}

			tfVersion, _ := version.NewVersion(c.version)
			iso := runtime.InitStepRunner{
				TerraformExecutor: terraform,
				DefaultTFVersion:  tfVersion,
			}
			When(terraform.RunCommandWithVersion(matchers.AnyModelsProjectCommandContext(), AnyString(), AnyStringSlice(), matchers2.AnyMapOfStringToString(), matchers2.AnyPtrToGoVersionVersion(), AnyString())).
				ThenReturn("output", nil)

			output, err := iso.Run(ctx, []string{"extra", "args"}, "/path", map[string]string(nil))
			Ok(t, err)
			// When there is no error, should not return init output to PR.
			Equals(t, "", output)

			// If using init then we specify -input=false but not for get.
			expArgs := []string{c.expCmd, "-input=false", "-no-color", "-upgrade", "extra", "args"}
			if c.expCmd == "get" {
				expArgs = []string{c.expCmd, "-no-color", "-upgrade", "extra", "args"}
			}
			terraform.VerifyWasCalledOnce().RunCommandWithVersion(ctx, "/path", expArgs, map[string]string(nil), tfVersion, "workspace")
		})
	}
}

func TestRun_ShowInitOutputOnError(t *testing.T) {
	// If there was an error during init then we want the output to be returned.
	RegisterMockTestingT(t)
	tfClient := mocks.NewMockClient()
	logger := logging.NewNoopLogger(t)
	When(tfClient.RunCommandWithVersion(matchers.AnyModelsProjectCommandContext(), AnyString(), AnyStringSlice(), matchers2.AnyMapOfStringToString(), matchers2.AnyPtrToGoVersionVersion(), AnyString())).
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
	}, nil, "/path", map[string]string(nil))
	ErrEquals(t, "error", err)
	Equals(t, "output", output)
}

func TestRun_InitOmitsUpgradeFlagIfLockFilePresent(t *testing.T) {
	tmpDir, cleanup := TempDir(t)
	defer cleanup()
	lockFilePath := filepath.Join(tmpDir, ".terraform.lock.hcl")
	err := ioutil.WriteFile(lockFilePath, nil, 0600)
	Ok(t, err)

	logger := logging.NewNoopLogger(t)
	ctx := models.ProjectCommandContext{
		Workspace:  "workspace",
		RepoRelDir: ".",
		Log:        logger,
	}

	RegisterMockTestingT(t)
	terraform := mocks.NewMockClient()

	tfVersion, _ := version.NewVersion("0.14.0")
	iso := runtime.InitStepRunner{
		TerraformExecutor: terraform,
		DefaultTFVersion:  tfVersion,
	}
	When(terraform.RunCommandWithVersion(matchers.AnyModelsProjectCommandContext(), AnyString(), AnyStringSlice(), matchers2.AnyMapOfStringToString(), matchers2.AnyPtrToGoVersionVersion(), AnyString())).
		ThenReturn("output", nil)

	output, err := iso.Run(ctx, []string{"extra", "args"}, tmpDir, map[string]string(nil))
	Ok(t, err)
	// When there is no error, should not return init output to PR.
	Equals(t, "", output)

	expectedArgs := []string{"init", "-input=false", "-no-color", "extra", "args"}
	terraform.VerifyWasCalledOnce().RunCommandWithVersion(ctx, tmpDir, expectedArgs, map[string]string(nil), tfVersion, "workspace")
}

func TestRun_InitKeepsUpgradeFlagIfLockFileNotPresent(t *testing.T) {
	tmpDir, cleanup := TempDir(t)
	defer cleanup()

	RegisterMockTestingT(t)
	terraform := mocks.NewMockClient()
	logger := logging.NewNoopLogger(t)
	ctx := models.ProjectCommandContext{
		Workspace:  "workspace",
		RepoRelDir: ".",
		Log:        logger,
	}

	tfVersion, _ := version.NewVersion("0.14.0")
	iso := runtime.InitStepRunner{
		TerraformExecutor: terraform,
		DefaultTFVersion:  tfVersion,
	}
	When(terraform.RunCommandWithVersion(matchers.AnyModelsProjectCommandContext(), AnyString(), AnyStringSlice(), matchers2.AnyMapOfStringToString(), matchers2.AnyPtrToGoVersionVersion(), AnyString())).
		ThenReturn("output", nil)

	output, err := iso.Run(ctx, []string{"extra", "args"}, tmpDir, map[string]string(nil))
	Ok(t, err)
	// When there is no error, should not return init output to PR.
	Equals(t, "", output)

	expectedArgs := []string{"init", "-input=false", "-no-color", "-upgrade", "extra", "args"}
	terraform.VerifyWasCalledOnce().RunCommandWithVersion(ctx, tmpDir, expectedArgs, map[string]string(nil), tfVersion, "workspace")
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
	ctx := models.ProjectCommandContext{
		Workspace:  "workspace",
		RepoRelDir: ".",
		Log:        logger,
	}

	tfVersion, _ := version.NewVersion("0.13.0")
	iso := runtime.InitStepRunner{
		TerraformExecutor: terraform,
		DefaultTFVersion:  tfVersion,
	}
	When(terraform.RunCommandWithVersion(matchers.AnyModelsProjectCommandContext(), AnyString(), AnyStringSlice(), matchers2.AnyMapOfStringToString(), matchers2.AnyPtrToGoVersionVersion(), AnyString())).
		ThenReturn("output", nil)

	output, err := iso.Run(ctx, []string{"extra", "args"}, tmpDir, map[string]string(nil))
	Ok(t, err)
	// When there is no error, should not return init output to PR.
	Equals(t, "", output)

	expectedArgs := []string{"init", "-input=false", "-no-color", "-upgrade", "extra", "args"}
	terraform.VerifyWasCalledOnce().RunCommandWithVersion(ctx, tmpDir, expectedArgs, map[string]string(nil), tfVersion, "workspace")
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
			ctx := models.ProjectCommandContext{
				Workspace:  "workspace",
				RepoRelDir: ".",
				Log:        logger,
			}

			tfVersion, _ := version.NewVersion("0.10.0")
			iso := runtime.InitStepRunner{
				TerraformExecutor: terraform,
				DefaultTFVersion:  tfVersion,
			}
			When(terraform.RunCommandWithVersion(matchers.AnyModelsProjectCommandContext(), AnyString(), AnyStringSlice(), matchers2.AnyMapOfStringToString(), matchers2.AnyPtrToGoVersionVersion(), AnyString())).
				ThenReturn("output", nil)

			output, err := iso.Run(ctx, c.extraArgs, "/path", map[string]string(nil))
			Ok(t, err)
			// When there is no error, should not return init output to PR.
			Equals(t, "", output)

			terraform.VerifyWasCalledOnce().RunCommandWithVersion(ctx, "/path", c.expectedArgs, map[string]string(nil), tfVersion, "workspace")
		})
	}
}
