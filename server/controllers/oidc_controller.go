package controllers

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/runatlantis/atlantis/server/logging"
)

// OIDCController handles OIDC requests.
type OIDCController struct {
	AtlantisURL *url.URL
	Logger      logging.SimpleLogging
}

// respond is a helper function to respond and log the response. lvl is the log
// level to log at, code is the HTTP response code.
func (l *OIDCController) respond(w http.ResponseWriter, lvl logging.LogLevel, responseCode int, format string, args ...interface{}) {
	response := fmt.Sprintf(format, args...)
	l.Logger.Log(lvl, response)
	w.WriteHeader(responseCode)
	fmt.Fprintln(w, response)
}
