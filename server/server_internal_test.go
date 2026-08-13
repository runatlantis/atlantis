// Copyright 2017 HootSuite Media Inc.
// SPDX-License-Identifier: Apache-2.0
// Modified hereafter by contributors to runatlantis/atlantis.

package server

import (
	"context"
	"errors"
	"testing"
	"testing/synctest"
	"time"

	"github.com/runatlantis/atlantis/server/core/db"
	"github.com/runatlantis/atlantis/server/core/db/mocks"
	"github.com/runatlantis/atlantis/server/core/ownership"
	"github.com/runatlantis/atlantis/server/events"
	"github.com/runatlantis/atlantis/server/logging"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestServer_ShutdownOrdersOwnershipAndDatabase(t *testing.T) {
	ctrl := gomock.NewController(t)
	database := mocks.NewMockDatabase(ctrl)
	var calls []string
	database.EXPECT().Close().DoAndReturn(func() error {
		calls = append(calls, "database-close")
		return nil
	})
	owners := &shutdownOwnerStore{calls: &calls}
	s := &Server{
		OwnerStore:            owners,
		commandExecutorWaiter: &recordingCommandWaiter{calls: &calls},
		Drainer:               &events.Drainer{},
		Logger:                logging.NewNoopLogger(t),
		database:              database,
	}
	httpServer := &recordingHTTPShutdowner{calls: &calls}

	assert.NoError(t, s.shutdown(httpServer, time.Second))
	assert.Equal(t, []string{"begin-drain", "http-shutdown", "command-wait", "owner-close", "database-close"}, calls)
}

func TestServer_ShutdownDoesNotReleaseClaimsAfterHTTPTimeout(t *testing.T) {
	var calls []string
	owners := &shutdownOwnerStore{calls: &calls}
	s := &Server{
		OwnerStore: owners,
		Drainer:    &events.Drainer{},
		Logger:     logging.NewNoopLogger(t),
	}
	httpServer := &recordingHTTPShutdowner{calls: &calls, err: context.DeadlineExceeded}

	err := s.shutdown(httpServer, time.Second)
	assert.ErrorContains(t, err, "while shutting down HTTP server")
	assert.Equal(t, []string{"begin-drain", "http-shutdown"}, calls)
}

type recordingHTTPShutdowner struct {
	calls *[]string
	err   error
}

type recordingCommandWaiter struct {
	calls *[]string
}

func (w *recordingCommandWaiter) Wait() {
	*w.calls = append(*w.calls, "command-wait")
}

func (s *recordingHTTPShutdowner) Shutdown(context.Context) error {
	*s.calls = append(*s.calls, "http-shutdown")
	return s.err
}

type shutdownOwnerStore struct {
	calls *[]string
}

func (s *shutdownOwnerStore) Claim(context.Context, ownership.Key) (ownership.Record, error) {
	return ownership.Record{}, nil
}
func (s *shutdownOwnerStore) Current(context.Context, ownership.Key) (ownership.Record, bool, error) {
	return ownership.Record{}, false, nil
}
func (s *shutdownOwnerStore) Admit(context.Context, ownership.Key, string) (bool, error) {
	return true, nil
}
func (s *shutdownOwnerStore) Owns(ownership.Key, string) bool { return false }
func (s *shutdownOwnerStore) Release(context.Context, ownership.Key, string) error {
	return nil
}
func (s *shutdownOwnerStore) BeginDrain() {
	*s.calls = append(*s.calls, "begin-drain")
}
func (s *shutdownOwnerStore) Ready(context.Context) error { return nil }
func (s *shutdownOwnerStore) Close() error {
	*s.calls = append(*s.calls, "owner-close")
	return nil
}

func TestServer_CloseDatabase(t *testing.T) {

	timeout := time.Second

	type databaseCase struct {
		description      string
		closeFn          func() error
		expectedErr      string
		expectedDuration time.Duration
	}

	cases := []databaseCase{
		{
			description: "closes successfully",
			closeFn:     func() error { return nil },
		},
		{
			description: "returns database error",
			closeFn:     func() error { return errors.New("boom") },
			expectedErr: "boom",
		},
		{
			description: "times out after 1s",
			closeFn: func() error {
				time.Sleep(1500 * time.Millisecond)
				return nil
			},
			expectedErr:      "timed out",
			expectedDuration: time.Second,
		},
		{
			description: "nil database",
			closeFn:     nil, // nil means database itself is nil
		},
	}

	for _, tt := range cases {
		t.Run(tt.description, func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				var database db.Database
				if tt.closeFn != nil {
					ctrl := gomock.NewController(t)
					m := mocks.NewMockDatabase(ctrl)
					closeFn := tt.closeFn
					m.EXPECT().Close().DoAndReturn(func() error {
						return closeFn()
					})
					database = m
				}

				s := &Server{
					database: database,
					Logger:   logging.NewNoopLogger(t),
				}

				start := time.Now()
				err := s.closeDatabase(timeout)
				duration := time.Since(start)

				assert.Equal(t, tt.expectedDuration, duration)

				//nolint:testifylint // testing error behavior, not precondition
				if tt.expectedErr == "" {
					assert.NoError(t, err)
				} else {
					assert.ErrorContains(t, err, tt.expectedErr)
				}

				// Make sure enough fake time so nothing is left running
				time.Sleep(2 * time.Second)
			})
		})
	}
}
