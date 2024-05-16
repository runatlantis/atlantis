---
title: Atlantis 0.4.4 Now Supports Bitbucket
lang: en-US
---

# Atlantis 0.4.4 Now Supports Bitbucket

::: info
This post was originally written on July 25th, 2018

Original post: <https://medium.com/runatlantis/atlantis-0-4-4-now-supports-bitbucket-86c53a550b45>
:::

![](/blog/atlantis-0-4-4-now-supports-bitbucket/pic1.webp)

Atlantis is an [open source](https://github.com/runatlantis/atlantis) platform for using Terraform in teams. I'm happy to announce that the [latest release](https://github.com/runatlantis/atlantis/releases) of Atlantis (0.4.4) now supports both Bitbucket Cloud (bitbucket.org) **and** Bitbucket Server (aka Stash).

![](/blog/atlantis-0-4-4-now-supports-bitbucket/pic2.gif)

Atlantis now supports the three major Git hosts: GitHub, GitLab and Bitbucket. The rest of this post will talk about how to use Atlantis with Bitbucket.

## What is Atlantis?

Atlantis is a self-hosted application that listens for Terraform pull request events via webhooks. It runs `terraform plan` and `apply` remotely and comments back on the pull request with the output.

With Atlantis, you collaborate on the Terraform pull request itself instead of running `terraform apply` from your own computers which can be dangerous:

Check out <www.runatlantis.io> for more information.

## Getting Started

The easiest way to try out Atlantis with Bitbucket is to run Atlantis locally on your own computer. Eventually you'll want to deploy it as a standalone app but this is the easiest way to try it out. Follow [these instructions](https://www.runatlantis.io/guide/getting-started.html) to get Atlantis running locally.

Create a Pull Request
If you've got the Atlantis webhook configured for your repository and Atlantis is running, it's time to create a new pull request. I recommend adding a `null_resource` to one of your Terraform files for the the test pull request. It won't actually create anything so it's safe to use as a test.

Using the web editor, open up one of your Terraform files and add:

```tf
resource "null_resource" "example" {}
```

![](/blog/atlantis-0-4-4-now-supports-bitbucket/pic3.webp)

Click Commit and select **Create a pull request for this change**.

![](/blog/atlantis-0-4-4-now-supports-bitbucket/pic4.webp)

Wait a few seconds and then refresh. Atlantis should have automatically run `terraform plan` and commented back on the pull request:

![](/blog/atlantis-0-4-4-now-supports-bitbucket/pic5.webp)

Now it's easier for your colleagues to review the pull request because they can see the `terraform plan` output.

### Terraform Apply

Since all we're doing is adding a null resource, I think it's safe to run `terraform apply`. To do so, I add a comment to the pull request: `atlantis apply`:

![](/blog/atlantis-0-4-4-now-supports-bitbucket/pic6.webp)

Atlantis is listening for pull request comments and will run `terraform apply` remotely and comment back with the output:

![](/blog/atlantis-0-4-4-now-supports-bitbucket/pic7.webp)

### Pull Request Approvals

If you don't want anyone to be able to `terraform apply`, you can run Atlantis with `--require-approval` or add that setting to your [atlantis.yaml file](https://www.runatlantis.io/guide/atlantis-yaml-use-cases.html#requiring-approvals-for-production).

This will ensure that the pull request has been approved before someone can run `apply`.

## Other Features

### Customizable Commands

Apart from being able to `plan` and `apply` from the pull request, Atlantis also enables you to customize the exact commands that are run via an `atlantis.yaml` config file. For example to use the `-var-file` flag:

```yaml{14}
# atlantis.yaml
version: 2
projects:
- name: staging
  dir: "."
  workflow: staging

workflows:
  staging:
    plan:
      steps:
      - init
      - plan:
          extra_args: ["-var-file", "staging.tfvars"]
```

### Locking For Coordination

![](/blog/atlantis-0-4-4-now-supports-bitbucket/pic8.webp)

Atlantis will prevent other pull requests from running against the same directory as an open pull request so that each plan is applied atomically. Once the first pull request is merged, other pull requests are unlocked.

## Next Steps

If you're interested in using Atlantis with Bitbucket, check out our Getting Started docs. Happy Terraforming!
