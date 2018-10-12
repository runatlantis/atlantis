package server

import (
	"fmt"
	"net/url"
	"strings"
)

// NormalizeBaseURL ensures the given URL is a valid base URL for Atlantis.
//
// URLs that are fundamentally invalid (e.g. "hi") will return an error.
// Otherwise, the returned URL will have no trailing slashes and be guaranteed
// to be suitable for use as a base URL.
func NormalizeBaseURL(u *url.URL) (*url.URL, error) {
	if !u.IsAbs() {
		return nil, fmt.Errorf("Base URLs must be absolute.")
	}
	if !(u.Scheme == "http" || u.Scheme == "https") {
		return nil, fmt.Errorf("Base URLs must be HTTP or HTTPS.")
	}
	out := *u
	out.Path = strings.TrimRight(out.Path, "/")
	return &out, nil
}
