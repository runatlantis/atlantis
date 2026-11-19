# Atlantis threat model

## What this project does and where untrusted input enters

Atlantis is a self-hosted Go server for Terraform/OpenTofu pull request automation. It receives
webhooks from a VCS provider (GitHub, GitLab, Bitbucket Cloud/Server, Azure DevOps, Gitea), clones
the pull request, runs `terraform plan`/`apply` in a per-PR working directory, and comments the
results back. It usually runs with cloud credentials that can change production infrastructure,
plus a VCS token or GitHub App key with write access to the repositories it serves.

Untrusted input, roughly from most to least exposed:

- **HTTP requests to `POST /events`** from anyone who can reach the server. Webhook authentication
  is per provider: HMAC signature (GitHub, Gitea), shared token (GitLab), basic auth (Azure DevOps).
  Requests that fail authentication must be rejected before any state changes.
- **Pull request metadata** set by the PR author: branch names, PR title/body, head repository
  (forks), changed file paths, and commit contents.
- **PR comments** (`atlantis plan/apply/unlock/approve_policies/import/state ...`) from any user who
  can comment, including flags (`-p`, `-d`, `-w`, `--verbose`) and extra arguments after `--` that are
  passed to Terraform. Parsing is in `server/events/comment_parser.go`.
- **Repo-level `atlantis.yaml`** from the PR branch, parsed in `server/core/config/`. Which keys it
  may set is limited by the operator's server-side repo config (`allowed_overrides`,
  `allow_custom_workflows`).
- **The web UI and REST API**: `/locks`, `/lock`, `/jobs/{id}` and its websocket, `/api/plan`,
  `/api/apply`, `/api/locks`, `/api/drift/*` (token protected), `/github-app/*`, `/status`.

Trusted: the operator, the server flags and server-side repo config, the VCS provider's API
responses, and the host filesystem outside the Atlantis data directory.

## Components that matter most / least

Most important:

- Webhook validation and event parsing: `server/controllers/events/` and the provider packages in
  `server/events/vcs/*`.
- Authorization of who may run which command: repo allowlist (`--repo-allowlist`), team/user
  allowlists, fork PR handling (`--allow-fork-prs`), apply requirements (`approved`, `mergeable`,
  `undiverged`), policy-check approval, and locking (`server/core/locking`, `server/events/*lock*`).
- Path and command construction from PR-controlled strings: workspace names, project directories,
  repo-relative paths, var-file paths (`--var-file-allowlist`), and anything that reaches a shell or
  `exec.Command` (`server/core/runtime/`, `server/events/working_dir.go`).
- Credential handling: VCS tokens and GitHub App keys must not reach logs, PR comments, job output
  pages, or the environment of steps a PR author controls beyond what the operator configured.
- The REST API token check and the web UI's lock/job endpoints (`server/controllers/`).

Less important: the VitePress docs site (`runatlantis.io/`), `e2e/`, `.github/scripts/`, generated
mocks, and `testdrive/`. Out of scope.

## How to exercise it

- The checkout is at `/src`; the server binary is `/usr/local/bin/atlantis`. Terraform and conftest
  are on `PATH`. The network is disabled; `GOPROXY=off` and `GOTOOLCHAIN=local` stop downloads.
- Unit tests: `runuser -u atlantis -- env HOME=/home/atlantis make test` (`go test -short` over the
  root module's packages). Run them as the `atlantis` user: as root, two tests that check unreadable
  files are rejected (`TestParseRepoCfg_BadPermissions`, `TestGenerateRCFile_ErrIfCannotRead`) fail.
- Webhook-to-comment integration tests drive the whole pipeline with fixture payloads and real git
  repos: `go test ./server/controllers/events/ -run TestGitHubWorkflow` (fixtures in
  `server/controllers/events/testdata/`). Tests that need a Terraform provider from the registry
  will fail offline; that is expected.
- Comment parsing, config parsing, and webhook parsing have table-driven tests next to the code and
  are good places to add inputs.
- Start a server locally with a fake provider for manual probing, for example:
  `atlantis server --gh-user=x --gh-token=x --repo-allowlist='*' --gh-webhook-secret=s` and send
  signed payloads to `http://localhost:4141/events`.

## How you rate severity

- **Critical**: unauthenticated remote code execution on the Atlantis host, or leaking VCS/cloud
  credentials to an unauthenticated attacker.
- **High**: a PR author or commenter executing commands, reading/writing files outside the PR's
  working directory, or reaching credentials beyond what the operator's configuration allows them;
  bypassing webhook authentication, the repo allowlist, team allowlists, fork restrictions, apply
  requirements, policy approvals, or locks; running `apply` without the required approvals.
- **Medium**: denial of service from a single small request (crash, unbounded memory/disk, lock
  starvation), cross-PR data exposure in the UI or job output, stored XSS in the web UI.
- **Low**: information disclosure without credentials (versions, repo names), issues needing
  operator misconfiguration that the docs already warn against.

## Anything to leave alone

Atlantis runs Terraform on code from the pull request, and that is code execution by design. Do not
report any of these, which are documented in `runatlantis.io/docs/security.md`:

- Code execution through Terraform itself during `plan` or `apply`: malicious providers, the
  `external` data source, `local-exec` provisioners, modules from arbitrary sources.
- Custom `run` steps or workflows in a repo-level `atlantis.yaml` when the operator has enabled
  `allow_custom_workflows` or allowed the `workflow` override.
- Anything that requires `--repo-allowlist='*'` without a webhook secret, `--allow-fork-prs` on a
  public repository, or other protections the operator chose to turn off.

These are still in scope when Atlantis lets an attacker reach them without the operator having
opted in, for example a workspace name that becomes a shell command, a repo-level config key that
should have been rejected by `allowed_overrides`, or a var-file path that escapes the allowlist.
