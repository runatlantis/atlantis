# Atlantis - Terraform Pull Request Automation

**What it does:** Self-hosted Go application that listens for Terraform PR webhooks, runs `terraform plan/apply`, and comments results back to PRs.

**Stack:** Go 1.25.8 • 381 Go files • Gorilla Mux • Cobra CLI • Viper config • VitePress docs (Node.js) • Docker deployment

**Key Info:** ~35MB repo, server + CLI app, E2E tests with Playwright, integration tests with Terraform

> **AI Usage Policy:** Before contributing with AI assistance, read the [AI_USAGE_POLICY.md](AI_USAGE_POLICY.md).

## Build & Test (Always from repo root)

**Prerequisites:** Go 1.25.8 (from go.mod) • Node 20+ & npm 10+ (website) • Docker • Terraform 1.11.1+ (integration tests)

**Build:** `make build-service` → creates `./atlantis` binary (~51MB, 30-60s first run, 10s subsequent). Clean: `make clean`

**Test:** `make test` (unit, ~60s) • `make test-all` (includes integration, ~5min) • `make docker/test-all` (CI environment)
⚠️ **Known failing test:** `TestNewServer_GitHubUser` in server/server_test.go - pre-existing, ignore it

**Lint/Format:** `make check-fmt` (ALWAYS works) • `make fmt` (auto-format)
⚠️ **Known issue:** `make lint` fails with Go 1.25+ version mismatch, and `check-lint` was removed. Use `make check-fmt` locally; CI handles linting.

**Mocks:** `make go-generate` (regenerate after interface changes) • `make regen-mocks` (delete & regenerate all)

**Website (VitePress):** `npm install` (required first) • `npm run website:dev` (<http://localhost:8080>) • `npm run website:build` • `npm run e2e` (Playwright)
📁 Docs: `runatlantis.io/docs/*.md` • Config: `runatlantis.io/.vitepress/config.js`
📝 Markdown linting: Done in CI with `DavidAnson/markdownlint-cli2-action`. For local linting, use an editor extension or run the GitHub Actions workflow.

## Architecture

**Structure:** `cmd/` (CLI) • `server/` (main app: controllers, core logic, events, vcs integrations) • `e2e/` (E2E tests) • `runatlantis.io/` (docs) • `scripts/` (build utils)

**Key paths:** `main.go` (entry) • `cmd/server.go` • `server/server.go` (init) • `server/router.go` • `server/controllers/events/events_controller.go` (webhooks)

**Core logic:** `server/core/config/` (parsing), `server/core/runtime/` (Terraform execution), `server/core/terraform/tfclient/` (TF client)
**Localization:** `server/i18n/` (embedded YAML catalogs and runtime overrides), `server/events/templates/i18n/<lang>/` (localized markdown template overrides)

**VCS providers:** `server/events/vcs/{github,gitlab,bitbucketcloud,bitbucketserver,azuredevops,gitea}/`

**Config files:** `.golangci.yml` (lint), `go.mod`, `Makefile`, `Dockerfile`, `docker-compose.yml`, `.pre-commit-config.yaml`

## CI Workflows (.github/workflows/)

**test.yml:** `make test-all` + `make check-fmt` in `ghcr.io/runatlantis/testing-env:latest` • E2E for GitHub/GitLab (skips on forks - no secrets)
**lint.yml:** golangci-lint via GitHub Action • Path filtered (Go files only)
**website.yml:** markdownlint-cli2-action → lychee link check → `npm install && npm run website:build` → `npm run e2e` (Playwright)
**Others:** pr-lint (Conventional Commits), codeql, scorecard, dependency-review

**Replicate CI locally:** `make test-all && make check-fmt` OR use Docker: `docker run --rm -v $(pwd):/atlantis ghcr.io/runatlantis/testing-env:latest sh -c "cd /atlantis && make test-all"`

**E2E tests:** Complex setup (ngrok + credentials). CI handles it. Local optional. See `./scripts/e2e.sh` for details.

- **GitHub E2E auth:** Supports two modes — GitHub App (preferred) via `ATLANTIS_GH_APP_ID` + `ATLANTIS_GH_APP_KEY` + `ATLANTIS_GH_APP_SLUG`, or PAT via `ATLANTIS_GH_USER` + `ATLANTIS_GH_TOKEN`. App auth avoids org 2FA restrictions.
- **E2E test code:** `e2e/` has its own `go.mod` (separate module from root). Build: `cd e2e && make build`. Run: `make run`.
- **Atlantis server in E2E:** Started by `scripts/e2e.sh` — reads GitHub auth from env vars automatically (viper). No explicit flags needed.

## Development Workflows

**Before commit:** `make test` → `make check-fmt` → `make go-generate` (if interfaces changed) → `make build-service` (verify)

**VCS provider:** Create `server/events/vcs/<provider>/` → Implement `Client` interface (`server/events/vcs/common/common.go`) → Update `server/server.go`

**VCS pagination:** When following provider-returned pagination links, validate the next URL against the configured provider API origin before issuing the request. Bitbucket Cloud diffstat pagination uses this guard in both modified-file and mergeability checks.

**Config changes:** Edit `server/core/config/valid/` or `raw/` → Update `server/user_config.go` → Test in `server/core/config/*_test.go`
**i18n changes:** Keep command routing/template selection keyed on stable command identifiers (`command.Name`), and use localized titles only for display text.

**Terraform execution:** Modify `server/core/terraform/tfclient/terraform_client.go` or `server/core/runtime/*_step_runner.go` (uses `hashicorp/hc-install`)

**Combined VCS statuses:** `server/events/commit_status_updater.go` uses `models.ProjectCounts`. For failed apply/policy status text, count actual errored projects with `Errored`, not `Total - Success`, because planned or untouched projects may still be pending. For apply status text, `NoChanges` is a subset of `Success` and should be reported as up to date, not applied.

**VCS status contexts:** Always pass generated commit status contexts through `truncateContext` before calling `UpdateStatus`. GitHub rejects contexts longer than 255 characters, and project names or workflow hook descriptions can exceed that limit.

**GitHub team allowlist:** `GH_TEAM_ALLOWLIST` honors GitHub child-team membership. Keep hierarchy expansion available to both initial command authorization and later project filtering such as generated `policy_check` contexts.

**Apply requirements:** API apply requests must populate `command.Context.PullRequestStatus` before evaluating apply requirements, and refresh it after the API plan phase. If pull status cannot be fetched, keep the fail-closed behavior for `approved` and `mergeable`.

**GitLab mergeable applies:** `mergeable` apply requirements are scoped per project only when GitLab reports blocking Atlantis per-project plan statuses for other projects. Preserve the conservative behavior: the current project's own plan, external CI, approvals, conflicts, and unknown blockers still fail the requirement.

**GitLab mergeability status filtering:** GitLab `PullIsMergeable` must filter commit statuses to the merge request's current ref/SHA before deciding which statuses block mergeability. Do not let stale statuses from another branch, ref, or earlier commit contaminate the MR's mergeability result.

**Apply locks:** Global apply lock check errors must reject `atlantis apply`, set the command as errored, update the apply status to failed, and comment that the lock backend must be reachable. Do not fall back to allowing applies when the lock backend is unavailable.

**Plan statistics:** `models.NewPlanSuccessStats` parses Terraform/OpenTofu summary lines including `to forget`. Preserve import/add/change/destroy/forget counts when changing plan rendering or parser regexes.

**GitHub App merge checkout:** In `server/events/working_dir.go`, `CheckoutMerge` with `GithubAppEnabled` and a PR number fetches `pull/<n>/head` from `origin` and intentionally skips the `source` remote. Preserve this fork-safe behavior when changing clone, fetch, or divergence logic.

**GitHub Enterprise Cloud:** `*.ghe.com` hosts use `https://api.<tenant>.ghe.com/` for REST and `/graphql` on that API host for GraphQL, not the GitHub Enterprise Server `/api/v3` and `/api/graphql` paths. For GitHub App credentials on GHE Cloud, preserve the app-level JWT `/app` lookup used to derive the bot slug.

**Pending plans:** `server/events/pending_plan_finder.go` should only scan workspace clone roots. Skip files, symlinks, and non-git directories under the pull directory before running `git ls-files` so stray directories do not look like pending Atlantis plans.

**Working dir ref refresh:** After `server/events/working_dir.go` resets or redoes a merge for a newer ref, untracked `.tfplan` files must be removed while preserving `.terragrunt-cache`. Stale plan files from a previous ref must not survive branch or merge checkout refreshes.

**Command output rendering:** `server/core/runtime/models/shell_command_runner.go` must preserve long single-line stdout/stderr output by keeping the enlarged scanner buffer. `PlanSuccess.DiffMarkdownFormattedTerraformOutput` should keep heredoc and multiline-string diff markers aligned so changed content remains colorized in markdown diff blocks.

## Known Issues

1. **TestNewServer_GitHubUser fails:** Pre-existing in main. Ignore it.
2. **E2E tests skip on forks:** Expected (no secrets). Maintainers run them.
3. **Website needs npm install first:** Always run `npm install` before `npm run website:*` commands.
4. **docker-compose needs atlantis.env:** Create file per CONTRIBUTING.md template for local webhook testing.
5. **E2E GitHub tests require GitHub App secrets:** CI needs `ATLANTISBOT_GH_APP_ID`, `ATLANTISBOT_GH_APP_KEY`, `ATLANTISBOT_GH_APP_SLUG` secrets configured. PAT auth (`ATLANTISBOT_GITHUB_USERNAME`/`ATLANTISBOT_GITHUB_TOKEN`) is deprecated due to org 2FA requirements (#6311).

## Code Style

**Logging:** Use `ctx.Log` • lowercase • quote strings with `%q` • NO colons (reserved for errors) • Levels: debug/info/warn/error
**Errors:** Lowercase • `fmt.Errorf("context: %w", err)` not `%s` • Describe action, not "failed to" • Example: "running git clone: no executable"
**Testing:** Tests in `{package}_test` • Internal: `{file}_internal_test.go` • Use `import . "github.com/runatlantis/atlantis/testing"` • `Assert()`, `Equals()`, `Ok()`
**Commits:** Conventional Commits (`fix:`, `feat:`, etc.) • Sign with `-s` (DCO)

## Pre-PR Checklist

✓ `make test-all` (ignore TestNewServer_GitHubUser) ✓ `make check-fmt` ✓ `make go-generate` (if interfaces changed) ✓ Website builds (docs changes)
✓ Conventional Commits format ✓ Signed commits (-s) ✓ Tests added ✓ Docs updated

## Quick Commands

**Daily:** `make build-service` • `make test` • `make check-fmt`
**Pre-commit:** `make test-all` • `make check-fmt`
**Website:** `npm install` • `npm run website:dev` • `npm run website:build`
**Coverage:** `make test-coverage-html`
**Docker:** `make docker/dev` • `docker-compose up`

---

**Trust these instructions first.** Search codebase only if info is incomplete/incorrect. Validated 2026-04-20 • Go 1.25.8
