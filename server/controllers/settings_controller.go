// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package controllers

import (
	"net/http"

	"github.com/runatlantis/atlantis/server/controllers/web_templates"
	"github.com/runatlantis/atlantis/server/logging"
)

// SettingsController handles the settings page
type SettingsController struct {
	template               web_templates.TemplateWriter
	globalApplyLockEnabled bool
	isLocked               func() bool
	version                string
	cleanedBasePath        string
	logger                 logging.SimpleLogging
}

func NewSettingsController(
	template web_templates.TemplateWriter,
	globalApplyLockEnabled bool,
	isLocked func() bool,
	version string,
	cleanedBasePath string,
	logger logging.SimpleLogging,
) *SettingsController {
	return &SettingsController{
		template:               template,
		globalApplyLockEnabled: globalApplyLockEnabled,
		isLocked:               isLocked,
		version:                version,
		cleanedBasePath:        cleanedBasePath,
		logger:                 logger,
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

	renderTemplate(w, c.template, data, c.logger)
}
