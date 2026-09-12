# Locking

When `plan` is run, the directory and Terraform workspace are **Locked** until the pull request is merged or closed, or the plan is manually deleted.

If another user attempts to `plan` for the same directory and workspace in a different pull request
they'll see this error:

![Lock Comment](./images/lock-comment.png)

Which links them to the pull request that holds the lock.

::: warning NOTE
Only the directory in the repo and Terraform workspace are locked, not the whole repo.
:::

Atlantis also checks the global apply lock before running `atlantis apply`. If Atlantis cannot reach the lock backend while checking that global lock, it fails closed and rejects the apply until the backend is reachable again.

## Why

1. Because `atlantis apply` is being done before the pull request is merged, after
an apply your `main` branch does not represent the most up-to-date version of your infrastructure
anymore. With locking, you can ensure that no other changes will be made until the
pull request is merged.

::: tip Why not apply on merge?
Sometimes `terraform apply` fails. If the apply were to fail after the pull
request was merged, you would need to create a new pull request to fix it.
With locking + applying on the branch, you effectively mimic merging to main
but with the added ability to re-plan/apply multiple times if things don't work.
:::
2. If there is already a `plan` in progress, other users won't see a plan that
will be made invalid after the in-progress plan is applied.

## Viewing Locks

To view locks, go to the URL that Atlantis is hosted at:

![Locks View](./images/locks-ui.png)

You can click on a lock to view its details:

<p align="center">
    <img src="./images/lock-detail-ui.png" alt="Lock Detail View" height="400px">
</p>

## Unlocking

The project and workspace will be automatically unlocked when the PR is merged or closed.

To unlock the project and workspace without completing an `apply` and merging, comment `atlantis unlock` on the PR,
or click the link at the bottom of the plan comment to discard the plan and delete the lock where
it says **"To discard this plan click here"**:

![Locks View](./images/lock-delete-comment.png)

The link will take you to the lock detail view where you can click **Discard Plan and Unlock**
to delete the lock.

<p align="center">
    <img src="./images/lock-detail-ui.png" alt="Lock Detail View" height="400px">
</p>

Once a plan is discarded, you'll need to run `plan` again prior to running `apply` when you go back to that pull request.

## Locking All Projects Before Planning

By default Atlantis locks each project immediately before it plans that project,
so a run over many projects interleaves locking and planning:

```plain
lock project A -> plan project A -> lock project B -> plan project B
```

On busy repositories with large pull requests, such as a provider version bump
touching every project, this leaves a window open. Another pull request can
take the lock for a project that has not been reached yet, halfway through a long
run. Atlantis then fails on that lock and discards the plans it had already
produced.

Setting [`repo_locks: {mode: on_apply}`](repo-level-atlantis-yaml.md#repolocks)
does not solve this: it removes locking from planning altogether, so two pull
requests can plan the same project at the same time and only discover they
disagree when one of them applies. `--lock-all-projects-before-plan` keeps the
default guarantee that only one pull request can be planning a given project at
a time. It only changes _when_ the lock for each project is taken, not
whether one is taken.

Starting Atlantis with
[`--lock-all-projects-before-plan`](server-configuration.md#-lock-all-projects-before-plan)
acquires every lock up front instead:

```plain
lock project A -> lock project B -> plan project A -> plan project B
```

If any lock cannot be acquired, no plans are run at all and the locks this run
already took are released again, so a competing pull request is never blocked by
a run that gave up. The pull request comment names the project that was blocked
and who holds its lock, exactly as it does today.

Notes:

* Projects configured with `repo_locks: {mode: on_apply}` or `mode: disabled` are
  not pre-locked.
* Projects that end up producing no plan have their locks released when the
  run finishes, so pre-locking never leaves a project locked with no plan to
  apply. This covers a run stopped with `atlantis cancel`, and an earlier
  execution order group that failed with `abort_on_execution_order_fail`.
* If the Atlantis server itself dies mid-run, the pre-acquired locks stay behind
  until the pull request is closed or someone runs `atlantis unlock`. That is the
  same recovery path as any other interrupted run, but pre-locking makes it
  affect more projects at once.

## Relationship to Terraform State Locking

Atlantis does not conflict with [Terraform State Locking](https://developer.hashicorp.com/terraform/language/state/locking). Under the hood, all
Atlantis is doing is running `terraform plan` and `apply` and so all of the
locking built in to those commands by Terraform isn't affected.

In more detail, Terraform state locking locks the state while you run `terraform apply`
so that multiple applies can't run concurrently. Atlantis's locking is at a higher
level because it prevents multiple pull requests from working on the same state.

## Locking and Drift Detection

When drift detection or remediation runs via the [API](api-endpoints.md) with `PR: 0` (non-PR workflow), Atlantis still acquires and releases working directory locks to prevent concurrent operations on the same project. However, since these operations are not associated with a pull request, they do not create PR-level locks visible in the Locks UI. The working directory lock is released automatically after each drift detection or remediation operation completes.
