# Terraform Versions

You can customize which version of Terraform Atlantis defaults to by setting
the `--default-tf-version` flag (ex. `--default-tf-version=v1.3.7`).

## Via `atlantis.yaml`
If you wish to use a different version than the default for a specific repo or project, you need
to create an `atlantis.yaml` file and set the `terraform_version` key:
```yaml
version: 3
projects:
- dir: .
  terraform_version: v1.1.5
```
See [atlantis.yaml Use Cases](repo-level-atlantis-yaml.html#terraform-versions) for more details.

## Via terraform config
Alternatively, one can use the terraform configuration block's `required_version` key to specify an exact version (`x.y.z` or `= x.y.z`), or as of [atlantis v0.21.0](https://github.com/runatlantis/atlantis/releases/tag/v0.21.0), a comparison or pessimistic [version constraint](https://developer.hashicorp.com/terraform/language/expressions/version-constraints#version-constraint-syntax):
#### Exactly version 1.2.9
```tf
terraform {
  required_version = "= 1.2.9"
}
```
#### Any patch/tiny version of minor version 1.2 (1.2.z)
```tf
terraform {
  required_version = "~> 1.2.0"
}
```
#### Any minor version of major version 1 (1.y.z)
```tf
terraform {
  required_version = "~> 1.2"
}
```
#### Any version that is at least 1.2.0
```tf
terraform {
  required_version = ">= 1.2.0"
}
```
See [Terraform `required_version`](https://developer.hashicorp.com/terraform/language/settings#specifying-a-required-terraform-version) for reference.

::: tip NOTE
Atlantis will automatically download the latest version that fulfills the constraint specified.
A `terraform_version` specified in the `atlantis.yaml` file takes precedence over both the [`--default-tf-version`](server-configuration.html#default-tf-version) flag and the `required_version` in the terraform hcl.
:::

::: tip NOTE
The Atlantis [latest docker image](https://github.com/runatlantis/atlantis/pkgs/container/atlantis/9854680?tag=latest) tends to have recent versions of Terraform, but there may be a delay as new versions are released. The highest version of Terraform allowed in your code is the version specified by `DEFAULT_TERRAFORM_VERSION` in the image your server is running.
:::
