#!/usr/bin/env node
// Re-pads markdown table cells so every row's pipes line up with the header.
//
// The English docs keep their tables column-aligned and markdownlint enforces
// that (MD060, style "aligned"). Translated cells are rarely the same width as
// the source, so a translated table fails the same rule on every row.
// Realigning is purely cosmetic for the rendered page but keeps the locale
// trees passing the same lint the English tree does.
//
// Rules: skip fenced code blocks; a table is a run of lines starting with `|`
// whose second line is a delimiter row; cells are split on every `|` (the docs
// never escape pipes); rows whose cell count differs from the header are left
// untouched so a stray pipe cannot corrupt a row.
//
// Usage:
//   node scripts/align-tables.mjs es [...more locales]

import { readFile, writeFile, readdir } from 'node:fs/promises';
import { existsSync } from 'node:fs';
import path from 'node:path';

const SITE_DIR = 'runatlantis.io';

const FENCE_RE = /^\s*(`{3,}|~{3,})/;
const TABLE_ROW_RE = /^\s*\|.*\|\s*$/;
const DELIMITER_RE = /^\s*\|(\s*:?-+:?\s*\|)+\s*$/;

async function markdownFiles(root) {
  if (!existsSync(root)) return [];
  const entries = await readdir(root, { recursive: true, withFileTypes: true });
  return entries
    .filter((entry) => entry.isFile() && entry.name.endsWith('.md'))
    .map((entry) => path.join(entry.parentPath ?? entry.path ?? root, entry.name));
}

// "| a | b |" -> ["a", "b"]
const splitCells = (line) => {
  const trimmed = line.trim();
  return trimmed.slice(1, trimmed.endsWith('|') ? -1 : undefined).split('|').map((c) => c.trim());
};

const delimiterCell = (cell, width) => {
  const left = cell.startsWith(':');
  const right = cell.endsWith(':');
  const dashes = '-'.repeat(Math.max(3, width - (left ? 1 : 0) - (right ? 1 : 0)));
  return `${left ? ':' : ''}${dashes}${right ? ':' : ''}`;
};

function alignTable(lines) {
  const rows = lines.map(splitCells);
  const header = rows[0];
  const columns = header.length;
  const alignable = rows.map((cells) => cells.length === columns);

  const widths = header.map((_, col) =>
    Math.max(
      3,
      ...rows.filter((cells, i) => alignable[i] && i !== 1).map((cells) => cells[col].length),
    ),
  );

  return lines.map((line, i) => {
    if (!alignable[i]) return line;
    const cells = rows[i].map((cell, col) =>
      i === 1 ? delimiterCell(cell, widths[col]) : cell.padEnd(widths[col]),
    );
    return `| ${cells.join(' | ')} |`;
  });
}

function alignFile(source) {
  const lines = source.split('\n');
  const out = [];
  let inFence = false;
  let fence = '';

  for (let i = 0; i < lines.length; i += 1) {
    const line = lines[i];
    const fenceMatch = line.match(FENCE_RE);
    if (fenceMatch) {
      if (!inFence) {
        inFence = true;
        fence = fenceMatch[1];
      } else if (fenceMatch[1].startsWith(fence[0]) && fenceMatch[1].length >= fence.length) {
        inFence = false;
      }
      out.push(line);
      continue;
    }
    if (inFence || !TABLE_ROW_RE.test(line) || !DELIMITER_RE.test(lines[i + 1] ?? '')) {
      out.push(line);
      continue;
    }

    let end = i;
    while (end < lines.length && TABLE_ROW_RE.test(lines[end])) end += 1;
    out.push(...alignTable(lines.slice(i, end)));
    i = end - 1;
  }

  return out.join('\n');
}

const locales = process.argv.slice(2);
if (locales.length === 0) {
  console.error('usage: node scripts/align-tables.mjs <locale> [...locales]');
  process.exit(2);
}

let changed = 0;
for (const locale of locales) {
  for (const file of await markdownFiles(path.join(SITE_DIR, locale))) {
    const original = await readFile(file, 'utf8');
    const updated = alignFile(original);
    if (updated !== original) {
      await writeFile(file, updated);
      changed += 1;
      console.log(`aligned tables: ${file}`);
    }
  }
}
console.log(`Table alignment complete (${changed} file(s) changed).`);
