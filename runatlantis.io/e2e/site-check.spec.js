import { test, expect } from '@playwright/test';

const origin = process.env.ATLANTIS_NORMAL_DOCS_URL || 'http://127.0.0.1:8080';

test('page should load without errors', async ({ page }) => {
  // Listen for any errors that occur within the page
  page.on('pageerror', error => {
    console.error('Page error:', error.message);
    throw new Error(`Page error: ${error.message}`);
  });

  // page.goto resolves on 404 and 500 responses, so check the status
  const response = await page.goto(`${origin}/`);
  expect(response?.status()).toBe(200);
  await expect(page.getByRole('heading', { level: 1, name: 'Terraform Pull Request Automation' })).toBeVisible();
});
