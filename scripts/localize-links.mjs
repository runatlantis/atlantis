#!/usr/bin/env node
// Rewrites relative links and image paths in translated pages so a partially
// translated locale still builds.
//
// VitePress fails the build on dead relative links. A translated page keeps its
// source's link targets verbatim, so `./docs/installation-guide.md` inside
// runatlantis.io/es/guide.md would resolve to runatlantis.io/es/docs/... and
// break until that page is translated too.
//
// Rule: if the localized target exists, point at it. Otherwise point back into
// the English tree with a `../` prefix. Readers land on the English page instead
// of a 404, and the locale can grow one page at a time.
//
// Usage:
//   node scripts/localize-links.mjs es [...more locales]

import { readFile, writeFile, readdir } from 'node:fs/promises';
import { existsSync } from 'node:fs';
import path from 'node:path';

const SITE_DIR = 'runatlantis.io';

// Matches the target of a markdown link or image: ](target) or ](target "title")
const LINK_RE = /(\]\()([^)\s]+)((?:\s+"[^"]*")?\))/g;

// These docs embed raw HTML for anything markdown can't express (sized images,
// nested lists with inline images), so src=/href= need the same treatment.
const HTML_ATTR_RE = /(<[^>]*?\b(?:src|href)=)(["'])([^"']+)(\2)/g;

const isExternal = (target) =>
  /^(?:[a-z][a-z0-9+.-]*:|\/\/|#)/i.test(target);

async function markdownFiles(root) {
  if (!existsSync(root)) return [];
  const entries = await readdir(root, { recursive: true, withFileTypes: true });
  return entries
    .filter((entry) => entry.isFile() && entry.name.endsWith('.md'))
    .map((entry) => path.join(entry.parentPath ?? entry.path ?? root, entry.name));
}

function rewriteTarget(target, fileDir, localeRoot, englishRoot) {
  if (isExternal(target)) return target;

  const [rawPath, anchor = ''] = target.split(/(?=#)/, 2);
  if (rawPath === '') return target; // pure anchor
  if (path.isAbsolute(rawPath)) return target; // site-absolute, already resolvable

  const resolved = path.resolve(fileDir, rawPath);
  if (existsSync(resolved)) return target; // localized target exists

  // Same path, but in the English tree.
  const relativeToLocale = path.relative(localeRoot, resolved);
  const englishTarget = path.join(englishRoot, relativeToLocale);
  if (!existsSync(englishTarget)) return target; // leave it; the build will flag it

  let rewritten = path.relative(fileDir, englishTarget);
  if (!rewritten.startsWith('.')) rewritten = `./${rewritten}`;
  return `${rewritten}${anchor}`;
}

const locales = process.argv.slice(2);
if (locales.length === 0) {
  console.error('usage: node scripts/localize-links.mjs <locale> [...locales]');
  process.exit(2);
}

let rewrittenFiles = 0;

for (const locale of locales) {
  const localeRoot = path.resolve(SITE_DIR, locale);
  const englishRoot = path.resolve(SITE_DIR);

  for (const file of await markdownFiles(path.join(SITE_DIR, locale))) {
    const original = await readFile(file, 'utf8');
    const fileDir = path.dirname(path.resolve(file));

    const updated = original
      .replace(LINK_RE, (match, open, target, close) => {
        const next = rewriteTarget(target, fileDir, localeRoot, englishRoot);
        return next === target ? match : `${open}${next}${close}`;
      })
      .replace(HTML_ATTR_RE, (match, open, quote, target, close) => {
        const next = rewriteTarget(target, fileDir, localeRoot, englishRoot);
        return next === target ? match : `${open}${quote}${next}${close}`;
      });

    if (updated !== original) {
      await writeFile(file, updated);
      rewrittenFiles += 1;
      console.log(`rewrote links: ${file}`);
    }
  }
}

console.log(`Link localization complete (${rewrittenFiles} file(s) changed).`);
