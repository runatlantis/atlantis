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
	"testing"
	"testing/synctest"
	"time"

	"github.com/runatlantis/atlantis/server/core/db"
	"github.com/runatlantis/atlantis/server/core/db/mocks"
	"github.com/runatlantis/atlantis/server/logging"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
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
