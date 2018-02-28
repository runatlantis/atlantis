package events

import (
	"errors"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/google/go-github/github"
	"github.com/lkysow/go-gitlab"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/events/vcs"
	"github.com/spf13/pflag"
)

const gitlabPullOpened = "opened"
const usagesCols = 90

// multiLineRegex is used to ignore multi-line comments since those aren't valid
// Atlantis commands.
var multiLineRegex = regexp.MustCompile(`.*\r?\n.+`)

//go:generate pegomock generate -m --use-experimental-model-gen --package mocks -o mocks/mock_event_parsing.go EventParsing

type Command struct {
	Name      CommandName
	Workspace string
	Verbose   bool
	Flags     []string
	// Dir is the path relative to the repo root to run the command in.
	// If empty string then it wasn't specified. "." is the root of the repo.
	// Dir will never end in "/".
	Dir string
}

type EventParsing interface {
	DetermineCommand(comment string, vcsHost vcs.Host) CommandParseResult
	ParseGithubIssueCommentEvent(comment *github.IssueCommentEvent) (baseRepo models.Repo, user models.User, pullNum int, err error)
	ParseGithubPull(pull *github.PullRequest) (models.PullRequest, models.Repo, error)
	ParseGithubRepo(ghRepo *github.Repository) (models.Repo, error)
	ParseGitlabMergeEvent(event gitlab.MergeEvent) (models.PullRequest, models.Repo)
	ParseGitlabMergeCommentEvent(event gitlab.MergeCommentEvent) (baseRepo models.Repo, headRepo models.Repo, user models.User)
	ParseGitlabMergeRequest(mr *gitlab.MergeRequest) models.PullRequest
}

type EventParser struct {
	GithubUser  string
	GithubToken string
	GitlabUser  string
	GitlabToken string
}

// CommandParseResult describes the result of parsing a comment as a command.
type CommandParseResult struct {
	// Command is the successfully parsed command. Will be nil if
	// CommentResponse or Ignore is set.
	Command *Command
	// CommentResponse is set when we should respond immediately to the command
	// for example for atlantis help.
	CommentResponse string
	// Ignore is set to true when we should just ignore this comment.
	Ignore bool
}

// DetermineCommand parses the comment as an Atlantis command.
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
// nolint: gocyclo
func (e *EventParser) DetermineCommand(comment string, vcsHost vcs.Host) CommandParseResult {
	if multiLineRegex.MatchString(comment) {
		return CommandParseResult{Ignore: true}
	}

	// strings.Fields strips out newlines but that's okay since we've removed
	// multiline strings above.
	args := strings.Fields(comment)
	if len(args) < 1 {
		return CommandParseResult{Ignore: true}
	}

	// Helpfully warn the user if they're using "terraform" instead of "atlantis"
	if args[0] == "terraform" {
		return CommandParseResult{CommentResponse: DidYouMeanAtlantisComment}
	}

	// Atlantis can be invoked using the name of the VCS host user we're
	// running under. Need to be able to match against that user.
	vcsUser := e.GithubUser
	if vcsHost == vcs.Gitlab {
		vcsUser = e.GitlabUser
	}
	executableNames := []string{"run", "atlantis", "@" + vcsUser}

	// If the comment doesn't start with the name of our 'executable' then
	// ignore it.
	if !e.stringInSlice(args[0], executableNames) {
		return CommandParseResult{Ignore: true}
	}

	// If they've just typed the name of the executable then give them the help
	// output.
	if len(args) == 1 {
		return CommandParseResult{CommentResponse: HelpComment}
	}
	command := args[1]

	// Help output.
	if e.stringInSlice(command, []string{"help", "-h", "--help"}) {
		return CommandParseResult{CommentResponse: HelpComment}
	}

	// Need to have a plan or apply at this point.
	if !e.stringInSlice(command, []string{"plan", "apply"}) {
		return CommandParseResult{CommentResponse: fmt.Sprintf("```\nError: unknown command %q.\nRun 'atlantis --help' for usage.\n```", command)}
	}

	var workspace string
	var dir string
	var verbose bool
	var extraArgs []string
	var flagSet *pflag.FlagSet
	var name CommandName

	// Set up the flag parsing depending on the command.
	const defaultWorkspace = "default"
	switch command {
	case "plan":
		name = Plan
		flagSet = pflag.NewFlagSet("plan", pflag.ContinueOnError)
		flagSet.SetOutput(ioutil.Discard)
		flagSet.StringVarP(&workspace, "workspace", "w", defaultWorkspace, "Switch to this Terraform workspace before planning.")
		flagSet.StringVarP(&dir, "dir", "d", "", "Which directory to run plan in relative to root of repo. Use '.' for root. If not specified, will attempt to run plan for all Terraform projects we think were modified in this changeset.")
		flagSet.BoolVarP(&verbose, "verbose", "", false, "Append Atlantis log to comment.")
	case "apply":
		name = Apply
		flagSet = pflag.NewFlagSet("apply", pflag.ContinueOnError)
		flagSet.SetOutput(ioutil.Discard)
		flagSet.StringVarP(&workspace, "workspace", "w", defaultWorkspace, "Apply the plan for this Terraform workspace.")
		flagSet.StringVarP(&dir, "dir", "d", "", "Apply the plan for this directory, relative to root of repo. Use '.' for root. If not specified, will run apply against all plans created for this workspace.")
		flagSet.BoolVarP(&verbose, "verbose", "", false, "Append Atlantis log to comment.")
	default:
		return CommandParseResult{CommentResponse: fmt.Sprintf("Error: unknown command %q – this is a bug", command)}
	}

	// Now parse the flags.
	// It's safe to use [2:] because we know there's at least 2 elements in args.
	err := flagSet.Parse(args[2:])
	if err == pflag.ErrHelp {
		return CommandParseResult{CommentResponse: fmt.Sprintf("```\nUsage of %s:\n%s\n```", command, flagSet.FlagUsagesWrapped(usagesCols))}
	}
	if err != nil {
		return CommandParseResult{CommentResponse: fmt.Sprintf("```\nError: %s.\nUsage of %s:\n%s\n```", err.Error(), command, flagSet.FlagUsagesWrapped(usagesCols))}
	}
	// We only use the extra args after the --. For example given a comment:
	// "atlantis plan -bad-option -- -target=hi"
	// we only append "-target=hi" to the eventual command.
	// todo: keep track of the args we're discarding and include that with
	//       comment as a warning.
	if flagSet.ArgsLenAtDash() != -1 {
		extraArgsUnsafe := flagSet.Args()[flagSet.ArgsLenAtDash():]
		// Quote all extra args so there isn't a security issue when we append
		// them to the terraform commands, ex. "; cat /etc/passwd"
		for _, arg := range extraArgsUnsafe {
			extraArgs = append(extraArgs, fmt.Sprintf(`"%s"`, arg))
		}
	}

	// If dir is specified, must ensure it's a valid path.
	if dir != "" {
		validatedDir := filepath.Clean(dir)
		// Join with . so the path is relative. This helps us if they use '/',
		// and is safe to do if their path is relative since it's a no-op.
		validatedDir = filepath.Join(".", validatedDir)
		// Need to clean again to resolve relative validatedDirs.
		validatedDir = filepath.Clean(validatedDir)
		// Detect relative dirs since they're not allowed.
		if strings.HasPrefix(validatedDir, "..") {
			return CommandParseResult{CommentResponse: fmt.Sprintf("Error: Using a relative path %q with -d/--dir is not allowed", dir)}
		}

		dir = validatedDir
	}
	// Because we use the workspace name as a file, need to make sure it's
	// not doing something weird like being a relative dir.
	if strings.Contains(workspace, "..") {
		return CommandParseResult{CommentResponse: "Error: Value for -w/--workspace can't contain '..'"}
	}

	return CommandParseResult{
		Command: &Command{Name: name, Verbose: verbose, Workspace: workspace, Dir: dir, Flags: extraArgs},
	}
}

func (e *EventParser) ParseGithubIssueCommentEvent(comment *github.IssueCommentEvent) (baseRepo models.Repo, user models.User, pullNum int, err error) {
	baseRepo, err = e.ParseGithubRepo(comment.Repo)
	if err != nil {
		return
	}
	if comment.Comment == nil || comment.Comment.User.GetLogin() == "" {
		err = errors.New("comment.user.login is null")
		return
	}
	commentorUsername := comment.Comment.User.GetLogin()
	user = models.User{
		Username: commentorUsername,
	}
	pullNum = comment.Issue.GetNumber()
	if pullNum == 0 {
		err = errors.New("issue.number is null")
		return
	}
	return
}

func (e *EventParser) ParseGithubPull(pull *github.PullRequest) (models.PullRequest, models.Repo, error) {
	var pullModel models.PullRequest
	var headRepoModel models.Repo

	commit := pull.Head.GetSHA()
	if commit == "" {
		return pullModel, headRepoModel, errors.New("head.sha is null")
	}
	url := pull.GetHTMLURL()
	if url == "" {
		return pullModel, headRepoModel, errors.New("html_url is null")
	}
	branch := pull.Head.GetRef()
	if branch == "" {
		return pullModel, headRepoModel, errors.New("head.ref is null")
	}
	authorUsername := pull.User.GetLogin()
	if authorUsername == "" {
		return pullModel, headRepoModel, errors.New("user.login is null")
	}
	num := pull.GetNumber()
	if num == 0 {
		return pullModel, headRepoModel, errors.New("number is null")
	}

	headRepoModel, err := e.ParseGithubRepo(pull.Head.Repo)
	if err != nil {
		return pullModel, headRepoModel, err
	}

	pullState := models.Closed
	if pull.GetState() == "open" {
		pullState = models.Open
	}

	return models.PullRequest{
		Author:     authorUsername,
		Branch:     branch,
		HeadCommit: commit,
		URL:        url,
		Num:        num,
		State:      pullState,
	}, headRepoModel, nil
}

func (e *EventParser) ParseGithubRepo(ghRepo *github.Repository) (models.Repo, error) {
	var repo models.Repo
	repoFullName := ghRepo.GetFullName()
	if repoFullName == "" {
		return repo, errors.New("repository.full_name is null")
	}
	repoOwner := ghRepo.Owner.GetLogin()
	if repoOwner == "" {
		return repo, errors.New("repository.owner.login is null")
	}
	repoName := ghRepo.GetName()
	if repoName == "" {
		return repo, errors.New("repository.name is null")
	}
	repoSanitizedCloneURL := ghRepo.GetCloneURL()
	if repoSanitizedCloneURL == "" {
		return repo, errors.New("repository.clone_url is null")
	}

	// Construct HTTPS repo clone url string with username and password.
	repoCloneURL := strings.Replace(repoSanitizedCloneURL, "https://", fmt.Sprintf("https://%s:%s@", e.GithubUser, e.GithubToken), -1)

	return models.Repo{
		Owner:             repoOwner,
		FullName:          repoFullName,
		CloneURL:          repoCloneURL,
		SanitizedCloneURL: repoSanitizedCloneURL,
		Name:              repoName,
	}, nil
}

func (e *EventParser) ParseGitlabMergeEvent(event gitlab.MergeEvent) (models.PullRequest, models.Repo) {
	modelState := models.Closed
	if event.ObjectAttributes.State == gitlabPullOpened {
		modelState = models.Open
	}
	// GitLab also has a "merged" state, but we map that to Closed so we don't
	// need to check for it.

	pull := models.PullRequest{
		URL:        event.ObjectAttributes.URL,
		Author:     event.User.Username,
		Num:        event.ObjectAttributes.IID,
		HeadCommit: event.ObjectAttributes.LastCommit.ID,
		Branch:     event.ObjectAttributes.SourceBranch,
		State:      modelState,
	}

	cloneURL := e.addGitlabAuth(event.Project.GitHTTPURL)
	// Get owner and name from PathWithNamespace because the fields
	// event.Project.Name and event.Project.Owner can have capitals.
	owner, name := e.getOwnerAndName(event.Project.PathWithNamespace)
	repo := models.Repo{
		FullName:          event.Project.PathWithNamespace,
		Name:              name,
		SanitizedCloneURL: event.Project.GitHTTPURL,
		Owner:             owner,
		CloneURL:          cloneURL,
	}
	return pull, repo
}

// addGitlabAuth adds gitlab username/password to the cloneURL.
// We support http and https URLs because GitLab's docs have http:// URLs whereas
// their API responses have https://.
// Ex. https://gitlab.com/owner/repo.git => https://uname:pass@gitlab.com/owner/repo.git
func (e *EventParser) addGitlabAuth(cloneURL string) string {
	httpsReplaced := strings.Replace(cloneURL, "https://", fmt.Sprintf("https://%s:%s@", e.GitlabUser, e.GitlabToken), -1)
	return strings.Replace(httpsReplaced, "http://", fmt.Sprintf("http://%s:%s@", e.GitlabUser, e.GitlabToken), -1)
}

// getOwnerAndName takes pathWithNamespace that should look like "owner/repo"
// and returns "owner", "repo"
func (e *EventParser) getOwnerAndName(pathWithNamespace string) (string, string) {
	pathSplit := strings.Split(pathWithNamespace, "/")
	if len(pathSplit) > 1 {
		return pathSplit[0], pathSplit[1]
	}
	return "", ""
}

// ParseGitlabMergeCommentEvent creates Atlantis models out of a GitLab event.
func (e *EventParser) ParseGitlabMergeCommentEvent(event gitlab.MergeCommentEvent) (baseRepo models.Repo, headRepo models.Repo, user models.User) {
	// Get owner and name from PathWithNamespace because the fields
	// event.Project.Name and event.Project.Owner can have capitals.
	owner, name := e.getOwnerAndName(event.Project.PathWithNamespace)
	baseRepo = models.Repo{
		FullName:          event.Project.PathWithNamespace,
		Name:              name,
		SanitizedCloneURL: event.Project.GitHTTPURL,
		Owner:             owner,
		CloneURL:          e.addGitlabAuth(event.Project.GitHTTPURL),
	}
	user = models.User{
		Username: event.User.Username,
	}
	owner, name = e.getOwnerAndName(event.MergeRequest.Source.PathWithNamespace)
	headRepo = models.Repo{
		FullName:          event.MergeRequest.Source.PathWithNamespace,
		Name:              name,
		SanitizedCloneURL: event.MergeRequest.Source.GitHTTPURL,
		Owner:             owner,
		CloneURL:          e.addGitlabAuth(event.MergeRequest.Source.GitHTTPURL),
	}
	return
}

func (e *EventParser) ParseGitlabMergeRequest(mr *gitlab.MergeRequest) models.PullRequest {
	pullState := models.Closed
	if mr.State == gitlabPullOpened {
		pullState = models.Open
	}
	// GitLab also has a "merged" state, but we map that to Closed so we don't
	// need to check for it.

	return models.PullRequest{
		URL:        mr.WebURL,
		Author:     mr.Author.Username,
		Num:        mr.IID,
		HeadCommit: mr.SHA,
		Branch:     mr.SourceBranch,
		State:      pullState,
	}
}

func (e *EventParser) stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

var HelpComment = "```cmake\n" +
	`atlantis
Terraform automation and collaboration for your team

Usage:
  atlantis <command> [options]

Commands:
  plan   Runs 'terraform plan' for the changes in this pull request.
  apply  Runs 'terraform apply' on the plans generated by 'atlantis plan'.
  help   View help.

Flags:
  -h, --help   help for atlantis

Use "atlantis [command] --help" for more information about a command.
`

var DidYouMeanAtlantisComment = "Did you mean to use `atlantis` instead of `terraform`?"
