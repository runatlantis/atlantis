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
//
package events

import (
	"fmt"
	"io/ioutil"
	"net/url"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/events/yaml"
	"github.com/spf13/pflag"
)

const (
	WorkspaceFlagLong  = "workspace"
	WorkspaceFlagShort = "w"
	DirFlagLong        = "dir"
	DirFlagShort       = "d"
	ProjectFlagLong    = "project"
	ProjectFlagShort   = "p"
	VerboseFlagLong    = "verbose"
	VerboseFlagShort   = ""
)

// multiLineRegex is used to ignore multi-line comments since those aren't valid
// Atlantis commands. If the second line just has newlines then we let it pass
// through because when you double click on a comment in GitHub and then you
// paste it again, GitHub adds two newlines and so we wanted to allow copying
// and pasting GitHub comments.
var multiLineRegex = regexp.MustCompile(`.*\r?\n[^\r\n]+`)

//go:generate pegomock generate -m --use-experimental-model-gen --package mocks -o mocks/mock_comment_parsing.go CommentParsing

// CommentParsing handles parsing pull request comments.
type CommentParsing interface {
	// Parse attempts to parse a pull request comment to see if it's an Atlantis
	// command.
	Parse(comment string, vcsHost models.VCSHostType) CommentParseResult
}

// CommentParser implements CommentParsing
type CommentParser struct {
	GithubUser  string
	GithubToken string
	GitlabUser  string
	GitlabToken string
}

// CommentParseResult describes the result of parsing a comment as a command.
type CommentParseResult struct {
	// Command is the successfully parsed command. Will be nil if
	// CommentResponse or Ignore is set.
	Command *CommentCommand
	// CommentResponse is set when we should respond immediately to the command
	// for example for atlantis help.
	CommentResponse string
	// Ignore is set to true when we should just ignore this comment.
	Ignore bool
}

// Parse parses the comment as an Atlantis command.
//
// Valid commands contain:
// - The initial "executable" name, 'run' or 'atlantis' or '@GithubUser'
//   where GithubUser is the API user Atlantis is running as.
// - Then a command, either 'plan', 'apply', or 'help'.
// - Then optional flags, then an optional separator '--' followed by optional
//   extra flags to be appended to the terraform plan/apply command.
//
// Examples:
// - atlantis help
// - run plan
// - @GithubUser plan -w staging
// - atlantis plan -w staging -d dir --verbose
// - atlantis plan --verbose -- -key=value -key2 value2
//
func (e *CommentParser) Parse(comment string, vcsHost models.VCSHostType) CommentParseResult {
	if multiLineRegex.MatchString(comment) {
		return CommentParseResult{Ignore: true}
	}

	// strings.Fields strips out newlines but that's okay since we've removed
	// multiline strings above.
	args := strings.Fields(comment)
	if len(args) < 1 {
		return CommentParseResult{Ignore: true}
	}

	// Helpfully warn the user if they're using "terraform" instead of "atlantis"
	if args[0] == "terraform" {
		return CommentParseResult{CommentResponse: DidYouMeanAtlantisComment}
	}

	// Atlantis can be invoked using the name of the VCS host user we're
	// running under. Need to be able to match against that user.
	vcsUser := e.GithubUser
	if vcsHost == models.Gitlab {
		vcsUser = e.GitlabUser
	}
	executableNames := []string{"run", "atlantis", "@" + vcsUser}

	// If the comment doesn't start with the name of our 'executable' then
	// ignore it.
	if !e.stringInSlice(args[0], executableNames) {
		return CommentParseResult{Ignore: true}
	}

	// If they've just typed the name of the executable then give them the help
	// output.
	if len(args) == 1 {
		return CommentParseResult{CommentResponse: HelpComment}
	}
	command := args[1]

	// Help output.
	if e.stringInSlice(command, []string{"help", "-h", "--help"}) {
		return CommentParseResult{CommentResponse: HelpComment}
	}

	// Need to have a plan or apply at this point.
	if !e.stringInSlice(command, []string{PlanCommand.String(), ApplyCommand.String()}) {
		return CommentParseResult{CommentResponse: fmt.Sprintf("```\nError: unknown command %q.\nRun 'atlantis --help' for usage.\n```", command)}
	}

	var workspace string
	var dir string
	var project string
	var verbose bool
	var extraArgs []string
	var flagSet *pflag.FlagSet
	var name CommandName

	// Set up the flag parsing depending on the command.
	switch command {
	case PlanCommand.String():
		name = PlanCommand
		flagSet = pflag.NewFlagSet(PlanCommand.String(), pflag.ContinueOnError)
		flagSet.SetOutput(ioutil.Discard)
		flagSet.StringVarP(&workspace, WorkspaceFlagLong, WorkspaceFlagShort, "", "Switch to this Terraform workspace before planning.")
		flagSet.StringVarP(&dir, DirFlagLong, DirFlagShort, "", "Which directory to run plan in relative to root of repo, ex. 'child/dir'.")
		flagSet.StringVarP(&project, ProjectFlagLong, ProjectFlagShort, "", fmt.Sprintf("Which project to run plan for. Refers to the name of the project configured in %s. Cannot be used at same time as workspace or dir flags.", yaml.AtlantisYAMLFilename))
		flagSet.BoolVarP(&verbose, VerboseFlagLong, VerboseFlagShort, false, "Append Atlantis log to comment.")
	case ApplyCommand.String():
		name = ApplyCommand
		flagSet = pflag.NewFlagSet(ApplyCommand.String(), pflag.ContinueOnError)
		flagSet.SetOutput(ioutil.Discard)
		flagSet.StringVarP(&workspace, WorkspaceFlagLong, WorkspaceFlagShort, "", "Apply the plan for this Terraform workspace.")
		flagSet.StringVarP(&dir, DirFlagLong, DirFlagShort, "", "Apply the plan for this directory, relative to root of repo, ex. 'child/dir'.")
		flagSet.StringVarP(&project, ProjectFlagLong, ProjectFlagShort, "", fmt.Sprintf("Apply the plan for this project. Refers to the name of the project configured in %s. Cannot be used at same time as workspace or dir flags.", yaml.AtlantisYAMLFilename))
		flagSet.BoolVarP(&verbose, VerboseFlagLong, VerboseFlagShort, false, "Append Atlantis log to comment.")
	default:
		return CommentParseResult{CommentResponse: fmt.Sprintf("Error: unknown command %q – this is a bug", command)}
	}

	// Now parse the flags.
	// It's safe to use [2:] because we know there's at least 2 elements in args.
	err := flagSet.Parse(args[2:])
	if err == pflag.ErrHelp {
		return CommentParseResult{CommentResponse: fmt.Sprintf("```\nUsage of %s:\n%s\n```", command, flagSet.FlagUsagesWrapped(usagesCols))}
	}
	if err != nil {
		return CommentParseResult{CommentResponse: e.errMarkdown(err.Error(), command, flagSet)}
	}

	var unusedArgs []string
	if flagSet.ArgsLenAtDash() == -1 {
		unusedArgs = flagSet.Args()
	} else {
		unusedArgs = flagSet.Args()[0:flagSet.ArgsLenAtDash()]
	}
	if len(unusedArgs) > 0 {
		return CommentParseResult{CommentResponse: e.errMarkdown(fmt.Sprintf("unknown argument(s) – %s", strings.Join(unusedArgs, " ")), command, flagSet)}
	}

	if flagSet.ArgsLenAtDash() != -1 {
		extraArgsUnsafe := flagSet.Args()[flagSet.ArgsLenAtDash():]
		// Quote all extra args so there isn't a security issue when we append
		// them to the terraform commands, ex. "; cat /etc/passwd"
		for _, arg := range extraArgsUnsafe {
			quotesEscaped := strings.Replace(arg, `"`, `\"`, -1)
			extraArgs = append(extraArgs, fmt.Sprintf(`"%s"`, quotesEscaped))
		}
	}

	dir, err = e.validateDir(dir)
	if err != nil {
		return CommentParseResult{CommentResponse: e.errMarkdown(err.Error(), command, flagSet)}
	}

	// Use the same validation that Terraform uses: https://git.io/vxGhU. Plus
	// we also don't allow '..'. We don't want the workspace to contain a path
	// since we create files based on the name.
	if workspace != url.PathEscape(workspace) || strings.Contains(workspace, "..") {
		return CommentParseResult{CommentResponse: e.errMarkdown(fmt.Sprintf("invalid workspace: %q", workspace), command, flagSet)}
	}

	// If project is specified, dir or workspace should not be set. Since we
	// dir/workspace have defaults we can't detect if the user set the flag
	// to the default or didn't set the flag so there is an edge case here we
	// don't detect, ex. atlantis plan -p project -d . -w default won't cause
	// an error.
	if project != "" && (workspace != "" || dir != "") {
		err := fmt.Sprintf("cannot use -%s/--%s at same time as -%s/--%s or -%s/--%s", ProjectFlagShort, ProjectFlagLong, DirFlagShort, DirFlagLong, WorkspaceFlagShort, WorkspaceFlagLong)
		return CommentParseResult{CommentResponse: e.errMarkdown(err, command, flagSet)}
	}

	return CommentParseResult{
		Command: NewCommentCommand(dir, extraArgs, name, verbose, workspace, project),
	}
}

func (e *CommentParser) validateDir(dir string) (string, error) {
	if dir == "" {
		return dir, nil
	}
	validatedDir := filepath.Clean(dir)
	// Join with . so the path is relative. This helps us if they use '/',
	// and is safe to do if their path is relative since it's a no-op.
	validatedDir = filepath.Join(".", validatedDir)
	// Need to clean again to resolve relative validatedDirs.
	validatedDir = filepath.Clean(validatedDir)
	// Detect relative dirs since they're not allowed.
	if strings.HasPrefix(validatedDir, "..") {
		return "", fmt.Errorf("using a relative path %q with -%s/--%s is not allowed", dir, DirFlagShort, DirFlagLong)
	}

	return validatedDir, nil
}

func (e *CommentParser) stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

func (e *CommentParser) errMarkdown(errMsg string, command string, flagSet *pflag.FlagSet) string {
	return fmt.Sprintf("```\nError: %s.\nUsage of %s:\n%s```", errMsg, command, flagSet.FlagUsagesWrapped(usagesCols))
}

var HelpComment = "```cmake\n" +
	`atlantis
Terraform automation and collaboration for your team

Usage:
  atlantis <command> [options] -- [terraform options]

Examples:
  # run plan in the root directory passing the -target flag to terraform
  atlantis plan -d . -- -target=resource

  # apply all unapplied plans
  atlantis apply

  # apply the plan for the root directory and staging workspace
  atlantis apply -d . -w staging

Commands:
  plan   Runs 'terraform plan' for the changes in this pull request.
  apply  Runs 'terraform apply' on all unapplied plans.
         To only apply a specific plan, use the -d, -w and -p flags.
  help   View help.

Flags:
  -h, --help   help for atlantis

Use "atlantis [command] --help" for more information about a command.
`

var DidYouMeanAtlantisComment = "Did you mean to use `atlantis` instead of `terraform`?"
