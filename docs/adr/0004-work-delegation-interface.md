# 4. Work Delegation Interface

Date: 2026-08-07

## Status

Proposed

## Context

Atlantis runs terraform in same process which serves webhooks. One server holds VCS
integration and also forks every `terraform plan` and `apply` as child process, with
concurrency limited by `--parallel-pool-size`.

This couples two workloads which have completely different shape. Webhook server lives
long and is almost always idle. Terraform executions are short, bursty and hungry for
memory — single `terraform plan` against big provider schema holds hundreds of MB, and
AWS provider alone unpacks to more than 500MB. So operator who sizes deployment must
provision for peak concurrency and then pay for it 24/7, also when nothing happens. And
when burst does not fit into the box, it takes webhook server down together with itself.

Same coupling is what blocks horizontal scaling. Two replicas behind load balancer means
replica which receives `apply` webhook is maybe not the one which did run `plan`. This
is the failure reported again and again in #1571.

From four pieces of shared state, three have already some story:

- project locks and pull status — BoltDB or Redis, by `--locking-db-type`, shared is fine
- plan artifacts — local FS or S3, by `--enable-external-stores` from #6312, shared is fine
- job output and streamed logs — in process memory, no backend at all, not shared
- working directory and clone — local `--data-dir`, but it can be built again, so not a problem

Issue #6694 proposes to unify how those backends are selected, and #6742 covers key and
path conventions which come out from it.

What does not exist is seam in the codebase where execution can be given to something
else. Requests for this are open for years — #260 from 2018, #1571 from 2021, #3791 from
2023. One concrete attempt, helm-level request proxy in front of Atlantis
(`SneaksAndData:atlantis:gateway`, 2023), was living outside of codebase and never
merged, because it solved problem only for one VCS and one deployment topology.

Position stated on #3791 is that specific flavors should not land piece by piece:

> I think the best approach here would be to define a standard interface for delegating
> work within Atlantis and then implement specific flavors, rather than bringing in this
> functionality piecemeal.

This ADR proposes such interface. It does not implement any flavor.

Not in scope here:

- choosing remote execution technology. Kubernetes Jobs, worker binary over gRPC,
  Teleport, queue frameworks — all of them are flavors and can be argued separately,
  after seam exists
- changing default behaviour. In-process execution stays default and stays only
  supported configuration, until some flavor is proposed and accepted on own merits
- changing meaning of `--parallel-pool-size`
- multi-tenancy or per-repo isolation guarantees

## Decision

Delegation happens on level of project command, so one pair of `(project, command)`.
This is already unit of work in every other place of codebase:

```go
// server/events/project_command_pool_executor.go
type prjCmdRunnerFunc func(ctx command.ProjectContext) command.ProjectCommandOutput
```

`runProjectCmdsParallel` already fans out exactly these across sized wait group. Same
granularity is used by project lock, by `jobID` for output streaming, and by
`planstore.PlanStore.Save` and `Load`, which are keyed on `command.ProjectContext`. So
every boundary which matters is drawn there already.

Two other places were considered and rejected.

Per shell command, so `models.ShellCommandRunner.Run`, is too fine. Workflow steps share
working directory, `.terraform` directory and environment which is accumulated by `env`
and `multienv` steps. To delegate each step separately means to ship working directory
back and forward on every step, and custom `run` steps anyway expect local filesystem
which survives between steps.

Per whole command, so all projects together, is too coarse. Per-project parallelism is
exactly the property which operators want to scale, so to collapse it kills the purpose.

Interface then looks like this:

```go
// server/core/delegation
//
// Executor runs single project command. Local implementation runs it in
// process, exactly like Atlantis does it today. Other implementations may
// run it somewhere else.
type Executor interface {
    // Execute runs cmd until it finishes and returns result. Implementation
    // must be at-most-once for side effects: returned error means command
    // maybe did run and maybe did not, so caller must never retry apply
    // without operator who looks on it.
    Execute(ctx context.Context, req Request) (command.ProjectCommandOutput, error)

    // Name tells which flavor this is, for logs and for /status endpoint.
    Name() string
}

type Request struct {
    // RunID is unique per delegation attempt and works as fencing token.
    // Implementation which cannot prove exclusivity for RunID must not
    // execute at all.
    RunID string

    // Project is explicit wire representation of command.ProjectContext.
    // ProjectContext is not serialized as it is, because it carries logger
    // and fields which come from VCS and must not leave controller.
    Project ProjectSpec

    // Source tells executor how to materialize repository on the commit
    // which is under test.
    Source SourceRef
}
```

`ProjectSpec` is new wire type and it is versioned explicitly.
`command.ProjectContext` is not serialized directly, because it carries
`logging.SimpleLogging` and fields derived from VCS, and to gob it would turn every
future field into question about wire compatibility.

Default `LocalExecutor` wraps existing `DefaultProjectCommandRunner` and changes no
behaviour. For registration it follows whatever selection mechanism #6694 will settle
on, instead of inventing third one.

What crosses boundary, in direction to executor: project spec, workflow steps which must
run, environment for the run, and way to get repository on right commit. Back from
executor: `command.ProjectCommandOutput`, plan artifact and log stream.

Plan artifact needs no new mechanism, `PlanStore` externalizes it already and remote
executor writes through same interface. Log stream is different story.
`jobs.ProjectCommandOutputHandler` today keeps output in process memory and registers
in-process channels by `jobID`. Remote executor has nowhere to send output, and web UI
sitting on another replica has nowhere to read it. This is the one hard dependency —
job output must become pluggable backend before any out-of-process flavor can work.

`SourceRef` is left open on purpose, because two possible answers have different
security properties and choice should be made together with concrete flavor:

- executor clones from VCS by itself. Simple, but then every executor needs VCS
  credentials, which widens blast radius in serious way
- controller uploads prepared working directory to object storage and executor takes it
  from there. Keeps credentials in one place and handles `merge` checkout strategy
  correctly, but moves the tree two times

Phases are then:

- extract `Executor` and implement `LocalExecutor` over existing runner. No behaviour
  change, no new flags, no new dependencies. Can be reviewed alone
- make job output pluggable, following backend selection from #6694 and key conventions
  from #6742
- define and version wire contract, including at-most-once semantics and how `RunID`
  fences against existing project lock
- propose first out-of-process flavor as separate ADR

First and third phase are useful also in case no flavor is ever merged, because they
make execution path testable without forking of processes.

## Consequences

Operator gets finally possibility to size webhook server for what it really does, which
is almost nothing, and let execution burst somewhere elastic. Together with Redis
locking and external plan store, replicas stop to be a data sharing problem. Per-project
resource limits and per-project blast radius become possible to express, first time.

Price is real also. Every out-of-process flavor brings version skew between controller
and whatever runs terraform — workflow semantics live in both of them, so they must be
pinned together. Log streaming gets network hop and "tail the running plan" experience
becomes harder to keep responsive. To debug failed run means to look in two places. And
secrets must reach executor somehow, which is security review, not config option.

Nothing from this lands on operators who do not opt in. `LocalExecutor` is default and
it is current code path.

Questions which stay open:

- does `SourceRef` clone on executor side or ship working directory, as written above
- how pre- and post-workflow hooks interact with delegation. They run per repo, not per
  project, so they sit on controller side of the seam. But `run` steps inside workflow
  do not sit there, and this difference should be documented before somebody is
  surprised by it
- should policy checks be delegated. Today they are project command, so by default they
  fall in scope, but conftest has different resource profile than terraform
- `--parallel-pool-size` today limits local concurrency. With remote executor it limits
  instead how many delegations are in flight, which is different thing under same name
