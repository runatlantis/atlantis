package deploy

import (
	"github.com/runatlantis/atlantis/server/events/metrics"
	"time"

	"github.com/pkg/errors"
	"github.com/runatlantis/atlantis/server/neptune/workflows/activities"
	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/config/logger"
	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/deploy/revision"
	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/deploy/revision/queue"
	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/deploy/terraform"
	workflowMetrics "github.com/runatlantis/atlantis/server/neptune/workflows/internal/metrics"
	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/sideeffect"
	temporalInternal "github.com/runatlantis/atlantis/server/neptune/workflows/internal/temporal"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

const (
	TaskQueue = "deploy"

	// signals
	NewRevisionSignalID = "new-revision"

	RevisionReceiveTimeout    = 60 * time.Minute
	ActiveDeployWorkflowStat  = "active"
	SuccessDeployWorkflowStat = "success"
)

type workerActivities struct {
	*activities.Github
	*activities.Deploy
}

type RunnerAction int64

const (
	OnCancel RunnerAction = iota
	OnTimeout
	OnReceive
)

type container interface {
	IsEmpty() bool
}

type SignalReceiver interface {
	Receive(c workflow.ReceiveChannel, more bool)
}

type QueueWorker interface {
	Work(ctx workflow.Context)
	GetState() queue.WorkerState
}

func Workflow(ctx workflow.Context, request Request, tfWorkflow terraform.Workflow) error {
	options := workflow.ActivityOptions{
		TaskQueue:           TaskQueue,
		StartToCloseTimeout: 5 * time.Second,
	}
	ctx = workflow.WithActivityOptions(ctx, options)

	runner, err := newRunner(ctx, request, tfWorkflow)

	if err != nil {
		return errors.Wrap(err, "initializing workflow runner")
	}

	// blocking call
	return runner.Run(ctx)
}

type Runner struct {
	Timeout                  time.Duration
	Queue                    container
	QueueWorker              QueueWorker
	RevisionReceiver         SignalReceiver
	NewRevisionSignalChannel workflow.ReceiveChannel
	Scope                    workflowMetrics.Scope
}

func newRunner(ctx workflow.Context, request Request, tfWorkflow terraform.Workflow) (*Runner, error) {
	// inject dependencies

	// temporal effectively "injects" this, it just cares about the method names,
	// so we're modeling our own DI around this.
	var a *workerActivities

	metricsHandler := workflow.GetMetricsHandler(ctx).WithTags(map[string]string{
		metrics.RepoTag: request.Repo.FullName,
		metrics.RootTag: request.Root.Name,
	})
	scope := workflowMetrics.NewScope(metricsHandler, "workflow", "deploy")

	lockStateUpdater := queue.LockStateUpdater{
		Activities: a,
	}
	revisionQueue := queue.NewQueue(func(ctx workflow.Context, d *queue.Deploy) {
		lockStateUpdater.UpdateQueuedRevisions(ctx, d)
	}, scope)

	worker, err := queue.NewWorker(ctx, revisionQueue, scope, a, tfWorkflow, request.Repo.FullName, request.Root.Name)
	if err != nil {
		return nil, err
	}

	revisionReceiver := revision.NewReceiver(ctx, revisionQueue, a, sideeffect.GenerateUUID, worker,
		scope.SubScopeWithTags(map[string]string{
			workflowMetrics.SignalNameTag: NewRevisionSignalID,
		}))

	return &Runner{
		Queue:                    revisionQueue,
		Timeout:                  RevisionReceiveTimeout,
		QueueWorker:              worker,
		RevisionReceiver:         revisionReceiver,
		NewRevisionSignalChannel: workflow.GetSignalChannel(ctx, NewRevisionSignalID),
		Scope:                    scope,
	}, nil
}

func (r *Runner) shutdown() {
	r.Scope.Gauge(ActiveDeployWorkflowStat).Update(0)
	r.Scope.Counter(SuccessDeployWorkflowStat).Inc(1)
}

func (r *Runner) Run(ctx workflow.Context) error {
	r.Scope.Gauge(ActiveDeployWorkflowStat).Update(1)
	defer r.shutdown()

	var action RunnerAction
	workerCtx, shutdownWorker := workflow.WithCancel(ctx)

	wg := workflow.NewWaitGroup(ctx)
	wg.Add(1)

	// if this panics in anyway, we'll need to ship a fix to the running workflows, else risk dropping
	// signals
	// should we have some way of persisting our queue in case of workflow termination?
	// Let's address this in a followup
	workflow.Go(workerCtx, func(ctx workflow.Context) {
		defer wg.Done()
		r.QueueWorker.Work(ctx)
	})

	onTimeout := func(f workflow.Future) {
		err := f.Get(ctx, nil)

		if temporal.IsCanceledError(err) {
			action = OnCancel
			return
		}

		action = OnTimeout
	}

	s := temporalInternal.SelectorWithTimeout{
		Selector: workflow.NewSelector(ctx),
	}
	s.AddReceive(r.NewRevisionSignalChannel, func(c workflow.ReceiveChannel, more bool) {
		r.RevisionReceiver.Receive(c, more)
		action = OnReceive

	})
	cancelTimer, _ := s.AddTimeout(ctx, r.Timeout, onTimeout)

	// main loop which handles external signals
	// and in turn signals the queue worker
	for {
		// blocks until a configured callback fires
		s.Select(ctx)

		switch action {
		case OnCancel:
			continue
		case OnReceive:
			cancelTimer()
			cancelTimer, _ = s.AddTimeout(ctx, r.Timeout, onTimeout)
			continue
		}

		logger.Info(ctx, "revision receiver timeout")

		// Since we timed out, let's determine whether we can shutdown our worker.
		// If we have no incoming revisions and the worker is just waiting, thats the first sign.
		// We also need to ensure that we're also checking whether the queue is empty since the worker can be in a waiting state
		// if the queue is locked (ie. if the workflow has just started up with prior deploy state)
		if !s.HasPending() && r.QueueWorker.GetState() != queue.WorkingWorkerState && r.Queue.IsEmpty() {
			logger.Info(ctx, "initiating worker shutdown")
			shutdownWorker()
			break
		}

		// basically keep on adding timeouts until we can either break this loop or get another signal
		// we need to use the timeoutCtx to ensure that this gets cancelled when when the receive is ready
		cancelTimer, _ = s.AddTimeout(ctx, r.Timeout, onTimeout)
	}
	// wait on cancellation so we can gracefully terminate, unsure if temporal handles this for us,
	// but just being safe.
	wg.Wait(ctx)

	return nil
}
