# Terraform Versions

You can customize which version of Terraform Atlantis defaults to by setting
the `--default-tf-version` flag (ex. `--default-tf-version=v0.12.1`).

If you wish to use a different version than the default for a specific repo or project, you need
to create an `atlantis.yaml` file and set the `terraform_version` key:
```yaml
version: 3
projects:
- dir: .
  terraform_version: v0.10.5
```
See [atlantis.yaml Use Cases](repo-level-atlantis-yaml.html#terraform-versions) for more details.

::: tip NOTE
Atlantis will automatically download the version specified.
:::
