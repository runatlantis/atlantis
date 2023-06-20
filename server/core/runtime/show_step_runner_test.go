package runtime

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/hashicorp/go-version"
	. "github.com/petergtz/pegomock/v4"
	"github.com/runatlantis/atlantis/server/core/terraform/mocks"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/logging"
	. "github.com/runatlantis/atlantis/testing"
)

func TestShowStepRunnner(t *testing.T) {
	logger := logging.NewNoopLogger(t)
	path := t.TempDir()
	resultPath := filepath.Join(path, "test-default.json")
	envs := map[string]string{"key": "val"}
	tfVersion, _ := version.NewVersion("0.12")
	context := command.ProjectContext{
		Workspace:   "default",
		ProjectName: "test",
		Log:         logger,
	}

	RegisterMockTestingT(t)

	mockExecutor := mocks.NewMockClient()

	subject := showStepRunner{
		terraformExecutor: mockExecutor,
		defaultTFVersion:  tfVersion,
	}

	t.Run("success", func(t *testing.T) {

		When(mockExecutor.RunCommandWithVersion(
			context, path, []string{"show", "-json", filepath.Join(path, "test-default.tfplan")}, envs, tfVersion, context.Workspace,
		)).ThenReturn("success", nil)

		r, err := subject.Run(context, []string{}, path, envs)

		Ok(t, err)

		actual, _ := os.ReadFile(resultPath)

		actualStr := string(actual)
		Assert(t, actualStr == "success", fmt.Sprintf("expected '%s' to be success", actualStr))
		Assert(t, r == "success", fmt.Sprintf("expected '%s' to be success", r))

	})

	t.Run("success w/ version override", func(t *testing.T) {

		v, _ := version.NewVersion("0.13.0")

		contextWithVersionOverride := command.ProjectContext{
			Workspace:        "default",
			ProjectName:      "test",
			Log:              logger,
			TerraformVersion: v,
		}

		When(mockExecutor.RunCommandWithVersion(
			contextWithVersionOverride, path, []string{"show", "-json", filepath.Join(path, "test-default.tfplan")}, envs, v, context.Workspace,
		)).ThenReturn("success", nil)

		r, err := subject.Run(contextWithVersionOverride, []string{}, path, envs)

		Ok(t, err)

		actual, _ := os.ReadFile(resultPath)

		actualStr := string(actual)
		Assert(t, actualStr == "success", "got expected result")
		Assert(t, r == "success", "returned expected result")

	})

	t.Run("failure running command", func(t *testing.T) {
		When(mockExecutor.RunCommandWithVersion(
			context, path, []string{"show", "-json", filepath.Join(path, "test-default.tfplan")}, envs, tfVersion, context.Workspace,
		)).ThenReturn("success", errors.New("error"))

		_, err := subject.Run(context, []string{}, path, envs)

		Assert(t, err != nil, "error is returned")

	})

}
