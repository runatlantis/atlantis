# v0.1.2 (UNRELEASED)
## Features
* all flags passed to `atlantis plan` or `atlantis apply` will now be passed through to `terraform`. ([#131](https://github.com/hootsuite/atlantis/pull/131))

## Bug Fixes
* Fix command parsing when comment ends with newline. ([#131](https://github.com/hootsuite/atlantis/pull/131))
* Plan and Apply outputs are shown in new line. ([#132](https://github.com/hootsuite/atlantis/pull/132))

## Downloads
* [atlantis_darwin_amd64.zip](https://github.com/hootsuite/atlantis/releases/download/v0.1.2/atlantis_darwin_amd64.zip)
* [atlantis_linux_386.zip](https://github.com/hootsuite/atlantis/releases/download/v0.1.2/atlantis_linux_386.zip)
* [atlantis_linux_amd64.zip](https://github.com/hootsuite/atlantis/releases/download/v0.1.2/atlantis_linux_amd64.zip)
* [atlantis_linux_arm.zip](https://github.com/hootsuite/atlantis/releases/download/v0.1.2/atlantis_linux_arm.zip)

# v0.1.1
## Backwards Incompatibilities / Notes:
* `--aws-assume-role-arn` and `--aws-region` flags removed. Instead, to name the
assume role session with the GitHub username of the user running the Atlantis command
use the `atlantis_user` terraform variable alongside Terraform's
[built-in support](https://www.terraform.io/docs/providers/aws/#assume-role) for assume role
(see https://github.com/hootsuite/atlantis/blob/master/README.md#assume-role-session-names)
* Atlantis has a docker image now ([#123](https://github.com/hootsuite/atlantis/pull/123)). Here is how you can try it out:

```bash
docker run hootsuite/atlantis:v0.1.1 server --gh-user=GITHUB_USERNAME --gh-token=GITHUB_TOKEN
```

## Improvements
* Support for HTTPS cloning using GitHub username and token provided to atlantis server ([#117](https://github.com/hootsuite/atlantis/pull/117))
* Adding `post_plan` and `post_apply` commands ([#102](https://github.com/hootsuite/atlantis/pull/102))
* Adding the ability to verify webhook secret ([#120](https://github.com/hootsuite/atlantis/pull/120))

## Downloads

* [atlantis_darwin_amd64.zip](https://github.com/hootsuite/atlantis/releases/download/v0.1.1/atlantis_darwin_amd64.zip)
* [atlantis_linux_386.zip](https://github.com/hootsuite/atlantis/releases/download/v0.1.1/atlantis_linux_386.zip)
* [atlantis_linux_amd64.zip](https://github.com/hootsuite/atlantis/releases/download/v0.1.1/atlantis_linux_amd64.zip)
* [atlantis_linux_arm.zip](https://github.com/hootsuite/atlantis/releases/download/v0.1.1/atlantis_linux_arm.zip)
