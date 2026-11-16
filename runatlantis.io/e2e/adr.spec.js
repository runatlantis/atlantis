import { test, expect } from '@playwright/test';

const origin = process.env.ATLANTIS_NORMAL_DOCS_URL || 'http://127.0.0.1:8080';
test.use({ channel: process.env.ATLANTIS_DOCS_BROWSER || 'chromium' });

test('publishes current ADRs under Contributing in both locales', async ({ page, request }) => {
  for (const [locale, label] of [['', 'Architecture Decisions'], ['/es', 'Decisiones de arquitectura']]) {
    await page.goto(`${origin}${locale}/contributing.html`);
    await page.locator('.VPSidebar').getByRole('link', { name: label, exact: true }).click();
    await expect(page).toHaveURL(`${origin}${locale}/contributing/adr/`);
    await expect(page.locator('h1#architecture-decision-records')).toBeVisible();
    const row = page.locator('main tbody tr').first();
    const status = await row.getByRole('cell').last().innerText();
    expect(status.trim()).not.toBe('');
    const record = row.getByRole('link').first();
    const filename = (await record.getAttribute('href')).split('/').pop().replace(/\.html$/, '.md');
    await expect(page.locator('main a[href="https://github.com/runatlantis/atlantis/blob/main/docs/adr/template.md"]')).toBeVisible();
    expect((await request.get(`${origin}${locale}/contributing/adr/template.html`)).status()).toBe(404);
    await record.click();
    await expect(page.locator('#status + p')).toHaveText(status);
    await expect(page.getByRole('link', { name: 'View source', exact: true })).toHaveAttribute('href', `https://github.com/runatlantis/atlantis/blob/main/docs/adr/${filename}`);
    if (locale) await expect(page.getByText('Registro original en inglés.')).toBeVisible();
  }
});
