package queue

import (
	"context"
	"fmt"

	key "github.com/runatlantis/atlantis/server/neptune/context"

	"github.com/pkg/errors"
	"github.com/runatlantis/atlantis/server/neptune/workflows/activities"
	"github.com/runatlantis/atlantis/server/neptune/workflows/activities/deployment"
	"github.com/runatlantis/atlantis/server/neptune/workflows/activities/github"
	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/config/logger"
	terraformWorkflow "github.com/runatlantis/atlantis/server/neptune/workflows/internal/deploy/terraform"
	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/metrics"
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
	Run(ctx workflow.Context, deploymentInfo terraformWorkflow.DeploymentInfo, diffDirection activities.DiffDirection) error
}

type dbActivities interface {
	FetchLatestDeployment(ctx context.Context, request activities.FetchLatestDeploymentRequest) (activities.FetchLatestDeploymentResponse, error)
	StoreLatestDeployment(ctx context.Context, request activities.StoreLatestDeploymentRequest) error
}

type githubActivities interface {
	GithubCompareCommit(ctx context.Context, request activities.CompareCommitRequest) (activities.CompareCommitResponse, error)
	GithubUpdateCheckRun(ctx context.Context, request activities.UpdateCheckRunRequest) (activities.UpdateCheckRunResponse, error)
}

type deployerActivities interface {
	dbActivities
	githubActivities
}

type Deployer struct {
	Activities              deployerActivities
	TerraformWorkflowRunner terraformWorkflowRunner
}

const (
	DirectionBehindSummary   = "This revision is behind the current revision and will not be deployed.  If this is intentional, revert the default branch to this revision to trigger a new deployment."
	RerunNotIdenticalSummary = "This revision is not identical to the last revision with an attempted deploy. Reruns are only supported on the most recent deploy."
	UpdateCheckRunRetryCount = 5
)

func (p *Deployer) Deploy(ctx workflow.Context, requestedDeployment terraformWorkflow.DeploymentInfo, latestDeployment *deployment.Info, scope metrics.Scope) (*deployment.Info, error) {
	commitDirection, err := p.getDeployRequestCommitDirection(ctx, requestedDeployment, latestDeployment, scope)
	if err != nil {
		return nil, err
	}
	if commitDirection == activities.DirectionBehind {
		scope.Counter("invalid_commit_direction_err").Inc(1)
		// always returns error for caller to skip revision
		p.updateCheckRun(ctx, requestedDeployment, github.CheckRunFailure, DirectionBehindSummary, nil)
		return nil, NewValidationError("requested revision %s is behind latest deployed revision %s", requestedDeployment.Revision, latestDeployment.Revision)
	}
	if requestedDeployment.Root.Rerun && commitDirection != activities.DirectionIdentical {
		scope.Counter("invalid_rerun_err").Inc(1)
		// always returns error for caller to skip revision
		p.updateCheckRun(ctx, requestedDeployment, github.CheckRunFailure, RerunNotIdenticalSummary, nil)
		return nil, NewValidationError("requested revision %s is a re-run attempt but not identical to the latest deployed revision %s", requestedDeployment.Revision, latestDeployment.Revision)
	}

	// don't wrap this err as it's not necessary and will mess with any err type assertions we might need to do
	err = p.TerraformWorkflowRunner.Run(ctx, requestedDeployment, commitDirection)

	// No need to persist deployment if it's a PlanRejectionError
	if _, ok := err.(*terraformWorkflow.PlanRejectionError); ok {
		return nil, err
	}

	latestDeployment = requestedDeployment.BuildPersistableInfo()
	if persistErr := p.persistLatestDeployment(ctx, latestDeployment); persistErr != nil {
		logger.Error(ctx, "unable to persist deployment, proceeding with in-memory value", key.ErrKey, persistErr)
	}

	// Count this as deployment as latest if it's not a PlanRejectionError which means it is a TerraformClientError
	// We do this as a safety measure to avoid deploying out of order revision after a failed deploy since it could still
	// mutate the state file
	return latestDeployment, err
}

func (p *Deployer) FetchLatestDeployment(ctx workflow.Context, repoName, rootName string) (*deployment.Info, error) {
	var resp activities.FetchLatestDeploymentResponse
	err := workflow.ExecuteActivity(ctx, p.Activities.FetchLatestDeployment, activities.FetchLatestDeploymentRequest{
		FullRepositoryName: repoName,
		RootName:           rootName,
	}).Get(ctx, &resp)
	if err != nil {
		return nil, errors.Wrap(err, "fetching latest deployment")
	}
	return resp.DeploymentInfo, nil
}

func (p *Deployer) getDeployRequestCommitDirection(ctx workflow.Context, deployRequest terraformWorkflow.DeploymentInfo, latestDeployment *deployment.Info, scope metrics.Scope) (activities.DiffDirection, error) {
	// this means we are deploying this root for the first time
	if latestDeployment == nil {
		scope.Counter("first_deployment").Inc(1)
		return activities.DirectionAhead, nil
	}
	var compareCommitResp activities.CompareCommitResponse
	err := workflow.ExecuteActivity(ctx, p.Activities.GithubCompareCommit, activities.CompareCommitRequest{
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
func (p *Deployer) updateCheckRun(ctx workflow.Context, deployRequest terraformWorkflow.DeploymentInfo, state github.CheckRunState, summary string, actions []github.CheckRunAction) {
	ctx = workflow.WithRetryPolicy(ctx, temporal.RetryPolicy{
		MaximumAttempts: UpdateCheckRunRetryCount,
	})
	err := workflow.ExecuteActivity(ctx, p.Activities.GithubUpdateCheckRun, activities.UpdateCheckRunRequest{
		Title:   terraformWorkflow.BuildCheckRunTitle(deployRequest.Root.Name),
		State:   state,
		Repo:    deployRequest.Repo,
		ID:      deployRequest.CheckRunID,
		Summary: summary,
		Actions: actions,
	}).Get(ctx, nil)
	if err != nil {
		logger.Error(ctx, "unable to update check run with validation error", key.ErrKey, err)
	}
}

func (p *Deployer) persistLatestDeployment(ctx workflow.Context, deploymentInfo *deployment.Info) error {
	// retry indefinitely since until we can guarantee persistance on shutdown
	// TODO: Persist deployment on shutdown
	err := workflow.ExecuteActivity(ctx, p.Activities.StoreLatestDeployment, activities.StoreLatestDeploymentRequest{
		DeploymentInfo: deploymentInfo,
	}).Get(ctx, nil)
	if err != nil {
		return errors.Wrap(err, "persisting deployment info")
	}
	return nil
}
