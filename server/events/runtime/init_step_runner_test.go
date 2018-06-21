package runtime_test

import (
	"testing"

	version "github.com/hashicorp/go-version"
	. "github.com/petergtz/pegomock"
	"github.com/runatlantis/atlantis/server/events/mocks/matchers"
	"github.com/runatlantis/atlantis/server/events/models"
	matchers2 "github.com/runatlantis/atlantis/server/events/run/mocks/matchers"
	"github.com/runatlantis/atlantis/server/events/runtime"
	"github.com/runatlantis/atlantis/server/events/terraform/mocks"
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

			tfVersion, _ := version.NewVersion(c.version)
			logger := logging.NewNoopLogger()
			iso := runtime.InitStepRunner{
				TerraformExecutor: terraform,
				DefaultTFVersion:  tfVersion,
			}
			When(terraform.RunCommandWithVersion(matchers.AnyPtrToLoggingSimpleLogger(), AnyString(), AnyStringSlice(), matchers2.AnyPtrToGoVersionVersion(), AnyString())).
				ThenReturn("output", nil)

			output, err := iso.Run(models.ProjectCommandContext{
				Log:         logger,
				Workspace:   "workspace",
				RepoRelPath: ".",
			}, []string{"extra", "args"}, "/path")
			Ok(t, err)
			// Shouldn't return output since we don't print init output to PR.
			Equals(t, "", output)

			terraform.VerifyWasCalledOnce().RunCommandWithVersion(logger, "/path", []string{c.expCmd, "-no-color", "extra", "args"}, tfVersion, "workspace")
		})
	}
}
