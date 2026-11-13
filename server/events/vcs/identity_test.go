// Copyright 2026 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package vcs_test

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"testing/synctest"
	"time"

	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/events/vcs"
	"github.com/runatlantis/atlantis/server/events/vcs/azuredevops"
	"github.com/runatlantis/atlantis/server/events/vcs/bitbucketcloud"
	"github.com/runatlantis/atlantis/server/events/vcs/bitbucketserver"
	"github.com/runatlantis/atlantis/server/events/vcs/gitea"
	"github.com/runatlantis/atlantis/server/events/vcs/github"
	atlantisgitlab "github.com/runatlantis/atlantis/server/events/vcs/gitlab"
	"github.com/runatlantis/atlantis/server/logging"
	. "github.com/runatlantis/atlantis/testing"
	gitlab "gitlab.com/gitlab-org/api/client-go"
)

func newIdentityClient(t *testing.T, host models.VCSHostType, configuredUser, identity string, status int) vcs.Client {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/api/v1/version" {
			_, _ = fmt.Fprint(w, `{"version":"1.25.4"}`)
			return
		}
		w.WriteHeader(status)
		switch host {
		case models.Gitlab:
			Equals(t, "/api/v4/user", r.URL.Path)
			_, _ = fmt.Fprintf(w, `{"username":%q}`, identity)
		case models.Gitea:
			Equals(t, "/api/v1/user", r.URL.Path)
			_, _ = fmt.Fprintf(w, `{"login":%q}`, identity)
		case models.BitbucketCloud:
			Equals(t, "/2.0/user", r.URL.Path)
			apiUser, token, _ := r.BasicAuth()
			Equals(t, "api@example.com", apiUser)
			Equals(t, "token", token)
			_, _ = fmt.Fprintf(w, `{"account_id":"different-from-uuid","uuid":%q}`, identity)
		case models.AzureDevops:
			Equals(t, "/owner/_apis/connectionData", r.URL.Path)
			_, token, _ := r.BasicAuth()
			Equals(t, "token", token)
			_, _ = fmt.Fprintf(w, `{"authenticatedUser":{"id":%q},"authorizedUser":{"id":"different-user"}}`, identity)
		default:
			t.Errorf("unexpected identity request for %s", host)
		}
	}))
	t.Cleanup(server.Close)
	return newIdentityClientForServer(t, host, configuredUser, server.URL, server.Client())
}

func newIdentityClientForServer(t *testing.T, host models.VCSHostType, configuredUser, baseURL string, httpClient *http.Client) vcs.Client {
	t.Helper()
	logger := logging.NewNoopLogger(t)
	var client vcs.Client
	var err error
	switch host {
	case models.Github:
		client, err = github.New("github.com", &github.UserCredentials{User: configuredUser, Token: "token"}, github.Config{}, 0, logger)
	case models.BitbucketServer:
		client, err = bitbucketserver.NewClient(httpClient, configuredUser, "token", baseURL, "")
	case models.BitbucketCloud:
		bb := bitbucketcloud.New(httpClient, configuredUser, "token", "api@example.com", "")
		bb.BaseURL = baseURL
		client = bb
	case models.Gitea:
		client, err = gitea.New(baseURL, configuredUser, "token", 100, logger)
	case models.Gitlab:
		var sdk *gitlab.Client
		sdk, err = gitlab.NewClient("token", gitlab.WithBaseURL(baseURL), gitlab.WithHTTPClient(httpClient), gitlab.WithCustomRetryMax(0))
		client = &atlantisgitlab.Client{Client: sdk}
	case models.AzureDevops:
		var ado *azuredevops.Client
		ado, err = azuredevops.New("dev.azure.com", configuredUser, "token")
		parsedURL, parseErr := url.Parse(baseURL + "/")
		Ok(t, parseErr)
		ado.Client.BaseURL = *parsedURL
		client = ado
	}
	Ok(t, err)
	return client
}

func TestIsAtlantisUser(t *testing.T) {
	for _, host := range []models.VCSHostType{models.Github, models.Gitlab, models.BitbucketCloud, models.BitbucketServer, models.AzureDevops, models.Gitea} {
		t.Run(host.String(), func(t *testing.T) {
			client := newIdentityClient(t, host, "atlantis", "atlantis", http.StatusOK)
			repo := models.Repo{FullName: "owner/project/repo", VCSHost: models.VCSHost{Type: host}}
			for _, c := range []struct {
				author string
				self   bool
			}{
				{"atlantis", true}, {"human", false}, {"other-atlantis[bot]", false}, {"", false},
			} {
				t.Run(c.author, func(t *testing.T) {
					self, err := client.IsAtlantisUser(repo, models.User{Username: c.author, ID: c.author})
					Ok(t, err)
					Equals(t, c.self, self)
				})
			}
			if host == models.BitbucketCloud {
				// Regression for the previous package-wide UUID cache.
				other := newIdentityClient(t, host, "other", "other", http.StatusOK)
				self, err := other.IsAtlantisUser(repo, models.User{Username: "atlantis", ID: "atlantis"})
				Ok(t, err)
				Equals(t, false, self)
			}
		})
	}
}

func TestIsAtlantisUser_MissingIdentity(t *testing.T) {
	for _, host := range []models.VCSHostType{models.Github, models.Gitlab, models.BitbucketCloud, models.BitbucketServer, models.AzureDevops, models.Gitea} {
		t.Run(host.String(), func(t *testing.T) {
			client := newIdentityClient(t, host, "", "", http.StatusOK)
			for _, author := range []string{"", "human"} {
				self, _ := client.IsAtlantisUser(models.Repo{FullName: "owner/project/repo"}, models.User{Username: author, ID: author})
				Equals(t, false, self)
			}
		})
	}
}

func TestIsAtlantisUser_AuthenticationError(t *testing.T) {
	for _, host := range []models.VCSHostType{models.Gitlab, models.BitbucketCloud, models.AzureDevops, models.Gitea} {
		t.Run(host.String(), func(t *testing.T) {
			client := newIdentityClient(t, host, "atlantis", "atlantis", http.StatusUnauthorized)
			self, err := client.IsAtlantisUser(models.Repo{FullName: "owner/project/repo"}, models.User{Username: "atlantis", ID: "atlantis"})
			Assert(t, err != nil, "lookup error must be returned")
			Equals(t, false, self)
		})
	}
}

func TestIsAtlantisUser_StableID(t *testing.T) {
	for _, host := range []models.VCSHostType{models.BitbucketCloud, models.AzureDevops} {
		t.Run(host.String(), func(t *testing.T) {
			client := newIdentityClient(t, host, "configured-name", "account-id", http.StatusOK)
			for _, c := range []struct {
				user models.User
				self bool
			}{
				{models.User{Username: "renamed-account", ID: "account-id"}, true},
				{models.User{Username: "configured-name", ID: "other-id"}, false},
				{models.User{Username: "configured-name"}, false},
			} {
				self, err := client.IsAtlantisUser(models.Repo{FullName: "owner/project/repo"}, c.user)
				Ok(t, err)
				Equals(t, c.self, self)
			}
		})
	}
}

func TestIsAtlantisUser_CachesOnlySuccessfulLookups(t *testing.T) {
	for _, host := range []models.VCSHostType{models.Gitlab, models.Gitea, models.BitbucketCloud, models.AzureDevops} {
		for _, failure := range []string{"forbidden", "missing identity", "malformed response"} {
			t.Run(host.String()+"/"+failure, func(t *testing.T) {
				var requests atomic.Int32
				var otherAccount atomic.Bool
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "application/json")
					if r.URL.Path == "/api/v1/version" {
						_, _ = fmt.Fprint(w, `{"version":"1.25.4"}`)
						return
					}
					if requests.Add(1) == 1 {
						switch failure {
						case "forbidden":
							w.WriteHeader(http.StatusForbidden)
							_, _ = fmt.Fprint(w, `{}`)
						case "missing identity":
							_, _ = fmt.Fprint(w, `{}`)
						case "malformed response":
							_, _ = fmt.Fprint(w, `{`)
						}
						return
					}
					identity := "self"
					if otherAccount.Load() {
						identity = "other"
					}
					_, _ = fmt.Fprint(w, identityResponse(identity))
				}))
				t.Cleanup(server.Close)
				client := newIdentityClientForServer(t, host, "configured", server.URL, server.Client())
				repo := models.Repo{FullName: "owner/project/repo"}
				user := models.User{Username: "self", ID: "self"}
				self, _ := client.IsAtlantisUser(repo, user)
				Equals(t, false, self)
				var wg sync.WaitGroup
				for range 12 {
					wg.Go(func() {
						self, err := client.IsAtlantisUser(repo, user)
						Ok(t, err)
						Equals(t, true, self)
					})
				}
				wg.Wait()
				Equals(t, int32(2), requests.Load())
				otherAccount.Store(true)
				self, err := client.IsAtlantisUser(repo, user)
				Ok(t, err)
				Equals(t, true, self)
				Equals(t, int32(2), requests.Load())
				other := newIdentityClientForServer(t, host, "configured", server.URL, server.Client())
				self, err = other.IsAtlantisUser(repo, user)
				Ok(t, err)
				Equals(t, false, self)
				Equals(t, int32(3), requests.Load())
			})
		}
	}
}

func TestIsAtlantisUser_AzureOrganizationCache(t *testing.T) {
	requests := make(map[string]int)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests[r.URL.Path]++
		owner, _, _ := strings.Cut(strings.TrimPrefix(r.URL.Path, "/"), "/")
		_, _ = fmt.Fprint(w, identityResponse(owner))
	}))
	t.Cleanup(server.Close)
	client := newIdentityClientForServer(t, models.AzureDevops, "configured", server.URL, server.Client())
	for range 2 {
		for _, owner := range []string{"one", "two"} {
			self, err := client.IsAtlantisUser(models.Repo{FullName: owner + "/project/repo"}, models.User{ID: owner})
			Ok(t, err)
			Equals(t, true, self)
		}
	}
	Equals(t, map[string]int{"/one/_apis/connectionData": 1, "/two/_apis/connectionData": 1}, requests)
}

type identityRoundTripFunc func(*http.Request) (*http.Response, error)

func (f identityRoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func identityResponse(identity string) string {
	return fmt.Sprintf(`{"username":%q,"login":%q,"uuid":%q,"authenticatedUser":{"id":%q}}`, identity, identity, identity, identity)
}

func TestIsAtlantisUser_Timeout(t *testing.T) {
	for _, host := range []models.VCSHostType{models.Gitlab, models.Gitea, models.BitbucketCloud, models.AzureDevops} {
		t.Run(host.String(), func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				previous := http.DefaultTransport
				t.Cleanup(func() { http.DefaultTransport = previous })
				stalled := true
				requests := 0
				http.DefaultTransport = identityRoundTripFunc(func(req *http.Request) (*http.Response, error) {
					body := identityResponse("self")
					if req.URL.Path == "/api/v1/version" {
						body = `{"version":"1.25.4"}`
					} else {
						requests++
						_, bounded := req.Context().Deadline()
						Assert(t, bounded, "identity lookup must have a deadline")
						if stalled {
							<-req.Context().Done()
							return nil, req.Context().Err()
						}
					}
					return &http.Response{StatusCode: http.StatusOK, Header: http.Header{"Content-Type": {"application/json"}}, Body: io.NopCloser(strings.NewReader(body))}, nil
				})
				client := newIdentityClientForServer(t, host, "configured", "http://identity.test", &http.Client{})
				repo := models.Repo{FullName: "owner/project/repo"}
				user := models.User{Username: "self", ID: "self"}
				start := time.Now()
				self, err := client.IsAtlantisUser(repo, user)
				Assert(t, errors.Is(err, context.DeadlineExceeded), "expected lookup timeout, got %v", err)
				Equals(t, false, self)
				Equals(t, 10*time.Second, time.Since(start))
				stalled = false
				self, err = client.IsAtlantisUser(repo, user)
				Ok(t, err)
				Equals(t, true, self)
				Equals(t, 2, requests)
			})
		})
	}
}

func TestIsAtlantisUser_SharesStalledLookup(t *testing.T) {
	for _, host := range []models.VCSHostType{models.Gitlab, models.Gitea, models.BitbucketCloud, models.AzureDevops} {
		t.Run(host.String(), func(t *testing.T) {
			previous := http.DefaultTransport
			t.Cleanup(func() { http.DefaultTransport = previous })
			var requests atomic.Int32
			firstStarted := make(chan struct{})
			firstFinished := make(chan struct{})
			stalled := true
			http.DefaultTransport = identityRoundTripFunc(func(req *http.Request) (*http.Response, error) {
				body := identityResponse("self")
				switch req.URL.Path {
				case "/api/v1/version":
					body = `{"version":"1.25.4"}`
				case "/other/_apis/connectionData":
					body = identityResponse("other")
				default:
					first := requests.Add(1) == 1
					if stalled {
						if first {
							close(firstStarted)
							defer close(firstFinished)
						}
						// Use a short real deadline so a regression to a locked
						// lookup also lets the test finish.
						ctx, cancel := context.WithTimeout(req.Context(), 200*time.Millisecond)
						defer cancel()
						<-ctx.Done()
						return nil, ctx.Err()
					}
				}
				return &http.Response{StatusCode: http.StatusOK, Header: http.Header{"Content-Type": {"application/json"}}, Body: io.NopCloser(strings.NewReader(body))}, nil
			})
			client := newIdentityClientForServer(t, host, "configured", "http://identity.test", &http.Client{})
			repo := models.Repo{FullName: "owner/project/repo"}
			user := models.User{Username: "self", ID: "self"}
			var wg sync.WaitGroup
			var selves [3]bool
			var errs [3]error
			start := make(chan struct{})
			for i := range errs {
				wg.Go(func() {
					<-start
					selves[i], errs[i] = client.IsAtlantisUser(repo, user)
				})
			}
			close(start)
			if host == models.AzureDevops {
				<-firstStarted
				otherRepo := models.Repo{FullName: "other/project/repo"}
				self, err := client.IsAtlantisUser(otherRepo, models.User{ID: "other"})
				if err != nil || !self {
					t.Errorf("lookup for another organization returned self=%t, err=%v", self, err)
				}
				select {
				case <-firstFinished:
					t.Error("lookup for another organization waited for the stalled lookup")
				default:
				}
			}
			wg.Wait()
			for i, err := range errs {
				Assert(t, errors.Is(err, context.DeadlineExceeded), "expected lookup timeout, got %v", err)
				Equals(t, false, selves[i])
			}
			Equals(t, int32(1), requests.Load())

			stalled = false
			for range 2 {
				self, err := client.IsAtlantisUser(repo, user)
				Ok(t, err)
				Equals(t, true, self)
			}
			Equals(t, int32(2), requests.Load())
		})
	}
}
