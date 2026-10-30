import { test, expect } from '@playwright/test';
import { execFileSync } from 'node:child_process';

const origin = process.env.ATLANTIS_VERSIONED_DOCS_URL;
const older = '/versions/v0.28.0/docs/server-configuration.html';
const newer = '/versions/v0.48.0/docs/server-configuration.html';
const main = '/versions/main/docs/server-configuration.html';
test.use({ channel: process.env.ATLANTIS_DOCS_BROWSER || 'chromium', viewport: { width: 1440, height: 1000 } });

test.describe('versioned configuration documentation', () => {
  test.skip(!origin, 'Set ATLANTIS_VERSIONED_DOCS_URL to the local versioned build.');

  test('leaves the normal website build on its original documentation route', async ({ page }) => {
    const normal = process.env.ATLANTIS_NORMAL_DOCS_URL;
    test.skip(!normal, 'Set ATLANTIS_NORMAL_DOCS_URL to the normal website build.');
    await page.goto(`${normal}/docs/server-configuration.html`, { waitUntil: 'networkidle' });
    await expect(page).toHaveURL(`${normal}/docs/server-configuration.html`);
    await expect(page.getByLabel('Documentation version', { exact: true })).toHaveCount(0);
    await expect(page.locator('h1#server-configuration')).toBeVisible();
  });

  test('populates the selector from the generated release manifest', async ({ page, request }) => {
    const response = await request.get(`${origin}/doc-versions.json`);
    expect(response.ok()).toBeTruthy();
    const releases = await response.json();
    expect(releases.every(({ version }) => version === 'main' || version.localeCompare('v0.28.0', undefined, { numeric: true }) >= 0)).toBeTruthy();
    await page.goto(`${origin}/docs/server-configuration.html`);
    const options = await page.getByLabel('Documentation version', { exact: true }).locator('option').evaluateAll((items) => items.map((item) => item.value));
    expect(options).toEqual(['main', ...releases.filter((release) => release.version !== 'main').map((release) => release.version)]);
    const selector = page.getByLabel('Documentation version', { exact: true });
    await expect(selector.locator('option').first()).toHaveText('Main (next)');
    await expect(selector.locator('option').nth(1)).toHaveText(`${releases.find((release) => release.version !== 'main').version} (latest)`);
  });

  test('embeds only version names and loads the page inventory when switching', async ({ page }) => {
    let requests = 0;
    await page.route('**/doc-versions.json', (route) => { requests++; return route.continue(); });
    await page.goto(`${origin}/docs/server-configuration.html`);
    const selector = page.getByLabel('Documentation version', { exact: true });
    await expect(selector).toBeEnabled();
    expect(requests).toBe(0);
    const compact = await page.evaluate(() => window.__VP_SITE_DATA__.themeConfig.docVersions.every((release) => Object.keys(release).join(',') === 'version'));
    expect(compact).toBeTruthy();
    await selector.selectOption('v0.28.0');
    await expect(page).toHaveURL(`${origin}${older}`);
    expect(requests).toBe(1);
  });

  test('allows retrying a switch when the page inventory cannot be fetched', async ({ page }) => {
    let requests = 0;
    await page.route('**/doc-versions.json', (route) => ++requests === 1 ? route.fulfill({ status: 503, body: 'Unavailable' }) : route.continue());
    await page.goto(`${origin}${older}`);
    const selector = page.getByLabel('Documentation version', { exact: true });
    await selector.selectOption('main');
    await expect(page.getByRole('alert')).toContainText('Documentation versions could not be loaded');
    await expect(selector).toBeEnabled();
    await expect(selector).toHaveValue('v0.28.0');
    await selector.selectOption('main');
    await expect(page).toHaveURL(`${origin}${main}`);
    expect(requests).toBe(2);
  });

  test('serves latest release content at the root without a runtime redirect', async ({ page, request }) => {
    const releases = await (await request.get(`${origin}/doc-versions.json`)).json();
    const latest = releases.find((release) => release.version !== 'main').version;
    const root = await (await request.get(`${origin}/docs/server-configuration.html`)).text();
    const release = await (await request.get(`${origin}/versions/${latest}/docs/server-configuration.html`)).text();
    await page.goto(`${origin}/docs/server-configuration.html#port`, { waitUntil: 'networkidle' });
    const content = await page.evaluate(([root, release]) => [root, release].map((html) => new DOMParser().parseFromString(html, 'text/html').querySelector('.vp-doc').textContent), [root, release]);
    expect(content[0]).toEqual(content[1]);
    await expect(page).toHaveURL(`${origin}/docs/server-configuration.html#port`);
    const selector = page.locator('.VPNavBar').getByLabel('Documentation version', { exact: true });
    await expect(selector).toHaveValue(latest);
    await selector.selectOption('main');
    await expect(page).toHaveURL(`${origin}${main}#port`);
    await expect(page.getByLabel('Documentation notice')).toContainText('Development documentation from main.');
    await expect(page.getByRole('link', { name: 'Current website', exact: true })).toHaveCount(0);
    await page.goto(`${origin}/contributing.html`);
    await expect(page.getByLabel('Documentation version', { exact: true })).toHaveCount(0);
    await page.locator('.VPNavBar').getByRole('link', { name: 'Docs', exact: true }).click();
    await expect(page).toHaveURL(`${origin}/docs.html`);
    await expect(page.locator('main h1')).toHaveText('Atlantis Documentation');
    await expect(page.locator('main')).not.toContainText('Release snapshot');
  });

  test('deploy previews serve the working tree at root routes with Main selected', async ({ page, request }) => {
    const preview = process.env.ATLANTIS_PREVIEW_DOCS_URL;
    test.skip(!preview, 'Set ATLANTIS_PREVIEW_DOCS_URL to a CONTEXT=deploy-preview build.');
    for (const locale of ['', '/es']) {
      const root = await (await request.get(`${preview}${locale}/docs/server-configuration.html`)).text();
      const main = await (await request.get(`${preview}/versions/main${locale}/docs/server-configuration.html`)).text();
      await page.goto(`${preview}${locale}/docs/server-configuration.html#port`, { waitUntil: 'networkidle' });
      const content = await page.evaluate(([root, main]) => [root, main].map((html) => new DOMParser().parseFromString(html, 'text/html').querySelector('.vp-doc').textContent), [root, main]);
      expect(content[0]).toEqual(content[1]);
      await expect(page).toHaveURL(`${preview}${locale}/docs/server-configuration.html#port`);
      await expect(page.getByLabel('Documentation version', { exact: true })).toHaveValue('main');
      await expect(page.getByLabel('Documentation notice')).toContainText('Development documentation from main.');
      await expect(page.locator('.edit-link-button')).toHaveAttribute('href', `https://github.com/runatlantis/atlantis/edit/main/runatlantis.io${locale}/docs/server-configuration.md`);
      await page.getByLabel('Documentation version', { exact: true }).selectOption('v0.48.0');
      await expect(page).toHaveURL(`${preview}/versions/v0.48.0${locale}/docs/server-configuration.html#port`);
    }
  });

  test('preserves the original documentation landing pages in English and Spanish', async ({ page }) => {
    for (const prefix of ['', '/versions/v0.48.0', '/versions/main']) {
      for (const [locale, heading] of [['', 'Atlantis Documentation'], ['/es', 'Documentación de Atlantis']]) {
        await page.goto(`${origin}${prefix}${locale}/docs.html`);
        await expect(page.locator('main h1')).toHaveText(heading);
        await expect(page.locator('main')).not.toContainText('Release snapshot');
        await expect(page.locator('main a[href$="/docs/installation-guide.html"]')).toBeVisible();
      }
    }
  });

  test('preserves documentation landing pages in supported releases and returns to Main', async ({ page }) => {
    for (const [version, heading] of [['v0.28.0', 'Atlantis Documentation'], ['v0.40.0', 'Atlantis Documentation']]) {
      await page.goto(`${origin}/docs.html`);
      await page.getByLabel('Documentation version', { exact: true }).selectOption(version);
      await expect(page).toHaveURL(`${origin}/versions/${version}/docs.html`);
      await expect(page.locator('main h1')).toHaveText(heading);
      await expect(page.locator('main')).not.toContainText('Release snapshot');
      await page.getByLabel('Documentation version', { exact: true }).selectOption('main');
      await expect(page).toHaveURL(`${origin}/versions/main/docs.html`);
      await expect(page.locator('main h1')).toHaveText('Atlantis Documentation');
    }
  });

  test('preserves current-site edit links and their language', async ({ page }) => {
    for (const [locale, label] of [['', 'Edit this page'], ['/es', 'Editar esta página en GitHub']]) {
      for (const relative of ['contributing', 'docs/server-configuration']) {
        await page.goto(`${origin}${locale}/${relative}.html`);
        await expect(page.locator('.edit-link-button')).toHaveText(label);
        await expect(page.locator('.edit-link-button')).toHaveAttribute('href', `https://github.com/runatlantis/atlantis/edit/main/runatlantis.io${locale}/${relative}.md`);
      }
    }
  });

  test('preserves Git timestamps for current pages and latest-release documentation at the root', async ({ page, request }) => {
    const releases = await (await request.get(`${origin}/doc-versions.json`)).json();
    const latest = releases.find((release) => release.version !== 'main').version;
    for (const [ref, source] of [['HEAD', 'contributing.md'], ['HEAD', 'blog/2018/atlantis-0-4-4-now-supports-bitbucket.md'], [latest, 'docs/server-configuration.md']]) {
      const timestamp = Number(execFileSync('git', ['log', '-1', '--format=%at', ref, '--', `runatlantis.io/${source}`], { encoding: 'utf8' }).trim()) * 1000;
      await page.goto(`${origin}/${source.replace(/\.md$/, '.html')}`);
      await expect(page.locator('.last-updated time')).toHaveAttribute('datetime', new Date(timestamp).toISOString());
      const normal = process.env.ATLANTIS_NORMAL_DOCS_URL;
      if (normal && ref === 'HEAD') {
        const html = await (await request.get(`${normal}/${source.replace(/\.md$/, '.html')}`)).text();
        const normalDate = await page.evaluate((html) => new DOMParser().parseFromString(html, 'text/html').querySelector('.last-updated time')?.getAttribute('datetime'), html);
        expect(normalDate).toEqual(new Date(timestamp).toISOString());
      }
    }
  });

  test('links release sources to matching files in every locale', async ({ page }) => {
    for (const [version, relative, original] of [
      ['v0.48.0', 'docs/server-configuration.md', 'docs/server-configuration.md'],
      ['v0.48.0', 'es/docs/server-configuration.md', 'es/docs/server-configuration.md'],
      ['main', 'es/docs/api-endpoints.md', 'es/docs/api-endpoints.md'],
      ['v0.28.0', 'docs.md', 'docs.md'],
    ]) {
      await page.goto(`${origin}/versions/${version}/${relative.replace(/index\.md$/, '').replace(/\.md$/, '.html')}`);
      await expect(page.locator('.edit-link-button')).toHaveAttribute('href', `https://github.com/runatlantis/atlantis/blob/${version}/runatlantis.io/${original}`);
      if (version !== 'main') execFileSync('git', ['cat-file', '-e', `${version}:runatlantis.io/${original}`]);
    }
    await page.goto(`${origin}/versions/v0.48.0/`);
    await expect(page.locator('.edit-link-button')).toHaveAttribute('href', 'https://github.com/runatlantis/atlantis/blob/v0.48.0/runatlantis.io/docs.md');
  });

  test('restores the selector after navigating Back', async ({ page }) => {
    await page.goto(`${origin}${older}`);
    await page.getByLabel('Documentation version', { exact: true }).selectOption('main');
    await expect(page).toHaveURL(`${origin}${main}`);
    await page.goBack();
    const selector = page.getByLabel('Documentation version', { exact: true });
    await expect(selector).toBeEnabled();
    await expect(selector).toHaveValue('v0.28.0');
    await selector.selectOption('v0.48.0');
    await expect(page).toHaveURL(`${origin}${newer}`);
  });

  test('resets a pending selection when a cached page is restored', async ({ page }) => {
    await page.goto(`${origin}${older}`);
    let pendingRequest;
    await page.route(`${origin}${main}`, (route) => { pendingRequest = route; });
    const selector = page.getByLabel('Documentation version', { exact: true });
    await selector.selectOption('main');
    await expect.poll(() => Boolean(pendingRequest)).toBeTruthy();
    await expect(selector).toBeDisabled();
    await page.evaluate(() => window.dispatchEvent(new PageTransitionEvent('pageshow', { persisted: true })));
    await expect(selector).toBeEnabled();
    await expect(selector).toHaveValue('v0.28.0');
    await pendingRequest.abort();
  });

  test('links Spanish ADR navigation and selects Spanish release documentation', async ({ page, request }) => {
    const releases = await (await request.get(`${origin}/doc-versions.json`)).json();
    const latest = releases.find((release) => release.version !== 'main').version;
    await page.goto(`${origin}/es/contributing.html`);
    await page.locator('.VPSidebar').getByRole('link', { name: 'Decisiones de arquitectura', exact: true }).click();
    await expect(page).toHaveURL(`${origin}/es/contributing/adr/`);
    await page.goto(`${origin}/es/docs/server-configuration.html`);
    await expect(page.getByLabel('Documentation version', { exact: true })).toHaveValue(latest);
    await page.getByLabel('Documentation version', { exact: true }).selectOption('main');
    await expect(page).toHaveURL(`${origin}/versions/main/es/docs/server-configuration.html`);
    await page.getByLabel('Documentation version', { exact: true }).selectOption('v0.28.0');
    await expect(page).toHaveURL(`${origin}${older}`);
    await page.goto(`${origin}/versions/v0.48.0/es/`);
    await expect(page.locator('main h1')).toHaveText('Documentación de Atlantis');
  });

  test('places the selector beside Documentation and keeps mobile navigation usable', async ({ page }, testInfo) => {
    await page.emulateMedia({ colorScheme: 'dark' });
    for (const width of [320, 390, 768, 960, 1024, 1440]) {
      await page.setViewportSize({ width, height: 900 });
      for (const pathname of [older, '/docs/server-configuration.html', '/contributing.html']) {
        await page.goto(`${origin}${pathname}`);
        const selector = page.getByLabel('Documentation version', { exact: true });
        if (pathname === '/contributing.html') {
          await expect(selector).toHaveCount(0);
          expect(await page.evaluate(() => document.documentElement.scrollWidth)).toBeLessThanOrEqual(width);
          continue;
        }
        await expect(selector).toHaveCount(1);
        await expect(selector).toBeVisible();
        const control = await selector.boundingBox();
        const header = await page.locator('.VPNavBar').boundingBox();
        expect(control.y).toBeGreaterThanOrEqual(header.y);
        expect(control.y + control.height).toBeLessThanOrEqual(header.y + header.height);
        expect(control.x + control.width).toBeLessThanOrEqual(width);
        expect(await page.evaluate(() => document.documentElement.scrollWidth)).toBeLessThanOrEqual(width);
        await expect(page.locator('main select')).toHaveCount(0);
        await expect(page.getByLabel('Documentation notice')).toHaveCount(0);
        await expect(page.locator('main')).not.toContainText('Release snapshot');
        await expect(page.getByRole('link', { name: 'Current website', exact: true })).toHaveCount(0);
        if (width >= 768) {
          const documentation = await page.locator('.VPNavBarMenu').getByRole('link', { name: pathname === older ? 'Documentation' : 'Docs', exact: true }).boundingBox();
          expect(control.x).toBeGreaterThanOrEqual(documentation.x + documentation.width);
          expect(control.x - documentation.x - documentation.width).toBeLessThanOrEqual(16);
          const utilities = await page.locator('.VPNavBarAppearance:visible, .VPNavBarExtra:visible').first().boundingBox();
          expect(control.x + control.width).toBeLessThanOrEqual(utilities.x);
        }
        if ([390, 768, 1024, 1440].includes(width)) await page.screenshot({ path: testInfo.outputPath(`navbar-${width}.png`) });
      }
    }
  });

  test('switches through every release in the generated manifest', async ({ page, request }) => {
    const response = await request.get(`${origin}/doc-versions.json`);
    const releases = await response.json();
    test.setTimeout(30_000 + releases.length * 5_000);
    const errors = [];
    page.on('pageerror', (error) => errors.push(error.message));
    await page.goto(`${origin}/docs/server-configuration.html`);
    for (const { version } of releases.filter((release) => release.version !== 'main')) {
      await page.getByLabel('Documentation version', { exact: true }).selectOption(version);
      await expect(page).toHaveURL(`${origin}/versions/${version}/docs/server-configuration.html`);
      await expect(page.getByLabel('Documentation version', { exact: true })).toHaveValue(version);
      await expect(page.getByLabel('Documentation notice')).toHaveCount(0);
      await expect(page.locator('main h3 .VPBadge').filter({ hasText: /^v\d+\./ })).toHaveCount(0);
    }
    expect(errors).toEqual([]);
  });

  test('switches releases, preserves a common section, and supports direct reload', async ({ page }) => {
    await page.goto(`${origin}${main}#port`);
    await page.getByLabel('Documentation version', { exact: true }).selectOption('v0.28.0');
    await expect(page).toHaveURL(`${origin}${older}#port`);
    await expect(page.getByLabel('Documentation version', { exact: true })).toHaveValue('v0.28.0');
    await expect(page.getByLabel('Documentation notice')).toHaveCount(0);
    await page.reload();
    await expect(page.locator('#port')).toBeVisible();
    await page.getByLabel('Documentation version', { exact: true }).selectOption('v0.48.0');
    await expect(page).toHaveURL(`${origin}${newer}#port`);
  });

  test('drops malformed fragments and still switches releases', async ({ page }) => {
    await page.goto(`${origin}${main}#%E0`);
    await page.getByLabel('Documentation version', { exact: true }).selectOption('v0.28.0');
    await expect(page).toHaveURL(`${origin}${older}`);
    await expect(page.getByLabel('Documentation version', { exact: true })).toHaveValue('v0.28.0');
    await expect(page.getByLabel('Documentation version', { exact: true })).toBeEnabled();
    await expect(page.getByRole('alert')).toHaveCount(0);
  });

  test('keeps real v0.28 relative links and sidebar navigation inside the release', async ({ page }, testInfo) => {
    const failures = [];
    page.on('pageerror', (error) => failures.push(error.message));
    await page.goto(`${origin}${older}`);
    await expect(page.getByLabel('Documentation version', { exact: true })).toBeEnabled();
    await page.screenshot({ path: testInfo.outputPath('versioned-config.png') });
    const config = page.locator('main a[href="./server-side-repo-config.html#reference"]').first();
    await config.click();
    await expect(page).toHaveURL(`${origin}/versions/v0.28.0/docs/server-side-repo-config.html#reference`);
    await expect(page.locator('#reference')).toBeVisible();
    const sidebarLinks = await page.locator('.VPSidebar a[href]').evaluateAll((links) => links.map((a) => a.getAttribute('href')));
    expect(sidebarLinks.length).toBeGreaterThan(0);
    expect(sidebarLinks.every((href) => href.startsWith('/versions/v0.28.0/') || /^\/(contributing|blog)(\/|\.)/.test(href))).toBeTruthy();
    await expect(page.getByRole('button', { name: 'Search' }).first()).toBeVisible();
    expect(failures).toEqual([]);
  });

  test('keeps contributing and blog navigation on the current site across releases', async ({ page }, testInfo) => {
    for (const version of ['v0.28.0', 'v0.48.0']) {
      await page.goto(`${origin}/versions/${version}/docs/server-configuration.html`);
      const sidebar = page.locator('.VPSidebar');
      await expect(sidebar.getByRole('link', { name: 'Contributing', exact: true })).toHaveAttribute('href', '/contributing.html');
      await expect(sidebar.getByRole('link', { name: 'Blog', exact: true })).toHaveAttribute('href', '/blog.html');
      if (version === 'v0.28.0') {
        await sidebar.getByRole('link', { name: 'Glossary', exact: true }).scrollIntoViewIfNeeded();
        await page.screenshot({ path: testInfo.outputPath('shared-navigation.png') });
      }
      await sidebar.getByRole('link', { name: 'Glossary', exact: true }).click();
      await expect(page).toHaveURL(`${origin}/contributing/glossary.html`);
      await expect(page.locator('h1#glossary')).toBeVisible();
    }
  });

  test('keeps search results inside the selected release', async ({ page }) => {
    await page.goto(`${origin}${older}`);
    await expect(page.getByLabel('Documentation version', { exact: true })).toBeEnabled();
    await page.getByRole('button', { name: 'Search' }).first().click();
    await page.getByRole('searchbox').fill('server configuration');
    const results = page.locator('#localsearch-list a.result');
    await expect(results.first()).toBeVisible();
    const hrefs = await results.evaluateAll((links) => links.map((a) => a.getAttribute('href')));
    expect(hrefs.every((href) => href.startsWith('/versions/v0.28.0/'))).toBeTruthy();
    await results.first().click();
    await expect(page).toHaveURL(new RegExp('/versions/v0\\.28\\.0/'));
  });

  test('explains unavailable pages and sections when switching to an older release', async ({ page }) => {
    await page.goto(`${origin}/versions/main/docs/sending-notifications-via-webhooks.html`);
    await page.getByLabel('Documentation version', { exact: true }).selectOption('v0.28.0');
    const fallback = new URL(`${origin}/versions/v0.28.0/docs.html`);
    fallback.searchParams.set('missing-page', 'docs/sending-notifications-via-webhooks.md');
    await expect(page).toHaveURL(fallback.href);
    await expect(page.locator('main h1')).toHaveText('Atlantis Documentation');
    await expect(page.getByRole('status')).toContainText('That page is unavailable in this release. Showing the documentation index.');
    await page.goto(`${origin}${main}#enable-profiling-api`);
    await page.getByLabel('Documentation version', { exact: true }).selectOption('v0.28.0');
    await expect(page.getByRole('status')).toContainText('section “enable-profiling-api” is unavailable');
    await expect(page).toHaveURL(new RegExp(`${older.replaceAll('.', '\\.')}\\?missing-section=enable-profiling-api$`));
  });

  test('links release navigation to current ADRs', async ({ page }, testInfo) => {
    await page.emulateMedia({ colorScheme: 'dark' });
    await page.goto(`${origin}${older}`);
    await page.locator('.VPSidebar').getByRole('link', { name: 'Architecture Decisions', exact: true }).click();
    await expect(page).toHaveURL(`${origin}/contributing/adr/`);
    await expect(page.locator('h1#architecture-decision-records')).toBeVisible();
    await expect(page.getByLabel('Documentation version', { exact: true })).toHaveCount(0);
    await page.screenshot({ path: testInfo.outputPath('adrs-index.png') });
    const record = page.locator('main tbody tr').first().getByRole('link').first();
    const recordPath = new URL(await record.getAttribute('href'), page.url()).pathname;
    await record.click();
    await expect(page).toHaveURL(`${origin}${recordPath}`);
    await expect(page.getByRole('link', { name: 'View source', exact: true })).toBeVisible();
    await page.screenshot({ path: testInfo.outputPath('adr-record.png') });
    await page.goto(`${origin}/es${recordPath}`);
    await expect(page.getByText('Registro original en inglés.')).toBeVisible();
    await expect(page.getByLabel('Documentation version', { exact: true })).toHaveCount(0);
  });
});
