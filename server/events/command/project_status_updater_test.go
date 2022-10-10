package command_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/lyft/feature"
)

type testJobURLGenerator struct {
	expectedURL string
	expectedErr error
}

func (t *testJobURLGenerator) GenerateProjectJobURL(jobID string) (string, error) {
	return t.expectedURL, t.expectedErr
}

type testJobCloser struct {
	called bool
}

func (t *testJobCloser) CloseJob(jobID string, repo models.Repo) {
	t.called = true
}

type testCommitStatusUpdater struct {
	expectedStatusID string
	expectedError    error
}

func (t *testCommitStatusUpdater) UpdateProject(ctx context.Context, projectCtx command.ProjectContext, cmdName fmt.Stringer, status models.CommitStatus, url string, statusID string) (string, error) {
	return t.expectedStatusID, t.expectedError
}

func TestProjectStatusUpdater_CloseJobWhenOperationComplete(t *testing.T) {

	jobURLGenerator := testJobURLGenerator{
		expectedURL: "url",
		expectedErr: nil,
	}

	jobCloser := testJobCloser{}

	commitStatusUpdater := testCommitStatusUpdater{
		expectedStatusID: "1234",
		expectedError:    nil,
	}

	prjStatusUpdater := command.ProjectStatusUpdater{
		ProjectJobURLGenerator:     &jobURLGenerator,
		FeatureAllocator:           testFeatureAllocator{isChecksEnabled: true},
		JobCloser:                  &jobCloser,
		ProjectCommitStatusUpdater: &commitStatusUpdater,
	}

	statusID, err := prjStatusUpdater.UpdateProjectStatus(command.ProjectContext{}, models.SuccessCommitStatus)

	if err != nil {
		t.FailNow()
	}

	if statusID != "1234" {
		t.FailNow()
	}

	if jobCloser.called != true {
		t.FailNow()
	}
}

func TestProjectStatusUpdater_DoNotCloseJobWhenInProgress(t *testing.T) {

	jobURLGenerator := testJobURLGenerator{
		expectedURL: "url",
		expectedErr: nil,
	}

	jobCloser := testJobCloser{}

	commitStatusUpdater := testCommitStatusUpdater{
		expectedStatusID: "1234",
		expectedError:    nil,
	}

	prjStatusUpdater := command.ProjectStatusUpdater{
		ProjectJobURLGenerator:     &jobURLGenerator,
		FeatureAllocator:           testFeatureAllocator{isChecksEnabled: true},
		JobCloser:                  &jobCloser,
		ProjectCommitStatusUpdater: &commitStatusUpdater,
	}

	statusID, err := prjStatusUpdater.UpdateProjectStatus(command.ProjectContext{}, models.PendingCommitStatus)

	if err != nil {
		t.FailNow()
	}

	if statusID != "1234" {
		t.FailNow()
	}

	if jobCloser.called != false {
		t.FailNow()
	}
}

type testFeatureAllocator struct {
	isChecksEnabled bool
}

func (t testFeatureAllocator) ShouldAllocate(featureID feature.Name, featureCtx feature.FeatureContext) (bool, error) {
	return t.isChecksEnabled, nil
}
