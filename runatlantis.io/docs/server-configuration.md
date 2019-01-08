# Server Configuration
This documentation explains how to configure the Atlantis server and how to deal
with credentials.

[[toc]]

Configuration for `atlantis server` can be specified via command line flags, environment variables or a YAML config file.
Config file values are overridden by environment variables which in turn are overridden by flags.

## YAML
To use a yaml config file, run atlantis with `--config /path/to/config.yaml`.
The keys of your config file should be the same as the flag, ex.
```yaml
---
gh-token: ...
log-level: ...
```

## Environment Variables
All flags can be specified as environment variables. You need to convert the flag's `-`'s to `_`'s, uppercase all the letters and prefix with `ATLANTIS_`.
For example, `--gh-user` can be set via the environment variable `ATLANTIS_GH_USER`.

To see a list of all flags and their descriptions run `atlantis server --help`

::: warning
The flag `--atlantis-url` is set by the environment variable `ATLANTIS_ATLANTIS_URL` **NOT** `ATLANTIS_URL`.
:::

## Repo Whitelist
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
