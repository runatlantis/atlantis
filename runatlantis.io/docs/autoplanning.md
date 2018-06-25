# Autoplanning
On any **new** pull request or **new commit** to an existing pull request, Atlantis will attempt to
run `terraform plan` in the directories it thinks hold modified Terraform projects.

The algorithm it uses is as follows:
1. Get list of all modified files in pull request
1. Filter to those containing `.tf`
1. Get the directories that those files are in
1. If the directory path doesn't contain `modules/` then try to run `plan` in that directory
1. If it does contain `modules/` look at the directory one level above `modules/`. If it
contains a `main.tf` run plan in that directory, otherwise ignore the change.

todo: add example

If you would like to configure how Atlantis determines which directory to run in
or disable it all together you need to create an `atlantis.yaml` file.
See
* Disabling Autoplanning
* Configur
