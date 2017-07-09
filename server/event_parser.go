package server

import (
	"errors"
	"fmt"
	"regexp"

	"github.com/google/go-github/github"
	"github.com/hootsuite/atlantis/models"
)

type EventParser struct {
	GithubUser string
}

func (e *EventParser) DetermineCommand(comment *github.IssueCommentEvent) (*Command, error) {
	// regex matches:
	// the initial "executable" name, 'run' or 'atlantis' or '@GithubUser' where GithubUser is the api user atlantis is running as
	// then a command, either 'plan', 'apply', or 'help'
	// then an optional environment and an optional --verbose flag
	//
	// examples:
	// atlantis help
	// run plan
	// @GithubUser plan staging
	// atlantis plan staging --verbose
	//
	// capture groups:
	// 1: the command, i.e. plan/apply/help
	// 2: the environment OR the --verbose flag (if they didn't specify and environment)
	// 3: the --verbose flag (if they specified an environment)
	atlantisCommentRegex := `^(?:run|atlantis|@` + e.GithubUser + `)[[:blank:]]+(plan|apply|help)(?:[[:blank:]]+([a-zA-Z0-9_-]+))?[[:blank:]]*(--verbose)?$`
	runPlanMatcher := regexp.MustCompile(atlantisCommentRegex)

	commentBody := comment.Comment.GetBody()
	if commentBody == "" {
		return nil, errors.New("comment.body is null")
	}

	// extract the command and environment. ex. for "atlantis plan staging", the command is "plan", and the environment is "staging"
	match := runPlanMatcher.FindStringSubmatch(commentBody)
	if len(match) < 4 {
		var truncated = commentBody
		if len(truncated) > 30 {
			truncated = truncated[0:30] + "..."
		}
		return nil, errors.New("not an Atlantis command")
	}

	// depending on the comment, the command/env/verbose may be in different matching groups
	// if there is no env (ex. just atlantis plan --verbose) then, verbose would be in the 2nd group
	// if there is an env, then verbose would be in the 3rd
	command := match[1]
	env := match[2]
	verboseFlag := match[3]
	if verboseFlag == "" && env == "--verbose" {
		verboseFlag = env
		env = ""
	}

	// now we're ready to actually look at the values
	verbose := false
	if verboseFlag == "--verbose" {
		verbose = true
	}
	// if env not specified, use terraform's default
	if env == "" {
		env = "default"
	}

	c := &Command{Verbose: verbose, Environment: env}
	switch command {
	case "plan":
		c.Name = Plan
	case "apply":
		c.Name = Apply
	case "help":
		c.Name = Help
	default:
		return nil, fmt.Errorf("something went wrong with our regex, the command we parsed %q was not apply or plan", match[1])
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
	repoSSHURL := ghRepo.GetSSHURL()
	if repoSSHURL == "" {
		return repo, errors.New("repository.ssh_url is null")
	}
	return models.Repo{
		Owner:    repoOwner,
		FullName: repoFullName,
		SSHURL:   repoSSHURL,
		Name:     repoName,
	}, nil
}
