# atlantis.yaml Use Cases

An `atlantis.yaml` file can be placed in the root of each repository to configure
how Atlantis runs. This documentation describes some use cases.

::: tip
Looking for the full atlantis.yaml reference? See [atlantis.yaml Reference](../docs/atlantis-yaml-reference.html).
:::

[[toc]]

## Disabling Autoplanning
```yaml
version: 2
projects:
- dir: project1
  autoplan:
    enabled: false
```
This will stop Atlantis automatically running plan when `project1/` is updated
in a pull request.

## Configuring Autoplanning
Given the directory structure:
```
.
├── modules
│   └── module1
│       ├── main.tf
│       ├── outputs.tf
│       └── submodule
│           ├── main.tf
│           └── outputs.tf
└── project1
    └── main.tf
```
If you wanted Atlantis to autoplan `project1/` whenever any `.tf` file under `module1/`
changed or any `.tf` or `.tfvars` file under `project1/` changed, you could use the following configuration:

```yaml
version: 2
projects:
- dir: project1
  autoplan:
    when_modified: ["../modules/**/*.tf", "*.tf*"]
```
Note:
* `when_modified` uses the [`.dockerignore` syntax](https://docs.docker.com/engine/reference/builder/#dockerignore-file)
* The paths are relative to the project's directory.

## Supporting Terraform Workspaces
```yaml
version: 2
projects:
- dir: project1
  workspace: staging
- dir: project1
  workspace: production
```
With the above config, when Atlantis determines that the configuration for the `project1` dir has changed,
it will run plan for both the `staging` and `production` workspaces.

If you want to `plan` or `apply` for a specific workspace you can use
```
atlantis plan -w staging -d project1
```
and
```
atlantis apply -w staging -d project1
```

## Using .tfvars files
Given the structure:
```
.
└── project1
    ├── main.tf
    ├── production.tfvars
    └── staging.tfvars
```

If you wanted Atlantis to automatically run plan with `-var-file staging.tfvars` and `-var-file production.tfvars`
you could use the following config:

```yaml
version: 2
projects:
# If two or more projects have the same dir and workspace, they must also have
# a 'name' key to differentiate them.
- name: project1-staging
  dir: project1
  # NOTE: the key here is 'workflow' not 'workspace'
  workflow: staging
- name: project1-production
  dir: project1
  workflow: production

workflows:
  staging:
    plan:
      steps:
      - init
      - plan:
          extra_args: ["-var-file", "staging.tfvars"]
  production:
    plan:
      steps:
      - init
      - plan:
          extra_args: ["-var-file", "production.tfvars"]
```
Here we're defining two projects with the same directory but with different
`workflow`s.

If you wanted to manually plan one of these projects you could use
```
atlantis plan -p project1-staging
```
Where `-p` refers to the project name.

When you want to apply the plan, you can run
```
atlantis apply -p project1-staging
```

::: warning Why can't you use atlantis apply -d project1?
Because Atlantis outputs the plan for both workflows into the `project1` directory
so it needs a way to differentiate between the plans.
:::

## Adding extra arguments to Terraform commands
If you need to append flags to `terraform plan` or `apply` temporarily, you can
append flags on a comment following `--`, for example commenting:
```
atlantis plan -- -lock=false
```
Would cause atlantis to run `terraform plan -lock=false`.

If you always need to do this for a project's `init`, `plan` or `apply` commands
then you must define the project's steps and set the `extra_args` key for the
command you need to modify.

```yaml
version: 2
projects:
- dir: project1
  workflow: myworkflow
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

## Terragrunt
Atlantis supports running custom commands in place of the default Atlantis
commands. We can use this functionality to enable
[Terragrunt](https://github.com/gruntwork-io/terragrunt).

Given a directory structure:
```
.
├── atlantis.yaml
├── live
│   ├── prod
│   │   └── terraform.tfvars
│   └── staging
│       └── terraform.tfvars
└── modules
    └── ...
```

You could use an `atlantis.yaml` with these contents:
```yaml
version: 2
projects:
- dir: live/staging
  workflow: terragrunt
- dir: live/prod
  workflow: terragrunt
workflows:
  terragrunt:
    plan:
      steps:
      - run: terragrunt plan -no-color -out $PLANFILE
    apply:
      steps:
      - run: terragrunt apply -no-color $PLANFILE
```

## Running custom commands
Atlantis supports running custom commands. In this example, we want to run
a script after every `apply`:

```yaml
version: 2
projects:
- dir: project1
  workflow: myworkflow
workflows:
  myworkflow:
    apply:
      steps:
      - apply
      - run: ./my-custom-script.sh
```

::: tip
Note how we're not specifying the `plan` key under `myworkflow`. If the `plan` key
isn't set, Atlantis will use the default plan workflow which is what we want in this case.
:::

## Terraform Versions
If you'd like to use a different version of Terraform than what is in Atlantis'
`PATH` then set the `terraform_version` key:

```yaml
version: 2
projects:
- dir: project1
  terraform_version: 0.10.0
```

Atlantis will then execute all Terraform commands with `terraform0.10.0` instead
of `terraform`. This requires that the 0.10.0 binary is in Atlantis's `PATH` with the
name `terraform0.10.0`.

## Requiring Approvals For Production
In this example, we only want to require `apply` approvals for the `production` directory.
```yaml
version: 2
projects:
- dir: staging
- dir: production
  apply_requirements: [approved]
```
:::tip
By default, there are no apply requirements so we only need to specify the `apply_requirements` key for production.
:::


## Custom Backend Config
If you need to specify the `-backend-config` flag to `terraform init` you'll need to use an `atlantis.yaml` file.
In this example, we're using custom backend files to configure two remote states, one for each environment.
We're then using `.tfvars` files to load different variables for each environment.

```yaml
version: 2
projects:
- name: staging
  dir: .
  workflow: staging
- name: production
  dir: .
  workflow: production
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
          extra_args: [-var-file=production.tfvars]
```
::: warning NOTE
We have to use a custom `run` step to `rm -rf .terraform` because otherwise Terraform
will complain in-between commands since the backend config has changed.
:::

## Next Steps
Check out the full [`atlantis.yaml` Reference](../docs/atlantis-yaml-reference.html) for more details.
