---
layout: HomeCustom
pageClass: home-custom
heroImage: /hero.png
heroText: Atlantis
actionText: Get Started →
actionLink: /guide/
---

## How it works
* You deploy Atlantis internally. You don't have to give your cloud credentials to a third party.
    * It runs as a golang binary or Docker container.
* Expose it with a URL that is accessible by github/gitlab.com or your private git host.
* Add its URL to your GitHub or GitLab repository so it can receive webhooks.
* When a Terraform pull request is opened, Atlantis will run `terraform plan` and comment
with the output back to the pull request.
    * The exact `terraform plan` command is configurable.
* If the `plan` looks good, users can comment on the pull request `atlantis apply` to apply the plan.
    * You can require pull request approval before running `apply` is allowed.
