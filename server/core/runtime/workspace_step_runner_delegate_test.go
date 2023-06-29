package runtime

import (
	"errors"
	"testing"

	"github.com/hashicorp/go-version"
	. "github.com/petergtz/pegomock/v4"
	"github.com/runatlantis/atlantis/server/core/terraform/mocks"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/logging"
	. "github.com/runatlantis/atlantis/testing"
)

func TestRun_NoWorkspaceIn08(t *testing.T) {
	// We don't want any workspace commands to be run in 0.8.
	RegisterMockTestingT(t)
	terraform := mocks.NewMockClient()
	tfVersion, _ := version.NewVersion("0.8")
	workspace := "default"
	logger := logging.NewNoopLogger(t)
	ctx := command.ProjectContext{
		Log:       logger,
		Workspace: workspace,
	}
	s := NewWorkspaceStepRunnerDelegate(terraform, tfVersion, &NullRunner{})

	_, err := s.Run(ctx, []string{"extra", "args"}, "/path", map[string]string(nil))
	Ok(t, err)

	// Verify that no env or workspace commands were run
	terraform.VerifyWasCalled(Never()).RunCommandWithVersion(ctx,
		"/path",
		[]string{"env",
			"select",
			"workspace"},
		map[string]string(nil),
		tfVersion,
		workspace)
	terraform.VerifyWasCalled(Never()).RunCommandWithVersion(ctx,
		"/path",
		[]string{"workspace",
			"select",
			"workspace"},
		map[string]string(nil),
		tfVersion,
		workspace)
}

func TestRun_ErrWorkspaceIn08(t *testing.T) {
	// If they attempt to use a workspace other than default in 0.8
	// we should error.
	RegisterMockTestingT(t)
	terraform := mocks.NewMockClient()
	tfVersion, _ := version.NewVersion("0.8")
	logger := logging.NewNoopLogger(t)
	workspace := "notdefault"
	s := NewWorkspaceStepRunnerDelegate(terraform, tfVersion, &NullRunner{})

	_, err := s.Run(command.ProjectContext{
		Log:       logger,
		Workspace: workspace,
	}, []string{"extra", "args"}, "/path", map[string]string(nil))
	ErrEquals(t, "terraform version 0.8.0 does not support workspaces", err)
}

func TestRun_SwitchesWorkspace(t *testing.T) {
	RegisterMockTestingT(t)

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
			terraform := mocks.NewMockClient()
			tfVersion, _ := version.NewVersion(c.tfVersion)
			logger := logging.NewNoopLogger(t)
			ctx := command.ProjectContext{
				Log:       logger,
				Workspace: "workspace",
			}
			s := NewWorkspaceStepRunnerDelegate(terraform, tfVersion, &NullRunner{})

			_, err := s.Run(ctx, []string{"extra", "args"}, "/path", map[string]string(nil))
			Ok(t, err)

			// Verify that env select was called as well as plan.
			terraform.VerifyWasCalledOnce().RunCommandWithVersion(ctx,
				"/path",
				[]string{c.expWorkspaceCmd,
					"select",
					"workspace"},
				map[string]string(nil),
				tfVersion,
				"workspace")
		})
	}
}

func TestRun_CreatesWorkspace(t *testing.T) {
	// Test that if `workspace select` fails, we call `workspace new`.
	RegisterMockTestingT(t)

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
			terraform := mocks.NewMockClient()
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
			s := NewWorkspaceStepRunnerDelegate(terraform, tfVersion, &NullRunner{})

			// Ensure that we actually try to switch workspaces by making the
			// output of `workspace show` to be a different name.
			When(terraform.RunCommandWithVersion(ctx, "/path", []string{"workspace", "show"}, map[string]string(nil), tfVersion, "workspace")).ThenReturn("diffworkspace\n", nil)

			expWorkspaceArgs := []string{c.expWorkspaceCommand, "select", "workspace"}
			When(terraform.RunCommandWithVersion(ctx, "/path", expWorkspaceArgs, map[string]string(nil), tfVersion, "workspace")).ThenReturn("", errors.New("workspace does not exist"))

			_, err := s.Run(ctx, []string{"extra", "args"}, "/path", map[string]string(nil))
			Ok(t, err)

			// Verify that env select was called as well as plan.
			terraform.VerifyWasCalledOnce().RunCommandWithVersion(ctx, "/path", expWorkspaceArgs, map[string]string(nil), tfVersion, "workspace")
		})
	}
}

func TestRun_NoWorkspaceSwitchIfNotNecessary(t *testing.T) {
	// Tests that if workspace show says we're on the right workspace we don't
	// switch.
	RegisterMockTestingT(t)
	terraform := mocks.NewMockClient()
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
	s := NewWorkspaceStepRunnerDelegate(terraform, tfVersion, &NullRunner{})
	When(terraform.RunCommandWithVersion(ctx, "/path", []string{"workspace", "show"}, map[string]string(nil), tfVersion, "workspace")).ThenReturn("workspace\n", nil)

	_, err := s.Run(ctx, []string{"extra", "args"}, "/path", map[string]string(nil))
	Ok(t, err)

	// Verify that workspace select was never called.
	terraform.VerifyWasCalled(Never()).RunCommandWithVersion(ctx, "/path", []string{"workspace", "select", "workspace"}, map[string]string(nil), tfVersion, "workspace")
}
