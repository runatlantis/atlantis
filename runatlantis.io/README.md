---
layout: HomeCustom
pageClass: home-custom
heroImage: /hero.png
heroText: Atlantis
actionText: Get Started →
actionLink: /guide/
---

## How it works
* Deploy Atlantis internally. You don't have to give your cloud credentials to a third party.
    * It runs as a golang binary or Docker container.
* Add its URL to your GitHub or GitLab repository so it can receive webhooks.
* When a Terraform pull request is opened, Atlantis runs `terraform plan` and comments
with the output back to the pull request.
    * The exact `terraform plan` command is configurable.
    * This directory and Terraform workspace within the repository are now "Locked".
    Other pull requests cannot `plan` against the same directory/workspace until the plan
    is applied or deleted and the pull request is merged.
* If the `plan` looks good, users can comment on the pull request `atlantis apply` to apply the plan.
    * You can require pull request approval before running `apply` is allowed.
