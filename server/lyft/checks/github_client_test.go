package checks_test

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/runatlantis/atlantis/server/core/db"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/events/vcs"
	"github.com/runatlantis/atlantis/server/events/vcs/types"
	"github.com/runatlantis/atlantis/server/logging"
	"github.com/runatlantis/atlantis/server/lyft/checks"
	"github.com/runatlantis/atlantis/server/lyft/feature"

	. "github.com/runatlantis/atlantis/testing"
)

var checkRunRespFormat = `{
	"id": 4,
	"head_sha": "ce587453ced02b1526dfb4cb910479d431683101",
	"node_id": "MDg6Q2hlY2tSdW40",
	"external_id": "42",
	"url": "https://api.github.com/repos/github/hello-world/check-runs/4",
	"html_url": "https://github.com/github/hello-world/runs/4",
	"details_url": "https://example.com",
	"status": "in_progress",
	"conclusion": null,
	"started_at": "2018-05-04T01:14:52Z",
	"completed_at": null,
	"name": "%s",
	"check_suite": {
	  "id": 5
	},
	"output": {
		"title": "Mighty Readme report",
		"summary": "There are 0 failures, 2 warnings, and 1 notice.",
		"text": "Output text"
	}
  }`

func TestUpdateStatus_FeatureAllocation(t *testing.T) {

	cases := []struct {
		name             string
		shouldAllocate   bool
		isCommitStatus   bool
		isCheckRunStatus bool
	}{
		{
			name:             "use default status update when checks is not enabled",
			shouldAllocate:   false,
			isCommitStatus:   true,
			isCheckRunStatus: false,
		},
		{
			name:             "use github checks when checks is enabled",
			shouldAllocate:   true,
			isCommitStatus:   false,
			isCheckRunStatus: true,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {

			// Reset at the start of each test
			commitStatus := false
			checkRunStatus := false
			statusName := "atlantis/plan"

			checksClientWrapper, boltdb, repo := setup(t, c.shouldAllocate, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch r.RequestURI {
				case "/api/v3/repos/owner/repo/statuses/ref":
					commitStatus = true

				case "/api/v3/repos/owner/repo/check-runs":
					checkRunStatus = true
					w.Write([]byte(fmt.Sprintf(checkRunRespFormat, statusName)))

				default:
					t.Errorf("got unexpected request at %q", r.RequestURI)
					http.Error(w, "not found", http.StatusNotFound)
					return
				}
			}))
			defer disableSSLVerification()()

			checksClientWrapper.UpdateStatus(context.TODO(), types.UpdateStatusRequest{
				Repo:       repo,
				Ref:        "ref",
				State:      models.PendingCommitStatus,
				StatusName: statusName,
			})

			// Assert the right status update is used
			if commitStatus != c.isCommitStatus || checkRunStatus != c.isCheckRunStatus {
				t.FailNow()
			}

			// Check if it was persisted to boltdb
			persistedCheckRunStatus, err := boltdb.GetCheckRunForStatus("atlantis/plan", repo, "ref")
			if c.isCheckRunStatus && (err != nil || persistedCheckRunStatus == nil) {
				t.FailNow()
			}
		})
	}
}

func TestUpdateStatus_PersistCheckRunOutput(t *testing.T) {

	cases := []struct {
		name                string
		statusName          string
		shouldPersistOutput bool
	}{
		{
			name:                "persist checkrun output in bolt db when policy_check command",
			statusName:          "atlantis/plan",
			shouldPersistOutput: false,
		},
		{
			name:                "should not perist checkrun output when not policy_check",
			statusName:          "atlantis/policy_check",
			shouldPersistOutput: true,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {

			checksClientWrapper, boltdb, repo := setup(t, true, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch r.RequestURI {

				case "/api/v3/repos/owner/repo/check-runs":
					w.Write([]byte(fmt.Sprintf(checkRunRespFormat, c.statusName)))

				default:
					t.Errorf("got unexpected request at %q", r.RequestURI)
					http.Error(w, "not found", http.StatusNotFound)
					return
				}
			}))
			defer disableSSLVerification()()

			checksClientWrapper.UpdateStatus(context.TODO(), types.UpdateStatusRequest{
				Repo:       repo,
				Ref:        "ref",
				State:      models.PendingCommitStatus,
				StatusName: c.statusName,
			})

			checkRunStatus, err := boltdb.GetCheckRunForStatus(c.statusName, repo, "ref")
			Ok(t, err)

			// Assert checkrun was persisted
			if checkRunStatus == nil {
				t.FailNow()
			}

			// Assert checkrun output was persisted when necessary
			if (c.shouldPersistOutput && checkRunStatus.Output == "") ||
				(!c.shouldPersistOutput && checkRunStatus.Output != "") {
				t.FailNow()
			}

		})
	}
}

func TestUpdateStatus_PopulatesOutputWhenEmpty(t *testing.T) {
	cases := []struct {
		name                     string
		expectedOutput           string
		populateOutputFromBoltDb bool
		output                   string
	}{
		{
			name:                     "populate output from boltdb for policy_check when output in req is empty",
			populateOutputFromBoltDb: true,
			output:                   "",
			expectedOutput:           "Original output",
		},
		{
			name:                     "do not populate output from boltdb for policy_check when output in req is not empty",
			populateOutputFromBoltDb: false,
			output:                   "Updated output",
			expectedOutput:           "Updated output",
		},
	}
	statusName := "atlantis/policy_check"

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {

			var output string
			checksClientWrapper, boltdb, repo := setup(t, true, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch r.RequestURI {

				case "/api/v3/repos/owner/repo/check-runs/4":
					body, err := ioutil.ReadAll(r.Body)
					Ok(t, err)
					m := make(map[string]interface{})
					err = json.Unmarshal(body, &m)
					Ok(t, err)

					receivedOutput := m["output"].(map[string]interface{})
					output = receivedOutput["text"].(string)

					w.Write([]byte(fmt.Sprintf(checkRunRespFormat, statusName)))

				default:
					t.Errorf("got unexpected request at %q", r.RequestURI)
					http.Error(w, "not found", http.StatusNotFound)
					return
				}
			}))
			defer disableSSLVerification()()

			// Populate boltdb
			boltdb.UpdateCheckRunForStatus("atlantis/policy_check", repo, "ref", models.CheckRunStatus{
				ID:     "4",
				Output: "Original output",
			})

			updateStatusReq := types.UpdateStatusRequest{
				Repo:       repo,
				Ref:        "ref",
				State:      models.SuccessCommitStatus,
				StatusName: statusName,
			}

			if c.output != "" {
				updateStatusReq.Output = c.output
			}

			checksClientWrapper.UpdateStatus(context.TODO(), updateStatusReq)

			if c.expectedOutput != output {
				t.FailNow()
			}

			// Assert last status update is persisted to bolt db
			checkRunStatus, err := boltdb.GetCheckRunForStatus(statusName, repo, "ref")
			Ok(t, err)

			if checkRunStatus.Output != "Output text" {
				t.FailNow()
			}
		})
	}
}

func TestUpdateStatus_ErrorWhenCheckRunDoesNotExist(t *testing.T) {
	dataDir, cleanup := TempDir(t)
	defer cleanup()

	boltdb, err := db.New(dataDir)
	Ok(t, err)

	checksClientWrapper := checks.ChecksClientWrapper{
		FeatureAllocator: &mockFeatureAllocator{shouldAllocate: true},
		Logger:           logging.NewNoopCtxLogger(t),
		Db:               boltdb,
	}

	repo := models.Repo{
		Owner:    "owner",
		Name:     "repo",
		FullName: "owner/repo",
	}

	err = checksClientWrapper.UpdateStatus(context.TODO(), types.UpdateStatusRequest{
		Repo:        repo,
		Ref:         "ref",
		State:       models.SuccessCommitStatus,
		StatusName:  "atlantis/plan",
		Description: "Hello World",
	})

	// assert same error
	ErrEquals(t, "checkrun dne in db", err)
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

func setup(t *testing.T, shouldAllocate bool, handlerFunc http.HandlerFunc) (checks.ChecksClientWrapper, db.BoltDB, models.Repo) {
	testServer := httptest.NewTLSServer(handlerFunc)

	testServerURL, err := url.Parse(testServer.URL)
	Ok(t, err)
	mergeabilityChecker := vcs.NewPullMergeabilityChecker("atlantis")
	client, err := vcs.NewGithubClient(testServerURL.Host, &vcs.GithubUserCredentials{"user", "pass"}, logging.NewNoopCtxLogger(t), mergeabilityChecker)
	Ok(t, err)

	dataDir, cleanup := TempDir(t)
	defer cleanup()

	boltdb, err := db.New(dataDir)
	Ok(t, err)

	repo := models.Repo{
		Owner: "owner",
		Name:  "repo",
	}

	return checks.ChecksClientWrapper{
		GithubClient:     client,
		FeatureAllocator: &mockFeatureAllocator{shouldAllocate: shouldAllocate},
		Logger:           logging.NewNoopCtxLogger(t),
		Db:               boltdb,
	}, *boltdb, repo
}

type mockFeatureAllocator struct {
	shouldAllocate bool
}

func (m *mockFeatureAllocator) ShouldAllocate(featureID feature.Name, featureCtx feature.FeatureContext) (bool, error) {
	return m.shouldAllocate, nil
}
