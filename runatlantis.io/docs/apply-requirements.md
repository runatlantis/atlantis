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
Please be aware that in GitHub **any user with read permissions** can approve a pull request.

In GitLab, you [can set](https://docs.gitlab.com/ee/user/project/merge_requests/merge_request_approvals.html#editing-approvals) who is allowed to approve.
:::

## Next Steps
* For more information on GitHub pull request reviews and approvals see: [https://help.github.com/articles/about-pull-request-reviews/](https://help.github.com/articles/about-pull-request-reviews/)
* For more information on GitLab merge request reviews and approvals (only supported on GitLab Enterprise) see: [https://docs.gitlab.com/ee/user/project/merge_requests/merge_request_approvals.html](https://docs.gitlab.com/ee/user/project/merge_requests/merge_request_approvals.html).
* For more information on Bitbucket pull request reviews and approvas see: [https://confluence.atlassian.com/bitbucket/pull-requests-and-code-review-223220593.html](https://confluence.atlassian.com/bitbucket/pull-requests-and-code-review-223220593.html)