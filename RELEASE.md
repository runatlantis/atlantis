# Releases

## Cadence

Atlantis follows a monthly release cadence to provide regular, predictable updates while maintaining stability for users.

### Release Schedule

- **Frequency**: Once per month
- **Timing**: First week or last week of the month, but only once per month
- **Release Day**: Typically Tuesday or Wednesday to allow for weekend buffer

### Versioning

Atlantis follows [Semantic Versioning](https://semver.org/) (SemVer):

- **Major releases** (`x.0.0`): Breaking changes
- **Minor releases** (`0.x.0`): Backward-compatible new features
- **Patch releases** (`0.0.x`): Bug fixes, security patches, documentation, dependency updates, and runtime image refreshes

### Release Branches

- **Main branch**: Contains the latest development work
- **Release branches**: Created for major/minor releases, for example `release-0.44`
- **Hotfixes**: Applied to both `main` and relevant release branches when an older release line also needs the fix

### Communication

- **Release Announcements**: Posted on GitHub Releases and community channels
- **Breaking Changes**: Clearly documented in release notes and migration guides
- **Security Updates**: Communicated through security advisories when appropriate

### Release Criteria

A release is ready when:

- [ ] **Tests**: All required tests pass.
- [ ] **Documentation**: Documentation is updated for user-visible changes.
- [ ] **Release notes**: Release notes are current and reviewed.
- [ ] **Critical bugs**: No known critical bugs remain open for the release line.
- [ ] **Security**: Security scans pass.
- [ ] **Performance**: Benchmarks are acceptable for performance-sensitive changes.

### Emergency Releases

In case of critical security vulnerabilities or severe bugs:

1. **Immediate Assessment**: Evaluate severity and impact.
1. **Hotfix Development**: Create a targeted fix.
1. **Expedited Testing**: Focus validation on the affected behavior.
1. **Emergency Release**: Release outside the normal cadence if necessary.

### Contributing to Releases

- **Feature Requests**: Submit early in the month for consideration.
- **Bug Reports**: Report immediately for faster resolution.
- **Testing**: Help test release candidates.
- **Documentation**: Contribute release notes and migration guidance.

For detailed information about contributing to Atlantis, see [CONTRIBUTING.md](CONTRIBUTING.md).

## Release Process

### Prepare the Release

1. Fetch the latest default branch, release branches, and tags:

   ```sh
   git fetch origin --tags --prune \
     +refs/heads/main:refs/remotes/origin/main \
     '+refs/heads/release-*:refs/remotes/origin/release-*'
   ```

1. Confirm the latest release and verify the new tag does not already exist:

   ```sh
   gh release list --repo runatlantis/atlantis --limit 5
   git ls-remote --tags origin refs/tags/vX.Y.Z
   ```

1. Choose the release target:

   - For a major or minor release, create or update the release branch, for example `release-0.44`.
   - For a patch on the current release line, release from `main`.
   - For a patch on an older release line, release from that release branch and do not mark it as the latest release.

1. Review the commits since the previous release on the chosen target branch:

   ```sh
   git log --first-parent --reverse --oneline vPREVIOUS..origin/TARGET_BRANCH
   ```

1. Choose the version increment:

   - Use a **major** release for breaking changes.
   - Use a **minor** release for backward-compatible new user-facing features.
   - Use a **patch** release for fixes, security updates, documentation, dependency updates, and runtime image refreshes.

1. Verify GitHub Actions on the release target before publishing. At minimum, check the current `main` or `release-*` branch runs for `tester`, `website`, `CodeQL`, `atlantis-image`, and `testing-env-image`. For releases from `main`, also confirm the latest Scorecard run is passing.

### Write Release Notes

1. Start from GitHub generated notes for the chosen tag and previous tag.
1. Curate the generated notes before publishing. PR labels can put changes in noisy or incorrect sections.
1. Add a short Highlights section when the release includes important provider behavior, apply/plan safety changes, runtime image changes, security updates, or compatibility fixes.
1. Keep contributor attribution and the full changelog link from the generated notes.
1. Call out runtime image base changes explicitly, including Debian or Alpine base updates and notable pinned package updates.

Use the Highlights section for changes users should notice before scanning the full changelog. Keep it short, usually two to five bullets, and prefer concrete outcomes over PR titles. Good highlight candidates include:

- New or improved VCS hosting support, such as GitHub Enterprise Server compatibility.
- Apply, plan, policy-check, locking, or mergeability behavior changes.
- Runtime image changes, including Debian or Alpine image updates.
- Security fixes, fail-closed behavior, or dependency hardening.
- Operational fixes that reduce noisy failures, stuck plans, or incorrect commit statuses.

### Publish the GitHub Release

1. Go to [GitHub Releases](https://github.com/runatlantis/atlantis/releases) and draft a new release, or use `gh release create`.
1. Prefix the version with `v`, for example `v0.44.1`.
1. Use the tag as the release title.
1. Set the target to the chosen release target branch or commit.
1. Mark the release as latest only when it is the newest release on the current release line.
1. Publish the release.

Example CLI flow:

```sh
gh release create vX.Y.Z \
  --repo runatlantis/atlantis \
  --target TARGET_BRANCH \
  --title vX.Y.Z \
  --notes-file release-notes.md \
  --latest
```

### Verify the Release

After publishing, verify the release itself, the tag, and the workflows triggered by the tag:

```sh
gh release view vX.Y.Z --repo runatlantis/atlantis
git ls-remote --tags origin refs/tags/vX.Y.Z
gh release list --repo runatlantis/atlantis --limit 5
```

Confirm these workflows complete successfully:

- `.github/workflows/release.yml`, which runs GoReleaser and uploads release assets.
- `.github/workflows/atlantis-image.yml`, which publishes versioned and `latest` image tags for Alpine and Debian images.

### Update the Helm Chart

After publishing an Atlantis release, check the official Helm chart in `runatlantis/helm-charts`:

1. Update `charts/atlantis/Chart.yaml`.
1. Set `appVersion` to the new Atlantis release tag.
1. Bump the chart `version` for the chart release.
1. Run chart documentation and lint checks.
1. Open a separate pull request in `runatlantis/helm-charts`.

Typical chart validation:

```sh
make docs
helm lint charts/atlantis
helm template atlantis charts/atlantis
git diff --check
```

### Backporting Fixes

Atlantis uses a [cherry-pick-bot](https://github.com/googleapis/repo-automation-bots/tree/main/packages/cherry-pick-bot) from Google. The bot assists in maintaining changes across release branches by cherry-picking merged pull requests into new pull requests.

Maintainers and core contributors can add a comment to a pull request:

```sh
/cherry-pick target-branch-name
```

`target-branch-name` is the branch to cherry-pick to. The bot will cherry-pick the merged commit to a new branch created from the target branch and open a pull request.

The bot immediately tries to cherry-pick a merged pull request. On an unmerged pull request, it waits until merge. You can comment multiple times on a pull request for multiple release branches.

#### Manual Backporting Fixes

The bot can fail to cherry-pick if the feature branch history is not linear. In that case, manually cherry-pick the squashed merge commit from `main` to the release branch.

1. Switch to the release branch intended for the fix.
1. Run `git cherry-pick <sha>` with the commit hash from `main`.
1. Push the newly cherry-picked commit to the remote release branch.

### Release History

For detailed information about past releases, see [GitHub Releases](https://github.com/runatlantis/atlantis/releases).

### GPG Signing Setup

Release binaries are signed with GPG. The `checksums.txt` file gets a detached signature (`checksums.txt.sig`) that allows users to verify release integrity.

#### Generating the signing key (one-time setup)

Run the following on a secure, trusted machine. Replace the passphrase with a strong, randomly generated value:

```sh
gpg --batch --gen-key <<EOF
Key-Type: EDDSA
Key-Curve: ed25519
Subkey-Type: ECDH
Subkey-Curve: cv25519
Name-Real: Atlantis Release Signing
Name-Email: cncf-atlantis-maintainers@lists.cncf.io
Expire-Date: 2y
Passphrase: <your-secure-passphrase>
%commit
EOF
```

A 2-year expiry is recommended; it limits the impact of a key compromise while keeping maintenance reasonable. When the key expires, repeat this process, update the repository secrets, and publish the new public key. Use `gpg --quick-set-expire <fingerprint> 2y` to extend the existing key's expiry without replacing it.

#### Setting the repository secrets

After generating the key, add the following [repository secrets](https://docs.github.com/en/actions/security-guides/encrypted-secrets):

```sh
# Show the fingerprint of the newly created key
gpg --list-secret-keys --keyid-format=long

# Export the armored private key — store as the GPG_PRIVATE_KEY secret
gpg --armor --export-secret-keys <fingerprint>

# The passphrase used above becomes the GPG_PASSPHRASE secret
```

| Secret | Value |
|---|---|
| `GPG_PRIVATE_KEY` | Output of `gpg --armor --export-secret-keys <fingerprint>` |
| `GPG_PASSPHRASE` | The passphrase chosen during key generation |

The public key should be published (e.g. to a keyserver or in the repository) so users can verify signatures:

```sh
gpg --armor --export <fingerprint>
```

#### Verifying a release (for users)

```sh
# Import the Atlantis release public key
gpg --import atlantis-release-public.asc

# Verify the checksums file
gpg --verify checksums.txt.sig checksums.txt

# Then verify a downloaded binary against the checksums
sha256sum --check --ignore-missing checksums.txt
```

---

_This document is maintained by the Atlantis maintainers. For questions about the release process, please open an issue or contact the maintainers._
