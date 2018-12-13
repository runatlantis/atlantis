package runtime_test

import (
	"io/ioutil"
	"os"
	"path"
	"testing"

	version "github.com/hashicorp/go-version"
	. "github.com/petergtz/pegomock"
	"github.com/pkg/errors"
	"github.com/runatlantis/atlantis/server/events/mocks/matchers"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/events/runtime"
	"github.com/runatlantis/atlantis/server/events/terraform/mocks"
	matchers2 "github.com/runatlantis/atlantis/server/events/terraform/mocks/matchers"
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

			tfVersion, _ := version.NewVersion(c.version)
			iso := runtime.InitStepRunner{
				TerraformExecutor: terraform,
				DefaultTFVersion:  tfVersion,
			}
			When(terraform.RunCommandWithVersion(matchers.AnyPtrToLoggingSimpleLogger(), AnyString(), AnyStringSlice(), matchers2.AnyPtrToGoVersionVersion(), AnyString())).
				ThenReturn("output", nil)

			dir, cleanup := TempDir(t)
			defer cleanup()

			output, err := iso.Run(models.ProjectCommandContext{
				Workspace:  "workspace",
				RepoRelDir: ".",
			}, []string{"extra", "args"}, dir)
			Ok(t, err)
			// When there is no error, should not return init output to PR.
			Equals(t, "", output)

			// If using init then we specify -input=false but not for get.
			expArgs := []string{c.expCmd, "-input=false", "-no-color", "extra", "args"}
			if c.expCmd == "get" {
				expArgs = []string{c.expCmd, "-no-color", "extra", "args"}
			}
			terraform.VerifyWasCalledOnce().RunCommandWithVersion(nil, dir, expArgs, tfVersion, "workspace")
		})
	}
}

func TestRun_ShowInitOutputOnError(t *testing.T) {
	// If there was an error during init then we want the output to be returned.
	RegisterMockTestingT(t)
	tfClient := mocks.NewMockClient()
	When(tfClient.RunCommandWithVersion(matchers.AnyPtrToLoggingSimpleLogger(), AnyString(), AnyStringSlice(), matchers2.AnyPtrToGoVersionVersion(), AnyString())).
		ThenReturn("output", errors.New("error"))

	tfVersion, _ := version.NewVersion("0.11.0")
	iso := runtime.InitStepRunner{
		TerraformExecutor: tfClient,
		DefaultTFVersion:  tfVersion,
	}

	dir, cleanup := TempDir(t)
	defer cleanup()

	output, err := iso.Run(models.ProjectCommandContext{
		Workspace:  "workspace",
		RepoRelDir: ".",
	}, nil, dir)
	ErrEquals(t, "error", err)
	Equals(t, "output", output)

	// Check error was written to `workspace.tfinit-error`
	errorFilePath := path.Join(dir, "workspace.tfinit-error")
	contents, err := ioutil.ReadFile(errorFilePath)
	Ok(t, err)
	Equals(t, output, string(contents))
	err = os.Remove(errorFilePath)
	Ok(t, err)
}
