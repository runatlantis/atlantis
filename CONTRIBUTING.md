# Contributing

# Table of Contents

<!-- toc -->

- [Reporting Issues](#reporting-issues)
- [Reporting Security Issues](#reporting-security-issues)
- [Creating a Pull Request](#creating-a-pull-request)
  - [Resolving Comments](#resolving-comments)
- [Developing](#developing)

<!-- tocstop -->

# Reporting Issues
* When reporting issues, please include the output of `atlantis version`.
* Also include the steps required to reproduce the problem if possible and applicable. This information will help us review and fix your issue faster.
* When sending lengthy log-files, consider posting them as a gist (https://gist.github.com). Don't forget to remove sensitive data from your logfiles before posting (you can replace those parts with "REDACTED").

# Reporting Security Issues
We take security issues seriously. Please report a security vulnerability to the maintainers using [private vulnerability reporting](https://github.com/runatlantis/atlantis/security/advisories/new).

# Creating a Pull Request
* Fork the [Atlantis repo](https://github.com/runatlantis/atlantis)
* Create a new branch, commit your changes
  * Make sure to sign your commits, for example by adding `-s` when committing, see more [here](https://probot.github.io/apps/dco/).
* Create a PR
  * Make sure your title follows Conventional Commits by using a prefix like `fix:` or `feat:`, see more [here](https://www.conventionalcommits.org/en/v1.0.0/).
  * Link to any issues, including one you may have made

If you have any questions about the contribution process, see [Atlantis Contributors on Slack](https://cloud-native.slack.com/archives/C07T45G27EZ).

## Resolving Comments

It is the PR author's responsibility to review and resolve all comments on their PRs. Even if no change is made, it's still important to engage with the feedback from the community.

This also applies to comments from Copilot, see our [AI Usage Policy](AI_USAGE_POLICY.md#copilot) for more details.

# Developing

For local development setup, test commands, code style, and mock regeneration guidance, see [DEVELOPING.md](DEVELOPING.md).
