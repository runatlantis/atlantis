# 4. Generalize policy evaluation around workflow stages

Date: 2026-09-04

## Status

Proposed

## Context

Atlantis policy checking is currently built around a post-plan Conftest
workflow. After a successful plan, Atlantis runs the `policy_check` project
command, interprets its output, records policy status and approvals, and gates
apply through the `policies_passed` requirement. Other policy engines can be
used through custom workflow steps, but must fit Conftest-oriented output and
approval behavior.

The current framework provides important governance and usability features:

- Atlantis controls policy-set selection, ownership, and required approvals;
- policy approval is independent from pull-request review;
- denied plans cannot be applied until their policies pass or are approved;
- sticky approvals remain valid while the findings they cover remain
  unchanged.

However, policy evaluation is not separated cleanly from its first evaluator.
Conftest execution, acquisition, source resolution, parsing, persistence, and
rendering assumptions cross runtime and event orchestration. Custom policy
checks are mapped to policy sets by position and interpreted from
human-readable output. This makes additional evaluators difficult to integrate
without reproducing Conftest behavior.

The current post-plan command also limits where policy can protect a workflow.
A post-plan policy can determine whether a plan is eligible to apply, but it
cannot prevent arbitrary pre-plan workflow code, enforce conditions immediately
before apply, or protect import and state-removal operations.

Policy engines expose different native result formats and identifiers.
Conftest reports messages and optional metadata, while tools such as Checkov
provide check and resource identifiers, and custom engines may return entirely
different structures. Native evaluator output is therefore not a suitable
Atlantis persistence, approval, or rendering contract.

Atlantis needs a small policy domain that preserves its existing governance and
sticky-approval behavior while separating evaluator execution from workflow
orchestration. Policy evaluations must attach to existing operation boundaries,
produce normalized findings, and render consistently regardless of evaluator.

## Decision

Atlantis will own a tool-neutral policy domain based on structured findings.
Conftest will become one adapter to that domain. Custom command evaluators may
use a simple exit-code protocol or a versioned JSON policy protocol; both are
normalized into the same domain result.

Policy evaluation will attach to boundaries around existing workflow stages:

```text
plan.before       plan.after
apply.before
import.before
state_rm.before
```

Atlantis will continue to own policy selection, input collection, decisions,
approvals, persistence, VCS status, and Markdown rendering. Evaluators will
only evaluate the supplied request and return findings.

Conftest is the first built-in evaluator, not a special policy domain. Atlantis
can add peer built-in evaluators such as Checkov by implementing the same
interface and normalization contract. A new adapter may add its own shared
runtime configuration and policy-set block, but does not change workflow
orchestration, decisions, approvals, persistence, or rendering.

### Configuration

The new framework has a separate server-controlled configuration path. It does
not extend the legacy `policies` schema.

```yaml
policy_framework:
  evaluators:
    conftest:
      version: 0.66.0
    checkov:
      binary: /usr/local/bin/checkov

  policy_sets:
    - name: infrastructure
      source:
        repository: https://github.com/example/terraform-policies.git
        ref: main
        path: conftest/infrastructure

      conftest:
        args:
          - --namespace=terraform

      approve_count: 1
      owners:
        teams: [platform]
      prevent_self_approve: true

      workflow_stages:
        plan.after: {}

        apply.before:
          inputs:
            - $SHOWFILE
            - change-request.json

    - name: security
      checkov: {}
      approve_count: -1
      workflow_stages:
        plan.after: {}

    - name: scanner
      command:
        argv:
          - /usr/local/bin/atlantis-scanner-policy
          - evaluate
        protocol: policy-v1
        timeout: 2m
        max_output_bytes: 1048576

      approve_count: -1

      workflow_stages:
        plan.after:
          inputs:
            - $SHOWFILE
            - reports/compliance.json

repos:
  - id: github.com/example/infrastructure
    policy_framework:
      enabled: true
```

The Checkov entries illustrate how another built-in evaluator fits beside
Conftest. They define the extension shape, not a requirement to implement
Checkov in the first framework change.

The configuration has five concepts:

1. `evaluators` enables and configures shared built-in evaluator runtimes
   managed by Atlantis. This decision implements Conftest; other types can be
   added as peers.
2. A `policy_set` identifies an optional policy source, one evaluator, and
   Atlantis-owned approval rules.
3. `workflow_stages` is a map from a unique stage boundary to its input
   selection. A map prevents attaching the same policy set to the same boundary
   more than once.
4. A policy set contains exactly one built-in evaluator block (`conftest` in
   this decision) or a `command` block.
5. A server-side repository rule enables the framework for that repository.

The supported `workflow_stages` keys are `plan.before`, `plan.after`,
`apply.before`, `import.before`, and `state_rm.before`. Validation rejects other
keys.

Policy-set definitions are global and inert until a matching `repos` rule sets
`policy_framework.enabled: true`. In the initial model, enabling the framework
applies every global policy set to every effective project in that repository.
Atlantis resolves policy inputs only after repository enablement and effective
project configuration are known.

`repos[].policy_framework` is an object so future policy-set or project
selection can be added without changing its configuration type. The initial
release does not define those selectors.

`evaluators.conftest.version` tells Atlantis to download and cache the selected
release using its existing version cache. It may instead specify exactly one
installed binary:

```yaml
evaluators:
  conftest:
    binary: /usr/local/bin/conftest
```

`version` and `binary` are mutually exclusive.

`conftest.args` configures evaluator behavior such as namespace selection. The
adapter owns the policy path, JSON output mode, and positional input files;
validation rejects arguments that override those values.

A policy source is either a server-local directory:

```yaml
source:
  path: /etc/atlantis/policies/infrastructure
```

or a Git repository:

```yaml
source:
  repository: https://github.com/example/terraform-policies.git
  ref: main
  path: conftest/infrastructure
```

Without `repository`, `path` is local. With `repository`, `ref` is required and
`path` selects a directory within the checkout. Conftest requires a source;
evaluators that carry their own policies may omit it. Policy sources contain
policy definitions, while `workflow_stages[].inputs` select evidence to
evaluate.

A `command` is configured completely within its policy set. Its `argv` may
start a compiled binary, a script with a shebang, or an explicit interpreter:

```yaml
command:
  argv:
    - /bin/bash
    - /etc/atlantis/policies/change-window.sh
  protocol: exit-code
```

Atlantis executes `argv` directly without an implicit shell. The configuration
is server-controlled; a pull request cannot introduce or replace an evaluator
executable. `protocol` is required and accepts `exit-code` or `policy-v1`.

`approve_count` has the following meaning:

- a positive value permits a denial to be overridden by that many eligible
  owners;
- `-1`, or omission, makes denials non-overridable;
- `0` is invalid;
- warnings never block;
- evaluator errors always fail closed at a gate and cannot be approved.

A denial is overridden when at least `approve_count` distinct, currently
eligible owners have approval snapshots covering every current denial finding.
Repeated approval by the same user is idempotent. The non-overridable default
is intentional for this new configuration path; legacy configuration retains
its existing default and behavior.

`prevent_self_approve` compares the authenticated approval actor with the
pull-request author. When enabled, an approval from the pull-request author
does not count, regardless of who executed the protected workflow. This rule
has the same meaning at every policy point.

There is no separate enforcement field. Evaluator finding type and Atlantis
approval configuration completely determine the result.

### Workflow-stage boundaries

Policy points bound the complete effective built-in or custom project workflow:

```text
resolve request, effective project configuration, and authorization
  -> prepare the current project workdir under the serialization lock,
     including clone and merge
  -> validate ordinary command requirements
  -> evaluate operation.before
  -> run every configured workflow step; the standard plan step produces
     $PLANFILE and its canonical JSON representation at $SHOWFILE
  -> for plan, evaluate plan.after
  -> persist and publish the result
```

`plan.before` always runs before the first configured plan step.
`plan.after` always runs after the last successful plan step. Custom workflow
steps do not move those boundaries, and policy cannot be inserted between
arbitrary workflow steps.

`before` policies are evaluated for every operation attempt. They may inspect
the saved artifact, but they may also decide from current execution context.
For example, `apply.before` can enforce blackout or maintenance windows,
incident freezes, and actor-specific authorization that was not known—or may
have changed—when `plan.after` ran.

| Policy point | What it protects | Example policy |
| --- | --- | --- |
| `plan.before` | Arbitrary plan workflow code, data lookups, credential use, and initialization | Allow only approved lookup or configuration-generation scripts before planning |
| `plan.after` | Whether a completed plan may be applied | Deny deletion of a protected database or public exposure of a service |
| `apply.before` | Infrastructure mutation and arbitrary apply workflow code | Deny apply during a blackout window or active incident freeze |
| `import.before` | Importing a protected or inappropriate resource into state | Require approval before importing a production database or externally managed resource |
| `state_rm.before` | Removing ownership or tracking of protected resources | Deny removal of protected resource addresses without break-glass approval |

Policy decisions affect workflow as follows:

| Boundary | Effect of an uncleared denial |
| --- | --- |
| Any `before` boundary | Do not start the workflow stage |
| `plan.after` | Keep the plan, but do not make it apply-eligible |

Policy points are limited to gates that can still prevent the protected action.
Post-apply, post-import, and post-state-removal checks cannot undo a completed
mutation and are outside this decision.

### Relationship to command requirements

New-framework policies enforce themselves at their configured boundaries;
users do not add `policies_passed` to command requirements. Existing
`plan_requirements`, `apply_requirements`, and `import_requirements` retain
their current meanings and are evaluated before the corresponding `before`
policy.

| Operation | Existing requirements | Intrinsic policy checks |
| --- | --- | --- |
| Plan | `plan_requirements` | `plan.before`, then `plan.after` on success |
| Apply | `apply_requirements` and a current apply-eligible plan | `apply.before` |
| Import | `import_requirements` | `import.before` |
| State removal | Existing command validation | `state_rm.before` |

A plan becomes apply-eligible only after the plan workflow succeeds and all of
its configured plan-boundary evaluations are cleared. Those evaluations remain
current only while the corresponding saved plan remains current.

`policies_passed` remains a legacy-only implementation requirement and is
still injected for legacy policy checking. It is neither injected nor read for
a new-framework project. Atlantis rejects `policies_passed` in the effective
requirements of a repository using `policy_framework`; it does not silently
ignore or remove the requirement.

### Inputs

Inputs belong to a workflow-stage entry because the same policy set may need
different evidence at different boundaries.

Omitting `inputs` for `plan.after` defaults to:

```yaml
inputs:
  - $SHOWFILE
```

An explicit list replaces the default. Before points have no default file
input.

A Conftest stage requires at least one input; a command evaluator may require no
file inputs. Inputs are explicit file paths; glob patterns are not supported.
Atlantis will:

- anchor ordinary paths to the project working directory;
- allow Atlantis-owned runtime paths such as `$SHOWFILE` explicitly;
- reject escaping paths and symlinks;
- require every configured input to resolve to a regular file.

The standard plan step owns `$SHOWFILE` generation. Policy coordination only
resolves and reads the artifact; it does not invoke Terraform. A custom
workflow that replaces the standard plan step is responsible for producing
the configured inputs, including `$SHOWFILE` when the default is retained.

Files generated by plan are available to `plan.after`, but not to
`plan.before`. A `before` policy may consume repository files or artifacts from
an earlier operation.

Atlantis resolves a policy source before invoking its evaluator:

```go
type PolicySourceResolver interface {
	Resolve(
		ctx context.Context,
		source PolicySource,
	) (path string, err error)
}
```

Atlantis resolves a repository source to an immutable local checkout before
invoking an evaluator. Resolution failures are non-approvable evaluation
errors, and Git credentials remain owned by Atlantis.

### Evaluator contract

The internal evaluator interface evaluates one configured policy set at one
workflow-stage boundary:

```go
type Evaluator interface {
	Evaluate(
		ctx context.Context,
		req EvaluationRequest,
	) (EvaluationResponse, error)
}

type EvaluationRequest struct {
	Version       int
	PolicySet     string
	WorkflowStage WorkflowStage
	PolicyPath    string
	Inputs        []string
}
```

An evaluator error means the adapter could not produce a trustworthy normalized
result. A native engine's nonzero exit status does not by itself determine
whether the result is a denial or an evaluator error.

A command evaluator runs in the project working directory and receives the JSON
representation of `EvaluationRequest` on standard input. It may ignore standard
input and use configured files, environment, or an external service instead.
Its top-level fields are `version`, `policy_set`,
`workflow_stage`, optional `policy_path`, and `inputs`; `inputs` contains the
resolved file paths. The command also receives the documented project and
workflow environment normally available to a workflow run step.

The `exit-code` protocol normalizes process output as follows:

```text
exit 0       -> passed; stdout may become evaluation diagnostic
exit 1       -> denied; non-empty stdout becomes one denial finding
other exit   -> evaluator error
```

Because stdout is one finding, formatting changes affect sticky approval
identity. Tools that cannot reliably distinguish denial from error by exit code
require a wrapper or the `policy-v1` protocol.

Timeout, cancellation, or oversized output is an evaluator error under either
command protocol.

The `policy-v1` protocol writes exactly one versioned response to standard
output:

```json
{
  "version": 1,
  "findings": [
    {
      "type": "denial",
      "rule_id": "prevent-destroy",
      "subject": "aws_db_instance.primary",
      "message": "database instance will be deleted",
      "details": "Triggered by the production retention policy.",
      "metadata": {
        "action": "delete",
        "environment": "production"
      }
    }
  ],
  "diagnostic": "18 rules evaluated; 2 findings"
}
```

Only `version`, `findings`, `finding.type`, and `finding.message` are required.
`findings` may be empty. `rule_id`, `subject`, `details`, `metadata`, and the
top-level `diagnostic` are optional.

The contract intentionally has no evaluator-supplied outcome, blocking flag,
approval fields, general severity scale, or fingerprint. Atlantis derives:

```text
evaluator or protocol failure       -> error
one or more denial findings         -> denied
warnings only, or no findings       -> passed
```

Under `policy-v1`, a command exits zero for every valid response, including
denials. A nonzero exit, invalid response, or unsupported version is an
evaluator error. Atlantis strictly validates the response. Standard error is
bounded process diagnostic output: it is logged and may accompany an evaluator
error, but is not interpreted as a finding or persisted as the structured
`diagnostic`.

Within `policy-v1`, Atlantis may add optional request fields. Command evaluators
must ignore request fields they do not understand. Atlantis rejects unknown
top-level and finding fields so misspelled or unsupported results cannot
silently affect a policy decision; arbitrary keys remain valid inside
`metadata`. Changing required fields or the semantics of an existing field
requires a new protocol version.

### Normalized findings and metadata

The common finding is intentionally small:

```go
type Finding struct {
	Type     FindingType // warning or denial
	RuleID   string
	Subject  string
	Message  string
	Details  string
	Metadata json.RawMessage
	Digest   string
}
```

`rule_id` is descriptive, optional, and not necessarily unique. Findings are
stored as a slice and never keyed or deduplicated by rule ID.

`message` describes the finding and is the only required descriptive field.
`subject` identifies the affected object, not the location of the policy rule.

`metadata` is a JSON object for selected, stable, decision-bearing engine
information, not a copy of the native result. Atlantis canonicalizes and
persists it but assigns no engine-specific meaning. Evaluator adapters remove
volatile native fields before normalization.

`details` is optional plain-text context for one finding. `details` and the
evaluation-level `diagnostic` are presentation-only. Volatile locations, code
excerpts, traces, timestamps, and process output belong there. Information
whose change must invalidate an approval belongs in `message`, `subject`, or
`metadata`.

For Conftest, the adapter maps:

| Conftest value | Finding value |
| --- | --- |
| warning / failure | warning / denial |
| `msg` | `message` |
| `metadata.rule_id`, then `metadata.id`, then `metadata.query` | `rule_id` |
| `metadata.subject`; `metadata.resource` as a compatibility alias | `subject` |
| `metadata.details` | `details` |
| nested `metadata.metadata` and other policy-supplied stable values | `metadata` |
| `loc`, outputs, traces, and success or exception summaries | `diagnostic` |

Because Conftest has no canonical finding ID, its query becomes one digest
component when used as the fallback rule ID. Conftest policies may return
strings, which produce findings with only a message, or use the documented
Atlantis convention:

```rego
deny contains finding if {
  change := input.resource_changes[_]
  change.change.actions == ["delete"]

  finding := {
    "msg": "production resources must not be deleted",
    "rule_id": "prevent-production-delete",
    "subject": change.address,
    "details": "Deletion requires an approved exception.",
    "metadata": {
      "action": "delete",
      "environment": "production",
    },
  }
}
```

The Conftest adapter invokes a controlled JSON output mode, normalizes valid
findings, and treats compilation, parsing, process, and protocol failures as
evaluator errors. It does not classify results through text matching.

### Evaluation currentness and finding identity

Workflow lifecycle, rather than persisted plan, attempt, evaluator, or policy
fingerprints, determines whether an evaluation is current. A successful new
plan or explicit recheck replaces `plan.after`; every `before` point
reevaluates on the next operation attempt. The complete retirement rules are
defined with the persisted state below.

Atlantis uses one digest for sticky approval: the identity of a normalized
finding.

```text
SHA256(canonical JSON {
  digest schema,
  policy set,
  workflow stage,
  finding type,
  rule_id,
  subject,
  message,
  metadata
})
```

The digest excludes plan and input identity, finding order, evaluator runtime,
policy source and content identity, `details`, and `diagnostic`. These describe
how an evaluation ran rather than the normalized concern it produced.
Decision-bearing information must therefore appear in the included finding
fields. An implementation change that produces the same normalized concern
preserves approval; a change to an included field invalidates it.

Atlantis computes finding digests. An evaluator cannot provide or override
them.

Identical normalized findings have identical digests. Atlantis retains the
sorted digest multiset, including duplicates. Approval coverage uses multiset
containment so an additional identical occurrence is still a new concern.

### Persistence and sticky approvals

New policy state is additive and separate from the legacy fields:

```go
type PolicyFrameworkStatus struct {
	Version     int
	Evaluations []PolicyEvaluation
	Approvals   []PolicyApproval
}

type ProjectStatus struct {
	// Existing project status fields remain.
	PolicyFramework *PolicyFrameworkStatus
}

type PolicyEvaluation struct {
	PolicySet     string
	WorkflowStage WorkflowStage
	Findings      []Finding
	Diagnostic    string
	Error         string
}

type PolicyApproval struct {
	PolicySet     string
	WorkflowStage WorkflowStage
	User          string
	Covered       []string // sorted denial-digest multiset
}
```

Persisting normalized findings is deliberate. Legacy approval state does not
retain enough policy-result content to reconstruct a complete report during an
approval command. The new framework therefore renders approvals and
withdrawals from persisted findings without rerunning the evaluator. Approval
requires a current evaluation, while withdrawal may remove retained sticky
coverage without one. Recheck replaces the current evaluation, and approved
findings remain visible as overridden.

Each project has at most one current evaluation for a policy set and workflow
stage. Workflow transitions, rather than duplicated plan or attempt identifiers
inside policy state, determine whether that evaluation remains current:

- a fresh evaluation replaces its current slot;
- invalidating, replacing, discarding, or consuming a saved plan retires its
  `plan.after` and `apply.before` evaluations;
- a new import or state-removal attempt replaces the corresponding `before`
  evaluation;
- a new pull-request head retires evaluations from the previous head.

A retired evaluation cannot be approved or used for gating. Retirement is part
of the corresponding workflow state transition.

Sticky approval reconciliation is always enabled; it is not a policy-set
option. This is required for `before` gates: retrying an approved operation
must reevaluate current evidence without discarding approval for unchanged
findings. Skipping reevaluation would permit stale evidence, while clearing
approval on every evaluation would create an approval loop.

Approval snapshots the current sorted denial-digest multiset per user. After
reevaluation, an approval counts only when its owner remains eligible and its
snapshot contains the current denial multiset. Removed findings remain covered;
new, changed, or additional findings require renewed approval. Warnings and
errors are not included.

Approval snapshots are retained for the pull request rather than pruned when a
finding disappears. If an unchanged finding later returns, its earlier approval
counts again. Conftest exceptions affect which denial findings are present; they
are diagnostic information, not separate findings. Removing an exception may
therefore restore a previously approved denial. This preserves legacy sticky
approval behavior, and an owner can remove retained coverage explicitly with
`--clear-policy-approval`.

Atlantis currently persists project status as JSON inside both BoltDB and
Redis. Adding `PolicyFrameworkStatus` does not require rewriting existing
records. Missing new fields mean no current framework evaluation or approval.
Legacy approval state is not converted into new approval state; switching a
repository to the new framework requires one fresh evaluation per applicable
project and, if needed, fresh approval.

The persisted numeric values `ErroredPolicyCheckStatus` and
`PassedPolicyCheckStatus` remain decodable for legacy records and must not be
reordered. The new framework never writes them.

Redis currently updates the complete `PullStatus` using a non-atomic
read-modify-write sequence. Concurrent lifecycle and policy updates may
therefore overwrite one another. This is a pre-existing coordination-store
limitation and should be tracked separately; this ADR does not redefine the
storage transaction model.

### Markdown and VCS status

The new framework uses one evaluator-independent policy report composed with
the existing operation result. Go deterministically converts normalized
findings into a presentation model and escapes evaluator-controlled strings as
untrusted plain text. The report groups findings by project, policy set, policy
point, and type; shows messages, optional rule and subject, and approval state;
places details, metadata, and diagnostics in bounded expandable sections; and
distinguishes policy denials from evaluator and operation errors.

Templates perform presentation only. Digesting, approval coverage, and policy
decisions occur before template execution. Conftest-specific templates remain
only on the legacy path.

Approval and withdrawal responses render current persisted findings, when
present, together with their updated policy-set approval state.

The new framework does not publish a separate `policy_check` VCS status.
Policies are intrinsic to the workflow result exposed by its existing status,
while persisted operation and policy facts remain separate.

| Situation | New-framework VCS result |
| --- | --- |
| `plan.before` denied | Plan blocked; no plan was produced |
| Plan workflow failed | Plan failed |
| Plan succeeded but `plan.after` denied or errored | Plan blocked with a saved, ineligible plan and policy detail |
| Current `plan.after` becomes clear through approval or recheck | The saved plan becomes eligible and the plan VCS status becomes successful |
| `apply.before` denied or errored | Apply blocked; the saved plan remains |
| `apply.before` is approved | Apply remains blocked until an explicit retry succeeds |
| Import or state removal is blocked | Report the command result; do not create a durable policy merge blocker |

The renderer states whether Terraform ran and whether a saved plan exists, so a
policy-blocked workflow is not mistaken for a Terraform execution error. The
operation's top-level runner owns its VCS status. Approval and recheck paths
recompute an affected workflow status from persisted state rather than writing
an independent policy status.

### Event and runtime integration

The new policy domain will live outside `server/events`. Event orchestration
will depend on a narrow coordinator rather than Conftest or policy result
parsing:

```go
type PolicyCoordinator interface {
	Evaluate(
		ctx context.Context,
		req StageEvaluationRequest,
	) PolicyReport
}
```

The coordinator selects policy sets for the boundary, resolves inputs, invokes
their adapters, normalizes findings, reconciles approvals, and returns a
complete `PolicyReport`.

At each policy point, Atlantis evaluates every applicable policy set and
returns all results, even when one set denies or errors. The gate clears only
when every set has passed or has sufficient valid approval; any evaluator
error blocks it. Evaluation order is not part of the contract, but rendered
results are deterministic.

`DefaultProjectCommandRunner` invokes the coordinator at the policy points
defined above. Workdir serialization remains held through evaluation and
persistence of the combined operation and policy result.

When a `before` evaluation prevents execution, Atlantis releases workdir
serialization and every project lock acquired for that command. It preserves
only a pre-existing lock associated with a retained saved plan. In particular,
a blocked `plan.before` evaluation cannot leave a project locked without a
plan. Evaluator errors follow the same cleanup path as denials.

Workflow commands support explicit policy-only modes:

```go
type CommandRequest struct {
	Operation Operation
	Mode      CommandMode // execute, approve_policies, clear_policy_approval, or recheck_policies
}
```

The mode is selected before operation authorization, requirements, locks,
hooks, or workflow construction. `--approve-policies` and
`--clear-policy-approval` use policy-approval authorization rather than
authorization for the named operation and modify only approval state. The
three policy-only modes are mutually exclusive and accept project selectors
and `--policy-set`.

`ProjectCommandOutput` will gain a composable policy report alongside its plan,
apply, import, or state-removal result:

```go
type ProjectCommandOutput struct {
	// Existing operation-specific fields remain.
	PolicyReport *PolicyReport
}
```

The database updater must merge these two result dimensions independently. A
policy report from a blocked `before` evaluation updates policy state without
deriving or overwriting the project's plan/apply status. A successful operation
updates its existing lifecycle status independently. This replaces the current
assumption that every project-command result maps to exactly one project
lifecycle status.

### Pull-request recovery and retry behavior

Policy state does not replace plan/apply state. The apply flow is:

```text
select unapplied plans -> evaluate apply.before -> apply
                              | denied or error
                              v
                    retain plan and stop attempt
                              |
                     approve or mitigate
                              |
                    explicitly retry apply
```

Approval or mitigation never resumes an operation. A new operation command
reevaluates its `before` policies before running workflow steps. An apply retry
retains the saved plan and selects only unapplied projects; completed projects
are neither rolled back nor reapplied.

The recovery action depends on the policy point and result:

| Result | Recovery |
| --- | --- |
| Any `before` denial | Approve through the corresponding operation or mitigate, then explicitly retry that operation |
| `before` evaluator or protocol error | Repair the evaluator, policy source, or input, then explicitly retry; errors cannot be approved |
| `plan.after` denial | Run `plan --approve-policies`, or mitigate and recheck; replan if plan inputs changed |
| `plan.after` evaluator or protocol error | Repair and recheck the saved plan, or replan if it is stale; errors cannot be approved |
| Stale or missing plan evidence | Replan; approval cannot make stale evidence current |

`atlantis plan --recheck-policies` reevaluates `plan.after` policy sets against
current saved plans without generating a new Terraform plan. It uses normal
plan project selectors, preserves the plan, and updates findings, sticky
approvals, VCS status, and apply eligibility. It does not run plan
requirements, `plan.before`, workflow steps, Terraform commands, or any other
operation. It reads the existing `$SHOWFILE`; if the saved plan or required
evidence is stale or missing, the command requires a new plan.

New-framework approvals use the protected workflow as their natural scope:

```text
atlantis plan --approve-policies
atlantis apply --approve-policies
atlantis import --approve-policies
atlantis state rm --approve-policies
```

These policy-only modes approve every current denial in their workflow scope
for which the caller is eligible. Existing project selectors narrow the
projects. An optional `--policy-set <name>` selector further narrows approval;
omitting it selects all applicable policy sets. `plan --approve-policies`
covers current `plan.before` and `plan.after` findings. The other commands cover
their respective `before` point.

Approvers withdraw approval through the equivalent workflow-scoped mode:

```text
atlantis plan --clear-policy-approval
atlantis apply --clear-policy-approval
atlantis import --clear-policy-approval
atlantis state rm --clear-policy-approval
```

This removes only the caller's approval snapshots for the selected projects,
policy sets, and policy points, including sticky coverage retained without a
current denial. Other users' approvals remain intact. Atlantis then recomputes
workflow eligibility and the affected VCS status from the remaining state.

`--policy-set` also narrows `plan --recheck-policies` to the named policy set at
`plan.after`, and narrows approval withdrawal in the same way as approval.
The authorization, self-approval, approval-count, and sticky-coverage rules are
the same at every policy point.

Resolving a `plan.after` denial through approval or recheck may make the saved
plan apply-eligible without rerunning Terraform.

## Open questions

### Policy-point terminology

Should the configuration and protocol use `policy_points` and `policy_point`
instead of `workflow_stages` and `workflow_stage`? Values such as
`plan.before` identify boundaries around a workflow, while Atlantis already
uses `Stage` for configured step sequences.

See [Configuration](#configuration) and
[Evaluator contract](#evaluator-contract).

### Policy state limits

Structured findings may substantially increase the per-PR `PullStatus`,
particularly in monorepos. What limits should apply per evaluation, project,
and pull request? Exceeding a limit on decision-bearing findings or metadata
must fail the evaluation rather than silently omit evidence; display-only
details and diagnostics may be truncated. Defaults should be informed by
representative Conftest results.

See [Persistence and sticky approvals](#persistence-and-sticky-approvals).

### Legacy migration

Should the new framework replace legacy policy checking, or coexist with it
during migration?

If coexistence is retained, both implementations may be enabled globally, but a
repository must use exactly one. Their configuration, evaluations, approvals,
commands, and VCS statuses are never combined. Legacy behavior should remain
behind a compatibility boundary and may share low-level Conftest acquisition
and process execution with the new adapter, but not normalization, persistence,
approval, or rendering.

Replacing legacy policy checking would remove `policies`, `policy_check`,
`custom_policy_check`, and the standalone `approve_policies` path. This produces
a smaller implementation but requires users to migrate their configuration
immediately.

Existing approvals need not be migrated in either case. The choice does not
affect the new evaluator, finding, approval, persistence, or rendering
contracts, but must be resolved before implementation.

See [Configuration](#configuration),
[Persistence and sticky approvals](#persistence-and-sticky-approvals), and
[Event and runtime integration](#event-and-runtime-integration).

## Consequences

| Benefits | Costs |
| --- | --- |
| Conftest becomes a peer adapter, additional built-ins fit laterally, and arbitrary tools gain simple exit-code and JSON contracts | Operation runners, rendering, and persistence gain a composable policy report |
| One policy set can protect several workflow stages | If retained, both policy implementations coexist during migration |
| Sticky approvals use normalized findings instead of rendered output | New Conftest normalization requires compatibility tests |
| Local or Git policy sources can be shared across evaluators | Atlantis owns Git source resolution and credentials |
| Events and Markdown no longer parse evaluator formats | The existing Redis `PullStatus` lost-update risk remains separate storage work |
| Workflow-scoped approval avoids ambiguous cross-workflow targeting | Command parsing and authorization must distinguish execution from policy-only modes |
| Abandoned administrative commands do not become durable merge blockers | New-framework branch protection uses workflow results rather than `policy_check` |

## Alternatives considered

| Alternative | Reason rejected |
| --- | --- |
| Use Conftest JSON as the domain contract | It lacks canonical finding identity and does not generalize to other engines |
| SARIF | It is larger than needed and still requires engine-specific normalization |
| Unspecified human-readable custom output | Without an explicit exit-code contract, text and status cannot reliably separate denial from evaluator failure |
| Existing workflow hooks or custom `run` steps | They can execute policy tools but do not provide normalized findings, policy ownership, sticky approvals, persisted decisions, or intrinsic per-project workflow gating |
| Let each evaluator fetch policies | It duplicates source acquisition, credentials, caching, and materialization behavior across adapters |
| Named evaluator instances | One shared runtime per built-in evaluator type is sufficient initially; inline commands cover custom runtimes |
| Policy between arbitrary steps | Arbitrary steps may mutate state, so intermediate gates are unsafe |
| Post-mutation policy points | They cannot prevent a completed mutation and require a separate audit and recovery model |
| One policy VCS status across workflows | Attempt-local failures could outlive abandoned commands and block unrelated PR progress |
| One VCS status per policy point | It creates missing and noisy branch-protection contexts |
| Standalone new-framework `approve_policies` | Workflow-scoped modes provide a natural target without a stage selector |
| Standalone recheck for `before` policies | Only an operation attempt can authoritatively evaluate its current execution context |
| Persist plan or attempt IDs in policy state | Workflow lifecycle transitions can retire dependent evaluations directly |
| Hash the whole plan or output | Unrelated changes would invalidate sticky approvals |
| Definition or evaluation fingerprints in policy state | Workflow lifecycle identifies current evaluations; normalized finding identity is sufficient for sticky approval |
| Change legacy types in place | Additive state provides safer decoding and migration |

## References

- [Atlantis policy checking](https://www.runatlantis.io/docs/policy-checking)
- [Atlantis custom policy checks](https://www.runatlantis.io/docs/custom-policy-checks)
- [Conftest result types](https://github.com/open-policy-agent/conftest/blob/v0.68.2/output/result.go)
- [Conftest JSON output](https://github.com/open-policy-agent/conftest/blob/v0.68.2/output/json.go)
- [Conftest metadata](https://www.conftest.dev/#metadata)
- [Checkov report model](https://github.com/bridgecrewio/checkov/blob/main/checkov/common/output/report.py)
- [Checkov record model](https://github.com/bridgecrewio/checkov/blob/main/checkov/common/output/record.py)
- [OPA REST API](https://www.openpolicyagent.org/docs/rest-api)
- [cnspec check-result schema](https://mondoo.com/docs/reporting/export/schema/check)
- [Sentinel JSON tracing](https://developer.hashicorp.com/sentinel/docs/writing/tracing#json-output)
- [Sticky policy approvals PR #6271](https://github.com/runatlantis/atlantis/pull/6271)
- [ADR convention](https://github.com/runatlantis/atlantis/blob/main/docs/adr/0001-record-architecture-decisions.md)
