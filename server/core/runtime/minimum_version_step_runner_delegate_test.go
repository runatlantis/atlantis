package runtime

import (
	"testing"

	"github.com/hashicorp/go-version"
	"go.uber.org/mock/gomock"
	"github.com/runatlantis/atlantis/server/core/runtime/mocks"
	"github.com/runatlantis/atlantis/server/events/command"
	. "github.com/runatlantis/atlantis/testing"
)

func TestRunMinimumVersionDelegate(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDelegate := mocks.NewMockRunner(ctrl)

	tfVersion12, _ := version.NewVersion("0.12.0")
	tfVersion11, _ := version.NewVersion("0.11.15")

	// these stay the same for all tests
	extraArgs := []string{"extra", "args"}
	envs := map[string]string{}
	path := ""

	expectedOut := "some valid output from delegate"

	t.Run("default version success", func(t *testing.T) {
		subject := &minimumVersionStepRunnerDelegate{
			defaultTfVersion: tfVersion12,
			minimumVersion:   tfVersion12,
			delegate:         mockDelegate,
		}

		ctx := command.ProjectContext{}

		mockDelegate.EXPECT().Run(ctx, extraArgs, path, envs).Return(expectedOut, nil)

		output, err := subject.Run(
			ctx,
			extraArgs,
			path,
			envs,
		)

		Equals(t, expectedOut, output)
		Ok(t, err)
	})

	t.Run("ctx version success", func(t *testing.T) {
		subject := &minimumVersionStepRunnerDelegate{
			defaultTfVersion: tfVersion11,
			minimumVersion:   tfVersion12,
			delegate:         mockDelegate,
		}

		ctx := command.ProjectContext{
			TerraformVersion: tfVersion12,
		}

		mockDelegate.EXPECT().Run(ctx, extraArgs, path, envs).Return(expectedOut, nil)

		output, err := subject.Run(
			ctx,
			extraArgs,
			path,
			envs,
		)

		Equals(t, expectedOut, output)
		Ok(t, err)
	})

	t.Run("default version failure", func(t *testing.T) {
		subject := &minimumVersionStepRunnerDelegate{
			defaultTfVersion: tfVersion11,
			minimumVersion:   tfVersion12,
			delegate:         mockDelegate,
		}

		ctx := command.ProjectContext{}

		mockDelegate.EXPECT().Run(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
		
		output, err := subject.Run(
			ctx,
			extraArgs,
			path,
			envs,
		)

		Equals(t, "Version: 0.11.15 is unsupported for this step. Minimum version is: 0.12.0", output)
		Ok(t, err)
	})

	t.Run("ctx version failure", func(t *testing.T) {
		subject := &minimumVersionStepRunnerDelegate{
			defaultTfVersion: tfVersion12,
			minimumVersion:   tfVersion12,
			delegate:         mockDelegate,
		}

		ctx := command.ProjectContext{
			TerraformVersion: tfVersion11,
		}

		mockDelegate.EXPECT().Run(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
		
		output, err := subject.Run(
			ctx,
			extraArgs,
			path,
			envs,
		)

		Equals(t, "Version: 0.11.15 is unsupported for this step. Minimum version is: 0.12.0", output)
		Ok(t, err)
	})

}
