// Copyright 2026 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package events

import (
	"encoding/json"
	"errors"
	"testing"

	coreevents "github.com/runatlantis/atlantis/server/events"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/logging"
	"github.com/stretchr/testify/require"
)

func TestHandleCommentEvent_DispatchesCredentialFreeCommand(t *testing.T) {
	baseRepo := routedTestRepo(t, "owner/repo", "base-token")
	headRepo := routedTestRepo(t, "fork/repo", "head-token")
	pull := models.PullRequest{Num: 12, HeadCommit: "abc123", BaseRepo: baseRepo}
	dispatcher := &recordingDispatcher{}
	runner := &recordingLegacyRunner{}
	controller := routedTestController(t, dispatcher, runner, &recordingLegacyPullCleaner{})
	controller.CommentParser = staticCommentParser{result: coreevents.CommentParseResult{
		Command: &coreevents.CommentCommand{Name: command.Apply},
	}}

	response := controller.handleCommentEvent(
		logging.NewNoopLogger(t), baseRepo, &headRepo, &pull, models.User{Username: "alice"},
		12, "atlantis apply", 1, models.Github,
	)

	require.Zero(t, response.err.code)
	require.Equal(t, "Processing...", response.body)
	require.Equal(t, 0, runner.commentCalls)
	require.Len(t, dispatcher.comments, 1)
	payload, err := json.Marshal(dispatcher.comments[0])
	require.NoError(t, err)
	require.NotContains(t, string(payload), "base-token")
	require.NotContains(t, string(payload), "head-token")
	require.Equal(t, 12, dispatcher.comments[0].OwnershipKey().PullNum)
}

func TestHandleCommentEvent_DispatchFailureIsUnavailable(t *testing.T) {
	baseRepo := routedTestRepo(t, "owner/repo", "base-token")
	dispatcher := &recordingDispatcher{commentErr: errors.New("redis unavailable")}
	runner := &recordingLegacyRunner{}
	controller := routedTestController(t, dispatcher, runner, &recordingLegacyPullCleaner{})
	controller.CommentParser = staticCommentParser{result: coreevents.CommentParseResult{
		Command: &coreevents.CommentCommand{Name: command.Plan},
	}}

	response := controller.handleCommentEvent(
		logging.NewNoopLogger(t), baseRepo, nil, nil, models.User{}, 12,
		"atlantis plan", 1, models.Github,
	)

	require.Equal(t, 503, response.err.code)
	require.ErrorContains(t, response.err.err, "redis unavailable")
	require.Equal(t, 0, runner.commentCalls)
}

func TestHandlePullRequestEvent_DispatchesAutoplanAndCleanup(t *testing.T) {
	baseRepo := routedTestRepo(t, "owner/repo", "base-token")
	headRepo := routedTestRepo(t, "fork/repo", "head-token")
	pull := models.PullRequest{Num: 12, HeadCommit: "abc123", BaseRepo: baseRepo}
	dispatcher := &recordingDispatcher{}
	runner := &recordingLegacyRunner{}
	cleaner := &recordingLegacyPullCleaner{}
	controller := routedTestController(t, dispatcher, runner, cleaner)

	response := controller.handlePullRequestEvent(
		logging.NewNoopLogger(t), baseRepo, headRepo, pull, models.User{Username: "alice"}, models.OpenedPullEvent,
	)
	require.Zero(t, response.err.code)
	require.Equal(t, "Processing...", response.body)
	require.Len(t, dispatcher.autoplans, 1)
	require.Equal(t, 0, runner.autoplanCalls)

	response = controller.handlePullRequestEvent(
		logging.NewNoopLogger(t), baseRepo, headRepo, pull, models.User{Username: "alice"}, models.ClosedPullEvent,
	)
	require.Zero(t, response.err.code)
	require.Equal(t, "Pull request cleaned successfully", response.body)
	require.Len(t, dispatcher.closed, 1)
	require.Equal(t, 0, cleaner.calls)
}

func TestHandlePullRequestEvent_DispatchFailureIsUnavailable(t *testing.T) {
	baseRepo := routedTestRepo(t, "owner/repo", "base-token")
	headRepo := routedTestRepo(t, "fork/repo", "head-token")
	pull := models.PullRequest{Num: 12, BaseRepo: baseRepo}
	tests := []struct {
		name      string
		eventType models.PullRequestEventType
		dispatch  *recordingDispatcher
	}{
		{name: "autoplan", eventType: models.OpenedPullEvent, dispatch: &recordingDispatcher{autoplanErr: errors.New("forward failed")}},
		{name: "cleanup", eventType: models.ClosedPullEvent, dispatch: &recordingDispatcher{closedErr: errors.New("forward failed")}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			controller := routedTestController(t, test.dispatch, &recordingLegacyRunner{}, &recordingLegacyPullCleaner{})
			response := controller.handlePullRequestEvent(
				logging.NewNoopLogger(t), baseRepo, headRepo, pull, models.User{}, test.eventType,
			)

			require.Equal(t, 503, response.err.code)
			require.ErrorContains(t, response.err.err, "forward failed")
		})
	}
}

func routedTestController(t *testing.T, dispatcher coreevents.CommandDispatcher, runner coreevents.CommandRunner, cleaner coreevents.PullCleaner) *VCSEventsController {
	t.Helper()
	allowlist, err := coreevents.NewRepoAllowlistChecker("*")
	require.NoError(t, err)
	return &VCSEventsController{
		CommandDispatcher:    dispatcher,
		CommandRunner:        runner,
		PullCleaner:          cleaner,
		RepoAllowlistChecker: allowlist,
		TestingMode:          true,
	}
}

func routedTestRepo(t *testing.T, fullName, token string) models.Repo {
	t.Helper()
	repo, err := models.NewRepo(models.Github, fullName, "https://github.com/"+fullName, "atlantis", token, "")
	require.NoError(t, err)
	return repo
}

type staticCommentParser struct {
	result coreevents.CommentParseResult
}

func (p staticCommentParser) Parse(string, models.VCSHostType) coreevents.CommentParseResult {
	return p.result
}

type recordingDispatcher struct {
	comments    []coreevents.CommentDispatch
	autoplans   []coreevents.AutoplanDispatch
	closed      []coreevents.PullClosedDispatch
	commentErr  error
	autoplanErr error
	closedErr   error
}

func (d *recordingDispatcher) DispatchComment(command coreevents.CommentDispatch) error {
	d.comments = append(d.comments, command)
	return d.commentErr
}
func (d *recordingDispatcher) DispatchAutoplan(command coreevents.AutoplanDispatch) error {
	d.autoplans = append(d.autoplans, command)
	return d.autoplanErr
}
func (d *recordingDispatcher) DispatchPullClosed(command coreevents.PullClosedDispatch) error {
	d.closed = append(d.closed, command)
	return d.closedErr
}

type recordingLegacyRunner struct {
	commentCalls  int
	autoplanCalls int
}

func (r *recordingLegacyRunner) RunCommentCommand(models.Repo, *models.Repo, *models.PullRequest, models.User, int, *coreevents.CommentCommand) {
	r.commentCalls++
}
func (r *recordingLegacyRunner) RunAutoplanCommand(models.Repo, models.Repo, models.PullRequest, models.User) {
	r.autoplanCalls++
}

type recordingLegacyPullCleaner struct {
	calls int
}

func (c *recordingLegacyPullCleaner) CleanUpPull(logging.SimpleLogging, models.Repo, models.PullRequest) error {
	c.calls++
	return nil
}
