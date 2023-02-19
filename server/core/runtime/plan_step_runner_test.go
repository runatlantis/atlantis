package runtime_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/hashicorp/go-version"
	. "github.com/petergtz/pegomock"
	"github.com/pkg/errors"
	"github.com/runatlantis/atlantis/server/core/runtime"
	runtimemocks "github.com/runatlantis/atlantis/server/core/runtime/mocks"
	runtimemodels "github.com/runatlantis/atlantis/server/core/runtime/models"
	"github.com/runatlantis/atlantis/server/core/terraform/mocks"
	matchers2 "github.com/runatlantis/atlantis/server/core/terraform/mocks/matchers"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/mocks/matchers"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/logging"

	. "github.com/runatlantis/atlantis/testing"
)

func TestRun_AddsEnvVarFile(t *testing.T) {
	// Test that if env/workspace.tfvars file exists we use -var-file option.
	RegisterMockTestingT(t)
	terraform := mocks.NewMockClient()
	commitStatusUpdater := runtimemocks.NewMockStatusUpdater()
	asyncTfExec := runtimemocks.NewMockAsyncTFExec()

	// Create the env/workspace.tfvars file.
	tmpDir := t.TempDir()
	err := os.MkdirAll(filepath.Join(tmpDir, "env"), 0700)
	Ok(t, err)
	envVarsFile := filepath.Join(tmpDir, "env/workspace.tfvars")
	err = os.WriteFile(envVarsFile, nil, 0600)
	Ok(t, err)

	// Using version >= 0.10 here so we don't expect any env commands.
	tfVersion, _ := version.NewVersion("0.10.0")
	logger := logging.NewNoopLogger(t)
	s := runtime.NewPlanStepRunner(terraform, tfVersion, commitStatusUpdater, asyncTfExec)

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
	commitStatusUpdater := runtimemocks.NewMockStatusUpdater()
	asyncTfExec := runtimemocks.NewMockAsyncTFExec()
	tfVersion, _ := version.NewVersion("0.10.0")
	logger := logging.NewNoopLogger(t)
	s := runtime.NewPlanStepRunner(terraform, tfVersion, commitStatusUpdater, asyncTfExec)
	ctx := command.ProjectContext{
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
	commitStatusUpdater := runtimemocks.NewMockStatusUpdater()
	asyncTfExec := runtimemocks.NewMockAsyncTFExec()
	tfVersion, _ := version.NewVersion("0.10.0")
	s := runtime.NewPlanStepRunner(terraform, tfVersion, commitStatusUpdater, asyncTfExec)
	When(terraform.RunCommandWithVersion(
		matchers.AnyCommandProjectContext(),
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
	actOutput, err := s.Run(command.ProjectContext{Workspace: "default"}, nil, "", map[string]string(nil))
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
	commitStatusUpdater := runtimemocks.NewMockStatusUpdater()
	asyncTfExec := runtimemocks.NewMockAsyncTFExec()
	tfVersion, _ := version.NewVersion("0.10.0")
	s := runtime.NewPlanStepRunner(terraform, tfVersion, commitStatusUpdater, asyncTfExec)
	expOutput := "expected output"
	expErrMsg := "error!"
	When(terraform.RunCommandWithVersion(
		matchers.AnyCommandProjectContext(),
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
	actOutput, actErr := s.Run(command.ProjectContext{Workspace: "default"}, nil, "", map[string]string(nil))
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
			commitStatusUpdater := runtimemocks.NewMockStatusUpdater()
			asyncTfExec := runtimemocks.NewMockAsyncTFExec()
			When(terraform.RunCommandWithVersion(
				matchers.AnyCommandProjectContext(),
				AnyString(),
				AnyStringSlice(),
				matchers2.AnyMapOfStringToString(),
				matchers2.AnyPtrToGoVersionVersion(),
				AnyString())).ThenReturn("output", nil)

			tfVersion, _ := version.NewVersion(c.tfVersion)
			s := runtime.NewPlanStepRunner(terraform, tfVersion, commitStatusUpdater, asyncTfExec)
			ctx := command.ProjectContext{
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
	cases := []struct {
		name         string
		tfVersion    string
		remoteOpsErr string
	}{
		{
			name:      "0.11.15 error",
			tfVersion: "0.11.15",
			remoteOpsErr: `Error: Saving a generated plan is currently not supported!

The "remote" backend does not support saving the generated execution
plan locally at this time.

`,
		},
		{
			name:      "0.12.* error",
			tfVersion: "0.12.0",
			remoteOpsErr: `Error: Saving a generated plan is currently not supported

The "remote" backend does not support saving the generated execution plan
locally at this time.

`,
		},
		{
			name:      "1.1.0 error",
			tfVersion: "1.1.0",
			remoteOpsErr: `╷
│ Error: Saving a generated plan is currently not supported
│ 
│ Terraform Cloud does not support saving the generated execution plan
│ locally at this time.
╵
`,
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {

			logger := logging.NewNoopLogger(t)
			// Now that mocking is set up, we're ready to run the plan.
			ctx := command.ProjectContext{
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
			commitStatusUpdater := runtimemocks.NewMockStatusUpdater()
			tfVersion, _ := version.NewVersion(c.tfVersion)
			asyncTf := &remotePlanMock{}
			s := runtime.NewPlanStepRunner(terraform, tfVersion, commitStatusUpdater, asyncTf)
			absProjectPath := t.TempDir()

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
			if tfVersion.GreaterThanOrEqual(version.Must(version.NewVersion("0.12.0"))) {
				expPlanArgs = []string{"plan",
					"-input=false",
					"-refresh",
					"-out",
					fmt.Sprintf("%q", filepath.Join(absProjectPath, "default.tfplan")),
					"extra",
					"args",
					"comment",
					"args",
				}
			}

			planErr := errors.New("exit status 1: err")
			planOutput := "\n" + c.remoteOpsErr
			asyncTf.LinesToSend = remotePlanOutput
			When(terraform.RunCommandWithVersion(ctx, absProjectPath, expPlanArgs, map[string]string(nil), tfVersion, "default")).
				ThenReturn(planOutput, planErr)

			output, err := s.Run(ctx, []string{"extra", "args"}, absProjectPath, map[string]string(nil))
			Ok(t, err)
			Assert(t, strings.Contains(output, `
An execution plan has been generated and is shown below.
Resource actions are indicated with the following symbols:
- destroy

Terraform will perform the following actions:

- null_resource.hi[1]


Plan: 0 to add, 0 to change, 1 to destroy.`), "expect plan success")

			expRemotePlanArgs := []string{"plan", "-input=false", "-refresh", "-no-color", "extra", "args", "comment", "args"}
			Equals(t, expRemotePlanArgs, asyncTf.CalledArgs)

			// Verify that the fake plan file we write has the correct contents.
			bytes, err := os.ReadFile(filepath.Join(absProjectPath, "default.tfplan"))
			Ok(t, err)
			Assert(t, strings.HasPrefix(string(bytes), "Atlantis: this plan was created by remote ops"), "expect remote plan")

			// Ensure that the status was updated with the runURL.
			runURL := "https://app.terraform.io/app/lkysow-enterprises/atlantis-tfe-test/runs/run-is4oVvJfrkud1KvE"
			commitStatusUpdater.VerifyWasCalledOnce().UpdateProject(ctx, command.Plan, models.PendingCommitStatus, runURL, nil)
			commitStatusUpdater.VerifyWasCalledOnce().UpdateProject(ctx, command.Plan, models.SuccessCommitStatus, runURL, nil)
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

func (r *remotePlanMock) RunCommandAsync(ctx command.ProjectContext, path string, args []string, envs map[string]string, v *version.Version, workspace string) (chan<- string, <-chan runtimemodels.Line) {
	r.CalledArgs = args
	in := make(chan string)
	out := make(chan runtimemodels.Line)
	go func() {
		for _, line := range strings.Split(r.LinesToSend, "\n") {
			out <- runtimemodels.Line{Line: line}
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
