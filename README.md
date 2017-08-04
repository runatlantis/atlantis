# Atlantis

<p align="center">
  <img src="./docs/atlantis-logo.png" alt="Atlantis Logo"/><br><br>
  A unified workflow for collaborating on Terraform through GitHub
</p>

## Walkthrough Video
[![Atlantis Walkthrough](./docs/atlantis-walkthrough-icon.png)](https://www.youtube.com/watch?v=TmIPWda0IKg)

[![CircleCI](https://circleci.com/gh/hootsuite/atlantis/tree/master.svg?style=shield&circle-token=08bf5b34233b0e168a9dd73e01cafdcf7dc4bf16)](https://circleci.com/gh/hootsuite/atlantis/tree/master)
![SuperDopeBadge](https://img.shields.io/badge/Hightower-super%20dope-b9f2ff.svg)

* [Features](#features)
* [Getting Started](#getting-started)
* [Project Structure](#project-structure)
* [Environments](#environments)
* [Terraform Versions](#terraform-versions)
* [Project-Specific Customization](#project-specific-customization)
* [Locking](#locking)
* [Approvals](#approvals)
* [Production-Ready Deployment](#production-ready-deployment)
* [Server Configuration](#server-configuration)
* [AWS Credentials](#aws-credentials)
* [Glossary](#glossary)
    * [Project](#project)
    * [Environment](#environment)

## Features
➜ Collaborate on Terraform with your team
- Run terraform `plan` and `apply` **from GitHub pull requests** so everyone can review the output
- **Lock environments** until pull requests are merged to prevent concurrent modification and confusion

➜ Developers can write Terraform safely
- **No need to distribute AWS credentials** to your whole team. Developers can submit Terraform changes and run `plan` and `apply` directly from the pull request
- Optionally, require a **review and approval** prior to running `apply`

➜ Also
- No more **copy-pasted code across environments**. Atlantis supports using an `env/{env}.tfvars` file per environment so you can write your base configuration once
- Support **multiple versions of Terraform** with a simple project config file


## Getting Started
Download from https://github.com/hootsuite/atlantis/releases

Run
```
./atlantis bootstrap
```
This will walk you through running Atlantis locally. It will
- fork an example terraform project
- install terraform (if not already in your PATH)
- install ngrok so we can expose Atlantis to GitHub
- start Atlantis

If you're ready to permanently set up Atlantis see [Production-Ready Deployment](#production-ready-deployment)

## Project Structure
Atlantis supports several Terraform project structures:
- a single Terraform project at the repo root
```
.
├── main.tf
└── ...
```
-  multiple project folders in a single repo (monorepo)
```
.
├── project1
│   ├── main.tf
|   └── ...
└── project2
    ├── main.tf
    └── ...
```
-  one folder per environment
```
.
├── staging
│   ├── main.tf
|   └── ...
└── production
    ├── main.tf
    └── ...
```
-  using `env/{env}.tfvars` to define environment specific variables. This works in both multi-project repos and single-project repos.
```
.
├── env
│   ├── production.tfvars
│   └── staging.tfvars
└── main.tf
```
or
```
.
├── project1
│   ├── env
│   │   ├── production.tfvars
│   │   └── staging.tfvars
│   └── main.tf
└── project2
    ├── env
    │   ├── production.tfvars
    │   └── staging.tfvars
    └── main.tf
```
With the above project structure you can de-duplicate your Terraform code between environments without requiring extensive use of modules. At Hootsuite we've found this project format to be very successful and use it in all of our 100+ Terraform repositories.

## Environments
Terraform recently introduced [State Environments](https://www.terraform.io/docs/state/environments.html) that
> allows a single folder of Terraform configurations to manage multiple distinct infrastructure resources

If you're using a Terraform version >= 0.9.0, Atlantis supports environments through an additional argument to the `atlantis plan` and `atlantis apply` commands.
For example,
```
atlantis plan staging
```

If an environment is specified Atlantis will use `terraform env select {env}` prior to running `terraform plan` or `terraform apply`.

If you're using the `env/{env}.tfvars` [project structure](#project-structure) we will also append `-tfvars=env/{env}.tfvars` to `plan` and `apply`.

If no environment is specified we will use `default` as the environment.

## Terraform Versions
By default, Atlantis will use the `terraform` executable that is in its path. To use a specific version of Terraform just install that version on the server that Atlantis is running on.

If you would like to use a different version of Terraform for some projects but not for others
1. Install the desired version of Terraform into the `$PATH` of where Atlantis is running and name it `terraform{version}`, ex. `terraform0.8.8`.
2. In the project root (which is not necessarily the repo root) of any project that needs a specific version, create an `atlantis.yaml` file as follows
```
---
terraform_version: 0.8.8 # set to desired version
```

So your project structure will look like
```
.
├── main.tf
└── atlantis.yaml
```
Now when Atlantis executes it will use the `terraform{version}` executable.

## Project-Specific Customization
An `atlantis.yaml` config file in your project root (which is not necessarily the repo root) can be used to customize
- what commands Atlantis runs **before** `plan` and `apply` with `pre_plan` and `pre_apply`
- additional arguments to be supplied to specific terraform commands with `extra_arguments`
- what version of Terraform to use (see [Terraform Versions](#terraform-versions))

The schema of the `atlantis.yaml` project config file is

```yaml
# atlantis.yaml
---
terraform_version: 0.8.8 # optional version
pre_plan:
  commands:
  - "curl http://example.com"
pre_apply:
  commands:
  - "curl http://example.com"
extra_arguments:
  - command: plan
    arguments:
    - "-tfvars=myvars.tfvars"
```

When running the `pre_plan` and `pre_apply` commands the following environment variables are available
- `ENVIRONMENT`: if an environment argument is supplied to `atlantis plan` or `atlantis apply` this will
be the value of that argument. Else it will be `default`
- `ATLANTIS_TERRAFORM_VERSION`: local version of `terraform` or the version from `terraform_version` if specified, ex. `0.10.0`
- `WORKSPACE`: absolute path to the root of the project on disk

## Locking
When `plan` is run, the [project](#project) and [environment](#environment) are **Locked** until an `apply` succeeds **and** the pull request is merged.
This protects against concurrent modifications to the same set of infrastructure and prevents
users from seeing a `plan` that will be invalid if another pull request is merged.

To unlock the project and environment without completing an `apply` and merging, click the link
at the bottom of the plan comment to discard the plan and delete the lock.
Once a plan is discarded, you'll need to run `plan` again prior to running `apply`.

## Approvals
If you'd like to require pull requests to be approved prior to a user running `atlantis apply` simply run Atlantis with the `--require-approval` flag.
By default, no approval is required.

For more information on pull request reviews and approvals see: https://help.github.com/articles/about-pull-request-reviews/

## Production-Ready Deployment
### Install Terraform
`terraform` needs to be in the `$PATH` for Atlantis.
Download from https://www.terraform.io/downloads.html
```
unzip path/to/terraform_*.zip -d /usr/local/bin
```
Check that it's in your `$PATH`
```
$ terraform version
Terraform v0.10.0
```
If you want to use a different version of Terraform see [Terraform Versions](#terraform-versions)

### Hosting Atlantis
Atlantis needs to be hosted somewhere that github.com or your GitHub Enterprise installation can reach. Developers in your organization also need to be able to access Atlantis to view the UI and to delete locks.

By default Atlantis runs on port `4141`. This can be changed with the `--port` flag.

### Add GitHub Webhook
Once you've decided where to host Atlantis you can add it as a Webhook to GitHub.
If you already have a GitHub organization we recommend installing the webhook at the **organization level** rather than on each repository, however both methods will work.

> If you're not sure if you have a GitHub organization see https://help.github.com/articles/differences-between-user-and-organization-accounts/

If you're installing on the organization, navigate to your organization's page and click **Settings**.
If installing on a single repository, navigate to the repository home page and click **Settings**.
- Select **Webhooks** or **Hooks** in the sidebar
- Click **Add webhook**
- set **Payload URL** to `http://$URL/events` where `$URL` is where Atlantis is hosted. **Be sure to add `/events`**
- set **Content type** to `application/json`
- leave **Secret** blank
- select **Let me select individual events**
- check the boxes
	- **Pull request review**
	- **Push**
	- **Issue comment**
	- **Pull request**
- leave **Active** checked
- click **Add webhook**

### Create a GitHub Token
We recommend creating a new user in GitHub named **atlantis** that performs all API actions however you can use any user.
Once you've created the user (or have decided to use an existing user) you need to create a personal access token.
- follow [https://help.github.com/articles/creating-a-personal-access-token-for-the-command-line/#creating-a-token](https://help.github.com/articles/creating-a-personal-access-token-for-the-command-line/#creating-a-token)
- copy the access token

### Start Atlantis
Now you're ready to start Atlantis! Run
```
$ atlantis server --atlantis-url $URL --gh-username $USERNAME --gh-token $TOKEN
2049/10/6 00:00:00 [WARN] server: Atlantis started - listening on port 4141
```

- `$URL` is the URL that Atlantis can be reached at
- `$USERNAME` is the GitHub username you generated the token for
- `$TOKEN` is the access token you created. If you don't want this to be passed in as an argument for security reasons you can specify it in a config file (see [Configuration](#configuration)) or as an environment variable: `ATLANTIS_GH_TOKEN`

Atlantis is now running!
**We recommend running it under something like Systemd or Supervisord.**

### Testing Out Atlantis

If you'd like to test out Atlantis before running it on your own repositories you can fork our example repo.

- Fork https://github.com/hootsuite/atlantis-example
- If you didn't add the Webhook as to your organization add Atlantis as a Webhook to the forked repo (see [Add GitHub Webhook](#add-github-webhook))
- Now that Atlantis can receive events you should be able to comment on a pull request to trigger Atlantis. Create a pull request
	- Click **Branches** on your forked repo's homepage
	- click the **New pull request** button next to the `example` branch
	- Change the `base` to `{your-repo}/master`
	- click **Create pull request**
- Now you can test out Atlantis
	- Create a comment `atlantis help` to see what commands you can run from the pull request
	- `atlantis plan` will run `terraform plan` behind the scenes. You should see the output commented back on the pull request. You should also see some logs show up where you're running `atlantis server`
	- `atlantis apply` will run `terraform apply`. Since our pull request creates a `null_resource` (which does nothing) this is safe to do.

## Server Configuration
Atlantis configuration can be specified via command line flags or a YAML config file.
The `gh-token` flag can also be specified via an `ATLANTIS_GH_TOKEN` environment variable.
Config file values are overridden by environment variables which in turn are overridden by flags.

To use a yaml config file, run atlantis with `--config /path/to/config.yaml`.
The keys of your config file should be the same as the flag, ex.
```yaml
---
gh-token: ...
log-level: ...
```

To see a list of all flags and their descriptions run `atlantis server --help`

## AWS Credentials
Atlantis simply shells out to `terraform` so you don't need to do anything special with AWS credentials.
As long as `terraform` works where you're hosting Atlantis, then Atlantis will work.
See https://www.terraform.io/docs/providers/aws/#authentication for more detail.

### Assume Role Session Names
Atlantis provides the ability to use AWS's Assume Role and **dynamically name the session** with the GitHub username of whoever commented `atlantis apply`.

This is used at Hootsuite so AWS API actions can be correlated with a specific user.
To take advantage of this feature, simply set the `--aws-assume-role-arn` flag to the
role to be assumed: `arn:aws:iam::ACCOUNT_ID:role/ROLE_NAME`.

If you're using Terraform's [built-in support](https://www.terraform.io/docs/providers/aws/#assume-role) for assume role then there is no need to set this flag (unless you also want your sessions to take the name of the GitHub user).

## Glossary
#### Project
A Terraform project. Multiple projects can be in a single GitHub repo.
We identify a project by its repo **and** the path to the root of the project within that repo.

#### Environment
A Terraform environment. See [terraform docs](https://www.terraform.io/docs/state/environments.html) for more information.
