# Atlantis

<p align="center">
  <img src="./runatlantis.io/.vuepress/public/hero.png" alt="Atlantis Logo"/><br><br>
  <b>Terraform Pull Request Automation</b>
</p>

## Azure Devops Support (alpha)

This fork contains Azure Devops support using a fork of the [go-azuredevops](https://github.com/mcdafydd/go-azuredevops/tree/webhook-support) library.  It will build and is somewhat functional now, but beware that work is still ongoing to fix panics, add more tests, fix work item comment support, and update documentation.

## Resources
* How to get started: [www.runatlantis.io/guide](https://www.runatlantis.io/guide)
* Full documentation: [www.runatlantis.io/docs](https://www.runatlantis.io/docs)
* Download the latest release: [github.com/runatlantis/atlantis/releases/latest](https://github.com/runatlantis/atlantis/releases/latest)
* Get help in our [Slack channel](https://thawing-headland-22460.herokuapp.com) or our [Gitter channel](https://gitter.im/runatlantis/Lobby)
* Start Contributing: [CONTRIBUTING.md](CONTRIBUTING.md)

## What is Atlantis?
A self-hosted golang application that listens for Terraform pull request events via webhooks.

## What does it do?
Runs `terraform plan` and `apply` remotely and comments back on the pull request with the output.

## Why should you use it?
* Make Terraform changes visible to your whole team.
* Enable non-operations engineers to collaborate on Terraform.
* Standardize your Terraform workflows.

## Badges!
[![SuperDopeBadge](./runatlantis.io/.vuepress/public/hightower-super-dope.svg)](https://twitter.com/kelseyhightower/status/893260922222813184)
[![Go Report Card](https://goreportcard.com/badge/github.com/runatlantis/atlantis)](https://goreportcard.com/report/github.com/runatlantis/atlantis)
[![codecov](https://codecov.io/gh/runatlantis/atlantis/branch/master/graph/badge.svg)](https://codecov.io/gh/runatlantis/atlantis)
[![CircleCI](https://circleci.com/gh/runatlantis/atlantis/tree/master.svg?style=shield)](https://circleci.com/gh/runatlantis/atlantis/tree/master)
[![Slack Status](https://thawing-headland-22460.herokuapp.com/badge.svg)](https://thawing-headland-22460.herokuapp.com)
[![Gitter chat](https://badges.gitter.im/runatlantis.png)](https://gitter.im/runatlantis)
