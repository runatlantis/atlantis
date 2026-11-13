// Copyright 2026 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package controllers

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/runatlantis/atlantis/server/events/vcs/github"
	"github.com/runatlantis/atlantis/server/logging"
	. "github.com/runatlantis/atlantis/testing"
)

type countingExchanger struct {
	calls int
}

func (c *countingExchanger) ExchangeCode(_ logging.SimpleLogging, _ string) (*github.GithubAppTemporarySecrets, error) {
	c.calls++
	return &github.GithubAppTemporarySecrets{ID: 1}, nil
}

func newClockedController(t *testing.T, now *time.Time) (*GithubAppController, *countingExchanger) {
	t.Helper()
	atlantisURL, err := url.Parse("https://atlantis.example.com")
	Ok(t, err)
	exchanger := &countingExchanger{}
	return &GithubAppController{
		AtlantisURL:    atlantisURL,
		Logger:         logging.NewNoopLogger(t),
		GithubHostname: "github.com",
		Exchanger:      exchanger,
		now:            func() time.Time { return *now },
	}, exchanger
}

// startInternalSetup calls New and returns the only pending flow's state and
// the binding cookie set on the response.
func startInternalSetup(t *testing.T, c *GithubAppController) (string, *http.Cookie) {
	t.Helper()
	w := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", "/github-app/setup", nil)
	before := map[string]bool{}
	for s := range c.flows {
		before[s] = true
	}
	c.New(w, r)
	Equals(t, http.StatusOK, w.Result().StatusCode)
	var state string
	for s := range c.flows {
		if !before[s] {
			state = s
		}
	}
	Assert(t, state != "", "expected New to store a pending flow")
	for _, ck := range w.Result().Cookies() {
		if ck.Name == GithubAppSetupCookieName {
			return state, ck
		}
	}
	t.Fatalf("expected New to set the %s cookie", GithubAppSetupCookieName)
	return "", nil
}

func callback(c *GithubAppController, state string, cookie *http.Cookie) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", "/github-app/exchange-code?code=abc123&state="+state, nil)
	r.AddCookie(&http.Cookie{Name: cookie.Name, Value: cookie.Value}) //nolint:gosec // G124: request cookies carry no attributes
	c.ExchangeCode(w, r)
	return w
}

func TestGithubAppController_ExchangeCode_ExpiredState(t *testing.T) {
	now := time.Date(2026, 10, 7, 12, 0, 0, 0, time.UTC)
	c, exchanger := newClockedController(t, &now)
	state, cookie := startInternalSetup(t, c)

	// Just past the TTL the flow must be refused without contacting GitHub.
	now = now.Add(githubAppStateTTL + time.Second)
	w := callback(c, state, cookie)

	Equals(t, http.StatusBadRequest, w.Result().StatusCode)
	Equals(t, 0, exchanger.calls)
	_, stillPending := c.flows[state]
	Assert(t, !stillPending, "expected the expired flow to be dropped")
}

func TestGithubAppController_ExchangeCode_StateWithinTTLIsAccepted(t *testing.T) {
	now := time.Date(2026, 10, 7, 12, 0, 0, 0, time.UTC)
	c, exchanger := newClockedController(t, &now)
	state, cookie := startInternalSetup(t, c)

	now = now.Add(githubAppStateTTL - time.Second)
	w := callback(c, state, cookie)

	Equals(t, http.StatusOK, w.Result().StatusCode)
	Equals(t, 1, exchanger.calls)
	Equals(t, true, c.SetupComplete())
}

func TestGithubAppController_New_LimitsPendingFlows(t *testing.T) {
	now := time.Date(2026, 10, 7, 12, 0, 0, 0, time.UTC)
	c, _ := newClockedController(t, &now)
	first, firstCookie := startInternalSetup(t, c)
	for i := 1; i < githubAppMaxPendingFlows; i++ {
		startInternalSetup(t, c)
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", "/github-app/setup", nil)
	c.New(w, r)
	Equals(t, http.StatusTooManyRequests, w.Result().StatusCode)
	Equals(t, githubAppMaxPendingFlows, len(c.flows))

	// Existing flows are not evicted by the flood.
	Equals(t, http.StatusOK, callback(c, first, firstCookie).Result().StatusCode)
}

func TestGithubAppController_New_PrunesExpiredFlows(t *testing.T) {
	now := time.Date(2026, 10, 7, 12, 0, 0, 0, time.UTC)
	c, _ := newClockedController(t, &now)
	for range githubAppMaxPendingFlows {
		startInternalSetup(t, c)
	}

	now = now.Add(githubAppStateTTL + time.Second)
	startInternalSetup(t, c)
	Equals(t, 1, len(c.flows))
}
