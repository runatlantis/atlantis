# Checkout Strategy

You can configure how Atlantis checks out the code from your pull request via
the `--checkout-strategy` flag or the `ATLANTIS_CHECKOUT_STRATEGY` environment
variable that get passed to the `atlantis server` command.

Atlantis supports `branch` and `merge` strategies.

## Branch
If set to `branch` (the default), Atlantis will check out the source branch
of the pull request.

For example, given the following git history:
![Git History](./images/branch-strategy.png)

If the pull request was asking to merge `branch` into `main`,
Atlantis would check out `branch` at commit `C3`.

## Merge
The problem with the `branch` strategy, is that if users push branches that are
out of date with `main`, then their `terraform plan` could be deleting
some resources that were configured in the main branch.

For example, in the above diagram if commits `C4` and `C5` have modified the
terraform state and added new resources, then when Atlantis runs `terraform plan`
at commit `C3`, because the code doesn't have the changes from `C4` and `C5`,
Terraform will try to delete those resources.

To fix this, users could merge `main` into their branch, *or* you can run
Atlantis with `--checkout-strategy=merge`. With this strategy, Atlantis will
try to perform a merge locally by:

* Checking out the destination branch of the pull request (ex. `main`)
* Locally performing a `git merge {source branch}`
* Then running its Terraform commands

In this example, the code that Atlantis would be operating on would look like:
![Git History](./images/merge-strategy.png)
Where Atlantis is using its local commit `C6`.

:::tip NOTE
Atlantis doesn't actually commit this merge anywhere. It just uses it locally.
:::

:::warning
Atlantis only performs this merge during the `terraform plan` phase. If another
commit is pushed to `main` **after** Atlantis runs `plan`, nothing will happen.
:::

To optimize cloning time, Atlantis can perform a shallow clone by specifying the `--checkout-depth` flag. The cloning is performed in a following manner:

- Shallow clone of the default branch is performed with depth of `--checkout-depth` value of zero (full clone).
- `branch` is retrieved, including the same amount of commits.
- Merge base of the default branch and `branch` is checked for existence in the shallow clone.
- If the merge base is not present, it means that either of the branches are ahead of the merge base by more than `--checkout-depth` commits. In this case full repo history is fetched.

If the commit history often diverges by more than the default checkout depth then the `--checkout-depth` flag should be tuned to avoid full fetches.