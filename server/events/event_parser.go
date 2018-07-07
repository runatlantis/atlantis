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
	"strings"

	"github.com/google/go-github/github"
	"github.com/lkysow/go-gitlab"
	"github.com/pkg/errors"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/events/vcs/bitbucketcloud"
	"github.com/runatlantis/atlantis/server/events/vcs/bitbucketserver"
	"gopkg.in/go-playground/validator.v9"
)

const gitlabPullOpened = "opened"
const usagesCols = 90

type CommandInterface interface {
	CommandName() CommandName
	IsVerbose() bool
	IsAutoplan() bool
}

type AutoplanCommand struct{}

func (c AutoplanCommand) CommandName() CommandName {
	return PlanCommand
}

func (c AutoplanCommand) IsVerbose() bool {
	return false
}

func (c AutoplanCommand) IsAutoplan() bool {
	return true
}

type CommentCommand struct {
	// RepoRelDir is the path relative to the repo root to run the command in.
	// Will never end in "/". If empty then the comment specified no directory.
	RepoRelDir string
	// CommentArgs are the extra arguments appended to comment,
	// ex. atlantis plan -- -target=resource
	Flags   []string
	Name    CommandName
	Verbose bool
	// Workspace is the name of the Terraform workspace to run the command in.
	// If empty then the comment specified no workspace.
	Workspace string
	// ProjectName is the name of a project to run the command on. It refers to a
	// project specified in an atlantis.yaml file.
	// If empty then the comment specified no project.
	ProjectName string
}

// IsForSpecificProject returns true if the command is for a specific dir, workspace
// or project name. Otherwise it's a command like "atlantis plan" or "atlantis
// apply".
func (c CommentCommand) IsForSpecificProject() bool {
	return c.RepoRelDir != "" || c.Workspace != "" || c.ProjectName != ""
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
	// If repoRelDir was empty we want to keep it that way to indicate that it
	// wasn't specified in the comment.
	if repoRelDir != "" {
		repoRelDir = path.Clean(repoRelDir)
		if repoRelDir == "/" {
			repoRelDir = "."
		}
	}
	return &CommentCommand{
		RepoRelDir:  repoRelDir,
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
	GetBitbucketCloudEventType(eventTypeHeader string) models.PullRequestEventType
	ParseBitbucketServerPullEvent(body []byte) (pull models.PullRequest, baseRepo models.Repo, headRepo models.Repo, user models.User, err error)
	ParseBitbucketServerCommentEvent(body []byte) (pull models.PullRequest, baseRepo models.Repo, headRepo models.Repo, user models.User, comment string, err error)
	GetBitbucketServerEventType(eventTypeHeader string) models.PullRequestEventType
}

type EventParser struct {
	GithubUser         string
	GithubToken        string
	GitlabUser         string
	GitlabToken        string
	BitbucketUser      string
	BitbucketToken     string
	BitbucketServerURL string
}

// GetBitbucketCloudEventType translates the bitbucket header name into a pull
// request event type.
func (e *EventParser) GetBitbucketCloudEventType(eventTypeHeader string) models.PullRequestEventType {
	switch eventTypeHeader {
	case bitbucketcloud.PullCreatedHeader:
		return models.OpenedPullEvent
	case bitbucketcloud.PullUpdatedHeader:
		return models.UpdatedPullEvent
	case bitbucketcloud.PullFulfilledHeader, bitbucketcloud.PullRejectedHeader:
		return models.ClosedPullEvent
	}
	return models.OtherPullEvent
}

func (e *EventParser) ParseBitbucketCloudCommentEvent(body []byte) (pull models.PullRequest, baseRepo models.Repo, headRepo models.Repo, user models.User, comment string, err error) {
	var event bitbucketcloud.CommentEvent
	if err = json.Unmarshal(body, &event); err != nil {
		err = errors.Wrap(err, "parsing json")
		return
	}
	if err = validator.New().Struct(event); err != nil {
		err = errors.Wrapf(err, "API response %q was missing fields", string(body))
		return
	}
	pull, baseRepo, headRepo, user, err = e.parseCommonBitbucketCloudEventData(event.CommonEventData)
	comment = *event.Comment.Content.Raw
	return
}

func (e *EventParser) parseCommonBitbucketCloudEventData(event bitbucketcloud.CommonEventData) (pull models.PullRequest, baseRepo models.Repo, headRepo models.Repo, user models.User, err error) {
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
		models.BitbucketCloud,
		*event.PullRequest.Source.Repository.FullName,
		*event.PullRequest.Source.Repository.Links.HTML.HREF,
		e.BitbucketUser,
		e.BitbucketToken)
	if err != nil {
		return
	}
	baseRepo, err = models.NewRepo(
		models.BitbucketCloud,
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
	var event bitbucketcloud.PullRequestEvent
	if err = json.Unmarshal(body, &event); err != nil {
		err = errors.Wrap(err, "parsing json")
		return
	}
	if err = validator.New().Struct(event); err != nil {
		err = errors.Wrapf(err, "API response %q was missing fields", string(body))
		return
	}
	pull, baseRepo, headRepo, user, err = e.parseCommonBitbucketCloudEventData(event.CommonEventData)
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

func (e *EventParser) GetBitbucketServerEventType(eventTypeHeader string) models.PullRequestEventType {
	switch eventTypeHeader {
	case bitbucketserver.PullCreatedHeader:
		return models.OpenedPullEvent
	case bitbucketserver.PullMergedHeader, bitbucketserver.PullDeclinedHeader:
		return models.ClosedPullEvent
	}
	return models.OtherPullEvent
}

func (e *EventParser) ParseBitbucketServerCommentEvent(body []byte) (pull models.PullRequest, baseRepo models.Repo, headRepo models.Repo, user models.User, comment string, err error) {
	var event bitbucketserver.CommentEvent
	if err = json.Unmarshal(body, &event); err != nil {
		err = errors.Wrap(err, "parsing json")
		return
	}
	if err = validator.New().Struct(event); err != nil {
		err = errors.Wrapf(err, "API response %q was missing fields", string(body))
		return
	}
	pull, baseRepo, headRepo, user, err = e.parseCommonBitbucketServerEventData(event.CommonEventData)
	comment = *event.Comment.Text
	return
}

func (e *EventParser) parseCommonBitbucketServerEventData(event bitbucketserver.CommonEventData) (pull models.PullRequest, baseRepo models.Repo, headRepo models.Repo, user models.User, err error) {
	var prState models.PullRequestState
	switch *event.PullRequest.State {
	case "OPEN":
		prState = models.OpenPullState
	case "MERGED":
		prState = models.ClosedPullState
	case "DECLINED":
		prState = models.ClosedPullState
	default:
		err = fmt.Errorf("unable to determine pull request state from %q, this is a bug!", *event.PullRequest.State)
		return
	}

	headRepoSlug := *event.PullRequest.FromRef.Repository.Slug
	headRepoFullname := fmt.Sprintf("%s/%s", *event.PullRequest.FromRef.Repository.Project.Name, headRepoSlug)
	headRepoCloneURL := fmt.Sprintf("%s/scm/%s/%s.git", e.BitbucketServerURL, strings.ToLower(*event.PullRequest.FromRef.Repository.Project.Key), headRepoSlug)
	headRepo, err = models.NewRepo(
		models.BitbucketServer,
		headRepoFullname,
		headRepoCloneURL,
		e.BitbucketUser,
		e.BitbucketToken)
	if err != nil {
		return
	}

	baseRepoSlug := *event.PullRequest.ToRef.Repository.Slug
	baseRepoFullname := fmt.Sprintf("%s/%s", *event.PullRequest.ToRef.Repository.Project.Name, baseRepoSlug)
	baseRepoCloneURL := fmt.Sprintf("%s/scm/%s/%s.git", e.BitbucketServerURL, strings.ToLower(*event.PullRequest.ToRef.Repository.Project.Key), baseRepoSlug)
	baseRepo, err = models.NewRepo(
		models.BitbucketServer,
		baseRepoFullname,
		baseRepoCloneURL,
		e.BitbucketUser,
		e.BitbucketToken)
	if err != nil {
		return
	}

	pull = models.PullRequest{
		Num:        *event.PullRequest.ID,
		HeadCommit: *event.PullRequest.FromRef.LatestCommit,
		URL:        fmt.Sprintf("%s/projects/%s/repos/%s/pull-requests/%d", e.BitbucketServerURL, *event.PullRequest.ToRef.Repository.Project.Key, *event.PullRequest.ToRef.Repository.Slug, *event.PullRequest.ID),
		Branch:     *event.PullRequest.FromRef.DisplayID,
		Author:     *event.Actor.Username,
		State:      prState,
		BaseRepo:   baseRepo,
	}
	user = models.User{
		Username: *event.Actor.Username,
	}
	return
}

func (e *EventParser) ParseBitbucketServerPullEvent(body []byte) (pull models.PullRequest, baseRepo models.Repo, headRepo models.Repo, user models.User, err error) {
	var event bitbucketserver.PullRequestEvent
	if err = json.Unmarshal(body, &event); err != nil {
		err = errors.Wrap(err, "parsing json")
		return
	}
	if err = validator.New().Struct(event); err != nil {
		err = errors.Wrapf(err, "API response %q was missing fields", string(body))
		return
	}
	pull, baseRepo, headRepo, user, err = e.parseCommonBitbucketServerEventData(event.CommonEventData)
	return
}
