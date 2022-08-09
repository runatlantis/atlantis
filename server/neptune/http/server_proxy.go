package http

import (
	"github.com/runatlantis/atlantis/server/logging"
	"net/http"
)

type ServerProxy struct {
	*http.Server
	SSLCertFile string
	SSLKeyFile  string
	Logger      logging.Logger
}

func (p *ServerProxy) ListenAndServe() error {
	if p.SSLCertFile != "" && p.SSLKeyFile != "" {
		return p.Server.ListenAndServeTLS(p.SSLCertFile, p.SSLKeyFile)
	} else {
		return p.Server.ListenAndServe()
	}
}
