---
aside: false
---
# Atlantis Contributing Documentation

These docs are for users who want to contribute to the Atlantis project. This
can vary from writing documentation, helping the community on Slack, discussing
issues, or writing code.

:::tip Looking to get started or use Atlantis?
If you're new, check out the [Guide](./guide.md) or the
[Documentation](./docs.md).
:::

## Next Steps

### Preview versioned configuration documentation

Install the website dependencies with `npm install`, then build selected release tags:

```sh
npm run website:build:versioned -- --versions v0.28.0,v0.40.0,v0.48.0 --outDir /tmp/atlantis-versioned-docs
python3 -m http.server 8081 --bind 127.0.0.1 --directory /tmp/atlantis-versioned-docs
```

Open `http://127.0.0.1:8081/docs/server-configuration.html`. Root documentation is built from the latest included release, marked `(latest)` in the header selector. `Main (next)` opens development documentation. Without `--versions`, the build selects the newest patch from each of the three most recent minor releases available in local Git tags. Fetch any missing release tags before building.

Versioned documentation starts at v0.28.0, the first VitePress release. Older documentation remains available in [GitHub release tags](https://github.com/runatlantis/atlantis/tags). Use `--versions all` to build every supported stable tag, or `--list-versions` to list them. Explicit versions below v0.28.0 are rejected. The selector reads the generated manifest, so new releases require no source edits.

For accurate page dates, run `node scripts/website/fetch-history.mjs` before building from a shallow checkout.

Set `ATLANTIS_DOCS_BUILD_CONCURRENCY=4` to render four snapshots at a time. The default is one; supported values are 1–8.

The build uses the current theme for release snapshots and Main. The root combines the latest release's documentation with current Contributing and Blog pages. Spanish documentation follows the same release selection when a translation exists. Historical Markdown, assets, and build workspaces are read from Git into temporary directories; only the final site is written to `--outDir`. Generated pages and assets are not committed. Documentation navigation and search stay within their release; Contributing and Blog link to the current site. Switching to an unavailable page or section displays a notice.

The latest release is rendered at both the root and its permanent release URL because their base paths differ. Source links open the original tagged file; Spanish fallback pages link to their English source.

Run link-rendering tests and the browser tests against that preview:

```sh
npm run website:test:versioned
ATLANTIS_VERSIONED_DOCS_URL=http://127.0.0.1:8081 npx playwright test versioned-docs.spec.js
```

The browser tests use all three example release tags above. Set `ATLANTIS_DOCS_BROWSER=chrome` to use an installed Chrome browser instead of Playwright's Chromium.

Netlify fetches release tags and full history, then publishes the versioned build with `--versions all`. In deploy previews (`CONTEXT=deploy-preview`), root documentation shows the PR's working tree with Main (next) selected; production defaults to the latest release. Historical versions remain available in both. Set `CONTEXT=deploy-preview` on the local versioned build to check this behavior.

New tags appear on the next website deployment. CI builds representative release tags and runs the versioning and browser tests. `website:build` remains available for a current-source preview without version selection.

### Contributing to Atlantis

- [Architecture Decisions](./contributing/adr/index.md)&nbsp;&nbsp;–&nbsp;&nbsp;Current proposals, decisions, and their status.
- [Events Controller](./contributing/events-controller.md)&nbsp;&nbsp;–&nbsp;&nbsp;How do the events work?
