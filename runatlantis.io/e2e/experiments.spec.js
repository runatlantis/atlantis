import { test, expect } from '@playwright/test';

const origin = process.env.ATLANTIS_VERSIONED_DOCS_URL;
const main = '/versions/main/docs/server-configuration.html';
test.use({ channel: process.env.ATLANTIS_DOCS_BROWSER || 'chromium', viewport: { width: 1440, height: 1000 } });

test.describe('experiment documentation', () => {
  test.skip(!origin, 'Set ATLANTIS_VERSIONED_DOCS_URL to the local versioned build.');

  test('keeps current-site ADR links on VitePress client navigation', async ({ page }) => {
    const normal = process.env.ATLANTIS_NORMAL_DOCS_URL;
    test.skip(!normal, 'Set ATLANTIS_NORMAL_DOCS_URL to the normal website build.');
    await page.goto(`${normal}/docs/experiments.html`);
    await page.evaluate(() => { window.navigationMarker = 'preserved'; });
    const link = page.getByRole('link', { name: 'feature lifecycle ADR', exact: true });
    await expect(link).toHaveAttribute('href', '/contributing/adr/0004-govern-feature-lifecycles.html');
    await expect(link).not.toHaveAttribute('target', '_self');
    await link.click();
    await expect(page).toHaveURL(`${normal}/contributing/adr/0004-govern-feature-lifecycles.html`);
    expect(await page.evaluate(() => window.navigationMarker)).toBe('preserved');
  });

  test('serves experiment documentation only in versions whose source contains it', async ({ page, request }) => {
    const versions = await (await request.get(`${origin}/doc-versions.json`)).json();
    for (const release of versions) {
      const available = release.pages.includes('docs/experiments.md');
      const response = await request.get(`${origin}/versions/${release.version}/docs/experiments.html`);
      expect(response.status(), release.version).toBe(available ? 200 : 404);
    }
    await page.goto(`${origin}/versions/main/docs/experiments.html`);
    await expect(page.locator('h1#experiments')).toBeVisible();
    await expect(page.getByRole('columnheader', { name: 'Releases', exact: true })).toBeVisible();
    await expect(page.locator('main table')).toContainText('v0.45.0+');
    await expect(page.locator('main table')).toContainText('unreleased (Main only)');
    await expect(page.locator('main')).not.toContainText('Durable storage');
    await expect(page.locator('.VPSidebar').getByRole('link', { name: 'Experiments', exact: true })).toBeVisible();
    const latest = versions.find((release) => release.version !== 'main');
    await page.getByLabel('Documentation version', { exact: true }).selectOption(latest.version);
    if (!latest.pages.includes('docs/experiments.md')) {
      await expect(page.getByRole('status')).toContainText('That page is unavailable');
      await expect(page.locator('.VPSidebar').getByRole('link', { name: 'Experiments', exact: true })).toHaveCount(0);
    }
  });

  test('identifies drift configuration as experimental and links its limitations', async ({ page }, testInfo) => {
    await page.emulateMedia({ colorScheme: 'dark' });
    await page.goto(`${origin}${main}`);
    for (const flag of ['enable-drift-detection', 'enable-drift-remediation']) {
      const heading = page.locator(`h3#${flag}`);
      await expect(heading).toContainText('Experimental');
      await heading.scrollIntoViewIfNeeded();
      await page.screenshot({ path: testInfo.outputPath(`${flag}.png`) });
      await heading.locator('+ p a').click();
      await expect(page).toHaveURL(`${origin}/versions/main/docs/experiments.html#drift-detection`);
      await expect(page.locator('#drift-detection')).toBeVisible();
      await page.goBack();
    }
    await page.goto(`${origin}/versions/main/docs/api-endpoints.html#drift-detection-and-remediation-alpha`);
    const heading = page.locator('#drift-detection-and-remediation-alpha');
    await expect(heading).toContainText('Drift Detection And Remediation (Experimental)');
    await expect(heading).toBeInViewport();
    const warning = heading.locator('+ .warning');
    await expect(warning.locator('.custom-block-title')).toHaveText('EXPERIMENTAL FEATURE - SUBJECT TO CHANGE');
    await page.screenshot({ path: testInfo.outputPath('experimental-drift-api.png') });
    await warning.getByRole('link', { name: 'drift detection experiment', exact: true }).click();
    await expect(page).toHaveURL(`${origin}/versions/main/docs/experiments.html#drift-detection`);
  });

  test('keeps API and drift experiments on separate trackers and graduation checklists', async ({ page }) => {
    await page.goto(`${origin}/versions/main/docs/experiments.html`);
    const table = page.locator('main table');
    const api = table.locator('tbody tr').filter({ has: page.getByRole('link', { name: 'API', exact: true }) });
    const drift = table.locator('tbody tr').filter({ has: page.getByRole('link', { name: 'Drift detection', exact: true }) });
    await expect(api).toHaveCount(1);
    await expect(drift).toHaveCount(1);
    await expect(api.getByRole('link', { name: 'API development #6162', exact: true })).toHaveAttribute('href', 'https://github.com/runatlantis/atlantis/issues/6162');
    await expect(drift.getByRole('link', { name: 'Development #7033', exact: true })).toHaveAttribute('href', 'https://github.com/runatlantis/atlantis/issues/7033');
    for (const row of [api, drift]) {
      await expect(row.getByRole('link', { name: 'ADR 0002', exact: true })).toHaveAttribute('href', '/contributing/adr/0002-api-enhancement-drift-detection.html');
    }
    await expect(page.locator('#api-graduation-checklist + ul')).toContainText('compatibility of legacy response bodies');
    await expect(page.locator('#drift-detection-graduation-checklist + ul')).toContainText('drift classification');
    await expect(page.locator('main .warning')).toContainText('Existing API and drift installations must add the matching experiment names');
    await expect(page.locator('main')).not.toContainText('experiment opt-in is optional');

    await page.goto(`${origin}/versions/main/docs/api-endpoints.html`);
    await page.locator('main .warning').first().getByRole('link', { name: 'API experiment', exact: true }).click();
    await expect(page).toHaveURL(`${origin}/versions/main/docs/experiments.html#api`);
    await page.goto(`${origin}${main}`);
    await expect(page.locator('h3#api-secret')).toContainText('Experimental');
    await page.locator('h3#api-secret + p').getByRole('link', { name: 'API experiment', exact: true }).click();
    await expect(page).toHaveURL(`${origin}/versions/main/docs/experiments.html#api`);
  });

  test('links experiment designs to current ADRs', async ({ page }) => {
    await page.goto(`${origin}/versions/main/docs/experiments.html`);
    await page.getByRole('link', { name: 'feature lifecycle ADR', exact: true }).click();
    await expect(page).toHaveURL(`${origin}/contributing/adr/0004-govern-feature-lifecycles.html`);
    await expect(page.locator('#status + p')).not.toHaveText('');
    await expect(page.getByLabel('Documentation version', { exact: true })).toHaveCount(0);
  });
});
