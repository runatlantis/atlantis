package vcs

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
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

var projectID = 4580910

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

	changesPending, err := os.ReadFile("testdata/gitlab-changes-pending.json")
	Ok(t, err)

	changesAvailable, err := os.ReadFile("testdata/gitlab-changes-available.json")
	Ok(t, err)

	for _, c := range cases {
		t.Run(fmt.Sprintf("Gitlab returns MR changes after %d attempts", c.attempts), func(t *testing.T) {
			numAttempts := 0
			testServer := httptest.NewServer(
				http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					switch r.RequestURI {
					case "/api/v4/projects/lkysow%2Fatlantis-example/merge_requests/8312/changes?page=1&per_page=100":
						w.WriteHeader(200)
						numAttempts++
						if numAttempts < c.attempts {
							w.Write(changesPending) // nolint: errcheck
							t.Logf("returning changesPending for attempt %d", numAttempts)
							return
						}
						t.Logf("returning changesAvailable for attempt %d", numAttempts)
						w.Write(changesAvailable) // nolint: errcheck
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
				logger:          logging.NewNoopLogger(t),
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
	mergeSuccess, err := os.ReadFile("testdata/github-pull-request.json")
	Ok(t, err)

	pipelineSuccess, err := os.ReadFile("testdata/gitlab-pipeline-success.json")
	Ok(t, err)

	projectSuccess, err := os.ReadFile("testdata/gitlab-project-success.json")
	Ok(t, err)

	cases := []struct {
		description string
		glResponse  []byte
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
			[]byte(`{"message":"405 Method Not Allowed"}`),
			405,
			"405 {message: 405 Method Not Allowed}",
		},
		{
			"406",
			[]byte(`{"message":"406 Branch cannot be merged"}`),
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
						w.Write(c.glResponse) // nolint: errcheck
					case "/api/v4/projects/runatlantis%2Fatlantis/merge_requests/1":
						w.WriteHeader(http.StatusOK)
						w.Write(pipelineSuccess) // nolint: errcheck
					case "/api/v4/projects/4580910":
						w.WriteHeader(http.StatusOK)
						w.Write(projectSuccess) // nolint: errcheck
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
				logger:  logging.NewNoopLogger(t),
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
	pipelineSuccess, err := os.ReadFile("testdata/gitlab-pipeline-success.json")
	Ok(t, err)

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
						exp := fmt.Sprintf(`{"state":"%s","ref":"patch-1-merger","context":"src","target_url":"https://google.com","description":"description"}`, c.expState)
						Equals(t, exp, string(body))
						defer r.Body.Close()  // nolint: errcheck
						w.Write([]byte("{}")) // nolint: errcheck
					case "/api/v4/projects/runatlantis%2Fatlantis/merge_requests/1":
						w.WriteHeader(http.StatusOK)
						w.Write(pipelineSuccess) // nolint: errcheck
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
				logger:  logging.NewNoopLogger(t),
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
	defaultMr := 1
	noHeadPipelineMR := 2
	ciMustPassSuccessMR := 3
	ciMustPassFailureMR := 4

	pipelineSuccess, err := os.ReadFile("testdata/gitlab-pipeline-success.json")
	Ok(t, err)

	projectSuccess, err := os.ReadFile("testdata/gitlab-project-success.json")
	Ok(t, err)

	detailedMergeStatusCiMustPass, err := os.ReadFile("testdata/gitlab-detailed-merge-status-ci-must-pass.json")
	Ok(t, err)

	headPipelineNotAvailable, err := os.ReadFile("testdata/gitlab-head-pipeline-not-available.json")
	Ok(t, err)

	cases := []struct {
		statusName    string
		status        models.CommitStatus
		gitlabVersion []string
		mrID          int
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
			fmt.Sprintf("%s/apply", vcsStatusName),
			models.FailedCommitStatus,
			gitlabServerVersions,
			ciMustPassSuccessMR,
			true,
		},
		{
			fmt.Sprintf("%s/plan", vcsStatusName),
			models.FailedCommitStatus,
			gitlabServerVersions,
			ciMustPassFailureMR,
			false,
		},
		{
			fmt.Sprintf("%s/apply: resource/default", vcsStatusName),
			models.FailedCommitStatus,
			gitlabServerVersions,
			noHeadPipelineMR,
			true,
		},
		{
			fmt.Sprintf("%s/apply", vcsStatusName),
			models.FailedCommitStatus,
			gitlabServerVersions,
			noHeadPipelineMR,
			true,
		},
		{
			fmt.Sprintf("%s/plan: resource/default", vcsStatusName),
			models.FailedCommitStatus,
			gitlabServerVersions,
			noHeadPipelineMR,
			false,
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
						case fmt.Sprintf("/api/v4/projects/runatlantis%%2Fatlantis/merge_requests/%v", defaultMr):
							w.WriteHeader(http.StatusOK)
							w.Write(pipelineSuccess) // nolint: errcheck
						case fmt.Sprintf("/api/v4/projects/runatlantis%%2Fatlantis/merge_requests/%v", noHeadPipelineMR):
							w.WriteHeader(http.StatusOK)
							w.Write(headPipelineNotAvailable) // nolint: errcheck
						case fmt.Sprintf("/api/v4/projects/runatlantis%%2Fatlantis/merge_requests/%v", ciMustPassSuccessMR):
							w.WriteHeader(http.StatusOK)
							w.Write(detailedMergeStatusCiMustPass) // nolint: errcheck
						case fmt.Sprintf("/api/v4/projects/runatlantis%%2Fatlantis/merge_requests/%v", ciMustPassFailureMR):
							w.WriteHeader(http.StatusOK)
							w.Write(detailedMergeStatusCiMustPass) // nolint: errcheck
						case fmt.Sprintf("/api/v4/projects/%v", projectID):
							w.WriteHeader(http.StatusOK)
							w.Write(projectSuccess) // nolint: errcheck
						case fmt.Sprintf("/api/v4/projects/%v/repository/commits/67cb91d3f6198189f433c045154a885784ba6977/statuses", projectID):
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
							err := json.NewEncoder(w).Encode(v)
							Ok(t, err)
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

				mergeable, err := client.PullIsMergeable(repo, models.PullRequest{
					Num:        c.mrID,
					BaseRepo:   repo,
					HeadCommit: "67cb91d3f6198189f433c045154a885784ba6977",
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
	planComments := [3]string{"Ran Plan for 2 projects:", "Ran Plan for dir: `stack1` workspace: `default`", "Ran Plan for 2 projects:"}
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

	repo := models.Repo{
		FullName: "runatlantis/atlantis",
		Owner:    "runatlantis",
		Name:     "atlantis",
		VCSHost: models.VCSHost{
			Type:     models.Gitlab,
			Hostname: "gitlab.com",
		},
	}

	cases := []struct {
		dir                  string
		processedComments    int
		processedCommentIds  []string
		processedPlanComment []string
	}{
		{
			"",
			2,
			[]string{planCommentIDs[0], planCommentIDs[1]},
			[]string{planComments[0], planComments[1]},
		},
		{
			"stack1",
			1,
			[]string{planCommentIDs[1]},
			[]string{planComments[1]},
		},
		{
			"stack2",
			0,
			[]string{},
			[]string{},
		},
	}

	for _, c := range cases {
		t.Run(c.dir, func(t *testing.T) {
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
							json.NewDecoder(r.Body).Decode(&body) // nolint: errcheck
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

			err = client.HidePrevCommandComments(repo, pullNum, command.Plan.TitleString(), c.dir)
			Ok(t, err)

			// Check the correct number of plan comments have been processed
			Equals(t, c.processedComments, len(gotNotePutCalls))
			// Check the correct comments have been processed
			for i := 0; i < c.processedComments; i++ {
				Equals(t, c.processedCommentIds[i], gotNotePutCalls[i].noteID)
				Equals(t, summaryHeader, gotNotePutCalls[i].comment[0])
				Equals(t, c.processedPlanComment[i], gotNotePutCalls[i].comment[1])
				Equals(t, summaryFooter, gotNotePutCalls[i].comment[2])
			}
		})
	}
}

func TestGithubClient_GetPullLabels(t *testing.T) {
	mergeSuccessWithLabel, err := os.ReadFile("testdata/gitlab-merge-success-with-label.json")
	Ok(t, err)

	testServer := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.RequestURI {
			case "/api/v4/projects/runatlantis%2Fatlantis/merge_requests/1":
				w.WriteHeader(http.StatusOK)
				w.Write(mergeSuccessWithLabel) // nolint: errcheck
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
		logger:  logging.NewNoopLogger(t),
	}

	labels, err := client.GetPullLabels(models.Repo{
		FullName: "runatlantis/atlantis",
	}, models.PullRequest{
		Num: 1,
	})
	Ok(t, err)
	Equals(t, []string{"work in progress"}, labels)
}

func TestGithubClient_GetPullLabels_EmptyResponse(t *testing.T) {
	pipelineSuccess, err := os.ReadFile("testdata/gitlab-pipeline-success.json")
	Ok(t, err)

	testServer := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.RequestURI {
			case "/api/v4/projects/runatlantis%2Fatlantis/merge_requests/1":
				w.WriteHeader(http.StatusOK)
				w.Write(pipelineSuccess) // nolint: errcheck
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
		logger:  logging.NewNoopLogger(t),
	}

	labels, err := client.GetPullLabels(models.Repo{
		FullName: "runatlantis/atlantis",
	}, models.PullRequest{
		Num: 1,
	})
	Ok(t, err)
	Equals(t, 0, len(labels))
}
