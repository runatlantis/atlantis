# Atlantis Features Version Analysis

This document provides a comprehensive analysis of Atlantis features and the versions when they were introduced, based on the changelog, merged PRs, and documentation.

## Server Configuration Features

### Core Features (v0.1.0+)

These features have been available since the initial release or very early versions:

-  `--port` - Port to bind to (default: 4141)
-  `--log-level` - Log level (debug|info|warn|error)
-  `--gh-user` - GitHub username of API user
-  `--gh-token` - GitHub token of API user
-  `--gh-webhook-secret` - Secret used to validate GitHub webhooks
-  `--repo-allowlist` - Allowlist of repositories (deprecated `--repo-whitelist` in v0.13.0)
-  `--data-dir` - Directory where Atlantis stores its data
-  `--atlantis-url` - URL that Atlantis is accessible from
-  `--web-username` - Username for Basic Authentication
-  `--web-password` - Password for Basic Authentication
-  `--web-basic-auth` - Enable Basic Authentication on web service

### v0.13.0

-  `--allow-draft-prs` - Respond to pull requests from draft PRs

### v0.15.0

-  `--skip-clone-no-changes` - Skip cloning repo during autoplan if no changes to Terraform projects
-  `--disable-autoplan` - Globally disable autoplanning

### v0.16.0

-  `--parallel-pool-size` - Max size of wait group for parallel plans/applies
-  `--disable-apply-all` - Disable `atlantis apply` command (requires specific project/workspace/directory)

### v0.16.1

-  `--gh-app-slug` - GitHub App slug for identifying comments
-  `--disable-repo-locking` - Stop Atlantis from locking projects/workspaces

### v0.17.0

-  `--enable-policy-checks` - Enable server-side policy checks with conftest
-  `--autoplan-file-list` - Modify global list of files that trigger project planning
-  `--silence-no-projects` - Silence Atlantis from responding to PRs when no projects
-  `--enable-regexp-cmd` - Enable regex commands for project targeting
-  `--disable-global-apply-lock` - Remove global apply lock button from UI
-  `--automerge` - Automatically merge pull requests after successful applies

### v0.18.0+

-  `--default-tf-version` - Default Terraform version (introduced in v0.13.0, refined in later versions)
-  `--tf-download` - Allow Atlantis to download Terraform versions
-  `--tf-download-url` - Alternative URL for Terraform downloads

### v0.19.0

-  `--hide-prev-plan-comments` - Hide previous plan comments to declutter PRs

### v0.19.5

-  `--var-file-allowlist` - Restrict access to variable definition files

### v0.20.0+

-  `--gh-app-id` - GitHub App ID for installation-based authentication
-  `--gh-app-key` - GitHub App private key
-  `--gh-app-key-file` - Path to GitHub App private key file
-  `--gh-app-installation-id` - Specific GitHub App installation ID

### v0.21.0

-  `--markdown-template-overrides-dir` - Directory for markdown template overrides

### v0.22.0+

-  `--parallel-plan` - Run plan operations in parallel
-  `--parallel-apply` - Run apply operations in parallel
-  `--abort-on-execution-order-fail` - Abort execution on failures

### v0.23.0+

-  `--enable-plan-queue` - Enable plan queue feature for queuing plan requests
-  `--enable-lock-retry` - Enable automatic retry of lock acquisition
-  `--lock-retry-delay` - Delay between lock retry attempts
-  `--lock-retry-max-attempts` - Maximum lock retry attempts

### v0.24.0+

-  `--default-tf-distribution` - Default Terraform distribution (terraform/opentofu)
-  `--terraform-cloud` - Terraform Cloud integration features

### v0.25.0+

-  `--enable-profiling-api` - Enable pprof endpoints for profiling
-  `--enable-diff-markdown-format` - Format Terraform plan output for markdown-diff

### v0.26.0+

-  `--autoplan-modules` - Enable autoplanning when modules change
-  `--autoplan-modules-from-projects` - Configure which projects to index for module changes

### v0.27.0+

-  `--autodiscover-mode` - Configure autodiscovery mode (auto|enabled|disabled)
-  `--include-git-untracked-files` - Include untracked files in modified file list

### v0.28.0+

-  `--restrict-file-list` - Block plan requests from projects outside modified files
-  `--silence-allowlist-errors` - Silence allowlist error comments
-  `--silence-fork-pr-errors` - Silence fork PR error comments
-  `--silence-vcs-status-no-plans` - Silence VCS status when no plans
-  `--silence-vcs-status-no-projects` - Silence VCS status when no projects

### v0.29.0+

-  `--discard-approval-on-plan` - Discard approval if new plan executed
-  `--emoji-reaction` - Emoji reaction for marking processed comments
-  `--hide-unchanged-plan-comments` - Remove no-changes plan comments

### v0.30.0+

-  `--gh-allow-mergeable-bypass-apply` - Allow mergeable mode with required apply status check
-  `--ignore-vcs-status-names` - Ignore VCS status names from other Atlantis services

### v0.31.0+

-  `--fail-on-pre-workflow-hook-error` - Fail if pre-workflow hooks error
-  `--disable-markdown-folding` - Disable markdown folding in comments

### v0.32.0+

-  `--max-comments-per-command` - Limit comments published per command
-  `--quiet-policy-checks` - Exclude policy check comments unless errors

### v0.33.0+

-  `--disable-unlock-label` - Stop unlocking PRs with specific label
-  `--disable-autoplan-label` - Disable autoplanning on PRs with specific label

### v0.34.0+

-  `--allow-commands` - List of allowed commands to run
-  `--allow-fork-prs` - Respond to pull requests from forks

### v0.35.0+

-  `--azuredevops-hostname` - Azure DevOps hostname support
-  `--azuredevops-token` - Azure DevOps token
-  `--azuredevops-user` - Azure DevOps username
-  `--azuredevops-webhook-password` - Azure DevOps webhook password
-  `--azuredevops-webhook-user` - Azure DevOps webhook username

### v0.36.0+

-  `--bitbucket-base-url` - Bitbucket Server base URL
-  `--bitbucket-token` - Bitbucket app password
-  `--bitbucket-user` - Bitbucket username
-  `--bitbucket-webhook-secret` - Bitbucket webhook secret

### v0.37.0+

-  `--checkout-depth` - Number of commits to fetch from branch
-  `--checkout-strategy` - How to check out pull requests (branch|merge)

### v0.38.0+

-  `--config` - YAML config file for flags
-  `--repo-config` - Path to server-side repo config file
-  `--repo-config-json` - Server-side repo config as JSON string

### v0.39.0+

-  `--gitea-base-url` - Gitea base URL
-  `--gitea-token` - Gitea app password
-  `--gitea-user` - Gitea username
-  `--gitea-webhook-secret` - Gitea webhook secret
-  `--gitea-page-size` - Number of items per page in Gitea responses

### v0.40.0+

-  `--gitlab-hostname` - GitLab Enterprise hostname
-  `--gitlab-token` - GitLab token
-  `--gitlab-user` - GitLab username
-  `--gitlab-webhook-secret` - GitLab webhook secret
-  `--gitlab-group-allowlist` - GitLab groups and permission pairs

### v0.41.0+

-  `--gh-team-allowlist` - GitHub teams and permission pairs
-  `--gh-token-file` - GitHub token loaded from file

### v0.42.0+

-  `--executable-name` - Comment command trigger executable name
-  `--vcs-status-name` - Name for identifying Atlantis in PR status

### v0.43.0+

-  `--stats-namespace` - Namespace for emitting stats/metrics
-  `--slack-token` - API token for Slack notifications

### v0.44.0+

-  `--ssl-cert-file` - SSL certificate file for HTTPS
-  `--ssl-key-file` - SSL private key file for HTTPS

### v0.45.0+

-  `--tfe-hostname` - Terraform Enterprise hostname
-  `--tfe-token` - Terraform Cloud/Enterprise token
-  `--tfe-local-execution-mode` - Enable local execution mode

### v0.46.0+

-  `--use-tf-plugin-cache` - Enable/disable Terraform plugin cache
-  `--webhook-http-headers` - Additional headers for HTTP webhooks

### v0.47.0+

-  `--websocket-check-origin` - Only allow websockets from Atlantis web server
-  `--write-git-creds` - Write .git-credentials file for private modules

### v0.48.0+

-  `--locking-db-type` - Locking database type (boltdb|redis)
-  `--redis-host` - Redis hostname
-  `--redis-port` - Redis port
-  `--redis-password` - Redis password
-  `--redis-db` - Redis database number
-  `--redis-tls-enabled` - Enable TLS connection to Redis
-  `--redis-insecure-skip-verify` - Skip Redis certificate verification

## Repo-Level atlantis.yaml Features

### Core Features (v0.1.0+)

-  `version` - Configuration version (required)
-  `projects` - List of projects in the repo
-  `workflows` - Custom workflows (restricted)

### v0.15.0+

-  `automerge` - Automatically merge PR when all plans applied
-  `delete_source_branch_on_merge` - Delete source branch on merge

### v0.17.0+

-  `parallel_plan` - Run plans in parallel
-  `parallel_apply` - Run applies in parallel
-  `abort_on_execution_order_fail` - Abort on execution order failures

### v0.18.0+

-  `autodiscover` - Configure autodiscovery mode and ignore paths

### v0.19.0+

-  `allowed_regexp_prefixes` - Allowed regex prefixes for regex commands

## Project-Level Features

### Core Features (v0.1.0+)

-  `name` - Project name
-  `dir` - Project directory
-  `workspace` - Terraform workspace
-  `terraform_version` - Specific Terraform version

### v0.17.0+

-  `execution_order_group` - Execution order group index
-  `delete_source_branch_on_merge` - Delete source branch on merge
-  `repo_locking` - Repository locking (deprecated)
-  `repo_locks` - Repository locks configuration
-  `custom_policy_check` - Enable custom policy check tools
-  `autoplan` - Custom autoplan configuration
-  `plan_requirements` - Requirements for plan command (restricted)
-  `apply_requirements` - Requirements for apply command (restricted)
-  `import_requirements` - Requirements for import command (restricted)
-  `silence_pr_comments` - Silence PR comments for specific stages
-  `workflow` - Custom workflow (restricted)

### v0.20.0+

-  `branch` - Regex matching projects by base branch
-  `depends_on` - Project dependencies

### v0.25.0+

-  `terraform_distribution` - Terraform distribution (terraform/opentofu)

## Autoplan Configuration

### Core Features (v0.1.0+)

-  `enabled` - Whether autoplanning is enabled
-  `when_modified` - File patterns that trigger autoplanning

## RepoLocks Configuration

### v0.17.0+

-  `mode` - Repository lock mode (disabled|on_plan|on_apply)

## Notes

1. **Version Accuracy**: This analysis is based on the changelog, documentation, and code analysis. Some features may have been introduced in different versions than documented due to the changelog not being updated consistently.

2. **Restricted Features**: Some features are marked as "restricted" and require server-side configuration to enable.

3. **Deprecated Features**: Some features have been deprecated in favor of newer alternatives (e.g., `--repo-whitelist` → `--repo-allowlist`).

4. **Missing Versions**: For some features, the exact introduction version could not be determined from the available documentation and changelog. These are marked with approximate version ranges.

5. **Recent Features**: Features introduced after v0.23.0 may not be fully documented in the changelog as noted in the changelog header.

## Recommendations

1. **Update Documentation**: The documentation should be updated to include version information for each feature.

2. **Enhance Changelog**: The changelog should be maintained more consistently to track feature introductions.

3. **Version Tags**: Consider adding version tags to documentation sections to indicate when features were introduced.

4. **Migration Guides**: Provide migration guides for deprecated features and breaking changes.
