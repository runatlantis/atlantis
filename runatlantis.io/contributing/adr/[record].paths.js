import { readFileSync } from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';
import { recordFiles } from '../../.vitepress/adr-records.mjs';

const source = fileURLToPath(new URL('../../../docs/adr/', import.meta.url));
const repository = 'https://github.com/runatlantis/atlantis/blob/main/docs/adr/';

export default {
  paths() {
    return recordFiles(source).map((file) => {
      let content = readFileSync(path.join(source, file), 'utf8')
        .replace(/\]\(README\.md(?=[)#])/g, '](index.md')
        .replace(/\]\(template\.md(?=[)#])/g, `](${repository}template.md`)
        .replace(/\]\((\.\.\/[^\s)]+)\)/g, (_, target) => `](${new URL(target, repository).href})`);
      if (file === 'README.md') {
        content = content.replace(/^(# .+\n)/, '$1\nThese records reflect current proposals and decisions from main. Acceptance does not imply release availability; check the release documentation for supported features.\n');
      }
      content += `\n\n[View source](${repository}${file})\n`;
      return { params: { record: file === 'README.md' ? 'index' : file.slice(0, -3) }, content };
    });
  },
};
