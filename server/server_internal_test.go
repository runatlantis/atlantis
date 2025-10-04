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

	. "github.com/petergtz/pegomock/v4"
	"github.com/runatlantis/atlantis/server/core/locking"
	"github.com/runatlantis/atlantis/server/core/locking/mocks"
	"github.com/runatlantis/atlantis/server/logging"
	"github.com/stretchr/testify/assert"
)

func TestServer_CloseBackend(t *testing.T) {
	type backendCase struct {
		description string
		closeFn     func() error
		expectedErr string
	}

	cases := []backendCase{
		{
			description: "closes successfully",
			closeFn:     func() error { return nil },
		},
		{
			description: "returns backend error",
			closeFn:     func() error { return errors.New("boom") },
			expectedErr: "boom",
		},
		{
			description: "times out after 1s",
			closeFn: func() error {
				time.Sleep(1500 * time.Millisecond)
				return nil
			},
			expectedErr: "timed out",
		},
		{
			description: "nil backend",
			closeFn:     nil, // nil means backend itself is nil
		},
	}

	for _, tt := range cases {
		t.Run(tt.description, func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				var backend locking.Backend
				if tt.closeFn != nil {
					m := mocks.NewMockBackend(WithT(t))
					When(m.Close()).Then(func([]Param) ReturnValues {
						return []ReturnValue{tt.closeFn()}
					})
					backend = m
				}

				s := &Server{
					backend: backend,
					Logger:  logging.NewNoopLogger(t),
				}

				err := s.closeBackend(time.Second)

				// "sleep" until after longest timeout
				time.Sleep(1 * time.Second)

				if tt.expectedErr == "" {
					assert.NoError(t, err)
				} else {
					assert.ErrorContains(t, err, tt.expectedErr)
				}
			})
		})
	}
}
