# v0.3.9

## Features
* None

## Bugfixes
* Fix GitLab approvals not actually checking approval ([#114](https://github.com/runatlantis/atlantis/issues/114))

## Backwards Incompatibilities / Notes:
* None

## Downloads
* [atlantis_darwin_amd64.zip](https://github.com/runatlantis/atlantis/releases/download/v0.3.9/atlantis_darwin_amd64.zip)
* [atlantis_linux_386.zip](https://github.com/runatlantis/atlantis/releases/download/v0.3.9/atlantis_linux_386.zip)
* [atlantis_linux_amd64.zip](https://github.com/runatlantis/atlantis/releases/download/v0.3.9/atlantis_linux_amd64.zip)
* [atlantis_linux_arm.zip](https://github.com/runatlantis/atlantis/releases/download/v0.3.9/atlantis_linux_arm.zip)

# v0.3.8

## Features
* Terraform 0.11.7 in Docker image
* Docker build now verifies terraform install via checksum

## Bugfixes
* None

## Backwards Incompatibilities / Notes:
* None

## Downloads
* [atlantis_darwin_amd64.zip](https://github.com/runatlantis/atlantis/releases/download/v0.3.8/atlantis_darwin_amd64.zip)
* [atlantis_linux_386.zip](https://github.com/runatlantis/atlantis/releases/download/v0.3.8/atlantis_linux_386.zip)
* [atlantis_linux_amd64.zip](https://github.com/runatlantis/atlantis/releases/download/v0.3.8/atlantis_linux_amd64.zip)
* [atlantis_linux_arm.zip](https://github.com/runatlantis/atlantis/releases/download/v0.3.8/atlantis_linux_arm.zip)

# v0.3.7

## Bugfixes
* `--repo-whitelist` is now case insensitive. Fixes ([#95](https://github.com/runatlantis/atlantis/issues/95)).

## Backwards Incompatibilities / Notes:
* None

## Downloads
* [atlantis_darwin_amd64.zip](https://github.com/runatlantis/atlantis/releases/download/v0.3.7/atlantis_darwin_amd64.zip)
* [atlantis_linux_386.zip](https://github.com/runatlantis/atlantis/releases/download/v0.3.7/atlantis_linux_386.zip)
* [atlantis_linux_amd64.zip](https://github.com/runatlantis/atlantis/releases/download/v0.3.7/atlantis_linux_amd64.zip)
* [atlantis_linux_arm.zip](https://github.com/runatlantis/atlantis/releases/download/v0.3.7/atlantis_linux_arm.zip)

# v0.3.6

## Features
* `atlantis server -h` has newlines between flags so it's easier to read ([#91](https://github.com/runatlantis/atlantis/issues/91)).

## Bugfixes
* `atlantis bootstrap` uses a custom ngrok config file so it should work even
if the user is already running another ngrok tunnel ([#93](https://github.com/runatlantis/atlantis/issues/93)).

## Backwards Incompatibilities / Notes:
* None

## Downloads
* [atlantis_darwin_amd64.zip](https://github.com/runatlantis/atlantis/releases/download/v0.3.6/atlantis_darwin_amd64.zip)
* [atlantis_linux_386.zip](https://github.com/runatlantis/atlantis/releases/download/v0.3.6/atlantis_linux_386.zip)
* [atlantis_linux_amd64.zip](https://github.com/runatlantis/atlantis/releases/download/v0.3.6/atlantis_linux_amd64.zip)
* [atlantis_linux_arm.zip](https://github.com/runatlantis/atlantis/releases/download/v0.3.6/atlantis_linux_arm.zip)

# v0.3.5

## Features
* Log a warning if unable to update commit status. ([#84](https://github.com/runatlantis/atlantis/issues/84))

## Backwards Incompatibilities / Notes:
* None

## Downloads
* [atlantis_darwin_amd64.zip](https://github.com/runatlantis/atlantis/releases/download/v0.3.5/atlantis_darwin_amd64.zip)
* [atlantis_linux_386.zip](https://github.com/runatlantis/atlantis/releases/download/v0.3.5/atlantis_linux_386.zip)
* [atlantis_linux_amd64.zip](https://github.com/runatlantis/atlantis/releases/download/v0.3.5/atlantis_linux_amd64.zip)
* [atlantis_linux_arm.zip](https://github.com/runatlantis/atlantis/releases/download/v0.3.5/atlantis_linux_arm.zip)

# v0.3.4
## Description
This release delivers some speed improvements through caching plugins and
not running `terraform workspace select` unnecessarily. In my testing it saves ~20s per run.

## Features
* All config flags can now be specified by environment variables. Fixes ([#38](https://github.com/runatlantis/atlantis/issues/38)).
  * Completed thanks to @psalaberria002!
* Run terraform with the `TF_PLUGIN_CACHE_DIR` env var set. Fixes ([#34](https://github.com/runatlantis/atlantis/issues/34)).
  * This will cache plugins and make `terraform init` faster. Terraform will still download new versions of plugins. See https://www.terraform.io/docs/configuration/providers.html#provider-plugin-cache for more details.
  * In my testing this saves >10s per run.
* Run terraform with `TF_IN_AUTOMATION=true` so the output won't contain suggestions to run commands that you can't run via Atlantis. ([#82](https://github.com/runatlantis/atlantis/pull/82)).
* Don't run `terraform workspace select` unless we actually need to switch workspaces. ([#82](https://github.com/runatlantis/atlantis/pull/82)).
  * In my testing this saves ~10s.

## Bug Fixes
* Validate that workspace doesn't contain a path when running ex. `atlantis plan -w /jdlkj`. This was already not a valid workspace name according to Terraform. ([#78](https://github.com/runatlantis/atlantis/pull/78)).
* Error out if `ngrok` is already running when running `atlantis bootstrap` ([#81](https://github.com/runatlantis/atlantis/pull/81)).

## Backwards Incompatibilities / Notes:
* None

## Downloads
* [atlantis_darwin_amd64.zip](https://github.com/runatlantis/atlantis/releases/download/v0.3.4/atlantis_darwin_amd64.zip)
* [atlantis_linux_386.zip](https://github.com/runatlantis/atlantis/releases/download/v0.3.4/atlantis_linux_386.zip)
* [atlantis_linux_amd64.zip](https://github.com/runatlantis/atlantis/releases/download/v0.3.4/atlantis_linux_amd64.zip)
* [atlantis_linux_arm.zip](https://github.com/runatlantis/atlantis/releases/download/v0.3.4/atlantis_linux_arm.zip)

# v0.3.3

## Features
* Atlantis version shown in footer of web UI. Fixes ([#33](https://github.com/runatlantis/atlantis/issues/33)).

## Bug Fixes
* GitHub comments greater than the max length will be split into multiple comments. Fixes ([#55](https://github.com/runatlantis/atlantis/issues/55)).

## Backwards Incompatibilities / Notes:
* None

## Downloads
* [atlantis_darwin_amd64.zip](https://github.com/runatlantis/atlantis/releases/download/v0.3.3/atlantis_darwin_amd64.zip)
* [atlantis_linux_386.zip](https://github.com/runatlantis/atlantis/releases/download/v0.3.3/atlantis_linux_386.zip)
* [atlantis_linux_amd64.zip](https://github.com/runatlantis/atlantis/releases/download/v0.3.3/atlantis_linux_amd64.zip)
* [atlantis_linux_arm.zip](https://github.com/runatlantis/atlantis/releases/download/v0.3.3/atlantis_linux_arm.zip)

# v0.3.2

## Description
This release focused on some security issues reported by @eriksw, thanks Erik!
By default, Atlantis will be more secure now and you'll have to specify which repositories
you want it to work on.

## Features
* New flag `--allow-fork-prs` added to `atlantis server` controls whether Atlantis will operate on pull requests from forks. Defaults to `false`.
This flag was added because on a public repository anyone could open up a pull request to your repo and use your Atlantis
install.
* New mandatory flag `--repo-whitelist` added to `atlantis server` controls which repos Atlantis will operate on. This flag was added
so that if a webhook secret is compromised (or you're not using webhook secrets) Atlantis won't be used on repos you don't control.
* Warn if running `atlantis server` without any webhook secrets set. This is dangerous because without a webhook secret, an attacker
could spoof requests to Atlantis.
* Make CLI output more readable by setting a fixed column width.

## Bug Fixes
* None

## Backwards Incompatibilities / Notes:
* Must set `--allow-fork-prs` now if you want to run Atlantis on pull requests from forked repos.
* Must set `--repo-whitelist` in order to start `atlantis server`. See `atlantis server --help` for how that flag works.

## Downloads
* [atlantis_darwin_amd64.zip](https://github.com/runatlantis/atlantis/releases/download/v0.3.2/atlantis_darwin_amd64.zip)
* [atlantis_linux_386.zip](https://github.com/runatlantis/atlantis/releases/download/v0.3.2/atlantis_linux_386.zip)
* [atlantis_linux_amd64.zip](https://github.com/runatlantis/atlantis/releases/download/v0.3.2/atlantis_linux_amd64.zip)
* [atlantis_linux_arm.zip](https://github.com/runatlantis/atlantis/releases/download/v0.3.2/atlantis_linux_arm.zip)

# v0.3.1
## Features
* None

## Bug Fixes
* Run apply in correct directory when using `-d` flag. Fixes ([#22](https://github.com/runatlantis/atlantis/issues/22))

## Backwards Incompatibilities / Notes:
* None

## Downloads
* [atlantis_darwin_amd64.zip](https://github.com/runatlantis/atlantis/releases/download/v0.3.1/atlantis_darwin_amd64.zip)
* [atlantis_linux_386.zip](https://github.com/runatlantis/atlantis/releases/download/v0.3.1/atlantis_linux_386.zip)
* [atlantis_linux_amd64.zip](https://github.com/runatlantis/atlantis/releases/download/v0.3.1/atlantis_linux_amd64.zip)
* [atlantis_linux_arm.zip](https://github.com/runatlantis/atlantis/releases/download/v0.3.1/atlantis_linux_arm.zip)

# v0.3.0
## Features
* Fix security issue where Atlantis wasn't escaping the optional "extra args" that could be appended to comments ([#16](https://github.com/runatlantis/atlantis/pull/16))
  * example exploit: `atlantis plan ; cat /etc/passwd`
* Atlantis moved to new repo: `atlantisrun/atlantis`. Read why [here](https://medium.com/runatlantis/moving-atlantis-to-runatlantis-atlantis-on-github-4efc025bb05f)
* New -w/--workspace and -d/--dir flags in comments ([#14](https://github.com/runatlantis/atlantis/pull/14))
  * You can now specify which directory to plan/apply in, ex. `atlantis plan -d dir1/dir2`
* Better feedback from atlantis when asking for help via comments, ex. `atlantis plan -h`

## Bug Fixes
* Convert `--data-dir` paths to absolute from relative. Fixes ([#245](https://github.com/hootsuite/atlantis/issues/245))
* Don't run plan in the parent of `modules/` unless there's a `main.tf` present. Fixes ([#12](https://github.com/runatlantis/atlantis/issues/12))

## Backwards Incompatibilities / Notes:
* You must use the `-w` flag to specify a workspace when commenting now
  * Previously: `atlantis plan staging`, now: `atlantis plan -w staging`
* You must use a double-dash between Atlantis flags and extra args to be appended to the terraform command
  * Previously: `atlantis plan -target=resource`, now: `atlantis plan -- -target=resource`
* Atlantis will no longer run `plan` in the parent directory of `modules/` unless there is a `main.tf` in that directory.

## Downloads
* [atlantis_darwin_amd64.zip](https://github.com/runatlantis/atlantis/releases/download/v0.3.0/atlantis_darwin_amd64.zip)
* [atlantis_linux_386.zip](https://github.com/runatlantis/atlantis/releases/download/v0.3.0/atlantis_linux_386.zip)
* [atlantis_linux_amd64.zip](https://github.com/runatlantis/atlantis/releases/download/v0.3.0/atlantis_linux_amd64.zip)
* [atlantis_linux_arm.zip](https://github.com/runatlantis/atlantis/releases/download/v0.3.0/atlantis_linux_arm.zip)

# v0.2.4
## Features
* SSL support added ([#233](https://github.com/hootsuite/atlantis/pull/233))

## Bug Fixes
* GitLab custom URL for GitLab Enterprise installations now works ([#231](https://github.com/hootsuite/atlantis/pull/231))

## Backwards Incompatibilities / Notes:
None

## Downloads
* [atlantis_darwin_amd64.zip](https://github.com/runatlantis/atlantis/releases/download/v0.2.4/atlantis_darwin_amd64.zip)
* [atlantis_linux_386.zip](https://github.com/runatlantis/atlantis/releases/download/v0.2.4/atlantis_linux_386.zip)
* [atlantis_linux_amd64.zip](https://github.com/runatlantis/atlantis/releases/download/v0.2.4/atlantis_linux_amd64.zip)
* [atlantis_linux_arm.zip](https://github.com/runatlantis/atlantis/releases/download/v0.2.4/atlantis_linux_arm.zip)

# v0.2.3
## Features
None

## Bug Fixes
* Use `env` instead of `workspace` for Terraform 0.9.*

## Backwards Incompatibilities / Notes:
None

## Downloads
* [atlantis_darwin_amd64.zip](https://github.com/runatlantis/atlantis/releases/download/v0.2.3/atlantis_darwin_amd64.zip)
* [atlantis_linux_386.zip](https://github.com/runatlantis/atlantis/releases/download/v0.2.3/atlantis_linux_386.zip)
* [atlantis_linux_amd64.zip](https://github.com/runatlantis/atlantis/releases/download/v0.2.3/atlantis_linux_amd64.zip)
* [atlantis_linux_arm.zip](https://github.com/runatlantis/atlantis/releases/download/v0.2.3/atlantis_linux_arm.zip)

# v0.2.2
## Features
* Terraform 0.11 is now supported ([#219](https://github.com/hootsuite/atlantis/pull/219))
* Safe shutdown on `SIGTERM`/`SIGINT` ([#215](https://github.com/hootsuite/atlantis/pull/215))

## Bug Fixes
None

## Backwards Incompatibilities / Notes:
* The environment variables available when executing commands have changed:
  * `WORKSPACE` => `DIR` - this is the absolute path to the project directory on disk
  * `ENVIRONMENT` => `WORKSPACE` - this is the name of the Terraform workspace that we're running in (ex. default)
* The schema for storing locks changed. Any old locks will still be held but you will be unable to discard them in the UI.
**To fix this, either merge all the open pull requests before upgrading OR delete the `~/.atlantis/atlantis.db` file.**
This is safe to do because you'll just need to re-run `plan` to get your plan back.

## Downloads
* [atlantis_darwin_amd64.zip](https://github.com/runatlantis/atlantis/releases/download/v0.2.2/atlantis_darwin_amd64.zip)
* [atlantis_linux_386.zip](https://github.com/runatlantis/atlantis/releases/download/v0.2.2/atlantis_linux_386.zip)
* [atlantis_linux_amd64.zip](https://github.com/runatlantis/atlantis/releases/download/v0.2.2/atlantis_linux_amd64.zip)
* [atlantis_linux_arm.zip](https://github.com/runatlantis/atlantis/releases/download/v0.2.2/atlantis_linux_arm.zip)

# v0.2.1
## Features
* Don't ignore changes in `modules` directories anymore. ([#211](https://github.com/hootsuite/atlantis/pull/211))

## Bug Fixes
* Don't set `as_user` to true for Slack webhooks so we can integrate as a workspace app. ([#206](https://github.com/hootsuite/atlantis/pull/206))

## Backwards Incompatibilities / Notes:
None

## Downloads
* [atlantis_darwin_amd64.zip](https://github.com/runatlantis/atlantis/releases/download/v0.2.1/atlantis_darwin_amd64.zip)
* [atlantis_linux_386.zip](https://github.com/runatlantis/atlantis/releases/download/v0.2.1/atlantis_linux_386.zip)
* [atlantis_linux_amd64.zip](https://github.com/runatlantis/atlantis/releases/download/v0.2.1/atlantis_linux_amd64.zip)
* [atlantis_linux_arm.zip](https://github.com/runatlantis/atlantis/releases/download/v0.2.1/atlantis_linux_arm.zip)

# v0.2.0
## Features
* GitLab is now supported! ([#190](https://github.com/hootsuite/atlantis/pull/190))
* Slack notifications. ([#199](https://github.com/hootsuite/atlantis/pull/199))

## Bug Fixes
None

## Backwards Incompatibilities / Notes:
None

## Downloads
* [atlantis_darwin_amd64.zip](https://github.com/runatlantis/atlantis/releases/download/v0.2.0/atlantis_darwin_amd64.zip)
* [atlantis_linux_386.zip](https://github.com/runatlantis/atlantis/releases/download/v0.2.0/atlantis_linux_386.zip)
* [atlantis_linux_amd64.zip](https://github.com/runatlantis/atlantis/releases/download/v0.2.0/atlantis_linux_amd64.zip)
* [atlantis_linux_arm.zip](https://github.com/runatlantis/atlantis/releases/download/v0.2.0/atlantis_linux_arm.zip)

# v0.1.3
## Features
* Environment variables are passed through to `extra_arguments`. ([#150](https://github.com/hootsuite/atlantis/pull/150))
* Tested hundreds of lines of code. Test coverage now at 60%. ([https://codecov.io/gh/hootsuite/atlantis](https://codecov.io/gh/hootsuite/atlantis))

## Bug Fixes
* Modules in list of changed files weren't being filtered. ([#193](https://github.com/hootsuite/atlantis/pull/193))
* Nil pointer error in bootstrap mode. ([#181](https://github.com/hootsuite/atlantis/pull/181))

## Backwards Incompatibilities / Notes:
None

## Downloads
* [atlantis_darwin_amd64.zip](https://github.com/runatlantis/atlantis/releases/download/v0.1.3/atlantis_darwin_amd64.zip)
* [atlantis_linux_386.zip](https://github.com/runatlantis/atlantis/releases/download/v0.1.3/atlantis_linux_386.zip)
* [atlantis_linux_amd64.zip](https://github.com/runatlantis/atlantis/releases/download/v0.1.3/atlantis_linux_amd64.zip)
* [atlantis_linux_arm.zip](https://github.com/runatlantis/atlantis/releases/download/v0.1.3/atlantis_linux_arm.zip)

# v0.1.2
## Features
* all flags passed to `atlantis plan` or `atlantis apply` will now be passed through to `terraform`. ([#131](https://github.com/hootsuite/atlantis/pull/131))

## Bug Fixes
* Fix command parsing when comment ends with newline. ([#131](https://github.com/hootsuite/atlantis/pull/131))
* Plan and Apply outputs are shown in new line. ([#132](https://github.com/hootsuite/atlantis/pull/132))

## Downloads
* [atlantis_darwin_amd64.zip](https://github.com/runatlantis/atlantis/releases/download/v0.1.2/atlantis_darwin_amd64.zip)
* [atlantis_linux_386.zip](https://github.com/runatlantis/atlantis/releases/download/v0.1.2/atlantis_linux_386.zip)
* [atlantis_linux_amd64.zip](https://github.com/runatlantis/atlantis/releases/download/v0.1.2/atlantis_linux_amd64.zip)
* [atlantis_linux_arm.zip](https://github.com/runatlantis/atlantis/releases/download/v0.1.2/atlantis_linux_arm.zip)

# v0.1.1
## Backwards Incompatibilities / Notes:
* `--aws-assume-role-arn` and `--aws-region` flags removed. Instead, to name the
assume role session with the GitHub username of the user running the Atlantis command
use the `atlantis_user` terraform variable alongside Terraform's
[built-in support](https://www.terraform.io/docs/providers/aws/#assume-role) for assume role
(see https://github.com/runatlantis/atlantis/blob/master/README.md#assume-role-session-names)
* Atlantis has a docker image now ([#123](https://github.com/hootsuite/atlantis/pull/123)). Here is how you can try it out:

```bash
docker run runatlantis/atlantis:v0.1.1 server --gh-user=GITHUB_USERNAME --gh-token=GITHUB_TOKEN
```

## Improvements
* Support for HTTPS cloning using GitHub username and token provided to atlantis server ([#117](https://github.com/hootsuite/atlantis/pull/117))
* Adding `post_plan` and `post_apply` commands ([#102](https://github.com/hootsuite/atlantis/pull/102))
* Adding the ability to verify webhook secret ([#120](https://github.com/hootsuite/atlantis/pull/120))

## Downloads

* [atlantis_darwin_amd64.zip](https://github.com/runatlantis/atlantis/releases/download/v0.1.1/atlantis_darwin_amd64.zip)
* [atlantis_linux_386.zip](https://github.com/runatlantis/atlantis/releases/download/v0.1.1/atlantis_linux_386.zip)
* [atlantis_linux_amd64.zip](https://github.com/runatlantis/atlantis/releases/download/v0.1.1/atlantis_linux_amd64.zip)
* [atlantis_linux_arm.zip](https://github.com/runatlantis/atlantis/releases/download/v0.1.1/atlantis_linux_arm.zip)
