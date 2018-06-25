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

## Next Steps
For more information on GitHub pull request reviews and approvals see: [https://help.github.com/articles/about-pull-request-reviews/](https://help.github.com/articles/about-pull-request-reviews/)

For more information on GitLab merge request reviews and approvals (only supported on GitLab Enterprise) see: [https://docs.gitlab.com/ee/user/project/merge_requests/merge_request_approvals.html](https://docs.gitlab.com/ee/user/project/merge_requests/merge_request_approvals.html).