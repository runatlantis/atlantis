# Release Cadence

## Overview

Atlantis follows a **monthly release cadence** to provide regular, predictable updates while maintaining stability for users.

## Release Schedule

-  **Frequency**: Once per month
-  **Timing**: First week OR last week of every month (but only once per month)
-  **Release Day**: Typically Tuesday or Wednesday to allow for weekend buffer

## Versioning

Atlantis follows [Semantic Versioning](https://semver.org/) (SemVer):

-  **Major releases** (x.0.0): Breaking changes
-  **Minor releases** (0.x.0): New features, backward compatible
-  **Patch releases** (0.0.x): Bug fixes and security patches

## Release Branches

-  **Main branch**: Contains the latest development work
-  **Release branches**: Created for major/minor releases (e.g., `release-0.20`)
-  **Hotfixes**: Applied to both main and relevant release branches

## Communication

-  **Release Announcements**: Posted on GitHub Releases and community channels
-  **Breaking Changes**: Clearly documented in release notes and migration guides
-  **Security Updates**: Immediately communicated through security advisories

## Release Criteria

A release is ready when:

1. ✅ All tests pass
2. ✅ Documentation is updated
3. ✅ Release notes are current
4. ✅ No known critical bugs
5. ✅ Security scan passes
6. ✅ Performance benchmarks are acceptable

## Emergency Releases

In case of critical security vulnerabilities or severe bugs:

1. **Immediate Assessment**: Evaluate severity and impact
2. **Hotfix Development**: Create targeted fix
3. **Expedited Testing**: Focused testing on the fix
4. **Emergency Release**: Release outside normal cadence if necessary

## Contributing to Releases

-  **Feature Requests**: Submit early in the month for consideration
-  **Bug Reports**: Report immediately for faster resolution
-  **Testing**: Help test release candidates
-  **Documentation**: Contribute to release notes and migration guides

For detailed information about contributing to Atlantis, see [CONTRIBUTING.md](./CONTRIBUTING.md).

## Release History

For detailed information about past releases, see:

-  [GitHub Releases](https://github.com/runatlantis/atlantis/releases)

---

_This document is maintained by the Atlantis maintainers. For questions about the release process, please open an issue or contact the maintainers._
