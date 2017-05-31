package main

import (
	"errors"
	"fmt"
	"github.com/google/go-github/github"
	"regexp"
)

type RequestParser struct{}

type CommandType int

const (
	Apply CommandType = iota
	Plan
	Help
)

type Command struct {
	verbose     bool
	environment string
	commandType CommandType
}

func (r *RequestParser) determineCommand(comment *github.IssueCommentEvent) (*Command, error) {
	atlantisCommentRegex := `^(?:run|atlantis) (plan|apply|help)([[:blank:]])?([a-zA-Z0-9_-]+)?\s*(--verbose)?$`
	runPlanMatcher := regexp.MustCompile(atlantisCommentRegex)

	commentComment := comment.Comment
	if commentComment == nil {
		return nil, errors.New("key 'comment.comment' is null")
	}

	commentBody := commentComment.Body
	if commentBody == nil {
		return nil, errors.New("key 'comment.comment.body' is null")
	}

	// extract the command and environment. ex. for "run plan staging", the command is "plan", and the environment is "staging"
	match := runPlanMatcher.FindStringSubmatch(*commentBody)
	if len(match) < 5 {
		var truncated = *commentBody
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

func (r *RequestParser) extractCommentData(comment *github.IssueCommentEvent, params *ExecutionContext) error {
	missingField := "<nil>"

	owner := comment.Repo.Owner.Login
	if owner == nil {
		return errors.New("key 'comment.repo.owner.login' is null")
	}
	repoName := comment.Repo.Name
	if repoName == nil {
		return errors.New("key 'comment.repo.name' is null")
	}
	pullNum := comment.Issue.Number
	if pullNum == nil {
		return errors.New("key 'comment.issue.number' is null")
	}
	pullCreator := comment.Issue.User.Login
	if pullCreator == nil {
		return errors.New("key 'comment.issue.user.login' is null")
	}
	commentorUsername := comment.Comment.User.Login
	if commentorUsername == nil {
		return errors.New("key 'comment.comment.user.login' is null")
	}
	commentorEmail := comment.Comment.User.Email
	if commentorEmail == nil {
		commentorEmail = &missingField
	}
	repoSSHURL := comment.Repo.SSHURL
	if repoSSHURL == nil {
		return errors.New("key 'comment.repo.sshurl' is null")
	}
	htmlURL := comment.Issue.HTMLURL
	if htmlURL == nil {
		return errors.New("key 'comment.issue.htmlUrl' is null")
	}
	params.requesterUsername = *commentorUsername
	params.requesterEmail = *commentorEmail
	params.repoOwner = *owner
	params.pullNum = *pullNum
	params.pullCreator = *pullCreator
	params.repoName = *repoName
	params.repoSSHUrl = *repoSSHURL
	params.htmlUrl = *htmlURL
	return nil
}

func (r *RequestParser) extractPullData(pull *github.PullRequest, params *ExecutionContext) error {
	commit := pull.Head.SHA
	if commit == nil {
		return errors.New("key 'pull.head.sha' is null")
	}
	base := pull.Base.SHA
	if base == nil {
		return errors.New("key 'pull.base.sha' is null")
	}
	pullLink := pull.HTMLURL
	if pullLink == nil {
		return errors.New("key 'pull.html_url' is null")
	}
	branch := pull.Head.Ref
	if branch == nil {
		return errors.New("key 'pull.head.ref' is null")
	}
	params.branch = *branch
	params.head = *commit
	params.base = *base
	params.pullLink = *pullLink
	return nil
}
