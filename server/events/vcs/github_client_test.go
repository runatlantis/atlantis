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
	"time"

	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/events/vcs"
	"github.com/runatlantis/atlantis/server/logging"
	. "github.com/runatlantis/atlantis/testing"

	"github.com/shurcooL/githubv4"
)

// GetModifiedFiles should make multiple requests if more than one page
// and concat results.
func TestGithubClient_GetModifiedFiles(t *testing.T) {
	logger := logging.NewNoopCtxLogger(t)
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
	mergeabilityChecker := vcs.NewPullMergeabilityChecker("atlantis")
	client, err := vcs.NewGithubClient(testServerURL.Host, &vcs.GithubUserCredentials{"user", "pass"}, logger, mergeabilityChecker)
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
	mergeabilityChecker := vcs.NewPullMergeabilityChecker("atlantis")
	client, err := vcs.NewGithubClient(testServerURL.Host, &vcs.GithubUserCredentials{"user", "pass"}, logging.NewNoopCtxLogger(t), mergeabilityChecker)
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

func TestGithubClient_PaginatesComments(t *testing.T) {
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

	mergeabilityChecker := vcs.NewPullMergeabilityChecker("atlantis")
	client, err := vcs.NewGithubClient(testServerURL.Host, &vcs.GithubUserCredentials{"user", "pass"}, logging.NewNoopCtxLogger(t), mergeabilityChecker)
	Ok(t, err)
	defer disableSSLVerification()()

	err = client.HidePrevCommandComments(
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
		"Plan",
	)
	Ok(t, err)
	Equals(t, 2, len(gotMinimizeCalls))
	Equals(t, "2", gotMinimizeCalls[0].Variables.Input.SubjectID)
	Equals(t, "8", gotMinimizeCalls[1].Variables.Input.SubjectID)
	Equals(t, githubv4.ReportedContentClassifiersOutdated, gotMinimizeCalls[0].Variables.Input.Classifier)
	Equals(t, githubv4.ReportedContentClassifiersOutdated, gotMinimizeCalls[1].Variables.Input.Classifier)
}

func TestGithubClient_HideOldComments(t *testing.T) {
	// Only comment 6 should be minimized, because it's by the same Atlantis bot user
	// and it has "plan" in the first line of the comment body.
	issueResp := `[
	{"node_id": "1", "body": "asd\nplan\nasd", "user": {"login": "someone-else"}},
	{"node_id": "2", "body": "asd plan\nasd", "user": {"login": "someone-else"}},
	{"node_id": "3", "body": "asdasdasd\nasdasdasd", "user": {"login": "someone-else"}},
	{"node_id": "4", "body": "asdasdasd\nasdasdasd", "user": {"login": "user"}},
	{"node_id": "5", "body": "asd\nplan\nasd", "user": {"login": "user"}},
	{"node_id": "6", "body": "asd plan\nasd", "user": {"login": "user"}},
	{"node_id": "7", "body": "asdasdasd", "user": {"login": "user"}},
	{"node_id": "8", "body": "asd plan\nasd", "user": {"login": "user"}},
	{"node_id": "9", "body": "Continued Plan from previous comment\nasd", "user": {"login": "user"}}
]`
	minimizeResp := "{}"
	type graphQLCall struct {
		Variables struct {
			Input githubv4.MinimizeCommentInput `json:"input"`
		} `json:"variables"`
	}
	gotMinimizeCalls := make([]graphQLCall, 0, 1)
	testServer := httptest.NewTLSServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.Method + " " + r.RequestURI {
			// This gets the pull request's comments.
			case "GET /api/v3/repos/owner/repo/issues/123/comments?direction=asc&sort=created":
				w.Write([]byte(issueResp)) // nolint: errcheck
				return
			case "POST /api/graphql":
				if accept, has := r.Header["Accept"]; !has || accept[0] != "application/vnd.github.queen-beryl-preview+json" {
					t.Error("missing preview header")
					http.Error(w, "bad request", http.StatusBadRequest)
					return
				}
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

	mergeabilityChecker := vcs.NewPullMergeabilityChecker("atlantis")
	client, err := vcs.NewGithubClient(testServerURL.Host, &vcs.GithubUserCredentials{"user", "pass"}, logging.NewNoopCtxLogger(t), mergeabilityChecker)
	Ok(t, err)
	defer disableSSLVerification()()

	err = client.HidePrevCommandComments(
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
		"Plan",
	)
	Ok(t, err)
	Equals(t, 3, len(gotMinimizeCalls))
	Equals(t, "6", gotMinimizeCalls[0].Variables.Input.SubjectID)
	Equals(t, "9", gotMinimizeCalls[2].Variables.Input.SubjectID)
	Equals(t, githubv4.ReportedContentClassifiersOutdated, gotMinimizeCalls[0].Variables.Input.Classifier)
}

func TestGithubClient_PullIsApproved(t *testing.T) {
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
			"state": "APPROVED",
			"html_url": "https://github.com/octocat/Hello-World/pull/12#pullrequestreview-%d",
			"pull_request_url": "https://api.github.com/repos/octocat/Hello-World/pulls/12",
			"_links": {
			  "html": {
				"href": "https://github.com/octocat/Hello-World/pull/12#pullrequestreview-%d"
			  },
			  "pull_request": {
				"href": "https://api.github.com/repos/octocat/Hello-World/pulls/12"
			  }
			},
			"submitted_at": "2019-11-17T17:43:43Z"
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
	mergeabilityChecker := vcs.NewPullMergeabilityChecker("atlantis")
	client, err := vcs.NewGithubClient(testServerURL.Host, &vcs.GithubUserCredentials{"user", "pass"}, logging.NewNoopCtxLogger(t), mergeabilityChecker)
	Ok(t, err)
	defer disableSSLVerification()()

	approvalStatus, err := client.PullIsApproved(models.Repo{
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

	timeOfApproval, err := time.Parse("2006-01-02T15:04:05Z", "2019-11-17T17:43:43Z")
	Ok(t, err)

	expApprovalStatus := models.ApprovalStatus{
		IsApproved: true,
		ApprovedBy: "octocat",
		Date:       timeOfApproval,
	}
	Equals(t, expApprovalStatus, approvalStatus)
}

func TestGithubClient_PullIsMergeable(t *testing.T) {
	combinedStatusJSON := `{
		"state": "success",
		"statuses": [%s]
	}`
	statusJSON := `{
        "url": "https://api.github.com/repos/octocat/Hello-World/statuses/6dcb09b5b57875f334f61aebed695e2e4193db5e",
        "avatar_url": "https://github.com/images/error/other_user_happy.gif",
        "id": 2,
        "node_id": "MDY6U3RhdHVzMg==",
        "state": "%s",
        "description": "Testing has completed successfully",
        "target_url": "https://ci.example.com/2000/output",
        "context": "%s",
        "created_at": "2012-08-20T01:19:13Z",
        "updated_at": "2012-08-20T01:19:13Z"
	  }`
	checksJSON := `{
		"check_runs": [
		  {
			"id": 4,
			"status": "%s",
			"conclusion": "%s",
			"name": "mighty_readme",
			"check_suite": {
			  "id": 5
			}
		  }
		]
	}`
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
	jsBytes, err := os.ReadFile("fixtures/github-pull-request.json")
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
					case "/api/v3/repos/owner/repo/commits/2/status?per_page=100":
						_, _ = w.Write([]byte(
							fmt.Sprintf(combinedStatusJSON, fmt.Sprintf(statusJSON, "success", "some_status")),
						)) // nolint: errcheck
						return
					case "/api/v3/repos/owner/repo/commits/2/check-runs?per_page=100":
						_, _ = w.Write([]byte(fmt.Sprintf(checksJSON, "completed", "success")))
						return
					default:
						t.Errorf("got unexpected request at %q", r.RequestURI)
						http.Error(w, "not found", http.StatusNotFound)
						return
					}
				}))
			testServerURL, err := url.Parse(testServer.URL)
			Ok(t, err)
			mergeabilityChecker := vcs.NewPullMergeabilityChecker("atlantis")
			client, err := vcs.NewGithubClient(testServerURL.Host, &vcs.GithubUserCredentials{"user", "pass"}, logging.NewNoopCtxLogger(t), mergeabilityChecker)
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
				Num:        1,
				HeadCommit: "2",
			})
			Ok(t, err)
			Equals(t, c.expMergeable, actMergeable)
		})
	}
}

// TODO: move this test to the mergeability checker itself
func TestGithubClient_PullisMergeable_BlockedStatus(t *testing.T) {
	// Use a real GitHub json response and edit the mergeable_state field.
	jsBytes, err := os.ReadFile("fixtures/github-pull-request.json")
	Ok(t, err)
	json := string(jsBytes)

	pullResponse := strings.Replace(json,
		`"mergeable_state": "clean"`,
		fmt.Sprintf(`"mergeable_state": "%s"`, "blocked"),
		1,
	)

	combinedStatusJSON := `{
		"state": "success",
		"statuses": [%s]
	}`
	statusJSON := `{
        "url": "https://api.github.com/repos/octocat/Hello-World/statuses/6dcb09b5b57875f334f61aebed695e2e4193db5e",
        "avatar_url": "https://github.com/images/error/other_user_happy.gif",
        "id": 2,
        "node_id": "MDY6U3RhdHVzMg==",
        "state": "%s",
        "description": "Testing has completed successfully",
        "target_url": "https://ci.example.com/2000/output",
        "context": "%s",
        "created_at": "2012-08-20T01:19:13Z",
        "updated_at": "2012-08-20T01:19:13Z"
	  }`
	checksJSON := `{
			"id": 4,
			"status": "%s",
			"conclusion": "%s",
			"name": "%s",
			"check_suite": {
			  "id": 5
			}
	}`
	combinedChecksJSON := `{
		"total_count": %d,
		"check_runs": [%s]
	}`

	completedCheckResponse := fmt.Sprintf(checksJSON, "completed", "success", "mighty-readme")

	cases := []struct {
		description  string
		statuses     []string
		checks       []string
		expMergeable bool
	}{
		{
			"apply-failure",
			[]string{},
			[]string{
				completedCheckResponse,
				fmt.Sprintf(checksJSON, "complete", "failure", "atlantis/apply"),
			},
			true,
		},
		{
			"apply-project-failure",
			[]string{},
			[]string{
				fmt.Sprintf(checksJSON, "complete", "failure", "atlantis/apply: terraform_cloud_workspace"),
				completedCheckResponse,
			},
			true,
		},
		{
			"plan+apply-failure",
			[]string{
				fmt.Sprintf(statusJSON, "failure", "atlantis/plan"),
			},
			[]string{
				completedCheckResponse,
				fmt.Sprintf(checksJSON, "complete", "failure", "atlantis/apply"),
			},
			false,
		},
		{
			"apply-failure-checks-failed",
			[]string{},
			[]string{
				fmt.Sprintf(checksJSON, "complete", "failure", "mighty-readme"),
				fmt.Sprintf(checksJSON, "complete", "failure", "atlantis/apply"),
			},
			false,
		},
		{
			"plan-success-checks-in-progress",
			[]string{
				fmt.Sprintf(statusJSON, "success", "atlantis/plan"),
			},
			[]string{
				fmt.Sprintf(checksJSON, "in_progress", "", "mighty-readme"),
			},
			false,
		},
	}

	for _, c := range cases {

		t.Run("blocked/"+c.description, func(t *testing.T) {
			testServer := httptest.NewTLSServer(
				http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					switch r.RequestURI {
					case "/api/v3/repos/owner/repo/commits/2/status?per_page=100":
						_, _ = w.Write([]byte(
							fmt.Sprintf(combinedStatusJSON, strings.Join(c.statuses, ",")),
						)) // nolint: errcheck
						return
					case "/api/v3/repos/owner/repo/commits/2/check-runs?per_page=100":
						_, _ = w.Write([]byte(
							fmt.Sprintf(combinedChecksJSON, len(c.checks), strings.Join(c.checks, ",")),
						))
						return
					case "/api/v3/repos/owner/repo/pulls/1":
						w.Write([]byte(pullResponse)) // nolint: errcheck
						return
					default:
						t.Errorf("got unexpected request at %q", r.RequestURI)
						http.Error(w, "not found", http.StatusNotFound)
						return
					}
				}))

			defer testServer.Close()

			testServerURL, err := url.Parse(testServer.URL)
			Ok(t, err)
			mergeabilityChecker := vcs.NewPullMergeabilityChecker("atlantis")
			client, err := vcs.NewGithubClient(testServerURL.Host, &vcs.GithubUserCredentials{"user", "pass"}, logging.NewNoopCtxLogger(t), mergeabilityChecker)
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
				Num:        1,
				HeadCommit: "2",
			})
			Ok(t, err)
			Equals(t, c.expMergeable, actMergeable)
		})
	}

}

func TestGithubClient_MarkdownPullLink(t *testing.T) {
	mergeabilityChecker := vcs.NewPullMergeabilityChecker("atlantis")
	client, err := vcs.NewGithubClient("hostname", &vcs.GithubUserCredentials{"user", "pass"}, logging.NewNoopCtxLogger(t), mergeabilityChecker)
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
	mergeabilityChecker := vcs.NewPullMergeabilityChecker("atlantis")
	client, err := vcs.NewGithubClient(testServerURL.Host, &vcs.GithubUserCredentials{"user", "pass"}, logging.NewNoopCtxLogger(t), mergeabilityChecker)
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
	err = client.CreateComment(repo, pull.Num, comment, "plan")
	Ok(t, err)
	err = client.CreateComment(repo, pull.Num, comment, "")
	Ok(t, err)

	body := strings.Split(githubComments[1].Body, "\n")
	firstSplit := strings.ToLower(body[0])
	body = strings.Split(githubComments[3].Body, "\n")
	secondSplit := strings.ToLower(body[0])

	Equals(t, 4, len(githubComments))
	Assert(t, strings.Contains(firstSplit, "plan"), fmt.Sprintf("comment should contain the command name but was %q", firstSplit))
	Assert(t, strings.Contains(secondSplit, "continued from previous comment"), fmt.Sprintf("comment should contain no reference to the command name but was %q", secondSplit))
}

// Test that we retry the get pull request call if it 404s.
func TestGithubClient_Retry404(t *testing.T) {
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
	mergeabilityChecker := vcs.NewPullMergeabilityChecker("atlantis")
	client, err := vcs.NewGithubClient(testServerURL.Host, &vcs.GithubUserCredentials{"user", "pass"}, logging.NewNoopCtxLogger(t), mergeabilityChecker)
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
	_, err = client.GetPullRequest(repo, 1)
	Ok(t, err)
	Equals(t, 3, numCalls)
}
