# Terraform Enterprise

Atlantis integrates seamlessly with Terraform Enterprise, whether you're using:
* New [Free Remote State Management](https://app.terraform.io/signup)
* Terraform Enterprise SaaS
* Private Terraform Enterprise

Read the docs below :point_down: depending on your use-case.
[[toc]]

## Using Atlantis With Free Remote State Storage
To use Atlantis with Free Remote State Storage, you need to:
1. Migrate your state to Terraform Enterprise. See [Getting Started with the Terraform Enterprise Free Tier](https://www.terraform.io/docs/enterprise/free/index.html#enable-remote-state-in-terraform-configurations)
1. Update any projects that are referencing the state you migrated to use the new location
1. [Generate a Terraform Enterprise Token](#generating-a-terraform-enterprise-token)
1. [Pass the token to Atlantis](#passing-the-token-to-atlantis)

That's it! Atlantis will run as normal and your state will be stored in Terraform
Enterprise.

## Using Atlantis With Full Terraform Enterprise
Atlantis integrates with the full version of Terraform Enterprise (TFE) via the [remote backend](https://www.terraform.io/docs/backends/types/remote.html).

Atlantis will run `terraform` commands as usual, however those commands will
actually be executed *remotely* in Terraform Enterprise.

### Why?
Using Atlantis with TFE gives you access to Terraform Enterprise features like:
* Real-time streaming output
* Ability to cancel in-progress commands
* Secret variables
* [Sentinel](https://www.hashicorp.com/sentinel)

**Without** having to change your pull request workflow.

### Getting Started
To use Atlantis with Terraform Enterprise, you need to:
1. Migrate your state to Terraform Enterprise. See [Migrating State from Terraform Open Source](https://www.terraform.io/docs/enterprise/migrate/index.html)
1. Update any projects that are referencing the state you migrated to use the new location
1. [Generate a Terraform Enterprise Token](#generating-a-terraform-enterprise-token)
1. [Pass the token to Atlantis](#passing-the-token-to-atlantis)

## Generating a Terraform Enterprise Token
Atlantis needs a TFE Token that it will use to access the TFE API.
Using a **Team Token is recommended**, however you can also use a User Token.

### Team Token
To generate a team token, click on **Settings** in the top bar, then **Teams** in
the sidebar, then scroll down to **Team API Token**.

### User Token
To generate a user token, click on your avatar, then **User Settings**, then
**Tokens** in the sidebar.

## Passing The Token To Atlantis
The token can be passed to Atlantis via the `ATLANTIS_TFE_TOKEN` environment variable.

You can also use the `--tfe-token` flag, however your token would then be easily
viewable in the process list.

That's it! Atlantis should be able to perform Terraform operations using TFE's
remote state backend now.

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
