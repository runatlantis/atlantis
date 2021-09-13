# Terraform Cloud/Enterprise

::: tip NOTE
Terraform Enterprise was [recently renamed](https://www.hashicorp.com/blog/introducing-terraform-cloud-remote-state-management) Terraform Cloud
and Private Terraform Enterprise was renamed Terraform Enterprise.
:::

Atlantis integrates seamlessly with Terraform Cloud and Terraform Enterprise, whether you're using:
* [Free Remote State Management](https://app.terraform.io)
* Terraform Cloud Paid Tiers
* A Private Installation of Terraform Enterprise

Read the docs below :point_down: depending on your use-case.
[[toc]]

## Using Atlantis With Free Remote State Storage
To use Atlantis with Free Remote State Storage, you need to:
1. Migrate your state to Terraform Cloud. See [Migrating State from Local Terraform](https://www.terraform.io/docs/cloud/migrate/index.html)
1. Update any projects that are referencing the state you migrated to use the new location
1. [Generate a Terraform Cloud/Enterprise Token](#generating-a-terraform-cloud-enterprise-token)
1. [Pass the token to Atlantis](#passing-the-token-to-atlantis)

That's it! Atlantis will run as normal and your state will be stored in Terraform
Cloud.

## Using Atlantis With Terraform Cloud Remote Operations or Terraform Enterprise
Atlantis integrates with the full version of Terraform Cloud and Terraform Enterprise
via the [remote backend](https://www.terraform.io/docs/backends/types/remote.html).

Atlantis will run `terraform` commands as usual, however those commands will
actually be executed *remotely* in Terraform Cloud or Terraform Enterprise.

### Why?
Using Atlantis with Terraform Cloud or Terraform Enterprise gives you access to features like:
* Real-time streaming output
* Ability to cancel in-progress commands
* Secret variables
* [Sentinel](https://www.hashicorp.com/sentinel)

**Without** having to change your pull request workflow.

### Getting Started
To use Atlantis with Terraform Cloud Remote Operations or Terraform Enterprise, you need to:
1. Migrate your state to Terraform Cloud/Enterprise. See [Migrating State from Local Terraform](https://www.terraform.io/docs/cloud/migrate/index.html)
1. Update any projects that are referencing the state you migrated to use the new location
1. [Generate a Terraform Cloud/Enterprise Token](#generating-a-terraform-cloud-enterprise-token)
1. [Pass the token to Atlantis](#passing-the-token-to-atlantis)

## Generating a Terraform Cloud/Enterprise Token
Atlantis needs a Terraform Cloud/Enterprise Token that it will use to access the API.
Using a **Team Token is recommended**, however you can also use a User Token.

### Team Token
To generate a team token, click on **Settings** in the top bar, then **Teams** in
the sidebar.
Choose an existing team or create a new one.
Enable the **Manage Workspaces** permission, then scroll down to **Team API Token**.

### User Token
To generate a user token, click on your avatar, then **User Settings**, then
**Tokens** in the sidebar.
Ensure the **Manage Workspaces** permission is enabled for this user's team.

## Passing The Token To Atlantis
The token can be passed to Atlantis via the `ATLANTIS_TFE_TOKEN` environment variable.

You can also use the `--tfe-token` flag, however your token would then be easily
viewable in the process list.

If you're hosting your own Terraform Enterprise installation, set the `--tfe-hostname`
flag to its hostname.

That's it! Atlantis should be able to perform Terraform operations using Terraform Cloud/Enterprise's
remote state backend now.

:::warning
The Terraform Cloud/Enterprise integration only works with the built-in
`plan` and `apply` steps. It does not work with custom `run` steps that replace
plan or apply.
:::

:::tip NOTE
Under the hood, Atlantis is generating a `~/.terraformrc` file.
If you already had a `~/.terraformrc` file where Atlantis is running,
 then you'll need to manually
add the credentials block to that file:
```
...
credentials "app.terraform.io" {
  token = "xxxx"
}
```
instead of using the `ATLANTIS_TFE_TOKEN` environment variable, since Atlantis
won't overwrite your `.terraformrc` file.
:::
