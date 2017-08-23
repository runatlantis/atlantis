package server

import (
	"errors"
	"fmt"
	"strings"

	"github.com/google/go-github/github"
	"github.com/hootsuite/atlantis/models"
)

type EventParser struct {
	GithubUser  string
	GithubToken string
}

// DetermineCommand parses the comment as an atlantis command. If it succeeds,
// it returns the command. Otherwise it returns error.
func (e *EventParser) DetermineCommand(comment *github.IssueCommentEvent) (*Command, error) {
	// valid commands contain:
	// the initial "executable" name, 'run' or 'atlantis' or '@GithubUser' where GithubUser is the api user atlantis is running as
	// then a command, either 'plan', 'apply', or 'help'
	// then an optional environment argument, an optional --verbose flag and any other flags
	//
	// examples:
	// atlantis help
	// run plan
	// @GithubUser plan staging
	// atlantis plan staging --verbose
	// atlantis plan staging --verbose -key=value -key2 value2
	commentBody := comment.Comment.GetBody()
	if commentBody == "" {
		return nil, errors.New("comment.body is null")
	}
	err := errors.New("not an Atlantis command")
	args := strings.Fields(commentBody)
	if len(args) < 2 {
		return nil, err
	}

	env := "default"
	verbose := false
	var flags []string

	if !e.stringInSlice(args[0], []string{"run", "atlantis", "@" + e.GithubUser}) {
		return nil, err
	}
	if !e.stringInSlice(args[1], []string{"plan", "apply", "help"}) {
		return nil, err
	}
	if args[1] == "help" {
		return &Command{Name: Help}, nil
	}
	command := args[1]

	if len(args) > 2 {
		flags = args[2:]

		// if the third arg doesn't start with '-' then we assume it's an
		// environment not a flag
		if !strings.HasPrefix(args[2], "-") {
			env = args[2]
			flags = args[3:]
		}

		// check for --verbose specially and then remove any additional
		// occurrences
		if e.stringInSlice("--verbose", flags) {
			verbose = true
			flags = e.removeOccurrences("--verbose", flags)
		}
	}

	c := &Command{Verbose: verbose, Environment: env, Flags: flags}
	switch command {
	case "plan":
		c.Name = Plan
	case "apply":
		c.Name = Apply
	default:
		return nil, fmt.Errorf("something went wrong parsing the command, the command we parsed %q was not apply or plan", command)
	}
	return c, nil
}

func (e *EventParser) ExtractCommentData(comment *github.IssueCommentEvent, ctx *CommandContext) error {
	repo, err := e.ExtractRepoData(comment.Repo)
	if err != nil {
		return err
	}
	pullNum := comment.Issue.GetNumber()
	if pullNum == 0 {
		return errors.New("issue.number' is null")
	}
	pullCreator := comment.Issue.User.GetLogin()
	if pullCreator == "" {
		return errors.New("issue.user.login' is null")
	}
	commentorUsername := comment.Comment.User.GetLogin()
	if commentorUsername == "" {
		return errors.New("comment.user.login is null")
	}
	htmlURL := comment.Issue.GetHTMLURL()
	if htmlURL == "" {
		return errors.New("comment.issue.html_url is null")
	}
	ctx.BaseRepo = repo
	ctx.User = models.User{
		Username: commentorUsername,
	}
	ctx.Pull = models.PullRequest{
		Num: pullNum,
	}
	return nil
}

func (e *EventParser) ExtractPullData(pull *github.PullRequest) (models.PullRequest, models.Repo, error) {
	var pullModel models.PullRequest
	var headRepoModel models.Repo

	commit := pull.Head.GetSHA()
	if commit == "" {
		return pullModel, headRepoModel, errors.New("head.sha is null")
	}
	base := pull.Base.GetSHA()
	if base == "" {
		return pullModel, headRepoModel, errors.New("base.sha is null")
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

	headRepoModel, err := e.ExtractRepoData(pull.Head.Repo)
	if err != nil {
		return pullModel, headRepoModel, err
	}

	return models.PullRequest{
		BaseCommit: base,
		Author:     authorUsername,
		Branch:     branch,
		HeadCommit: commit,
		URL:        url,
		Num:        num,
	}, headRepoModel, nil
}

func (e *EventParser) ExtractRepoData(ghRepo *github.Repository) (models.Repo, error) {
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

	// construct HTTPS repo clone url string with username and password
	repoCloneURL := strings.Replace(repoSanitizedCloneURL, "https://", fmt.Sprintf("https://%s:%s@", e.GithubUser, e.GithubToken), -1)

	return models.Repo{
		Owner:             repoOwner,
		FullName:          repoFullName,
		CloneURL:          repoCloneURL,
		SanitizedCloneURL: repoSanitizedCloneURL,
		Name:              repoName,
	}, nil
}

func (e *EventParser) stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

func (e *EventParser) removeOccurrences(a string, list []string) []string {
	var out []string
	for _, b := range list {
		if b != a {
			out = append(out, b)
		}
	}
	return out
}
