// Copyright 2026 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package controllers

import (
	"net/http"

	"github.com/runatlantis/atlantis/server/events/command"
)

func (a *APIController) StateShow(w http.ResponseWriter, r *http.Request) {
	a.apiRunInspection(w, r, command.StateShow)
}
