# Using Atlantis

Atlantis triggers commands via pull request comments.
![Help Command](./images/pr-comment-help.png)

::: tip
You can use following executable names.
* `atlantis help`
  * `atlantis` is executable name. You can configure by [Executable Name](/docs/server-configuration.html#executable-name).
* `run help`
  * `run` is a global executable name.
* `@GithubUser help`
  * `@GithubUser` is the VCS host user which you connected to Atlantis by user token.
:::

Currently, Atlantis supports the following commands.
[[toc]]

---
## atlantis help
```bash
atlantis help
```
### Explanation
View help

---
## atlantis version
```bash
atlantis version
```

### Explanation
Print the output of 'terraform version'.

---
## atlantis plan
```bash
atlantis plan [options] -- [terraform plan flags]
```
### Explanation
Runs `terraform plan` on the pull request's branch. You may wish to re-run plan after Atlantis has already done
so if you've changed some resources manually.

### Examples
```bash
# Runs plan for any projects that Atlantis thinks were modified.
# If an `atlantis.yaml` file is specified, runs plan on the projects that
# were modified as determined by the `when_modified` config.
atlantis plan

# Runs plan in the root directory of the repo with workspace `default`.
atlantis plan -d .

# Runs plan in the `project1` directory of the repo with workspace `default`
atlantis plan -d project1

# Runs plan in the root directory of the repo with workspace `staging`
atlantis plan -w staging
```

### Options
* `-d directory` Which directory to run plan in relative to root of repo. Use `.` for root.
    * Ex. `atlantis plan -d child/dir`
* `-p project` Which project to run plan for. Refers to the name of the project configured in the repo's [`atlantis.yaml` file](repo-level-atlantis-yaml.html). Cannot be used at same time as `-d` or `-w` because the project defines this already.
* `-w workspace` Switch to this [Terraform workspace](https://developer.hashicorp.com/terraform/language/state/workspaces) before planning. Defaults to `default`. Ignore this if Terraform workspaces are unused.
* `--verbose` Append Atlantis log to comment.

::: warning NOTE
A `atlantis plan` (without flags), like autoplans, discards all plans previously created with `atlantis plan` `-p`/`-d`/`-w`
:::

### Additional Terraform flags

If `terraform plan` requires additional arguments, like `-target=resource` or `-var 'foo=bar'` or `-var-file myfile.tfvars`
you can append them to the end of the comment after `--`, ex.
```
atlantis plan -d dir -- -var foo='bar'
```
If you always need to append a certain flag, see [Custom Workflow Use Cases](custom-workflows.html#adding-extra-arguments-to-terraform-commands).

---
## atlantis apply
```bash
atlantis apply [options] -- [terraform apply flags]
```
### Explanation
Runs `terraform apply` for the plan that matches the directory/project/workspace.

::: tip
If no directory/project/workspace is specified, ex. `atlantis apply`, this command will apply **all unapplied plans from this pull request**.
This includes all projects that have been planned manually with `atlantis plan` `-p`/`-d`/`-w` since the last autoplan or `atlantis plan` command.
:::

### Examples
```bash
# Runs apply for all unapplied plans from this pull request.
atlantis apply

# Runs apply in the root directory of the repo with workspace `default`.
atlantis apply -d .

# Runs apply in the `project1` directory of the repo with workspace `default`
atlantis apply -d project1

# Runs apply in the root directory of the repo with workspace `staging`
atlantis apply -w staging
```

### Options
* `-d directory` Apply the plan for this directory, relative to root of repo. Use `.` for root.
* `-p project` Apply the plan for this project. Refers to the name of the project configured in the repo's [`atlantis.yaml` file](repo-level-atlantis-yaml.html). Cannot be used at same time as `-d` or `-w`.
* `-w workspace` Apply the plan for this [Terraform workspace](https://developer.hashicorp.com/terraform/language/state/workspaces). Ignore this if Terraform workspaces are unused.
* `--auto-merge-disabled` Disable [automerge](automerging.html) for this apply command.
* `--verbose` Append Atlantis log to comment.

### Additional Terraform flags

Because Atlantis under the hood is running `terraform apply plan.tfplan`, any Terraform options that would change the `plan` are ignored, ex:
* `-target=resource`
* `-var 'foo=bar'`
* `-var-file=myfile.tfvars`

They're ignored because they can't be specified for an already generated planfile.
If you would like to specify these flags, do it while running `atlantis plan`.

---
## atlantis import
```bash
atlantis import [options] ADDRESS ID -- [terraform import flags]
```
### Explanation
Runs `terraform import` that matches the directory/project/workspace.
This command discards the terraform plan result. After an import and before an apply, another `atlantis plan` must be run again.

To allow the `import` command requires [--allow-commands](/docs/server-configuration.html#allow-commands) configuration.

### Examples
```bash
# Runs import
atlantis import ADDRESS ID

# Runs import in the root directory of the repo with workspace `default`
atlantis import -d . ADDRESS ID

# Runs import in the `project1` directory of the repo with workspace `default`
atlantis import -d project1 ADDRESS ID

# Runs import in the root directory of the repo with workspace `staging`
atlantis import -w staging ADDRESS ID
```

::: tip
* If import for_each resources, it requires a single quoted address.
  * ex. `atlantis import 'aws_instance.example["foo"]' i-1234567890abcdef0`
:::

### Options
* `-d directory` Import a resource for this directory, relative to root of repo. Use `.` for root.
* `-p project` Import a resource for this project. Refers to the name of the project configured in the repo's [`atlantis.yaml`](repo-level-atlantis-yaml.html) repo configuration file. This cannot be used at the same time as `-d` or `-w`.
* `-w workspace` Import a resource for a specific [Terraform workspace](https://developer.hashicorp.com/terraform/language/state/workspaces). Ignore this if Terraform workspaces are unused.

### Additional Terraform flags

If `terraform import` requires additional arguments, like `-var 'foo=bar'` or `-var-file myfile.tfvars`
append them to the end of the comment after `--`, e.g.
```
atlantis import -d dir 'aws_instance.example["foo"]' i-1234567890abcdef0 -- -var foo='bar'
```
If a flag is needed to be always appended, see [Custom Workflow Use Cases](custom-workflows.html#adding-extra-arguments-to-terraform-commands).

---
## atlantis state rm
```bash
atlantis state [options] rm ADDRESS... -- [terraform state rm flags]
```
### Explanation
Runs `terraform state rm` that matches the directory/project/workspace.
This command discards the terraform plan result. After run state rm and before an apply, another `atlantis plan` must be run again.

To allow the `state` command requires [--allow-commands](/docs/server-configuration.html#allow-commands) configuration.

### Examples
```bash
# Runs state rm
atlantis state rm ADDRESS1 ADDRESS2

# Runs state rm in the root directory of the repo with workspace `default`
atlantis state -d . rm ADDRESS

# Runs state rm in the `project1` directory of the repo with workspace `default`
atlantis state -d project1 rm ADDRESS

# Runs state rm in the root directory of the repo with workspace `staging`
atlantis state -w staging rm ADDRESS
```

::: tip
* If run state rm to for_each resources, it requires a single quoted address.
  * ex. `atlantis state rm 'aws_instance.example["foo"]'`
:::

### Options
* `-d directory` Run state rm a resource for this directory, relative to root of repo. Use `.` for root.
* `-p project` Run state rm a resource for this project. Refers to the name of the project configured in the repo's [`atlantis.yaml`](repo-level-atlantis-yaml.html) repo configuration file. This cannot be used at the same time as `-d` or `-w`.
* `-w workspace` Run state rm a resource for a specific [Terraform workspace](https://developer.hashicorp.com/terraform/language/state/workspaces). Ignore this if Terraform workspaces are unused.

### Additional Terraform flags

If `terraform state rm` requires additional arguments, like `-lock=false'`
append them to the end of the comment after `--`, e.g.
```
atlantis state -d dir rm 'aws_instance.example["foo"]' -- -lock=false
```
If a flag is needed to be always appended, see [Custom Workflow Use Cases](custom-workflows.html#adding-extra-arguments-to-terraform-commands).

---
## atlantis unlock
```bash
atlantis unlock
```

### Explanation
Removes all atlantis locks and discards all plans for this PR.
To unlock a specific plan you can use the Atlantis UI.

---
## atlantis approve_policies
```bash
atlantis approve_policies
```

### Explanation
Approves all current policy checking failures for the PR.

See also [policy checking](/docs/policy-checking.html).

### Options
* `--verbose` Append Atlantis log to comment.
