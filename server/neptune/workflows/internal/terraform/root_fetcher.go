package terraform

import (
	"github.com/pkg/errors"
	"github.com/runatlantis/atlantis/server/neptune/workflows/activities"
	"github.com/runatlantis/atlantis/server/neptune/workflows/activities/terraform"
	"go.temporal.io/sdk/workflow"
)

type RootFetcher struct {
	Request Request
	Ga      githubActivities
	Ta      terraformActivities
}

// Fetch returns a local root and a cleanup function
func (r *RootFetcher) Fetch(ctx workflow.Context) (*terraform.LocalRoot, func(workflow.Context) error, error) {
	var fetchRootResponse activities.FetchRootResponse
	err := workflow.ExecuteActivity(ctx, r.Ga.GithubFetchRoot, activities.FetchRootRequest{
		Repo:         r.Request.Repo,
		Root:         r.Request.Root,
		DeploymentID: r.Request.DeploymentID,
		Revision:     r.Request.Revision,
	}).Get(ctx, &fetchRootResponse)

	if err != nil {
		return nil, func(_ workflow.Context) error { return nil }, err
	}

	return fetchRootResponse.LocalRoot, func(c workflow.Context) error {
		var cleanupResponse activities.CleanupResponse
		err = workflow.ExecuteActivity(c, r.Ta.Cleanup, activities.CleanupRequest{ //nolint:gosimple // unnecessary to add a method to convert reponses
			DeployDirectory: fetchRootResponse.DeployDirectory,
		}).Get(ctx, &cleanupResponse)
		if err != nil {
			return errors.Wrap(err, "cleaning up")
		}
		return nil
	}, nil
}
