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

## How to set merge method for automerge

If automerge is enabled, you can use `--merge-method` option
for `atlantis apply` command to specify which merge method use.

```shell
atlantis apply --merge-method squash
```

Implemented only for GitHub. You can choose one of them:
- merge
- rebase
- squash

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
