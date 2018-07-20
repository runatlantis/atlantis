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
	"encoding/json"
	"fmt"
	"path"
	"regexp"
	"strings"

	"github.com/google/go-github/github"
	"github.com/lkysow/go-gitlab"
	"github.com/pkg/errors"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/events/vcs/bitbucket"
	"gopkg.in/go-playground/validator.v9"
)

const gitlabPullOpened = "opened"
const usagesCols = 90

// multiLineRegex is used to ignore multi-line comments since those aren't valid
// Atlantis commands.
var multiLineRegex = regexp.MustCompile(`.*\r?\n.+`)

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

//go:generate pegomock generate -m --use-experimental-model-gen --package mocks -o mocks/mock_event_parsing.go EventParsing

type EventParsing interface {
	ParseGithubIssueCommentEvent(comment *github.IssueCommentEvent) (baseRepo models.Repo, user models.User, pullNum int, err error)
	// ParseGithubPull returns the pull request, base repo and head repo.
	ParseGithubPull(pull *github.PullRequest) (models.PullRequest, models.Repo, models.Repo, error)
	// ParseGithubPullEvent returns the pull request, head repo and user that
	// caused the event. Base repo is available as a field on PullRequest.
	ParseGithubPullEvent(pullEvent *github.PullRequestEvent) (pull models.PullRequest, pullEventType models.PullRequestEventType, baseRepo models.Repo, headRepo models.Repo, user models.User, err error)
	ParseGithubRepo(ghRepo *github.Repository) (models.Repo, error)
	// ParseGitlabMergeEvent returns the pull request, base repo, head repo and
	// user that caused the event.
	ParseGitlabMergeEvent(event gitlab.MergeEvent) (pull models.PullRequest, pullEventType models.PullRequestEventType, baseRepo models.Repo, headRepo models.Repo, user models.User, err error)
	ParseGitlabMergeCommentEvent(event gitlab.MergeCommentEvent) (baseRepo models.Repo, headRepo models.Repo, user models.User, err error)
	ParseGitlabMergeRequest(mr *gitlab.MergeRequest, baseRepo models.Repo) models.PullRequest
	ParseBitbucketCloudPullEvent(body []byte) (pull models.PullRequest, baseRepo models.Repo, headRepo models.Repo, user models.User, err error)
	ParseBitbucketCloudCommentEvent(body []byte) (pull models.PullRequest, baseRepo models.Repo, headRepo models.Repo, user models.User, comment string, err error)
	GetBitbucketEventType(eventTypeHeader string) models.PullRequestEventType
}

type EventParser struct {
	GithubUser     string
	GithubToken    string
	GitlabUser     string
	GitlabToken    string
	BitbucketUser  string
	BitbucketToken string
}

// GetBitbucketEventType translates the bitbucket header name into a pull
// request event type.
func (e *EventParser) GetBitbucketEventType(eventTypeHeader string) models.PullRequestEventType {
	switch eventTypeHeader {
	case "pullrequest:created":
		return models.OpenedPullEvent
	case "pullrequest:updated":
		return models.UpdatedPullEvent
	case "pullrequest:fulfilled", "pullrequest:rejected":
		return models.ClosedPullEvent
	}
	return models.OtherPullEvent
}

func (e *EventParser) ParseBitbucketCloudCommentEvent(body []byte) (pull models.PullRequest, baseRepo models.Repo, headRepo models.Repo, user models.User, comment string, err error) {
	var event bitbucket.CommentEvent
	if err = json.Unmarshal(body, &event); err != nil {
		err = errors.Wrap(err, "parsing json")
		return
	}
	if err = validator.New().Struct(event); err != nil {
		return
	}
	pull, baseRepo, headRepo, user, err = e.parseCommonBitbucketEventData(event.CommonEventData)
	comment = *event.Comment.Content.Raw
	return
}

func (e *EventParser) parseCommonBitbucketEventData(event bitbucket.CommonEventData) (pull models.PullRequest, baseRepo models.Repo, headRepo models.Repo, user models.User, err error) {
	var prState models.PullRequestState
	switch *event.PullRequest.State {
	case "OPEN":
		prState = models.OpenPullState
	case "MERGED":
		prState = models.ClosedPullState
	case "SUPERSEDED":
		prState = models.ClosedPullState
	case "DECLINE":
		prState = models.ClosedPullState
	default:
		err = fmt.Errorf("unable to determine pull request state from %q, this is a bug!", *event.PullRequest.State)
		return
	}

	headRepo, err = models.NewRepo(
		models.Bitbucket,
		*event.PullRequest.Source.Repository.FullName,
		*event.PullRequest.Source.Repository.Links.HTML.HREF,
		e.BitbucketUser,
		e.BitbucketToken)
	if err != nil {
		return
	}
	baseRepo, err = models.NewRepo(
		models.Bitbucket,
		*event.Repository.FullName,
		*event.Repository.Links.HTML.HREF,
		e.BitbucketUser,
		e.BitbucketToken)
	if err != nil {
		return
	}

	pull = models.PullRequest{
		Num:        *event.PullRequest.ID,
		HeadCommit: *event.PullRequest.Source.Commit.Hash,
		URL:        *event.PullRequest.Links.HTML.HREF,
		Branch:     *event.PullRequest.Source.Branch.Name,
		Author:     *event.Actor.Username,
		State:      prState,
		BaseRepo:   baseRepo,
	}
	user = models.User{
		Username: *event.Actor.Username,
	}
	return
}

func (e *EventParser) ParseBitbucketCloudPullEvent(body []byte) (pull models.PullRequest, baseRepo models.Repo, headRepo models.Repo, user models.User, err error) {
	var event bitbucket.PullRequestEvent
	if err = json.Unmarshal(body, &event); err != nil {
		err = errors.Wrap(err, "parsing json")
		return
	}
	if err = validator.New().Struct(event); err != nil {
		return
	}
	pull, baseRepo, headRepo, user, err = e.parseCommonBitbucketEventData(event.CommonEventData)
	return
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

func (e *EventParser) ParseGithubPullEvent(pullEvent *github.PullRequestEvent) (pull models.PullRequest, pullEventType models.PullRequestEventType, baseRepo models.Repo, headRepo models.Repo, user models.User, err error) {
	if pullEvent.PullRequest == nil {
		err = errors.New("pull_request is null")
		return
	}
	pull, baseRepo, headRepo, err = e.ParseGithubPull(pullEvent.PullRequest)
	if err != nil {
		return
	}
	if pullEvent.Sender == nil {
		err = errors.New("sender is null")
		return
	}
	senderUsername := pullEvent.Sender.GetLogin()
	if senderUsername == "" {
		err = errors.New("sender.login is null")
		return
	}
	switch pullEvent.GetAction() {
	case "opened":
		pullEventType = models.OpenedPullEvent
	case "synchronize":
		pullEventType = models.UpdatedPullEvent
	case "closed":
		pullEventType = models.ClosedPullEvent
	default:
		pullEventType = models.OtherPullEvent
	}
	user = models.User{Username: senderUsername}
	return
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

	pullState := models.ClosedPullState
	if pull.GetState() == "open" {
		pullState = models.OpenPullState
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

func (e *EventParser) ParseGitlabMergeEvent(event gitlab.MergeEvent) (pull models.PullRequest, eventType models.PullRequestEventType, baseRepo models.Repo, headRepo models.Repo, user models.User, err error) {
	modelState := models.ClosedPullState
	if event.ObjectAttributes.State == gitlabPullOpened {
		modelState = models.OpenPullState
	}
	// GitLab also has a "merged" state, but we map that to Closed so we don't
	// need to check for it.

	baseRepo, err = models.NewRepo(models.Gitlab, event.Project.PathWithNamespace, event.Project.GitHTTPURL, e.GitlabUser, e.GitlabToken)
	if err != nil {
		return
	}
	headRepo, err = models.NewRepo(models.Gitlab, event.ObjectAttributes.Source.PathWithNamespace, event.ObjectAttributes.Source.GitHTTPURL, e.GitlabUser, e.GitlabToken)
	if err != nil {
		return
	}

	pull = models.PullRequest{
		URL:        event.ObjectAttributes.URL,
		Author:     event.User.Username,
		Num:        event.ObjectAttributes.IID,
		HeadCommit: event.ObjectAttributes.LastCommit.ID,
		Branch:     event.ObjectAttributes.SourceBranch,
		State:      modelState,
		BaseRepo:   baseRepo,
	}

	switch event.ObjectAttributes.Action {
	case "open":
		eventType = models.OpenedPullEvent
	case "update":
		eventType = models.UpdatedPullEvent
	case "merge", "close":
		eventType = models.ClosedPullEvent
	default:
		eventType = models.OtherPullEvent
	}

	user = models.User{
		Username: event.User.Username,
	}

	return
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
	pullState := models.ClosedPullState
	if mr.State == gitlabPullOpened {
		pullState = models.OpenPullState
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
