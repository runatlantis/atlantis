package server

import (
	"errors"
	"fmt"
	"github.com/google/go-github/github"
	"github.com/hootsuite/atlantis/models"
	"regexp"
)

type EventParser struct{}

func (e *EventParser) DetermineCommand(comment *github.IssueCommentEvent) (*Command, error) {
	// for legacy, also support "run" instead of atlantis
	atlantisCommentRegex := `^(?:run|atlantis) (plan|apply|help)([[:blank:]])?([a-zA-Z0-9_-]+)?\s*(--verbose)?$`
	runPlanMatcher := regexp.MustCompile(atlantisCommentRegex)

	commentBody := comment.Comment.GetBody()
	if commentBody == "" {
		return nil, errors.New("comment.body is null")
	}

	// extract the command and environment. ex. for "atlantis plan staging", the command is "plan", and the environment is "staging"
	match := runPlanMatcher.FindStringSubmatch(commentBody)
	if len(match) < 5 {
		var truncated = commentBody
		if len(truncated) > 30 {
			truncated = truncated[0:30] + "..."
		}
		return nil, errors.New("not an Atlantis command")
	}

	verbose := false
	if match[4] == "--verbose" {
		verbose = true
	}
	command := &Command{verbose: verbose, environment: match[3]}
	switch match[1] {
	case "plan":
		command.commandType = Plan
	case "apply":
		command.commandType = Apply
	case "help":
		command.commandType = Help
	default:
		return nil, fmt.Errorf("something went wrong with our regex, the command we parsed %q was not apply or plan", match[1])
	}
	return command, nil
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
	ctx.Repo = repo
	ctx.User = models.User{
		Username: commentorUsername,
	}
	ctx.Pull = models.PullRequest{
		Num: pullNum,
	}
	return nil
}

func (e *EventParser) ExtractPullData(pull *github.PullRequest) (models.PullRequest, error) {
	var pullModel models.PullRequest
	commit := pull.Head.GetSHA()
	if commit == "" {
		return pullModel, errors.New("head.sha is null")
	}
	base := pull.Base.GetSHA()
	if base == "" {
		return pullModel, errors.New("base.sha is null")
	}
	url := pull.GetHTMLURL()
	if url == "" {
		return pullModel, errors.New("html_url is null")
	}
	branch := pull.Head.GetRef()
	if branch == "" {
		return pullModel, errors.New("head.ref is null")
	}
	authorUsername := pull.User.GetLogin()
	if authorUsername == "" {
		return pullModel, errors.New("user.login is null")
	}
	num := pull.GetNumber()
	if num == 0 {
		return pullModel, errors.New("number is null")
	}
	return models.PullRequest{
		BaseCommit: base,
		Author:     authorUsername,
		Branch:     branch,
		HeadCommit: commit,
		URL:        url,
		Num:        num,
	}, nil
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
		Owner: repoOwner,
		FullName: repoFullName,
		SSHURL: repoSSHURL,
		Name: repoName,
	}, nil
}
