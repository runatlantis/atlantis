package deploy

import (
	"time"

	"github.com/runatlantis/atlantis/server/neptune/workflows/activities"
	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/config/logger"
	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/deploy/revision"
	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/deploy/revision/queue"
	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/deploy/terraform"
	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/sideeffect"
	temporalInternal "github.com/runatlantis/atlantis/server/neptune/workflows/internal/temporal"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

const (
	TaskQueue = "deploy"

	// signals
	NewRevisionSignalID = "new-revision"

	RevisionReceiveTimeout = 60 * time.Minute
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

type SignalReceiver interface {
	Receive(c workflow.ReceiveChannel, more bool)
}

type QueueWorker interface {
	Work(ctx workflow.Context)
	GetState() queue.WorkerState
}

func Workflow(ctx workflow.Context, request Request, tfWorkflow terraform.Workflow) error {
	options := workflow.ActivityOptions{
		TaskQueue:              TaskQueue,
		ScheduleToCloseTimeout: 5 * time.Second,
	}
	ctx = workflow.WithActivityOptions(ctx, options)

	runner := newRunner(ctx, request, tfWorkflow)

	// blocking call
	return runner.Run(ctx)
}

type Runner struct {
	QueueWorker              QueueWorker
	RevisionReceiver         SignalReceiver
	NewRevisionSignalChannel workflow.ReceiveChannel
}

func newRunner(ctx workflow.Context, request Request, tfWorkflow terraform.Workflow) *Runner {
	// inject dependencies

	// temporal effectively "injects" this, it just cares about the method names,
	// so we're modeling our own DI around this.
	var a *workerActivities

	lockStateUpdater := queue.LockStateUpdater{
		Activities: a,
	}
	revisionQueue := queue.NewQueue(func(ctx workflow.Context, d *queue.Deploy) {
		lockStateUpdater.UpdateQueuedRevisions(ctx, d)
	})
	revisionReceiver := revision.NewReceiver(ctx, revisionQueue, a, sideeffect.GenerateUUID)
	tfWorkflowRunner := terraform.NewWorkflowRunner(a, tfWorkflow)
	deployer := &queue.Deployer{
		Activities:              a,
		TerraformWorkflowRunner: tfWorkflowRunner,
	}

	worker := &queue.Worker{
		Queue:    revisionQueue,
		Deployer: deployer,
	}

	return &Runner{
		QueueWorker:              worker,
		RevisionReceiver:         revisionReceiver,
		NewRevisionSignalChannel: workflow.GetSignalChannel(ctx, NewRevisionSignalID),
	}
}

func (r *Runner) Run(ctx workflow.Context) error {
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
	cancelTimer, _ := s.AddTimeout(ctx, RevisionReceiveTimeout, onTimeout)

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
			cancelTimer, _ = s.AddTimeout(ctx, RevisionReceiveTimeout, onTimeout)
			continue
		}

		logger.Info(ctx, "revision receiver timeout")

		// check state here since if we timed out, we're probably not susceptible to the queue
		// worker being in a waiting state right before it's about to start working on an item.
		if !s.HasPending() && r.QueueWorker.GetState() != queue.WorkingWorkerState {
			shutdownWorker()
			break
		}

		// basically keep on adding timeouts until we can either break this loop or get another signal
		// we need to use the timeoutCtx to ensure that this gets cancelled when when the receive is ready
		cancelTimer, _ = s.AddTimeout(ctx, RevisionReceiveTimeout, onTimeout)
	}
	// wait on cancellation so we can gracefully terminate, unsure if temporal handles this for us,
	// but just being safe.
	wg.Wait(ctx)

	return nil
}
