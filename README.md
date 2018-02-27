# Atlantis

<p align="center">
  <img src="./docs/atlantis-logo.png" alt="Atlantis Logo"/><br><br>
  A unified workflow for collaborating on Terraform through GitHub and GitLab
</p>

## Walkthrough Video
[![Atlantis Walkthrough](./docs/atlantis-walkthrough-icon.png)](https://www.youtube.com/watch?v=TmIPWda0IKg)

Read about [Why We Built Atlantis](https://medium.com/runatlantis/introducing-atlantis-6570d6de7281)

[![CircleCI](https://circleci.com/gh/runatlantis/atlantis/tree/master.svg?style=shield)](https://circleci.com/gh/runatlantis/atlantis/tree/master)
[![SuperDopeBadge](https://img.shields.io/badge/Hightower-extra%20dope-b9f2ff.svg)](https://twitter.com/kelseyhightower/status/893260922222813184)
[![Slack Status](https://thawing-headland-22460.herokuapp.com/badge.svg)](https://thawing-headland-22460.herokuapp.com)

* [Features](#features)
* [Atlantis Works With](#atlantis-works-with)
* [Getting Started](#getting-started)
* [Pull/Merge Request Commands](#pullmerge-request-commands)
* [Project Structure](#project-structure)
* [Workspaces/Environments](#workspacesenvironments)
* [Terraform Versions](#terraform-versions)
* [Project-Specific Customization](#project-specific-customization)
* [Locking](#locking)
* [Approvals](#approvals)
* [Production-Ready Deployment](#production-ready-deployment)
    * [Docker](#docker)
* [Server Configuration](#server-configuration)
* [AWS Credentials](#aws-credentials)
* [Glossary](#glossary)
    * [Project](#project)
    * [Workspace/Environment](#workspaceenvironment)
* [FAQ](#faq)
* [Contributing](#contributing)
* [Credits](#credits)

## Features
➜ Collaborate on Terraform with your team
- Run terraform `plan` and `apply` **from GitHub pull requests** so everyone can review the output
- **Lock workspaces** until pull requests are merged to prevent concurrent modification and confusion

➜ Developers can write Terraform safely
- **No need to distribute AWS credentials** to your whole team. Developers can submit Terraform changes and run `plan` and `apply` directly from the pull/merge request
- Optionally, require a **review and approval** prior to running `apply`

➜ Also
- Support **multiple versions of Terraform** with a simple project config file

## Atlantis Works With
* GitHub (public, private or enterprise) and GitLab (public, private or enterprise)
* Any Terraform version (see [Terraform Versions](#terraform-version))
* Can be run with a [single binary](https://github.com/runatlantis/atlantis/releases) or with our [Docker image](https://hub.docker.com/r/runatlantis/atlantis/)
* Any repository structure

## Getting Started
Download from [https://github.com/runatlantis/atlantis/releases](https://github.com/runatlantis/atlantis/releases)

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

## Pull/Merge Request Commands
Atlantis currently supports three commands that can be run via pull request comments (or merge request comments on GitLab):

![Help Command](./docs/pr-comment-help.png)
#### `atlantis help`
View help

---
![Plan Command](./docs/pr-comment-plan.png)
#### `atlantis plan [options] -- [terraform plan flags]`
Runs `terraform plan` for the changes in this pull request.

Options:
* `-d directory` Which directory to run plan in relative to root of repo. Use `.` for root. If not specified, will attempt to run plan for all Terraform projects we think were modified in this changeset.
* `-w workspace` Switch to this [Terraform workspace](https://www.terraform.io/docs/state/workspaces.html) before planning. Defaults to `default`. If not using Terraform workspaces you can ignore this.
* `--verbose` Append Atlantis log to comment.

Additional Terraform flags:

If you need to run `terraform plan` with additional arguments, like `-target=resource` or `-var 'foo-bar'`
you can append them to the end of the comment after `--`, ex.
```
atlantis plan -d dir -- -var 'foo=bar'
```
If you always need to append a certain flag, see [Project-Specific Customization](#project-specific-customization).

---
![Apply Command](./docs/pr-comment-apply.png)
#### `atlantis apply [options] -- [terraform apply flags]`
Runs `terraform apply` for the plans that match the directory and workspace.

Options:
* `-d directory` Apply the plan for this directory, relative to root of repo. Use `.` for root. If not specified, will run apply against all plans created for this workspace.
* `-w workspace` Apply the plan for this [Terraform workspace](https://www.terraform.io/docs/state/workspaces.html). Defaults to `default`. If not using Terraform workspaces you can ignore this.
* `--verbose` Append Atlantis log to comment.

Additional Terraform flags:

Same as with `atlantis plan`.

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
-  one folder per set of configuration
```
.
├── staging
│   ├── main.tf
|   └── ...
└── production
    ├── main.tf
    └── ...
```
-  using `env/{env}.tfvars` to define workspace specific variables. This works in both multi-project repos and single-project repos.
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
With the above project structure you can de-duplicate your Terraform code between workspaces/environments without requiring extensive use of modules. At Hootsuite we found this project format to be very successful and use it in all of our 100+ Terraform repositories.

## Workspaces/Environments
Terraform introduced [Workspaces](https://www.terraform.io/docs/state/workspaces.html) in 0.9. They allow for
> a single directory of Terraform configuration to be used to manage multiple distinct sets of infrastructure resources

If you're using a Terraform version >= 0.9.0, Atlantis supports workspaces through the `-w` flag.
For example,
```
atlantis plan -w staging
```

If a workspace is specified, Atlantis will use `terraform workspace select {workspace}` prior to running `terraform plan` or `terraform apply`.

If you're using the `env/{env}.tfvars` [project structure](#project-structure) we will also append `-tfvars=env/{env}.tfvars` to `plan` and `apply`.

If no workspace is specified, we'll use the `default` workspace by default.
This replicates Terraform's default behaviour which also uses the `default` workspace.

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
    - "-tfvars=myvars.tfvars"
```

When running the `pre_plan`, `post_plan`, `pre_apply`, and `post_apply` commands the following environment variables are available
- `WORKSPACE`: if a workspace argument is supplied to `atlantis plan` or `atlantis apply`, ex `atlantis plan -w staging`, this will
be the value of that argument. Else it will be `default`
- `ATLANTIS_TERRAFORM_VERSION`: local version of `terraform` or the version from `terraform_version` if specified, ex. `0.10.0`
- `DIR`: absolute path to the root of the project on disk

## Locking
When `plan` is run, the [project](#project) and [workspace](#workspaceenvironment) (**but not the whole repo**) are **Locked** until an `apply` succeeds **and** the pull request/merge request is merged.
This protects against concurrent modifications to the same set of infrastructure and prevents
users from seeing a `plan` that will be invalid if another pull request is merged.

If you have multiple directories inside a single repository, only the directory will be locked. Not the whole repo.

To unlock the project and workspace without completing an `apply` and merging, click the link
at the bottom of the plan comment to discard the plan and delete the lock.
Once a plan is discarded, you'll need to run `plan` again prior to running `apply` when you go back to that pull request.

## Approvals
If you'd like to require pull/merge requests to be approved prior to a user running `atlantis apply` simply run Atlantis with the `--require-approval` flag.
By default, no approval is required.

For more information on GitHub pull request reviews and approvals see: https://help.github.com/articles/about-pull-request-reviews/

For more information on GitLab merge request reviews and approvals (only supported on GitLab Enterprise) see: https://docs.gitlab.com/ee/user/project/merge_requests/merge_request_approvals.html.

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
Atlantis needs to be hosted somewhere that github.com/gitlab.com or your GitHub/GitLab Enterprise installation can reach. Developers in your organization also need to be able to access Atlantis to view the UI and to delete locks.

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
- leave **Secret** blank or set this to a random key (https://www.random.org/strings/). If you set it, you'll need to use the `--gh-webhook-secret` option when you start Atlantis
- select **Let me select individual events**
- check the boxes
	- **Pull request review**
	- **Push**
	- **Issue comment**
	- **Pull request**
- leave **Active** checked
- click **Add webhook**

### Add GitLab Webhook
If you're using GitLab, navigate to your project's home page in GitLab
- Click **Settings > Integrations** in the sidebar
- set **URL** to `http://$URL/events` where `$URL` is where Atlantis is hosted. **Be sure to add `/events`**
- leave **Secret Token** blank or set this to a random key (https://www.random.org/strings/). If you set it, you'll need to use the `--gitlab-webhook-secret` option when you start Atlantis
- check the boxes
    - **Push events**
    - **Comments**
    - **Merge Request events**
- leave **Enable SSL verification** checked
- click **Add webhook**

### Create a GitHub Token
We recommend creating a new user in GitHub named **atlantis** that performs all API actions, however you can use any user.
Once you've created the user (or have decided to use an existing user) you need to create a personal access token.
- follow [https://help.github.com/articles/creating-a-personal-access-token-for-the-command-line/#creating-a-token](https://help.github.com/articles/creating-a-personal-access-token-for-the-command-line/#creating-a-token)
- copy the access token

### Create a GitLab Token
We recommend creating a new user in GitLab named **atlantis** that performs all API actions, however you can use any user.
Once you've created the user (or have decided to use an existing user) you need to create a personal access token.
- follow [https://docs.gitlab.com/ce/user/profile/personal_access_tokens.html#creating-a-personal-access-token](https://docs.gitlab.com/ce/user/profile/personal_access_tokens.html#creating-a-personal-access-token)
- create a token with **api** scope
- copy the access token

### Start Atlantis
Now you're ready to start Atlantis!

If you're using GitHub, run:
```
$ atlantis server --atlantis-url $URL --gh-user $USERNAME --gh-token $TOKEN --gh-webhook-secret $SECRET
2049/10/6 00:00:00 [WARN] server: Atlantis started - listening on port 4141
```

If you're using GitLab, run:
```
$ atlantis server --atlantis-url $URL --gitlab-user $USERNAME --gitlab-token $TOKEN --gitlab-webhook-secret $SECRET
2049/10/6 00:00:00 [WARN] server: Atlantis started - listening on port 4141
```

- `$URL` is the URL that Atlantis can be reached at
- `$USERNAME` is the GitHub/GitLab username you generated the token for
- `$TOKEN` is the access token you created. If you don't want this to be passed in as an argument for security reasons you can specify it in a config file (see [Configuration](#configuration)) or as an environment variable: `ATLANTIS_GH_TOKEN` or `ATLANTIS_GITLAB_TOKEN`
- `$SECRET` is the random key you used for the webhook secret. If you left the secret blank then don't specify this flag. If you don't want this to be passed in as an argument for security reasons you can specify it in a config file (see [Configuration](#configuration)) or as an environment variable: `ATLANTIS_GH_WEBHOOK_SECRET` or `ATLANTIS_GITLAB_WEBHOOK_SECRET`

Atlantis is now running!
**We recommend running it under something like Systemd or Supervisord.**

### Docker
Atlantis also ships inside a docker image. Run the docker image:

```bash
docker run runatlantis/atlantis:latest server <required options>
```

#### Usage
If you need to modify the Docker image that we provide, for instance to add a specific version of Terraform, you can do something like this:

* Create a custom docker file
```bash
vim Dockerfile-custom
```

```dockerfile
FROM runatlantis/atlantis

# copy a terraform binary of the version you need
COPY terraform /usr/local/bin/terraform
```

* Build docker image

```bash
docker build -t {YOUR_DOCKER_ORG}/atlantis-custom -f Dockerfile-custom .
```

* Run docker image

```bash
docker run {YOUR_DOCKER_ORG}/atlantis-custom server --gh-user=GITHUB_USERNAME --gh-token=GITHUB_TOKEN
```


### Testing Out Atlantis on GitHub

If you'd like to test out Atlantis before running it on your own repositories you can fork our example repo.

- Fork https://github.com/runatlantis/atlantis-example
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
The `gh-token` and `gitlab-token` flags can also be specified via the `ATLANTIS_GH_TOKEN` and `ATLANTIS_GITLAB_TOKEN` environment variables respectively.
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

### Multiple AWS Accounts
Atlantis supports multiple AWS accounts through the use of Terraform's
[AWS Authentication](https://www.terraform.io/docs/providers/aws/#authentication).

If you're using the [Shared Credentials file](https://www.terraform.io/docs/providers/aws/#shared-credentials-file)
you'll need to ensure the server that Atlantis is executing on has the corresponding credentials file.

If you're using [Assume role](https://www.terraform.io/docs/providers/aws/#assume-role)
you'll need to ensure that the credentials file has a `default` profile that is able
to assume all required roles.

[Environment variables](https://www.terraform.io/docs/providers/aws/#environment-variables) authentication
won't work for multiple accounts since Atlantis wouldn't know which environment variables to execute
Terraform with.

### Assume Role Session Names
Atlantis injects the Terraform variable `atlantis_user` and sets it to the GitHub username of
the user that is running the Atlantis command. This can be used to dynamically name the assume role
session. This is used at Hootsuite so AWS API actions can be correlated with a specific user.

To take advantage of this feature, use Terraform's [built-in support](https://www.terraform.io/docs/providers/aws/#assume-role) for assume role
and use the `atlantis_user` terraform variable

```hcl
provider "aws" {
  assume_role {
    role_arn     = "arn:aws:iam::ACCOUNT_ID:role/ROLE_NAME"
    session_name = "${var.atlantis_user}"
  }
}

# need to define the atlantis_user variable to avoid terraform errors
variable "atlantis_user" {
  default = "atlantis_user"
}
```

If you're also using the [S3 Backend](https://www.terraform.io/docs/backends/types/s3.html)
make sure to add the `role_arn` option:

```hcl
terraform {
  backend "s3" {
    bucket   = "mybucket"
    key      = "path/to/my/key"
    region   = "us-east-1"
    role_arn = "arn:aws:iam::ACCOUNT_ID:role/ROLE_NAME"
    # can't use var.atlantis_user as the session name because
    # interpolations are not allowed in backend configuration
    # session_name = "${var.atlantis_user}" WON'T WORK
  }
}
```

Terraform doesn't support interpolations in backend config so you will not be
able to use `session_name = "${var.atlantis_user}"`. However, the backend assumed
role is only used for state-related API actions. Any other API actions will be performed using
the assumed role specified in the `aws` provider and will have the session named as the GitHub user.

## Glossary
#### Project
A Terraform project. Multiple projects can be in a single GitHub repo.
We identify a project by its repo **and** the path to the root of the project within that repo.

#### Workspace/Environment
A Terraform workspace. See [terraform docs](https://www.terraform.io/docs/state/workspaces.html) for more information.

## FAQ
**Q: Does Atlantis affect Terraform [remote state](https://www.terraform.io/docs/state/remote.html)?**

A: No. Atlantis does not interfere with Terraform remote state in any way. Under the hood, Atlantis is simply executing `terraform plan` and `terraform apply`.

**Q: How does Atlantis locking interact with Terraform [locking](https://www.terraform.io/docs/state/locking.html)?**

A: Atlantis provides locking of pull requests that prevents concurrent modification of the same infrastructure (Terraform project) whereas Terraform locking only prevents two concurrent `terraform apply`'s from happening. 

Terraform locking can be used alongside Atlantis locking since Atlantis is simply executing terraform commands.

**Q: How to run Atlantis in high availability mode? Does it need to be?**

A: Atlantis server can easily be run under the supervision of a init system like `upstart` or `systemd` to make sure `atlantis server` is always running. 

Atlantis currently stores all locking and Terraform plans locally on disk under the `--data-dir` directory (defaults to `~/.atlantis`). Because of this there is currently no way to run two or more Atlantis instances concurrently.

However, if you were to lose the data, all you would need to do is run `atlantis plan` again on the pull requests that are open. If someone tries to run `atlantis apply` after the data has been lost then they will get an error back, so they will have to re-plan anyway.

**Q: How to add SSL to Atlantis server?**

A: First, you'll need to get a public/private key pair to serve over SSL.
These need to be in a directory accessible by Atlantis. Then start `atlantis server` with the `--ssl-cert-file` and `--ssl-key-file` flags.
See `atlantis server --help` for more information.


## Contributing
Want to contribute? Check out [CONTRIBUTING](https://github.com/runatlantis/atlantis/blob/master/CONTRIBUTING.md).

Thank you to these awesome contributors!
- [@nicholas-wu-hs](https://github.com/nicholas-wu-hs)
- [@nadavshatz](https://github.com/nadavshatz)
- [@jwieringa](https://github.com/jwieringa)
- [@suhussai](https://github.com/suhussai)
- [@mootpt](https://github.com/mootpt)
- [@codec](https://github.com/codec)
- [@nick-hollingsworth-hs](https://github.com/nick-hollingsworth-hs)
- [@mpmsimo](https://github.com/mpmsimo)
- [@hussfelt](https://github.com/hussfelt)

## Credits
* Atlantis Logo: Icon made by [freepik](https://www.flaticon.com/authors/freepik) from www.flaticon.com

