import assert from 'node:assert/strict';
import { execFileSync, spawnSync } from 'node:child_process';
import { existsSync, mkdirSync, mkdtempSync, readFileSync, rmSync, symlinkSync, writeFileSync } from 'node:fs';
import { tmpdir } from 'node:os';
import path from 'node:path';
import { fileURLToPath, pathToFileURL } from 'node:url';
import test from 'node:test';
import { createMarkdownRenderer } from 'vitepress';
import { configureVersions, flagVersionBadges, historicalLink, historicalSidebar, validateSharedLink, versionedLinks, versionTarget } from '../../runatlantis.io/.vitepress/versioning.mjs';
import { MIN_DOC_VERSION, englishFallback, fetchDocumentationHistory, gitLastUpdated, includeThemeFile, isSupportedRelease, mergePublicAssets } from './build-utils.mjs';
import { collectSharedPages } from '../../runatlantis.io/.vitepress/shared-pages.mjs';

const source = fileURLToPath(new URL('../../runatlantis.io', import.meta.url));
const version = 'v0.28.0';
const base = `/versions/${version}/`;
const md = await createMarkdownRenderer(source, { config: (renderer) => renderer.use(flagVersionBadges).use(versionedLinks, version) }, base);
const render = (text) => md.render(text, { path: path.join(source, 'docs/server-configuration.md') });

test('release flag headings omit version badges while preserving status, anchors and examples', () => {
  const flag = render('### `--port` <Badge text="v0.1.3+" type="info"/>');
  assert.ok(flag.includes('id="port"'));
  assert.ok(!flag.includes('<Badge'));
  for (const status of ['Experimental', 'Deprecated']) {
    assert.ok(render(`### \`--feature\` <Badge text="${status}" type="warning"/>`).includes(`text="${status}"`));
  }
  assert.ok(render('## Configuration <Badge text="v0.1.3+"/>').includes('text="v0.1.3+"'));
  const example = render('```markdown\n### `--port` <Badge text="v0.1.3+"/>\n```');
  assert.ok(example.includes('v0.1.3+'));
});

test('VitePress keeps relative, root-relative, and historical absolute links inside the release', () => {
  for (const destination of [
    'server-side-repo-config.md#reference',
    '/docs/server-side-repo-config.md#reference',
    'https://www.runatlantis.io/docs/server-side-repo-config.html#reference',
    'http://runatlantis.io/docs/server-side-repo-config.html#reference',
    '//www.runatlantis.io/docs/server-side-repo-config.html#reference',
  ]) {
    const html = render(`[Configuration](${destination})`);
    const href = html.match(/href="([^"]+)"/)[1];
    assert.equal(new URL(href, `https://preview.invalid${base}docs/server-configuration.html`).pathname,
      `${base}docs/server-side-repo-config.html`, destination);
    assert.ok(href.endsWith('#reference'));
    assert.ok(!html.includes('target="_blank"'));
  }
});

test('raw HTML anchors receive the release prefix exactly once', () => {
  const html = render('<a href="/docs/server-side-repo-config.html#reference">Config</a>');
  assert.ok(html.includes(`href="${base}docs/server-side-repo-config.html#reference"`));
  assert.equal(historicalLink(`${base}docs/index.html`, version, true), `${base}docs/index.html`);
});

test('ADR links from a release open the current contributing section', () => {
  const href = '/contributing/adr/0004-govern-feature-lifecycles.md';
  const html = render(`[Feature lifecycle](${href})`);
  assert.ok(html.includes('href="/contributing/adr/0004-govern-feature-lifecycles.html"'));
  assert.ok(html.includes('target="_self"'));
  assert.ok(!html.includes(base));
});

test('shared links are validated even when target=_self bypasses VitePress', () => {
  const pages = collectSharedPages(source);
  assert.ok(pages.includes('contributing/adr/index.md'));
  assert.ok(!pages.some((page) => page.includes('[record]') || page.endsWith('/template.md')));
  for (const version of ['', 'v0.28.0']) {
    const renderer = new md.constructor({ html: true }).use(versionedLinks, version, [], pages);
    assert.equal(renderer.render('[ADR](/contributing/adr/0004-govern-feature-lifecycles.md)', { relativePath: 'docs/experiments.md' }).includes('target="_self"'), Boolean(version));
    assert.throws(() => renderer.render('[Missing](/contributing/nonexistent.md)', { relativePath: 'docs/example.md' }), /Broken current-site link/);
    assert.throws(() => renderer.render('<a href="/blog/nonexistent.html" target="_self">Missing</a>', { relativePath: 'contributing.md' }), /Broken current-site link/);
  }
});

test('current-site Markdown links retain VitePress routing and extension normalization', () => {
  // VitePress caches one renderer per process, so isolate the root-base renderer.
  const html = execFileSync(process.execPath, ['--input-type=module', '-e', `
    import { createMarkdownRenderer } from 'vitepress';
    import { versionedLinks } from './runatlantis.io/.vitepress/versioning.mjs';
    import { collectSharedPages } from './runatlantis.io/.vitepress/shared-pages.mjs';
    const source = ${JSON.stringify(source)};
    const renderer = await createMarkdownRenderer(source, { config: (md) => md.use(versionedLinks, '', [], collectSharedPages(source)) }, '/');
    for (const relativePath of ['contributing.md', 'docs/experiments.md']) {
      console.log(renderer.render('[Blog](/blog/2018/atlantis-0-4-4-now-supports-bitbucket.md)', { relativePath }));
    }
  `], { cwd: path.dirname(source), encoding: 'utf8' });
  assert.equal(html.match(/href="\/blog\/2018\/atlantis-0-4-4-now-supports-bitbucket.html"/g)?.length, 2);
  assert.ok(!html.includes('target="_self"'));
});

test('shared-link validation accepts existing assets and rejects missing ones', () => {
  const directory = mkdtempSync(path.join(tmpdir(), 'docs-shared-assets-'));
  try {
    mkdirSync(path.join(directory, 'blog/2024'), { recursive: true });
    mkdirSync(path.join(directory, 'public/contributing'), { recursive: true });
    writeFileSync(path.join(directory, 'blog/2024/talk.pdf'), 'PDF');
    writeFileSync(path.join(directory, 'public/contributing/diagram.svg'), '<svg/>');
    const inventory = collectSharedPages(directory);
    validateSharedLink('/blog/2024/talk.pdf#page=2', inventory, 'docs/faq.md');
    validateSharedLink('/contributing/diagram.svg', inventory, 'docs/faq.md');
    assert.throws(() => validateSharedLink('/blog/2024/missing.pdf', inventory, 'docs/faq.md'), /Broken current-site link.*missing.pdf in docs\/faq.md/);
  } finally {
    rmSync(directory, { recursive: true, force: true });
  }
});

test('malformed shared URLs identify the link and its source file', () => {
  assert.throws(() => validateSharedLink('/blog/100%-coverage', [], 'docs/faq.md'), /Invalid current-site link \/blog\/100%-coverage in docs\/faq.md/);
});

test('plain and versioned VitePress builds reject missing shared pages', () => {
  const directory = mkdtempSync(path.join(tmpdir(), 'docs-shared-link-build-'));
  try {
    const site = path.join(directory, 'site');
    mkdirSync(path.join(site, '.vitepress'), { recursive: true });
    symlinkSync(path.resolve(source, '../node_modules'), path.join(site, 'node_modules'), 'dir');
    writeFileSync(path.join(site, 'package.json'), '{"type":"module"}');
    const plugin = pathToFileURL(path.join(source, '.vitepress/versioning.mjs')).href;
    const inventory = pathToFileURL(path.join(source, '.vitepress/shared-pages.mjs')).href;
    for (const version of ['', 'v0.28.0']) {
      writeFileSync(path.join(site, '.vitepress/config.js'), `import { versionedLinks } from ${JSON.stringify(plugin)};\nimport { collectSharedPages } from ${JSON.stringify(inventory)};\nexport default { base: ${JSON.stringify(version ? `/versions/${version}/` : '/')}, markdown: { config(md) { md.use(versionedLinks, ${JSON.stringify(version)}, [], collectSharedPages(${JSON.stringify(site)})); } } };\n`);
      writeFileSync(path.join(site, 'index.md'), '# Test\n\n[Missing](/contributing/nonexistent.md)\n');
      const result = spawnSync(process.execPath, [path.resolve(source, '../node_modules/vitepress/bin/vitepress.js'), 'build', site, '--outDir', path.join(directory, 'output')], { encoding: 'utf8' });
      assert.notEqual(result.status, 0);
      assert.match(result.stdout + result.stderr, /Broken current-site link \/contributing\/nonexistent\.(?:html|md)/);
    }
  } finally {
    rmSync(directory, { recursive: true, force: true });
  }
});

test('external links, same-page fragments, and code samples retain their meaning', () => {
  assert.ok(render('[External](https://example.com/docs/reference)').includes('href="https://example.com/docs/reference"'));
  assert.ok(render('[External](https://www.runatlantis.io.evil.example/docs/reference)').includes('www.runatlantis.io.evil.example'));
  assert.ok(render('[Section](#reference)').includes('href="#reference"'));
  assert.ok(render('`https://www.runatlantis.io/docs/server-configuration.html`').includes('<code>https://www.runatlantis.io/docs/server-configuration.html</code>'));
  assert.equal(historicalLink('https://www.runatlantis.io/blog', version), 'https://www.runatlantis.io/blog');
});

test('root release compatibility applies to documentation and preserves current blog links', () => {
  const renderer = new md.constructor().use(versionedLinks, '');
  const link = '[Guide](https://www.runatlantis.io/guide.html)';
  assert.ok(renderer.render(link, { relativePath: 'docs/server-configuration.md' }).includes('href="/guide.html"'));
  assert.ok(renderer.render(link, { relativePath: 'blog/2018/example.md' }).includes('href="https://www.runatlantis.io/guide.html"'));
});

test('current Spanish pages fall back to English documentation when a release has no translation', () => {
  const renderer = new md.constructor().use(versionedLinks, '', [{ version, pages: ['docs.md', 'guide.md'] }]);
  const html = renderer.render('[Docs](./docs.md) [Guide](./guide.md)', { relativePath: 'es/contributing.md' });
  assert.ok(html.includes('href="/docs.md"'));
  assert.ok(html.includes('href="/guide.md"'));
});

test('preview links use the working tree translation inventory instead of the latest release', () => {
  const versions = [
    { version, pages: ['docs.md'] },
    { version: 'main', pages: ['docs.md', 'es/docs.md'] },
  ];
  const renderer = new md.constructor().use(versionedLinks, '', versions, undefined, 'main');
  const html = renderer.render('[Docs](./docs.md)', { relativePath: 'es/contributing.md' });
  assert.ok(html.includes('href="./docs.md"'));
});

test('the sidebar omits pages that did not exist in a release', () => {
  assert.equal(historicalSidebar([{ text: 'Configuración', link: '/es/docs/server-configuration' }], ['docs/server-configuration.md'])[0].link, '/docs/server-configuration');
  assert.equal(historicalSidebar([{ text: 'Guide', link: '/guide' }], ['guide/index.md'])[0].link, '/guide/');
  assert.deepEqual(historicalSidebar([{ text: 'Config', items: [
    { text: 'Server', link: '/docs/server-configuration' },
    { text: 'Experiments', link: '/docs/experiments' },
  ] }], ['docs/server-configuration.md']), [{ text: 'Config', link: undefined, items: [
    { text: 'Server', link: '/docs/server-configuration' },
  ] }]);
});

test('the current website scopes documentation links to the latest release and retains contributing links', () => {
  const config = { themeConfig: {}, locales: { root: { themeConfig: {} } } };
  const sidebar = [{ text: 'Docs', items: [
    { text: 'Server', link: '/docs/server-configuration' },
    { text: 'Experiments', link: '/docs/experiments' },
  ] }, { text: 'Contributing', link: '/contributing' }];
  const versions = [
    { version, pages: ['docs/server-configuration.md'] },
    { version: 'main', pages: ['docs/server-configuration.md', 'docs/experiments.md'] },
  ];
  const result = configureVersions(config, '', versions, sidebar);
  const rendered = JSON.stringify(result.locales.root.themeConfig.sidebar);
  assert.ok(!rendered.includes('Experiments'));
  assert.ok(rendered.includes('Server'));
  assert.ok(rendered.includes('Contributing'));
});

test('HTML theme data includes version names without per-release page inventories', () => {
  const versions = [{ version, pages: Array.from({ length: 5000 }, (_, i) => `docs/page-${i}.md`) }];
  const result = configureVersions({ themeConfig: {}, locales: { root: { themeConfig: {} } } }, '', versions, []);
  assert.deepEqual(result.themeConfig.docVersions, [{ version }]);
  assert.ok(JSON.stringify(result.themeConfig).length < 1000);
});

const configuration = () => ({
  themeConfig: {}, head: [],
  locales: { root: { themeConfig: {} }, es: { themeConfig: { sidebar: [], editLink: { pattern: 'edit/main/:path' } } } },
});

test('preview roots select Main and use its sidebar and page dates while retaining root routes', async () => {
  const versions = [
    { version, pages: ['docs.md'] },
    { version: 'main', pages: ['docs.md', 'docs/new-feature.md', 'es/docs.md', 'es/docs/new-feature.md'] },
  ];
  const metadata = {
    rootVersion: 'main',
    sources: { main: { 'docs/new-feature.md': 'docs/new-feature.md' } },
    lastUpdated: { main: { 'docs/new-feature.md': 2000 } },
  };
  const sidebar = [{ text: 'New feature', link: '/docs/new-feature' }];
  const config = configureVersions(configuration(), '', versions, sidebar, metadata);
  assert.equal(config.themeConfig.docVersion, 'main');
  assert.equal(config.base, undefined);
  assert.equal(config.themeConfig.i18nRouting, undefined);
  assert.equal(config.locales.root.themeConfig.sidebar[0].link, '/docs/new-feature');
  const page = { relativePath: 'docs/new-feature.md', filePath: 'docs/new-feature.md', frontmatter: {} };
  await config.transformPageData(page);
  assert.equal(page.lastUpdated, 2000);
  assert.equal(page.frontmatter.editLink, undefined);
});

test('root timestamps follow the release for docs and HEAD for current contributing and blog pages', async () => {
  const metadata = {
    sources: { [version]: { 'docs/server-configuration.md': 'docs/server-configuration.md' } },
    lastUpdated: { [version]: { 'docs/server-configuration.md': 1000 }, main: { 'contributing.md': 2000, 'blog/2018/example.md': 3000 } },
  };
  const config = configureVersions(configuration(), '', [{ version, pages: [] }], [], metadata);
  for (const [filePath, timestamp] of [['docs/server-configuration.md', 1000], ['contributing.md', 2000], ['blog/2018/example.md', 3000]]) {
    const page = { relativePath: filePath, filePath, frontmatter: {}, lastUpdated: null };
    await config.transformPageData(page);
    assert.equal(page.lastUpdated, timestamp);
  }
  const hidden = { relativePath: 'contributing.md', filePath: 'contributing.md', frontmatter: { lastUpdated: false } };
  await config.transformPageData(hidden);
  assert.equal(hidden.lastUpdated, undefined);
});

test('root pages preserve their edit actions and translated labels', async () => {
  const config = configuration();
  const english = { pattern: 'https://github.com/runatlantis/atlantis/edit/main/runatlantis.io/:path', text: 'Edit this page on GitHub' };
  const spanish = { ...english, text: 'Editar esta página en GitHub' };
  config.themeConfig.editLink = english;
  config.locales.es.themeConfig.editLink = spanish;
  configureVersions(config, '', [{ version, pages: ['docs.md'] }], []);
  assert.equal(config.themeConfig.editLink, english);
  assert.equal(config.locales.es.themeConfig.editLink, spanish);
  const page = { relativePath: 'contributing.md', filePath: 'contributing.md', frontmatter: {} };
  await config.transformPageData(page);
  assert.equal(page.frontmatter.editLink, undefined);
});

test('release source links use original files in every locale and hide generated pages', async () => {
  const sources = {
    'docs/server-configuration.md': 'docs/server-configuration.md',
    'es/docs/server-configuration.md': 'es/docs/server-configuration.md',
    'es/docs/experiments.md': 'docs/experiments.md',
  };
  const config = configureVersions(configuration(), version, [{ version, pages: Object.keys(sources) }], [], { sources: { [version]: sources } });
  for (const [relativePath, original] of Object.entries(sources)) {
    const page = { relativePath, frontmatter: {} };
    await config.transformPageData(page);
    const expected = `https://github.com/runatlantis/atlantis/blob/${version}/runatlantis.io/${original}`;
    assert.equal(config.themeConfig.editLink.pattern(page), expected);
    assert.equal(config.locales.es.themeConfig.editLink.pattern(page), expected);
  }
  const generated = { relativePath: 'index.md', frontmatter: {} };
  await config.transformPageData(generated);
  assert.equal(generated.frontmatter.editLink, false);
});

test('theme copying works from a checkout beneath cache or dist', () => {
  for (const parent of ['cache', 'dist']) {
    const root = `/private/tmp/${parent}/checkout/runatlantis.io/.vitepress`;
    assert.equal(includeThemeFile(root, root), true);
    assert.equal(includeThemeFile(root, `${root}/config.ts`), true);
    assert.equal(includeThemeFile(root, `${root}/components/DocumentationVersion.vue`), true);
    assert.equal(includeThemeFile(root, `${root}/dist/index.html`), false);
    assert.equal(includeThemeFile(root, `${root}/cache/deps.json`), false);
    assert.equal(includeThemeFile(root, `${root}/config.ts.timestamp-123.mjs`), false);
  }
});

test('timestamps match per-file Git history at the selected ref', () => {
  const repo = path.dirname(source);
  for (const [ref, file] of [['HEAD', 'contributing.md'], [version, 'docs/server-configuration.md']]) {
    const expected = Number(execFileSync('git', ['log', '-1', '--format=%at', ref, '--', `runatlantis.io/${file}`], { cwd: repo, encoding: 'utf8' }).trim()) * 1000;
    assert.equal(gitLastUpdated(repo, ref)[file], expected);
  }
});

test('timestamps follow per-file merge history when a merged branch’s edit was discarded', () => {
  const repo = mkdtempSync(path.join(tmpdir(), 'docs-merge-dates-'));
  const identity = { GIT_AUTHOR_NAME: 'Docs fixture', GIT_AUTHOR_EMAIL: 'docs@example.invalid', GIT_COMMITTER_NAME: 'Docs fixture', GIT_COMMITTER_EMAIL: 'docs@example.invalid' };
  const git = (args, dates = {}) => execFileSync('git', ['-c', 'commit.gpgsign=false', '-c', 'core.hooksPath=/dev/null', ...args], { cwd: repo, encoding: 'utf8', env: { ...process.env, ...identity, ...dates }, stdio: 'pipe' }).trim();
  const commit = (name, author, committer) => { git(['add', '.']); git(['commit', '-m', name], { GIT_AUTHOR_DATE: author, GIT_COMMITTER_DATE: committer }); };
  try {
    git(['init', '--initial-branch=main']);
    mkdirSync(path.join(repo, 'runatlantis.io'));
    const file = path.join(repo, 'runatlantis.io/contributing.md');
    const other = path.join(repo, 'runatlantis.io/blog.md');
    writeFileSync(file, 'base\n'); writeFileSync(other, 'base\n');
    commit('base', '2000-01-01T00:00:00Z', '2000-01-01T00:00:00Z');
    git(['switch', '-c', 'feature']);
    writeFileSync(file, 'discarded feature edit\n'); writeFileSync(other, 'accepted feature edit\n');
    commit('feature', '2001-01-01T00:00:00Z', '2006-01-01T00:00:00Z');
    git(['switch', 'main']);
    writeFileSync(file, 'main edit\n');
    commit('main', '2002-01-01T00:00:00Z', '2005-01-01T00:00:00Z');
    assert.throws(() => git(['merge', '--no-ff', '--no-commit', 'feature']));
    writeFileSync(file, 'main edit\n');
    commit('merge', '2003-01-01T00:00:00Z', '2007-01-01T00:00:00Z');
    const expected = Date.parse('2002-01-01T00:00:00Z');
    assert.equal(Number(git(['log', '-1', '--format=%at', '--', 'runatlantis.io/contributing.md'])) * 1000, expected);
    assert.equal(gitLastUpdated(repo, 'HEAD')['contributing.md'], expected);
  } finally {
    rmSync(repo, { recursive: true, force: true });
  }
});

test('deployment restores shallow history and uses author dates rather than committer or cutoff dates', () => {
  const directory = mkdtempSync(path.join(tmpdir(), 'docs-history-check-'));
  const remote = path.join(directory, 'remote');
  const shallow = path.join(directory, 'shallow');
  const firstAuthor = '2000-01-01T12:00:00Z';
  const secondAuthor = '2002-02-02T12:00:00Z';
  const git = (cwd, args, env = {}) => execFileSync('git', args, { cwd, encoding: 'utf8', env: { ...process.env, ...env }, stdio: 'pipe' }).trim();
  const identity = { GIT_AUTHOR_NAME: 'Docs fixture', GIT_AUTHOR_EMAIL: 'docs@example.invalid', GIT_COMMITTER_NAME: 'Docs fixture', GIT_COMMITTER_EMAIL: 'docs@example.invalid' };
  try {
    git(directory, ['init', '--initial-branch=main', remote]);
    mkdirSync(path.join(remote, 'runatlantis.io/docs'), { recursive: true });
    writeFileSync(path.join(remote, 'runatlantis.io/contributing.md'), '# Contributing\n');
    git(remote, ['add', '.']);
    git(remote, ['-c', 'commit.gpgsign=false', '-c', 'core.hooksPath=/dev/null', 'commit', '-m', 'first'], { ...identity, GIT_AUTHOR_DATE: firstAuthor, GIT_COMMITTER_DATE: '2010-01-01T12:00:00Z' });
    git(remote, ['-c', 'tag.gpgsign=false', 'tag', 'v1.0.0']);
    writeFileSync(path.join(remote, 'runatlantis.io/docs/server-configuration.md'), '# Configuration\n');
    git(remote, ['add', '.']);
    git(remote, ['-c', 'commit.gpgsign=false', '-c', 'core.hooksPath=/dev/null', 'commit', '-m', 'second'], { ...identity, GIT_AUTHOR_DATE: secondAuthor, GIT_COMMITTER_DATE: '2020-02-02T12:00:00Z' });
    git(directory, ['clone', '--depth=1', '--no-tags', pathToFileURL(remote).href, shallow]);
    assert.equal(git(shallow, ['rev-parse', '--is-shallow-repository']), 'true');
    assert.equal(gitLastUpdated(shallow, 'HEAD')['contributing.md'], Date.parse(secondAuthor));
    fetchDocumentationHistory(shallow);
    assert.equal(git(shallow, ['rev-parse', '--is-shallow-repository']), 'false');
    assert.equal(gitLastUpdated(shallow, 'HEAD')['contributing.md'], Date.parse(firstAuthor));
    assert.equal(gitLastUpdated(shallow, 'HEAD')['docs/server-configuration.md'], Date.parse(secondAuthor));
    assert.equal(git(shallow, ['tag', '--list']), 'v1.0.0');
    fetchDocumentationHistory(shallow);
  } finally {
    rmSync(directory, { recursive: true, force: true });
  }
});

test('root public assets keep current files and backfill release-only files', () => {
  const directory = mkdtempSync(path.join(tmpdir(), 'docs-assets-check-'));
  try {
    for (const name of ['release', 'current']) mkdirSync(path.join(directory, name));
    writeFileSync(path.join(directory, 'current/logo.svg'), 'current');
    writeFileSync(path.join(directory, 'release/logo.svg'), 'historical');
    writeFileSync(path.join(directory, 'release/archived.png'), 'archive-only');
    mergePublicAssets(path.join(directory, 'release'), path.join(directory, 'current'));
    assert.equal(readFileSync(path.join(directory, 'current/logo.svg'), 'utf8'), 'current');
    assert.equal(readFileSync(path.join(directory, 'current/archived.png'), 'utf8'), 'archive-only');
  } finally {
    rmSync(directory, { recursive: true, force: true });
  }
});

test('switching versions preserves known pages and explicitly identifies unavailable pages', () => {
  const versions = [{ version, pages: ['docs.md', 'docs/server-configuration.md'] }];
  assert.deepEqual(versionTarget('es/docs/server-configuration.md', version, versions), {
    path: `${base}docs/server-configuration.html`, missingPage: false,
  });
  assert.deepEqual(versionTarget('docs/server-configuration.md', version, versions), {
    path: `${base}docs/server-configuration.html`, missingPage: false,
  });
  assert.deepEqual(versionTarget('docs/experiments.md', version, versions), {
    path: `${base}docs.html`, missingPage: true,
  });
  assert.deepEqual(versionTarget('docs/server-configuration.md', '', versions), {
    path: '/docs/server-configuration.html', missingPage: false,
  });
});

test('missing pages fall back to the documentation index in the reader’s language when available', () => {
  for (const [pages, destination] of [
    [['es/docs.md', 'docs.md'], 'es/docs.html'],
    [['es/index.md', 'docs.md'], 'es/'],
    [['docs.md'], 'docs.html'],
    [['index.md'], ''],
  ]) {
    assert.deepEqual(versionTarget('es/docs/missing.md', version, [{ version, pages }]), {
      path: base + destination, missingPage: true,
    });
  }
});

test('English fallback notices preserve front matter and VitePress clears stale output on rebuild', () => {
  const directory = mkdtempSync(path.join(tmpdir(), 'docs-fallback-build-'));
  try {
    const site = path.join(directory, 'site');
    const output = path.join(directory, 'output');
    mkdirSync(path.join(site, '.vitepress'), { recursive: true });
    mkdirSync(path.join(output, 'versions/stale'), { recursive: true });
    symlinkSync(path.resolve(source, '../node_modules'), path.join(site, 'node_modules'), 'dir');
    writeFileSync(path.join(site, 'package.json'), '{"type":"module"}');
    writeFileSync(path.join(site, '.vitepress/config.js'), 'export default { titleTemplate: false };');
    writeFileSync(path.join(output, 'versions/stale/index.html'), 'stale release');
    writeFileSync(path.join(output, 'removed.html'), 'stale page');
    const original = '---\ntitle: Original title\nlayout: doc\naside: false\n---\n# Original heading\n';
    mkdirSync(path.join(site, 'docs'));
    writeFileSync(path.join(site, 'docs/example.md'), original);
    writeFileSync(path.join(site, 'index.md'), englishFallback(original, 'docs/example.md'));
    execFileSync(process.execPath, [path.resolve(source, '../node_modules/vitepress/bin/vitepress.js'), 'build', site, '--outDir', output], { stdio: 'pipe' });
    const html = readFileSync(path.join(output, 'index.html'), 'utf8');
    assert.match(html, /<title>Original title<\/title>/);
    assert.match(html, /Esta página está disponible en inglés/);
    assert.match(html, /id="original-heading"/);
    assert.ok(!html.includes('class="VPDocAsideOutline"'));
    assert.equal(existsSync(path.join(output, 'versions/stale')), false);
    assert.equal(existsSync(path.join(output, 'removed.html')), false);
    for (const original of ['# Heading\n', '---\r\ntitle: CRLF\r\n---\r\n# Heading\r\n']) {
      const fallback = englishFallback(original, 'docs/example.md');
      assert.ok(fallback.endsWith('# Heading' + (original.includes('\r') ? '\r\n' : '\n')));
      if (original.startsWith('---')) assert.ok(fallback.startsWith('---\r\ntitle: CRLF\r\n---\r\n'));
    }
    assert.ok(englishFallback('---\n---\n# Heading\n', 'docs/example.md').startsWith('---\n---\n\n::: info'));
  } finally {
    rmSync(directory, { recursive: true, force: true });
  }
});

test('the build refuses to overwrite a directory it does not own', () => {
  const directory = mkdtempSync(path.join(tmpdir(), 'docs-output-check-'));
  try {
    writeFileSync(path.join(directory, 'keep.txt'), 'existing content');
    const result = spawnSync(process.execPath, [fileURLToPath(new URL('./build-versioned.mjs', import.meta.url)), '--versions', version, '--outDir', directory]);
    assert.notEqual(result.status, 0);
    assert.match(result.stderr.toString(), /choose an empty directory/);
    assert.equal(readFileSync(path.join(directory, 'keep.txt'), 'utf8'), 'existing content');
  } finally {
    rmSync(directory, { recursive: true });
  }
});

test('only locally available stable release tags can become snapshots', () => {
  const directory = mkdtempSync(path.join(tmpdir(), 'docs-tag-check-'));
  try {
    const result = spawnSync(process.execPath, [fileURLToPath(new URL('./build-versioned.mjs', import.meta.url)), '--versions', 'HEAD', '--outDir', directory]);
    assert.notEqual(result.status, 0);
    assert.match(result.stderr.toString(), /release tag is not available locally/);
  } finally {
    rmSync(directory, { recursive: true });
  }
});

test('release discovery lists stable documentation tags without creating build output', () => {
  const directory = mkdtempSync(path.join(tmpdir(), 'docs-discovery-check-'));
  try {
    const output = path.join(directory, 'site');
    const result = spawnSync(process.execPath, [fileURLToPath(new URL('./build-versioned.mjs', import.meta.url)), '--list-versions', '--outDir', output]);
    assert.equal(result.status, 0, result.stderr.toString());
    const versions = result.stdout.toString().trim().split('\n');
    assert.ok(versions.length > 3);
    assert.ok(versions.includes(MIN_DOC_VERSION));
    assert.ok(versions.every(isSupportedRelease));
    assert.ok(!versions.includes('v0.27.0'));
    assert.ok(versions.every((tag) => /^v\d+\.\d+\.\d+$/.test(tag)));
    assert.deepEqual(versions, [...new Set(versions)].sort((a, b) => b.localeCompare(a, undefined, { numeric: true })));
    assert.equal(existsSync(output), false);
  } finally {
    rmSync(directory, { recursive: true });
  }
});

test('documentation support starts at v0.28.0 and includes future stable releases', () => {
  for (const tag of ['v0.4.0', 'v0.27.0', 'v0.27.99', 'v0.28.0-rc.1', 'main', 'HEAD']) assert.equal(isSupportedRelease(tag), false, tag);
  for (const tag of [MIN_DOC_VERSION, 'v0.28.2', 'v0.100.0', 'v1.0.0']) assert.equal(isSupportedRelease(tag), true, tag);
});

test('explicit requests below the release floor are rejected before writing output', () => {
  const directory = mkdtempSync(path.join(tmpdir(), 'docs-version-floor-'));
  try {
    const output = path.join(directory, 'site');
    for (const version of ['v0.4.0', 'v0.27.0']) {
      const result = spawnSync(process.execPath, [fileURLToPath(new URL('./build-versioned.mjs', import.meta.url)), '--versions', version, '--outDir', output]);
      assert.notEqual(result.status, 0);
      assert.match(result.stderr.toString(), /documentation starts at v0\.28\.0/);
      assert.equal(existsSync(output), false);
    }
  } finally {
    rmSync(directory, { recursive: true });
  }
});
