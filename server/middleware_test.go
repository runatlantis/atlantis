// Copyright 2026 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package server_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/runatlantis/atlantis/server"
	"github.com/runatlantis/atlantis/server/logging"
	. "github.com/runatlantis/atlantis/testing"
	"github.com/urfave/negroni/v3"
)

func TestRequestLogger_WebAuthentication(t *testing.T) {
	requestLogger := server.NewRequestLogger(&server.Server{
		Logger:            logging.NewNoopLogger(t),
		WebAuthentication: true,
		WebUsername:       "user",
		WebPassword:       "password",
	})

	tests := []struct {
		name           string
		path           string
		expStatus      int
		expNextHandler bool
	}{
		{
			name:           "ready endpoint is public",
			path:           "/readyz",
			expStatus:      http.StatusNoContent,
			expNextHandler: true,
		},
		{
			name:           "root remains protected",
			path:           "/",
			expStatus:      http.StatusUnauthorized,
			expNextHandler: false,
		},
		{
			name:           "ready endpoint subpath remains protected",
			path:           "/readyz/details",
			expStatus:      http.StatusUnauthorized,
			expNextHandler: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			responseWriter := negroni.NewResponseWriter(recorder)
			request := httptest.NewRequest(http.MethodGet, test.path, nil)
			nextHandlerCalled := false

			requestLogger.ServeHTTP(responseWriter, request, func(w http.ResponseWriter, _ *http.Request) {
				nextHandlerCalled = true
				w.WriteHeader(http.StatusNoContent)
			})

			Equals(t, test.expStatus, recorder.Code)
			Equals(t, test.expNextHandler, nextHandlerCalled)
		})
	}
}
