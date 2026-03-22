# Atlantis - Terraform Pull Request Automation

**What it does:** Self-hosted Go application that listens for Terraform PR webhooks, runs `terraform plan/apply`, and comments results back to PRs.

**Stack:** Go 1.25.4 • 381 Go files • Gorilla Mux • Cobra CLI • Viper config • VitePress docs (Node.js) • Docker deployment

**Key Info:** ~35MB repo, server + CLI app, E2E tests with Playwright, integration tests with Terraform

## Build & Test (Always from repo root)

**Prerequisites:** Go 1.25.4 (from go.mod, not .tool-versions) • Node 20+ & npm 10+ (website) • Docker • Terraform 1.11.1+ (integration tests)

**Build:** `make build-service` → creates `./atlantis` binary (~51MB, 30-60s first run, 10s subsequent). Clean: `make clean`

**Test:** `make test` (unit, ~60s) • `make test-all` (includes integration, ~5min) • `make docker/test-all` (CI environment)
⚠️ **Known failing test:** `TestNewServer_GitHubUser` in server/server_test.go - pre-existing, ignore it

**Lint/Format:** `make check-fmt` (ALWAYS works) • `make fmt` (auto-format)
⚠️ **Known issue:** `make lint` and `make check-lint` fail with Go 1.25+ version mismatch. Use `make check-fmt` locally, CI handles linting.

**Mocks:** `make go-generate` (regenerate after interface changes) • `make regen-mocks` (delete & regenerate all)

**Website (VitePress):** `npm install` (required first) • `npm run website:dev` (http://localhost:8080) • `npm run website:build` • `npm run website:lint` • `npm run e2e` (Playwright)
📁 Docs: `runatlantis.io/docs/*.md` • Config: `runatlantis.io/.vitepress/config.js`

## Architecture

**Structure:** `cmd/` (CLI) • `server/` (main app: controllers, core logic, events, vcs integrations) • `e2e/` (E2E tests) • `runatlantis.io/` (docs) • `scripts/` (build utils)

**Key paths:** `main.go` (entry) • `cmd/server.go` • `server/server.go` (init) • `server/router.go` • `server/controllers/events/events_controller.go` (webhooks)

**Core logic:** `server/core/config/` (parsing), `server/core/runtime/` (Terraform execution), `server/core/terraform/tfclient/` (TF client)

**VCS providers:** `server/events/vcs/{github,gitlab,bitbucketcloud,bitbucketserver,azuredevops,gitea}/`

**Config files:** `.golangci.yml` (lint), `go.mod`, `Makefile`, `Dockerfile`, `docker-compose.yml`, `.pre-commit-config.yaml`

## CI Workflows (.github/workflows/)

**test.yml:** `make test-all` + `make check-fmt` in `ghcr.io/runatlantis/testing-env:latest` • E2E for GitHub/GitLab (skips on forks - no secrets)
**lint.yml:** golangci-lint via GitHub Action • Path filtered (Go files only)
**website.yml:** markdownlint → lychee link check → `npm install && npm run website:build` → `npm run e2e` (Playwright)
**Others:** pr-lint (Conventional Commits), codeql, scorecard, dependency-review

**Replicate CI locally:** `make test-all && make check-fmt` OR use Docker: `docker run --rm -v $(pwd):/atlantis ghcr.io/runatlantis/testing-env:latest sh -c "cd /atlantis && make test-all"`

**E2E tests:** Complex setup (ngrok + credentials). CI handles it. Local optional. See `./scripts/e2e.sh` for details.

## Development Workflows

**Before commit:** `make test` → `make check-fmt` → `make go-generate` (if interfaces changed) → `make build-service` (verify)

**VCS provider:** Create `server/events/vcs/<provider>/` → Implement `Client` interface (`server/events/vcs/common/common.go`) → Update `server/server.go`

**Config changes:** Edit `server/core/config/valid/` or `raw/` → Update `server/user_config.go` → Test in `server/core/config/*_test.go`

**Terraform execution:** Modify `server/core/terraform/tfclient/terraform_client.go` or `server/core/runtime/*_step_runner.go` (uses `hashicorp/hc-install`)

## Known Issues

1. **golangci-lint Go 1.25+ incompatibility:** `make lint`/`make check-lint` fail. Use `make check-fmt` locally; CI handles linting.
2. **TestNewServer_GitHubUser fails:** Pre-existing in main. Ignore it.
3. **E2E tests skip on forks:** Expected (no secrets). Maintainers run them.
4. **Website needs npm install first:** Always run `npm install` before `npm run website:*` commands.
5. **docker-compose needs atlantis.env:** Create file per CONTRIBUTING.md template for local webhook testing.

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
**Website:** `npm install` • `npm run website:dev` • `npm run website:lint`
**Coverage:** `make test-coverage-html`
**Docker:** `make docker/dev` • `docker-compose up`

---

**Trust these instructions first.** Search codebase only if info is incomplete/incorrect. Validated 2026-01-30 • Go 1.25.4
