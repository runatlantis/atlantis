# Automerging

Atlantis can be configured to automatically merge a pull request after all plans have
been successfully applied.

![Automerge](./images/automerge.png)

## How To Enable

Automerging can be enabled either by:

1. Passing the `--automerge` flag to `atlantis server`. This sets the parameter globally; however, explicit declaration in the repo config will be respected and take priority.
1. Setting `automerge: true` in the repo's `atlantis.yaml` file:

    ```yaml
    version: 3
    automerge: true
    projects:
    - dir: .
    ```

    :::tip NOTE
    If a repo has an `atlantis.yaml` file, then each project in the repo needs
    to be configured under the `projects` key.
    :::

## How to Disable

If automerge is enabled, you can disable it for a single `atlantis apply`
command with the `--auto-merge-disabled` option.

## How to set the merge method for automerge

If automerge is enabled, you can set a default merge method with the
`--automerge-method` server flag or `ATLANTIS_AUTOMERGE_METHOD` environment
variable.

```shell
atlantis server --automerge-method <method>
```

You can override the server default for a single `atlantis apply` command with
the `--auto-merge-method` option.

```shell
atlantis apply --auto-merge-method <method>
```

The `method` must be one of:

- merge
- rebase
- squash

This is currently only implemented for the GitHub VCS.

## Requirements

### All Plans Must Succeed

When automerge is enabled, **all plans** in a pull request **must succeed** before
**any** plans can be applied.

For example, imagine this scenario:

1. I open a pull request that makes changes to two Terraform projects, in `dir1/`
   and `dir2/`.
1. The plan for `dir2/` fails because my Terraform syntax is wrong.

In this scenario, I can't run

```shell
atlantis apply -d dir1
```

Even though that plan succeeded, because **all** plans must succeed for **any** plans
to be saved.

Once I fix the issue in `dir2`, I can push a new commit which will trigger an
autoplan. Then I will be able to apply both plans.

### All Plans must be applied

If multiple projects/dirs/workspaces are configured to be planned automatically,
then they should all be applied before Atlantis automatically merges the PR.

## Permissions

The Atlantis VCS user must have the ability to merge pull requests.

## GitHub Merge Queue

If the PR's base branch requires a [merge queue](https://docs.github.com/en/repositories/configuring-branches-and-merges-in-your-repository/configuring-pull-request-merges/managing-a-merge-queue), Atlantis cannot merge the PR directly via the REST API (GitHub returns `405 Method Not Allowed`). When [`--gh-merge-queue-enabled`](server-configuration.md#--gh-merge-queue-enabled) is set, Atlantis routes automerge through the GraphQL `enablePullRequestAutoMerge` mutation, which adds the PR to the merge queue. GitHub then runs the queue's required checks against the merge candidate commit and merges the PR once they pass.

Detection runs in two layers:

- A pre-flight GraphQL query reads both `branchProtectionRule.requiresMergeQueue` (classic branch protection) and the base ref's rulesets, looking for an active `MERGE_QUEUE` rule. Either signal short-circuits to `enablePullRequestAutoMerge`.
- If the direct REST merge later returns `405` with a merge-queue indication, Atlantis falls back to `enablePullRequestAutoMerge`. This is a safety net for edge cases such as a token without ruleset read scope or repos with more than 100 ruleset rules on the base ref.

For this to work end-to-end:

- Set [`--gh-merge-queue-enabled`](server-configuration.md#--gh-merge-queue-enabled) so Atlantis (a) posts `success` for `plan`/`apply`/`policy_check` on `merge_group` events, and (b) takes the queue-aware automerge path. With the flag off, Atlantis uses direct REST merge for all repos and surfaces a clear error if the base branch enforces a queue.
- Subscribe the GitHub webhook to the `merge_group` event (see [Configuring Webhooks](configuring-webhooks.md#github)).
- The Atlantis VCS user / GitHub App must have permission to enable auto-merge on the repo.

The `--auto-merge-method` setting is honored for non-queue branches; for branches enforcing a merge queue GitHub uses the queue's configured method and ignores any per-PR override.
