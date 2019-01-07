# Server Configuration
This documentation explains how to configure the Atlantis server and how to deal
with credentials.

[[toc]]

Configuration for `atlantis server` can be specified via command line flags, environment variables or a YAML config file.
Config file values are overridden by environment variables which in turn are overridden by flags.

## YAML
To use a yaml config file, run atlantis with `--config /path/to/config.yaml`.
The keys of your config file should be the same as the flag, ex.
```yaml
---
gh-token: ...
log-level: ...
```

## Environment Variables
All flags can be specified as environment variables. You need to convert the flag's `-`'s to `_`'s, uppercase all the letters and prefix with `ATLANTIS_`.
For example, `--gh-user` can be set via the environment variable `ATLANTIS_GH_USER`.

To see a list of all flags and their descriptions run `atlantis server --help`

::: warning
The flag `--atlantis-url` is set by the environment variable `ATLANTIS_ATLANTIS_URL` **NOT** `ATLANTIS_URL`.
:::

## AWS Credentials
Atlantis simply shells out to `terraform` so you don't need to do anything special with AWS credentials.
As long as `terraform` commands works where you're hosting Atlantis, then Atlantis will work.
See [https://www.terraform.io/docs/providers/aws/#authentication](https://www.terraform.io/docs/providers/aws/#authentication) for more detail.

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
Atlantis injects 5 Terraform variables that can be used to dynamically name the assume role session name.
Setting the `session_name` allows you to trace API calls made through Atlantis back to a specific
user and repo via CloudWatch:

```bash
provider "aws" {
  assume_role {
    role_arn     = "arn:aws:iam::ACCOUNT_ID:role/ROLE_NAME"
    session_name = "${var.atlantis_user}-${var.atlantis_repo_owner}-${var.atlantis_repo_name}-${var.atlantis_pull_num}"
  }
}
```

Atlantis runs `terraform` with the following variables:
| `-var` Argument                           | Description                                                                                                                           |
|-------------------------------------------|---------------------------------------------------------------------------------------------------------------------------------------|
| `atlantis_user=lkysow`               | The VCS username of who is running the plan command.                                                                                  |
| `atlantis_repo=runatlantis/atlantis` | The full name of the repo the pull request is in. NOTE: This variable can't be used in the AWS session name because it contains a `/`. |
| `atlantis_repo_owner=runatlantis`    | The name of the **owner** of the repo the pull request is in.                                                                         |
| `atlantis_repo_name=atlantis`        | The name of the repo the pull request is in.                                                                                          |
| `atlantis_pull_num=200`              | The pull request number.                                                                                                              |

If you want to use `assume_role` with Atlantis and you're also using the [S3 Backend](https://www.terraform.io/docs/backends/types/s3.html),
make sure to add the `role_arn` option:

```bash
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

## Repo Whitelist
Atlantis requires you to specify a whitelist of repositories it will accept webhooks from via the `--repo-whitelist` flag.

Notes:
* Accepts a comma separated list, ex. `definition1,definition2`
* Format is `{hostname}/{owner}/{repo}`, ex. `github.com/runatlantis/atlantis`
* `*` matches any characters, ex. `github.com/runatlantis/*` will match all repos in the runatlantis organization
* For Bitbucket Server: `{hostname}` is the domain without scheme and port, `{owner}` is the name of the project (not the key), and `{repo}` is the repo name

Examples:
* Whitelist `myorg/repo1` and `myorg/repo2` on `github.com`
  * `--repo-whitelist=github.com/myorg/repo1,github.com/myorg/repo2`
* Whitelist all repos under `myorg` on `github.com`
  * `--repo-whitelist='github.com/myorg/*'`
* Whitelist all repos in my GitHub Enterprise installation
  * `--repo-whitelist='github.yourcompany.com/*'`
* Whitelist all repositories
  * `--repo-whitelist='*'`
