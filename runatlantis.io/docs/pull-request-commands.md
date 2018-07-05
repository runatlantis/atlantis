# Pull Request Commands
Atlantis currently supports three commands that can be run via pull request comments:
[[toc]]

## atlantis help
![Help Command](./images/pr-comment-help.png)
```bash
atlantis help
```
**Explanation**: View help

---
## atlantis plan
![Plan Command](./images/pr-comment-plan.png)
```bash
atlantis plan [options] -- [terraform plan flags]
```
**Explanation**: Runs `terraform plan` on the pull request's branch. You may wish to re-run plan after Atlantis has already done
so if you've changed some resources manually.

Options:
* `-d directory` Which directory to run plan in relative to root of repo. Use `.` for root. Defaults to root.
    * Ex. `atlantis plan -d child/dir`
* `-p project` Which project to run plan for. Refers to the name of the project configured in the repo's [`atlantis.yaml` file](/docs/atlantis-yaml-reference.html). Cannot be used at same time as `-d` or `-w` because the project defines this already.
* `-w workspace` Switch to this [Terraform workspace](https://www.terraform.io/docs/state/workspaces.html) before planning. Defaults to `default`. If not using Terraform workspaces you can ignore this.
* `--verbose` Append Atlantis log to comment.

Additional Terraform flags:

If you need to run `terraform plan` with additional arguments, like `-target=resource` or `-var 'foo-bar'` or `-var-file myfile.tfvars`
you can append them to the end of the comment after `--`, ex.
```
atlantis plan -d dir -- -var 'foo=bar'
```
If you always need to append a certain flag, see [Project-Specific Customization](#project-specific-customization).

---
## atlantis apply
![Apply Command](./images/pr-comment-apply.png)
```bash
atlantis apply [options] -- [terraform apply flags]
```
**Explanation**: Runs `terraform apply` for the plan generated previously that matches the directory/project/workspace.

Options:
* `-d directory` Apply the plan for this directory, relative to root of repo. Use `.` for root. Defaults to root.
* `-p project` Apply the plan for this project. Refers to the name of the project configured in the repo's [`atlantis.yaml` file](/docs/atlantis-yaml-reference.html). Cannot be used at same time as `-d` or `-w`.
* `-w workspace` Apply the plan for this [Terraform workspace](https://www.terraform.io/docs/state/workspaces.html). Defaults to `default`. If not using Terraform workspaces you can ignore this.
* `--verbose` Append Atlantis log to comment.

Additional Terraform flags:

Because Atlantis under the hood is running `terraform apply` on the planfile generated in the previous step, any Terraform options that would change the `plan` are ignored:
* `-target=resource`
* `-var 'foo=bar'`
* `-var-file=myfile.tfvars`
If you would like to specify these flags, do it while running `atlantis plan`.
