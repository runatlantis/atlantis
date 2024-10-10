# Server Configuration

This page explains how to configure the `atlantis server` command.

Configuration to `atlantis server` can be specified via command line flags,
 environment variables, a config file or a mix of the three.

## Environment Variables

All flags can be specified as environment variables.

1. Take the flag name, ex. `--gh-user`
1. Ignore the first `--` => `gh-user`
1. Convert the `-`'s to `_`'s => `gh_user`
1. Uppercase all the letters => `GH_USER`
1. Prefix with `ATLANTIS_` => `ATLANTIS_GH_USER`

::: warning NOTE
To set a boolean flag use `true` or `false` as the value.
:::

::: warning NOTE
The flag `--atlantis-url` is set by the environment variable `ATLANTIS_ATLANTIS_URL` **NOT** `ATLANTIS_URL`.
:::

## Config File

All flags can also be specified via a YAML config file.

To use a YAML config file, run `atlantis server --config /path/to/config.yaml`.

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

## Flags

### `--allow-commands`

  ```bash
  atlantis server --allow-commands=version,plan,apply,unlock,approve_policies
  # or
  ATLANTIS_ALLOW_COMMANDS='version,plan,apply,unlock,approve_policies'
  ```

  List of allowed commands to be run on the Atlantis server, Defaults to `version,plan,apply,unlock,approve_policies`

  Notes:

* Accepts a comma separated list, ex. `command1,command2`.
* `version`, `plan`, `apply`, `unlock`, `approve_policies`, `import`, `state` and `all` are available.
* `all` is a special keyword that allows all commands. If pass `all` then all other commands will be ignored.

### `--allow-draft-prs`

  ```bash
  atlantis server --allow-draft-prs
  # or
  ATLANTIS_ALLOW_DRAFT_PRS=true
  ```

  Respond to pull requests from draft prs. Defaults to `false`.

### `--allow-fork-prs`

  ```bash
  atlantis server --allow-fork-prs
  # or
  ATLANTIS_ALLOW_FORK_PRS=true
  ```

  Respond to pull requests from forks. Defaults to `false`.

  :::warning SECURITY WARNING
  Potentially dangerous to enable
  because if attackers can create a pull request to your repo then they can cause Atlantis
  to run arbitrary code. This can happen because
  Atlantis will automatically run `terraform plan`
  which can run arbitrary code if given a malicious Terraform configuration.
  :::

### `--api-secret`

  ```bash
  atlantis server --api-secret="secret"
  # or (recommended)
  ATLANTIS_API_SECRET="secret"
  ```

  Required secret used to validate requests made to the [`/api/*` endpoints](api-endpoints.md).

### `--atlantis-url`

  ```bash
  atlantis server --atlantis-url="https://my-domain.com:9090/basepath"
  # or
  ATLANTIS_ATLANTIS_URL=https://my-domain.com:9090/basepath
  ```

  Specify the URL that Atlantis is accessible from. Used in the Atlantis UI
  and in links from pull request comments. Defaults to `http://$(hostname):$port`
  where `$port` is from the [`--port`](#port) flag. Supports a basepath if you're hosting Atlantis under a path.

  Notes:

* If a load balancer with a non http/https port (not the one defined in the `--port` flag) is used, update the URL to include the port like in the example above.
* This URL is used as the `details` link next to each atlantis job to view the job's logs.

### `--autodiscover-mode`

  ```bash
  atlantis server --autodiscover-mode="<auto|enabled|disabled>"
  # or
  ATLANTIS_AUTODISCOVER_MODE="<auto|enabled|disabled>"
  ```

  Sets auto discover mode, default is `auto`. When set to `auto`, projects in a repo will be discovered by
  Atlantis when there are no projects configured in the repo config. If one or more projects are defined
  in the repo config then auto discovery will be completely disabled.

  When set to `enabled` projects will be discovered unconditionally. If an auto discovered project is already
  defined in the projects section of the repo config, the project from the repo config will take precedence over
  the auto discovered project.

  When set to `disabled` projects will never be discovered, even if there are no projects configured in the repo config.

### `--automerge`

  ```bash
  atlantis server --automerge
  # or
  ATLANTIS_AUTOMERGE=true
  ```

  Automatically merge pull requests after all plans have been successfully applied.
  Defaults to `false`. See [Automerging](automerging.md) for more details.

### `--autoplan-file-list`

  ```bash
  # NOTE: Use single quotes to avoid shell expansion of *.
  atlantis server --autoplan-file-list='**/*.tf,project1/*.pkr.hcl'
  # or
  ATLANTIS_AUTOPLAN_FILE_LIST='**/*.tf,project1/*.pkr.hcl'
  ```

  List of file patterns that Atlantis will use to check if a directory contains modified files that should trigger project planning.

  Notes:

* Accepts a comma separated list, ex. `pattern1,pattern2`.
* Patterns use the [`.dockerignore` syntax](https://docs.docker.com/engine/reference/builder/#dockerignore-file)
* List of file patterns will be used by both automatic and manually run plans.
* When not set, defaults to all `.tf`, `.tfvars`, `.tfvars.json`,  `terragrunt.hcl` and `.terraform.lock.hcl` files
    (`--autoplan-file-list='**/*.tf,**/*.tfvars,**/*.tfvars.json,**/terragrunt.hcl,**/.terraform.lock.hcl'`).
* Setting `--autoplan-file-list` will override the defaults. You **must** add `**/*.tf` and other defaults if you want to include them.
* A custom [Workflow](repo-level-atlantis-yaml.md#configuring-planning) that uses autoplan `when_modified` will ignore this value.

  Examples:

* Autoplan when any `*.tf` or `*.tfvars` file is modified.
  * `--autoplan-file-list='**/*.tf,**/*.tfvars'`
* Autoplan when any `*.tf` file is modified except in `project2/` directory
  * `--autoplan-file-list='**/*.tf,!project2'`
* Autoplan when any `*.tf` files or `.yml` files in subfolder of `project1` is modified.
  * `--autoplan-file-list='**/*.tf,project2/**/*.yml'`

::: warning NOTE
By default, changes to modules will not trigger autoplanning. See the flags below.
:::

::: warning NOTE
If any projects are defined in a repo atlantis.yaml file, the logic for this flag will not execute. See issue [#3122](https://github.com/runatlantis/atlantis/issues/3122).
:::

### `--autoplan-modules`

```bash
atlantis server --autoplan-modules
# or
ATLANTIS_AUTOPLAN_MODULES=true
```

Defaults to `false`. When set to `true`, Atlantis will trace the local modules of included projects.
Included project are projects with files included by `--autoplan-file-list`.
After tracing, Atlantis will plan any project that includes a changed module. This is equivalent to setting
`--autoplan-modules-from-projects` to the value of `--autoplan-file-list`. See below.

::: warning NOTE
If any projects are defined in a repo atlantis.yaml file, the logic for this flag will not execute. See issue [#3122](https://github.com/runatlantis/atlantis/issues/3122).
:::

### `--autoplan-modules-from-projects`

```bash
atlantis server --autoplan-modules-from-projects='**/init.tf'
# or
ATLANTIS_AUTOPLAN_MODULES_FROM_PROJECTS='**/init.tf'
```

Enables auto-planing of projects when a module dependency in the same repository has changed.
This is a list of file patterns like `autoplan-file-list`.

These patterns select **projects** to index based on the files matched. The index maps modules to the projects that depends on them,
including projects that include the module via other modules. When a module file matching `autoplan-file-list` changes,
all indexed projects will be planned.

Current default is "" (disabled).

Examples:

* `**/*.tf` - will index all projects that have a `.tf` file in their directory, and plan them whenever an in-repo module dependency has changed.
* `**/*.tf,!foo,!bar` - will index all projects containing `.tf` except `foo` and `bar` and plan them whenever an in-repo module dependency has changed.
   This allows projects to opt-out of auto-planning when a module dependency changes.

::: warning NOTE
Modules that are not selected by autoplan-file-list will not be indexed and dependant projects will not be planned. This
flag allows the *projects* to index to be selected, but the trigger for a plan must be a file in `autoplan-file-list`.
:::

::: warning NOTE
This flag overrides `--autoplan-modules`. If you wish to disable auto-planning of modules, set this flag to an empty string,
and set `--autoplan-modules` to `false`.
:::

### `--azuredevops-hostname`

  ```bash
  atlantis server --azuredevops-hostname="dev.azure.com"
  # or
  ATLANTIS_AZUREDEVOPS_HOSTNAME="dev.azure.com"
  ```

  Azure DevOps hostname to support cloud and self hosted instances. Defaults to `dev.azure.com`.

### `--azuredevops-token`

  ```bash
  atlantis server --azuredevops-token="RandomStringProducedByAzureDevOps"
  # or (recommended)
  ATLANTIS_AZUREDEVOPS_TOKEN="RandomStringProducedByAzureDevOps"
  ```

  Azure DevOps token of API user.

### `--azuredevops-user`

  ```bash
  atlantis server --azuredevops-user="username@example.com"
  # or
  ATLANTIS_AZUREDEVOPS_USER="username@example.com"
  ```

  Azure DevOps username of API user.

### `--azuredevops-webhook-password`

  ```bash
  atlantis server --azuredevops-webhook-password="password123"
  # or (recommended)
  ATLANTIS_AZUREDEVOPS_WEBHOOK_PASSWORD="password123"
  ```

  Azure DevOps basic authentication password for inbound webhooks (see
  [docs](https://docs.microsoft.com/en-us/azure/devops/service-hooks/authorize?view=azure-devops)).

  ::: warning SECURITY WARNING
  If not specified, Atlantis won't be able to validate that the
  incoming webhook call came from your Azure DevOps org. This means that an
  attacker could spoof calls to Atlantis and cause it to perform malicious
  actions. Should be specified via the `ATLANTIS_AZUREDEVOPS_WEBHOOK_PASSWORD` environment
  variable.
  :::

### `--azuredevops-webhook-user`

  ```bash
  atlantis server --azuredevops-webhook-user="username@example.com"
  # or
  ATLANTIS_AZUREDEVOPS_WEBHOOK_USER="username@example.com"
  ```

  Azure DevOps basic authentication username for inbound webhooks.

### `--bitbucket-base-url`

  ```bash
  atlantis server --bitbucket-base-url="http://bitbucket.corp:7990/basepath"
  # or
  ATLANTIS_BITBUCKET_BASE_URL="http://bitbucket.corp:7990/basepath"
  ```

  Base URL of Bitbucket Server (aka Stash) installation. Must include
  `http://` or `https://`. If using Bitbucket Cloud (bitbucket.org), do not set. Defaults to
  `https://api.bitbucket.org`.

### `--bitbucket-token`

  ```bash
  atlantis server --bitbucket-token="token"
  # or (recommended)
  ATLANTIS_BITBUCKET_TOKEN="token"
  ```

  Bitbucket app password of API user.

### `--bitbucket-user`

  ```bash
  atlantis server --bitbucket-user="myuser"
  # or
  ATLANTIS_BITBUCKET_USER="myuser"
  ```

  Bitbucket username of API user.

### `--bitbucket-webhook-secret`

  ```bash
  atlantis server --bitbucket-webhook-secret="secret"
  # or (recommended)
  ATLANTIS_BITBUCKET_WEBHOOK_SECRET="secret"
  ```

  Secret used to validate Bitbucket webhooks. Only Bitbucket Server supports webhook secrets.
  For Bitbucket.org, see [Security](security.md#bitbucket-cloud-bitbucket-org) for mitigations.

  ::: warning SECURITY WARNING
  If not specified, Atlantis won't be able to validate that the incoming webhook call came from Bitbucket.
  This means that an attacker could spoof calls to Atlantis and cause it to perform malicious actions.
  :::

### `--checkout-depth`

  ```bash
  atlantis server --checkout-depth=0
  # or
  ATLANTIS_CHECKOUT_DEPTH=0
  ```

  The number of commits to fetch from the branch. Used if `--checkout-strategy=merge` since the `--checkout-strategy=branch` (default) checkout strategy always defaults to a shallow clone using a depth of 1.
  Defaults to `0`. See [Checkout Strategy](checkout-strategy.md) for more details.

### `--checkout-strategy`

  ```bash
  atlantis server --checkout-strategy="<branch|merge>"
  # or
  ATLANTIS_CHECKOUT_STRATEGY="<branch|merge>"
  ```

  How to check out pull requests. Use either `branch` or `merge`.
  Defaults to `branch`. See [Checkout Strategy](checkout-strategy.md) for more details.

### `--config`

  ```bash
  atlantis server --config="my/config/file.yaml"
  # or
  ATLANTIS_CONFIG="my/config/file.yaml"
  ```

  YAML config file where flags can also be set. See [Config File](#config-file) for more details.

### `--data-dir`

  ```bash
  atlantis server --data-dir="path/to/data/dir"
  # or
  ATLANTIS_DATA_DIR="path/to/data/dir"
  ```

  Directory where Atlantis will store its data. Will be created if it doesn't exist.
  Defaults to `~/.atlantis`. Atlantis will store its database, checked out repos, Terraform plans and downloaded
  Terraform binaries here. If Atlantis loses this directory, [locks](locking.md)
  will be lost and unapplied plans will be lost.

  Note that the atlantis user is restricted to `~/.atlantis`.
  If you set the `--data-dir` flag to a path outside of Atlantis its home directory, ensure that you grant the atlantis user the correct permissions.

### `--default-tf-version`

  ```bash
  atlantis server --default-tf-version="v0.12.31"
  # or
  ATLANTIS_DEFAULT_TF_VERSION="v0.12.31"
  ```

  Terraform version to default to. Will download to `<data-dir>/bin/terraform<version>`
  if not in `PATH`. See [Terraform Versions](terraform-versions.md) for more details.

### `--disable-apply-all`

  ```bash
  atlantis server --disable-apply-all
  # or
  ATLANTIS_DISABLE_APPLY_ALL=true
  ```

  Disable `atlantis apply` command so a specific project/workspace/directory has to
  be specified for applies.

### `--disable-autoplan`

  ```bash
  atlantis server --disable-autoplan
  # or
  ATLANTIS_DISABLE_AUTOPLAN=true
  ```

  Disable atlantis auto planning.

### `--disable-autoplan-label`

  ```bash
  atlantis server --disable-autoplan-label="no-autoplan"
  # or
  ATLANTIS_DISABLE_AUTOPLAN_LABEL="no-autoplan"
  ```

  Disable atlantis auto planning only on pull requests with the specified label.

  If `disable-autoplan` property is `true`, this flag has no effect.

### `--disable-markdown-folding`

  ```bash
  atlantis server --disable-markdown-folding
  # or
  ATLANTIS_DISABLE_MARKDOWN_FOLDING=true
  ```

  Disable folding in markdown output using the `<details>` html tag.

### `--disable-repo-locking`

  ```bash
  atlantis server --disable-repo-locking
  # or
  ATLANTIS_DISABLE_REPO_LOCKING=true
  ```

  Stops atlantis from locking projects and or workspaces when running terraform.

### `--disable-unlock-label`

  ```bash
  atlantis server --disable-unlock-label do-not-unlock
  # or
  ATLANTIS_DISABLE_UNLOCK_LABEL="do-not-unlock"
  ```

  Stops atlantis from unlocking a pull request with this label. Defaults to "" (feature disabled).

### `--emoji-reaction`

  ```bash
  atlantis server --emoji-reaction thumbsup
  # or
  ATLANTIS_EMOJI_REACTION=thumbsup
  ```

  The emoji reaction to use for marking processed comments. Currently supported on Azure DevOps, GitHub and GitLab. If not specified, Atlantis will not use an emoji reaction.
  Defaults to "" (empty string).

### `--enable-diff-markdown-format`

  ```bash
  atlantis server --enable-diff-markdown-format
  # or
  ATLANTIS_ENABLE_DIFF_MARKDOWN_FORMAT=true
  ```

  Enable Atlantis to format Terraform plan output into a markdown-diff friendly format for color-coding purposes.

  Useful to enable for use with GitHub.

### `--enable-policy-checks`

  ```bash
  atlantis server --enable-policy-checks
  # or
  ATLANTIS_ENABLE_POLICY_CHECKS=true
  ```

  Enables atlantis to run server side policies on the result of a terraform plan. Policies are defined in [server side repo config](server-side-repo-config.md#reference).

### `--enable-regexp-cmd`

  ```bash
  atlantis server --enable-regexp-cmd
  # or
  ATLANTIS_ENABLE_REGEXP_CMD=true
  ```

  Enable Atlantis to use regular expressions to run plan/apply commands against defined project names when `-p` flag is passed with it.

  This can be used to run all defined projects (with the `name` key) in `atlantis.yaml` using `atlantis plan -p .*`.

  The flag will only allow the regexes listed in the [`allowed_regexp_prefixes`](repo-level-atlantis-yaml.md#reference) key defined in the repo `atlantis.yaml` file. If the key is undefined, its value defaults to `[]` which will allow any regex.

  This will not work with `-d` yet and to use `-p` the repo projects must be defined in the repo `atlantis.yaml` file.

  This will bypass `--restrict-file-list` if regex is used, normal commands will stil be blocked if necessary.

  ::: warning SECURITY WARNING
  It's not supposed to be used with `--disable-apply-all`.
  The command `atlantis apply -p .*` will bypass the restriction and run apply on every projects.
  :::

### `--executable-name`

  ```bash
  atlantis server --executable-name="atlantis"
  # or
  ATLANTIS_EXECUTABLE_NAME="atlantis"
  ```

  Comment command trigger executable name. Defaults to `atlantis`.

  This is useful when running multiple Atlantis servers against a single repository.

### `--fail-on-pre-workflow-hook-error`

  ```bash
  atlantis server --fail-on-pre-workflow-hook-error
  # or
  ATLANTIS_FAIL_ON_PRE_WORKFLOW_HOOK_ERROR=true
  ```

  Fail and do not run the requested Atlantis command if any of the pre workflow hooks error.

### `--gitea-base-url`

  ```bash
  atlantis server --gitea-base-url="http://your-gitea.corp:7990/basepath"
  # or
  ATLANTIS_GITEA_BASE_URL="http://your-gitea.corp:7990/basepath"
  ```

  Base URL of Gitea installation. Must include `http://` or `https://`. Defaults to `https://gitea.com` if left empty/absent.

### `--gitea-token`

  ```bash
  atlantis server --gitea-token="token"
  # or (recommended)
  ATLANTIS_GITEA_TOKEN="token"
  ```

  Gitea app password of API user.

### `--gitea-user`

  ```bash
  atlantis server --gitea-user="myuser"
  # or
  ATLANTIS_GITEA_USER="myuser"
  ```

  Gitea username of API user.

### `--gitea-webhook-secret`

  ```bash
  atlantis server --gitea-webhook-secret="secret"
  # or (recommended)
  ATLANTIS_GITEA_WEBHOOK_SECRET="secret"
  ```

  Secret used to validate Gitea webhooks.

  ::: warning SECURITY WARNING
  If not specified, Atlantis won't be able to validate that the incoming webhook call came from Gitea.
  This means that an attacker could spoof calls to Atlantis and cause it to perform malicious actions.
  :::

### `--gitea-page-size`

  ```bash
  atlantis server --gitea-page-size=30
  # or (recommended)
  ATLANTIS_GITEA_PAGE_SIZE=30
  ```

  Number of items on a single page in Gitea paged responses.

  ::: warning Configuration dependent
  The default value conforms to the Gitea server's standard config setting: DEFAULT_PAGING_NUM
 The highest valid value depends on the Gitea server's config setting: MAX_RESPONSE_ITEMS
  :::

### `--gh-allow-mergeable-bypass-apply`

  ```bash
  atlantis server --gh-allow-mergeable-bypass-apply
  # or
  ATLANTIS_GH_ALLOW_MERGEABLE_BYPASS_APPLY=true
  ```

  Feature flag to enable ability to use `mergeable` mode with required apply status check.

### `--gh-app-id`

  ```bash
  atlantis server --gh-app-id="00000"
  # or
  ATLANTIS_GH_APP_ID="00000"
  ```

  GitHub app ID. If set, GitHub authentication will be performed as [an installation](https://docs.github.com/en/rest/apps/installations).

  ::: tip
  A GitHub app can be created by starting Atlantis first, then pointing your browser at

  ```shell
  $(hostname)/github-app/setup
  ```

  You'll be redirected to GitHub to create a new app, and will then be redirected to

  ```shell
  $(hostname)/github-app/exchange-code?code=some-code
  ```

  After which Atlantis will display your new app's credentials: your app's ID, its generated `--gh-webhook-secret` and the contents of the file for `--gh-app-key-file`. Update your Atlantis config accordingly, and restart the server.
  :::

### `--gh-app-key`

  ```bash
  atlantis server --gh-app-key="-----BEGIN RSA PRIVATE KEY-----(...)"
  # or
  ATLANTIS_GH_APP_KEY="-----BEGIN RSA PRIVATE KEY-----(...)"
  ```

  The PEM encoded private key for the GitHub App.

  ::: warning SECURITY WARNING
  The contents of the private key will be visible by anyone that can run `ps` or look at the shell history of the machine where Atlantis is running. Use `--gh-app-key-file` to mitigate that risk.
  :::

### `--gh-app-key-file`

  ```bash
  atlantis server --gh-app-key-file="path/to/app-key.pem"
  # or
  ATLANTIS_GH_APP_KEY_FILE="path/to/app-key.pem"
  ```

  Path to a GitHub App PEM encoded private key file. If set, GitHub authentication will be performed as [an installation](https://docs.github.com/en/rest/apps/installations).

### `--gh-app-slug`

  ```bash
  atlantis server --gh-app-slug="myappslug"
  # or
  ATLANTIS_GH_APP_SLUG="myappslug"
  ```

  A slugged version of GitHub app name shown in pull requests comments, etc (not `Atlantis App` but something like `atlantis-app`). Atlantis uses the value of this parameter to identify the comments it has left on GitHub pull requests. This is used for functions such as `--hide-prev-plan-comments`. You need to obtain this value from your GitHub app, one way is to go to your App settings and open "Public page" from the left sidebar. Your `--gh-app-slug` value will be the last part of the URL, e.g `https://github.com/apps/<slug>`.

### `--gh-hostname`

  ```bash
  atlantis server --gh-hostname="my.github.enterprise.com"
  # or
  ATLANTIS_GH_HOSTNAME="my.github.enterprise.com"
  ```

  Hostname of your GitHub Enterprise installation. If using [GitHub.com](https://github.com),
  don't set. Defaults to `github.com`.

### `--gh-app-installation-id`

  ```bash
  atlantis server --gh-app-installation-id="123"
  # or
  ATLANTIS_GH_APP_INSTALLATION_ID="123"
  ```

The installation ID of a specific instance of a GitHub application. Normally this value is
derived by querying GitHub for the list of installations of the ID supplied via `--gh-app-id` and selecting
the first one found and where multiple installations results in an error. Use this flag if you have multiple
instances of Atlantis but you want to use a single already-installed GitHub app for all of them. You would normally do this if
you are running a proxy as your single GitHub application that will proxy to an appropriate Atlantis instance
based on the organization or user that triggered the webhook.

### `--gh-org`

  ```bash
  atlantis server --gh-org="myorgname"
  # or
  ATLANTIS_GH_ORG="myorgname"
  ```

  GitHub organization name. Set to enable creating a private GitHub app for this organization.

### `--gh-team-allowlist`

  ```bash
  atlantis server --gh-team-allowlist="myteam:plan, secteam:apply, DevOps Team:apply, DevOps Team:import"
  # or
  ATLANTIS_GH_TEAM_ALLOWLIST="myteam:plan, secteam:apply, DevOps Team:apply, DevOps Team:import"
  ```

  In versions v0.21.0 and later, the GitHub team name can be a name or a slug.

  In versions v0.20.1 and below, the Github team name required the case sensitive team name.

  Comma-separated list of GitHub teams and permission pairs.

  By default, any team can plan and apply.

  ::: warning NOTE
  You should use the Team name as the variable, not the slug, even if it has spaces or special characters.
  i.e., "Engineering Team:plan, Infrastructure Team:apply"
  :::

### `--gh-token`

  ```bash
  atlantis server --gh-token="token"
  # or (recommended)
  ATLANTIS_GH_TOKEN="token"
  ```

  GitHub token of API user.

### `--gh-user`

  ```bash
  atlantis server --gh-user="myuser"
  # or
  ATLANTIS_GH_USER="myuser"
  ```

   GitHub username of API user.

### `--gh-webhook-secret`

  ```bash
  atlantis server --gh-webhook-secret="secret"
  # or (recommended)
  ATLANTIS_GH_WEBHOOK_SECRET="secret"
  ```

  Secret used to validate GitHub webhooks (see [GitHub: Validating webhook deliveries](https://docs.github.com/en/webhooks/using-webhooks/validating-webhook-deliveries)).

  ::: warning SECURITY WARNING
  If not specified, Atlantis won't be able to validate that the incoming webhook call came from GitHub.
  This means that an attacker could spoof calls to Atlantis and cause it to perform malicious actions.
  :::

### `--gitlab-hostname`

  ```bash
  atlantis server --gitlab-hostname="my.gitlab.enterprise.com"
  # or
  ATLANTIS_GITLAB_HOSTNAME="my.gitlab.enterprise.com"
  ```

  Hostname of your GitLab Enterprise installation. If using [Gitlab.com](https://gitlab.com),
  don't set. Defaults to `gitlab.com`.

### `--gitlab-token`

  ```bash
  atlantis server --gitlab-token="token"
  # or (recommended)
  ATLANTIS_GITLAB_TOKEN="token"
  ```

  GitLab token of API user.

### `--gitlab-user`

  ```bash
  atlantis server --gitlab-user="myuser"
  # or
  ATLANTIS_GITLAB_USER="myuser"
  ```

  GitLab username of API user.

### `--gitlab-webhook-secret`

  ```bash
  atlantis server --gitlab-webhook-secret="secret"
  # or (recommended)
  ATLANTIS_GITLAB_WEBHOOK_SECRET="secret"
  ```

  Secret used to validate GitLab webhooks.

  ::: warning SECURITY WARNING
  If not specified, Atlantis won't be able to validate that the incoming webhook call came from GitLab.
  This means that an attacker could spoof calls to Atlantis and cause it to perform malicious actions.
  :::

### `--help`

  ```bash
  atlantis server --help
  ```

  View help.

### `--hide-prev-plan-comments`

  ```bash
  atlantis server --hide-prev-plan-comments
  # or
  ATLANTIS_HIDE_PREV_PLAN_COMMENTS=true
  ```

  Hide previous plan comments to declutter PRs. This is only supported in
  GitHub and GitLab currently. This is not enabled by default. When using Github App, you need to set `--gh-app-slug` to enable this feature.

### `--hide-unchanged-plan-comments`

  ```bash
  atlantis server --hide-unchanged-plan-comments
  # or
  ATLANTIS_HIDE_UNCHANGED_PLAN_COMMENTS=true
  ```

Remove no-changes plan comments from the pull request.

This is useful when you have many projects and want to keep the pull request clean from useless comments.

### `--include-git-untracked-files`

  ```bash
  atlantis server --include-git-untracked-files
  # or
  ATLANTIS_INCLUDE_GIT_UNTRACKED_FILES=true
  ```

  Include git untracked files in the Atlantis modified file list.
  Used for example with CDKTF pre-workflow hooks that dynamically generate
  Terraform files.

### `--locking-db-type`

  ```bash
  atlantis server --locking-db-type="<boltdb|redis>"
  # or
  ATLANTIS_LOCKING_DB_TYPE="<boltdb|redis>"
  ```

  The locking database type to use for storing plan and apply locks. Defaults to `boltdb`.

  Notes:

* If set to `boltdb`, only one process may have access to the boltdb instance.
* If set to `redis`, then `--redis-host`, `--redis-port`, and `--redis-password` must be set.

### `--log-level`

  ```bash
  atlantis server --log-level="<debug|info|warn|error>"
  # or
  ATLANTIS_LOG_LEVEL="<debug|info|warn|error>"
  ```

  Log level. Defaults to `info`.

### `--markdown-template-overrides-dir`

  ```bash
  atlantis server --markdown-template-overrides-dir="path/to/templates/"
  # or
  ATLANTIS_MARKDOWN_TEMPLATE_OVERRIDES_DIR="path/to/templates/"
  ```

  This will be available in v0.21.0.

  Directory where Atlantis will read in overrides for markdown templates used to render comments on pull requests.
  Markdown template overrides may be specified either in individual files, or all together in a single file. All template
  override files *must* have the `.tmpl` extension, otherwise they will not be parsed.

  Markdown templates which may have overrides can be found [here](https://github.com/runatlantis/atlantis/tree/main/server/events/templates)

  Please be mindful that settings like `--enable-diff-markdown-format` depend on logic defined in the templates. It is
  possible to diverge from expected behavior, if care is not taken when overriding default templates.

  Defaults to the atlantis home directory `/home/atlantis/.markdown_templates/` in `/$HOME/.markdown_templates`.

### `--max-comments-per-command`

  ```bash
  atlantis server --max-comments-per-command=100
  # or
  ATLANTIS_MAX_COMMENTS_PER_COMMAND=100
  ```

  Limit the number of comments published after a command is executed, to prevent spamming your VCS and Atlantis to get throttled as a result. Defaults to `100`. Set this option to `0` to disable log truncation. Note that the truncation will happen on the top of the command output, to preserve the most important parts of the output, often displayed at the end.

### `--parallel-apply`

  ```bash
  atlantis server --parallel-apply
  # or
  ATLANTIS_PARALLEL_APPLY=true
  ```

  Whether to run apply operations in parallel. Defaults to `false`. Explicit declaration in [repo config](repo-level-atlantis-yaml.md#run-plans-and-applies-in-parallel) takes precedence.

### `--parallel-plan`

  ```bash
  atlantis server --parallel-plan
  # or
  ATLANTIS_PARALLEL_PLAN=true
  ```

  Whether to run plan operations in parallel. Defaults to `false`. Explicit declaration in [repo config](repo-level-atlantis-yaml.md#run-plans-and-applies-in-parallel) takes precedence.

### `--parallel-pool-size`

  ```bash
  atlantis server --parallel-pool-size=100
  # or
  ATLANTIS_PARALLEL_POOL_SIZE=100
  ```

  Max size of the wait group that runs parallel plans and applies (if enabled). Defaults to `15`

### `--port`

  ```bash
  atlantis server --port=4141
  # or
  ATLANTIS_PORT=4141
  ```

  Port to bind to. Defaults to `4141`.

### `--quiet-policy-checks`

  ```bash
  atlantis server --quiet-policy-checks
  # or
  ATLANTIS_QUIET_POLICY_CHECKS=true
  ```

  Exclude policy check comments from pull requests unless there's an actual error from conftest. This also excludes warnings. Defaults to `false`.

### `--redis-db`

  ```bash
  atlantis server --redis-db=0
  # or
  ATLANTIS_REDIS_DB=0
  ```

  The Redis Database to use when using a Locking DB type of `redis`. Defaults to `0`.

### `--redis-host`

  ```bash
  atlantis server --redis-host="localhost"
  # or
  ATLANTIS_REDIS_HOST="localhost"
  ```

  The Redis Hostname for when using a Locking DB type of `redis`.

### `--redis-insecure-skip-verify`

  ```bash
  atlantis server --redis-insecure-skip-verify=false
  # or
  ATLANTIS_REDIS_INSECURE_SKIP_VERIFY=false
  ```

  Controls whether the Redis client verifies the Redis server's certificate chain and host name. If true, accepts any certificate presented by the server and any host name in that certificate. Defaults to `false`.

  ::: warning SECURITY WARNING
  If this is enabled, TLS is susceptible to machine-in-the-middle attacks unless custom verification is used.
  :::

### `--redis-password`

  ```bash
  atlantis server --redis-password="password123"
  # or (recommended)
  ATLANTIS_REDIS_PASSWORD="password123"
  ```

  The Redis Password for when using a Locking DB type of `redis`.

### `--redis-port`

  ```bash
  atlantis server --redis-port=6379
  # or
  ATLANTIS_REDIS_PORT=6379
  ```

  The Redis Port for when using a Locking DB type of `redis`. Defaults to `6379`.

### `--redis-tls-enabled`

  ```bash
  atlantis server --redis-tls-enabled=false
  # or
  ATLANTIS_REDIS_TLS_ENABLED=false
  ```

  Enables a TLS connection, with min version of 1.2, to Redis when using a Locking DB type of `redis`. Defaults to `false`.

### `--repo-allowlist`

  ```bash
  # NOTE: Use single quotes to avoid shell expansion of *.
  atlantis server --repo-allowlist='github.com/myorg/*'
  # or
  ATLANTIS_REPO_ALLOWLIST='github.com/myorg/*'
  ```

  Atlantis requires you to specify an allowlist of repositories it will accept webhooks from.

  Notes:

* Accepts a comma separated list, ex. `definition1,definition2`
* Format is `{hostname}/{owner}/{repo}`, ex. `github.com/runatlantis/atlantis`
* `*` matches any characters, ex. `github.com/runatlantis/*` will match all repos in the runatlantis organization
* An entry beginning with `!` negates it, ex. `github.com/foo/*,!github.com/foo/bar` will match all github repos in the `foo` owner *except* `bar`.
* For Bitbucket Server: `{hostname}` is the domain without scheme and port, `{owner}` is the name of the project (not the key), and `{repo}` is the repo name
  * User (not project) repositories take on the format: `{hostname}/{full name}/{repo}` (e.g., `bitbucket.example.com/Jane Doe/myatlantis` for username `jdoe` and full name `Jane Doe`, which is not very intuitive)
* For Azure DevOps the allowlist takes one of two forms: `{owner}.visualstudio.com/{project}/{repo}` or `dev.azure.com/{owner}/{project}/{repo}`
* Microsoft is in the process of changing Azure DevOps to the latter form, so it may be safest to always specify both formats in your repo allowlist for each repository until the change is complete.

  Examples:

* Allowlist `myorg/repo1` and `myorg/repo2` on `github.com`
  * `--repo-allowlist=github.com/myorg/repo1,github.com/myorg/repo2`
* Allowlist all repos under `myorg` on `github.com`
  * `--repo-allowlist='github.com/myorg/*'`
* Allowlist all repos under `myorg` on `github.com`, excluding `myorg/untrusted-repo`
  * `--repo-allowlist='github.com/myorg/*,!github.com/myorg/untrusted-repo'`
* Allowlist all repos in my GitHub Enterprise installation
  * `--repo-allowlist='github.yourcompany.com/*'`
* Allowlist all repos under `myorg` project `myproject` on Azure DevOps
  * `--repo-allowlist='myorg.visualstudio.com/myproject/*,dev.azure.com/myorg/myproject/*'`
* Allowlist all repositories
  * `--repo-allowlist='*'`

### `--repo-config`

  ```bash
  atlantis server --repo-config="path/to/repos.yaml"
  # or
  ATLANTIS_REPO_CONFIG="path/to/repos.yaml"
  ```

  Path to a YAML server-side repo config file. See [Server Side Repo Config](server-side-repo-config.md).

### `--repo-config-json`

  ```bash
  atlantis server --repo-config-json='{"repos":[{"id":"/.*/", "apply_requirements":["mergeable"]}]}'
  # or
  ATLANTIS_REPO_CONFIG_JSON='{"repos":[{"id":"/.*/", "apply_requirements":["mergeable"]}]}'
  ```

  Specify server-side repo config as a JSON string. Useful if you don't want to write a config file to disk.
  See [Server Side Repo Config](server-side-repo-config.md) for more details.

  ::: tip
  If specifying a [Workflow](custom-workflows.md#reference), [step](custom-workflows.md#step)'s
  can be specified as follows:

  ```json
  {
    "repos": [],
    "workflows": {
      "custom": {
        "plan": {
          "steps": [
            "init",
            {
              "plan": {
                "extra_args": ["extra", "args"]
              }
            },
            {
              "run": "my custom command"
            }
          ]
        }
      }
    }
  }
  ```

  :::

### `--restrict-file-list`

  ```bash
  atlantis server --restrict-file-list
  # or (recommended)
  ATLANTIS_RESTRICT_FILE_LIST=true
  ```

  `--restrict-file-list` will block plan requests from projects outside the files modified in the pull request.
  This will not block plan requests with regex if using the `--enable-regexp-cmd` flag, in these cases commands
  like `atlantis plan -p .*` will still work if used. normal commands will stil be blocked if necessary.
  Defaults to `false`.

### `--silence-allowlist-errors`

  ```bash
  atlantis server --silence-allowlist-errors
  # or
  ATLANTIS_SILENCE_ALLOWLIST_ERRORS=true
  ```

  Some users use the `--repo-allowlist` flag to control which repos Atlantis
  responds to. Normally, if Atlantis receives a pull request webhook from a repo not listed
  in the allowlist, it will comment back with an error. This flag disables that commenting.

  Some users find this useful because they prefer to add the Atlantis webhook
  at an organization level rather than on each repo.

### `--silence-fork-pr-errors`

  ```bash
  atlantis server --silence-fork-pr-errors
  # or
  ATLANTIS_SILENCE_FORK_PR_ERRORS=true
  ```

  Normally, if Atlantis receives a pull request webhook from a fork and --allow-fork-prs is not set,
  it will comment back with an error. This flag disables that commenting.

### `--silence-no-projects`

  ```bash
  atlantis server --silence-no-projects
  # or
  ATLANTIS_SILENCE_NO_PROJECTS=true
  ```

  `--silence-no-projects` will tell Atlantis to ignore PRs if none of the modified files are part of a project defined in the `atlantis.yaml` file.
  This flag ensures an Atlantis server only responds to its explicitly declared projects.
  This has no effect if projects are undefined in the repo level `atlantis.yaml`.
  This also silences targeted commands (eg. `atlantis plan -d mydir` or `atlantis apply -p myproj`) so if the project is not in the repo config `atlantis.yaml`, these commands will not run or report back in a comment.

  This is useful when running multiple Atlantis servers against a single repository so you can
  delegate work to each Atlantis server. Also useful when used with pre_workflow_hooks to dynamically generate an `atlantis.yaml` file.

### `--silence-vcs-status-no-plans`

  ```bash
  atlantis server --silence-vcs-status-no-plans
  # or
  ATLANTIS_SILENCE_VCS_STATUS_NO_PLANS=true
  ```

  `--silence-vcs-status-no-plans` will tell Atlantis to ignore setting VCS status on plans if none of the modified files are part of a project defined in the `atlantis.yaml` file.

### `--silence-vcs-status-no-projects`

  ```bash
  atlantis server --silence-vcs-status-no-projects
  # or
  ATLANTIS_SILENCE_VCS_STATUS_NO_PROJECTS=true
  ```

  `--silence-vcs-status-no-projects` will tell Atlantis to ignore setting VCS status on any command if none of the modified files are part of a project defined in the `atlantis.yaml` file.

### `--skip-clone-no-changes`

  ```bash
  atlantis server --skip-clone-no-changes
  # or
  ATLANTIS_SKIP_CLONE_NO_CHANGES=true
  ```

  `--skip-clone-no-changes` will skip cloning the repo during autoplan if there are no changes to Terraform projects. This will only apply for GitHub and GitLab and only for repos that have `atlantis.yaml` file. Defaults to `false`.

### `--slack-token`

  ```bash
  atlantis server --slack-token=token
  # or (recommended)
  ATLANTIS_SLACK_TOKEN='token'
  ```

  API token for Slack notifications. See [Using Slack hooks](using-slack-hooks.md).

### `--ssl-cert-file`

  ```bash
  atlantis server --ssl-cert-file="/etc/ssl/certs/my-cert.crt"
  # or
  ATLANTIS_SSL_CERT_FILE="/etc/ssl/certs/my-cert.crt"
  ```

  File containing x509 Certificate used for serving HTTPS.
  If the cert is signed by a CA, the file should be the concatenation
  of the server's certificate, any intermediates, and the CA's certificate.

### `--ssl-key-file`

  ```bash
  atlantis server --ssl-key-file="/etc/ssl/private/my-cert.key"
  # or
  ATLANTIS_SSL_KEY_FILE="/etc/ssl/private/my-cert.key"
  ```

  File containing x509 private key matching `--ssl-cert-file`.

### `--stats-namespace`

  ```bash
  atlantis server --stats-namespace="myatlantis"
  # or
  ATLANTIS_STATS_NAMESPACE="myatlantis"
  ```

  Namespace for emitting stats/metrics. See [stats](stats.md) section.

### `--tf-distribution`
  ```bash
  atlantis server --tf-distribution="terraform"
  # or
  ATLANTIS_TF_DISTRIBUTION="terraform"
  ```
  Which TF distribution to use. Can be set to `terraform` or `opentofu`.

### `--tf-download`

  ```bash
  atlantis server --tf-download=false
  # or
  ATLANTIS_TF_DOWNLOAD=false
  ```

Defaults to `true`. Allow Atlantis to list and download additional versions of Terraform.
Setting this to `false` can be useful in an air-gapped environment where a download mirror is not available.

### `--tf-download-url`

  ```bash
  atlantis server --tf-download-url="https://releases.company.com"
  # or
  ATLANTIS_TF_DOWNLOAD_URL="https://releases.company.com"
  ```

  An alternative URL to download Terraform versions if they are missing. Useful in an airgapped
  environment where releases.hashicorp.com is not available. Directory structure of the custom
  endpoint should match that of releases.hashicorp.com.

  This has no impact if `--tf-download` is set to `false`.

  This setting is not yet supported when `--tf-distribution` is set to `opentofu`.

### `--tfe-hostname`

  ```bash
  atlantis server --tfe-hostname="my-terraform-enterprise.company.com"
  # or
  ATLANTIS_TFE_HOSTNAME="my-terraform-enterprise.company.com"
  ```

  Hostname of your Terraform Enterprise installation to be used in conjunction with
  `--tfe-token`. See [Terraform Cloud](terraform-cloud.md) for more details.
  If using Terraform Cloud (i.e. you don't have your own Terraform Enterprise installation)
  no need to set since it defaults to `app.terraform.io`.

### `--tfe-local-execution-mode`

  ```bash
  atlantis server --tfe-local-execution-mode
  # or
  ATLANTIS_TFE_LOCAL_EXECUTION_MODE=true
  ```

  Enable if you're using local execution mode (instead of TFE/C's remote execution mode). See [Terraform Cloud](terraform-cloud.md) for more details.

### `--tfe-token`

  ```bash
  atlantis server --tfe-token="xxx.atlasv1.yyy"
  # or (recommended)
  ATLANTIS_TFE_TOKEN='xxx.atlasv1.yyy'
  ```

  A token for Terraform Cloud/Terraform Enterprise integration. See [Terraform Cloud](terraform-cloud.md) for more details.

### `--use-tf-plugin-cache`

```bash
atlantis server --use-tf-plugin-cache=false
# or
ATLANTIS_USE_TF_PLUGIN_CACHE=false
```

Set to false if you want to disable terraform plugin cache.

This flag is useful when having multiple projects that need to run a plan and apply in the same PR to avoid the race condition of `plugin_cache_dir` concurrently, this is a terraform known issue, more info:

* [plugin_cache_dir concurrently discussion](https://github.com/hashicorp/terraform/issues/31964)
* [PR to improve the situation](https://github.com/hashicorp/terraform/pull/33479)

The effect of the race condition is more evident when using parallel configuration to run plan and apply, by disabling the use of plugin cache will impact in the performance when starting a new plan or apply, but in large atlantis deployments with multiple projects and shared modules the use of `--parallel_plan` and `--parallel_apply` is mandatory for an efficient managment of the PRs.

### `--var-file-allowlist`

  ```bash
  atlantis server --var-file-allowlist='/path/to/tfvars/dir'
  # or
  ATLANTIS_VAR_FILE_ALLOWLIST='/path/to/tfvars/dir'
  ```

  Comma-separated list of additional directory paths where [variable definition files](https://developer.hashicorp.com/terraform/language/values/variables#variable-definitions-tfvars-files) can be read from.
  The paths in this argument should be absolute paths. Relative paths and globbing are currently not supported.
  If this argument is not provided, it defaults to Atlantis' data directory, determined by the `--data-dir` argument.

### `--vcs-status-name`

  ```bash
  atlantis server --vcs-status-name="atlantis-dev"
  # or
  ATLANTIS_VCS_STATUS_NAME="atlantis-dev"
  ```

  Name used to identify Atlantis when updating a pull request status. Defaults to `atlantis`.

  This is useful when running multiple Atlantis servers against a single repository so you can
  give each Atlantis server its own unique name to prevent the statuses clashing.

### `--web-basic-auth`

  ```bash
  atlantis server --web-basic-auth
  # or
  ATLANTIS_WEB_BASIC_AUTH=true
  ```

  Enable Basic Authentication on the Atlantis web service.

### `--web-password`

  ```bash
  atlantis server --web-password="atlantis"
  # or
  ATLANTIS_WEB_PASSWORD="atlantis"
  ```

  Password used for Basic Authentication on the Atlantis web service. Defaults to `atlantis`.

### `--web-username`

  ```bash
  atlantis server --web-username="atlantis"
  # or
  ATLANTIS_WEB_USERNAME="atlantis"
  ```

  Username used for Basic Authentication on the Atlantis web service. Defaults to `atlantis`.

### `--websocket-check-origin`

  ```bash
  atlantis server --websocket-check-origin
  # or
  ATLANTIS_WEBSOCKET_CHECK_ORIGIN=true
  ```

  Only allow websockets connection when they originate from the running Atlantis web server

### `--write-git-creds`

  ```bash
  atlantis server --write-git-creds
  # or
  ATLANTIS_WRITE_GIT_CREDS=true
  ```

  Write out a .git-credentials file with the provider user and token to allow
  cloning private modules over HTTPS or SSH. See [here](https://git-scm.com/docs/git-credential-store) for more information.

  Follow the `git::ssh` syntax to avoid using a custom `.gitconfig` with an `insteadOf`.

  ```hcl
  module "private_submodule" {
    source = "git::ssh://git@github.com/<org>/<repo>//modules/<some-module-name>?ref=v1.2.3"

    # ...
  }
  ```

  ::: warning SECURITY WARNING
  This does write secrets to disk and should only be enabled in a secure environment.
  :::
