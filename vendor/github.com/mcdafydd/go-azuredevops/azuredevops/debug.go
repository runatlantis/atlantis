// +build debug

package azuredevops

import (
	"log"
	"net/http"
	"net/http/httputil"
)

// Idea from:
// https://dave.cheney.net/2014/09/28/using-build-to-switch-between-debug-and-release
func debug(fmt string, args ...interface{}) {
	log.Printf(fmt, args...)
}

func debugReq(req *http.Request) {
	dump, err := httputil.DumpRequest(req, true)
	debug(string(dump), err)
}
