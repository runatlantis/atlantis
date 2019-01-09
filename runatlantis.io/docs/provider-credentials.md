# Provider Credentials

## AWS
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
