# Customization
An `atlantis.yaml` config file in your project root (which is not necessarily the repo root) can be used to customize
- what commands Atlantis runs **before** `init`, `get`, `plan` and `apply` with `pre_init`, `pre_get`, `pre_plan` and `pre_apply`
- what commands Atlantis runs **after** `plan` and `apply` with `post_plan` and `post_apply`
- additional arguments to be supplied to specific terraform commands with `extra_arguments`
    - the commmands that we support adding extra args to are `init`, `get`, `plan` and `apply`
- what version of Terraform to use (see [Terraform Versions](#terraform-versions))

The schema of the `atlantis.yaml` project config file is

```yaml
# atlantis.yaml
---
terraform_version: 0.8.8 # optional version
# pre_init commands are run when the Terraform version is >= 0.9.0
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
extra_arguments:
  - command_name: plan
    arguments:
    - "-var-file=terraform.tfvars"
```

When running the `pre_plan`, `post_plan`, `pre_apply`, and `post_apply` commands the following environment variables are available
- `WORKSPACE`: if a workspace argument is supplied to `atlantis plan` or `atlantis apply`, ex `atlantis plan -w staging`, this will
be the value of that argument. Else it will be `default`
- `ATLANTIS_TERRAFORM_VERSION`: local version of `terraform` or the version from `terraform_version` if specified, ex. `0.8.8`
- `DIR`: absolute path to the root of the project on disk