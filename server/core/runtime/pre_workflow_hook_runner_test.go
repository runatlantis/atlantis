package runtime_test

import (
	"strings"
	"testing"

	"github.com/hashicorp/go-version"
	. "github.com/petergtz/pegomock/v4"
	"github.com/runatlantis/atlantis/server/core/runtime"
	"github.com/runatlantis/atlantis/server/core/terraform/mocks"
	"github.com/runatlantis/atlantis/server/events/models"
	jobmocks "github.com/runatlantis/atlantis/server/jobs/mocks"
	"github.com/runatlantis/atlantis/server/logging"
	. "github.com/runatlantis/atlantis/testing"
)

func TestPreWorkflowHookRunner_Run(t *testing.T) {

	defaultShell := "sh"
	defaultShellArgs := "-c"

	cases := []struct {
		Command        string
		Shell          string
		ShellArgs      string
		ExpOut         string
		ExpErr         string
		ExpDescription string
	}{
		{
			Command:        "",
			Shell:          defaultShell,
			ShellArgs:      defaultShellArgs,
			ExpOut:         "",
			ExpErr:         "",
			ExpDescription: "",
		},
		{
			Command:        "echo hi",
			Shell:          defaultShell,
			ShellArgs:      defaultShellArgs,
			ExpOut:         "hi\r\n",
			ExpErr:         "",
			ExpDescription: "",
		},
		{
			Command:        `printf \'your main.tf file does not provide default region.\\ncheck\'`,
			Shell:          defaultShell,
			ShellArgs:      defaultShellArgs,
			ExpOut:         `'your`,
			ExpErr:         "",
			ExpDescription: "",
		},
		{
			Command:        `printf 'your main.tf file does not provide default region.\ncheck'`,
			Shell:          defaultShell,
			ShellArgs:      defaultShellArgs,
			ExpOut:         "your main.tf file does not provide default region.\r\ncheck",
			ExpErr:         "",
			ExpDescription: "",
		},
		{
			Command:        "echo 'a",
			Shell:          defaultShell,
			ShellArgs:      defaultShellArgs,
			ExpOut:         "sh: 1: Syntax error: Unterminated quoted string\r\n",
			ExpErr:         "exit status 2: running \"sh -c echo 'a\" in",
			ExpDescription: "",
		},
		{
			Command:        "echo hi >> file && cat file",
			Shell:          defaultShell,
			ShellArgs:      defaultShellArgs,
			ExpOut:         "hi\r\n",
			ExpErr:         "",
			ExpDescription: "",
		},
		{
			Command:        "lkjlkj",
			Shell:          defaultShell,
			ShellArgs:      defaultShellArgs,
			ExpOut:         "sh: 1: lkjlkj: not found\r\n",
			ExpErr:         "exit status 127: running \"sh -c lkjlkj\" in",
			ExpDescription: "",
		},
		{
			Command:        "echo base_repo_name=$BASE_REPO_NAME base_repo_owner=$BASE_REPO_OWNER head_repo_name=$HEAD_REPO_NAME head_repo_owner=$HEAD_REPO_OWNER head_branch_name=$HEAD_BRANCH_NAME head_commit=$HEAD_COMMIT base_branch_name=$BASE_BRANCH_NAME pull_num=$PULL_NUM pull_url=$PULL_URL pull_author=$PULL_AUTHOR",
			Shell:          defaultShell,
			ShellArgs:      defaultShellArgs,
			ExpOut:         "base_repo_name=basename base_repo_owner=baseowner head_repo_name=headname head_repo_owner=headowner head_branch_name=add-feat head_commit=12345abcdef base_branch_name=main pull_num=2 pull_url=https://github.com/runatlantis/atlantis/pull/2 pull_author=acme\r\n",
			ExpErr:         "",
			ExpDescription: "",
		},
		{
			Command:        "echo user_name=$USER_NAME",
			Shell:          defaultShell,
			ShellArgs:      defaultShellArgs,
			ExpOut:         "user_name=acme-user\r\n",
			ExpErr:         "",
			ExpDescription: "",
		},
		{
			Command:        "echo something > $OUTPUT_STATUS_FILE",
			Shell:          defaultShell,
			ShellArgs:      defaultShellArgs,
			ExpOut:         "",
			ExpErr:         "",
			ExpDescription: "something",
		},
		{
			Command:        "echo shell test 1",
			Shell:          "bash",
			ShellArgs:      defaultShellArgs,
			ExpOut:         "shell test 1\r\n",
			ExpErr:         "",
			ExpDescription: "",
		},
		{
			Command:        "echo shell test 2",
			Shell:          defaultShell,
			ShellArgs:      "-cx",
			ExpOut:         "+ echo shell test 2\r\nshell test 2\r\n",
			ExpErr:         "",
			ExpDescription: "",
		},
		{
			Command:        "echo shell test 3",
			Shell:          "bash",
			ShellArgs:      "-cv",
			ExpOut:         "echo shell test 3\r\nshell test 3\r\n",
			ExpErr:         "",
			ExpDescription: "",
		},
	}

	for _, c := range cases {
		var err error

		Ok(t, err)

		RegisterMockTestingT(t)
		terraform := mocks.NewMockClient()
		When(terraform.EnsureVersion(Any[logging.SimpleLogging](), Any[*version.Version]())).
			ThenReturn(nil)

		logger := logging.NewNoopLogger(t)
		tmpDir := t.TempDir()

		projectCmdOutputHandler := jobmocks.NewMockProjectCommandOutputHandler()
		r := runtime.DefaultPreWorkflowHookRunner{
			OutputHandler: projectCmdOutputHandler,
		}
		t.Run(c.Command, func(t *testing.T) {
			ctx := models.WorkflowHookCommandContext{
				BaseRepo: models.Repo{
					Name:  "basename",
					Owner: "baseowner",
				},
				HeadRepo: models.Repo{
					Name:  "headname",
					Owner: "headowner",
				},
				Pull: models.PullRequest{
					Num:        2,
					URL:        "https://github.com/runatlantis/atlantis/pull/2",
					HeadBranch: "add-feat",
					HeadCommit: "12345abcdef",
					BaseBranch: "main",
					Author:     "acme",
				},
				User: models.User{
					Username: "acme-user",
				},
				Log:         logger,
				CommandName: "plan",
			}
			_, desc, err := r.Run(ctx, c.Command, c.Shell, c.ShellArgs, tmpDir)
			if c.ExpErr != "" {
				ErrContains(t, c.ExpErr, err)
			} else {
				Ok(t, err)
			}
			// Replace $DIR in the exp with the actual temp dir. We do this
			// here because when constructing the cases we don't yet know the
			// temp dir.
			Equals(t, c.ExpDescription, desc)
			expOut := strings.Replace(c.ExpOut, "$DIR", tmpDir, -1)
			projectCmdOutputHandler.VerifyWasCalledOnce().SendWorkflowHook(
				Any[models.WorkflowHookCommandContext](), Eq(expOut), Eq(false))
		})
	}
}
