# Atlantis Repository - Coding Agent Instructions

## Repository Overview

**Atlantis** is a self-hosted Terraform pull request automation tool written in Go. It listens for Terraform pull request events via webhooks and automatically runs `terraform plan`, `import`, and `apply` remotely, commenting back on the pull request with the output. This enables collaborative Terraform workflows for development teams.

### Project Characteristics
- **Primary Language:** Go 1.25.4
- **Repository Size:** ~35MB with 381+ Go files
- **Type:** Server application with CLI interface
- **Frameworks:** Gorilla Mux (HTTP routing), Cobra (CLI), Viper (configuration)
- **Target Runtime:** Linux (primary), cross-platform via Docker
- **Website:** VitePress-based documentation site (Node.js/npm)
- **Testing:** Go tests, E2E tests (Playwright), integration tests with Terraform

## Build Instructions

### Prerequisites
- **Go:** 1.25.4 (as specified in go.mod; use go.mod version, not .tool-versions)
- **Node.js:** 20.x+ (for website development)
- **npm:** 10.x+
- **Docker:** Required for containerized development/testing
- **Terraform:** 1.11.1+ (for integration and E2E tests)

### Building the Binary

**Always run build commands from the repository root directory.**

```bash
# Build the Atlantis binary for Linux
make build-service
# or
make build
```

This creates the `atlantis` binary in the repository root (~51MB). Build time: 30-60 seconds on first run (downloads dependencies), ~10 seconds on subsequent builds.

**Build Details:**
- Uses `CGO_ENABLED=0 GOOS=linux GOARCH=amd64`
- Binary output: `./atlantis`
- Clean build artifacts: `make clean` (removes the atlantis binary)

### Running Tests

**Important: There is a known failing test in server/server_test.go (TestNewServer_GitHubUser). This is a pre-existing issue and should be ignored.**

```bash
# Run unit tests (excludes integration tests, ~30-60 seconds)
make test

# Run all tests including integration tests (requires Terraform, ~5 minutes)
make test-all

# Run tests in Docker (recommended for CI-like environment)
make docker/test
make docker/test-all
```

**Test Command Details:**
- `make test`: Runs `go test -short` on all packages except e2e, static, mocks, and testing
- `make test-all`: Runs `go test -timeout=300s` (includes integration tests)
- Tests require no special setup; they use mocked dependencies
- Integration tests (`make test-all`) actually execute Terraform commands

### Linting and Formatting

**Critical: golangci-lint version compatibility issue exists with Go 1.25+**

```bash
# Check code formatting (ALWAYS works, use this)
make check-fmt

# Local linting (may fail due to golangci-lint version mismatch)
make lint
# Requires: golangci-lint installed via `go install`

# CI linting (downloads correct version, may still fail with Go 1.25+)
make check-lint
```

**Known Issue:** The `make check-lint` and `make lint` commands may fail with error:
```
Error: can't load config: the Go language version (go1.24) used to build golangci-lint is lower than the targeted Go version (1.25.4)
```

**Workaround:** This is a known issue with golangci-lint not yet supporting Go 1.25+. The CI workflow uses a GitHub Action that handles this correctly. For local development, rely on `make check-fmt` and the CI pipeline for linting validation.

### Formatting Code

```bash
# Auto-format all Go files
make fmt

# Check formatting without changes
make check-fmt
```

### Mock Generation

```bash
# Regenerate all mocks (if you modify interfaces)
make go-generate

# Delete and regenerate all mocks
make regen-mocks
```

**When to regenerate mocks:**
- After modifying any interface with a `//go:generate pegomock generate` comment
- If you see compilation errors about missing mock methods
- Run `go generate <file>` on the specific file or `make go-generate` for all

## Website Development

The documentation website is built with VitePress and located in `runatlantis.io/`.

```bash
# Install dependencies (run from repo root)
npm install

# Run development server (http://localhost:8080)
npm run website:dev

# Build website
npm run website:build

# Lint markdown
npm run website:lint

# Auto-fix markdown issues
npm run website:lint-fix

# Run Playwright E2E tests for website
npm run e2e
```

**Website Files:**
- Source: `runatlantis.io/docs/*.md`
- Configuration: `runatlantis.io/.vitepress/config.js`
- Linting: `.markdownlint.yaml`, `.lycheeignore`

## Project Layout and Architecture

### Root Directory Structure
```
.github/            # GitHub workflows, issue templates, configs
cmd/                # CLI command implementations (cobra commands)
e2e/                # End-to-end test suite (separate Go module)
runatlantis.io/     # VitePress documentation website
scripts/            # Build and utility scripts
server/             # Main application server code
  controllers/      # HTTP request handlers
  core/             # Core business logic
    config/         # Configuration parsing and validation
    runtime/        # Terraform/OpenTofu runtime execution
    terraform/      # Terraform client and helpers
    locking/        # Lock management
    db/             # Database interfaces
    boltdb/         # BoltDB implementation
    redis/          # Redis implementation
  events/           # Event processing (plan, apply, etc.)
    vcs/            # VCS provider integrations (GitHub, GitLab, etc.)
  logging/          # Structured logging
  metrics/          # Metrics collection
  jobs/             # Background job management
testdata/           # Test fixtures and golden files
testdrive/          # Interactive test drive command
testing/            # Shared testing utilities
main.go             # Application entry point
```

### Key Files and Their Purposes

**Configuration Files:**
- `.golangci.yml` - golangci-lint configuration (linter rules)
- `.pre-commit-config.yaml` - Pre-commit hooks (gitleaks, shellcheck, etc.)
- `go.mod` / `go.sum` - Go module dependencies
- `package.json` / `package-lock.json` - npm dependencies (website)
- `Makefile` - Build automation
- `Dockerfile` - Production Docker image
- `Dockerfile.dev` - Development Docker image
- `docker-compose.yml` - Local development with ngrok
- `.tool-versions` - asdf version manager (reference only)

**Entry Points:**
- `main.go` - Application entry point, sets up commands
- `cmd/server.go` - Server command implementation
- `cmd/version.go` - Version command
- `cmd/testdrive.go` - Interactive testdrive command

**Core Server:**
- `server/server.go` - Main server initialization and setup
- `server/router.go` - HTTP route configuration
- `server/controllers/events/events_controller.go` - Webhook event handling

### VCS Provider Support
Atlantis supports multiple VCS providers with separate client implementations:
- GitHub: `server/events/vcs/github/`
- GitLab: `server/events/vcs/gitlab/`
- Bitbucket Cloud: `server/events/vcs/bitbucketcloud/`
- Bitbucket Server: `server/events/vcs/bitbucketserver/`
- Azure DevOps: `server/events/vcs/azuredevops/`
- Gitea: `server/events/vcs/gitea/`

## Continuous Integration and Validation

### GitHub Workflows (`.github/workflows/`)

**Primary CI Workflows:**

1. **`test.yml`** - Main test pipeline
   - Triggers: Push to main/release branches, PRs
   - Steps: `make test-all`, `make check-fmt`
   - Runs in Docker container: `ghcr.io/runatlantis/testing-env:latest`
   - Uses path filtering (only runs for Go file changes)
   - Includes E2E tests for GitHub and GitLab (requires secrets)
   - **Expected failures:** Skip E2E tests on forked PRs (missing secrets)

2. **`lint.yml`** - Linting pipeline
   - Triggers: PRs to main/release branches
   - Steps: golangci-lint via GitHub Action
   - Uses path filtering (only runs for Go file changes)
   - Uses golangci-lint v2.7.2

3. **`website.yml`** - Documentation website validation
   - Triggers: PRs affecting `runatlantis.io/`, `package*.json`
   - Steps:
     - `markdownlint` - Markdown linting
     - `lychee` - Link checking
     - `npm install && npm run website:build` - Build verification
     - `npm run e2e` - Playwright E2E tests
   - Uses path filtering

4. **`pr-lint.yml`** - PR title validation (Conventional Commits)
5. **`codeql.yml`** - CodeQL security analysis
6. **`scorecard.yml`** - OpenSSF Scorecard
7. **`dependency-review.yml`** - Dependency security check

### Pre-commit Hooks (`.pre-commit-config.yaml`)
- gitleaks (secret scanning)
- golangci-lint
- shellcheck
- eslint
- end-of-file-fixer
- trailing-whitespace

**These are optional locally but recommended.** Install with: `pre-commit install`

### Replicating CI Locally

```bash
# 1. Run the same tests as CI
make test-all
make check-fmt

# 2. For exact CI environment, use Docker
docker run --rm -v $(pwd):/atlantis \
  ghcr.io/runatlantis/testing-env:latest \
  sh -c "cd /atlantis && make test-all && make check-fmt"

# 3. Build and test with docker-compose (requires atlantis.env setup)
make build-service
docker-compose up --detach
docker-compose logs --follow

# 4. Validate website changes
npm install
npm run website:build
npx playwright install --with-deps
npm run e2e
```

### E2E Tests

E2E tests are located in `e2e/` directory and are a separate Go module.

**Running E2E tests requires:**
1. Built atlantis binary: `make build-service`
2. ngrok installed and configured with auth token
3. GitHub/GitLab credentials set as environment variables
4. Script execution: `./scripts/e2e.sh`

**Environment variables needed:**
- `ATLANTIS_GH_USER` / `ATLANTIS_GH_TOKEN` (for GitHub)
- `ATLANTIS_GITLAB_USER` / `ATLANTIS_GITLAB_TOKEN` (for GitLab)
- `NGROK_AUTH_TOKEN`

**The CI workflow handles E2E tests automatically.** Local E2E testing is optional and complex to set up.

## Common Development Workflows

### Making Code Changes

1. **Always** run `make test` before committing
2. **Always** run `make check-fmt` before committing
3. Regenerate mocks if interfaces changed: `make go-generate`
4. If adding dependencies, update `go.mod`: `go get <package>`
5. Build to verify: `make build-service`

### Adding a New VCS Provider

1. Create directory: `server/events/vcs/<provider>/`
2. Implement `Client` interface from `server/events/vcs/common/common.go`
3. Add tests
4. Update server initialization in `server/server.go`
5. Regenerate mocks if needed

### Modifying Configuration Parsing

- Config structures: `server/core/config/valid/` and `server/core/config/raw/`
- Validation logic: `server/core/config/valid/global_cfg.go`
- User config: `server/user_config.go`
- Tests: `server/core/config/*_test.go`

### Working with Terraform Execution

- Terraform client: `server/core/terraform/tfclient/terraform_client.go`
- Runtime execution: `server/core/runtime/`
- Step runners: `server/core/runtime/*_step_runner.go`
- Version management: Uses `hashicorp/hc-install` for downloading TF/OpenTofu

## Known Issues and Workarounds

### 1. golangci-lint Version Incompatibility
**Issue:** golangci-lint v1.64.4 built with Go 1.24 cannot lint code targeting Go 1.25.4
**Workaround:** Use `make check-fmt` locally. CI workflow handles linting correctly via GitHub Action.
**When this matters:** Any linting during local development or in custom CI scripts.

### 2. Failing Unit Test: TestNewServer_GitHubUser
**Issue:** Pre-existing nil pointer dereference in `server/server_test.go:TestNewServer_GitHubUser`
**Status:** Known issue, exists in main branch
**Workaround:** Ignore this test failure. It's unrelated to most code changes.
**When this matters:** Running `make test` or `make test-all` will show 1 failure.

### 3. E2E Tests on Forked PRs
**Issue:** E2E tests (`e2e-github`, `e2e-gitlab` in test.yml) skip on forked PRs due to missing secrets
**Status:** Expected behavior for security reasons
**Workaround:** None needed. Maintainers run E2E tests on their branches.

### 4. Website Build Requires npm install
**Issue:** Running `npm run website:build` or `npm run website:dev` fails without `npm install`
**Workaround:** **Always** run `npm install` first if package.json or package-lock.json changed.
**When this matters:** Any website-related changes.

### 5. Docker Compose Development Setup
**Issue:** `docker-compose up` requires `atlantis.env` file with credentials
**Workaround:** Create `atlantis.env` file following the template in CONTRIBUTING.md
**When this matters:** Local development with webhook testing via ngrok.

## Code Style and Conventions

### Logging (from CONTRIBUTING.md)
- Use `ctx.Log` (available in most methods, pass it down if not)
- Levels: debug (developers), info (users), warn (maybe problem), error (definitely problem)
- **ALWAYS** lowercase (first letter auto-capitalized)
- **ALWAYS** quote string variables with `%q` in format strings
- **NEVER** use colons (`:`) in logs (reserved for error chains)

### Errors (from CONTRIBUTING.md)
- **ALWAYS** lowercase unless required
- **ALWAYS** use `fmt.Errorf("context: %w", err)` not `%s`
- **NEVER** use "error occurred", "failed to", "unable to" - describe the action instead
- Example: "setting up workspace: running git clone: no executable \"git\""

### Testing (from CONTRIBUTING.md)
- Place tests under `{package}_test` to test external interfaces
- For internal testing: `{file}_internal_test.go`
- Use testing utilities: `import . "github.com/runatlantis/atlantis/testing"`
- Use `Assert()`, `Equals()`, `Ok()` for assertions

### Commit Messages
- Must use Conventional Commits format: `fix:`, `feat:`, `docs:`, etc.
- Sign commits with `-s` flag (DCO requirement)

## Validation Checklist

Before submitting a PR, ensure:

- [ ] `make test-all` passes (ignore TestNewServer_GitHubUser failure)
- [ ] `make check-fmt` passes
- [ ] Mocks regenerated if interfaces modified (`make go-generate`)
- [ ] Website builds if docs changed (`npm install && npm run website:build`)
- [ ] Commit messages follow Conventional Commits
- [ ] Commits are signed (DCO)
- [ ] PR title uses correct prefix (`fix:`, `feat:`, etc.)
- [ ] Code follows logging and error conventions
- [ ] Tests added for new functionality
- [ ] Documentation updated if behavior changed

## Additional Commands Reference

```bash
# Coverage
make test-coverage         # Generate coverage report
make test-coverage-html    # Generate HTML coverage report

# Docker development
make docker/dev            # Build dev Docker image
make build-service         # Build binary for docker-compose

# Release (maintainers only)
make release              # Create release packages

# E2E dependencies (for E2E testing setup)
make end-to-end-deps      # Install E2E test dependencies
make end-to-end-tests     # Run E2E tests (requires full setup)
```

## Quick Reference: Most Common Commands

```bash
# Start new work
make build-service        # Build the binary
make test                 # Run unit tests
make check-fmt            # Check formatting

# Before committing
make test-all             # Run all tests including integration
make check-fmt            # Verify formatting

# Website work
npm install               # Install dependencies
npm run website:dev       # Local dev server
npm run website:lint      # Check markdown
```

## Trust These Instructions

These instructions are comprehensive and validated against the current repository state. When working on this codebase:
- **Trust** these build and test commands as the correct sequence
- **Refer** to these instructions before searching the codebase for build info
- **Update** these instructions if you discover errors or missing information
- **Search** the codebase only if information here is incomplete or found to be incorrect

---

*Last validated: 2026-01-30*
*Repository: runatlantis/atlantis*
*Go version: 1.25.4*
