# Pre Workflow Hooks

Pre workflow hooks can be defined to run scripts before default or custom
workflows are executed. Pre workflow hooks differ from [custom
workflows](custom-workflows.html#custom-run-command) in several ways.

1. Pre workflow hooks do not require the repository configuration to be
   present. This can be utilized to [dynamically generate repo configs](pre-workflow-hooks.html#dynamic-repo-config-generation).
2. Pre workflow hooks are run outside of Atlantis commands. Which means
   they do not surface their output back to the PR as a comment.

[[toc]]

## Usage

Pre workflow hooks can only be specified in the Server-Side Repo Config under the
`repos` key.
::: tip Note
`pre-workflow-hooks` do not prevent Atlantis from executing its
workflows(`plan`, `apply`) even if a `run` command exits with an error.
::: 

## Use Cases

### Dynamic Repo Config Generation

To generate the repo `atlantis.yaml` before Atlantis can parse it,
add a `run` command to `pre_workflow_hooks`. Your Repo config will be generated
right before Atlantis parses it.

```yaml
repos:
    - id: /.*/
      pre_workflow_hooks:
        - run: ./repo-config-generator.sh
          description: Generating configs
```

## Customizing the Shell

By default, the command will be run using the 'sh' shell with an argument of '-c'. This
can be customized using the `shell` and `shellArgs` keys.

Example:

```yaml
repos:
    - id: /.*/
      pre_workflow_hooks:
        - run: |
            echo "generating atlantis.yaml"
            terragrunt-atlantis-config generate --output atlantis.yaml --autoplan --parallel
          description: Generating atlantis.yaml
          shell: bash
          shellArgs: -cv
```

## Reference

### Custom `run` Command

This is very similar to the [custom workflow run
command](custom-workflows.html#custom-run-command).

```yaml
- run: custom-command
```

| Key         | Type   | Default | Required | Description          |
| ----------- | ------ | ------- | -------- | -------------------- |
| run         | string | none    | no       | Run a custom command |
| description | string | none    | no       | Pre hook description |
| shell       | string | 'sh'    | no       | The shell to use for running the command |
| shellArgs   | string | '-c'    | no       | The shell arguments to use for running the command |

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
  * `PULL_URL` - Pull request URL, ex. `https://github.com/runatlantis/atlantis/pull/2`.
  * `PULL_AUTHOR` - Username of the pull request author, ex. `acme-user`.
  * `DIR` - The absolute path to the root of the cloned repository. 
  * `USER_NAME` - Username of the VCS user running command, ex. `acme-user`. During an autoplan, the user will be the Atlantis API user, ex. `atlantis`.
  * `COMMENT_ARGS` - Any additional flags passed in the comment on the pull request. Flags are separated by commas and
      every character is escaped, ex. `atlantis plan -- arg1 arg2` will result in `COMMENT_ARGS=\a\r\g\1,\a\r\g\2`.
  * `COMMAND_NAME` - The name of the command that is being executed, i.e. `plan`, `apply` etc.
  * `OUTPUT_STATUS_FILE` - An output file to customize the success or failure status. ex. `echo 'failure' > $OUTPUT_STATUS_FILE`.
:::
