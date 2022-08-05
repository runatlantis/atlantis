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

func (p *ServerProxy) ListenAndServe() {
	var err error
	if p.SSLCertFile != "" && p.SSLKeyFile != "" {
		err = p.Server.ListenAndServeTLS(p.SSLCertFile, p.SSLKeyFile)
	} else {
		err = p.Server.ListenAndServe()
	}
	if err != nil && err != http.ErrServerClosed {
		p.Logger.Error(err.Error())
	}
}
