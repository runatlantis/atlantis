# Apply Requirements

## Approved
If you'd like to require pull/merge requests to be approved prior to a user running `atlantis apply` simply run Atlantis with the `--require-approval` flag.
By default, no approval is required. If you want to configure this on a per-repo/project basis, for example to only require approvals for your production
configuration you must use an `atlantis.yaml` file:

```yaml
version: 2
projects:
- dir: .
  apply_requirements: [approved]
```

::: danger
A pull request approval might not be as secure as you'd expect:
* In GitHub **any user with read permissions** to the repo can approve a pull request.
* In GitLab, you [can set](https://docs.gitlab.com/ee/user/project/merge_requests/merge_request_approvals.html#editing-approvals) who is allowed to approve.
:::

::: tip Note
In Bitbucket Cloud (bitbucket.org), a user can approve their own pull request.
Atlantis does not count that as an approval and requires at least one user that
is not the author of the pull request to approve it.
:::

## Mergeable

You may also set the `mergeable` requirement for pull/merge requests. A PR/MR must be marked as `mergeable` before a user is able to run `atlantis apply`. This is helpful in prevent certain edge cases, such as applying stale code in an outdated PR which has passed all other checks. You can enable this feature by passing the `--require-mergeable` flag to Atlantis, or by using an `atlantis.yaml` file like this:

```yaml
version: 2
projects:
- dir: .
  apply_requirements: [mergeable, approved]
```

::: tip Note
The meaning of "mergeable" may be slightly different depending on your Git provider. In the case of GitHub and GitLab, it means that a PR/MR has no conflicts and satisfies all required approvals and checks. On BitBucket, it means that merging is possible and there are no conflicts.
:::

## Next Steps
* For more information on GitHub pull request reviews and approvals see: [https://help.github.com/articles/about-pull-request-reviews/](https://help.github.com/articles/about-pull-request-reviews/)
* For more information on GitLab merge request reviews and approvals (only supported on GitLab Enterprise) see: [https://docs.gitlab.com/ee/user/project/merge_requests/merge_request_approvals.html](https://docs.gitlab.com/ee/user/project/merge_requests/merge_request_approvals.html).
* For more information on Bitbucket pull request reviews and approvals see: [https://confluence.atlassian.com/bitbucket/pull-requests-and-code-review-223220593.html](https://confluence.atlassian.com/bitbucket/pull-requests-and-code-review-223220593.html)