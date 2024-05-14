import { test } from '@playwright/test';

test('page should load without errors', async ({ page }) => {
  // Listen for any errors that occur within the page
  page.on('pageerror', error => {
    console.error('Page error:', error.message);
    throw new Error(`Page error: ${error.message}`);
  });

  // Navigate to the URL
  await page.goto('http://localhost:8080/');
});
