package gitlab

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	ajk = BasicUser{
		ID:        3614858,
		Name:      "Alex Kalderimis",
		Username:  "alexkalderimis",
		State:     "active",
		AvatarURL: "https://assets.gitlab-static.net/uploads/-/system/user/avatar/3614858/avatar.png",
		WebURL:    "https://gitlab.com/alexkalderimis",
	}
	tk = BasicUser{
		ID:        2535118,
		Name:      "Thong Kuah",
		Username:  "tkuah",
		State:     "active",
		AvatarURL: "https://secure.gravatar.com/avatar/f7b51bdd49a4914d29504d7ff4c3f7b9?s=80&d=identicon",
		WebURL:    "https://gitlab.com/tkuah",
	}
	getOpts = GetMergeRequestsOptions{}
	labels  = Labels{
		"GitLab Enterprise Edition",
		"backend",
		"database",
		"database::reviewed",
		"design management",
		"feature",
		"frontend",
		"group::knowledge",
		"missed:12.1",
	}
	pipelineCreation = time.Date(2019, 8, 19, 9, 50, 58, 157000000, time.UTC)
	pipelineUpdate   = time.Date(2019, 8, 19, 19, 22, 29, 647000000, time.UTC)
	pipelineBasic    = PipelineInfo{
		ID:        77056819,
		SHA:       "8e0b45049b6253b8984cde9241830d2851168142",
		Ref:       "delete-designs-v2",
		Status:    "success",
		WebURL:    "https://gitlab.com/gitlab-org/gitlab-ee/pipelines/77056819",
		CreatedAt: &pipelineCreation,
		UpdatedAt: &pipelineUpdate,
	}
	pipelineStarted  = time.Date(2019, 8, 19, 9, 51, 6, 545000000, time.UTC)
	pipelineFinished = time.Date(2019, 8, 19, 19, 22, 29, 632000000, time.UTC)
	pipelineDetailed = Pipeline{
		ID:         77056819,
		SHA:        "8e0b45049b6253b8984cde9241830d2851168142",
		Ref:        "delete-designs-v2",
		Status:     "success",
		WebURL:     "https://gitlab.com/gitlab-org/gitlab-ee/pipelines/77056819",
		BeforeSHA:  "3fe568caacb261b63090886f5b879ca0d9c6f4c3",
		Tag:        false,
		User:       &ajk,
		CreatedAt:  &pipelineCreation,
		UpdatedAt:  &pipelineUpdate,
		StartedAt:  &pipelineStarted,
		FinishedAt: &pipelineFinished,
		Duration:   4916,
		Coverage:   "82.68",
		DetailedStatus: &DetailedStatus{
			Icon:        "status_warning",
			Text:        "passed",
			Label:       "passed with warnings",
			Group:       "success-with-warnings",
			Tooltip:     "passed",
			HasDetails:  true,
			DetailsPath: "/gitlab-org/gitlab-ee/pipelines/77056819",
			Favicon:     "https://gitlab.com/assets/ci_favicons/favicon_status_success-8451333011eee8ce9f2ab25dc487fe24a8758c694827a582f17f42b0a90446a2.png",
		},
	}
)

func TestGetMergeRequest(t *testing.T) {
	mux, server, client := setup()
	defer teardown(server)

	path := "/api/v4/projects/namespace/name/merge_requests/123"

	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		mustWriteHTTPResponse(t, w, "testdata/get_merge_request.json")
	})

	mergeRequest, _, err := client.MergeRequests.GetMergeRequest("namespace/name", 123, &getOpts)

	require.NoError(t, err)

	require.Equal(t, mergeRequest.ID, 33092005)
	require.Equal(t, mergeRequest.SHA, "8e0b45049b6253b8984cde9241830d2851168142")
	require.Equal(t, mergeRequest.IID, 14656)
	require.Equal(t, mergeRequest.Reference, "!14656")
	require.Equal(t, mergeRequest.ProjectID, 278964)
	require.Equal(t, mergeRequest.SourceBranch, "delete-designs-v2")
	require.Equal(t, mergeRequest.TaskCompletionStatus.Count, 9)
	require.Equal(t, mergeRequest.TaskCompletionStatus.CompletedCount, 8)
	require.Equal(t, mergeRequest.Title, "Add deletion support for designs")
	require.Equal(t, mergeRequest.Description,
		"## What does this MR do?\r\n\r\nThis adds the capability to destroy/hide designs.")
	require.Equal(t, mergeRequest.WebURL,
		"https://gitlab.com/gitlab-org/gitlab-ee/merge_requests/14656")
	require.Equal(t, mergeRequest.MergeStatus, "can_be_merged")
	require.Equal(t, mergeRequest.Author, &ajk)
	require.Equal(t, mergeRequest.Assignee, &tk)
	require.Equal(t, mergeRequest.Assignees, []*BasicUser{&tk})
	require.Equal(t, mergeRequest.Labels, labels)
	require.Equal(t, mergeRequest.Squash, true)
	require.Equal(t, mergeRequest.UserNotesCount, 245)
	require.Equal(t, mergeRequest.Pipeline, &pipelineBasic)
	require.Equal(t, mergeRequest.HeadPipeline, &pipelineDetailed)
	mrCreation := time.Date(2019, 7, 11, 22, 34, 43, 500000000, time.UTC)
	require.Equal(t, mergeRequest.CreatedAt, &mrCreation)
	mrUpdate := time.Date(2019, 8, 20, 9, 9, 56, 690000000, time.UTC)
	require.Equal(t, mergeRequest.UpdatedAt, &mrUpdate)
	require.Equal(t, mergeRequest.HasConflicts, true)
}

func TestListProjectMergeRequests(t *testing.T) {
	mux, server, client := setup()
	defer teardown(server)

	path := "/api/v4/projects/278964/merge_requests"

	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		mustWriteHTTPResponse(t, w, "testdata/get_merge_requests.json")
	})
	opts := ListProjectMergeRequestsOptions{}

	mergeRequests, _, err := client.MergeRequests.ListProjectMergeRequests(278964, &opts)

	require.NoError(t, err)
	require.Equal(t, 20, len(mergeRequests))

	validStates := []string{"opened", "closed", "locked", "merged"}
	mergeStatuses := []string{"can_be_merged", "cannot_be_merged"}
	allCreatedBefore := time.Date(2019, 8, 21, 0, 0, 0, 0, time.UTC)
	allCreatedAfter := time.Date(2019, 8, 17, 0, 0, 0, 0, time.UTC)

	for _, mr := range mergeRequests {
		require.Equal(t, 278964, mr.ProjectID)
		require.Contains(t, validStates, mr.State)
		assert.Less(t, mr.CreatedAt.Unix(), allCreatedBefore.Unix())
		assert.Greater(t, mr.CreatedAt.Unix(), allCreatedAfter.Unix())
		assert.LessOrEqual(t, mr.CreatedAt.Unix(), mr.UpdatedAt.Unix())
		assert.LessOrEqual(t, mr.TaskCompletionStatus.CompletedCount, mr.TaskCompletionStatus.Count)
		require.Contains(t, mergeStatuses, mr.MergeStatus)
		// list requests do not provide these fields:
		assert.Nil(t, mr.Pipeline)
		assert.Nil(t, mr.HeadPipeline)
		assert.Equal(t, "", mr.DiffRefs.HeadSha)
	}
}
