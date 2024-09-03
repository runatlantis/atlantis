# Repo and Project Permissions

Sometimes it may be necessary to limit who can run which commands, such as
restricting who can apply changes to production, while allowing more
freedom for dev and test environments.

## Authorization Workflow

Atlantis performs two authorization checks to verify a user has the necessary
permissions to run a command:

1. After a command has been validated, before var files, repo metadata, or
   pull request statuses are checked and validated.
2. After pre workflow hooks have run, repo configuration processed, and 
   affected projects determined.

::: tip Note
The first check should be considered as validating the user for a repository
as a whole, while the second check is for validating a user for a specific
project in that repo.
:::

### Why check permissions twice?
The way Atlantis is currently designed, not all relevant information may be
available when the first check happens.  In particular, affected projects
are not known because pre workflow hooks haven't run yet, so repositories
that use hooks to generate or modify repo configurations won't know which
projects to check permissions for.

## Configuring permissions

Atlantis has two options for allowing instance administrators to configure
permissions.

### Server option [`--gh-team-allowlist`](server-configuration.md#gh-team-allowlist)

The `--gh-team-allowlist` option allows administrators to configure a global
set of permissions that apply to all repositories.  For most use cases, this
should be sufficient.

### External command

For administrators that require more granular and specific permission
definitions, an external command can be defined in the [server side repo
configuration](server-side-repo-config.md#teamauthz).  This command will receive
information about the command, repo, project, and GitHub teams the user is a
member of, allowing administrators to integrate the permissions validation
with other systems or business requirements.  An example would be allowing
users to apply changes to lower environments like dev and test environments
while restricting changes to production or other sensitive environments.

::: warning
These options are mutually exclusive.  If an external command is defined,
the `--gh-team-allowlist` option is ignored.
:::

## Example

### Restrict production changes
This example shows a simple example of how a script could be used to restrict
production changes to a specific team, while allowing anyone to work on other 
environments.  For brevity, this example assumes each user is a member of a
single team.

`server-side-repo-config.yaml`
```yaml
team_authz:
  command: "/scripts/example.sh"
```

`example.sh`
```shell
#!/bin/bash

# Define name of team allowed to make production changes
PROD_TEAM="example-org/prod-deployers"

# Set variables from command-line arguments for convenience
COMMAND="$1"
REPO="$2"
TEAM="$3"

# Check if we are running the 'apply' command on prod
if [ "${COMMAND}" == "apply" -a "${PROJECT_NAME}" == "prod" ]
then
   # Only the prod team can make this change
   if [ "${TEAM}" == "${PROD_TEAM}" ]
   then
      echo "pass"
      exit 0
   fi
   
   # Print reason for failing and exit
   echo "user \"${USER_NAME}\" must be a member of \"${PROD_TEAM}\" to apply changes to production."
   exit 0
fi 

# Any other command and environment is okay
echo "pass"
exit 0
```
## Reference

### External Command Execution

External commands are executed on every authorization check with arguments and
environment variables containing context about the command being checked. The
command is executed using the following format:

```shell
external_command [external_args...] atlantis_command repo [teams...]
```

| Key                | Optional | Description                                                                               |
|--------------------|----------|-------------------------------------------------------------------------------------------|
| `external_command` | no       | Command defined in [server side repo configuration](server-side-repo-config.md)           |
| `external_args`    | yes      | Command arguments defined in [server side repo configuration](server-side-repo-config.md) |
| `atlantis_command` | no       | The atlantis command being run (`plan`, `apply`, etc)                                     |
| `repo`             | no       | The full name of the repo being executed (format: `owner/repo_name`)                      |
| `teams`            | yes      | A list of zero or more teams of the user executing the command                            | 


The following environment variables are passed to the command on every execution:

| Key                  | Description                                                                                                                                                                                                                           |
|----------------------|---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `BASE_REPO_NAME`     | Name of the repository that the pull request will be merged into, ex. `atlantis`.                                                                                                                                                     |
| `BASE_REPO_OWNER`    | Owner of the repository that the pull request will be merged into, ex. `runatlantis`.                                                                                                                                                 |
| `COMMAND_NAME`       | The name of the command that is being executed, i.e. `plan`, `apply` etc.                                                                                                                                                             |
| `USER_NAME`          | Username of the VCS user running command, ex. `acme-user`. During an autoplan, the user will be the Atlantis API user, ex. `atlantis`.                                                                                                |


The following environment variables are also passed to the command when checking project authorization:

| Key                  | Description                                                                                                                                                                                                                           |
|----------------------|---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `BASE_BRANCH_NAME`   | Name of the base branch of the pull request (the branch that the pull request is getting merged into)                                                                                                                                 |
| `COMMENT_ARGS`       | Any additional flags passed in the comment on the pull request. Flags are separated by commas and every character is escaped, ex. `atlantis plan -- arg1 arg2` will result in `COMMENT_ARGS=\a\r\g\1,\a\r\g\2`.                       |
| `HEAD_REPO_NAME`     | Name of the repository that is getting merged into the base repository, ex. `atlantis`.                                                                                                                                               |
| `HEAD_REPO_OWNER`    | Owner of the repository that is getting merged into the base repository, ex. `acme-corp`.                                                                                                                                             |
| `HEAD_BRANCH_NAME`   | Name of the head branch of the pull request (the branch that is getting merged into the base)                                                                                                                                         |
| `HEAD_COMMIT`        | The sha256 that points to the head of the branch that is being pull requested into the base. If the pull request is from Bitbucket Cloud the string will only be 12 characters long because Bitbucket Cloud truncates its commit IDs. |
| `PROJECT_NAME`       | Name of the project the command is being executed on                                                                                                                                                                                  |
| `PULL_NUM`           | Pull request number or ID, ex. `2`.                                                                                                                                                                                                   |
| `PULL_URL`           | Pull request URL, ex. `https://github.com/runatlantis/atlantis/pull/2`.                                                                                                                                                               |
| `PULL_AUTHOR`        | Username of the pull request author, ex. `acme-user`.                                                                                                                                                                                 |
| `REPO_ROOT`          | The absolute path to the root of the cloned repository.                                                                                                                                                                               |
| `REPO_REL_PATH`      | Path to the project relative to `REPO_ROOT`                                                                                                                                                                                           |

### External Command Result Handling

Atlantis determines if a user is authorized to run the requested command by 
checking if the external command exited with code `0` and if the last line
of output is `pass`.

```
# Psuedo-code of Atlantis evaluation of external commands

user_authorized =
  external_command.exit_code == 0
  && external_command.output.last_line == 'pass' 
```

::: tip
* A non-zero exit code means the command failed to evaluate the request for
some reason (bad configuration, missing dependencies, solar flares, etc).
* If the command was able to run successfully, but determined the user is not
authorized, it should still exit with code `0`.
   * The command output could contain the reasoning for the authorization failure.
:::
