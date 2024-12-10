---
title: "Putting The Dev Into DevOps: Why Your Developers Should Write Terraform Too"
lang: en-US
---

# Putting The Dev Into DevOps: Why Your Developers Should Write Terraform Too

::: info
This post was originally written on August 29th, 2018

Original post: <https://medium.com/runatlantis/putting-the-dev-into-devops-why-your-developers-should-write-terraform-too-d3c079dfc6a8>
:::

[Terraform](https://www.terraform.io/) is an amazing tool for provisioning infrastructure. Terraform enables your operators to perform their work faster and more reliably.

**But if only your ops team is writing Terraform, you're missing out.**

Terraform is not just a tool that makes ops teams more effective. Adopting Terraform is an opportunity to turn all of your developers into operators (at least for smaller tasks). This can make your entire engineering team more effective and create a better relationship between developers and operators.

### Quick Aside — What is Terraform?

Terraform is two things. It's a language for describing infrastructure:

```tf
resource "aws_instance" "example" {
  ami           = "ami-2757f631"
  instance_type = "t2.micro"
}
```

And it's a CLI tool that reads Terraform code and makes API calls to AWS (or any other cloud provider) to provision that infrastructure.

In this example, we're using the CLI to run `terraform apply` which will create an EC2 instance:

```sh
$ terraform apply

Terraform will perform the following actions:

  # aws_instance.example
  + aws_instance.example
      ami:              "ami-2757f631"
      instance_type:    "t2.micro"
      ...

Plan: 1 to add, 0 to change, 0 to destroy.

Do you want to perform these actions?
  Terraform will perform the actions described above.
  Only 'yes' will be accepted to approve.

  Enter a value: yes

aws_instance.example: Creating...
  ami:              "" => "ami-2757f631"
  instance_type:    "" => "t2.micro"
  ...

aws_instance.example: Still creating... (10s elapsed)
aws_instance.example: Creation complete

Apply complete! Resources: 1 added, 0 changed, 0 destroyed.
```

## Terraform Adoption From A Dev's Perspective

Adopting Terraform is great for your operations team's effectiveness but it doesn't change much for devs. Before Terraform adoption, devs typically interacted with an ops team like this:

![](/blog/putting-the-dev-into-devops/pic1.webp)

1. **Dev: Creates ticket asking for some ops work**
2. **Dev: Waits**
3. _Ops: Looks at ticket when in queue_
4. _Ops: Does work_
5. _Ops: Updates ticket_
6. **Dev: Continues their work**

After the Ops team adopts Terraform, the workflow from a dev's perspective is the same!

![](/blog/putting-the-dev-into-devops/pic2.webp)

1. **Dev: Creates ticket asking for some ops work**
2. **Dev: Waits**
3. _Ops: Looks at ticket when in queue_
4. _Ops: Does work. This time using Terraform (TF)_
5. _Ops: Updates ticket_
6. **Dev: Continues their work**

With Terraform, there's less of Step 2 (Dev: Waits) but apart from that, not much has changed.

> If only ops is writing Terraform, your developers' experience is the same.

## Devs Want To Help

Developers would love to help out with operations work. They know that for small changes they should be able to do the work themselves (with a review from ops). For example:

- Adding a new security group rule
- Increasing the size of an autoscaling group
- Using a larger instance because their app needs more memory

Developers could make all of these changes because they're small and well defined. Also, previous examples of doing the same thing can guide them.

## ...But Often They're Not Allowed

In many organizations, devs are locked out of the cloud console.

![](/blog/putting-the-dev-into-devops/pic3.webp)

They might be locked out for good reasons:

- Security — You can do a lot of damage with full access to a cloud console
- Compliance — Maybe your compliance requires only certain groups to have access
- Cost — Devs might spin up some expensive resources and then forget about them

Even if they have access, operations can be complicated:

- It's often difficult to do seemingly simple things (think adding a security group rule that also requires peering VPCs). This means that just having access sometimes isn't enough. Devs might need help from an expert to get things done.

## Enter Terraform

With Terraform, everything changes. Or at least it can.

Now Devs can see in code how infrastructure is built. They can see the exact spot where security group rules are configured:

```tf
resource "aws_security_group_rule" "allow_all" {
  type              = "ingress"
  from_port         = 0
  to_port           = 65535
  protocol          = "tcp"
  cidr_blocks       = ["0.0.0.0/0"]
  security_group_id = "sg-123456"
}

resource "aws_security_group_rule" "allow_office" {
  ...
}
```

Or where the size of the autoscaling group is set:

```tf
resource "aws_autoscaling_group" "asg" {
  name               = "my-asg"
  max_size           = 5
  desired_capacity   = 4
  min_size           = 2
  ...
}
```

Devs understand code (surprise!) so it's a lot easier for them to make those small changes.

Here's the new workflow:

![](/blog/putting-the-dev-into-devops/pic4.webp)

1. **Dev: Writes Terraform code**
2. **Dev: Creates pull request**
3. _Ops: Reviews pull request_
4. **Dev: Applies the change with Terraform (TF)**
5. **Dev: Continues their work**

Now:

- Devs are making small changes themselves. This saves time and increases the speed of the whole engineering organization.
- Devs can see exactly what is required to make the change. This means there's less back and forth over a ticket: “Okay so I know you need the security group opened between server A and B, but on which ports and with which protocol?”
- Devs start to see how infrastructure is built. This increases cooperation between dev and ops because they can understand each other's work.

Great! But there's another problem.

## Devs Are Locked Out Of Terraform Too

In order to execute Terraform you need to have cloud credentials! It's really hard to write Terraform without being able to run `terraform init` and `terraform plan`, for the same reason it would be hard to write code if you could never run it locally!

So are we back at square one?

## Enter Atlantis

[Atlantis](https://www.runatlantis.io/) is an [open source](https://github.com/runatlantis/atlantis) tool for running Terraform from pull requests. With Atlantis, Terraform is run on a separate server (Atlantis is self-hosted) so you don't need to give out credentials to everyone. Access is controlled through pull request approvals.

Here's what the workflow looks like:

### Step 1 — Create a Pull Request

A developer creates a pull request with their change to add a security group rule.

![](/blog/putting-the-dev-into-devops/pic5.webp)

### Step 2 — Atlantis Runs Terraform Plan

Atlantis automatically runs `terraform plan` and comments back on the pull request with the output. Now developers can fix their Terraform errors before asking for a review.

![](/blog/putting-the-dev-into-devops/pic6.webp)

### Step 3 — Fix The Terraform

The developer pushes a new commit that fixes their error and Atlantis comments back with the valid `terraform plan` output. Now the developer can verify that the plan output looks good.

![](/blog/putting-the-dev-into-devops/pic7.webp)

### Step 4 — Get Approval

You'll probably want to run Atlantis with the --require-approval flag that requires pull requests to be Approved before running atlantis apply.

![](/blog/putting-the-dev-into-devops/pic8.webp)

### Step 4a — Actually Get Approval

An operator can now come along and review the changes and the output of `terraform plan`. This is much faster than doing the change themselves.

![](/blog/putting-the-dev-into-devops/pic9.webp)

### Step 5 — Apply

To apply the changes, the developer or operator comments “atlantis apply”.

![](/blog/putting-the-dev-into-devops/pic10.webp)

## Success

Now we've got a workflow that makes everyone happy:

- Devs can write Terraform and iterate on the pull request until the `terraform plan` looks good
- Operators can review pull requests and approve the changes before they're applied

Now developers can make small operations changes and learn more about how infrastructure is built. Everyone can work more effectively and with a shared understanding that enhances collaboration.

## Does It Work In Practice?

Atlantis has been used by my previous company, Hootsuite, for over 2 years. It's used daily by 20 operators but it's also used occasionally by over 60 developers!
Another company uses Atlantis to manage 600+ Terraform repos collaborated on by over 300 developers and operators.

## Next Steps

- If you'd like to learn more about Terraform, check out HashiCorp's [Introduction to Terraform](https://developer.hashicorp.com/terraform/intro)
- If you'd like to try out Atlantis, go to <www.runatlantis.io>
- If you have any questions, reach out to me on Twitter ([at]lkysow) or in the comments below.

## Credits

- Thanks to [Seth Vargo](https://medium.com/@sethvargo) for his talk [Version-Controlled Infrastructure with GitHub](https://www.youtube.com/watch?v=2TWqi7dLSro) that inspired a lot of this post.
- Thanks to Isha for reading drafts of this post.
- Icons in graphics from made by [Freepik](http://freepik.com/) from [Flaticon](https://www.flaticon.com/) and licensed by [CC 3.0](https://creativecommons.org/licenses/by/3.0/)
