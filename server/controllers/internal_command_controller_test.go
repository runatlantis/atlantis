// Copyright 2026 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package controllers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/runatlantis/atlantis/server/controllers"
	"github.com/runatlantis/atlantis/server/core/ownership"
	"github.com/runatlantis/atlantis/server/events"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/stretchr/testify/require"
)

func TestInternalCommandController_AcceptsExactOwnedClaim(t *testing.T) {
	dispatch := internalCommentDispatch()
	owners := &internalOwnerStore{
		current: ownership.Record{ReplicaID: "replica-1", ClaimID: "claim-1"},
		found:   true,
		owns:    true,
	}
	executor := &internalCommandExecutor{}
	controller := controllers.InternalCommandController{
		Token:     "internal-secret",
		ReplicaID: "replica-1",
		Owners:    owners,
		Executor:  executor,
	}
	recorder := performInternalRequest(t, controller.Comment, dispatch, "internal-secret", "claim-1")

	require.Equal(t, http.StatusAccepted, recorder.Code)
	require.Equal(t, []string{"claim-1"}, executor.commentClaims)
	require.Equal(t, 1, owners.currentCalls)
}

func TestInternalCommandController_RejectsBadTokenBeforeReadingOwnership(t *testing.T) {
	owners := &internalOwnerStore{}
	controller := controllers.InternalCommandController{
		Token: "internal-secret", ReplicaID: "replica-1", Owners: owners, Executor: &internalCommandExecutor{},
	}
	recorder := performInternalRequest(t, controller.Comment, internalCommentDispatch(), "wrong-secret", "claim-1")

	require.Equal(t, http.StatusUnauthorized, recorder.Code)
	require.Equal(t, 0, owners.currentCalls)
}

func TestInternalCommandController_RejectsMalformedPayload(t *testing.T) {
	controller := controllers.InternalCommandController{
		Token: "internal-secret", ReplicaID: "replica-1", Owners: &internalOwnerStore{}, Executor: &internalCommandExecutor{},
	}
	tests := []string{
		`{"base_repo":{},"pull_num":12,"unknown":true}`,
		`{"base_repo":`,
		strings.Repeat("x", (1<<20)+1),
	}
	for _, body := range tests {
		req := httptest.NewRequest(http.MethodPost, events.InternalCommentCommandPath, strings.NewReader(body))
		req.Header.Set(events.InternalCommandTokenHeader, "internal-secret")
		req.Header.Set(events.OwnershipClaimHeader, "claim-1")
		recorder := httptest.NewRecorder()

		controller.Comment(recorder, req)
		require.Equal(t, http.StatusBadRequest, recorder.Code)
	}
}

func TestInternalCommandController_RejectsInvalidOwnershipKeyBeforeRedis(t *testing.T) {
	owners := &internalOwnerStore{}
	controller := controllers.InternalCommandController{
		Token: "internal-secret", ReplicaID: "replica-1", Owners: owners, Executor: &internalCommandExecutor{},
	}
	dispatch := internalCommentDispatch()
	dispatch.PullNum = 0
	recorder := performInternalRequest(t, controller.Comment, dispatch, "internal-secret", "claim-1")

	require.Equal(t, http.StatusBadRequest, recorder.Code)
	require.Equal(t, 0, owners.currentCalls)
}

func TestInternalCommandController_FencesStaleOwnership(t *testing.T) {
	tests := []struct {
		name    string
		owners  *internalOwnerStore
		claimID string
	}{
		{name: "missing", owners: &internalOwnerStore{}, claimID: "claim-1"},
		{name: "different replica", owners: &internalOwnerStore{found: true, current: ownership.Record{ReplicaID: "replica-2", ClaimID: "claim-1"}, owns: true}, claimID: "claim-1"},
		{name: "different claim", owners: &internalOwnerStore{found: true, current: ownership.Record{ReplicaID: "replica-1", ClaimID: "claim-2"}, owns: true}, claimID: "claim-1"},
		{name: "previous process", owners: &internalOwnerStore{found: true, current: ownership.Record{ReplicaID: "replica-1", ClaimID: "claim-1"}, owns: false}, claimID: "claim-1"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			executor := &internalCommandExecutor{}
			controller := controllers.InternalCommandController{
				Token: "internal-secret", ReplicaID: "replica-1", Owners: tt.owners, Executor: executor,
			}
			recorder := performInternalRequest(t, controller.Comment, internalCommentDispatch(), "internal-secret", tt.claimID)

			require.Equal(t, http.StatusConflict, recorder.Code)
			require.Empty(t, executor.commentClaims)
		})
	}
}

func TestInternalCommandController_ReturnsUnavailableWithoutExecuting(t *testing.T) {
	tests := []struct {
		name     string
		owners   *internalOwnerStore
		executor *internalCommandExecutor
	}{
		{name: "not ready", owners: &internalOwnerStore{readyErr: errors.New("redis unavailable")}, executor: &internalCommandExecutor{}},
		{name: "current failed", owners: &internalOwnerStore{currentErr: errors.New("redis unavailable")}, executor: &internalCommandExecutor{}},
		{name: "executor failed", owners: &internalOwnerStore{found: true, current: ownership.Record{ReplicaID: "replica-1", ClaimID: "claim-1"}, owns: true}, executor: &internalCommandExecutor{commentErr: errors.New("disk unavailable")}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			controller := controllers.InternalCommandController{
				Token: "internal-secret", ReplicaID: "replica-1", Owners: tt.owners, Executor: tt.executor,
			}
			recorder := performInternalRequest(t, controller.Comment, internalCommentDispatch(), "internal-secret", "claim-1")

			require.Equal(t, http.StatusServiceUnavailable, recorder.Code)
		})
	}
}

func performInternalRequest(t *testing.T, handler http.HandlerFunc, payload any, token, claimID string) *httptest.ResponseRecorder {
	t.Helper()
	body, err := json.Marshal(payload)
	require.NoError(t, err)
	req := httptest.NewRequest(http.MethodPost, events.InternalCommentCommandPath, bytes.NewReader(body))
	req.Header.Set(events.InternalCommandTokenHeader, token)
	req.Header.Set(events.OwnershipClaimHeader, claimID)
	recorder := httptest.NewRecorder()
	handler(recorder, req)
	return recorder
}

func internalCommentDispatch() events.CommentDispatch {
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

type internalOwnerStore struct {
	current      ownership.Record
	found        bool
	currentErr   error
	readyErr     error
	owns         bool
	currentCalls int
}

func (s *internalOwnerStore) Claim(context.Context, ownership.Key) (ownership.Record, error) {
	return ownership.Record{}, errors.New("unexpected claim")
}
func (s *internalOwnerStore) Current(context.Context, ownership.Key) (ownership.Record, bool, error) {
	s.currentCalls++
	return s.current, s.found, s.currentErr
}
func (s *internalOwnerStore) Owns(ownership.Key, string) bool { return s.owns }
func (s *internalOwnerStore) Release(context.Context, ownership.Key, string) error {
	return nil
}
func (s *internalOwnerStore) BeginDrain()                 {}
func (s *internalOwnerStore) Ready(context.Context) error { return s.readyErr }
func (s *internalOwnerStore) Close() error                { return nil }

type internalCommandExecutor struct {
	commentClaims []string
	commentErr    error
}

func (e *internalCommandExecutor) ExecuteComment(_ events.CommentDispatch, claimID string) error {
	e.commentClaims = append(e.commentClaims, claimID)
	return e.commentErr
}
func (e *internalCommandExecutor) ExecuteAutoplan(events.AutoplanDispatch, string) error {
	return nil
}
func (e *internalCommandExecutor) ExecutePullClosed(events.PullClosedDispatch, string) error {
	return nil
}
