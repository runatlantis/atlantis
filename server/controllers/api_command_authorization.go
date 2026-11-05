// Copyright 2026 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package controllers

import (
	"crypto/subtle"
	"fmt"
	"net/http"
	"slices"

	"github.com/runatlantis/atlantis/server/events/command"
)

func (a *APIController) requireAPICommand(w http.ResponseWriter, r *http.Request, name command.Name) bool {
	if len(a.APISecret) == 0 {
		a.apiReportLegacyError(w, http.StatusServiceUnavailable, fmt.Errorf("ignoring request since API is disabled"))
		return false
	}
	if subtle.ConstantTimeCompare([]byte(r.Header.Get(atlantisTokenHeader)), a.APISecret) != 1 {
		a.apiReportLegacyError(w, http.StatusUnauthorized, fmt.Errorf("header %s did not match expected secret", atlantisTokenHeader))
		return false
	}
	allowed := a.AllowAPICommands
	if allowed == nil {
		allowed = []command.Name{command.Plan, command.Apply}
	}
	if !slices.Contains(allowed, name) {
		a.apiReportLegacyError(w, http.StatusForbidden, fmt.Errorf("API command %q is disabled by --allow-api-commands", name))
		return false
	}
	return true
}
