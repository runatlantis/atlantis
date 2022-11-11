package runtime

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/hashicorp/go-version"
	. "github.com/petergtz/pegomock"
	"github.com/runatlantis/atlantis/server/core/terraform/mocks"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/logging"
	. "github.com/runatlantis/atlantis/testing"
)

func TestShowStepRunnner(t *testing.T) {
	logger := logging.NewNoopCtxLogger(t)
	path := t.TempDir()
	resultPath := filepath.Join(path, "test-default.json")
	envs := map[string]string{"key": "val"}
	tfVersion, _ := version.NewVersion("0.12")
	ctx := context.Background()
	prjCtx := command.ProjectContext{
		Workspace:   "default",
		ProjectName: "test",
		Log:         logger,
	}

	RegisterMockTestingT(t)

	mockExecutor := mocks.NewMockClient()

	subject := ShowStepRunner{
		TerraformExecutor: mockExecutor,
		DefaultTFVersion:  tfVersion,
	}

	t.Run("success", func(t *testing.T) {

		When(mockExecutor.RunCommandWithVersion(
			ctx, prjCtx, path, []string{"show", "-json", filepath.Join(path, "test-default.tfplan")}, envs, tfVersion, prjCtx.Workspace,
		)).ThenReturn("success", nil)

		_, err := subject.Run(ctx, prjCtx, []string{}, path, envs)

		Ok(t, err)

		actual, _ := os.ReadFile(resultPath)

		actualStr := string(actual)
		Assert(t, actualStr == "success", fmt.Sprintf("expected '%s' to be success", actualStr))

	})

	t.Run("success w/ version override", func(t *testing.T) {

		v, _ := version.NewVersion("0.13.0")

		ctx := context.Background()
		prjCtx := command.ProjectContext{
			Workspace:        "default",
			ProjectName:      "test",
			Log:              logger,
			TerraformVersion: v,
		}

		When(mockExecutor.RunCommandWithVersion(
			ctx, prjCtx, path, []string{"show", "-json", filepath.Join(path, "test-default.tfplan")}, envs, v, prjCtx.Workspace,
		)).ThenReturn("success", nil)

		_, err := subject.Run(ctx, prjCtx, []string{}, path, envs)

		Ok(t, err)

		actual, _ := os.ReadFile(resultPath)

		actualStr := string(actual)
		Assert(t, actualStr == "success", "got expected result")

	})

	t.Run("failure running command", func(t *testing.T) {
		When(mockExecutor.RunCommandWithVersion(
			ctx, prjCtx, path, []string{"show", "-json", filepath.Join(path, "test-default.tfplan")}, envs, tfVersion, prjCtx.Workspace,
		)).ThenReturn("err", errors.New("error"))

		r, err := subject.Run(ctx, prjCtx, []string{}, path, envs)

		Assert(t, err != nil, "error is returned")
		Assert(t, r == "err", "returned expected result")

	})

}
