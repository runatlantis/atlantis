// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package controllers

import (
	"bytes"
	"net/http"

	"github.com/runatlantis/atlantis/server/controllers/web_templates"
	"github.com/runatlantis/atlantis/server/logging"
)

// renderTemplate executes a template into a buffer first, then writes to the
// response. This prevents partial HTML output when template execution fails.
func renderTemplate(w http.ResponseWriter, tmpl web_templates.TemplateWriter, data any, logger logging.SimpleLogging) {
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		if logger != nil {
			logger.Err("template execution failed: %s", err)
		}
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	buf.WriteTo(w) //nolint:errcheck
}
