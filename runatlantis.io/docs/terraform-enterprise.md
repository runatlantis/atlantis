# Terraform Enterprise

Atlantis integrates seamlessly with Terraform Enterprise's new [Free Remote State Management](https://app.terraform.io/signup).

[[toc]]

## Migrating to TFE's Remote State
If you're using a different state backend, you first need to migrate your state
to use Terraform Enterprise (TFE). Read [TODO: use right link](https://www.terraform.io/docs/enterprise/migrate/index.html)
for more information on how to migrate.

## Configuring Atlantis
Once you've migrated your state to TFE, and your code is using the TFE backend:

```bash
# Example configuration
terraform {
  backend "remote" {
    organization = "company"
    workspaces {
      name = "my-app-prod"
    }
  }
}
```

You need to provide Atlantis with a TFE [User Token](https://www.terraform.io/docs/enterprise/users-teams-organizations/users.html#api-tokens)
that it will use to access the TFE API.

You can provide this token by either:
1. Setting the `--tfe-token` flag, or the `ATLANTIS_TFE_TOKEN` environment variable
1. Creating a `.terraformrc` file in the home directory of whichever user is executing Atlantis
    with the following contents:
    ```json
    credentials "app.terraform.io" {
      token = "xxxxxx.hunter2.zzzzzzzzzzzzz"
    }
    ```

Notes:
* If you specify the `--tfe-token` or `ATLANTIS_TFE_TOKEN` environment variable,
    on startup, Atlantis will generate a config file to `~/.terraformrc`. If
    this file already exists, Atlantis will error.
* If you're using the Atlantis Docker image, the `.terraformrc` file should be
   placed in `/home/atlantis/.terraformrc`
