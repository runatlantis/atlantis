# Terraform Versions

You can customize which version of Terraform Atlantis defaults to by setting
the `--default-tf-version` flag (ex. `--default-tf-version=v0.12.0`).

## Via `atlantis.yaml`
If you wish to use a different version than the default for a specific repo or project, you need
to create an `atlantis.yaml` file and set the `terraform_version` key:
```yaml
version: 3
projects:
- dir: .
  terraform_version: v0.10.5
```
See [atlantis.yaml Use Cases](repo-level-atlantis-yaml.html#terraform-versions) for more details.

## Via terraform config
Alternatively, one can use the terraform configuration block's `required_version` key to specify an *exact* version:
```tf
terraform {
  required_version = "0.12.0"
}
```
See [Terraform `required_version`](https://www.terraform.io/docs/configuration/terraform.html#specifying-a-required-terraform-version) for reference.

::: tip NOTE
Atlantis will automatically download the version specified.
:::

::: tip NOTE
The Atlantis [latest docker image](https://hub.docker.com/layers/runatlantis/atlantis/latest/images/sha256-4f80472e20bd899b03a619e593f9e7b9a55d9e630850de443b988295f63f5c7a?context=explore) tends to have recent versions of Terraform, but there may be a delay as new versions are released. The highest version of Terraform allowed in your code is the version specified by `DEFAULT_TERRAFORM_VERSION` in the image your server is running.
:::
