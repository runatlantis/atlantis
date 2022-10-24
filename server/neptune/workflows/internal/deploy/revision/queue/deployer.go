package queue

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	"github.com/runatlantis/atlantis/server/neptune/workflows/activities"
	"github.com/runatlantis/atlantis/server/neptune/workflows/activities/deployment"
	"github.com/runatlantis/atlantis/server/neptune/workflows/activities/github"
	terraformWorkflow "github.com/runatlantis/atlantis/server/neptune/workflows/internal/deploy/terraform"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

type ValidationError struct {
	error
}

func NewValidationError(msg string, format ...interface{}) *ValidationError {
	return &ValidationError{fmt.Errorf(msg, format...)}
}

type terraformWorkflowRunner interface {
	Run(ctx workflow.Context, deploymentInfo terraformWorkflow.DeploymentInfo) error
}

type dbActivities interface {
	FetchLatestDeployment(ctx context.Context, request activities.FetchLatestDeploymentRequest) (activities.FetchLatestDeploymentResponse, error)
	StoreLatestDeployment(ctx context.Context, request activities.StoreLatestDeploymentRequest) error
}

type githubActivities interface {
	CompareCommit(ctx context.Context, request activities.CompareCommitRequest) (activities.CompareCommitResponse, error)
	UpdateCheckRun(ctx context.Context, request activities.UpdateCheckRunRequest) (activities.UpdateCheckRunResponse, error)
}

type workerActivities interface {
	dbActivities
	githubActivities
}

type Deployer struct {
	Activities              workerActivities
	TerraformWorkflowRunner terraformWorkflowRunner
}

const (
	DirectionBehindSummary   = "This revision is behind the current revision and will not be deployed.  If this is intentional, revert the default branch to this revision to trigger a new deployment."
	UpdateCheckRunRetryCount = 5

	DivergedCommitsSummary = "The current deployment has diverged from the default branch, so we have locked the root. This is most likely the result of this PR performing a manual deployment. To override that lock and allow the main branch to perform new deployments, select the Unlock button."
)

func (p *Deployer) Deploy(ctx workflow.Context, requestedDeployment terraformWorkflow.DeploymentInfo, latestDeployment *deployment.Info) (*deployment.Info, error) {
	latestDeployment, err := p.fetchLatestDeployment(ctx, requestedDeployment, latestDeployment)
	if err != nil {
		return nil, err
	}
	commitDirection := activities.DirectionAhead
	// only fetch if last deployment exists
	if latestDeployment != nil {
		commitDirection, err = p.getDeployRequestCommitDirection(ctx, requestedDeployment, latestDeployment)
		if err != nil {
			return nil, err
		}
	}
	switch commitDirection {
	case activities.DirectionBehind:
		// always returns error for caller to skip revision
		if err = p.updateCheckRun(ctx, requestedDeployment, github.CheckRunFailure, DirectionBehindSummary, nil); err != nil {
			return nil, errors.Wrap(err, "updating check run")
		}

		return nil, NewValidationError("requested revision %s is behind latest deployed revision %s", requestedDeployment.Revision, latestDeployment.Revision)
	}
	err = p.TerraformWorkflowRunner.Run(ctx, requestedDeployment)
	if err != nil {
		return nil, errors.Wrap(err, "running terraform workflow")
	}
	latestDeployment = p.buildLatestDeployment(requestedDeployment)

	// TODO: Persist deployment on shutdown if it fails instead of blocking
	if err = p.persistLatestDeployment(ctx, latestDeployment); err != nil {
		return nil, errors.Wrap(err, "failed to persist latest deploy job")
	}
	return latestDeployment, nil
}

func (p *Deployer) getDeployRequestCommitDirection(ctx workflow.Context, deployRequest terraformWorkflow.DeploymentInfo, latestDeployment *deployment.Info) (activities.DiffDirection, error) {
	var compareCommitResp activities.CompareCommitResponse
	err := workflow.ExecuteActivity(ctx, p.Activities.CompareCommit, activities.CompareCommitRequest{
		DeployRequestRevision:  deployRequest.Revision,
		LatestDeployedRevision: latestDeployment.Revision,
		Repo:                   deployRequest.Repo,
	}).Get(ctx, &compareCommitResp)
	if err != nil {
		return "", errors.Wrap(err, "unable to determine deploy request commit direction")
	}
	return compareCommitResp.CommitComparison, nil
}

// worker should not block on updating check runs for invalid deploy requests so let's retry for UpdateCheckrunRetryCount only
func (p *Deployer) updateCheckRun(ctx workflow.Context, deployRequest terraformWorkflow.DeploymentInfo, state github.CheckRunState, summary string, actions []github.CheckRunAction) error {
	ctx = workflow.WithRetryPolicy(ctx, temporal.RetryPolicy{
		MaximumAttempts: UpdateCheckRunRetryCount,
	})
	return workflow.ExecuteActivity(ctx, p.Activities.UpdateCheckRun, activities.UpdateCheckRunRequest{
		Title:   terraformWorkflow.BuildCheckRunTitle(deployRequest.Root.Name),
		State:   state,
		Repo:    deployRequest.Repo,
		ID:      deployRequest.CheckRunID,
		Summary: summary,
		Actions: actions,
	}).Get(ctx, nil)
}

func (p *Deployer) fetchLatestDeployment(ctx workflow.Context, deploymentInfo terraformWorkflow.DeploymentInfo, latestDeployment *deployment.Info) (*deployment.Info, error) {
	// Skip fetching latest deployment it it's already in memory
	if latestDeployment != nil {
		return latestDeployment, nil
	}
	var resp activities.FetchLatestDeploymentResponse
	err := workflow.ExecuteActivity(ctx, p.Activities.FetchLatestDeployment, activities.FetchLatestDeploymentRequest{
		FullRepositoryName: deploymentInfo.Repo.GetFullName(),
		RootName:           deploymentInfo.Root.Name,
	}).Get(ctx, &resp)
	if err != nil {
		return nil, errors.Wrap(err, "fetching latest deployment")
	}
	return resp.DeploymentInfo, nil
}

func (p *Deployer) buildLatestDeployment(deployRequest terraformWorkflow.DeploymentInfo) *deployment.Info {
	return &deployment.Info{
		Version:  deployment.InfoSchemaVersion,
		ID:       deployRequest.ID.String(),
		Revision: deployRequest.Revision,
		Root: deployment.Root{
			Name: deployRequest.Root.Name,
		},
		Repo: deployment.Repo{
			Name:  deployRequest.Repo.Name,
			Owner: deployRequest.Repo.Owner,
		},
	}
}

func (p *Deployer) persistLatestDeployment(ctx workflow.Context, deploymentInfo *deployment.Info) error {
	err := workflow.ExecuteActivity(ctx, p.Activities.StoreLatestDeployment, activities.StoreLatestDeploymentRequest{
		DeploymentInfo: deploymentInfo,
	}).Get(ctx, nil)
	if err != nil {
		return errors.Wrap(err, "persisting deployment info")
	}
	return nil
}
