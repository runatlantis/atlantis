import { readFileSync } from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';
import { recordFiles } from '../../.vitepress/adr-records.mjs';

const source = fileURLToPath(new URL('../../../docs/adr/', import.meta.url));
const repository = 'https://github.com/runatlantis/atlantis/blob/main/docs/adr/';

// createRecordPaths builds the ADR pages for one locale. assets is the path
// from that locale's ADR pages to docs/adr/assets, so images next to the
// records resolve from every locale.
export function createRecordPaths(assets) {
  return {
    paths() {
      return recordFiles(source).map((file) => {
        let content = readFileSync(path.join(source, file), 'utf8')
          .replace(/\]\(README\.md(?=[)#])/g, '](index.md')
          .replace(/\]\(template\.md(?=[)#])/g, `](${repository}template.md`)
          .replace(/\]\((\.\.\/[^\s)]+)\)/g, (_, target) => `](${new URL(target, repository).href})`)
          .replace(/(!\[[^\]]*\])\(assets\//g, `$1(${assets}`);
        if (file === 'README.md') {
          content = content.replace(/^(# .+\n)/, '$1\nThese records reflect current proposals and decisions from main. Acceptance does not imply release availability; check the release documentation for supported features.\n');
        }
        content += `\n\n[View source](${repository}${file})\n`;
        return { params: { record: file === 'README.md' ? 'index' : file.slice(0, -3) }, content };
      });
    },
  };
}

export default createRecordPaths('../../../docs/adr/assets/');
