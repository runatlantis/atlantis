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

## How to retry a failed automerge

Some VCS providers can briefly reject a merge right after Atlantis applies. With
GitHub branch protection or repository rulesets, the required status checks
(including `atlantis/apply`) are not always registered as passing by the time
Atlantis calls the merge API, so the merge fails with an error such as
`required status checks are expected`. These failures are transient and clear
within a few seconds.

Set the `--automerge-retry-count` server flag (or `ATLANTIS_AUTOMERGE_RETRY_COUNT`
environment variable) to retry the merge, with an exponential backoff, before
giving up.

```shell
atlantis server --automerge-retry-count 3
```

Defaults to `0`, which attempts the merge exactly once.

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
