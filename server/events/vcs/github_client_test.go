package vcs_test

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/events/vcs"
	"github.com/runatlantis/atlantis/server/logging"
	. "github.com/runatlantis/atlantis/testing"

	"github.com/shurcooL/githubv4"
)

// GetModifiedFiles should make multiple requests if more than one page
// and concat results.
func TestGithubClient_GetModifiedFiles(t *testing.T) {
	logger := logging.NewNoopLogger(t)
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
	client, err := vcs.NewGithubClient(testServerURL.Host, &vcs.GithubUserCredentials{"user", "pass", ""}, vcs.GithubConfig{}, 0, logger)
	Ok(t, err)
	defer disableSSLVerification()()

	files, err := client.GetModifiedFiles(
		logger,
		models.Repo{
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
	logger := logging.NewNoopLogger(t)
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
	client, err := vcs.NewGithubClient(testServerURL.Host, &vcs.GithubUserCredentials{"user", "pass", ""}, vcs.GithubConfig{}, 0, logging.NewNoopLogger(t))
	Ok(t, err)
	defer disableSSLVerification()()

	files, err := client.GetModifiedFiles(
		logger,
		models.Repo{
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

func TestGithubClient_PaginatesComments(t *testing.T) {
	logger := logging.NewNoopLogger(t)
	calls := 0
	issueResps := []string{
		`[
	{"node_id": "1", "body": "asd\nplan\nasd", "user": {"login": "someone-else"}},
	{"node_id": "2", "body": "asd plan\nasd", "user": {"login": "user"}}
]`,
		`[
	{"node_id": "3", "body": "asd", "user": {"login": "someone-else"}},
	{"node_id": "4", "body": "asdasd", "user": {"login": "someone-else"}}
]`,
		`[
	{"node_id": "5", "body": "asd plan", "user": {"login": "someone-else"}},
	{"node_id": "6", "body": "asd\nplan", "user": {"login": "user"}}
]`,
		`[
	{"node_id": "7", "body": "asd", "user": {"login": "user"}},
	{"node_id": "8", "body": "asd plan \n asd", "user": {"login": "user"}}
]`,
	}
	minimizeResp := "{}"
	type graphQLCall struct {
		Variables struct {
			Input githubv4.MinimizeCommentInput `json:"input"`
		} `json:"variables"`
	}
	gotMinimizeCalls := make([]graphQLCall, 0, 2)
	testServer := httptest.NewTLSServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.Method + " " + r.RequestURI {
			case "POST /api/graphql":
				defer r.Body.Close() // nolint: errcheck
				body, err := io.ReadAll(r.Body)
				if err != nil {
					t.Errorf("read body error: %v", err)
					http.Error(w, "server error", http.StatusInternalServerError)
					return
				}
				call := graphQLCall{}
				err = json.Unmarshal(body, &call)
				if err != nil {
					t.Errorf("parse body error: %v", err)
					http.Error(w, "server error", http.StatusInternalServerError)
					return
				}
				gotMinimizeCalls = append(gotMinimizeCalls, call)
				w.Write([]byte(minimizeResp)) // nolint: errcheck
				return
			default:
				if r.Method != "GET" || !strings.HasPrefix(r.RequestURI, "/api/v3/repos/owner/repo/issues/123/comments") {
					t.Errorf("got unexpected request at %q", r.RequestURI)
					http.Error(w, "not found", http.StatusNotFound)
					return
				}
				if (calls + 1) < len(issueResps) {
					w.Header().Add(
						"Link",
						fmt.Sprintf(
							`<http://%s/api/v3/repos/owner/repo/issues/123/comments?page=%d&per_page=100>; rel="next"`,
							r.Host,
							calls+1,
						),
					)
				}
				w.Write([]byte(issueResps[calls])) // nolint: errcheck
				calls++
			}
		}),
	)

	testServerURL, err := url.Parse(testServer.URL)
	Ok(t, err)

	client, err := vcs.NewGithubClient(testServerURL.Host, &vcs.GithubUserCredentials{"user", "pass", ""}, vcs.GithubConfig{}, 0, logging.NewNoopLogger(t))
	Ok(t, err)
	defer disableSSLVerification()()

	err = client.HidePrevCommandComments(
		logger,
		models.Repo{
			FullName:          "owner/repo",
			Owner:             "owner",
			Name:              "repo",
			CloneURL:          "",
			SanitizedCloneURL: "",
			VCSHost: models.VCSHost{
				Hostname: "github.com",
				Type:     models.Github,
			},
		},
		123,
		command.Plan.TitleString(),
		"",
	)
	Ok(t, err)
	Equals(t, 2, len(gotMinimizeCalls))
	Equals(t, "2", gotMinimizeCalls[0].Variables.Input.SubjectID)
	Equals(t, "8", gotMinimizeCalls[1].Variables.Input.SubjectID)
	Equals(t, githubv4.ReportedContentClassifiersOutdated, gotMinimizeCalls[0].Variables.Input.Classifier)
	Equals(t, githubv4.ReportedContentClassifiersOutdated, gotMinimizeCalls[1].Variables.Input.Classifier)
}

func TestGithubClient_HideOldComments(t *testing.T) {
	logger := logging.NewNoopLogger(t)
	atlantisUser := "AtlantisUser"
	pullRequestNum := 123
	issueResp := strings.ReplaceAll(`[
	{"node_id": "1", "body": "asd\nplan\nasd", "user": {"login": "someone-else"}},
	{"node_id": "2", "body": "asd plan\nasd", "user": {"login": "someone-else"}},
	{"node_id": "3", "body": "asdasdasd\nasdasdasd", "user": {"login": "someone-else"}},
	{"node_id": "4", "body": "asdasdasd\nasdasdasd", "user": {"login": "AtlantisUser"}},
	{"node_id": "5", "body": "asd\nplan\nasd", "user": {"login": "AtlantisUser"}},
	{"node_id": "6", "body": "Ran Plan for 2 projects:", "user": {"login": "AtlantisUser"}},
	{"node_id": "7", "body": "Ran Apply for 2 projects:", "user": {"login": "AtlantisUser"}},
	{"node_id": "8", "body": "Ran Plan for dir: 'stack1' workspace: 'default'", "user": {"login": "AtlantisUser"}},
	{"node_id": "9", "body": "Ran Plan for dir: 'stack2' workspace: 'default'", "user": {"login": "AtlantisUser"}},
	{"node_id": "10", "body": "Continued Plan from previous comment\nasd", "user": {"login": "AtlantisUser"}}
]`, "'", "`")
	minimizeResp := "{}"
	type graphQLCall struct {
		Variables struct {
			Input githubv4.MinimizeCommentInput `json:"input"`
		} `json:"variables"`
	}

	cases := []struct {
		dir                 string
		processedComments   int
		processedCommentIds []string
	}{
		{
			// With no dir specified, comments 6, 8, 9 and 10 should be minimized.
			"",
			4,
			[]string{"6", "8", "9", "10"},
		},
		{
			// With a dir of "stack1", comment 8 should be minimized.
			"stack1",
			1,
			[]string{"8"},
		},
		{
			// With a dir of "stack2", comment 9 should be minimized.
			"stack2",
			1,
			[]string{"9"},
		},
	}

	for _, c := range cases {
		t.Run(c.dir, func(t *testing.T) {
			gotMinimizeCalls := make([]graphQLCall, 0, 1)
			testServer := httptest.NewTLSServer(
				http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					switch r.Method + " " + r.RequestURI {
					// This gets the pull request's comments.
					case fmt.Sprintf("GET /api/v3/repos/owner/repo/issues/%v/comments?direction=asc&sort=created", pullRequestNum):
						w.Write([]byte(issueResp)) // nolint: errcheck
						return
					case "POST /api/graphql":
						defer r.Body.Close() // nolint: errcheck
						body, err := io.ReadAll(r.Body)
						if err != nil {
							t.Errorf("read body error: %v", err)
							http.Error(w, "server error", http.StatusInternalServerError)
							return
						}
						call := graphQLCall{}
						err = json.Unmarshal(body, &call)
						if err != nil {
							t.Errorf("parse body error: %v", err)
							http.Error(w, "server error", http.StatusInternalServerError)
							return
						}
						gotMinimizeCalls = append(gotMinimizeCalls, call)
						w.Write([]byte(minimizeResp)) // nolint: errcheck
						return
					default:
						t.Errorf("got unexpected request at %q", r.RequestURI)
						http.Error(w, "not found", http.StatusNotFound)
						return
					}
				}),
			)

			testServerURL, err := url.Parse(testServer.URL)
			Ok(t, err)

			client, err := vcs.NewGithubClient(testServerURL.Host, &vcs.GithubUserCredentials{atlantisUser, "pass", ""}, vcs.GithubConfig{}, 0,
				logging.NewNoopLogger(t))
			Ok(t, err)
			defer disableSSLVerification()()

			err = client.HidePrevCommandComments(
				logger,
				models.Repo{
					FullName:          "owner/repo",
					Owner:             "owner",
					Name:              "repo",
					CloneURL:          "",
					SanitizedCloneURL: "",
					VCSHost: models.VCSHost{
						Hostname: "github.com",
						Type:     models.Github,
					},
				},
				pullRequestNum,
				command.Plan.TitleString(),
				c.dir,
			)
			Ok(t, err)
			Equals(t, c.processedComments, len(gotMinimizeCalls))
			for i := 0; i < c.processedComments; i++ {
				Equals(t, c.processedCommentIds[i], gotMinimizeCalls[i].Variables.Input.SubjectID)
				Equals(t, githubv4.ReportedContentClassifiersOutdated, gotMinimizeCalls[i].Variables.Input.Classifier)
			}
		})
	}
}

func TestGithubClient_UpdateStatus(t *testing.T) {
	logger := logging.NewNoopLogger(t)
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
						body, err := io.ReadAll(r.Body)
						Ok(t, err)
						exp := fmt.Sprintf(`{"state":"%s","target_url":"https://google.com","description":"description","context":"src"}%s`, c.expState, "\n")
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
			client, err := vcs.NewGithubClient(testServerURL.Host, &vcs.GithubUserCredentials{"user", "pass", ""}, vcs.GithubConfig{}, 0, logging.NewNoopLogger(t))
			Ok(t, err)
			defer disableSSLVerification()()

			err = client.UpdateStatus(
				logger,
				models.Repo{
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
				}, c.status, "src", "description", "https://google.com")
			Ok(t, err)
		})
	}
}

func TestGithubClient_PullIsApproved(t *testing.T) {
	logger := logging.NewNoopLogger(t)
	respTemplate := `[
		{
			"id": %d,
			"node_id": "MDE3OlB1bGxSZXF1ZXN0UmV2aWV3ODA=",
			"user": {
			  "login": "octocat",
			  "id": 1,
			  "node_id": "MDQ6VXNlcjE=",
			  "avatar_url": "https://github.com/images/error/octocat_happy.gif",
			  "gravatar_id": "",
			  "url": "https://api.github.com/users/octocat",
			  "html_url": "https://github.com/octocat",
			  "followers_url": "https://api.github.com/users/octocat/followers",
			  "following_url": "https://api.github.com/users/octocat/following{/other_user}",
			  "gists_url": "https://api.github.com/users/octocat/gists{/gist_id}",
			  "starred_url": "https://api.github.com/users/octocat/starred{/owner}{/repo}",
			  "subscriptions_url": "https://api.github.com/users/octocat/subscriptions",
			  "organizations_url": "https://api.github.com/users/octocat/orgs",
			  "repos_url": "https://api.github.com/users/octocat/repos",
			  "events_url": "https://api.github.com/users/octocat/events{/privacy}",
			  "received_events_url": "https://api.github.com/users/octocat/received_events",
			  "type": "User",
			  "site_admin": false
			},
			"body": "Here is the body for the review.",
			"commit_id": "ecdd80bb57125d7ba9641ffaa4d7d2c19d3f3091",
			"state": "CHANGES_REQUESTED",
			"html_url": "https://github.com/octocat/Hello-World/pull/12#pullrequestreview-%d",
			"pull_request_url": "https://api.github.com/repos/octocat/Hello-World/pulls/12",
			"_links": {
			  "html": {
				"href": "https://github.com/octocat/Hello-World/pull/12#pullrequestreview-%d"
			  },
			  "pull_request": {
				"href": "https://api.github.com/repos/octocat/Hello-World/pulls/12"
			  }
			}
		  }
]`
	firstResp := fmt.Sprintf(respTemplate, 80, 80, 80)
	secondResp := fmt.Sprintf(respTemplate, 81, 81, 81)
	testServer := httptest.NewTLSServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.RequestURI {
			// The first request should hit this URL.
			case "/api/v3/repos/owner/repo/pulls/1/reviews?per_page=300":
				// We write a header that means there's an additional page.
				w.Header().Add("Link", `<https://api.github.com/resource?page=2>; rel="next",
      <https://api.github.com/resource?page=2>; rel="last"`)
				w.Write([]byte(firstResp)) // nolint: errcheck
				return
				// The second should hit this URL.
			case "/api/v3/repos/owner/repo/pulls/1/reviews?page=2&per_page=300":
				w.Write([]byte(secondResp)) // nolint: errcheck
			default:
				t.Errorf("got unexpected request at %q", r.RequestURI)
				http.Error(w, "not found", http.StatusNotFound)
				return
			}
		}))

	testServerURL, err := url.Parse(testServer.URL)
	Ok(t, err)
	client, err := vcs.NewGithubClient(testServerURL.Host, &vcs.GithubUserCredentials{"user", "pass", ""}, vcs.GithubConfig{}, 0, logging.NewNoopLogger(t))
	Ok(t, err)
	defer disableSSLVerification()()

	approvalStatus, err := client.PullIsApproved(
		logger,
		models.Repo{
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
	Equals(t, false, approvalStatus.IsApproved)
}

func TestGithubClient_PullIsMergeable(t *testing.T) {
	logger := logging.NewNoopLogger(t)
	vcsStatusName := "atlantis-test"
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
	jsBytes, err := os.ReadFile("testdata/github-pull-request.json")
	Ok(t, err)
	prJSON := string(jsBytes)

	for _, c := range cases {
		t.Run(c.state, func(t *testing.T) {
			response := strings.Replace(prJSON,
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
			client, err := vcs.NewGithubClient(testServerURL.Host, &vcs.GithubUserCredentials{"user", "pass", ""}, vcs.GithubConfig{}, 0, logging.NewNoopLogger(t))
			Ok(t, err)
			defer disableSSLVerification()()

			actMergeable, err := client.PullIsMergeable(
				logger,
				models.Repo{
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
				}, vcsStatusName, []string{})
			Ok(t, err)
			Equals(t, c.expMergeable, actMergeable)
		})
	}
}

func TestGithubClient_PullIsMergeableWithAllowMergeableBypassApply(t *testing.T) {
	logger := logging.NewNoopLogger(t)
	vcsStatusName := "atlantis"
	ignoreVCSStatusNames := []string{"other-atlantis"}
	cases := []struct {
		state                     string
		statusCheckRollupFilePath string
		reviewDecision            string
		expMergeable              bool
	}{
		{
			"dirty",
			"ruleset-atlantis-apply-pending.json",
			`"REVIEW_REQUIRED"`,
			false,
		},
		{
			"unknown",
			"ruleset-atlantis-apply-pending.json",
			`"REVIEW_REQUIRED"`,
			false,
		},
		{
			"blocked",
			"ruleset-atlantis-apply-pending.json",
			`"REVIEW_REQUIRED"`,
			false,
		},
		{
			"blocked",
			"ruleset-atlantis-apply-pending.json",
			`"APPROVED"`,
			true,
		},
		{
			"blocked",
			"ruleset-atlantis-apply-pending.json",
			"null",
			true,
		},
		{
			"behind",
			"ruleset-atlantis-apply-pending.json",
			`"REVIEW_REQUIRED"`,
			false,
		},
		{
			"random",
			"ruleset-atlantis-apply-pending.json",
			`"REVIEW_REQUIRED"`,
			false,
		},
		{
			"unstable",
			"ruleset-atlantis-apply-pending.json",
			`"REVIEW_REQUIRED"`,
			true,
		},
		{
			"has_hooks",
			"ruleset-atlantis-apply-pending.json",
			`"APPROVED"`,
			true,
		},
		{
			"clean",
			"ruleset-atlantis-apply-pending.json",
			`"APPROVED"`,
			true,
		},
		{
			"",
			"ruleset-atlantis-apply-pending.json",
			`"APPROVED"`,
			false,
		},
		{
			"blocked",
			"ruleset-atlantis-apply-expected.json",
			`"APPROVED"`,
			true,
		},
		{
			"blocked",
			"ruleset-optional-check-failed.json",
			`"APPROVED"`,
			true,
		},
		{
			"blocked",
			"ruleset-optional-status-failed.json",
			`"APPROVED"`,
			true,
		},
		{
			"blocked",
			"ruleset-check-pending.json",
			`"APPROVED"`,
			false,
		},
		{
			"blocked",
			"ruleset-check-pending-other-atlantis.json",
			`"APPROVED"`,
			true,
		},
		{
			"blocked",
			"ruleset-check-skipped.json",
			`"APPROVED"`,
			true,
		},
		{
			"blocked",
			"ruleset-check-neutral.json",
			`"APPROVED"`,
			true,
		},
		{
			"blocked",
			"ruleset-evaluate-workflow-failed.json",
			`"APPROVED"`,
			true,
		},
		{
			"blocked",
			"branch-protection-expected.json",
			`"APPROVED"`,
			false,
		},
		{
			"blocked",
			"branch-protection-failed.json",
			`"APPROVED"`,
			false,
		},
		{
			"blocked",
			"branch-protection-passed.json",
			`"APPROVED"`,
			true,
		},
		{
			"blocked",
			"ruleset-check-expected.json",
			`"APPROVED"`,
			false,
		},
		{
			"blocked",
			"ruleset-check-failed.json",
			`"APPROVED"`,
			false,
		},
		{
			"blocked",
			"ruleset-check-failed-other-atlantis.json",
			`"APPROVED"`,
			true,
		},
		{
			"blocked",
			"ruleset-check-passed.json",
			`"APPROVED"`,
			true,
		},
		{
			"blocked",
			"ruleset-workflow-expected.json",
			`"APPROVED"`,
			false,
		},
		{
			"blocked",
			"ruleset-workflow-failed.json",
			`"APPROVED"`,
			false,
		},
		{
			"blocked",
			"ruleset-workflow-passed.json",
			`"APPROVED"`,
			true,
		},
		{
			"blocked",
			"ruleset-workflow-passed-multiple-runs.json",
			`"APPROVED"`,
			true,
		},
		{
			"blocked",
			"ruleset-workflow-passed-sha-match.json",
			`"APPROVED"`,
			true,
		},
		{
			"blocked",
			"ruleset-workflow-passed-sha-mismatch.json",
			`"APPROVED"`,
			false,
		},
	}

	// Use a real GitHub json response and edit the mergeable_state field.
	jsBytes, err := os.ReadFile("testdata/github-pull-request.json")
	Ok(t, err)
	prJSON := string(jsBytes)

	jsBytes, err = os.ReadFile("testdata/github-pull-request-mergeability/repository-id.json")
	Ok(t, err)
	repoIdJSON := string(jsBytes)

	for _, c := range cases {
		t.Run(c.state, func(t *testing.T) {
			response := strings.Replace(prJSON,
				`"mergeable_state": "clean"`,
				fmt.Sprintf(`"mergeable_state": "%s"`, c.state),
				1,
			)

			// PR review decision and checks statuses Response
			jsBytes, err = os.ReadFile("testdata/github-pull-request-mergeability/" + c.statusCheckRollupFilePath)
			Ok(t, err)
			prMergeableStatusJSON := string(jsBytes)

			// PR review decision and checks statuses Response
			prMergeableStatus := strings.Replace(prMergeableStatusJSON,
				`"reviewDecision": null,`,
				fmt.Sprintf(`"reviewDecision": %s,`, c.reviewDecision),
				1,
			)

			testServer := httptest.NewTLSServer(
				http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					switch r.RequestURI {
					case "/api/v3/repos/octocat/repo/pulls/1":
						w.Write([]byte(response)) // nolint: errcheck
						return
					case "/api/graphql":
						body, err := io.ReadAll(r.Body)
						if err != nil {
							t.Errorf("read body error: %v", err)
							http.Error(w, "", http.StatusInternalServerError)
							return
						}
						if strings.Contains(string(body), "pullRequest(") {
							w.Write([]byte(prMergeableStatus)) // nolint: errcheck
							return
						} else if strings.Contains(string(body), "databaseId") {
							w.Write([]byte(repoIdJSON)) // nolint: errcheck
							return
						}
						t.Errorf("got unexpected request at %q", r.RequestURI)
						http.Error(w, "not found", http.StatusNotFound)
						return
					default:
						t.Errorf("got unexpected request at %q", r.RequestURI)
						http.Error(w, "not found", http.StatusNotFound)
						return
					}
				}))
			testServerURL, err := url.Parse(testServer.URL)
			Ok(t, err)
			client, err := vcs.NewGithubClient(testServerURL.Host, &vcs.GithubUserCredentials{"user", "pass", ""}, vcs.GithubConfig{AllowMergeableBypassApply: true}, 0, logging.NewNoopLogger(t))
			Ok(t, err)
			defer disableSSLVerification()()

			actMergeable, err := client.PullIsMergeable(
				logger,
				models.Repo{
					FullName:          "octocat/repo",
					Owner:             "octocat",
					Name:              "repo",
					CloneURL:          "",
					SanitizedCloneURL: "",
					VCSHost: models.VCSHost{
						Type:     models.Github,
						Hostname: "github.com",
					},
				}, models.PullRequest{
					Num: 1,
				}, vcsStatusName, ignoreVCSStatusNames)
			Ok(t, err)
			Equals(t, c.expMergeable, actMergeable)
		})
	}
}

func TestGithubClient_MergePullHandlesError(t *testing.T) {
	logger := logging.NewNoopLogger(t)
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

	jsBytes, err := os.ReadFile("testdata/github-repo.json")
	Ok(t, err)

	for _, c := range cases {
		t.Run(c.message, func(t *testing.T) {
			testServer := httptest.NewTLSServer(
				http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					switch r.RequestURI {
					case "/api/v3/repos/owner/repo":
						w.Write(jsBytes) // nolint: errcheck
						return
					case "/api/v3/repos/owner/repo/pulls/1/merge":
						body, err := io.ReadAll(r.Body)
						Ok(t, err)
						exp := "{\"merge_method\":\"merge\"}\n"
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
			client, err := vcs.NewGithubClient(testServerURL.Host, &vcs.GithubUserCredentials{"user", "pass", ""}, vcs.GithubConfig{}, 0, logging.NewNoopLogger(t))
			Ok(t, err)
			defer disableSSLVerification()()

			err = client.MergePull(
				logger,
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
				}, models.PullRequestOptions{
					DeleteSourceBranchOnMerge: false,
				})

			if c.expErr == "" {
				Ok(t, err)
			} else {
				ErrContains(t, c.expErr, err)
			}
		})
	}
}

// Test that if the pull request only allows a certain merge method that we
// use that method
func TestGithubClient_MergePullCorrectMethod(t *testing.T) {
	logger := logging.NewNoopLogger(t)
	cases := map[string]struct {
		allowMerge        bool
		allowRebase       bool
		allowSquash       bool
		mergeMethodOption string
		expMethod         string
		expErr            string
	}{
		"all true": {
			allowMerge:  true,
			allowRebase: true,
			allowSquash: true,
			expMethod:   "merge",
		},
		"all false (edge case)": {
			allowMerge:  false,
			allowRebase: false,
			allowSquash: false,
			expMethod:   "merge",
		},
		"merge: false rebase: true squash: true": {
			allowMerge:  false,
			allowRebase: true,
			allowSquash: true,
			expMethod:   "rebase",
		},
		"merge: false rebase: false squash: true": {
			allowMerge:  false,
			allowRebase: false,
			allowSquash: true,
			expMethod:   "squash",
		},
		"merge: false rebase: true squash: false": {
			allowMerge:  false,
			allowRebase: true,
			allowSquash: false,
			expMethod:   "rebase",
		},
		"all true: merge with merge: overrided by command": {
			allowMerge:        true,
			allowRebase:       true,
			allowSquash:       true,
			mergeMethodOption: "merge",
			expMethod:         "merge",
		},
		"all true: merge with rebase: overrided by command": {
			allowMerge:        true,
			allowRebase:       true,
			allowSquash:       true,
			mergeMethodOption: "rebase",
			expMethod:         "rebase",
		},
		"all true: merge with squash: overrided by command": {
			allowMerge:        true,
			allowRebase:       true,
			allowSquash:       true,
			mergeMethodOption: "squash",
			expMethod:         "squash",
		},
		"merge with merge: overridden by command: merge not allowed": {
			allowMerge:        false,
			allowRebase:       true,
			allowSquash:       true,
			mergeMethodOption: "merge",
			expMethod:         "",
			expErr:            "Merge method 'merge' is not allowed by the repository Pull Request settings",
		},
		"merge with rebase: overridden by command: rebase not allowed": {
			allowMerge:        true,
			allowRebase:       false,
			allowSquash:       true,
			mergeMethodOption: "rebase",
			expMethod:         "",
			expErr:            "Merge method 'rebase' is not allowed by the repository Pull Request settings",
		},
		"merge with squash: overridden by command: squash not allowed": {
			allowMerge:        true,
			allowRebase:       true,
			allowSquash:       false,
			mergeMethodOption: "squash",
			expMethod:         "",
			expErr:            "Merge method 'squash' is not allowed by the repository Pull Request settings",
		},
		"merge with unknown: overridden by command: unknown doesn't exist": {
			allowMerge:        true,
			allowRebase:       true,
			allowSquash:       true,
			mergeMethodOption: "unknown",
			expMethod:         "",
			expErr:            "Merge method 'unknown' is unknown. Specify one of the valid values: 'merge, rebase, squash'",
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {

			// Modify response.
			jsBytes, err := os.ReadFile("testdata/github-repo.json")
			Ok(t, err)
			resp := string(jsBytes)
			resp = strings.Replace(resp,
				`"allow_squash_merge": true`,
				fmt.Sprintf(`"allow_squash_merge": %t`, c.allowSquash),
				-1)
			resp = strings.Replace(resp,
				`"allow_merge_commit": true`,
				fmt.Sprintf(`"allow_merge_commit": %t`, c.allowMerge),
				-1)
			resp = strings.Replace(resp,
				`"allow_rebase_merge": true`,
				fmt.Sprintf(`"allow_rebase_merge": %t`, c.allowRebase),
				-1)

			testServer := httptest.NewTLSServer(
				http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					switch r.RequestURI {
					case "/api/v3/repos/runatlantis/atlantis":
						w.Write([]byte(resp)) // nolint: errcheck
						return
					case "/api/v3/repos/runatlantis/atlantis/pulls/1/merge":
						body, err := io.ReadAll(r.Body)
						Ok(t, err)
						defer r.Body.Close() // nolint: errcheck
						type bodyJSON struct {
							MergeMethod string `json:"merge_method"`
						}
						expBody := bodyJSON{
							MergeMethod: c.expMethod,
						}
						expBytes, err := json.Marshal(expBody)
						Ok(t, err)
						Equals(t, string(expBytes)+"\n", string(body))

						resp := `{"sha":"6dcb09b5b57875f334f61aebed695e2e4193db5e","merged":true,"message":"Pull Request successfully merged"}`
						w.Write([]byte(resp)) // nolint: errcheck
					default:
						t.Errorf("got unexpected request at %q", r.RequestURI)
						http.Error(w, "not found", http.StatusNotFound)
						return
					}
				}))

			testServerURL, err := url.Parse(testServer.URL)
			Ok(t, err)
			client, err := vcs.NewGithubClient(testServerURL.Host, &vcs.GithubUserCredentials{"user", "pass", ""}, vcs.GithubConfig{}, 0, logging.NewNoopLogger(t))
			Ok(t, err)
			defer disableSSLVerification()()

			err = client.MergePull(
				logger,
				models.PullRequest{
					BaseRepo: models.Repo{
						FullName:          "runatlantis/atlantis",
						Owner:             "runatlantis",
						Name:              "atlantis",
						CloneURL:          "",
						SanitizedCloneURL: "",
						VCSHost: models.VCSHost{
							Type:     models.Github,
							Hostname: "github.com",
						},
					},
					Num: 1,
				}, models.PullRequestOptions{
					DeleteSourceBranchOnMerge: false,
					MergeMethod:               c.mergeMethodOption,
				})

			if c.expErr == "" {
				Ok(t, err)
			} else {
				ErrContains(t, c.expErr, err)
			}
		})
	}
}

func TestGithubClient_MarkdownPullLink(t *testing.T) {
	client, err := vcs.NewGithubClient("hostname", &vcs.GithubUserCredentials{"user", "pass", ""}, vcs.GithubConfig{}, 0, logging.NewNoopLogger(t))
	Ok(t, err)
	pull := models.PullRequest{Num: 1}
	s, _ := client.MarkdownPullLink(pull)
	exp := "#1"
	Equals(t, exp, s)
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

func TestGithubClient_SplitComments(t *testing.T) {
	logger := logging.NewNoopLogger(t)
	type githubComment struct {
		Body string `json:"body"`
	}
	githubComments := make([]githubComment, 0, 1)

	testServer := httptest.NewTLSServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			switch r.Method + " " + r.RequestURI {
			case "POST /api/v3/repos/runatlantis/atlantis/issues/1/comments":
				defer r.Body.Close() // nolint: errcheck
				body, err := io.ReadAll(r.Body)
				if err != nil {
					t.Errorf("read body error: %v", err)
					http.Error(w, "server error", http.StatusInternalServerError)
					return
				}
				requestBody := githubComment{}
				err = json.Unmarshal(body, &requestBody)
				if err != nil {
					t.Errorf("parse body error: %v", err)
					http.Error(w, "server error", http.StatusInternalServerError)
					return
				}
				githubComments = append(githubComments, requestBody)
				return
			default:
				t.Errorf("got unexpected request at %q", r.RequestURI)
				http.Error(w, "not found", http.StatusNotFound)
				return
			}
		}))

	testServerURL, err := url.Parse(testServer.URL)
	Ok(t, err)
	client, err := vcs.NewGithubClient(testServerURL.Host, &vcs.GithubUserCredentials{"user", "pass", ""}, vcs.GithubConfig{}, 0, logging.NewNoopLogger(t))
	Ok(t, err)
	defer disableSSLVerification()()
	pull := models.PullRequest{Num: 1}
	repo := models.Repo{
		FullName:          "runatlantis/atlantis",
		Owner:             "runatlantis",
		Name:              "atlantis",
		CloneURL:          "",
		SanitizedCloneURL: "",
		VCSHost: models.VCSHost{
			Type:     models.Github,
			Hostname: "github.com",
		},
	}
	// create an extra long string
	comment := strings.Repeat("a", 65537)
	err = client.CreateComment(logger, repo, pull.Num, comment, command.Plan.String())
	Ok(t, err)
	err = client.CreateComment(logger, repo, pull.Num, comment, "")
	Ok(t, err)

	body := strings.Split(githubComments[1].Body, "\n")
	firstSplit := strings.ToLower(body[0])
	body = strings.Split(githubComments[3].Body, "\n")
	secondSplit := strings.ToLower(body[0])

	Equals(t, 4, len(githubComments))
	Assert(t, strings.Contains(firstSplit, command.Plan.String()), fmt.Sprintf("comment should contain the command name but was %q", firstSplit))
	Assert(t, strings.Contains(secondSplit, "continued from previous comment"), fmt.Sprintf("comment should contain no reference to the command name but was %q", secondSplit))
}

// Test that we retry the get pull request call if it 404s.
func TestGithubClient_Retry404(t *testing.T) {
	logger := logging.NewNoopLogger(t)
	var numCalls = 0

	testServer := httptest.NewTLSServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			switch r.Method + " " + r.RequestURI {
			case "GET /api/v3/repos/runatlantis/atlantis/pulls/1":
				defer r.Body.Close() // nolint: errcheck
				numCalls++
				if numCalls < 3 {
					w.WriteHeader(404)
				} else {
					w.WriteHeader(200)
				}
				return
			default:
				t.Errorf("got unexpected request at %q", r.RequestURI)
				http.Error(w, "not found", http.StatusNotFound)
				return
			}
		}))

	testServerURL, err := url.Parse(testServer.URL)
	Ok(t, err)
	client, err := vcs.NewGithubClient(testServerURL.Host, &vcs.GithubUserCredentials{"user", "pass", ""}, vcs.GithubConfig{}, 0, logging.NewNoopLogger(t))
	Ok(t, err)
	defer disableSSLVerification()()
	repo := models.Repo{
		FullName:          "runatlantis/atlantis",
		Owner:             "runatlantis",
		Name:              "atlantis",
		CloneURL:          "",
		SanitizedCloneURL: "",
		VCSHost: models.VCSHost{
			Type:     models.Github,
			Hostname: "github.com",
		},
	}
	_, err = client.GetPullRequest(logger, repo, 1)
	Ok(t, err)
	Equals(t, 3, numCalls)
}

// Test that we retry the get pull request files call if it 404s.
func TestGithubClient_Retry404Files(t *testing.T) {
	logger := logging.NewNoopLogger(t)
	var numCalls = 0

	testServer := httptest.NewTLSServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			switch r.Method + " " + r.RequestURI {
			case "GET /api/v3/repos/runatlantis/atlantis/pulls/1/files?per_page=300":
				defer r.Body.Close() // nolint: errcheck
				numCalls++
				if numCalls < 3 {
					w.WriteHeader(404)
				} else {
					w.WriteHeader(200)
				}
				return
			default:
				t.Errorf("got unexpected request at %q", r.RequestURI)
				http.Error(w, "not found", http.StatusNotFound)
				return
			}
		}))

	testServerURL, err := url.Parse(testServer.URL)
	Ok(t, err)
	client, err := vcs.NewGithubClient(testServerURL.Host, &vcs.GithubUserCredentials{"user", "pass", ""}, vcs.GithubConfig{}, 0, logging.NewNoopLogger(t))
	Ok(t, err)
	defer disableSSLVerification()()
	repo := models.Repo{
		FullName:          "runatlantis/atlantis",
		Owner:             "runatlantis",
		Name:              "atlantis",
		CloneURL:          "",
		SanitizedCloneURL: "",
		VCSHost: models.VCSHost{
			Type:     models.Github,
			Hostname: "github.com",
		},
	}
	pr := models.PullRequest{Num: 1}
	_, err = client.GetModifiedFiles(logger, repo, pr)
	Ok(t, err)
	Equals(t, 3, numCalls)
}

// GetTeamNamesForUser returns a list of team names for a user.
func TestGithubClient_GetTeamNamesForUser(t *testing.T) {
	logger := logging.NewNoopLogger(t)
	// Mocked GraphQL response for two teams
	resp := `{
		"data":{
		  "organization": {
			"teams":{
				"edges":[
					{"node":{"name": "Frontend Developers", "slug":"frontend-developers"}},
					{"node":{"name": "Employees", "slug":"employees"}}
				],
				"pageInfo":{
					"endCursor":"Y3Vyc29yOnYyOpHOAFMoLQ==",
					"hasNextPage":false
				}
			}
		}
	  }
	}`
	testServer := httptest.NewTLSServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.RequestURI {
			case "/api/graphql":
				w.Write([]byte(resp)) // nolint: errcheck
			default:
				t.Errorf("got unexpected request at %q", r.RequestURI)
				http.Error(w, "not found", http.StatusNotFound)
				return
			}
		}))
	testServerURL, err := url.Parse(testServer.URL)
	Ok(t, err)
	client, err := vcs.NewGithubClient(testServerURL.Host, &vcs.GithubUserCredentials{"user", "pass", ""}, vcs.GithubConfig{}, 0, logger)
	Ok(t, err)
	defer disableSSLVerification()()

	teams, err := client.GetTeamNamesForUser(models.Repo{
		Owner: "testrepo",
	}, models.User{
		Username: "testuser",
	})
	Ok(t, err)
	Equals(t, []string{"Frontend Developers", "frontend-developers", "Employees", "employees"}, teams)
}

func TestGithubClient_DiscardReviews(t *testing.T) {
	type ResponseDef struct {
		httpCode int
		body     string
	}
	type fields struct {
		responses []ResponseDef
	}
	type args struct {
		repo models.Repo
		pull models.PullRequest
	}

	queryResponseSingleReview := `{
  "data": {
    "repository": {
      "pullRequest": {
        "reviewDecision": "APPROVED",
        "reviews": {
          "nodes": [
            {
              "id": "PRR_kwDOFxULt85HBb7A",
              "submittedAt": "2022-11-23T12:28:30Z",
              "author": {
                "login": "atlantis-test"
              }
            }
          ]
        }
      }
    }
  }
}`
	queryResponseMultipleReviews := `{
  "data": {
    "repository": {
      "pullRequest": {
        "reviewDecision": "APPROVED",
        "reviews": {
          "nodes": [
            {
              "id": "PRR_kwDOFxULt85HBb7A",
              "submittedAt": "2022-11-23T12:28:30Z",
              "author": {
                "login": "atlantis-test"
              }
            },
            {
              "id": "PRR_kwDOFxULt85HBb7B",
              "submittedAt": "2022-11-23T14:28:30Z",
              "author": {
                "login": "atlantis-test2"
              }
            }
          ]
        }
      }
    }
  }
}`
	mutationResponseSingleReviewDismissal := `{
  "data": {
    "dismissPullRequestReview": {
      "pullRequestReview": {
        "id": "PRR_kwDOFxULt85HBb7A"
      }
    }
  }
}`
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "return no error if dismissing a single approval",
			fields: fields{
				responses: []ResponseDef{
					{
						httpCode: 200,
						body:     queryResponseSingleReview,
					},
					{
						httpCode: 200,
						body:     mutationResponseSingleReviewDismissal,
					},
				},
			},
			args:    args{},
			wantErr: false,
		},
		{
			name: "return no error if dismissing multiple reviews",
			fields: fields{
				responses: []ResponseDef{
					{
						httpCode: 200,
						body:     queryResponseMultipleReviews,
					},
					{
						httpCode: 200,
						body:     mutationResponseSingleReviewDismissal,
					},
					{
						httpCode: 200,
						body:     mutationResponseSingleReviewDismissal,
					},
				},
			},
			args:    args{},
			wantErr: false,
		},
		{
			name: "return error if query fails",
			fields: fields{
				responses: []ResponseDef{
					{
						httpCode: 500,
						body:     "",
					},
				},
			},
			args:    args{},
			wantErr: true,
		},
		{
			name: "return error if mutating fails",
			fields: fields{
				responses: []ResponseDef{
					{
						httpCode: 200,
						body:     queryResponseSingleReview,
					},
					{
						httpCode: 500,
						body:     "",
					},
				},
			},
			args:    args{},
			wantErr: true,
		},
		{
			name: "return error if dismissing fails after already dismissing one",
			fields: fields{
				responses: []ResponseDef{
					{
						httpCode: 200,
						body:     queryResponseMultipleReviews,
					},
					{
						httpCode: 200,
						body:     mutationResponseSingleReviewDismissal,
					},
					{
						httpCode: 500,
						body:     "",
					},
				},
			},
			args:    args{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Mocked GraphQL response for two teams
			responseIndex := 0
			responseLength := len(tt.fields.responses)
			testServer := httptest.NewTLSServer(
				http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					if r.RequestURI != "/api/graphql" {
						t.Errorf("got unexpected request at %q", r.RequestURI)
						http.Error(w, "not found", http.StatusNotFound)
						return
					}
					Assert(t, responseIndex < responseLength, "requesting more responses than are defined")
					response := tt.fields.responses[responseIndex]
					responseIndex++
					w.WriteHeader(response.httpCode)
					w.Write([]byte(response.body)) // nolint: errcheck
				}))
			testServerURL, err := url.Parse(testServer.URL)
			Ok(t, err)
			client, err := vcs.NewGithubClient(testServerURL.Host, &vcs.GithubUserCredentials{"user", "pass", ""}, vcs.GithubConfig{}, 0, logging.NewNoopLogger(t))
			Ok(t, err)
			defer disableSSLVerification()()
			if err := client.DiscardReviews(tt.args.repo, tt.args.pull); (err != nil) != tt.wantErr {
				t.Errorf("DiscardReviews() error = %v, wantErr %v", err, tt.wantErr)
			}
			Equals(t, responseLength, responseIndex) // check if all defined requests have been used
		})
	}
}

func TestGithubClient_GetPullLabels(t *testing.T) {
	logger := logging.NewNoopLogger(t)
	resp := `{
	  "url": "https://api.github.com/repos/runatlantis/atlantis/pulls/1",
	  "id": 167530667,
	  "merge_commit_sha": "3fe6aa34bc25ac3720e639fcad41b428e83bdb37",
	  "labels": [
		{
		  "id": 1303230720,
		  "node_id": "MDU6TGFiZWwxMzAzMjMwNzIw",
		  "url": "https://api.github.com/repos/runatlantis/atlantis/labels/docs",
		  "name": "docs",
		  "color": "d87165",
		  "default": false,
		  "description": "Documentation"
		},
		{
		  "id": 2552271640,
		  "node_id": "MDU6TGFiZWwyNTUyMjcxNjQw",
		  "url": "https://api.github.com/repos/runatlantis/atlantis/labels/go",
		  "name": "go",
		  "color": "16e2e2",
		  "default": false,
		  "description": "Pull requests that update Go code"
		},
		{
		  "id": 2696098981,
		  "node_id": "MDU6TGFiZWwyNjk2MDk4OTgx",
		  "url": "https://api.github.com/repos/runatlantis/atlantis/labels/needs%20tests",
		  "name": "needs tests",
		  "color": "FBB1DE",
		  "default": false,
		  "description": "Change requires tests"
		},
		{
		  "id": 4439792681,
		  "node_id": "LA_kwDOBy76Zc8AAAABCKHcKQ",
		  "url": "https://api.github.com/repos/runatlantis/atlantis/labels/work-in-progress",
		  "name": "work-in-progress",
		  "color": "B1E20A",
		  "default": false,
		  "description": ""
		}
	  ]
	}`
	testServer := httptest.NewTLSServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.RequestURI {
			case "/api/v3/repos/runatlantis/atlantis/pulls/1":
				w.Write([]byte(resp)) // nolint: errcheck
			default:
				t.Errorf("got unexpected request at %q", r.RequestURI)
				http.Error(w, "not found", http.StatusNotFound)
				return
			}
		}))
	testServerURL, err := url.Parse(testServer.URL)
	Ok(t, err)
	client, err := vcs.NewGithubClient(testServerURL.Host, &vcs.GithubUserCredentials{"user", "pass", ""}, vcs.GithubConfig{}, 0, logger)
	Ok(t, err)
	defer disableSSLVerification()()

	labels, err := client.GetPullLabels(
		logger,
		models.Repo{
			Owner: "runatlantis",
			Name:  "atlantis",
		}, models.PullRequest{
			Num: 1,
		})
	Ok(t, err)
	Equals(t, []string{"docs", "go", "needs tests", "work-in-progress"}, labels)
}

func TestGithubClient_GetPullLabels_EmptyResponse(t *testing.T) {
	logger := logging.NewNoopLogger(t)
	resp := `{
	  "url": "https://api.github.com/repos/runatlantis/atlantis/pulls/1",
	  "id": 167530667,
	  "merge_commit_sha": "3fe6aa34bc25ac3720e639fcad41b428e83bdb37",
	  "labels": []
	}`
	testServer := httptest.NewTLSServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.RequestURI {
			case "/api/v3/repos/runatlantis/atlantis/pulls/1":
				w.Write([]byte(resp)) // nolint: errcheck
			default:
				t.Errorf("got unexpected request at %q", r.RequestURI)
				http.Error(w, "not found", http.StatusNotFound)
				return
			}
		}))
	testServerURL, err := url.Parse(testServer.URL)
	Ok(t, err)
	client, err := vcs.NewGithubClient(testServerURL.Host, &vcs.GithubUserCredentials{"user", "pass", ""}, vcs.GithubConfig{}, 0, logger)
	Ok(t, err)
	defer disableSSLVerification()()

	labels, err := client.GetPullLabels(
		logger,
		models.Repo{
			Owner: "runatlantis",
			Name:  "atlantis",
		},
		models.PullRequest{
			Num: 1,
		})
	Ok(t, err)
	Equals(t, 0, len(labels))
}
