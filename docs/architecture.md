plan file names
- named as {owner}_{repo}_{pull number}
- saved to s3 at {bucket}/{prefix}/{owner}/{repo}/{owner}_{repo}_{pull}{_optional_relative_path}.tfplan{.optional env}, bucket=hootsuite-terraform, prefix=plans
  - planfilename = ".tfplan.{env}" if at root, otherwise "_relative_plan_path.tfplan.{env}". If no env then no .{env} suffix
- the trick with the plan names is when we're running an apply, we pull out all plans that match: plans/{owner}/{repo}/{owner}_{repo}_{pullNum}
  - now we still need to know where to run the plan from, which is why we include the path as part of the filename, ex. ...{pullNum}_{relative}_{path}.tfplan.{env}.
  - we *could* figure out the plan paths from the pull request, however we would still need to know which plan file is which, hence needing to have the relative path as part of the plan

results:
  - apply/plan gives either a "Terraform Apply:" or a "Terraform Plan:" prefix and the same for each path "Terraform plan for `{path}`"
  - either a setup error/failure (basically anything pre-path fork)
  - once it gets to the path phase, each path can have a success/error/failure
  - so what needs a template?
    - pre-path-execution setup failures/successes with a "{Command} {Statused}" prefix, ex. **Apply Failed**:
    - path-execution results

naming:
  - `Atlantis` has (currently) three `commands`: `plan`, `apply`, and `help`
  - `Terraform` has `commands` like `plan`, `apply`, `get`, `remote`
  - When a user comments with an `Atlantis` `command`, they trigger an `execution`
    - We determine the `Atlantis` `command` (ex. `plan` or `apply`)
      - We also determine the `context` of the `execution`, or `executionContext`. This consists of data like the `repo owner`, the `pull request number`, the `commentor username`, etc.
    - From the `command` type, we know which `executor` to use. Either the `PlanExecutor`, or the `ApplyExecutor`
    - We then tell the `executor` to `execute` the `command`
    - There may be multiple `terraform projects` involved in an `execution` because we may have to run `terraform plan/apply` in multiple `paths`.
    - The `executor` returns to us the `executionResult`
      - each `result` has one of three `statuses`: `successful`, `failed`, `errored`
        - `successful` is when... it was successful
        - `failed` is when it didn't work but because of a user-solvable problem, like their terraform was wrong, or that environment doesn't exist
        - `errored` is when there was an internal Atlantis error that couldn't be fixed by the user, like a request to github timed out
      - `result` has sub-types to represent different results. For example there is a `PlanResult` that has the output of `terraform plan`, and the url to discard the plan and lock
    - We then need to add a `comment` back to the `pull request`, so we `render` the `executionResult` using a `renderer`. In this case, we need the `GithubCommentRenderer`
      - We can then add the `comment` to the `pull request` by using the `githubClient`

components:
  clients for external services:
    - github (created)
    - s3 (created)
  clients for on-host "services":
    - git
    - aws cli
    - terraform
  commands:
    - plan
    - apply
    - help

logging guidelines:
  - all lowercase (first letter of each line will be autocapitalized)
  - levels:
    - debug is for developers of atlantis
    - info is for users (expected that people run on info level)
    - warn is for something that might be a problem but we're not sure
    - error is for something that's definitely a problem
  - don't log any output or multiple lines unless at debug level
  - quote any string variables using %q in the fmt string, ex. `ctx.Log.Info("cleaning clone dir %q", dir)` => `Cleaning clone directory "/tmp/atlantis/lkysow/atlantis-terraform-test/3"`
  - if something is an error, ex. we couldn't clean up the workspace, use the words "failed to" in the log
  - never use colons "`:`" in a log since that's used to separate error descriptions and causes
    - if you need to have a break in your comment, either use `-` or `,` ex. `failed to clean directory, continuing regardless` or `POST /404 - Response code 404`

Glossary

* **Run**: Encompasses the two steps (plan and apply) for modifying infrastructure in a specific environment
* **Run Lock**: When a run has started but is not yet completed, the infrastructure and environment that's being modified is "locked" against
other runs being started for the same set of infrastructure and environment. We determine what infrastructure is being modified by combining the
repository name, the directory in the repository at which the terraform commands need to be run, and the environment that's being modified
* **Run Path**: The path relative to the repository's root at which terraform commands need to be executed for this Run
* **Run Key**: The unique id for the set of infrastructure that is being modified in a Run. It is a combination of the repository name, run path, and environment
* **Run Id**: The id for this specific Run
