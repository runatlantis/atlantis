package queue

import (
	"context"
	"fmt"

	key "github.com/runatlantis/atlantis/server/neptune/context"

	metricNames "github.com/runatlantis/atlantis/server/events/metrics"

	"github.com/pkg/errors"
	internalContext "github.com/runatlantis/atlantis/server/neptune/context"
	"github.com/runatlantis/atlantis/server/neptune/workflows/activities"
	"github.com/runatlantis/atlantis/server/neptune/workflows/activities/deployment"
	tfModel "github.com/runatlantis/atlantis/server/neptune/workflows/activities/terraform"
	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/config/logger"
	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/deploy/terraform"
	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/metrics"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

type queue interface {
	IsEmpty() bool
	CanPop() bool
	Pop() (terraform.DeploymentInfo, error)
	SetLockForMergedItems(ctx workflow.Context, state LockState)
}

type deployer interface {
	Deploy(ctx workflow.Context, requestedDeployment terraform.DeploymentInfo, latestDeployment *deployment.Info, scope metrics.Scope) (*deployment.Info, error)
}

type workerActivities interface {
	deployerActivities
	AuditJob(ctx context.Context, request activities.AuditJobRequest) error
}

type WorkerState string

const (
	WaitingWorkerState  WorkerState = "waiting"
	WorkingWorkerState  WorkerState = "working"
	CompleteWorkerState WorkerState = "complete"

	UnlockSignalName = "unlock"
)

type UnlockSignalRequest struct {
	User string
}

type CurrentDeploymentStatus int

type CurrentDeployment struct {
	Deployment terraform.DeploymentInfo
	Status     CurrentDeploymentStatus
}

const (
	InProgressStatus CurrentDeploymentStatus = iota
	CompleteStatus
)

type Worker struct {
	Queue    queue
	Deployer deployer

	// mutable
	state             WorkerState
	latestDeployment  *deployment.Info
	currentDeployment CurrentDeployment
}

type actionType string

const (
	canceled = "canceled"
	process  = "process"
	receive  = "receive"
)

func NewWorker(
	ctx workflow.Context,
	q queue,
	a workerActivities,
	tfWorkflow terraform.Workflow,
	repoName, rootName string,
	githubCheckRunCache CheckRunClient,
) (*Worker, error) {
	tfWorkflowRunner := terraform.NewWorkflowRunner(a, tfWorkflow, githubCheckRunCache)
	deployer := &Deployer{
		Activities:              a,
		TerraformWorkflowRunner: tfWorkflowRunner,
		GithubCheckRunCache:     githubCheckRunCache,
	}

	latestDeployment, err := deployer.FetchLatestDeployment(ctx, repoName, rootName)
	if err != nil {
		return nil, errors.Wrap(err, "fetching current deployment")
	}

	// we don't persist lock state anywhere so in the case of workflow completion we need to rebuild
	// the lock state
	if latestDeployment != nil && latestDeployment.Root.Trigger == string(tfModel.ManualTrigger) {
		q.SetLockForMergedItems(ctx, LockState{
			Status:   LockedStatus,
			Revision: latestDeployment.Revision,
		})
	}

	return &Worker{
		Queue:            q,
		Deployer:         deployer,
		latestDeployment: latestDeployment,
	}, nil
}

// Work pops work off the queue and if the queue is empty,
// it waits for the queue to be non-empty or a cancelation signal
func (w *Worker) Work(ctx workflow.Context) {
	// set to complete once we return else callers could think we are still working based on the 'working' state.
	defer func() {
		w.state = CompleteWorkerState
	}()

	var currentAction actionType
	callback := func(f workflow.Future) {
		err := f.Get(ctx, nil)

		if temporal.IsCanceledError(err) {
			currentAction = canceled
			return
		}

		if err != nil {
			logger.Error(ctx, fmt.Sprintf("Unknown error %s, worker is shutting down", err.Error()))
			currentAction = canceled
			return
		}

		currentAction = process
	}
	selector := workflow.NewSelector(ctx)
	selector.AddFuture(w.awaitWork(ctx), callback)

	var request UnlockSignalRequest
	selector.AddReceive(workflow.GetSignalChannel(ctx, UnlockSignalName), func(c workflow.ReceiveChannel, more bool) {
		_ = c.Receive(ctx, &request)
		currentAction = receive
	})

	for {
		if w.Queue.IsEmpty() {
			w.state = WaitingWorkerState
		}

		selector.Select(ctx)

		var currentDeployment *deployment.Info
		var err error
		switch currentAction {
		case canceled:
			logger.Info(ctx, "Received cancelled signal, worker is shutting down")
			return
		case process:
			logger.Info(ctx, "Processing... ")
		case receive:
			logger.Info(ctx, "Received unlock signal... ")
			workflow.GetMetricsHandler(ctx).WithTags(map[string]string{metricNames.SignalNameTag: UnlockSignalName}).
				Counter(metricNames.SignalReceive).
				Inc(1)
			w.Queue.SetLockForMergedItems(ctx, LockState{
				Status: UnlockedStatus,
			})
			continue
		default:
			logger.Warn(ctx, fmt.Sprintf("%s action not configured. This is probably a bug, skipping for now", currentAction))
			return
		}

		w.state = WorkingWorkerState
		msg, err := w.Queue.Pop()
		if err != nil {
			logger.Error(ctx, "failed to pop next revision off of queue, this is most definitely a bug.", key.ErrKey, err)
			continue
		}

		ctx = setContextKeys(ctx, msg)

		scope := metrics.NewScope(
			ctx,
			"revision",
		)

		w.emitRevisionRequestStats(scope, msg)
		currentDeployment, err = w.deploy(ctx, msg, w.latestDeployment, scope)

		// since there was no error we can safely count this as our latest deploy
		if err == nil {
			scope.Counter("success").Inc(1)
			w.latestDeployment = currentDeployment
			selector.AddFuture(w.awaitWork(ctx), callback)
			continue
		}

		var readableErr string
		switch e := err.(type) {
		case *ValidationError:
			readableErr = "validation"
			logger.Error(ctx, "deploy validation failed, moving to next one", key.ErrKey, e)
		case *terraform.PlanRejectionError:
			readableErr = "plan_rejected"
			logger.Warn(ctx, "Plan rejected")
		default:

			// If it's not a ValidationError or PlanRejectionError, it's most likely a TerraformClientError and it is possible the state file
			// is mutated if the apply operation failed. If the next deployment in the queue is behind this commit [out of order] and we try to deploy
			// it, we'd essentially go back in history which is not safe for terraform state files. So, as a safety measure, we'll update the failed
			// deployment as latest deployment and allow redeploy as long as the failed deploy is the latest deployment.
			w.latestDeployment = currentDeployment

			readableErr = "unknown"
			logger.Error(ctx, "failed to deploy revision, moving to next one", key.ErrKey, err)
		}

		scope.SubScopeWithTags(map[string]string{"error_type": readableErr}).Counter("failure")

		selector.AddFuture(w.awaitWork(ctx), callback)
	}
}

func (w *Worker) emitRevisionRequestStats(scope metrics.Scope, request terraform.DeploymentInfo) {
	if request.Root.Rerun {
		scope.Counter("rerun_requested").Inc(1)
	}

	planMode := request.Root.Plan.GetPlanMode().String()
	if planMode != "" {
		scope.Counter(fmt.Sprintf("%s_plan_mode_requested", planMode)).Inc(1)
	}
}

func (w *Worker) setCurrentDeploymentState(state CurrentDeployment) {
	w.currentDeployment = state
}

func (w *Worker) GetCurrentDeploymentState() CurrentDeployment {
	return w.currentDeployment
}

func (w *Worker) GetLatestDeployment() *deployment.Info {
	return w.latestDeployment
}

func (w *Worker) awaitWork(ctx workflow.Context) workflow.Future {
	future, settable := workflow.NewFuture(ctx)

	workflow.Go(ctx, func(ctx workflow.Context) {
		err := workflow.Await(ctx, func() bool {
			return w.Queue.CanPop()
		})

		settable.SetError(err)
	})

	return future
}

func setContextKeys(ctx workflow.Context, requestedDeployment terraform.DeploymentInfo) workflow.Context {
	ctx = workflow.WithValue(ctx, internalContext.SHAKey, requestedDeployment.Revision)
	ctx = workflow.WithValue(ctx, internalContext.DeploymentIDKey, requestedDeployment.ID)
	ctx = workflow.WithValue(ctx, internalContext.PlanMode, string(requestedDeployment.Root.Plan.GetPlanMode()))
	ctx = workflow.WithValue(ctx, internalContext.Trigger, string(requestedDeployment.Root.Trigger))

	return ctx
}

func (w *Worker) deploy(ctx workflow.Context, requestedDeployment terraform.DeploymentInfo, latestDeployment *deployment.Info, scope metrics.Scope) (*deployment.Info, error) {
	w.setCurrentDeploymentState(CurrentDeployment{
		Deployment: requestedDeployment,
		Status:     InProgressStatus,
	})
	defer w.setCurrentDeploymentState(CurrentDeployment{
		Deployment: requestedDeployment,
		Status:     CompleteStatus,
	})

	return w.Deployer.Deploy(ctx, requestedDeployment, latestDeployment, scope)
}

func (w *Worker) GetState() WorkerState {
	return w.state
}
