package vcs_test

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/events/vcs"
	"github.com/runatlantis/atlantis/server/logging"
	"github.com/stretchr/testify/assert"
)

// TODO: move this test to the mergeability checker itself
func TestLyftGithubClient_PullisMergeable_BlockedStatus(t *testing.T) {
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

	// Use a real GitHub json response and edit the mergeable_state field.
	jsBytes, err := ioutil.ReadFile("fixtures/github-pull-request.json")
	assert.NoError(t, err)
	json := string(jsBytes)

	pullResponse := strings.Replace(json,
		`"mergeable_state": "clean"`,
		fmt.Sprintf(`"mergeable_state": "%s"`, "blocked"),
		1,
	)

	completedCheckResponse := fmt.Sprintf(checksJSON, "completed", "success")

	cases := []struct {
		description    string
		statuses       []string
		checksResponse string
		expMergeable   bool
	}{
		{
			"sq-pending+owners-success+check-success",
			[]string{
				fmt.Sprintf(statusJSON, "pending", "sq-ready-to-merge"),
				fmt.Sprintf(statusJSON, "success", "_owners-check"),
			},
			completedCheckResponse,
			true,
		},
		{
			"sq-pending+owners-missing",
			[]string{
				fmt.Sprintf(statusJSON, "pending", "sq-ready-to-merge"),
			},
			completedCheckResponse,
			false,
		},
		{
			"sq-pending+owners-failure",
			[]string{
				fmt.Sprintf(statusJSON, "pending", "sq-ready-to-merge"),
				fmt.Sprintf(statusJSON, "failure", "_owners-check"),
			},
			completedCheckResponse,
			false,
		},
		{
			"sq-pending+apply-failure",
			[]string{
				fmt.Sprintf(statusJSON, "pending", "sq-ready-to-merge"),
				fmt.Sprintf(statusJSON, "success", "_owners-check"),
				fmt.Sprintf(statusJSON, "failure", "atlantis/apply"),
			},
			completedCheckResponse,
			true,
		},
		{
			"sq-pending+apply-failure+check-failure",
			[]string{
				fmt.Sprintf(statusJSON, "pending", "sq-ready-to-merge"),
				fmt.Sprintf(statusJSON, "success", "_owners-check"),
				fmt.Sprintf(statusJSON, "failure", "atlantis/apply"),
			},
			fmt.Sprintf(checksJSON, "in_progress", ""),
			false,
		},
		{
			"sq-pending+check_pending",
			[]string{
				fmt.Sprintf(statusJSON, "pending", "sq-ready-to-merge"),
				fmt.Sprintf(statusJSON, "success", "_owners-check"),
			},
			fmt.Sprintf(checksJSON, "in_progress", ""),
			false,
		},
		{
			"sq-pending+check_failure",
			[]string{
				fmt.Sprintf(statusJSON, "pending", "sq-ready-to-merge"),
				fmt.Sprintf(statusJSON, "success", "_owners-check"),
			},
			fmt.Sprintf(checksJSON, "complete", "failure"),
			false,
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
						_, _ = w.Write([]byte(c.checksResponse))
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
			client, err := vcs.NewGithubClient(testServerURL.Host, &vcs.GithubUserCredentials{"user", "pass"}, logging.NewNoopLogger(t), mergeabilityChecker)
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
