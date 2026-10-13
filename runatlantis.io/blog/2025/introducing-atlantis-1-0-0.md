---
title: Introducing Atlantis 1.0.0!
lang: en-US
---

# Atlantis 1.0.0

The Core Team is excited to announce the release of Atlantis 1.0.0! This release is many years in the making, and we're excited to be bringing Atlantis into this next chapter.

## Questions

### What should we expect from this release?

We don't expect 1.0.0 to be any different than any of our "minor" releases starting with `0`. It is primarily an indication that we believe the product is stable enough to warrant 1.0.0 release, and to show our commitment to both stability and backwards compatibility.

### Why now?

Atlantis has reached level of maturity, both as a product as well as a project, that justifies the designation that comes with a stable 1.0.0 release.

One technical issue that this release solves is that, right now, when the Atlantis team publishes a new feature release by incrementing from `0.X.Y` to `0.{X+1}.0`, it is impossible to tell if there are breaking changes in the release without reading the release notes. In a post-1.0.0 world, breaking changes will always be accompanied by a bump in the major version (see below), hence the version number will encode more meaning.

### How will we decide whether to increment major, minor, or patch for a given release?

We are roughly guided by the recommendations [here](https://semver.org/), which roughly say that bug fixes go in patch releases, feature changes go in minor releases, and backwards incompatible changes go in major releases.

As mentioned above, right now backwards incompatible changes are included together with "normal" releases. As we get experience separating these, we will develop a more clear understanding of what it means for a change to be backwards incompatible. For now here are a few guidelines:
- Changes to server settings, repo config, or the API that cause previously working setups to stop working
- Changes to core behavior like when applies and plans are run, except when gated by a new flag or setting

### When will there be an Atlantis 2.0.0?

We do not expect to need to increment the major version very often, as we value backwards compatibility. We reserve the right to push out major version changes, but we expect and hope that most development to Atlantis can be incremental and non-breaking.