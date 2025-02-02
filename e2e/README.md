# End to end tests

Tests run against actual repos in various VCS providers

## Configuration

### Gitlab

User: https://gitlab.com/atlantis-tests
Email: maintainers@runatlantis.io

To rotate token:
1. Login to account
2. Select avatar -> Edit Profile -> Access tokens -> Add new token
3. Create a new token, and upload it to Github Action as environment secret `ATLANTIS_GITLAB_TOKEN`.
