package command_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/lyft/feature"
)

type testJobUrlGenerator struct {
	expectedUrl string
	expectedErr error
}

func (t *testJobUrlGenerator) GenerateProjectJobURL(jobID string) (string, error) {
	return t.expectedUrl, t.expectedErr
}

type testJobCloser struct {
	called bool
}

func (t *testJobCloser) CloseJob(jobID string, repo models.Repo) {
	t.called = true
}

type testCommitStatusUpdater struct {
	expectedStatusId string
	expectedError    error
}

func (t *testCommitStatusUpdater) UpdateProject(ctx context.Context, projectCtx command.ProjectContext, cmdName fmt.Stringer, status models.CommitStatus, url string, statusId string) (string, error) {
	return t.expectedStatusId, t.expectedError
}

func TestProjectStatusUpdater_CloseJobWhenOperationComplete(t *testing.T) {

	jobUrlGenerator := testJobUrlGenerator{
		expectedUrl: "url",
		expectedErr: nil,
	}

	jobCloser := testJobCloser{}

	commitStatusUpdater := testCommitStatusUpdater{
		expectedStatusId: "1234",
		expectedError:    nil,
	}

	prjStatusUpdater := command.ProjectStatusUpdater{
		ProjectJobURLGenerator:     &jobUrlGenerator,
		FeatureAllocator:           testFeatureAllocator{isChecksEnabled: true},
		JobCloser:                  &jobCloser,
		ProjectCommitStatusUpdater: &commitStatusUpdater,
	}

	statusId, err := prjStatusUpdater.UpdateProjectStatus(command.ProjectContext{}, models.SuccessCommitStatus)

	if err != nil {
		t.FailNow()
	}

	if statusId != "1234" {
		t.FailNow()
	}

	if jobCloser.called != true {
		t.FailNow()
	}
}

func TestProjectStatusUpdater_DoNotCloseJobWhenInProgress(t *testing.T) {

	jobUrlGenerator := testJobUrlGenerator{
		expectedUrl: "url",
		expectedErr: nil,
	}

	jobCloser := testJobCloser{}

	commitStatusUpdater := testCommitStatusUpdater{
		expectedStatusId: "1234",
		expectedError:    nil,
	}

	prjStatusUpdater := command.ProjectStatusUpdater{
		ProjectJobURLGenerator:     &jobUrlGenerator,
		FeatureAllocator:           testFeatureAllocator{isChecksEnabled: true},
		JobCloser:                  &jobCloser,
		ProjectCommitStatusUpdater: &commitStatusUpdater,
	}

	statusId, err := prjStatusUpdater.UpdateProjectStatus(command.ProjectContext{}, models.PendingCommitStatus)

	if err != nil {
		t.FailNow()
	}

	if statusId != "1234" {
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

type testPrjCmdBuilder struct {
	expectedResp struct {
		prjCtxs []command.ProjectContext
		err     error
	}
}
