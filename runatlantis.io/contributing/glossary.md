# Glossary

The Atlantis community uses many words and phrases to work more efficiently.
You will find the most common ones and their meaning on this page.

## Pull / Merge Request Event

The different VCSs have different names for merging changes. Atlantis uses the
name Pull Request as the abstraction. The VCS provider implements this
abstraction and forwards the call to the respective function.

## VCS

VCS stands for Version Control System.

Atlantis supports only git as a Version Control System. However, there is
support for multiple VCS Providers. Currently, it supports the following
providers:

- [Azure DevOps](https://azure.microsoft.com/en-us/products/devops)
- [BitBucket](https://bitbucket.org/)
- [GitHub](https://github.com/)
- [GitLab](https://gitlab.com/)
- [Gitea](https://gitea.com/)

The term VCS is used for both git and the different VCS providers.
