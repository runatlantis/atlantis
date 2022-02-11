package runtime_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/hashicorp/go-version"
	"github.com/runatlantis/atlantis/server/core/terraform"
	mocks2 "github.com/runatlantis/atlantis/server/events/mocks"

	. "github.com/petergtz/pegomock"
	"github.com/pkg/errors"
	"github.com/runatlantis/atlantis/server/core/runtime"
	"github.com/runatlantis/atlantis/server/core/terraform/mocks"
	matchers2 "github.com/runatlantis/atlantis/server/core/terraform/mocks/matchers"
	"github.com/runatlantis/atlantis/server/events/mocks/matchers"
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
	ctx := models.ProjectCommandContext{
		Log:                logger,
		EscapedCommentArgs: []string{"comment", "args"},
		Workspace:          workspace,
		RepoRelDir:         ".",
		User:               models.User{Username: "username"},
		Pull: models.PullRequest{
			Num: 2,
		},
		BaseRepo: models.Repo{
			FullName: "owner/repo",
			Owner:    "owner",
			Name:     "repo",
		},
	}
	s := runtime.PlanStepRunner{
		DefaultTFVersion:  tfVersion,
		TerraformExecutor: terraform,
	}

	When(terraform.RunCommandWithVersion(matchers.AnyModelsProjectCommandContext(), AnyString(), AnyStringSlice(), matchers2.AnyMapOfStringToString(), matchers2.AnyPtrToGoVersionVersion(), AnyString())).
		ThenReturn("output", nil)
	output, err := s.Run(ctx, []string{"extra", "args"}, "/path", map[string]string(nil))
	Ok(t, err)

	Equals(t, "output", output)
	terraform.VerifyWasCalledOnce().RunCommandWithVersion(
		ctx,
		"/path",
		[]string{"plan",
			"-input=false",
			"-refresh",
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
		map[string]string(nil),
		tfVersion,
		workspace)

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
	s := runtime.PlanStepRunner{
		TerraformExecutor: terraform,
		DefaultTFVersion:  tfVersion,
	}

	When(terraform.RunCommandWithVersion(matchers.AnyModelsProjectCommandContext(), AnyString(), AnyStringSlice(), matchers2.AnyMapOfStringToString(), matchers2.AnyPtrToGoVersionVersion(), AnyString())).
		ThenReturn("output", nil)
	_, err := s.Run(models.ProjectCommandContext{
		Log:        logger,
		Workspace:  workspace,
		RepoRelDir: ".",
		User:       models.User{Username: "username"},
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
			ctx := models.ProjectCommandContext{
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
			s := runtime.PlanStepRunner{
				TerraformExecutor: terraform,
				DefaultTFVersion:  tfVersion,
			}

			When(terraform.RunCommandWithVersion(matchers.AnyModelsProjectCommandContext(), AnyString(), AnyStringSlice(), matchers2.AnyMapOfStringToString(), matchers2.AnyPtrToGoVersionVersion(), AnyString())).
				ThenReturn("output", nil)
			output, err := s.Run(ctx, []string{"extra", "args"}, "/path", map[string]string(nil))
			Ok(t, err)

			Equals(t, "output", output)
			// Verify that env select was called as well as plan.
			terraform.VerifyWasCalledOnce().RunCommandWithVersion(ctx,
				"/path",
				[]string{c.expWorkspaceCmd,
					"select",
					"workspace"},
				map[string]string(nil),
				tfVersion,
				"workspace")
			terraform.VerifyWasCalledOnce().RunCommandWithVersion(ctx,
				"/path",
				[]string{"plan",
					"-input=false",
					"-refresh",
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
			ctx := models.ProjectCommandContext{
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
			s := runtime.PlanStepRunner{
				TerraformExecutor: terraform,
				DefaultTFVersion:  tfVersion,
			}

			// Ensure that we actually try to switch workspaces by making the
			// output of `workspace show` to be a different name.
			When(terraform.RunCommandWithVersion(ctx, "/path", []string{"workspace", "show"}, map[string]string(nil), tfVersion, "workspace")).ThenReturn("diffworkspace\n", nil)

			expWorkspaceArgs := []string{c.expWorkspaceCommand, "select", "workspace"}
			When(terraform.RunCommandWithVersion(ctx, "/path", expWorkspaceArgs, map[string]string(nil), tfVersion, "workspace")).ThenReturn("", errors.New("workspace does not exist"))

			expPlanArgs := []string{"plan",
				"-input=false",
				"-refresh",
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
			When(terraform.RunCommandWithVersion(ctx, "/path", expPlanArgs, map[string]string(nil), tfVersion, "workspace")).ThenReturn("output", nil)

			output, err := s.Run(ctx, []string{"extra", "args"}, "/path", map[string]string(nil))
			Ok(t, err)

			Equals(t, "output", output)
			// Verify that env select was called as well as plan.
			terraform.VerifyWasCalledOnce().RunCommandWithVersion(ctx, "/path", expWorkspaceArgs, map[string]string(nil), tfVersion, "workspace")
			terraform.VerifyWasCalledOnce().RunCommandWithVersion(ctx, "/path", expPlanArgs, map[string]string(nil), tfVersion, "workspace")
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
	ctx := models.ProjectCommandContext{
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
	s := runtime.PlanStepRunner{
		TerraformExecutor: terraform,
		DefaultTFVersion:  tfVersion,
	}
	When(terraform.RunCommandWithVersion(ctx, "/path", []string{"workspace", "show"}, map[string]string(nil), tfVersion, "workspace")).ThenReturn("workspace\n", nil)

	expPlanArgs := []string{"plan",
		"-input=false",
		"-refresh",
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
	When(terraform.RunCommandWithVersion(ctx, "/path", expPlanArgs, map[string]string(nil), tfVersion, "workspace")).ThenReturn("output", nil)

	output, err := s.Run(ctx, []string{"extra", "args"}, "/path", map[string]string(nil))
	Ok(t, err)

	Equals(t, "output", output)
	terraform.VerifyWasCalledOnce().RunCommandWithVersion(ctx, "/path", expPlanArgs, map[string]string(nil), tfVersion, "workspace")

	// Verify that workspace select was never called.
	terraform.VerifyWasCalled(Never()).RunCommandWithVersion(ctx, "/path", []string{"workspace", "select", "workspace"}, map[string]string(nil), tfVersion, "workspace")
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
	err = os.WriteFile(envVarsFile, nil, 0600)
	Ok(t, err)

	// Using version >= 0.10 here so we don't expect any env commands.
	tfVersion, _ := version.NewVersion("0.10.0")
	logger := logging.NewNoopLogger(t)
	s := runtime.PlanStepRunner{
		TerraformExecutor: terraform,
		DefaultTFVersion:  tfVersion,
	}

	expPlanArgs := []string{"plan",
		"-input=false",
		"-refresh",
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
	ctx := models.ProjectCommandContext{
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
	When(terraform.RunCommandWithVersion(ctx, tmpDir, expPlanArgs, map[string]string(nil), tfVersion, "workspace")).ThenReturn("output", nil)

	output, err := s.Run(ctx, []string{"extra", "args"}, tmpDir, map[string]string(nil))
	Ok(t, err)

	// Verify that env select was never called since we're in version >= 0.10
	terraform.VerifyWasCalled(Never()).RunCommandWithVersion(ctx, tmpDir, []string{"env", "select", "workspace"}, map[string]string(nil), tfVersion, "workspace")
	terraform.VerifyWasCalledOnce().RunCommandWithVersion(ctx, tmpDir, expPlanArgs, map[string]string(nil), tfVersion, "workspace")
	Equals(t, "output", output)
}

func TestRun_UsesDiffPathForProject(t *testing.T) {
	// Test that if running for a project, uses a different path for the plan
	// file.
	RegisterMockTestingT(t)
	terraform := mocks.NewMockClient()
	tfVersion, _ := version.NewVersion("0.10.0")
	logger := logging.NewNoopLogger(t)
	s := runtime.PlanStepRunner{
		TerraformExecutor: terraform,
		DefaultTFVersion:  tfVersion,
	}
	ctx := models.ProjectCommandContext{
		Log:                logger,
		Workspace:          "default",
		RepoRelDir:         ".",
		User:               models.User{Username: "username"},
		EscapedCommentArgs: []string{"comment", "args"},
		ProjectName:        "projectname",
		Pull: models.PullRequest{
			Num: 2,
		},
		BaseRepo: models.Repo{
			FullName: "owner/repo",
			Owner:    "owner",
			Name:     "repo",
		},
	}
	When(terraform.RunCommandWithVersion(ctx, "/path", []string{"workspace", "show"}, map[string]string(nil), tfVersion, "workspace")).ThenReturn("workspace\n", nil)

	expPlanArgs := []string{"plan",
		"-input=false",
		"-refresh",
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
	When(terraform.RunCommandWithVersion(ctx, "/path", expPlanArgs, map[string]string(nil), tfVersion, "default")).ThenReturn("output", nil)

	output, err := s.Run(ctx, []string{"extra", "args"}, "/path", map[string]string(nil))
	Ok(t, err)
	Equals(t, "output", output)
}

// Test that we format the plan output for better rendering.
func TestRun_PlanFmt(t *testing.T) {
	rawOutput := `Refreshing Terraform state in-memory prior to plan...
The refreshed state will be used to calculate this plan, but will not be
persisted to local or remote state storage.


------------------------------------------------------------------------

An execution plan has been generated and is shown below.
Resource actions are indicated with the following symbols:
  + create
  ~ update in-place
  - destroy

Terraform will perform the following actions:

+ null_resource.test[0]
      id: <computed>

  + null_resource.test[1]
      id: <computed>

  ~ aws_security_group_rule.allow_all
      description: "" => "test3"

  - aws_security_group_rule.allow_all
`
	RegisterMockTestingT(t)
	terraform := mocks.NewMockClient()
	tfVersion, _ := version.NewVersion("0.10.0")
	s := runtime.PlanStepRunner{
		TerraformExecutor: terraform,
		DefaultTFVersion:  tfVersion,
	}
	When(terraform.RunCommandWithVersion(
		matchers.AnyModelsProjectCommandContext(),
		AnyString(),
		AnyStringSlice(),
		matchers2.AnyMapOfStringToString(),
		matchers2.AnyPtrToGoVersionVersion(),
		AnyString())).
		Then(func(params []Param) ReturnValues {
			// This code allows us to return different values depending on the
			// tf command being run while still using the wildcard matchers above.
			tfArgs := params[2].([]string)
			if stringSliceEquals(tfArgs, []string{"workspace", "show"}) {
				return []ReturnValue{"default", nil}
			} else if tfArgs[0] == "plan" {
				return []ReturnValue{rawOutput, nil}
			} else {
				return []ReturnValue{"", errors.New("unexpected call to RunCommandWithVersion")}
			}
		})
	actOutput, err := s.Run(models.ProjectCommandContext{Workspace: "default"}, nil, "", map[string]string(nil))
	Ok(t, err)
	Equals(t, `
An execution plan has been generated and is shown below.
Resource actions are indicated with the following symbols:
+ create
~ update in-place
- destroy

Terraform will perform the following actions:

+ null_resource.test[0]
      id: <computed>

+ null_resource.test[1]
      id: <computed>

~ aws_security_group_rule.allow_all
      description: "" => "test3"

- aws_security_group_rule.allow_all
`, actOutput)
}

// Test that even if there's an error, we get the returned output.
func TestRun_OutputOnErr(t *testing.T) {
	RegisterMockTestingT(t)
	terraform := mocks.NewMockClient()
	tfVersion, _ := version.NewVersion("0.10.0")
	s := runtime.PlanStepRunner{
		TerraformExecutor: terraform,
		DefaultTFVersion:  tfVersion,
	}
	expOutput := "expected output"
	expErrMsg := "error!"
	When(terraform.RunCommandWithVersion(
		matchers.AnyModelsProjectCommandContext(),
		AnyString(),
		AnyStringSlice(),
		matchers2.AnyMapOfStringToString(),
		matchers2.AnyPtrToGoVersionVersion(),
		AnyString())).
		Then(func(params []Param) ReturnValues {
			// This code allows us to return different values depending on the
			// tf command being run while still using the wildcard matchers above.
			tfArgs := params[2].([]string)
			if stringSliceEquals(tfArgs, []string{"workspace", "show"}) {
				return []ReturnValue{"default\n", nil}
			} else if tfArgs[0] == "plan" {
				return []ReturnValue{expOutput, errors.New(expErrMsg)}
			} else {
				return []ReturnValue{"", errors.New("unexpected call to RunCommandWithVersion")}
			}
		})
	actOutput, actErr := s.Run(models.ProjectCommandContext{Workspace: "default"}, nil, "", map[string]string(nil))
	ErrEquals(t, expErrMsg, actErr)
	Equals(t, expOutput, actOutput)
}

// Test that if we're using 0.12, we don't set the optional -var atlantis_repo_name
// flags because in >= 0.12 you can't set -var flags if those variables aren't
// being used.
func TestRun_NoOptionalVarsIn012(t *testing.T) {
	RegisterMockTestingT(t)

	expPlanArgs := []string{
		"plan",
		"-input=false",
		"-refresh",
		"-out",
		fmt.Sprintf("%q", "/path/default.tfplan"),
		"extra",
		"args",
		"comment",
		"args",
	}

	cases := []struct {
		name      string
		tfVersion string
	}{
		{
			"stable version",
			"0.12.0",
		},
		{
			"with prerelease",
			"0.14.0-rc1",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			terraform := mocks.NewMockClient()
			When(terraform.RunCommandWithVersion(
				matchers.AnyModelsProjectCommandContext(),
				AnyString(),
				AnyStringSlice(),
				matchers2.AnyMapOfStringToString(),
				matchers2.AnyPtrToGoVersionVersion(),
				AnyString())).ThenReturn("output", nil)

			tfVersion, _ := version.NewVersion(c.tfVersion)
			s := runtime.PlanStepRunner{
				TerraformExecutor: terraform,
				DefaultTFVersion:  tfVersion,
			}
			ctx := models.ProjectCommandContext{
				Workspace:          "default",
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

			output, err := s.Run(ctx, []string{"extra", "args"}, "/path", map[string]string(nil))
			Ok(t, err)
			Equals(t, "output", output)

			terraform.VerifyWasCalledOnce().RunCommandWithVersion(ctx, "/path", expPlanArgs, map[string]string(nil), tfVersion, "default")
		})
	}

}

// Test plans if using remote ops.
func TestRun_RemoteOps(t *testing.T) {
	cases := map[string]string{
		"0.11.15 error": `Error: Saving a generated plan is currently not supported!

The "remote" backend does not support saving the generated execution
plan locally at this time.

`,
		"0.12.* error": `Error: Saving a generated plan is currently not supported

The "remote" backend does not support saving the generated execution plan
locally at this time.

`,
	}
	for name, remoteOpsErr := range cases {
		t.Run(name, func(t *testing.T) {

			logger := logging.NewNoopLogger(t)
			// Now that mocking is set up, we're ready to run the plan.
			ctx := models.ProjectCommandContext{
				Log:                logger,
				Workspace:          "default",
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
			RegisterMockTestingT(t)
			terraform := mocks.NewMockClient()

			tfVersion, _ := version.NewVersion("0.11.12")
			updater := mocks2.NewMockCommitStatusUpdater()
			asyncTf := &remotePlanMock{}
			s := runtime.PlanStepRunner{
				TerraformExecutor:   terraform,
				DefaultTFVersion:    tfVersion,
				AsyncTFExec:         asyncTf,
				CommitStatusUpdater: updater,
			}
			absProjectPath, cleanup := TempDir(t)
			defer cleanup()

			// First, terraform workspace gets run.
			When(terraform.RunCommandWithVersion(
				ctx,
				absProjectPath,
				[]string{"workspace", "show"},
				map[string]string(nil),
				tfVersion,
				"default")).ThenReturn("default\n", nil)

			// Then the first call to terraform plan should return the remote ops error.
			expPlanArgs := []string{"plan",
				"-input=false",
				"-refresh",
				"-out",
				fmt.Sprintf("%q", filepath.Join(absProjectPath, "default.tfplan")),
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

			planErr := errors.New("exit status 1: err")
			planOutput := "\n" + remoteOpsErr
			asyncTf.LinesToSend = remotePlanOutput
			When(terraform.RunCommandWithVersion(ctx, absProjectPath, expPlanArgs, map[string]string(nil), tfVersion, "default")).
				ThenReturn(planOutput, planErr)

			output, err := s.Run(ctx, []string{"extra", "args"}, absProjectPath, map[string]string(nil))
			Ok(t, err)
			Equals(t, `
An execution plan has been generated and is shown below.
Resource actions are indicated with the following symbols:
- destroy

Terraform will perform the following actions:

- null_resource.hi[1]


Plan: 0 to add, 0 to change, 1 to destroy.`, output)

			expRemotePlanArgs := []string{"plan", "-input=false", "-refresh", "extra", "args", "comment", "args"}
			Equals(t, expRemotePlanArgs, asyncTf.CalledArgs)

			// Verify that the fake plan file we write has the correct contents.
			bytes, err := os.ReadFile(filepath.Join(absProjectPath, "default.tfplan"))
			Ok(t, err)
			Equals(t, `Atlantis: this plan was created by remote ops

An execution plan has been generated and is shown below.
Resource actions are indicated with the following symbols:
  - destroy

Terraform will perform the following actions:

  - null_resource.hi[1]


Plan: 0 to add, 0 to change, 1 to destroy.`, string(bytes))

			// Ensure that the status was updated with the runURL.
			runURL := "https://app.terraform.io/app/lkysow-enterprises/atlantis-tfe-test/runs/run-is4oVvJfrkud1KvE"
			updater.VerifyWasCalledOnce().UpdateProject(ctx, models.PlanCommand, models.PendingCommitStatus, runURL)
			updater.VerifyWasCalledOnce().UpdateProject(ctx, models.PlanCommand, models.SuccessCommitStatus, runURL)
		})
	}
}

// Test striping output method
func TestStripRefreshingFromPlanOutput(t *testing.T) {
	tfVersion0135, _ := version.NewVersion("0.13.5")
	tfVersion0140, _ := version.NewVersion("0.14.0")
	cases := []struct {
		out       string
		tfVersion *version.Version
	}{
		{
			remotePlanOutput,
			tfVersion0135,
		},
		{
			`Running plan in the remote backend. Output will stream here. Pressing Ctrl-C
will stop streaming the logs, but will not stop the plan running remotely.

Preparing the remote plan...

To view this run in a browser, visit:
https://app.terraform.io/app/lkysow-enterprises/atlantis-tfe-test/runs/run-is4oVvJfrkud1KvE

Waiting for the plan to start...

Terraform v0.14.0

Configuring remote state backend...
Initializing Terraform configuration...
2019/02/20 22:40:52 [DEBUG] Using modified User-Agent: Terraform/0.14.0TFE/202eeff
Refreshing Terraform state in-memory prior to plan...
The refreshed state will be used to calculate this plan, but will not be
persisted to local or remote state storage.

null_resource.hi: Refreshing state... (ID: 217661332516885645)
null_resource.hi[1]: Refreshing state... (ID: 6064510335076839362)

An execution plan has been generated and is shown below.
Resource actions are indicated with the following symbols:
  - destroy

Terraform will perform the following actions:

  - null_resource.hi[1]


Plan: 0 to add, 0 to change, 1 to destroy.`,
			tfVersion0140,
		},
	}

	for _, c := range cases {
		output := runtime.StripRefreshingFromPlanOutput(c.out, c.tfVersion)
		Equals(t, `
An execution plan has been generated and is shown below.
Resource actions are indicated with the following symbols:
  - destroy

Terraform will perform the following actions:

  - null_resource.hi[1]


Plan: 0 to add, 0 to change, 1 to destroy.`, output)
	}
}

type remotePlanMock struct {
	// LinesToSend will be sent on the channel.
	LinesToSend string
	// CalledArgs is what args we were called with.
	CalledArgs []string
}

func (r *remotePlanMock) RunCommandAsync(ctx models.ProjectCommandContext, path string, args []string, envs map[string]string, v *version.Version, workspace string) (chan<- string, <-chan terraform.Line) {
	r.CalledArgs = args
	in := make(chan string)
	out := make(chan terraform.Line)
	go func() {
		for _, line := range strings.Split(r.LinesToSend, "\n") {
			out <- terraform.Line{Line: line}
		}
		close(out)
		close(in)
	}()
	return in, out
}

func stringSliceEquals(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}

var remotePlanOutput = `Running plan in the remote backend. Output will stream here. Pressing Ctrl-C
will stop streaming the logs, but will not stop the plan running remotely.

Preparing the remote plan...

To view this run in a browser, visit:
https://app.terraform.io/app/lkysow-enterprises/atlantis-tfe-test/runs/run-is4oVvJfrkud1KvE

Waiting for the plan to start...

Terraform v0.11.11

Configuring remote state backend...
Initializing Terraform configuration...
2019/02/20 22:40:52 [DEBUG] Using modified User-Agent: Terraform/0.11.11 TFE/202eeff
Refreshing Terraform state in-memory prior to plan...
The refreshed state will be used to calculate this plan, but will not be
persisted to local or remote state storage.

null_resource.hi: Refreshing state... (ID: 217661332516885645)
null_resource.hi[1]: Refreshing state... (ID: 6064510335076839362)

------------------------------------------------------------------------

An execution plan has been generated and is shown below.
Resource actions are indicated with the following symbols:
  - destroy

Terraform will perform the following actions:

  - null_resource.hi[1]


Plan: 0 to add, 0 to change, 1 to destroy.`
