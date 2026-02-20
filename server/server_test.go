// Copyright 2017 HootSuite Media Inc.
//
// Licensed under the Apache License, Version 2.0 (the License);
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//    http://www.apache.org/licenses/LICENSE-2.0
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an AS IS BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
// Modified hereafter by contributors to runatlantis/atlantis.

package server_test

import (
	"bytes"
	"crypto/tls"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/runatlantis/atlantis/cmd"
	"github.com/runatlantis/atlantis/server"
	. "github.com/runatlantis/atlantis/testing"
)

const (
	testAtlantisVersion = "1.0.0"
	testAtlantisUrl     = "http://example.com"
	testLockingDBType   = cmd.DefaultLockingDBType
	testGitHubHostName  = cmd.DefaultGHHostname
	testGitHubUser      = "user"
)

func TestNewServer_GitHubUser(t *testing.T) {
	t.Log("Run through NewServer constructor")
	tmpDir := t.TempDir()
	_, err := server.NewServer(
		server.UserConfig{
			DataDir:        tmpDir,
			AtlantisURL:    testAtlantisUrl,
			LockingDBType:  testLockingDBType,
			GithubHostname: testGitHubHostName,
			GithubUser:     testGitHubUser,
		}, server.Config{
			AtlantisVersion: testAtlantisVersion,
		},
	)
	Ok(t, err)
}

// todo: test what happens if we set different flags. The generated config should be different.

func TestNewServer_InvalidAtlantisURL(t *testing.T) {
	tmpDir := t.TempDir()
	_, err := server.NewServer(server.UserConfig{
		DataDir:     tmpDir,
		AtlantisURL: "example.com",
	}, server.Config{
		AtlantisURLFlag: "atlantis-url",
	})
	ErrEquals(t, "parsing --atlantis-url flag \"example.com\": http or https must be specified", err)
}

func TestHealthz(t *testing.T) {
	s := server.Server{}
	req, _ := http.NewRequest("GET", "/healthz", bytes.NewBuffer(nil))
	w := httptest.NewRecorder()
	s.Healthz(w, req)

	resp := w.Result()
	defer resp.Body.Close()
	Equals(t, http.StatusOK, resp.StatusCode)
	body, _ := io.ReadAll(resp.Body)
	Equals(t, "application/json", resp.Header["Content-Type"][0])
	Equals(t,
		`{
  "status": "ok"
}`, string(body))
}

type mockRW struct{}

var _ http.ResponseWriter = mockRW{}
var mh = http.Header{}

func (w mockRW) WriteHeader(int)           {}
func (w mockRW) Write([]byte) (int, error) { return 0, nil }
func (w mockRW) Header() http.Header       { return mh }

var w = mockRW{}
var s = &server.Server{}

func BenchmarkHealthz(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		s.Healthz(w, nil)
	}
}

func TestGetCertificate(t *testing.T) {
	s := server.Server{}
	clientHelloInfo := &tls.ClientHelloInfo{}

	// Initial certificate load
	s.SSLCertFile = "../testdata/cert.pem"
	s.SSLKeyFile = "../testdata/key.pem"
	cert, err := s.GetSSLCertificate(clientHelloInfo)
	Ok(t, err)

	// Certificate reload
	s.SSLCertFile = "../testdata/cert2.pem"
	s.SSLKeyFile = "../testdata/key2.pem"
	s.CertLastRefreshTime = s.CertLastRefreshTime.Add(-1 * time.Second)
	s.KeyLastRefreshTime = s.KeyLastRefreshTime.Add(-1 * time.Second)
	newCert, err := s.GetSSLCertificate(clientHelloInfo)

	Ok(t, err)
	Assert(
		t,
		!bytes.Equal(bytes.Join(cert.Certificate, nil), bytes.Join(newCert.Certificate, nil)),
		"Certificate expected to rotate")
}

func TestParseAtlantisURL(t *testing.T) {
	cases := []struct {
		In     string
		ExpErr string
		ExpURL string
	}{
		// Valid URLs should work.
		{
			In:     "https://example.com",
			ExpURL: "https://example.com",
		},
		{
			In:     "http://example.com",
			ExpURL: "http://example.com",
		},
		{
			In:     "http://example.com/",
			ExpURL: "http://example.com",
		},
		{
			In:     "http://example.com",
			ExpURL: "http://example.com",
		},
		{
			In:     "http://example.com:4141",
			ExpURL: "http://example.com:4141",
		},
		{
			In:     "http://example.com:4141/",
			ExpURL: "http://example.com:4141",
		},
		{
			In:     "http://example.com/baseurl",
			ExpURL: "http://example.com/baseurl",
		},
		{
			In:     "http://example.com/baseurl/",
			ExpURL: "http://example.com/baseurl",
		},
		{
			In:     "http://example.com/baseurl/test",
			ExpURL: "http://example.com/baseurl/test",
		},

		// Must be valid URL.
		{
			In:     "::",
			ExpErr: "parse \"::\": missing protocol scheme",
		},

		// Must be absolute.
		{
			In:     "/hi",
			ExpErr: "http or https must be specified",
		},

		// Must have http or https scheme..
		{
			In:     "localhost/test",
			ExpErr: "http or https must be specified",
		},
		{
			In:     "http0://localhost/test",
			ExpErr: "http or https must be specified",
		},
	}

	for _, c := range cases {
		t.Run(c.In, func(t *testing.T) {
			act, err := server.ParseAtlantisURL(c.In)
			if c.ExpErr != "" {
				ErrEquals(t, c.ExpErr, err)
			} else {
				Ok(t, err)
				Equals(t, c.ExpURL, act.String())
			}
		})
	}
}
