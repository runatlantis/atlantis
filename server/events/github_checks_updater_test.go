// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package events_test

import (
	"errors"
	"testing"

	"github.com/google/go-github/v71/github"
	"github.com/runatlantis/atlantis/server/events"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/logging"
	. "github.com/runatlantis/atlantis/testing"
)

// fakeChecksClient is a test double for GithubChecksClientInterface.
type fakeChecksClient struct {
	createCalls []github.CreateCheckRunOptions
	updateCalls []updateCheckRunCall
	findResults map[string]int64 // name -> id to return from FindCheckRunID
	createErr   error
	updateErr   error
	findErr     error
	nextID      int64
}

type updateCheckRunCall struct {
	id   int64
	opts github.UpdateCheckRunOptions
}

func (f *fakeChecksClient) CreateCheckRun(_ logging.SimpleLogging, _ models.Repo, opts github.CreateCheckRunOptions) (int64, error) {
	if f.createErr != nil {
		return 0, f.createErr
	}
	f.createCalls = append(f.createCalls, opts)
	f.nextID++
	return f.nextID, nil
}

func (f *fakeChecksClient) UpdateCheckRun(_ logging.SimpleLogging, _ models.Repo, id int64, opts github.UpdateCheckRunOptions) error {
	if f.updateErr != nil {
		return f.updateErr
	}
	f.updateCalls = append(f.updateCalls, updateCheckRunCall{id: id, opts: opts})
	return nil
}

func (f *fakeChecksClient) FindCheckRunID(_ logging.SimpleLogging, _ models.Repo, _, name string) (int64, error) {
	if f.findErr != nil {
		return 0, f.findErr
	}
	if f.findResults != nil {
		return f.findResults[name], nil
	}
	return 0, nil
}

func makeTestPull() models.PullRequest {
	return models.PullRequest{
		Num:        42,
		HeadCommit: "abc123",
		BaseRepo: models.Repo{
			Owner:    "myorg",
			Name:     "myrepo",
			FullName: "myorg/myrepo",
		},
	}
}

func makeTestRepo() models.Repo {
	return models.Repo{
		Owner:    "myorg",
		Name:     "myrepo",
		FullName: "myorg/myrepo",
	}
}

func makeTestProjectContext(workspace, relDir, projectName string) command.ProjectContext {
	return command.ProjectContext{
		Log:         logging.NewNoopLogger(nil),
		Pull:        makeTestPull(),
		BaseRepo:    makeTestRepo(),
		Workspace:   workspace,
		RepoRelDir:  relDir,
		ProjectName: projectName,
	}
}

// TestParseCheckRunExternalID verifies the external_id round-trip encoding.
func TestParseCheckRunExternalID(t *testing.T) {
	t.Log("when externalID contains the separator, it should be parsed as workspace::relDir")
	projectName, workspace, relDir := events.ParseCheckRunExternalID("default::infra/staging")
	Equals(t, "", projectName)
	Equals(t, "default", workspace)
	Equals(t, "infra/staging", relDir)

	t.Log("when externalID has no separator, it should be treated as a projectName")
	projectName, workspace, relDir = events.ParseCheckRunExternalID("my-project")
	Equals(t, "my-project", projectName)
	Equals(t, "", workspace)
	Equals(t, "", relDir)

	t.Log("when externalID is empty, all fields should be empty")
	projectName, workspace, relDir = events.ParseCheckRunExternalID("")
	Equals(t, "", projectName)
	Equals(t, "", workspace)
	Equals(t, "", relDir)
}

// TestGithubChecksUpdater_UpdateCombined_CreatesThenUpdates verifies that the
// first UpdateCombined call creates a check run and subsequent ones update it.
func TestGithubChecksUpdater_UpdateCombined_CreatesThenUpdates(t *testing.T) {
	t.Log("UpdateCombined should create a check run on first call and update on second")
	client := &fakeChecksClient{}
	updater := events.NewGithubChecksUpdater(client, "atlantis")

	logger := logging.NewNoopLogger(t)
	repo := makeTestRepo()
	pull := makeTestPull()

	// First call → create + update (pending)
	err := updater.UpdateCombined(logger, repo, pull, models.PendingCommitStatus, command.Plan)
	Ok(t, err)
	Equals(t, 1, len(client.createCalls))
	Equals(t, 1, len(client.updateCalls))
	Equals(t, "atlantis/plan", client.createCalls[0].Name)
	Equals(t, "in_progress", client.createCalls[0].GetStatus())
	Equals(t, "", client.createCalls[0].GetConclusion())

	// Second call → only update (success), no new create
	err = updater.UpdateCombined(logger, repo, pull, models.SuccessCommitStatus, command.Plan)
	Ok(t, err)
	Equals(t, 1, len(client.createCalls)) // still 1 create
	Equals(t, 2, len(client.updateCalls))
	Equals(t, "completed", client.updateCalls[1].opts.GetStatus())
	Equals(t, "success", client.updateCalls[1].opts.GetConclusion())
}

// TestGithubChecksUpdater_UpdateCombined_UsesExistingCheckRun verifies that
// when FindCheckRunID returns an existing ID, no new check run is created.
func TestGithubChecksUpdater_UpdateCombined_UsesExistingCheckRun(t *testing.T) {
	t.Log("UpdateCombined should use an existing check run ID from the API instead of creating a new one")
	client := &fakeChecksClient{
		findResults: map[string]int64{"atlantis/plan": 99},
	}
	updater := events.NewGithubChecksUpdater(client, "atlantis")

	logger := logging.NewNoopLogger(t)
	repo := makeTestRepo()
	pull := makeTestPull()

	err := updater.UpdateCombined(logger, repo, pull, models.SuccessCommitStatus, command.Plan)
	Ok(t, err)
	// No create since FindCheckRunID returned an existing ID.
	Equals(t, 0, len(client.createCalls))
	Equals(t, 1, len(client.updateCalls))
	Equals(t, int64(99), client.updateCalls[0].id)
}

// TestGithubChecksUpdater_UpdateProject_PlanSuccess verifies that a successful
// plan check run includes the plan output, diff summary, and Apply action.
func TestGithubChecksUpdater_UpdateProject_PlanSuccess(t *testing.T) {
	t.Log("UpdateProject for a successful plan should include plan output and an Apply action")
	client := &fakeChecksClient{}
	updater := events.NewGithubChecksUpdater(client, "atlantis")

	ctx := makeTestProjectContext("default", "infra/staging", "")
	planOutput := "Terraform will perform the following actions:\n\n  + null_resource.test\n\nPlan: 1 to add, 0 to change, 0 to destroy."
	result := &command.ProjectCommandOutput{
		PlanSuccess: &models.PlanSuccess{
			TerraformOutput: planOutput,
		},
	}

	err := updater.UpdateProject(ctx, command.Plan, models.SuccessCommitStatus, "https://atlantis/logs/1", result)
	Ok(t, err)

	Equals(t, 1, len(client.createCalls))
	// check the update call
	Equals(t, 1, len(client.updateCalls))
	update := client.updateCalls[0]
	Equals(t, "completed", update.opts.GetStatus())
	Equals(t, "success", update.opts.GetConclusion())
	Assert(t, update.opts.Output != nil, "output should not be nil")
	Assert(t, update.opts.Output.Text != nil, "output text should not be nil")
	Assert(t, len(*update.opts.Output.Text) > 0, "output text should not be empty")
	// Should have the Apply action
	Equals(t, 1, len(update.opts.Actions))
	Equals(t, "apply", update.opts.Actions[0].Identifier)
	Equals(t, "Apply", update.opts.Actions[0].Label)
	// external_id should encode workspace::relDir
	Assert(t, update.opts.ExternalID != nil, "external_id should not be nil")
	Equals(t, "default::infra/staging", *update.opts.ExternalID)
}

// TestGithubChecksUpdater_UpdateProject_PlanSuccessNoChanges verifies that a
// plan check run with no changes does NOT have an Apply action.
func TestGithubChecksUpdater_UpdateProject_PlanSuccessNoChanges(t *testing.T) {
	t.Log("UpdateProject for a plan with no changes should not include an Apply action")
	client := &fakeChecksClient{}
	updater := events.NewGithubChecksUpdater(client, "atlantis")

	ctx := makeTestProjectContext("default", ".", "")
	result := &command.ProjectCommandOutput{
		PlanSuccess: &models.PlanSuccess{
			TerraformOutput: "No changes. Infrastructure is up-to-date.",
		},
	}

	err := updater.UpdateProject(ctx, command.Plan, models.SuccessCommitStatus, "", result)
	Ok(t, err)

	Equals(t, 1, len(client.updateCalls))
	update := client.updateCalls[0]
	// No Apply action since no changes
	Equals(t, 0, len(update.opts.Actions))
}

// TestGithubChecksUpdater_UpdateProject_PlanFailure verifies that a failed plan
// check run uses the failure conclusion and includes the error text.
func TestGithubChecksUpdater_UpdateProject_PlanFailure(t *testing.T) {
	t.Log("UpdateProject for a failed plan should use failure conclusion and include error output")
	client := &fakeChecksClient{}
	updater := events.NewGithubChecksUpdater(client, "atlantis")

	ctx := makeTestProjectContext("default", ".", "")
	result := &command.ProjectCommandOutput{
		Error: errors.New("Error: invalid configuration"),
	}

	err := updater.UpdateProject(ctx, command.Plan, models.FailedCommitStatus, "", result)
	Ok(t, err)

	Equals(t, 1, len(client.updateCalls))
	update := client.updateCalls[0]
	Equals(t, "completed", update.opts.GetStatus())
	Equals(t, "failure", update.opts.GetConclusion())
	// No Apply action on failure
	Equals(t, 0, len(update.opts.Actions))
	// error output is included
	Assert(t, update.opts.Output != nil, "output should not be nil")
	Assert(t, update.opts.Output.Text != nil, "output text should not be nil")
	Assert(t, len(*update.opts.Output.Text) > 0, "error text should not be empty")
}

// TestGithubChecksUpdater_UpdateProject_WithProjectName verifies that when a
// project name is set, it is used as the external_id (not workspace::relDir).
func TestGithubChecksUpdater_UpdateProject_WithProjectName(t *testing.T) {
	t.Log("UpdateProject should encode the project name as external_id when set")
	client := &fakeChecksClient{}
	updater := events.NewGithubChecksUpdater(client, "atlantis")

	ctx := makeTestProjectContext("default", ".", "my-project")

	err := updater.UpdateProject(ctx, command.Apply, models.SuccessCommitStatus, "", &command.ProjectCommandOutput{
		ApplySuccess: "Apply complete! Resources: 0 added, 0 changed, 0 destroyed.",
	})
	Ok(t, err)

	Equals(t, 1, len(client.updateCalls))
	update := client.updateCalls[0]
	Assert(t, update.opts.ExternalID != nil, "external_id should not be nil")
	Equals(t, "my-project", *update.opts.ExternalID)
}

// TestGithubChecksUpdater_TruncatesLongOutput verifies that plan output
// exceeding the GitHub check run limit is truncated.
func TestGithubChecksUpdater_TruncatesLongOutput(t *testing.T) {
	t.Log("plan output exceeding 65535 characters should be truncated")
	client := &fakeChecksClient{}
	updater := events.NewGithubChecksUpdater(client, "atlantis")

	ctx := makeTestProjectContext("default", ".", "")
	// Build a string longer than the limit
	longOutput := make([]byte, 70000)
	for i := range longOutput {
		longOutput[i] = 'x'
	}
	result := &command.ProjectCommandOutput{
		PlanSuccess: &models.PlanSuccess{
			TerraformOutput: string(longOutput) + " Plan: 1 to add, 0 to change, 0 to destroy.",
		},
	}

	err := updater.UpdateProject(ctx, command.Plan, models.SuccessCommitStatus, "", result)
	Ok(t, err)

	Equals(t, 1, len(client.updateCalls))
	update := client.updateCalls[0]
	Assert(t, update.opts.Output != nil, "output should not be nil")
	Assert(t, update.opts.Output.Text != nil, "output text should not be nil")
	// Truncated to at most 65535 characters
	Assert(t, len(*update.opts.Output.Text) <= 65535, "output text should be truncated to ≤65535 chars")
}

// TestGithubChecksUpdater_UpdateCombinedCount verifies the count summary text.
func TestGithubChecksUpdater_UpdateCombinedCount(t *testing.T) {
	t.Log("UpdateCombinedCount should include the count in the summary")
	client := &fakeChecksClient{}
	updater := events.NewGithubChecksUpdater(client, "atlantis")

	logger := logging.NewNoopLogger(t)
	repo := makeTestRepo()
	pull := makeTestPull()

	err := updater.UpdateCombinedCount(logger, repo, pull, models.SuccessCommitStatus, command.Plan, 2, 3)
	Ok(t, err)

	Equals(t, 1, len(client.updateCalls))
	update := client.updateCalls[0]
	Assert(t, update.opts.Output != nil, "output should not be nil")
	Assert(t, update.opts.Output.Summary != nil, "summary should not be nil")
	Assert(t, len(*update.opts.Output.Summary) > 0, "summary should not be empty")
	// Summary should contain the counts.
	Assert(t, contains(*update.opts.Output.Summary, "2/3"), "summary should contain '2/3' but got: "+*update.opts.Output.Summary)
}

// TestGithubChecksUpdater_CreateError propagates errors from CreateCheckRun.
func TestGithubChecksUpdater_CreateError(t *testing.T) {
	t.Log("errors from CreateCheckRun should be propagated")
	client := &fakeChecksClient{createErr: errors.New("api error")}
	updater := events.NewGithubChecksUpdater(client, "atlantis")

	logger := logging.NewNoopLogger(t)
	repo := makeTestRepo()
	pull := makeTestPull()

	err := updater.UpdateCombined(logger, repo, pull, models.PendingCommitStatus, command.Plan)
	Assert(t, err != nil, "expected error from CreateCheckRun to be propagated")
}

// TestGithubChecksUpdater_UpdateError propagates errors from UpdateCheckRun.
func TestGithubChecksUpdater_UpdateError(t *testing.T) {
	t.Log("errors from UpdateCheckRun should be propagated")
	client := &fakeChecksClient{
		updateErr:   errors.New("update api error"),
		findResults: map[string]int64{"atlantis/plan": 5},
	}
	updater := events.NewGithubChecksUpdater(client, "atlantis")

	logger := logging.NewNoopLogger(t)
	repo := makeTestRepo()
	pull := makeTestPull()

	err := updater.UpdateCombined(logger, repo, pull, models.SuccessCommitStatus, command.Plan)
	Assert(t, err != nil, "expected error from UpdateCheckRun to be propagated")
}

// contains is a simple substring helper used in assertions.
func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(sub) == 0 || (len(s) > 0 && indexOf(s, sub) >= 0))
}

func indexOf(s, sub string) int {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}
