# v0.1.1 (Unreleased)
### Backwards Incompatibilities / Notes:
* `--aws-assume-role-arn` and `--aws-region` flags removed. Instead, to name the
assume role session with the GitHub username of the user running the Atlantis command
use the `atlantis_user` terraform variable alongside Terraform's
[built-in support](https://www.terraform.io/docs/providers/aws/#assume-role) for assume role
(see https://github.com/hootsuite/atlantis/blob/master/README.md#assume-role-session-names)
* Atlantis has a docker image now ([#123](https://github.com/hootsuite/atlantis/pull/123)). Here is how you can try it out:

```bash
docker run -it hootsuite/atlantis server --gh-user=GITHUB_USERNAME --gh-token=GITHUB_TOKEN
```

### Improvements
* Support for HTTPS cloning using GitHub username and token provided to atlantis server ([#117](https://github.com/hootsuite/atlantis/pull/117))
* Adding `post_plan` and `post_apply` commands ([#102](https://github.com/hootsuite/atlantis/pull/102))
* Adding the ability to verify webhook secret ([#120](https://github.com/hootsuite/atlantis/pull/120))

### Bug Fixes

### Features

### Core Changes
