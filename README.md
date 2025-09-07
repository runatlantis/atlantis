
<div align="right">
  <details>
    <summary >🌐 Language</summary>
    <div>
      <div align="center">
        <a href="https://openaitx.github.io/view.html?user=runatlantis&project=atlantis&lang=en">English</a>
        | <a href="https://openaitx.github.io/view.html?user=runatlantis&project=atlantis&lang=zh-CN">简体中文</a>
        | <a href="https://openaitx.github.io/view.html?user=runatlantis&project=atlantis&lang=zh-TW">繁體中文</a>
        | <a href="https://openaitx.github.io/view.html?user=runatlantis&project=atlantis&lang=ja">日本語</a>
        | <a href="https://openaitx.github.io/view.html?user=runatlantis&project=atlantis&lang=ko">한국어</a>
        | <a href="https://openaitx.github.io/view.html?user=runatlantis&project=atlantis&lang=hi">हिन्दी</a>
        | <a href="https://openaitx.github.io/view.html?user=runatlantis&project=atlantis&lang=th">ไทย</a>
        | <a href="https://openaitx.github.io/view.html?user=runatlantis&project=atlantis&lang=fr">Français</a>
        | <a href="https://openaitx.github.io/view.html?user=runatlantis&project=atlantis&lang=de">Deutsch</a>
        | <a href="https://openaitx.github.io/view.html?user=runatlantis&project=atlantis&lang=es">Español</a>
        | <a href="https://openaitx.github.io/view.html?user=runatlantis&project=atlantis&lang=it">Italiano</a>
        | <a href="https://openaitx.github.io/view.html?user=runatlantis&project=atlantis&lang=ru">Русский</a>
        | <a href="https://openaitx.github.io/view.html?user=runatlantis&project=atlantis&lang=pt">Português</a>
        | <a href="https://openaitx.github.io/view.html?user=runatlantis&project=atlantis&lang=nl">Nederlands</a>
        | <a href="https://openaitx.github.io/view.html?user=runatlantis&project=atlantis&lang=pl">Polski</a>
        | <a href="https://openaitx.github.io/view.html?user=runatlantis&project=atlantis&lang=ar">العربية</a>
        | <a href="https://openaitx.github.io/view.html?user=runatlantis&project=atlantis&lang=fa">فارسی</a>
        | <a href="https://openaitx.github.io/view.html?user=runatlantis&project=atlantis&lang=tr">Türkçe</a>
        | <a href="https://openaitx.github.io/view.html?user=runatlantis&project=atlantis&lang=vi">Tiếng Việt</a>
        | <a href="https://openaitx.github.io/view.html?user=runatlantis&project=atlantis&lang=id">Bahasa Indonesia</a>
        | <a href="https://openaitx.github.io/view.html?user=runatlantis&project=atlantis&lang=as">অসমীয়া</
      </div>
    </div>
  </details>
</div>

# Atlantis <!-- omit in toc -->

[![Latest Release](https://img.shields.io/github/release/runatlantis/atlantis.svg)](https://github.com/runatlantis/atlantis/releases/latest)
[![SuperDopeBadge](./runatlantis.io/public/hightower-super-dope.svg)](https://twitter.com/kelseyhightower/status/893260922222813184)
[![Go Report Card](https://goreportcard.com/badge/github.com/runatlantis/atlantis)](https://goreportcard.com/report/github.com/runatlantis/atlantis)
[![Go Reference](https://pkg.go.dev/badge/github.com/runatlantis/atlantis.svg)](https://pkg.go.dev/github.com/runatlantis/atlantis)
[![Slack](https://img.shields.io/badge/Join-Atlantis%20Community%20Slack-red)](https://slack.cncf.io/)
[![OpenSSF Scorecard](https://api.scorecard.dev/projects/github.com/runatlantis/atlantis/badge)](https://scorecard.dev/viewer/?uri=github.com/runatlantis/atlantis)
[![OpenSSF Best Practices](https://www.bestpractices.dev/projects/9428/badge)](https://www.bestpractices.dev/projects/9428)

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
* Get help in our [Slack channel](https://slack.cncf.io/) in channel #atlantis and development in #atlantis-contributors
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
