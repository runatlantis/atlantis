#!/usr/bin/env node
// Verifies that a machine-translated docs page still carries the same
// executable content as its English source.
//
// The translator parses markdown into an AST and skips code nodes, so fenced
// blocks are structurally safe. Everything else on this list is plain text to
// the parser and is where real breakage lands: VitePress container markers,
// <Badge> markup, and link targets. A mistranslated --gh-webhook-secret example
// is worse for a reader than an untranslated page, so this runs in CI and fails
// the build rather than warning.
//
// Usage:
//   node scripts/check-translation-integrity.mjs es [...more locales]

import { readFile, readdir } from 'node:fs/promises';
import { existsSync } from 'node:fs';
import path from 'node:path';

const SITE_DIR = 'runatlantis.io';

// Source trees that get translated. Mirrors the include list in i18n.json.
// '.' means the top-level pages only (not recursive).
const TRANSLATED_DIRS = [
  { dir: 'docs', recursive: true },
  { dir: 'guide', recursive: true },
  { dir: '.', recursive: false },
];

const FENCE_RE = /^([ \t]*)(`{3,}|~{3,})([^\n]*)\n([\s\S]*?)^[ \t]*\2[ \t]*$/gm;
const INLINE_CODE_RE = /(?<!`)`([^`\n]+)`(?!`)/g;
const LINK_RE = /\]\(([^)\s]+)(?:\s+"[^"]*")?\)/g;
const CONTAINER_RE = /^:::+\s*([a-z-]+)/gm;
const BADGE_RE = /<Badge\b[^>]*>/g;
const HEADING_RE = /^#{1,6}\s+\S/gm;
const FRONTMATTER_RE = /^---\r?\n([\s\S]*?)\r?\n---/;

const collect = (text, re, pick = (m) => m[1]) =>
  [...text.matchAll(re)].map(pick);

/** Code blocks must survive byte for byte, including the info string. */
const codeBlocks = (text) =>
  collect(text, FENCE_RE, (m) => `${m[3].trim()}\n${m[4]}`);

/** Frontmatter keys, plus the values that VitePress interprets rather than renders. */
const frontmatter = (text) => {
  const match = FRONTMATTER_RE.exec(text);
  if (!match) return { keys: [], structural: {} };
  const keys = [];
  const structural = {};
  for (const line of match[1].split('\n')) {
    const kv = /^([A-Za-z0-9_-]+):\s*(.*)$/.exec(line);
    if (!kv) continue;
    keys.push(kv[1]);
    if (['layout', 'aside', 'outline', 'sidebar', 'navbar', 'editLink'].includes(kv[1])) {
      structural[kv[1]] = kv[2].trim();
    }
  }
  return { keys: keys.sort(), structural };
};

const sameMultiset = (a, b) => {
  if (a.length !== b.length) return false;
  const sorted = (xs) => [...xs].sort();
  return sorted(a).every((value, i) => value === sorted(b)[i]);
};

/** Resolves a markdown link target to an absolute path plus its anchor, or null if it isn't a relative file link. */
function resolveTarget(target, fromDir) {
  if (/^(?:[a-z][a-z0-9+.-]*:|\/\/)/i.test(target)) return null; // external
  const hash = target.indexOf('#');
  const filePart = hash === -1 ? target : target.slice(0, hash);
  const anchor = hash === -1 ? '' : target.slice(hash);
  if (filePart === '') return null; // pure anchor
  if (path.isAbsolute(filePart)) return { file: filePart, anchor }; // site-absolute
  return { file: path.resolve(fromDir, filePart), anchor };
}

function compare(sourceText, targetText, sourceDir, targetDir, locale) {
  const problems = [];

  const srcBlocks = codeBlocks(sourceText);
  const dstBlocks = codeBlocks(targetText);
  if (srcBlocks.length !== dstBlocks.length) {
    problems.push(`fenced code blocks: source has ${srcBlocks.length}, translation has ${dstBlocks.length}`);
  } else {
    srcBlocks.forEach((block, i) => {
      if (block !== dstBlocks[i]) {
        problems.push(`fenced code block #${i + 1} was modified by translation`);
      }
    });
  }

  const srcInline = collect(sourceText, INLINE_CODE_RE);
  const dstInline = collect(targetText, INLINE_CODE_RE);
  if (!sameMultiset(srcInline, dstInline)) {
    const missing = srcInline.filter((c) => !dstInline.includes(c));
    problems.push(
      `inline code spans differ` + (missing.length ? ` (missing: ${missing.slice(0, 5).map((c) => `\`${c}\``).join(', ')})` : ''),
    );
  }

  const srcLinks = collect(sourceText, LINK_RE);
  const dstLinks = collect(targetText, LINK_RE);
  if (srcLinks.length !== dstLinks.length) {
    problems.push(`link count differs: source ${srcLinks.length}, translation ${dstLinks.length}`);
  } else {
    // Links are compared after resolution, not as literal strings:
    // scripts/localize-links.mjs legitimately rewrites `./docs/x.md` to
    // `../docs/x.md` so a partial locale still builds. A link is acceptable if
    // it resolves either to the same file as the source or to that file's
    // counterpart inside the locale.
    srcLinks.forEach((srcLink, i) => {
      const dstLink = dstLinks[i];
      if (srcLink === dstLink) return;

      const srcResolved = resolveTarget(srcLink, sourceDir);
      const dstResolved = resolveTarget(dstLink, targetDir);
      if (srcResolved === null || dstResolved === null) {
        problems.push(`link #${i + 1} changed: '${srcLink}' became '${dstLink}'`);
        return;
      }
      if (srcResolved.anchor !== dstResolved.anchor) {
        problems.push(`link #${i + 1} anchor changed: '${srcLink}' became '${dstLink}'`);
        return;
      }

      const localeTwin = path.join(SITE_DIR, locale, path.relative(SITE_DIR, srcResolved.file));
      if (dstResolved.file !== srcResolved.file && dstResolved.file !== path.resolve(localeTwin)) {
        problems.push(`link #${i + 1} points somewhere else: '${srcLink}' became '${dstLink}'`);
      }
    });
  }

  const srcContainers = collect(sourceText, CONTAINER_RE);
  const dstContainers = collect(targetText, CONTAINER_RE);
  if (srcContainers.join('|') !== dstContainers.join('|')) {
    problems.push(
      `VitePress container keywords differ: source [${srcContainers.join(', ')}] vs translation [${dstContainers.join(', ')}]`,
    );
  }

  const srcBadges = collect(sourceText, BADGE_RE, (m) => m[0]);
  const dstBadges = collect(targetText, BADGE_RE, (m) => m[0]);
  if (!sameMultiset(srcBadges, dstBadges)) {
    problems.push(`<Badge> markup differs: source ${srcBadges.length}, translation ${dstBadges.length}`);
  }

  const srcHeadings = collect(sourceText, HEADING_RE, (m) => m[0]).length;
  const dstHeadings = collect(targetText, HEADING_RE, (m) => m[0]).length;
  if (srcHeadings !== dstHeadings) {
    problems.push(`heading count differs: source ${srcHeadings}, translation ${dstHeadings}`);
  }

  const srcFm = frontmatter(sourceText);
  const dstFm = frontmatter(targetText);
  if (srcFm.keys.join(',') !== dstFm.keys.join(',')) {
    problems.push(`frontmatter keys differ: [${srcFm.keys}] vs [${dstFm.keys}]`);
  }
  for (const [key, value] of Object.entries(srcFm.structural)) {
    if (dstFm.structural[key] !== value) {
      problems.push(`frontmatter '${key}' must not be translated: '${value}' became '${dstFm.structural[key]}'`);
    }
  }

  return problems;
}

async function sourceFiles() {
  const files = [];
  for (const { dir, recursive } of TRANSLATED_DIRS) {
    const base = path.join(SITE_DIR, dir);
    if (!existsSync(base)) continue;
    const entries = await readdir(base, { recursive, withFileTypes: true });
    for (const entry of entries) {
      if (!entry.isFile() || !entry.name.endsWith('.md')) continue;
      const absolute = path.join(entry.parentPath ?? entry.path ?? base, entry.name);
      files.push(path.relative(SITE_DIR, absolute));
    }
  }
  return [...new Set(files)].sort();
}

const locales = process.argv.slice(2);
if (locales.length === 0) {
  console.error('usage: node scripts/check-translation-integrity.mjs <locale> [...locales]');
  process.exit(2);
}

let failures = 0;
let checked = 0;

for (const locale of locales) {
  for (const relative of await sourceFiles()) {
    const target = path.join(SITE_DIR, locale, relative);
    if (!existsSync(target)) continue; // not translated yet — that is allowed
    const source = path.join(SITE_DIR, relative);
    const [sourceText, targetText] = await Promise.all([
      readFile(source, 'utf8'),
      readFile(target, 'utf8'),
    ]);
    checked += 1;
    const problems = compare(
      sourceText,
      targetText,
      path.dirname(path.resolve(source)),
      path.dirname(path.resolve(target)),
      locale,
    );
    if (problems.length > 0) {
      failures += 1;
      console.error(`\n✗ ${target}`);
      for (const problem of problems) console.error(`    ${problem}`);
    }
  }
}

if (failures > 0) {
  console.error(`\n${failures} of ${checked} translated file(s) failed integrity checks.`);
  process.exit(1);
}

console.log(`Translation integrity: ${checked} file(s) checked, no problems found.`);
