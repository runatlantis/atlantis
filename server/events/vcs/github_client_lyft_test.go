package vcs_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/events/vcs"
	"github.com/runatlantis/atlantis/server/logging"
	"github.com/stretchr/testify/assert"
)

// TODO: move this test to the mergeability checker itself
func TestLyftGithubClient_PullisMergeable_BlockedStatus(t *testing.T) {
	checkJSON := `{
		"id": 4,
		"status": "%s",
		"conclusion": "%s",
		"name": "%s",
		"check_suite": {
		  "id": 5
		}
	}`
	combinedStatusJSON := `{
		"state": "success",
		"statuses": [%s]
	}`
	combinedChecksJSON := `{
		"check_runs": [%s]
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

	// Use a real GitHub json response and edit the mergeable_state field.
	jsBytes, err := os.ReadFile("fixtures/github-pull-request.json")
	assert.NoError(t, err)
	json := string(jsBytes)

	pullResponse := strings.Replace(json,
		`"mergeable_state": "clean"`,
		fmt.Sprintf(`"mergeable_state": "%s"`, "blocked"),
		1,
	)

	cases := []struct {
		description  string
		statuses     []string
		checks       []string
		expMergeable bool
	}{
		{
			"sq-pending+owners-success+check-success",
			[]string{
				fmt.Sprintf(statusJSON, "pending", "sq-ready-to-merge"),
				fmt.Sprintf(statusJSON, "success", "_owners-check"),
			},
			[]string{
				fmt.Sprintf(checkJSON, "completed", "success", "check-name"),
			},
			true,
		},
		{
			"sq-pending+owners-missing",
			[]string{
				fmt.Sprintf(statusJSON, "pending", "sq-ready-to-merge"),
			},
			[]string{
				fmt.Sprintf(checkJSON, "completed", "success", "check-name"),
			},
			false,
		},
		{
			"sq-pending+owners-failure",
			[]string{
				fmt.Sprintf(statusJSON, "pending", "sq-ready-to-merge"),
				fmt.Sprintf(statusJSON, "failure", "_owners-check"),
			},
			[]string{
				fmt.Sprintf(checkJSON, "completed", "success", "check-name"),
			},
			false,
		},
		{
			"sq-pending+apply-failure",
			[]string{
				fmt.Sprintf(statusJSON, "pending", "sq-ready-to-merge"),
				fmt.Sprintf(statusJSON, "success", "_owners-check"),
			},
			[]string{
				fmt.Sprintf(checkJSON, "completed", "success", "check-name"),
				fmt.Sprintf(checkJSON, "completed", "failure", "atlantis/apply"),
			},
			true,
		},
		{
			"sq-pending+apply-failure+check-failure",
			[]string{
				fmt.Sprintf(statusJSON, "pending", "sq-ready-to-merge"),
				fmt.Sprintf(statusJSON, "success", "_owners-check"),
				fmt.Sprintf(statusJSON, "failure", "atlantis/apply"),
			},
			[]string{
				fmt.Sprintf(checkJSON, "in_progress", "", "check-name"),
			},
			false,
		},
		{
			"sq-pending+check_pending",
			[]string{
				fmt.Sprintf(statusJSON, "pending", "sq-ready-to-merge"),
				fmt.Sprintf(statusJSON, "success", "_owners-check"),
			},
			[]string{
				fmt.Sprintf(checkJSON, "in_progress", "", "check-name"),
			},
			false,
		},
		{
			"sq-pending+check_failure",
			[]string{
				fmt.Sprintf(statusJSON, "pending", "sq-ready-to-merge"),
				fmt.Sprintf(statusJSON, "success", "_owners-check"),
			},
			[]string{
				fmt.Sprintf(checkJSON, "complete", "failure", "check-name"),
			},
			false,
		},
		{
			"sq-pending-check+owners-success+check-success",
			[]string{
				fmt.Sprintf(statusJSON, "pending", "sq-ready-to-merge"),
				fmt.Sprintf(statusJSON, "success", "_owners-check"),
			},
			[]string{
				fmt.Sprintf(checkJSON, "queued", "", "sq-ready-to-merge"),
				fmt.Sprintf(checkJSON, "completed", "success", "check-name"),
			},
			true,
		},
	}

	for _, c := range cases {

		t.Run("blocked/"+c.description, func(t *testing.T) {
			testServer := httptest.NewTLSServer(
				http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					switch r.RequestURI {
					case "/api/v3/repos/owner/repo/pulls/1":
						w.Write([]byte(pullResponse)) // nolint: errcheck
						return
					case "/api/v3/repos/owner/repo/commits/2/status?per_page=100":
						_, _ = w.Write([]byte(
							fmt.Sprintf(combinedStatusJSON, strings.Join(c.statuses, ",")),
						)) // nolint: errcheck
						return
					case "/api/v3/repos/owner/repo/commits/2/check-runs?per_page=100":
						_, _ = w.Write([]byte(
							fmt.Sprintf(combinedChecksJSON, strings.Join(c.checks, ",")),
						))
						return
					default:
						t.Errorf("got unexpected request at %q", r.RequestURI)
						http.Error(w, "not found", http.StatusNotFound)
						return
					}
				}))

			defer testServer.Close()

			testServerURL, err := url.Parse(testServer.URL)
			assert.NoError(t, err)
			mergeabilityChecker := vcs.NewLyftPullMergeabilityChecker("atlantis")
			client, err := vcs.NewGithubClient(testServerURL.Host, &vcs.GithubUserCredentials{"user", "pass"}, logging.NewNoopCtxLogger(t), mergeabilityChecker)
			assert.NoError(t, err)
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
			assert.NoError(t, err)
			assert.Equal(t, c.expMergeable, actMergeable)
		})
	}

}
