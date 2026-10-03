// Copyright 2026 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

(() => {
  const form = document.getElementById('dashboard-filters');
  const search = document.getElementById('metadata-search');
  const fields = ['repo', 'workspace', 'user'];
  const selects = fields.map(field => document.getElementById(`${field}-filter`));
  const pulls = [...document.querySelectorAll('.pull-row')];
  const locks = [...document.querySelectorAll('.lock-list li[data-repo]')];
  const rows = [...pulls, ...locks];
  const repositories = [...document.querySelectorAll('.repository')];
  const collapsedState = new Map();
  let wasFiltering = false;

  // Use the actual rendered metadata for both dropdowns and search, keeping
  // repository/workspace/user values out of HTML strings built by JavaScript.
  fields.forEach((field, index) => {
    const values = [...new Set(rows.map(row => row.dataset[field]).filter(Boolean))].sort((a, b) => a.localeCompare(b));
    values.forEach(value => selects[index].add(new Option(value, value)));
  });

  const normalize = value => value.normalize('NFKD').replace(/\p{M}/gu, '').toLowerCase();
  const metadata = new Map(rows.map(row => [row, fields.map(field => normalize(row.dataset[field]))]));

  // A subsequence match accepts abbreviations such as "infr prod octo" without
  // a search dependency. Anchor abbreviations to a name segment so "octo"
  // doesn't also match the scattered o-c-t-o letters in "production".
  // Every term must match at least one metadata field.
  function fuzzyMatch(term, value) {
    if (value.includes(term)) return true;
    return value.split(/[\/_.\s-]+/u).some(segment => {
      if (segment[0] !== term[0]) return false;
      let next = 0;
      for (const character of segment) {
        if (character === term[next]) next++;
        if (next === term.length) return true;
      }
      return false;
    });
  }

  const requestCount = entries => new Set(entries.map(row => row.dataset.requestId).filter(Boolean)).size;

  function applyFilters() {
    const terms = normalize(search.value).trim().split(/\s+/).filter(Boolean);
    const active = terms.length > 0 || selects.some(select => select.value !== '');
    if (active && !wasFiltering) repositories.forEach(repo => collapsedState.set(repo, repo.open));

    rows.forEach(row => {
      const exactMatch = fields.every((field, index) => !selects[index].value || row.dataset[field] === selects[index].value);
      const searchMatch = terms.every(term => metadata.get(row).some(value => fuzzyMatch(term, value)));
      row.hidden = !(exactMatch && searchMatch);
    });

    repositories.forEach(repo => {
      repo.querySelectorAll('.workspace').forEach(workspace => {
        const count = [...workspace.querySelectorAll('.pull-row')].filter(row => !row.hidden).length;
        workspace.hidden = count === 0;
        workspace.querySelector('.workspace-count').textContent = `${count} entr${count === 1 ? 'y' : 'ies'}`;
      });
      repo.hidden = ![...repo.querySelectorAll('.workspace')].some(workspace => !workspace.hidden);
      if (active && !repo.hidden) repo.open = true;
      else if (!active && wasFiltering) repo.open = collapsedState.get(repo);
    });

    const visiblePulls = pulls.filter(row => !row.hidden);
    const visibleLocks = locks.filter(row => !row.hidden);
    document.getElementById('pull-count').textContent = requestCount(visiblePulls);
    document.getElementById('lock-count').textContent = visibleLocks.length;
    document.getElementById('no-pull-matches').hidden = pulls.length === 0 || visiblePulls.length > 0;
    document.getElementById('no-lock-matches').hidden = locks.length === 0 || visibleLocks.length > 0;
    document.querySelector('.lock-list').hidden = locks.length > 0 && visibleLocks.length === 0;
    document.getElementById('clear-filters').disabled = !active;
    const hookGroups = pulls.filter(row => !row.dataset.requestId);
    const hookSummary = hookGroups.length ? ` and ${hookGroups.filter(row => !row.hidden).length} of ${hookGroups.length} workflow hook groups` : '';
    document.getElementById('filter-results').textContent = `Showing ${requestCount(visiblePulls)} of ${requestCount(pulls)} pull requests and ${visibleLocks.length} of ${locks.length} locks${hookSummary}`;
    wasFiltering = active;
  }

  form.addEventListener('submit', event => event.preventDefault());
  search.addEventListener('input', applyFilters);
  selects.forEach(select => select.addEventListener('change', applyFilters));
  form.addEventListener('reset', () => {
    search.value = '';
    selects.forEach(select => { select.value = ''; });
    applyFilters();
  });
  applyFilters();
  form.hidden = false;
})();
