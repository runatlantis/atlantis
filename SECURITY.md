# Security Policy

## Reporting a Vulnerability

We take security issues seriously. Please report a security vulnerability to the maintainers using [private vulnerability reporting](https://github.com/runatlantis/atlantis/security/advisories/new).

If you can't use GitHub, email the security contact at <jose.amengual@gmail.com>.

Configuration for [Anthropic's OSS Scanner](https://red.anthropic.com/oss-scanner/) is in [`.oss-scanner/`](.oss-scanner/): a Dockerfile that builds Atlantis and runs its tests offline, and the [threat model](.oss-scanner/threat_model.md) that scopes what gets reported. Update the threat model when a change moves a trust boundary.

## Maintained releases

Only the two latest minor releases are maintained. For example, if `0.29.7` is the latest, then `0.29.x` and `0.28.x` will receive security fixes.
