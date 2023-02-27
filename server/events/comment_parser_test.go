// Copyright 2017 HootSuite Media Inc.
//
// Licensed under the Apache License, Version 2.0 (the License);
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//    http://www.apache.org/licenses/LICENSE-2.0
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an AS IS BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
// Modified hereafter by contributors to runatlantis/atlantis.

package events_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/runatlantis/atlantis/server/events"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/models"
	. "github.com/runatlantis/atlantis/testing"
	"github.com/stretchr/testify/assert"
)

var commentParser = events.CommentParser{
	GithubUser:     "github-user",
	GitlabUser:     "gitlab-user",
	ExecutableName: "atlantis",
	AllowCommands:  command.AllCommentCommands,
}

func TestNewCommentParser(t *testing.T) {
	type args struct {
		githubUser      string
		gitlabUser      string
		bitbucketUser   string
		azureDevopsUser string
		executableName  string
		allowCommands   []command.Name
	}
	tests := []struct {
		name string
		args args
		want *events.CommentParser
	}{
		{
			name: "duplicate allow commands filtered",
			args: args{
				allowCommands: []command.Name{command.Plan, command.Plan, command.Plan},
			},
			want: &events.CommentParser{
				AllowCommands: []command.Name{command.Plan},
			},
		},
		{
			name: "comment un-available commands filtered",
			args: args{
				// PolicyCheck and Autoplan cannot be used on comment command, so filtered
				allowCommands: []command.Name{command.Plan, command.Apply, command.Unlock, command.PolicyCheck, command.ApprovePolicies, command.Autoplan, command.Version, command.Import},
			},
			want: &events.CommentParser{
				AllowCommands: []command.Name{command.Version, command.Plan, command.Apply, command.Unlock, command.ApprovePolicies, command.Import},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, events.NewCommentParser(tt.args.githubUser, tt.args.gitlabUser, tt.args.bitbucketUser, tt.args.azureDevopsUser, tt.args.executableName, tt.args.allowCommands), "NewCommentParser(%v, %v, %v, %v, %v, %v)", tt.args.githubUser, tt.args.gitlabUser, tt.args.bitbucketUser, tt.args.azureDevopsUser, tt.args.executableName, tt.args.allowCommands)
		})
	}
}

func TestParse_Ignored(t *testing.T) {
	ignoreComments := []string{
		"",
		"a",
		"abc",
		"atlantis plan\nbut with newlines",
		"terraform plan\nbut with newlines",
		"This shouldn't error, but it does.",
	}
	for _, c := range ignoreComments {
		r := commentParser.Parse(c, models.Github)
		Assert(t, r.Ignore, "expected Ignore to be true for comment %q", c)
	}
}

func TestParse_ExecutableName(t *testing.T) {
	cases := []struct {
		user      string
		expIgnore bool
	}{
		{"custom-executable-name", false},
		{"run", false},
		{"@github-user", false},
		{"github-user", true},
		{"atlantis", true},
	}
	for _, c := range cases {
		t.Run(c.user, func(t *testing.T) {
			var commentParser = events.CommentParser{
				GithubUser:     "github-user",
				ExecutableName: "custom-executable-name",
			}
			comment := fmt.Sprintf("%s help", c.user)
			r := commentParser.Parse(comment, models.Github)
			Assert(t, r.Ignore == c.expIgnore, "expected Ignore %q, but got %q", c.expIgnore, r.Ignore)
		})
	}
}

func TestParse_HelpResponse(t *testing.T) {
	allowCommandsCases := [][]command.Name{
		command.AllCommentCommands,
		{}, // empty case
	}
	helpComments := []string{
		"run",
		"atlantis",
		"@github-user",
		"atlantis help",
		"atlantis --help",
		"atlantis -h",
		"atlantis help something else",
		"atlantis help plan",
	}
	for _, allowCommandCase := range allowCommandsCases {
		for _, c := range helpComments {
			t.Run(fmt.Sprintf("%s with allow commands %v", c, allowCommandCase), func(t *testing.T) {
				commentParser := events.CommentParser{
					GithubUser:     "github-user",
					ExecutableName: "atlantis",
					AllowCommands:  allowCommandCase,
				}
				r := commentParser.Parse(c, models.Github)
				Equals(t, commentParser.HelpComment(), r.CommentResponse)
			})
		}
	}
}

func TestParse_UnusedArguments(t *testing.T) {
	t.Log("if there are unused flags we return an error")
	cases := []struct {
		Command command.Name
		Args    string
		Unused  string
	}{
		{
			command.Plan,
			"-d . arg",
			"arg",
		},
		{
			command.Plan,
			"arg -d .",
			"arg",
		},
		{
			command.Plan,
			"arg",
			"arg",
		},
		{
			command.Plan,
			"arg arg2",
			"arg arg2",
		},
		{
			command.Plan,
			"-d . arg -w kjj arg2",
			"arg arg2",
		},
		{
			command.Apply,
			"-d . arg",
			"arg",
		},
		{
			command.Apply,
			"arg arg2",
			"arg arg2",
		},
		{
			command.Apply,
			"arg arg2 -- useful",
			"arg arg2",
		},
		{
			command.Apply,
			"arg arg2 --",
			"arg arg2",
		},
		{
			command.ApprovePolicies,
			"arg arg2 --",
			"arg arg2",
		},
		{
			command.Import,
			"arg --",
			"arg",
		},
		{
			command.Import,
			"arg1 arg2 arg3 --",
			"arg1 arg2 arg3",
		},
	}
	for _, c := range cases {
		comment := fmt.Sprintf("atlantis %s %s", c.Command.String(), c.Args)
		t.Run(comment, func(t *testing.T) {
			r := commentParser.Parse(comment, models.Github)
			var usage string
			switch c.Command {
			case command.Plan:
				usage = PlanUsage
			case command.Apply:
				usage = ApplyUsage
			case command.ApprovePolicies:
				usage = ApprovePolicyUsage
			case command.Import:
				usage = ImportUsage
			}
			Equals(t, fmt.Sprintf("```\nError: unknown argument(s) – %s.\n%s```", c.Unused, usage), r.CommentResponse)
		})
	}
}

func TestParse_UnknownShorthandFlag(t *testing.T) {
	comment := "atlantis unlock -d ."
	r := commentParser.Parse(comment, models.Github)

	Equals(t, UnlockUsage, r.CommentResponse)
}

func TestParse_DidYouMeanAtlantis(t *testing.T) {
	t.Log("given a comment that should result in a 'did you mean atlantis'" +
		"response, should set CommentParseResult.CommentResult")
	comments := []string{
		"terraform",
		"terraform help",
		"terraform --help",
		"terraform -h",
		"terraform plan",
		"terraform apply",
		"terraform plan -w workspace -d . -- test",
	}
	for _, c := range comments {
		r := commentParser.Parse(c, models.Github)
		Assert(t, r.CommentResponse == fmt.Sprintf(events.DidYouMeanAtlantisComment, "atlantis", "terraform"),
			"For comment %q expected CommentResponse==%q but got %q", c, events.DidYouMeanAtlantisComment, r.CommentResponse)
	}
}

func TestParse_InvalidCommand(t *testing.T) {
	t.Log("given a comment with an invalid atlantis command, should return " +
		"a warning.")
	comments := []string{
		"atlantis paln",
		"atlantis Plan",
		"atlantis appely apply",
	}
	cp := events.NewCommentParser(
		"github-user",
		"gitlab-user",
		"bitbucket-user",
		"azure-devops-user",
		"atlantis",
		[]command.Name{
			command.Version,
			command.Unlock,
			command.Apply,
			command.Plan,
			command.Apply, // duplicate command is filtered
		},
	)
	for _, c := range comments {
		r := cp.Parse(c, models.Github)
		exp := fmt.Sprintf("```\nError: unknown command %q.\nRun 'atlantis --help' for usage.\nAvailable commands(--allow-commands): version, plan, apply, unlock\n```", strings.Fields(c)[1])
		Equals(t, exp, r.CommentResponse)
	}
}

func TestParse_SubcommandUsage(t *testing.T) {
	t.Log("given a comment asking for the usage of a subcommand should " +
		"return help")
	tests := []struct {
		input    string
		expUsage string
	}{
		{"atlantis plan -h", "plan"},
		{"atlantis plan --help", "plan"},
		{"atlantis apply -h", "apply"},
		{"atlantis apply --help", "apply"},
		{"atlantis approve_policies -h", "approve_policies"},
		{"atlantis approve_policies --help", "approve_policies"},
		{"atlantis import -h", "import ADDRESS ID"},
		{"atlantis import --help", "import ADDRESS ID"},
		{"atlantis state -h", "state [rm ADDRESS...]"},
		{"atlantis state --help", "state [rm ADDRESS...]"},
	}
	for _, c := range tests {
		r := commentParser.Parse(c.input, models.Github)
		exp := "Usage of " + c.expUsage
		Assert(t, strings.Contains(r.CommentResponse, exp),
			"For comment %q expected CommentResponse %q to contain %q", c, r.CommentResponse, exp)
		Assert(t, !strings.Contains(r.CommentResponse, "Error:"),
			"For comment %q expected CommentResponse %q to not contain %q", c, r.CommentResponse, "Error: ")
	}
}

func TestParse_InvalidFlags(t *testing.T) {
	t.Log("given a comment with a valid atlantis command but invalid" +
		" flags, should return a warning and the proper usage")
	cases := []struct {
		comment string
		exp     string
	}{
		{
			"atlantis plan -e",
			"Error: unknown shorthand flag: 'e' in -e",
		},
		{
			"atlantis plan --abc",
			"Error: unknown flag: --abc",
		},
		{
			"atlantis apply -e",
			"Error: unknown shorthand flag: 'e' in -e",
		},
		{
			"atlantis apply --abc",
			"Error: unknown flag: --abc",
		},
		{
			"atlantis import --abc",
			"Error: unknown flag: --abc",
		},
		{
			"atlantis state rm --abc",
			"Error: unknown flag: --abc",
		},
	}
	for _, c := range cases {
		r := commentParser.Parse(c.comment, models.Github)
		Assert(t, strings.Contains(r.CommentResponse, c.exp),
			"For comment %q expected CommentResponse %q to contain %q", c.comment, r.CommentResponse, c.exp)
		Assert(t, strings.Contains(r.CommentResponse, "Usage of "),
			"For comment %q expected CommentResponse %q to contain %q", c.comment, r.CommentResponse, "Usage of ")
	}
}

func TestParse_RelativeDirPath(t *testing.T) {
	t.Log("if -d is used with a relative path, should return an error")
	comments := []string{
		"atlantis plan -d ..",
		"atlantis apply -d ..",
		"atlantis import -d .. address id",
		"atlantis state -d .. rm address",
		// These won't return an error because we prepend with . when parsing.
		//"atlantis plan -d /..",
		//"atlantis apply -d /..",
		//"atlantis import -d /.. address id",
		//"atlantis state rm -d /.. address",
		"atlantis plan -d ./..",
		"atlantis apply -d ./..",
		"atlantis import -d ./.. address id",
		"atlantis state -d ./.. rm address",
		"atlantis plan -d a/b/../../..",
		"atlantis apply -d a/../..",
		"atlantis import -d a/../.. address id",
		"atlantis state -d a/../.. rm address id",
	}
	for _, c := range comments {
		r := commentParser.Parse(c, models.Github)
		exp := "Error: using a relative path"
		Assert(t, strings.Contains(r.CommentResponse, exp),
			"For comment %q expected CommentResponse %q to contain %q", c, r.CommentResponse, exp)
	}
}

// If there's multiple lines but it's whitespace, allow the command. This
// occurs when you copy and paste via GitHub.
func TestParse_Multiline(t *testing.T) {
	comments := []string{
		"atlantis plan\n",
		"atlantis plan\n\n",
		"atlantis plan\r\n",
		"atlantis plan\r\n\r\n",
		"\natlantis plan",
		"\r\natlantis plan",
		"\natlantis plan\n",
		"\r\natlantis plan\r\n",
	}
	for _, comment := range comments {
		t.Run(comment, func(t *testing.T) {
			r := commentParser.Parse(comment, models.Github)
			Equals(t, "", r.CommentResponse)
			Equals(t, &events.CommentCommand{
				RepoRelDir:  "",
				Flags:       nil,
				Name:        command.Plan,
				Verbose:     false,
				Workspace:   "",
				ProjectName: "",
			}, r.Command)
		})
	}
}

func TestParse_InvalidWorkspace(t *testing.T) {
	t.Log("if -w is used with '..' or '/', should return an error")
	comments := []string{
		"atlantis plan -w ..",
		"atlantis apply -w ..",
		"atlantis import -w .. address id",
		"atlantis import -w .. rm address",
		"atlantis plan -w /",
		"atlantis apply -w /",
		"atlantis import -w / address id",
		"atlantis state -w / rm address",
		"atlantis plan -w ..abc",
		"atlantis apply -w abc..",
		"atlantis import -w abc.. address id",
		"atlantis state -w abc.. rm address",
		"atlantis plan -w abc..abc",
		"atlantis apply -w ../../../etc/passwd",
		"atlantis import -w ../../../etc/passwd address id",
		"atlantis state -w ../../../etc/passwd rm address",
	}
	for _, c := range comments {
		r := commentParser.Parse(c, models.Github)
		exp := "Error: invalid workspace"
		Assert(t, strings.Contains(r.CommentResponse, exp),
			"For comment %q expected CommentResponse %q to contain %q", c, r.CommentResponse, exp)
	}
}

func TestParse_UsingProjectAtSameTimeAsWorkspaceOrDir(t *testing.T) {
	cases := []string{
		"atlantis plan -w workspace -p project",
		"atlantis plan -d dir -p project",
		"atlantis plan -d dir -w workspace -p project",
	}
	for _, c := range cases {
		t.Run(c, func(t *testing.T) {
			r := commentParser.Parse(c, models.Github)
			exp := "Error: cannot use -p/--project at same time as -d/--dir or -w/--workspace"
			Assert(t, strings.Contains(r.CommentResponse, exp),
				"For comment %q expected CommentResponse %q to contain %q", c, r.CommentResponse, exp)
		})
	}
}

func TestParse_Parsing(t *testing.T) {
	cases := []struct {
		flags        string
		expWorkspace string
		expDir       string
		expVerbose   bool
		expExtraArgs string
		expProject   string
	}{
		// Test defaults.
		{
			"",
			"",
			"",
			false,
			"",
			"",
		},
		// Test each short flag individually.
		{
			"-w workspace",
			"workspace",
			"",
			false,
			"",
			"",
		},
		{
			"-d dir",
			"",
			"dir",
			false,
			"",
			"",
		},
		{
			"-p project",
			"",
			"",
			false,
			"",
			"project",
		},
		{
			"--verbose",
			"",
			"",
			true,
			"",
			"",
		},
		// Test each long flag individually.
		{
			"--workspace workspace",
			"workspace",
			"",
			false,
			"",
			"",
		},
		{
			"--dir dir",
			"",
			"dir",
			false,
			"",
			"",
		},
		{
			"--project project",
			"",
			"",
			false,
			"",
			"project",
		},
		// Test all of them with different permutations.
		{
			"-w workspace -d dir --verbose",
			"workspace",
			"dir",
			true,
			"",
			"",
		},
		{
			"-d dir -w workspace --verbose",
			"workspace",
			"dir",
			true,
			"",
			"",
		},
		{
			"--verbose -w workspace -d dir",
			"workspace",
			"dir",
			true,
			"",
			"",
		},
		{
			"-p project --verbose",
			"",
			"",
			true,
			"",
			"project",
		},
		{
			"--verbose -p project",
			"",
			"",
			true,
			"",
			"project",
		},
		// Test that flags after -- are ignored
		{
			"-w workspace -d dir -- --verbose",
			"workspace",
			"dir",
			false,
			"--verbose",
			"",
		},
		{
			"-w workspace -- -d dir --verbose",
			"workspace",
			"",
			false,
			"-d dir --verbose",
			"",
		},
		// Test the extra args parsing.
		{
			"--",
			"",
			"",
			false,
			"",
			"",
		},
		{
			"-w workspace -d dir --verbose -- arg one -two --three &&",
			"workspace",
			"dir",
			true,
			"arg one -two --three &&",
			"",
		},
		// Test whitespace.
		{
			"\t-w\tworkspace\t-d\tdir\t--verbose\t--\targ\tone\t-two\t--three\t&&",
			"workspace",
			"dir",
			true,
			"arg one -two --three &&",
			"",
		},
		{
			"   -w   workspace   -d   dir   --verbose   --   arg   one   -two   --three   &&",
			"workspace",
			"dir",
			true,
			"arg one -two --three &&",
			"",
		},
		// Test that the dir string is normalized.
		{
			"-d /",
			"",
			".",
			false,
			"",
			"",
		},
		{
			"-d /adir",
			"",
			"adir",
			false,
			"",
			"",
		},
		{
			"-d .",
			"",
			".",
			false,
			"",
			"",
		},
		{
			"-d ./",
			"",
			".",
			false,
			"",
			"",
		},
		{
			"-d ./adir",
			"",
			"adir",
			false,
			"",
			"",
		},
		{
			"-d \"dir with space\"",
			"",
			"dir with space",
			false,
			"",
			"",
		},
	}

	for _, test := range cases {
		for _, cmdName := range []string{"plan", "apply", "import 'some[\"addr\"]' id", "state rm 'some[\"addr\"]'"} {
			comment := fmt.Sprintf("atlantis %s %s", cmdName, test.flags)
			t.Run(comment, func(t *testing.T) {
				r := commentParser.Parse(comment, models.Github)
				Assert(t, r.CommentResponse == "", "CommentResponse should have been empty but was %q for comment %q", r.CommentResponse, comment)
				Assert(t, test.expDir == r.Command.RepoRelDir, "exp dir to equal %q but was %q for comment %q", test.expDir, r.Command.RepoRelDir, comment)
				Assert(t, test.expWorkspace == r.Command.Workspace, "exp workspace to equal %q but was %q for comment %q", test.expWorkspace, r.Command.Workspace, comment)
				Assert(t, test.expVerbose == r.Command.Verbose, "exp verbose to equal %v but was %v for comment %q", test.expVerbose, r.Command.Verbose, comment)
				actExtraArgs := strings.Join(r.Command.Flags, " ")
				if cmdName == "plan" {
					Assert(t, r.Command.Name == command.Plan, "did not parse comment %q as plan command", comment)
					Assert(t, test.expExtraArgs == actExtraArgs, "exp extra args to equal %v but got %v for comment %q", test.expExtraArgs, actExtraArgs, comment)
				}
				if cmdName == "apply" {
					Assert(t, r.Command.Name == command.Apply, "did not parse comment %q as apply command", comment)
					Assert(t, test.expExtraArgs == actExtraArgs, "exp extra args to equal %v but got %v for comment %q", test.expExtraArgs, actExtraArgs, comment)
				}
				if cmdName == "approve_policies" {
					Assert(t, r.Command.Name == command.ApprovePolicies, "did not parse comment %q as approve_policies command", comment)
					Assert(t, test.expExtraArgs == actExtraArgs, "exp extra args to equal %v but got %v for comment %q", test.expExtraArgs, actExtraArgs, comment)
				}
				if strings.HasPrefix(cmdName, "import") {
					expExtraArgs := "some[\"addr\"] id" // import use default args with `some["addr"] id`
					if test.expExtraArgs != "" {
						expExtraArgs = fmt.Sprintf("%s %s", test.expExtraArgs, expExtraArgs)
					}
					Assert(t, r.Command.Name == command.Import, "did not parse comment %q as import command", comment)
					Assert(t, expExtraArgs == actExtraArgs, "exp extra args to equal %v but got %v for comment %q", expExtraArgs, actExtraArgs, comment)
				}
				if strings.HasPrefix(cmdName, "state rm") {
					expExtraArgs := "some[\"addr\"]" // state rm use default args with `some["addr"]`
					if test.expExtraArgs != "" {
						expExtraArgs = fmt.Sprintf("%s %s", test.expExtraArgs, expExtraArgs)
					}
					Assert(t, r.Command.Name == command.State, "did not parse comment %q as state command", comment)
					Assert(t, r.Command.SubName == "rm", "did not parse comment %q as state rm subcommand", comment)
					Assert(t, expExtraArgs == actExtraArgs, "exp extra args to equal %v but got %v for comment %q", expExtraArgs, actExtraArgs, comment)
				}
			})
		}
	}
}

func TestBuildPlanApplyVersionComment(t *testing.T) {
	cases := []struct {
		repoRelDir        string
		workspace         string
		project           string
		autoMergeDisabled bool
		commentArgs       []string
		expPlanFlags      string
		expApplyFlags     string
		expVersionFlags   string
	}{
		{
			repoRelDir:      ".",
			workspace:       "default",
			project:         "",
			commentArgs:     nil,
			expPlanFlags:    "-d .",
			expApplyFlags:   "-d .",
			expVersionFlags: "-d .",
		},
		{
			repoRelDir:      "dir",
			workspace:       "default",
			project:         "",
			commentArgs:     nil,
			expPlanFlags:    "-d dir",
			expApplyFlags:   "-d dir",
			expVersionFlags: "-d dir",
		},
		{
			repoRelDir:      ".",
			workspace:       "workspace",
			project:         "",
			commentArgs:     nil,
			expPlanFlags:    "-w workspace",
			expApplyFlags:   "-w workspace",
			expVersionFlags: "-w workspace",
		},
		{
			repoRelDir:      "dir",
			workspace:       "workspace",
			project:         "",
			commentArgs:     nil,
			expPlanFlags:    "-d dir -w workspace",
			expApplyFlags:   "-d dir -w workspace",
			expVersionFlags: "-d dir -w workspace",
		},
		{
			repoRelDir:      ".",
			workspace:       "default",
			project:         "project",
			commentArgs:     nil,
			expPlanFlags:    "-p project",
			expApplyFlags:   "-p project",
			expVersionFlags: "-p project",
		},
		{
			repoRelDir:      "dir",
			workspace:       "workspace",
			project:         "project",
			commentArgs:     nil,
			expPlanFlags:    "-p project",
			expApplyFlags:   "-p project",
			expVersionFlags: "-p project",
		},
		{
			repoRelDir:      ".",
			workspace:       "default",
			project:         "",
			commentArgs:     []string{`"arg1"`, `"arg2"`},
			expPlanFlags:    "-d . -- arg1 arg2",
			expApplyFlags:   "-d .",
			expVersionFlags: "-d .",
		},
		{
			repoRelDir:      "dir",
			workspace:       "workspace",
			project:         "",
			commentArgs:     []string{`"arg1"`, `"arg2"`, `arg3`},
			expPlanFlags:    "-d dir -w workspace -- arg1 arg2 arg3",
			expApplyFlags:   "-d dir -w workspace",
			expVersionFlags: "-d dir -w workspace",
		},
		{
			repoRelDir:      "dir with spaces",
			workspace:       "default",
			project:         "",
			expPlanFlags:    "-d \"dir with spaces\"",
			expApplyFlags:   "-d \"dir with spaces\"",
			expVersionFlags: "-d \"dir with spaces\"",
		},
		{
			repoRelDir:        "dir",
			workspace:         "workspace",
			project:           "",
			autoMergeDisabled: true,
			commentArgs:       []string{`"arg1"`, `"arg2"`, `arg3`},
			expPlanFlags:      "-d dir -w workspace -- arg1 arg2 arg3",
			expApplyFlags:     "-d dir -w workspace --auto-merge-disabled",
			expVersionFlags:   "-d dir -w workspace",
		},
	}

	for _, c := range cases {
		t.Run(c.expPlanFlags, func(t *testing.T) {
			for _, cmd := range []command.Name{command.Plan, command.Apply, command.Version} {
				switch cmd {
				case command.Plan:
					actComment := commentParser.BuildPlanComment(c.repoRelDir, c.workspace, c.project, c.commentArgs)
					Equals(t, fmt.Sprintf("atlantis plan %s", c.expPlanFlags), actComment)
				case command.Apply:
					actComment := commentParser.BuildApplyComment(c.repoRelDir, c.workspace, c.project, c.autoMergeDisabled)
					Equals(t, fmt.Sprintf("atlantis apply %s", c.expApplyFlags), actComment)
				}
			}
		})
	}
}

func TestCommentParser_HelpComment(t *testing.T) {
	cases := []struct {
		name          string
		allowCommands []command.Name
		expectResult  string
	}{
		{
			name:          "all commands allowed",
			allowCommands: command.AllCommentCommands,
			expectResult: "```cmake\n" +
				`atlantis
Terraform Pull Request Automation

Usage:
  atlantis <command> [options] -- [terraform options]

Examples:
  # show atlantis help
  atlantis help

  # run plan in the root directory passing the -target flag to terraform
  atlantis plan -d . -- -target=resource

  # apply all unapplied plans from this pull request
  atlantis apply

  # apply the plan for the root directory and staging workspace
  atlantis apply -d . -w staging

Commands:
  plan     Runs 'terraform plan' for the changes in this pull request.
           To plan a specific project, use the -d, -w and -p flags.
  apply    Runs 'terraform apply' on all unapplied plans from this pull request.
           To only apply a specific plan, use the -d, -w and -p flags.
  unlock   Removes all atlantis locks and discards all plans for this PR.
           To unlock a specific plan you can use the Atlantis UI.
  approve_policies
           Approves all current policy checking failures for the PR.
  version  Print the output of 'terraform version'
  import ADDRESS ID
           Runs 'terraform import' for the passed address resource.
           To import a specific project, use the -d, -w and -p flags.
  state rm ADDRESS...
           Runs 'terraform state rm' for the passed address resource.
           To remove a specific project resource, use the -d, -w and -p flags.
  help     View help.

Flags:
  -h, --help   help for atlantis

Use "atlantis [command] --help" for more information about a command.` +
				"\n```",
		},
		{
			name:          "all commands disallowed",
			allowCommands: []command.Name{},
			expectResult: "```cmake\n" +
				`atlantis
Terraform Pull Request Automation

Usage:
  atlantis <command> [options] -- [terraform options]

Examples:
  # show atlantis help
  atlantis help

Commands:
  help     View help.

Flags:
  -h, --help   help for atlantis

Use "atlantis [command] --help" for more information about a command.` +
				"\n```",
		},
		{
			name: "partial commands allowed",
			allowCommands: []command.Name{
				command.Apply,
				command.Unlock,
			},
			expectResult: "```cmake\n" +
				`atlantis
Terraform Pull Request Automation

Usage:
  atlantis <command> [options] -- [terraform options]

Examples:
  # show atlantis help
  atlantis help

  # apply all unapplied plans from this pull request
  atlantis apply

  # apply the plan for the root directory and staging workspace
  atlantis apply -d . -w staging

Commands:
  apply    Runs 'terraform apply' on all unapplied plans from this pull request.
           To only apply a specific plan, use the -d, -w and -p flags.
  unlock   Removes all atlantis locks and discards all plans for this PR.
           To unlock a specific plan you can use the Atlantis UI.
  help     View help.

Flags:
  -h, --help   help for atlantis

Use "atlantis [command] --help" for more information about a command.` +
				"\n```",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			commentParser := events.CommentParser{
				ExecutableName: "atlantis",
				AllowCommands:  c.allowCommands,
			}
			Equals(t, commentParser.HelpComment(), c.expectResult)
		})
	}
}

func TestParse_VCSUsername(t *testing.T) {
	cp := events.CommentParser{
		GithubUser:      "gh",
		GitlabUser:      "gl",
		BitbucketUser:   "bb",
		AzureDevopsUser: "ad",
		ExecutableName:  "atlantis",
	}
	cases := []struct {
		vcs  models.VCSHostType
		user string
	}{
		{
			vcs:  models.Github,
			user: "gh",
		},
		{
			vcs:  models.Gitlab,
			user: "gl",
		},
		{
			vcs:  models.BitbucketServer,
			user: "bb",
		},
		{
			vcs:  models.BitbucketCloud,
			user: "bb",
		},
		{
			vcs:  models.AzureDevops,
			user: "ad",
		},
	}

	for _, c := range cases {
		t.Run(c.vcs.String(), func(t *testing.T) {
			r := cp.Parse(fmt.Sprintf("@%s %s", c.user, "help"), c.vcs)
			Equals(t, cp.HelpComment(), r.CommentResponse)
		})
	}
}

var PlanUsage = `Usage of plan:
  -d, --dir string         Which directory to run plan in relative to root of repo,
                           ex. 'child/dir'.
  -p, --project string     Which project to run plan for. Refers to the name of the
                           project configured in a repo config file. Cannot be used
                           at same time as workspace or dir flags.
      --verbose            Append Atlantis log to comment.
  -w, --workspace string   Switch to this Terraform workspace before planning.
`

var ApplyUsage = `Usage of apply:
      --auto-merge-disabled   Disable automerge after apply.
  -d, --dir string            Apply the plan for this directory, relative to root of
                              repo, ex. 'child/dir'.
  -p, --project string        Apply the plan for this project. Refers to the name of
                              the project configured in a repo config file. Cannot
                              be used at same time as workspace or dir flags.
      --verbose               Append Atlantis log to comment.
  -w, --workspace string      Apply the plan for this Terraform workspace.
`

var ApprovePolicyUsage = `Usage of approve_policies:
  -d, --dir string          Approve policies for this directory, relative to root of
                            repo, ex. 'child/dir'.
      --policy-set string   Approve policies for this project. Refers to the name of
                            the project configured in a repo config file. Cannot be
                            used at same time as workspace or dir flags.
  -p, --project string      Approve policies for this project. Refers to the name of
                            the project configured in a repo config file. Cannot be
                            used at same time as workspace or dir flags.
      --verbose             Append Atlantis log to comment.
  -w, --workspace string    Approve policies for this Terraform workspace.
`

var UnlockUsage = "`Usage of unlock:`\n\n ```cmake\n" +
	`atlantis unlock

  Unlocks the entire PR and discards all plans in this PR.
  Arguments or flags are not supported at the moment.
  If you need to unlock a specific project please use the atlantis UI.` +
	"\n```"

var ImportUsage = `Usage of import ADDRESS ID:
  -d, --dir string         Which directory to run import in relative to root of
                           repo, ex. 'child/dir'.
  -p, --project string     Which project to run import for. Refers to the name of
                           the project configured in a repo config file. Cannot be
                           used at same time as workspace or dir flags.
      --verbose            Append Atlantis log to comment.
  -w, --workspace string   Switch to this Terraform workspace before importing.
`
