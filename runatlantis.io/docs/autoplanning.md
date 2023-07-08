# Autoplanning
On any **new** pull request or **new commit** to an existing pull request, Atlantis will attempt to
run `terraform plan` in the directories it thinks hold modified Terraform projects.

The algorithm it uses is as follows:
1. Get list of all modified files in pull request
1. Filter to those containing `.tf`
1. Get the directories that those files are in
1. If the directory path doesn't contain `modules/` then try to run `plan` in that directory
1. If it does contain `modules/` look at the directory one level above `modules/`. If it
contains a `main.tf` run plan in that directory, otherwise ignore the change (see below for exceptions).

## Example
Given the directory structure:
```
.
├── modules
│   └── module1
│       └── main.tf
└── project1
    ├── main.tf
    └── modules
        └── module1
            └── main.tf
```

* If `project1/main.tf` were modified, we would run `plan` in `project1`
* If `modules/module1/main.tf` were modified, we would not automatically run `plan` because we couldn't determine the location of the terraform project
    * You could use an [atlantis.yaml](repo-level-atlantis-yaml.html#configuring-planning) file to specify which projects to plan when this module changed
    * You could enable [module autoplanning](server-configuration.html#autoplan-modules) which indexes projects to their local module dependencies.
    * Or you could manually plan with `atlantis plan -d <dir>`
* If `project1/modules/module1/main.tf` were modified, we would look one level above `project1/modules`
into `project1/`, see that there was a `main.tf` file and so run plan in `project1/`

## Bitbucket-Specific Notes
Bitbucket does not have a webhook that triggers only upon a new PR or commit. To fix this we cache the last commit to see if it has changed. If the cache is emptied, Atlantis will think your commit is new and you may see extra plans.
This scenario can happen if:
* Atlantis restarts
* You are running multiple Atlantis instances behind a load balancer

## Customizing
If you would like to customize how Atlantis determines which directory to run in
or disable it all together you need to create an `atlantis.yaml` file.
See
* [Disabling Autoplanning](repo-level-atlantis-yaml.html#disabling-autoplanning)
* [Configuring Planning](repo-level-atlantis-yaml.html#configuring-planning)
