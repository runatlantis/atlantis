# Atlantis <!-- omit in toc -->

[![Latest Release](https://img.shields.io/github/release/runatlantis/atlantis.svg)](https://github.com/runatlantis/atlantis/releases/latest)
[![SuperDopeBadge](./runatlantis.io/.vitepress/public/hightower-super-dope.svg)](https://twitter.com/kelseyhightower/status/893260922222813184)
[![Go Report Card](https://goreportcard.com/badge/github.com/runatlantis/atlantis)](https://goreportcard.com/report/github.com/runatlantis/atlantis)
[![Go Reference](https://pkg.go.dev/badge/github.com/runatlantis/atlantis.svg)](https://pkg.go.dev/github.com/runatlantis/atlantis)
[![codecov](https://codecov.io/gh/runatlantis/atlantis/branch/main/graph/badge.svg)](https://codecov.io/gh/runatlantis/atlantis)
[![CircleCI](https://circleci.com/gh/runatlantis/atlantis/tree/main.svg?style=shield)](https://circleci.com/gh/runatlantis/atlantis/tree/main)
[![Slack](https://img.shields.io/badge/Join-Atlantis%20Community%20Slack-red)](https://join.slack.com/t/atlantis-community/shared_invite/zt-9xlxtxtc-CUSKB1ATt_sQy6um~LDPNw)

<p align="center">
  <img src="./runatlantis.io/public/hero.png" alt="Atlantis Logo"/><br><br>
  <b>Terraform Pull Request Automation</b>
</p>

- [Resources](#resources)
- [What is Atlantis?](#what-is-atlantis)
- [What does it do?](#what-does-it-do)
- [Why should you use it?](#why-should-you-use-it)
- [Stargazers over time](#stargazers-over-time)

## Resources
* How to get started: [www.runatlantis.io/guide](https://www.runatlantis.io/guide)
* Full documentation: [www.runatlantis.io/docs](https://www.runatlantis.io/docs)
* Download the latest release: [github.com/runatlantis/atlantis/releases/latest](https://github.com/runatlantis/atlantis/releases/latest)
* Get help in our [Slack channel](https://join.slack.com/t/atlantis-community/shared_invite/zt-9xlxtxtc-CUSKB1ATt_sQy6um~LDPNw)
* Start Contributing: [CONTRIBUTING.md](CONTRIBUTING.md)

## What is Atlantis?
A self-hosted golang application that listens for Terraform pull request events via webhooks.

## What does it do?
Runs `terraform plan`, `import`, `apply` remotely and comments back on the pull request with the output.

## Why should you use it?
* Make Terraform changes visible to your whole team.
* Enable non-operations engineers to collaborate on Terraform.
* Standardize your Terraform workflows.

## Stargazers over time

[![Stargazers over time](https://starchart.cc/runatlantis/atlantis.svg)](https://starchart.cc/runatlantis/atlantis)

## User Survey

In April 2024, the Core Atlantis Team put together an anonymous survey for Atlantis users to help us understand the community needs and prioritize our roadmap. If you are a user of Atlantis, please take 5 minutes to fill it out <https://docs.google.com/forms/d/1fOGWkdinDV2_46CZvzQRdz8401ypZR8Z-iwkNNt3EX0>
