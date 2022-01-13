# Post Workflow Hooks

Post workflow hooks can be defined to run scripts after default or custom
workflows are executed. Post workflow hooks differ from [custom
workflows](custom-workflows.html#custom-run-command) in that they are run
outside of Atlantis commands. Which means they do not surface their output
back to the PR as a comment.

[[toc]]

## Usage

Post workflow hooks can only be specified in the Server-Side Repo Config under
`repos` key.

## Use Cases

### Cost estimation reporting

You can add a post workflow hook to perform custom reporting after all workflows
have finished.

In this example we use a custom workflow to generate cost estimates for each
workflow using [Infracost](https://www.infracost.io/docs/integrations/atlantis/), then create a summary report after all workflows have completed.


```yaml
# repos.yaml
workflows:
  myworkflow:
    plan:
      steps:
      - init
      - plan
      - run: infracost breakdown --path=$PLANFILE --format=json --out-file=/tmp/$BASE_REPO_OWNER-$BASE_REPO_NAME-$PULL_NUM-$WORKSPACE-$REPO_REL_DIR-infracost.json
repos:
  - id: /.*/
    workflow: myworkflow
    post_workflow_hooks:
      - run: infracost output --path=/tmp/$BASE_REPO_OWNER-$BASE_REPO_NAME-$PULL_NUM-*-infracost.json --format=github-comment --out-file=/tmp/infracost-comment.md
      # Now report the output as desired, e.g. post to GitHub as a comment.
      # ...
```

## Reference

### Custom `run` Command

This is very similar to [custom workflow run
command](custom-workflows.html#custom-run-command).

```yaml
- run: custom-command
```

| Key | Type   | Default | Required | Description          |
| --- | ------ | ------- | -------- | -------------------- |
| run | string | none    | no       | Run a custom command |

::: tip Notes
* `run` commands are executed with the following environment variables:
  * `BASE_REPO_NAME` - Name of the repository that the pull request will be merged into, ex. `atlantis`.
  * `BASE_REPO_OWNER` - Owner of the repository that the pull request will be merged into, ex. `runatlantis`.
  * `HEAD_REPO_NAME` - Name of the repository that is getting merged into the base repository, ex. `atlantis`.
  * `HEAD_REPO_OWNER` - Owner of the repository that is getting merged into the base repository, ex. `acme-corp`.
  * `HEAD_BRANCH_NAME` - Name of the head branch of the pull request (the branch that is getting merged into the base)
  * `HEAD_COMMIT` - The sha256 that points to the head of the branch that is being pull requested into the base. If the pull request is from Bitbucket Cloud the string will only be 12 characters long because Bitbucket Cloud truncates its commit IDs.
  * `BASE_BRANCH_NAME` - Name of the base branch of the pull request (the branch that the pull request is getting merged into)
  * `PULL_NUM` - Pull request number or ID, ex. `2`.
  * `PULL_AUTHOR` - Username of the pull request author, ex. `acme-user`.
  * `DIR` - The absolute path to the root of the cloned repository.
  * `USER_NAME` - Username of the VCS user running command, ex. `acme-user`. During an autoplan, the user will be the Atlantis API user, ex. `atlantis`.
:::
