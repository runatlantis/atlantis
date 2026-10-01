# Top Issues Ranking Automation Exploration

Date: 2026-06-29

## Summary Recommendation

Implement with changes.

Atlantis already asks users in both issue templates to vote with thumbs-up reactions and avoid "+1" comments. A maintained ranking issue would make that existing signal visible to maintainers and contributors. The current data looks useful: the top open Atlantis issues by thumbs-up reactions are mostly recognizable feature requests and bugs, not random noise.

Do not copy OpenTofu's implementation verbatim. Reimplement the idea for Atlantis because OpenTofu's Go file is MPL-2.0 licensed and carries OpenTofu/HashiCorp copyright headers, while Atlantis is Apache-2.0. The Atlantis version should also explicitly exclude PRs, use Atlantis labels, exclude stale/noise labels, format titles as Markdown links, include labels/comments/updated date, and make the target issue configurable.

## OpenTofu Implementation Findings

References:

- Tracking issue: https://github.com/opentofu/opentofu/issues/1496
- Workflow: https://github.com/opentofu/opentofu/blob/main/.github/workflows/update-top-issues-ranking.yml
- Script directory: https://github.com/opentofu/opentofu/tree/main/.github/scripts/update_top_issues_ranking

| Area | OpenTofu behavior |
| --- | --- |
| Trigger model | Manual workflow_dispatch plus daily schedule at 10:00 UTC. |
| Repository guard | Job-level guard: repository owner must be opentofu. |
| Runner | ubuntu-latest. |
| Permissions | contents: read and issues: write. |
| Actions | checkout and setup-go pinned by SHA with version comments. |
| Script invocation | cd into .github/scripts/update_top_issues_ranking, download modules, run main.go with owner, repo, and tracking issue number 1496. |
| Script language | Standalone Go module. |
| Dependencies | github.com/google/go-github/v61 and indirect go-querystring. |
| Auth | Reads GITHUB_TOKEN; falls back to unauthenticated client if absent. |
| Issue listing | Lists all open repository issues with pagination, 100 per page. |
| PR exclusion | No explicit PR exclusion. GitHub issue APIs can return PRs through issue endpoints, so a copied implementation could include labeled PRs. |
| Threshold | Keeps issues with at least 2 thumbs-up reactions. |
| Sorting | Sorts descending by thumbs-up count. Ties are not explicitly stabilized. |
| Label grouping | Hard-coded groups: enhancement and bug. |
| Max rows | Top 20 per group. |
| Template | Short intro plus Top Enhancements and Top Bugs sections; rows are numbered entries with bare issue URLs and reaction counts. |
| Timestamp | None. |
| Roadmap disclaimer | None beyond saying rankings are based on reactions. |
| Failure modes | Missing args can panic; invalid issue number or GitHub API errors call log.Fatal; no dry-run; no retry/backoff; no target issue validation; no explicit PR filtering. |
| License/copying | main.go is MPL-2.0 with OpenTofu and HashiCorp copyright headers. Atlantis should reimplement rather than copy. |

## Atlantis Repository Conventions

### Labels

Relevant live labels:

- feature: New functionality/enhancement.
- bug: Something is not working.
- rfc: Requirements for Comment.
- needs discussion: Large change needing community/maintainer review.
- help wanted: Good feature for contributors.
- quick-win: Obviously worth doing and under roughly 4 hours.
- regression: Bug introduced in a new version.
- question: Further information is requested.
- Exclusion candidates: Stale, duplicate, wont-do, spam, waiting-on-response.

There is no live enhancement label in the label list, although .github/release.yml references enhancement in release-note categories. Atlantis issue templates use feature.

### Issue Templates

.github/ISSUE_TEMPLATE/bug_report.md and .github/ISSUE_TEMPLATE/feature_request.md both ask users to add a thumbs-up reaction to the original issue and not leave "+1" comments. That makes this automation aligned with current project guidance.

### GitHub Actions Style

- Actions are generally pinned to full SHAs with a trailing version comment.
- Many jobs use step-security/harden-runner with egress-policy: audit.
- Newer jobs commonly use ubuntu-24.04, although some existing scheduled/security jobs still use ubuntu-latest.
- Permissions are explicit and scoped.
- Scheduled workflows already exist, including stale.yml daily, scorecard.yml weekly, and codeql.yml weekly.

### Go Tooling

- Root module: github.com/runatlantis/atlantis.
- Root go.mod uses go 1.26.4.
- Root module already depends on github.com/google/go-github/v88.
- There is no existing .github/scripts directory. A workflow-local .github/scripts/update_top_issues_ranking directory is acceptable, but it would be new.
- Prefer a standalone Go module under .github/scripts for isolation from product dependencies. Use go 1.26.4 and go-github/v88 for consistency.

## Live Atlantis Issue Signal

Data sampled with GitHub GraphQL search on 2026-06-29 using open issues sorted by thumbs-up reactions.

- Open issues overall: 710.
- Open issues excluding Stale, duplicate, wont-do, and spam: 280.

Top open issues overall after excluding stale/noise:

| Rank | Issue | Thumbs-up | Comments | Updated | Labels |
| --- | --- | ---: | ---: | --- | --- |
| 1 | [#3245 Drift Detection](https://github.com/runatlantis/atlantis/issues/3245) | 243 | 31 | 2026-05-25 | feature, help wanted, needs discussion, rfc |
| 2 | [#260 Run plan/apply for multiple projects in parallel](https://github.com/runatlantis/atlantis/issues/260) | 139 | 48 | 2024-02-21 | feature, help wanted |
| 3 | [#2172 Feature request: allow apply on merge](https://github.com/runatlantis/atlantis/issues/2172) | 127 | 46 | 2024-12-05 | feature, help wanted |
| 4 | [#2242 Error using shared Terraform providers with a large number of parallel project plans](https://github.com/runatlantis/atlantis/issues/2242) | 126 | 23 | 2025-04-17 | bug, work-in-progress |
| 5 | [#4452 SAML/SSO Authentication for Atlantis Web UI and API](https://github.com/runatlantis/atlantis/issues/4452) | 111 | 6 | 2026-06-27 | feature |
| 6 | [#4319 Metrics for PRs are never expiring](https://github.com/runatlantis/atlantis/issues/4319) | 60 | 16 | 2026-06-17 | bug, regression, never-stale |
| 7 | [#2125 gitlab: atlantis/apply job stays pending](https://github.com/runatlantis/atlantis/issues/2125) | 52 | 34 | 2026-03-18 | bug, help wanted, provider/gitlab |
| 8 | [#5415 Allow merge method as configurable option in atlantis.yaml](https://github.com/runatlantis/atlantis/issues/5415) | 47 | 3 | 2025-11-01 | feature |
| 9 | [#941 Support atlantis destroy subcommand](https://github.com/runatlantis/atlantis/issues/941) | 47 | 15 | 2026-05-06 | feature |
| 10 | [#391 Support prerequisite apply of projects for apply_requirements](https://github.com/runatlantis/atlantis/issues/391) | 44 | 13 | 2025-08-22 | feature, help wanted |

Useful observations:

- Ranking by thumbs-up is meaningful for Atlantis; the top list surfaces recurring user asks.
- Excluding Stale matters. Without exclusion, several high-vote stale items appear near the top.
- Feature ranking is strong: top feature requests have 40 to 240+ thumbs-up reactions.
- Bug ranking is also strong: top bugs have 19 to 126+ thumbs-up reactions.
- rfc has only 2 open issues, so combine it with needs discussion.
- question is comparatively noisy and often stale; do not include it in v1.

## Proposed Atlantis Design

### Tracking Issue

Maintainers should create the tracking issue manually first. The workflow/script should receive the issue number from configuration rather than hard-coding it deep in Go code.

Suggested title:

Top-Ranking Atlantis Issues 📊

Suggested initial body:

This issue is updated automatically with open Atlantis issues ranked by thumbs-up reactions.

The ranking is community signal for maintainers and contributors. It is not a roadmap commitment, priority guarantee, or acceptance decision.

Please vote by adding a thumbs-up reaction to the original issue instead of leaving "+1" comments.

### Categories And Label Mapping

| Category | Labels included | Rationale |
| --- | --- | --- |
| Top Feature Requests | feature | Matches the feature template and main user-demand surface. |
| Top Bugs | bug | Captures operational pain and correctness issues. |
| Top Regressions | regression | Small but important set; useful for triage urgency. |
| Top RFC / Needs Discussion | rfc, needs discussion | Strategic or larger design items; combined because rfc alone is sparse. |
| Top Contributor-Friendly Issues | help wanted, quick-win | Useful for contributors looking for high-impact work. |

Recommended exclusions for all categories:

- Closed issues.
- Pull requests.
- The tracking issue itself.
- Stale.
- duplicate.
- wont-do.
- spam.
- waiting-on-response.

Do not include question in v1 unless maintainers explicitly want a support-signal section.

### Ranking Parameters

| Parameter | Recommendation |
| --- | --- |
| Minimum thumbs-up threshold | 10. Atlantis has enough reaction volume that OpenTofu's threshold of 2 would include lower-signal entries. |
| Max rows per category | 15. Enough to show breadth without oversized issue bodies. |
| Sort | Thumbs-up descending, comment count descending, updated date descending, issue number ascending. |
| Row format | Markdown table with linked issue title, thumbs-up count, comments, updated date, and labels. |
| Generated timestamp | Include UTC timestamp. |
| Roadmap disclaimer | Include an explicit signal-not-roadmap note. |
| Target issue number | Workflow env var or workflow input, for example TOP_ISSUES_TRACKING_ISSUE. |
| Dry run | Add --dry-run to print generated Markdown without editing. |

### Sample Generated Markdown

Based on current live data and proposed exclusions/threshold:

## Top-Ranking Atlantis Issues 📊

Last updated: 2026-06-29 00:00 UTC.

This issue is updated automatically from open Atlantis issues ranked by thumbs-up reactions. Rankings are community signal for maintainers and contributors, not a roadmap commitment or acceptance decision.

Excluded labels: Stale, duplicate, wont-do, spam, waiting-on-response.

### Top Feature Requests

| Rank | Issue | Thumbs-up | Comments | Updated | Labels |
| --- | --- | ---: | ---: | --- | --- |
| 1 | [#3245 Drift Detection](https://github.com/runatlantis/atlantis/issues/3245) | 243 | 31 | 2026-05-25 | feature, help wanted, needs discussion, rfc |
| 2 | [#260 Run plan/apply for multiple projects in parallel](https://github.com/runatlantis/atlantis/issues/260) | 139 | 48 | 2024-02-21 | feature, help wanted |
| 3 | [#2172 Feature request: allow apply on merge](https://github.com/runatlantis/atlantis/issues/2172) | 127 | 46 | 2024-12-05 | feature, help wanted |
| 4 | [#4452 SAML/SSO Authentication for Atlantis Web UI and API](https://github.com/runatlantis/atlantis/issues/4452) | 111 | 6 | 2026-06-27 | feature |
| 5 | [#5415 Allow merge method as configurable option in atlantis.yaml](https://github.com/runatlantis/atlantis/issues/5415) | 47 | 3 | 2025-11-01 | feature |

### Top Bugs

| Rank | Issue | Thumbs-up | Comments | Updated | Labels |
| --- | --- | ---: | ---: | --- | --- |
| 1 | [#2242 Error using shared Terraform providers with a large number of parallel project plans](https://github.com/runatlantis/atlantis/issues/2242) | 126 | 23 | 2025-04-17 | bug, work-in-progress |
| 2 | [#4319 Metrics for PRs are never expiring](https://github.com/runatlantis/atlantis/issues/4319) | 60 | 16 | 2026-06-17 | bug, regression, never-stale |
| 3 | [#2125 gitlab: atlantis/apply job stays pending](https://github.com/runatlantis/atlantis/issues/2125) | 52 | 34 | 2026-03-18 | bug, help wanted, provider/gitlab |
| 4 | [#2412 Transient errors during parallel plan from .terraform.lock: plugins are not consistent](https://github.com/runatlantis/atlantis/issues/2412) | 41 | 27 | 2026-04-20 | bug, help wanted |
| 5 | [#2129 Invalid Key when attempting to stream logs over websocket](https://github.com/runatlantis/atlantis/issues/2129) | 41 | 18 | 2025-03-10 | bug, help wanted, streaming-logs |

### Top Regressions

| Rank | Issue | Thumbs-up | Comments | Updated | Labels |
| --- | --- | ---: | ---: | --- | --- |
| 1 | [#4319 Metrics for PRs are never expiring](https://github.com/runatlantis/atlantis/issues/4319) | 60 | 16 | 2026-06-17 | bug, regression, never-stale |
| 2 | [#5048 Unclearable locks with mergeability requirements after upgrade](https://github.com/runatlantis/atlantis/issues/5048) | 12 | 14 | 2025-03-13 | bug, regression |
| 3 | [#3031 post_workflow_hook errors after apply when using --automerge](https://github.com/runatlantis/atlantis/issues/3031) | 11 | 19 | 2024-09-19 | bug, regression, hook, collab, never-stale |
| 4 | [#4636 Gitlab Formatting and reaction issues in v0.28](https://github.com/runatlantis/atlantis/issues/4636) | 11 | 14 | 2024-09-06 | bug, regression |

### Top RFC / Needs Discussion

| Rank | Issue | Thumbs-up | Comments | Updated | Labels |
| --- | --- | ---: | ---: | --- | --- |
| 1 | [#3245 Drift Detection](https://github.com/runatlantis/atlantis/issues/3245) | 243 | 31 | 2026-05-25 | feature, help wanted, needs discussion, rfc |
| 2 | [#671 Run terraform workflows from local CLI command without a VCS](https://github.com/runatlantis/atlantis/issues/671) | 17 | 8 | 2023-01-17 | feature, help wanted, needs discussion |
| 3 | [#3202 More maintainers, reviewers, early adopters, and doc writers](https://github.com/runatlantis/atlantis/issues/3202) | 17 | 3 | 2025-03-30 | help wanted, needs discussion, never-stale |
| 4 | [#6158 WebUI Refactor](https://github.com/runatlantis/atlantis/issues/6158) | 11 | 19 | 2026-03-22 | feature, refactoring, needs discussion, website, breaking-change |

## Security And Permissions

Recommended workflow shape:

- Triggers: workflow_dispatch and daily schedule, for example 15 10 * * *.
- Guard: github.repository == runatlantis/atlantis.
- Top-level permissions: contents: read.
- Job permissions: issues: write.
- Token: GITHUB_TOKEN is sufficient for listing issues and editing one issue in the same repository.
- Fork safety: no pull_request trigger; repository guard prevents accidental writes from copied repos.
- Include step-security/harden-runner in audit mode to match Atlantis conventions.
- Pin checkout/setup-go actions by SHA with version comments.

The script should validate the target issue number and only edit that configured issue. It should not create issues, add labels, close issues, comment, or edit any other issue.

## Maintenance Cost

### Go Script Approach

Recommended.

Pros:

- Easy to unit test filtering, grouping, sorting, and rendering.
- Fits Atlantis' Go-heavy tooling.
- go-github is already familiar and present in the root module.
- Pagination and reaction counts are straightforward.
- Safer than complex shell/JQ once categories and exclusions grow.

Cons:

- Adds a small separate module and go.sum.
- Needs occasional dependency updates.
- Must avoid copying OpenTofu MPL-2.0 code directly.

Recommended structure:

- .github/workflows/update-top-issues-ranking.yml
- .github/scripts/update_top_issues_ranking/go.mod
- .github/scripts/update_top_issues_ranking/go.sum
- .github/scripts/update_top_issues_ranking/main.go
- .github/scripts/update_top_issues_ranking/ranking.tmpl
- .github/scripts/update_top_issues_ranking/main_test.go

### Root Module Reuse

Not recommended for v1. The root module is large and product-focused. A workflow utility should not add churn to product dependencies or root test surfaces.

### Shell / gh / API Script

Possible but not preferred. It avoids a Go module, but it is harder to unit test, more fragile around pagination/quoting, and less clear for deterministic sorting and Markdown escaping.

## Implementation Notes

Recommended script behavior:

1. Parse flags: owner, repo, issue-number, min-thumbs-up, max-per-category, dry-run.
2. Read GITHUB_TOKEN and require it in non-dry-run mode.
3. List open issues with pagination.
4. Explicitly skip pull requests.
5. Skip issues with excluded labels.
6. Skip the tracking issue itself.
7. Skip issues below threshold.
8. Group by category label mappings.
9. Sort deterministically by thumbs-up, comments, updated date, and issue number.
10. Render Markdown using text/template.
11. In dry-run mode, print to stdout.
12. In write mode, edit only the configured tracking issue.

## Exact Files Likely Added Or Changed

Likely added:

- .github/workflows/update-top-issues-ranking.yml
- .github/scripts/update_top_issues_ranking/go.mod
- .github/scripts/update_top_issues_ranking/go.sum
- .github/scripts/update_top_issues_ranking/main.go
- .github/scripts/update_top_issues_ranking/ranking.tmpl
- .github/scripts/update_top_issues_ranking/main_test.go

Optional later changes:

- .github/ISSUE_TEMPLATE/bug_report.md
- .github/ISSUE_TEMPLATE/feature_request.md

The optional template change would link to the ranking issue after maintainers create it. Do not include that in the initial implementation unless maintainers explicitly want it.

## Implementation Risks

- Reaction count is community signal, not severity. High-vote features can dominate urgent lower-vote bugs.
- Stale issues can still have high historical support. Excluding them makes rankings more actionable but may hide dormant demand.
- Label hygiene matters.
- Multi-label issues can appear in multiple categories. That is acceptable if framed as multiple views, but it may feel repetitive.
- The ranking issue itself can become a target for reactions/comments; the script should exclude it from rankings.
- GitHub issue endpoints can include PRs; the script must explicitly skip PRs.
- Copying OpenTofu code would introduce MPL-2.0 license considerations.
- Scheduled workflow permissions must remain limited to the one job that edits the configured issue.

## Validation Plan

1. Create or identify the tracking issue manually.
2. Add unit tests for PR exclusion, excluded labels, threshold filtering, deterministic sort ties, category grouping, tracking issue self-exclusion, and Markdown escaping.
3. Run script tests: cd .github/scripts/update_top_issues_ranking && go test ./...
4. Run dry-run locally against live data with GITHUB_TOKEN from gh auth token.
5. Run gofmt over the script directory.
6. Validate workflow syntax with actionlint if available.
7. Review generated Markdown manually for links, tables, excluded labels, and absence of PRs.
8. After merge, run workflow_dispatch on runatlantis/atlantis and verify only the configured tracking issue body changes.
