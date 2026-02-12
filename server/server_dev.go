//go:build dev

package server

import (
	"os"

	web_templates "github.com/runatlantis/atlantis/server/controllers/web_templates"
)

//nolint:gochecknoinits // Dev-only: enables template hot reloading when built with -tags dev.
func init() {
	templatesDir := "server/controllers/web_templates/templates"
	if _, err := os.Stat(templatesDir); os.IsNotExist(err) {
		templatesDir = "controllers/web_templates/templates"
	}
	if _, err := os.Stat(templatesDir); err == nil {
		web_templates.SetDevMode(true, templatesDir)
	}
}