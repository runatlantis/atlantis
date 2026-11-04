import { execFileSync, spawn } from 'node:child_process';
import { cpSync, existsSync, mkdirSync, mkdtempSync, readdirSync, readFileSync, realpathSync, rmSync, symlinkSync, writeFileSync } from 'node:fs';
import { tmpdir } from 'node:os';
import path from 'node:path';
import { fileURLToPath } from 'node:url';
import { MIN_DOC_VERSION, englishFallback, gitLastUpdated, includeThemeFile, isSupportedRelease, mergePublicAssets } from './build-utils.mjs';
import { collectSharedPages } from '../../runatlantis.io/.vitepress/shared-pages.mjs';

const repo = path.resolve(path.dirname(fileURLToPath(import.meta.url)), '../..');
const args = process.argv.slice(2);
const concurrency = Number(process.env.ATLANTIS_DOCS_BUILD_CONCURRENCY || 1);
if (!Number.isInteger(concurrency) || concurrency < 1 || concurrency > 8) throw new Error('ATLANTIS_DOCS_BUILD_CONCURRENCY must be an integer from 1 to 8');
let requested;
let listVersions = false;
let output = path.join(tmpdir(), 'atlantis-versioned-docs');
for (let i = 0; i < args.length; i++) {
  if (args[i] === '--versions' && args[i + 1]) requested = args[++i].split(',');
  else if (args[i] === '--outDir' && args[i + 1]) output = path.resolve(args[++i]);
  else if (args[i] === '--list-versions') listVersions = true;
  else throw new Error(`unknown or incomplete option ${args[i]}`);
}
if (repo.startsWith(path.resolve(output) + path.sep) || path.resolve(output) === repo) throw new Error('output must be a dedicated build directory');
const outputMarker = path.join(output, '.atlantis-docs-build');
if (existsSync(output) && readdirSync(output).length && !existsSync(outputMarker) && !existsSync(path.join(output, 'doc-versions.json'))) {
  throw new Error('output directory is not an existing versioned documentation build; choose an empty directory');
}
const git = (...argv) => execFileSync('git', argv, { cwd: repo, maxBuffer: 128 * 1024 * 1024 });
const stable = git('tag', '--list', 'v*').toString().trim().split('\n').filter(isSupportedRelease);
stable.sort((a, b) => b.localeCompare(a, undefined, { numeric: true }));
const objects = execFileSync('git', ['cat-file', '--batch-check=%(objecttype)'], {
  cwd: repo, input: stable.map((tag) => `${tag}:runatlantis.io/docs/server-configuration.md`).join('\n') + '\n',
}).toString().trim().split('\n');
const available = stable.filter((_, i) => objects[i] === 'blob');
if (listVersions) {
  console.log(available.join('\n'));
  process.exit(0);
}
const seen = new Set();
const tags = requested?.length === 1 && requested[0] === 'all' ? available : requested ?? available.filter((tag) => {
  const minor = tag.split('.').slice(0, 2).join('.');
  if (seen.has(minor)) return false;
  seen.add(minor);
  return true;
}).slice(0, 3);
if (!tags.length || new Set(tags).size !== tags.length) throw new Error('select at least one distinct release tag');
for (const tag of tags) {
  if (/^v\d+\.\d+\.\d+$/.test(tag) && !isSupportedRelease(tag)) throw new Error(`documentation starts at ${MIN_DOC_VERSION}; unsupported release tag: ${tag}`);
  if (!/^v\d+\.\d+\.\d+$/.test(tag) || !available.includes(tag)) throw new Error(`release tag is not available locally with configuration documentation: ${tag}`);
}
tags.sort((a, b) => b.localeCompare(a, undefined, { numeric: true }));
const rootVersion = process.env.CONTEXT === 'deploy-preview' ? 'main' : tags[0];

const workspace = mkdtempSync(path.join(realpathSync(tmpdir()), 'atlantis-doc-build-'));
function pagesIn(directory, prefix = '') {
  return readdirSync(directory, { withFileTypes: true }).flatMap((entry) => {
    if (entry.name.startsWith('.') || entry.name === 'node_modules') return [];
    const relative = path.posix.join(prefix, entry.name);
    return entry.isDirectory() ? pagesIn(path.join(directory, entry.name), relative) : relative.endsWith('.md') ? [relative] : [];
  });
}
const snapshots = [];
const folders = ['docs', 'guide', 'logs', 'hosting', 'terraform', 'public', 'es/docs', 'es/guide'];
const landingPages = ['docs.md', 'guide.md', 'es/docs.md', 'es/guide.md'];
try {
  for (const version of [tags[0], 'main', ...tags.slice(1)]) {
    let historical = path.join(repo, 'runatlantis.io');
    if (version !== 'main') {
      const archive = path.join(workspace, `${version}-source`);
      mkdirSync(archive);
      execFileSync('tar', ['-xf', '-', '-C', archive], { input: git('archive', '--format=tar', version, 'runatlantis.io') });
      historical = path.join(archive, 'runatlantis.io');
    }
    const site = path.join(workspace, version);
    mkdirSync(site);
    const theme = path.join(repo, 'runatlantis.io/.vitepress');
    cpSync(theme, path.join(site, '.vitepress'), { recursive: true, filter: (source) => includeThemeFile(theme, source) });
    for (const folder of folders) {
      if (existsSync(path.join(historical, folder))) cpSync(path.join(historical, folder), path.join(site, folder), { recursive: true });
    }
    for (const file of landingPages) {
      if (existsSync(path.join(historical, file))) cpSync(path.join(historical, file), path.join(site, file));
    }
    const sources = Object.fromEntries(pagesIn(site).map((file) => [file, file]));
    cpSync(path.join(site, 'docs.md'), path.join(site, 'index.md'));
    sources['index.md'] = 'docs.md';
    if (existsSync(path.join(site, 'es/docs'))) {
      // Keep language-menu routes valid when a release has partial translations.
      for (const folder of ['docs', 'guide']) {
        if (!existsSync(path.join(site, folder))) continue;
        for (const page of pagesIn(path.join(site, folder))) {
          const translated = path.join(site, 'es', folder, page);
          if (existsSync(translated)) continue;
          mkdirSync(path.dirname(translated), { recursive: true });
          writeFileSync(translated, englishFallback(readFileSync(path.join(site, folder, page), 'utf8'), `${folder}/${page}`));
          const original = sources[`${folder}/${page}`];
          if (original) sources[`es/${folder}/${page}`] = original;
        }
      }
      cpSync(path.join(site, 'es/docs.md'), path.join(site, 'es/index.md'));
      sources['es/index.md'] = 'es/docs.md';
    }
    symlinkSync(path.join(repo, 'node_modules'), path.join(site, 'node_modules'), 'dir');
    writeFileSync(path.join(site, 'package.json'), '{"type":"module"}\n');
    snapshots.push({ version, site, pages: pagesIn(site), sources });
  }
  const manifest = snapshots.map(({ version, pages }) => ({ version, pages }));
  const manifestFile = path.join(workspace, 'doc-versions.json');
  writeFileSync(manifestFile, JSON.stringify(manifest));
  const metadataFile = path.join(workspace, 'build-metadata.json');
  writeFileSync(metadataFile, JSON.stringify({
    rootVersion,
    sharedPages: collectSharedPages(path.join(repo, 'runatlantis.io')),
    sources: Object.fromEntries(snapshots.map(({ version, sources }) => [version, sources])),
    lastUpdated: { main: gitLastUpdated(repo, 'HEAD'), ...(rootVersion !== 'main' ? { [rootVersion]: gitLastUpdated(repo, rootVersion) } : {}) },
  }));
  const rootSite = path.join(workspace, 'current');
  const rootDocs = snapshots.find(({ version }) => version === rootVersion);
  cpSync(path.join(repo, 'runatlantis.io'), rootSite, { recursive: true, filter: (source) => {
    const relative = path.relative(path.join(repo, 'runatlantis.io'), source).split(path.sep).join('/');
    return !/^(?:\.vitepress\/(?:dist|cache)|node_modules|e2e)(\/|$)/.test(relative)
      && !/^(?:(?:es\/)?(?:docs|guide)(?:\/|\.md$)|logs\/|hosting\/|terraform\/)/.test(relative);
  } });
  for (const folder of folders) {
    if (!existsSync(path.join(rootDocs.site, folder))) continue;
    if (folder === 'public') mergePublicAssets(path.join(rootDocs.site, folder), path.join(rootSite, folder));
    else cpSync(path.join(rootDocs.site, folder), path.join(rootSite, folder), { recursive: true });
  }
  for (const file of landingPages) {
    if (existsSync(path.join(rootDocs.site, file))) cpSync(path.join(rootDocs.site, file), path.join(rootSite, file));
  }
  mkdirSync(path.join(workspace, 'docs'), { recursive: true });
  symlinkSync(path.join(repo, 'docs/adr'), path.join(workspace, 'docs/adr'), 'dir');
  symlinkSync(path.join(repo, 'node_modules'), path.join(rootSite, 'node_modules'), 'dir');
  writeFileSync(path.join(rootSite, 'package.json'), '{"type":"module"}\n');
  async function build({ version, site }) {
    const destination = version ? path.join(output, 'versions', version) : output;
    console.log(`Building ${version || 'main'} → ${destination}`);
    await new Promise((resolve, reject) => {
      const child = spawn(process.execPath, [path.join(repo, 'node_modules/vitepress/bin/vitepress.js'), 'build', site, '--outDir', destination], {
        cwd: repo, stdio: 'inherit', env: { ...process.env, ATLANTIS_DOC_VERSION: version, ATLANTIS_DOC_VERSIONS_FILE: manifestFile, ATLANTIS_DOC_BUILD_METADATA_FILE: metadataFile },
      });
      child.on('error', reject);
      child.on('close', (code) => code === 0 ? resolve() : reject(new Error(`building ${version || 'main website'} failed`)));
    });
  }
  await build({ version: '', site: rootSite });
  let next = 0;
  let failure;
  await Promise.all(Array.from({ length: Math.min(concurrency, snapshots.length) }, async () => {
    while (!failure && next < snapshots.length) {
      try { await build(snapshots[next++]); }
      catch (error) { failure = error; }
    }
  }));
  if (failure) throw failure;
  writeFileSync(path.join(output, 'doc-versions.json'), JSON.stringify(manifest, null, 2) + '\n');
  console.log(`Versioned documentation built at ${output}`);
} finally {
  rmSync(workspace, { recursive: true, force: true });
  if (existsSync(output)) writeFileSync(outputMarker, 'Generated documentation output. Rebuildable from Git tags.\n');
}
