// +build !debug

package azuredevops

import "net/http"

// Idea from:
// https://dave.cheney.net/2014/09/28/using-build-to-switch-between-debug-and-release
func debug(fmt string, args ...interface{}) {}

func debugReq(*http.Request) {}
