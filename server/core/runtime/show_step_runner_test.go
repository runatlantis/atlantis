package runtime

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/hashicorp/go-version"
	"go.uber.org/mock/gomock"
	tf "github.com/runatlantis/atlantis/server/core/terraform"
	"github.com/runatlantis/atlantis/server/core/terraform/mocks"
	tfclientmocks "github.com/runatlantis/atlantis/server/core/terraform/tfclient/mocks"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/logging"
	. "github.com/runatlantis/atlantis/testing"
)

func TestShowStepRunnner(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	
	logger := logging.NewNoopLogger(t)
	path := t.TempDir()
	resultPath := filepath.Join(path, "test-default.json")
	envs := map[string]string{"key": "val"}
	mockDownloader := mocks.NewMockDownloader(ctrl)
	tfDistribution := tf.NewDistributionTerraformWithDownloader(mockDownloader)
	tfVersion, _ := version.NewVersion("0.12")
	context := command.ProjectContext{
		Workspace:   "default",
		ProjectName: "test",
		Log:         logger,
	}

	mockExecutor := tfclientmocks.NewMockClient(ctrl)

	subject := showStepRunner{
		terraformExecutor:     mockExecutor,
		defaultTfDistribution: tfDistribution,
		defaultTFVersion:      tfVersion,
	}

	t.Run("success", func(t *testing.T) {

		mockExecutor.EXPECT().RunCommandWithVersion(
			context, path, []string{"show", "-json", filepath.Join(path, "test-default.tfplan")}, envs, tfDistribution, tfVersion, context.Workspace,
		).Return("success", nil)

		r, err := subject.Run(context, []string{}, path, envs)

		Ok(t, err)

		actual, _ := os.ReadFile(resultPath)

		actualStr := string(actual)
		Assert(t, actualStr == "success", fmt.Sprintf("expected '%s' to be success", actualStr))
		Assert(t, r == "success", fmt.Sprintf("expected '%s' to be success", r))

	})

	t.Run("success w/ version override", func(t *testing.T) {

		v, _ := version.NewVersion("0.13.0")
		mockDownloader := mocks.NewMockDownloader()
		d := tf.NewDistributionTerraformWithDownloader(mockDownloader)

		contextWithVersionOverride := command.ProjectContext{
			Workspace:        "default",
			ProjectName:      "test",
			Log:              logger,
			TerraformVersion: v,
		}

		mockExecutor.EXPECT().RunCommandWithVersion(
			contextWithVersionOverride, path, []string{"show", "-json", filepath.Join(path, "test-default.tfplan")}, envs, d, v, context.Workspace,
		).Return("success", nil)

		r, err := subject.Run(contextWithVersionOverride, []string{}, path, envs)

		Ok(t, err)

		actual, _ := os.ReadFile(resultPath)

		actualStr := string(actual)
		Assert(t, actualStr == "success", "got expected result")
		Assert(t, r == "success", "returned expected result")

	})

	t.Run("success w/ distribution override", func(t *testing.T) {

		v, _ := version.NewVersion("0.13.0")
		mockDownloader := mocks.NewMockDownloader()
		d := tf.NewDistributionTerraformWithDownloader(mockDownloader)
		projTFDistribution := "opentofu"

		contextWithDistributionOverride := command.ProjectContext{
			Workspace:             "default",
			ProjectName:           "test",
			Log:                   logger,
			TerraformDistribution: &projTFDistribution,
		}

		mockExecutor.EXPECT().RunCommandWithVersion(
			gomock.Eq(contextWithDistributionOverride), gomock.Eq(path), gomock.Eq([]string{"show", "-json", filepath.Join(path, "test-default.tfplan")}), gomock.Eq(envs), gomock.Not(gomock.Eq(d)), gomock.Not(gomock.Eq(v)), gomock.Eq(context.Workspace),
		).Return("success", nil)

		r, err := subject.Run(contextWithDistributionOverride, []string{}, path, envs)

		Ok(t, err)

		actual, _ := os.ReadFile(resultPath)

		actualStr := string(actual)
		Assert(t, actualStr == "success", "got expected result")
		Assert(t, r == "success", "returned expected result")

	})

	t.Run("failure running command", func(t *testing.T) {
		mockExecutor.EXPECT().RunCommandWithVersion(
			context, path, []string{"show", "-json", filepath.Join(path, "test-default.tfplan")}, envs, tfDistribution, tfVersion, context.Workspace,
		).Return("success", errors.New("error"))

		_, err := subject.Run(context, []string{}, path, envs)

		Assert(t, err != nil, "error is returned")

	})

}
