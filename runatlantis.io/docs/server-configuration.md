# Server Configuration
This page explains how to configure the `atlantis server` command.

Configuration to `atlantis server` can be specified via command line flags,
 environment variables, a config file or a mix of the three.

[[toc]]

## Flags
To see which flags are available, run `atlantis server --help`.

## Environment Variables
All flags can be specified as environment variables.

1. Take the flag name, ex. `--gh-user`
1. Ignore the first `--` => `gh-user`
1. Convert the `-`'s to `_`'s => `gh_user`
1. Uppercase all the letters => `GH_USER`
1. Prefix with `ATLANTIS_` => `ATLANTIS_GH_USER`

::: warning NOTE
The flag `--atlantis-url` is set by the environment variable `ATLANTIS_ATLANTIS_URL` **NOT** `ATLANTIS_URL`.
:::

## Config File
All flags can also be specified via a YAML config file.

To use a YAML config file, run `atlantis serer --config /path/to/config.yaml`.

The keys of your config file should be the same as the flag names, ex.
```yaml
gh-token: ...
log-level: ...
```

::: warning
The config file you pass to `--config` is different from the `--repo-config` file.
The `--config` config file is only used as an alternate way of setting `atlantis server` flags.
:::

## Precedence
Values are chosen in this order:
1. Flags
1. Environment Variables
1. Config File


## Notes On Specific Flags
### `--automerge`
See [Automerging](automerging.html)

### `--checkout-strategy`
See [Checkout Strategy](checkout-strategy.html)

### `--default-tf-version`
See [Terraform Versions](terraform-versions.html)

### `--repo-whitelist`
Atlantis requires you to specify a whitelist of repositories it will accept webhooks from via the `--repo-whitelist` flag.

Notes:
* Accepts a comma separated list, ex. `definition1,definition2`
* Format is `{hostname}/{owner}/{repo}`, ex. `github.com/runatlantis/atlantis`
* `*` matches any characters, ex. `github.com/runatlantis/*` will match all repos in the runatlantis organization
* For Bitbucket Server: `{hostname}` is the domain without scheme and port, `{owner}` is the name of the project (not the key), and `{repo}` is the repo name

Examples:
* Whitelist `myorg/repo1` and `myorg/repo2` on `github.com`
  * `--repo-whitelist=github.com/myorg/repo1,github.com/myorg/repo2`
* Whitelist all repos under `myorg` on `github.com`
  * `--repo-whitelist='github.com/myorg/*'`
* Whitelist all repos in my GitHub Enterprise installation
  * `--repo-whitelist='github.yourcompany.com/*'`
* Whitelist all repositories
  * `--repo-whitelist='*'`
  
### `--silence-whitelist-errors`
Some users use the `--repo-whitelist` flag to control which repos Atlantis
responds to. Normally, if Atlantis receives a pull request webhook from a repo not listed
in the whitelist, it will comment back with an error. This flag disables that commenting.

Some users find this useful because they prefer to add the Atlantis webhook
at an organization level rather than on each repo.

### `--tfe-token`
A token for Terraform Enterprise integration. See [Terraform Enterprise](terraform-enterprise.html) for more details.
