import assert from 'node:assert/strict';
import { execFileSync } from 'node:child_process';
import { existsSync, mkdirSync, mkdtempSync, readFileSync, rmSync, writeFileSync } from 'node:fs';
import { tmpdir } from 'node:os';
import path from 'node:path';
import { fileURLToPath } from 'node:url';
import test from 'node:test';
import records from '../../runatlantis.io/contributing/adr/[record].paths.js';

test('ADR routes publish the index and numbered records while linking to the template on GitHub', () => {
  const pages = records.paths();
  assert.ok(pages.every(({ params }) => params.record === 'index' || /^\d{4}-/.test(params.record)));
  assert.ok(!pages.some(({ params }) => params.record === 'template'));
  assert.ok(pages.find(({ params }) => params.record === 'index').content.includes('https://github.com/runatlantis/atlantis/blob/main/docs/adr/template.md'));
});

test('translation staging skips ADR sources and copy-back preserves hand-written locale files', () => {
  const repository = fileURLToPath(new URL('../../', import.meta.url));
  const directory = mkdtempSync(path.join(tmpdir(), 'adr-translations-'));
  const put = (file, content) => {
    const target = path.join(directory, file);
    mkdirSync(path.dirname(target), { recursive: true });
    writeFileSync(target, content);
  };
  const read = (file) => readFileSync(path.join(directory, file), 'utf8');
  const wrapper = 'contributing/adr/[record].md';
  const loader = 'contributing/adr/[record].paths.js';
  try {
    for (const section of ['docs', 'guide', 'blog', 'contributing']) put(`runatlantis.io/${section}/example.md`, '# Example');
    put('runatlantis.io/index.md', '# Home');
    put(`runatlantis.io/${wrapper}`, '<!-- @content -->');
    put(`runatlantis.io/es/${wrapper}`, 'Registro original en inglés.\n\n<!-- @content -->');
    put(`runatlantis.io/es/${loader}`, 'export { default } from "../../../contributing/adr/[record].paths.js";');
    put('runatlantis.io/es/contributing/example.md', '# Original');
    put('runatlantis.io/es/stale.md', '# Stale');
    const originalWrapper = read(`runatlantis.io/es/${wrapper}`);
    const originalLoader = read(`runatlantis.io/es/${loader}`);

    execFileSync('bash', [path.join(repository, 'scripts/stage-translation-sources.sh'), 'es'], { cwd: directory });
    assert.equal(existsSync(path.join(directory, '.lingo-src/en', wrapper)), false);
    assert.equal(read('.lingo-src/en/contributing/example.md'), '# Example');
    assert.equal(read(`.lingo-src/es/${wrapper}`), originalWrapper);

    const workflow = readFileSync(path.join(repository, '.github/workflows/translate-docs.yml'), 'utf8');
    const command = workflow.match(/^\s+(rsync -a --delete .+)$/m)[1];
    const copyBack = () => execFileSync('bash', ['-c', `dir="$1"; locale="$2"; ${command}`, 'copy-translations', '.lingo-src/es/', 'es'], { cwd: directory });
    put(`.lingo-src/es/${wrapper}`, '# Unexpected translation');
    put(`.lingo-src/es/${loader}`, '// Unexpected loader');
    put('.lingo-src/es/contributing/example.md', '# Traducido');
    rmSync(path.join(directory, '.lingo-src/es/stale.md'));
    copyBack();
    assert.equal(read(`runatlantis.io/es/${wrapper}`), originalWrapper);
    assert.equal(read(`runatlantis.io/es/${loader}`), originalLoader);
    assert.equal(read('runatlantis.io/es/contributing/example.md'), '# Traducido');
    assert.equal(existsSync(path.join(directory, 'runatlantis.io/es/stale.md')), false);

    rmSync(path.join(directory, '.lingo-src/es/contributing/adr'), { recursive: true });
    copyBack();
    assert.equal(read(`runatlantis.io/es/${wrapper}`), originalWrapper);
    assert.equal(read(`runatlantis.io/es/${loader}`), originalLoader);
  } finally {
    rmSync(directory, { recursive: true, force: true });
  }
});

