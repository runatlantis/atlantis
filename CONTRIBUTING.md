# Contributing <!-- omit in toc -->

This document contains information on how to contribute to Atlantis.  If you're a maintainer, please
also see the [DEVELOPING.md](DEVELOPING.md) file for information on how to develop Atlantis locally.

# Table of Contents <!-- omit in toc -->
- [Reporting Issues](#reporting-issues)
- [Reporting Security Issues](#reporting-security-issues)
- [Updating The Website](#updating-the-website)

# Reporting Issues
* When reporting issues, please include the output of `atlantis version`.
* Also include the steps required to reproduce the problem if possible and applicable. This information will help us review and fix your issue faster.
* When sending lengthy log-files, consider posting them as a gist (https://gist.github.com). Don't forget to remove sensitive data from your logfiles before posting (you can replace those parts with "REDACTED").

# Reporting Security Issues
We take security issues seriously. Please report a security vulnerability to the maintainers using [private vulnerability reporting](https://github.com/runatlantis/atlantis/security/advisories/new).

# Updating The Website
* To view the generated website locally, run `npm website:dev` and then
open your browser to http://localhost:8080.
* The website will be regenerated when your pull request is merged to main.
