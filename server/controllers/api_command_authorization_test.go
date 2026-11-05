// Copyright 2026 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package controllers_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/petergtz/pegomock/v4"
	"github.com/runatlantis/atlantis/server/events/command"
	. "github.com/runatlantis/atlantis/testing"
)

func TestAPICommandAllowlistCanDisablePlanAndApply(t *testing.T) {
	for _, name := range []command.Name{command.Plan, command.Apply} {
		t.Run(name.String(), func(t *testing.T) {
			ac, _, runner := setup(t)
			ac.AllowAPICommands = []command.Name{command.Output}
			req := httptest.NewRequest(http.MethodPost, "/api/"+name.String(), bytes.NewBufferString(`{}`))
			req.Header.Set(atlantisTokenHeader, atlantisToken)
			w := httptest.NewRecorder()
			if name == command.Plan {
				ac.Plan(w, req)
			} else {
				ac.Apply(w, req)
			}
			Equals(t, http.StatusForbidden, w.Code)
			runner.VerifyWasCalled(Never()).Plan(Any[command.ProjectContext]())
			runner.VerifyWasCalled(Never()).Apply(Any[command.ProjectContext]())
		})
	}
}
