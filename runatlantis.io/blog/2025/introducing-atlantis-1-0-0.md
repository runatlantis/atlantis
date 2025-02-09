---
title: Introducing Atlantis 1.0.0!
lang: en-US
---

# Atlantis 1.0.0

The core team is excited to announce the release of Atlantis 1.0.0! This release is many years in the making, and we're excited to be bringing Atlantis into this next chapter.

## Questions

### Why now?

Atlantis has reached level of maturity, both as a product as well as a project, that justifies the designation that comes with a stable 1.0.0 release.

One technical issue that this release solves is that, right now, when the Atlantis team publishes a new feature release by incrementing from `0.X.Y` to `0.{X+1}.0`, it is impossible to tell if there are breaking changes in the release without reading the release notes. In a post-1.0.0 world, breaking changes will always be accompanied by a bump in the major version (see below), hence the version number will encode more meaning.

### What should we expect from this release?

We don't expect 1.0.0 to be any different than any of our "minor" releases starting with `0`. It is primarily an indication that we believe the product is stable enough to warrant 1.0.0 release, and to show our commitment to both stability and backwards compatibility.

### Will there be do an Atlantis 2.0.0?

There are no immediate plans to release Atlantis 2.0.0, nor do we think it will never happen. We are roughly following semver guidelines described [here](https://semver.org/), which note that the major version should be incremented if "backward incompatible changes are introduced to the public API". (TODO: What does that mean for Atlantis?). If that ever happens, we reserve the right to release Atlantis 2.0.0.

### How will we decide whether to increment major, minor, or patch for a given release?

We are roughly guided by the recommendations [here](https://semver.org/), which roughly say that bug fixes go in patch releases, feature changes go in minor releases, and backwards incompatible changes go in major releases.

As mentioned above, right now backwards incompatible changes are included together with "normal" releases. As we get experience separating these, we will develop a more clear understanding of what it means for a change to be backwards incompatible. For now here are a few guidelines:
- Changes to server or repo config such that previously specified valid flag and configuration immediately fail
- Changes to behavior like when applies and plans are run, except when gated by a new flag or setting
