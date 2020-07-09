# Custom Workflows

Custom workflows can be defined to override the default commands that Atlantis
runs.

[[toc]]

## Usage
Custom workflows can be specified in the Server-Side Repo Config or in the Repo-Level
`atlantis.yaml` files.

**Notes**
* If you want to allow repos to select their own workflows, they must have the
`allowed_overrides: [workflow]` setting. See [server-side repo config use cases](server-side-repo-config.html#allow-repos-to-choose-a-server-side-workflow) for more details.
* If in addition you also want to allow repos to define their own workflows, they must have the
`allow_custom_workflows: true` setting. See [server-side repo config use cases](server-side-repo-config.html#allow-repos-to-define-their-own-workflows) for more details.


## Use Cases
### .tfvars files
Given the structure:
```
.
└── project1
    ├── main.tf
    ├── production.tfvars
    └── staging.tfvars
```

If you wanted Atlantis to automatically run plan with `-var-file staging.tfvars` and `-var-file production.tfvars`
you could define two workflows:
```yaml
# repos.yaml or atlantis.yaml
workflows:
  staging:
    plan:
      steps:
      - init
      - plan:
          extra_args: ["-var-file", "staging.tfvars"]
    # NOTE: no need to define the apply stage because it will default
    # to the normal apply stage.
    
  production:
    plan:
      steps:
      - init
      - plan:
          extra_args: ["-var-file", "production.tfvars"]
```
Then in your repo-level `atlantis.yaml` file, you would reference the workflows:
```yaml
# atlantis.yaml
version: 3
projects:
# If two or more projects have the same dir and workspace, they must also have
# a 'name' key to differentiate them.
- name: project1-staging
  dir: project1
  workflow: staging
- name: project1-production
  dir: project1
  workflow: production

workflows:
  # If you didn't define the workflows in your server-side repos.yaml config,
  # you would define them here instead.
```
When you want to apply the plans, you can comment
```
atlantis apply -p project1-staging
```
and
```
atlantis apply -p project1-production
```
Where `-p` refers to the project name.

### Adding extra arguments to Terraform commands
If you need to append flags to `terraform plan` or `apply` temporarily, you can
append flags on a comment following `--`, for example commenting:
```
atlantis plan -- -lock=false
```

If you always need to do this for a project's `init`, `plan` or `apply` commands
then you must define a custom workflow and set the `extra_args` key for the
command you need to modify.

```yaml
# atlantis.yaml or repos.yaml
workflows:
  myworkflow:
    plan:
      steps:
      - init:
          extra_args: ["-lock=false"]
      - plan:
          extra_args: ["-lock=false"]
    apply:
      steps:
      - apply:
          extra_args: ["-lock=false"]
```

### Custom init/plan/apply Commands
If you want to customize `terraform init`, `plan` or `apply` in ways that
aren't supported by `extra_args`, you can completely override those commands.

In this example, we're not using any of the built-in commands and are instead
using our own.

```yaml
# atlantis.yaml or repos.yaml
workflows:
  myworkflow:
    plan:
      steps:
      - run: terraform init -input=false -no-color
      
      # If you're using workspaces you need to select the workspace using the
      # $WORKSPACE environment variable.
      - run: terraform workspace select -no-color $WORKSPACE
      
      # You MUST output the plan using -out $PLANFILE because Atlantis expects
      # plans to be in a specific location.
      - run: terraform plan -input=false -refresh -no-color -out $PLANFILE
    apply:
      steps:
      # Again, you must use the $PLANFILE environment variable.
      - run: terraform apply -no-color $PLANFILE
```

### Terragrunt
Atlantis supports running custom commands in place of the default Atlantis
commands. We can use this functionality to enable
[Terragrunt](https://github.com/gruntwork-io/terragrunt).

You can either use your repo's `atlantis.yaml` file or the Atlantis server's `repos.yaml` file.

Given a directory structure:
```
.
└── live
    ├── prod
    │   └── terragrunt.hcl
    └── staging
        └── terragrunt.hcl
```

If using the server `repos.yaml` file, you would use the following config:

```yaml
# repos.yaml
# Specify TERRAGRUNT_TFPATH environment variable to accomodate setting --default-tf-version
repos:
- id: "/.*/"
  workflow: terragrunt
workflows:
  terragrunt:
    plan:
      steps:
      - env:
          name: TERRAGRUNT_TFPATH
          command: 'echo "terraform${ATLANTIS_TERRAFORM_VERSION}"'
      - run: terragrunt plan -no-color -out=$PLANFILE
    apply:
      steps:
      - env:
          name: TERRAGRUNT_TFPATH
          command: 'echo "terraform${ATLANTIS_TERRAFORM_VERSION}"'
      - run: terragrunt apply -no-color $PLANFILE
```

If using the repo's `atlantis.yaml` file you would use the following config:
```yaml
version: 3
projects:
- dir: live/staging
  workflow: terragrunt
- dir: live/prod
  workflow: terragrunt
workflows:
  terragrunt:
    plan:
      steps:
      - env:
          name: TERRAGRUNT_TFPATH
          command: 'echo "terraform${ATLANTIS_TERRAFORM_VERSION}"'
      - run: terragrunt plan -no-color -out $PLANFILE
    apply:
      steps:
      - env:
          name: TERRAGRUNT_TFPATH
          command: 'echo "terraform${ATLANTIS_TERRAFORM_VERSION}"'
      - run: terragrunt apply -no-color $PLANFILE
```

**NOTE:** If using the repo's `atlantis.yaml` file, you will need to specify each directory that is a Terragrunt project.


::: warning
Atlantis will need to have the `terragrunt` binary in its PATH.
If you're using Docker you can build your own image, see [Customization](/docs/deployment.html#customization).
:::

If you don't want to create/manage the repo's `atlantis.yaml` file yourself, you can use the tool [terragrunt-atlantis-config](https://github.com/transcend-io/terragrunt-atlantis-config) to generate it.

The `terragrunt-atlantis-config` tool is a community project and not maintained by the Atlantis team.

### Running custom commands
Atlantis supports running completely custom commands. In this example, we want to run
a script after every `apply`:

```yaml
# repos.yaml or atlantis.yaml
workflows:
  myworkflow:
    apply:
      steps:
      - apply
      - run: ./my-custom-script.sh
```

::: tip Notes
* We don't need to write a `plan` key under `myworkflow`. If `plan`
isn't set, Atlantis will use the default plan workflow which is what we want in this case.
* A custom command will only terminate if all output file descriptors are closed.
Therefore a custom command can only be sent to the background (e.g. for an SSH tunnel during
the terraform run) when its output is redirected to a different location. For example, Atlantis
will execute a custom script containing the following code to create a SSH tunnel correctly: 
`ssh -f -M -S /tmp/ssh_tunnel -L 3306:database:3306 -N bastion 1>/dev/null 2>&1`. Without
the redirect, the script would block the Atlantis workflow.
:::

### Custom Backend Config
If you need to specify the `-backend-config` flag to `terraform init` you'll need to use a custom workflow.
In this example, we're using custom backend files to configure two remote states, one for each environment.
We're then using `.tfvars` files to load different variables for each environment.

```yaml
# repos.yaml or atlantis.yaml
workflows:
  staging:
    plan:
      steps:
      - run: rm -rf .terraform
      - init:
          extra_args: [-backend-config=staging.backend.tfvars]
      - plan:
          extra_args: [-var-file=staging.tfvars]
  production:
    plan:
      steps:
      - run: rm -rf .terraform
      - init:
          extra_args: [-backend-config=production.backend.tfvars]
      - plan:
          extra_args: [-var-file=production.tfvars
```
::: warning NOTE
We have to use a custom `run` step to `rm -rf .terraform` because otherwise Terraform
will complain in-between commands since the backend config has changed.
:::

You would then reference the workflows in your repo-level `atlantis.yaml`:
```yaml
version: 3
projects:
- name: staging
  dir: .
  workflow: staging
- name: production
  dir: .
  workflow: production
```

## Reference
### Workflow
```yaml
plan:
apply:
```

| Key   | Type            | Default               | Required | Description                    |
|-------|-----------------|-----------------------|----------|--------------------------------|
| plan  | [Stage](#stage) | `steps: [init, plan]` | no       | How to plan for this project.  |
| apply | [Stage](#stage) | `steps: [apply]`      | no       | How to apply for this project. |

### Stage
```yaml
steps:
- run: custom-command
- init
- plan:
    extra_args: [-lock=false]
```

| Key   | Type                 | Default | Required | Description                                                                                   |
|-------|----------------------|---------|----------|-----------------------------------------------------------------------------------------------|
| steps | array[[Step](#step)] | `[]`    | no       | List of steps for this stage. If the steps key is empty, no steps will be run for this stage. |

### Step
#### Built-In Commands: init, plan, apply
Steps can be a single string for a built-in command.
```yaml
- init
- plan
- apply
```
| Key             | Type   | Default | Required | Description                                                                                            |
| --------------- | ------ | ------- | -------- | ------------------------------------------------------------------------------------------------------ |
| init/plan/apply | string | none    | no       | Use a built-in command without additional configuration. Only `init`, `plan` and `apply` are supported |

#### Built-In Command With Extra Args
A map from string to `extra_args` for a built-in command with extra arguments.
```yaml
- init:
    extra_args: [arg1, arg2]
- plan:
    extra_args: [arg1, arg2]
- apply:
    extra_args: [arg1, arg2]
```
| Key             | Type                               | Default | Required | Description                                                                                                                                         |
|-----------------|------------------------------------|---------|----------|-----------------------------------------------------------------------------------------------------------------------------------------------------|
| init/plan/apply | map[`extra_args` -> array[string]] | none    | no       | Use a built-in command and append `extra_args`. Only `init`, `plan` and `apply` are supported as keys and only `extra_args` is supported as a value |

#### Custom `run` Command
Or a custom command
```yaml
- run: custom-command
```
| Key | Type   | Default | Required | Description          |
|-----|--------|---------|----------|----------------------|
| run | string | none    | no       | Run a custom command |

::: tip Notes
* `run` steps are executed with the following environment variables:
  * `WORKSPACE` - The Terraform workspace used for this project, ex. `default`.
    * NOTE: if the step is executed before `init` then Atlantis won't have switched to this workspace yet.
  * `ATLANTIS_TERRAFORM_VERSION` - The version of Terraform used for this project, ex. `0.11.0`.
  * `DIR` - Absolute path to the current directory.
  * `PLANFILE` - Absolute path to the location where Atlantis expects the plan to
  either be generated (by plan) or already exist (if running apply). Can be used to
  override the built-in `plan`/`apply` commands, ex. `run: terraform plan -out $PLANFILE`.
  * `BASE_REPO_NAME` - Name of the repository that the pull request will be merged into, ex. `atlantis`.
  * `BASE_REPO_OWNER` - Owner of the repository that the pull request will be merged into, ex. `runatlantis`.
  * `HEAD_REPO_NAME` - Name of the repository that is getting merged into the base repository, ex. `atlantis`.
  * `HEAD_REPO_OWNER` - Owner of the repository that is getting merged into the base repository, ex. `acme-corp`.
  * `HEAD_BRANCH_NAME` - Name of the head branch of the pull request (the branch that is getting merged into the base)
  * `BASE_BRANCH_NAME` - Name of the base branch of the pull request (the branch that the pull request is getting merged into)
  * `PROJECT_NAME` - Name of the project configured in `atlantis.yaml`. If no project name is configured this will be an empty string.
  * `PULL_NUM` - Pull request number or ID, ex. `2`.
  * `PULL_AUTHOR` - Username of the pull request author, ex. `acme-user`.
  * `REPO_REL_DIR` - The relative path of the project in the repository. For example if your project is in `dir1/dir2/` then this will be set to `"dir1/dir2"`. If your project is at the root this will be `"."`.
  * `USER_NAME` - Username of the VCS user running command, ex. `acme-user`. During an autoplan, the user will be the Atlantis API user, ex. `atlantis`.
  * `COMMENT_ARGS` - Any additional flags passed in the comment on the pull request. Flags are separated by commas and
  every character is escaped, ex. `atlantis plan -- arg1 arg2` will result in `COMMENT_ARGS=\a\r\g\1,\a\r\g\2`.
* A custom command will only terminate if all output file descriptors are closed.
Therefore a custom command can only be sent to the background (e.g. for an SSH tunnel during
the terraform run) when its output is redirected to a different location. For example, Atlantis
will execute a custom script containing the following code to create a SSH tunnel correctly: 
`ssh -f -M -S /tmp/ssh_tunnel -L 3306:database:3306 -N bastion 1>/dev/null 2>&1`. Without
the redirect, the script would block the Atlantis workflow.
* If a workflow step returns a non-zero exit code, the workflow will stop. 
:::

#### Environment Variable `env` Command
The `env` command allows you to set environment variables that will be available
to all steps defined **below** the `env` step.

You can set hard coded values via the `value` key, or set dynamic values via
the `command` key which allows you to run any command and uses the output
as the environment variable value.
```yaml
- env:
    name: ENV_NAME
    value: hard-coded-value
- env:
    name: ENV_NAME_2
    command: 'echo "dynamic-value-$(date)"'
```
| Key             | Type                               | Default | Required | Description                                                                                                                                         |
|-----------------|------------------------------------|---------|----------|-----------------------------------------------------------------------------------------------------------------------------------------------------|
| env | map[`name` -> string, `value` -> string, `command` -> string] | none    | no       | Set environment variables for subsequent steps |

::: tip Notes
* `env` `command`'s can use any of the built-in environment variables available
  to `run` commands. 
:::
