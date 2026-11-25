import { existsSync, readdirSync } from 'node:fs';
import path from 'node:path';
import { recordFiles } from './adr-records.mjs';

export function collectSharedPages(site) {
  const pages = [];
  function walk(directory, prefix) {
    if (!existsSync(directory)) return;
    for (const entry of readdirSync(directory, { withFileTypes: true })) {
      const relative = `${prefix}/${entry.name}`;
      if (entry.isDirectory()) walk(path.join(directory, entry.name), relative);
      else if (!entry.name.includes('[')) pages.push(relative);
    }
  }
  for (const prefix of ['contributing', 'blog', 'es/contributing', 'es/blog']) {
    if (existsSync(path.join(site, `${prefix}.md`))) pages.push(`${prefix}.md`);
    walk(path.join(site, prefix), prefix);
    walk(path.join(site, 'public', prefix), prefix);
  }
  for (const locale of ['', 'es/']) {
    if (!existsSync(path.join(site, locale, 'contributing/adr/[record].md'))) continue;
    for (const file of recordFiles(path.resolve(site, '../docs/adr'))) {
      pages.push(`${locale}contributing/adr/${file === 'README.md' ? 'index.md' : file}`);
    }
  }
  return [...new Set(pages)].sort();
}
