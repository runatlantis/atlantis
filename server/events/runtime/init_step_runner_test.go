package runtime_test

import (
	"testing"

	version "github.com/hashicorp/go-version"
	. "github.com/petergtz/pegomock"
	"github.com/pkg/errors"

	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/events/runtime"
	"github.com/runatlantis/atlantis/server/events/terraform/mocks"
	matchers2 "github.com/runatlantis/atlantis/server/events/terraform/mocks/matchers"
	"github.com/runatlantis/atlantis/server/logging"
	logging_matchers "github.com/runatlantis/atlantis/server/logging/mocks/matchers"
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
			When(terraform.RunCommandWithVersion(logging_matchers.AnyLoggingSimpleLogging(), AnyString(), AnyStringSlice(), matchers2.AnyMapOfStringToString(), matchers2.AnyPtrToGoVersionVersion(), AnyString())).
				ThenReturn("output", nil)

			output, err := iso.Run(models.ProjectCommandContext{
				Workspace:  "workspace",
				RepoRelDir: ".",
				Log:        logger,
			}, []string{"extra", "args"}, "/path", map[string]string(nil))
			Ok(t, err)
			// When there is no error, should not return init output to PR.
			Equals(t, "", output)

			// If using init then we specify -input=false but not for get.
			expArgs := []string{c.expCmd, "-input=false", "-no-color", "-upgrade", "extra", "args"}
			if c.expCmd == "get" {
				expArgs = []string{c.expCmd, "-no-color", "-upgrade", "extra", "args"}
			}
			terraform.VerifyWasCalledOnce().RunCommandWithVersion(logger, "/path", expArgs, map[string]string(nil), tfVersion, "workspace")
		})
	}
}

func TestRun_ShowInitOutputOnError(t *testing.T) {
	// If there was an error during init then we want the output to be returned.
	RegisterMockTestingT(t)
	tfClient := mocks.NewMockClient()
	logger := logging.NewNoopLogger(t)
	When(tfClient.RunCommandWithVersion(logging_matchers.AnyLoggingSimpleLogging(), AnyString(), AnyStringSlice(), matchers2.AnyMapOfStringToString(), matchers2.AnyPtrToGoVersionVersion(), AnyString())).
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
