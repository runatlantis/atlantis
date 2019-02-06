package vcs_test

import (
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/events/vcs"
	. "github.com/runatlantis/atlantis/testing"
)

// GetModifiedFiles should make multiple requests if more than one page
// and concat results.
func TestGithubClient_GetModifiedFiles(t *testing.T) {
	respTemplate := `[
  {
    "sha": "bbcd538c8e72b8c175046e27cc8f907076331401",
    "filename": "%s",
    "status": "added",
    "additions": 103,
    "deletions": 21,
    "changes": 124,
    "blob_url": "https://github.com/octocat/Hello-World/blob/6dcb09b5b57875f334f61aebed695e2e4193db5e/file1.txt",
    "raw_url": "https://github.com/octocat/Hello-World/raw/6dcb09b5b57875f334f61aebed695e2e4193db5e/file1.txt",
    "contents_url": "https://api.github.com/repos/octocat/Hello-World/contents/file1.txt?ref=6dcb09b5b57875f334f61aebed695e2e4193db5e",
    "patch": "@@ -132,7 +132,7 @@ module Test @@ -1000,7 +1000,7 @@ module Test"
  }
]`
	firstResp := fmt.Sprintf(respTemplate, "file1.txt")
	secondResp := fmt.Sprintf(respTemplate, "file2.txt")
	testServer := httptest.NewTLSServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.RequestURI {
			// The first request should hit this URL.
			case "/api/v3/repos/owner/repo/pulls/1/files?per_page=300":
				// We write a header that means there's an additional page.
				w.Header().Add("Link", `<https://api.github.com/resource?page=2>; rel="next",
      <https://api.github.com/resource?page=2>; rel="last"`)
				w.Write([]byte(firstResp)) // nolint: errcheck
				return
				// The second should hit this URL.
			case "/api/v3/repos/owner/repo/pulls/1/files?page=2&per_page=300":
				w.Write([]byte(secondResp)) // nolint: errcheck
			default:
				t.Errorf("got unexpected request at %q", r.RequestURI)
				http.Error(w, "not found", http.StatusNotFound)
				return
			}
		}))

	testServerURL, err := url.Parse(testServer.URL)
	Ok(t, err)
	client, err := vcs.NewGithubClient(testServerURL.Host, "user", "pass")
	Ok(t, err)
	defer disableSSLVerification()()

	files, err := client.GetModifiedFiles(models.Repo{
		FullName:          "owner/repo",
		Owner:             "owner",
		Name:              "repo",
		CloneURL:          "",
		SanitizedCloneURL: "",
		VCSHost: models.VCSHost{
			Type:     models.Github,
			Hostname: "github.com",
		},
	}, models.PullRequest{
		Num: 1,
	})
	Ok(t, err)
	Equals(t, []string{"file1.txt", "file2.txt"}, files)
}

// GetModifiedFiles should include the source and destination of a moved
// file.
func TestGithubClient_GetModifiedFilesMovedFile(t *testing.T) {
	resp := `[
  {
    "sha": "bbcd538c8e72b8c175046e27cc8f907076331401",
    "filename": "new/filename.txt",
    "previous_filename": "previous/filename.txt",
    "status": "renamed",
    "additions": 103,
    "deletions": 21,
    "changes": 124,
    "blob_url": "https://github.com/octocat/Hello-World/blob/6dcb09b5b57875f334f61aebed695e2e4193db5e/file1.txt",
    "raw_url": "https://github.com/octocat/Hello-World/raw/6dcb09b5b57875f334f61aebed695e2e4193db5e/file1.txt",
    "contents_url": "https://api.github.com/repos/octocat/Hello-World/contents/file1.txt?ref=6dcb09b5b57875f334f61aebed695e2e4193db5e",
    "patch": "@@ -132,7 +132,7 @@ module Test @@ -1000,7 +1000,7 @@ module Test"
  }
]`
	testServer := httptest.NewTLSServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.RequestURI {
			// The first request should hit this URL.
			case "/api/v3/repos/owner/repo/pulls/1/files?per_page=300":
				w.Write([]byte(resp)) // nolint: errcheck
				return
			default:
				t.Errorf("got unexpected request at %q", r.RequestURI)
				http.Error(w, "not found", http.StatusNotFound)
				return
			}
		}))

	testServerURL, err := url.Parse(testServer.URL)
	Ok(t, err)
	client, err := vcs.NewGithubClient(testServerURL.Host, "user", "pass")
	Ok(t, err)
	defer disableSSLVerification()()

	files, err := client.GetModifiedFiles(models.Repo{
		FullName:          "owner/repo",
		Owner:             "owner",
		Name:              "repo",
		CloneURL:          "",
		SanitizedCloneURL: "",
		VCSHost: models.VCSHost{
			Type:     models.Github,
			Hostname: "github.com",
		},
	}, models.PullRequest{
		Num: 1,
	})
	Ok(t, err)
	Equals(t, []string{"new/filename.txt", "previous/filename.txt"}, files)
}

func TestGithubClient_UpdateStatus(t *testing.T) {
	cases := []struct {
		status   models.CommitStatus
		expState string
	}{
		{
			models.PendingCommitStatus,
			"pending",
		},
		{
			models.SuccessCommitStatus,
			"success",
		},
		{
			models.FailedCommitStatus,
			"failure",
		},
	}

	for _, c := range cases {
		t.Run(c.status.String(), func(t *testing.T) {
			testServer := httptest.NewTLSServer(
				http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					switch r.RequestURI {
					case "/api/v3/repos/owner/repo/statuses/":
						body, err := ioutil.ReadAll(r.Body)
						Ok(t, err)
						exp := fmt.Sprintf(`{"state":"%s","description":"description","context":"Atlantis"}%s`, c.expState, "\n")
						Equals(t, exp, string(body))
						defer r.Body.Close() // nolint: errcheck
						w.WriteHeader(http.StatusOK)
					default:
						t.Errorf("got unexpected request at %q", r.RequestURI)
						http.Error(w, "not found", http.StatusNotFound)
						return
					}
				}))

			testServerURL, err := url.Parse(testServer.URL)
			Ok(t, err)
			client, err := vcs.NewGithubClient(testServerURL.Host, "user", "pass")
			Ok(t, err)
			defer disableSSLVerification()()

			err = client.UpdateStatus(models.Repo{
				FullName:          "owner/repo",
				Owner:             "owner",
				Name:              "repo",
				CloneURL:          "",
				SanitizedCloneURL: "",
				VCSHost: models.VCSHost{
					Type:     models.Github,
					Hostname: "github.com",
				},
			}, models.PullRequest{
				Num: 1,
			}, c.status, "description")
			Ok(t, err)
		})
	}
}

func TestGithubClient_PullIsMergeable(t *testing.T) {
	cases := []struct {
		state        string
		expMergeable bool
	}{
		{
			"dirty",
			false,
		},
		{
			"unknown",
			false,
		},
		{
			"blocked",
			false,
		},
		{
			"behind",
			false,
		},
		{
			"random",
			false,
		},
		{
			"unstable",
			true,
		},
		{
			"has_hooks",
			true,
		},
		{
			"clean",
			true,
		},
		{
			"",
			false,
		},
	}

	// Use a real GitHub json response and edit the mergeable_state field.
	jsBytes, err := ioutil.ReadFile("fixtures/github-pull-request.json")
	Ok(t, err)
	json := string(jsBytes)

	for _, c := range cases {
		t.Run(c.state, func(t *testing.T) {
			response := strings.Replace(json,
				`"mergeable_state": "clean"`,
				fmt.Sprintf(`"mergeable_state": "%s"`, c.state),
				1,
			)

			testServer := httptest.NewTLSServer(
				http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					switch r.RequestURI {
					case "/api/v3/repos/owner/repo/pulls/1":
						w.Write([]byte(response)) // nolint: errcheck
						return
					default:
						t.Errorf("got unexpected request at %q", r.RequestURI)
						http.Error(w, "not found", http.StatusNotFound)
						return
					}
				}))
			testServerURL, err := url.Parse(testServer.URL)
			Ok(t, err)
			client, err := vcs.NewGithubClient(testServerURL.Host, "user", "pass")
			Ok(t, err)
			defer disableSSLVerification()()

			actMergeable, err := client.PullIsMergeable(models.Repo{
				FullName:          "owner/repo",
				Owner:             "owner",
				Name:              "repo",
				CloneURL:          "",
				SanitizedCloneURL: "",
				VCSHost: models.VCSHost{
					Type:     models.Github,
					Hostname: "github.com",
				},
			}, models.PullRequest{
				Num: 1,
			})
			Ok(t, err)
			Equals(t, c.expMergeable, actMergeable)
		})
	}
}

func TestGithubClient_MergePull(t *testing.T) {
	cases := []struct {
		code    int
		message string
		merged  string
		expErr  string
	}{
		{
			code:    200,
			message: "Pull Request successfully merged",
			merged:  "true",
			expErr:  "",
		},
		{
			code:    405,
			message: "Pull Request is not mergeable",
			expErr:  "405 Pull Request is not mergeable []",
		},
		{
			code:    409,
			message: "Head branch was modified. Review and try the merge again.",
			expErr:  "409 Head branch was modified. Review and try the merge again. []",
		},
	}

	for _, c := range cases {
		t.Run(c.message, func(t *testing.T) {
			testServer := httptest.NewTLSServer(
				http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					switch r.RequestURI {
					case "/api/v3/repos/owner/repo/pulls/1/merge":
						body, err := ioutil.ReadAll(r.Body)
						Ok(t, err)
						exp := "{\"commit_message\":\"[Atlantis] Automatically merging after successful apply\"}\n"
						Equals(t, exp, string(body))
						var resp string
						if c.code == 200 {
							resp = fmt.Sprintf(`{"message":"%s","merged":%s}%s`, c.message, c.merged, "\n")
						} else {
							resp = fmt.Sprintf(`{"message":"%s"}%s`, c.message, "\n")
						}
						defer r.Body.Close() // nolint: errcheck
						w.WriteHeader(c.code)
						w.Write([]byte(resp)) // nolint: errcheck
					default:
						t.Errorf("got unexpected request at %q", r.RequestURI)
						http.Error(w, "not found", http.StatusNotFound)
						return
					}
				}))

			testServerURL, err := url.Parse(testServer.URL)
			Ok(t, err)
			client, err := vcs.NewGithubClient(testServerURL.Host, "user", "pass")
			Ok(t, err)
			defer disableSSLVerification()()

			err = client.MergePull(
				models.PullRequest{
					BaseRepo: models.Repo{
						FullName:          "owner/repo",
						Owner:             "owner",
						Name:              "repo",
						CloneURL:          "",
						SanitizedCloneURL: "",
						VCSHost: models.VCSHost{
							Type:     models.Github,
							Hostname: "github.com",
						},
					},
					Num: 1,
				})

			if c.expErr == "" {
				Ok(t, err)
			} else {
				ErrContains(t, c.expErr, err)
			}
		})
	}
}

// disableSSLVerification disables ssl verification for the global http client
// and returns a function to be called in a defer that will re-enable it.
func disableSSLVerification() func() {
	orig := http.DefaultTransport.(*http.Transport).TLSClientConfig
	// nolint: gosec
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	return func() {
		http.DefaultTransport.(*http.Transport).TLSClientConfig = orig
	}
}
