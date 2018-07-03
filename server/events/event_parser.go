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
	"path"
	"regexp"
	"strings"

	"github.com/google/go-github/github"
	"github.com/lkysow/go-gitlab"
	"github.com/pkg/errors"
	"github.com/runatlantis/atlantis/server/events/models"
)

const gitlabPullOpened = "opened"
const usagesCols = 90

// multiLineRegex is used to ignore multi-line comments since those aren't valid
// Atlantis commands.
var multiLineRegex = regexp.MustCompile(`.*\r?\n.+`)

//go:generate pegomock generate -m --use-experimental-model-gen --package mocks -o mocks/mock_event_parsing.go EventParsing

type CommandInterface interface {
	CommandName() CommandName
	IsVerbose() bool
	IsAutoplan() bool
}

type AutoplanCommand struct{}

func (c AutoplanCommand) CommandName() CommandName {
	return Plan
}

func (c AutoplanCommand) IsVerbose() bool {
	return false
}

func (c AutoplanCommand) IsAutoplan() bool {
	return true
}

type CommentCommand struct {
	// RepoRelDir is the path relative to the repo root to run the command in.
	// Will never be an empty string and will never end in "/".
	RepoRelDir string
	// CommentArgs are the extra arguments appended to comment,
	// ex. atlantis plan -- -target=resource
	Flags     []string
	Name      CommandName
	Verbose   bool
	Workspace string
	// ProjectName is the name of a project to run the command on. It refers to a
	// project specified in an atlantis.yaml file.
	ProjectName string
}

func (c CommentCommand) CommandName() CommandName {
	return c.Name
}

func (c CommentCommand) IsVerbose() bool {
	return c.Verbose
}

func (c CommentCommand) IsAutoplan() bool {
	return false
}

func (c CommentCommand) String() string {
	return fmt.Sprintf("command=%q verbose=%t dir=%q workspace=%q project=%q flags=%q", c.Name.String(), c.Verbose, c.RepoRelDir, c.Workspace, c.ProjectName, strings.Join(c.Flags, ","))
}

// NewCommentCommand constructs a CommentCommand, setting all missing fields to defaults.
func NewCommentCommand(repoRelDir string, flags []string, name CommandName, verbose bool, workspace string, project string) *CommentCommand {
	// If repoRelDir was an empty string, this will return '.'.
	validDir := path.Clean(repoRelDir)
	if validDir == "/" {
		validDir = "."
	}
	if workspace == "" {
		workspace = DefaultWorkspace
	}
	return &CommentCommand{
		RepoRelDir:  validDir,
		Flags:       flags,
		Name:        name,
		Verbose:     verbose,
		Workspace:   workspace,
		ProjectName: project,
	}
}

type EventParsing interface {
	ParseGithubIssueCommentEvent(comment *github.IssueCommentEvent) (baseRepo models.Repo, user models.User, pullNum int, err error)
	// ParseGithubPull returns the pull request, base repo and head repo.
	ParseGithubPull(pull *github.PullRequest) (models.PullRequest, models.Repo, models.Repo, error)
	// ParseGithubPullEvent returns the pull request, head repo and user that
	// caused the event. Base repo is available as a field on PullRequest.
	ParseGithubPullEvent(pullEvent *github.PullRequestEvent) (pull models.PullRequest, baseRepo models.Repo, headRepo models.Repo, user models.User, err error)
	ParseGithubRepo(ghRepo *github.Repository) (models.Repo, error)
	// ParseGitlabMergeEvent returns the pull request, base repo, head repo and
	// user that caused the event.
	ParseGitlabMergeEvent(event gitlab.MergeEvent) (models.PullRequest, models.Repo, models.Repo, models.User, error)
	ParseGitlabMergeCommentEvent(event gitlab.MergeCommentEvent) (baseRepo models.Repo, headRepo models.Repo, user models.User, err error)
	ParseGitlabMergeRequest(mr *gitlab.MergeRequest, baseRepo models.Repo) models.PullRequest
}

type EventParser struct {
	GithubUser  string
	GithubToken string
	GitlabUser  string
	GitlabToken string
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

func (e *EventParser) ParseGithubPullEvent(pullEvent *github.PullRequestEvent) (models.PullRequest, models.Repo, models.Repo, models.User, error) {
	if pullEvent.PullRequest == nil {
		return models.PullRequest{}, models.Repo{}, models.Repo{}, models.User{}, errors.New("pull_request is null")
	}
	pull, baseRepo, headRepo, err := e.ParseGithubPull(pullEvent.PullRequest)
	if err != nil {
		return models.PullRequest{}, models.Repo{}, models.Repo{}, models.User{}, err
	}
	if pullEvent.Sender == nil {
		return models.PullRequest{}, models.Repo{}, models.Repo{}, models.User{}, errors.New("sender is null")
	}
	senderUsername := pullEvent.Sender.GetLogin()
	if senderUsername == "" {
		return models.PullRequest{}, models.Repo{}, models.Repo{}, models.User{}, errors.New("sender.login is null")
	}
	return pull, baseRepo, headRepo, models.User{Username: senderUsername}, nil
}

func (e *EventParser) ParseGithubPull(pull *github.PullRequest) (pullModel models.PullRequest, baseRepo models.Repo, headRepo models.Repo, err error) {
	commit := pull.Head.GetSHA()
	if commit == "" {
		err = errors.New("head.sha is null")
		return
	}
	url := pull.GetHTMLURL()
	if url == "" {
		err = errors.New("html_url is null")
		return
	}
	branch := pull.Head.GetRef()
	if branch == "" {
		err = errors.New("head.ref is null")
		return
	}
	authorUsername := pull.User.GetLogin()
	if authorUsername == "" {
		err = errors.New("user.login is null")
		return
	}
	num := pull.GetNumber()
	if num == 0 {
		err = errors.New("number is null")
		return
	}

	baseRepo, err = e.ParseGithubRepo(pull.Base.Repo)
	if err != nil {
		return
	}
	headRepo, err = e.ParseGithubRepo(pull.Head.Repo)
	if err != nil {
		return
	}

	pullState := models.Closed
	if pull.GetState() == "open" {
		pullState = models.Open
	}

	pullModel = models.PullRequest{
		Author:     authorUsername,
		Branch:     branch,
		HeadCommit: commit,
		URL:        url,
		Num:        num,
		State:      pullState,
		BaseRepo:   baseRepo,
	}
	return
}

func (e *EventParser) ParseGithubRepo(ghRepo *github.Repository) (models.Repo, error) {
	return models.NewRepo(models.Github, ghRepo.GetFullName(), ghRepo.GetCloneURL(), e.GithubUser, e.GithubToken)
}

func (e *EventParser) ParseGitlabMergeEvent(event gitlab.MergeEvent) (models.PullRequest, models.Repo, models.Repo, models.User, error) {
	modelState := models.Closed
	if event.ObjectAttributes.State == gitlabPullOpened {
		modelState = models.Open
	}
	// GitLab also has a "merged" state, but we map that to Closed so we don't
	// need to check for it.

	baseRepo, err := models.NewRepo(models.Gitlab, event.Project.PathWithNamespace, event.Project.GitHTTPURL, e.GitlabUser, e.GitlabToken)
	if err != nil {
		return models.PullRequest{}, models.Repo{}, models.Repo{}, models.User{}, err
	}
	headRepo, err := models.NewRepo(models.Gitlab, event.ObjectAttributes.Source.PathWithNamespace, event.ObjectAttributes.Source.GitHTTPURL, e.GitlabUser, e.GitlabToken)
	if err != nil {
		return models.PullRequest{}, models.Repo{}, models.Repo{}, models.User{}, err
	}

	pull := models.PullRequest{
		URL:        event.ObjectAttributes.URL,
		Author:     event.User.Username,
		Num:        event.ObjectAttributes.IID,
		HeadCommit: event.ObjectAttributes.LastCommit.ID,
		Branch:     event.ObjectAttributes.SourceBranch,
		State:      modelState,
		BaseRepo:   baseRepo,
	}

	user := models.User{
		Username: event.User.Username,
	}

	return pull, baseRepo, headRepo, user, err
}

// ParseGitlabMergeCommentEvent creates Atlantis models out of a GitLab event.
func (e *EventParser) ParseGitlabMergeCommentEvent(event gitlab.MergeCommentEvent) (baseRepo models.Repo, headRepo models.Repo, user models.User, err error) {
	// Parse the base repo first.
	repoFullName := event.Project.PathWithNamespace
	cloneURL := event.Project.GitHTTPURL
	baseRepo, err = models.NewRepo(models.Gitlab, repoFullName, cloneURL, e.GitlabUser, e.GitlabToken)
	if err != nil {
		return
	}
	user = models.User{
		Username: event.User.Username,
	}

	// Now parse the head repo.
	headRepoFullName := event.MergeRequest.Source.PathWithNamespace
	headCloneURL := event.MergeRequest.Source.GitHTTPURL
	headRepo, err = models.NewRepo(models.Gitlab, headRepoFullName, headCloneURL, e.GitlabUser, e.GitlabToken)
	return
}

// ParseGitlabMergeRequest parses the merge requests and returns a pull request
// model. We require passing in baseRepo because although can't get this information
// from the merge request, the only caller of this function already has that
// data. This means we can construct the pull request object correctly.
func (e *EventParser) ParseGitlabMergeRequest(mr *gitlab.MergeRequest, baseRepo models.Repo) models.PullRequest {
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
		BaseRepo:   baseRepo,
	}
}
