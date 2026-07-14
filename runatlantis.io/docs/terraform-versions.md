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

See [atlantis.yaml Use Cases](repo-level-atlantis-yaml.md#terraform-versions) for more details.

## Via terraform config

Alternatively, one can use the terraform configuration block's `required_version` key to specify an exact version (`x.y.z` or `= x.y.z`), or as of [atlantis v0.21.0](https://github.com/runatlantis/atlantis/releases/tag/v0.21.0), a comparison or pessimistic [version constraint](https://developer.hashicorp.com/terraform/language/expressions/version-constraints#version-constraint-syntax):

### Exactly version 1.2.9

```tf
terraform {
  required_version = "= 1.2.9"
}
```

### Any patch/tiny version of minor version 1.2 (1.2.z)

```tf
terraform {
  required_version = "~> 1.2.0"
}
```

### Any minor version of major version 1 (1.y.z)

```tf
terraform {
  required_version = "~> 1.2"
}
```

### Any version that is at least 1.2.0

```tf
terraform {
  required_version = ">= 1.2.0"
}
```

See [Terraform `required_version`](https://developer.hashicorp.com/terraform/language/terraform#terraform-required_version) for reference.

::: tip NOTE
Atlantis will automatically download the latest version that fulfills the constraint specified.
A `terraform_version` specified in the `atlantis.yaml` file takes precedence over both the [`--default-tf-version`](server-configuration.md#default-tf-version) flag and the `required_version` in the terraform hcl.
When a project sets `terraform_distribution`, Atlantis resolves the `required_version`
constraint against that distribution. For example, an OpenTofu project resolves to an
OpenTofu version instead of a Terraform version.
:::

## OpenTofu `.tofu` file support

When the effective distribution is OpenTofu, Atlantis reads `required_version`
from `.tofu` and `.tofu.json` files in addition to `.tf` and `.tf.json`. The effective
distribution is OpenTofu when either:

- A project sets `terraform_distribution: opentofu` in `atlantis.yaml`
- The server default is `--default-tf-distribution=opentofu` and the project does not override it

If a project explicitly sets `terraform_distribution: terraform`, Atlantis uses the
Terraform version-detection path (`.tf` / `.tf.json` only) even if the server default is OpenTofu.

OpenTofu file precedence is respected: a `.tofu` file overrides a same-basename `.tf`
file, and `.tofu.json` overrides a same-basename `.tf.json` file. Files with different
basenames both contribute constraints independently.

Terraform distribution is unaffected and continues to read only `.tf` / `.tf.json` files.

::: warning Known limitation
Module autoplanning (`--autoplan-modules`) dependency indexing still relies on
`terraform-config-inspect` and does not fully understand `.tofu` / `.tofu.json`:

1. Module source blocks defined only in `.tofu` / `.tofu.json` files are not indexed.
   Projects using these files for `module {}` blocks will not be planned when shared
   modules change.
2. Shared module directories containing only `.tofu` / `.tofu.json` files may not be
   recognized as modules by the dependency index.

Direct file-change autoplanning (without `--autoplan-modules`) is fully supported for
`.tofu` projects. As a workaround for module dependencies, include shared module paths
in explicit `autoplan.when_modified` patterns, or keep module source declarations in
`.tf` files until full `.tofu` module indexing is implemented.

Terraform Cloud workspace detection (`cloud { workspaces { ... } }`) supports `.tf`,
`.tf.json`, `.tofu`, and `.tofu.json` files for **autodiscovered projects**. `.tofu` and
`.tofu.json` are only scanned when the server default distribution
(`--default-tf-distribution`) is OpenTofu. Same-basename precedence applies in OpenTofu
mode: `.tofu` overrides `main.tf`, `.tofu.json` overrides `main.tf.json`. Projects with
Terraform server default read `.tf` and `.tf.json` but ignore `.tofu` / `.tofu.json`.

For explicitly configured projects in `atlantis.yaml`, set the `workspace:` field directly
— workspace HCL scanning is not used for configured projects regardless of distribution.
:::

::: tip NOTE
The Atlantis [latest docker image](https://github.com/runatlantis/atlantis/pkgs/container/atlantis/9854680?tag=latest) tends to have recent versions of Terraform, but there may be a delay as new versions are released. The highest version of Terraform allowed in your code is the version specified by `DEFAULT_TERRAFORM_VERSION` in the image your server is running.
:::
