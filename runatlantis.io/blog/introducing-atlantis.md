---
title: Introducing Atlantis
lang: en-US
---

# Introducing Atlantis

::: info
This post was originally written on September 11th, 2017

Original post: https://medium.com/runatlantis/introducing-atlantis-6570d6de7281
:::

We’re very excited to announce the open source release of Atlantis! Atlantis is a tool for
collaborating on Terraform that’s been in use at Hootsuite for over a year. The core
functionality of Atlantis enables developers and operators to run `terraform plan` and
`apply` directly from Terraform pull requests. Atlantis then comments back on the pull
request with the output of the commands:

![](/blog/intro/intro1.gif)

This is a simple feature, however it has had a massive effect on how our team writes Terraform.
By bringing a Terraform workflow to pull requests, Atlantis helped our Ops team collaborate
better on Terraform and also enabled our entire development team to write and execute Terraform safely.

Atlantis was built to solve two problems that arose at Hootsuite as we adopted Terraform:

### 1. Effective Collaboration
What’s the best way to collaborate on Terraform in a team setting?

### 2. Developers Writing Terraform
How can we enable our developers to write and apply Terraform safely?

## Effective Collaboration
When writing Terraform, there are a number of workflows you can follow. The simplest workflow is just using `master`:

![](/blog/intro/intro2.webp)

In this workflow, you work on `master` and run `terraform` locally.
The problem with this workflow is that there is no collaboration or code review.
So we start to use pull requests:

![](/blog/intro/intro3.webp)

We still run `terraform plan` locally, but once we’re satisfied with the changes we create a pull request for review. When the pull request is approved, we run `apply` locally.

This workflow is an improvement, but there are still problems. The first problem is that it’s hard to review just the diff on the pull request. To properly review a change, you really need to see the output from `terraform plan`.

![](/blog/intro/intro4.webp)

What looks like a small change...

![](/blog/intro/intro5.webp)

...can have a big plan

The second problem is that now it’s easy for `master` to get out of sync with what’s actually been applied. This can happen if you merge a pull request without running `apply` or if the `apply` has an error halfway through, you forget to fix it and then you merge to `master`. Now what’s in `master` isn’t actually what’s running on production. At best, this causes confusion the next time someone runs `terraform plan`. At worst, it causes an outage when someone assumes that what’s in `master` is actually running, and depends on it.

With the Atlantis workflow, these problems are solved:

![](/blog/intro/intro6.webp)

Now it’s easy to review changes because you see the `terraform plan` output on the pull request.

![](/blog/intro/intro7.webp)

Pull requests are easy to review since you can see the plan

It’s also easy to ensure that the pull request is `terraform apply`’d before merging to master because you can see the actual `apply` output on the pull request.

![](/blog/intro/intro8.webp)

So, Atlantis makes working on Terraform within an operations team much easier, but how does it help with getting your whole team to write Terraform?

## Developers Writing Terraform

Terraform usually starts out being used by the Ops team. As a result of using Terraform, the Ops team becomes much faster at making infrastructure changes, but the way developers request those changes remains the same: they use a ticketing system or chat to ask operations for help, the request goes into a queue and later Ops responds that the task is complete.

Soon however, the Ops team starts to realize that it’s possible for developers to make some of these Terraform changes themselves! There are some problems that arise though:

- Developers don’t have the credentials to actually run Terraform commands
- If you give them credentials, it’s hard to review what is actually being applied

With Atlantis, these problems are solved. All `terraform plan` and `apply` commands are run from the pull request. This means developers don’t need to have any credentials to run Terraform locally. Of course, this can be dangerous: how can you ensure developers (who might be new to Terraform) aren’t applying things they shouldn’t? The answer is code reviews and approvals.

Since Atlantis comments back with the `plan` output directly on the pull request, it’s easy for an operations engineer to review exactly what changes will be applied. And Atlantis can run in `require-approval` mode, that will require a GitHub pull request approval before allowing `apply` to be run:

![](/blog/intro/intro9.webp)

With Atlantis, developers are able to write and apply Terraform safely. They submit pull requests, can run `atlantis plan` until their change looks good and then get approval from Ops to `apply`.

Since the introduction of Atlantis at Hootsuite, we’ve had **78** contributors to our Terraform repositories, **58** of whom are developers (**75%**).

## Where we are now

Since the introduction of Atlantis at Hootsuite we’ve grown to 144 Terraform repositories [^1] that manage thousands of Amazon resources. Atlantis is used for every single Terraform change throughout our organization.

## Getting started with Atlantis

If you’d like to try out Atlantis for your team you can download the latest release from https://github.com/runatlantis/atlantis/releases. If you run `atlantis testdrive` you can get started in less than 5 minutes. To read more about Atlantis go to https://www.runatlantis.io/.

Check out our video for more information:

<iframe src="https://cdn.embedly.com/widgets/media.html?src=https%3A%2F%2Fwww.youtube.com%2Fembed%2FTmIPWda0IKg%3Ffeature%3Doembed&amp;url=http%3A%2F%2Fwww.youtube.com%2Fwatch%3Fv%3DTmIPWda0IKg&amp;image=https%3A%2F%2Fi.ytimg.com%2Fvi%2FTmIPWda0IKg%2Fhqdefault.jpg&amp;key=a19fcc184b9711e1b4764040d3dc5c07&amp;type=text%2Fhtml&amp;schema=youtube" allowfullscreen="" frameborder="0" height="480" width="640" title="Atlantis Walkthrough" class="fr n gh dv bg" scrolling="no"></iframe>

[^1]: We split our Terraform up into multiple states, each with its own repository (see [1], [2], [3]).

[1]: https://blog.gruntwork.io/how-to-manage-terraform-state-28f5697e68fa
[2]: https://charity.wtf/2016/03/30/terraform-vpc-and-why-you-want-a-tfstate-file-per-env/
[3]: https://www.nclouds.com/blog/terraform-multi-state-management/
