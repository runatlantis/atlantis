package workflows

import (
	"os"
	"path/filepath"
	"time"

	"github.com/pkg/errors"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/lyft/temporal/activities"
	"go.temporal.io/sdk/workflow"
)

const (
	// signal names
	PlanReviewSignal     = "plan-review"
	NewCommitAddedSignal = "new-commit-added"

	TaskQueue = "atlantis-tq"
)

const (
	Approved PlanReviewStatus = iota
	Discard
)

type Repo struct {
	Owner string
	Name  string
}

type DeployRequest struct {
	Repo   Repo
	Branch string
}

type DeployStatus int
type PlanReviewStatus int

type PlanReview struct {
	User   string
	Status PlanReviewStatus
}

type Deploy struct {
	DataDir string
}

func (d *Deploy) Run(ctx workflow.Context, request DeployRequest) error {
	logger := workflow.GetLogger(ctx)
	options := workflow.ActivityOptions{
		TaskQueue:              TaskQueue,
		ScheduleToCloseTimeout: 30 * time.Second,
	}

	ctx = workflow.WithActivityOptions(ctx, options)

	// used to receive new commits for the repo ready for deployment
	commitAddedSelector := workflow.NewSelector(ctx)
	commitAddedCh := workflow.GetSignalChannel(ctx, NewCommitAddedSignal)

	var revision string
	commitAddedSelector.AddReceive(commitAddedCh, func(c workflow.ReceiveChannel, more bool) {
		c.Receive(ctx, &revision)

		logger.Info("Received revision %s for %s", request.Repo.Name)
	})

	// loop forever to basically emulate a queue that is receiving new commits
	for {

		// Ensures that if we don't have pending commits to deploy, we can break out of this workflow.
		// if we do we run through a deployment for the revision
		if !commitAddedSelector.HasPending() {
			break
		}

		commitAddedSelector.Select(ctx)

		// fetch repository details
		getRepositoryResponse, err := getRepository(ctx, request)
		if err != nil {
			return err
		}
		repo := getRepositoryResponse.Repo
		repoDir := filepath.Join(d.DataDir, repo.Owner, repo.Name)

		// execute the actual deployment related activities
		err = executeDeploy(ctx, request, repo, repoDir, revision)

		if err != nil {
			return err
		}
	}

	return nil
}

func executeDeploy(ctx workflow.Context, request DeployRequest, repo models.Repo, repoDir, revision string) error {
	logger := workflow.GetLogger(ctx)

	opts := &workflow.SessionOptions{
		CreationTimeout:  time.Minute,
		ExecutionTimeout: 30 * time.Minute,
	}

	// sessions ensure that we are running our plans and applies on the same worker
	// which basically ensures this workflow can run on multiple hosts
	sessionCtx, err := workflow.CreateSession(ctx, opts)
	if err != nil {
		return errors.Wrapf(err, "creating deployment session")
	}

	defer workflow.CompleteSession(sessionCtx)

	// execute set of activities that make up a terraform plan
	planResponse, err := plan(sessionCtx, request, repo, repoDir, revision)
	if err != nil {
		return err
	}

	// block until we get a valid plan review
	planReview := awaitPlanReview(sessionCtx)

	if planReview.Status == Discard {
		logger.Warn("Deployment cancelled due to discarded plan")
		return nil
	}

	// since the plan has now been approved let's apply
	return apply(sessionCtx, planResponse, repoDir)
}

func getRepository(ctx workflow.Context, request DeployRequest) (activities.GetRepositoryResponse, error) {
	var vcsClient *activities.VCSClientWrapper
	var response activities.GetRepositoryResponse
	err := workflow.ExecuteActivity(ctx, vcsClient.GetRepository, &activities.GetRepositoryRequest{
		Owner: request.Repo.Owner,
		Repo:  request.Repo.Name,
	}).Get(ctx, &response)

	if err != nil {
		return activities.GetRepositoryResponse{}, errors.Wrap(err, " executing get repository activity")
	}

	return response, nil
}

func plan(ctx workflow.Context, request DeployRequest, repo models.Repo, repoDir string, revision string) (activities.PlanResponse, error) {
	// MmkdirAll guarantees this is replayable
	err := os.MkdirAll(repoDir, os.ModePerm)
	if err != nil {
		return activities.PlanResponse{}, errors.Wrapf(err, "creating repo directory")
	}

	var cloneResponse activities.CloneActivityResponse
	err = workflow.ExecuteActivity(
		ctx,
		activities.Clone,
		&activities.CloneActivityRequest{
			Repo:     repo,
			Revision: revision,
			Branch:   request.Branch,
			Dir:      repoDir,
		}).Get(ctx, &cloneResponse)

	if err != nil {
		return activities.PlanResponse{}, errors.Wrap(err, " executing clone activity")
	}

	err = workflow.ExecuteActivity(ctx, activities.Init, &activities.InitRequest{RootDir: cloneResponse.Dir}).Get(ctx, nil)
	if err != nil {
		return activities.PlanResponse{}, errors.Wrap(err, " executing terraform init activity")
	}

	var planResponse activities.PlanResponse
	err = workflow.ExecuteActivity(ctx, activities.Plan, &activities.PlanRequest{RootDir: cloneResponse.Dir}).Get(ctx, &planResponse)
	if err != nil {
		return activities.PlanResponse{}, errors.Wrap(err, " executing terraform plan activity")
	}

	return planResponse, nil
}

func apply(ctx workflow.Context, planResponse activities.PlanResponse, repoDir string) error {
	err := workflow.ExecuteActivity(ctx, activities.Apply, &activities.ApplyRequest{RootDir: repoDir, Planfile: planResponse.Planfile}).Get(ctx, nil)

	if err != nil {
		return errors.Wrap(err, "executing terraform apply activity")
	}

	return nil
}

func awaitPlanReview(ctx workflow.Context) PlanReview {
	logger := workflow.GetLogger(ctx)
	planReviewSelector := workflow.NewSelector(ctx)
	planReviewCh := workflow.GetSignalChannel(ctx, PlanReviewSignal)

	var planReview PlanReview
	planReviewSelector.AddReceive(planReviewCh, func(c workflow.ReceiveChannel, more bool) {
		c.Receive(ctx, &planReview)
		logger.Info("Received plan review")
	})

	// Forever loop until we get a valid plan review status
	for {
		planReviewSelector.Select(ctx)

		if planReview.Status == Discard || planReview.Status == Approved {
			break
		} else {
			logger.Warn("Unsupported plan review status %d", planReview.Status)
		}
	}

	return planReview
}
