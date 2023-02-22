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

package events

import (
	"bytes"
	"fmt"
	"io"
	"net/url"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"

	"github.com/google/shlex"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/utils"
	"github.com/spf13/pflag"
)

const (
	workspaceFlagLong          = "workspace"
	workspaceFlagShort         = "w"
	dirFlagLong                = "dir"
	dirFlagShort               = "d"
	projectFlagLong            = "project"
	projectFlagShort           = "p"
	autoMergeDisabledFlagLong  = "auto-merge-disabled"
	autoMergeDisabledFlagShort = ""
	verboseFlagLong            = "verbose"
	verboseFlagShort           = ""
)

// multiLineRegex is used to ignore multi-line comments since those aren't valid
// Atlantis commands. If the second line just has newlines then we let it pass
// through because when you double click on a comment in GitHub and then you
// paste it again, GitHub adds two newlines and so we wanted to allow copying
// and pasting GitHub comments.
var multiLineRegex = regexp.MustCompile(`.*\r?\n[^\r\n]+`)

//go:generate pegomock generate -m --package mocks -o mocks/mock_comment_parsing.go CommentParsing

// CommentParsing handles parsing pull request comments.
type CommentParsing interface {
	// Parse attempts to parse a pull request comment to see if it's an Atlantis
	// command.
	Parse(comment string, vcsHost models.VCSHostType) CommentParseResult
}

//go:generate pegomock generate -m --package mocks -o mocks/mock_comment_building.go CommentBuilder

// CommentBuilder builds comment commands that can be used on pull requests.
type CommentBuilder interface {
	// BuildPlanComment builds a plan comment for the specified args.
	BuildPlanComment(repoRelDir string, workspace string, project string, commentArgs []string) string
	// BuildApplyComment builds an apply comment for the specified args.
	BuildApplyComment(repoRelDir string, workspace string, project string, autoMergeDisabled bool) string
}

// CommentParser implements CommentParsing
type CommentParser struct {
	GithubUser      string
	GitlabUser      string
	BitbucketUser   string
	AzureDevopsUser string
	ExecutableName  string
	AllowCommands   []command.Name
}

// NewCommentParser returns a CommentParser
func NewCommentParser(githubUser, gitlabUser, bitbucketUser, azureDevopsUser, executableName string, allowCommands []command.Name) *CommentParser {
	var commentAllowCommands []command.Name
	for _, acceptableCommand := range command.AllCommentCommands {
		for _, allowCommand := range allowCommands {
			if acceptableCommand == allowCommand {
				commentAllowCommands = append(commentAllowCommands, allowCommand)
				break // for distinct
			}
		}
	}

	return &CommentParser{
		GithubUser:      githubUser,
		GitlabUser:      gitlabUser,
		BitbucketUser:   bitbucketUser,
		AzureDevopsUser: azureDevopsUser,
		ExecutableName:  executableName,
		AllowCommands:   commentAllowCommands,
	}
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
//   - The initial "executable" name, 'run' or 'atlantis' or '@GithubUser'
//     where GithubUser is the API user Atlantis is running as.
//   - Then a command: 'plan', 'apply', 'unlock', 'version, 'approve_policies',
//     or 'help'.
//   - Then optional flags, then an optional separator '--' followed by optional
//     extra flags to be appended to the terraform plan/apply command.
//
// Examples:
// - atlantis help
// - run apply
// - @GithubUser plan -w staging
// - atlantis plan -w staging -d dir --verbose
// - atlantis plan --verbose -- -key=value -key2 value2
// - atlantis unlock
// - atlantis version
// - atlantis approve_policies
// - atlantis import ADDRESS ID
func (e *CommentParser) Parse(rawComment string, vcsHost models.VCSHostType) CommentParseResult {
	comment := strings.TrimSpace(rawComment)

	if multiLineRegex.MatchString(comment) {
		return CommentParseResult{Ignore: true}
	}

	// We first use strings.Fields to parse and do an initial evaluation.
	// Later we use a proper shell parser and re-parse.
	args := strings.Fields(comment)
	if len(args) < 1 {
		return CommentParseResult{Ignore: true}
	}

	// Helpfully warn the user if they're using "terraform" instead of "atlantis"
	if args[0] == "terraform" && e.ExecutableName != "terraform" {
		return CommentParseResult{CommentResponse: fmt.Sprintf(DidYouMeanAtlantisComment, e.ExecutableName, "terraform")}
	}

	// Helpfully warn the user that the command might be misspelled
	if utils.IsSimilarWord(args[0], e.ExecutableName) {
		return CommentParseResult{CommentResponse: fmt.Sprintf(DidYouMeanAtlantisComment, e.ExecutableName, args[0])}
	}

	// Atlantis can be invoked using the name of the VCS host user we're
	// running under. Need to be able to match against that user.
	var vcsUser string
	switch vcsHost {
	case models.Github:
		vcsUser = e.GithubUser
	case models.Gitlab:
		vcsUser = e.GitlabUser
	case models.BitbucketCloud, models.BitbucketServer:
		vcsUser = e.BitbucketUser
	case models.AzureDevops:
		vcsUser = e.AzureDevopsUser
	}
	executableNames := []string{"run", e.ExecutableName, "@" + vcsUser}
	if !e.stringInSlice(args[0], executableNames) {
		return CommentParseResult{Ignore: true}
	}

	// Now that we know Atlantis is being invoked, re-parse using a shell-style
	// parser.
	args, err := shlex.Split(comment)
	if err != nil {
		return CommentParseResult{CommentResponse: fmt.Sprintf("```\nError parsing command: %s\n```", err)}
	}
	if len(args) < 1 {
		return CommentParseResult{Ignore: true}
	}

	// If they've just typed the name of the executable then give them the help
	// output.
	if len(args) == 1 {
		return CommentParseResult{CommentResponse: e.HelpComment()}
	}
	cmd := args[1]

	// Help output.
	if e.stringInSlice(cmd, []string{"help", "-h", "--help"}) {
		return CommentParseResult{CommentResponse: e.HelpComment()}
	}

	// Need to have allow commands at this point.
	if !e.isAllowedCommand(cmd) {
		var allowCommandList []string
		for _, allowCommand := range e.AllowCommands {
			allowCommandList = append(allowCommandList, allowCommand.String())
		}
		return CommentParseResult{CommentResponse: fmt.Sprintf("```\nError: unknown command %q.\nRun '%s --help' for usage.\nAvailable commands(--allow-commands): %s\n```", cmd, e.ExecutableName, strings.Join(allowCommandList, ", "))}
	}

	var workspace string
	var dir string
	var project string
	var verbose, autoMergeDisabled bool
	var flagSet *pflag.FlagSet
	var name command.Name

	// Set up the flag parsing depending on the command.
	switch cmd {
	case command.Plan.String():
		name = command.Plan
		flagSet = pflag.NewFlagSet(command.Plan.String(), pflag.ContinueOnError)
		flagSet.SetOutput(io.Discard)
		flagSet.StringVarP(&workspace, workspaceFlagLong, workspaceFlagShort, "", "Switch to this Terraform workspace before planning.")
		flagSet.StringVarP(&dir, dirFlagLong, dirFlagShort, "", "Which directory to run plan in relative to root of repo, ex. 'child/dir'.")
		flagSet.StringVarP(&project, projectFlagLong, projectFlagShort, "", "Which project to run plan for. Refers to the name of the project configured in a repo config file. Cannot be used at same time as workspace or dir flags.")
		flagSet.BoolVarP(&verbose, verboseFlagLong, verboseFlagShort, false, "Append Atlantis log to comment.")
	case command.Apply.String():
		name = command.Apply
		flagSet = pflag.NewFlagSet(command.Apply.String(), pflag.ContinueOnError)
		flagSet.SetOutput(io.Discard)
		flagSet.StringVarP(&workspace, workspaceFlagLong, workspaceFlagShort, "", "Apply the plan for this Terraform workspace.")
		flagSet.StringVarP(&dir, dirFlagLong, dirFlagShort, "", "Apply the plan for this directory, relative to root of repo, ex. 'child/dir'.")
		flagSet.StringVarP(&project, projectFlagLong, projectFlagShort, "", "Apply the plan for this project. Refers to the name of the project configured in a repo config file. Cannot be used at same time as workspace or dir flags.")
		flagSet.BoolVarP(&autoMergeDisabled, autoMergeDisabledFlagLong, autoMergeDisabledFlagShort, false, "Disable automerge after apply.")
		flagSet.BoolVarP(&verbose, verboseFlagLong, verboseFlagShort, false, "Append Atlantis log to comment.")
	case command.ApprovePolicies.String():
		name = command.ApprovePolicies
		flagSet = pflag.NewFlagSet(command.ApprovePolicies.String(), pflag.ContinueOnError)
		flagSet.SetOutput(io.Discard)
		flagSet.BoolVarP(&verbose, verboseFlagLong, verboseFlagShort, false, "Append Atlantis log to comment.")
	case command.Unlock.String():
		name = command.Unlock
		flagSet = pflag.NewFlagSet(command.Unlock.String(), pflag.ContinueOnError)
		flagSet.SetOutput(io.Discard)
	case command.Version.String():
		name = command.Version
		flagSet = pflag.NewFlagSet(command.Version.String(), pflag.ContinueOnError)
		flagSet.StringVarP(&workspace, workspaceFlagLong, workspaceFlagShort, "", "Switch to this Terraform workspace before running version.")
		flagSet.StringVarP(&dir, dirFlagLong, dirFlagShort, "", "Which directory to run version in relative to root of repo, ex. 'child/dir'.")
		flagSet.StringVarP(&project, projectFlagLong, projectFlagShort, "", "Print the version for this project. Refers to the name of the project configured in a repo config file.")
		flagSet.BoolVarP(&verbose, verboseFlagLong, verboseFlagShort, false, "Append Atlantis log to comment.")
	case command.Import.String():
		name = command.Import
		flagSet = pflag.NewFlagSet(command.Import.String(), pflag.ContinueOnError)
		flagSet.SetOutput(io.Discard)
		flagSet.StringVarP(&workspace, workspaceFlagLong, workspaceFlagShort, "", "Switch to this Terraform workspace before importing.")
		flagSet.StringVarP(&dir, dirFlagLong, dirFlagShort, "", "Which directory to run import in relative to root of repo, ex. 'child/dir'.")
		flagSet.StringVarP(&project, projectFlagLong, projectFlagShort, "", "Which project to run import for. Refers to the name of the project configured in a repo config file. Cannot be used at same time as workspace or dir flags.")
		flagSet.BoolVarP(&verbose, verboseFlagLong, verboseFlagShort, false, "Append Atlantis log to comment.")
	case command.State.String():
		name = command.State
		flagSet = pflag.NewFlagSet(command.State.String(), pflag.ContinueOnError)
		flagSet.SetOutput(io.Discard)
		flagSet.StringVarP(&workspace, workspaceFlagLong, workspaceFlagShort, "", "Switch to this Terraform workspace before processing tfstate.")
		flagSet.StringVarP(&dir, dirFlagLong, dirFlagShort, "", "Which directory to run state command in relative to root of repo, ex. 'child/dir'.")
		flagSet.StringVarP(&project, projectFlagLong, projectFlagShort, "", "Which project to run state command for. Refers to the name of the project configured in a repo config file. Cannot be used at same time as workspace or dir flags.")
		flagSet.BoolVarP(&verbose, verboseFlagLong, verboseFlagShort, false, "Append Atlantis log to comment.")
	default:
		return CommentParseResult{CommentResponse: fmt.Sprintf("Error: unknown command %q – this is a bug", cmd)}
	}

	subName, extraArgs, errResult := e.parseArgs(name, args, flagSet)
	if errResult != "" {
		return CommentParseResult{CommentResponse: errResult}
	}

	dir, err = e.validateDir(dir)
	if err != nil {
		return CommentParseResult{CommentResponse: e.errMarkdown(err.Error(), cmd, flagSet)}
	}

	// Use the same validation that Terraform uses: https://git.io/vxGhU. Plus
	// we also don't allow '..'. We don't want the workspace to contain a path
	// since we create files based on the name.
	if workspace != url.PathEscape(workspace) || strings.Contains(workspace, "..") {
		return CommentParseResult{CommentResponse: e.errMarkdown(fmt.Sprintf("invalid workspace: %q", workspace), cmd, flagSet)}
	}

	// If project is specified, dir or workspace should not be set. Since we
	// dir/workspace have defaults we can't detect if the user set the flag
	// to the default or didn't set the flag so there is an edge case here we
	// don't detect, ex. atlantis plan -p project -d . -w default won't cause
	// an error.
	if project != "" && (workspace != "" || dir != "") {
		err := fmt.Sprintf("cannot use -%s/--%s at same time as -%s/--%s or -%s/--%s", projectFlagShort, projectFlagLong, dirFlagShort, dirFlagLong, workspaceFlagShort, workspaceFlagLong)
		return CommentParseResult{CommentResponse: e.errMarkdown(err, cmd, flagSet)}
	}

	return CommentParseResult{
		Command: NewCommentCommand(dir, extraArgs, name, subName, verbose, autoMergeDisabled, workspace, project),
	}
}

func (e *CommentParser) parseArgs(name command.Name, args []string, flagSet *pflag.FlagSet) (string, []string, string) {
	// Now parse the flags.
	// It's safe to use [2:] because we know there's at least 2 elements in args.
	err := flagSet.Parse(args[2:])
	if err == pflag.ErrHelp {
		return "", nil, fmt.Sprintf("```\nUsage of %s:\n%s\n```", name.DefaultUsage(), flagSet.FlagUsagesWrapped(usagesCols))
	}
	if err != nil {
		if name == command.Unlock {
			return "", nil, fmt.Sprintf(UnlockUsage, e.ExecutableName)
		}
		return "", nil, e.errMarkdown(err.Error(), name.String(), flagSet)
	}

	var commandArgs []string // commandArgs are the arguments that are passed before `--` without any parameter flags.
	if flagSet.ArgsLenAtDash() == -1 {
		commandArgs = flagSet.Args()
	} else {
		commandArgs = flagSet.Args()[0:flagSet.ArgsLenAtDash()]
	}

	// If command require subcommand, get it from command args
	var subCommand string
	availableSubCommands := name.SubCommands()
	if len(availableSubCommands) > 0 { // command requires a subcommand
		if len(commandArgs) < 1 {
			return "", nil, e.errMarkdown("subcommand required", name.String(), flagSet)
		}
		subCommand, commandArgs = commandArgs[0], commandArgs[1:]
		isAvailableSubCommand := utils.SlicesContains(availableSubCommands, subCommand)
		if !isAvailableSubCommand {
			errMsg := fmt.Sprintf("invalid subcommand %s (not %s)", subCommand, strings.Join(availableSubCommands, ", "))
			return "", nil, e.errMarkdown(errMsg, name.String(), flagSet)
		}
	}

	// check command args count requirements
	commandArgCount, err := name.CommandArgCount(subCommand)
	if err != nil {
		return "", nil, e.errMarkdown(err.Error(), name.String(), flagSet)
	}
	if !commandArgCount.IsMatchCount(len(commandArgs)) {
		return "", nil, e.errMarkdown(fmt.Sprintf("unknown argument(s) – %s", strings.Join(commandArgs, " ")), name.DefaultUsage(), flagSet)
	}

	var extraArgs []string // command extra_args
	if flagSet.ArgsLenAtDash() != -1 {
		extraArgs = append(extraArgs, flagSet.Args()[flagSet.ArgsLenAtDash():]...)
	}

	// pass commandArgs into extraArgs after extra args.
	// - after comment_parser, we will use extra_args only.
	// - terraform command args accept after options like followings
	//   - e.g.
	//     - from: `atlantis import ADDRESS ID -- -var foo=bar
	//     - to: `terraform import -var foo=bar ADDRESS ID`
	//   - e.g.
	//     - from: `atlantis state rm ADDRESS1 ADDRESS2 -- -var foo=bar
	//     - to: `terraform state rm -var foo=bar ADDRESS1 ADDRESS2` (subcommand=rm)
	extraArgs = append(extraArgs, commandArgs...)
	return subCommand, extraArgs, ""
}

// BuildPlanComment builds a plan comment for the specified args.
func (e *CommentParser) BuildPlanComment(repoRelDir string, workspace string, project string, commentArgs []string) string {
	flags := e.buildFlags(repoRelDir, workspace, project, false)
	commentFlags := ""
	if len(commentArgs) > 0 {
		var flagsWithoutQuotes []string
		for _, f := range commentArgs {
			f = strings.TrimPrefix(f, "\"")
			f = strings.TrimSuffix(f, "\"")
			flagsWithoutQuotes = append(flagsWithoutQuotes, f)
		}
		commentFlags = fmt.Sprintf(" -- %s", strings.Join(flagsWithoutQuotes, " "))
	}
	return fmt.Sprintf("%s %s%s%s", e.ExecutableName, command.Plan.String(), flags, commentFlags)
}

// BuildApplyComment builds an apply comment for the specified args.
func (e *CommentParser) BuildApplyComment(repoRelDir string, workspace string, project string, autoMergeDisabled bool) string {
	flags := e.buildFlags(repoRelDir, workspace, project, autoMergeDisabled)
	return fmt.Sprintf("%s %s%s", e.ExecutableName, command.Apply.String(), flags)
}

func (e *CommentParser) buildFlags(repoRelDir string, workspace string, project string, autoMergeDisabled bool) string {
	// Add quotes if dir has spaces.
	if strings.Contains(repoRelDir, " ") {
		repoRelDir = fmt.Sprintf("%q", repoRelDir)
	}

	var flags string
	switch {
	// If project is specified we can just use its name.
	case project != "":
		flags = fmt.Sprintf(" -%s %s", projectFlagShort, project)
	case repoRelDir == DefaultRepoRelDir && workspace == DefaultWorkspace:
		// If it's the root and default workspace then we just need to specify one
		// of the flags and the other will get defaulted.
		flags = fmt.Sprintf(" -%s %s", dirFlagShort, DefaultRepoRelDir)
	case repoRelDir == DefaultRepoRelDir:
		// If dir is the default then we just need to specify workspace.
		flags = fmt.Sprintf(" -%s %s", workspaceFlagShort, workspace)
	case workspace == DefaultWorkspace:
		// If workspace is the default then we just need to specify the dir.
		flags = fmt.Sprintf(" -%s %s", dirFlagShort, repoRelDir)
	default:
		// Otherwise we have to specify both flags.
		flags = fmt.Sprintf(" -%s %s -%s %s", dirFlagShort, repoRelDir, workspaceFlagShort, workspace)
	}
	if autoMergeDisabled {
		flags = fmt.Sprintf("%s --%s", flags, autoMergeDisabledFlagLong)
	}
	return flags
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
		return "", fmt.Errorf("using a relative path %q with -%s/--%s is not allowed", dir, dirFlagShort, dirFlagLong)
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

func (e *CommentParser) isAllowedCommand(cmd string) bool {
	for _, allowed := range e.AllowCommands {
		if allowed.String() == cmd {
			return true
		}
	}
	return false
}

func (e *CommentParser) errMarkdown(errMsg string, cmd string, flagSet *pflag.FlagSet) string {
	return fmt.Sprintf("```\nError: %s.\nUsage of %s:\n%s```", errMsg, cmd, flagSet.FlagUsagesWrapped(usagesCols))
}

func (e *CommentParser) HelpComment() string {
	buf := &bytes.Buffer{}
	var tmpl = template.Must(template.New("").Parse(helpCommentTemplate))
	if err := tmpl.Execute(buf, struct {
		ExecutableName       string
		AllowVersion         bool
		AllowPlan            bool
		AllowApply           bool
		AllowUnlock          bool
		AllowApprovePolicies bool
		AllowImport          bool
		AllowState           bool
	}{
		ExecutableName:       e.ExecutableName,
		AllowVersion:         e.isAllowedCommand(command.Version.String()),
		AllowPlan:            e.isAllowedCommand(command.Plan.String()),
		AllowApply:           e.isAllowedCommand(command.Apply.String()),
		AllowUnlock:          e.isAllowedCommand(command.Unlock.String()),
		AllowApprovePolicies: e.isAllowedCommand(command.ApprovePolicies.String()),
		AllowImport:          e.isAllowedCommand(command.Import.String()),
		AllowState:           e.isAllowedCommand(command.State.String()),
	}); err != nil {
		return fmt.Sprintf("Failed to render template, this is a bug: %v", err)
	}
	return buf.String()
}

var helpCommentTemplate = "```cmake\n" +
	`atlantis
Terraform Pull Request Automation

Usage:
  {{ .ExecutableName }} <command> [options] -- [terraform options]

Examples:
  # show atlantis help
  {{ .ExecutableName }} help
{{- if .AllowPlan }}

  # run plan in the root directory passing the -target flag to terraform
  {{ .ExecutableName }} plan -d . -- -target=resource
{{- end }}
{{- if .AllowApply }}

  # apply all unapplied plans from this pull request
  {{ .ExecutableName }} apply

  # apply the plan for the root directory and staging workspace
  {{ .ExecutableName }} apply -d . -w staging
{{- end }}

Commands:
{{- if .AllowPlan }}
  plan     Runs 'terraform plan' for the changes in this pull request.
           To plan a specific project, use the -d, -w and -p flags.
{{- end }}
{{- if .AllowApply }}
  apply    Runs 'terraform apply' on all unapplied plans from this pull request.
           To only apply a specific plan, use the -d, -w and -p flags.
{{- end }}
{{- if .AllowUnlock }}
  unlock   Removes all atlantis locks and discards all plans for this PR.
           To unlock a specific plan you can use the Atlantis UI.
{{- end }}
{{- if .AllowApprovePolicies }}
  approve_policies
           Approves all current policy checking failures for the PR.
{{- end }}
{{- if .AllowVersion }}
  version  Print the output of 'terraform version'
{{- end }}
{{- if .AllowImport }}
  import ADDRESS ID
           Runs 'terraform import' for the passed address resource.
           To import a specific project, use the -d, -w and -p flags.
{{- end }}
{{- if .AllowState }}
  state rm ADDRESS...
           Runs 'terraform state rm' for the passed address resource.
           To remove a specific project resource, use the -d, -w and -p flags.
{{- end }}
  help     View help.

Flags:
  -h, --help   help for atlantis

Use "{{ .ExecutableName }} [command] --help" for more information about a command.` +
	"\n```"

// DidYouMeanAtlantisComment is the comment we add to the pull request when
// someone runs a misspelled command or terraform instead of atlantis.
var DidYouMeanAtlantisComment = "Did you mean to use `%s` instead of `%s`?"

// UnlockUsage is the comment we add to the pull request when someone runs
// `atlantis unlock` with flags.

var UnlockUsage = "`Usage of unlock:`\n\n ```cmake\n" +
	`%s unlock

  Unlocks the entire PR and discards all plans in this PR.
  Arguments or flags are not supported at the moment.
  If you need to unlock a specific project please use the atlantis UI.` +
	"\n```"
