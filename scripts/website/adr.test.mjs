import assert from 'node:assert/strict';
import test from 'node:test';
import records from '../../runatlantis.io/contributing/adr/[record].paths.js';

test('ADR routes publish the index and numbered records while linking to the template on GitHub', () => {
  const pages = records.paths();
  assert.ok(pages.every(({ params }) => params.record === 'index' || /^\d{4}-/.test(params.record)));
  assert.ok(!pages.some(({ params }) => params.record === 'template'));
  assert.ok(pages.find(({ params }) => params.record === 'index').content.includes('https://github.com/runatlantis/atlantis/blob/main/docs/adr/template.md'));
});

