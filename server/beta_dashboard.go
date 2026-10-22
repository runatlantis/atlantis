// Copyright 2026 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"net/http"

	"github.com/runatlantis/atlantis/server/controllers/web_templates"
)

// BetaDashboard is an opt-in view of the same data as the classic dashboard.
// It uses the existing routes and middleware; it does not change job execution.
func (s *Server) BetaDashboard(w http.ResponseWriter, _ *http.Request) {
	s.renderIndex(w, web_templates.BetaDashboardTemplate)
}
