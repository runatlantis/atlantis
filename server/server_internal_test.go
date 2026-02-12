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

package server

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/synctest"
	"time"

	. "github.com/petergtz/pegomock/v4"
	"github.com/runatlantis/atlantis/server/core/db"
	"github.com/runatlantis/atlantis/server/core/db/mocks"
	"github.com/runatlantis/atlantis/server/logging"
	"github.com/stretchr/testify/assert"
)

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
					m := mocks.NewMockDatabase(WithT(t))
					When(m.Close()).Then(func([]Param) ReturnValues {
						return []ReturnValue{tt.closeFn()}
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

func TestRequireCSRFHeader(t *testing.T) {
	handlerCalled := false
	inner := func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		w.WriteHeader(http.StatusOK)
	}
	wrapped := requireCSRFHeader(inner)

	t.Run("rejects request without CSRF header", func(t *testing.T) {
		handlerCalled = false
		req := httptest.NewRequest(http.MethodDelete, "/locks?id=test", nil)
		w := httptest.NewRecorder()

		wrapped(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
		assert.Contains(t, w.Body.String(), "CSRF header required")
		assert.False(t, handlerCalled, "inner handler should not be called")
	})

	t.Run("rejects request with empty CSRF header", func(t *testing.T) {
		handlerCalled = false
		req := httptest.NewRequest(http.MethodDelete, "/locks?id=test", nil)
		req.Header.Set("X-Atlantis-CSRF", "")
		w := httptest.NewRecorder()

		wrapped(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
		assert.Contains(t, w.Body.String(), "CSRF header required")
		assert.False(t, handlerCalled, "inner handler should not be called")
	})

	t.Run("allows request with CSRF header", func(t *testing.T) {
		handlerCalled = false
		req := httptest.NewRequest(http.MethodDelete, "/locks?id=test", nil)
		req.Header.Set("X-Atlantis-CSRF", "1")
		w := httptest.NewRecorder()

		wrapped(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.True(t, handlerCalled, "inner handler should be called")
	})

	t.Run("allows request with any non-empty CSRF header value", func(t *testing.T) {
		handlerCalled = false
		req := httptest.NewRequest(http.MethodPost, "/apply/lock", nil)
		req.Header.Set("X-Atlantis-CSRF", "anything")
		w := httptest.NewRecorder()

		wrapped(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.True(t, handlerCalled, "inner handler should be called")
	})
}
