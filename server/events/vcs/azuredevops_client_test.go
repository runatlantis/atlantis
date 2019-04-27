package vcs

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/mcdafydd/go-azuredevops/azuredevops"
	"github.com/runatlantis/atlantis/server/events/models"
	. "github.com/runatlantis/atlantis/testing"
)

// Test that the base url gets set properly.
func TestNewAzureDevopsClient_BaseURL(t *testing.T) {
	azuredevopsClientUnderTest = true
	defer func() { azuredevopsClientUnderTest = false }()
	cases := []struct {
		Hostname   string
		Account    string
		Project    string
		ExpBaseURL string
	}{
		{
			"dev.azure.com",
			"testorg",
			"testproject",
			"https://dev.azure.com/testorg/testproject",
		},
		{
			"visualstudio.com",
			"testorg",
			"testproject",
			"https://testorg.visualstudio.com/testproject",
		},
	}

	for _, c := range cases {
		t.Run(c.Hostname, c.Account, c.Project, func(t *testing.T) {
			client, err := NewAzureDevopsClient(c.Hostname, c.Account, c.Project, "token", nil)
			Ok(t, err)
			Equals(t, c.ExpBaseURL, client.Client.BaseURL().String())
		})
	}
}

func TestAzureDevopsClient_MergePull(t *testing.T) {
	cases := []struct {
		description string
		glResponse  string
		code        int
		expErr      string
	}{
		{
			"success",
			mergeSuccess,
			200,
			"",
		},
		{
			"405",
			`{"message":"405 Method Not Allowed"}`,
			405,
			"405 {message: 405 Method Not Allowed}",
		},
		{
			"406",
			`{"message":"406 Branch cannot be merged"}`,
			406,
			"406 {message: 406 Branch cannot be merged}",
		},
	}

	for _, c := range cases {
		t.Run(c.description, func(t *testing.T) {
			testServer := httptest.NewServer(
				http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					switch r.RequestURI {
					// The first request should hit this URL.
					case "/api/v4/projects/runatlantis%2Fatlantis/merge_requests/1/merge":
						w.WriteHeader(c.code)
						w.Write([]byte(c.glResponse)) // nolint: errcheck
					default:
						t.Errorf("got unexpected request at %q", r.RequestURI)
						http.Error(w, "not found", http.StatusNotFound)
					}
				}))

			var httpClient = &http.Client{
				Timeout: time.Second * 10,
			}
			internalClient := azuredevops.NewClient(account, project, "token", httpClient)
			Ok(t, internalClient.SetBaseURL(testServer.URL))
			client := &AzureDevopsClient{
				Client:  internalClient,
				Version: nil,
			}

			err := client.MergePull(models.PullRequest{
				Num: 1,
				BaseRepo: models.Repo{
					FullName: "runatlantis/atlantis",
					Owner:    "runatlantis",
					Name:     "atlantis",
				},
			})
			if c.expErr == "" {
				Ok(t, err)
			} else {
				ErrContains(t, c.expErr, err)
				ErrContains(t, "unable to merge merge request, it may not be in a mergeable state", err)
			}
		})
	}
}

func TestAzureDevopsClient_UpdateStatus(t *testing.T) {
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
			"failed",
		},
	}
	for _, c := range cases {
		t.Run(c.expState, func(t *testing.T) {
			gotRequest := false
			testServer := httptest.NewServer(
				http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					switch r.RequestURI {
					case "/api/v4/projects/runatlantis%2Fatlantis/statuses/sha":
						gotRequest = true

						body, err := ioutil.ReadAll(r.Body)
						Ok(t, err)
						exp := fmt.Sprintf(`{"state":"%s","context":"src","target_url":"https://google.com","description":"description"}`, c.expState)
						Equals(t, exp, string(body))
						defer r.Body.Close()  // nolint: errcheck
						w.Write([]byte("{}")) // nolint: errcheck
					default:
						t.Errorf("got unexpected request at %q", r.RequestURI)
						http.Error(w, "not found", http.StatusNotFound)
					}
				}))

			var httpClient = &http.Client{
				Timeout: time.Second * 10,
			}
			internalClient := azuredevops.NewClient(account, project, "token", httpClient)
			Ok(t, internalClient.SetBaseURL(testServer.URL))
			client := &AzureDevopsClient{
				Client:  internalClient,
				Version: nil,
			}

			repo := models.Repo{
				FullName: "runatlantis/atlantis",
				Owner:    "runatlantis",
				Name:     "atlantis",
			}
			err := client.UpdateStatus(repo, models.PullRequest{
				Num:        1,
				BaseRepo:   repo,
				HeadCommit: "sha",
			}, c.status, "src", "description", "https://google.com")
			Ok(t, err)
			Assert(t, gotRequest, "expected to get the request")
		})
	}
}

var mergeSuccess = `{"id":22461274,"iid":13,"project_id":4580910,"title":"Update main.tf","description":"","state":"merged","created_at":"2019-01-15T18:27:29.375Z","updated_at":"2019-01-25T17:28:01.437Z","merged_by":{"id":1755902,"name":"Luke Kysow","username":"lkysow","state":"active","avatar_url":"https://secure.gravatar.com/avatar/25fd57e71590fe28736624ff24d41c5f?s=80\u0026d=identicon","web_url":"https://azuredevops.com/lkysow"},"merged_at":"2019-01-25T17:28:01.459Z","closed_by":null,"closed_at":null,"target_branch":"patch-1","source_branch":"patch-1-merger","upvotes":0,"downvotes":0,"author":{"id":1755902,"name":"Luke Kysow","username":"lkysow","state":"active","avatar_url":"https://secure.gravatar.com/avatar/25fd57e71590fe28736624ff24d41c5f?s=80\u0026d=identicon","web_url":"https://azuredevops.com/lkysow"},"assignee":null,"source_project_id":4580910,"target_project_id":4580910,"labels":[],"work_in_progress":false,"milestone":null,"merge_when_pipeline_succeeds":false,"merge_status":"can_be_merged","sha":"cb86d70f464632bdfbe1bb9bc0f2f9d847a774a0","merge_commit_sha":"c9b336f1c71d3e64810b8cfa2abcfab232d6bff6","user_notes_count":0,"discussion_locked":null,"should_remove_source_branch":null,"force_remove_source_branch":false,"web_url":"https://azuredevops.com/lkysow/atlantis-example/merge_requests/13","time_stats":{"time_estimate":0,"total_time_spent":0,"human_time_estimate":null,"human_total_time_spent":null},"squash":false,"subscribed":true,"changes_count":"1","latest_build_started_at":null,"latest_build_finished_at":null,"first_deployed_to_production_at":null,"pipeline":null,"diff_refs":{"base_sha":"67cb91d3f6198189f433c045154a885784ba6977","head_sha":"cb86d70f464632bdfbe1bb9bc0f2f9d847a774a0","start_sha":"67cb91d3f6198189f433c045154a885784ba6977"},"merge_error":null,"approvals_before_merge":null}`

var getChangesResp = `{
  "changeCounts": {
    "Add": 456
  },
  "changes": [
    {
      "item": {
        "gitObjectType": "blob",
        "path": "/MyWebSite/MyWebSite/favicon.ico",
        "url": "https://dev.azure.com/fabrikam/_apis/git/repositories/278d5cd2-584d-4b63-824a-2ba458937249/items/MyWebSite/MyWebSite/favicon.ico?versionType=Commit"
      },
      "changeType": "add"
    },
    {
      "item": {
        "gitObjectType": "tree",
        "path": "/MyWebSite/MyWebSite/fonts",
        "isFolder": true,
        "url": "https://dev.azure.com/fabrikam/_apis/git/repositories/278d5cd2-584d-4b63-824a-2ba458937249/items/MyWebSite/MyWebSite/fonts?versionType=Commit"
      },
      "changeType": "add"
    }
  ]
}`
