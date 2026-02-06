// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package controllers

import (
	"net/http"

	"github.com/runatlantis/atlantis/server/controllers/web_templates"
)

// SettingsController handles the settings page
type SettingsController struct {
	template               web_templates.TemplateWriter
	globalApplyLockEnabled bool
	isLocked               func() bool
	version                string
	cleanedBasePath        string
}

func NewSettingsController(
	template web_templates.TemplateWriter,
	globalApplyLockEnabled bool,
	isLocked func() bool,
	version string,
	cleanedBasePath string,
) *SettingsController {
	return &SettingsController{
		template:               template,
		globalApplyLockEnabled: globalApplyLockEnabled,
		isLocked:               isLocked,
		version:                version,
		cleanedBasePath:        cleanedBasePath,
	}
}

func (c *SettingsController) Get(w http.ResponseWriter, r *http.Request) {
	applyLockActive := c.isLocked()

	data := web_templates.SettingsData{
		LayoutData: web_templates.LayoutData{
			AtlantisVersion: c.version,
			CleanedBasePath: c.cleanedBasePath,
			ActiveNav:       "settings",
			ApplyLockActive: applyLockActive,
		},
		GlobalApplyLockEnabled: c.globalApplyLockEnabled,
		Version:                c.version,
	}

	if err := c.template.Execute(w, data); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}
