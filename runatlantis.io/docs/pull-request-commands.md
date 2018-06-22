# Pull Request Commands
Atlantis currently supports three commands that can be run via pull request comments (or merge request comments on GitLab):

![Help Command](./images/pr-comment-help.png)
## `atlantis help`
View help

---
![Plan Command](./images/pr-comment-plan.png)
## `atlantis plan [options] -- [terraform plan flags]`
Runs `terraform plan` for the changes in this pull request.

Options:
* `-d directory` Which directory to run plan in relative to root of repo. Use `.` for root. If not specified, will attempt to run plan for all Terraform projects we think were modified in this changeset.
* `-w workspace` Switch to this [Terraform workspace](https://www.terraform.io/docs/state/workspaces.html) before planning. Defaults to `default`. If not using Terraform workspaces you can ignore this.
* `--verbose` Append Atlantis log to comment.

Additional Terraform flags:

If you need to run `terraform plan` with additional arguments, like `-target=resource` or `-var 'foo-bar'`
you can append them to the end of the comment after `--`, ex.
```
atlantis plan -d dir -- -var 'foo=bar'
```
If you always need to append a certain flag, see [Project-Specific Customization](#project-specific-customization).

---
![Apply Command](./images/pr-comment-apply.png)
## `atlantis apply [options] -- [terraform apply flags]`
Runs `terraform apply` for the plans that match the directory and workspace.

Options:
* `-d directory` Apply the plan for this directory, relative to root of repo. Use `.` for root. If not specified, will run apply against all plans created for this workspace.
* `-w workspace` Apply the plan for this [Terraform workspace](https://www.terraform.io/images/state/workspaces.html). Defaults to `default`. If not using Terraform workspaces you can ignore this.
* `--verbose` Append Atlantis log to comment.

Additional Terraform flags:

Same as with `atlantis plan`.