package controllers_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/runatlantis/atlantis/server/controllers"
	"github.com/runatlantis/atlantis/server/controllers/web_templates"
	. "github.com/runatlantis/atlantis/testing"
)

func TestSettingsController_Get(t *testing.T) {
	t.Run("renders settings page", func(t *testing.T) {
		tmpl := web_templates.SettingsTemplate
		controller := controllers.NewSettingsController(
			tmpl,
			false, // globalApplyLockEnabled
			func() bool { return false }, // isLocked
			"v0.1.0",
			"",
		)

		req := httptest.NewRequest(http.MethodGet, "/settings", nil)
		w := httptest.NewRecorder()

		controller.Get(w, req)

		Ok(t, nil)
		Equals(t, http.StatusOK, w.Code)
		Assert(t, strings.Contains(w.Body.String(), "Settings"), "should contain Settings title")
		Assert(t, strings.Contains(w.Body.String(), "v0.1.0"), "should contain version")
	})

	t.Run("shows apply lock state when enabled", func(t *testing.T) {
		tmpl := web_templates.SettingsTemplate
		controller := controllers.NewSettingsController(
			tmpl,
			true,                         // globalApplyLockEnabled
			func() bool { return true },  // isLocked
			"v0.1.0",
			"",
		)

		req := httptest.NewRequest(http.MethodGet, "/settings", nil)
		w := httptest.NewRecorder()

		controller.Get(w, req)

		Equals(t, http.StatusOK, w.Code)
		Assert(t, strings.Contains(w.Body.String(), "Enable Apply"), "should show enable button when locked")
	})
}
