# Top Issues Ranking Implementation Prototype

Date: 2026-06-29

## Design Summary

This prototype adds an Atlantis-native scheduled/manual GitHub Actions workflow and a standalone Go utility that updates one configured tracking issue with open Atlantis issues ranked by thumbs-up reactions.

The implementation is a fresh Atlantis Apache-2.0 implementation, not a copy of OpenTofu's MPL-2.0 Go source.

The script:

- lists all open issues with pagination
- explicitly excludes pull requests
- excludes the tracking issue itself
- excludes closed issues and configured noise labels
- filters by minimum thumbs-up threshold
- groups issues into Atlantis-specific categories
- sorts deterministically
- renders Markdown tables
- prints the generated Markdown in dry-run mode
- edits only the configured tracking issue in write mode

## Label / Category Mapping

| Category | Labels |
| --- | --- |
| Top Feature Requests | feature |
| Top Bugs | bug |
| Top Regressions | regression |
| Top RFC / Needs Discussion | rfc, needs discussion |
| Top Contributor-Friendly Issues | help wanted, quick-win |

Excluded labels:

- Stale
- duplicate
- wont-do
- spam
- waiting-on-response

Defaults:

- Minimum thumbs-up threshold: 10
- Max rows per category: 15

## Tracking Issue Setup Steps

Tracking issue:

- URL: <https://github.com/runatlantis/atlantis/issues/6608>
- Issue number: 6608

Repository variable:

- TOP_ISSUES_TRACKING_ISSUE is set to 6608.
- Recreate or repair it with:

      gh variable set TOP_ISSUES_TRACKING_ISSUE --repo runatlantis/atlantis --body 6608

Manual verification after the workflow is merged:

1. Open the "Update top issues ranking" workflow in GitHub Actions.
2. Run workflow_dispatch with issue_number left blank to use TOP_ISSUES_TRACKING_ISSUE, or set issue_number to 6608 for an explicit override.
3. Confirm the workflow updates <https://github.com/runatlantis/atlantis/issues/6608> and no other issue.

The workflow intentionally fails with a clear error if no tracking issue number is configured.

## Dry-Run Command

From the script directory:

    GITHUB_TOKEN=$(gh auth token) go run . --owner runatlantis --repo atlantis --issue-number 6608 --dry-run

The dry-run does not edit GitHub and prints the generated Markdown to stdout.

## Production Workflow Command / Path

Workflow:

    .github/workflows/update-top-issues-ranking.yml

Production command run by the workflow:

    cd .github/scripts/update_top_issues_ranking
    go run . --owner runatlantis --repo atlantis --issue-number "$issue_number"

## Security / Permission Notes

- The workflow has no pull_request trigger.
- The job is guarded with github.repository == 'runatlantis/atlantis'.
- Top-level permissions are contents: read.
- The update job adds issues: write because it edits the configured tracking issue.
- The script requires GITHUB_TOKEN in write mode.
- The script only calls Issues.Edit for the configured issue number.
- step-security/harden-runner runs in audit mode.
- Actions are pinned by full SHA with version comments.

## Limitations

- Reaction count is community signal, not severity or maintainer priority.
- Stale issues can still have meaningful historical demand, but are excluded by default for actionability.
- Issues with multiple labels can appear in multiple categories.
- Label hygiene affects category accuracy.
- The workflow will fail if TOP_ISSUES_TRACKING_ISSUE is removed or set to an invalid issue number.

## Follow-Up Options

- Link the tracking issue from bug and feature issue templates after maintainers create it.
- Add a separate "Recently updated high-signal issues" section.
- Make categories and excluded labels configurable through a YAML file.
- Add a workflow_dispatch dry-run mode that uploads generated Markdown as an artifact.
