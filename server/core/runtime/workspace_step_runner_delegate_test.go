package runtime

import (
	"errors"
	"testing"

	"github.com/hashicorp/go-version"
	"go.uber.org/mock/gomock"
	tf "github.com/runatlantis/atlantis/server/core/terraform"
	"github.com/runatlantis/atlantis/server/core/terraform/mocks"
	tfclientmocks "github.com/runatlantis/atlantis/server/core/terraform/tfclient/mocks"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/logging"
	. "github.com/runatlantis/atlantis/testing"
)

func TestRun_NoWorkspaceIn08(t *testing.T) {
	// We don't want any workspace commands to be run in 0.8.
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	terraform := tfclientmocks.NewMockClient(ctrl)
	mockDownloader := mocks.NewMockDownloader(ctrl)
	tfDistribution := tf.NewDistributionTerraformWithDownloader(mockDownloader)
	tfVersion, _ := version.NewVersion("0.8")
	workspace := "default"
	logger := logging.NewNoopLogger(t)
	ctx := command.ProjectContext{
		Log:       logger,
		Workspace: workspace,
	}
	s := NewWorkspaceStepRunnerDelegate(terraform, tfDistribution, tfVersion, &NullRunner{})

	_, err := s.Run(ctx, []string{"extra", "args"}, "/path", map[string]string(nil))
	Ok(t, err)

	// Verify that no env or workspace commands were run
	// TODO: Convert Never() expectation: terraform.EXPECT().RunCommandWithVersion(...).Times(0)
}

func TestRun_ErrWorkspaceIn08(t *testing.T) {
	// If they attempt to use a workspace other than default in 0.8
	// we should error.
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	terraform := tfclientmocks.NewMockClient(ctrl)
	mockDownloader := mocks.NewMockDownloader(ctrl)
	tfDistribution := tf.NewDistributionTerraformWithDownloader(mockDownloader)
	tfVersion, _ := version.NewVersion("0.8")
	logger := logging.NewNoopLogger(t)
	workspace := "notdefault"
	s := NewWorkspaceStepRunnerDelegate(terraform, tfDistribution, tfVersion, &NullRunner{})

	_, err := s.Run(command.ProjectContext{
		Log:       logger,
		Workspace: workspace,
	}, []string{"extra", "args"}, "/path", map[string]string(nil))
	ErrEquals(t, "terraform version 0.8.0 does not support workspaces", err)
}

func TestRun_SwitchesWorkspace(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockDownloader := mocks.NewMockDownloader(ctrl)
	tfDistribution := tf.NewDistributionTerraformWithDownloader(mockDownloader)

	cases := []struct {
		tfVersion       string
		expWorkspaceCmd string
	}{
		{
			"0.9.0",
			"env",
		},
		{
			"0.9.11",
			"env",
		},
		{
			"0.10.0",
			"workspace",
		},
		{
			"0.11.0",
			"workspace",
		},
	}

	for _, c := range cases {
		t.Run(c.tfVersion, func(t *testing.T) {
			terraform := tfclientmocks.NewMockClient(ctrl)
			tfVersion, _ := version.NewVersion(c.tfVersion)
			logger := logging.NewNoopLogger(t)
			ctx := command.ProjectContext{
				Log:       logger,
				Workspace: "workspace",
			}
			s := NewWorkspaceStepRunnerDelegate(terraform, tfDistribution, tfVersion, &NullRunner{})

			_, err := s.Run(ctx, []string{"extra", "args"}, "/path", map[string]string(nil))
			Ok(t, err)

			// TODO: Convert to gomock expectation: terraform.EXPECT().RunCommandWithVersion(...)
		})
	}
}

func TestRun_SwitchesWorkspaceDistribution(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockDownloader := mocks.NewMockDownloader(ctrl)
	tfDistribution := tf.NewDistributionTerraformWithDownloader(mockDownloader)

	cases := []struct {
		tfVersion       string
		tfDistribution  string
		expWorkspaceCmd string
	}{
		{
			"0.9.0",
			"opentofu",
			"env",
		},
		{
			"0.9.11",
			"terraform",
			"env",
		},
		{
			"0.10.0",
			"terraform",
			"workspace",
		},
		{
			"0.11.0",
			"opentofu",
			"workspace",
		},
	}

	for _, c := range cases {
		t.Run(c.tfVersion, func(t *testing.T) {
			terraform := tfclientmocks.NewMockClient(ctrl)
			tfVersion, _ := version.NewVersion(c.tfVersion)
			logger := logging.NewNoopLogger(t)
			ctx := command.ProjectContext{
				Log:                   logger,
				Workspace:             "workspace",
				TerraformDistribution: &c.tfDistribution,
			}
			s := NewWorkspaceStepRunnerDelegate(terraform, tfDistribution, tfVersion, &NullRunner{})

			_, err := s.Run(ctx, []string{"extra", "args"}, "/path", map[string]string(nil))
			Ok(t, err)

			// TODO: Convert to gomock expectation: terraform.EXPECT().RunCommandWithVersion(...)
		})
	}
}

func TestRun_CreatesWorkspace(t *testing.T) {
	// Test that if `workspace select` fails, we call `workspace new`.
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cases := []struct {
		tfVersion           string
		expWorkspaceCommand string
	}{
		{
			"0.9.0",
			"env",
		},
		{
			"0.9.11",
			"env",
		},
		{
			"0.10.0",
			"workspace",
		},
		{
			"0.11.0",
			"workspace",
		},
	}

	for _, c := range cases {
		t.Run(c.tfVersion, func(t *testing.T) {
			terraform := tfclientmocks.NewMockClient(ctrl)
			mockDownloader := mocks.NewMockDownloader(ctrl)
			tfDistribution := tf.NewDistributionTerraformWithDownloader(mockDownloader)
			tfVersion, _ := version.NewVersion(c.tfVersion)
			logger := logging.NewNoopLogger(t)
			ctx := command.ProjectContext{
				Log:                logger,
				Workspace:          "workspace",
				RepoRelDir:         ".",
				User:               models.User{Username: "username"},
				EscapedCommentArgs: []string{"comment", "args"},
				Pull: models.PullRequest{
					Num: 2,
				},
				BaseRepo: models.Repo{
					FullName: "owner/repo",
					Owner:    "owner",
					Name:     "repo",
				},
			}
			s := NewWorkspaceStepRunnerDelegate(terraform, tfDistribution, tfVersion, &NullRunner{})

			// Ensure that we actually try to switch workspaces by making the
			// output of `workspace show` to be a different name.
			terraform.EXPECT().RunCommandWithVersion(ctx, "/path", []string{"workspace", "show"}, map[string]string(nil), tfDistribution, tfVersion, "workspace").Return("diffworkspace\n", nil)

			expWorkspaceArgs := []string{c.expWorkspaceCommand, "select", "workspace"}
			terraform.EXPECT().RunCommandWithVersion(ctx, "/path", expWorkspaceArgs, map[string]string(nil), tfDistribution, tfVersion, "workspace").Return("", errors.New("workspace does not exist"))

			_, err := s.Run(ctx, []string{"extra", "args"}, "/path", map[string]string(nil))
			Ok(t, err)

			// TODO: Convert to gomock expectation: terraform.EXPECT().RunCommandWithVersion(...)
		})
	}
}

func TestRun_NoWorkspaceSwitchIfNotNecessary(t *testing.T) {
	// Tests that if workspace show says we're on the right workspace we don't
	// switch.
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	terraform := tfclientmocks.NewMockClient(ctrl)
	mockDownloader := mocks.NewMockDownloader(ctrl)
	tfDistribution := tf.NewDistributionTerraformWithDownloader(mockDownloader)
	tfVersion, _ := version.NewVersion("0.10.0")
	logger := logging.NewNoopLogger(t)
	ctx := command.ProjectContext{
		Log:                logger,
		Workspace:          "workspace",
		RepoRelDir:         ".",
		User:               models.User{Username: "username"},
		EscapedCommentArgs: []string{"comment", "args"},
		Pull: models.PullRequest{
			Num: 2,
		},
		BaseRepo: models.Repo{
			FullName: "owner/repo",
			Owner:    "owner",
			Name:     "repo",
		},
	}
	s := NewWorkspaceStepRunnerDelegate(terraform, tfDistribution, tfVersion, &NullRunner{})
	terraform.EXPECT().RunCommandWithVersion(ctx, "/path", []string{"workspace", "show"}, map[string]string(nil), tfDistribution, tfVersion, "workspace").Return("workspace\n", nil)

	_, err := s.Run(ctx, []string{"extra", "args"}, "/path", map[string]string(nil))
	Ok(t, err)

	// TODO: Convert Never() expectation: terraform.EXPECT().RunCommandWithVersion(...).Times(0)
}
