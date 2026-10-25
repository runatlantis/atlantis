// Copyright 2026 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package controllers_test

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/runatlantis/atlantis/server/controllers"
	"github.com/runatlantis/atlantis/server/events/vcs/github"
	"github.com/runatlantis/atlantis/server/logging"
	. "github.com/runatlantis/atlantis/testing"
)

// fakeExchanger records calls made to the GitHub manifest conversion API.
type fakeExchanger struct {
	mu    sync.Mutex
	calls []string
}

func (f *fakeExchanger) ExchangeCode(_ logging.SimpleLogging, code string) (*github.GithubAppTemporarySecrets, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.calls = append(f.calls, code)
	return &github.GithubAppTemporarySecrets{
		ID:            42,
		Key:           "-----BEGIN RSA PRIVATE KEY-----\nfake\n-----END RSA PRIVATE KEY-----",
		Name:          "atlantis-test",
		WebhookSecret: "fake-webhook-secret",
		URL:           "https://github.com/apps/atlantis-test",
	}, nil
}

func (f *fakeExchanger) Calls() []string {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]string(nil), f.calls...)
}

func newGithubAppController(t *testing.T) (*controllers.GithubAppController, *fakeExchanger) {
	t.Helper()
	atlantisURL, err := url.Parse("https://atlantis.example.com")
	Ok(t, err)
	exchanger := &fakeExchanger{}
	return &controllers.GithubAppController{
		AtlantisURL:    atlantisURL,
		Logger:         logging.NewNoopLogger(t),
		GithubHostname: "github.com",
		Exchanger:      exchanger,
	}, exchanger
}

var stateInTarget = regexp.MustCompile(`settings/apps/new\?state=([A-Za-z0-9_-]+)"`)

// setupFlow is what a browser holds after opening the setup page: the state
// GitHub will echo back and the flow-binding cookie Atlantis set.
type setupFlow struct {
	state  string
	cookie *http.Cookie
}

func newSetup(c *controllers.GithubAppController) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", "/github-app/setup", nil)
	c.New(w, r)
	return w
}

// startSetup renders the setup page and returns the state nonce embedded in
// the manifest form's target URL together with the binding cookie.
func startSetup(t *testing.T, c *controllers.GithubAppController) setupFlow {
	t.Helper()
	w := newSetup(c)
	Equals(t, http.StatusOK, w.Result().StatusCode)
	match := stateInTarget.FindStringSubmatch(w.Body.String())
	Assert(t, len(match) == 2, "expected setup form target to carry a state parameter, body:\n%s", w.Body.String())
	cookie := findCookie(w, controllers.GithubAppSetupCookieName)
	Assert(t, cookie != nil, "expected setup page to set the %s cookie", controllers.GithubAppSetupCookieName)
	return setupFlow{state: match[1], cookie: cookie}
}

func findCookie(w *httptest.ResponseRecorder, name string) *http.Cookie {
	for _, c := range w.Result().Cookies() {
		if c.Name == name {
			return c
		}
	}
	return nil
}

// exchange performs the GitHub redirect back to Atlantis, presenting cookie
// when it is not nil.
func exchange(c *controllers.GithubAppController, query string, cookie *http.Cookie) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", "/github-app/exchange-code?"+query, nil)
	if cookie != nil {
		r.AddCookie(&http.Cookie{Name: cookie.Name, Value: cookie.Value})
	}
	c.ExchangeCode(w, r)
	return w
}

func (f setupFlow) callback(c *controllers.GithubAppController, code string) *httptest.ResponseRecorder {
	return exchange(c, "code="+code+"&state="+f.state, f.cookie)
}

func TestGithubAppController_New_EmbedsStateInFormTarget(t *testing.T) {
	c, _ := newGithubAppController(t)
	flow := startSetup(t, c)
	Assert(t, len(flow.state) >= 32, "expected a long random state, got %q", flow.state)
}

func TestGithubAppController_New_GeneratesFreshStatePerRequest(t *testing.T) {
	c, _ := newGithubAppController(t)
	first := startSetup(t, c)
	second := startSetup(t, c)
	Assert(t, first.state != second.state, "expected each setup page to carry a fresh state")
	Assert(t, first.cookie.Value != second.cookie.Value, "expected each setup page to carry a fresh binding")
}

func TestGithubAppController_New_SetsBindingCookie(t *testing.T) {
	c, _ := newGithubAppController(t)
	flow := startSetup(t, c)

	Assert(t, flow.cookie.HttpOnly, "expected HttpOnly cookie")
	Assert(t, flow.cookie.Secure, "expected Secure cookie for an https Atlantis URL")
	Equals(t, http.SameSiteLaxMode, flow.cookie.SameSite)
	Equals(t, "/github-app/", flow.cookie.Path)
	Assert(t, flow.cookie.MaxAge > 0, "expected the cookie to expire with the flow")
	Assert(t, flow.cookie.Value != flow.state, "expected the binding to differ from the state")
	Assert(t, len(flow.cookie.Value) >= 32, "expected a long random binding, got %q", flow.cookie.Value)
}

func TestGithubAppController_New_CookieNotSecureForPlainHTTP(t *testing.T) {
	c, _ := newGithubAppController(t)
	plain, err := url.Parse("http://atlantis.example.com/basepath")
	Ok(t, err)
	c.AtlantisURL = plain

	flow := startSetup(t, c)

	Assert(t, !flow.cookie.Secure, "expected no Secure flag when neither the request nor the Atlantis URL use TLS")
	Equals(t, "/basepath/github-app/", flow.cookie.Path)
}

func TestGithubAppController_ExchangeCode_HappyPath(t *testing.T) {
	c, exchanger := newGithubAppController(t)
	flow := startSetup(t, c)

	w := flow.callback(c, "abc123")

	Equals(t, http.StatusOK, w.Result().StatusCode)
	Equals(t, []string{"abc123"}, exchanger.Calls())
	body := w.Body.String()
	Assert(t, strings.Contains(body, "Github app created successfully"), "expected success page, got:\n%s", body)
	Assert(t, strings.Contains(body, "fake-webhook-secret"), "expected webhook secret in success page")
	cleared := findCookie(w, controllers.GithubAppSetupCookieName)
	Assert(t, cleared != nil && cleared.MaxAge < 0, "expected the binding cookie to be cleared after use")
}

func TestGithubAppController_ExchangeCode_StateWithoutCookieIsRejected(t *testing.T) {
	c, exchanger := newGithubAppController(t)
	flow := startSetup(t, c)

	w := exchange(c, "code=abc123&state="+flow.state, nil)

	Equals(t, http.StatusBadRequest, w.Result().StatusCode)
	Equals(t, 0, len(exchanger.Calls()))
	// The admin's own browser can still finish the flow.
	Equals(t, http.StatusOK, flow.callback(c, "abc123").Result().StatusCode)
	Equals(t, []string{"abc123"}, exchanger.Calls())
}

func TestGithubAppController_ExchangeCode_StateWithOtherBrowsersCookieIsRejected(t *testing.T) {
	c, exchanger := newGithubAppController(t)
	browserA := startSetup(t, c)
	browserB := startSetup(t, c)

	w := exchange(c, "code=abc123&state="+browserA.state, browserB.cookie)

	Equals(t, http.StatusBadRequest, w.Result().StatusCode)
	Equals(t, 0, len(exchanger.Calls()))
	Equals(t, http.StatusOK, browserA.callback(c, "abc123").Result().StatusCode)
	Equals(t, []string{"abc123"}, exchanger.Calls())
}

func TestGithubAppController_ExchangeCode_ForgedCookieIsRejected(t *testing.T) {
	c, exchanger := newGithubAppController(t)
	flow := startSetup(t, c)
	forged := &http.Cookie{Name: controllers.GithubAppSetupCookieName, Value: "forged"}

	w := exchange(c, "code=abc123&state="+flow.state, forged)

	Equals(t, http.StatusBadRequest, w.Result().StatusCode)
	Equals(t, 0, len(exchanger.Calls()))
}

func TestGithubAppController_ExchangeCode_MissingStateKeepsFlow(t *testing.T) {
	c, exchanger := newGithubAppController(t)
	flow := startSetup(t, c)

	w := exchange(c, "code=abc123", flow.cookie)

	Equals(t, http.StatusBadRequest, w.Result().StatusCode)
	Equals(t, 0, len(exchanger.Calls()))
	Equals(t, http.StatusOK, flow.callback(c, "abc123").Result().StatusCode)
}

func TestGithubAppController_ExchangeCode_WrongStateKeepsFlow(t *testing.T) {
	c, exchanger := newGithubAppController(t)
	flow := startSetup(t, c)

	w := exchange(c, "code=abc123&state=not-the-right-one", flow.cookie)

	Equals(t, http.StatusBadRequest, w.Result().StatusCode)
	Equals(t, 0, len(exchanger.Calls()))
	Equals(t, http.StatusOK, flow.callback(c, "abc123").Result().StatusCode)
	Equals(t, []string{"abc123"}, exchanger.Calls())
}

func TestGithubAppController_ExchangeCode_NoSetupStarted(t *testing.T) {
	c, exchanger := newGithubAppController(t)

	w := exchange(c, "code=abc123&state=anything", nil)

	Equals(t, http.StatusBadRequest, w.Result().StatusCode)
	Equals(t, 0, len(exchanger.Calls()))
}

func TestGithubAppController_ExchangeCode_ReusedState(t *testing.T) {
	c, exchanger := newGithubAppController(t)
	flow := startSetup(t, c)

	// The flow is spent by the first callback; replaying it must fail.
	Equals(t, http.StatusOK, flow.callback(c, "abc123").Result().StatusCode)

	second := flow.callback(c, "def456")
	Equals(t, http.StatusBadRequest, second.Result().StatusCode)
	Equals(t, []string{"abc123"}, exchanger.Calls())
}

func TestGithubAppController_ExchangeCode_FailureReturnsToIdle(t *testing.T) {
	c, _ := newGithubAppController(t)
	failing := &failingExchanger{}
	c.Exchanger = failing
	flow := startSetup(t, c)

	Equals(t, http.StatusInternalServerError, flow.callback(c, "abc123").Result().StatusCode)

	// The failed flow is consumed.
	Equals(t, http.StatusBadRequest, flow.callback(c, "abc123").Result().StatusCode)
	Equals(t, 1, failing.Calls())

	// A fresh setup can be started and reaches GitHub again.
	retry := startSetup(t, c)
	Equals(t, http.StatusInternalServerError, retry.callback(c, "def456").Result().StatusCode)
	Equals(t, 2, failing.Calls())
	Equals(t, false, c.SetupComplete())
}

func TestGithubAppController_ExchangeCode_OnlyOneExchangeAtATime(t *testing.T) {
	c, _ := newGithubAppController(t)
	blocking := newBlockingExchanger()
	c.Exchanger = blocking
	browserA := startSetup(t, c)
	browserB := startSetup(t, c)

	done := make(chan *httptest.ResponseRecorder)
	go func() { done <- browserA.callback(c, "abc123") }()
	select {
	case <-blocking.entered:
	case <-time.After(5 * time.Second):
		t.Fatal("exchanger was never called")
	}

	// While A is exchanging nothing else may start or finish a flow.
	Equals(t, http.StatusConflict, newSetup(c).Result().StatusCode)
	Equals(t, http.StatusConflict, browserB.callback(c, "def456").Result().StatusCode)
	Equals(t, http.StatusConflict, browserA.callback(c, "abc123").Result().StatusCode)

	close(blocking.release)
	Equals(t, http.StatusOK, (<-done).Result().StatusCode)
	Equals(t, 1, blocking.Calls())
	Equals(t, true, c.SetupComplete())

	// B's flow was never exchanged and setup is now closed.
	Equals(t, http.StatusBadRequest, browserB.callback(c, "def456").Result().StatusCode)
	Equals(t, 1, blocking.Calls())
}

func TestGithubAppController_ExchangeCode_EmptyCodeWritesSingleResponse(t *testing.T) {
	c, exchanger := newGithubAppController(t)
	flow := startSetup(t, c)

	w := exchange(c, "state="+flow.state, flow.cookie)

	Equals(t, http.StatusBadRequest, w.Result().StatusCode)
	Equals(t, 0, len(exchanger.Calls()))
	Equals(t, 1, strings.Count(w.Body.String(), "\n"))
	Assert(t, !strings.Contains(w.Body.String(), "<html"), "expected a plain error, not the setup page")
}

func TestGithubAppController_ExchangeCode_SecondExchangeAfterSuccessIsRejected(t *testing.T) {
	c, exchanger := newGithubAppController(t)
	flow := startSetup(t, c)
	Equals(t, http.StatusOK, flow.callback(c, "abc123").Result().StatusCode)

	// Starting setup again in the same process must be refused too.
	Equals(t, http.StatusBadRequest, newSetup(c).Result().StatusCode)

	second := flow.callback(c, "def456")
	Equals(t, http.StatusBadRequest, second.Result().StatusCode)
	Equals(t, []string{"abc123"}, exchanger.Calls())
}

func TestGithubAppController_SetupCompleteRejectsBothEndpoints(t *testing.T) {
	c, exchanger := newGithubAppController(t)
	c.GithubSetupComplete = true

	Equals(t, http.StatusBadRequest, newSetup(c).Result().StatusCode)
	Equals(t, http.StatusBadRequest, exchange(c, "code=abc123&state=x", nil).Result().StatusCode)
	Equals(t, 0, len(exchanger.Calls()))
}

type failingExchanger struct {
	mu    sync.Mutex
	calls int
}

func (f *failingExchanger) ExchangeCode(_ logging.SimpleLogging, _ string) (*github.GithubAppTemporarySecrets, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.calls++
	return nil, errBoom
}

func (f *failingExchanger) Calls() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.calls
}

// blockingExchanger signals entered on each call and then waits for release,
// so a test can act while an exchange is in flight.
type blockingExchanger struct {
	entered chan struct{}
	release chan struct{}
	mu      sync.Mutex
	calls   int
}

func newBlockingExchanger() *blockingExchanger {
	return &blockingExchanger{entered: make(chan struct{}, 10), release: make(chan struct{})}
}

func (b *blockingExchanger) ExchangeCode(_ logging.SimpleLogging, _ string) (*github.GithubAppTemporarySecrets, error) {
	b.mu.Lock()
	b.calls++
	b.mu.Unlock()
	b.entered <- struct{}{}
	<-b.release
	return &github.GithubAppTemporarySecrets{ID: 7, Name: "blocked"}, nil
}

func (b *blockingExchanger) Calls() int {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.calls
}

var errBoom = &exchangeError{}

type exchangeError struct{}

func (*exchangeError) Error() string { return "boom" }
