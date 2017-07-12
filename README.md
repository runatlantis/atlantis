# atlantis
[![CircleCI](https://circleci.com/gh/hootsuite/atlantis/tree/master.svg?style=shield&circle-token=08bf5b34233b0e168a9dd73e01cafdcf7dc4bf16)](https://circleci.com/gh/hootsuite/atlantis/tree/master)

A unified workflow for collaborating on Terraform through GitHub.

## Features
-

## Getting Started
// todo: atlantis bootstrap workflow

## Configuration
Atlantis configuration can be specified via cli flags or a yaml config file.
The `gh-password` flag can also be specified via an `ATLANTIS_GH_PASSWORD` environment variable.
Config file values are overridden by environment variables which in turn are overridden by flags.

To use a yaml config file, run atlantis with `--config /path/to/config.yaml`.
The keys of your config file should be the same as the flag, ex.
```yaml
---
log-level: debug
```

To see a list of all flags and their descriptions run `atlantis server -h`

### AWS Credentials
Atlantis handles AWS credentials in the same way as Terraform.
It looks in the regular places for AWS credentials in this order:
1. `AWS_ACCESS_KEY_ID` and `AWS_SECRET_ACCESS_KEY` environment variables
2. The AWS credentials file located at `~/.aws/credentials`
3. Instance profile credentials if Atlantis is running on an EC2 instance

One additional feature of Atlantis is the ability to use the AWS Security Token Service (STS)
to assume a role (specified by the `--aws-assume-role-arn` flag) and **dynamically
name the session** with the GitHub username of whoever is running `atlantis apply`.
To take advantage of this feature, simply set the `--aws-assume-role-arn` flag.

## Project Locking
When `plan` is run, the project and environment are Locked until an `apply` succeeds and the pull request is merged.
This protects against concurrent modifications to the same set of infrastructure and prevents
users from seeing a `plan` that will be invalid if another pull request is merged.

To unlock the project and environment without completing an `apply`, click the link
at the bottom of each plan to discard the plan and delete the lock.

## Glossary
#### Project
A Terraform project. Multiple projects can be in a single GitHub repo.

#### Environment
A Terraform environment. See [terraform docs](https://www.terraform.io/docs/state/environments.html) for more information.
