// Copyright 2026 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package events_test

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/runatlantis/atlantis/server/core/ownership"
	"github.com/runatlantis/atlantis/server/events"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/logging"
	"github.com/stretchr/testify/require"
)

func TestRoutedCommandDispatcher_ExecutesOnlyExactLocalClaim(t *testing.T) {
	dispatch := testCommentDispatch()
	owners := &dispatchOwnerStore{
		records: []ownership.Record{{ReplicaID: "replica-0", ClaimID: "claim-1", AdvertiseURL: "http://replica-0"}},
		owns:    map[string]bool{"claim-1": true},
	}
	local := &recordingCommandExecutor{}
	forwarder := &recordingCommandForwarder{}
	dispatcher := events.NewRoutedCommandDispatcher("replica-0", owners, local, forwarder)

	require.NoError(t, dispatcher.DispatchComment(dispatch))
	require.Equal(t, []string{"claim-1"}, local.commentClaims)
	require.Empty(t, forwarder.commentOwners)
}

func TestRoutedCommandDispatcher_ForwardsRemoteOwner(t *testing.T) {
	dispatch := testCommentDispatch()
	owners := &dispatchOwnerStore{
		records: []ownership.Record{{ReplicaID: "replica-1", ClaimID: "claim-1", AdvertiseURL: "http://replica-1"}},
	}
	local := &recordingCommandExecutor{}
	forwarder := &recordingCommandForwarder{}
	dispatcher := events.NewRoutedCommandDispatcher("replica-0", owners, local, forwarder)

	require.NoError(t, dispatcher.DispatchComment(dispatch))
	require.Empty(t, local.commentClaims)
	require.Equal(t, []string{"claim-1"}, forwarder.commentOwners)
}

func TestRoutedCommandDispatcher_DoesNotAdoptClaimFromPreviousProcess(t *testing.T) {
	dispatch := testCommentDispatch()
	owners := &dispatchOwnerStore{
		records: []ownership.Record{{ReplicaID: "replica-0", ClaimID: "old-process-claim", AdvertiseURL: "http://replica-0"}},
		owns:    map[string]bool{"old-process-claim": false},
	}
	local := &recordingCommandExecutor{}
	forwarder := &recordingCommandForwarder{}
	dispatcher := events.NewRoutedCommandDispatcher("replica-0", owners, local, forwarder)

	require.NoError(t, dispatcher.DispatchComment(dispatch))
	require.Empty(t, local.commentClaims)
	require.Equal(t, []string{"old-process-claim"}, forwarder.commentOwners)
}

func TestRoutedCommandDispatcher_RetriesOnceWhenOwnerChanges(t *testing.T) {
	dispatch := testCommentDispatch()
	owners := &dispatchOwnerStore{
		records: []ownership.Record{
			{ReplicaID: "replica-1", ClaimID: "expired", AdvertiseURL: "http://replica-1"},
			{ReplicaID: "replica-2", ClaimID: "replacement", AdvertiseURL: "http://replica-2"},
		},
	}
	forwarder := &recordingCommandForwarder{commentErrs: []error{events.ErrOwnershipChanged, nil}}
	dispatcher := events.NewRoutedCommandDispatcher("replica-0", owners, &recordingCommandExecutor{}, forwarder)

	require.NoError(t, dispatcher.DispatchComment(dispatch))
	require.Equal(t, []string{"expired", "replacement"}, forwarder.commentOwners)
	require.Equal(t, 2, owners.claimCalls)
}

func TestRoutedCommandDispatcher_FailsClosed(t *testing.T) {
	dispatch := testCommentDispatch()
	t.Run("redis unavailable", func(t *testing.T) {
		owners := &dispatchOwnerStore{claimErr: errors.New("redis unavailable")}
		local := &recordingCommandExecutor{}
		forwarder := &recordingCommandForwarder{}
		dispatcher := events.NewRoutedCommandDispatcher("replica-0", owners, local, forwarder)

		require.EqualError(t, dispatcher.DispatchComment(dispatch), "claiming command owner: redis unavailable")
		require.Empty(t, local.commentClaims)
		require.Empty(t, forwarder.commentOwners)
	})

	t.Run("forward failed", func(t *testing.T) {
		owners := &dispatchOwnerStore{records: []ownership.Record{{ReplicaID: "replica-1", ClaimID: "claim-1"}}}
		local := &recordingCommandExecutor{}
		forwarder := &recordingCommandForwarder{commentErrs: []error{errors.New("connection refused")}}
		dispatcher := events.NewRoutedCommandDispatcher("replica-0", owners, local, forwarder)

		require.EqualError(t, dispatcher.DispatchComment(dispatch), "forwarding command to replica-1: connection refused")
		require.Empty(t, local.commentClaims)
	})
}

func TestLocalCommandExecutor_DeletesStalePlanBeforeCommentCommand(t *testing.T) {
	baseRepo := hydratedRepo(t, "owner/repo")
	headRepo := hydratedRepo(t, "fork/repo")
	hydrator := &recordingRepoHydrator{repos: map[string]models.Repo{
		"owner/repo": baseRepo,
		"fork/repo":  headRepo,
	}}
	var sequence []string
	workingDir := &recordingPullStateCleaner{delete: func(_ logging.SimpleLogging, repo models.Repo, pull models.PullRequest) error {
		sequence = append(sequence, "delete")
		require.Equal(t, baseRepo, repo)
		require.Equal(t, 12, pull.Num)
		require.Equal(t, baseRepo, pull.BaseRepo)
		return nil
	}}
	runner := &recordingCommandRunner{comment: func(_ models.Repo, gotHead *models.Repo, gotPull *models.PullRequest, _ models.User, pullNum int, cmd *events.CommentCommand) {
		sequence = append(sequence, "run")
		require.Equal(t, &headRepo, gotHead)
		require.Nil(t, gotPull, "nil pull semantics must be preserved for the existing runner")
		require.Equal(t, 12, pullNum)
		require.Equal(t, command.Apply, cmd.Name)
	}}
	executor := &events.LocalCommandExecutor{
		Hydrator:    hydrator,
		Runner:      runner,
		WorkingDir:  workingDir,
		ClaimGuard:  events.NewLocalClaimGuard(),
		Logger:      logging.NewNoopLogger(t),
		TestingMode: true,
	}
	dispatch := testCommentDispatch()
	dispatch.HeadRepo = &events.RepoRef{FullName: "fork/repo", CloneURL: "https://github.com/fork/repo.git", VCSHost: dispatch.BaseRepo.VCSHost}
	dispatch.Command = &events.CommentCommand{Name: command.Apply}

	require.NoError(t, executor.ExecuteComment(dispatch, "claim-1"))
	require.NoError(t, executor.ExecuteComment(dispatch, "claim-1"))
	require.Equal(t, []string{"delete", "run", "run"}, sequence)
}

func TestLocalCommandExecutor_DoesNotRunWhenLocalResetFails(t *testing.T) {
	baseRepo := hydratedRepo(t, "owner/repo")
	runner := &recordingCommandRunner{}
	executor := &events.LocalCommandExecutor{
		Hydrator: &recordingRepoHydrator{repos: map[string]models.Repo{"owner/repo": baseRepo}},
		Runner:   runner,
		WorkingDir: &recordingPullStateCleaner{delete: func(logging.SimpleLogging, models.Repo, models.PullRequest) error {
			return errors.New("disk busy")
		}},
		ClaimGuard:  events.NewLocalClaimGuard(),
		Logger:      logging.NewNoopLogger(t),
		TestingMode: true,
	}

	require.EqualError(t, executor.ExecuteComment(testCommentDispatch(), "claim-1"), "resetting local pull state for new owner: disk busy")
	require.Equal(t, 0, runner.commentCalls)
}

func TestLocalCommandExecutor_PullCloseReleasesOnlyAfterSuccessfulCleanup(t *testing.T) {
	baseRepo := hydratedRepo(t, "owner/repo")
	dispatch := events.PullClosedDispatch{
		BaseRepo: events.RepoRef{FullName: "owner/repo", CloneURL: "https://github.com/owner/repo.git", VCSHost: baseRepo.VCSHost},
		Pull:     events.PullRef{Num: 12},
	}
	owners := &dispatchOwnerStore{}
	cleanupErr := errors.New("cleanup failed")
	pullCleaner := &recordingPullCleaner{errs: []error{cleanupErr, nil}}
	executor := &events.LocalCommandExecutor{
		Hydrator:    &recordingRepoHydrator{repos: map[string]models.Repo{"owner/repo": baseRepo}},
		PullCleaner: pullCleaner,
		WorkingDir: &recordingPullStateCleaner{delete: func(logging.SimpleLogging, models.Repo, models.PullRequest) error {
			return nil
		}},
		ClaimGuard: events.NewLocalClaimGuard(),
		Owners:     owners,
		Logger:     logging.NewNoopLogger(t),
	}

	require.ErrorIs(t, executor.ExecutePullClosed(dispatch, "claim-1"), cleanupErr)
	require.Empty(t, owners.releasedClaims)
	require.NoError(t, executor.ExecutePullClosed(dispatch, "claim-1"))
	require.Equal(t, []string{"claim-1"}, owners.releasedClaims)
}

func TestLocalCommandExecutor_WaitTracksAcceptedAsyncCommand(t *testing.T) {
	baseRepo := hydratedRepo(t, "owner/repo")
	started := make(chan struct{})
	release := make(chan struct{})
	runner := &recordingCommandRunner{comment: func(models.Repo, *models.Repo, *models.PullRequest, models.User, int, *events.CommentCommand) {
		close(started)
		<-release
	}}
	executor := &events.LocalCommandExecutor{
		Hydrator: &recordingRepoHydrator{repos: map[string]models.Repo{"owner/repo": baseRepo}},
		Runner:   runner,
		Logger:   logging.NewNoopLogger(t),
	}
	require.NoError(t, executor.ExecuteComment(testCommentDispatch(), ""))
	<-started

	waited := make(chan struct{})
	go func() {
		executor.Wait()
		close(waited)
	}()
	select {
	case <-waited:
		t.Fatal("Wait returned while an accepted command was still running")
	default:
	}
	close(release)
	<-waited
}

func testCommentDispatch() events.CommentDispatch {
	return events.CommentDispatch{
		BaseRepo: events.RepoRef{
			FullName: "owner/repo",
			CloneURL: "https://github.com/owner/repo.git",
			VCSHost:  models.VCSHost{Type: models.Github, Hostname: "github.com"},
		},
		PullNum: 12,
		Command: &events.CommentCommand{Name: command.Plan},
	}
}

func hydratedRepo(t *testing.T, fullName string) models.Repo {
	t.Helper()
	repo, err := models.NewRepo(models.Github, fullName, "https://github.com/"+fullName, "user", "token", "")
	require.NoError(t, err)
	return repo
}

type dispatchOwnerStore struct {
	mu             sync.Mutex
	records        []ownership.Record
	claimErr       error
	claimCalls     int
	owns           map[string]bool
	readyErr       error
	releaseErr     error
	releasedClaims []string
}

func (s *dispatchOwnerStore) Claim(context.Context, ownership.Key) (ownership.Record, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.claimCalls++
	if s.claimErr != nil {
		return ownership.Record{}, s.claimErr
	}
	if len(s.records) == 0 {
		return ownership.Record{}, errors.New("no test ownership record")
	}
	idx := s.claimCalls - 1
	if idx >= len(s.records) {
		idx = len(s.records) - 1
	}
	return s.records[idx], nil
}

func (s *dispatchOwnerStore) Current(context.Context, ownership.Key) (ownership.Record, bool, error) {
	return ownership.Record{}, false, nil
}

func (s *dispatchOwnerStore) Owns(_ ownership.Key, claimID string) bool {
	return s.owns[claimID]
}

func (s *dispatchOwnerStore) Release(_ context.Context, _ ownership.Key, claimID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.releaseErr == nil {
		s.releasedClaims = append(s.releasedClaims, claimID)
	}
	return s.releaseErr
}

func (s *dispatchOwnerStore) BeginDrain() {}
func (s *dispatchOwnerStore) Ready(context.Context) error {
	return s.readyErr
}
func (s *dispatchOwnerStore) Close() error { return nil }

type recordingCommandExecutor struct {
	commentClaims []string
}

func (e *recordingCommandExecutor) ExecuteComment(_ events.CommentDispatch, claimID string) error {
	e.commentClaims = append(e.commentClaims, claimID)
	return nil
}
func (e *recordingCommandExecutor) ExecuteAutoplan(events.AutoplanDispatch, string) error { return nil }
func (e *recordingCommandExecutor) ExecutePullClosed(events.PullClosedDispatch, string) error {
	return nil
}

type recordingCommandForwarder struct {
	commentOwners []string
	commentErrs   []error
}

func (f *recordingCommandForwarder) ForwardComment(owner ownership.Record, _ events.CommentDispatch) error {
	f.commentOwners = append(f.commentOwners, owner.ClaimID)
	idx := len(f.commentOwners) - 1
	if idx < len(f.commentErrs) {
		return f.commentErrs[idx]
	}
	return nil
}
func (f *recordingCommandForwarder) ForwardAutoplan(ownership.Record, events.AutoplanDispatch) error {
	return nil
}
func (f *recordingCommandForwarder) ForwardPullClosed(ownership.Record, events.PullClosedDispatch) error {
	return nil
}

type recordingRepoHydrator struct {
	repos map[string]models.Repo
}

func (h *recordingRepoHydrator) HydrateRepo(ref events.RepoRef) (models.Repo, error) {
	repo, ok := h.repos[ref.FullName]
	if !ok {
		return models.Repo{}, errors.New("unknown repo")
	}
	return repo, nil
}

type recordingPullStateCleaner struct {
	delete func(logging.SimpleLogging, models.Repo, models.PullRequest) error
}

func (c *recordingPullStateCleaner) Delete(logger logging.SimpleLogging, repo models.Repo, pull models.PullRequest) error {
	return c.delete(logger, repo, pull)
}

type recordingCommandRunner struct {
	comment      func(models.Repo, *models.Repo, *models.PullRequest, models.User, int, *events.CommentCommand)
	commentCalls int
}

func (r *recordingCommandRunner) RunCommentCommand(base models.Repo, head *models.Repo, pull *models.PullRequest, user models.User, pullNum int, cmd *events.CommentCommand) {
	r.commentCalls++
	if r.comment != nil {
		r.comment(base, head, pull, user, pullNum, cmd)
	}
}

func (r *recordingCommandRunner) RunAutoplanCommand(models.Repo, models.Repo, models.PullRequest, models.User) {
}

type recordingPullCleaner struct {
	errs  []error
	calls int
}

func (c *recordingPullCleaner) CleanUpPull(logging.SimpleLogging, models.Repo, models.PullRequest) error {
	idx := c.calls
	c.calls++
	if idx < len(c.errs) {
		return c.errs[idx]
	}
	return nil
}
