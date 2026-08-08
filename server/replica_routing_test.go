// Copyright 2026 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package server_test

import (
	"context"
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/gorilla/mux"
	"github.com/runatlantis/atlantis/server"
	"github.com/runatlantis/atlantis/server/controllers"
	events_controllers "github.com/runatlantis/atlantis/server/controllers/events"
	"github.com/runatlantis/atlantis/server/core/ownership"
	"github.com/runatlantis/atlantis/server/events"
	"github.com/runatlantis/atlantis/server/logging"
	"github.com/stretchr/testify/require"
)

func TestNewServer_WiresReplicaRoutingWithSharedRedis(t *testing.T) {
	redisServer := miniredis.RunT(t)
	host, portString, err := net.SplitHostPort(redisServer.Addr())
	require.NoError(t, err)
	port, err := strconv.Atoi(portString)
	require.NoError(t, err)

	s, err := server.NewServer(server.UserConfig{
		DataDir:              t.TempDir(),
		AtlantisURL:          "http://example.com",
		LockingDBType:        "redis",
		RedisHost:            host,
		RedisPort:            port,
		GithubHostname:       "github.com",
		GithubUser:           "user",
		ReplicaAdvertiseURL:  "http://replica-0:4141",
		InternalCommandToken: "internal-token",
		OwnershipTTLSeconds:  30,
	}, server.Config{AtlantisVersion: "test"})
	require.NoError(t, err)
	t.Cleanup(func() {
		if s.OwnerStore != nil {
			require.NoError(t, s.OwnerStore.Close())
		}
	})

	require.True(t, s.EnableReplicaRouting)
	require.NotNil(t, s.OwnerStore)
	require.NotNil(t, s.InternalCommandController)
	_, routed := s.VCSEventsController.CommandDispatcher.(*events.RoutedCommandDispatcher)
	require.True(t, routed)

	record, err := s.OwnerStore.Claim(context.Background(), ownership.Key{
		VCSHostname: "github.com", RepoFullName: "owner/repo", PullNum: 12,
	})
	require.NoError(t, err)
	hostname, err := os.Hostname()
	require.NoError(t, err)
	require.Equal(t, hostname, record.ReplicaID)
}

func TestNewServer_ExplicitReplicaIDOverridesHostname(t *testing.T) {
	redisServer := miniredis.RunT(t)
	host, portString, err := net.SplitHostPort(redisServer.Addr())
	require.NoError(t, err)
	port, err := strconv.Atoi(portString)
	require.NoError(t, err)

	s, err := server.NewServer(server.UserConfig{
		DataDir:              t.TempDir(),
		AtlantisURL:          "http://example.com",
		LockingDBType:        "redis",
		RedisHost:            host,
		RedisPort:            port,
		GithubHostname:       "github.com",
		GithubUser:           "user",
		ReplicaID:            "replica-override",
		ReplicaAdvertiseURL:  "http://replica-0:4141",
		InternalCommandToken: "internal-token",
		OwnershipTTLSeconds:  30,
	}, server.Config{AtlantisVersion: "test"})
	require.NoError(t, err)
	t.Cleanup(func() {
		if s.OwnerStore != nil {
			require.NoError(t, s.OwnerStore.Close())
		}
	})

	record, err := s.OwnerStore.Claim(context.Background(), ownership.Key{
		VCSHostname: "github.com", RepoFullName: "owner/repo", PullNum: 12,
	})
	require.NoError(t, err)
	require.Equal(t, "replica-override", record.ReplicaID)
}

func TestNewServer_DisabledReplicaRoutingPreservesLegacyDispatch(t *testing.T) {
	redisServer := miniredis.RunT(t)
	host, portString, err := net.SplitHostPort(redisServer.Addr())
	require.NoError(t, err)
	port, err := strconv.Atoi(portString)
	require.NoError(t, err)

	s, err := server.NewServer(server.UserConfig{
		DataDir:        t.TempDir(),
		AtlantisURL:    "http://example.com",
		LockingDBType:  "redis",
		RedisHost:      host,
		RedisPort:      port,
		GithubHostname: "github.com",
		GithubUser:     "user",
	}, server.Config{AtlantisVersion: "test"})
	require.NoError(t, err)

	require.False(t, s.EnableReplicaRouting)
	require.Nil(t, s.OwnerStore)
	require.Nil(t, s.InternalCommandController)
	require.Nil(t, s.VCSEventsController.CommandDispatcher)
}

func TestSetupRoutes_ReplicaRoutingRoutesAreConditional(t *testing.T) {
	newServer := func(enabled bool) *server.Server {
		return &server.Server{
			Router:                    mux.NewRouter(),
			APIController:             &controllers.APIController{},
			StatusController:          &controllers.StatusController{},
			LocksController:           &controllers.LocksController{},
			GithubAppController:       &controllers.GithubAppController{},
			JobsController:            &controllers.JobsController{},
			VCSEventsController:       &events_controllers.VCSEventsController{},
			InternalCommandController: &controllers.InternalCommandController{},
			EnableReplicaRouting:      enabled,
			Logger:                    logging.NewNoopLogger(t),
		}
	}

	t.Run("enabled", func(t *testing.T) {
		s := newServer(true)
		s.SetupRoutes()
		for _, path := range []string{
			events.InternalCommentCommandPath,
			events.InternalAutoplanCommandPath,
			events.InternalPullClosedCommandPath,
		} {
			req := httptest.NewRequest(http.MethodPost, path, nil)
			var match mux.RouteMatch
			require.True(t, s.Router.Match(req, &match), path)
		}
	})

	t.Run("disabled", func(t *testing.T) {
		s := newServer(false)
		s.SetupRoutes()
		req := httptest.NewRequest(http.MethodPost, events.InternalCommentCommandPath, nil)
		var match mux.RouteMatch
		require.False(t, s.Router.Match(req, &match))
	})
}

func TestReadyz_IncludesOwnershipLeaseHealth(t *testing.T) {
	owners := &serverOwnerStore{}
	s := &server.Server{OwnerStore: owners}

	t.Run("ready", func(t *testing.T) {
		recorder := httptest.NewRecorder()
		s.Readyz(recorder, httptest.NewRequest(http.MethodGet, "/readyz", nil))
		require.Equal(t, http.StatusOK, recorder.Code)
	})

	t.Run("lease renewal unhealthy", func(t *testing.T) {
		owners.readyErr = errors.New("ownership renewal failed")
		recorder := httptest.NewRecorder()
		s.Readyz(recorder, httptest.NewRequest(http.MethodGet, "/readyz", nil))
		require.Equal(t, http.StatusServiceUnavailable, recorder.Code)
		require.Contains(t, recorder.Body.String(), "ownership renewal failed")
	})
}

type serverOwnerStore struct {
	readyErr error
}

func (s *serverOwnerStore) Claim(context.Context, ownership.Key) (ownership.Record, error) {
	return ownership.Record{}, nil
}
func (s *serverOwnerStore) Current(context.Context, ownership.Key) (ownership.Record, bool, error) {
	return ownership.Record{}, false, nil
}
func (s *serverOwnerStore) Owns(ownership.Key, string) bool { return false }
func (s *serverOwnerStore) Release(context.Context, ownership.Key, string) error {
	return nil
}
func (s *serverOwnerStore) BeginDrain()                 {}
func (s *serverOwnerStore) Ready(context.Context) error { return s.readyErr }
func (s *serverOwnerStore) Close() error                { return nil }
