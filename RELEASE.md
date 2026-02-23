# Releases

## Cadence

Atlantis follows a **monthly release cadence** to provide regular, predictable updates while maintaining stability for users.

### Release Schedule

-  **Frequency**: Once per month
-  **Timing**: First week OR last week of every month (but only once per month)
-  **Release Day**: Typically Tuesday or Wednesday to allow for weekend buffer

### Versioning

Atlantis follows [Semantic Versioning](https://semver.org/) (SemVer):

-  **Major releases** (x.0.0): Breaking changes
-  **Minor releases** (0.x.0): New features, backward compatible
-  **Patch releases** (0.0.x): Bug fixes and security patches

### Release Branches

-  **Main branch**: Contains the latest development work
-  **Release branches**: Created for major/minor releases (e.g., `release-0.20`)
-  **Hotfixes**: Applied to both main and relevant release branches

### Communication

-  **Release Announcements**: Posted on GitHub Releases and community channels
-  **Breaking Changes**: Clearly documented in release notes and migration guides
-  **Security Updates**: Immediately communicated through security advisories

### Release Criteria

A release is ready when:

1. ✅ All tests pass
2. ✅ Documentation is updated
3. ✅ Release notes are current
4. ✅ No known critical bugs
5. ✅ Security scan passes
6. ✅ Performance benchmarks are acceptable

### Emergency Releases

In case of critical security vulnerabilities or severe bugs:

1. **Immediate Assessment**: Evaluate severity and impact
2. **Hotfix Development**: Create targeted fix
3. **Expedited Testing**: Focused testing on the fix
4. **Emergency Release**: Release outside normal cadence if necessary

### Contributing to Releases

-  **Feature Requests**: Submit early in the month for consideration
-  **Bug Reports**: Report immediately for faster resolution
-  **Testing**: Help test release candidates
-  **Documentation**: Contribute to release notes and migration guides

For detailed information about contributing to Atlantis, see [CONTRIBUTING.md](./CONTRIBUTING.md).

## Release Process

### Creating a New Release
1. (Major/Minor release only) Create a new release branch `release-x.y`
1. Go to https://github.com/runatlantis/atlantis/releases and click "Draft a new release"
    1. Prefix version with `v` and increment based on last release.
    1. The title of the release is the same as the tag (ex. v0.2.2)
    1. Fill in description by clicking on the "Generate Release Notes" button.
        1. You may have to manually move around some commit titles as they are determined by PR labels (see .github/labeler.yml & .github/release.yml)
    1. (Latest Major/Minor branches only) Make sure the release is set as latest
        1. Don't set "latest release" for patches on older release branches.
1. Check and update the default version in `Chart.yaml` in [the official Helm chart](https://github.com/runatlantis/helm-charts/blob/main/charts/atlantis/values.yaml) as needed.

### Backporting Fixes
Atlantis now uses a [cherry-pick-bot](https://github.com/googleapis/repo-automation-bots/tree/main/packages/cherry-pick-bot) from Google. The bot assists in maintaining changes across releases branches by easily cherry-picking changes via pull requests.

Maintainers and Core Contributors can add a comment to a pull request:

```sh
/cherry-pick target-branch-name
```

target-branch-name is the branch to cherry-pick to. cherry-pick-bot will cherry-pick the merged commit to a new branch (created from the target branch) and open a new pull request to the target branch.

The bot will immediately try to cherry-pick a merged PR. On unmerged pull request, it will not do anything immediately, but wait until merge. You can comment multiple times on a PR for multiple release branches.

#### Manual Backporting Fixes
The bot will fail to cherry-pick if the feature branches' git history is not linear (merge commits instead of rebase). In that case, you will need to manually cherry-pick the squashed merged commit from main to the release branch

1. Switch to the release branch intended for the fix.
1. Run `git cherry-pick <sha>` with the commit hash from the main branch.
1. Push the newly cherry-picked commit up to the remote release branch.

### Release History

For detailed information about past releases, see:

-  [GitHub Releases](https://github.com/runatlantis/atlantis/releases)

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
```
