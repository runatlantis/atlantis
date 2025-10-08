package vcs

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/go-version"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/logging"
	gitlab "gitlab.com/gitlab-org/api/client-go"

	. "github.com/runatlantis/atlantis/testing"
)

var projectID = 4580910

const gitlabPipelineSuccessMrID = 488598

const updateStatusDescription = "description"
const updateStatusTargetUrl = "https://google.com"
const updateStatusSrc = "src"
const updateStatusHeadBranch = "test"

/* UpdateStatus request JSON body object */
type UpdateStatusJsonBody struct {
	State       string `json:"state"`
	Context     string `json:"context"`
	TargetUrl   string `json:"target_url"`
	Description string `json:"description"`
	PipelineId  int    `json:"pipeline_id"`
	Ref         string `json:"ref"`
}

/* GetCommit response last_pipeline JSON object */
type GetCommitResponseLastPipeline struct {
	ID int `json:"id"`
}

/* GetCommit response JSON object */
type GetCommitResponse struct {
	LastPipeline GetCommitResponseLastPipeline `json:"last_pipeline"`
}

/* Empty struct for JSON marshalling */
type EmptyStruct struct{}

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
			client, err := NewGitlabClient(c.Hostname, "token", []string{}, log)
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
	logger := logging.NewNoopLogger(t)
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
			}

			filenames, err := client.GetModifiedFiles(
				logger,
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
	logger := logging.NewNoopLogger(t)
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
			}

			err = client.MergePull(
				logger,
				models.PullRequest{
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
	logger := logging.NewNoopLogger(t)

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

						var updateStatusJsonBody UpdateStatusJsonBody
						err := json.NewDecoder(r.Body).Decode(&updateStatusJsonBody)
						Ok(t, err)

						Equals(t, c.expState, updateStatusJsonBody.State)
						Equals(t, updateStatusSrc, updateStatusJsonBody.Context)
						Equals(t, updateStatusTargetUrl, updateStatusJsonBody.TargetUrl)
						Equals(t, updateStatusDescription, updateStatusJsonBody.Description)
						Equals(t, gitlabPipelineSuccessMrID, updateStatusJsonBody.PipelineId)

						defer r.Body.Close() // nolint: errcheck

						setStatusJsonResponse, err := json.Marshal(EmptyStruct{})
						Ok(t, err)

						_, err = w.Write(setStatusJsonResponse)
						Ok(t, err)

					case "/api/v4/projects/runatlantis%2Fatlantis/repository/commits/sha":
						w.WriteHeader(http.StatusOK)

						getCommitResponse := GetCommitResponse{
							LastPipeline: GetCommitResponseLastPipeline{
								ID: gitlabPipelineSuccessMrID,
							},
						}
						getCommitJsonResponse, err := json.Marshal(getCommitResponse)
						Ok(t, err)

						_, err = w.Write(getCommitJsonResponse)
						Ok(t, err)

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
			err = client.UpdateStatus(
				logger,
				repo,
				models.PullRequest{
					Num:        1,
					BaseRepo:   repo,
					HeadCommit: "sha",
					HeadBranch: updateStatusHeadBranch,
				},
				c.status,
				updateStatusSrc,
				updateStatusDescription,
				updateStatusTargetUrl,
			)
			Ok(t, err)
			Assert(t, gotRequest, "expected to get the request")
		})
	}
}

func TestGitlabClient_UpdateStatusGetCommitRetryable(t *testing.T) {
	logger := logging.NewNoopLogger(t)

	cases := []struct {
		title                     string
		status                    models.CommitStatus
		commitsWithNoLastPipeline int
		expNumberOfRequests       int
		expRefOrPipelineId        string
	}{
		// Ensure that GetCommit with last pipeline id sets the pipeline id.
		{
			title:                     "GetCommit with a pipeline id",
			status:                    models.PendingCommitStatus,
			commitsWithNoLastPipeline: 0,
			expNumberOfRequests:       1,
			expRefOrPipelineId:        "PipelineId",
		},
		// Ensure that 1 x GetCommit with no pipelines sets the pipeline id.
		{
			title:                     "1 x GetCommit with no last pipeline id",
			status:                    models.PendingCommitStatus,
			commitsWithNoLastPipeline: 1,
			expNumberOfRequests:       2,
			expRefOrPipelineId:        "PipelineId",
		},
		// Ensure that 2 x GetCommit with no last pipeline id sets the ref.
		{
			title:                     "2 x GetCommit with no last pipeline id",
			status:                    models.PendingCommitStatus,
			commitsWithNoLastPipeline: 2,
			expNumberOfRequests:       2,
			expRefOrPipelineId:        "Ref",
		},
	}

	for _, c := range cases {
		t.Run(c.title, func(t *testing.T) {
			handledNumberOfRequests := 0

			testServer := httptest.NewServer(
				http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					switch r.RequestURI {
					case "/api/v4/projects/runatlantis%2Fatlantis/statuses/sha":
						var updateStatusJsonBody UpdateStatusJsonBody
						err := json.NewDecoder(r.Body).Decode(&updateStatusJsonBody)
						Ok(t, err)

						Equals(t, "running", updateStatusJsonBody.State)
						Equals(t, updateStatusSrc, updateStatusJsonBody.Context)
						Equals(t, updateStatusTargetUrl, updateStatusJsonBody.TargetUrl)
						Equals(t, updateStatusDescription, updateStatusJsonBody.Description)
						if c.expRefOrPipelineId == "Ref" {
							Equals(t, updateStatusHeadBranch, updateStatusJsonBody.Ref)
						} else {
							Equals(t, gitlabPipelineSuccessMrID, updateStatusJsonBody.PipelineId)
						}

						defer r.Body.Close()

						getCommitJsonResponse, err := json.Marshal(EmptyStruct{})
						Ok(t, err)

						_, err = w.Write(getCommitJsonResponse)
						Ok(t, err)

					case "/api/v4/projects/runatlantis%2Fatlantis/repository/commits/sha":
						handledNumberOfRequests++
						noCommitLastPipeline := handledNumberOfRequests <= c.commitsWithNoLastPipeline

						w.WriteHeader(http.StatusOK)
						if noCommitLastPipeline {
							getCommitJsonResponse, err := json.Marshal(EmptyStruct{})
							Ok(t, err)

							_, err = w.Write(getCommitJsonResponse)
							Ok(t, err)
						} else {
							getCommitResponse := GetCommitResponse{
								LastPipeline: GetCommitResponseLastPipeline{
									ID: gitlabPipelineSuccessMrID,
								},
							}
							getCommitJsonResponse, err := json.Marshal(getCommitResponse)
							Ok(t, err)

							_, err = w.Write(getCommitJsonResponse)
							Ok(t, err)
						}

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
				Client:          internalClient,
				Version:         nil,
				PollingInterval: 10 * time.Millisecond,
			}

			repo := models.Repo{
				FullName: "runatlantis/atlantis",
				Owner:    "runatlantis",
				Name:     "atlantis",
			}

			err = client.UpdateStatus(
				logger,
				repo,
				models.PullRequest{
					Num:        1,
					BaseRepo:   repo,
					HeadCommit: "sha",
					HeadBranch: updateStatusHeadBranch,
				},
				c.status,
				updateStatusSrc,
				updateStatusDescription,
				updateStatusTargetUrl,
			)
			Ok(t, err)

			Assert(t, c.expNumberOfRequests == handledNumberOfRequests,
				fmt.Sprintf("expected %d number of requests, but processed %d", c.expNumberOfRequests, handledNumberOfRequests))
		})
	}
}

func TestGitlabClient_UpdateStatusSetCommitStatusConflictRetryable(t *testing.T) {
	logger := logging.NewNoopLogger(t)

	cases := []struct {
		status              models.CommitStatus
		numberOfConflicts   int
		expNumberOfRequests int
		expState            string
		expError            bool
	}{
		// Ensure that 0 x 409 Conflict succeeds
		{
			status:              models.PendingCommitStatus,
			numberOfConflicts:   0,
			expNumberOfRequests: 1,
			expState:            "running",
		},
		// Ensure that 5 x 409 Conflict still succeeds
		{
			status:              models.PendingCommitStatus,
			numberOfConflicts:   5,
			expNumberOfRequests: 6,
			expState:            "running",
		},
		// Ensure that 10 x 409 Conflict still fail due to running out of retries
		{
			status:              models.FailedCommitStatus,
			numberOfConflicts:   100, // anything larger than 10 is fine
			expNumberOfRequests: 10,
			expState:            "failed",
			expError:            true,
		},
	}
	for _, c := range cases {
		t.Run(c.expState, func(t *testing.T) {
			handledNumberOfRequests := 0

			testServer := httptest.NewServer(
				http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					switch r.RequestURI {
					case "/api/v4/projects/runatlantis%2Fatlantis/statuses/sha":
						handledNumberOfRequests++
						shouldSendConflict := handledNumberOfRequests <= c.numberOfConflicts

						var updateStatusJsonBody UpdateStatusJsonBody
						err := json.NewDecoder(r.Body).Decode(&updateStatusJsonBody)
						Ok(t, err)

						Equals(t, c.expState, updateStatusJsonBody.State)
						Equals(t, updateStatusSrc, updateStatusJsonBody.Context)
						Equals(t, updateStatusTargetUrl, updateStatusJsonBody.TargetUrl)
						Equals(t, updateStatusDescription, updateStatusJsonBody.Description)

						defer r.Body.Close() // nolint: errcheck

						if shouldSendConflict {
							w.WriteHeader(http.StatusConflict)
						}

						getCommitJsonResponse, err := json.Marshal(EmptyStruct{})
						Ok(t, err)

						_, err = w.Write(getCommitJsonResponse)
						Ok(t, err)

					case "/api/v4/projects/runatlantis%2Fatlantis/repository/commits/sha":
						w.WriteHeader(http.StatusOK)

						getCommitResponse := GetCommitResponse{
							LastPipeline: GetCommitResponseLastPipeline{
								ID: gitlabPipelineSuccessMrID,
							},
						}
						getCommitJsonResponse, err := json.Marshal(getCommitResponse)
						Ok(t, err)

						_, err = w.Write(getCommitJsonResponse)
						Ok(t, err)

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
				Client:          internalClient,
				Version:         nil,
				PollingInterval: 10 * time.Millisecond,
			}

			repo := models.Repo{
				FullName: "runatlantis/atlantis",
				Owner:    "runatlantis",
				Name:     "atlantis",
			}
			err = client.UpdateStatus(
				logger,
				repo,
				models.PullRequest{
					Num:        1,
					BaseRepo:   repo,
					HeadCommit: "sha",
					HeadBranch: "test",
				},
				c.status,
				updateStatusSrc,
				updateStatusDescription,
				updateStatusTargetUrl,
			)

			if c.expError {
				ErrContains(t, "failed to update commit status for 'runatlantis/atlantis' @ 'sha' to 'src' after 10 attempts", err)
				ErrContains(t, "409", err)
			} else {
				Ok(t, err)
			}

			Assert(t, c.expNumberOfRequests == handledNumberOfRequests,
				fmt.Sprintf("expected %d number of requests, but processed %d", c.expNumberOfRequests, handledNumberOfRequests))
		})
	}
}

func mustReadFile(t *testing.T, filename string) []byte {
	ret, err := os.ReadFile(filename)
	Ok(t, err)
	return ret
}

func TestGitlabClient_PullIsMergeable(t *testing.T) {
	logger := logging.NewNoopLogger(t)
	gitlabClientUnderTest = true
	gitlabVersionOver15_6 := "15.8.3-ee"
	gitlabVersion15_6 := "15.6.0-ee"
	gitlabVersionUnder15_6 := "15.3.2-ce"
	gitlabServerVersions := []string{gitlabVersionOver15_6, gitlabVersion15_6, gitlabVersionUnder15_6}
	vcsStatusName := "atlantis-test"
	defaultMr := 1
	noHeadPipelineMR := 2
	ciMustPassMR := 3
	needRebaseMR := 4
	remainingApprovalsMR := 5
	blockingDiscussionsUnresolvedMR := 6
	workInProgressMR := 7
	pipelineSkippedMR := 8

	// Any IsMergeable logic that depends on data from the project itself is too difficult to test here.
	// See TestGitlabClient_gitlabPullIsMergeable

	projectSuccess, err := os.ReadFile("testdata/gitlab-project-success.json")
	Ok(t, err)

	mrs := map[int][]byte{
		defaultMr:                       mustReadFile(t, "testdata/gitlab-pipeline-success.json"),
		noHeadPipelineMR:                mustReadFile(t, "testdata/gitlab-head-pipeline-not-available.json"),
		ciMustPassMR:                    mustReadFile(t, "testdata/gitlab-detailed-merge-status-ci-must-pass.json"),
		needRebaseMR:                    mustReadFile(t, "testdata/gitlab-detailed-merge-status-need-rebase.json"),
		remainingApprovalsMR:            mustReadFile(t, "testdata/gitlab-pipeline-remaining-approvals.json"),
		blockingDiscussionsUnresolvedMR: mustReadFile(t, "testdata/gitlab-pipeline-blocking-discussions-unresolved.json"),
		workInProgressMR:                mustReadFile(t, "testdata/gitlab-pipeline-work-in-progress.json"),
		pipelineSkippedMR:               mustReadFile(t, "testdata/gitlab-pipeline-with-pipeline-skipped.json"),
	}

	cases := []struct {
		statusName    string
		status        models.CommitStatus
		gitlabVersion []string
		mrID          int
		expState      models.MergeableStatus
	}{
		{
			fmt.Sprintf("%s/apply: resource/default", vcsStatusName),
			models.FailedCommitStatus,
			gitlabServerVersions,
			defaultMr,
			models.MergeableStatus{
				IsMergeable: true,
			},
		},
		{
			fmt.Sprintf("%s/apply", vcsStatusName),
			models.FailedCommitStatus,
			gitlabServerVersions,
			defaultMr,
			models.MergeableStatus{
				IsMergeable: true,
			},
		},
		{
			fmt.Sprintf("%s/plan: resource/default", vcsStatusName),
			models.FailedCommitStatus,
			gitlabServerVersions,
			defaultMr,
			models.MergeableStatus{
				IsMergeable: false,
				Reason:      fmt.Sprintf("Pipeline %s/plan: resource/default has status failed", vcsStatusName),
			},
		},
		{
			fmt.Sprintf("%s/plan", vcsStatusName),
			models.PendingCommitStatus,
			gitlabServerVersions,
			defaultMr,
			models.MergeableStatus{
				IsMergeable: false,
				Reason:      fmt.Sprintf("Pipeline %s/plan has status pending", vcsStatusName),
			},
		},
		{
			fmt.Sprintf("%s/plan", vcsStatusName),
			models.SuccessCommitStatus,
			gitlabServerVersions,
			defaultMr,
			models.MergeableStatus{
				IsMergeable: true,
			},
		},
		{
			fmt.Sprintf("%s/apply", vcsStatusName),
			models.FailedCommitStatus,
			gitlabServerVersions,
			ciMustPassMR,
			models.MergeableStatus{
				IsMergeable: true,
			},
		},
		{
			fmt.Sprintf("%s/plan", vcsStatusName),
			models.FailedCommitStatus,
			gitlabServerVersions,
			ciMustPassMR,
			models.MergeableStatus{
				IsMergeable: false,
				Reason:      fmt.Sprintf("Pipeline %s/plan has status failed", vcsStatusName),
			},
		},
		{
			fmt.Sprintf("%s/apply", vcsStatusName),
			models.FailedCommitStatus,
			gitlabServerVersions,
			needRebaseMR,
			models.MergeableStatus{
				IsMergeable: true,
			},
		},
		{
			fmt.Sprintf("%s/apply: resource/default", vcsStatusName),
			models.FailedCommitStatus,
			gitlabServerVersions,
			noHeadPipelineMR,
			models.MergeableStatus{
				IsMergeable: true,
			},
		},
		{
			fmt.Sprintf("%s/apply", vcsStatusName),
			models.FailedCommitStatus,
			gitlabServerVersions,
			noHeadPipelineMR,
			models.MergeableStatus{
				IsMergeable: true,
			},
		},
		{
			fmt.Sprintf("%s/plan: resource/default", vcsStatusName),
			models.FailedCommitStatus,
			gitlabServerVersions,
			noHeadPipelineMR,
			models.MergeableStatus{
				IsMergeable: false,
				Reason:      fmt.Sprintf("Pipeline %s/plan: resource/default has status failed", vcsStatusName),
			},
		},
		{
			fmt.Sprintf("%s/plan", vcsStatusName),
			models.PendingCommitStatus,
			gitlabServerVersions,
			noHeadPipelineMR,
			models.MergeableStatus{
				IsMergeable: false,
				Reason:      fmt.Sprintf("Pipeline %s/plan has status pending", vcsStatusName),
			},
		},
		{
			fmt.Sprintf("%s/plan", vcsStatusName),
			models.FailedCommitStatus,
			gitlabServerVersions,
			noHeadPipelineMR,
			models.MergeableStatus{
				IsMergeable: false,
				Reason:      fmt.Sprintf("Pipeline %s/plan has status failed", vcsStatusName),
			},
		},
		{
			fmt.Sprintf("%s/plan", vcsStatusName),
			models.SuccessCommitStatus,
			gitlabServerVersions,
			noHeadPipelineMR,
			models.MergeableStatus{
				IsMergeable: true,
			},
		},
		{
			fmt.Sprintf("%s/plan", vcsStatusName),
			models.SuccessCommitStatus,
			gitlabServerVersions,
			remainingApprovalsMR,
			models.MergeableStatus{
				IsMergeable: false,
				Reason:      "Still require 2 approvals",
			},
		},
		{
			fmt.Sprintf("%s/plan", vcsStatusName),
			models.SuccessCommitStatus,
			gitlabServerVersions,
			blockingDiscussionsUnresolvedMR,
			models.MergeableStatus{
				IsMergeable: false,
				Reason:      "Blocking discussions unresolved",
			},
		},
		{
			fmt.Sprintf("%s/plan", vcsStatusName),
			models.SuccessCommitStatus,
			gitlabServerVersions,
			workInProgressMR,
			models.MergeableStatus{
				IsMergeable: false,
				Reason:      "Work in progress",
			},
		},
		{
			fmt.Sprintf("%s/plan", vcsStatusName),
			models.SuccessCommitStatus,
			gitlabServerVersions,
			pipelineSkippedMR,
			models.MergeableStatus{
				IsMergeable: false,
				Reason:      "Pipeline was skipped",
			},
		},
	}
	for _, serverVersion := range gitlabServerVersions {
		for _, c := range cases {
			t.Run(c.statusName, func(t *testing.T) {
				testServer := httptest.NewServer(
					http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						switch {
						case r.RequestURI == "/api/v4/":
							// Rate limiter requests.
							w.WriteHeader(http.StatusOK)

						case strings.HasPrefix(r.RequestURI, "/api/v4/projects/runatlantis%2Fatlantis/merge_requests/"):
							// Extract merge request ID
							mrPart := strings.TrimPrefix(r.RequestURI, "/api/v4/projects/runatlantis%2Fatlantis/merge_requests/")
							mrID, err := strconv.Atoi(mrPart)
							if err != nil {
								t.Errorf("invalid MR id in URI %q", r.RequestURI)
								http.Error(w, "bad request", http.StatusBadRequest)
								return
							}
							response, ok := mrs[mrID]
							if !ok {
								t.Errorf("invalid MR id %d", mrID)
								http.Error(w, "not found", http.StatusNotFound)
								return
							}

							w.WriteHeader(http.StatusOK)
							w.Write(response) // nolint: errcheck

						case r.RequestURI == fmt.Sprintf("/api/v4/projects/%v", projectID):
							w.WriteHeader(http.StatusOK)
							w.Write(projectSuccess) // nolint: errcheck
						case r.RequestURI == fmt.Sprintf("/api/v4/projects/%v/repository/commits/67cb91d3f6198189f433c045154a885784ba6977/statuses", projectID):
							w.WriteHeader(http.StatusOK)
							response := fmt.Sprintf(`[{"id":133702594,"sha":"67cb91d3f6198189f433c045154a885784ba6977","ref":"patch-1","status":"%s","name":"%s","target_url":null,"description":"ApplySuccess","created_at":"2018-12-12T18:31:57.957Z","started_at":null,"finished_at":"2018-12-12T18:31:58.480Z","allow_failure":false,"coverage":null,"author":{"id":1755902,"username":"lkysow","name":"LukeKysow","state":"active","avatar_url":"https://secure.gravatar.com/avatar/25fd57e71590fe28736624ff24d41c5f?s=80&d=identicon","web_url":"https://gitlab.com/lkysow"}}]`, c.status, c.statusName)
							w.Write([]byte(response)) // nolint: errcheck
						case r.RequestURI == "/api/v4/version":
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

				mergeable, err := client.PullIsMergeable(
					logger,
					repo,
					models.PullRequest{
						Num:        c.mrID,
						BaseRepo:   repo,
						HeadCommit: "67cb91d3f6198189f433c045154a885784ba6977",
					}, vcsStatusName, []string{})

				Ok(t, err)
				Equals(t, c.expState, mergeable)
			})
		}
	}
}

func TestGitlabClient_gitlabIsMergeable(t *testing.T) {
	// Test the helper gitlabIsMergeable directly

	cases := []struct {
		description                 string
		mr                          *gitlab.MergeRequest
		project                     *gitlab.Project
		supportsDetailedMergeStatus bool
		expected                    models.MergeableStatus
	}{
		{
			description: "requires approvals",
			mr: &gitlab.MergeRequest{
				ApprovalsBeforeMerge: 2,
			},
			project: &gitlab.Project{},
			expected: models.MergeableStatus{
				IsMergeable: false,
				Reason:      "Still require 2 approvals",
			},
		},
		{
			description: "blocking discussions unresolved",
			mr: &gitlab.MergeRequest{
				BlockingDiscussionsResolved: false,
			},
			project: &gitlab.Project{},
			expected: models.MergeableStatus{
				IsMergeable: false,
				Reason:      "Blocking discussions unresolved",
			},
		},
		{
			description: "work in progress",
			mr: &gitlab.MergeRequest{
				BlockingDiscussionsResolved: true,
				WorkInProgress:              true,
			},
			project: &gitlab.Project{},
			expected: models.MergeableStatus{
				IsMergeable: false,
				Reason:      "Work in progress",
			},
		},
		{
			description: "pipeline skipped and not allowed",
			mr: &gitlab.MergeRequest{
				BlockingDiscussionsResolved: true,
				HeadPipeline:                &gitlab.Pipeline{Status: "skipped"},
			},
			project: &gitlab.Project{
				AllowMergeOnSkippedPipeline: false,
			},
			expected: models.MergeableStatus{
				IsMergeable: false,
				Reason:      "Pipeline was skipped",
			},
		},
		{
			description: "pipeline skipped and is allowed",
			mr: &gitlab.MergeRequest{
				BlockingDiscussionsResolved: true,
				HeadPipeline:                &gitlab.Pipeline{Status: "skipped"},
				DetailedMergeStatus:         "mergeable",
			},
			supportsDetailedMergeStatus: true,
			project: &gitlab.Project{
				AllowMergeOnSkippedPipeline: true,
			},
			expected: models.MergeableStatus{
				IsMergeable: true,
			},
		},
		{
			description: "detailed merge status mergeable",
			mr: &gitlab.MergeRequest{
				BlockingDiscussionsResolved: true,
				DetailedMergeStatus:         "mergeable",
			},
			project:                     &gitlab.Project{},
			supportsDetailedMergeStatus: true,
			expected:                    models.MergeableStatus{IsMergeable: true},
		},
		{
			description: "detailed merge status need_rebase",
			mr: &gitlab.MergeRequest{
				BlockingDiscussionsResolved: true,
				DetailedMergeStatus:         "need_rebase",
			},
			project:                     &gitlab.Project{},
			supportsDetailedMergeStatus: true,
			expected: models.MergeableStatus{
				IsMergeable: false,
				Reason:      "Merge status is need_rebase",
			},
		},
		{
			description: "detailed merge status not mergeable",
			mr: &gitlab.MergeRequest{
				BlockingDiscussionsResolved: true,
				DetailedMergeStatus:         "blocked",
			},
			project:                     &gitlab.Project{},
			supportsDetailedMergeStatus: true,
			expected: models.MergeableStatus{
				IsMergeable: false,
				Reason:      "Merge status is blocked",
			},
		},
		{
			description: "detailed merge status can_be_merged (not a valid detailed status)",
			mr: &gitlab.MergeRequest{
				BlockingDiscussionsResolved: true,
				DetailedMergeStatus:         "can_be_merged",
			},
			project:                     &gitlab.Project{},
			supportsDetailedMergeStatus: true,
			expected: models.MergeableStatus{
				IsMergeable: false,
				Reason:      "Merge status is can_be_merged",
			},
		},
		{
			description: "legacy merge status can_be_merged",
			mr: &gitlab.MergeRequest{
				BlockingDiscussionsResolved: true,
				MergeStatus:                 "can_be_merged",
			},
			project:  &gitlab.Project{},
			expected: models.MergeableStatus{IsMergeable: true},
		},
		{
			description: "legacy merge status cannot be merged",
			mr: &gitlab.MergeRequest{
				BlockingDiscussionsResolved: true,
				MergeStatus:                 "cannot_be_merged",
			},
			project: &gitlab.Project{},
			expected: models.MergeableStatus{
				IsMergeable: false,
				Reason:      "Merge status is cannot_be_merged",
			},
		},
	}

	for _, c := range cases {
		t.Run(c.description, func(t *testing.T) {
			actual := gitlabIsMergeable(c.mr, c.project, c.supportsDetailedMergeStatus)
			Equals(t, c.expected, actual)
		})
	}
}

func TestGitlabClient_MarkdownPullLink(t *testing.T) {
	logger := logging.NewNoopLogger(t)
	gitlabClientUnderTest = true
	defer func() { gitlabClientUnderTest = false }()
	client, err := NewGitlabClient("gitlab.com", "token", []string{}, logger)
	Ok(t, err)
	pull := models.PullRequest{Num: 1}
	s, _ := client.MarkdownPullLink(pull)
	exp := "!1"
	Equals(t, exp, s)
}

func TestGitlabClient_HideOldComments(t *testing.T) {
	logger := logging.NewNoopLogger(t)
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
			}

			err = client.HidePrevCommandComments(logger, repo, pullNum, command.Plan.TitleString(), c.dir)
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

func TestGitlabClient_GetPullLabels(t *testing.T) {
	logger := logging.NewNoopLogger(t)
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
	}

	labels, err := client.GetPullLabels(
		logger,
		models.Repo{
			FullName: "runatlantis/atlantis",
		},
		models.PullRequest{
			Num: 1,
		},
	)
	Ok(t, err)
	Equals(t, []string{"work in progress"}, labels)
}

func TestGitlabClient_GetPullLabels_EmptyResponse(t *testing.T) {
	logger := logging.NewNoopLogger(t)
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
	}

	labels, err := client.GetPullLabels(
		logger,
		models.Repo{
			FullName: "runatlantis/atlantis",
		}, models.PullRequest{
			Num: 1,
		})
	Ok(t, err)
	Equals(t, 0, len(labels))
}

// GetTeamNamesForUser returns the names of the GitLab groups that the user belongs to.
func TestGitlabClient_GetTeamNamesForUser(t *testing.T) {
	logger := logging.NewNoopLogger(t)

	groupMembershipSuccess, err := os.ReadFile("testdata/gitlab-group-membership-success.json")
	Ok(t, err)

	userSuccess, err := os.ReadFile("testdata/gitlab-user-success.json")
	Ok(t, err)

	userEmpty, err := os.ReadFile("testdata/gitlab-user-none.json")
	Ok(t, err)

	multipleUsers, err := os.ReadFile("testdata/gitlab-user-multiple.json")
	Ok(t, err)

	configuredGroups := []string{"someorg/group1", "someorg/group2", "someorg/group3", "someorg/group4"}

	cases := []struct {
		userName string
		expErr   string
		expTeams []string
	}{
		{
			userName: "testuser",
			expTeams: []string{"someorg/group1", "someorg/group2"},
		},
		{
			userName: "none",
			expErr:   "GET /users returned no user",
		},
		{
			userName: "multiuser",
			expErr:   "GET /users returned more than 1 user",
		},
	}
	for _, c := range cases {
		t.Run(c.userName, func(t *testing.T) {

			testServer := httptest.NewServer(
				http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					switch r.RequestURI {
					case "/api/v4/users?username=testuser":
						w.WriteHeader(http.StatusOK)
						w.Write(userSuccess) // nolint: errcheck
					case "/api/v4/users?username=none":
						w.WriteHeader(http.StatusOK)
						w.Write(userEmpty) // nolint: errcheck
					case "/api/v4/users?username=multiuser":
						w.WriteHeader(http.StatusOK)
						w.Write(multipleUsers) // nolint: errcheck
					case "/api/v4/groups/someorg%2Fgroup1/members/123", "/api/v4/groups/someorg%2Fgroup2/members/123":
						w.WriteHeader(http.StatusOK)
						w.Write(groupMembershipSuccess) // nolint: errcheck
					case "/api/v4/groups/someorg%2Fgroup3/members/123":
						http.Error(w, "forbidden", http.StatusForbidden)
					case "/api/v4/groups/someorg%2Fgroup4/members/123":
						http.Error(w, "not found", http.StatusNotFound)
					default:
						t.Errorf("got unexpected request at %q", r.RequestURI)
						http.Error(w, "not found", http.StatusNotFound)
					}
				}))
			internalClient, err := gitlab.NewClient("token", gitlab.WithBaseURL(testServer.URL))
			Ok(t, err)
			client := &GitlabClient{
				Client:           internalClient,
				Version:          nil,
				ConfiguredGroups: configuredGroups,
			}

			teams, err := client.GetTeamNamesForUser(
				logger,
				models.Repo{
					Owner: "someorg",
				}, models.User{
					Username: c.userName,
				})
			if c.expErr == "" {
				Ok(t, err)
				Equals(t, c.expTeams, teams)
			} else {
				ErrContains(t, c.expErr, err)

			}

		})
	}
}

func TestGithubClient_DiscardReviews(t *testing.T) {
	logger := logging.NewNoopLogger(t)
	cases := []struct {
		description   string
		repoFullName  string
		pullReqeustId int
		wantErr       bool
	}{
		{
			"success",
			"runatlantis/atlantis",
			42,
			false,
		},
		{
			"error",
			"runatlantis/atlantis",
			32,
			true,
		},
	}
	for _, c := range cases {
		t.Run(c.description, func(t *testing.T) {
			testServer := httptest.NewServer(
				http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					switch r.RequestURI {
					case "/api/v4/projects/runatlantis%2Fatlantis/merge_requests/42/reset_approvals":
						w.WriteHeader(http.StatusOK)
					case "/api/v4/projects/runatlantis%2Fatlantis/merge_requests/32/reset_approvals":
						http.Error(w, "No bot token", http.StatusUnauthorized)
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
				FullName: c.repoFullName,
			}

			pr := models.PullRequest{
				Num: c.pullReqeustId,
			}

			if err := client.DiscardReviews(logger, repo, pr); (err != nil) != c.wantErr {
				t.Errorf("DiscardReviews() error = %v", err)
			}
		})
	}
}

func TestGitlabClient_UpdateStatusTransitionAlreadyComplete(t *testing.T) {
	logger := logging.NewNoopLogger(t)

	testServer := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.RequestURI {
			case "/api/v4/projects/runatlantis%2Fatlantis/statuses/sha":
				w.WriteHeader(http.StatusBadRequest)
				_, err := w.Write([]byte(`{"message": {"state": ["Cannot transition status via :run from :running"]}}`))
				Ok(t, err)

			case "/api/v4/projects/runatlantis%2Fatlantis/repository/commits/sha":
				w.WriteHeader(http.StatusOK)

				getCommitResponse := GetCommitResponse{
					LastPipeline: GetCommitResponseLastPipeline{
						ID: gitlabPipelineSuccessMrID,
					},
				}
				getCommitJsonResponse, err := json.Marshal(getCommitResponse)
				Ok(t, err)

				_, err = w.Write(getCommitJsonResponse)
				Ok(t, err)

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
		Client:          internalClient,
		Version:         nil,
		PollingInterval: 10 * time.Millisecond,
	}

	repo := models.Repo{
		FullName: "runatlantis/atlantis",
		Owner:    "runatlantis",
		Name:     "atlantis",
	}
	err = client.UpdateStatus(
		logger,
		repo,
		models.PullRequest{
			Num:        1,
			BaseRepo:   repo,
			HeadCommit: "sha",
			HeadBranch: "test",
		},
		models.PendingCommitStatus,
		updateStatusSrc,
		updateStatusDescription,
		updateStatusTargetUrl,
	)

	Ok(t, err)
}
