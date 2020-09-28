# Provider Credentials
Atlantis runs Terraform by simply executing `terraform plan` and `apply` commands
on the server Atlantis is hosted on.
Just like when you run Terraform locally, Atlantis needs credentials for your
specific provider.

It's up to you how you provide credentials for your specific provider to Atlantis:
* The Atlantis [Helm Chart](deployment.html#kubernetes-helm-chart) and 
    [AWS Fargate Module](deployment.html#aws-fargate) have their own mechanisms for provider
    credentials. Read their docs.
* If you're running Atlantis in a cloud then many clouds have ways to give cloud API access
  to applications running on them, ex:
    * [AWS EC2 Roles](https://registry.terraform.io/providers/hashicorp/aws/latest/docs) (Search for "EC2 Role")
    * [GCE Instance Service Accounts](https://www.terraform.io/docs/providers/google/provider_reference.html#configuration-reference)
* Many users set environment variables, ex. `AWS_ACCESS_KEY`, where Atlantis is running.
* Others create the necessary config files, ex. `~/.aws/credentials`, where Atlantis is running.
* Use the [HashiCorp Vault Provider](https://registry.terraform.io/providers/hashicorp/vault/latest/docs)
  to obtain provider credentials.

:::tip
As a general rule, if you can `ssh` or `exec` into the server where Atlantis is
running and run `terraform` commands like you would locally, then Atlantis will work.
:::


## AWS Specific Info

### Multiple AWS Accounts
Atlantis supports multiple AWS accounts through the use of Terraform's
[AWS Authentication](https://registry.terraform.io/providers/hashicorp/aws/latest/docs) (Search for "Authentication").

If you're using the [Shared Credentials file](https://registry.terraform.io/providers/hashicorp/aws/latest/docs) (Search for "Shared Credentials file")
you'll need to ensure the server that Atlantis is executing on has the corresponding credentials file.

If you're using [Assume role](https://registry.terraform.io/providers/hashicorp/aws/latest/docs) (Search for "Assume role")
you'll need to ensure that the credentials file has a `default` profile that is able
to assume all required roles.

Using multiple [Environment variables](https://registry.terraform.io/providers/hashicorp/aws/latest/docs) (Search for "Environment variables")
won't work for multiple accounts since Atlantis wouldn't know which environment variables to execute
Terraform with.

### Assume Role Session Names
If you're using Terraform < 0.12, Atlantis injects 5 Terraform variables that can be used to dynamically name the assume role session name.
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
| `-var` Argument                      | Description                                                                                                                            |
|--------------------------------------|----------------------------------------------------------------------------------------------------------------------------------------|
| `atlantis_user=lkysow`               | The VCS username of who is running the plan command.                                                                                   |
| `atlantis_repo=runatlantis/atlantis` | The full name of the repo the pull request is in. NOTE: This variable can't be used in the AWS session name because it contains a `/`. |
| `atlantis_repo_owner=runatlantis`    | The name of the **owner** of the repo the pull request is in.                                                                          |
| `atlantis_repo_name=atlantis`        | The name of the repo the pull request is in.                                                                                           |
| `atlantis_pull_num=200`              | The pull request number.                                                                                                               |

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

:::tip Why does this not work in TF >= 0.12?
In Terraform >= 0.12, you're not allowed to set any `-var` flags if those variables
aren't being used. Since we can't know if you're using these `atlantis_*` variables,
we can't set the `-var` flag.

You can still set these variables yourself using the `extra_args` configuration.
:::

## Next Steps
* If you want to configure Atlantis further, read [Configuring Atlantis](configuring-atlantis.html)
* If you're ready to use Atlantis, read [Using Atlantis](using-atlantis.html)
