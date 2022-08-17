package deploy

import (
	"time"

	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/activities"
	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/config"
	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/config/logger"
	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/deploy/revision"
	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/deploy/revision/queue"
	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/github"
	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/steps"
	temporalInternal "github.com/runatlantis/atlantis/server/neptune/workflows/internal/temporal"
	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/terraform"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

const (
	TaskQueue = "deploy"

	// signals
	NewRevisionSignalID = "new-revision"

	RevisionReceiveTimeout = 60 * time.Minute
)

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

func Workflow(ctx workflow.Context, request Request, terraformWorkflow func(ctx workflow.Context, request terraform.Request) error) error {
	options := workflow.ActivityOptions{
		TaskQueue:              TaskQueue,
		ScheduleToCloseTimeout: 5 * time.Second,
	}
	ctx = workflow.WithActivityOptions(ctx, options)

	// set relevant logging context
	ctx = workflow.WithValue(ctx, config.RepositoryLogKey, request.Repository.FullName)
	ctx = workflow.WithValue(ctx, config.ProjectLogKey, request.Root.Name)
	ctx = workflow.WithValue(ctx, config.GHRequestIDLogKey, request.GHRequestID)

	runner := newRunner(ctx, request, terraformWorkflow)

	// blocking call
	return runner.Run(ctx)
}

type Runner struct {
	QueueWorker              QueueWorker
	RevisionReceiver         SignalReceiver
	NewRevisionSignalChannel workflow.ReceiveChannel
}

func newRunner(ctx workflow.Context, request Request, terraformWorkflow func(ctx workflow.Context, request terraform.Request) error) *Runner {
	// convert to internal types, we should probably move these into another struct
	repo := github.Repo{
		Name:  request.Repository.Name,
		Owner: request.Repository.Owner,
		URL:   request.Repository.URL,
		Credentials: github.AppCredentials{
			InstallationToken: request.Repository.Credentials.InstallationToken,
		},
	}
	root := steps.Root{
		Name:  request.Root.Name,
		Apply: steps.Job{Steps: convertToInternalSteps(request.Root.Apply.Steps)},
		Plan:  steps.Job{Steps: convertToInternalSteps(request.Root.Plan.Steps)},
	}

	// inject dependencies

	// temporal effectively "injects" this, it just cares about the method names,
	// so we're modeling our own DI around this.
	var a *activities.Deploy

	revisionQueue := queue.NewQueue()
	revisionReceiver := revision.NewReceiver(ctx, revisionQueue, repo, a)

	worker := &queue.Worker{
		Queue:             revisionQueue,
		Activities:        a,
		Repo:              repo,
		Root:              root,
		TerraformWorkflow: terraformWorkflow,
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

func convertToInternalSteps(requestSteps []Step) []steps.Step {
	var terraformSteps []steps.Step
	for _, step := range requestSteps {
		terraformSteps = append(terraformSteps, steps.Step{
			StepName:    step.StepName,
			ExtraArgs:   step.ExtraArgs,
			RunCommand:  step.RunCommand,
			EnvVarName:  step.EnvVarName,
			EnvVarValue: step.EnvVarValue,
		})
	}
	return terraformSteps
}
