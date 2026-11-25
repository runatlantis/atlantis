# 6. Group related projects

Date: 2026-11-24

## Status

Proposed

Related issue: [#7138](https://github.com/runatlantis/atlantis/issues/7138).

## Context

A repository with several production projects has to name them one by one or run the command across the repository. It also has to repeat any workflow they share. Numeric execution groups can put staging before production, but the numbers give no clue what runs in each stage.

## Decision

A repository can put related projects in a named group, select them together, and set defaults for them. Settings on a project take precedence over defaults from its group. Projects and groups can also name what should run before them. Operators can define reusable groups in `repos.yaml` alongside `workflows`. Each repository can define its own groups in `atlantis.yaml` and use the server groups listed in `allowed_project_groups`. Server group names are reserved. A project can belong to at most one group. Group names and the names of projects that apply to the PR's target branch share a namespace. Atlantis rejects duplicates in that set.

### Configuration

Here, the server defines `staging` and the repository defines `prod`. A workflow chosen in repo config follows the existing workflow restrictions, whether a project or its group chooses it.

The YAML examples illustrate the intended behavior. Field names and layout may change during implementation.

```yaml
# Server-side repos.yaml
repos:
  - id: github.com/acme/payments-infra
    allowed_project_groups: [staging]
    allowed_overrides: [workflow]
    allowed_workflows: [guarded]

workflows:
  guarded:
    plan:
      steps: [init, plan]
    apply:
      steps:
        - run: /etc/atlantis/check-change-window
        - apply

project_groups:
  - name: staging
```

```yaml
# Repository atlantis.yaml
version: 3
project_groups:
  - name: prod
    workflow: guarded
    depends_on: [staging]

projects:
  - name: staging-api
    dir: terraform/staging/api
    group: staging
  - name: prod-database
    dir: terraform/prod/database
    group: prod
  - name: prod-api
    dir: terraform/prod/api
    group: prod
    run_after: [prod-database]
```

Atlantis evaluates autoplan for each project. `atlantis plan --group prod` explicitly selects all projects in `prod`. `atlantis apply --group prod` selects the pending production plans. Both production projects inherit the server's `guarded` workflow from their group. If the PR affects `staging-api`, that project must have applied or have a current plan with no changes. `prod-api` waits for the database when both are selected. A later apply that selects only `prod-api` still checks staging and can proceed without selecting the database.

### Order and dependencies

| Setting | What it does |
| --- | --- |
| `run_after` | Waits for the named work to finish when it's part of the same command. The next project can still run after a failure. |
| `depends_on` | Orders work within a command and skips projects whose prerequisites fail. Before apply, each prerequisite affected by the PR must have applied or have a current plan with no changes, even if it wasn't selected. |

Projects can name other projects or groups. Groups can name other groups. Atlantis matches dependency names exactly and rejects unknown names. It expands each group to its projects and rejects cycles. It reports unmet prerequisites without adding work to the command. Independent projects can run in parallel. With `abort_on_failure`, any failure skips work that hasn't started. Locks and results stay per project.

## Transition

Project groups and named ordering ship alongside numeric ordering. The same release deprecates `execution_order_group`. Repositories can keep their numeric ordering and `abort_on_execution_order_fail` during the deprecation period. A repository uses one ordering model at a time. Existing project-to-project `depends_on` works with either. Removal follows Atlantis's deprecation policy.

## Consequences

Operators can run a set of projects by name and set shared defaults once. Atlantis has to validate dependencies across projects and groups, and users still get a result for each project.

## Alternatives considered

- Keep numeric ordering and use groups for selection and workflow defaults. Repositories could avoid an ordering migration, and operators would still use numbers to say what runs first within a command.
- Use `depends_on` for all named ordering. A targeted apply could be blocked by a prerequisite that previously only affected the order of a batch apply.

## References

- [Atlantis repository configuration](https://www.runatlantis.io/docs/repo-level-atlantis-yaml)
