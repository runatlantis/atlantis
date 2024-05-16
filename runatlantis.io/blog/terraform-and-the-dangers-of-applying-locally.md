---
title: Terraform And The Dangers Of Applying Locally
lang: en-US
---

# Terraform And The Dangers Of Applying Locally

::: info
This post was originally written on July 13th, 2018

Original post: https://medium.com/runatlantis/terraform-and-the-dangers-of-applying-locally-543563782a73
:::

If you're using Terraform then at some point you've likely ran a `terraform apply` that reverted someone else's change!

Here's how that tends to happen:

## The Setup

Say we have two developers: Alice and Bob. Alice needs to add a new security group rule. She checks out a new branch, adds her rule and creates a pull request:

![](/blog/terraform-and-the-dangers-of-applying-locally/pic1.webp)

When she runs `terraform plan` locally she sees what she expects.

![](/blog/terraform-and-the-dangers-of-applying-locally/pic2.webp)

Meanwhile, Bob is working on an emergency fix. He checks out a new branch and adds a different security group rule called `emergency`:

![](/blog/terraform-and-the-dangers-of-applying-locally/pic3.webp)

And, because it's an emergency, he **immediately runs apply**:

![](/blog/terraform-and-the-dangers-of-applying-locally/pic4.webp)

Now back to Alice. She's just gotten approval on her pull request change and so she runs `terraform apply`:

![](/blog/terraform-and-the-dangers-of-applying-locally/pic5.webp)

Did you catch what happened? Did you notice that the `apply` deleted Bob's rule?

![](/blog/terraform-and-the-dangers-of-applying-locally/pic6.webp)

In this example, it wasn't too hard to see. However if the plan is much longer, or if the change is less obvious then it can be easy to miss.

## Possible Solutions

There are some ways to avoid this:

### Use terraform plan `-out`

If Alice had run `terraform plan -out plan.tfplan` then when she ran `terraform apply plan.tfplan` she would see:

![](/blog/terraform-and-the-dangers-of-applying-locally/pic7.webp)

The problem with this solution is that few people run `terraform plan` anymore, much less `terraform plan -out`!

<iframe src="https://cdn.embedly.com/widgets/media.html?type=text%2Fhtml&amp;key=a19fcc184b9711e1b4764040d3dc5c07&amp;schema=twitter&amp;url=https%3A//twitter.com/sethvargo/status/989979940098424832&amp;image=https%3A//i.embed.ly/1/image%3Furl%3Dhttps%253A%252F%252Fpbs.twimg.com%252Fprofile_images%252F808025120296013825%252FfrGuc14s_400x400.jpg%26key%3Da19fcc184b9711e1b4764040d3dc5c07" allowfullscreen="" frameborder="0" height="249" width="680" title="Seth Vargo on Twitter" class="fr n gh dv bg" scrolling="no"></iframe>

It's easier to just run `terraform apply` and humans will take the easier path most of the time.

### Wrap `terraform apply` to ensure up to date with `master`

Another possible solution is to write a wrapper script that ensures our branch is up to date with `master`. But this doesn't solve the problem of Bob running `apply` locally and not yet merging to `master`. In this case, Alice's branch would have been up to date with `master` but not the latest apply'd state.

### Be more disciplined!

What if everyone:

- ALWAYS created a branch, got a pull request review, merged to `master` and then ran apply. And also everyone
- ALWAYS checked to ensure their branch was rebased from `master`. And also everyone
- ALWAYS carefully inspected the `terraform plan` output and made sure it was exactly what they expected

...then we wouldn't have a problem!

Unfortunately this is not a real solution. We're all human and we're all going to make mistakes. Relying on people to follow a complicated process 100% of the time is not a solution because it doesn't work.

## Core Problem

The core problem is that everyone is applying from their own workstations and it's up to them to ensure that they're up to date and that they keep `master` up to date. This is like developers deploying to production from their laptops.

### What if, instead of applying locally, a remote system did the apply's?

This is why we built [Atlantis](https://www.runatlantis.io/) – an open source project for Terraform automation by pull request. You could also accomplished this with your own CI system or with [Terraform Enterprise](https://www.hashicorp.com/products/terraform). Here's how Atlantis solves this issue:

When Alice makes her change, she creates a pull request and Atlantis automatically runs `terraform plan` and comments on the pull request.

When Bob makes his change, he creates a pull request and Atlantis automatically runs `terraform plan` and comments on the pull request.

![](/blog/terraform-and-the-dangers-of-applying-locally/pic8.webp)

Atlantis also **locks the directory** to ensure that no one else can run `plan` or `apply` until Alice's plan has been intentionally deleted or she merges the pull request.

If Bob creates a pull request for his emergency change he'd see this error:

![](/blog/terraform-and-the-dangers-of-applying-locally/pic9.webp)

Alice can then comment `atlantis apply` and Atlantis will run the apply itself:

![](/blog/terraform-and-the-dangers-of-applying-locally/pic10.webp)

Finally, she merges the pull request and unlocks Bob's branch:

![](/blog/terraform-and-the-dangers-of-applying-locally/pic11.webp)

### But what if Bob ran `apply` locally?

In that case, Alice is still okay because when Atlantis ran `terraform plan` it used `-out`. If Alice tries to apply that plan, Terraform will give an error because the plan was generated against an old state.

### Why does Atlantis run `apply` on the branch and not after a merge to `master`?

We do this because `terraform apply` fails quite often, despite `terraform plan` succeeding. Usually it's because of a dependency issue between resources or because the cloud provider requires a certain format or a certain field to be set. Regardless, in practice we've found that `apply` fails a lot.

By locking the directory, we're essentially ensuring that the branch being `apply`'d is `"master"` since no one else can modify that state. We then get the benefit of being able to iterate on the pull request and push small fixes until we're sure that the changeset is `apply`'d. If `apply` failed after merging to `master`, we'd have to open new pull requests over and over again. There is definitely a tradeoff here, however we believe it's the right tradeoff.

## Conclusion

In conclusion, running `terraform apply` when you're working with a team of operators can be dangerous. Look to solutions like your own CI, Atlantis or Terraform Enterprise to ensure you're always working off the latest code that was `apply`'d.

If you'd like to try Atlantis, you can get started here: https://www.runatlantis.io/guide/
