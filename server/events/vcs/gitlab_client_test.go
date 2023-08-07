package vcs

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"path"
	"strings"
	"testing"
	"time"

	version "github.com/hashicorp/go-version"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/logging"
	gitlab "github.com/xanzy/go-gitlab"

	. "github.com/runatlantis/atlantis/testing"
)

// Test that the base url gets set properly.
func TestNewGitlabClient_BaseURL(t *testing.T) {
	gitlabClientUnderTest = true
	defer func() { gitlabClientUnderTest = false }()
	cases := []struct {
		Hostname   string
		ExpBaseURL string
	}{
		{
			"gitlab.com",
			"https://gitlab.com/api/v4/",
		},
		{
			"custom.domain",
			"https://custom.domain/api/v4/",
		},
		{
			"http://custom.domain",
			"http://custom.domain/api/v4/",
		},
		{
			"http://custom.domain:8080",
			"http://custom.domain:8080/api/v4/",
		},
		{
			"https://custom.domain",
			"https://custom.domain/api/v4/",
		},
		{
			"https://custom.domain/",
			"https://custom.domain/api/v4/",
		},
		{
			"https://custom.domain/basepath/",
			"https://custom.domain/basepath/api/v4/",
		},
	}

	for _, c := range cases {
		t.Run(c.Hostname, func(t *testing.T) {
			log := logging.NewNoopLogger(t)
			client, err := NewGitlabClient(c.Hostname, "token", log)
			Ok(t, err)
			Equals(t, c.ExpBaseURL, client.Client.BaseURL().String())
		})
	}
}

// This function gets called even if GitlabClient is nil
// so we need to test that.
func TestGitlabClient_SupportsCommonMarkNil(t *testing.T) {
	var gl *GitlabClient
	Equals(t, false, gl.SupportsCommonMark())
}

func TestGitlabClient_SupportsCommonMark(t *testing.T) {
	cases := []struct {
		version string
		exp     bool
	}{
		{
			"11.0",
			false,
		},
		{
			"11.1",
			true,
		},
		{
			"11.2",
			true,
		},
		{
			"12.0",
			true,
		},
	}

	for _, c := range cases {
		t.Run(c.version, func(t *testing.T) {
			vers, err := version.NewVersion(c.version)
			Ok(t, err)
			gl := GitlabClient{
				Version: vers,
			}
			Equals(t, c.exp, gl.SupportsCommonMark())
		})
	}
}

func TestGitlabClient_GetModifiedFiles(t *testing.T) {
	cases := []struct {
		attempts int
	}{
		{1}, {2}, {3},
	}

	for _, c := range cases {
		t.Run(fmt.Sprintf("Gitlab returns MR changes after %d attempts", c.attempts), func(t *testing.T) {
			numAttempts := 0
			testServer := httptest.NewServer(
				http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					switch r.RequestURI {
					case "/api/v4/projects/lkysow%2Fatlantis-example/merge_requests/8312/changes?page=1&per_page=100":
						w.WriteHeader(200)
						numAttempts += 1
						if numAttempts < c.attempts {
							w.Write([]byte(changesPending)) // nolint: errcheck
							t.Logf("returning changesPending for attempt %d", numAttempts)
							return
						}
						t.Logf("returning changesAvailable for attempt %d", numAttempts)
						w.Write([]byte(changesAvailable)) // nolint: errcheck
					default:
						t.Errorf("got unexpected request at %q", r.RequestURI)
						http.Error(w, "not found", http.StatusNotFound)
					}
				}))

			internalClient, err := gitlab.NewClient("token", gitlab.WithBaseURL(testServer.URL))
			Ok(t, err)
			client := &GitlabClient{
				Client:          internalClient,
				Version:         nil,
				PollingInterval: time.Second * 0,
				PollingTimeout:  time.Second * 10,
			}

			filenames, err := client.GetModifiedFiles(
				models.Repo{
					FullName: "lkysow/atlantis-example",
					Owner:    "lkysow",
					Name:     "atlantis-example",
				},
				models.PullRequest{
					Num: 8312,
					BaseRepo: models.Repo{
						FullName: "lkysow/atlantis-example",
						Owner:    "lkysow",
						Name:     "atlantis-example",
					},
				})
			Ok(t, err)
			Equals(t, []string{"somefile.yaml"}, filenames)
		})
	}

}

func TestGitlabClient_MergePull(t *testing.T) {
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
					case "/api/v4/projects/runatlantis%2Fatlantis/merge_requests/1":
						w.WriteHeader(http.StatusOK)
						w.Write([]byte(pipelineSuccess)) // nolint: errcheck
					case "/api/v4/projects/4580910":
						w.WriteHeader(http.StatusOK)
						w.Write([]byte(projectSuccess)) // nolint: errcheck
					case "/api/v4/":
						// Rate limiter requests.
						w.WriteHeader(http.StatusOK)
					default:
						t.Errorf("got unexpected request at %q", r.RequestURI)
						http.Error(w, "not found", http.StatusNotFound)
					}
				}))

			internalClient, err := gitlab.NewClient("token", gitlab.WithBaseURL(testServer.URL))
			Ok(t, err)
			client := &GitlabClient{
				Client:  internalClient,
				Version: nil,
			}

			err = client.MergePull(models.PullRequest{
				Num: 1,
				BaseRepo: models.Repo{
					FullName: "runatlantis/atlantis",
					Owner:    "runatlantis",
					Name:     "atlantis",
				},
			}, models.PullRequestOptions{
				DeleteSourceBranchOnMerge: false,
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

func TestGitlabClient_UpdateStatus(t *testing.T) {
	cases := []struct {
		status   models.CommitStatus
		expState string
	}{
		{
			models.PendingCommitStatus,
			"running",
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

						body, err := io.ReadAll(r.Body)
						Ok(t, err)
						exp := fmt.Sprintf(`{"state":"%s","ref":"test","context":"src","target_url":"https://google.com","description":"description"}`, c.expState)
						Equals(t, exp, string(body))
						defer r.Body.Close()  // nolint: errcheck
						w.Write([]byte("{}")) // nolint: errcheck
					case "/api/v4/projects/runatlantis%2Fatlantis/merge_requests/1":
						w.WriteHeader(http.StatusOK)
						w.Write([]byte(pipelineSuccess)) // nolint: errcheck
					case "/api/v4/":
						// Rate limiter requests.
						w.WriteHeader(http.StatusOK)
					default:
						t.Errorf("got unexpected request at %q", r.RequestURI)
						http.Error(w, "not found", http.StatusNotFound)
					}
				}))

			internalClient, err := gitlab.NewClient("token", gitlab.WithBaseURL(testServer.URL))
			Ok(t, err)
			client := &GitlabClient{
				Client:  internalClient,
				Version: nil,
			}

			repo := models.Repo{
				FullName: "runatlantis/atlantis",
				Owner:    "runatlantis",
				Name:     "atlantis",
			}
			err = client.UpdateStatus(repo, models.PullRequest{
				Num:        1,
				BaseRepo:   repo,
				HeadCommit: "sha",
				HeadBranch: "test",
			}, c.status, "src", "description", "https://google.com")
			Ok(t, err)
			Assert(t, gotRequest, "expected to get the request")
		})
	}
}

func TestGitlabClient_PullIsMergeable(t *testing.T) {
	gitlabClientUnderTest = true
	gitlabVersionOver15_6 := "15.8.3-ee"
	gitlabVersion15_6 := "15.6.0-ee"
	gitlabVersionUnder15_6 := "15.3.2-ce"
	gitlabServerVersions := []string{gitlabVersionOver15_6, gitlabVersion15_6, gitlabVersionUnder15_6}
	vcsStatusName := "atlantis-test"
	defaultMR = 1
	noHeadPipelineMR = 2
	cases := []struct {
		statusName    string
		status        models.CommitStatus
		gitlabVersion []string
		mrId          int
		expState      bool
	}{
		{
			fmt.Sprintf("%s/apply: resource/default", vcsStatusName),
			models.FailedCommitStatus,
			gitlabServerVersions,
			defaultMr,
			true,
		},
		{
			fmt.Sprintf("%s/apply", vcsStatusName),
			models.FailedCommitStatus,
			gitlabServerVersions,
			defaultMr,
			true,
		},
		{
			fmt.Sprintf("%s/plan: resource/default", vcsStatusName),
			models.FailedCommitStatus,
			gitlabServerVersions,
			defaultMr,
			false,
		},
		{
			fmt.Sprintf("%s/plan", vcsStatusName),
			models.PendingCommitStatus,
			gitlabServerVersions,
			defaultMr,
			false,
		},
		{
			fmt.Sprintf("%s/plan", vcsStatusName),
			models.SuccessCommitStatus,
			gitlabServerVersions,
			defaultMr,
			true,
		},
		{
			fmt.Sprintf("%s/plan", vcsStatusName),
			models.PendingCommitStatus,
			gitlabServerVersions,
			noHeadPipelineMR,
			false,
		},
		{
			fmt.Sprintf("%s/plan", vcsStatusName),
			models.FailedCommitStatus,
			gitlabServerVersions,
			noHeadPipelineMR,
			false,
		},
		{
			fmt.Sprintf("%s/plan", vcsStatusName),
			models.SuccessCommitStatus,
			gitlabServerVersions,
			noHeadPipelineMR,
			true,
		},
	}
	for _, serverVersion := range gitlabServerVersions {
		for _, c := range cases {
			t.Run(c.statusName, func(t *testing.T) {
				testServer := httptest.NewServer(
					http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						switch r.RequestURI {
						case "/api/v4/":
							// Rate limiter requests.
							w.WriteHeader(http.StatusOK)
						case "/api/v4/projects/runatlantis%2Fatlantis/merge_requests/1":
							w.WriteHeader(http.StatusOK)
							w.Write([]byte(pipelineSuccess)) // nolint: errcheck
						case "/api/v4/projects/runatlantis%2Fatlantis/merge_requests/2":
							w.WriteHeader(http.StatusOK)
							w.Write([]byte(headPipelineNotAvailable)) // nolint: errcheck
						case "/api/v4/projects/4580910":
							w.WriteHeader(http.StatusOK)
							w.Write([]byte(projectSuccess)) // nolint: errcheck
						case "/api/v4/projects/4580910/repository/commits/67cb91d3f6198189f433c045154a885784ba6977/statuses":
							w.WriteHeader(http.StatusOK)
							response := fmt.Sprintf(`[{"id":133702594,"sha":"67cb91d3f6198189f433c045154a885784ba6977","ref":"patch-1","status":"%s","name":"%s","target_url":null,"description":"ApplySuccess","created_at":"2018-12-12T18:31:57.957Z","started_at":null,"finished_at":"2018-12-12T18:31:58.480Z","allow_failure":false,"coverage":null,"author":{"id":1755902,"username":"lkysow","name":"LukeKysow","state":"active","avatar_url":"https://secure.gravatar.com/avatar/25fd57e71590fe28736624ff24d41c5f?s=80&d=identicon","web_url":"https://gitlab.com/lkysow"}}]`, c.status, c.statusName)
							w.Write([]byte(response)) // nolint: errcheck
						case "/api/v4/version":
							w.WriteHeader(http.StatusOK)
							w.Header().Set("Content-Type", "application/json")
							type version struct {
								Version string
							}
							v := version{Version: serverVersion}
							json.NewEncoder(w).Encode(v)
						default:
							t.Errorf("got unexpected request at %q", r.RequestURI)
							http.Error(w, "not found", http.StatusNotFound)
						}
					}))

				internalClient, err := gitlab.NewClient("token", gitlab.WithBaseURL(testServer.URL))
				Ok(t, err)
				client := &GitlabClient{
					Client:  internalClient,
					Version: nil,
				}

				repo := models.Repo{
					FullName: "runatlantis/atlantis",
					Owner:    "runatlantis",
					Name:     "atlantis",
					VCSHost: models.VCSHost{
						Type:     models.Gitlab,
						Hostname: "gitlab.com",
					},
				}

				mergeable, err := client.PullIsMergeable(repo, models.PullRequest{
					Num:        c.mrId,
					BaseRepo:   repo,
					HeadCommit: "sha",
				}, vcsStatusName)

				Ok(t, err)
				Equals(t, c.expState, mergeable)
			})
		}
	}
}

func TestGitlabClient_MarkdownPullLink(t *testing.T) {
	gitlabClientUnderTest = true
	defer func() { gitlabClientUnderTest = false }()
	client, err := NewGitlabClient("gitlab.com", "token", nil)
	Ok(t, err)
	pull := models.PullRequest{Num: 1}
	s, _ := client.MarkdownPullLink(pull)
	exp := "!1"
	Equals(t, exp, s)
}

func TestGitlabClient_HideOldComments(t *testing.T) {
	type notePutCallDetails struct {
		noteID  string
		comment []string
	}
	type jsonBody struct {
		Body string
	}

	authorID := 1
	authorUserName := "pipin"
	authorEmail := "admin@example.com"
	pullNum := 123

	userCommentIDs := [1]string{"1"}
	planCommentIDs := [2]string{"3", "5"}
	systemCommentIDs := [1]string{"4"}
	summaryCommentIDs := [1]string{"2"}
	planComments := [3]string{"plan comment 1", "plan comment 2", "plan comment 3"}
	summaryHeader := fmt.Sprintf("<!--- +-Superseded Command-+ ---><details><summary>Superseded Atlantis %s</summary>",
		command.Plan.TitleString())
	summaryFooter := "</details>"
	lineFeed := "\\n"

	issueResp := "[" +
		fmt.Sprintf(`{"id":%s,"body":"User comment","author":{"id": %d, "username":"%s", "email":"%s"},"system": false,"project_id": %d}`,
			userCommentIDs[0], authorID, authorUserName, authorEmail, pullNum) + "," +
		fmt.Sprintf(`{"id":%s,"body":"%s","author":{"id": %d, "username":"%s", "email":"%s"},"system": false,"project_id": %d}`,
			summaryCommentIDs[0], summaryHeader+lineFeed+planComments[2]+lineFeed+summaryFooter, authorID, authorUserName, authorEmail, pullNum) + "," +
		fmt.Sprintf(`{"id":%s,"body":"%s","author":{"id": %d, "username":"%s", "email":"%s"},"system": false,"project_id": %d}`,
			planCommentIDs[0], planComments[0], authorID, authorUserName, authorEmail, pullNum) + "," +
		fmt.Sprintf(`{"id":%s,"body":"System comment","author":{"id": %d, "username":"%s", "email":"%s"},"system": true,"project_id": %d}`,
			systemCommentIDs[0], authorID, authorUserName, authorEmail, pullNum) + "," +
		fmt.Sprintf(`{"id":%s,"body":"%s","author":{"id": %d, "username":"%s", "email":"%s"},"system": false,"project_id": %d}`,
			planCommentIDs[1], planComments[1], authorID, authorUserName, authorEmail, pullNum) +
		"]"

	gitlabClientUnderTest = true
	defer func() { gitlabClientUnderTest = false }()
	gotNotePutCalls := make([]notePutCallDetails, 0, 1)
	testServer := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.Method {
			case "GET":
				switch r.RequestURI {
				case "/api/v4/user":
					w.WriteHeader(http.StatusOK)
					w.Header().Set("Content-Type", "application/json")
					response := fmt.Sprintf(`{"id": %d,"username": "%s", "email": "%s"}`, authorID, authorUserName, authorEmail)
					w.Write([]byte(response)) // nolint: errcheck
				case fmt.Sprintf("/api/v4/projects/runatlantis%%2Fatlantis/merge_requests/%d/notes?order_by=created_at&sort=asc", pullNum):
					w.WriteHeader(http.StatusOK)
					response := issueResp
					w.Write([]byte(response)) // nolint: errcheck
				default:
					t.Errorf("got unexpected request at %q", r.RequestURI)
					http.Error(w, "not found", http.StatusNotFound)
				}
			case "PUT":
				switch {
				case strings.HasPrefix(r.RequestURI, fmt.Sprintf("/api/v4/projects/runatlantis%%2Fatlantis/merge_requests/%d/notes/", pullNum)):
					w.WriteHeader(http.StatusOK)
					var body jsonBody
					json.NewDecoder(r.Body).Decode(&body)
					notePutCallDetail := notePutCallDetails{
						noteID:  path.Base(r.RequestURI),
						comment: strings.Split(body.Body, "\n"),
					}
					gotNotePutCalls = append(gotNotePutCalls, notePutCallDetail)
					response := "{}"
					w.Write([]byte(response)) // nolint: errcheck
				default:
					t.Errorf("got unexpected request at %q", r.RequestURI)
					http.Error(w, "not found", http.StatusNotFound)
				}
			default:
				t.Errorf("got unexpected method at %q", r.Method)
				http.Error(w, "not found", http.StatusNotFound)
			}
		}),
	)

	internalClient, err := gitlab.NewClient("token", gitlab.WithBaseURL(testServer.URL))
	Ok(t, err)
	client := &GitlabClient{
		Client:  internalClient,
		Version: nil,
		logger:  logging.NewNoopLogger(t),
	}

	repo := models.Repo{
		FullName: "runatlantis/atlantis",
		Owner:    "runatlantis",
		Name:     "atlantis",
		VCSHost: models.VCSHost{
			Type:     models.Gitlab,
			Hostname: "gitlab.com",
		},
	}

	err = client.HidePrevCommandComments(repo, pullNum, command.Plan.TitleString())
	Ok(t, err)

	// Check the correct number of plan comments have been processed
	Equals(t, len(planCommentIDs), len(gotNotePutCalls))
	// Check the first plan comment has been currectly summarised
	Equals(t, planCommentIDs[0], gotNotePutCalls[0].noteID)
	Equals(t, summaryHeader, gotNotePutCalls[0].comment[0])
	Equals(t, planComments[0], gotNotePutCalls[0].comment[1])
	Equals(t, summaryFooter, gotNotePutCalls[0].comment[2])
	// Check the second plan comment has been currectly summarised
	Equals(t, planCommentIDs[1], gotNotePutCalls[1].noteID)
	Equals(t, summaryHeader, gotNotePutCalls[1].comment[0])
	Equals(t, planComments[1], gotNotePutCalls[1].comment[1])
	Equals(t, summaryFooter, gotNotePutCalls[1].comment[2])
}

var mergeSuccess = `{"id":22461274,"iid":13,"project_id":4580910,"title":"Update main.tf","description":"","state":"merged","created_at":"2019-01-15T18:27:29.375Z","updated_at":"2019-01-25T17:28:01.437Z","merged_by":{"id":1755902,"name":"Luke Kysow","username":"lkysow","state":"active","avatar_url":"https://secure.gravatar.com/avatar/25fd57e71590fe28736624ff24d41c5f?s=80\u0026d=identicon","web_url":"https://gitlab.com/lkysow"},"merged_at":"2019-01-25T17:28:01.459Z","closed_by":null,"closed_at":null,"target_branch":"patch-1","source_branch":"patch-1-merger","upvotes":0,"downvotes":0,"author":{"id":1755902,"name":"Luke Kysow","username":"lkysow","state":"active","avatar_url":"https://secure.gravatar.com/avatar/25fd57e71590fe28736624ff24d41c5f?s=80\u0026d=identicon","web_url":"https://gitlab.com/lkysow"},"assignee":null,"source_project_id":4580910,"target_project_id":4580910,"labels":[],"work_in_progress":false,"milestone":null,"merge_when_pipeline_succeeds":false,"merge_status":"can_be_merged","detailed_merge_status":"mergeable","sha":"cb86d70f464632bdfbe1bb9bc0f2f9d847a774a0","merge_commit_sha":"c9b336f1c71d3e64810b8cfa2abcfab232d6bff6","user_notes_count":0,"discussion_locked":null,"should_remove_source_branch":null,"force_remove_source_branch":false,"web_url":"https://gitlab.com/lkysow/atlantis-example/merge_requests/13","time_stats":{"time_estimate":0,"total_time_spent":0,"human_time_estimate":null,"human_total_time_spent":null},"squash":false,"subscribed":true,"changes_count":"1","latest_build_started_at":null,"latest_build_finished_at":null,"first_deployed_to_production_at":null,"pipeline":null,"diff_refs":{"base_sha":"67cb91d3f6198189f433c045154a885784ba6977","head_sha":"cb86d70f464632bdfbe1bb9bc0f2f9d847a774a0","start_sha":"67cb91d3f6198189f433c045154a885784ba6977"},"merge_error":null,"approvals_before_merge":null}`
var pipelineSuccess = `{"id": 22461274,"iid": 13,"project_id": 4580910,"title": "Update main.tf","description": "","state": "opened","created_at": "2019-01-15T18:27:29.375Z","updated_at": "2019-01-25T17:28:01.437Z","merged_by": null,"merged_at": null,"closed_by": null,"closed_at": null,"target_branch": "patch-1","source_branch": "patch-1-merger","user_notes_count": 0,"upvotes": 0,"downvotes": 0,"author": {"id": 1755902,"name": "Luke Kysow","username": "lkysow","state": "active","avatar_url": "https://secure.gravatar.com/avatar/25fd57e71590fe28736624ff24d41c5f?s=80\u0026d=identicon","web_url": "https://gitlab.com/lkysow"},"assignee": null,"reviewers": [],"source_project_id": 4580910,"target_project_id": 4580910,"labels": [],"work_in_progress": false,"milestone": null,"merge_when_pipeline_succeeds": false,"merge_status": "can_be_merged","detailed_merge_status": "mergeable","sha": "cb86d70f464632bdfbe1bb9bc0f2f9d847a774a0","merge_commit_sha": null,"squash_commit_sha": null,"discussion_locked": null,"should_remove_source_branch": null,"force_remove_source_branch": true,"reference": "!13","references": {"short": "!13","relative": "!13","full": "lkysow/atlantis-example!13"},"web_url": "https://gitlab.com/lkysow/atlantis-example/merge_requests/13","time_stats": {"time_estimate": 0,"total_time_spent": 0,"human_time_estimate": null,"human_total_time_spent": null},"squash": true,"task_completion_status": {"count": 0,"completed_count": 0},"has_conflicts": false,"blocking_discussions_resolved": true,"approvals_before_merge": null,"subscribed": false,"changes_count": "1","latest_build_started_at": "2019-01-15T18:27:29.375Z","latest_build_finished_at": "2019-01-25T17:28:01.437Z","first_deployed_to_production_at": null,"pipeline": {"id": 488598,"sha": "67cb91d3f6198189f433c045154a885784ba6977","ref": "patch-1-merger","status": "success","created_at": "2019-01-15T18:27:29.375Z","updated_at": "2019-01-25T17:28:01.437Z","web_url": "https://gitlab.com/lkysow/atlantis-example/-/pipelines/488598"},"head_pipeline": {"id": 488598,"sha": "67cb91d3f6198189f433c045154a885784ba6977","ref": "patch-1-merger","status": "success","created_at": "2019-01-15T18:27:29.375Z","updated_at": "2019-01-25T17:28:01.437Z","web_url": "https://gitlab.com/lkysow/atlantis-example/-/pipelines/488598","before_sha": "0000000000000000000000000000000000000000","tag": false,"yaml_errors": null,"user": {"id": 1755902,"name": "Luke Kysow","username": "lkysow","state": "active","avatar_url": "https://secure.gravatar.com/avatar/25fd57e71590fe28736624ff24d41c5f?s=80\u0026d=identicon","web_url": "https://gitlab.com/lkysow"},"started_at": "2019-01-15T18:27:29.375Z","finished_at": "2019-01-25T17:28:01.437Z","committed_at": null,"duration": 31,"coverage": null,"detailed_status": {"icon": "status_success","text": "passed","label": "passed","group": "success","tooltip": "passed","has_details": true,"details_path": "/lkysow/atlantis-example/-/pipelines/488598","illustration": null,"favicon": "/assets/ci_favicons/favicon_status_success-8451333011eee8ce9f2ab25dc487fe24a8758c694827a582f17f42b0a90446a2.png"}},"diff_refs": {"base_sha": "67cb91d3f6198189f433c045154a885784ba6977","head_sha": "cb86d70f464632bdfbe1bb9bc0f2f9d847a774a0","start_sha": "67cb91d3f6198189f433c045154a885784ba6977"},"merge_error": null,"first_contribution": false,"user": {"can_merge": true}}`
var projectSuccess = `{"id": 4580910,"description": "","name": "atlantis-example","name_with_namespace": "lkysow / atlantis-example","path": "atlantis-example","path_with_namespace": "lkysow/atlantis-example","created_at": "2018-04-30T13:44:28.367Z","default_branch": "patch-1","tag_list": [],"ssh_url_to_repo": "git@gitlab.com:lkysow/atlantis-example.git","http_url_to_repo": "https://gitlab.com/lkysow/atlantis-example.git","web_url": "https://gitlab.com/lkysow/atlantis-example","readme_url": "https://gitlab.com/lkysow/atlantis-example/-/blob/main/README.md","avatar_url": "https://gitlab.com/uploads/-/system/project/avatar/4580910/avatar.png","forks_count": 0,"star_count": 7,"last_activity_at": "2021-06-29T21:10:43.968Z","namespace": {"id": 1,"name": "lkysow","path": "lkysow","kind": "group","full_path": "lkysow","parent_id": 1,"avatar_url": "/uploads/-/system/group/avatar/1651/platform.png","web_url": "https://gitlab.com/groups/lkysow"},"_links": {"self": "https://gitlab.com/api/v4/projects/4580910","issues": "https://gitlab.com/api/v4/projects/4580910/issues","merge_requests": "https://gitlab.com/api/v4/projects/4580910/merge_requests","repo_branches": "https://gitlab.com/api/v4/projects/4580910/repository/branches","labels": "https://gitlab.com/api/v4/projects/4580910/labels","events": "https://gitlab.com/api/v4/projects/4580910/events","members": "https://gitlab.com/api/v4/projects/4580910/members"},"packages_enabled": false,"empty_repo": false,"archived": false,"visibility": "private","resolve_outdated_diff_discussions": false,"container_registry_enabled": false,"container_expiration_policy": {"cadence": "1d","enabled": false,"keep_n": 10,"older_than": "90d","name_regex": ".*","name_regex_keep": null,"next_run_at": "2021-05-01T13:44:28.397Z"},"issues_enabled": true,"merge_requests_enabled": true,"wiki_enabled": false,"jobs_enabled": true,"snippets_enabled": true,"service_desk_enabled": false,"service_desk_address": null,"can_create_merge_request_in": true,"issues_access_level": "private","repository_access_level": "enabled","merge_requests_access_level": "enabled","forking_access_level": "enabled","wiki_access_level": "disabled","builds_access_level": "enabled","snippets_access_level": "enabled","pages_access_level": "private","operations_access_level": "disabled","analytics_access_level": "enabled","emails_disabled": null,"shared_runners_enabled": true,"lfs_enabled": false,"creator_id": 818,"import_status": "none","import_error": null,"open_issues_count": 0,"runners_token": "1234456","ci_default_git_depth": 50,"ci_forward_deployment_enabled": true,"public_jobs": true,"build_git_strategy": "fetch","build_timeout": 3600,"auto_cancel_pending_pipelines": "enabled","build_coverage_regex": null,"ci_config_path": "","shared_with_groups": [],"only_allow_merge_if_pipeline_succeeds": true,"allow_merge_on_skipped_pipeline": false,"restrict_user_defined_variables": false,"request_access_enabled": true,"only_allow_merge_if_all_discussions_are_resolved": true,"remove_source_branch_after_merge": true,"printing_merge_request_link_enabled": true,"merge_method": "merge","suggestion_commit_message": "","auto_devops_enabled": false,"auto_devops_deploy_strategy": "continuous","autoclose_referenced_issues": true,"repository_storage": "default","approvals_before_merge": 0,"mirror": false,"external_authorization_classification_label": null,"marked_for_deletion_at": null,"marked_for_deletion_on": null,"requirements_enabled": false,"compliance_frameworks": [],"permissions": {"project_access": null,"group_access": {"access_level": 50,"notification_level": 3}}}`
var changesPending = `{"id":8312,"iid":102,"target_branch":"main","source_branch":"TestBranch","project_id":3771,"title":"Update somefile.yaml","state":"opened","created_at":"2023-03-14T13:43:17.895Z","updated_at":"2023-03-14T13:43:17.895Z","upvotes":0,"downvotes":0,"author":{"id":1755902,"name":"Luke Kysow","username":"lkysow","state":"active","avatar_url":"https://secure.gravatar.com/avatar/25fd57e71590fe28736624ff24d41c5f?s=80\\u0026d=identicon","web_url":"https://gitlab.com/lkysow"},"assignee":null,"assignees":[],"reviewers":[],"source_project_id":3771,"target_project_id":3771,"labels":"","description":"","draft":false,"work_in_progress":false,"milestone":null,"merge_when_pipeline_succeeds":false,"detailed_merge_status":"checking","merge_error":"","merged_by":null,"merged_at":null,"closed_by":null,"closed_at":null,"subscribed":false,"sha":"cb86d70f464632bdfbe1bb9bc0f2f9d847a774a0","merge_commit_sha":"","squash_commit_sha":"","user_notes_count":0,"changes_count":"","should_remove_source_branch":false,"force_remove_source_branch":true,"allow_collaboration":false,"web_url":"https://gitlab.com/lkysow/atlantis-example/merge_requests/13","references":{"short":"!13","relative":"!13","full":"lkysow/atlantis-example!13"},"discussion_locked":false,"changes":[],"user":{"can_merge":true},"time_stats":{"human_time_estimate":"","human_total_time_spent":"","time_estimate":0,"total_time_spent":0},"squash":false,"pipeline":null,"head_pipeline":null,"diff_refs":{"base_sha":"","head_sha":"","start_sha":""},"diverged_commits_count":0,"rebase_in_progress":false,"approvals_before_merge":0,"reference":"!13","first_contribution":false,"task_completion_status":{"count":0,"completed_count":0},"has_conflicts":false,"blocking_discussions_resolved":true,"overflow":false,"merge_status":"checking"}`
var changesAvailable = `{"id":8312,"iid":102,"target_branch":"main","source_branch":"TestBranch","project_id":3771,"title":"Update somefile.yaml","state":"opened","created_at":"2023-03-14T13:43:17.895Z","updated_at":"2023-03-14T13:43:59.978Z","upvotes":0,"downvotes":0,"author":{"id":1755902,"name":"Luke Kysow","username":"lkysow","state":"active","avatar_url":"https://secure.gravatar.com/avatar/25fd57e71590fe28736624ff24d41c5f?s=80\\u0026d=identicon","web_url":"https://gitlab.com/lkysow"},"assignee":null,"assignees":[],"reviewers":[],"source_project_id":3771,"target_project_id":3771,"labels":[],"description":"","draft":false,"work_in_progress":false,"milestone":null,"merge_when_pipeline_succeeds":false,"detailed_merge_status":"not_approved","merge_error":"","merged_by":null,"merged_at":null,"closed_by":null,"closed_at":null,"subscribed":false,"sha":"cb86d70f464632bdfbe1bb9bc0f2f9d847a774a0","merge_commit_sha":null,"squash_commit_sha":null,"user_notes_count":0,"changes_count":"1","should_remove_source_branch":null,"force_remove_source_branch":true,"allow_collaboration":false,"web_url":"https://gitlab.com/lkysow/atlantis-example/merge_requests/13","references":{"short":"!13","relative":"!13","full":"lkysow/atlantis-example!13"},"discussion_locked":null,"changes":[{"old_path":"somefile.yaml","new_path":"somefile.yaml","a_mode":"100644","b_mode":"100644","diff":"--- a/somefile.yaml\\ +++ b/somefile.yaml\\ @@ -1 +1 @@\\ -gud\\ +good","new_file":false,"renamed_file":false,"deleted_file":false}],"user":{"can_merge":true},"time_stats":{"human_time_estimate":null,"human_total_time_spent":null,"time_estimate":0,"total_time_spent":0},"squash":false,"pipeline":null,"head_pipeline":null,"diff_refs":{"base_sha":"67cb91d3f6198189f433c045154a885784ba6977","head_sha":"cb86d70f464632bdfbe1bb9bc0f2f9d847a774a0","start_sha":"67cb91d3f6198189f433c045154a885784ba6977"},"approvals_before_merge":null,"reference":"!13","task_completion_status":{"count":0,"completed_count":0},"has_conflicts":false,"blocking_discussions_resolved":true,"overflow":false,"merge_status":"can_be_merged"}`
var headPipelineNotAvailable = `{"id": 22461274,"iid": 13,"project_id": 4580910,"title": "Update main.tf","description": "","state": "opened","created_at": "2019-01-15T18:27:29.375Z","updated_at": "2019-01-25T17:28:01.437Z","merged_by": null,"merged_at": null,"closed_by": null,"closed_at": null,"target_branch": "patch-1","source_branch": "patch-1-merger","user_notes_count": 0,"upvotes": 0,"downvotes": 0,"author": {"id": 1755902,"name": "Luke Kysow","username": "lkysow","state": "active","avatar_url": "https://secure.gravatar.com/avatar/25fd57e71590fe28736624ff24d41c5f?s=80\u0026d=identicon","web_url": "https://gitlab.com/lkysow"},"assignee": null,"reviewers": [],"source_project_id": 4580910,"target_project_id": 4580910,"labels": [],"work_in_progress": false,"milestone": null,"merge_when_pipeline_succeeds": false,"merge_status": "can_be_merged","detailed_merge_status": "mergeable","sha": "cb86d70f464632bdfbe1bb9bc0f2f9d847a774a0","merge_commit_sha": null,"squash_commit_sha": null,"discussion_locked": null,"should_remove_source_branch": null,"force_remove_source_branch": true,"reference": "!13","references": {"short": "!13","relative": "!13","full": "lkysow/atlantis-example!13"},"web_url": "https://gitlab.com/lkysow/atlantis-example/merge_requests/13","time_stats": {"time_estimate": 0,"total_time_spent": 0,"human_time_estimate": null,"human_total_time_spent": null},"squash": true,"task_completion_status": {"count": 0,"completed_count": 0},"has_conflicts": false,"blocking_discussions_resolved": true,"approvals_before_merge": null,"subscribed": false,"changes_count": "1","latest_build_started_at": "2019-01-15T18:27:29.375Z","latest_build_finished_at": "2019-01-25T17:28:01.437Z","first_deployed_to_production_at": null,"pipeline": {"id": 488598,"sha": "67cb91d3f6198189f433c045154a885784ba6977","ref": "patch-1-merger","status": "success","created_at": "2019-01-15T18:27:29.375Z","updated_at": "2019-01-25T17:28:01.437Z","web_url": "https://gitlab.com/lkysow/atlantis-example/-/pipelines/488598"},"head_pipeline": null,"diff_refs": {"base_sha": "67cb91d3f6198189f433c045154a885784ba6977","head_sha": "cb86d70f464632bdfbe1bb9bc0f2f9d847a774a0","start_sha": "67cb91d3f6198189f433c045154a885784ba6977"},"merge_error": null,"first_contribution": false,"user": {"can_merge": true}}`
