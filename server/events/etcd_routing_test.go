// Copyright 2017 HootSuite Media Inc.
// SPDX-License-Identifier: Apache-2.0
// Modified hereafter by contributors to runatlantis/atlantis.

package events_test

import (
	"context"
	"encoding/json"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/runatlantis/atlantis/server/core/etcd"
	"github.com/runatlantis/atlantis/server/events"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/events/vcs"
	"github.com/runatlantis/atlantis/server/logging"
	. "github.com/runatlantis/atlantis/testing"
)

// fakeCoordinator is a hand-written EtcdCoordinator double. Route returns a
// configured result and records the last command; Execute records the command,
// optionally invokes run, and returns a configured outcome.
type fakeCoordinator struct {
	mu sync.Mutex

	routeResult etcd.Result
	routeErr    error
	routed      *etcd.Command

	execOutcome etcd.ExecuteOutcome
	execRun     bool // whether Execute should invoke run
	executed    *etcd.Command
	ran         bool
	reopened    int
	reopenGen   int64 // lifecycle generation ReopenPull returns
	done        chan struct{}
}

func newFakeCoordinator() *fakeCoordinator {
	return &fakeCoordinator{
		routeResult: etcd.Result{Status: http.StatusAccepted},
		execOutcome: etcd.ExecuteRan,
		execRun:     true,
		done:        make(chan struct{}, 1),
	}
}

func (f *fakeCoordinator) Route(_ context.Context, cmd etcd.Command) (etcd.Result, error) {
	f.mu.Lock()
	c := cmd
	f.routed = &c
	f.mu.Unlock()
	return f.routeResult, f.routeErr
}

func (f *fakeCoordinator) Execute(cmd etcd.Command, run func() bool) etcd.ExecuteOutcome {
	f.mu.Lock()
	c := cmd
	f.executed = &c
	f.mu.Unlock()
	if f.execRun {
		f.ran = run()
	}
	select {
	case f.done <- struct{}{}:
	default:
	}
	return f.execOutcome
}

func (f *fakeCoordinator) ReopenPull(_ context.Context, _, _ string, _ int) (int64, error) {
	f.mu.Lock()
	f.reopened++
	g := f.reopenGen
	f.mu.Unlock()
	return g, nil
}

func (f *fakeCoordinator) waitExecuted(t *testing.T) {
	t.Helper()
	select {
	case <-f.done:
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for coordinator.Execute")
	}
}

// recordingRunner records the delegate CommandRunner calls.
type recordingRunner struct {
	mu       sync.Mutex
	comments []*events.CommentCommand
	autoplan int
	lastPull models.PullRequest
}

func (r *recordingRunner) RunCommentCommand(_ models.Repo, _ *models.Repo, _ *models.PullRequest, _ models.User, _ int, cmd *events.CommentCommand) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.comments = append(r.comments, cmd)
}

func (r *recordingRunner) RunAutoplanCommand(_ models.Repo, _ models.Repo, pull models.PullRequest, _ models.User) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.autoplan++
	r.lastPull = pull
}

// commentRecordingVCS embeds vcs.Client and only overrides CreateComment.
type commentRecordingVCS struct {
	vcs.Client
	mu       sync.Mutex
	comments []string
}

func (c *commentRecordingVCS) CreateComment(_ logging.SimpleLogging, _ models.Repo, _ int, comment string, _ string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.comments = append(c.comments, comment)
	return nil
}

func (c *commentRecordingVCS) commentCount() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return len(c.comments)
}

func testRepo() models.Repo {
	return models.Repo{
		FullName: "owner/repo",
		VCSHost:  models.VCSHost{Hostname: "github.com", Type: models.Github},
	}
}

func TestEtcdCommandRouter_IngressComment_LocalAdmit(t *testing.T) {
	coord := newFakeCoordinator()
	delegate := &recordingRunner{}
	vcsClient := &commentRecordingVCS{}
	router := events.NewEtcdCommandRouter(delegate, coord, vcsClient, logging.NewNoopLogger(t))

	repo := testRepo()
	cmd := &events.CommentCommand{Name: command.Plan, RepoRelDir: "dir", Workspace: "default"}
	router.RunCommentCommand(repo, nil, nil, models.User{Username: "u"}, 7, cmd)

	// A 202 must not produce a fail-closed comment.
	Equals(t, 0, vcsClient.commentCount())

	coord.mu.Lock()
	routed := coord.routed
	coord.mu.Unlock()
	Assert(t, routed != nil, "expected a routed command")
	Equals(t, "webhook", routed.Identity.SourceKind)
	Equals(t, "github.com", routed.Identity.VCSHostname)
	Equals(t, etcd.PullScope{VCSHostname: "github.com", Repository: "owner/repo", PullNum: 7}, routed.Scope)

	// The body must decode back to the original comment command.
	var payload struct {
		Kind    string                 `json:"kind"`
		PullNum int                    `json:"pull_num"`
		Comment *events.CommentCommand `json:"comment"`
	}
	Ok(t, json.Unmarshal(routed.Body, &payload))
	Equals(t, "comment", payload.Kind)
	Equals(t, 7, payload.PullNum)
	Assert(t, payload.Comment != nil, "expected comment in payload")
	Equals(t, command.Plan, payload.Comment.Name)
	Equals(t, "dir", payload.Comment.RepoRelDir)
}

// routedDeliveryID returns the DeliveryID of the last routed command.
func (f *fakeCoordinator) routedDeliveryID() string {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.routed == nil {
		return ""
	}
	return f.routed.Identity.DeliveryID
}

// TestEtcdCommandRouter_CommentDedupIdentity proves the routed identity is stable
// across redeliveries of the same comment (a provider redelivery deduplicates),
// distinct for a different comment id, and non-deduplicating when no comment id is
// present (design §509).
func TestEtcdCommandRouter_CommentDedupIdentity(t *testing.T) {
	coord := newFakeCoordinator()
	router := events.NewEtcdCommandRouter(&recordingRunner{}, coord, &commentRecordingVCS{}, logging.NewNoopLogger(t))
	repo := testRepo()
	user := models.User{Username: "u"}

	// Same comment id -> identical identity (redelivery deduplicates).
	cmd := &events.CommentCommand{Name: command.Plan, DeliveryID: "100"}
	router.RunCommentCommand(repo, nil, nil, user, 7, cmd)
	first := coord.routedDeliveryID()
	Assert(t, first != "", "delivery id should be set from the comment id")
	router.RunCommentCommand(repo, nil, nil, user, 7, cmd)
	Equals(t, first, coord.routedDeliveryID())

	// A different comment id -> a different identity (a re-issued command runs).
	cmd2 := &events.CommentCommand{Name: command.Plan, DeliveryID: "101"}
	router.RunCommentCommand(repo, nil, nil, user, 7, cmd2)
	Assert(t, coord.routedDeliveryID() != first, "distinct comment id must not deduplicate")

	// No comment id -> fresh id each time (matches non-HA behavior, no dedup).
	noID := &events.CommentCommand{Name: command.Plan}
	router.RunCommentCommand(repo, nil, nil, user, 7, noID)
	a := coord.routedDeliveryID()
	router.RunCommentCommand(repo, nil, nil, user, 7, noID)
	Assert(t, a != coord.routedDeliveryID(), "absent comment id must not deduplicate")
}

// TestEtcdCommandRouter_AutoplanReopenBumpsIdentity proves the lifecycle
// generation is folded into the autoplan identity, so a close+reopen at the same
// head commit (which bumps the generation) is NOT suppressed by the still-cached
// admission record from before the close.
func TestEtcdCommandRouter_AutoplanReopenBumpsIdentity(t *testing.T) {
	coord := newFakeCoordinator()
	router := events.NewEtcdCommandRouter(&recordingRunner{}, coord, &commentRecordingVCS{}, logging.NewNoopLogger(t))
	repo := testRepo()
	user := models.User{Username: "u"}
	pull := models.PullRequest{Num: 7, HeadCommit: "abc123"}

	coord.reopenGen = 0
	router.RunAutoplanCommand(repo, repo, pull, user)
	gen0 := coord.routedDeliveryID()

	// A reopen bumps the lifecycle generation; the same head commit must now map
	// to a distinct identity so the autoplan runs again.
	coord.reopenGen = 1
	router.RunAutoplanCommand(repo, repo, pull, user)
	Assert(t, coord.routedDeliveryID() != gen0, "a reopen (new generation) must not reuse the pre-close autoplan identity")
}

// TestEtcdCommandRouter_AutoplanDedupIdentity proves autoplan identity is derived
// from the head commit: a redelivered push for the same commit deduplicates, a new
// commit does not (design §509).
func TestEtcdCommandRouter_AutoplanDedupIdentity(t *testing.T) {
	coord := newFakeCoordinator()
	router := events.NewEtcdCommandRouter(&recordingRunner{}, coord, &commentRecordingVCS{}, logging.NewNoopLogger(t))
	repo := testRepo()
	user := models.User{Username: "u"}

	pull := models.PullRequest{Num: 7, HeadCommit: "abc123"}
	router.RunAutoplanCommand(repo, repo, pull, user)
	first := coord.routedDeliveryID()
	Assert(t, first != "", "delivery id should be derived from the head commit")
	router.RunAutoplanCommand(repo, repo, pull, user)
	Equals(t, first, coord.routedDeliveryID())

	newPush := models.PullRequest{Num: 7, HeadCommit: "def456"}
	router.RunAutoplanCommand(repo, repo, newPush, user)
	Assert(t, coord.routedDeliveryID() != first, "a new head commit must not deduplicate")
}

func TestEtcdCommandRouter_IngressUnavailable_FailsClosed(t *testing.T) {
	coord := newFakeCoordinator()
	coord.routeResult = etcd.Result{Status: http.StatusServiceUnavailable, Message: "backend down"}
	vcsClient := &commentRecordingVCS{}
	router := events.NewEtcdCommandRouter(&recordingRunner{}, coord, vcsClient, logging.NewNoopLogger(t))

	router.RunCommentCommand(testRepo(), nil, nil, models.User{}, 7, &events.CommentCommand{Name: command.Plan})

	Equals(t, 1, vcsClient.commentCount())
}

func TestEtcdCommandRouter_IngressRouteError_FailsClosed(t *testing.T) {
	coord := newFakeCoordinator()
	coord.routeErr = context.DeadlineExceeded
	vcsClient := &commentRecordingVCS{}
	router := events.NewEtcdCommandRouter(&recordingRunner{}, coord, vcsClient, logging.NewNoopLogger(t))

	router.RunAutoplanCommand(testRepo(), testRepo(), models.PullRequest{Num: 3}, models.User{})

	Equals(t, 1, vcsClient.commentCount())
}

func TestEtcdCommandRouter_ExecutorRunsComment(t *testing.T) {
	coord := newFakeCoordinator()
	delegate := &recordingRunner{}
	router := events.NewEtcdCommandRouter(delegate, coord, &commentRecordingVCS{}, logging.NewNoopLogger(t))

	body, err := json.Marshal(map[string]any{
		"kind":      "comment",
		"base_repo": testRepo(),
		"user":      models.User{Username: "u"},
		"pull_num":  9,
		"comment":   &events.CommentCommand{Name: command.Apply},
	})
	Ok(t, err)

	Ok(t, router.Register(context.Background(), etcd.Command{
		Identity: etcd.CommandIdentity{SourceKind: "webhook", VCSHostname: "github.com", DeliveryID: "d1"},
		Scope:    etcd.PullScope{VCSHostname: "github.com", Repository: "owner/repo", PullNum: 9},
		Body:     body,
	}))

	coord.waitExecuted(t)
	// Give the run closure a moment to record.
	delegate.mu.Lock()
	got := len(delegate.comments)
	delegate.mu.Unlock()
	Equals(t, 1, got)
}

func TestEtcdCommandRouter_ExecutorRunsAutoplan(t *testing.T) {
	coord := newFakeCoordinator()
	delegate := &recordingRunner{}
	router := events.NewEtcdCommandRouter(delegate, coord, &commentRecordingVCS{}, logging.NewNoopLogger(t))

	repo := testRepo()
	body, err := json.Marshal(map[string]any{
		"kind":      "autoplan",
		"base_repo": repo,
		"head_repo": repo,
		"pull":      models.PullRequest{Num: 5},
		"user":      models.User{Username: "u"},
		"pull_num":  5,
	})
	Ok(t, err)

	Ok(t, router.Register(context.Background(), etcd.Command{
		Identity: etcd.CommandIdentity{SourceKind: "autoplan", VCSHostname: "github.com", DeliveryID: "d2"},
		Scope:    etcd.PullScope{VCSHostname: "github.com", Repository: "owner/repo", PullNum: 5},
		Body:     body,
	}))

	coord.waitExecuted(t)
	delegate.mu.Lock()
	autoplan := delegate.autoplan
	pullNum := delegate.lastPull.Num
	delegate.mu.Unlock()
	Equals(t, 1, autoplan)
	Equals(t, 5, pullNum)
}

func TestEtcdCommandRouter_ExecutorBlocked_FailsClosed(t *testing.T) {
	coord := newFakeCoordinator()
	coord.execRun = false
	coord.execOutcome = etcd.ExecuteBlocked
	delegate := &recordingRunner{}
	vcsClient := &commentRecordingVCS{}
	router := events.NewEtcdCommandRouter(delegate, coord, vcsClient, logging.NewNoopLogger(t))

	body, err := json.Marshal(map[string]any{
		"kind":      "comment",
		"base_repo": testRepo(),
		"user":      models.User{},
		"pull_num":  9,
		"comment":   &events.CommentCommand{Name: command.Apply},
	})
	Ok(t, err)

	Ok(t, router.Register(context.Background(), etcd.Command{
		Identity: etcd.CommandIdentity{SourceKind: "webhook", VCSHostname: "github.com", DeliveryID: "d3"},
		Scope:    etcd.PullScope{VCSHostname: "github.com", Repository: "owner/repo", PullNum: 9},
		Body:     body,
	}))

	coord.waitExecuted(t)
	// Blocked must not run the delegate, and must surface a fail-closed comment.
	// Poll briefly for the async comment.
	deadline := time.Now().Add(5 * time.Second)
	for vcsClient.commentCount() == 0 && time.Now().Before(deadline) {
		time.Sleep(10 * time.Millisecond)
	}
	delegate.mu.Lock()
	ran := len(delegate.comments)
	delegate.mu.Unlock()
	Equals(t, 0, ran)
	Equals(t, 1, vcsClient.commentCount())
}

func TestEtcdCommandRouter_Register_RejectsBadBody(t *testing.T) {
	coord := newFakeCoordinator()
	router := events.NewEtcdCommandRouter(&recordingRunner{}, coord, &commentRecordingVCS{}, logging.NewNoopLogger(t))
	err := router.Register(context.Background(), etcd.Command{Body: []byte("not json")})
	Assert(t, err != nil, "expected an error decoding a bad body")
}
