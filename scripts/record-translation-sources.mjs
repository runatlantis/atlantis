#!/usr/bin/env node
// Records which English source each translated page was generated from.
//
// check-translation-integrity.mjs skips a committed translation whose source
// changed after it was translated. Merge order cannot tell that apart: a
// translation pull request generated before a source change can merge after
// it, and its pages then look current while being older than their source. So
// translate-docs records the git blob hash of the English source for every page
// it (re)wrote, in i18n.sources.json, and the checker compares against that.
//
// Run after the integrity checks and quarantine, from the repository root. A
// page counts as rewritten when git reports it modified or untracked; pages the
// run left alone, including quarantined ones restored to their committed
// version, keep their existing entry.
//
// Usage:
//   node scripts/record-translation-sources.mjs es [...more locales]

import { readFile, writeFile } from 'node:fs/promises';
import { existsSync } from 'node:fs';
import { execFileSync } from 'node:child_process';
import path from 'node:path';

const SITE_DIR = 'runatlantis.io';
const SOURCES_FILE = 'i18n.sources.json';

const git = (...args) => execFileSync('git', args, { encoding: 'utf8' }).trim();

const locales = process.argv.slice(2);
if (locales.length === 0) {
  console.error('usage: node scripts/record-translation-sources.mjs <locale> [...locales]');
  process.exit(2);
}

const sources = existsSync(SOURCES_FILE)
  ? JSON.parse(await readFile(SOURCES_FILE, 'utf8'))
  : {};

let recorded = 0;
for (const locale of locales) {
  const localeDir = path.join(SITE_DIR, locale);
  // -z keeps paths with spaces intact; -uall lists files inside new directories.
  // Not through git(): trimming would eat the leading space of the first
  // " M path" status entry.
  const status = execFileSync('git', ['status', '--porcelain', '-z', '-uall', '--', localeDir], { encoding: 'utf8' })
    .split('\0')
    .filter(Boolean);
  for (const entry of status) {
    const code = entry.slice(0, 2);
    const target = entry.slice(3);
    if (code.includes('D') || !target.endsWith('.md')) continue;
    const relative = path.relative(localeDir, target);
    const source = path.join(SITE_DIR, relative);
    if (!existsSync(source)) continue;
    sources[`${locale}/${relative.split(path.sep).join('/')}`] = git('hash-object', source);
    recorded += 1;
  }
}

const sorted = Object.fromEntries(Object.entries(sources).sort(([a], [b]) => a.localeCompare(b)));
await writeFile(SOURCES_FILE, `${JSON.stringify(sorted, null, 2)}\n`);
console.log(`Recorded the English source of ${recorded} translated page(s) in ${SOURCES_FILE}.`);
