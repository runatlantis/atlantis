package runtime_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/hashicorp/go-version"
	. "github.com/petergtz/pegomock"
	"github.com/pkg/errors"
	"github.com/runatlantis/atlantis/server/events/mocks/matchers"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/events/runtime"
	"github.com/runatlantis/atlantis/server/events/terraform/mocks"
	matchers2 "github.com/runatlantis/atlantis/server/events/terraform/mocks/matchers"
	"github.com/runatlantis/atlantis/server/events/yaml/valid"
	"github.com/runatlantis/atlantis/server/logging"
	. "github.com/runatlantis/atlantis/testing"
)

func TestRun_NoWorkspaceIn08(t *testing.T) {
	// We don't want any workspace commands to be run in 0.8.
	RegisterMockTestingT(t)
	terraform := mocks.NewMockClient()

	tfVersion, _ := version.NewVersion("0.8")
	logger := logging.NewNoopLogger()
	workspace := "default"
	s := runtime.PlanStepRunner{
		DefaultTFVersion:  tfVersion,
		TerraformExecutor: terraform,
	}

	When(terraform.RunCommandWithVersion(matchers.AnyPtrToLoggingSimpleLogger(), AnyString(), AnyStringSlice(), matchers2.AnyPtrToGoVersionVersion(), AnyString())).
		ThenReturn("output", nil)
	output, err := s.Run(models.ProjectCommandContext{
		Log:         logger,
		CommentArgs: []string{"comment", "args"},
		Workspace:   workspace,
		RepoRelDir:  ".",
		User:        models.User{Username: "username"},
		Pull: models.PullRequest{
			Num: 2,
		},
		BaseRepo: models.Repo{
			FullName: "owner/repo",
			Owner:    "owner",
			Name:     "repo",
		},
	}, []string{"extra", "args"}, "/path")
	Ok(t, err)

	Equals(t, "output", output)
	terraform.VerifyWasCalledOnce().RunCommandWithVersion(
		logger,
		"/path",
		[]string{"plan",
			"-input=false",
			"-refresh",
			"-no-color",
			"-out",
			"\"/path/default.tfplan\"",
			"-var",
			"atlantis_user=\"username\"",
			"-var",
			"atlantis_repo=\"owner/repo\"",
			"-var",
			"atlantis_repo_name=\"repo\"",
			"-var",
			"atlantis_repo_owner=\"owner\"",
			"-var",
			"atlantis_pull_num=2",
			"extra",
			"args",
			"comment",
			"args"},
		tfVersion,
		workspace)

	// Verify that no env or workspace commands were run
	terraform.VerifyWasCalled(Never()).RunCommandWithVersion(logger,
		"/path",
		[]string{"env",
			"select",
			"-no-color",
			"workspace"},
		tfVersion,
		workspace)
	terraform.VerifyWasCalled(Never()).RunCommandWithVersion(logger,
		"/path",
		[]string{"workspace",
			"select",
			"-no-color",
			"workspace"},
		tfVersion,
		workspace)
}

func TestRun_ErrWorkspaceIn08(t *testing.T) {
	// If they attempt to use a workspace other than default in 0.8
	// we should error.
	RegisterMockTestingT(t)
	terraform := mocks.NewMockClient()

	tfVersion, _ := version.NewVersion("0.8")
	logger := logging.NewNoopLogger()
	workspace := "notdefault"
	s := runtime.PlanStepRunner{
		TerraformExecutor: terraform,
		DefaultTFVersion:  tfVersion,
	}

	When(terraform.RunCommandWithVersion(matchers.AnyPtrToLoggingSimpleLogger(), AnyString(), AnyStringSlice(), matchers2.AnyPtrToGoVersionVersion(), AnyString())).
		ThenReturn("output", nil)
	_, err := s.Run(models.ProjectCommandContext{
		Log:        logger,
		Workspace:  workspace,
		RepoRelDir: ".",
		User:       models.User{Username: "username"},
	}, []string{"extra", "args"}, "/path")
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
			logger := logging.NewNoopLogger()

			s := runtime.PlanStepRunner{
				TerraformExecutor: terraform,
				DefaultTFVersion:  tfVersion,
			}

			When(terraform.RunCommandWithVersion(matchers.AnyPtrToLoggingSimpleLogger(), AnyString(), AnyStringSlice(), matchers2.AnyPtrToGoVersionVersion(), AnyString())).
				ThenReturn("output", nil)
			output, err := s.Run(models.ProjectCommandContext{
				Log:         logger,
				Workspace:   "workspace",
				RepoRelDir:  ".",
				User:        models.User{Username: "username"},
				CommentArgs: []string{"comment", "args"},
				Pull: models.PullRequest{
					Num: 2,
				},
				BaseRepo: models.Repo{
					FullName: "owner/repo",
					Owner:    "owner",
					Name:     "repo",
				},
			}, []string{"extra", "args"}, "/path")
			Ok(t, err)

			Equals(t, "output", output)
			// Verify that env select was called as well as plan.
			terraform.VerifyWasCalledOnce().RunCommandWithVersion(logger,
				"/path",
				[]string{c.expWorkspaceCmd,
					"select",
					"-no-color",
					"workspace"},
				tfVersion,
				"workspace")
			terraform.VerifyWasCalledOnce().RunCommandWithVersion(logger,
				"/path",
				[]string{"plan",
					"-input=false",
					"-refresh",
					"-no-color",
					"-out",
					"\"/path/workspace.tfplan\"",
					"-var",
					"atlantis_user=\"username\"",
					"-var",
					"atlantis_repo=\"owner/repo\"",
					"-var",
					"atlantis_repo_name=\"repo\"",
					"-var",
					"atlantis_repo_owner=\"owner\"",
					"-var",
					"atlantis_pull_num=2",
					"extra",
					"args",
					"comment",
					"args"},
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
			logger := logging.NewNoopLogger()
			s := runtime.PlanStepRunner{
				TerraformExecutor: terraform,
				DefaultTFVersion:  tfVersion,
			}

			// Ensure that we actually try to switch workspaces by making the
			// output of `workspace show` to be a different name.
			When(terraform.RunCommandWithVersion(logger, "/path", []string{"workspace", "show"}, tfVersion, "workspace")).ThenReturn("diffworkspace\n", nil)

			expWorkspaceArgs := []string{c.expWorkspaceCommand, "select", "-no-color", "workspace"}
			When(terraform.RunCommandWithVersion(logger, "/path", expWorkspaceArgs, tfVersion, "workspace")).ThenReturn("", errors.New("workspace does not exist"))

			expPlanArgs := []string{"plan",
				"-input=false",
				"-refresh",
				"-no-color",
				"-out",
				"\"/path/workspace.tfplan\"",
				"-var",
				"atlantis_user=\"username\"",
				"-var",
				"atlantis_repo=\"owner/repo\"",
				"-var",
				"atlantis_repo_name=\"repo\"",
				"-var",
				"atlantis_repo_owner=\"owner\"",
				"-var",
				"atlantis_pull_num=2",
				"extra",
				"args",
				"comment",
				"args"}
			When(terraform.RunCommandWithVersion(logger, "/path", expPlanArgs, tfVersion, "workspace")).ThenReturn("output", nil)

			output, err := s.Run(models.ProjectCommandContext{
				Log:         logger,
				Workspace:   "workspace",
				RepoRelDir:  ".",
				User:        models.User{Username: "username"},
				CommentArgs: []string{"comment", "args"},
				Pull: models.PullRequest{
					Num: 2,
				},
				BaseRepo: models.Repo{
					FullName: "owner/repo",
					Owner:    "owner",
					Name:     "repo",
				},
			}, []string{"extra", "args"}, "/path")
			Ok(t, err)

			Equals(t, "output", output)
			// Verify that env select was called as well as plan.
			terraform.VerifyWasCalledOnce().RunCommandWithVersion(logger, "/path", expWorkspaceArgs, tfVersion, "workspace")
			terraform.VerifyWasCalledOnce().RunCommandWithVersion(logger, "/path", expPlanArgs, tfVersion, "workspace")
		})
	}
}

func TestRun_NoWorkspaceSwitchIfNotNecessary(t *testing.T) {
	// Tests that if workspace show says we're on the right workspace we don't
	// switch.
	RegisterMockTestingT(t)
	terraform := mocks.NewMockClient()
	tfVersion, _ := version.NewVersion("0.10.0")
	logger := logging.NewNoopLogger()
	s := runtime.PlanStepRunner{
		TerraformExecutor: terraform,
		DefaultTFVersion:  tfVersion,
	}
	When(terraform.RunCommandWithVersion(logger, "/path", []string{"workspace", "show"}, tfVersion, "workspace")).ThenReturn("workspace\n", nil)

	expPlanArgs := []string{"plan",
		"-input=false",
		"-refresh",
		"-no-color",
		"-out",
		"\"/path/workspace.tfplan\"",
		"-var",
		"atlantis_user=\"username\"",
		"-var",
		"atlantis_repo=\"owner/repo\"",
		"-var",
		"atlantis_repo_name=\"repo\"",
		"-var",
		"atlantis_repo_owner=\"owner\"",
		"-var",
		"atlantis_pull_num=2",
		"extra",
		"args",
		"comment",
		"args"}
	When(terraform.RunCommandWithVersion(logger, "/path", expPlanArgs, tfVersion, "workspace")).ThenReturn("output", nil)

	output, err := s.Run(models.ProjectCommandContext{
		Log:         logger,
		Workspace:   "workspace",
		RepoRelDir:  ".",
		User:        models.User{Username: "username"},
		CommentArgs: []string{"comment", "args"},
		Pull: models.PullRequest{
			Num: 2,
		},
		BaseRepo: models.Repo{
			FullName: "owner/repo",
			Owner:    "owner",
			Name:     "repo",
		},
	}, []string{"extra", "args"}, "/path")
	Ok(t, err)

	Equals(t, "output", output)
	terraform.VerifyWasCalledOnce().RunCommandWithVersion(logger, "/path", expPlanArgs, tfVersion, "workspace")

	// Verify that workspace select was never called.
	terraform.VerifyWasCalled(Never()).RunCommandWithVersion(logger, "/path", []string{"workspace", "select", "-no-color", "workspace"}, tfVersion, "workspace")
}

func TestRun_AddsEnvVarFile(t *testing.T) {
	// Test that if env/workspace.tfvars file exists we use -var-file option.
	RegisterMockTestingT(t)
	terraform := mocks.NewMockClient()

	// Create the env/workspace.tfvars file.
	tmpDir, cleanup := TempDir(t)
	defer cleanup()
	err := os.MkdirAll(filepath.Join(tmpDir, "env"), 0700)
	Ok(t, err)
	envVarsFile := filepath.Join(tmpDir, "env/workspace.tfvars")
	err = ioutil.WriteFile(envVarsFile, nil, 0644)
	Ok(t, err)

	// Using version >= 0.10 here so we don't expect any env commands.
	tfVersion, _ := version.NewVersion("0.10.0")
	logger := logging.NewNoopLogger()
	s := runtime.PlanStepRunner{
		TerraformExecutor: terraform,
		DefaultTFVersion:  tfVersion,
	}

	expPlanArgs := []string{"plan",
		"-input=false",
		"-refresh",
		"-no-color",
		"-out",
		fmt.Sprintf("%q", filepath.Join(tmpDir, "workspace.tfplan")),
		"-var",
		"atlantis_user=\"username\"",
		"-var",
		"atlantis_repo=\"owner/repo\"",
		"-var",
		"atlantis_repo_name=\"repo\"",
		"-var",
		"atlantis_repo_owner=\"owner\"",
		"-var",
		"atlantis_pull_num=2",
		"extra",
		"args",
		"comment",
		"args",
		"-var-file",
		envVarsFile,
	}
	When(terraform.RunCommandWithVersion(logger, tmpDir, expPlanArgs, tfVersion, "workspace")).ThenReturn("output", nil)

	output, err := s.Run(models.ProjectCommandContext{
		Log:         logger,
		Workspace:   "workspace",
		RepoRelDir:  ".",
		User:        models.User{Username: "username"},
		CommentArgs: []string{"comment", "args"},
		Pull: models.PullRequest{
			Num: 2,
		},
		BaseRepo: models.Repo{
			FullName: "owner/repo",
			Owner:    "owner",
			Name:     "repo",
		},
	}, []string{"extra", "args"}, tmpDir)
	Ok(t, err)

	// Verify that env select was never called since we're in version >= 0.10
	terraform.VerifyWasCalled(Never()).RunCommandWithVersion(logger, tmpDir, []string{"env", "select", "-no-color", "workspace"}, tfVersion, "workspace")
	terraform.VerifyWasCalledOnce().RunCommandWithVersion(logger, tmpDir, expPlanArgs, tfVersion, "workspace")
	Equals(t, "output", output)
}

func TestRun_UsesDiffPathForProject(t *testing.T) {
	// Test that if running for a project, uses a different path for the plan
	// file.
	RegisterMockTestingT(t)
	terraform := mocks.NewMockClient()
	tfVersion, _ := version.NewVersion("0.10.0")
	logger := logging.NewNoopLogger()
	s := runtime.PlanStepRunner{
		TerraformExecutor: terraform,
		DefaultTFVersion:  tfVersion,
	}
	When(terraform.RunCommandWithVersion(logger, "/path", []string{"workspace", "show"}, tfVersion, "workspace")).ThenReturn("workspace\n", nil)

	expPlanArgs := []string{"plan",
		"-input=false",
		"-refresh",
		"-no-color",
		"-out",
		"\"/path/projectname-default.tfplan\"",
		"-var",
		"atlantis_user=\"username\"",
		"-var",
		"atlantis_repo=\"owner/repo\"",
		"-var",
		"atlantis_repo_name=\"repo\"",
		"-var",
		"atlantis_repo_owner=\"owner\"",
		"-var",
		"atlantis_pull_num=2",
		"extra",
		"args",
		"comment",
		"args",
	}
	When(terraform.RunCommandWithVersion(logger, "/path", expPlanArgs, tfVersion, "default")).ThenReturn("output", nil)

	projectName := "projectname"
	output, err := s.Run(models.ProjectCommandContext{
		Log:         logger,
		Workspace:   "default",
		RepoRelDir:  ".",
		User:        models.User{Username: "username"},
		CommentArgs: []string{"comment", "args"},
		ProjectConfig: &valid.Project{
			Name: &projectName,
		},
		Pull: models.PullRequest{
			Num: 2,
		},
		BaseRepo: models.Repo{
			FullName: "owner/repo",
			Owner:    "owner",
			Name:     "repo",
		},
	}, []string{"extra", "args"}, "/path")
	Ok(t, err)
	Equals(t, "output", output)
}
