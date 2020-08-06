package gitlab

import (
	"fmt"
	"net/http"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var testRevertCommitTargetBranch = "release"

func TestGetCommit(t *testing.T) {
	mux, server, client := setup()
	defer teardown(server)

	mux.HandleFunc("/api/v4/projects/1/repository/commits/b0b3a907f41409829b307a28b82fdbd552ee5a27", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		mustWriteHTTPResponse(t, w, "testdata/get_commit.json")
	})

	commit, resp, err := client.Commits.GetCommit("1", "b0b3a907f41409829b307a28b82fdbd552ee5a27", nil)
	if err != nil {
		t.Fatalf("Commits.GetCommit returned error: %v, response: %v", err, resp)
	}

	updatedAt := time.Date(2019, 11, 4, 15, 39, 03, 935000000, time.UTC)
	createdAt := time.Date(2019, 11, 4, 15, 38, 53, 154000000, time.UTC)
	want := &Commit{
		ID:             "6104942438c14ec7bd21c6cd5bd995272b3faff6",
		ShortID:        "6104942438c",
		Title:          "Sanitize for network graph",
		AuthorName:     "randx",
		AuthorEmail:    "dmitriy.zaporozhets@gmail.com",
		CommitterName:  "Dmitriy",
		CommitterEmail: "dmitriy.zaporozhets@gmail.com",
		Message:        "Sanitize for network graph",
		ParentIDs:      []string{"ae1d9fb46aa2b07ee9836d49862ec4e2c46fbbba"},
		Stats:          &CommitStats{Additions: 15, Deletions: 10, Total: 25},
		Status:         BuildState(Running),
		LastPipeline: &PipelineInfo{
			ID:        8,
			Ref:       "master",
			SHA:       "2dc6aa325a317eda67812f05600bdf0fcdc70ab0",
			Status:    "created",
			WebURL:    "https://gitlab.com/gitlab-org/gitlab-ce/pipelines/54268416",
			UpdatedAt: &updatedAt,
			CreatedAt: &createdAt,
		},
		ProjectID: 13083,
	}

	assert.Equal(t, want, commit)
}

func TestGetCommitStatuses(t *testing.T) {
	mux, server, client := setup()
	defer teardown(server)

	mux.HandleFunc("/api/v4/projects/1/repository/commits/b0b3a907f41409829b307a28b82fdbd552ee5a27/statuses", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		fmt.Fprint(w, `[{"id":1}]`)
	})

	opt := &GetCommitStatusesOptions{Ref: String("master"), Stage: String("test"), Name: String("ci/jenkins"), All: Bool(true)}
	statuses, _, err := client.Commits.GetCommitStatuses("1", "b0b3a907f41409829b307a28b82fdbd552ee5a27", opt)

	if err != nil {
		t.Errorf("Commits.GetCommitStatuses returned error: %v", err)
	}

	want := []*CommitStatus{{ID: 1}}
	if !reflect.DeepEqual(want, statuses) {
		t.Errorf("Commits.GetCommitStatuses returned %+v, want %+v", statuses, want)
	}
}

func TestSetCommitStatus(t *testing.T) {
	mux, server, client := setup()
	defer teardown(server)

	mux.HandleFunc("/api/v4/projects/1/statuses/b0b3a907f41409829b307a28b82fdbd552ee5a27", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "POST")
		fmt.Fprint(w, `{"id":1}`)
	})

	opt := &SetCommitStatusOptions{State: Running, Ref: String("master"), Name: String("ci/jenkins"), Context: String(""), TargetURL: String("http://abc"), Description: String("build")}
	status, _, err := client.Commits.SetCommitStatus("1", "b0b3a907f41409829b307a28b82fdbd552ee5a27", opt)

	if err != nil {
		t.Errorf("Commits.SetCommitStatus returned error: %v", err)
	}

	want := &CommitStatus{ID: 1}
	if !reflect.DeepEqual(want, status) {
		t.Errorf("Commits.SetCommitStatus returned %+v, want %+v", status, want)
	}
}

func TestRevertCommit_NoOptions(t *testing.T) {
	mux, server, client := setup()
	defer teardown(server)

	mux.HandleFunc("/api/v4/projects/1/repository/commits/b0b3a907f41409829b307a28b82fdbd552ee5a27/revert", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "POST")
		mustWriteHTTPResponse(t, w, "testdata/get_commit.json")
	})

	commit, resp, err := client.Commits.RevertCommit("1", "b0b3a907f41409829b307a28b82fdbd552ee5a27", nil)
	if err != nil {
		t.Fatalf("Commits.RevertCommit returned error: %v, response: %v", err, resp)
	}

	updatedAt := time.Date(2019, 11, 4, 15, 39, 03, 935000000, time.UTC)
	createdAt := time.Date(2019, 11, 4, 15, 38, 53, 154000000, time.UTC)
	want := &Commit{
		ID:             "6104942438c14ec7bd21c6cd5bd995272b3faff6",
		ShortID:        "6104942438c",
		Title:          "Sanitize for network graph",
		AuthorName:     "randx",
		AuthorEmail:    "dmitriy.zaporozhets@gmail.com",
		CommitterName:  "Dmitriy",
		CommitterEmail: "dmitriy.zaporozhets@gmail.com",
		Message:        "Sanitize for network graph",
		ParentIDs:      []string{"ae1d9fb46aa2b07ee9836d49862ec4e2c46fbbba"},
		Stats:          &CommitStats{Additions: 15, Deletions: 10, Total: 25},
		Status:         BuildState(Running),
		LastPipeline: &PipelineInfo{
			ID:        8,
			Ref:       "master",
			SHA:       "2dc6aa325a317eda67812f05600bdf0fcdc70ab0",
			Status:    "created",
			WebURL:    "https://gitlab.com/gitlab-org/gitlab-ce/pipelines/54268416",
			UpdatedAt: &updatedAt,
			CreatedAt: &createdAt,
		},
		ProjectID: 13083,
	}

	assert.Equal(t, want, commit)
}

func TestRevertCommit_WithOptions(t *testing.T) {
	mux, server, client := setup()
	defer teardown(server)

	mux.HandleFunc("/api/v4/projects/1/repository/commits/b0b3a907f41409829b307a28b82fdbd552ee5a27/revert", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "POST")
		testBody(t, r, `{"branch":"release"}`)
		mustWriteHTTPResponse(t, w, "testdata/get_commit.json")
	})

	commit, resp, err := client.Commits.RevertCommit("1", "b0b3a907f41409829b307a28b82fdbd552ee5a27", &RevertCommitOptions{
		Branch: &testRevertCommitTargetBranch,
	})
	if err != nil {
		t.Fatalf("Commits.RevertCommit returned error: %v, response: %v", err, resp)
	}

	updatedAt := time.Date(2019, 11, 4, 15, 39, 03, 935000000, time.UTC)
	createdAt := time.Date(2019, 11, 4, 15, 38, 53, 154000000, time.UTC)
	want := &Commit{
		ID:             "6104942438c14ec7bd21c6cd5bd995272b3faff6",
		ShortID:        "6104942438c",
		Title:          "Sanitize for network graph",
		AuthorName:     "randx",
		AuthorEmail:    "dmitriy.zaporozhets@gmail.com",
		CommitterName:  "Dmitriy",
		CommitterEmail: "dmitriy.zaporozhets@gmail.com",
		Message:        "Sanitize for network graph",
		ParentIDs:      []string{"ae1d9fb46aa2b07ee9836d49862ec4e2c46fbbba"},
		Stats:          &CommitStats{Additions: 15, Deletions: 10, Total: 25},
		Status:         BuildState(Running),
		LastPipeline: &PipelineInfo{
			ID:        8,
			Ref:       "master",
			SHA:       "2dc6aa325a317eda67812f05600bdf0fcdc70ab0",
			Status:    "created",
			WebURL:    "https://gitlab.com/gitlab-org/gitlab-ce/pipelines/54268416",
			UpdatedAt: &updatedAt,
			CreatedAt: &createdAt,
		},
		ProjectID: 13083,
	}

	assert.Equal(t, want, commit)
}

func TestGetGPGSignature(t *testing.T) {
	mux, server, client := setup()
	defer teardown(server)

	mux.HandleFunc("/api/v4/projects/1/repository/commits/b0b3a907f41409829b307a28b82fdbd552ee5a27/signature", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		mustWriteHTTPResponse(t, w, "testdata/get_signature.json")
	})

	sig, resp, err := client.Commits.GetGPGSiganature("1", "b0b3a907f41409829b307a28b82fdbd552ee5a27", nil)
	if err != nil {
		t.Fatalf("Commits.GetGPGSignature returned error: %v, response: %v", err, resp)
	}

	want := &GPGSignature{
		KeyID:              7977,
		KeyPrimaryKeyID:    "627C5F589F467F17",
		KeyUserName:        "Dmitriy Zaporozhets",
		KeyUserEmail:       "dmitriy.zaporozhets@gmail.com",
		VerificationStatus: "verified",
		KeySubkeyID:        0,
	}

	assert.Equal(t, want, sig)
}
