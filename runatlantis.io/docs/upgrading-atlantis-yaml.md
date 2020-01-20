# Upgrading atlantis.yaml

## Upgrading From v2 To v3
Atlantis version `v0.7.0` introduced a new version 3 of `atlantis.yaml`.

**If you're not using [custom `run` steps](custom-workflows.html#custom-run-command),
 then you can upgrade from `version: 2` to `version: 3` without any changes.**

**NOTE:** Version 2 **is not being deprecated** and there is no need to upgrade your version
if you don't wish to do so.

The only change from v2 to v3 is that we're parsing custom `run` steps differently.
```yaml
# atlantis.yaml
workflows:
  custom:
    plan:
      steps:
      - run: my custom command
```
<center><i>An example workflow using a custom run step</i></center>

Previously, we used a library that would parse the custom step prior to running
it. Now, we just run the step directly. This will only affect your steps if they were using shell escaping of some sort.
For example, if your step was previously:
```yaml
# version: 2
- run: "printf \'print me\'"
```

You can now write this in version 3 as:
```yaml
# version: 3
- run: "printf 'print me'"
```


## Upgrading From V1 To V3
If you are upgrading from an **old** Atlantis version `<=v0.3.10` (from before July 4, 2018)
you'll need to follow the following steps.

### Single atlantis.yaml
If you had multiple `atlantis.yaml` files per directory then you'll need to
consolidate them into a single `atlantis.yaml` file at the root of the repo.

For example, if you had a directory structure:
```
.
├── project1
│   └── atlantis.yaml
└── project2
    └── atlantis.yaml
```

Then your new structure would look like:
```
.
├── atlantis.yaml
├── project1
└── project2
```

And your `atlantis.yaml` would look something like:
```yaml
version: 2
projects:
- dir: project1
  terraform_version: my-version
  workflow: project1-workflow
- dir: project2
  terraform_version: my-version
  workflow: project2-workflow
workflows:
  project1-workflow:
    ...
  project2-workflow:
    ...
```

We will talk more about `workflows` below.

### Terraform Version
The `terraform_version` key moved from being a top-level key to being per `project`
so if before your `atlantis.yaml` was in directory `mydir` and looked like:
```yaml
terraform_version: 0.11.0
```

Then your new config would be:
```yaml
version: 2
projects:
- dir: mydir
  terraform_version: 0.11.0
```

### Workflows
Workflows are the new way to set all `pre_*`, `post_*` and `extra_arguments`.

Each `project` can have a custom workflow via the `workflow` key.
```yaml
version: 2
projects:
- dir: .
  workflow: myworkflow
```

Workflows are defined as a top-level key:
```yaml
version: 2
projects:
...

workflows:
  myworkflow:
  ...
```

To start with, determine whether you're customizing commands that happen during
`plan` or `apply`. You then set that key under the workflow's name:
```yaml
...
workflows:
  myworkflow:
    plan:
      steps:
      ...
    apply:
      steps:
      ...
```

If you're not customizing a specific stage then you can omit that key. For example
if you're only customizing the commands that happen during `plan` then your config
will look like:
```yaml
...
workflows:
  myworkflow:
    plan:
      steps:
      ...
```

#### Extra Arguments
`extra_arguments` is now specified as follows. Given a previous config:
```yaml
extra_arguments:
  - command_name: init
    arguments:
    - "-lock=false"
  - command_name: plan
    arguments:
    - "-lock=false"
  - command_name: apply
    arguments:
    - "-lock=false"
```

Your config would now look like:
```yaml
...
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


#### Pre/Post Commands
Instead of using `pre_*` or `post_*`, you now can insert your custom commands
before/after the built-in commands. Given a previous config:

```yaml
pre_init:
  commands:
  - "curl http://example.com"
# pre_get commands are run when the Terraform version is < 0.9.0
pre_get:
  commands:
  - "curl http://example.com"
pre_plan:
  commands:
  - "curl http://example.com"
post_plan:
  commands:
  - "curl http://example.com"
pre_apply:
  commands:
  - "curl http://example.com"
post_apply:
  commands:
  - "curl http://example.com"
```

Your config would now look like:
```yaml
...
workflows:
  myworkflow:
    plan:
      steps:
      - run: curl http://example.com
      - init
      - plan
      - run: curl http://example.com
    apply:
      steps:
      - run: curl http://example.com
      - apply
      - run: curl http://example.com
```

::: tip
It's important to include the built-in commands: `init`, `plan` and `apply`.
Otherwise Atlantis won't run the necessary commands to actually plan/apply.
:::
