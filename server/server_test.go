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
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/gorilla/mux"
	. "github.com/petergtz/pegomock"
	"github.com/runatlantis/atlantis/server"
	"github.com/runatlantis/atlantis/server/controllers/templates"
	tMocks "github.com/runatlantis/atlantis/server/controllers/templates/mocks"
	"github.com/runatlantis/atlantis/server/core/locking/mocks"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/logging"
	. "github.com/runatlantis/atlantis/testing"
)

func TestNewServer(t *testing.T) {
	t.Log("Run through NewServer constructor")
	tmpDir, err := os.MkdirTemp("", "")
	Ok(t, err)
	_, err = server.NewServer(server.UserConfig{
		DataDir:     tmpDir,
		AtlantisURL: "http://example.com",
	}, server.Config{})
	Ok(t, err)
}

// todo: test what happens if we set different flags. The generated config should be different.

func TestNewServer_InvalidAtlantisURL(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "")
	Ok(t, err)
	_, err = server.NewServer(server.UserConfig{
		DataDir:     tmpDir,
		AtlantisURL: "example.com",
	}, server.Config{
		AtlantisURLFlag: "atlantis-url",
	})
	ErrEquals(t, "parsing --atlantis-url flag \"example.com\": http or https must be specified", err)
}

func TestIndex_LockErr(t *testing.T) {
	t.Log("index should return a 503 if unable to list locks")
	RegisterMockTestingT(t)
	l := mocks.NewMockLocker()
	When(l.List()).ThenReturn(nil, errors.New("err"))
	s := server.Server{
		Locker: l,
	}
	req, _ := http.NewRequest("GET", "", bytes.NewBuffer(nil))
	w := httptest.NewRecorder()
	s.Index(w, req)
	ResponseContains(t, w, 503, "Could not retrieve locks: err")
}

func TestIndex_Success(t *testing.T) {
	t.Log("Index should render the index template successfully.")
	RegisterMockTestingT(t)
	l := mocks.NewMockLocker()
	al := mocks.NewMockApplyLocker()
	// These are the locks that we expect to be rendered.
	now := time.Now()
	locks := map[string]models.ProjectLock{
		"lkysow/atlantis-example/./default": {
			Pull: models.PullRequest{
				Num: 9,
			},
			Project: models.Project{
				RepoFullName: "lkysow/atlantis-example",
			},
			Time: now,
		},
	}
	When(l.List()).ThenReturn(locks, nil)
	it := tMocks.NewMockTemplateWriter()
	r := mux.NewRouter()
	atlantisVersion := "0.3.1"
	// Need to create a lock route since the server expects this route to exist.
	r.NewRoute().Path("/lock").
		Queries("id", "{id}").Name(server.LockViewRouteName)
	u, err := url.Parse("https://example.com")
	Ok(t, err)
	s := server.Server{
		Locker:          l,
		ApplyLocker:     al,
		IndexTemplate:   it,
		Router:          r,
		AtlantisVersion: atlantisVersion,
		AtlantisURL:     u,
		Logger:          logging.NewNoopLogger(t),
	}
	req, _ := http.NewRequest("GET", "", bytes.NewBuffer(nil))
	w := httptest.NewRecorder()
	s.Index(w, req)
	it.VerifyWasCalledOnce().Execute(w, templates.IndexData{
		ApplyLock: templates.ApplyLockData{
			Locked:        false,
			Time:          time.Time{},
			TimeFormatted: "01-01-0001 00:00:00",
		},
		Locks: []templates.LockIndexData{
			{
				LockPath:      "/lock?id=lkysow%252Fatlantis-example%252F.%252Fdefault",
				RepoFullName:  "lkysow/atlantis-example",
				PullNum:       9,
				Time:          now,
				TimeFormatted: now.Format("02-01-2006 15:04:05"),
			},
		},
		AtlantisVersion: atlantisVersion,
	})
	ResponseContains(t, w, http.StatusOK, "")
}

func TestHealthz(t *testing.T) {
	s := server.Server{}
	req, _ := http.NewRequest("GET", "/healthz", bytes.NewBuffer(nil))
	w := httptest.NewRecorder()
	s.Healthz(w, req)
	Equals(t, http.StatusOK, w.Result().StatusCode)
	body, _ := io.ReadAll(w.Result().Body)
	Equals(t, "application/json", w.Result().Header["Content-Type"][0])
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
	for i := 0; i < b.N; i++ {
		s.Healthz(w, nil)
	}
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
