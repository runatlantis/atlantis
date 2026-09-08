// Copyright 2026 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package controllers_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/runatlantis/atlantis/server/controllers"
	"github.com/runatlantis/atlantis/server/core/db"
	redisdb "github.com/runatlantis/atlantis/server/core/redis"
	"github.com/runatlantis/atlantis/server/events"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/logging"
	"github.com/stretchr/testify/require"
)

func TestPublicationRecoveryAPI_ExactOwnerAndAcceptedStatus(t *testing.T) {
	server := miniredis.RunT(t)
	now := time.Unix(1000, 0)
	server.SetTime(now)
	port, err := strconv.Atoi(server.Port())
	require.NoError(t, err)
	store, err := redisdb.New(server.Host(), port, "", false, false, 0)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, store.Close()) })
	pull := models.PullRequest{Num: 1, HeadCommit: "head", BaseRepo: models.Repo{FullName: "owner/repo", VCSHost: models.VCSHost{Hostname: "github.com"}}}
	project := command.ProjectContext{Workspace: "default", RepoRelDir: "."}
	_, err = store.BeginPlanGeneration(pull, "G1", []command.ProjectContext{project}, false, command.NoClaim{})
	require.NoError(t, err)
	accepted, err := store.UpdatePullWithResults(pull, []command.ProjectResult{{Command: command.Plan, Workspace: "default", RepoRelDir: ".", PlanGeneration: "G1", ProjectCommandOutput: command.ProjectCommandOutput{PlanSuccess: &models.PlanSuccess{}}}}, command.NoClaim{})
	require.NoError(t, err)
	_, err = store.AcquirePublicationLease(context.Background(), pull, "original", time.Minute)
	require.NoError(t, err)
	require.NoError(t, store.BeginPublication(context.Background(), pull, db.PublicationFence{Owner: "original"}))
	allowlist, err := events.NewRepoAllowlistChecker("github.com/owner/repo")
	require.NoError(t, err)
	controller := &controllers.APIController{APISecret: []byte("test-secret"), Logger: logging.NewNoopLogger(t), PublicationRecovery: store, RepoAllowlistChecker: allowlist}
	recoverBody := `{"repository":"owner/repo","hostname":"github.com","pull_num":1,"expected_owner":"original","operation_stopped_and_reconciled":true}`
	request := func(method, body, token string, cancelled bool) *httptest.ResponseRecorder {
		req := httptest.NewRequest(method, "/api/publication-lease?repository=owner/repo&hostname=github.com&pull_num=1", strings.NewReader(body))
		req.Header.Set("X-Atlantis-Token", token)
		if cancelled {
			ctx, cancel := context.WithCancel(req.Context())
			cancel()
			req = req.WithContext(ctx)
		}
		response := httptest.NewRecorder()
		if method == http.MethodGet {
			controller.PublicationLease(response, req)
		} else {
			controller.RecoverPublicationLease(response, req)
		}
		return response
	}
	response := request(http.MethodGet, "", "test-secret", false)
	require.Equal(t, http.StatusOK, response.Code)
	var inspected struct {
		Owner      string `json:"owner"`
		Publishing bool   `json:"publishing"`
	}
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &inspected))
	require.Equal(t, "original", inspected.Owner)
	require.True(t, inspected.Publishing)
	require.Equal(t, http.StatusConflict, request(http.MethodPost, recoverBody, "test-secret", false).Code, "a live owner cannot be recovered")
	server.SetTime(now.Add(time.Minute))
	for _, tc := range []struct {
		name, body, token string
		cancelled         bool
		code              int
	}{
		{"authentication", recoverBody, "wrong", false, http.StatusUnauthorized},
		{"acknowledgement", strings.ReplaceAll(recoverBody, "true", "false"), "test-secret", false, http.StatusBadRequest},
		{"wrong owner", strings.ReplaceAll(recoverBody, "original", "wrong"), "test-secret", false, http.StatusConflict},
		{"allowlist", strings.ReplaceAll(recoverBody, "owner/repo", "other/repo"), "test-secret", false, http.StatusForbidden},
		{"cancellation", recoverBody, "test-secret", true, http.StatusServiceUnavailable},
	} {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.code, request(http.MethodPost, tc.body, tc.token, tc.cancelled).Code)
			lease, err := store.GetPublicationLease(context.Background(), pull)
			require.NoError(t, err)
			require.Equal(t, "original", lease.Owner)
			require.True(t, lease.Publishing)
		})
	}
	require.Equal(t, http.StatusNoContent, request(http.MethodPost, recoverBody, "test-secret", false).Code)
	require.Equal(t, http.StatusNotFound, request(http.MethodGet, "", "test-secret", false).Code)
	after, err := store.GetPullStatus(pull)
	require.NoError(t, err)
	require.Equal(t, accepted, *after)
	_, err = store.AcquirePublicationLease(context.Background(), pull, "replacement", time.Minute)
	require.NoError(t, err)
	require.Equal(t, http.StatusConflict, request(http.MethodPost, recoverBody, "test-secret", false).Code)
}
