# Maintainers Guide

This guide is for Atlantis maintainers. It covers processes for handling releases, backports, and other maintainer-specific tasks.

## Table of Contents
- [Backporting Fixes](#backporting-fixes)
  - [Manual Backporting Fixes](#manual-backporting-fixes)
- [Creating a New Release](#creating-a-new-release)

## Backporting Fixes

We use the [Mergify](https://mergify.com/) bot to automatically backport changes to previous release branches.

To use Mergify for backporting, add a label to your PR in the format `backport-<branch>` where `<branch>` is the target release branch.

For example, to backport to the `release-0.27` branch, add the label `backport-release-0.27`.

### Manual Backporting Fixes

If Mergify cannot automatically backport a change, you can do it manually:

1. Checkout the target release branch:
   ```sh
   git checkout release-0.27
   git pull origin release-0.27
   ```

2. Create a new branch for the backport:
   ```sh
   git checkout -b backport/my-fix-to-release-0.27
   ```

3. Cherry-pick the commit from main:
   ```sh
   git cherry-pick <commit-hash>
   ```

4. Resolve any conflicts that arise

5. Push the branch and create a PR targeting the release branch:
   ```sh
   git push origin backport/my-fix-to-release-0.27
   ```

6. In the PR description, mention that this is a backport and link to the original PR

## Creating a New Release

### Prerequisites

- You must be a maintainer with write access to the repository
- Ensure all changes for the release are merged to the main branch
- All tests should be passing on the main branch

### Release Process

1. **Determine the version number**
   
   We follow [Semantic Versioning](https://semver.org/):
   - **MAJOR**: Incompatible API changes
   - **MINOR**: Backward-compatible functionality additions
   - **PATCH**: Backward-compatible bug fixes

2. **Update the CHANGELOG**
   
   Ensure the CHANGELOG.md is updated with all notable changes since the last release.

3. **Create a new release branch (if it's a minor or major release)**
   
   For patch releases, use the existing release branch.
   ```sh
   git checkout -b release-0.28
   git push origin release-0.28
   ```

4. **Create a new release on GitHub**
   
   - Go to https://github.com/runatlantis/atlantis/releases
   - Click "Draft a new release"
   - Choose "Choose a tag" and enter the new version (e.g., `v0.28.0`)
   - Click "Create new tag: v0.28.0 on publish"
   - Set the target to the appropriate branch (main for new releases, release branch for patches)
   - Fill in the release title: `v0.28.0`
   - Copy the relevant section from CHANGELOG.md into the release description
   - Click "Publish release"

5. **Verify the Docker images are built**
   
   Docker images are automatically built and pushed via GitHub Actions. Verify they appear at:
   https://github.com/runatlantis/atlantis/pkgs/container/atlantis

6. **Update the website documentation**
   
   If there are documentation changes, ensure they are deployed to the website.

7. **Announce the release**
   
   - Post in the #atlantis Slack channel
   - Tweet from the official Twitter account (if applicable)

### Post-Release Tasks

- Monitor for any critical issues reported after the release
- Be prepared to create a patch release if needed
- Update any external documentation or integrations

---

**Note**: This guide is a living document. If you find something unclear or missing, please submit a PR to improve it!
