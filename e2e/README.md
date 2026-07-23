# End to end tests

Tests run against the `runatlantis/atlantis-tests` fixture repository on hosted
VCS providers.

## Lifecycle scenarios

- `ScenarioPlanOnly` waits for autoplan and asserts configured project statuses
  and comment markers. It does not execute `ApplyCommand` fields.
- `ScenarioPlanThenApply` waits for autoplan, posts its configured apply command,
  rejects stale aggregate results, and asserts apply project statuses and a new
  apply comment marker.
- `ScenarioPlanThenReplanThenApply` pushes a second mutation, requires a new
  plan result and marker, and proves targeted apply consumed generation 2.
- `ScenarioPlanThenApplyExpectFailure` requires a new failed apply result and
  error marker while rejecting a configured built-in apply success marker.
- `ScenarioOnApplyLockPreservation` is an opt-in two-pull-request apply-lock
  lifecycle.

All scenarios share clone, branch creation, fixture mutation, push, pull-request
creation, and cleanup helpers. Plan-then-apply timeout diagnostics include the
pull-request URL, aggregate status, relevant project statuses, and recent comments.

## Configuration

### Gitlab

User: https://gitlab.com/atlantis-tests
Email: maintainers@runatlantis.io

To rotate a token:

1. Login to account
2. Select avatar -> Edit Profile -> Access tokens -> Add new token
3. Create a new token, and upload it to GitHub Actions as environment secret `ATLANTIS_GITLAB_TOKEN`.
